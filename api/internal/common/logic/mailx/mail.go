package mailx

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	"github.com/micro/go-micro/v2/client"
	"gopkg.in/gomail.v2"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/global/proto/mail"
	config "rxcsoft.cn/pit3/srv/global/proto/mail-config"
)

// log出力
const (
	MailProcessName = "Mail"
	ActionSendMail  = "SendMail"
)

// SendMail 发邮件
func SendMail(db string, recipients []string, cc []string, subject string, body string) error {
	loggerx.SystemLog(false, false, ActionSendMail, fmt.Sprintf("Process FindConfig:%s", loggerx.MsgProcessStarted))
	// 获取邮箱服务器连接信息
	configService := config.NewConfigService("global", client.DefaultClient)

	var req config.FindConfigRequest
	req.Database = "system"
	res, err := configService.FindConfig(context.TODO(), &req)
	if err != nil {
		loggerx.SystemLog(true, false, ActionSendMail, fmt.Sprintf(loggerx.MsgProcessError, "FindConfig", err))
		return err
	}
	loggerx.SystemLog(false, false, ActionSendMail, fmt.Sprintf("Process FindConfig:%s", loggerx.MsgProcessEnded))

	mailConfig := res.GetConfig()
	// 邮箱配置验证
	if len(mailConfig.Host) == 0 || len(mailConfig.Port) == 0 || len(mailConfig.Mail) == 0 || len(mailConfig.Password) == 0 {
		return fmt.Errorf("email configuration is incorrect")
	}

	// 转换端口类型为int
	port, e := strconv.Atoi(mailConfig.Port)
	if e != nil {
		return e
	}
	// 连接邮箱服务器
	d := gomail.NewDialer(mailConfig.Host, port, mailConfig.Mail, mailConfig.Password)
	if mailConfig.Ssl == "true" {
		d.SSL = true
	}
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// 编辑邮件信息
	m := gomail.NewMessage()
	// 设置带别名送件人
	nickName := "Pro-Ship Incorporated."
	m.SetHeader("From", m.FormatAddress(mailConfig.Mail, nickName))
	// 设置收件人
	if len(recipients) > 0 {
		m.SetHeader("To", recipients...)
	} else {
		return fmt.Errorf("recipient cannot be empty")
	}
	// 设置抄送人
	m.SetHeader("Cc", cc...)
	// 设置邮件主题
	m.SetHeader("Subject", subject)
	// 设置邮件正文
	m.SetBody("text/html", body)
	// // 设置附件名
	// if attachFile != "" {
	// 	pos := strings.LastIndex(attachFile, "/")
	// 	name := attachFile[pos+1:]
	// 	m.Attach(attachFile,
	// 		gomail.Rename(name),
	// 	)
	// }

	// 发送编辑好的邮件
	er := d.DialAndSend(m)
	if er != nil {
		loggerx.SystemLog(true, false, ActionSendMail, fmt.Sprintf(loggerx.MsgProcessError, ActionSendMail, er))
		return er
	}

	// 添加邮件发送记录
	loggerx.SystemLog(false, false, ActionSendMail, fmt.Sprintf("Process AddMail:%s", loggerx.MsgProcessStarted))
	mailService := mail.NewMailService("global", client.DefaultClient)

	var AddReq mail.AddMailRequest

	AddReq.Database = db
	AddReq.Sender = mailConfig.Mail
	AddReq.Recipients = recipients
	AddReq.Ccs = cc
	AddReq.Subject = subject
	AddReq.Content = body
	AddReq.SendTime = time.Now().Format("2006-01-02 15:04:05")

	_, addErr := mailService.AddMail(context.TODO(), &AddReq)
	if addErr != nil {
		loggerx.SystemLog(true, false, ActionSendMail, fmt.Sprintf(loggerx.MsgProcessError, "AddMail", addErr))
		return addErr
	}
	loggerx.SystemLog(false, false, ActionSendMail, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddMail"))

	loggerx.SystemLog(false, false, ActionSendMail, fmt.Sprintf("Process AddMail:%s", loggerx.MsgProcessEnded))
	return nil
}
