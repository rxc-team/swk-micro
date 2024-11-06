package admin

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/help"
)

// Help 帮助文档
type Help struct{}

// log出力
const (
	HelpProcessName = "Help"
	ActionFindHelps = "FindHelps"
	ActionFindHelp  = "FindHelp"
)

// FindHelp 获取单个帮助文档
// @Router /helps/{help_id} [get]
func (t *Help) FindHelp(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindHelp, loggerx.MsgProcessStarted)

	helpService := help.NewHelpService("global", client.DefaultClient)

	var req help.FindHelpRequest
	req.HelpId = c.Param("help_id")
	req.Database = "system"

	response, err := helpService.FindHelp(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindHelp, err)
		return
	}

	loggerx.InfoLog(c, ActionFindHelp, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, HelpProcessName, ActionFindHelp)),
		Data:    response.GetHelp(),
	})
}

// FindHelps 获取多个帮助文档
// @Router /helps [get]
func (t *Help) FindHelps(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindHelps, loggerx.MsgProcessStarted)

	helpService := help.NewHelpService("global", client.DefaultClient)

	var req help.FindHelpsRequest
	req.Title = c.Query("title")
	req.Type = c.Query("type")
	req.Tag = c.Query("tag")
	isDev := c.Query("is_dev")
	if isDev == "true" {
		req.LangCd = c.Query("lang_cd")
	} else {
		req.LangCd = sessionx.GetCurrentLanguage(c)
	}
	req.Database = "system"

	response, err := helpService.FindHelps(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindHelps, err)
		return
	}

	loggerx.InfoLog(c, ActionFindHelps, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, HelpProcessName, ActionFindHelps)),
		Data:    response.GetHelps(),
	})
}
