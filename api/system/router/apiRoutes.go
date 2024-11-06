package router

import (
	"github.com/gin-gonic/gin"
	"rxcsoft.cn/pit3/api/system/handler"
)

// InitRouter 初始化路由
func InitRouter(router *gin.Engine) error {

	status := new(handler.SystemInfo)
	router.GET("/system/api/v1/release", status.GetReleaseStatus)
	router.GET("/system/api/v1/config", status.GetStatusAndIP)
	router.POST("/system/api/v1/config", status.SetStatusAndIP)

	return nil
}
