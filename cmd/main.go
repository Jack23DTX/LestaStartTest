package main

import (
	"LestaStartTest/internal/controllers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Загрузка переменных окружения
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {

	// Загрузка порта из .env
	port, exist := os.LookupEnv("MAIN_PORT")
	if !exist {
		log.Fatal("No ports in .env")
	}

	// Инициализация роутеров и запуск сервера
	r := gin.Default()

	// Загрузка шаблонов страниц
	r.LoadHTMLGlob("internal/templates/*.html")

	// Роутеры
	r.GET("/", controllers.UploadPage)
	r.POST("/", controllers.UploadFileHandler)

	// API-эндпойнты
	r.GET("/status", controllers.StatusHandler)
	r.GET("/metrics", controllers.MetricsHandler)
	r.GET("/version", controllers.VersionHandler)

	// Запуск сервера
	r.Run(port)
}
