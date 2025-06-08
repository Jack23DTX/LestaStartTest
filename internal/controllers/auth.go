package controllers

import (
	"net/http"
	"strconv"

	"LestaStartTest/internal/db"
	"LestaStartTest/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func LoginPage(c *gin.Context) {
	errorMsg := c.Query("error")
	c.HTML(http.StatusOK, "login.html", gin.H{"error": errorMsg})
}

func RegisterPage(c *gin.Context) {
	errorMsg := c.Query("error")
	c.HTML(http.StatusOK, "register.html", gin.H{"error": errorMsg})
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login?error=Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.Redirect(http.StatusFound, "/login?error=Invalid credentials")
		return
	}

	c.SetCookie("auth", strconv.Itoa(int(user.ID)), 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Проверка на существование пользователя
	var count int64
	db.DB.Model(&models.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		c.Redirect(http.StatusFound, "/register?error=User already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=Error registering user")
		return
	}

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login?error=Database error registering user")
		return
	}

	c.SetCookie("auth", strconv.Itoa(int(user.ID)), 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("auth")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(cookie)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("userID", uint(userID))
		c.Next()
	}
}

func Logout(c *gin.Context) {
	c.SetCookie("auth", "", -1, "/", "", false, true)

	c.Set("userID", nil)

	c.Redirect(http.StatusFound, "/login?message=You have been logged out")
}
