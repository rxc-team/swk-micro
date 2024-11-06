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
	"rxcsoft.cn/pit3/srv/manage/proto/level"
)

// Level 授权等级
type Level struct{}

// log出力使用
const (
	LevelProcessName  = "Level"
	LevelFindLevels   = "FindLevels"
	LevelFindLevel    = "FindLevel"
	LevelAddLevel     = "AddLevel"
	LevelModifyLevel  = "ModifyLevel"
	LevelDeleteLevel  = "DeleteLevel"
	LevelDeleteLevels = "DeleteLevels"
)

// FindLevels 获取所有授权等级
// @Router /levels [get]
func (f *Level) FindLevels(c *gin.Context) {
	loggerx.InfoLog(c, LevelFindLevels, loggerx.MsgProcessStarted)

	levelService := level.NewLevelService("manage", client.DefaultClient)

	var req level.FindLevelsRequest
	response, err := levelService.FindLevels(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, LevelFindLevels, err)
		return
	}

	loggerx.InfoLog(c, LevelFindLevels, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LevelProcessName, LevelFindLevels)),
		Data:    response.GetLevels(),
	})
}

// FindLevel 获取授权等级
// @Router /levels/{level_id} [get]
func (f *Level) FindLevel(c *gin.Context) {
	loggerx.InfoLog(c, LevelFindLevel, loggerx.MsgProcessStarted)

	levelService := level.NewLevelService("manage", client.DefaultClient)

	var req level.FindLevelRequest
	req.LevelId = c.Param("level_id")
	response, err := levelService.FindLevel(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, LevelFindLevel, err)
		return
	}

	loggerx.InfoLog(c, LevelFindLevel, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LevelProcessName, LevelFindLevel)),
		Data:    response.GetLevel(),
	})
}

// AddLevel 添加授权等级
// @Router /levels [post]
func (f *Level) AddLevel(c *gin.Context) {
	loggerx.InfoLog(c, LevelAddLevel, loggerx.MsgProcessStarted)

	levelService := level.NewLevelService("manage", client.DefaultClient)

	var req level.AddLevelRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, LevelAddLevel, err)
		return
	}
	// 从共通中获取
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := levelService.AddLevel(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, LevelAddLevel, err)
		return
	}
	loggerx.SuccessLog(c, LevelAddLevel, fmt.Sprintf("Level[%s] Create Success", response.GetLevelId()))

	loggerx.InfoLog(c, LevelAddLevel, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LevelProcessName, LevelAddLevel)),
		Data:    response,
	})
}

// ModifyLevel 更新授权等级
// @Router /levels/{level_id} [put]
func (f *Level) ModifyLevel(c *gin.Context) {
	loggerx.InfoLog(c, LevelModifyLevel, loggerx.MsgProcessStarted)

	levelService := level.NewLevelService("manage", client.DefaultClient)

	var req level.ModifyLevelRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, LevelModifyLevel, err)
		return
	}
	// 从path中获取参数
	req.LevelId = c.Param("level_id")
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := levelService.ModifyLevel(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, LevelModifyLevel, err)
		return
	}
	loggerx.SuccessLog(c, LevelModifyLevel, fmt.Sprintf(loggerx.MsgProcesSucceed, LevelModifyLevel))

	loggerx.InfoLog(c, LevelModifyLevel, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, LevelProcessName, LevelModifyLevel)),
		Data:    response,
	})
}

// DeleteLevels 硬删除多个授权等级
// @Router /levels [delete]
func (f *Level) DeleteLevels(c *gin.Context) {
	loggerx.InfoLog(c, LevelDeleteLevels, loggerx.MsgProcessStarted)

	levelService := level.NewLevelService("manage", client.DefaultClient)

	var req level.DeleteLevelsRequest
	req.LevelIds = c.QueryArray("level_ids")

	response, err := levelService.DeleteLevels(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, LevelDeleteLevels, err)
		return
	}
	loggerx.SuccessLog(c, LevelDeleteLevels, fmt.Sprintf(loggerx.MsgProcesSucceed, LevelDeleteLevels))

	loggerx.InfoLog(c, LevelDeleteLevels, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, LevelProcessName, LevelDeleteLevels)),
		Data:    response,
	})
}
