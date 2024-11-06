package typesx

// 字段
type DownloadField struct {
	FieldId       string `json:"field_id"`
	FieldName     string `json:"field_name"`
	FieldType     string `json:"field_type"`
	IsImage       bool   `json:"is_image"`
	AsTitle       bool   `json:"as_title"`
	DisplayOrder  int64  `json:"display_order"`
	DisplayDigits int64  `json:"display_digits"`
	Precision     int64  `json:"precision"`
	Prefix        string `json:"prefix"`
	Format        string `json:"fromat"`
}

// FieldList 字段排序
type DownloadFields []*DownloadField

//排序规则：按displayOrder排序（由小到大）
func (list DownloadFields) Len() int {
	return len(list)
}

func (list DownloadFields) Less(i, j int) bool {
	return list[i].DisplayOrder < list[j].DisplayOrder
}

func (list DownloadFields) Swap(i, j int) {
	var temp *DownloadField = list[i]
	list[i] = list[j]
	list[j] = temp
}
