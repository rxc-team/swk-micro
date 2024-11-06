/*
 * @Description:常量
 * @Author: RXC 廖云江
 * @Date: 2019-08-20 09:47:57
 * @LastEditors: RXC 廖云江
 * @LastEditTime: 2019-08-22 15:41:54
 */

package model

const (
	// TimeFormat 日期格式化format
	TimeFormat = "2006-01-02 03:04:05"
)

// GetReportNameKey 获取报表名的前缀
func GetReportNameKey(appID, rsID string) string {
	return "apps." + appID + ".reports." + rsID
}

// GetDashboardNameKey 获取仪表盘名的前缀
func GetDashboardNameKey(appID, rsID string) string {
	return "apps." + appID + ".dashboards." + rsID
}
