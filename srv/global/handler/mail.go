package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/mail"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Mail 邮件
type Mail struct{}

// log出力使用
const (
	ActionFindMails = "FindMails"
	ActionAddMail   = "AddMail"
)

// FindMails 获取邮件
func (m *Mail) FindMails(ctx context.Context, req *mail.FindMailsRequest, rsp *mail.MailsResponse) error {
	utils.InfoLog(ActionFindMails, utils.MsgProcessStarted)

	mails, err := model.FindMails(req.GetDatabase(), req.GetRecipient(), req.GetCc(), req.GetSubject(), req.GetContent(), req.GetAnnex(), req.GetSendTime())
	if err != nil {
		utils.ErrorLog(ActionFindMails, err.Error())
		return err
	}

	res := &mail.MailsResponse{}
	for _, mail := range mails {
		res.Mails = append(res.Mails, mail.ToProto())
	}

	*rsp = *res
	utils.InfoLog(ActionFindMails, utils.MsgProcessEnded)

	return nil
}

// AddMail 添加邮件记录
func (m *Mail) AddMail(ctx context.Context, req *mail.AddMailRequest, rsp *mail.AddMailResponse) error {
	utils.InfoLog(ActionAddMail, utils.MsgProcessStarted)

	params := model.Mail{
		Sender:     req.Sender,
		Recipients: req.Recipients,
		Ccs:        req.Ccs,
		Subject:    req.Subject,
		Content:    req.Content,
		Annex:      req.Annex,
		SendTime:   req.SendTime,
	}

	err := model.AddMail(req.Database, &params)
	if err != nil {
		utils.ErrorLog(ActionAddMail, err.Error())
		return err
	}

	utils.InfoLog(ActionAddMail, utils.MsgProcessEnded)
	return nil
}
