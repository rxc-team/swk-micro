package initx

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/global/proto/logger"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
)

// log出力
const (
	systemProcessName = "System"
	actionInitApp     = "InitApp"

	defaultDomain      = "proship.co.jp"
	defaultDomainEnv   = "DEFAULT_DOMAIN"
	defaultApp         = "system"
	defaultGroup       = "root"
	defaultPassword    = "Rxc1234%"
	defaultPasswordEnv = "DEFAULT_PASSWORD"
)

// InitApp 系统初始化
func InitApp() {
	// 去掉重复数据，并添加email唯一索引
	addUserCollectionIndex()
	// 判断超级管理员用户是否存在
	if !existSuperAdmin() {
		addSuperAdminUser()
		createLoggerIndex()
	}
	// 添加默认的任务
	addDefaultSchedule()
}

func addDefaultSchedule() {
	userService := user.NewUserService("manage", client.DefaultClient)
	req := user.FindDefaultUserRequest{
		UserType: 2,
	}

	superUser, err := userService.FindDefaultUser(context.TODO(), &req)
	if err != nil {
		return
	}

	userId := superUser.GetUser().GetUserId()
	// 去掉重复任务数据，并添加schedule_name唯一索引
	addScheduleNameUniqueIndex(userId)
	// 添加备份任务
	addBackupSchedule(userId)
	// 添加备份清理任务
	addBackupClearSchedule(userId)
}

// 添加备份任务
func addBackupSchedule(userId string) {
	// 默认值
	db := "system"
	scheduleType := "db-backup"
	scheduleSpec := "TZ=Asia/Shanghai 0 2 * * ?"

	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)

	var freq schedule.SchedulesRequest
	freq.PageIndex = 1
	freq.PageSize = 1
	freq.Database = db
	freq.UserId = userId
	freq.ScheduleType = scheduleType
	freq.RunNow = false

	response, err := scheduleService.FindSchedules(context.TODO(), &freq)
	if err != nil {
		loggerx.ErrorLog("addBackupSchedule", err.Error())
		return
	}
	// 如果已经存在，则直接返回
	if response.GetTotal() == 1 {
		return
	}

	// 不存在的场合，添加
	domain := os.Getenv(defaultDomainEnv)
	if len(domain) == 0 {
		domain = defaultDomain
	}

	// 添加备份任务,每天00:00执行,（不存在的情况）

	var req schedule.AddRequest
	req.Writer = userId
	req.Database = db
	req.Params = make(map[string]string)
	req.Params["db"] = db
	req.Params["domain"] = domain
	req.Params["app_id"] = "system"
	req.Params["client_ip"] = "127.0.0.1"
	req.ScheduleName = "db backup"
	req.Spec = scheduleSpec
	req.Multi = 0
	req.RetryTimes = 1
	req.RetryInterval = 1000
	req.StartTime = time.Now().Format("2006-01-02")
	req.EndTime = "3000-01-01"
	req.ScheduleType = scheduleType
	req.RunNow = false
	req.Status = "1"

	_, err = scheduleService.AddSchedule(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("addDefaultSchedule", err.Error())
		return
	}
}

// 添加备份清理任务
func addBackupClearSchedule(userId string) {
	// 默认值
	db := "system"
	scheduleType := "db-backup-clean"
	scheduleSpec := "TZ=Asia/Shanghai 0 3 * * ?"

	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)

	var freq schedule.SchedulesRequest
	freq.PageIndex = 1
	freq.PageSize = 1
	freq.Database = db
	freq.UserId = userId
	freq.ScheduleType = scheduleType
	freq.RunNow = false

	response, err := scheduleService.FindSchedules(context.TODO(), &freq)
	if err != nil {
		loggerx.ErrorLog("addBackupClearSchedule", err.Error())
		return
	}
	// 如果已经存在，则直接返回
	if response.GetTotal() == 1 {
		return
	}

	// 不存在的场合，添加
	domain := os.Getenv(defaultDomainEnv)
	if len(domain) == 0 {
		domain = defaultDomain
	}

	// 添加备份任务,每天00:00执行,（不存在的情况）

	var req schedule.AddRequest
	req.Writer = userId
	req.Database = db
	req.Params = make(map[string]string)
	req.Params["db"] = db
	req.Params["domain"] = domain
	req.Params["app_id"] = "system"
	req.Params["client_ip"] = "127.0.0.1"
	req.ScheduleName = "db backup clean"
	req.Spec = scheduleSpec
	req.Multi = 0
	req.RetryTimes = 1
	req.RetryInterval = 1000
	req.StartTime = time.Now().Format("2006-01-02")
	req.EndTime = "3000-01-01"
	req.ScheduleType = scheduleType
	req.RunNow = false
	req.Status = "1"

	_, err = scheduleService.AddSchedule(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("addDefaultSchedule", err.Error())
		return
	}
}

