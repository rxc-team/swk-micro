package exist

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// log出力
const (
	defaultDomain    = "proship.co.jp"
	defaultDomainEnv = "DEFAULT_DOMAIN"

	ActionCheckAppExpired = "CheckAppExpired"
	ActioncheckAppExist   = "checkAppExist"
)

//CheckExist 存在检查中间件
func CheckExist() gin.HandlerFunc {
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
			exist := checkExist(u)

			if !exist {
				c.JSON(200, httpx.Response{
					Status:  1,
					Message: "app-not-exist",
				})
				c.Abort()
				return
			}
		}

		currentApp := u.GetCurrentApp()

		// 检查当前APP是否过期
		if currentApp != "system" {
			expired := CheckAppExpired(db, currentApp)
			if expired {
				c.JSON(200, httpx.Response{
					Status:  1,
					Message: "app-expired",
				})

				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// CheckAppExpired 检查当前APP是否过期
func CheckAppExpired(db, appID string) (expired bool) {
	appService := app.NewAppService("manage", client.DefaultClient)

	var req app.FindAppRequest
	req.AppId = appID
	req.Database = db
	response, err := appService.FindApp(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog(ActionCheckAppExpired, err.Error())
		return true
	}

	// 以天计
	endDate, err := time.Parse("2006-01-02", response.GetApp().GetEndTime())
	if err != nil {
		return true
	}
	nowDate, err := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if err != nil {
		return true
	}

	endTime := endDate.Unix()
	nowTime := nowDate.Unix()
	// 如果当前时间大于结束时间，则已经超过使用期限，停止用户使用
	return nowTime > endTime
}

// checkExist 判断当前用户的app中是否有有效的app，无效的情况下，返回false
func checkExist(userInfo *user.User) (exist bool) {

	// 判断用户是否已经被无效化
	if userInfo.GetDeletedBy() != "" {
		return false
	}

	domain := os.Getenv(defaultDomainEnv)
	if len(domain) == 0 {
		domain = defaultDomain
	}

	if userInfo.GetDomain() == domain {
		return true
	}

	var appReq app.FindAppsByIdsRequest
	appReq.Domain = userInfo.GetDomain()
	appReq.AppIdList = userInfo.GetApps()
	appReq.Database = userInfo.GetCustomerId()

	appService := app.NewAppService("manage", client.DefaultClient)
	res, err := appService.FindAppsByIds(context.TODO(), &appReq)
	if err != nil {
		loggerx.ErrorLog(ActioncheckAppExist, err.Error())
		return false
	}

	result := false

	for _, app := range res.GetApps() {
		if app.DeletedBy == "" {
			result = true
		}
	}

	return result
}
