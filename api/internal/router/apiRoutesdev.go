/*
 * @Description:API路由
 * @Author: RXC 廖云江
 * @Date: 2019-08-19 10:23:27
 * @LastEditors: Rxc 陳平
 * @LastEditTime: 2021-02-23 13:49:13
 */

package router

import (
	"github.com/gin-gonic/gin"
	"rxcsoft.cn/pit3/api/internal/handler/common"
	"rxcsoft.cn/pit3/api/internal/handler/dev"
	"rxcsoft.cn/pit3/api/internal/middleware/auth/jwt"
	"rxcsoft.cn/pit3/api/internal/middleware/exist"
	"rxcsoft.cn/pit3/api/internal/middleware/match"
	"rxcsoft.cn/pit3/api/internal/middleware/pv"
)

// 初始化需要验证的路由
func initAuthRouterDev(router *gin.Engine) {
	// 创建组
	v1 := router.Group("/internal/api/v1/dev")

	// 使用jwt校验
	v1.Use(jwt.APIJWTAuth())
	// 身份验证
	v1.Use(pv.DevPV())

	role := new(dev.Role)
	{
		roleRoute := v1.Group("/role")
		roleRoute.PUT("/whitelistClear/roles", role.WhitelistClear)
	}

	user := new(dev.User)
	{
		userRoute := v1.Group("/user")
		// 更新用户
		userRoute.PUT("/users/:user_id", user.ModifyUser)
		// 通过ID获取用户
		userRoute.GET("/users/:user_id", user.FindUser)
	}

	backup := new(dev.Backup)
	{
		backupRoute := v1.Group("/backup")
		// 查找多个备份记录
		backupRoute.GET("/backups", backup.FindBackups)
		// 查找一个备份记录
		backupRoute.GET("/backups/:backup_id", backup.FindBackup)
		// 添加备份记录
		backupRoute.POST("/backups", backup.AddBackup)
		// 删除多个备份记录
		backupRoute.DELETE("/phydel/backups", backup.HardDeleteTemplateBackups)
	}

	// language
	language := new(dev.Language)
	{
		languageRoute := v1.Group("/language")
		// 获取语言数据
		languageRoute.GET("/languages/search", language.FindLanguage)
	}

	// 使用app用户等存在验证
	v1.Use(exist.CheckExist())
	v1.Use(match.Macth())

	// app
	app := new(dev.App)
	{
		appRoute := v1.Group("/app")
		// 查找多个APP记录
		appRoute.GET("/apps", app.FindApps)
		// 通过ID查找单个APP记录
		appRoute.GET("/apps/:a_id", app.FindApp)
		// 添加单个APP记录
		appRoute.POST("/apps", app.AddApp)
		// 修改单个APP记录
		appRoute.PUT("/apps/:a_id", app.ModifyApp)
		// 删除选中的APP记录
		appRoute.DELETE("/apps", app.DeleteSelectApps)
		// 物理删除APP记录
		appRoute.DELETE("/phydel/apps", app.HardDeleteApps)
		// 恢复选中APP记录
		appRoute.PUT("/recover/apps", app.RecoverSelectApps)
	}

	// user
	{
		userRoute := v1.Group("/user")
		// 获取所有用户
		userRoute.GET("/users", user.FindUsers)
		// 获取开发平台默认管理员用户
		userRoute.GET("/default/user", user.FindDefaultUser)
		// 创建用户
		userRoute.POST("/users", user.AddUser)
		// 删除选中用户
		userRoute.DELETE("/users", user.DeleteSelectUsers)
		// 恢复选中用户
		userRoute.PUT("/recover/users", user.RecoverSelectUsers)
		// 恢复被锁用户
		userRoute.PUT("/unlock/users", user.UnlockSelectUsers)
	}

	// role
	{
		roleRoute := v1.Group("/role")
		// 查找多个角色
		roleRoute.GET("/roles", role.FindRoles)
		// 查找单个角色
		roleRoute.GET("/roles/:role_id", role.FindRole)
		// 添加角色
		roleRoute.POST("/roles", role.AddRole)
		// 更新角色
		roleRoute.PUT("/roles/:role_id", role.ModifyRole)
		// 删除选中角色
		roleRoute.DELETE("/roles", role.DeleteSelectRoles)
		// 物理删除选中角色
		roleRoute.DELETE("/phydel/roles", role.HardDeleteRoles)
		// 恢复选中角色
		roleRoute.PUT("/recover/roles", role.RecoverSelectRoles)
	}

	// group
	group := new(dev.Group)
	{
		groupRoute := v1.Group("/group")
		// 获取多个组
		groupRoute.GET("/groups", group.FindGroups)
	}

	mongoScript := new(dev.MongoScript)
	{
		scriptRoute := v1.Group("/script")
		scriptRoute.POST("/run", mongoScript.Run)
	}

	// datapatch脚本
	script := new(dev.Script)
	{
		scriptRoute := v1.Group("/script")
		// 获取job
		scriptRoute.GET("/scripts", script.FindScriptJobs)
		// 获取单个job
		scriptRoute.GET("/scripts/:script_id", script.FindScriptJob)
		// 添加
		scriptRoute.POST("/scripts", script.AddScriptJob)
		// 更新
		scriptRoute.PATCH("/scripts/:script_id", script.ModifyScript)
		// 执行job
		scriptRoute.POST("/scripts/:script_id", script.ExecScriptJob)
	}

	// customer
	customer := new(dev.Customer)
	{
		customerRoute := v1.Group("/customer")
		// 查找多个顾客记录
		customerRoute.GET("/customers", customer.FindCustomers)
		// 查找单个顾客记录
		customerRoute.GET("/customers/:customer_id", customer.FindCustomer)
		// 添加单个顾客记录
		customerRoute.POST("/customers", customer.AddCustomer)
		// 修改单个顾客记录
		customerRoute.PUT("/customers/:customer_id", customer.ModifyCustomer)
		// 删除选中顾客记录
		customerRoute.DELETE("/customers", customer.DeleteSelectCustomers)
		// 物理删除客户
		customerRoute.DELETE("/phydel/customers", customer.HardDeleteCustomers)
		// 恢复选中客户
		customerRoute.PUT("/recover/customers", customer.RecoverSelectCustomers)
	}

	// file
	file := new(dev.File)
	{
		fileRoute := v1.Group("/file")
		// 文件上传
		fileRoute.POST("/folders/:fo_id/upload", file.UploadDev)
		// 头像文件上传
		fileRoute.POST("/header/upload", file.HeaderFileUpload)
		// 删除头像或LOGO文件
		fileRoute.DELETE("/public/header/file", file.DeletePublicHeaderFile)
		// 下载
		fileRoute.GET("/download/folders/:fo_id/files/:file_id", file.Download)
		// 查找多个文件
		fileRoute.GET("/folders/:fo_id/files", file.FindFiles)
		// 硬删除单个文件
		fileRoute.DELETE("/folders/:fo_id/files/:file_id", file.HardDeleteFile)
	}

	// folder
	folder := new(dev.Folder)
	{
		folderRoute := v1.Group("/folder")
		// 查找多个文件夹
		folderRoute.GET("/folders", folder.FindFolders)
		// 添加文件夹
		folderRoute.POST("/folders", folder.AddFolder)
	}

	// language
	{
		languageRoute := v1.Group("/language")
		// 删除app语言数据
		languageRoute.DELETE("/apps/:a_id/languages", language.DeleteLanguageData)
	}

	// Validation
	validation := new(dev.Validation)
	{
		// 验证密码
		v1.POST("/validation/password", validation.PasswordValidation)
		// 验证多语言项目的唯一性
		v1.POST("/validation/unique", validation.UniqueValidation)
		// 验证角色名称唯一性
		v1.POST("/validation/rolename", validation.RoleNameValidation)
		// 验证用户名称或登录ID唯一性
		v1.POST("/validation/user", validation.UserDuplicated)
		// 验证客户名称、域名唯一性
		v1.POST("/validation/customer", validation.CustomerDuplicated)
		// 文件名称唯一性检查
		v1.POST("/validation/filename", validation.FileNameDuplicated)
		// 帮助文章名称唯一性检查
		v1.POST("/validation/helpname", validation.HelpNameDuplicated)
		// 类别名称唯一性检查
		v1.POST("/validation/typename", validation.TypeNameDuplicated)
		// 模板名称唯一性检查
		v1.POST("/validation/backupname", validation.BackUpNameDuplicated)
	}

	// helpType
	helpType := new(dev.Type)
	{
		typeRoute := v1.Group("/type")
		// 获取单个帮助文档类型
		typeRoute.GET("/types/:type_id", helpType.FindType)
		// 获取多个帮助文档类型
		typeRoute.GET("/types", helpType.FindTypes)
		// 添加帮助文档类型
		typeRoute.POST("/types", helpType.AddType)
		// 更新帮助文档类型
		typeRoute.PUT("/types/:type_id", helpType.ModifyType)
		// 硬删除多个帮助文档类型
		typeRoute.DELETE("/types", helpType.DeleteTypes)
	}

	// help
	help := new(dev.Help)
	{
		helpRoute := v1.Group("/help")
		// 获取单个帮助文档
		helpRoute.GET("/helps/:help_id", help.FindHelp)
		// 获取多个帮助文档
		helpRoute.GET("/helps", help.FindHelps)
		// 获取所有不重复帮助文档标签
		helpRoute.GET("/tags", help.FindTags)
		// 添加帮助文档
		helpRoute.POST("/helps", help.AddHelp)
		// 更新帮助文档
		helpRoute.PUT("/helps/:help_id", help.ModifyHelp)
		// 硬删除多个帮助文档
		helpRoute.DELETE("/helps", help.DeleteHelps)
	}

	schedule := new(dev.Schedule)
	{
		taskRoute := v1.Group("/schedule")
		// 获取当前用户的任务一览
		taskRoute.GET("/schedules", schedule.FindDefaultSchedules)
		// 添加任务
		taskRoute.POST("/schedules", schedule.AddSchedule)
		// 添加任务
		taskRoute.PUT("/schedules/:s_id", schedule.ModifySchedule)
		// 添加本地任务
		taskRoute.POST("/schedules/restore", schedule.AddRestoreSchedule)
		// 删除任务
		taskRoute.DELETE("/schedules", schedule.DeleteSchedule)
	}

	// config
	config := new(dev.Config)
	{
		configRoute := v1.Group("/config")
		// 获取邮件配置
		configRoute.GET("/configs/config", config.FindConfig)
		// 添加邮件配置
		configRoute.POST("/configs", config.AddConfig)
		// 更新邮件配置
		configRoute.PUT("/configs/:config_id", config.ModifyConfig)
	}

	// question
	question := new(dev.Question)
	{
		questionRoute := v1.Group("/question")
		// 获取单个问题
		questionRoute.GET("/questions/:question_id", question.FindQuestion)
		// 获取多个问题
		questionRoute.GET("/questions", question.FindQuestions)
		// 添加问题
		questionRoute.POST("/questions", question.AddQuestion)
		// 更新问题
		questionRoute.PUT("/questions/:question_id", question.ModifyQuestion)
		// 硬删除多个问题
		questionRoute.DELETE("/questions", question.DeleteQuestions)
	}

	// action
	action := new(dev.Action)
	{
		actionRoute := v1.Group("/action")
		// 获取单个操作
		actionRoute.GET("objs/:action_object/actions/:action_key", action.FindAction)
		// 获取多个操作
		actionRoute.GET("/actions", action.FindActions)
		// 添加操作
		actionRoute.POST("/actions", action.AddAction)
		// 更新操作
		actionRoute.PUT("objs/:action_object/actions/:action_key", action.ModifyAction)
		// 硬删除多个操作
		actionRoute.PUT("/actions", action.DeleteActions)
	}

	// allow
	allow := new(dev.Allow)
	{
		allowRoute := v1.Group("/allow")
		// 获取单个许可
		allowRoute.GET("/allows/:allow_id", allow.FindAllow)
		// 获取多个许可
		allowRoute.GET("/level/allows", allow.FindLevelAllows)
		// 获取多个许可
		allowRoute.GET("/check/allows", allow.CheckAllow)
		// 获取多个许可
		allowRoute.GET("/allows", allow.FindAllows)
		// 添加许可
		allowRoute.POST("/allows", allow.AddAllow)
		// 更新许可
		allowRoute.PUT("/allows/:allow_id", allow.ModifyAllow)
		// 硬删除多个许可
		allowRoute.DELETE("/allows", allow.DeleteAllows)
	}

	// level
	level := new(dev.Level)
	{
		levelRoute := v1.Group("/level")
		// 获取单个授权等级
		levelRoute.GET("/levels/:level_id", level.FindLevel)
		// 获取多个授权等级
		levelRoute.GET("/levels", level.FindLevels)
		// 添加授权等级
		levelRoute.POST("/levels", level.AddLevel)
		// 更新授权等级
		levelRoute.PUT("/levels/:level_id", level.ModifyLevel)
		// 硬删除多个授权等级
		levelRoute.DELETE("/levels", level.DeleteLevels)
	}

	// message
	message := new(common.Message)
	{
		messageRoute := v1.Group("/message")
		// 获取多个通知
		messageRoute.GET("/messages", message.FindMessages)
		// 添加通知
		messageRoute.POST("/messages", message.AddMessage)
		// 变更状态
		messageRoute.PATCH("/messages/:message_id", message.ChangeStatus)
		// 硬删除通知
		messageRoute.DELETE("/messages/:message_id", message.DeleteMessage)
		// 硬删除多个通知
		messageRoute.DELETE("/messages", message.DeleteMessages)
	}

	task := new(common.Task)
	{
		taskRoute := v1.Group("/task")
		// 获取当前用户的任务一览
		taskRoute.GET("/tasks", task.FindTasks)
		// 获取当前用户的任务一览
		taskRoute.GET("/tasks/:j_id", task.FindTask)
		// 获取当前用户的任务一览
		taskRoute.POST("/tasks", task.AddTask)
		// 获取当前用户的任务一览
		taskRoute.DELETE("/tasks/:j_id", task.DeleteTask)
	}

	taskHistory := new(common.TaskHistory)
	{
		taskRoute := v1.Group("/task")
		// 获取当前用户的任务一览
		taskRoute.GET("/histories", taskHistory.FindTaskHistories)
		// 下载任务履历
		taskRoute.GET("/histories/download", taskHistory.DownloadTaskHistory)
	}

	log := new(common.Log)
	{
		logRoute := v1.Group("/log")
		// 下载日志
		logRoute.GET("/logs", log.DownloadLog)
		// 查找日志
		logRoute.GET("/loggers", log.FindLogs)
	}
}