// 判断默认的超级管理员用户是否存在，不存在则添加
func existSuperAdmin() bool {
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process FindDefaultUser:%s", loggerx.MsgProcessStarted))
	userService := user.NewUserService("manage", client.DefaultClient)

	domain := os.Getenv(defaultDomainEnv)
	if len(domain) == 0 {
		domain = defaultDomain
	}

	req := user.FindDefaultUserRequest{
		UserType: 2,
	}

	superUser, err := userService.FindDefaultUser(context.TODO(), &req)
	if err != nil {
		return false
	}
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process FindDefaultUser:%s", loggerx.MsgProcessEnded))

	if superUser.GetUser().GetUserId() != "" {
		return true
	}

	return false
}

// 添加超级管理员用户
func addSuperAdminUser() error {
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddUser:%s", loggerx.MsgProcessStarted))

	// 添加默认的超级管理员角色
	roleID, e := addSuperAdminRole()
	if e != nil {
		return e
	}

	userService := user.NewUserService("manage", client.DefaultClient)

	domain := os.Getenv(defaultDomainEnv)
	if len(domain) == 0 {
		domain = defaultDomain
	}

	password := os.Getenv(defaultPasswordEnv)
	if len(password) == 0 {
		password = defaultPassword
	}

	email := cryptox.GenerateMailAddress("superadmin", domain)
	admin := user.AddUserRequest{
		UserName:   "SYSTEM",
		Email:      email,
		Password:   cryptox.GenerateMd5Password(password, email),
		CurrentApp: defaultApp,
		Timezone:   "Asia/Tokyo",
		Group:      defaultGroup,
		Roles:      []string{roleID},
		Apps:       []string{defaultApp},
		Language:   "zh-CN",
		Domain:     domain,
		CustomerId: "system",
		UserType:   2,
		Writer:     "system",
		Database:   "system",
	}
	response, err := userService.AddUser(context.TODO(), &admin)
	if err != nil {
		loggerx.SystemLog(true, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcessError, "AddUser", err))
		return err
	}
	loggerx.SystemLog(false, true, actionInitApp, fmt.Sprintf("Process AddUser [%s] Succeed", response.GetUserId()))

	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddUser:%s", loggerx.MsgProcessEnded))

	return nil
}

// 创建日志索引
func createLoggerIndex() error {
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process CreateLoggerIndex:%s", loggerx.MsgProcessStarted))

	loggerService := logger.NewLoggerService("global", client.DefaultClient)

	var req logger.CreateIndexRequest

	_, err := loggerService.CreateLoggerIndex(context.TODO(), &req)
	if err != nil {
		loggerx.SystemLog(true, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcessError, "CreateLoggerIndex", err))
		return err
	}
	loggerx.SystemLog(false, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcesSucceed, "CreateLoggerIndex"))
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process CreateLoggerIndex:%s", loggerx.MsgProcessEnded))

	return nil
}

// 添加超级管理员的角色
func addSuperAdminRole() (roleID string, err error) {
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddRole:%s", loggerx.MsgProcessStarted))

	roleService := role.NewRoleService("manage", client.DefaultClient)

	domain := os.Getenv(defaultDomainEnv)
	if len(domain) == 0 {
		domain = defaultDomain
	}

	req := role.AddRoleRequest{
		RoleName:    "SYSTEM",
		Description: "System super administrator",
		Domain:      domain,
		RoleType:    2,
		Writer:      "system",
		Database:    "system",
	}

	response, err := roleService.AddRole(context.TODO(), &req)
	if err != nil {
		loggerx.SystemLog(true, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcessError, "AddRole", err))
		return "", err
	}
	loggerx.SystemLog(false, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddRole"))
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddRole:%s", loggerx.MsgProcessEnded))

	return response.GetRoleId(), nil
}

// 去掉重复用户数据，并添加email唯一索引
func addUserCollectionIndex() (err error) {
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddUserCollectionIndex:%s", loggerx.MsgProcessStarted))
	userService := user.NewUserService("manage", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
	}

	var req user.AddUserIndexRequest
	req.Db = "system"
	_, err = userService.AddUserCollectionIndex(context.TODO(), &req, opss)
	if err != nil {
		loggerx.SystemLog(true, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcessError, "AddUserCollectionIndex", err))
		return err
	}
	loggerx.SystemLog(false, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddUserCollectionIndex"))
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddUserCollectionIndex:%s", loggerx.MsgProcessEnded))
	return nil
}

// 去掉重复任务数据，并添加schedule_name唯一索引
func addScheduleNameUniqueIndex(userID string) (err error) {
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddScheduleNameUniqueIndex:%s", loggerx.MsgProcessStarted))
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)
	var addReq schedule.ScheduleNameIndexRequest
	addReq.Db = "system"
	addReq.UserId = userID
	_, err = scheduleService.AddScheduleNameUniqueIndex(context.TODO(), &addReq)
	if err != nil {
		loggerx.SystemLog(true, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcessError, "AddScheduleNameUniqueIndex", err))
		return err
	}

	loggerx.SystemLog(false, true, actionInitApp, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddScheduleNameUniqueIndex"))
	loggerx.SystemLog(false, false, actionInitApp, fmt.Sprintf("Process AddScheduleNameUniqueIndex:%s", loggerx.MsgProcessEnded))
	return nil
}
