package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/server/grpc"
	lg "github.com/micro/go-plugins/logger/logrus/v2"

	"rxcsoft.cn/pit3/srv/journal/handler"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/pit3/srv/journal/proto/subject"
	"rxcsoft.cn/pit3/srv/journal/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func main() {

	service := micro.NewService(
		micro.Name("journal"),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
	)

	service.Server().Init(grpc.MaxMsgSize(100 * 1024 * 1024))

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
	journal.RegisterJournalServiceHandler(service.Server(), new(handler.Journal))
	subject.RegisterSubjectServiceHandler(service.Server(), new(handler.Subject))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
