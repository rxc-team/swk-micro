package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/outer/common/cryptox"
	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/common/slicex"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// User 用户
type User struct{}

// log出力
const (
	userProcessName         = "User"
	ActionFindUsers         = "FindUsers"
	ActionFindUser          = "FindUser"
	ActionFindGroupUsers    = "FindGroupUsers"
	ActionAddUser           = "AddUser"
	ActionModifyUser        = "ModifyUser"
	ActionDeleteUser        = "DeleteUser"
	ActionDeleteSelectUsers = "DeleteSelectUsers"
	ActionHardDeleteUsers   = "HardDeleteUsers"
	ActionCheckAction       = "CheckAction"
	defaultPassword         = "123"
)

// FindUsers 获取所有用户
// @Summary 获取所有用户
// @description 调用srv中的user服务，获取所有用户
// @Tags User
// @Accept json
// @Security JWT
// @Produce  json
// @Param user_name query string false "用户名"
// @Param email query string false "用户邮箱"
// @Param group query string false "组"
// @Param app query string false "app"
// @Param role query string false "角色"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /users [get]
func (u *User) FindUsers(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUsers, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest
	// 从query中获取参数
	req.UserName = c.Query("user_name")
	req.Email = c.Query("email")
	req.Group = c.Query("group")
	req.App = c.Query("app")
	req.Role = c.Query("role")
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.InvalidatedIn = c.Query("invalidated_in")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUsers, err)
		return
	}

	loggerx.InfoLog(c, ActionFindUsers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindUsers)),
		Data:    response.GetUsers(),
	})
}

// FindUser 获取用户
// @Summary 获取用户
// @description 调用srv中的user服务，通过用户ID获取用户相关信息
// @Tags User
// @Accept json
// @Security JWT
// @Produce  json
// @Param user_id path string true "用户ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /users/{user_id} [get]
func (u *User) FindUser(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUser, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUserRequest
	req.UserId = c.Param("user_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUser, err)
		return
	}

	loggerx.InfoLog(c, ActionFindUser, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindUser)),
		Data:    response.GetUser(),
	})
}

