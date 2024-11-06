package transferx

import (
	"encoding/json"
	"strconv"

	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
)

// TransferApprove 获取对应的数据类型的数据
func TransferApprove(value *approve.Value) (v interface{}) {
	switch value.DataType {
	case "text", "textarea":
		return value.GetValue()
	case "number":
		result, _ := strconv.ParseFloat(value.GetValue(), 64)
		return result
	case "autonum":
		return value.GetValue()
	case "date":
		return value.GetValue()
	case "time":
		return value.GetValue()
	case "switch":
		result, _ := strconv.ParseBool(value.GetValue())
		return result
	case "user":
		var result []string
		json.Unmarshal([]byte(value.GetValue()), &result)
		return result
	case "file":
		var result []typesx.FileValue
		json.Unmarshal([]byte(value.GetValue()), &result)
		return result
	case "options":
		return value.GetValue()
	case "lookup":
		// var result LookUp
		// json.Unmarshal([]byte(value.GetValue()), &result)
		return value.GetValue()
	}

	return nil
}
