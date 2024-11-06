package userx

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// 获取所有用户数据
func GetAllUser(db, app, domain string) (users []*user.User) {
	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest
	// 从query中获取参数
	req.App = app
	req.InvalidatedIn = "true"
	// 从共通中获取参数
	req.Domain = domain
	req.Database = db

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getAllUser", err.Error())
		return users
	}

	return response.GetUsers()
}

// 转换用户ID变更为用户名称
func TranUser(userID string, users []*user.User) string {
	for _, user := range users {
		if user.UserId == userID {
			return user.UserName
		}
	}

	return ""
}

// 获取系统管理端，能回答问题的用户
func GetQuestionUsers(c *gin.Context) map[string]string {
	// 获取系统管理端回答问题的角色id
	var roles []string
	var questionUsers = make(map[string]string)

	roleService := role.NewRoleService("manage", client.DefaultClient)
	var reqRole role.FindRolesRequest
	reqRole.Database = "system"
	reqRole.Domain = "proship.co.jp"

	resRole, err := roleService.FindRoles(context.TODO(), &reqRole)
	if err != nil {
		loggerx.ErrorLog("GetQuestionUsers", err.Error())
		return nil
	}
	for _, role := range resRole.GetRoles() {
		if role.RoleName == "SYSTEM" {
			roles = append(roles, role.RoleId)
			continue
		}
		for _, menu := range role.GetMenus() {
			if menu == "/customer/question/list" {
				roles = append(roles, role.RoleId)
				break
			}
		}
	}
	// 获取系统管理端回答问题的用户id
	userService := user.NewUserService("manage", client.DefaultClient)

	var reqUser user.FindUsersRequest
	reqUser.Database = "system"
	for _, r_id := range roles {
		reqUser.Role = r_id
		resUser, err := userService.FindUsers(context.TODO(), &reqUser)
		if err != nil {
			loggerx.ErrorLog("GetQuestionUsers", err.Error())
			return nil
		}
		for _, user := range resUser.GetUsers() {
			questionUsers[user.UserId] = user.UserName
		}
	}
	return questionUsers
}
