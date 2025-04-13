package main

import (
	"LestaStartTest/internal/controllers"

	"github.com/gin-gonic/gin"
)

func main() {

	// Инициализация роутеров и запуск сервера
	r := gin.Default()

	// Загрузка шаблонов страниц
	r.LoadHTMLGlob("internal/templates/*.html")

	// Роутеры
	r.GET("/", controllers.UploadPage)
	r.POST("/", controllers.UploadFileHandler)

	// Запуск сервера
	r.Run(":8080")
}
