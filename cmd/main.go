package main

import (
	"log"
	"os"

	"LestaStartTest/internal/controllers"
	"LestaStartTest/internal/db"
	"LestaStartTest/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Раздаём папку static/ по корню
	r.Static("/static", "./static")

	// Swagger маршруты
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Публичные API эндпоинты
	r.POST("/login", controllers.LoginAPI)
	r.POST("/register", controllers.RegisterAPI)

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth(), middleware.Handle401())
	{
		// Документы
		protected.GET("/documents", controllers.ListDocumentsAPI)
		protected.POST("/documents/upload", controllers.UploadAPI)
		protected.GET("/documents/:id", controllers.GetDocumentAPI)
		protected.GET("/documents/:id/statistics", controllers.DocumentStatisticsAPI)
		protected.DELETE("/documents/:id", controllers.DeleteDocumentAPI)

		// Коллекции
		protected.POST("/collections", controllers.CreateCollectionAPI)
		protected.GET("/collections", controllers.ListCollectionsAPI)
		protected.GET("/collections/:id", controllers.GetCollectionAPI)
		protected.GET("/collections/:id/statistics", controllers.CollectionStatisticsAPI)
		protected.POST("/collection/:collection_id/:document_id", controllers.AddDocumentToCollectionAPI)
		protected.DELETE("/collection/:collection_id/:document_id", controllers.RemoveDocumentFromCollectionAPI)
		protected.DELETE("/collections/:id", controllers.DeleteCollectionAPI)

		// Пользователь
		protected.PATCH("/user/:user_id", controllers.ChangePasswordAPI)
		protected.DELETE("/user/:user_id", controllers.DeleteUserAPI)

		// Аутентификация (выход из аккаунта)
		protected.GET("/logout", controllers.LogoutAPI)
	}

	// Системные эндпойнты
	r.GET("/api/status", controllers.StatusHandler)
	r.GET("/api/metrics", controllers.MetricsHandler)
	r.GET("/api/version", controllers.VersionHandler)

	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Запуск сервера
	r.Run(port)
}
