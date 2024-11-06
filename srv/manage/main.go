package main

import (
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	lg "github.com/micro/go-plugins/logger/logrus/v2"

	"rxcsoft.cn/pit3/srv/manage/handler"
	"rxcsoft.cn/pit3/srv/manage/proto/access"
	"rxcsoft.cn/pit3/srv/manage/proto/action"
	"rxcsoft.cn/pit3/srv/manage/proto/allow"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/level"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/script"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/manage/server"
	myLogger "rxcsoft.cn/utils/logger"
	utilsServer "rxcsoft.cn/utils/server"
)

var log = myLogger.New()

func main() {

	service := micro.NewService(
		micro.Name("manage"),
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
	user.RegisterUserServiceHandler(service.Server(), new(handler.User))
	app.RegisterAppServiceHandler(service.Server(), new(handler.App))
	customer.RegisterCustomerServiceHandler(service.Server(), new(handler.Customer))
	group.RegisterGroupServiceHandler(service.Server(), new(handler.Group))
	role.RegisterRoleServiceHandler(service.Server(), new(handler.Role))
	backup.RegisterBackupServiceHandler(service.Server(), new(handler.Backup))
	permission.RegisterPermissionServiceHandler(service.Server(), new(handler.Permission))
	action.RegisterActionServiceHandler(service.Server(), new(handler.Action))
	allow.RegisterAllowServiceHandler(service.Server(), new(handler.Allow))
	level.RegisterLevelServiceHandler(service.Server(), new(handler.Level))
	access.RegisterAccessServiceHandler(service.Server(), new(handler.Access))
	script.RegisterScriptServiceHandler(service.Server(), new(handler.Script))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("server start has error: %v", err)
	}
}
