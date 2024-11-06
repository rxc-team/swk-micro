package floatx

import (
	"fmt"
	"math"
)

// ToFixed 直接向下取整
func ToFixed(f float64, places int64) float64 {
	shift := math.Pow(10, float64(places))
	fv := 0.0000000001 + f //对浮点数产生.xxx999999999 计算不准进行处理
	return math.Floor(fv*shift) / shift
}

// ToFixedString 直接向下取整
func ToFixedString(f float64, places int64) string {
	shift := math.Pow(10, float64(places))
	fv := 0.0000000001 + f //对浮点数产生.xxx999999999 计算不准进行处理
	return fmt.Sprintf("%v", math.Floor(fv*shift)/shift)
}
