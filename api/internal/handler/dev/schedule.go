package dev

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	"rxcsoft.cn/utils/mq"
)

// Schedule 任务
type Schedule struct{}

// log出力使用
const (
	ScheduleProcessName      = "Schedule"
	ActionFindSchedules      = "FindSchedules"
	ActionFindSchedule       = "FindSchedule"
	ActionAddSchedule        = "AddSchedule"
	ActionAddRestoreSchedule = "AddRestoreSchedule"
	ActionModifySchedule     = "ModifySchedule"
	ActionDeleteSchedule     = "DeleteSchedule"
)

// FindSchedules 获取当前用户的所有任务列表
// @Router /schedules [get]
func (f *Schedule) FindDefaultSchedules(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindSchedules, loggerx.MsgProcessStarted)

	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)

	var req schedule.SchedulesRequest
	// 从query获取
	req.PageIndex = 1
	req.PageSize = 1
	req.Database = sessionx.GetUserCustomer(c)

	req.UserId = sessionx.GetAuthUserID(c)
	req.ScheduleType = c.Query("schedule_type")
	req.RunNow = false

	response, err := scheduleService.FindSchedules(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSchedules, err)
		return
	}

	if response.Total != 1 {
		loggerx.InfoLog(c, ActionFindSchedules, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionFindSchedules)),
			Data:    nil,
		})
		return
	}

	loggerx.InfoLog(c, ActionFindSchedules, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionFindSchedules)),
		Data:    response.Schedules[0],
	})
}

// AddSchedule 添加任务
// @Router /schedules [post]
func (f *Schedule) AddSchedule(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddSchedule, loggerx.MsgProcessStarted)
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)

	var req schedule.AddRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddSchedule, err)
		return
	}
	// 从共通中获取
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)
	req.Params["db"] = sessionx.GetUserCustomer(c)
	req.Params["domain"] = sessionx.GetUserDomain(c)
	req.Params["app_id"] = sessionx.GetCurrentApp(c)
	req.Params["client_ip"] = c.ClientIP()

	response, err := scheduleService.AddSchedule(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddSchedule, err)
		return
	}

	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["backup_name"] = req.GetScheduleName()
	loggerx.ProcessLog(c, ActionAddBackup, msg.L014, params)

	loggerx.SuccessLog(c, ActionAddSchedule, fmt.Sprintf("Schedule[%s] create  Success", response.GetScheduleId()))

	// 获取任务详情
	sc, err := findSchedule(response.GetScheduleId(), req.GetDatabase())
	if err != nil {
		httpx.GinHTTPError(c, ActionAddSchedule, err)
		return
	}
	job := new(jobx.Job)

	if req.GetRunNow() {
		job.Run(sc)
	} else {
		if sc.Status == "0" {
			loggerx.InfoLog(c, ActionAddSchedule, loggerx.MsgProcessEnded)
			c.JSON(200, httpx.Response{
				Status:  0,
				Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionAddSchedule)),
				Data:    response,
			})
			return
		}

		// 创建任务
		body, err := json.Marshal(sc)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddSchedule, err)
			return
		}

		bk := mq.NewBroker()

		err = bk.Publish("job.add", &broker.Message{
			Body: body,
		})

		if err != nil {
			loggerx.ErrorLog(ActionAddSchedule, err.Error())
			return
		}
	}

	loggerx.InfoLog(c, ActionAddSchedule, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionAddSchedule)),
		Data:    response,
	})
}

// ModifySchedule 更新任务状态
// @Router /schedules [post]
func (f *Schedule) ModifySchedule(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifySchedule, loggerx.MsgProcessStarted)
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)

	var req schedule.ModifyRequest
	req.ScheduleId = c.Param("s_id")
	req.Status = c.Query("status")
	// 从共通中获取
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	// 获取任务详情
	sc, err := findSchedule(req.GetScheduleId(), req.GetDatabase())
	if err != nil {
		httpx.GinHTTPError(c, ActionModifySchedule, err)
		return
	}

	response, err := scheduleService.ModifySchedule(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifySchedule, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifySchedule, fmt.Sprintf("Schedule[%s] update  Success", req.GetScheduleId()))

	if sc.Status != req.Status {
		sc.Status = req.Status

		// 任务
		body, err := json.Marshal(sc)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifySchedule, err)
			return
		}
		bk := mq.NewBroker()
		if sc.Status == "1" {
			err := bk.Publish("job.add", &broker.Message{
				Body: body,
			})

			if err != nil {
				loggerx.ErrorLog(ActionModifySchedule, err.Error())
				return
			}
		} else {
			err := bk.Publish("job.stop", &broker.Message{
				Body: body,
			})

			if err != nil {
				loggerx.ErrorLog(ActionModifySchedule, err.Error())
				return
			}
		}
	}

	loggerx.InfoLog(c, ActionModifySchedule, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionModifySchedule)),
		Data:    response,
	})
}

