/*
 * @Description:API路由
 * @Author: RXC 廖云江
 * @Date: 2019-08-19 10:23:27
 * @LastEditors: Rxc 陳平
 * @LastEditTime: 2021-02-23 13:49:13
 */

package router

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"rxcsoft.cn/pit3/api/internal/handler/common"
	"rxcsoft.cn/pit3/api/internal/middleware/auth/jwt"
	"rxcsoft.cn/utils/config"
)

var secret = []byte("sz36447sbjr4hnt47eah8gwahw2xm2d7")

// InitRouter 初始化路由
func InitRouter(router *gin.Engine) error {

	rf := config.GetConf("redis")

	// session 配置
	store, err := redis.NewStoreWithDB(10, "tcp", fmt.Sprintf("%s:%s", rf.Host, rf.Port), rf.Password, "10", secret)
	if err != nil {
		panic(err)
	}

	store.Options(sessions.Options{
		MaxAge: 86400,
		Path:   "/",
	})

	router.Use(sessions.Sessions("internal", store))

	// 初始化无验证的路由
	initUnAuthRouter(router)
	// 初始化需要验证的路由
	initAuthRouter(router)
	initAuthRouterWeb(router)
	initAuthRouterAdm(router)
	initAuthRouterDev(router)
	// 初始化ws路由
	initWsRouter(router)

	return nil
}

// 初始化无验证的路由
func initUnAuthRouter(router *gin.Engine) {

	// 登陆
	auth := new(common.Auth)

	router.POST("/internal/api/v1/refresh/token", auth.RefreshToken)
	// 密码操作
	password := new(common.Password)

	// 验证令牌
	router.POST("/internal/api/v1/validation/token", password.TokenValidation)
	// 邮箱
	mail := new(common.Mail)
	router.PATCH("internal/api/v1/active/mail", mail.ActiveMail)
	// 验证码
	captcha := new(common.Captcha)
	router.GET("/internal/api/v1/captcha", captcha.CreatCaptcha)
	router.POST("/internal/api/v1/captcha", captcha.VerifyCaptcha)
	// 二次验证
	router.GET("/internal/api/v1/second/captcha", captcha.CreatSecondCaptcha)
	router.POST("/internal/api/v1/second/captcha", captcha.VerifySecondCaptcha)

	router.Use(csrf.Middleware(csrf.Options{
		Secret: string(secret),
		ErrorFunc: func(c *gin.Context) {
			// c.String(500, "CSRF token invalid")
			// c.Abort()
		},
	}))
	//ping使用
	router.GET("/internal/api/v1/ping", func(c *gin.Context) {
		c.Header("X-CSRF-Token", csrf.GetToken(c))
		c.JSON(200, "ok")
	})
	router.POST("/internal/api/v1/login", auth.Login)
	router.POST("/internal/api/v1/password/reset", password.PasswordReset)
	router.POST("/internal/api/v1/password/reset/selected", password.SelectedPasswordReset)
	router.POST("/internal/api/v1/admin/password/reset", password.ResetAdminPassword)
	router.POST("/internal/api/v1/new/password", password.SetNewPassword)
}

func initWsRouter(router *gin.Engine) {
	// ws
	ws := new(common.WebSocket)
	router.POST("/internal/api/v1/send", ws.Send)

	// 创建组
	v1 := router.Group("/internal/api/v1")
	// 使用jwt校验
	v1.Use(jwt.WsJWTAuth())
	{
		v1.GET("/ws/:user_id", ws.Ws)
	}
}

// 初始化需要验证的路由
func initAuthRouter(router *gin.Engine) {

}
