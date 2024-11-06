package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/message"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Message 通知
type Message struct{}

// log出力使用
const (
	ActionFindMessage       = "FindMessage"
	ActionFindUpdateMessage = "FindUpdateMessage"
	ActionFindMessages      = "FindMessages"
	ActionAddMessage        = "AddMessage"
	ActionDeleteMessage     = "DeleteMessage"
	ActionChangeStatus      = "ChangeStatus"
	ActionDeleteMessages    = "DeleteMessages"
)

// FindMessage 获取单个通知
func (t *Message) FindMessage(ctx context.Context, req *message.FindMessageRequest, rsp *message.FindMessageResponse) error {
	utils.InfoLog(ActionFindMessage, utils.MsgProcessStarted)

	res, err := model.FindMessage(req.GetDatabase(), req.GetMessageId())
	if err != nil {
		utils.ErrorLog(ActionFindMessage, err.Error())
		return err
	}

	rsp.Message = res.ToProto()

	utils.InfoLog(ActionFindMessage, utils.MsgProcessEnded)
	return nil
}

// FindMessages 获取多个通知
func (t *Message) FindMessages(ctx context.Context, req *message.FindMessagesRequest, rsp *message.FindMessagesResponse) error {
	utils.InfoLog(ActionFindMessages, utils.MsgProcessStarted)

	messageList, err := model.FindMessages(req.GetDatabase(), req.GetRecipient(), req.GetDomain(), req.GetStatus(), req.GetMsgType(), req.GetLimit(), req.GetSkip())
	if err != nil {
		utils.ErrorLog(ActionFindMessages, err.Error())
		return err
	}

	res := &message.FindMessagesResponse{}

	for _, message := range messageList {
		res.Messages = append(res.Messages, message.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindMessages, utils.MsgProcessEnded)
	return nil
}

// FindUpdateMessage 获取系统更新通知
func (t *Message) FindUpdateMessage(ctx context.Context, req *message.FindUpdateMessageRequest, rsp *message.FindUpdateMessageResponse) error {
	utils.InfoLog(ActionFindUpdateMessage, utils.MsgProcessStarted)

	// 转换时间
	nowtime, err := time.ParseInLocation("2006-01-02 15:04:05", req.GetNowTime(), time.Local)
	if err != nil {
		utils.ErrorLog(ActionFindUpdateMessage, err.Error())
		return err
	}

	messageUpdate, err := model.FindUpdateMessage(req.GetDatabase(), req.GetDomain(), req.GetRecipient(), nowtime)
	if err != nil {
		utils.ErrorLog(ActionFindUpdateMessage, err.Error())
		return err
	}

	res := &message.FindUpdateMessageResponse{}
	res.Message = messageUpdate.ToProto()
	*rsp = *res

	utils.InfoLog(ActionFindUpdateMessage, utils.MsgProcessEnded)
	return nil
}

// AddMessage 添加通知
func (t *Message) AddMessage(ctx context.Context, req *message.AddMessageRequest, rsp *message.AddMessageResponse) error {
	utils.InfoLog(ActionAddMessage, utils.MsgProcessStarted)

	params := model.Message{
		Domain:    req.GetDomain(),
		Sender:    req.GetSender(),
		Content:   req.GetContent(),
		Recipient: req.GetRecipient(),
		Code:      req.GetCode(),
		Link:      req.GetLink(),
		Status:    req.GetStatus(),
		MsgType:   req.GetMsgType(),
		Object:    req.GetObject(),
		SendTime:  time.Now(),
	}
	if req.MsgType == "update" {
		// 转换时间
		endtime, err := time.ParseInLocation("2006-01-02 15:04:05", req.GetEndTime(), time.Local)
		if err != nil {
			utils.ErrorLog(ActionAddMessage, err.Error())
			return err
		}
		params.EndTime = endtime
	}

	id, err := model.AddMessage(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddMessage, err.Error())
		return err
	}

	rsp.MessageId = id

	utils.InfoLog(ActionAddMessage, utils.MsgProcessEnded)
	return nil
}

// ChangeStatus 变更状态
func (t *Message) ChangeStatus(ctx context.Context, req *message.ChangeStatusRequest, rsp *message.ChangeStatusResponse) error {
	utils.InfoLog(ActionChangeStatus, utils.MsgProcessStarted)

	err := model.ChangeStatus(req.GetDatabase(), req.GetMessageId())
	if err != nil {
		utils.ErrorLog(ActionChangeStatus, err.Error())
		return err
	}

	utils.InfoLog(ActionChangeStatus, utils.MsgProcessEnded)
	return nil
}

// DeleteMessage 硬删除通知
func (t *Message) DeleteMessage(ctx context.Context, req *message.DeleteMessageRequest, rsp *message.DeleteMessageResponse) error {
	utils.InfoLog(ActionDeleteMessage, utils.MsgProcessStarted)

	err := model.DeleteMessage(req.GetDatabase(), req.GetMessageId())
	if err != nil {
		utils.ErrorLog(ActionDeleteMessage, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteMessage, utils.MsgProcessEnded)
	return nil
}

// DeleteMessages 硬删除多个通知
func (t *Message) DeleteMessages(ctx context.Context, req *message.DeleteMessagesRequest, rsp *message.DeleteMessagesResponse) error {
	utils.InfoLog(ActionDeleteMessages, utils.MsgProcessStarted)

	err := model.DeleteMessages(req.GetDatabase(), req.GetUserId())
	if err != nil {
		utils.ErrorLog(ActionDeleteMessages, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteMessages, utils.MsgProcessEnded)
	return nil
}
