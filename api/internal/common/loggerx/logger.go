package loggerx

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"rxcsoft.cn/pit3/lib/logger"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// log出力
const (
	MsgProcessStarted = "Process Started"
	MsgProcessEnded   = "Process Ended"
	MsgProcesSucceed  = "Process %s Succeed"
	MsgProcessError   = "Process %s Error:%v"
)

// InfoLog 一般处理（调试使用，日志等，一般用于服务开始，服务当前进度，服务的数据以及服务结束显示log使用）
func InfoLog(c *gin.Context, action, msg string) {
	var userId, domain string
	//根据上下文获取载荷userInfo

	if userInfo, exit := c.Get("userInfo"); exit {
		if u, ok := userInfo.(*user.User); ok {
			userId = u.GetEmail()
			domain = u.GetDomain()
		}
	}

	log := logger.New()
	// 请求IP
	clientIP := c.ClientIP()

	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"user_id":    userId,
		"domain":     domain,
		"client_ip":  clientIP,
		"log_type":   logger.Micro,
		"process_id": action,
	}).Info(msg)
}

// LoginLog 登录处理（login专用）
func LoginLog(clientIP, userID, domain, action, errMsg string, isSuccesd bool) {
	log := logger.New()
	// 出力日志
	if isSuccesd {
		log.WithFields(logrus.Fields{
			"app_name":   "internal",
			"save":       "true",
			"user_id":    userID,
			"domain":     domain,
			"client_ip":  clientIP,
			"log_type":   logger.Login,
			"process_id": action,
		}).Info(errMsg)
	} else {
		log.WithFields(logrus.Fields{
			"app_name":   "internal",
			"save":       "true",
			"user_id":    userID,
			"domain":     domain,
			"client_ip":  clientIP,
			"log_type":   logger.Login,
			"process_id": action,
		}).Error(errMsg)
	}

}

// ProcessLog 处理进度处理
func ProcessLog(c *gin.Context, action, msg string, params map[string]string) {

	var userId, domain string
	//根据上下文获取载荷userInfo

	if userInfo, exit := c.Get("userInfo"); exit {
		if u, ok := userInfo.(*user.User); ok {
			userId = u.GetEmail()
			domain = u.GetDomain()
		}
	}

	log := logger.New()
	// 设置不输出到控制台
	log.SetOutput(ioutil.Discard)
	// 请求IP
	clientIP := c.ClientIP()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "true",
		"user_id":    userId,
		"domain":     domain,
		"client_ip":  clientIP,
		"log_type":   logger.Micro,
		"process_id": action,
		"params":     params,
	}).Info(msg)
}

// JobProcessLog 处理进度处理
func JobProcessLog(userID, domain, clientIP, action, msg string, params map[string]string) {
	log := logger.New()
	// 设置不输出到控制台
	log.SetOutput(ioutil.Discard)
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "true",
		"user_id":    userID,
		"domain":     domain,
		"client_ip":  clientIP,
		"log_type":   logger.Micro,
		"process_id": action,
		"params":     params,
	}).Info(msg)
}

// SuccessLog 成功处理（包含新规，更新，无效，删除等处理）
func SuccessLog(c *gin.Context, action, msg string) {

	var userId, domain string
	//根据上下文获取载荷userInfo

	if userInfo, exit := c.Get("userInfo"); exit {
		if u, ok := userInfo.(*user.User); ok {
			userId = u.GetEmail()
			domain = u.GetDomain()
		}
	}

	log := logger.New()
	// 请求IP
	clientIP := c.ClientIP()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"user_id":    userId,
		"domain":     domain,
		"client_ip":  clientIP,
		"log_type":   logger.Micro,
		"process_id": action,
	}).Info(msg)
}

// FailureLog 失败处理（包含新规，更新，无效，删除等处理）
func FailureLog(c *gin.Context, action, errMsg string) {

	var userId, domain string
	//根据上下文获取载荷userInfo

	if userInfo, exit := c.Get("userInfo"); exit {
		if u, ok := userInfo.(*user.User); ok {
			userId = u.GetEmail()
			domain = u.GetDomain()
		}
	}

	log := logger.New()
	// 请求IP
	clientIP := c.ClientIP()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"user_id":    userId,
		"domain":     domain,
		"client_ip":  clientIP,
		"log_type":   logger.Micro,
		"process_id": action,
	}).Error(errMsg)
}

// FatalLog 致命错误（系统启动等场合使用）
func FatalLog(action, errMsg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"log_type":   logger.System,
		"process_id": action,
	}).Fatal(errMsg)
}

// ErrorLog 普通错误（任务处理等场合使用）
func ErrorLog(action, errMsg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"log_type":   logger.Micro,
		"process_id": action,
	}).Error(errMsg)
}

// DebugLog 调试日志
func DebugLog(action, msg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"log_type":   logger.Micro,
		"process_id": action,
	}).Debug(msg)
}

// PanicLog 宕机错误（系统启动等场合使用）
func PanicLog(action, errMsg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"log_type":   logger.System,
		"process_id": action,
	}).Panic(errMsg)
}

// SystemLog app初始化处理（app初始化使用）
func SystemLog(isError, isSave bool, action, msg string) {
	log := logger.New()
	// var save string
	// if isSave {
	// 	save = "true"
	// } else {
	// 	save = "false"
	// }

	if isError {
		// 出力日志
		log.WithFields(logrus.Fields{
			"app_name":   "internal",
			"save":       "false",
			"domain":     "proship.co.jp",
			"client_ip":  "127.0.0.1",
			"log_type":   logger.Micro,
			"process_id": action,
		}).Error(msg)
		return
	}

	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "internal",
		"save":       "false",
		"domain":     "proship.co.jp",
		"client_ip":  "127.0.0.1",
		"log_type":   logger.Micro,
		"process_id": action,
	}).Info(msg)
}
