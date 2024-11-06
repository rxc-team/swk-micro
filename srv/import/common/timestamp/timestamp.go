package timestamp

import (
	"strings"
	"time"
)

func Timestamp() string {
	timestamp := time.Now().Format("20060102150405.000000")
	return strings.Replace(timestamp, ".", "", -1)
}
