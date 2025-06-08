package controllers

import (
	"net/http"

	"LestaStartTest/internal/monitoring"

	"github.com/gin-gonic/gin"
)

// StatusHandler - статус приложения
func StatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// MetricsHandler - метрики приложения
func MetricsHandler(c *gin.Context) {
	totalDocs, avgTime := monitoring.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"total_processed_documents": totalDocs,
		"avg_processed_documents":   avgTime,
	})
}

// VersionHandler - версия приложения
func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": "1.1.0"})
}
