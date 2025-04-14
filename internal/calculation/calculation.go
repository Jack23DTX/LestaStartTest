package calculation

import (
	"math"
	"strings"
	"unicode"
)

// Удаление пунктуации и приведение к нижнему регистру
func PunctuationRemoveAndLower(s string) string {
	l := strings.ToLower(s)
	var b strings.Builder
	for _, r := range l {
		if !unicode.IsPunct(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CountTf - Вычисление TF
func CountTf(documents []string) map[string]float64 {

	wordCount := make(map[string]int)
	totalWords := 0
	for _, doc := range documents {
		words := strings.Fields(doc)
		for _, word := range words {
			wordCount[word]++
			totalWords++
		}
	}

	tf := make(map[string]float64)
	for word, count := range wordCount {
		tf[word] = float64(count) / float64(totalWords)
	}

	return tf
}

// CountIdf - Вычисление IDF
func CountIdf(documents []string) map[string]float64 {

	documentsCount := len(documents)
	wordsDocumentCount := make(map[string]int)

	for _, doc := range documents {
		words := strings.Fields(doc)
		seen := make(map[string]bool)
		for _, word := range words {
			if !seen[word] {
				wordsDocumentCount[word]++
				seen[word] = true
			}
		}
	}

	idf := make(map[string]float64)
	for word, count := range wordsDocumentCount {
		idf[word] = math.Log(float64(documentsCount) / float64(count))
	}
	return idf
}
