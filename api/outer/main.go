package main

import (
	"fmt"
	"time"

	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/web"
	lg "github.com/micro/go-plugins/logger/logrus/v2"

	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/router"
	"rxcsoft.cn/pit3/api/outer/server"
	"rxcsoft.cn/pit3/api/outer/server/start"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
	"rxcsoft.cn/utils/storage/client"
)

var log = myLogger.New()

// @title PIT3对外服务api接口文档
// @version 1.0
// @description PIT3对外服务api接口文档

// @host localhost:8080
// @BasePath /outer/api/v1

// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
func main() {

	// 根据运行环境创建服务名称
	serviceName := "go.micro.api.outer"

	// 创建服务
	service := web.NewService(
		web.Name(serviceName),
		web.RegisterTTL(time.Second*30),
		web.RegisterInterval(time.Second*10),
	)

	// 设置log出力的配置
	logger.DefaultLogger = lg.NewLogger(lg.WithLogger(log))

	// 初始化配置文件
	utilsServer.Start()

	start.DBStart()

	// micro服务初始化
	service.Init()

	// 创建Handler (using Gin)
	host := server.NewWebHost().
		UseGinLogger().
		UseGinRecovery().
		UseExternalAPIRoutes(router.InitRouter).
		ConfigureServices(func() {
			go func() {
				client.InitStorageClient()
			}()
		})

	// 注册Handler
	service.Handle("/", host.GinEngine)

	// 运行服务
	if err := service.Run(); err != nil {
		loggerx.FatalLog("server start", fmt.Sprintf("server start has error: %v", err))
	}
}
