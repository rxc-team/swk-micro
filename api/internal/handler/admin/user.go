package admin

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/mailx"
	"rxcsoft.cn/pit3/api/internal/common/originx"
	"rxcsoft.cn/pit3/api/internal/common/slicex"
	"rxcsoft.cn/pit3/api/internal/common/storex"
	"rxcsoft.cn/pit3/api/internal/system/aclx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/utils/redisx"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// User 用户
type User struct{}

// log出力
const (
	userProcessName          = "User"
	ActionFindRelatedUsers   = "FindRelatedUsers"
	ActionFindUsers          = "FindUsers"
	ActionFindUser           = "FindUser"
	ActionFindDefaultUser    = "FindDefaultUser"
	ActionCheckAction        = "CheckAction"
	ActionFindGroupUsers     = "FindGroupUsers"
	ActionAddUser            = "AddUser"
	ActionModifyUser         = "ModifyUser"
	ActionDeleteUser         = "DeleteUser"
	ActionDeleteSelectUsers  = "DeleteSelectUsers"
	ActionHardDeleteUsers    = "HardDeleteUsers"
	ActionRecoverSelectUsers = "RecoverSelectUsers"
	ActionUnlockSelectUsers  = "UnlockSelectUsers"
	ActionUploadUsers        = "UploadUsers"
	ActionDownloadUsers      = "DownloadUsers"
	WEBUI_URL                = "WEBUI_URL"
)