// ModifyUser 更新用户
// @Summary 更新用户
// @description 调用srv中的user服务，更新用户
// @Tags User
// @Accept json
// @Security JWT
// @Produce  json
// @Param user_id path string true "用户ID"
// @Param body body user.ModifyRequest true "用户属性"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /users/{user_id} [put]
func (u *User) ModifyUser(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyUser, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)
	// 变更前查询用户信息以供日志使用
	var freq user.FindUserRequest
	freq.UserId = c.Param("user_id")
	freq.Database = sessionx.GetUserCustomer(c)
	fresponse, err := userService.FindUser(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyUser, err)
		return
	}
	userInfo := fresponse.GetUser()

	var req user.ModifyUserRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyUser, err)
		return
	}
	// 从path中获取参数
	req.UserId = c.Param("user_id")
	// 当前用户为更新者
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	if req.Password != "" {
		req.Password = cryptox.GenerateMd5Password(req.Password, req.Email)
		req.Email = ""
	}

	response, err := userService.ModifyUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyUser, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyUser, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyUser))

	// 如果传入的用户名为空，日志里采用之前的用户名
	if req.GetUserName() == "" {
		req.UserName = userInfo.GetUserName()
	}
	// 变更成功后，比较变更的结果，记录日志
	// 比较个人邮箱地址
	noticeEmail := userInfo.GetNoticeEmail()
	if noticeEmail != req.GetNoticeEmail() && req.GetNoticeEmail() != "" {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = req.GetUserName()
		params["personal_email_address"] = req.GetNoticeEmail()
		loggerx.ProcessLog(c, ActionModifyUser, msg.L040, params)
	}
	// 比较登录ID
	loginID := userInfo.GetEmail()
	if loginID != req.GetEmail() && req.GetEmail() != "" {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = req.GetUserName()
		params["login_id"] = req.GetEmail()
		loggerx.ProcessLog(c, ActionModifyUser, msg.L041, params)
	}
	// // 比较用户Group
	if userInfo.GetGroup() != req.GetGroup() && len(req.GetGroup()) > 0 {
		// 查找变更的group的数据
		groupService := group.NewGroupService("manage", client.DefaultClient)
		var fReq group.FindGroupRequest
		fReq.GroupId = req.GetGroup()
		fReq.Database = sessionx.GetUserCustomer(c)
		fResponse, err := groupService.FindGroup(context.TODO(), &fReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyUser, err)
			return
		}
		groupInfo := fResponse.GetGroup()
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = req.GetUserName()
		params["group_name"] = "{{" + groupInfo.GetGroupName() + "}}"
		loggerx.ProcessLog(c, ActionModifyUser, msg.L042, params)
	}
	// 比较用户角色
	if !slicex.StringSliceEqual(userInfo.GetRoles(), req.GetRoles()) && len(req.GetRoles()) > 0 {
		for _, key := range req.GetRoles() {
			// 获取变更后的角色的名称
			roleService := role.NewRoleService("manage", client.DefaultClient)

			var freq role.FindRoleRequest
			freq.RoleId = key
			freq.Database = sessionx.GetUserCustomer(c)
			fresponse, err := roleService.FindRole(context.TODO(), &freq)
			if err != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}
			roleInfo := fresponse.GetRole()
			params := make(map[string]string)
			params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
			params["user_name1"] = req.GetUserName()
			params["profile_name"] = roleInfo.GetRoleName()
			loggerx.ProcessLog(c, ActionModifyUser, msg.L043, params)
		}
	}

	// 比较用户app
	if !slicex.StringSliceEqual(userInfo.GetApps(), req.GetApps()) && len(req.GetApps()) > 0 {
		for _, key := range req.GetApps() {
			// 获取变更后app的名称
			appService := app.NewAppService("manage", client.DefaultClient)

			var freq app.FindAppRequest
			freq.AppId = key
			freq.Database = sessionx.GetUserCustomer(c)
			fresponse, err := appService.FindApp(context.TODO(), &freq)
			if err != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}
			appInfo := fresponse.GetApp()

			params := make(map[string]string)
			params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
			params["user_name1"] = req.GetUserName()
			params["app_name"] = "{{" + appInfo.GetAppName() + "}}"
			loggerx.ProcessLog(c, ActionModifyUser, msg.L044, params)
		}
	}

	loggerx.InfoLog(c, ActionModifyUser, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionModifyUser)),
		Data:    response,
	})
}

// CheckAction 判断用户的按钮权限
// @Summary 判断用户的按钮权限
// @description 调用srv中的user服务，判断用户的按钮权限
// @Tags User
// @Accept json
// @Security JWT
// @Produce  json
// @Param key query string false "按钮key""
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /check/actions/:key [get]
func (u *User) CheckAction(c *gin.Context) {
	loggerx.InfoLog(c, ActionCheckAction, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)
	var req user.FindUserRequest
	req.UserId = sessionx.GetAuthUserID(c)

	req.Database = sessionx.GetUserCustomer(c)

	// 获取用户情报
	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindUser:%s", loggerx.MsgProcessStarted))
	res, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionCheckAction, err)
		return
	}
	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindUser:%s", loggerx.MsgProcessEnded))

	datastoreID := c.Query("datastore_id")
	actionKey := c.Param("key")
	// 默认没有权限
	hasAccess := false

	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindRole:%s", loggerx.MsgProcessStarted))
	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	var preq permission.FindActionsRequest
	preq.RoleId = res.GetUser().GetRoles()
	preq.PermissionType = "app"
	preq.AppId = sessionx.GetCurrentApp(c)
	preq.ActionType = "datastore"
	preq.ObjectId = datastoreID
	preq.Database = sessionx.GetUserCustomer(c)
	pResp, err := pmService.FindActions(context.TODO(), &preq)
	if err != nil {
		httpx.GinHTTPError(c, ActionCheckAction, err)
		return
	}
	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindRole:%s", loggerx.MsgProcessEnded))

	for _, act := range pResp.GetActions() {
		if act.ObjectId == datastoreID {
			if val, exist := act.ActionMap[actionKey]; exist {
				hasAccess = val
				break
			}
		}
	}

	loggerx.InfoLog(c, ActionCheckAction, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionCheckAction)),
		Data:    hasAccess,
	})
}
