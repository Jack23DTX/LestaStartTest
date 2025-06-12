package controllers

import (
	"net/http"
	"os"

	"LestaStartTest/internal/monitoring"

	"github.com/gin-gonic/gin"
)

// StatusHandler - статус приложения
// @Summary Статус приложения
// @Description Проверяет состояние приложения.
// @Tags Системные
// @Produce json
// @Success 200 {object} map[string]string "Application is running"
// @Router /api/status [get]
func StatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// MetricsHandler - метрики приложения
// @Summary Метрики приложения
// @Description Возвращает статистику обработки документов.
// @Tags Системные
// @Produce json
// @Success 200 {object} map[string]string "Application metrics"
// @Router /api/metrics [get]
func MetricsHandler(c *gin.Context) {
	totalDocs, avgTime := monitoring.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"total_processed_documents": totalDocs,
		"avg_processed_documents":   avgTime,
	})
}

// VersionHandler - версия приложения
// @Summary Версия приложения
// @Description Возвращает текущую версию приложения.
// @Tags Системные
// @Produce json
// @Success 200 {object} map[string]string "Application version"
// @Router /api/version [get]
func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": os.Getenv("VERSION")})
}
