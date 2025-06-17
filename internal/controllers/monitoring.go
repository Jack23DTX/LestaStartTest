package controllers

import (
	"net/http"
	"os"

	"LestaStartTest/internal/monitoring"

	"github.com/gin-gonic/gin"
)

// StatusHandler – статус приложения
// @Summary Статус приложения
// @Description Проверяет, что сервис запущен.
// @Tags Системные
// @Produce json
// @Success 200 {object} map[string]string "Application is running"
// @Router /api/status [get]
func StatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// MetricsHandler – метрики приложения
// @Summary Метрики обработки документов
// @Description Возвращает общее число обработанных документов и среднее время обработки (нс).
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

// VersionHandler – версия приложения
// @Summary Текущая версия приложения
// @Description Возвращает значение переменной окружения VERSION.
// @Tags Системные
// @Produce json
// @Success 200 {object} map[string]string "Application version"
// @Router /api/version [get]
func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": os.Getenv("VERSION")})
}
