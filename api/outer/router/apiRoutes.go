package router

import (
	"github.com/gin-gonic/gin"

	"rxcsoft.cn/pit3/api/outer/handler"
	"rxcsoft.cn/pit3/api/outer/middleware/auth/jwt"
	"rxcsoft.cn/pit3/api/outer/middleware/exist"
)

// InitRouter 初始化路由
func InitRouter(router *gin.Engine) error {

	// 初始化无验证的路由
	initUnAuthRouter(router)
	// 初始化需要验证的路由
	initAuthRouter(router)

	return nil
}

// 初始化无验证的路由
func initUnAuthRouter(router *gin.Engine) {

	//ping使用
	router.GET("/outer/api/v1/ping", func(c *gin.Context) {
		c.JSON(200, "ok")
	})
	// 登陆
	auth := new(handler.Auth)
	router.POST("/outer/api/v1/login", auth.Login)
}

// 初始化需要验证的路由
func initAuthRouter(router *gin.Engine) {
	// 创建组
	v1 := router.Group("/outer/api/v1")

	// 使用jwt校验
	v1.Use(jwt.APIJWTAuth())

	user := new(handler.User)
	{
		userRoute := v1.Group("/user")
		// 更新用户
		userRoute.PUT("/users/:user_id", user.ModifyUser)
	}

	app := new(handler.App)
	{
		appRoute := v1.Group("/app")
		// 通过当前用户查找多个APP记录
		appRoute.GET("/user/apps", app.FindUserApp)
	}

	// language
	language := new(handler.Language)
	{
		languageRoute := v1.Group("/language")
		// 获取语言数据
		languageRoute.GET("/languages/search", language.FindLanguage)
	}

	// 使用存在验证
	v1.Use(exist.CheckExist())

	// user
	// user := new(handler.User)
	{
		userRoute := v1.Group("/user")
		// 获取所有用户
		userRoute.GET("/users", user.FindUsers)
		// 通过ID获取用户
		userRoute.GET("/users/:user_id", user.FindUser)
		// 判断用户的按钮权限
		userRoute.GET("/check/actions/:key", user.CheckAction)
	}

	// group
	group := new(handler.Group)
	{
		groupRoute := v1.Group("/group")
		// 获取多个组
		groupRoute.GET("/groups", group.FindGroups)
		// 获取单个组
		groupRoute.GET("/groups/:group_id", group.FindGroup)
	}

	// file
	file := new(handler.File)
	{
		fileRoute := v1.Group("/file")
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
	}

	// option
	option := new(handler.Option)
	{
		optionRoute := v1.Group("/option")
		// 获取所有选项
		optionRoute.GET("/options", option.FindOptions)
		optionRoute.GET("/label/options", option.FindOptionLabels)
		// 通过选项ID获取选项
		optionRoute.GET("/options/:o_id", option.FindOption)
	}

	// Validation
	validation := new(handler.Validation)
	{
		// 验证台账数据唯一性
		v1.POST("/validation/datastores/:id/items/unique", validation.ItemUniqueValidation)
		// 验证更新字段是否有流程
		v1.POST("/validation/datastores/:id/updateCheck", validation.UpdateFieldsValidation)
		// 验证特殊字符
		v1.POST("/validation/specialchar", validation.ValidSpecialChar)
	}

	// datastore
	datastores := new(handler.Datastore)
	{
		datastoreRoute := v1.Group("/datastore")
		// 查找多个台账
		datastoreRoute.GET("/datastores", datastores.FindDatastores)
		// 查找单个台账
		datastoreRoute.GET("/datastores/:d_id", datastores.FindDatastore)
	}

	// Item
	items := new(handler.Item)
	{
		itemRoute := v1.Group("/item")
		// 获取台账所有数据
		itemRoute.POST("/datastores/:d_id/items/search", items.FindItems)
		// 获取台账一条数据
		itemRoute.GET("/datastores/:d_id/items/:i_id", items.FindItem)
		// 添加台账数据
		itemRoute.POST("/datastores/:d_id/items", items.AddItem)
		// 上传csv文件，导入台账数据
		itemRoute.POST("/import/csv/datastores/:d_id/items", items.ImportCsvItem)
		// 更新台账数据
		itemRoute.PUT("/datastores/:d_id/items/:i_id", items.ModifyItem)
		// 盘点台账数据
		itemRoute.PATCH("/datastores/:d_id/items/:i_id", items.InventoryItem)
		// 盘点多条台账数据
		itemRoute.PATCH("/datastores/:d_id/items", items.MutilInventoryItem)
		// 删除单条台账数据
		itemRoute.DELETE("/datastores/:d_id/items/:i_id", items.DeleteItem)
	}

	// imp
	mapping := new(handler.Mapping)
	{
		mappingRoute := v1.Group("/mapping")
		// 导入数据
		mappingRoute.POST("/datastores/:d_id/upload", mapping.MappingUpload)
		// 导出数据
		mappingRoute.POST("/datastores/:d_id/download", mapping.MappingDownload)
	}

	// field
	field := new(handler.Field)
	{
		fieldRoute := v1.Group("/field")
		// 获取台账中所有的字段
		fieldRoute.GET("/datastores/:d_id/fields", field.FindFields)
	}

	// workflow
	workflow := new(handler.Workflow)
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
}