type BVParam struct {
	Domain          string
	Data            []string
	Header          []string
	LangData        *language.Language
	GroupList       []*group.Group
	AppList         []*app.App
	RoleList        []*role.Role
	DefaultTimezone string
	DefaultLanguage string
	Writer          string
	Database        string
}
type Timezone struct {
	Code  string
	Value string
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

	db := sessionx.GetUserCustomer(c)
	// 判断顾客用户数超限制否
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	cusReq := customer.FindCustomerRequest{
		CustomerId: db,
	}
	custom, err := customerService.FindCustomer(context.TODO(), &cusReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddUser, err)
		return
	}
	if custom.Customer.UsedUsers >= custom.Customer.MaxUsers {
		httpx.GinHTTPError(c, ActionAddUser, errors.New("ユーザー数が組織内の顧客の最大ユーザー制限を超えています"))
		return
	}

	ctx := c.Request.Context()

	rdb := redisx.New()
	// 判断是否有人在上传了
	count, err := rdb.Exists(ctx, db).Result()
	if err != nil {
		httpx.GinHTTPError(c, ActionUploadUsers, err)
		return
	}

	if count > 0 {
		httpx.GinHTTPError(c, ActionAddUser, errors.New("他の誰かがユーザーをアップロードしています。アップロードを続行するには、ユーザーがアップロードするのを待つ必要があります。"))
		return
	}

	// 十分钟后过期
	err = rdb.Set(ctx, db, "1", 10*time.Minute).Err()
	if err != nil {
		httpx.GinHTTPError(c, ActionUploadUsers, err)
		return
	}

	defer func() {
		// 上传结束后，删除改key，保证可以继续上传
		err = rdb.Del(context.TODO(), db).Err()
		if err != nil {
			return
		}
	}()

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

	aclx.SetUserCasbin(response.GetUserId(), req.GetRoles(), req.GetApps())

	// 更新顾客已用用户数
	customerUpReq := customer.ModifyUsedUsersRequest{
		CustomerId: sessionx.GetUserCustomer(c),
		UsedUsers:  1,
	}
	_, err = customerService.ModifyUsedUsers(context.TODO(), &customerUpReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddUser, err)
		return
	}

	// 用户情报验证成功,生成临时令牌发送给用户通知邮箱
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
		origin := originx.GetOrigin(false)
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

		// 添加用户成功后保存日志到DB
		param := make(map[string]string)
		param["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		param["user_name1"] = req.GetUserName()      // 新规的时候取传入参数

		loggerx.ProcessLog(c, ActionAddUser, msg.L035, param)

		loggerx.InfoLog(c, ActionAddUser, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionAddUser)),
			Data:    gin.H{},
		})
		return
	}

	loggerx.SuccessLog(c, ActionAddUser, fmt.Sprintf("User[%s] create Success", response.GetUserId()))

	// 添加用户成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
	params["user_name1"] = req.GetUserName()      // 新规的时候取传入参数

	loggerx.ProcessLog(c, ActionAddUser, msg.L035, params)

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
			origin := originx.GetOrigin(true)
			tpl := template.Must(template.ParseFiles("assets/html/mail.html"))
			params := map[string]string{
				"url": origin + "/mail_activate/" + userInfo.Email + "?email=" + req.NoticeEmail,
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

	if req.Password != "" {
		req.Password = cryptox.GenerateMd5Password(req.Password, req.Email)
		req.Email = ""
	}

	response, err := userService.ModifyUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyUser, err)
		return
	}

	// 更新用户和角色的关系
	// 更改了APP或者角色之后需要更新权限关系
	if len(req.Apps) > 0 || len(req.Roles) > 0 {
		aclx.SetUserCasbin(req.GetUserId(), req.GetRoles(), req.GetApps())
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
	// // 比较用户Group
	if userInfo.GetGroup() != req.GetGroup() && len(req.GetGroup()) > 0 {
		// 查找变更的group的数据
		groupService := group.NewGroupService("manage", client.DefaultClient)
		var fReq group.FindGroupRequest
		fReq.GroupId = req.GetGroup()
		fReq.Database = sessionx.GetUserCustomer(c)
		fResponse, err := groupService.FindGroup(context.TODO(), &fReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyUser, err)
			return
		}
		groupInfo := fResponse.GetGroup()
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = req.GetUserName()
		params["group_name"] = "{{" + groupInfo.GetGroupName() + "}}"
		loggerx.ProcessLog(c, ActionModifyUser, msg.L042, params)
	}
	// 比较用户角色
	if !slicex.StringSliceEqual(userInfo.GetRoles(), req.GetRoles()) && len(req.GetRoles()) > 0 {
		for _, key := range req.GetRoles() {
			// 获取变更后的角色的名称
			roleService := role.NewRoleService("manage", client.DefaultClient)

			var freq role.FindRoleRequest
			freq.RoleId = key
			freq.Database = sessionx.GetUserCustomer(c)
			fresponse, err := roleService.FindRole(context.TODO(), &freq)
			if err != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}
			roleInfo := fresponse.GetRole()
			params := make(map[string]string)
			params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
			params["user_name1"] = req.GetUserName()
			params["profile_name"] = roleInfo.GetRoleName()
			loggerx.ProcessLog(c, ActionModifyUser, msg.L043, params)
		}
	}

	// 比较用户app
	if !slicex.StringSliceEqual(userInfo.GetApps(), req.GetApps()) && len(req.GetApps()) > 0 {
		for _, key := range req.GetApps() {
			// 获取变更后app的名称
			appService := app.NewAppService("manage", client.DefaultClient)

			var freq app.FindAppRequest
			freq.AppId = key
			freq.Database = sessionx.GetUserCustomer(c)
			fresponse, err := appService.FindApp(context.TODO(), &freq)
			if err != nil {
				httpx.GinHTTPError(c, ActionModifyUser, err)
				return
			}
			appInfo := fresponse.GetApp()

			params := make(map[string]string)
			params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
			params["user_name1"] = req.GetUserName()
			params["app_name"] = "{{" + appInfo.GetAppName() + "}}"
			loggerx.ProcessLog(c, ActionModifyUser, msg.L044, params)
		}
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
			origin := originx.GetOrigin(true)
			tpl := template.Must(template.ParseFiles("assets/html/mail.html"))
			params := map[string]string{
				"url": origin + "/mail_activate/" + userInfo.Email + "?email=" + req.NoticeEmail,
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
			httpx.GinHTTPError(c, ActionModifyUser, fmt.Errorf("認証メールの送信間隔は60秒です"))
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
	// 将用户无效化之前查询用户名
	userNameList := make(map[string]string)
	for _, id := range req.GetUserIdList() {
		var freq user.FindUserRequest
		freq.UserId = id
		freq.Database = sessionx.GetUserCustomer(c)
		fresponse, err := userService.FindUser(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteSelectUsers, err)
			return
		}
		userNameList[id] = fresponse.GetUser().GetUserName()
	}
	response, err := userService.DeleteSelectUsers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectUsers, err)
		return
	}

	// 删除该用户的所有角色关系
	for _, userID := range req.UserIdList {
		aclx.ClearUserCasbin(userID)
	}

	loggerx.SuccessLog(c, ActionDeleteSelectUsers, fmt.Sprintf("SelectUsers[%s] Delete Success", req.GetUserIdList()))

	// 更新顾客已用用户数
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	// 更新
	customerUpReq := customer.ModifyUsedUsersRequest{
		CustomerId: sessionx.GetUserCustomer(c),
		UsedUsers:  -int32(len(req.UserIdList)),
	}
	_, err = customerService.ModifyUsedUsers(context.TODO(), &customerUpReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteUsers, err)
		return
	}

	// 将用户无效化成功后保存日志到DB
	for _, id := range req.GetUserIdList() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = userNameList[id]
		loggerx.ProcessLog(c, ActionDeleteSelectUsers, msg.L037, params)
	}
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

	// 判断顾客用户数超限制否
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	cusReq := customer.FindCustomerRequest{
		CustomerId: sessionx.GetUserCustomer(c),
	}
	custom, err := customerService.FindCustomer(context.TODO(), &cusReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddUser, err)
		return
	}
	if custom.Customer.UsedUsers >= custom.Customer.MaxUsers {
		httpx.GinHTTPError(c, ActionAddUser, errors.New("ユーザー数が組織内の顧客の最大ユーザー制限を超えています"))
		return
	}

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

	// 更新顾客已用用户数
	customerUpReq := customer.ModifyUsedUsersRequest{
		CustomerId: sessionx.GetUserCustomer(c),
		UsedUsers:  int32(len(req.GetUserIdList())),
	}
	_, err = customerService.ModifyUsedUsers(context.TODO(), &customerUpReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddUser, err)
		return
	}

	// 恢复用户成功后保存日志到DB
	for _, id := range req.GetUserIdList() {
		var freq user.FindUserRequest
		freq.UserId = id
		freq.Database = sessionx.GetUserCustomer(c)
		fresponse, err := userService.FindUser(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionRecoverSelectUsers, err)
			return
		}

		// 循环添加当前用户和角色以及app之间的关联
		aclx.SetUserCasbin(id, fresponse.GetUser().GetRoles(), fresponse.GetUser().GetApps())

		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = fresponse.GetUser().GetUserName()
		loggerx.ProcessLog(c, ActionRecoverSelectUsers, msg.L038, params)
	}

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

	// 解锁用户成功后保存日志到DB
	for _, id := range req.GetUserIdList() {
		var freq user.FindUserRequest
		freq.UserId = id
		freq.Database = req.GetDatabase()
		fresponse, err := userService.FindUser(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionUnlockSelectUsers, err)
			return
		}
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["user_name1"] = fresponse.GetUser().GetUserName()
		loggerx.ProcessLog(c, ActionUnlockSelectUsers, msg.L039, params)
	}

	loggerx.InfoLog(c, ActionUnlockSelectUsers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionUnlockSelectUsers)),
		Data:    response,
	})
}

