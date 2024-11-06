package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	lg "github.com/micro/go-plugins/logger/logrus/v2"

	"rxcsoft.cn/pit3/srv/task/handler"
	"rxcsoft.cn/pit3/srv/task/proto/history"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/pit3/srv/task/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func main() {

	service := micro.NewService(
		micro.Name("task"),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
	)

	// 设置log出力的配置
	logger.DefaultLogger = lg.NewLogger(lg.WithLogger(log))

	// 初始化配置文件
	utilsServer.Start()

	// 初始化配置文件
	server.LoadConfig()

	// 启动DB服务
	server.DBStart()

	// micro服务初始化
	service.Init()

	// 注册handler
	task.RegisterTaskServiceHandler(service.Server(), new(handler.Task))
	schedule.RegisterScheduleServiceHandler(service.Server(), new(handler.Schedule))
	history.RegisterHistoryServiceHandler(service.Server(), new(handler.History))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
