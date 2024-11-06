package wsx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/global/proto/message"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
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
	EndTime   string `json:"end_time"`
}

// SendMsg 发消息，用于task发消息需要JOBID
func SendMsg(param MessageParam) {

	jsonMsg, err := json.Marshal(&Message{
		Sender:    param.Sender,
		Recipient: param.Recipient,
		Domain:    param.Domain,
		Code:      param.Code,
		Content:   param.Content,
		MsgType:   param.MsgType,
		Link:      param.Link,
		Object:    param.Object,
		Status:    "unread",
		SendTime:  time.Now(),
	})
	if err != nil {
		return
	}
	// 发消息
	Manager.SendToUser([]byte(jsonMsg), param.Recipient)
}

// SendToUser 发消息给指定用户，用于更新字段或台账数据后系统发送通知给用户以及用户提问后系统发通知给管理员
func SendToUser(param MessageParam) {
	messageService := message.NewMessageService("global", client.DefaultClient)

	// 保存消息参数编辑
	var req message.AddMessageRequest
	req.Database = "system"
	req.Sender = param.Sender
	req.Recipient = param.Recipient
	req.Domain = param.Domain
	req.MsgType = param.MsgType
	req.Code = param.Code
	req.Link = param.Link
	req.Content = param.Content
	req.Status = param.Status
	req.Object = param.Object
	req.EndTime = param.EndTime

	// 消息类型判断
	if param.MsgType != "lang" {
		// 非刷新语言类型则先保存消息到DB
		response, err := messageService.AddMessage(context.TODO(), &req)
		if err != nil {
			return
		}
		// 发消息
		jsonMsg, err := json.Marshal(&Message{
			MessageID: response.GetMessageId(),
			Sender:    param.Sender,
			Recipient: param.Recipient,
			Domain:    param.Domain,
			Code:      param.Code,
			Content:   param.Content,
			MsgType:   param.MsgType,
			Link:      param.Link,
			Object:    param.Object,
			Status:    "unread",
			SendTime:  time.Now(),
		})
		if err != nil {
			return
		}
		Manager.SendToUser([]byte(jsonMsg), param.Recipient)
	} else {
		// 刷新语言类型则直接发送消息
		jsonMsg, err := json.Marshal(&Message{
			Sender:    param.Sender,
			Recipient: param.Recipient,
			Domain:    param.Domain,
			Code:      param.Code,
			Content:   param.Content,
			MsgType:   param.MsgType,
			Link:      param.Link,
			Object:    param.Object,
			Status:    "unread",
			SendTime:  time.Now(),
		})
		if err != nil {
			return
		}
		Manager.SendToUser([]byte(jsonMsg), param.Recipient)
	}
}

// SendToCompany 发消息给指定客户下所有用户
func SendToCompany(param MessageParam) {
	// 依据客户域名获取客户情报
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	var dreq customer.FindCustomerByDomainRequest
	dreq.Domain = param.Domain
	dresp, err := customerService.FindCustomerByDomain(context.TODO(), &dreq)
	if err != nil {
		loggerx.ErrorLog("SendToCompany", err.Error())
		return
	}

	// 通過domain查找客户下属所有用戶
	userService := user.NewUserService("manage", client.DefaultClient)
	var req user.FindUsersRequest
	req.Database = dresp.GetCustomer().GetCustomerId()
	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("SendToCompany", err.Error())
		return
	}
	// 添加dev端的消息,用于dev端消息添加时显示记录
	SendToUser(param)

	// 循环客户下属所有用戶
	for _, user := range response.GetUsers() {
		param.Recipient = user.GetUserId()
		// 发消息给用户
		SendToUser(param)
	}
}

// SendToDevCompany 发消息开发公司下所有用户
func SendToDevCompany(param MessageParam) {
	// 通過domain查找該公司下的所有用戶
	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest
	req.Database = "system"
	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("SendToCompany", err.Error())
		return
	}

	for _, user := range response.GetUsers() {
		param.Recipient = user.GetUserId()
		SendToUser(param)
	}
}

// SendToEveryone 发消息给所有客户的所有用户
func SendToEveryone(param MessageParam) {
	// 获取所有客户情报
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	var req customer.FindCustomersRequest
	response, err := customerService.FindCustomers(context.TODO(), &req)
	if err != nil {
		return
	}
	// 添加dev端的消息,用于dev端消息添加时显示记录
	SendToUser(param)
	// 循环客户情报
	for _, customer := range response.GetCustomers() {
		// 排除非客户的开发公司
		if customer.GetDomain() != "proship.co.jp" {
			param.Domain = customer.GetDomain()
			// 发消息给客户下属所有用户
			SendToCompany(param)
		}
	}
}

// SendToGroup 发消息给一个组织的所有用户
func SendToGroup(param MessageParam, db, groupID string) {
	// 通過groupid查找组下的所有用戶
	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest

	req.Domain = param.Domain
	req.Database = db
	req.Group = groupID

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		return
	}

	for _, user := range response.GetUsers() {
		param.Recipient = user.GetUserId()
		SendToUser(param)
	}
}

// SendToCurrentAndParentGroup 发消息给当前用户的同组织以及上级组织的用户，可用于用户对数据做了修改后通知上级知晓
func SendToCurrentAndParentGroup(param MessageParam, db, groupID string) {
	groupService := group.NewGroupService("manage", client.DefaultClient)

	var req group.FindGroupsRequest

	// 当前用户的domain下的所有组织
	req.Domain = param.Domain
	req.Database = db
	response, err := groupService.FindGroups(context.TODO(), &req)
	if err != nil {
		return
	}
	// 把list转换成map类型方便后面查找
	groupMap := make(map[string]*group.Group)
	for _, g := range response.GetGroups() {
		groupMap[g.GroupId] = g
	}
	parentGroup := groupID
	// 从当前组开始，循环给每个上级组织发送消息，当上级组织等于root时结束循环
	for {
		// 发送消息给组织下所有用户
		SendToGroup(param, db, parentGroup)
		if value, ok := groupMap[parentGroup]; ok {
			parentGroup = value.GetParentGroupId()
			if parentGroup == "root" {
				return
			}
		} else {
			return
		}

	}
}
