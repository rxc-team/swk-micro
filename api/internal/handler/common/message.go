package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/message"
)

// Message 通知
type Message struct{}

// log出力
const (
	MessageProcessName      = "Message"
	ActionFindMessage       = "FindMessage"
	ActionFindUpdateMessage = "FindUpdateMessage"
	ActionFindMessages      = "FindMessages"
	ActionAddMessage        = "AddMessage"
	ActionDeleteMessage     = "DeleteMessage"
	ActionChangeStatus      = "ChangeStatus"
	ActionDeleteMessages    = "DeleteMessages"
)

// FindMessages 获取多个通知
// @Router /messages [get]
func (t *Message) FindMessages(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindMessages, loggerx.MsgProcessStarted)

	messageService := message.NewMessageService("global", client.DefaultClient)

	limit := int64(0)
	limitStr := c.Query("limit")
	if len(limitStr) > 0 {
		result1, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindMessages, err)
			return
		}
		limit = result1
	}
	skip := int64(0)
	skipStr := c.Query("skip")
	if len(skipStr) > 0 {
		result2, err := strconv.ParseInt(skipStr, 10, 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindMessages, err)
			return
		}
		skip = result2
	}

	recipient := c.Query("recipient")
	domain := c.Query("domain")

	var req message.FindMessagesRequest
	req.Recipient = recipient
	req.Domain = domain
	if len(recipient) > 0 && len(domain) == 0 {
		req.Domain = sessionx.GetUserDomain(c)
	}

	req.Status = c.Query("status")
	req.MsgType = c.Query("msg_type")
	req.Limit = limit
	req.Skip = skip
	req.Database = "system"

	response, err := messageService.FindMessages(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindMessages, err)
		return
	}

	loggerx.InfoLog(c, ActionFindMessages, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, MessageProcessName, ActionFindMessages)),
		Data:    response.GetMessages(),
	})
}

// FindMessages 获取系统更新通知
// @Router /messages [get]
func (t *Message) FindUpdateMessage(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUpdateMessage, loggerx.MsgProcessStarted)

	messageService := message.NewMessageService("global", client.DefaultClient)

	var req message.FindUpdateMessageRequest
	req.Database = "system"
	req.NowTime = c.Query("now_time")
	req.Domain = sessionx.GetUserDomain(c)
	req.Recipient = sessionx.GetAuthUserID(c)

	response, err := messageService.FindUpdateMessage(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUpdateMessage, err)
		return
	}

	loggerx.InfoLog(c, ActionFindUpdateMessage, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, MessageProcessName, ActionFindUpdateMessage)),
		Data:    response.GetMessage(),
	})
}

// AddMessage 添加通知
// @Router /messages [post]
func (t *Message) AddMessage(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddMessage, loggerx.MsgProcessStarted)

	// messageService := message.NewMessageService("global", client.DefaultClient)

	var req message.AddMessageRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddMessage, err)
		return
	}
	sendType := c.Query("sendType")
	// req.Database = "system"

	// response, err := messageService.AddMessage(context.TODO(), &req)
	// if err != nil {
	// 	httpx.GinHTTPError(c, ActionAddMessage, err)
	// 	return
	// }
	// loggerx.SuccessLog(c, ActionAddMessage, fmt.Sprintf("Message[%s] create Success", response.GetMessageId()))

	// 发送消息到具体的用户
	if sendType == "select" {
		param := wsx.MessageParam{
			Sender:  req.GetSender(),
			Domain:  req.GetDomain(),
			MsgType: req.GetMsgType(),
			Code:    req.GetCode(),
			Link:    req.GetLink(),
			Content: req.GetContent(),
			Status:  "unread",
		}
		wsx.SendToCompany(param)
	} else {
		param := wsx.MessageParam{
			Sender:  req.GetSender(),
			MsgType: req.GetMsgType(),
			Code:    req.GetCode(),
			Link:    req.GetLink(),
			Content: req.GetContent(),
			Status:  "unread",
			EndTime: req.GetEndTime(),
		}
		wsx.SendToEveryone(param)
	}

	loggerx.InfoLog(c, ActionAddMessage, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, MessageProcessName, ActionAddMessage)),
		Data:    nil,
	})
}

