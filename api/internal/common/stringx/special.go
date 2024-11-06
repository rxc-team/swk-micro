package stringx

import (
	"regexp"
	"strings"
)

// 特殊字符检查，合法返回true,否则返回false
func SpecialCheck(value string, special string) bool {

	if len(special) == 0 {
		return true
	}

	specialReg := regexp.QuoteMeta(special)
	// 判断特殊字符是否包含减号
	hasMinus := strings.Contains(specialReg, "-")
	if hasMinus {
		specialReg = strings.Replace(specialReg, "-", "\\-", 1)
	}
	re := regexp.MustCompile("[" + specialReg + "]")
	return !re.MatchString(value)
}