// UploadUsers 批量上传用户
// @Router /upload/users [POST]
func (u *User) UploadUsers(c *gin.Context) {
	loggerx.InfoLog(c, ActionUploadUsers, loggerx.MsgProcessStarted)

	appID := sessionx.GetCurrentApp(c)
	jobID := c.PostForm("job_id")
	charEncoding := c.PostForm("encoding")
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)
	userID := sessionx.GetAuthUserID(c)
	db := sessionx.GetUserCustomer(c)

	ctx := c.Request.Context()

	rdb := redisx.New()
	// 判断是否有人在上传了
	count, err := rdb.Exists(ctx, db).Result()
	if err != nil {
		httpx.GinHTTPError(c, ActionUploadUsers, err)
		return
	}

	if count > 0 {
		httpx.GinHTTPError(c, ActionUploadUsers, errors.New("他の誰かがユーザーをアップロードしています。アップロードを続行するには、ユーザーがアップロードするのを待つ必要があります。"))
		return
	}

	// 十分钟后过期
	err = rdb.Set(ctx, db, "1", 10*time.Minute).Err()
	if err != nil {
		httpx.GinHTTPError(c, ActionUploadUsers, err)
		return
	}

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "User import",
		Origin:       "_",
		UserId:       userID,
		ShowProgress: true,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "user-csv-import",
		Steps:        []string{"start", "data-ready", "build-check-data", "upload", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionUploadUsers, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
		path := filex.WriteAndSaveFile(domain, appID, []string{fmt.Sprintf("the csv file type [%v] is not supported", files.Header.Get("content-type"))})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionUploadUsers, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "csv", files.Size) {
		path := filex.WriteAndSaveFile(domain, appID, []string{"the csv file ファイルサイズが設定サイズを超えています"})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionUploadUsers, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	go func() {

		defer func() {
			// 上传结束后，删除改key，保证可以继续上传
			err = rdb.Del(context.TODO(), db).Err()
			if err != nil {
				return
			}
		}()

		// 获取顾客信息
		customerService := customer.NewCustomerService("manage", client.DefaultClient)

		var creq customer.FindCustomerRequest
		creq.CustomerId = db
		cResp, err := customerService.FindCustomer(context.TODO(), &creq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "依存データの取得に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		customerInfo := cResp.GetCustomer()

		// 判断当前顾客的用户数是否已经达到设定值
		if customerInfo.UsedUsers == customerInfo.MaxUsers {
			path := filex.WriteAndSaveFile(domain, appID, []string{"組織の顧客の最大ユーザー数が使い果たされ、これ以上ユーザーを作成できなくなりました"})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ユーザー数の判別エラー",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 读取文件
		fs, err := files.Open()
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイルを開くことができません",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}
		defer fs.Close()

		var cr *csv.Reader
		// UTF-8格式的场合，直接读取
		if charEncoding == "utf-8" {
			cr = csv.NewReader(fs)
		} else {
			// ShiftJIS格式的场合，先转换为uft-8，再读取
			utfReader := transform.NewReader(fs, japanese.ShiftJIS.NewDecoder())
			cr = csv.NewReader(utfReader)
		}
		cr.LazyQuotes = true
		cr.Comma = 44

		var rows [][]string

		for {
			row, err := cr.Read()
			if err == io.EOF {
				break
			}
			// 出现读写错误，直接返回
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 数据验证错误，停止上传
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "ファイルの読み込みに失敗しました",
					CurrentStep: "data-ready",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}

			rows = append(rows, row)
		}

		// 上传数量不能超过最大用户量
		if customerInfo.UsedUsers+int32(len(rows)) > customerInfo.MaxUsers {
			path := filex.WriteAndSaveFile(domain, appID, []string{"アップロードしたユーザー数が組織内の最大ユーザー数を超えています。アップロードする前にアップロード数を減らしてください。"})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ユーザー数の判別エラー",
				CurrentStep: "build-check-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 一次上传数量不能超过1000件
		if int32(len(rows)) > 1001 {
			path := filex.WriteAndSaveFile(domain, appID, []string{"1回のアップロードでアップロードできるのは、最大1000個までです。"})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ユーザー数の判別エラー",
				CurrentStep: "build-check-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 发送消息 数据准备
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "依存データを取得します",
			CurrentStep: "data-ready",
			Database:    db,
		}, userID)

		// 获取所有组织
		groupService := group.NewGroupService("manage", client.DefaultClient)

		var greq group.FindGroupsRequest
		// 当前用户的domain
		greq.Domain = domain
		greq.Database = db

		gResp, err := groupService.FindGroups(context.TODO(), &greq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "依存データの取得に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		groupList := gResp.GetGroups()
		// 获取所有app
		appService := app.NewAppService("manage", client.DefaultClient)

		var areq app.FindAppsRequest
		areq.Domain = domain
		areq.Database = db

		aResp, err := appService.FindApps(context.TODO(), &areq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "依存データの取得に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		appIdList := aResp.GetApps()
		// 获取所有角色
		roleService := role.NewRoleService("manage", client.DefaultClient)

		var rreq role.FindRolesRequest
		// 从共通中获取参数
		rreq.Domain = domain
		rreq.Database = db

		rResp, err := roleService.FindRoles(context.TODO(), &rreq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "依存データの取得に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		rolesList := rResp.GetRoles()

		// 获取语言
		langData := langx.GetLanguageData(db, lang, domain)

		// 获取上传流
		userService := user.NewUserService("manage", client.DefaultClient)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		stream, err := userService.Upload(context.Background(), opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイルアップロードの初期化に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		defer stream.Close()

		// 发送消息 开始读取并验证数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "データの読み取りと検証を開始します",
			CurrentStep: "build-check-data",
			Database:    db,
		}, userID)

		var r *csv.Reader
		// UTF-8格式的场合，直接读取
		if charEncoding == "utf-8" {
			r = csv.NewReader(fs)
		} else {
			// ShiftJIS格式的场合，先转换为uft-8，再读取
			utfReader := transform.NewReader(fs, japanese.ShiftJIS.NewDecoder())
			r = csv.NewReader(utfReader)
		}
		r.LazyQuotes = true

		// 设置逗号分割
		r.Comma = 44

		var header []string
		var items []*user.UploadRequest
		var errorList []string
		// 针对大文件，一行一行的读取文件
		for index, row := range rows {
			if index == 0 {
				header = row
				// 去除utf-8 withbom的前缀
				header[0] = strings.Replace(header[0], "\uFEFF", "", -1)
				continue
			}

			// 验证中有错误，放入全局的验证错误中，等待全部验证完毕后一起返回
			parm := BVParam{
				Domain:          domain,
				Data:            row,
				Header:          header,
				LangData:        langData,
				GroupList:       groupList,
				AppList:         appIdList,
				RoleList:        rolesList,
				DefaultTimezone: customerInfo.DefaultTimezone,
				DefaultLanguage: customerInfo.DefaultLanguage,
				Writer:          userID,
				Database:        db,
			}

			result, err := BuildAndValidate(parm)
			if err != nil {
				errorList = append(errorList, fmt.Sprintf("第%d行目でエラーが発生しました。エラー内容：%s", index, err.Error()))
				continue
			}

			items = append(items, result)
		}

		if len(errorList) > 0 {
			path := filex.WriteAndSaveFile(domain, appID, errorList)

			// 发送消息 出现错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_046"),
				CurrentStep: "build-check-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 发送消息 开始上传数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "データのアップロードを開始します",
			CurrentStep: "upload",
			Database:    db,
		}, userID)

		go func() {
			for _, data := range items {
				err := stream.Send(data)
				if err == io.EOF {
					return
				}

				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 数据验证错误，停止上传
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     "ファイルのアップロード中にエラーが発生しました。",
						CurrentStep: "upload",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)
					return
				}
			}

			err := stream.Send(&user.UploadRequest{
				Status: user.SendStatus_COMPLETE,
			})

			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 数据验证错误，停止上传
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "ファイルのアップロード中にエラーが発生しました。",
					CurrentStep: "upload",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}
		}()

		// 发送消息 等待上传结果
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "アップロード結果を待っています",
			CurrentStep: "upload",
			Database:    db,
		}, userID)

		var inserted int64 = 0

		for {
			result, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 数据验证错误，停止上传
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "ファイルのアップロード中にエラーが発生しました。",
					CurrentStep: "upload",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}

			if result.Status == user.Status_FAILED {
				errorList = append(errorList, "アップロードに失敗")

				// 终止继续发送
				err := stream.Send(&user.UploadRequest{
					Status: user.SendStatus_COMPLETE,
				})

				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 数据验证错误，停止上传
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     "ファイルのアップロード中にエラーが発生しました。",
						CurrentStep: "upload",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)
					return
				}
				break
			}

			if result.Status == user.Status_SUCCESS {

				inserted = result.GetInserted()
				continue
			}
		}

		for _, data := range items {
			aclx.SetUserCasbin(data.GetUserId(), data.GetRoles(), data.GetApps())
		}

		// 更新顾客已用用户数
		customerUpReq := customer.ModifyUsedUsersRequest{
			CustomerId: sessionx.GetUserCustomer(c),
			UsedUsers:  int32(inserted),
		}
		_, err = customerService.ModifyUsedUsers(context.TODO(), &customerUpReq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "顧客ユーザー情報の更新中にエラーが発生しました",
				CurrentStep: "upload",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 表示有部分发生错误
		if len(errorList) > 0 {
			path := filex.WriteAndSaveFile(domain, appID, errorList)

			// 发送消息 出现错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_008"),
				CurrentStep: "end",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 完成全部导入
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_009"),
			CurrentStep: "end",
			Progress:    100,
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)

	}()

	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionUploadUsers)),
		Data:    gin.H{},
	})

}

