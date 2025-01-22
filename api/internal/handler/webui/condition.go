package webui

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/journal/proto/condition"
)

// Journal 分录
type Condition struct{}

// log出力使用
const (
	// JournalProcessName        = "Journal"
	ActionAddCondition   = "AddCondition"
	ActionFindConditions = "FindConditions"
)

// AddDatastoreMapping 添加分录下载设置
// @Router /download/setting[post]
func (f *Condition) AddCondition(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddCondition, loggerx.MsgProcessStarted)

	conditionService := condition.NewConditionService("journal", client.DefaultClient)

	var req condition.AddConditionRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddCondition, err)
		return
	}
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := conditionService.AddCondition(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddCondition, err)

		return
	}
	loggerx.SuccessLog(c, ActionAddCondition, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddCondition))

	loggerx.InfoLog(c, ActionAddCondition, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionAddCondition)),
		Data:    response,
	})
}

// FindDownloadSetting 查询分录下载设置
// @Router download/find[GET]
func (f *Condition) FindConditions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDownloadSetting, loggerx.MsgProcessStarted)

	conditionService := condition.NewConditionService("journal", client.DefaultClient)

	var req condition.FindConditionsRequest

	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := conditionService.FindConditions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDownloadSetting, err)

		return
	}
	loggerx.SuccessLog(c, ActionFindDownloadSetting, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionFindDownloadSetting))

	loggerx.InfoLog(c, ActionFindDownloadSetting, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDownloadSetting)),
		Data:    response,
	})
}
