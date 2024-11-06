package webui

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

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
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// User 用户
type User struct{}

// log出力
const (
	userProcessName        = "User"
	ActionFindRelatedUsers = "FindRelatedUsers"
	ActionFindUsers        = "FindUsers"
	ActionFindUser         = "FindUser"
	ActionCheckAction      = "CheckAction"
	ActionModifyUser       = "ModifyUser"
	WEBUI_URL              = "WEBUI_URL"
)

// CheckAction 判断用户的按钮权限
// @Router /check/actions/:key [get]
func (u *User) CheckAction(c *gin.Context) {
	loggerx.InfoLog(c, ActionCheckAction, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)
	var req user.FindUserRequest
	req.UserId = sessionx.GetAuthUserID(c)

	req.Database = sessionx.GetUserCustomer(c)

	// 获取用户情报
	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindUser:%s", loggerx.MsgProcessStarted))
	res, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionCheckAction, err)
		return
	}
	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindUser:%s", loggerx.MsgProcessEnded))

	datastoreID := c.Query("datastore_id")
	actionKey := c.Param("key")
	// 默认没有权限
	hasAccess := false

	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindRole:%s", loggerx.MsgProcessStarted))
	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	var preq permission.FindActionsRequest
	preq.RoleId = res.GetUser().GetRoles()
	preq.PermissionType = "app"
	preq.AppId = sessionx.GetCurrentApp(c)
	preq.ActionType = "datastore"
	preq.ObjectId = datastoreID
	preq.Database = sessionx.GetUserCustomer(c)
	pResp, err := pmService.FindActions(context.TODO(), &preq)
	if err != nil {
		httpx.GinHTTPError(c, ActionCheckAction, err)
		return
	}
	loggerx.InfoLog(c, ActionCheckAction, fmt.Sprintf("Process FindRole:%s", loggerx.MsgProcessEnded))

	for _, act := range pResp.GetActions() {
		if act.ObjectId == datastoreID {
			if val, exist := act.ActionMap[actionKey]; exist {
				hasAccess = val
				break
			}
		}
	}

	loggerx.InfoLog(c, ActionCheckAction, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionCheckAction)),
		Data:    hasAccess,
	})
}

// FindRelatedUsers 查找用户组&关联用户组的多个用户记录
// @Router /users/related [get]
func (u *User) FindRelatedUsers(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindRelatedUsers, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindRelatedUsersRequest

	req.GroupIDs = sessionx.GetRelatedGroups(c, c.Query("group"), sessionx.GetUserDomain(c))
	req.InvalidatedIn = c.Query("invalidated_in")
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := userService.FindRelatedUsers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindRelatedUsers, err)
		return
	}

	loggerx.InfoLog(c, ActionFindRelatedUsers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindRelatedUsers)),
		Data:    response.GetUsers(),
	})
}

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
		httpx.GinHTTPError(c, ActionFindUser, err)
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
			httpx.GinHTTPError(c, ActionFindUser, err)
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

// ModifySelf 更新用户
// @Router /users/{user_id} [put]
func (u *User) ModifySelf(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyUser, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	param := c.Param("user_id")
	owner := sessionx.GetAuthUserID(c)

	if param != owner {
		c.JSON(403, gin.H{
			"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
		})
		c.Abort()
		return
	}

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

	notSend := true
	// 个人通知邮箱有变更则发送邮箱激活邮件,更新邮箱状态为确认中
	if userInfo.GetNoticeEmail() != req.GetNoticeEmail() && req.GetNoticeEmail() != "" {
		store := storex.NewRedisStore(60)
		key := req.UserId
		val := store.Get(key, true)
		if len(val) == 0 {
			store.Set(key, "Send Mail Interval")
			notSend = false
			req.NoticeEmailStatus = "Verifying"
			// 定义收件人
			mailTo := []string{
				req.NoticeEmail,
			}
			// 定义抄送人
			mailCcTo := []string{}
			// 邮件主题
			subject := "Notification mailbox activation"
			// 邮件正文
			origin := originx.GetOrigin(false)

			tpl := template.Must(template.ParseFiles("assets/html/mail.html"))
			params := map[string]string{
				"url": origin + "/mail_activate/" + userInfo.Email + "?email=" + req.NoticeEmail,
			}

			var out bytes.Buffer
			err := tpl.Execute(&out, params)
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

	if req.Password != "" {
		req.Password = cryptox.GenerateMd5Password(req.Password, req.Email)
		req.Email = ""
	}

	response, err := userService.ModifyUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyUser, err)
		return
	}

	// 如果传入的用户名为空，日志里采用之前的用户名
	if req.GetUserName() == "" {
		req.UserName = userInfo.GetUserName()
	}
	// 变更成功后，比较变更的结果，记录日志
	// 比较个人邮箱地址
	noticeEmail := userInfo.GetNoticeEmail()
	if noticeEmail != req.GetNoticeEmail() && req.GetNoticeEmail() != "" {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = req.GetUserName()
		params["personal_email_address"] = req.GetNoticeEmail()
		loggerx.ProcessLog(c, ActionModifyUser, msg.L040, params)
	}
	// 比较登录ID
	loginID := userInfo.GetEmail()
	if loginID != req.GetEmail() && req.GetEmail() != "" {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = req.GetUserName()
		params["login_id"] = req.GetEmail()
		loggerx.ProcessLog(c, ActionModifyUser, msg.L041, params)
	}

	loggerx.SuccessLog(c, ActionModifyUser, fmt.Sprintf("User[%s] Update Success", req.GetUserId()))

	// 发送邮件验证用户通知邮箱
	if req.NoticeEmailStatus == "Verifying" && notSend {
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
			origin := originx.GetOrigin(false)

			tpl := template.Must(template.ParseFiles("assets/html/mail.html"))
			params := map[string]string{
				"url": origin + "/mail_activate/" + userInfo.Email + "?email=" + req.NoticeEmail,
			}

			var out bytes.Buffer
			err := tpl.Execute(&out, params)
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
