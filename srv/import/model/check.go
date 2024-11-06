package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/option"
)

// checkExist 判断表头是否存在
func CheckExist(f string, header []string) bool {
	for i := 0; i < len(header); i++ {
		if header[i] == f {
			return true
		}
	}
	return false
}

// checkDataType 判断数据的类型是否合法
func CheckDataType(it *item.ListItems) (result []*item.Error) {
	line := cast.ToInt64(it.Items["index"].Value)
	for field, value := range it.Items {
		if !checkDataValue(value) {
			result = append(result, &item.Error{
				CurrentLine: line,
				FieldId:     field,
				ErrorMsg:    fmt.Sprintf("データタイプが一致しません。フィールドデータのタイプは[%v]である必要があります", value.DataType),
			})
		}
	}
	return result
}

// checkFieldExist 判断字段是否存在
func CheckFieldExist(fieldID string, fields []*field.Field) (f *field.Field) {

	var res *field.Field

Loop:
	for _, field := range fields {
		if field.FieldId == fieldID {
			res = field
			break Loop
		}
	}

	return res
}

// checkDataValue 获取对应的数据类型的数据
func checkDataValue(value *item.Value) (r bool) {
	switch value.DataType {
	case "number":
		if len(value.GetValue()) == 0 {
			return true
		}
		if _, err := cast.ToFloat64E(value.GetValue()); err != nil {
			return false
		}
		return true
	case "date":
		if len(value.GetValue()) == 0 {
			return true
		}
		if _, err := time.Parse("2006-01-02", value.GetValue()); err != nil {
			return false
		}
		return true
	case "time":
		if len(value.GetValue()) == 0 {
			return true
		}
		if _, err := time.Parse("15:04:05", value.GetValue()); err != nil {
			return false
		}
		return true
	case "switch":
		if value.GetValue() == "" {
			return false
		}
		if _, err := cast.ToBoolE(value.GetValue()); err != nil {
			return false
		}
		return true
	case "user":
		if len(value.GetValue()) == 0 {
			return true
		}
		result := strings.Split(value.GetValue(), ",")
		if len(result) == 0 {
			return false
		}
		return true
	case "file":
		if len(value.GetValue()) == 0 {
			return true
		}
		var result []FileValue
		if err := json.Unmarshal([]byte(value.GetValue()), &result); err != nil {
			return false
		}
		return true
	case "lookup":
		return true
	case "autonum":
		return true
	case "function":
		return true
	default:
		return true
	}
}

// checkOptionValid 判断选项是否有效
func CheckOptionValid(group, value string, opList []*option.Option) bool {
	for _, o := range opList {
		if o.OptionId == group && o.OptionValue == value && len(o.DeletedBy) == 0 {
			return true
		}
	}

	return false
}
