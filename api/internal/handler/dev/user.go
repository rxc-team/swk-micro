package dev

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
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// User 用户
type User struct{}

// log出力
const (
	userProcessName          = "User"
	ActionFindUsers          = "FindUsers"
	ActionFindUser           = "FindUser"
	ActionFindDefaultUser    = "FindDefaultUser"
	ActionAddUser            = "AddUser"
	ActionModifyUser         = "ModifyUser"
	ActionDeleteSelectUsers  = "DeleteSelectUsers"
	ActionRecoverSelectUsers = "RecoverSelectUsers"
	ActionUnlockSelectUsers  = "UnlockSelectUsers"
	DEV_URL                  = "DEV_URL"
)

// FindUsers 获取所有用户
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
	req.InvalidatedIn = c.Query("invalidated_in")
	req.ErrorCount = c.Query("error_count")
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
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
// @Router /users/{user_id} [get]
func (u *User) FindUser(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUser, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUserRequest
	rType := c.Query("type")
	if len(rType) == 0 || rType == "0" {
		req.UserId = c.Param("user_id")
	} else {
		req.Type = 1
		req.Email = c.Param("user_id")
	}

	db := c.Query("database")
	if len(db) > 0 {
		req.Database = db
	} else {
		req.Database = sessionx.GetUserCustomer(c)
	}

	response, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUsers, err)
		return
	}

	if response.GetUser().GetCustomerId() != "system" {
		// 查找顾客信息，获取logo
		customerService := customer.NewCustomerService("manage", client.DefaultClient)

		loggerx.InfoLog(c, ActionFindUser, fmt.Sprintf("Process FindCustomer:%s", loggerx.MsgProcessStarted))
		var cReq customer.FindCustomerRequest
		cReq.CustomerId = response.GetUser().GetCustomerId()
		res, err := customerService.FindCustomer(context.TODO(), &cReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindCustomer, err)
			return
		}
		loggerx.InfoLog(c, ActionFindUser, fmt.Sprintf("Process FindCustomer:%s", loggerx.MsgProcessEnded))

		type User struct {
			user.User
			Logo         string `json:"logo"`
			CustomerName string `json:"customer_name"`
		}

		userInfo := &User{
			*response.User,
			res.GetCustomer().GetCustomerLogo(),
			res.GetCustomer().GetCustomerName(),
		}

		loggerx.InfoLog(c, ActionFindUser, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindUser)),
			Data:    userInfo,
		})
		return
	}
	loggerx.InfoLog(c, ActionFindUser, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindUser)),
		Data:    response.GetUser(),
	})
}

// FindDefaultUser 获取公司默认管理员用户
// @Router /default/user [get]
func (u *User) FindDefaultUser(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDefaultUser, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindDefaultUserRequest
	if c.Query("type") == "2" {
		req.UserType = 2
	} else if c.Query("type") == "1" {
		req.UserType = 1
	} else {
		req.UserType = 0
	}

	req.Database = c.Query("database")
	if req.Database == "" {
		req.Database = sessionx.GetUserCustomer(c)
	}

	response, err := userService.FindDefaultUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDefaultUser, err)
		return
	}

	loggerx.InfoLog(c, ActionFindDefaultUser, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindDefaultUser)),
		Data:    response.GetUser(),
	})
}

