package webui

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
)

// Role 角色
type Role struct{}

// log出力
const (
	RoleProcessName       = "Role"
	ActionFindRoles       = "FindRoles"
	ActionFindRole        = "FindRole"
	ActionFindUserActions = "FindUserActions"
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

// FindRole 获取用户操作
// @Router /user/roles/{role_id}/actions [get]
func (u *Role) FindUserActions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUserActions, loggerx.MsgProcessStarted)

	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	roles := sessionx.GetUserRoles(c)

	var req permission.FindActionsRequest
	req.RoleId = roles
	req.ActionType = "datastore"
	req.PermissionType = "app"
	req.AppId = sessionx.GetCurrentApp(c)

	req.Database = sessionx.GetUserCustomer(c)
	response, err := pmService.FindActions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUserActions, err)
		return
	}

	loggerx.InfoLog(c, ActionFindUserActions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionFindUserActions)),
		Data:    response.GetActions(),
	})
}
