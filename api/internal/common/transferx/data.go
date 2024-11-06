package transferx

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/cast"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/srv/database/proto/item"
)

// TransferData 获取对应的数据类型的数据
func TransferData(value *item.Value) (v interface{}) {
	switch value.DataType {
	case "text", "textarea":
		return value.GetValue()
	case "number":
		return cast.ToFloat32(value.GetValue())
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
		return value.GetValue()
	}

	return nil
}