// AddUser 添加用户
// @Router /users [post]
func (u *User) AddUser(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddUser, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.AddUserRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddUser, err)
		return
	}

	password := cryptox.GenerateRandPassword()

	req.Password = cryptox.GenerateMd5Password(password, req.GetEmail())
	// 从共通中获取参数
	req.UserType = 3
	req.Group = "root"
	req.CurrentApp = "system"
	req.Apps = []string{"system"}
	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.CustomerId = sessionx.GetUserCustomer(c)
	req.Database = sessionx.GetUserCustomer(c)
	req.NoticeEmailStatus = "Verifying"
	noticeEmail := req.GetNoticeEmail()
	response, err := userService.AddUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddUser, err)
		return
	}

	// 使用用户ID,生成临时令牌发送给用户通知邮箱
	token, err := cryptox.Password(response.GetUserId())
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionAddUser)),
			Data: gin.H{
				"code": "createTokenErr",
			},
		})
		return
	}
	// 临时令牌暂存redis
	store := storex.NewRedisStore(86400)
	store.Set(token, req.GetEmail())

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
		origin := originx.GetOriginDev()
		linkUrl := origin + "/password_reset/" + url.QueryEscape(token)
		tpl := template.Must(template.ParseFiles("assets/html/token.html"))
		params := map[string]string{
			"url": linkUrl,
		}

		var out bytes.Buffer
		err := tpl.Execute(&out, params)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddUser, err)
			return
		}

		er := mailx.SendMail(req.GetCustomerId(), mailTo, mailCcTo, subject, out.String())
		if er != nil {
			httpx.GinHTTPError(c, ActionAddUser, err)
			return
		}
		loggerx.SuccessLog(c, ActionAddUser, fmt.Sprintf("User[%s] Update(SendMail) Success", response.GetUserId()))

		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionAddUser)),
			Data:    gin.H{},
		})
		return
	}

	loggerx.SuccessLog(c, ActionAddUser, fmt.Sprintf("User[%s] create Success", response.GetUserId()))

	loggerx.InfoLog(c, ActionAddUser, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, userProcessName, ActionAddUser)),
		Data:    response,
	})
}

// ModifyUser 更新用户
// @Router /users/{user_id} [put]
func (u *User) ModifyUser(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyUser, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	// 变更前查询用户信息
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

	// 个人通知邮箱有变更则发送邮箱激活邮件,更新邮箱状态为确认中
	if userInfo.GetNoticeEmail() != req.GetNoticeEmail() && req.GetNoticeEmail() != "" {
		store := storex.NewRedisStore(60)
		key := req.UserId
		val := store.Get(key, true)
		if len(val) == 0 {
			store.Set(key, "Send Mail Interval")
			// 定义收件人
			mailTo := []string{
				req.NoticeEmail,
			}
			// 定义抄送人
			mailCcTo := []string{}
			// 邮件主题
			subject := "Notification mailbox activation"
			// 邮件正文
			origin := originx.GetOriginDev()
			tpl := template.Must(template.ParseFiles("assets/html/mail.html"))
			params := map[string]string{
				"url": origin + "/mail_activate/" + req.Email + "?email=" + req.NoticeEmail,
			}

			var out bytes.Buffer
			err = tpl.Execute(&out, params)
			if err != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}

			er := mailx.SendMail(sessionx.GetUserCustomer(c), mailTo, mailCcTo, subject, out.String())
			if er != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}
			loggerx.SuccessLog(c, ActionModifyUser, fmt.Sprintf("User[%s] Update(SendMail) Success", req.GetUserId()))

			req.NoticeEmailStatus = "Verifying"
		} else {
			httpx.GinHTTPError(c, ActionModifyUser, fmt.Errorf("认证邮件发送间隔时间为60s"))
			return
		}
	}

	if req.Password != "" {
		req.Password = cryptox.GenerateMd5Password(req.Password, req.Email)
		req.Email = ""
	}

	response, err := userService.ModifyUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyUser, err)
		return
	}

	loggerx.SuccessLog(c, ActionModifyUser, fmt.Sprintf("User[%s] Update Success", req.GetUserId()))

	// 发送邮件验证用户通知邮箱
	if req.NoticeEmailStatus == "Verifying" {
		store := storex.NewRedisStore(60)
		key := req.UserId
		val := store.Get(key, true)
		if len(val) == 0 {
			store.Set(key, "Send Mail Interval")

			// 定义收件人
			mailTo := []string{
				req.NoticeEmail,
			}
			// 定义抄送人
			mailCcTo := []string{}
			// 邮件主题
			subject := "Notification mailbox activation"
			// 邮件正文
			origin := originx.GetOriginDev()
			tpl := template.Must(template.ParseFiles("assets/html/mail.html"))
			params := map[string]string{
				"url": origin + "/mail_activate/" + req.Email + "?email=" + req.NoticeEmail,
			}

			var out bytes.Buffer
			err = tpl.Execute(&out, params)
			if err != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}

			er := mailx.SendMail(sessionx.GetUserCustomer(c), mailTo, mailCcTo, subject, out.String())
			if er != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}
			loggerx.SuccessLog(c, ActionModifyUser, fmt.Sprintf("User[%s] Update(SendMail) Success", req.GetUserId()))
		} else {
			httpx.GinHTTPError(c, ActionModifyUser, fmt.Errorf("认证邮件发送间隔时间为60s"))
			return
		}
	}

	loggerx.InfoLog(c, ActionModifyUser, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionModifyUser)),
		Data:    response,
	})
}

