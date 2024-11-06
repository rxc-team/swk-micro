package stringx

import (
	"strings"

	"rxcsoft.cn/pit3/api/internal/common/charsetx"
)

func AddEllipsis(value string, lineTotalWords float64, line int) string {
	var result []string
	var runes []rune

	words := []rune(value)

	var currentWords float64 = 0

	for i, v := range words {
		// 如果是换行
		if v == 10 && len(runes) > 0 {
			result = append(result, string(runes))
			runes = []rune{}
			currentWords = 0
			continue
		}
		// 将当前文字添加到当前行中
		runes = append(runes, v)

		if charsetx.IsASCIIChar(v) {
			currentWords += 0.5
		} else {
			currentWords += 1
		}

		if currentWords >= lineTotalWords {
			result = append(result, string(runes))
			runes = []rune{}
			currentWords = 0
			continue
		}

		if i == len(words)-1 {
			result = append(result, string(runes))
			runes = []rune{}
			currentWords = 0
			continue
		}
	}

	if len(result) > line {
		result = result[:line]
	}

	return strings.Join(result, "\n")
}