// AddRestoreSchedule 添加恢复任务
// @Router /schedules/local [post]
func (f *Schedule) AddRestoreSchedule(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddRestoreSchedule, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)

	// 需要恢复的DB（即顾客ID）
	backupDb := c.PostForm("db")
	// 需要恢复的domain（即顾客domain）
	backupDomain := c.PostForm("domain")
	// zip备份文件
	zipFile, err := c.FormFile("zipFile")
	if err != nil {
		httpx.GinHTTPError(c, ActionAddRestoreSchedule, err)
		return
	}
	// 文件类型检查
	if !filex.CheckSupport("zip", zipFile.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionAddRestoreSchedule, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "zip", zipFile.Size) {
		httpx.GinHTTPError(c, ActionAddRestoreSchedule, errors.New("maximum size of uploaded file is 1G"))
		return
	}
	// 创建备份文件文件夹
	dir := "backups/local/" + time.Now().Format("20060102150405") + "/"
	dirErr := filex.Mkdir(dir)
	if dirErr != nil {
		httpx.GinHTTPError(c, ActionAddRestoreSchedule, err)
		return
	}
	// 拷贝备份文件到备份文件文件夹
	zipFilePath := dir + zipFile.Filename
	zipFile.Filename = zipFilePath
	if err := c.SaveUploadedFile(zipFile, zipFile.Filename); err != nil {
		httpx.GinHTTPError(c, ActionAddRestoreSchedule, err)
		return
	}

	// 加入立马执行计划
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)
	var req schedule.AddRequest
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = db
	req.Params = make(map[string]string)
	req.Params["db"] = db
	req.Params["domain"] = domain
	req.Params["backup_db"] = backupDb
	req.Params["backup_domain"] = backupDomain
	req.Params["app_id"] = "system"
	req.Params["client_ip"] = c.ClientIP()
	req.Params["local_path"] = zipFilePath
	req.ScheduleName = "db restore" + strconv.FormatInt(time.Now().Unix(), 10)
	req.Spec = ""
	req.Multi = 0
	req.RetryTimes = 1
	req.RetryInterval = 1000
	req.StartTime = time.Now().Format("2006-01-02")
	req.EndTime = time.Now().Format("2006-01-02")
	req.ScheduleType = "db-restore"
	req.Status = "1"
	req.RunNow = true

	response, err := scheduleService.AddSchedule(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddRestoreSchedule, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddRestoreSchedule, fmt.Sprintf("Schedule[%s] create  Success", response.GetScheduleId()))

	// 获取任务计划详情
	sc, err := findSchedule(response.GetScheduleId(), req.GetDatabase())
	if err != nil {
		httpx.GinHTTPError(c, ActionAddRestoreSchedule, err)
		return
	}
	// 立马执行任务计划
	job := new(jobx.Job)
	job.Run(sc)

	loggerx.InfoLog(c, ActionAddRestoreSchedule, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionAddRestoreSchedule)),
		Data:    response,
	})
}

func findSchedule(id, db string) (s *schedule.Schedule, err error) {

	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)

	var req schedule.ScheduleRequest
	req.ScheduleId = id
	req.Database = db
	response, err := scheduleService.FindSchedule(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog(ActionFindSchedule, err.Error())
		return nil, err
	}

	return response.GetSchedule(), nil
}

// DeleteSchedule 删除任务
// @Router /schedules/{j_id} [delete]
func (f *Schedule) DeleteSchedule(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSchedule, loggerx.MsgProcessStarted)
	db := sessionx.GetUserCustomer(c)
	scheduleIds := c.QueryArray("schedule_ids")

	for _, id := range scheduleIds {
		// 获取任务详情
		sc, err := findSchedule(id, db)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddSchedule, err)
			return
		}

		// 删除任务
		body, err := json.Marshal(sc)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddSchedule, err)
			return
		}

		bk := mq.NewBroker()

		bk.Publish("job.delete", &broker.Message{
			Body: body,
		})
	}

	loggerx.InfoLog(c, ActionDeleteSchedule, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionDeleteSchedule)),
		Data:    nil,
	})
}
