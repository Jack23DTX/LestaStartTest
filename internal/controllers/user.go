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

func ChangePasswordAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	requestedID, _ := strconv.Atoi(c.Param("user_id"))

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

func DeleteUserAPI(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	requestedID, _ := strconv.Atoi(c.Param("user_id"))

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
