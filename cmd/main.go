package main

import (
	"LestaStartTest/internal/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("internal/templates/*.html")

	r.GET("/", controllers.UploadPage)
	r.POST("/", controllers.UploadFileHandler)

	r.Run(":8080")
}