// ChangeStatus 变更状态
// @Router /messages/{message_id} [patch]
func (t *Message) ChangeStatus(c *gin.Context) {
	loggerx.InfoLog(c, ActionChangeStatus, loggerx.MsgProcessStarted)

	messageService := message.NewMessageService("global", client.DefaultClient)

	var req message.ChangeStatusRequest
	req.MessageId = c.Param("message_id")
	req.Database = "system"

	var reqConfirm message.FindMessageRequest
	reqConfirm.MessageId = req.MessageId
	reqConfirm.Database = "system"
	res, err := messageService.FindMessage(context.TODO(), &reqConfirm)
	if err != nil {
		loggerx.FailureLog(c, ActionChangeStatus, fmt.Sprintf("Change message status has error: [%v]", err))
		return
	}

	// 只能修改自己消息
	if res.GetMessage().GetRecipient() != sessionx.GetAuthUserID(c) {
		c.JSON(403, gin.H{
			"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
		})
		c.Abort()
		return
	}

	response, err := messageService.ChangeStatus(context.TODO(), &req)
	if err != nil {
		loggerx.FailureLog(c, ActionChangeStatus, fmt.Sprintf("Change message status has error: [%v]", err))
		return
	}

	loggerx.SuccessLog(c, ActionChangeStatus, fmt.Sprintf("Message[%s] changed message status success", req.GetMessageId()))

	loggerx.InfoLog(c, ActionChangeStatus, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, MessageProcessName, ActionDeleteMessage)),
		Data:    response,
	})
}

// DeleteMessage 硬删除通知
// @Router /messages/{message_id} [delete]
func (t *Message) DeleteMessage(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteMessage, loggerx.MsgProcessStarted)

	messageService := message.NewMessageService("global", client.DefaultClient)

	var req message.DeleteMessageRequest
	req.MessageId = c.Param("message_id")
	req.Database = "system"

	var reqConfirm message.FindMessageRequest
	reqConfirm.MessageId = req.MessageId
	reqConfirm.Database = "system"
	res, err := messageService.FindMessage(context.TODO(), &reqConfirm)
	if err != nil {
		loggerx.FailureLog(c, ActionDeleteMessage, fmt.Sprintf("Change message status has error: [%v]", err))
		return
	}

	// 只能删除自己消息
	if res.GetMessage().GetRecipient() != sessionx.GetAuthUserID(c) {
		c.JSON(403, gin.H{
			"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
		})
		c.Abort()
		return
	}

	response, err := messageService.DeleteMessage(context.TODO(), &req)
	if err != nil {
		loggerx.FailureLog(c, ActionDeleteMessage, fmt.Sprintf("Delete message has error: [%v]", err))
		return
	}

	loggerx.SuccessLog(c, ActionDeleteMessage, fmt.Sprintf("Message[%s] Delete Success", req.GetMessageId()))

	loggerx.InfoLog(c, ActionDeleteMessage, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, MessageProcessName, ActionDeleteMessage)),
		Data:    response,
	})
}

// DeleteMessages 硬删除多个通知
// @Router /messages [delete]
func (t *Message) DeleteMessages(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteMessages, loggerx.MsgProcessStarted)

	messageService := message.NewMessageService("global", client.DefaultClient)

	var req message.DeleteMessagesRequest
	req.UserId = sessionx.GetAuthUserID(c)
	req.Database = "system"

	response, err := messageService.DeleteMessages(context.TODO(), &req)
	if err != nil {
		loggerx.FailureLog(c, ActionDeleteMessages, fmt.Sprintf("DeleteMessages has error: [%v]", err))
		return
	}
	loggerx.SuccessLog(c, ActionDeleteMessages, fmt.Sprintf("User[%s] messages has delete success", req.GetUserId()))

	loggerx.InfoLog(c, ActionDeleteMessages, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, MessageProcessName, ActionDeleteMessages)),
		Data:    response,
	})
}
