package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/logger"
)

// Logger 日志
type Logger struct{}

// log出力使用
const (
	ActionFindLoggers       = "FindLoggers"
	ActionAddLogger         = "AddLogger"
	ActionCreateLoggerIndex = "CreateLoggerIndex"
)

// FindLoggers 获取日志
func (l *Logger) FindLoggers(ctx context.Context, req *logger.LoggersRequest, rsp *logger.LoggersResponse) error {
	loggers, total, err := model.FindLoggers(
		req.GetAppName(),
		req.GetUserId(),
		req.GetDomain(),
		req.GetLogType(),
		req.GetClientIp(),
		req.GetLevel(),
		req.GetStartTime(),
		req.GetEndTime(),
		req.GetPageIndex(),
		req.GetPageSize())
	if err != nil {
		return err
	}

	res := &logger.LoggersResponse{}
	for _, log := range loggers {
		res.Loggers = append(res.Loggers, log.ToProto())
	}

	res.Total = total
	*rsp = *res

	return nil
}

// AddLogger 添加日志记录
func (l *Logger) AddLogger(ctx context.Context, req *logger.AddRequest, rsp *logger.AddResponse) error {

	t, err := time.Parse("2006-01-02 15:04:05.000000", req.GetTime())
	if err != nil {
		return err
	}

	params := model.Logger{
		AppName:   req.GetAppName(),
		UserID:    req.GetUserId(),
		Domain:    req.GetDomain(),
		LogType:   req.GetLogType(),
		ProcessID: req.GetProcessId(),
		ClientIP:  req.GetClientIp(),
		Source:    req.GetSource(),
		Msg:       req.GetMsg(),
		Time:      t,
		Level:     req.GetLevel(),
		Params:    req.GetParams(),
	}

	err = model.AddLogger(&params)
	if err != nil {
		return err
	}

	return nil
}

// CreateLoggerIndex 创建日志索引
func (l *Logger) CreateLoggerIndex(ctx context.Context, req *logger.CreateIndexRequest, rsp *logger.CreateIndexResponse) error {

	err := model.CreateLoggerIndex()
	if err != nil {
		return err
	}

	return nil
}
