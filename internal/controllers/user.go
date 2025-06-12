package controllers

import (
	"net/http"
	"os"
	"strconv"

	"LestaStartTest/internal/db"
	"LestaStartTest/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ChangePasswordAPI - изменение пароля пользователя
// @Summary Изменение пароля
// @Description Обновление пароля пользователя.
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Param NewPassword body string true "Новый пароль"
// @Success 200 {object} map[string]string "Password updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Password update failed or database error"
// @Router /user/{user_id} [patch]
func ChangePasswordAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	requestedID, _ := strconv.Atoi(c.Param("user_id"))
	if uint(requestedID) != userID {
		c.AbortWithStatus(403)
		return
	}

	if uint(requestedID) != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var input struct {
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password update failed"})
		return
	}

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
}

// DeleteUserAPI - удаление пользователя
// @Summary Удаление пользователя
// @Description Удаление пользователя, его документы и коллекции.
// @Tags Пользователь
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Success 200 {object} map[string]string "User deleted"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /user/{user_id} [delete]
func DeleteUserAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	requestedID, _ := strconv.Atoi(c.Param("user_id"))
	if userID != uint(requestedID) {
		c.AbortWithStatus(403)
		return
	}

	if uint(requestedID) != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	// Удаляем файлы пользователя
	var documents []models.Document
	db.DB.Where("user_id = ?", userID).Find(&documents)
	for _, doc := range documents {
		os.Remove(doc.OriginalPath) // Удаляем физический файл
	}

	db.DB.Where("collection_id IN (SELECT id FROM collections WHERE user_id = ?)", userID).Delete(&models.CollectionIDF{})

	tx := db.DB.Begin()
	tx.Where("user_id = ?", userID).Delete(&models.Document{})
	tx.Where("user_id = ?", userID).Delete(&models.Collection{})
	tx.Delete(&models.User{}, userID)
	tx.Commit()

	c.SetCookie("auth", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
