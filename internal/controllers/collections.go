package controllers

import (
	"LestaStartTest/internal/calculation"
	"LestaStartTest/internal/db"
	"LestaStartTest/internal/models"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CreateCollectionRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateCollectionAPI - эндпоинт для создания коллекций
func CreateCollectionAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	collection := models.Collection{
		UserID: userID,
		Name:   req.Name,
	}

	if err := db.DB.Create(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   collection.ID,
		"name": collection.Name,
	})
}

// RecalcCollectionIDF - эндпоинт перерасчета
func recalcCollectionIDF(collectionID uint, userID uint) error {
	var texts []string
	err := db.DB.Table("documents").
		Select("processed_content").
		Joins("JOIN collection_documents cd ON cd.document_id = documents.id").
		Where("cd.collection_id = ? AND documents.user_id = ?", collectionID, userID).
		Pluck("processed_content", &texts).Error
	if err != nil {
		return err
	}
	idfMap := calculation.CountIdf(texts)

	tx := db.DB.Begin()
	tx.Where("collection_id = ?", collectionID).Delete(&models.CollectionIDF{})
	for w, v := range idfMap {
		tx.Create(&models.CollectionIDF{CollectionID: collectionID, Word: w, IDFValue: v})
	}
	return tx.Commit().Error
}

func ListCollectionsAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var collections []models.Collection
	if err := db.DB.Where("user_id = ?", userID).Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := make([]gin.H, len(collections))
	for i, col := range collections {
		response[i] = gin.H{
			"id":   col.ID,
			"name": col.Name,
		}
	}

	c.JSON(http.StatusOK, gin.H{"collections": response})
}

func GetCollectionAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var collection models.Collection
	if err := db.DB.Preload("Documents").
		Where("id = ? AND user_id = ?", id, userID).
		First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	documents := make([]gin.H, len(collection.Documents))
	for i, doc := range collection.Documents {
		documents[i] = gin.H{
			"id":   doc.ID,
			"name": doc.Filename,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        collection.ID,
		"name":      collection.Name,
		"documents": documents,
	})
}

func CollectionStatisticsAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	var col models.Collection
	if err := db.DB.Preload("Documents").
		Where("id = ? AND user_id = ?", id, userID).
		First(&col).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	var combinedText strings.Builder
	for _, doc := range col.Documents {
		combinedText.WriteString(doc.ProcessedContent)
		combinedText.WriteString(" ")
	}

	tf := calculation.CountTf([]string{combinedText.String()})

	var idfRecs []models.CollectionIDF
	db.DB.Where("collection_id = ?", id).Find(&idfRecs)
	idfMap := make(map[string]float64, len(idfRecs))
	for _, rec := range idfRecs {
		idfMap[rec.Word] = rec.IDFValue
	}

	stats := make(map[string]gin.H, len(tf))
	for word, tfVal := range tf {
		stats[word] = gin.H{"tf": tfVal, "idf": idfMap[word]}
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

	c.JSON(http.StatusOK, gin.H{
		"collection_id": col.ID,
		"statistics":    result,
	})
}

// AddDocumentToCollectionAPI - эндпоинт для добавления документа в коллекцию
func AddDocumentToCollectionAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	collectionID, _ := strconv.Atoi(c.Param("collection_id"))
	documentID, _ := strconv.Atoi(c.Param("document_id"))

	var col models.Collection
	if err := db.DB.Where("id = ? AND user_id = ?", collectionID, userID).
		First(&col).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	var doc models.Document
	if err := db.DB.Where("id = ? AND user_id = ?", documentID, userID).
		First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	var count int64
	db.DB.Model(&models.Collection{}).
		Where("id = ? AND user_id = ?", collectionID, userID).
		Count(&count)
	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Collection doesn't belong to user"})
		return
	}

	if err := db.DB.Model(&col).
		Association("Documents").
		Append(&doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add document"})
		return
	}

	if err := recalcCollectionIDF(uint(collectionID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update IDF"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document added to collection"})

}

// RemoveDocumentFromCollectionAPI - эндпоинт для удаления документа в коллекции
func RemoveDocumentFromCollectionAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	collectionID, _ := strconv.Atoi(c.Param("collection_id"))
	documentID, _ := strconv.Atoi(c.Param("document_id"))

	var col models.Collection
	if err := db.DB.Where("id = ? AND user_id = ?", collectionID, userID).
		First(&col).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	var doc models.Document
	if err := db.DB.Where("id = ? AND user_id = ?", documentID, userID).
		First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Удаляем один раз
	if err := db.DB.Model(&col).
		Association("Documents").
		Delete(&doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove document"})
		return
	}

	// Пересчёт IDF
	if err := recalcCollectionIDF(uint(collectionID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update IDF"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document removed from collection"})
}

// DeleteCollectionAPI - эндпоинт для удаления коллекций
func DeleteCollectionAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	collectionID, _ := strconv.Atoi(c.Param("id"))

	// Проверка владельца коллекции
	var collection models.Collection
	if err := db.DB.Where("id = ? AND user_id = ?", collectionID, userID).
		First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	// Удаление связанных данных
	if err := db.DB.Where("collection_id = ?", collectionID).
		Delete(&models.CollectionIDF{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete IDF records"})
		return
	}

	// Удаление коллекции
	if err := db.DB.Delete(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete in Database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collection deleted"})
}
