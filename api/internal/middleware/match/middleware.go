package match

import (
	"github.com/gin-gonic/gin"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// log出力
const (
	ActionCheckAppMatch = "CheckAppMatch"
)

// Macth 当前app是否匹配验证
func Macth() gin.HandlerFunc {
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

		db := u.GetCustomerId()
		app := c.GetHeader("App")
		if db != "system" {
			if u.CurrentApp != app {
				c.JSON(200, httpx.Response{
					Status:  1,
					Message: "app-not-match",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
