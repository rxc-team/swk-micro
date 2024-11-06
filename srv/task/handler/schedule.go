package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/task/model"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	"rxcsoft.cn/pit3/srv/task/utils"
)

// Schedule 任务
type Schedule struct{}

// log出力使用
const (
	ScheduleProcessName              = "Schedule"
	ActionFindSchedules              = "FindSchedules"
	ActionFindSchedule               = "FindSchedule"
	ActionAddSchedule                = "AddSchedule"
	ActionModifySchedule             = "ModifySchedule"
	ActionDeleteSchedule             = "DeleteSchedule"
	ActionAddScheduleNameUniqueIndex = "AddScheduleNameUniqueIndex"
)

// FindSchedules 获取多个任务
func (f *Schedule) FindSchedules(ctx context.Context, req *schedule.SchedulesRequest, rsp *schedule.SchedulesResponse) error {
	utils.InfoLog(ActionFindSchedules, utils.MsgProcessStarted)

	schedules, total, err := model.FindSchedules(req.GetDatabase(), req.GetUserId(), req.GetScheduleType(), req.GetPageIndex(), req.GetPageSize(), req.GetRunNow())
	if err != nil {
		utils.ErrorLog(ActionFindSchedules, err.Error())
		return err
	}

	res := &schedule.SchedulesResponse{}
	for _, t := range schedules {
		res.Schedules = append(res.Schedules, t.ToProto())
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindSchedules, utils.MsgProcessEnded)
	return nil
}

// FindSchedule 通过JobID获取任务
func (f *Schedule) FindSchedule(ctx context.Context, req *schedule.ScheduleRequest, rsp *schedule.ScheduleResponse) error {
	utils.InfoLog(ActionFindSchedule, utils.MsgProcessStarted)

	res, err := model.FindSchedule(req.GetDatabase(), req.GetScheduleId())
	if err != nil {
		utils.ErrorLog(ActionFindSchedule, err.Error())
		return err
	}

	rsp.Schedule = res.ToProto()

	utils.InfoLog(ActionFindSchedule, utils.MsgProcessEnded)
	return nil
}

// AddSchedule 添加任务
func (f *Schedule) AddSchedule(ctx context.Context, req *schedule.AddRequest, rsp *schedule.AddResponse) error {
	utils.InfoLog(ActionAddSchedule, utils.MsgProcessStarted)

	param := model.Schedule{
		ScheduleName:  req.GetScheduleName(),
		ScheduleType:  req.GetScheduleType(),
		EntryID:       req.GetEntryId(),
		Spec:          req.GetSpec(),
		Multi:         req.GetMulti(),
		RetryTimes:    req.GetRetryTimes(),
		RetryInterval: req.GetRetryInterval(),
		RunNow:        req.GetRunNow(),
		StartTime:     req.GetStartTime(),
		EndTime:       req.GetEndTime(),
		Status:        req.GetStatus(),
		Params:        req.GetParams(),
		CreatedAt:     time.Now(),
		CreatedBy:     req.GetWriter(),
		UpdatedAt:     time.Now(),
		UpdatedBy:     req.GetWriter(),
	}

	id, err := model.AddSchedule(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionAddSchedule, err.Error())
		return err
	}

	rsp.ScheduleId = id

	utils.InfoLog(ActionAddSchedule, utils.MsgProcessEnded)

	return nil
}

// ModifySchedule 更新任务
func (f *Schedule) ModifySchedule(ctx context.Context, req *schedule.ModifyRequest, rsp *schedule.ModifyResponse) error {
	utils.InfoLog(ActionModifySchedule, utils.MsgProcessStarted)

	param := model.ModifyParam{
		ScheduleID: req.GetScheduleId(),
		EntryID:    req.GetEntryId(),
		Status:     req.GetStatus(),
		Database:   req.GetDatabase(),
		Writer:     req.GetWriter(),
	}

	err := model.ModifySchedule(&param)
	if err != nil {
		utils.ErrorLog(ActionModifySchedule, err.Error())
		return err
	}

	utils.InfoLog(ActionModifySchedule, utils.MsgProcessEnded)
	return nil
}

// DeleteSchedule 删除任务
func (f *Schedule) DeleteSchedule(ctx context.Context, req *schedule.DeleteRequest, rsp *schedule.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSchedule, utils.MsgProcessStarted)

	err := model.DeleteSchedule(req.GetDatabase(), req.GetScheduleIds())
	if err != nil {
		utils.ErrorLog(ActionDeleteSchedule, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSchedule, utils.MsgProcessEnded)
	return nil
}

// AddScheduleNameUniqueIndex 添加schedule_name唯一索引
func (f *Schedule) AddScheduleNameUniqueIndex(ctx context.Context, req *schedule.ScheduleNameIndexRequest, rsp *schedule.ScheduleNameIndexResponse) error {
	utils.InfoLog(ActionAddScheduleNameUniqueIndex, utils.MsgProcessStarted)

	err := model.AddScheduleNameUniqueIndex(ctx, req.GetDb(), req.GetUserId())
	if err != nil {
		utils.ErrorLog(ActionAddScheduleNameUniqueIndex, err.Error())
		return err
	}

	utils.InfoLog(ActionAddScheduleNameUniqueIndex, utils.MsgProcessEnded)
	return nil
}
