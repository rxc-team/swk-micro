package utils

import "regexp"

//匹配unique(组合)字段重复消息的所有field_id和field_value
func FieldMatch(regex, source string) []string {
	comp := regexp.MustCompile(regex)
	value := comp.FindAllString(source, -1)
	return value
}
