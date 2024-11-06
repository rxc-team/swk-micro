package mailx

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	"github.com/micro/go-micro/v2/client"
	"gopkg.in/gomail.v2"
	"rxcsoft.cn/pit3/srv/global/proto/mail"
	config "rxcsoft.cn/pit3/srv/global/proto/mail-config"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
)

// log出力
const (
	MailProcessName = "Mail"
	ActionSendMail  = "SendMail"
)

// SendMail 发邮件
func SendMail(db string, recipients []string, cc []string, subject string, body string) error {
	// 获取邮箱服务器连接信息
	configService := config.NewConfigService("global", client.DefaultClient)

	var req config.FindConfigRequest
	req.Database = "system"
	res, err := configService.FindConfig(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog(ActionSendMail, err.Error())
		return err
	}

	mailConfig := res.GetConfig()
	// 邮箱配置验证
	if len(mailConfig.Host) == 0 || len(mailConfig.Port) == 0 || len(mailConfig.Mail) == 0 || len(mailConfig.Password) == 0 {
		loggerx.ErrorLog(ActionSendMail, "Email configuration is incorrect")
		return fmt.Errorf("Email configuration is incorrect")
	}

	// 转换端口类型为int
	port, err := strconv.Atoi(mailConfig.Port)
	if err != nil {
		loggerx.ErrorLog(ActionSendMail, err.Error())
		return err
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
		loggerx.ErrorLog(ActionSendMail, "Recipient cannot be empty")
		return fmt.Errorf("Recipient cannot be empty")
	}
	// 设置抄送人
	m.SetHeader("Cc", cc...)
	// 设置邮件主题
	m.SetHeader("Subject", subject)
	// 设置邮件正文
	m.SetBody("text/html", body)

	// 发送编辑好的邮件
	err = d.DialAndSend(m)
	if err != nil {
		loggerx.ErrorLog(ActionSendMail, err.Error())
		return err
	}

	// 添加邮件发送记录
	mailService := mail.NewMailService("global", client.DefaultClient)

	var AddReq mail.AddMailRequest

	AddReq.Database = db
	AddReq.Sender = mailConfig.Mail
	AddReq.Recipients = recipients
	AddReq.Ccs = cc
	AddReq.Subject = subject
	AddReq.Content = body
	AddReq.SendTime = time.Now().Format("2006-01-02 15:04:05")

	_, err = mailService.AddMail(context.TODO(), &AddReq)
	if err != nil {
		loggerx.ErrorLog(ActionSendMail, err.Error())
		return err
	}
	return nil
}
