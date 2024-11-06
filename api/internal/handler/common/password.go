package common

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/mailx"
	"rxcsoft.cn/pit3/api/internal/common/originx"
	"rxcsoft.cn/pit3/api/internal/common/storex"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// Password 认证
type Password struct{}

// log出力
const (
	PasswordProcessName         = "Password"
	ActionPasswordReset         = "PasswordReset"
	ActionSelectedPasswordReset = "SelectedPasswordReset"
	ActionSetNewPassword        = "SetNewPassword"
	ActionTokenValidation       = "TokenValidation"
)

// PasswordReset 普通用户密码重置
// @Router /password/reset [POST]
func (a *Password) PasswordReset(c *gin.Context) {
	loggerx.InfoLog(c, ActionPasswordReset, loggerx.MsgProcessStarted)

	type UserParams struct {
		LoginID     string `json:"login_id"`
		NoticeEmail string `json:"notice_email"`
	}

	var params UserParams
	// 从body中获取参数
	if err := c.BindJSON(&params); err != nil {
		httpx.GinHTTPError(c, ActionPasswordReset, err)
		return
	}

	userService := user.NewUserService("manage", client.DefaultClient)

	// 参数-用户ID
	loginID := params.LoginID
	// 参数-用户通知邮箱
	noticeEmail := params.NoticeEmail
	// 根据用户ID查询用户信息
	var freq user.EmailRequest
	freq.Email = loginID
	fresponse, err := userService.FindUserByEmail(context.TODO(), &freq)
	// 查询用户信息失败,返回nofound
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
			Data: gin.H{
				"code": "nofound",
			},
		})
		return
	}

	// 用户情报变量定义编辑
	userInfo := fresponse.GetUser()
	db := userInfo.GetCustomerId()

	// 用户通知邮箱未验证,返回unverified
	if userInfo.GetNoticeEmailStatus() != "Verified" {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
			Data: gin.H{
				"code": "unverified",
			},
		})
		return
	}

	// 用户通知邮箱与输入参数通知邮箱不一致,返回nomatch
	if userInfo.GetNoticeEmail() != noticeEmail {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
			Data: gin.H{
				"code": "nomatch",
			},
		})
		return
	}

	// 使用用户ID,生成临时令牌发送给用户通知邮箱
	token, err := cryptox.Password(userInfo.UserId)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
			Data: gin.H{
				"code": "createTokenErr",
			},
		})
		return
	}

	// 临时令牌暂存redis
	store := storex.NewRedisStore(86400)
	// 存入用户邮箱
	store.Set(token, userInfo.Email)

	// 发送临时令牌给用户通知邮箱
	if len(noticeEmail) > 0 {
		// 发送密码重置邮件
		// 定义收件人
		mailTo := []string{
			noticeEmail,
		}
		// 定义抄送人
		mailCcTo := []string{}
		// 邮件主题
		subject := "Please set user login password"
		// 邮件正文
		origin := originx.GetOrigin(false)
		linkUrl := origin + "/password_reset/" + url.QueryEscape(token)
		tpl := template.Must(template.ParseFiles("assets/html/token.html"))
		params := map[string]string{
			"url": linkUrl,
		}

		var out bytes.Buffer
		err := tpl.Execute(&out, params)
		if err != nil {
			httpx.GinHTTPError(c, ActionPasswordReset, err)
			return
		}

		er := mailx.SendMail(db, mailTo, mailCcTo, subject, out.String())
		if er != nil {
			httpx.GinHTTPError(c, ActionPasswordReset, err)
			return
		}
		loggerx.SuccessLog(c, ActionPasswordReset, fmt.Sprintf("User[%s] Update(SendMail) Success", userInfo.GetUserId()))

		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
			Data:    gin.H{},
		})
		return
	}

	loggerx.InfoLog(c, ActionPasswordReset, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
		Data:    gin.H{},
	})
}

