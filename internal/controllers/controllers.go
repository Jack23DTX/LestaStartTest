package controllers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"LestaStartTest/internal/calculation"
	"LestaStartTest/internal/db"
	"LestaStartTest/internal/models"
	"LestaStartTest/internal/monitoring"

	"github.com/gin-gonic/gin"
)

// WordData - создание структуры для хранения информации о слове
type WordData struct {
	Word string
	TF   float64
	IDF  float64
}

// UploadPage - обработчик для формы загрузки файлов
func UploadPage(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}

// UploadFileHandler - обработчик для обработки текста и отображения таблицы
func UploadFileHandler(c *gin.Context) {
	start := time.Now() // Фиксируем время начала обработки
	userID := c.MustGet("userID").(uint)

	// Получение файлов
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting files"})
	}

	// Проверка на наличие загруженных файлов
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
	}

	var allContents []string
	var wg sync.WaitGroup
	processingErrors := make(chan error, len(files))
	uploadedDocuments := make([]models.Document, 0, len(files))

	// Создаем директорию пользователя
	userUploadDir := filepath.Join("uploads", strconv.Itoa(int(userID)))
	if err := os.MkdirAll(userUploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user directory"})
		return
	}

	var mu sync.Mutex

	// Обработка каждого файла в отдельной горутине
	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()

			// Открытие файла
			src, err := file.Open()
			if err != nil {
				processingErrors <- fmt.Errorf("error opening file %s: %w", file.Filename, err)
				return
			}
			defer src.Close()

			// Создание пути для сохранения
			filePath := filepath.Join(userUploadDir, file.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				processingErrors <- fmt.Errorf("error creating file %s: %w", filePath, err)
				return
			}
			defer dst.Close()

			// Копирование файла
			if _, err := io.Copy(dst, src); err != nil {
				processingErrors <- fmt.Errorf("error saving file %s: %w", filePath, err)
				return
			}

			// Чтение и обработка содержимого
			content, err := os.ReadFile(filePath)
			if err != nil {
				processingErrors <- fmt.Errorf("error reading file %s: %w", filePath, err)
				return
			}

			cleanContent := calculation.PunctuationRemoveAndLower(string(content))
			mu.Lock()
			allContents = append(allContents, cleanContent)
			uploadedDocuments = append(uploadedDocuments, models.Document{
				UserID:           userID,
				Filename:         file.Filename,
				OriginalPath:     filePath,
				ProcessedContent: cleanContent,
			})
			mu.Unlock()
		}(file)
	}

	// Ожидание завершения обработки файлов
	wg.Wait()
	close(processingErrors)

	// Проверка ошибок
	var errors []string
	for err := range processingErrors {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": errors})
		return
	}

	// Сохранение документов в БД
	tx := db.DB.Begin()
	for _, doc := range uploadedDocuments {
		if err := tx.Create(&doc).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error saving documents"})
			return
		}
	}
	tx.Commit()

	// Расчет статистики
	var tf map[string]float64
	var idf map[string]float64

	wg.Add(2)
	go func() {
		defer wg.Done()
		tf = calculation.CountTf(allContents)
	}()

	go func() {
		defer wg.Done()
		idf = calculation.CountIdf(allContents)
	}()
	wg.Wait()

	// Формирование результата
	var wordData []WordData
	for word, tfValue := range tf {
		wordData = append(wordData, WordData{
			Word: word,
			TF:   tfValue,
			IDF:  idf[word],
		})
	}

	sort.Slice(wordData, func(i, j int) bool {
		return wordData[i].IDF > wordData[j].IDF
	})

	if len(wordData) > 50 {
		wordData = wordData[:50]
	}

	// Получение информации о пользователе
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Отправка результата
	c.HTML(http.StatusOK, "result.html", gin.H{
		"words": wordData,
		"user":  user,
	})

	// Обновление метрик
	processingTime := time.Since(start).Nanoseconds()
	monitoring.UpdateMetrics(len(files), processingTime)
}
