package stringx

import "strconv"

// StringToInt string转int
func StringToInt(in string) (out int) {
	out, _ = strconv.Atoi(in)
	return
}
