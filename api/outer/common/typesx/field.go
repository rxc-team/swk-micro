package typesx

import "rxcsoft.cn/pit3/srv/database/proto/field"

// FieldList 字段排序
type FieldList []*field.Field

//排序规则：按displayOrder排序（由小到大）
func (list FieldList) Len() int {
	return len(list)
}

func (list FieldList) Less(i, j int) bool {
	return list[i].DisplayOrder < list[j].DisplayOrder
}

func (list FieldList) Swap(i, j int) {
	var temp *field.Field = list[i]
	list[i] = list[j]
	list[j] = temp
}
