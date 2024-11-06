package dev

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
)

// Language 语言
type Language struct{}

// log出力
const (
	LanguageProcessName      = "Language"
	ActionFindLanguage       = "FindLanguage"
	ActionDeleteLanguageData = "DeleteLanguageData"
)

// FindLanguage 获取语言数据
// @Router /languages/search [get]
func (l *Language) FindLanguage(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindLanguage, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.LangCd = c.Query("lang_cd")
	domain := c.Query("domain")

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
	}

	if domain != "proship.co.jp" {
		customerService := customer.NewCustomerService("manage", client.DefaultClient)
		var cReq customer.FindCustomerByDomainRequest
		cReq.Domain = domain
		cRes, err := customerService.FindCustomerByDomain(context.TODO(), &cReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindLanguage, err)
			return
		}

		req.Domain = domain
		req.Database = cRes.GetCustomer().CustomerId
		response, err := languageService.FindLanguage(context.TODO(), &req, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindLanguage, err)
			return
		}

		if response.GetApps() == nil {
			response.Apps = make(map[string]*language.App)
		} else {
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
		return
	}

	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionFindLanguage)),
		Data:    gin.H{},
	})
}

// DeleteLanguageData 删除app语言数据
// @Router /apps/{a_id}/languages [delete]
func (l *Language) DeleteLanguageData(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteLanguageData, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.DeleteLanguageDataRequest

	req.Domain = sessionx.GetUserDomain(c)
	req.AppId = c.Param("a_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := languageService.DeleteLanguageData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteLanguageData, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteLanguageData, fmt.Sprintf("Domain[%s] Language delete Success", req.GetDomain()))

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), req.Domain)

	loggerx.InfoLog(c, ActionDeleteLanguageData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionDeleteLanguageData)),
		Data:    response,
	})
}
