package userx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
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
