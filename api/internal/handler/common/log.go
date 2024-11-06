package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/logger"
	"rxcsoft.cn/pit3/srv/task/proto/task"
)

// Log 日志
type Log struct{}

// log出力
const (
	LogProcessName    = "Log"
	ActionDownloadLog = "DownloadLog"
	ActionFindLogs    = "FindLogs"
	format            = "[%s-%s-%s] %s <%s> [%s] [%s]: %s" // 类型（大类-APP-等级），时间（2006-01-02 15：04：05.000000），Ip，处理ID，消息
)

// DownloadLog 下载日志
// @Router /apps [get]
func (l *Log) DownloadLog(c *gin.Context) {
	// 开始处理log
	loggerx.InfoLog(c, ActionDownloadLog, loggerx.MsgProcessStarted)

	jobID := c.Query("job_id")
	userID := c.Query("user_id")
	domain := sessionx.GetUserDomain(c)
	appID := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)
	lang := sessionx.GetCurrentLanguage(c)
	timezone := sessionx.GetCurrentTimezone(c)

	clientIp := c.Query("client_ip")
	logType := c.Query("log_type")
	level := c.Query("level")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	isDev := c.Query("is_dev")
	appName := c.Query("app_name")

	go func() {

		// 创建任务
		jobx.CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      "log download",
			Origin:       "Log Download",
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(lang, "job.J_014"),
			TaskType:     "log-download",
			Steps:        []string{"start", "build-data", "save-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})

		// 任务开始
		loggerService := logger.NewLoggerService("global", client.DefaultClient)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		var req logger.LoggersRequest
		// 从path中获取参数
		req.UserId = userID
		req.ClientIp = clientIp
		req.LogType = logType
		req.Level = level
		req.StartTime = startTime
		req.EndTime = endTime
		if isDev == "true" {
			req.AppName = appName
		} else {
			req.Domain = domain
			req.AppName = "internal"
		}
		response, err := loggerService.FindLoggers(context.TODO(), &req, opss)
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
			Message:     i18n.Tr(lang, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		var logs []string

		langData := langx.GetLanguageData(db, lang, domain)

		for _, lg := range response.GetLoggers() {
			params := make(map[string]string)
			for key, value := range lg.GetParams() {
				if strings.HasPrefix(value, "{{") {
					ky := strings.Replace(strings.Replace(value, "{{", "", 1), "}}", "", 1)

					value = langx.GetLangValue(langData, ky, langx.DefaultResult)
					params[key] = value
				} else {
					params[key] = value
				}
			}

			local, err := time.LoadLocation(timezone)
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
			t, err := time.Parse("2006-01-02 15:04:05.000000", lg.Time)
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

			lcTime := t.In(local).Format("2006-01-02 15:04:05.000000")
			zone, _ := t.In(local).Zone()

			out := msg.GetObjMsg(lang, msg.Logger, lg.GetMsg(), params)
			if len(out) == 0 {
				out = lg.GetMsg()
			}
			s := fmt.Sprintf(format, lg.LogType, lg.AppName, lg.Level, fmt.Sprintf("%s %s", lcTime, zone), lg.ClientIp, lg.UserId, lg.ProcessId, out)
			logs = append(logs, s)
		}

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_043"),
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		file := filex.WriteAndSaveFile(domain, appID, logs)
		if file == nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{"save file has error"})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_042"),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			File: &task.File{
				Url:  file.MediaLink,
				Name: file.Name,
			},
			EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database: db,
		}, userID)

	}()

	// 处理结束log
	loggerx.InfoLog(c, ActionDownloadLog, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LogProcessName, ActionDownloadLog)),
		Data:    gin.H{},
	})
}

// FindLogs 获取日志记录
// @Router /logs [get]
func (l *Log) FindLogs(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindLogs, loggerx.MsgProcessStarted)

	loggerService := logger.NewLoggerService("global", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	var req logger.LoggersRequest
	req.UserId = c.Query("user_id")
	req.ClientIp = c.Query("client_ip")
	req.LogType = c.Query("log_type")
	req.Level = c.Query("level")
	req.StartTime = c.Query("start_time")
	req.EndTime = c.Query("end_time")
	req.PageIndex, _ = strconv.ParseInt(c.Query("page_index"), 0, 64)
	req.PageSize, _ = strconv.ParseInt(c.Query("page_size"), 0, 64)
	isDev := c.Query("is_dev")
	if isDev == "true" {
		req.AppName = c.Query("app_name")
	} else {
		req.Domain = sessionx.GetUserDomain(c)
		req.AppName = "internal"
	}

	response, err := loggerService.FindLoggers(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindLogs, err)
		return
	}

	loggerx.InfoLog(c, ActionFindLogs, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LogProcessName, ActionFindLogs)),
		Data: gin.H{
			"total":   response.GetTotal(),
			"loggers": response.GetLoggers(),
		},
	})
}
