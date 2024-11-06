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
	"rxcsoft.cn/pit3/srv/manage/proto/app"
)

// App App
type App struct{}

// log出力
const (
	AppProcessName = "App"
	// Action
	ActionFindApps            = "FindApps"
	ActionFindApp             = "FindApp"
	ActionFindUserApp         = "FindUserApp"
	ActionAddApp              = "AddApp"
	ActionModifyApp           = "ModifyApp"
	ActionModifyAppSort       = "ModifyAppSort"
	ActionDeleteApp           = "DeleteApp"
	ActionDeleteSelectApps    = "DeleteSelectApps"
	ActionHardDeleteApps      = "HardDeleteApps"
	ActionHardDeleteCopyApps  = "HardDeleteCopyApps"
	ActionRecoverSelectApps   = "RecoverSelectApps"
	ActionaddDefaultGroup     = "addDefaultGroup"
	ActionaddAppLangItem      = "addAppLangItem"
	ActionaddDefaultAdminUser = "addDefaultAdminUser"
	ActionaddDefaultAdminRole = "addDefaultAdminRole"
	ActionAddGroup            = "AddGroup"
	defaultPasswordEnv        = "DEFAULT_PASSWORD"
)

// FindApp 查找单个APP记录
// @Router /apps/{a_id} [get]
func (a *App) FindApp(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindApp, loggerx.MsgProcessStarted)

	appService := app.NewAppService("manage", client.DefaultClient)

	var req app.FindAppRequest
	req.AppId = c.Param("a_id")
	req.Database = c.Query("database")
	response, err := appService.FindApp(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApp, err)
		return
	}
	loggerx.InfoLog(c, ActionFindApp, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AppProcessName, ActionFindApp)),
		Data:    response.GetApp(),
	})
}

// FindUserApp 查找当前用户的所有APP
// @Router /user/apps [get]
func (a *App) FindUserApp(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUserApp, loggerx.MsgProcessStarted)

	apps := sessionx.GetUserApps(c)

	appService := app.NewAppService("manage", client.DefaultClient)

	var req app.FindAppsByIdsRequest
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)
	req.AppIdList = apps

	response, err := appService.FindAppsByIds(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUserApp, err)
		return
	}

	loggerx.InfoLog(c, ActionFindUserApp, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AppProcessName, ActionFindUserApp)),
		Data:    response.GetApps(),
	})
}
