package main

import (
	"time"

	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/server/grpc"
	lg "github.com/micro/go-plugins/logger/logrus/v2"
	"github.com/sirupsen/logrus"

	"rxcsoft.cn/pit3/srv/import/handler"
	"rxcsoft.cn/pit3/srv/import/proto/upload"
	"rxcsoft.cn/pit3/srv/import/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
	"rxcsoft.cn/utils/storage/client"
)

var log = myLogger.New()

func init() {
	i18n.SetDefaultLanguage("ja-JP")
}

func main() {

	service := micro.NewService(
		micro.Name("import"),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
		micro.AfterStart(func() error {
			client.InitStorageClient()
			return nil
		}),
	)

	log.SetLevel(logrus.DebugLevel)

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
	upload.RegisterUploadServiceHandler(service.Server(), new(handler.Upload))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
