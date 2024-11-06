package common

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/userx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/task/proto/history"
	"rxcsoft.cn/pit3/srv/task/proto/task"
)

// TaskHistory 任务履历
type TaskHistory struct{}

// log出力使用
const (
	TaskHistoryProcessName  = "TaskHistory"
	ActionFindTaskHistories = "FindTaskHistories"
	ActionDownloadTasks     = "DownloadTaskHistory"
	formatHistory           = "[%s] [%s] [%s] [%s] <%s> %s message content:[%s]" // 类型（时间），任务ID, 任务名称，当前step，用户，任务源
)

// FindTaskHistories 获取当前用户的所有任务履历
// @Router /tasks [get]
func (f *TaskHistory) FindTaskHistories(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindTasks, loggerx.MsgProcessStarted)

	historyService := history.NewHistoryService("task", client.DefaultClient)

	var req history.HistoriesRequest

	isAdmin := c.Query("isAdmin")

	// 从query获取
	index := c.Query("page_index")
	size := c.Query("page_size")
	pageIndex, _ := strconv.ParseInt(index, 10, 64)
	pageSize, _ := strconv.ParseInt(size, 10, 64)
	req.PageIndex = pageIndex
	req.PageSize = pageSize
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	if isAdmin != "true" {
		req.UserId = sessionx.GetAuthUserID(c)
	}

	response, err := historyService.FindHistories(context.TODO(), &req)
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

// DownloadTaskHistory 下载履历
// @Router /apps [get]
func (f *TaskHistory) DownloadTaskHistory(c *gin.Context) {
	// 开始处理log
	loggerx.InfoLog(c, ActionDownloadTasks, loggerx.MsgProcessStarted)

	jobID := c.Query("job_id")
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	appID := sessionx.GetCurrentApp(c)
	langcd := sessionx.GetCurrentLanguage(c)
	db := sessionx.GetUserCustomer(c)
	go func() {

		// 创建任务
		jobx.CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      "task history download",
			Origin:       "TaskHistory Download",
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(langcd, "job.J_014"),
			TaskType:     "th-download",
			Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})

		//开始任务
		historyService := history.NewHistoryService("task", client.DefaultClient)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		var req history.DownloadRequest
		// 获取参数
		req.Database = db
		req.AppId = appID

		response, err := historyService.DownloadHistories(context.TODO(), &req, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(langcd, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		// 添加头部字段
		var taskHistory []string

		allUsers := userx.GetAllUser(db, appID, domain)
		langData := langx.GetLanguageData(db, langcd, domain)
		for _, th := range response.GetHistories() {
			userName := userx.TranUser(th.UserId, allUsers)
			origin := langx.GetLangValue(langData, th.Origin, th.Origin)
			s := fmt.Sprintf(formatHistory, th.StartTime, th.JobId, th.JobName, th.CurrentStep, userName, origin, th.Message)
			taskHistory = append(taskHistory, s)
		}

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(langcd, "job.J_049"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(langcd, "job.J_043"),
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		file := filex.WriteAndSaveFile(domain, appID, taskHistory)
		if file == nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{"save file has error"})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(langcd, "job.J_042"),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

		} else {
			// 发送消息 写入保存文件成功，返回下载路径，任务结束
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(langcd, "job.J_028"),
				CurrentStep: "end",
				File: &task.File{
					Url:  file.MediaLink,
					Name: file.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

		}
	}()

	// 处理结束log
	loggerx.InfoLog(c, ActionDownloadTasks, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, TaskHistoryProcessName, ActionDownloadTasks)),
		Data:    gin.H{},
	})
}
