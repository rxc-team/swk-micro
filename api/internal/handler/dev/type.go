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
	types "rxcsoft.cn/pit3/srv/global/proto/help-type"
)

// Type 帮助文档类型
type Type struct{}

// log出力
const (
	TypeProcessName   = "Type"
	ActionFindTypes   = "FindTypes"
	ActionFindType    = "FindType"
	ActionAddType     = "AddType"
	ActionModifyType  = "ModifyType"
	ActionDeleteType  = "DeleteType"
	ActionDeleteTypes = "DeleteTypes"
)

// FindType 获取单个帮助文档类型
// @Router /types/{type_id} [get]
func (t *Type) FindType(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindType, loggerx.MsgProcessStarted)

	typeService := types.NewTypeService("global", client.DefaultClient)

	var req types.FindTypeRequest
	req.TypeId = c.Param("type_id")
	req.Database = "system"

	response, err := typeService.FindType(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindType, err)
		return
	}

	loggerx.InfoLog(c, ActionFindType, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, TypeProcessName, ActionFindType)),
		Data:    response.GetType(),
	})
}

// FindTypes 获取多个帮助文档类型
// @Router /types [get]
func (t *Type) FindTypes(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindTypes, loggerx.MsgProcessStarted)

	typeService := types.NewTypeService("global", client.DefaultClient)

	var req types.FindTypesRequest
	req.TypeName = c.Query("type_name")
	req.Show = c.Query("show")
	isDev := c.Query("is_dev")
	if isDev == "true" {
		req.LangCd = c.Query("lang_cd")
	} else {
		req.LangCd = sessionx.GetCurrentLanguage(c)
	}
	req.Database = "system"

	response, err := typeService.FindTypes(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindTypes, err)
		return
	}

	loggerx.InfoLog(c, ActionFindTypes, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, TypeProcessName, ActionFindTypes)),
		Data:    response.GetTypes(),
	})
}

// AddType 添加帮助文档类型
// @Router /types [post]
func (t *Type) AddType(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddType, loggerx.MsgProcessStarted)

	typeService := types.NewTypeService("global", client.DefaultClient)

	var req types.AddTypeRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddType, err)
		return
	}

	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = "system"

	response, err := typeService.AddType(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddType, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddType, fmt.Sprintf("Type[%s] create Success", response.GetTypeId()))

	loggerx.InfoLog(c, ActionAddType, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, TypeProcessName, ActionAddType)),
		Data:    response,
	})
}

// ModifyType 更新帮助文档类型
// @Router /types/{type_id} [put]
func (t *Type) ModifyType(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyType, loggerx.MsgProcessStarted)

	typeService := types.NewTypeService("global", client.DefaultClient)

	var req types.ModifyTypeRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyType, err)
		return
	}

	req.TypeId = c.Param("type_id")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = "system"

	response, err := typeService.ModifyType(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyType, err)
		return
	}

	loggerx.SuccessLog(c, ActionModifyType, fmt.Sprintf("Type[%s] Update Success", req.GetTypeId()))

	loggerx.InfoLog(c, ActionModifyType, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, TypeProcessName, ActionModifyType)),
		Data:    response,
	})
}

// DeleteTypes 硬删除多个帮助文档类型
// @Router /types [delete]
func (t *Type) DeleteTypes(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteTypes, loggerx.MsgProcessStarted)

	typeService := types.NewTypeService("global", client.DefaultClient)

	var req types.DeleteTypesRequest
	req.TypeIdList = c.QueryArray("type_id_list")
	req.Database = "system"

	response, err := typeService.DeleteTypes(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteTypes, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteTypes, fmt.Sprintf("Types[%s] HardDelete Success", req.GetTypeIdList()))

	loggerx.InfoLog(c, ActionDeleteTypes, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, TypeProcessName, ActionDeleteTypes)),
		Data:    response,
	})
}
