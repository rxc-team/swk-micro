package csv

import (
	"net/http"
	"os"
)

// getContentType 获取图片的类型
func getContentType(filePath string) string {

	fo, err := os.Open(filePath)
	if err != nil {
		return ""
	}

	defer fo.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = fo.Read(buffer)
	if err != nil {
		return ""
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType
}
