package casbin

import (
	"strings"

	"github.com/gin-gonic/gin"

	"rxcsoft.cn/pit3/api/internal/system/aclx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

func CheckAction() gin.HandlerFunc {
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
		if db != "system" {
			user := u.GetUserId()
			app := u.GetCurrentApp()
			objectId := getOubject(c)
			roleid := u.GetRoles()
			e, hasRole := aclx.GetCasbin_request(app, roleid, user, c.Request.URL.Path, c.Request.Method, objectId)
			e.EnableLog(false)

			if hasRole {
				//hasRole == trueなら、権限が必要ないリクエストのためそのまま通す
				c.Next()
			} else {
				//hasRole == falseなら、権限が必要なリクエスト（パス）のため、e.Enforceで権限確認を行う。
				hasRole, err := e.Enforce(user, db, app, objectId, c.Request.URL.Path, c.Request.Method)
				if err != nil {
					c.JSON(403, gin.H{
						"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
					})
					c.Abort()
					return
				}

				if hasRole {
					c.Next()
				} else {
					c.JSON(403, gin.H{
						"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
					})
					c.Abort()
					return
				}
			}
		}
	}
}

func getOubject(c *gin.Context) string {
	path := c.Request.URL.Path

	if strings.HasPrefix(path, "/internal/api/v1/web/item/") {
		return c.Param("d_id")
	}
	if strings.HasPrefix(path, "/internal/api/v1/web/history/") {
		return c.Param("d_id")
	}
	if strings.HasPrefix(path, "/internal/api/v1/web/mapping/") {
		return c.Param("d_id")
	}
	if strings.HasPrefix(path, "/internal/api/v1/web/report/") {
		return c.Param("rp_id")
	}
	if strings.HasPrefix(path, "/internal/api/v1/web/file/") {
		return c.Param("fo_id")
	}

	return ""

}