// SelectedPasswordReset 重置选中的普通用户密码
// @Router /password/reset//selected [POST]
func (a *Password) SelectedPasswordReset(c *gin.Context) {
	loggerx.InfoLog(c, ActionPasswordReset, loggerx.MsgProcessStarted)

	type UserParams struct {
		LoginID     string `json:"login_id"`
		NoticeEmail string `json:"notice_email"`
	}

	var params []UserParams
	// 从body中获取参数
	if err := c.BindJSON(&params); err != nil {
		httpx.GinHTTPError(c, ActionSelectedPasswordReset, err)
		return
	}

	userService := user.NewUserService("manage", client.DefaultClient)

	// 循环重置选中的用户的密码
	for _, userParam := range params {
		// 参数-用户ID
		loginID := userParam.LoginID
		// 参数-用户通知邮箱
		noticeEmail := userParam.NoticeEmail
		// 根据用户ID查询用户信息
		var freq user.EmailRequest
		freq.Email = loginID
		fresponse, err := userService.FindUserByEmail(context.TODO(), &freq)
		// 查询用户信息失败,返回nofound
		if err != nil {
			httpx.GinHTTPError(c, ActionSelectedPasswordReset, err)
			loggerx.ErrorLog(ActionSelectedPasswordReset, fmt.Sprintf("Email[%s] Password Update(SendMail) Fail", loginID))
			return
		}

		// 用户情报变量定义编辑
		userInfo := fresponse.GetUser()
		db := userInfo.GetCustomerId()

		// 使用用户ID,生成临时令牌发送给用户通知邮箱
		token, err := cryptox.Password(userInfo.UserId)
		if err != nil {
			httpx.GinHTTPError(c, ActionSelectedPasswordReset, err)
			loggerx.ErrorLog(ActionSelectedPasswordReset, fmt.Sprintf("Email[%s] Password Update(SendMail) Fail", loginID))
			return
		}

		// 临时令牌暂存redis
		store := storex.NewRedisStore(86400)
		// 存入用户邮箱
		store.Set(token, userInfo.Email)

		// 发送临时令牌给用户通知邮箱
		if len(noticeEmail) > 0 {
			// 发送密码重置邮件
			// 定义收件人
			mailTo := []string{
				noticeEmail,
			}
			// 定义抄送人
			mailCcTo := []string{}
			// 邮件主题
			subject := "Please set user login password"
			// 邮件正文
			origin := originx.GetOrigin(false)
			linkUrl := origin + "/password_reset/" + url.QueryEscape(token)
			tpl := template.Must(template.ParseFiles("assets/html/token.html"))
			params := map[string]string{
				"url": linkUrl,
			}

			var out bytes.Buffer
			err := tpl.Execute(&out, params)
			if err != nil {
				httpx.GinHTTPError(c, ActionSelectedPasswordReset, err)
				return
			}

			er := mailx.SendMail(db, mailTo, mailCcTo, subject, out.String())
			if er != nil {
				httpx.GinHTTPError(c, ActionSelectedPasswordReset, err)
				return
			}
			loggerx.SuccessLog(c, ActionSelectedPasswordReset, fmt.Sprintf("Email[%s] Update(SendMail) Success", userInfo.Email))
		}
	}

	loggerx.InfoLog(c, ActionSelectedPasswordReset, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionSelectedPasswordReset)),
		Data:    gin.H{},
	})
}

// SetNewPassword 设定新密码
// @Router /new/password [post]
func (a *Password) SetNewPassword(c *gin.Context) {
	loggerx.InfoLog(c, ActionSetNewPassword, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	type SetNewPasswordParams struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	var params SetNewPasswordParams
	// 从body中获取参数
	if err := c.BindJSON(&params); err != nil {
		httpx.GinHTTPError(c, ActionSetNewPassword, err)
		return
	}
	token := params.Token
	// 判断token是否为空
	if token == "" {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionSetNewPassword)),
			Data: gin.H{
				"code": "nullTokenErr",
			},
		})
		return
	}
	// 判断token是否过期
	store := storex.NewRedisStore(86400)
	// 通过token取出用户ID
	email := store.Get(token, true)
	if email == "" {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionSetNewPassword)),
			Data: gin.H{
				"code": "expiredTokenErr",
			},
		})
		return
	}

	// 获取用户信息
	var freq user.EmailRequest
	freq.Email = email
	fresponse, err := userService.FindUserByEmail(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionActiveMail, err)
		return
	}
	userInfo := fresponse.GetUser()

	// 验证当前用户是否和存入的token一致
	if !cryptox.CheckPassword(userInfo.UserId, token) {
		// token不一致的情况
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionSetNewPassword)),
			Data: gin.H{
				"code": "parseTokenErr",
			},
		})
		return
	}

	// 更新用户密码
	var req user.ModifyUserRequest
	req.UserId = userInfo.GetUserId()
	req.Writer = userInfo.GetUserId()
	req.Database = userInfo.GetCustomerId()
	// 通知邮箱状态更正为已验证
	req.NoticeEmailStatus = "Verified"
	// 新密码编辑
	req.Password = cryptox.GenerateMd5Password(params.NewPassword, email)

	response, err := userService.ModifyUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionSetNewPassword, err)
		return
	}
	loggerx.SuccessLog(c, ActionSetNewPassword, fmt.Sprintf("User[%s] Update Success", req.GetUserId()))

	loggerx.InfoLog(c, ActionSetNewPassword, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionSetNewPassword)),
		Data:    response,
	})
}

