package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/workflow/model"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
	"rxcsoft.cn/pit3/srv/workflow/utils"
)

// Workflow 流程
type Workflow struct{}

// log出力使用
const (
	WorkflowProcessName = "Workflow"

	ActionFindWorkflows     = "FindWorkflows"
	ActionFindUserWorkflows = "FindUserWorkflows"
	ActionFindWorkflow      = "FindWorkflow"
	ActionAddWorkflow       = "AddWorkflow"
	ActionModifyWorkflow    = "ModifyWorkflow"
	ActionDeleteWorkflow    = "DeleteWorkflow"
)

// FindWorkflows 获取多个流程
func (f *Workflow) FindWorkflows(ctx context.Context, req *workflow.WorkflowsRequest, rsp *workflow.WorkflowsResponse) error {
	utils.InfoLog(ActionFindWorkflows, utils.MsgProcessStarted)

	workflows, err := model.FindWorkflows(req.GetDatabase(), req.GetAppId(), req.GetIsValid(), req.GetGroupId(), req.GetObjectId(), req.GetAction())
	if err != nil {
		utils.ErrorLog(ActionFindWorkflows, err.Error())
		return err
	}

	res := &workflow.WorkflowsResponse{}
	for _, t := range workflows {
		res.Workflows = append(res.Workflows, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindWorkflows, utils.MsgProcessEnded)
	return nil
}

// FindWorkflows 获取多个流程
func (f *Workflow) FindUserWorkflows(ctx context.Context, req *workflow.UserWorkflowsRequest, rsp *workflow.UserWorkflowsResponse) error {
	utils.InfoLog(ActionFindUserWorkflows, utils.MsgProcessStarted)

	workflows, err := model.FindUserWorkflows(req.GetDatabase(), req.GetAppId(), req.GetObjectId(), req.GetGroupId(), req.GetAction())
	if err != nil {
		utils.ErrorLog(ActionFindUserWorkflows, err.Error())
		return err
	}

	res := &workflow.UserWorkflowsResponse{}
	for _, t := range workflows {
		res.Workflows = append(res.Workflows, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindUserWorkflows, utils.MsgProcessEnded)
	return nil
}

// FindWorkflow 通过JobID获取流程
func (f *Workflow) FindWorkflow(ctx context.Context, req *workflow.WorkflowRequest, rsp *workflow.WorkflowResponse) error {
	utils.InfoLog(ActionFindWorkflow, utils.MsgProcessStarted)

	res, err := model.FindWorkflow(req.GetDatabase(), req.GetWfId())
	if err != nil {
		utils.ErrorLog(ActionFindWorkflow, err.Error())
		return err
	}

	rsp.Workflow = res.ToProto()

	utils.InfoLog(ActionFindWorkflow, utils.MsgProcessEnded)
	return nil
}

// AddWorkflow 添加流程
func (f *Workflow) AddWorkflow(ctx context.Context, req *workflow.AddRequest, rsp *workflow.AddResponse) error {
	utils.InfoLog(ActionAddWorkflow, utils.MsgProcessStarted)

	param := model.Workflow{
		WorkflowName:    req.GetWfName(),
		MenuName:        req.GetMenuName(),
		IsValid:         req.GetIsValid(),
		GroupID:         req.GetGroupId(),
		AppID:           req.GetAppId(),
		WorkflowType:    req.GetWorkflowType(),
		AcceptOrDismiss: req.GetAcceptOrDismiss(),
		Params:          req.GetParams(),
		CreatedAt:       time.Now(),
		CreatedBy:       req.GetWriter(),
		UpdatedAt:       time.Now(),
		UpdatedBy:       req.GetWriter(),
	}

	id, err := model.AddWorkflow(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionAddWorkflow, err.Error())
		return err
	}

	rsp.WfId = id

	utils.InfoLog(ActionAddWorkflow, utils.MsgProcessEnded)

	return nil
}

// ModifyWorkflow 更新流程
func (f *Workflow) ModifyWorkflow(ctx context.Context, req *workflow.ModifyRequest, rsp *workflow.ModifyResponse) error {
	utils.InfoLog(ActionModifyWorkflow, utils.MsgProcessStarted)

	param := model.WorkflowUpdateParam{
		WfId:            req.GetWfId(),
		IsValid:         req.GetIsValid(),
		AcceptOrDismiss: req.GetAcceptOrDismiss(),
		Params:          req.GetParams(),
		Writer:          req.GetWriter(),
	}

	err := model.ModifyWorkflow(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionModifyWorkflow, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyWorkflow, utils.MsgProcessEnded)
	return nil
}

// DeleteWorkflow 删除流程
func (f *Workflow) DeleteWorkflow(ctx context.Context, req *workflow.DeleteRequest, rsp *workflow.DeleteResponse) error {
	utils.InfoLog(ActionDeleteWorkflow, utils.MsgProcessStarted)

	err := model.DeleteWorkflow(req.GetDatabase(), req.GetWorkflows())
	if err != nil {
		utils.ErrorLog(ActionDeleteWorkflow, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteWorkflow, utils.MsgProcessEnded)
	return nil
}
