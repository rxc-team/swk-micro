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
	"rxcsoft.cn/pit3/srv/manage/proto/action"
)

// Action 许可操作
type Action struct{}

// log出力使用
const (
	ActionProcessName   = "Action"
	ActionFindActions   = "FindActions"
	ActionFindAction    = "FindAction"
	ActionAddAction     = "AddAction"
	ActionModifyAction  = "ModifyAction"
	ActionDeleteAction  = "DeleteAction"
	ActionDeleteActions = "DeleteActions"
)

// FindActions 获取所有许可操作
// @Router /actions [get]
func (f *Action) FindActions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindActions, loggerx.MsgProcessStarted)

	actionService := action.NewActionService("manage", client.DefaultClient)

	var req action.FindActionsRequest
	// 从query获取
	req.ActionGroup = c.Query("action_group")

	response, err := actionService.FindActions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindActions, err)
		return
	}

	loggerx.InfoLog(c, ActionFindActions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ActionProcessName, ActionFindActions)),
		Data:    response.GetActions(),
	})
}

// FindAction 获取单个许可操作
// @Router /action/objs/{action_object}/actions/{action_key} [get]
func (f *Action) FindAction(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindAction, loggerx.MsgProcessStarted)

	actionService := action.NewActionService("manage", client.DefaultClient)

	var req action.FindActionRequest
	req.ActionObject = c.Param("action_object")
	req.ActionKey = c.Param("action_key")
	response, err := actionService.FindAction(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindAction, err)
		return
	}

	loggerx.InfoLog(c, ActionFindAction, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ActionProcessName, ActionFindAction)),
		Data:    response.GetAction(),
	})
}

// AddAction 添加许可操作
// @Router /actions [post]
func (f *Action) AddAction(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddAction, loggerx.MsgProcessStarted)

	actionService := action.NewActionService("manage", client.DefaultClient)

	var req action.AddActionRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddAction, err)
		return
	}
	// 从共通中获取
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := actionService.AddAction(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddAction, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddAction, fmt.Sprintf("Action[%s] Create Success", req.GetActionKey()))

	loggerx.InfoLog(c, ActionAddAction, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ActionProcessName, ActionAddAction)),
		Data:    response,
	})
}

// ModifyAction 更新许可操作
// @Router /action/objs/{action_object}/actions/{action_key} [put]
func (f *Action) ModifyAction(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyAction, loggerx.MsgProcessStarted)

	actionService := action.NewActionService("manage", client.DefaultClient)

	var req action.ModifyActionRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyAction, err)
		return
	}

	// 从path中获取参数
	req.ActionKey = c.Param("action_key")
	req.ActionObject = c.Param("action_object")

	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := actionService.ModifyAction(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyAction, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyAction, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyAction))

	loggerx.InfoLog(c, ActionModifyAction, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ActionProcessName, ActionModifyAction)),
		Data:    response,
	})
}

// DeleteActions 硬删除多个许可操作
// @Router /actions [PUT]
func (f *Action) DeleteActions(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteActions, loggerx.MsgProcessStarted)

	actionService := action.NewActionService("manage", client.DefaultClient)

	var req action.DeleteActionsRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionDeleteActions, err)
		return
	}

	response, err := actionService.DeleteActions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteActions, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteActions, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeleteActions))

	loggerx.InfoLog(c, ActionDeleteActions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ActionProcessName, ActionDeleteActions)),
		Data:    response,
	})
}
