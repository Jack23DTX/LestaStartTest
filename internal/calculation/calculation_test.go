package calculation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountTf(t *testing.T) {
	documents := []string{
		"это первая строка строка",
		"это вторая строка строка",
	}

	//первая	1/8	~ 0.125
	//вторая	1/8	~ 0.125
	//строка	4/8	~ 0.5
	//это		2/8	~ 0.25

	tf := CountTf(documents)

	assert.InDelta(t, 0.25, tf["это"], 0.0001, "Неверный tf для слова 'это'")
	assert.InDelta(t, 0.5, tf["строка"], 0.0001, "Неверный tf для слова 'строка'")
	assert.InDelta(t, 0.125, tf["первая"], 0.0001, "Неверный tf для слова 'первая'")
	assert.InDelta(t, 0.125, tf["вторая"], 0.0001, "Неверный tf для слова 'вторая'")
}

func TestCountIdf(t *testing.T) {
	documents := []string{
		"ослик суслик",
		"паукан ослик",
		"ослик суслик паукан",
	}

	idf := CountIdf(documents)

	assert.InDelta(t, 0.0, idf["ослик"], 0.0001, "Слово 'ослик' встречается в каждом документе, IDF приблизительно равен нулю")
	assert.True(t, idf["суслик"] > idf["ослик"], "Слово 'суслик' встречается реже, чем слово 'ослик'")
	assert.True(t, idf["паукан"] > 0.0, "Слово 'паукан' должно быть больше нуля")
}
