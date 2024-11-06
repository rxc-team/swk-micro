package utils

import (
	"github.com/sirupsen/logrus"
	"rxcsoft.cn/pit3/lib/logger"
)

// log出力
const (
	MsgProcessStarted = "Process Started"
	MsgProcessEnded   = "Process Ended"
)

// InfoLog 调试处理（调试使用，一般用于服务开始，服务当前进度，服务的数据以及服务结束显示log使用）
func InfoLog(action, msg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "workflow",
		"save":       "false",
		"log_type":   logger.Micro,
		"process_id": action,
	}).Info(msg)
}

// DebugLog 调试处理（调试使用，一般用于服务开始，服务当前进度，服务的数据以及服务结束显示log使用）
func DebugLog(action, msg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "workflow",
		"save":       "false",
		"log_type":   logger.Micro,
		"process_id": action,
	}).Debug(msg)
}

// ErrorLog 错误（系统启动等场合使用）
func ErrorLog(action, errMsg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "workflow",
		"save":       "false",
		"log_type":   logger.Micro,
		"process_id": action,
	}).Error(errMsg)
}

// FatalLog 致命错误（系统启动等场合使用）
func FatalLog(action, errMsg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "workflow",
		"save":       "false",
		"log_type":   logger.System,
		"process_id": action,
	}).Fatal(errMsg)
}

// PanicLog 宕机错误（系统启动等场合使用）
func PanicLog(action, errMsg string) {
	log := logger.New()
	// 出力日志
	log.WithFields(logrus.Fields{
		"app_name":   "workflow",
		"save":       "false",
		"log_type":   logger.System,
		"process_id": action,
	}).Panic(errMsg)
}
