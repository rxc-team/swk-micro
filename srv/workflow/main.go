package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	lg "github.com/micro/go-plugins/logger/logrus/v2"

	"rxcsoft.cn/pit3/srv/workflow/handler"
	"rxcsoft.cn/pit3/srv/workflow/proto/example"
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/proto/process"
	"rxcsoft.cn/pit3/srv/workflow/proto/relation"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
	"rxcsoft.cn/pit3/srv/workflow/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func main() {

	service := micro.NewService(
		micro.Name("workflow"),
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
	example.RegisterExampleServiceHandler(service.Server(), new(handler.Example))
	node.RegisterNodeServiceHandler(service.Server(), new(handler.Node))
	process.RegisterProcessServiceHandler(service.Server(), new(handler.Process))
	workflow.RegisterWfServiceHandler(service.Server(), new(handler.Workflow))
	relation.RegisterRelationServiceHandler(service.Server(), new(handler.Relation))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
