package controllers

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"LestaStartTest/internal/calculation"
	"LestaStartTest/internal/db"
	"LestaStartTest/internal/models"

	"github.com/gin-gonic/gin"
)

// DocumentResponse - структура для ответа API
type DocumentResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content,omitempty"`
}

// ListDocumentsAPI – список документов
// @Summary Список документов
// @Description Возвращает все документы текущего пользователя.
// @Tags Документы
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "{"documents":[]DocumentResponse}"
// @Failure 500 {object} map[string]string "Database error"
// @Router /api/documents [get]
func ListDocumentsAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var documents []models.Document
	if err := db.DB.Where("user_id = ?", userID).Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := make([]DocumentResponse, len(documents))
	for i, doc := range documents {
		response[i] = DocumentResponse{
			ID:   doc.ID,
			Name: doc.Filename,
		}
	}

	c.JSON(http.StatusOK, gin.H{"documents": response})
}

// GetDocumentAPI – получение документа
// @Summary Получение документа
// @Description Возвращает имя и содержимое документа по его ID.
// @Tags Документы
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID документа"
// @Success 200 {object} DocumentResponse
// @Failure 404 {object} map[string]string "Document not found"
// @Router /api/documents/{id} [get]
func GetDocumentAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var document models.Document
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, DocumentResponse{
		ID:      document.ID,
		Name:    document.Filename,
		Content: document.Content,
	})
}

// DocumentStatisticsAPI – статистика документа
// @Summary TF‑IDF статистика документа
// @Description Рассчитывает TF‑IDF слова внутри всех коллекций, где есть этот документ.
// @Tags Документы
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID документа"
// @Success 200 {object} map[string]interface{} "{"document_id":int,"statistics":map[string]object}"
// @Failure 400 {object} map[string]string "Document is not in any collection"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Failed to find collections"
// @Router /api/documents/{id}/statistics [get]
func DocumentStatisticsAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var document models.Document
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	var collections []models.Collection
	if err := db.DB.Preload("Documents").Where("id IN (SELECT collection_id FROM collection_documents WHERE document_id = ?)", document.ID).
		Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find collections"})
		return
	}

	if len(collections) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document is not in any collection"})
		return
	}

	var allDocuments []string
	for _, col := range collections {
		var docs []models.Document
		if err := db.DB.Model(&col).Association("Documents").Find(&docs); err != nil {
			continue
		}
		for _, doc := range docs {
			allDocuments = append(allDocuments, doc.ProcessedContent)
		}
	}

	uniqueDocs := make(map[string]bool)
	for _, doc := range allDocuments {
		uniqueDocs[doc] = true
	}

	corpus := make([]string, 0, len(uniqueDocs))
	for doc := range uniqueDocs {
		corpus = append(corpus, doc)
	}

	tf := calculation.CountTf([]string{document.ProcessedContent})
	idf := calculation.CountIdf(corpus)

	stats := make(map[string]gin.H)
	for word := range tf {
		stats[word] = gin.H{
			"tf":  tf[word],
			"idf": idf[word],
		}
	}

	type wordStat struct {
		Word string
		TF   float64
		IDF  float64
	}

	var statsSlice []wordStat
	for word, data := range stats {
		statsSlice = append(statsSlice, wordStat{
			Word: word,
			TF:   data["tf"].(float64),
			IDF:  data["idf"].(float64),
		})
	}

	sort.Slice(statsSlice, func(i, j int) bool {
		return statsSlice[i].TF < statsSlice[j].TF
	})

	if len(statsSlice) > 50 {
		statsSlice = statsSlice[:50]
	}

	result := make(map[string]gin.H)
	for _, item := range statsSlice {
		result[item.Word] = gin.H{"tf": item.TF, "idf": item.IDF}
	}

	// проверка на пустую статистик
	if len(stats) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"document_id": document.ID,
			"message":     "No statistics available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id": document.ID,
		"statistics":  result,
	})
}

// DeleteDocumentAPI – удаление документа
// @Summary Удаление документа
// @Description Удаляет документ, все связи и физический файл.
// @Tags Документы
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID документа"
// @Success 200 {object} map[string]string "{"message":"Document deleted"}"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Failed to delete document"
// @Router /api/documents/{id} [delete]
func DeleteDocumentAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var document models.Document
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	if err := db.DB.Model(&document).Association("Collections").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear associations"})
		return
	}

	if err := db.DB.Delete(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	if err := os.Remove(document.OriginalPath); err != nil {
		log.Printf("Failed to remove file: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}

// HuffmanEncodeAPI – кодирование Хаффмана
// @Summary Кодирование документа алгоритмом Хаффмана
// @Description Возвращает закодированное представление содержимого документа.
// @Tags Документы
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID документа"
// @Success 200 {object} map[string]interface{} "{"document_id":int,"huffman_encoded":string}"
// @Failure 400 {object} map[string]string "Invalid document ID or content too large"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Encoding failed"
// @Router /api/documents/{id}/huffman [get]
func HuffmanEncodeAPI(c *gin.Context) {
	documentID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	id, err := strconv.Atoi(documentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
	}
	var document models.Document
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	const maxSize = 10 * 1024 * 1024 // ограничение по памяти документа (10MB)
	if len(document.Content) > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document content too large"})
		return
	}

	encodedContent, err := calculation.Encode(document.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id":     document.ID,
		"huffman_encoded": encodedContent,
	})
}
