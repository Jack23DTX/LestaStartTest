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

// CreateCollectionAPI - создание новой коллекции
// @Summary Создание коллекции
// @Description Создает новую коллекцию для пользователя.
// @Tags Коллекции
// @Accept json
// @Produce json
// @Param CreateCollectionRequest body CreateCollectionRequest true "Данные коллекции"
// @Success 201 {object} map[string]string "Collection created"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Failed to create collection"
// @Router /collections [post]
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

// recalcCollectionIDF - пересчет IDF для коллекции
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

// ListCollectionsAPI - получение списка коллекций
// @Summary Список коллекций
// @Description Возвращает список коллекций, принадлежащих пользователю.
// @Tags Коллекции
// @Produce json
// @Success 200 {object} map[string]string "Collections"
// @Failure 500 {object} map[string]string "Database error"
// @Router /collections [get]
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

// GetCollectionAPI - получение коллекции по ID
// @Summary Получение коллекции
// @Description Возвращает информацию о коллекции.
// @Tags Коллекции
// @Produce json
// @Param id path int true "ID коллекции"
// @Success 200 {object} map[string]string "Collection details"
// @Failure 404 {object} map[string]string "Collection not found"
// @Router /collections/{id} [get]
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

// CollectionStatisticsAPI - получение статистики коллекции
// @Summary Статистика коллекции
// @Description Рассчитывает TF-IDF статистику для коллекции.
// @Tags Коллекции
// @Produce json
// @Param id path int true "ID коллекции"
// @Success 200 {object} map[string]string "Statistics"
// @Failure 404 {object} map[string]string "Collection not found"
// @Router /collections/{id}/statistics [get]
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

// AddDocumentToCollectionAPI - добавление документа в коллекцию
// @Summary Добавление документа в коллекцию
// @Description Добавляет документ в коллекцию пользователя.
// @Tags Коллекции
// @Produce json
// @Param collection_id path int true "ID коллекции"
// @Param document_id path int true "ID документа"
// @Success 200 {object} map[string]string "Document added to collection"
// @Failure 404 {object} map[string]string "Collection or Document not found"
// @Failure 500 {object} map[string]string "Failed to add document or update IDF"
// @Router /collection/{collection_id}/{document_id} [post]
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

// RemoveDocumentFromCollectionAPI - удаление документа из коллекции
// @Summary Удаление документа из коллекции
// @Description Удаляет документ из коллекции пользователя.
// @Tags Коллекции
// @Produce json
// @Param collection_id path int true "ID коллекции"
// @Param document_id path int true "ID документа"
// @Success 200 {object} map[string]string "Document removed from collection"
// @Failure 404 {object} map[string]string "Collection or Document not found"
// @Failure 500 {object} map[string]string "Failed to remove document"
// @Router /collection/{collection_id}/{document_id} [delete]
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

// DeleteCollectionAPI - удаление коллекции
// @Summary Удаление коллекции
// @Description Удаляет коллекцию и связанные данные.
// @Tags Коллекции
// @Produce json
// @Param id path int true "ID коллекции"
// @Success 200 {object} map[string]string "Collection deleted"
// @Failure 404 {object} map[string]string "Collection not found"
// @Failure 500 {object} map[string]string "Failed to delete collection or related data"
// @Router /collections/{id} [delete]
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
