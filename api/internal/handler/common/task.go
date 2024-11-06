package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
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

// FindTasks 获取当前用户的所有任务
// @Router /tasks [get]
func (f *Task) FindTasks(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindTasks, loggerx.MsgProcessStarted)

	taskService := task.NewTaskService("task", client.DefaultClient)

	var req task.TasksRequest
	// 从query获取
	index := c.Query("page_index")
	size := c.Query("page_size")
	pageIndex, _ := strconv.ParseInt(index, 10, 64)
	pageSize, _ := strconv.ParseInt(size, 10, 64)
	req.PageIndex = pageIndex
	req.PageSize = pageSize
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)

	req.UserId = sessionx.GetAuthUserID(c)

	response, err := taskService.FindTasks(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindTasks, err)
		return
	}

	loggerx.InfoLog(c, ActionFindTasks, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, TaskProcessName, ActionFindTasks)),
		Data:    response,
	})
}

//FindTask 获取任务
// @Router /tasks/{j_id} [get]
func (f *Task) FindTask(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindTask, loggerx.MsgProcessStarted)

	taskService := task.NewTaskService("task", client.DefaultClient)

	var req task.TaskRequest
	req.JobId = c.Param("j_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := taskService.FindTask(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindTask, err)
		return
	}

	loggerx.InfoLog(c, ActionFindTask, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, TaskProcessName, ActionFindTask)),
		Data:    response.GetTask(),
	})
}

//AddTask 添加任务
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

// DeleteTask 删除任务
// @Router /tasks/{j_id} [delete]
func (f *Task) DeleteTask(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteTask, loggerx.MsgProcessStarted)
	taskService := task.NewTaskService("task", client.DefaultClient)

	var req task.DeleteRequest
	// 从path中获取参数
	req.JobId = c.Param("j_id")
	req.AppId = sessionx.GetCurrentApp(c)
	req.UserId = sessionx.GetAuthUserID(c)

	// domain := sessionx.GetUserDomain(c)

	response, err := taskService.DeleteTask(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteTask, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteTask, fmt.Sprintf("Task[%s] delete  Success", req.GetJobId()))

	// // 删除对应的错误文件
	// errFile := c.Query("error_file")
	// if errFile != "" && errFile != "undefined" {
	// 	filex.DeleteFile(domain, errFile)
	// }
	// // 删除对应的下载文件
	// file := c.Query("file")
	// if file != "" && file != "undefined" {
	// 	filex.DeleteFile(domain, file)
	// }

	loggerx.InfoLog(c, ActionDeleteTask, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, TaskProcessName, ActionDeleteTask)),
		Data:    response,
	})
}
