package monitoring

import (
	"expvar"
)

var (
	totalDocs          = expvar.NewInt("total_processed_documents")
	totalProcessTimeNs = expvar.NewInt("total_processing_time_ns")
)

// UpdateMetrics - обновление метрик после обработки документов
func UpdateMetrics(docsProcessed int, processTimeNs int64) {
	totalDocs.Add(int64(docsProcessed))
	totalProcessTimeNs.Add(processTimeNs)
}

// GetMetrics - текущие значения метрик
func GetMetrics() (int64, float64) {
	docs := totalDocs.Value()
	timeNs := totalProcessTimeNs.Value()

	var avgTime float64
	if docs > 0 {
		avgTime = float64(timeNs) / float64(docs) / 1e6
	}

	return docs, avgTime
}