// DownloadUsers 批量下载用户
// @Router /download/users [POST]
func (u *User) DownloadUsers(c *gin.Context) {
	loggerx.InfoLog(c, ActionUploadUsers, loggerx.MsgProcessStarted)

	appID := sessionx.GetCurrentApp(c)
	jobID := c.Query("job_id")
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)
	userID := sessionx.GetAuthUserID(c)
	db := sessionx.GetUserCustomer(c)

	var req user.DownloadRequest
	// 从query中获取参数
	req.UserName = c.Query("user_name")
	req.Email = c.Query("email")
	req.Group = c.Query("group")
	req.App = c.Query("app")
	req.Role = c.Query("role")
	req.InvalidatedIn = c.Query("invalidated_in")
	req.ErrorCount = c.Query("error_count")
	// 从共通中获取参数
	req.Domain = domain
	req.Database = db

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "Download user",
		Origin:       "_",
		UserId:       userID,
		ShowProgress: true,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "user-csv-download",
		Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	go func() {

		// 发送消息 数据准备
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "依存データを取得します",
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		// 获取所有组织
		groupService := group.NewGroupService("manage", client.DefaultClient)

		var greq group.FindGroupsRequest
		// 当前用户的domain
		greq.Domain = domain
		greq.Database = db

		gResp, err := groupService.FindGroups(context.TODO(), &greq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "依存データの取得に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		groupList := gResp.GetGroups()
		// 获取所有app
		appService := app.NewAppService("manage", client.DefaultClient)

		var areq app.FindAppsRequest
		areq.Domain = domain
		areq.Database = db

		aResp, err := appService.FindApps(context.TODO(), &areq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "依存データの取得に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		appIdList := aResp.GetApps()
		// 获取所有角色
		roleService := role.NewRoleService("manage", client.DefaultClient)

		var rreq role.FindRolesRequest
		// 从共通中获取参数
		rreq.Domain = domain
		rreq.Database = db

		rResp, err := roleService.FindRoles(context.TODO(), &rreq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "依存データの取得に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		rolesList := rResp.GetRoles()

		// 获取语言
		langData := langx.GetLanguageData(db, lang, domain)

		// 获取上传流
		userService := user.NewUserService("manage", client.DefaultClient)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		stream, err := userService.Download(context.Background(), &req, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据查询错误
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイルアップロードの初期化に失敗しました",
				CurrentStep: "data-ready",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		defer stream.Close()

		// 发送消息 开始读写数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "データの読み書きを開始します",
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		timestamp := time.Now().Format("20060102150405")

		// 验证header传入字段内容是否全部正确
		header := []string{"UserID", "UserName", "NoticeEmail", "Password", "Profiles", "Group", "Apps", "Language", "Timezone"}
		// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
		header[0] = "\xEF\xBB\xBF" + header[0]

		headers := append([][]string{}, header)

		filex.Mkdir("temp/")

		// 写入文件到本地
		filename := "temp/tmp" + "_" + timestamp + ".csv"
		f, err := os.Create(filename)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "write-to-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		writer := csv.NewWriter(f)
		writer.WriteAll(headers)

		writer.Flush() // 此时才会将缓冲区数据写入

		var items [][]string

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "write-to-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			var row []string
			// 用户ID
			row = append(row, resp.User.GetEmail())
			// 用户名
			row = append(row, resp.User.GetUserName())
			// 通知邮箱
			row = append(row, resp.User.GetNoticeEmail())
			// 登录密码(默认下载为空)
			row = append(row, "")
			// 角色
			{
				var roles []string
				for _, r := range resp.User.GetRoles() {
				LP:
					for _, role := range rolesList {
						if r == role.RoleId {
							roles = append(roles, role.RoleName)
							break LP
						}
					}
				}
				row = append(row, strings.Join(roles, ","))
			}
			// 组织
			{
			LP1:
				for _, group := range groupList {
					if resp.User.GetGroup() == group.GroupId {
						row = append(row, langx.GetLangValue(langData, group.GroupName, ""))
						break LP1
					}
				}
			}
			// App
			{
				var apps []string
				for _, a := range resp.User.GetApps() {
				LP2:
					for _, app := range appIdList {
						if a == app.AppId {
							apps = append(apps, langx.GetLangValue(langData, app.AppName, ""))
							break LP2
						}
					}
				}
				row = append(row, strings.Join(apps, ","))
			}
			// Language
			{
				row = append(row, resp.User.GetLanguage())
			}
			// Timezone
			{
				row = append(row, resp.User.GetTimezone())
			}

			items = append(items, row)

		}

		defer f.Close()

		// 写入数据
		writer.WriteAll(items)
		writer.Flush() // 此时才会将缓冲区数据写入

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_029"),
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		fo, err := os.Open(filename)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		defer func() {
			fo.Close()
			os.Remove(filename)
		}()

		// 写入文件到 minio
		minioClient, err := storagecli.NewClient(domain)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		appRoot := "app_" + appID
		filePath := path.Join(appRoot, "csv", "check_"+timestamp+".csv")
		path, err := minioClient.SavePublicObject(fo, filePath, "text/csv")
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		// 判断顾客上传文件是否在设置的最大存储空间以内
		canUpload := filex.CheckCanUpload(domain, float64(path.Size))
		if canUpload {
			// 如果没有超出最大值，就对顾客的已使用大小进行累加
			err = filex.ModifyUsedSize(domain, float64(path.Size))
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
		} else {
			// 如果已达上限，则删除刚才上传的文件
			minioClient.DeleteObject(path.Name)
			path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_007"),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			Progress:    100,
			File: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database: db,
		}, userID)

	}()

	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, userProcessName, ActionDownloadUsers)),
		Data:    gin.H{},
	})

}

