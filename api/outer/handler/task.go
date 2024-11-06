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
	"rxcsoft.cn/pit3/srv/task/proto/task"
)

// Task 任务
type Task struct{}

// log出力使用
const (
	TaskProcessName  = "Task"
	ActionFindTasks  = "FindTasks"
	ActionFindTask   = "FindTask"
	ActionAddTask    = "AddTask"
	ActionModifyTask = "ModifyTask"
	ActionDeleteTask = "DeleteTask"
)

//AddTask 添加任务
// @Summary 添加任务
// @description 调用srv中的 task服务，添加任务
// @Tags Task
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "台账ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /tasks [post]
func (f *Task) AddTask(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddTask, loggerx.MsgProcessStarted)
	taskService := task.NewTaskService("task", client.DefaultClient)

	var req task.AddRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddTask, err)
		return
	}
	// 从共通中获取
	req.UserId = sessionx.GetAuthUserID(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := taskService.AddTask(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddTask, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddTask, fmt.Sprintf("Task[%s] create  Success", response.GetJobId()))

	loggerx.InfoLog(c, ActionAddTask, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, TaskProcessName, ActionAddTask)),
		Data:    response,
	})
}
