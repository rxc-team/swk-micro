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
	"rxcsoft.cn/pit3/api/internal/handler/admin"
	"rxcsoft.cn/pit3/api/internal/handler/common"
	"rxcsoft.cn/pit3/api/internal/middleware/auth/jwt"
	"rxcsoft.cn/pit3/api/internal/middleware/exist"
	"rxcsoft.cn/pit3/api/internal/middleware/match"
	"rxcsoft.cn/pit3/api/internal/middleware/pv"
)

// 初始化需要验证的路由
func initAuthRouterAdm(router *gin.Engine) {
	// 创建组
	v1 := router.Group("/internal/api/v1/admin")

	// 使用jwt校验
	v1.Use(jwt.APIJWTAuth())
	v1.Use(pv.AdminPV())

	user := new(admin.User)
	{
		userRoute := v1.Group("/user")
		// 更新用户
		userRoute.PUT("/users/:user_id", user.ModifyUser)
		// 通过ID获取用户
		userRoute.GET("/users/:user_id", user.FindUser)
	}

	app := new(admin.App)
	{
		appRoute := v1.Group("/app")
		// 通过当前用户查找多个APP记录
		appRoute.GET("/user/apps", app.FindUserApp)
	}

	// language
	language := new(admin.Language)
	{
		languageRoute := v1.Group("/language")
		// 获取语言数据
		languageRoute.GET("/languages/search", language.FindLanguage)
	}

	// 使用app用户等存在验证
	v1.Use(exist.CheckExist())
	v1.Use(match.Macth())

	// app
	{
		appRoute := v1.Group("/app")
		// 查找多个APP记录
		appRoute.GET("/apps", app.FindApps)
		// 通过ID查找单个APP记录
		appRoute.GET("/apps/:a_id", app.FindApp)
		// 添加单个APP记录
		appRoute.POST("/apps", app.CopyApp)
		// 修改单个APP记录
		appRoute.PUT("/apps/:a_id", app.ModifyApp)
		// 修改单个APP的config记录
		appRoute.PUT("/apps/:a_id/configs", app.ModifyAppConfigs)
		// 下一月度处理
		appRoute.PUT("/apps/nextMonth", app.NextMonth)
		// 删除选中的APP记录
		appRoute.DELETE("/apps", app.DeleteSelectApps)
		// 物理删除APP记录
		appRoute.DELETE("/phydel/apps", app.HardDeleteCopyApps)
		// 恢复选中APP记录
		appRoute.PUT("/recover/apps", app.RecoverSelectApps)
	}

	// user
	{
		userRoute := v1.Group("/user")
		// 查找用户组&关联用户组的多个用户记录
		userRoute.GET("/related/users", user.FindRelatedUsers)
		// 获取所有用户
		userRoute.GET("/users", user.FindUsers)
		// 获取公司默认管理员用户
		userRoute.GET("/default/user", user.FindDefaultUser)
		// 创建用户
		userRoute.POST("/users", user.AddUser)
		// 删除选中用户
		userRoute.DELETE("/users", user.DeleteSelectUsers)
		// 恢复选中用户
		userRoute.PUT("/recover/users", user.RecoverSelectUsers)
		// 恢复被锁用户
		userRoute.PUT("/unlock/users", user.UnlockSelectUsers)
		// 批量上传用户
		userRoute.POST("/upload/users", user.UploadUsers)
		// 批量下载用户
		userRoute.POST("/download/users", user.DownloadUsers)
	}

	// role
	role := new(admin.Role)
	{
		roleRoute := v1.Group("/role")
		// 查找多个角色
		roleRoute.GET("/roles", role.FindRoles)
		// 查找单个角色
		roleRoute.GET("/roles/:role_id", role.FindRole)
		// 查找单个角色的action
		roleRoute.GET("/roles/:role_id/actions", role.FindRoleActions)
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
	group := new(admin.Group)
	{
		groupRoute := v1.Group("/group")
		// 获取多个组
		groupRoute.GET("/groups", group.FindGroups)
		// 获取单个组
		groupRoute.GET("/groups/:group_id", group.FindGroup)
		// 添加组
		groupRoute.POST("/groups", group.AddGroup)
		// 更新组
		groupRoute.PUT("/groups/:group_id", group.ModifyGroup)
		// 物理删除选中组
		groupRoute.DELETE("/phydel/groups", group.HardDeleteGroups)
	}

	// customer
	customer := new(admin.Customer)
	{
		customerRoute := v1.Group("/customer")
		// 查找多个顾客记录
		customerRoute.GET("/customers", customer.FindCustomers)
		// 查找单个顾客记录
		customerRoute.GET("/customers/:customer_id", customer.FindCustomer)
		// 修改单个顾客记录
		customerRoute.PUT("/customers/:customer_id", customer.ModifyCustomer)
	}

	// file
	file := new(admin.File)
	{
		fileRoute := v1.Group("/file")
		// 文件上传
		fileRoute.POST("/folders/:fo_id/upload", file.UploadAdmin)
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
	folder := new(admin.Folder)
	{
		folderRoute := v1.Group("/folder")
		// 查找多个文件夹
		folderRoute.GET("/folders", folder.FindFolders)
		// 添加文件夹
		folderRoute.POST("/folders", folder.AddFolder)
		// 物理删除多个文件夹
		folderRoute.DELETE("/phydel/folders", folder.HardDeleteFolders)
	}

	// language
	{
		languageRoute := v1.Group("/language")
		// 添加APP语言数据
		languageRoute.POST("/apps/:a_id/languages/:lang_cd", language.AddLanguageData)
		// 添加APP中的语言数据
		languageRoute.POST("/languages/:lang_cd", language.AddAppLanguageData)
		// 添加Common语言数据
		languageRoute.POST("/common/languages/:lang_cd", language.AddCommonData)
		// 添加或更新多条语言数据
		languageRoute.POST("/import/csv", language.AddManyLanData)
		// // 删除App中的语言数据
		// languageRoute.DELETE("/languages/types/:type/keys/:key", language.DeleteAppLanguageData)
		// // 删除app语言数据
		// languageRoute.DELETE("/apps/:a_id/languages", language.DeleteLanguageData)
	}

	// Validation
	validation := new(admin.Validation)
	{
		// 验证密码
		v1.POST("/validation/password", validation.PasswordValidation)
		// 验证多语言项目的唯一性
		v1.POST("/validation/unique", validation.UniqueValidation)
		// 验证应用下属唯一性字段数值唯一性
		v1.GET("/validation/apps/:app_id/value/unique", validation.ValueUniqueValidation)
		// 验证台账的字段是否有被引用的报表
		v1.GET("/validation/datastores/:id/fields/:f_id/relation", validation.FieldRelationValidation)
		// 验证选项值的唯一性
		v1.POST("/validation/option", validation.OptionValueUinqueValidation)
		// 验证客户名称唯一性
		v1.POST("/validation/customer", validation.CustomerNameUinqueValidation)
		// 验证台账apiKey唯一性
		v1.POST("/validation/datastoreapikey", validation.DatastoreApiKeyUinqueValidation)
		// 验证字段ID唯一性
		v1.POST("/validation/field", validation.FieldIDUinqueValidation)
		// 验证文件名称唯一性
		v1.POST("/validation/filename", validation.FileNameUinqueValidation)
		// 验证文件夹名称唯一性
		v1.POST("/validation/foldername", validation.FolderNameDuplicated)
		// 验证问题标题唯一性
		v1.POST("/validation/questiontitle", validation.QuestionTitleDuplicated)
		// 验证角色名称唯一性
		v1.POST("/validation/rolename", validation.RoleNameDuplicated)
		// 验证用户名称或登录ID唯一性
		v1.POST("/validation/user", validation.UserDuplicated)
	}

	// option
	option := new(admin.Option)
	{
		optionRoute := v1.Group("/option")
		// 获取所有选项
		optionRoute.GET("/options", option.FindOptions)
		optionRoute.GET("/label/options", option.FindOptionLabels)
		// 通过选项ID获取选项
		optionRoute.GET("/options/:o_id", option.FindOption)
		// 添加选项
		optionRoute.POST("/options", option.AddOption)
		// 删除选中选项
		optionRoute.DELETE("/options", option.DeleteSelectOptions)
		// 物理删除选中选项
		optionRoute.DELETE("/phydel/options", option.HardDeleteOptions)
		// 通过选项ID和选项值删除选项值数据
		optionRoute.DELETE("/options/:o_id/values/:value", option.DeleteOptionChild)
		// 通过选项ID和选项值物理删除选项值数据
		optionRoute.DELETE("/phydel/options/:o_id/values/:value", option.HardDeleteOptionChild)
		// 恢复选中选项
		optionRoute.PUT("/recover/options", option.RecoverSelectOptions)
		// 恢复选中选项
		optionRoute.PUT("/recover/options/:o_id/values/:value", option.RecoverOptionChild)
		// csv下载所有选项
		optionRoute.GET("/download/options", option.DownloadCSVOptions)

	}

	// datastore
	datastores := new(admin.Datastore)
	{
		datastoreRoute := v1.Group("/datastore")
		// 查找多个台账
		datastoreRoute.GET("/datastores", datastores.FindDatastores)
		// 查找单个台账
		datastoreRoute.GET("/datastores/:d_id", datastores.FindDatastore)
		// 查找单个台账
		datastoreRoute.GET("/key/datastores/:api_key", datastores.FindDatastoreByKey)
		// 添加台账
		datastoreRoute.POST("/datastores", datastores.AddDatastore)
		// 更新台账
		datastoreRoute.PUT("/datastores/:d_id", datastores.ModifyDatastore)
		// 更新台账菜单排序
		datastoreRoute.PUT("/datastores/sort", datastores.ModifyDatastoreSort)
		// 物理删除多个台账
		datastoreRoute.DELETE("/phydel/datastores", datastores.HardDeleteDatastores)
		// 查找单个台账mapping
		datastoreRoute.GET("/datastores/:d_id/mappings/:m_id", datastores.FindDatastoreMapping)
		// 添加台账mapping
		datastoreRoute.POST("/datastores/:d_id/mappings", datastores.AddDatastoreMapping)
		// 更新台账mapping
		datastoreRoute.PUT("/datastores/:d_id/mappings/:m_id", datastores.ModifyDatastoreMapping)
		// 删除单个台账mapping
		datastoreRoute.DELETE("/datastores/:d_id/mappings/:m_id", datastores.DeleteDatastoreMapping)
		// 添加台账unique
		datastoreRoute.POST("/datastores/:d_id/unique", datastores.AddUniqueKey)
		// 删除台账unique
		datastoreRoute.DELETE("/datastores/:d_id/unique/:fields", datastores.DeleteUniqueKey)
		// 添加台账relation
		datastoreRoute.GET("/datastores/:d_id/relation", datastores.FindDatastoreRelations)
		// 添加台账relation
		datastoreRoute.POST("/datastores/:d_id/relation", datastores.AddRelation)
		// 删除台账relation
		datastoreRoute.DELETE("/datastores/:d_id/relation/:r_id", datastores.DeleteRelation)
	}

	// report
	reports := new(admin.Report)
	{
		reportRoute := v1.Group("/report")
		// 获取所属公司所属APP下所有报表情报
		reportRoute.GET("/reports", reports.FindReports)
		// 通过报表ID获取单个报表情报
		reportRoute.GET("/reports/:rp_id", reports.FindReport)
		// 通过报表ID获取报表数据
		reportRoute.POST("/reports/:rp_id/data", reports.FindReportData)
		// 通过报表ID获取报表数据
		reportRoute.POST("/gen/reports/:rp_id/data", reports.GenerateReportData)
		// 添加单个报表情报
		reportRoute.POST("/reports", reports.AddReport)
		// 更新单个报表情报
		reportRoute.PUT("/reports/:rp_id", reports.ModifyReport)
		// 物理删除多个报表情报
		reportRoute.DELETE("/phydel/reports", reports.HardDeleteReports)
	}

	// dashboard
	dashboards := new(admin.Dashboard)
	{
		dashboardRoute := v1.Group("/dashboard")
		// 获取所属公司所属APP所属报表下所有仪表盘情报
		dashboardRoute.GET("/dashboards", dashboards.FindDashboards)
		// 通过仪表盘ID获取单个仪表盘情报
		dashboardRoute.GET("/dashboards/:dash_id", dashboards.FindDashboard)
		// 添加单个仪表盘情报
		dashboardRoute.POST("/dashboards", dashboards.AddDashboard)
		// 更新单个仪表盘情报
		dashboardRoute.PUT("/dashboards/:dash_id", dashboards.ModifyDashboard)
		// 物理删除多个仪表盘情报
		dashboardRoute.DELETE("/phydel/dashboards", dashboards.HardDeleteDashboards)
	}

	// Item
	items := new(admin.Item)
	{
		itemRoute := v1.Group("/item")
		// 获取台账所有数据
		itemRoute.POST("/datastores/:d_id/items/search", items.FindItems)
		// 更新所有者
		itemRoute.PATCH("/datastores/:d_id/items", items.ChangeOwners)
		// 盘点台账盘点数据盘点状态重置
		itemRoute.PATCH("/apps/:app_id/inventory/reset", items.ResetInventoryItems)
	}

	// approve
	approve := new(admin.Approve)
	{
		templateRoute := v1.Group("/approve")
		// 获取台账所有数据
		templateRoute.GET("/approves", approve.FindApproveCount)
	}

	// field
	field := new(admin.Field)
	{
		fieldRoute := v1.Group("/field")
		// 获取APP中所有台账中的所有字段
		fieldRoute.GET("/app/fields", field.FindAppFields)
		// 获取台账中所有的字段
		fieldRoute.GET("/datastores/:d_id/fields", field.FindFields)
		// 获取一个字段
		fieldRoute.GET("/fields/:f_id", field.FindField)
		// 验证函数是否正确
		fieldRoute.POST("/func/verify", field.VerifyFunc)
		// 添加一个字段
		fieldRoute.POST("/datastores/:d_id/fields", field.AddField)
		// 更新字段
		fieldRoute.PUT("/fields/:f_id", field.ModifyField)
		// 删除选中的字段
		fieldRoute.DELETE("/fields", field.DeleteSelectFields)
		// 物理删除选中字段
		fieldRoute.DELETE("/phydel/datastores/:d_id/fields", field.HardDeleteFields)
		// 恢复选中字段
		fieldRoute.PUT("/recover/fields", field.RecoverSelectFields)
	}

	schedule := new(admin.Schedule)
	{
		taskRoute := v1.Group("/schedule")
		// 获取当前用户的任务一览
		taskRoute.GET("/schedules", schedule.FindSchedules)
		// 添加任务
		taskRoute.POST("/schedules", schedule.AddSchedule)
		// 删除任务
		taskRoute.DELETE("/schedules", schedule.DeleteSchedule)
	}

	// helpType
	helpType := new(admin.Type)
	{
		typeRoute := v1.Group("/type")
		// 获取单个帮助文档类型
		typeRoute.GET("/types/:type_id", helpType.FindType)
		// 获取多个帮助文档类型
		typeRoute.GET("/types", helpType.FindTypes)
	}

	// help
	help := new(admin.Help)
	{
		helpRoute := v1.Group("/help")
		// 获取单个帮助文档
		helpRoute.GET("/helps/:help_id", help.FindHelp)
		// 获取多个帮助文档
		helpRoute.GET("/helps", help.FindHelps)
	}

	// question
	question := new(admin.Question)
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

	// process
	process := new(admin.Process)
	{
		processRoute := v1.Group("/process")
		// 获取所有进程
		processRoute.GET("/processes/:user_id", process.FindsProcess)
		// 获取所有节点
		processRoute.GET("/node", process.FindNodes)
	}
	// workflow
	workflow := new(admin.Workflow)
	{
		workflowRoute := v1.Group("/workflow")
		// 获取单个流程
		workflowRoute.GET("/workflows/:wf_id", workflow.FindWorkflow)
		// 获取多个流程
		workflowRoute.GET("/workflows", workflow.FindWorkflows)
		// 添加流程
		workflowRoute.POST("/workflows", workflow.AddWorkflow)
		// 更新流程
		workflowRoute.PUT("/workflows/:wf_id", workflow.ModifyWorkflow)
		// 硬删除流程
		workflowRoute.DELETE("/workflows", workflow.DeleteWorkflow)
	}

	// allow
	allow := new(admin.Allow)
	{
		allowRoute := v1.Group("/allow")
		// 获取多个许可
		allowRoute.GET("/level/allows", allow.FindLevelAllows)
	}

	// print
	print := new(admin.Print)
	{
		printRoute := v1.Group("/print")
		// 获取台账打印设置
		printRoute.GET("/prints/:datastore_id", print.FindPrint)
		// 添加台账打印设置
		printRoute.POST("/prints", print.AddPrint)
		// 更新台账打印设置
		printRoute.PUT("/prints/:datastore_id", print.ModifyPrint)
	}

	// access
	access := new(admin.Access)
	{
		accessRoute := v1.Group("/ace")
		// 查找多个access
		accessRoute.GET("/access", access.FindAccess)
		// 查找单个access
		accessRoute.GET("/access/:access_id", access.FindOneAccess)
		// 添加access
		accessRoute.POST("/access", access.AddAccess)
		// 删除选中access
		accessRoute.DELETE("/access", access.DeleteSelectAccess)
		// 物理删除选中access
		accessRoute.DELETE("/phydel/access", access.HardDeleteAccess)
	}
	// access
	generate := new(admin.Generate)
	{
		generateRoute := v1.Group("/gen")
		// 文件上传
		generateRoute.POST("/upload", generate.Upload)
		// 获取配置
		generateRoute.GET("/config", generate.FindConfig)
		// 查找上传数据
		generateRoute.GET("/row/data", generate.FindRowData)
		// 设置台账
		generateRoute.POST("/database/set", generate.DatabaseSet)
		// 设置字段
		generateRoute.POST("/field/set", generate.FiledSet)
		// 查找列数据
		generateRoute.GET("/column/data", generate.FindColumnData)
		// 创建台账等信息
		generateRoute.POST("/database/create", generate.CreateDatabase)
		// 创建mapping
		generateRoute.POST("/mapping/create", generate.CreateMapping)
		// 完成
		generateRoute.POST("/complete", generate.Complete)
	}

	// message
	message := new(common.Message)
	{
		messageRoute := v1.Group("/message")
		// 获取多个通知
		messageRoute.GET("/messages", message.FindMessages)
		// 获取系统更新通知
		messageRoute.GET("/messages/update", message.FindUpdateMessage)
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
