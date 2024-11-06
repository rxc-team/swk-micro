package model

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"rxcsoft.cn/pit3/srv/database/proto/copy"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/template"
)

// GetLinkedParam 获取值
func GetLinkedParam(it interface{}) interface{} {
	itemMap := it.(map[string]interface{})
	dataType := itemMap["data_type"].(string)
	switch dataType {
	case "text", "textarea", "options", "number", "autonum", "time", "switch":
		return itemMap["value"]
	case "date":
		switch itemMap["value"].(type) {
		case primitive.DateTime:
			return itemMap["value"].(primitive.DateTime).Time().Format("2006-01-02")
		case time.Time:
			return itemMap["value"].(time.Time).Format("2006-01-02")
		default:
			return ""
		}
	case "user":
		jsonBytes, _ := json.Marshal(itemMap["value"])
		return string(jsonBytes)
	case "file":
		jsonBytes, _ := json.Marshal(itemMap["value"])
		return string(jsonBytes)
	case "lookup":
		return itemMap["value"]
	default:
		jsonBytes, _ := json.Marshal(itemMap["value"])
		return string(jsonBytes)
	}
}

// GetFuncParam 获取值
func GetFuncParam(v *Value) interface{} {
	switch v.DataType {
	case "text", "textarea", "options", "number", "autonum", "time", "switch":
		return v.Value
	case "date":
		switch v.Value.(type) {
		case primitive.DateTime:
			return v.Value.(primitive.DateTime).Time().Format("2006-01-02")
		case time.Time:
			return v.Value.(time.Time).Format("2006-01-02")
		default:
			return ""
		}
	case "user":
		jsonBytes, _ := json.Marshal(v.Value)
		return string(jsonBytes)
	case "file":
		jsonBytes, _ := json.Marshal(v.Value)
		return string(jsonBytes)
	case "lookup":
		return v.Value
	default:
		jsonBytes, _ := json.Marshal(v.Value)
		return string(jsonBytes)
	}
}

// ConvertItemValue 将mapvalue转换为item.Value类型
func ConvertItemValue(it interface{}) *item.Value {
	switch it.(type) {
	case primitive.M:
		itemMap := it.(primitive.M)
		dataType := itemMap["data_type"].(string)
		value := itemMap["value"].(string)
		return &item.Value{
			DataType: dataType,
			Value:    value,
		}
	default:
		itemMap := it.(*Value)
		return &item.Value{
			DataType: itemMap.DataType,
			Value:    GetValueFromModel(itemMap),
		}
	}
}

func GetValueFromProto(value *item.Value) (v interface{}) {

	switch value.DataType {
	case "text", "textarea":
		return value.GetValue()
	case "number":
		result, err := strconv.ParseFloat(value.GetValue(), 64)
		if err != nil {
			return 0
		}
		return result
	case "date":
		zone := time.Time{}
		if len(value.GetValue()) == 0 {
			return zone
		}
		date, err := time.Parse("2006-01-02", value.GetValue())
		if err != nil {
			return zone
		}
		return date
	case "time":
		return value.GetValue()
	case "switch":
		result, err := strconv.ParseBool(value.GetValue())
		if err != nil {
			return false
		}
		return result
	case "user":
		if len(value.GetValue()) == 0 {
			return []string{}
		}
		result := strings.Split(value.GetValue(), ",")
		return result
	case "file":
		return value.GetValue()
	case "options":
		return value.GetValue()
	case "lookup":
		return value.GetValue()
	}

	return ""
}

func GetCopyeValueFromProto(value *copy.Value) (v interface{}) {

	switch value.DataType {
	case "text", "textarea":
		return value.GetValue()
	case "number":
		result, err := strconv.ParseFloat(value.GetValue(), 64)
		if err != nil {
			return 0
		}
		return result
	case "date":
		zone := time.Time{}
		if len(value.GetValue()) == 0 {
			return zone
		}
		date, err := time.Parse("2006-01-02", value.GetValue())
		if err != nil {
			return zone
		}
		return date
	case "time":
		return value.GetValue()
	case "switch":
		result, err := strconv.ParseBool(value.GetValue())
		if err != nil {
			return false
		}
		return result
	case "user":
		if len(value.GetValue()) == 0 {
			return []string{}
		}
		result := strings.Split(value.GetValue(), ",")
		return result
	case "file":
		return value.GetValue()
	case "options":
		return value.GetValue()
	case "lookup":
		return value.GetValue()
	}

	return ""
}

func GetTemplateValueFromProto(value *template.Value) (v interface{}) {

	switch value.DataType {
	case "text", "textarea":
		return value.GetValue()
	case "number":
		return cast.ToFloat32(value.Value)
	case "date":
		if len(value.GetValue()) == 0 {
			return nil
		}
		date, err := time.Parse("2006-01-02", value.GetValue())
		if err != nil {
			return nil
		}
		return date
	case "time":
		return value.GetValue()
	case "switch":
		return cast.ToBool(value.Value)
	case "user":
		if len(value.GetValue()) == 0 {
			return []string{}
		}
		result := strings.Split(value.GetValue(), ",")
		return result
	case "file":
		return value.GetValue()
	case "options":
		return value.GetValue()
	case "lookup":
		return value.GetValue()
	}

	return ""
}

func GetNumberValue(value *Value) float32 {

	if value.DataType != "number" {
		return 0
	}

	if value.Value == nil {
		return 0
	}

	return cast.ToFloat32(value.Value)
}

func GetValueFromModel(value *Value) (v string) {
	switch value.DataType {
	case "text", "textarea":
		return cast.ToString(value.Value)
	case "options":
		return cast.ToString(value.Value)
	case "number":
		if value.Value == nil {
			return "0"
		}

		return cast.ToString(value.Value)
	case "autonum":
		if value.Value == nil {
			return ""
		}
		return cast.ToString(value.Value)
	case "date":
		if value.Value == nil {
			return ""
		}
		switch value.Value.(type) {
		case primitive.DateTime:
			return value.Value.(primitive.DateTime).Time().Format("2006-01-02")
		case time.Time:
			return value.Value.(time.Time).Format("2006-01-02")
		default:
			return ""
		}
	case "time":
		return cast.ToString(value.Value)
	case "switch":
		return cast.ToString(value.Value)
	case "user":
		if value.Value == nil {
			return ""
		}
		jsonBytes, err := json.Marshal(value.Value)
		if err != nil {
			return ""
		}
		return string(jsonBytes)
	case "file":
		return cast.ToString(value.Value)
	case "function":
		return cast.ToString(value.Value)
	case "lookup":
		return cast.ToString(value.Value)
	default:
		jsonBytes, _ := json.Marshal(value.Value)
		return string(jsonBytes)
	}
}

func CheckValueFromModel(value interface{}, dataType string) (err error) {
	if value == nil {
		return nil
	}

	switch dataType {
	case "text", "textarea":
		switch value.(type) {
		case string:
			return nil
		default:
			return errors.New("the result of the calculation is not a text type")
		}
	case "number":
		switch value.(type) {
		case int64:
			return nil
		case float64:
			return nil
		default:
			return errors.New("the result of the calculation is not a numeric type")
		}
	default:
		return nil
	}
}
