package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/workflow/model"
	"rxcsoft.cn/pit3/srv/workflow/proto/process"
	"rxcsoft.cn/pit3/srv/workflow/utils"
)

// Process 流程的进程
type Process struct{}

// log出力使用
const (
	ProcessProcessName = "Process"

	ActionFindProcesss  = "FindProcesss"
	ActionFindsProcesss = "FindsProcesss"
	ActionAddProcess    = "AddProcess"
	ActionModifyProcess = "ModifyProcess"
	ActionDeleteProcess = "DeleteProcess"
)

// FindProcesses 获取多个流程的进程
func (f *Process) FindProcesses(ctx context.Context, req *process.ProcessesRequest, rsp *process.ProcessesResponse) error {
	utils.InfoLog(ActionFindProcesss, utils.MsgProcessStarted)

	processs, err := model.FindProcesses(req.GetDatabase(), req.GetExId())
	if err != nil {
		utils.ErrorLog(ActionFindProcesss, err.Error())
		return err
	}

	res := &process.ProcessesResponse{}
	for _, t := range processs {
		res.Processes = append(res.Processes, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindProcesss, utils.MsgProcessEnded)
	return nil
}

// FindsProcesses 获取所有流程的进程
func (f *Process) FindsProcesses(ctx context.Context, req *process.FindsProcessesRequest, rsp *process.FindsProcessesResponse) error {
	utils.InfoLog(ActionFindsProcesss, utils.MsgProcessStarted)

	processs, err := model.FindsProcesses(req.GetUserId(), req.GetDatabase())
	if err != nil {
		utils.ErrorLog(ActionFindsProcesss, err.Error())
		return err
	}

	res := &process.FindsProcessesResponse{}
	for _, t := range processs {
		res.Processes = append(res.Processes, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindsProcesss, utils.MsgProcessEnded)
	return nil
}

// AddProcess 添加流程的进程
func (f *Process) AddProcess(ctx context.Context, req *process.AddRequest, rsp *process.AddResponse) error {
	utils.InfoLog(ActionAddProcess, utils.MsgProcessStarted)

	param := model.Process{
		ExampleID:   req.GetExId(),
		CurrentNode: req.GetCurrentNode(),
		UserID:      req.GetUserId(),
		ExpireDate:  req.GetExpireDate(),
		Status:      req.GetStatus(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	id, err := model.AddProcess(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionAddProcess, err.Error())
		return err
	}

	rsp.ProId = id

	utils.InfoLog(ActionAddProcess, utils.MsgProcessEnded)

	return nil
}

// ModifyProcess 更新流程的进程
func (f *Process) ModifyProcess(ctx context.Context, req *process.ModifyRequest, rsp *process.ModifyResponse) error {
	utils.InfoLog(ActionModifyProcess, utils.MsgProcessStarted)

	err := model.ModifyProcess(req.GetDatabase(), req.GetProId(), req.GetStatus(), req.GetComment(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionModifyProcess, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyProcess, utils.MsgProcessEnded)
	return nil
}

// DeleteProcess 删除流程的进程
func (f *Process) DeleteProcess(ctx context.Context, req *process.DeleteRequest, rsp *process.DeleteResponse) error {
	utils.InfoLog(ActionDeleteProcess, utils.MsgProcessStarted)

	err := model.DeleteProcess(req.GetDatabase(), req.GetExId())
	if err != nil {
		utils.ErrorLog(ActionDeleteProcess, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteProcess, utils.MsgProcessEnded)
	return nil
}
