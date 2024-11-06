package dev

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
)

// Role 角色
type Role struct{}

// log出力
const (
	RoleProcessName          = "Role"
	ActionFindRoles          = "FindRoles"
	ActionFindRole           = "FindRole"
	ActionAddRole            = "AddRole"
	ActionModifyRole         = "ModifyRole"
	ActionDeleteSelectRoles  = "DeleteSelectRoles"
	ActionHardDeleteRoles    = "HardDeleteRoles"
	ActionRecoverSelectRoles = "RecoverSelectRoles"
	ActionWhitelistClear     = "WhitelistClear"
)

// FindRoles 获取所有角色
// @Router /roles [get]
func (u *Role) FindRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindRoles, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.FindRolesRequest
	// 从query中获取参数
	req.RoleId = c.Query("role_id")
	req.RoleName = c.Query("role_name")
	req.Description = c.Query("description")
	req.InvalidatedIn = c.Query("invalidated_in")
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.FindRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindRoles, err)
		return
	}

	loggerx.InfoLog(c, ActionFindRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionFindRoles)),
		Data:    response.GetRoles(),
	})
}

// FindRole 获取角色
// @Router /roles/{role_id} [get]
func (u *Role) FindRole(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindRole, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.FindRoleRequest
	req.RoleId = c.Param("role_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := roleService.FindRole(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindRole, err)
		return
	}

	loggerx.InfoLog(c, ActionFindRole, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionFindRole)),
		Data:    response.GetRole(),
	})
}

// WhitelistClear 清空白名单
// @Router /role/whitelistClear/roles [PUT]
func (u *Role) WhitelistClear(c *gin.Context) {
	loggerx.InfoLog(c, ActionWhitelistClear, loggerx.MsgProcessStarted)

	var req role.WhitelistClearRequest

	req.Database = c.Query("database")
	req.Writer = sessionx.GetAuthUserID(c)

	roleService := role.NewRoleService("manage", client.DefaultClient)
	response, err := roleService.WhitelistClear(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionWhitelistClear, err)
		return
	}

	loggerx.SuccessLog(c, ActionWhitelistClear, fmt.Sprintf("customer[%s] administrator WhitelistClear Success", req.GetDatabase()))

	loggerx.InfoLog(c, ActionWhitelistClear, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionWhitelistClear)),
		Data:    response,
	})
}

// AddRole 添加角色
// @Router /roles [post]
func (u *Role) AddRole(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddRole, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.AddRoleRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddRole, err)
		return
	}
	req.RoleType = 3
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.AddRole(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddRole, err)
		return
	}

	loggerx.SuccessLog(c, ActionAddRole, fmt.Sprintf("Role[%s] create Success", response.GetRoleId()))

	loggerx.InfoLog(c, ActionAddRole, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionAddRole)),
		Data:    response,
	})
}

// ModifyRole 更新角色
// @Router /roles/{role_id} [put]
func (u *Role) ModifyRole(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyRole, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.ModifyRoleRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyRole, err)
		return
	}

	// 从path中获取参数
	req.RoleId = c.Param("role_id")
	// 从body中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.ModifyRole(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyRole, err)
		return
	}

	loggerx.SuccessLog(c, ActionModifyRole, fmt.Sprintf("Role[%s] update Success", req.GetRoleId()))

	loggerx.InfoLog(c, ActionModifyRole, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionModifyRole)),
		Data:    response,
	})
}

// DeleteSelectRoles 删除选中角色
// @Router /roles [delete]
func (u *Role) DeleteSelectRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSelectRoles, loggerx.MsgProcessStarted)
	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.DeleteSelectRolesRequest
	req.RoleIdList = c.QueryArray("role_id_list")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.DeleteSelectRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectRoles, err)
		return
	}

	loggerx.SuccessLog(c, ActionDeleteSelectRoles, fmt.Sprintf("Roles[%s] delete Success", req.GetRoleIdList()))

	loggerx.InfoLog(c, ActionDeleteSelectRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionDeleteSelectRoles)),
		Data:    response,
	})
}

// HardDeleteRoles 物理删除选中角色
// @Router /phydel/roles [delete]
func (u *Role) HardDeleteRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteRoles, loggerx.MsgProcessStarted)
	roleService := role.NewRoleService("manage", client.DefaultClient)
	var req role.HardDeleteRolesRequest
	req.RoleIdList = c.QueryArray("role_id_list")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.HardDeleteRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteRoles, err)
		return
	}

	loggerx.SuccessLog(c, ActionHardDeleteRoles, fmt.Sprintf("Roles[%s] physically delete Success", req.GetRoleIdList()))

	loggerx.InfoLog(c, ActionHardDeleteRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionHardDeleteRoles)),
		Data:    response,
	})
}

// RecoverSelectRoles 恢复选中角色
// @Router /recover/roles [PUT]
func (u *Role) RecoverSelectRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionRecoverSelectRoles, loggerx.MsgProcessStarted)
	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.RecoverSelectRolesRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectRoles, err)
		return
	}

	db := sessionx.GetUserCustomer(c)
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = db

	response, err := roleService.RecoverSelectRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectRoles, err)
		return
	}

	loggerx.SuccessLog(c, ActionRecoverSelectRoles, fmt.Sprintf("Roles[%s] recover Success", req.GetRoleIdList()))

	loggerx.InfoLog(c, ActionRecoverSelectRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I013, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionRecoverSelectRoles)),
		Data:    response,
	})
}
