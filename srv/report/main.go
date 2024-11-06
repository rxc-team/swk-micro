/*
 * @Description:报表main
 * @Author: RXC 廖云江
 * @Date: 2019-08-19 14:25:58
 * @LastEditors: RXC 廖云江
 * @LastEditTime: 2020-03-12 16:18:29
 */

package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	lg "github.com/micro/go-plugins/logger/logrus/v2"

	"rxcsoft.cn/pit3/srv/report/handler"
	"rxcsoft.cn/pit3/srv/report/proto/coldata"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
	"rxcsoft.cn/pit3/srv/report/proto/report"
	"rxcsoft.cn/pit3/srv/report/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func main() {

	service := micro.NewService(
		micro.Name("report"),
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
	report.RegisterReportServiceHandler(service.Server(), new(handler.Report))
	dashboard.RegisterDashboardServiceHandler(service.Server(), new(handler.Dashboard))
	coldata.RegisterColDataServiceHandler(service.Server(), new(handler.ColData))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