func BuildAndValidate(p BVParam) (*user.UploadRequest, error) {
	// 验证header传入字段内容是否全部正确
	headers := []string{"UserID", "UserName", "NoticeEmail", "Password", "Profiles", "Group", "Apps", "Language", "Timezone"}
	l, r := slicex.StringSliceCompare(p.Header, headers)
	if len(l) > 0 {
		// 说明当前header中有未包含的字段，提示无法上传
		return nil, fmt.Errorf("テーブルヘッダは、サポートされていないタイトルがあります【%v】", l)
	}
	if len(r) > 0 {
		// 说明当前header中有未包含的字段，提示无法上传
		return nil, fmt.Errorf("テーブルヘッダは次のタイトルがあります【%v】は正しくインポートされませんでした", r)
	}

	var u user.UploadRequest

	// 循环header，判断数据类型和存在判断
	for i, h := range p.Header {
		col := p.Data[i]
		// 用户ID，必须要以@domain结尾，然后ID只能是英文字母和下划线组成
		if h == "UserID" {
			// 必须验证
			if len(col) == 0 {
				return nil, fmt.Errorf("UserIDは必須フィールドです。正しい値を入力してください")
			}
			// 必须要以@domain结尾
			if !strings.HasSuffix(col, "@"+p.Domain) {
				return nil, fmt.Errorf("UserIDは【@%v】で終わる必要があります。正しい値を渡してください", p.Domain)
			}

			// 以@符号切分数据
			datas := strings.Split(col, "@")
			if len(datas) > 2 {
				return nil, fmt.Errorf("UserIDは、英語の文字とアンダースコアのみで構成できます。正しい値を入力してください")
			}

			// 正常分割的情况，判断是否是英文字母和下划线组成
			reg := regexp.MustCompile(`^[a-z0-9A-Z]+[-a-z0-9A-Z._]*`)
			if !reg.MatchString(datas[0]) {
				return nil, fmt.Errorf("UserIDは、英語の文字とアンダースコアのみで構成できます。正しい値を入力してください")
			}

			// 判断是否已经存在
			userService := user.NewUserService("manage", client.DefaultClient)

			var req user.FindUserRequest
			req.Type = 1
			req.Email = col
			req.Database = p.Database
			_, err := userService.FindUser(context.TODO(), &req)
			if err != nil {
				u.Email = col
				continue
			}
			return nil, fmt.Errorf("UserIDはすでに存在します。変更してください")
		}

		// 用户名，必须有值,且不存在
		if h == "UserName" {
			if len(col) == 0 {
				return nil, fmt.Errorf("UserNameは必須フィールドです。正しい値を入力してください")
			}

			// 验证是否唯一
			userService := user.NewUserService("manage", client.DefaultClient)

			var req user.FindUsersRequest
			req.UserName = col
			req.InvalidatedIn = "true"
			// 从共通中获取参数
			req.Domain = p.Domain
			req.Database = p.Database

			response, err := userService.FindUsers(context.TODO(), &req)
			if err != nil {
				return nil, err
			}

			if len(response.GetUsers()) == 0 {
				u.UserName = col
				continue
			}

			return nil, fmt.Errorf("UserNameはすでに存在します。変更してください")
		}

		// 用户通知邮箱，必须有值
		if h == "NoticeEmail" {
			if len(col) == 0 {
				return nil, fmt.Errorf("NoticeEmailは必須フィールドです。正しい値を入力してください")
			}
			u.NoticeEmail = col
		}

		// 密码，必须符合要求，而且必须有值
		if h == "Password" {
			if len(col) == 0 {
				return nil, fmt.Errorf("Passwordは必須フィールドです。正しい値を入力してください")
			}

			if b := cryptox.VerifyPassword(col, 8, 32); !b {
				return nil, fmt.Errorf("Passwordは8文字以上で、数字、大文字、特殊文字（％¥＃\u0026など）を含める必要があります。")
			}

			u.Password = col
			continue
		}

		// 角色以逗号分割，必须存在，并且必须有一个
		if h == "Profiles" {
			if len(col) == 0 {
				return nil, fmt.Errorf("Profileは必須フィールドです。正しい値を入力してください")
			}

			datas := strings.Split(col, ",")

			var roles []string

			// 判断是否存在
			for _, r := range datas {
				hasExist := false
			LP:
				for _, role := range p.RoleList {
					if r == role.RoleName {
						hasExist = true
						roles = append(roles, role.RoleId)
						break LP
					}
				}

				if !hasExist {
					return nil, fmt.Errorf("現在のProfile【%s】は存在しません。正しい値を渡してください", col)
				}
			}

			u.Roles = roles
			continue
		}

		// Group只有一个
		if h == "Group" {
			if len(col) == 0 {
				return nil, fmt.Errorf("Groupは必須フィールドです。正しい値を入力してください")
			}
			hasExist := false
			// 判断是否存在
		LP1:
			for _, group := range p.GroupList {
				// 从多语言获取组织名
				gName := langx.GetLangValue(p.LangData, group.GroupName, "")

				if col == gName {
					hasExist = true
					u.Group = group.GroupId
					break LP1
				}
			}

			if !hasExist {
				return nil, fmt.Errorf("現在のGroup【%s】は存在しません。正しい値を渡してください", col)
			}

			continue
		}

		// 角色以逗号分割，必须存在，并且必须有一个
		if h == "Apps" {
			if len(col) == 0 {
				return nil, fmt.Errorf("Appsは必須フィールドです。正しい値を入力してください")
			}

			datas := strings.Split(col, ",")

			var apps []string

			// 判断是否存在
			for _, a := range datas {
				hasExist := false
			LP2:
				for _, app := range p.AppList {
					// 从多语言获取组织名
					aName := langx.GetLangValue(p.LangData, app.AppName, "")

					if a == aName {
						hasExist = true
						apps = append(apps, app.AppId)
						break LP2
					}
				}

				if !hasExist {
					return nil, fmt.Errorf("現在のApp【%s】は存在しません。正しい値を渡してください", a)
				}
			}

			u.Apps = apps
			continue
		}
		// 语言必须有值，且必须为系统支持的语言
		if h == "Language" {
			if len(col) == 0 {
				return nil, fmt.Errorf("Languageは必須フィールドです。正しい値を入力してください")
			}
			// 获取所有语言数据
			var opss client.CallOption = func(o *client.CallOptions) {
				o.RequestTimeout = time.Minute * 5
				o.DialTimeout = time.Minute * 5
			}
			languageService := language.NewLanguageService("global", client.DefaultClient)
			var lgReq language.FindLanguagesRequest
			lgReq.Domain = p.Domain
			lgReq.Database = p.Database
			lgResponse, err := languageService.FindLanguages(context.TODO(), &lgReq, opss)
			if err != nil {
				return nil, fmt.Errorf("Languageはシステムに存在しません。正しい値を入力してください")
			}
			isOk := false
			for _, lang := range lgResponse.GetLanguageList() {
				if lang.LangCd == col {
					u.Language = col
					isOk = true
					break
				}
			}
			if !isOk {
				return nil, fmt.Errorf("Languageはシステムに存在しません。正しい値を入力してください")
			}
			continue
		}
		// 时区必须有值，且必须为系统支持的时区
		if h == "Timezone" {
			if len(col) == 0 {
				return nil, fmt.Errorf("Timezoneは必須フィールドです。正しい値を入力してください")
			}
			var timezone []*Timezone
			err := filex.ReadFile("assets/timezone/zone.json", &timezone)
			if err != nil {
				return nil, fmt.Errorf("Timezoneはシステムに存在しません。正しい値を入力してください")
			}
			isOk := false
			for _, tz := range timezone {
				if tz.Value == col {
					u.Timezone = col
					isOk = true
					break
				}
			}
			if !isOk {
				return nil, fmt.Errorf("Timezoneはシステムに存在しません。正しい値を入力してください")
			}
			continue
		}
	}

	// 设置默认值
	if len(u.Apps) > 0 {
		u.CurrentApp = u.Apps[0]
	}
	u.Password = cryptox.GenerateMd5Password(u.Password, u.Email)
	u.CustomerId = p.Database
	u.Domain = p.Domain
	u.NoticeEmailStatus = "UnVerified"
	u.Theme = "default"
	u.UserType = 2
	u.Writer = p.Writer
	u.Database = p.Database

	id := primitive.NewObjectID()
	u.UserId = id.Hex()

	u.Status = user.SendStatus_SECTION

	return &u, nil
}
