package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/server/grpc"
	lg "github.com/micro/go-plugins/logger/logrus/v2"
	"github.com/sirupsen/logrus"

	"rxcsoft.cn/pit3/srv/database/handler"
	av "rxcsoft.cn/pit3/srv/database/handler/approve"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/proto/check"
	"rxcsoft.cn/pit3/srv/database/proto/copy"
	"rxcsoft.cn/pit3/srv/database/proto/datahistory"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/feed"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/generate"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/database/proto/print"
	"rxcsoft.cn/pit3/srv/database/proto/query"
	"rxcsoft.cn/pit3/srv/database/proto/template"
	"rxcsoft.cn/pit3/srv/database/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func init() {
}

func main() {
	service := micro.NewService(
		micro.Name("database"),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
	)

	service.Server().Init(grpc.MaxMsgSize(100 * 1024 * 1024))

	log.SetLevel(logrus.DebugLevel)

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
	datastore.RegisterDataStoreServiceHandler(service.Server(), new(handler.Datastore))
	field.RegisterFieldServiceHandler(service.Server(), new(handler.Field))
	print.RegisterPrintServiceHandler(service.Server(), new(handler.Print))
	item.RegisterItemServiceHandler(service.Server(), new(handler.Item))
	option.RegisterOptionServiceHandler(service.Server(), new(handler.Option))
	query.RegisterQueryServiceHandler(service.Server(), new(handler.Query))
	datahistory.RegisterHistoryServiceHandler(service.Server(), new(handler.History))
	template.RegisterTemplateServiceHandler(service.Server(), new(handler.Template))
	feed.RegisterImportServiceHandler(service.Server(), new(handler.Import))
	approve.RegisterApproveServiceHandler(service.Server(), new(av.Approve))
	check.RegisterCheckHistoryServiceHandler(service.Server(), new(handler.CheckHistory))
	copy.RegisterCopyServiceHandler(service.Server(), new(handler.Copy))
	generate.RegisterGenerateServiceHandler(service.Server(), new(handler.Generate))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
