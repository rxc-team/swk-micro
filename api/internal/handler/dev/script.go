package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/scriptx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/script"
)

// Script Script
type Script struct{}

// log出力
const (
	scriptProcessName     = "Script"
	ActionFindScriptJobs  = "FindScriptJobs"
	ActionFindScriptJob   = "FindScriptJob"
	ActionAddScriptJob    = "AddScriptJob"
	ActionModifyScriptJob = "ModifyScriptJob"
	ActionExecScriptJob   = "ExecScriptJob"
)

// FindScriptJobs 获取所有Script
// @Router /scripts [get]
func (s *Script) FindScriptJobs(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindScriptJobs, loggerx.MsgProcessStarted)

	scriptService := script.NewScriptService("manage", client.DefaultClient)

	var req script.FindScriptJobsRequest
	// 从query中获取参数
	req.ScriptType = c.Query("script_name")
	req.ScriptVersion = c.Query("email")
	req.RanBy = c.Query("group")
	// 从共通中获取参数
	req.Database = sessionx.GetUserCustomer(c)

	response, err := scriptService.FindScriptJobs(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindScriptJobs, err)
		return
	}

	loggerx.InfoLog(c, ActionFindScriptJobs, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, scriptProcessName, ActionFindScriptJobs)),
		Data:    response.GetScriptJobs(),
	})
}

// FindScriptJob 获取Script
// @Router /scripts/{script_id} [get]
func (s *Script) FindScriptJob(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindScriptJob, loggerx.MsgProcessStarted)

	scriptService := script.NewScriptService("manage", client.DefaultClient)

	var req script.FindScriptJobRequest
	req.ScriptId = c.Param("script_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := scriptService.FindScriptJob(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindScriptJob, err)
		return
	}

	loggerx.InfoLog(c, ActionFindScriptJob, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, scriptProcessName, ActionFindScriptJob)),
		Data:    response.GetScriptJob(),
	})
}

// AddScriptJob 添加Script
// @Router /scripts [post]
func (s *Script) AddScriptJob(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddScriptJob, loggerx.MsgProcessStarted)

	scriptService := script.NewScriptService("manage", client.DefaultClient)

	var req script.AddRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddScriptJob, err)
		return
	}

	req.ScriptType = "javascript"
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := scriptService.AddScriptJob(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddScriptJob, err)
		return
	}

	loggerx.InfoLog(c, ActionAddScriptJob, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, scriptProcessName, ActionAddScriptJob)),
		Data:    response,
	})
}

// ModifyScript 更新Script
// @Router /scripts/{script_id} [put]
func (s *Script) ModifyScript(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyScriptJob, loggerx.MsgProcessStarted)

	scriptService := script.NewScriptService("manage", client.DefaultClient)

	var req script.ModifyRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyScriptJob, err)
		return
	}
	// 从path中获取参数
	req.ScriptId = c.Param("script_id")
	// 当前Script为更新者
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := scriptService.ModifyScriptJob(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyScriptJob, err)
		return
	}

	loggerx.InfoLog(c, ActionModifyScriptJob, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, scriptProcessName, ActionModifyScriptJob)),
		Data:    response,
	})
}

// ExecScriptJob 执行任务
// @Router /scripts/{script_id} [get]
func (s *Script) ExecScriptJob(c *gin.Context) {
	loggerx.InfoLog(c, ActionExecScriptJob, loggerx.MsgProcessStarted)

	sid := c.Param("script_id")

	scriptService := script.NewScriptService("manage", client.DefaultClient)

	var req script.FindScriptJobRequest
	req.ScriptId = sid
	req.Database = sessionx.GetUserCustomer(c)

	response, err := scriptService.FindScriptJob(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindScriptJob, err)
		return
	}

	scp := response.GetScriptJob()

	// 设置版本号
	v := os.Getenv("VERSION")
	if len(v) == 0 {
		v = "1.0.0"
	}

	if scp.GetScriptVersion() != v {
		httpx.GinHTTPError(c, ActionFindScriptJob, fmt.Errorf("バージョンが一致しないため、スクリプトを実行できません。 サーバーバージョン%s、スクリプトバージョン%s", v, scp.GetScriptVersion()))
		return
	}

	var sreq script.StartRequest
	// 从path中获取参数
	sreq.ScriptId = sid
	// 当前Script为更新者
	sreq.Writer = sessionx.GetAuthUserID(c)
	sreq.Database = sessionx.GetUserCustomer(c)

	_, err = scriptService.StartScriptJob(context.TODO(), &sreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionRun, err)
		return
	}

	job := scriptx.NewJob(sid)

	job.ExecJob()

	loggerx.InfoLog(c, ActionExecScriptJob, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, scriptProcessName, ActionExecScriptJob)),
		Data:    nil,
	})
}
