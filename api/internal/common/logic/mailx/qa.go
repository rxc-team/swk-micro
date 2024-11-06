package mailx

import (
	"bytes"
	"errors"
	"html/template"
	"strings"
)

// SendQuestionEmail 顾客更新或添加问题后系统发送邮件给管理员
func SendQuestionEmail(db, subject, message, noticeMail string) error {
	// 发送邮件
	// 定义收件人
	if len(noticeMail) == 0 {
		return errors.New("notification email does not exist")
	}
	mailTo := []string{
		noticeMail,
	}
	// 定义抄送人
	mailCcTo := []string{}
	// 邮件主题
	sub := subject
	// 邮件正文
	tpl := template.Must(template.ParseFiles("assets/html/message.html"))
	params := map[string]string{
		"msg": message,
	}

	var out bytes.Buffer
	err := tpl.Execute(&out, params)
	if err != nil {
		return err
	}
	c1 := strings.Replace(out.String(), "&gt;", ">", -1)
	c2 := strings.Replace(c1, "&lt;", "<", -1)
	// 发送邮件
	er := SendMail(db, mailTo, mailCcTo, sub, c2)
	if er != nil {
		return er
	}
	return nil
}
