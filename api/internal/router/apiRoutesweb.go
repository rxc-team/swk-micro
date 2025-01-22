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
	"rxcsoft.cn/pit3/api/internal/handler/webui"
	"rxcsoft.cn/pit3/api/internal/middleware/auth/jwt"
	"rxcsoft.cn/pit3/api/internal/middleware/exist"
	"rxcsoft.cn/pit3/api/internal/middleware/match"
	"rxcsoft.cn/pit3/api/internal/middleware/pv"
)

// 初始化需要验证的路由
func initAuthRouterWeb(router *gin.Engine) {
	// 创建组
	v1 := router.Group("/internal/api/v1/web")

	// 使用jwt校验
	v1.Use(jwt.APIJWTAuth())
	// 身份验证
	v1.Use(pv.WebPV())

	user := new(webui.User)
	{
		userRoute := v1.Group("/user")
		// 更新用户
		userRoute.PUT("/users/:user_id", user.ModifySelf)
		// 通过ID获取用户
		userRoute.GET("/users/:user_id", user.FindUser)
	}

	app := new(webui.App)
	{
		appRoute := v1.Group("/app")
		// 通过当前用户查找多个APP记录
		appRoute.GET("/user/apps", app.FindUserApp)
	}

	// language
	language := new(webui.Language)
	{
		languageRoute := v1.Group("/language")
		// 获取语言数据
		languageRoute.GET("/languages/search", language.FindLanguage)
	}

	// 使用app用户等存在验证
	v1.Use(exist.CheckExist())
	v1.Use(match.Macth())

	/* v1.Use(casbin.CheckAction()) */

	journal := new(webui.Journal)
	{
		journalRoute := v1.Group("/journal")
		// 查找多个分录记录
		journalRoute.GET("/journals", journal.FindJournals)
		// 导入分录记录
		journalRoute.POST("/journals", journal.ImportJournal)
		// 计算分录
		journalRoute.GET("/compute/journals", journal.JournalCompute)
		// 查找分录作成数据
		journalRoute.GET("/journals/findSakuseiData", journal.FindSakuseiData)
		// 分录确定
		journalRoute.GET("/journals/confim", journal.JournalConfim)
		// 修改分录记录
		journalRoute.PUT("/journals/:j_id", journal.ModifyJournal)
		// 添加分录下载设置
		journalRoute.POST("/download/setting", journal.AddDownloadSetting)
		// 查询分录下载设置
		journalRoute.GET("/download/find", journal.FindDownloadSetting)
		// 查询分录下载设置
		journalRoute.GET("/download/findSettings", journal.FindDownloadSettings)
		// 分录下载
		journalRoute.GET("/download", journal.SwkDownload)
	}

	subject := new(webui.Subject)
	{
		subjectRoute := v1.Group("/subject")
		// 查找多个科目记录
		subjectRoute.GET("/subjects", subject.FindSubjects)
		// 导入科目记录
		subjectRoute.POST("/subjects", subject.ImportSubject)
		// 修改科目记录
		subjectRoute.PUT("/subjects/:s_key", subject.ModifySubject)
	}

	condition := new(webui.Condition)
	{
		conditionRoute := v1.Group("/condition")
		// 添加条件
		conditionRoute.POST("/conditions", condition.AddCondition)
		// 查找多个条件
		conditionRoute.GET("/conditions", condition.FindConditions)
	}

	// app
	{
		appRoute := v1.Group("/app")
		// 通过ID查找单个APP记录
		appRoute.GET("/apps/:a_id", app.FindApp)
		// 更新基本设定
		appRoute.PUT("/apps/:a_id/swkSetting", app.ModifySwkSetting)
	}

	// user
	{
		userRoute := v1.Group("/user")
		// 查找用户组&关联用户组的多个用户记录
		userRoute.GET("/related/users", user.FindRelatedUsers)
		// 获取所有用户
		userRoute.GET("/users", user.FindUsers)
		// 判断用户的按钮权限
		userRoute.GET("/check/actions/:key", user.CheckAction)
	}

	// role
	role := new(webui.Role)
	{
		roleRoute := v1.Group("/role")
		// 查找多个角色
		roleRoute.GET("/roles", role.FindRoles)
		// 查找单个角色
		roleRoute.GET("/roles/:role_id", role.FindRole)
		// 查找单个角色
		roleRoute.GET("/user/roles/actions", role.FindUserActions)
	}

	// group
	group := new(webui.Group)
	{
		groupRoute := v1.Group("/group")
		// 获取多个组
		groupRoute.GET("/groups", group.FindGroups)
	}

	// file
	file := new(webui.File)
	{
		fileRoute := v1.Group("/file")
		// 文件上传
		fileRoute.POST("/folders/:fo_id/upload", file.Upload)
		// 台账数据中的文件字段上传数据
		fileRoute.POST("/item/upload", file.ItemUpload)
		// 头像文件上传
		fileRoute.POST("/header/upload", file.HeaderFileUpload)
		// 删除头像或LOGO文件
		fileRoute.DELETE("/public/header/file", file.DeletePublicHeaderFile)
		// 删除文件类型字段数据的文件
		fileRoute.DELETE("/public/data/file", file.DeletePublicDataFile)
		// 删除多个文件类型字段数据的文件
		fileRoute.DELETE("/public/data/files", file.DeletePublicDataFiles)
		// 下载
		fileRoute.GET("/download/folders/:fo_id/files/:file_id", file.Download)
		// 查找多个文件
		fileRoute.GET("/folders/:fo_id/files", file.FindFiles)
		// 硬删除单个文件
		fileRoute.DELETE("/folders/:fo_id/files/:file_id", file.HardDeleteFile)
		// 拷贝文件类型字段单个数据文件
		fileRoute.GET("/public/data/file/copy", file.CopyPublicDataFile)
	}

	// folder
	folder := new(webui.Folder)
	{
		folderRoute := v1.Group("/folder")
		// 查找多个文件夹
		folderRoute.GET("/folders", folder.FindFolders)
		// 添加文件夹
		folderRoute.POST("/folders", folder.AddFolder)
		// 物理删除多个文件夹
		folderRoute.DELETE("/phydel/folders", folder.HardDeleteFolders)
	}

	// query
	query := new(webui.Query)
	{
		queryRoute := v1.Group("/query")
		// 查找多个query
		queryRoute.GET("/queries", query.FindQueries)
		// 查找单个query
		queryRoute.GET("/queries/:q_id", query.FindQuery)
		// 添加单个query
		queryRoute.POST("/queries", query.AddQuery)
		// 删除单个query
		queryRoute.DELETE("/queries/:q_id", query.DeleteQuery)
	}

	// Validation
	validation := new(webui.Validation)
	{
		// 验证密码
		v1.POST("/validation/password", validation.PasswordValidation)
		// 验证数据唯一
		v1.POST("/validation/datastores/:id/items/unique", validation.ItemUniqueValidation)
		// 验证映射数据是否有流程
		v1.GET("/validation/datastores/:d_id/mappings/:m_id", validation.WorkflowExistValidation)
		// 验证特殊字符
		v1.POST("/validation/specialchar", validation.ValidSpecialChar)
		// 验证快捷方式名称唯一性
		v1.POST("/validation/queryname", validation.QueryNameDuplicated)
		// 验证文件名称唯一性
		v1.POST("/validation/filename", validation.FileNameDuplicated)
		// 验证问题标题唯一性
		v1.POST("/validation/questiontitle", validation.QuestionTitleDuplicated)
	}

	// template
	template := new(webui.Template)
	{
		tempRoute := v1.Group("/template")
		// 查询临时台账数据（可分页）
		tempRoute.GET("/templates", template.FindTemplateItems)
		// 删除临时台账数据
		tempRoute.DELETE("/templates/:template_id", template.DeleteTemplateItems)
	}

	// option
	option := new(webui.Option)
	{
		optionRoute := v1.Group("/option")
		// 获取所有选项
		optionRoute.GET("/options", option.FindOptions)
		optionRoute.GET("/label/options", option.FindOptionLabels)
		// 通过选项ID获取选项
		optionRoute.GET("/options/:o_id", option.FindOption)
	}

	// history
	history := new(webui.History)
	{
		historyRoute := v1.Group("/history")
		// 获取所有履历
		historyRoute.GET("/datastores/:d_id/histories", history.FindHistories)
		// 获取最新履历
		historyRoute.GET("/last/datastores/:d_id/histories", history.FindLastHistories)
		// 获取一条履历记录
		historyRoute.GET("/datastores/:d_id/histories/:h_id", history.FindHistory)
		// 履历下载
		historyRoute.GET("/datastores/:d_id/download", history.HistoryDownload)
	}

	// checkHistory
	checkHistory := new(webui.CheckHistory)
	{
		historyRoute := v1.Group("/checkhistory")
		// 获取所有履历
		historyRoute.POST("/datastores/:d_id/histories", checkHistory.FindHistories)
		// 获取所有履历
		historyRoute.POST("/datastores/:d_id/download", checkHistory.HistoryDownload)
	}

	// datastore
	datastores := new(webui.Datastore)
	{
		datastoreRoute := v1.Group("/datastore")
		// 查找多个台账
		datastoreRoute.GET("/datastores", datastores.FindDatastores)
		// 查找单个台账
		datastoreRoute.GET("/datastores/:d_id", datastores.FindDatastore)
		// 查找单个台账
		datastoreRoute.GET("/key/datastores/:api_key", datastores.FindDatastoreByKey)
	}

	// report
	reports := new(webui.Report)
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
		// 通过报表ID获取报表数据
		reportRoute.POST("/reports/:rp_id/download", reports.Download)
		// 获取总表所有数据
		reportRoute.POST("/colData", reports.FindColDatas)
		// 通过契约番号，年月搜索总表数据
		reportRoute.POST("/selectColData", reports.SelectColData)
		// 通过d_id和app_id创建总表数据
		reportRoute.POST("/:rp_id/create", reports.CreateColData)
		// 根据条件CSV下载总表数据
		reportRoute.GET("/downloadColData", reports.ColDataDownload)
	}

	// report 特殊报表的下载
	prs := new(webui.Prs)
	{
		reportRoute := v1.Group("/item")
		// 作成租赁物件本金返还预计表(明细表)数据,以csv文件下载
		reportRoute.POST("/datastores/:d_id/prs/download", prs.DownloadPrs)
	}

	// dashboard
	dashboards := new(webui.Dashboard)
	{
		dashboardRoute := v1.Group("/dashboard")
		// 获取所属公司所属APP所属报表下所有仪表盘情报
		dashboardRoute.GET("/dashboards", dashboards.FindDashboards)
		// 通过仪表盘ID获取仪表盘数据情报
		dashboardRoute.GET("/dashboards/:dash_id/data", dashboards.FindDashboardData)
	}

	// Item
	items := new(webui.Item)
	{
		itemRoute := v1.Group("/item")
		// 获取台账所有数据
		itemRoute.POST("/datastores/:d_id/items/search", items.FindItems)
		// 获取台账所有数据
		itemRoute.POST("/datastores/:d_id/items/print", items.PrintList)
		// 获取台账所有数据,以csv文件下载
		itemRoute.POST("/datastores/:d_id/items/download", items.Download)
		// 获取台账一条数据
		itemRoute.GET("/datastores/:d_id/items/:i_id", items.FindItem)
		// 获取台账一条数据(利子率マスタ)
		itemRoute.GET("/datastores/:d_id/rishiritsu", items.FindRishiritsu)
		// 查询台账未审批数据件数
		itemRoute.GET("/datastores/:d_id/unApprove", items.FindUnApproveItems)
		// 添加台账数据
		itemRoute.POST("/datastores/:d_id/items", items.AddItem)
		// 上传csv文件，导入台账数据
		itemRoute.POST("/import/csv/datastores/:d_id/items", items.ImportCsvItem)
		// 上传csv文件，批量盘点台账数据
		itemRoute.POST("/import/csv/datastores/:d_id/check/items", items.ImportCheckItems)
		// 上传csv文件，导入台账数据
		itemRoute.POST("/import/image/datastores/:d_id/items", items.ImportCsvItem)
		// 更新台账数据
		itemRoute.PUT("/datastores/:d_id/items/:i_id", items.ModifyItem)
		// 更新当前条件下的数据的所有者
		itemRoute.POST("/datastores/:d_id/items/owners", items.ChangeSelectOwners)
		// 更新当前itemid条件下的数据的所有者
		itemRoute.POST("/datastores/:d_id/item/owner", items.ChangeItemOwner)
		// 删除单条台账数据
		itemRoute.DELETE("/datastores/:d_id/items/:i_id", items.DeleteItem)
		// 删除该台账的所有数据记录
		itemRoute.DELETE("/clear/datastores/:d_id/items", items.DeleteDatastoreItems)
		// 删除契约台账所有数据记录
		itemRoute.DELETE("/clear/datastores/clearAll", items.DeleteAllDatastoreItems)
		// 删除该台账选中的数据记录
		itemRoute.DELETE("/clear/datastores/:d_id/items/selected", items.DeleteSelectedItems)
		// 修改标签出力时间
		itemRoute.PUT("/changeLabel/datastores/:d_id/items", items.ChangeLabelTime)
		// 生成支付数据(租赁系统用)
		itemRoute.POST("/generate/pay", items.GeneratePay)
		// 计算利息和偿还数据(租赁系统用)
		itemRoute.POST("/compute/leaserepay", items.ComputeLeaserepay)
		// 债务变更
		itemRoute.PUT("/datastores/:d_id/items/:i_id/debt", items.ChangeDebt)
		// 契约满了
		itemRoute.PUT("/datastores/:d_id/items/:i_id/contractExpire", items.ContractExpire)
		// 契约情报变更
		itemRoute.PUT("/datastores/:d_id/items/:i_id/contract", items.ModifyContract)
		// 中途解约
		itemRoute.PUT("/datastores/:d_id/items/:i_id/terminate", items.TerminateContract)
	}

	// approve
	approve := new(webui.Approve)
	{
		approveRoute := v1.Group("/approve")
		// 获取台账所有数据
		approveRoute.POST("/approves", approve.FindApproveItems)
		// 审批日志下载
		approveRoute.POST("/approves/log/download", approve.ApproveLogDownload)
		// 获取台账一条数据
		approveRoute.GET("/approves/:ex_id", approve.FindApproveItem)
		// 删除单条台账数据field
		approveRoute.DELETE("/approves", approve.DeleteApprove)
	}

	// field
	field := new(webui.Field)
	{
		fieldRoute := v1.Group("/field")
		// 获取APP中所有台账中的所有字段
		fieldRoute.GET("/app/fields", field.FindAppFields)
		// 获取台账中所有的字段
		fieldRoute.GET("/datastores/:d_id/fields", field.FindFields)
	}

	// imp
	mapping := new(webui.Mapping)
	{
		mappingRoute := v1.Group("/mapping")
		// 导入数据
		mappingRoute.POST("/datastores/:d_id/upload", mapping.MappingUpload)
		// 导出数据
		mappingRoute.POST("/datastores/:d_id/download", mapping.MappingDownload)
	}

	// helpType
	helpType := new(webui.Type)
	{
		typeRoute := v1.Group("/type")
		// 获取单个帮助文档类型
		typeRoute.GET("/types/:type_id", helpType.FindType)
		// 获取多个帮助文档类型
		typeRoute.GET("/types", helpType.FindTypes)
	}

	// help
	help := new(webui.Help)
	{
		helpRoute := v1.Group("/help")
		// 获取单个帮助文档
		helpRoute.GET("/helps/:help_id", help.FindHelp)
		// 获取多个帮助文档
		helpRoute.GET("/helps", help.FindHelps)
	}

	// question
	question := new(webui.Question)
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
	// workflow
	workflow := new(webui.Workflow)
	{
		workflowRoute := v1.Group("/workflow")
		// 获取单个流程
		workflowRoute.GET("/workflows/:wf_id", workflow.FindWorkflow)
		// 获取多个流程
		workflowRoute.GET("/workflows", workflow.FindWorkflows)
		// 获取多个流程
		workflowRoute.GET("/user/workflows", workflow.FindUserWorkflows)
		// 承认
		workflowRoute.POST("/admit", workflow.Admit)
		// 拒绝
		workflowRoute.POST("/dismiss", workflow.Dismiss)
	}

	// allow
	allow := new(webui.Allow)
	{
		allowRoute := v1.Group("/allow")
		// 获取多个许可
		allowRoute.GET("/check/allows", allow.CheckAllow)
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
}