// DeleteSelectUsers 删除选中用户
// @Router /users [delete]
func (u *User) DeleteSelectUsers(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSelectUsers, loggerx.MsgProcessStarted)

	var req user.DeleteSelectUsersRequest
	req.UserIdList = c.QueryArray("user_id_list")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	userService := user.NewUserService("manage", client.DefaultClient)
	response, err := userService.DeleteSelectUsers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectUsers, err)
		return
	}

	loggerx.SuccessLog(c, ActionDeleteSelectUsers, fmt.Sprintf("SelectUsers[%s] Delete Success", req.GetUserIdList()))

	loggerx.InfoLog(c, ActionDeleteSelectUsers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, userProcessName, ActionDeleteSelectUsers)),
		Data:    response,
	})
}

// RecoverSelectUsers 恢复选中用户
// @Router /recover/users [PUT]
func (u *User) RecoverSelectUsers(c *gin.Context) {
	loggerx.InfoLog(c, ActionRecoverSelectUsers, loggerx.MsgProcessStarted)

	var req user.RecoverSelectUsersRequest

	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectUsers, err)
		return
	}
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	userService := user.NewUserService("manage", client.DefaultClient)
	response, err := userService.RecoverSelectUsers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectUsers, err)
		return
	}

	loggerx.SuccessLog(c, ActionRecoverSelectUsers, fmt.Sprintf("SelectUsers[%s] Recover Success", req.GetUserIdList()))

	loggerx.InfoLog(c, ActionRecoverSelectUsers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I013, fmt.Sprintf(httpx.Temp, userProcessName, ActionRecoverSelectUsers)),
		Data:    response,
	})
}

// UnlockSelectUsers 恢复被锁用户
// @Router /unlock/users [PUT]
func (u *User) UnlockSelectUsers(c *gin.Context) {
	loggerx.InfoLog(c, ActionUnlockSelectUsers, loggerx.MsgProcessStarted)

	var req user.UnlockSelectUsersRequest

	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionUnlockSelectUsers, err)
		return
	}

	owner := sessionx.GetAuthUserID(c)
	for _, u := range req.GetUserIdList() {
		if u == owner {
			c.JSON(403, gin.H{
				"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
			})
			c.Abort()
			return
		}
	}

	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = c.Query("database")
	if req.Database == "" {
		req.Database = sessionx.GetUserCustomer(c)
	}

	userService := user.NewUserService("manage", client.DefaultClient)
	response, err := userService.UnlockSelectUsers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUnlockSelectUsers, err)
		return
	}

	loggerx.SuccessLog(c, ActionUnlockSelectUsers, fmt.Sprintf("SelectUsers[%s] Unlock Success", req.GetUserIdList()))

	loggerx.InfoLog(c, ActionUnlockSelectUsers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionUnlockSelectUsers)),
		Data:    response,
	})
}
