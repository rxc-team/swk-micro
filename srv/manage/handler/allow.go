package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/allow"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Allow 许可
type Allow struct{}

// log出力使用
const (
	AllowProcessName = "Allow"

	AllowFindAllows   = "FindAllows"
	AllowFindAllow    = "FindAllow"
	AllowAddAllow     = "AddAllow"
	AllowModifyAllow  = "ModifyAllow"
	AllowDeleteAllow  = "DeleteAllow"
	AllowDeleteAllows = "DeleteAllows"
)

// FindAllows 查找多个许可记录
func (r *Allow) FindAllows(ctx context.Context, req *allow.FindAllowsRequest, rsp *allow.FindAllowsResponse) error {
	utils.InfoLog(AllowFindAllows, utils.MsgProcessStarted)

	allows, err := model.FindAllows(ctx, req.GetAllowType(), req.GetObjectType())
	if err != nil {
		utils.ErrorLog(AllowFindAllows, err.Error())
		return err
	}

	res := &allow.FindAllowsResponse{}
	for _, r := range allows {
		res.Allows = append(res.Allows, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(AllowFindAllows, utils.MsgProcessEnded)
	return nil
}

// FindLevelAllows 查找多个许可记录
func (r *Allow) FindLevelAllows(ctx context.Context, req *allow.FindLevelAllowsRequest, rsp *allow.FindLevelAllowsResponse) error {
	utils.InfoLog(AllowFindAllows, utils.MsgProcessStarted)

	allows, err := model.FindLevelAllows(ctx, req.GetAllowList())
	if err != nil {
		utils.ErrorLog(AllowFindAllows, err.Error())
		return err
	}

	res := &allow.FindLevelAllowsResponse{}
	for _, r := range allows {
		res.Allows = append(res.Allows, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(AllowFindAllows, utils.MsgProcessEnded)
	return nil
}

// FindAllow 查找单个许可记录
func (r *Allow) FindAllow(ctx context.Context, req *allow.FindAllowRequest, rsp *allow.FindAllowResponse) error {
	utils.InfoLog(AllowFindAllow, utils.MsgProcessStarted)

	res, err := model.FindAllow(ctx, req.GetAllowId())
	if err != nil {
		utils.ErrorLog(AllowFindAllow, err.Error())
		return err
	}

	rsp.Allow = res.ToProto()

	utils.InfoLog(AllowFindAllow, utils.MsgProcessEnded)
	return nil
}

// AddAllow 添加单个许可记录
func (r *Allow) AddAllow(ctx context.Context, req *allow.AddAllowRequest, rsp *allow.AddAllowResponse) error {
	utils.InfoLog(AllowAddAllow, utils.MsgProcessStarted)

	var actions []*model.AAction

	for _, act := range req.GetActions() {
		actions = append(actions, &model.AAction{
			ApiKey:     act.ApiKey,
			ActionName: act.ActionName,
			GroupKey:   act.GroupKey,
		})
	}

	params := model.Allow{
		AllowName:  req.GetAllowName(),
		AllowType:  req.GetAllowType(),
		ObjectType: req.GetObjectType(),
		Actions:    actions,
		CreatedAt:  time.Now(),
		CreatedBy:  req.GetWriter(),
		UpdatedAt:  time.Now(),
		UpdatedBy:  req.GetWriter(),
	}

	err := model.AddAllow(ctx, &params)
	if err != nil {
		utils.ErrorLog(AllowAddAllow, err.Error())
		return err
	}

	utils.InfoLog(AllowAddAllow, utils.MsgProcessEnded)
	return nil
}

// ModifyAllow 更新许可的信息
func (r *Allow) ModifyAllow(ctx context.Context, req *allow.ModifyAllowRequest, rsp *allow.ModifyAllowResponse) error {
	utils.InfoLog(AllowModifyAllow, utils.MsgProcessStarted)

	var actions []model.AAction

	for _, act := range req.GetActions() {
		actions = append(actions, model.AAction{
			ApiKey:     act.ApiKey,
			ActionName: act.ActionName,
			GroupKey:   act.GroupKey,
		})
	}

	err := model.ModifyAllow(ctx, req.GetAllowId(), req.GetAllowName(), req.GetAllowType(), req.GetObjectType(), req.GetWriter(), actions)
	if err != nil {
		utils.ErrorLog(AllowModifyAllow, err.Error())
		return err
	}

	utils.InfoLog(AllowModifyAllow, utils.MsgProcessEnded)
	return nil
}

// DeleteAllow 硬删除单个许可操作记录
func (r *Allow) DeleteAllow(ctx context.Context, req *allow.DeleteAllowRequest, rsp *allow.DeleteAllowResponse) error {
	utils.InfoLog(AllowDeleteAllow, utils.MsgProcessStarted)

	err := model.DeleteAllow(ctx, req.GetAllowId())
	if err != nil {
		utils.ErrorLog(AllowDeleteAllow, err.Error())
		return err
	}

	utils.InfoLog(AllowDeleteAllow, utils.MsgProcessEnded)
	return nil
}

// DeleteAllows 硬删除多个许可操作记录
func (r *Allow) DeleteAllows(ctx context.Context, req *allow.DeleteAllowsRequest, rsp *allow.DeleteAllowsResponse) error {
	utils.InfoLog(AllowDeleteAllows, utils.MsgProcessStarted)

	err := model.DeleteAllows(ctx, req.GetAllowIds())
	if err != nil {
		utils.ErrorLog(AllowDeleteAllows, err.Error())
		return err
	}

	utils.InfoLog(AllowDeleteAllows, utils.MsgProcessEnded)
	return nil
}
