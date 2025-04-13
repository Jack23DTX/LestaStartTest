package controllers

import (
	"LestaStartTest/internal/calculation"
	"fmt"
	"io"
	"net/http"
	"sort"

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

	// Получение файлов
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при загрузке файла/файлов"})
	}

	// Проверка на наличие загруженных файлов
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не были загружены файлы"})
	}
	fmt.Println("Загруженные файлы:", files)

	var documents []string

	// Чтение содержимого
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при чтении файла/файлов"})
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при чтении файла"})
		}
		documents = append(documents, string(content))
	}

	// Вычисление TF и IDF
	tf := calculation.CountTf(documents)

	docs := documents
	idf := calculation.CountIdf(docs)

	// Инициализация слайса для хранения слов и их метрик
	var wordData []WordData
	for word, tfValue := range tf {
		wordData = append(wordData, WordData{
			Word: word,
			TF:   tfValue,
			IDF:  idf[word],
		})
	}

	// Сортировка по уменьшению idf
	sort.Slice(wordData, func(i, j int) bool {
		return wordData[i].IDF > wordData[j].IDF
	})

	// Корректировка количества слов в выборке, в случае если их больше 50-ти
	if len(wordData) > 50 {
		wordData = wordData[:50]
	}

	// Вывод результата
	c.HTML(http.StatusOK, "result.html", gin.H{
		"words": wordData,
	})

}
