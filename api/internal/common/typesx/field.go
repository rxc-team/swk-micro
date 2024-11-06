package typesx

import (
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
)

// FieldList 字段排序
type FieldList []*field.Field

// DashboardList 字段排序
type DashboardList []*dashboard.Dashboard

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

//排序规则：按created_at排序（由小到大）
func (list DashboardList) Len() int {
	return len(list)
}

func (list DashboardList) Less(i, j int) bool {
	return list[i].CreatedAt < list[j].CreatedAt
}

func (list DashboardList) Swap(i, j int) {
	var temp *dashboard.Dashboard = list[i]
	list[i] = list[j]
	list[j] = temp
}
