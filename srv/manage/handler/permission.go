package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Permission 许可权限
type Permission struct{}

// log出力使用
const (
	PermissionProcessName = "Permission"
	ActionFindPermissions = "FindPermissions"
)

// FindActions 查找多个许可权限记录
func (r *Permission) FindActions(ctx context.Context, req *permission.FindActionsRequest, rsp *permission.FindActionsResponse) error {
	utils.InfoLog(ActionFindActions, utils.MsgProcessStarted)

	permissions, err := model.FindPActions(ctx, req.GetDatabase(), req.GetAppId(), req.GetPermissionType(), req.GetActionType(), req.GetObjectId(), req.GetRoleId())
	if err != nil {
		utils.ErrorLog(ActionFindActions, err.Error())
		return err
	}

	res := &permission.FindActionsResponse{}
	for _, a := range permissions {
		res.Actions = append(res.Actions, a.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindActions, utils.MsgProcessEnded)
	return nil
}

// FindPermissions 查找多个许可权限记录
func (r *Permission) FindPermissions(ctx context.Context, req *permission.FindPermissionsRequest, rsp *permission.FindPermissionsResponse) error {
	utils.InfoLog(ActionFindPermissions, utils.MsgProcessStarted)

	permissions, err := model.FindPermissions(ctx, req.GetDatabase(), req.GetRoleId())
	if err != nil {
		utils.ErrorLog(ActionFindPermissions, err.Error())
		return err
	}

	res := &permission.FindPermissionsResponse{}
	for _, a := range permissions {
		res.Permission = append(res.Permission, a.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindPermissions, utils.MsgProcessEnded)
	return nil
}
