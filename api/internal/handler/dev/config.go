package dev

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	config "rxcsoft.cn/pit3/srv/global/proto/mail-config"
)

// Config 邮件配置
type Config struct{}

// log出力
const (
	ConfigProcessName  = "Config"
	ActionFindConfig   = "FindConfig"
	ActionFindConfigs  = "FindConfigs"
	ActionAddConfig    = "AddConfig"
	ActionModifyConfig = "ModifyConfig"
)

// FindConfig 获取邮件配置
// @Router /configs/config [get]
func (mc *Config) FindConfig(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindConfig, loggerx.MsgProcessStarted)

	configService := config.NewConfigService("global", client.DefaultClient)

	var req config.FindConfigRequest
	req.Database = "system"

	response, err := configService.FindConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindConfig, err)
		return
	}

	loggerx.InfoLog(c, ActionFindConfig, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ConfigProcessName, ActionFindConfig)),
		Data:    response.GetConfig(),
	})
}

// AddConfig 添加邮件配置
// @Router /configs [post]
func (mc *Config) AddConfig(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddConfig, loggerx.MsgProcessStarted)

	configService := config.NewConfigService("global", client.DefaultClient)

	var req config.AddConfigRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddConfig, err)
		return
	}
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = "system"

	response, err := configService.AddConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddConfig, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddConfig, fmt.Sprintf("Config[%s] create Success", response.GetConfigId()))

	loggerx.InfoLog(c, ActionAddConfig, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ConfigProcessName, ActionAddConfig)),
		Data:    response,
	})
}

// ModifyConfig 更新邮件配置
// @Router /configs/{config_id} [put]
func (mc *Config) ModifyConfig(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyConfig, loggerx.MsgProcessStarted)

	configService := config.NewConfigService("global", client.DefaultClient)

	var req config.ModifyConfigRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyConfig, err)
		return
	}
	req.ConfigId = c.Param("config_id")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = "system"

	response, err := configService.ModifyConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyConfig, err)
		return
	}

	loggerx.SuccessLog(c, ActionModifyConfig, fmt.Sprintf("Config[%s] Update Success", req.GetConfigId()))

	loggerx.InfoLog(c, ActionModifyConfig, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ConfigProcessName, ActionModifyConfig)),
		Data:    response,
	})
}
