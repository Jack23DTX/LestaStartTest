package main

import (
	"log"
	"os"

	"LestaStartTest/internal/controllers"
	"LestaStartTest/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Загрузка переменных окружения
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	db.Init()
}

func main() {

	// Загрузка порта из .env
	port := os.Getenv("MAIN_PORT")

	// Инициализация роутеров и запуск сервера
	r := gin.Default()

	// Загрузка шаблонов страниц
	r.LoadHTMLGlob("internal/templates/*.html")

	// Публичные роутеры
	r.GET("/login", controllers.LoginPage)
	r.POST("/login", controllers.Login)
	r.GET("/register", controllers.RegisterPage)
	r.POST("/register", controllers.Register)

	protected := r.Group("/")
	protected.Use(controllers.AuthMiddleware())
	{
		// Веб-страницы
		protected.GET("/", controllers.UploadPage)
		protected.POST("/upload", controllers.UploadFileHandler)
		protected.GET("/logout", controllers.Logout)

		// Документы
		protected.GET("/documents", controllers.ListDocumentsAPI)
		protected.GET("/documents/:id", controllers.GetDocumentAPI)
		protected.GET("/documents/:id/statistics", controllers.DocumentStatisticsAPI)
		protected.DELETE("/documents/:id", controllers.DeleteDocumentAPI)

		// Коллекции
		protected.GET("/collections", controllers.ListCollectionsAPI)
		protected.GET("/collections/:id", controllers.GetCollectionAPI)
		protected.GET("/collections/:id/statistics", controllers.CollectionStatisticsAPI)
		protected.POST("/collections/:collection_id/:document_id", controllers.AddDocumentToCollectionAPI)
		protected.DELETE("/collections/:collection_id/:document_id", controllers.RemoveDocumentFromCollectionAPI)

		// Пользователь
		protected.PATCH("/user/:user_id", controllers.ChangePasswordAPI)
		protected.DELETE("/user/:user_id", controllers.DeleteUserAPI)
	}

	// Системные эндпойнты
	r.GET("/status", controllers.StatusHandler)
	r.GET("/metrics", controllers.MetricsHandler)
	r.GET("/version", controllers.VersionHandler)

	// Запуск сервера
	r.Run(port)
}
