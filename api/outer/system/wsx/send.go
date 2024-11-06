package wsx

import (
	"context"
	"encoding/json"

	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/utils/mq"
)

// MessageParam 发送消息时的参数
type MessageParam struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Domain    string `json:"domain"`
	MsgType   string `json:"msg_type"`
	Code      string `json:"code"`
	Link      string `json:"link"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	Object    string `json:"object"`
}

// SendMsg 发消息，用于task发消息需要JOBID
func SendMsg(param MessageParam) {
	br := mq.NewBroker()
	body, err := json.Marshal(param)
	if err != nil {
		br.Publish("message.task", &broker.Message{
			Header: map[string]string{},
			Body:   body,
		})
	}
}

// SendToUser 发消息给指定用户，用于更新字段或台账数据后系统发送通知给用户以及用户提问后系统发通知给管理员
func SendToUser(param MessageParam) {
	br := mq.NewBroker()
	body, err := json.Marshal(param)
	if err != nil {
		br.Publish("message.send", &broker.Message{
			Header: map[string]string{},
			Body:   body,
		})
	}
}

// SendToCompany 发消息给一个公司下所有用户，用于超级管理员按doamin下发通知到公司下所有人员
func SendToCompany(param MessageParam) {

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var dreq customer.FindCustomerByDomainRequest
	dreq.Domain = param.Domain
	dresp, err := customerService.FindCustomerByDomain(context.TODO(), &dreq)
	if err != nil {
		loggerx.ErrorLog("SendToCompany", err.Error())
		return
	}

	// 通過domain查找該公司下的所有用戶
	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest

	req.Database = dresp.GetCustomer().GetCustomerId()

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("SendToCompany", err.Error())
		return
	}
	// 重新清空domain
	param.Domain = ""

	for _, user := range response.GetUsers() {
		param.Recipient = user.GetUserId()
		SendToUser(param)
	}
}
