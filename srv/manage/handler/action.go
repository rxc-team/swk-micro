package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/action"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Action 许可权限
type Action struct{}

// log出力使用
const (
	ActionProcessName = "Action"

	ActionFindActions   = "FindActions"
	ActionFindAction    = "FindAction"
	ActionAddAction     = "AddAction"
	ActionModifyAction  = "ModifyAction"
	ActionDeleteAction  = "DeleteAction"
	ActionDeleteActions = "DeleteActions"
)

// FindActions 查找多个许可权限记录
func (r *Action) FindActions(ctx context.Context, req *action.FindActionsRequest, rsp *action.FindActionsResponse) error {
	utils.InfoLog(ActionFindActions, utils.MsgProcessStarted)

	actions, err := model.FindActions(ctx, req.GetActionGroup())
	if err != nil {
		utils.ErrorLog(ActionFindActions, err.Error())
		return err
	}

	res := &action.FindActionsResponse{}
	for _, r := range actions {
		res.Actions = append(res.Actions, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindActions, utils.MsgProcessEnded)
	return nil
}

// FindAction 查找单个许可权限记录
func (r *Action) FindAction(ctx context.Context, req *action.FindActionRequest, rsp *action.FindActionResponse) error {
	utils.InfoLog(ActionFindAction, utils.MsgProcessStarted)

	res, err := model.FindAction(ctx, req.GetActionKey(), req.GetActionObject())
	if err != nil {
		utils.ErrorLog(ActionFindAction, err.Error())
		return err
	}

	rsp.Action = res.ToProto()

	utils.InfoLog(ActionFindAction, utils.MsgProcessEnded)
	return nil
}

// AddAction 添加单个许可权限记录
func (r *Action) AddAction(ctx context.Context, req *action.AddActionRequest, rsp *action.AddActionResponse) error {
	utils.InfoLog(ActionAddAction, utils.MsgProcessStarted)

	params := model.Action{
		ActionKey:    req.GetActionKey(),
		ActionName:   req.GetActionName(),
		ActionObject: req.GetActionObject(),
		ActionGroup:  req.GetActionGroup(),
		CreatedAt:    time.Now(),
		CreatedBy:    req.GetWriter(),
		UpdatedAt:    time.Now(),
		UpdatedBy:    req.GetWriter(),
	}

	err := model.AddAction(ctx, &params)
	if err != nil {
		utils.ErrorLog(ActionAddAction, err.Error())
		return err
	}

	utils.InfoLog(ActionAddAction, utils.MsgProcessEnded)
	return nil
}

// ModifyAction 更新许可权限的信息
func (r *Action) ModifyAction(ctx context.Context, req *action.ModifyActionRequest, rsp *action.ModifyActionResponse) error {
	utils.InfoLog(ActionModifyAction, utils.MsgProcessStarted)

	params := model.Action{
		ActionObject: req.GetActionObject(),
		ActionKey:    req.GetActionKey(),
		ActionName:   req.GetActionName(),
		ActionGroup:  req.GetActionGroup(),
		UpdatedAt:    time.Now(),
		UpdatedBy:    req.GetWriter(),
	}

	err := model.ModifyAction(ctx, params)
	if err != nil {
		utils.ErrorLog(ActionModifyAction, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyAction, utils.MsgProcessEnded)
	return nil
}

// DeleteAction 硬删除单个许可权限
func (r *Action) DeleteAction(ctx context.Context, req *action.DeleteActionRequest, rsp *action.DeleteActionResponse) error {
	utils.InfoLog(ActionDeleteAction, utils.MsgProcessStarted)

	err := model.DeleteAction(ctx, req.GetActionKey(), req.GetActionObject())
	if err != nil {
		utils.ErrorLog(ActionDeleteAction, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteAction, utils.MsgProcessEnded)
	return nil
}

// DeleteActions 硬删除多个许可权限
func (r *Action) DeleteActions(ctx context.Context, req *action.DeleteActionsRequest, rsp *action.DeleteActionsResponse) error {
	utils.InfoLog(ActionDeleteActions, utils.MsgProcessStarted)

	var dels []*model.ActionDelParam
	for _, d := range req.GetDels() {
		del := &model.ActionDelParam{
			ActionKey:    d.ActionKey,
			ActionObject: d.ActionObject,
		}
		dels = append(dels, del)
	}

	err := model.DeleteActions(ctx, dels)
	if err != nil {
		utils.ErrorLog(ActionDeleteActions, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteActions, utils.MsgProcessEnded)
	return nil
}
