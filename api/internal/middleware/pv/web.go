package pv

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// WebPV 身份类型拦截器
func WebPV() gin.HandlerFunc {
	return func(c *gin.Context) {
		//根据上下文获取载荷userInfo
		userInfo, exit := c.Get("userInfo")
		if !exit {
			c.JSON(401, gin.H{
				"message": msg.GetMsg("ja-JP", msg.Error, msg.E005),
			})
			c.Abort()
			return
		}
		u, exist := userInfo.(*user.User)
		if !exist {
			c.JSON(401, gin.H{
				"message": msg.GetMsg("ja-JP", msg.Error, msg.E005),
			})
			c.Abort()
			return
		}

		userType, err := getUserType(u)
		if err != nil {
			c.JSON(401, gin.H{
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		if userType > 1 {
			c.JSON(403, gin.H{
				"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getUserType(userInfo *user.User) (int, error) {
	roles := userInfo.GetRoles()
	db := userInfo.GetCustomerId()
	// 判断用户角色类型
	userFlg := 0
	for _, g := range roles {
		roleService := role.NewRoleService("manage", client.DefaultClient)

		var req role.FindRoleRequest
		req.RoleId = g
		req.Database = db
		response, err := roleService.FindRole(context.TODO(), &req)
		if err != nil {
			return 0, err
		}

		// 判断角色类型
		if response.Role.RoleType == 3 {
			// dev管理员
			userFlg = 3
		}
		// 判断角色类型
		if response.Role.RoleType == 2 {
			// 超级管理员
			userFlg = 2
		}
		if response.Role.RoleType == 1 {
			// 管理员
			userFlg = 1
		}
	}

	return userFlg, nil
}
