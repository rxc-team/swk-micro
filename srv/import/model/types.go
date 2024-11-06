package model

// FileValue 文件类型
type FileValue struct {
	URL  string `json:"url" bson:"url"`
	Name string `json:"name" bson:"name"`
}
