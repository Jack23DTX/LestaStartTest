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

// Структуры для API-ответов

type UploadResponse struct {
	ID       uint   `json:"id"`
	Filename string `json:"filename"`
}

type WordStat struct {
	Word string  `json:"word"`
	Tf   float64 `json:"tf"`
	Idf  float64 `json:"idf"`
}

type UploadResult struct {
	User      models.User      `json:"user"`
	Documents []UploadResponse `json:"documents"`
	TopWords  []WordStat       `json:"top_words"`
}

// UploadAPI - загрузка файлов через API
// @Summary Загрузка файлов
// @Description Загружает файлы, обрабатывает их содержимое и сохраняет в базе данных.
// @Tags Документы
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Файлы для загрузки"
// @Success 200 {object} map[string]string "Files uploaded and processed successfully"
// @Failure 400 {object} map[string]string "Error getting files or no files uploaded"
// @Failure 500 {object} map[string]string "Failed to save files or process documents"
// @Router /documents/upload [post]
func UploadAPI(c *gin.Context) {
	start := time.Now() // Фиксируем время начала обработки
	userID := c.MustGet("userID").(uint)

	// Получение информации о пользователе
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

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

	var (
		allContents       []string
		wg                sync.WaitGroup
		mu                sync.Mutex
		errCh             = make(chan error, len(files))
		uploadedDocuments = make([]models.Document, 0, len(files))
	)

	// Создаем директорию пользователя
	userUploadDir := filepath.Join("uploads", strconv.Itoa(int(userID)))
	if err := os.MkdirAll(userUploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user directory"})
		return
	}

	// Обработка каждого файла в отдельной горутине
	for _, f := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()

			// Открытие файла
			src, err := file.Open()
			if err != nil {
				errCh <- fmt.Errorf("error opening file %s: %w", file.Filename, err)
				return
			}
			defer src.Close()

			// Создание пути для сохранения
			filePath := filepath.Join(userUploadDir, file.Filename)
			dst, err := os.Create(filePath)
			if err != nil {
				errCh <- fmt.Errorf("error creating file %s: %w", filePath, err)
				return
			}
			defer dst.Close()

			// Копирование файла
			if _, err := io.Copy(dst, src); err != nil {
				errCh <- fmt.Errorf("error saving file %s: %w", filePath, err)
				return
			}

			// Чтение и обработка содержимого
			content, err := os.ReadFile(filePath)
			if err != nil {
				errCh <- fmt.Errorf("error reading file %s: %w", filePath, err)
				return
			}
			cleanContent := calculation.PunctuationRemoveAndLower(string(content))

			mu.Lock()
			allContents = append(allContents, cleanContent)
			uploadedDocuments = append(uploadedDocuments, models.Document{
				UserID:           userID,
				Filename:         file.Filename,
				OriginalPath:     filePath,
				Content:          string(content),
				ProcessedContent: cleanContent,
			})
			mu.Unlock()
		}(f)
	}

	// Ожидание завершения обработки файлов
	wg.Wait()
	close(errCh)

	// Проверка ошибок
	var errors []string
	for err := range errCh {
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
	var wordStats []WordStat
	for word, tfValue := range tf {
		wordStats = append(wordStats, WordStat{
			Word: word,
			Tf:   tfValue,
			Idf:  idf[word],
		})
	}

	// Сортировка по IDF в порядке убывания
	sort.Slice(wordStats, func(i, j int) bool {
		return wordStats[i].Idf > wordStats[j].Idf
	})

	// Берем топ-50 слов
	if len(wordStats) > 50 {
		wordStats = wordStats[:50]
	}

	// Формируем ответ с документами
	documentsResponse := make([]UploadResponse, len(uploadedDocuments))
	for i, doc := range uploadedDocuments {
		documentsResponse[i] = UploadResponse{
			ID:       doc.ID,
			Filename: doc.Filename,
		}
	}

	// Полный ответ
	result := UploadResult{
		User:      user,
		Documents: documentsResponse,
		TopWords:  wordStats,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded and processed successfully",
		"data":    result,
	})

	// Обновление метрик
	processingTime := time.Since(start).Nanoseconds()
	monitoring.UpdateMetrics(len(files), processingTime)
}
