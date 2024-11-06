package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/access"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Access 权限
type Access struct{}

// log出力使用
const (
	AccessProcessName = "Access"

	ActionFindUserAccesss     = "FindUserAccesss"
	ActionFindAccesss         = "FindAccesss"
	ActionFindAccess          = "FindAccess"
	ActionAddAccess           = "AddAccess"
	ActionModifyAccess        = "ModifyAccess"
	ActionDeleteAccess        = "DeleteAccess"
	ActionDeleteSelectAccess  = "DeleteSelectAccess"
	ActionHardDeleteAccesss   = "HardDeleteAccess"
	ActionRecoverSelectAccess = "RecoverSelectAccess"
)

// FindUserAccess 查找多个权限记录
func (r *Access) FindUserAccess(ctx context.Context, req *access.FindUserAccesssRequest, rsp *access.FindUserAccesssResponse) error {
	utils.InfoLog(ActionFindUserAccesss, utils.MsgProcessStarted)

	access, err := model.FindUserAccess(ctx, req.GetDatabase(), req.GetGroupId(), req.GetAppId(), req.GetDatastoreId(), req.GetOwner(), req.GetAction(), req.GetRoleId())
	if err != nil {
		utils.ErrorLog(ActionFindUserAccesss, err.Error())
		return err
	}

	rsp.AccessKeys = access

	utils.InfoLog(ActionFindUserAccesss, utils.MsgProcessEnded)
	return nil
}

// FindAccess 查找多个权限记录
func (r *Access) FindAccess(ctx context.Context, req *access.FindAccessRequest, rsp *access.FindAccessResponse) error {
	utils.InfoLog(ActionFindAccesss, utils.MsgProcessStarted)

	result, err := model.FindAccess(ctx, req.GetDatabase(), req.GetRoleId(), req.GetGroupId())
	if err != nil {
		utils.ErrorLog(ActionFindAccesss, err.Error())
		return err
	}

	res := &access.FindAccessResponse{}
	for _, r := range result {
		res.AccessList = append(res.AccessList, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindAccesss, utils.MsgProcessEnded)
	return nil
}

// FindOneAccess 查找单个权限记录
func (r *Access) FindOneAccess(ctx context.Context, req *access.FindOneAccessRequest, rsp *access.FindOneAccessResponse) error {
	utils.InfoLog(ActionFindAccess, utils.MsgProcessStarted)

	res, err := model.FindOneAccess(ctx, req.GetDatabase(), req.GetAccessId())
	if err != nil {
		utils.ErrorLog(ActionFindAccess, err.Error())
		return err
	}

	rsp.Access = res.ToProto()

	utils.InfoLog(ActionFindAccess, utils.MsgProcessEnded)
	return nil
}

// AddAccess 添加单个权限记录
func (r *Access) AddAccess(ctx context.Context, req *access.AddAccessRequest, rsp *access.AddAccessResponse) error {
	utils.InfoLog(ActionAddAccess, utils.MsgProcessStarted)

	apps := make(map[string]*model.AppData)

	for ak, av := range req.GetApps() {
		dataAccess := make(map[string]*model.DataAccess)
		for dk, dv := range av.DataAccess {
			var actions []*model.DataAction
			for _, a := range dv.Actions {
				actions = append(actions, &model.DataAction{
					GroupID:   a.GroupId,
					AccessKey: a.AccessKey,
					CanFind:   a.CanFind,
					CanUpdate: a.CanUpdate,
					CanDelete: a.CanDelete,
				})
			}

			dataAccess[dk] = &model.DataAccess{
				Actions: actions,
			}
		}

		apps[ak] = &model.AppData{
			DataAccess: dataAccess,
		}
	}

	params := model.Access{
		RoleID:    req.GetRoleId(),
		GroupID:   req.GetGroupId(),
		Apps:      apps,
		CreatedAt: time.Now(),
		CreatedBy: req.GetWriter(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	id, err := model.AddAccess(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddAccess, err.Error())
		return err
	}

	rsp.AccessId = id

	utils.InfoLog(ActionAddAccess, utils.MsgProcessEnded)
	return nil
}

// HardDeleteAccesss 物理删除选中权限
func (r *Access) HardDeleteAccess(ctx context.Context, req *access.HardDeleteAccessRequest, rsp *access.HardDeleteAccessResponse) error {
	utils.InfoLog(ActionHardDeleteAccesss, utils.MsgProcessStarted)

	err := model.HardDeleteAccess(ctx, req.GetDatabase(), req.GetAccessList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteAccesss, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteAccesss, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectAccesss 恢复选中权限
func (r *Access) RecoverSelectAccess(ctx context.Context, req *access.RecoverSelectAccessRequest, rsp *access.RecoverSelectAccessResponse) error {
	utils.InfoLog(ActionRecoverSelectAccess, utils.MsgProcessStarted)

	err := model.RecoverSelectAccess(ctx, req.GetDatabase(), req.GetAccessList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectAccess, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectAccess, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectAccess 恢复选中权限
func (r *Access) DeleteSelectAccess(ctx context.Context, req *access.DeleteSelectAccessRequest, rsp *access.DeleteSelectAccessResponse) error {
	utils.InfoLog(ActionDeleteSelectAccess, utils.MsgProcessStarted)

	err := model.DeleteSelectAccess(ctx, req.GetDatabase(), req.GetWriter(), req.GetAccessList())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectAccess, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectAccess, utils.MsgProcessEnded)
	return nil
}
