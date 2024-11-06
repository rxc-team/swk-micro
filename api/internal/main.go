package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/broker/rabbitmq/v2"
	lg "github.com/micro/go-plugins/logger/logrus/v2"
	"github.com/sirupsen/logrus"

	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/router"
	"rxcsoft.cn/pit3/api/internal/server"
	"rxcsoft.cn/pit3/api/internal/server/start"
	"rxcsoft.cn/pit3/api/internal/system/eventx"
	"rxcsoft.cn/pit3/api/internal/system/initx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/scriptx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"

	"rxcsoft.cn/pit3/lib/msg"
	myLogger "rxcsoft.cn/utils/logger"
	"rxcsoft.cn/utils/mq"
	utilsServer "rxcsoft.cn/utils/server"
	"rxcsoft.cn/utils/storage/client"
)

var log = myLogger.New()
var version = "1.0.0"

func init() {
	i18n.SetDefaultLanguage("ja-JP")
	// 设置版本号
	v := os.Getenv("VERSION")
	if len(v) > 0 {
		version = v
	}
}

// @title PIT3内部服务api接口文档
// @version 1.0
// @description PIT3内部服务api接口文档。

// @host localhost:8080
// @BasePath /internal/api/v1

// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
func main() {

	// 根据运行环境创建服务名称
	serviceName := "go.micro.api.internal"

	job := new(jobx.Job)

	// 创建服务
	service := web.NewService(
		web.Name(serviceName),
		web.Version(version),
		web.RegisterTTL(time.Second*30),
		web.RegisterInterval(time.Second*10),
		web.AfterStart(func() error {
			// utils.InitLocale()
			return nil
		}),
	)

	log.SetLevel(logrus.DebugLevel)
	// 设置log出力的配置
	logger.DefaultLogger = lg.NewLogger(lg.WithLogger(log))

	// 初始化配置文件
	utilsServer.Start()

	start.DBStart()

	// micro服务初始化
	service.Init()

	// 订阅事件
	bk := mq.NewBroker()
	bk.Subscribe("message.send", eventx.SendToUser, broker.Queue("message.send"), broker.DisableAutoAck(), rabbitmq.DurableQueue())
	bk.Subscribe("message.task", eventx.SendMsg, broker.Queue("message.task"), broker.DisableAutoAck(), rabbitmq.DurableQueue())
	bk.Subscribe("job.add", eventx.AddJob, broker.DisableAutoAck())
	bk.Subscribe("job.stop", eventx.StopJob, broker.DisableAutoAck())
	bk.Subscribe("job.delete", eventx.DeleteJob, broker.DisableAutoAck())
	bk.Subscribe("acl.refresh", eventx.LoadCasbinPolicy, broker.DisableAutoAck())

	// 创建Handler (using Gin)
	host := server.NewWebHost().
		UseGinLogger().
		UseGinRecovery().
		UseExternalAPIRoutes(router.InitRouter).
		ConfigureServices(func() {
			go func() {
				initx.InitApp()
				msg.LoadMsg()
				client.InitStorageClient()
				job.Initialize()
				scriptx.Init()
				wsx.Manager.Start()
			}()
		})

	// 注册Handler
	service.Handle("/", host.GinEngine)

	// 运行服务
	if err := service.Run(); err != nil {
		loggerx.FatalLog("server start", fmt.Sprintf("server start has error: %v", err))
	}
}
