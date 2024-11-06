package logger

import (
	"github.com/sirupsen/logrus"
)

// DefaultHook 本系统默认的Hook处理
type DefaultHook struct {
}

// NewHook 获取Hook
func NewHook() *DefaultHook {
	return &DefaultHook{}
}

// Fire 当前消息处理
func (h *DefaultHook) Fire(entry *logrus.Entry) error {

	save := entry.Data["save"]

	if save.(string) == "true" {

		logData := &Logger{
			Source: entry.Data["source"].(string),
			Level:  entry.Level.String(),
			Msg:    entry.Message,
			Time:   entry.Time,
		}

		if v, ok := entry.Data["app_name"]; ok {
			logData.AppName = v.(string)
		}
		if v, ok := entry.Data["user_id"]; ok {
			logData.UserID = v.(string)
		}
		if v, ok := entry.Data["domain"]; ok {
			logData.Domain = v.(string)
		}
		if v, ok := entry.Data["log_type"]; ok {
			logData.LogType = string(v.(LogType))
		}
		if v, ok := entry.Data["process_id"]; ok {
			logData.ProcessID = v.(string)
		}

		if v, ok := entry.Data["params"]; ok {
			logData.Params = v.(map[string]string)
		}

		if v, ok := entry.Data["client_ip"]; ok {
			logData.ClientIP = v.(string)
		}

		err := AddLogger(logData)
		if err != nil {
			return err
		}
	}

	return nil
}

// Levels 当前hook需要的消息层级
func (h *DefaultHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
