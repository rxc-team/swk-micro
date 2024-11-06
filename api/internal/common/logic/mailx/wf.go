package mailx

import (
	"bytes"
	"context"
	"html/template"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/originx"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

type EmailParam struct {
	Database       string
	UserID         string
	AppID          string
	WorkflowID     string
	DatastoreID    string
	Language       string
	CreateUserName string
	Opreate        string
}

func SendEmailToApprover(param EmailParam) error {
	var opreate string
	switch param.Opreate {
	case "insert":
		opreate = "追加"
	case "delete":
		opreate = "削除"
	case "update":
		opreate = "更新する"
	case "debt-change":
		opreate = "債務見積"
	case "info-change":
		opreate = "契約情報の変更"
	case "midway-cancel":
		opreate = "中途解約"
	case "contract-expire":
		opreate = "契約満了"
	}
	// 查询用户信息
	userService := user.NewUserService("manage", client.DefaultClient)
	var uReq user.FindUserRequest
	uReq.Database = param.Database
	uReq.UserId = param.UserID
	uRes, err := userService.FindUser(context.TODO(), &uReq)
	if err != nil {
		return err
	}
	if len(uRes.GetUser().GetNoticeEmail()) > 0 {
		// 发送密码重置邮件
		// 定义收件人
		mailTo := []string{
			uRes.GetUser().GetNoticeEmail(),
		}
		// 定义抄送人
		mailCcTo := []string{}
		// 邮件主题
		subject := "New data pending approval"
		// 邮件正文
		origin := originx.GetOrigin(false)
		linkUrl := origin + "/approve/" + param.WorkflowID + "/list"
		tpl := template.Must(template.ParseFiles("assets/html/workflow.html"))

		// 获取台账对应的多语言
		languageService := language.NewLanguageService("global", client.DefaultClient)

		var lgReq language.FindLanguagesRequest
		lgReq.Domain = uRes.GetUser().GetDomain()
		lgReq.Database = param.Database
		var dsName string

		lgResponse, err := languageService.FindLanguages(context.TODO(), &lgReq)
		if err != nil {
			return err
		}
		for _, lang := range lgResponse.GetLanguageList() {
			if lang.GetLangCd() == param.Language {
				dsName = lang.GetApps()[param.AppID].GetDatastores()[param.DatastoreID]
			}
		}

		params := map[string]string{
			"url":     linkUrl,
			"user":    param.CreateUserName,
			"opreate": opreate,
			"ds":      dsName,
		}

		var out bytes.Buffer
		err = tpl.Execute(&out, params)
		if err != nil {
			return err
		}

		er := SendMail(param.Database, mailTo, mailCcTo, subject, out.String())
		if er != nil {
			return err
		}
	}
	return nil
}
