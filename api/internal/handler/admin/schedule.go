package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"

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
func (f *Schedule) FindSchedules(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindSchedules, loggerx.MsgProcessStarted)

	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)

	var req schedule.SchedulesRequest
	// 从query获取
	index := c.Query("page_index")
	size := c.Query("page_size")
	pageIndex, _ := strconv.ParseInt(index, 10, 64)
	pageSize, _ := strconv.ParseInt(size, 10, 64)
	req.PageIndex = pageIndex
	req.PageSize = pageSize
	req.Database = sessionx.GetUserCustomer(c)

	req.UserId = c.Query("user_id")
	req.ScheduleType = c.Query("schedule_type")
	req.RunNow, _ = strconv.ParseBool(c.Query("run_now"))

	response, err := scheduleService.FindSchedules(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSchedules, err)
		return
	}

	loggerx.InfoLog(c, ActionFindSchedules, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ScheduleProcessName, ActionFindSchedules)),
		Data:    response,
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
	loggerx.ProcessLog(c, ActionAddSchedule, msg.L014, params)

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
