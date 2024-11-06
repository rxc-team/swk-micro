package typesx

// FieldInfo 字段信息
type FieldInfo struct {
	FieldID     string
	DataType    string
	AliasName   string
	DatastoreID string
	IsDynamic   bool
	Order       int64
}

// FieldInfoList 字段排序
type FieldInfoList []*FieldInfo

//排序规则：按displayOrder排序（由小到大）
func (list FieldInfoList) Len() int {
	return len(list)
}

func (list FieldInfoList) Less(i, j int) bool {
	return list[i].Order < list[j].Order
}

func (list FieldInfoList) Swap(i, j int) {
	var temp *FieldInfo = list[i]
	list[i] = list[j]
	list[j] = temp
}
