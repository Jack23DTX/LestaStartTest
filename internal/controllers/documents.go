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

func ListDocumentsAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var documents []models.Document
	if err := db.DB.Where("user_id = ?", userID).Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := make([]gin.H, len(documents))
	for i, doc := range documents {
		response[i] = gin.H{
			"id":   doc.ID,
			"name": doc.Filename,
		}
	}

	c.JSON(http.StatusOK, gin.H{"documents": response})
}

func GetDocumentAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var document models.Document
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      document.ID,
		"name":    document.Filename,
		"content": document.Content,
	})
}

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
			allDocuments = append(allDocuments, doc.Content)
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

	tf := calculation.CountTf([]string{document.Content})
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
