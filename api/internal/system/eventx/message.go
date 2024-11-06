package eventx

import (
	"encoding/json"

	"github.com/micro/go-micro/v2/broker"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
)

// SendToUser 发送消息
func SendToUser(event broker.Event) error {
	var param wsx.MessageParam
	err := json.Unmarshal(event.Message().Body, &param)
	if err != nil {
		loggerx.ErrorLog("SendToUser", err.Error())
		return err
	}

	wsx.SendToUser(param)

	event.Ack()
	return nil
}

// SendToUser 发送消息
func SendMsg(event broker.Event) error {
	var param wsx.MessageParam
	err := json.Unmarshal(event.Message().Body, &param)
	if err != nil {
		loggerx.ErrorLog("SendToUser", err.Error())
		return err
	}

	wsx.SendMsg(param)

	event.Ack()
	return nil
}
