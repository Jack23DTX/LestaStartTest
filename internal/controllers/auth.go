package controllers

import (
	"LestaStartTest/internal/db"
	"LestaStartTest/internal/middleware"
	"LestaStartTest/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Структуры запросов

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Message string `json:"message"`
	UserID  uint   `json:"user_id,omitempty"`
}

// LoginAPI - аутентификация пользователя
// @Summary Аутентификация пользователя
// @Description Проверка учетных данных и генерация JWT токена.
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param AuthRequest body AuthRequest true "Данные для аутентификации"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /login [post]
func LoginAPI(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials (login or password)"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials (login or password)"})
		return
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged successfully",
		"token":   token,
	})
}

// RegisterAPI - регистрация нового пользователя
// @Summary Регистрация нового пользователя
// @Description Создание учетной записи пользователя.
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param AuthRequest body AuthRequest true "Данные для регистрации"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "User already exists"
// @Router /register [post]
func RegisterAPI(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверка на существование пользователя
	var count int64
	db.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error registering user"})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error to register user"})
		return
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User successfully registered",
		"token":   token,
	})
}

// LogoutAPI - выход из аккаунта пользователя
// @Summary Выход из аккаунта
// @Description Завершение сеанса пользователя.
// @Tags Пользователь
// @Produce json
// @Success 200 {object} map[string]string "Successfully logged out"
// @Router /logout [get]
func LogoutAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
