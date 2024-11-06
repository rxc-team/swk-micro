package excelx

import "math"

func GetAxisY(count int) string {
	arr := [27]string{"", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	mod := count % 26
	divisor := int(math.Floor(float64(count) / 26))
	if mod == 0 && divisor > 0 {
		mod = 26
		divisor -= 1
	}

	return arr[divisor] + arr[mod]
}
