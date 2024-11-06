package webui

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
	TypeProcessName = "Type"
	ActionFindTypes = "FindTypes"
	ActionFindType  = "FindType"
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
