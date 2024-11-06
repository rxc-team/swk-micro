/*
 * @Description:
 * @Author: RXC 廖云江
 * @Date: 2020-11-03 16:53:45
 * @LastEditors: RXC 廖云江
 * @LastEditTime: 2020-12-07 10:52:54
 */

package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
)

// App App
type App struct{}

// log出力
const (
	AppProcessName    = "App"
	ActionFindUserApp = "FindUserApp"
)

// FindUserApp 查找当前用户的所有APP
// @Summary 查找当前用户的所有APP
// @description 调用srv中的app服务，通过用户ID,查找当前用户的所有APP
// @Tags App
// @Accept json
// @Security JWT
// @Produce  json
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /user/apps [get]
func (a *App) FindUserApp(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUserApp, loggerx.MsgProcessStarted)

	apps := sessionx.GetUserApps(c)
	var req app.FindAppsByIdsRequest
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)
	req.AppIdList = apps

	appService := app.NewAppService("manage", client.DefaultClient)
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
