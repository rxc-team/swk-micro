package pv

import (
	"github.com/gin-gonic/gin"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// DevPV 身份类型拦截器
func DevPV() gin.HandlerFunc {
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

		if !(userType == 2 || userType == 3) {
			c.JSON(403, gin.H{
				"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
