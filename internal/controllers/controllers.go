package controllers

import (
	"LestaStartTest/internal/calculation"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type WordData struct {
	Word string
	TF   float64
	IDF  float64
}

func UploadPage(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}

func UploadFileHandler(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при загрузке файла/файлов"})
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не были загружены файлы"})
	}
	fmt.Println("Загруженные файлы:", files)

	var documents []string

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

	tf := calculation.CountTf(documents)

	docs := documents
	idf := calculation.CountIdf(docs)

	// Инициализируем слайс для хранения слов и их метрик
	var wordData []WordData
	for word, tfValue := range tf {
		wordData = append(wordData, WordData{
			Word: word,
			TF:   tfValue,
			IDF:  idf[word],
		})
	}

	// Сортируем по уменьшению idf
	sort.Slice(wordData, func(i, j int) bool {
		return wordData[i].IDF > wordData[j].IDF
	})
	// Вывод 50 слов, в случае если их больше
	if len(wordData) > 50 {
		wordData = wordData[:50]
	}

	c.HTML(http.StatusOK, "result.html", gin.H{
		"words": wordData, // Вывод результата
	})

}
