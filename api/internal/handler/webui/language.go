package webui

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/language"
)

// Language 语言
type Language struct{}

// log出力
const (
	LanguageProcessName = "Language"
	ActionFindLanguage  = "FindLanguage"
)

// FindLanguage 获取语言数据
// @Router /languages/search [get]
func (l *Language) FindLanguage(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindLanguage, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.LangCd = c.Query("lang_cd")

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
	}

	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := languageService.FindLanguage(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindLanguage, err)
		return
	}
	// // 将空值更改为默认值
	for _, value := range response.GetApps() {
		if value.GetDashboards() == nil {
			value.Dashboards = make(map[string]string)
		}
		if value.GetDatastores() == nil {
			value.Datastores = make(map[string]string)
		}
		if value.GetFields() == nil {
			value.Fields = make(map[string]string)
		}
		if value.GetMappings() == nil {
			value.Mappings = make(map[string]string)
		}
		if value.GetOptions() == nil {
			value.Options = make(map[string]string)
		}
		if value.GetQueries() == nil {
			value.Queries = make(map[string]string)
		}
		if value.GetReports() == nil {
			value.Reports = make(map[string]string)
		}
		if value.GetStatuses() == nil {
			value.Statuses = make(map[string]string)
		}
		if value.GetWorkflows() == nil {
			value.Workflows = make(map[string]string)
		}
	}

	if response.GetCommon().GetGroups() == nil {
		response.GetCommon().Groups = make(map[string]string)
	}
	if response.GetCommon().GetWorkflows() == nil {
		response.GetCommon().Workflows = make(map[string]string)
	}

	loggerx.InfoLog(c, ActionFindLanguage, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionFindLanguage)),
		Data:    response,
	})

}
