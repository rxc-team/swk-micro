package dev

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
	HelpProcessName   = "Help"
	ActionFindHelps   = "FindHelps"
	ActionFindTags    = "FindTags"
	ActionFindHelp    = "FindHelp"
	ActionAddHelp     = "AddHelp"
	ActionModifyHelp  = "ModifyHelp"
	ActionDeleteHelp  = "DeleteHelp"
	ActionDeleteHelps = "DeleteHelps"
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

// FindTags 获取所有不重复帮助文档标签
// @Router /tags [get]
func (t *Help) FindTags(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindTags, loggerx.MsgProcessStarted)

	helpService := help.NewHelpService("global", client.DefaultClient)

	var req help.FindTagsRequest
	req.Database = "system"

	response, err := helpService.FindTags(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindTags, err)
		return
	}

	loggerx.InfoLog(c, ActionFindTags, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, HelpProcessName, ActionFindTags)),
		Data:    response.GetTags(),
	})
}

// AddHelp 添加帮助文档
// @Router /helps [post]
func (t *Help) AddHelp(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddHelp, loggerx.MsgProcessStarted)

	helpService := help.NewHelpService("global", client.DefaultClient)

	var req help.AddHelpRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddHelp, err)
		return
	}

	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = "system"

	response, err := helpService.AddHelp(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddHelp, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddHelp, fmt.Sprintf("Help[%s] create Success", response.GetHelpId()))

	loggerx.InfoLog(c, ActionAddHelp, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, HelpProcessName, ActionAddHelp)),
		Data:    response,
	})
}

// ModifyHelp 更新帮助文档
// @Router /helps/{help_id} [put]
func (t *Help) ModifyHelp(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyHelp, loggerx.MsgProcessStarted)

	helpService := help.NewHelpService("global", client.DefaultClient)

	var req help.ModifyHelpRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyHelp, err)
		return
	}

	req.HelpId = c.Param("help_id")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = "system"

	response, err := helpService.ModifyHelp(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyHelp, err)
		return
	}

	loggerx.SuccessLog(c, ActionModifyHelp, fmt.Sprintf("Help[%s] Update Success", req.GetHelpId()))

	loggerx.InfoLog(c, ActionModifyHelp, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, HelpProcessName, ActionModifyHelp)),
		Data:    response,
	})
}

// DeleteHelps 硬删除多个帮助文档
// @Router /helps [delete]
func (t *Help) DeleteHelps(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteHelps, loggerx.MsgProcessStarted)

	helpService := help.NewHelpService("global", client.DefaultClient)

	var req help.DeleteHelpsRequest
	req.HelpIdList = c.QueryArray("help_id_list")
	req.Database = "system"

	response, err := helpService.DeleteHelps(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteHelps, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteHelps, fmt.Sprintf("Helps[%s] HardDelete Success", req.GetHelpIdList()))

	loggerx.InfoLog(c, ActionDeleteHelps, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, HelpProcessName, ActionDeleteHelps)),
		Data:    response,
	})
}
