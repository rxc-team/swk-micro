/*
 * @Description:main
 * @Author: RXC 廖云江
 * @Date: 2019-10-18 09:19:31
 * @LastEditors: RXC 廖云江
 * @LastEditTime: 2020-12-08 14:38:00
 */

package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	mlogger "github.com/micro/go-micro/v2/logger"
	lg "github.com/micro/go-plugins/logger/logrus/v2"

	"rxcsoft.cn/pit3/srv/global/handler"
	"rxcsoft.cn/pit3/srv/global/proto/cache"
	"rxcsoft.cn/pit3/srv/global/proto/datapatch"
	"rxcsoft.cn/pit3/srv/global/proto/help"
	types "rxcsoft.cn/pit3/srv/global/proto/help-type"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/global/proto/logger"
	"rxcsoft.cn/pit3/srv/global/proto/mail"
	config "rxcsoft.cn/pit3/srv/global/proto/mail-config"
	"rxcsoft.cn/pit3/srv/global/proto/message"
	"rxcsoft.cn/pit3/srv/global/proto/question"
	"rxcsoft.cn/pit3/srv/global/proto/sequence"
	"rxcsoft.cn/pit3/srv/global/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func main() {

	service := micro.NewService(
		micro.Name("global"),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
	)

	// 设置log出力的配置
	mlogger.DefaultLogger = lg.NewLogger(lg.WithLogger(log))

	// 初始化配置文件
	utilsServer.Start()

	// 初始化配置文件
	server.LoadConfig()

	// 启动DB服务
	server.DBStart()

	// micro服务初始化
	service.Init()

	// 注册handler
	cache.RegisterCacheServiceHandler(service.Server(), new(handler.Cache))
	language.RegisterLanguageServiceHandler(service.Server(), new(handler.Language))
	logger.RegisterLoggerServiceHandler(service.Server(), new(handler.Logger))
	mail.RegisterMailServiceHandler(service.Server(), new(handler.Mail))
	config.RegisterConfigServiceHandler(service.Server(), new(handler.Config))
	types.RegisterTypeServiceHandler(service.Server(), new(handler.Type))
	help.RegisterHelpServiceHandler(service.Server(), new(handler.Help))
	message.RegisterMessageServiceHandler(service.Server(), new(handler.Message))
	question.RegisterQuestionServiceHandler(service.Server(), new(handler.Question))
	sequence.RegisterSequenceServiceHandler(service.Server(), new(handler.Sequence))
	datapatch.RegisterDataPatchServiceHandler(service.Server(), new(handler.DataPatch))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