// TokenValidation 验证令牌
// @Router /validation/token [post]
func (a *Password) TokenValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionTokenValidation, fmt.Sprintf("Process Login:%s", loggerx.MsgProcessStarted))

	type TokenParams struct {
		Token string `json:"token"`
	}

	var params TokenParams
	// 从body中获取参数
	if err := c.BindJSON(&params); err != nil {
		httpx.GinHTTPError(c, ActionTokenValidation, err)
		return
	}
	token := params.Token

	// 判断token是否为空
	if token == "" {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionTokenValidation)),
			Data: gin.H{
				"code": "nullTokenErr",
			},
		})
		return
	}

	// 判断token是否过期
	store := storex.NewRedisStore(86400)
	// 通过token取出用户ID
	email := store.Get(token, false)
	if email == "" {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionTokenValidation)),
			Data: gin.H{
				"code": "expiredTokenErr",
			},
		})
		return
	}

	loggerx.InfoLog(c, ActionTokenValidation, fmt.Sprintf("Process Login:%s", loggerx.MsgProcessEnded))
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionTokenValidation)),
		Data: gin.H{
			"code": "",
		},
	})
}

// ResetAdminPassword 管理员密码重置
// @Router /admin/password/reset [POST]
func (a *Password) ResetAdminPassword(c *gin.Context) {
	loggerx.InfoLog(c, ActionPasswordReset, loggerx.MsgProcessStarted)

	type ResetAdminParams struct {
		CustomerID string `json:"customer_id"`
	}

	userService := user.NewUserService("manage", client.DefaultClient)

	var params ResetAdminParams
	// 从body中获取参数
	if err := c.BindJSON(&params); err != nil {
		httpx.GinHTTPError(c, ActionPasswordReset, err)
		return
	}

	// 查询admin用户信息
	db := params.CustomerID
	uReq := user.FindDefaultUserRequest{
		UserType: 1,
		Database: db,
	}
	adminUser, _ := userService.FindDefaultUser(context.TODO(), &uReq)
	// 用户情报变量定义编辑
	userInfo := adminUser.GetUser()
	email := userInfo.NoticeEmail

	// 使用用户ID,生成临时令牌发送给用户通知邮箱
	token, err := cryptox.Password(userInfo.UserId)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
			Data: gin.H{
				"code": "createTokenErr",
			},
		})
		return
	}

	// 临时令牌暂存redis
	store := storex.NewRedisStore(86400)
	// 存入用户邮箱
	store.Set(token, userInfo.Email)

	// 发送临时令牌给用户通知邮箱
	if len(email) > 0 {
		// 发送密码重置邮件
		// 定义收件人
		mailTo := []string{
			email,
		}
		// 定义抄送人
		mailCcTo := []string{}
		// 邮件主题
		subject := "Please set user login password"
		// 邮件正文
		origin := originx.GetOrigin(true)
		linkUrl := origin + "/password_reset/" + url.QueryEscape(token)
		tpl := template.Must(template.ParseFiles("assets/html/token.html"))
		params := map[string]string{
			"url": linkUrl,
		}

		var out bytes.Buffer
		err := tpl.Execute(&out, params)
		if err != nil {
			httpx.GinHTTPError(c, ActionPasswordReset, err)
			return
		}

		er := mailx.SendMail(db, mailTo, mailCcTo, subject, out.String())
		if er != nil {
			httpx.GinHTTPError(c, ActionPasswordReset, err)
			return
		}
		loggerx.SuccessLog(c, ActionPasswordReset, fmt.Sprintf("User[%s] Update(SendMail) Success", userInfo.GetUserId()))

		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
			Data:    gin.H{},
		})
		return
	}

	loggerx.InfoLog(c, ActionPasswordReset, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionPasswordReset)),
		Data:    gin.H{},
	})
}
