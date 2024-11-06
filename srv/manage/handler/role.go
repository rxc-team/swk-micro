package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Role 角色
type Role struct{}

// log出力使用
const (
	RoleProcessName          = "Role"
	ActionFindRoles          = "FindRoles"
	ActionFindRole           = "FindRole"
	ActionAddRole            = "AddRole"
	ActionModifyRole         = "ModifyRole"
	ActionDeleteRole         = "DeleteRole"
	ActionDeleteSelectRoles  = "DeleteSelectRoles"
	ActionHardDeleteRoles    = "HardDeleteRoles"
	ActionRecoverSelectRoles = "RecoverSelectRoles"
	ActionWhitelistClear     = "WhitelistClear"
)

// FindRoles 查找多个角色记录
func (r *Role) FindRoles(ctx context.Context, req *role.FindRolesRequest, rsp *role.FindRolesResponse) error {
	utils.InfoLog(ActionFindRoles, utils.MsgProcessStarted)

	roles, err := model.FindRoles(ctx, req.GetDatabase(), req.GetRoleId(), req.GetRoleType(), req.GetRoleName(), req.GetDescription(), req.GetDomain(), req.GetInvalidatedIn())
	if err != nil {
		utils.ErrorLog(ActionFindRoles, err.Error())
		return err
	}

	res := &role.FindRolesResponse{}
	for _, r := range roles {
		res.Roles = append(res.Roles, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindRoles, utils.MsgProcessEnded)
	return nil
}

// FindRole 查找单个角色记录
func (r *Role) FindRole(ctx context.Context, req *role.FindRoleRequest, rsp *role.FindRoleResponse) error {
	utils.InfoLog(ActionFindRole, utils.MsgProcessStarted)

	res, err := model.FindRole(ctx, req.GetDatabase(), req.GetRoleId())
	if err != nil {
		utils.ErrorLog(ActionFindRole, err.Error())
		return err
	}

	rsp.Role = res.ToProto()

	utils.InfoLog(ActionFindRole, utils.MsgProcessEnded)
	return nil
}

// AddRole 添加单个角色记录
func (r *Role) AddRole(ctx context.Context, req *role.AddRoleRequest, rsp *role.AddRoleResponse) error {
	utils.InfoLog(ActionAddRole, utils.MsgProcessStarted)

	ps := make([]*model.Permission, 0)
	for _, p := range req.GetPermissions() {

		var as []*model.PAction

		for _, act := range p.GetActions() {

			fields := act.GetFields()
			if len(fields) == 0 {
				fields = make([]string, 0)
			}
			actions := act.GetActionMap()
			if len(actions) == 0 {
				actions = make(map[string]bool)
			}

			as = append(as, &model.PAction{
				ObjectId:  act.GetObjectId(),
				Fields:    fields,
				ActionMap: actions,
			})
		}

		ps = append(ps, &model.Permission{
			PermissionType: p.GetPermissionType(),
			AppId:          p.GetAppId(),
			ActionType:     p.GetActionType(),
			Actions:        as,
			CreatedAt:      time.Now(),
			CreatedBy:      req.GetWriter(),
			UpdatedAt:      time.Now(),
			UpdatedBy:      req.GetWriter(),
		})
	}

	ips := make([]model.IPSegment, 0)
	for _, ip := range req.GetIpSegments() {
		ips = append(ips, model.IPSegment{
			Start: ip.GetStart(),
			End:   ip.GetEnd(),
		})
	}

	params := model.Role{
		RoleName:    req.GetRoleName(),
		Description: req.GetDescription(),
		Domain:      req.GetDomain(),
		IPSegments:  ips,
		Menus:       req.GetMenus(),
		RoleType:    req.GetRoleType(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	id, err := model.AddRole(ctx, req.GetDatabase(), &params, ps)
	if err != nil {
		utils.ErrorLog(ActionAddRole, err.Error())
		return err
	}

	rsp.RoleId = id

	utils.InfoLog(ActionAddRole, utils.MsgProcessEnded)
	return nil
}

// ModifyRole 更新角色的信息
func (r *Role) ModifyRole(ctx context.Context, req *role.ModifyRoleRequest, rsp *role.ModifyRoleResponse) error {
	utils.InfoLog(ActionModifyRole, utils.MsgProcessStarted)

	var ps []*model.Permission
	for _, p := range req.GetPermissions() {

		var as []*model.PAction

		for _, act := range p.GetActions() {
			fields := act.GetFields()
			if len(fields) == 0 {
				fields = make([]string, 0)
			}
			actions := act.GetActionMap()
			if len(actions) == 0 {
				actions = make(map[string]bool)
			}

			as = append(as, &model.PAction{
				ObjectId:  act.GetObjectId(),
				Fields:    fields,
				ActionMap: actions,
			})
		}

		ps = append(ps, &model.Permission{
			RoleId:         req.GetRoleId(),
			PermissionType: p.GetPermissionType(),
			AppId:          p.GetAppId(),
			ActionType:     p.GetActionType(),
			Actions:        as,
			CreatedAt:      time.Now(),
			CreatedBy:      req.GetWriter(),
			UpdatedAt:      time.Now(),
			UpdatedBy:      req.GetWriter(),
		})
	}

	var ips []model.IPSegment
	for _, ip := range req.GetIpSegments() {
		ips = append(ips, model.IPSegment{
			Start: ip.GetStart(),
			End:   ip.GetEnd(),
		})
	}

	params := model.Role{
		RoleID:      req.GetRoleId(),
		RoleName:    req.GetRoleName(),
		Description: req.GetDescription(),
		IPSegments:  ips,
		Menus:       req.GetMenus(),
		RoleType:    req.RoleType,
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	err := model.ModifyRole(ctx, req.GetDatabase(), &params, ps)
	if err != nil {
		utils.ErrorLog(ActionModifyRole, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyRole, utils.MsgProcessEnded)
	return nil
}

// DeleteRole 删除单个角色
func (r *Role) DeleteRole(ctx context.Context, req *role.DeleteRoleRequest, rsp *role.DeleteRoleResponse) error {
	utils.InfoLog(ActionDeleteRole, utils.MsgProcessStarted)

	err := model.DeleteRole(ctx, req.GetDatabase(), req.GetRoleId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteRole, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteRole, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectRoles 删除选中角色
func (r *Role) DeleteSelectRoles(ctx context.Context, req *role.DeleteSelectRolesRequest, rsp *role.DeleteSelectRolesResponse) error {
	utils.InfoLog(ActionDeleteSelectRoles, utils.MsgProcessStarted)

	err := model.DeleteSelectRoles(ctx, req.GetDatabase(), req.GetRoleIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectRoles, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectRoles, utils.MsgProcessEnded)
	return nil
}

// HardDeleteRoles 物理删除选中角色
func (r *Role) HardDeleteRoles(ctx context.Context, req *role.HardDeleteRolesRequest, rsp *role.HardDeleteRolesResponse) error {
	utils.InfoLog(ActionHardDeleteRoles, utils.MsgProcessStarted)

	err := model.HardDeleteRoles(ctx, req.GetDatabase(), req.GetRoleIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteRoles, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteRoles, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectRoles 恢复选中角色
func (r *Role) RecoverSelectRoles(ctx context.Context, req *role.RecoverSelectRolesRequest, rsp *role.RecoverSelectRolesResponse) error {
	utils.InfoLog(ActionRecoverSelectRoles, utils.MsgProcessStarted)

	err := model.RecoverSelectRoles(ctx, req.GetDatabase(), req.GetRoleIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectRoles, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectRoles, utils.MsgProcessEnded)
	return nil
}

// WhitelistClear 清空白名单
func (r *Role) WhitelistClear(ctx context.Context, req *role.WhitelistClearRequest, rsp *role.WhitelistClearResponse) error {
	utils.InfoLog(ActionWhitelistClear, utils.MsgProcessStarted)

	err := model.WhitelistClear(ctx, req.GetWriter(), req.GetDatabase())
	if err != nil {
		utils.ErrorLog(ActionWhitelistClear, err.Error())
		return err
	}

	utils.InfoLog(ActionWhitelistClear, utils.MsgProcessEnded)
	return nil
}
