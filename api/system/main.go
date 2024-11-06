package main

import (
	"time"

	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/web"
	lg "github.com/micro/go-plugins/logger/logrus/v2"
	"github.com/sirupsen/logrus"
	"rxcsoft.cn/pit3/api/system/router"
	"rxcsoft.cn/pit3/api/system/server"
	"rxcsoft.cn/pit3/api/system/server/start"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func main() {
	// 根据运行环境创建服务名称
	serviceName := "go.micro.api.system"

	// 创建服务
	service := web.NewService(
		web.Name(serviceName),
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
	if err := service.Init(); err != nil {
		log.Fatal(err)
	}

	// 创建Handler (using Gin)
	host := server.NewWebHost().
		UseGinLogger().
		UseGinRecovery().
		UseExternalAPIRoutes(router.InitRouter).
		ConfigureServices(func() {
			go func() {
			}()
		})

	// 注册Handler
	service.Handle("/", host.GinEngine)

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
