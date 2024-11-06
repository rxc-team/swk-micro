package httpx

// log出力
const (
	Temp     = "API.%v.%v"
	TimeTemp = "%v"
)

// Response 正常返回结果
type Response struct {
	Status  int32       `json:"status" example:"0"`
	Message string      `json:"message" example:"更新成功"`
	Data    interface{} `json:"data"`
}

// ErrorResponse 错误返回结果
type ErrorResponse struct {
	Message string `json:"message" example:"更新失败"`
}
