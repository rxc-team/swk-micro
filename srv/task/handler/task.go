package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/task/model"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/pit3/srv/task/utils"
)

// Task 任务
type Task struct{}

// log出力使用
const (
	TaskProcessName = "Task"

	ActionFindTasks  = "FindTasks"
	ActionFindTask   = "FindTask"
	ActionAddTask    = "AddTask"
	ActionModifyTask = "ModifyTask"
	ActionDeleteTask = "DeleteTask"
)

// FindTasks 获取多个任务
func (f *Task) FindTasks(ctx context.Context, req *task.TasksRequest, rsp *task.TasksResponse) error {
	utils.InfoLog(ActionFindTasks, utils.MsgProcessStarted)

	tasks, total, err := model.FindTasks(req.GetDatabase(), req.GetUserId(), req.GetScheduleId(), req.GetAppId(), req.GetPageIndex(), req.GetPageSize())
	if err != nil {
		utils.ErrorLog(ActionFindTasks, err.Error())
		return err
	}

	res := &task.TasksResponse{}
	for _, t := range tasks {
		res.Tasks = append(res.Tasks, t.ToProto())
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindTasks, utils.MsgProcessEnded)
	return nil
}

// FindTask 通过JobID获取任务
func (f *Task) FindTask(ctx context.Context, req *task.TaskRequest, rsp *task.TaskResponse) error {
	utils.InfoLog(ActionFindTask, utils.MsgProcessStarted)

	res, err := model.FindTask(req.GetDatabase(), req.GetJobId())
	if err != nil {
		utils.ErrorLog(ActionFindTask, err.Error())
		return err
	}

	rsp.Task = res.ToProto()

	utils.InfoLog(ActionFindTask, utils.MsgProcessEnded)
	return nil
}

// AddTask 添加任务
func (f *Task) AddTask(ctx context.Context, req *task.AddRequest, rsp *task.AddResponse) error {
	utils.InfoLog(ActionAddTask, utils.MsgProcessStarted)

	param := model.Task{
		JobID:        req.GetJobId(),
		ScheduleID:   req.GetScheduleId(),
		JobName:      req.GetJobName(),
		Origin:       req.GetOrigin(),
		UserID:       req.GetUserId(),
		ShowProgress: req.GetShowProgress(),
		Progress:     req.GetProgress(),
		StartTime:    req.GetStartTime(),
		Message:      req.GetMessage(),
		TaskType:     req.GetTaskType(),
		Steps:        req.GetSteps(),
		CurrentStep:  req.GetCurrentStep(),
		AppID:        req.GetAppId(),
	}

	id, err := model.AddTask(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionAddTask, err.Error())
		return err
	}
	rsp.JobId = id

	utils.InfoLog(ActionAddTask, utils.MsgProcessEnded)

	return nil
}

// ModifyTask 更新任务
func (f *Task) ModifyTask(ctx context.Context, req *task.ModifyRequest, rsp *task.ModifyResponse) error {
	utils.InfoLog(ActionModifyTask, utils.MsgProcessStarted)

	param := model.Task{
		JobID:       req.GetJobId(),
		Progress:    req.GetProgress(),
		EndTime:     req.GetEndTime(),
		Message:     req.GetMessage(),
		CurrentStep: req.GetCurrentStep(),
		Insert:      req.GetInsert(),
		Update:      req.GetUpdate(),
		Total:       req.GetTotal(),
	}

	if req.GetFile() != nil {
		file := model.File{
			URL:  req.GetFile().Url,
			Name: req.GetFile().Name,
		}
		param.File = file
	}

	if req.GetErrorFile() != nil {
		errorFile := model.File{
			URL:  req.GetErrorFile().Url,
			Name: req.GetErrorFile().Name,
		}
		param.ErrorFile = errorFile
	}

	err := model.ModifyTask(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionModifyTask, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyTask, utils.MsgProcessEnded)
	return nil
}

// DeleteTask 删除任务
func (f *Task) DeleteTask(ctx context.Context, req *task.DeleteRequest, rsp *task.DeleteResponse) error {
	utils.InfoLog(ActionDeleteTask, utils.MsgProcessStarted)

	err := model.DeleteTask(req.GetAppId(), req.GetUserId(), req.GetJobId())
	if err != nil {
		utils.ErrorLog(ActionDeleteTask, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteTask, utils.MsgProcessEnded)
	return nil
}
