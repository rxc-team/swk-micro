package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/outer/common/containerx"
	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/common/stringx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/api/outer/system/wfx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
)

// Validation 验证
type Validation struct{}

type SpecialChar struct {
	Special string `json:"special"`
}

// log出力
const (
	ValidationProcessName        = "Validation"
	ActionItemUniqueValidation   = "ItemUniqueValidation"
	ActionValidSpecial           = "ValidSpecial"
	ActionUpdateFieldsValidation = "UpdateFieldsValidation"
)

// ItemUniqueValidation 验证数据唯一性
// @Summary 验证数据唯一性
// @description 调用srv中的服务，验证数据唯一性
// @Tags Validation
// @Accept json
// @Produce  json
// @Param d_id path string true "台账ID"
// @Param id path string true "字段ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /validation/datastores/{id}/fields/{f_id}/relation [get]
func (a *Validation) ItemUniqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionItemUniqueValidation, fmt.Sprintf("Process FindItems:%s", loggerx.MsgProcessStarted))

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.CountRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionItemUniqueValidation, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := itemService.FindCount(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionItemUniqueValidation, err)
		return
	}

	result := true

	if response.GetTotal() > 0 {
		result = false
	}

	loggerx.InfoLog(c, ActionItemUniqueValidation, fmt.Sprintf("Process FindItems:%s", loggerx.MsgProcessEnded))
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionItemUniqueValidation)),
		Data:    result,
	})
}

// ValidSpecialChar 验证特殊字符是否合法
// @Router /validation/specialchar [get]
func (a *Validation) ValidSpecialChar(c *gin.Context) {
	loggerx.InfoLog(c, ActionValidSpecial, loggerx.MsgProcessStarted)
	// 获取公共参数
	db := sessionx.GetUserCustomer(c)
	appId := sessionx.GetCurrentApp(c)
	var value SpecialChar
	err := c.BindJSON(&value)
	if err != nil {
		httpx.GinHTTPError(c, ActionValidSpecial, err)
		return
	}

	appService := app.NewAppService("manage", client.DefaultClient)

	var req app.FindAppRequest
	req.AppId = appId
	req.Database = db
	response, err := appService.FindApp(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionValidSpecial, err)
		return
	}
	specialChars := response.GetApp().GetConfigs().GetSpecial()
	var special bool = true
	if len(specialChars) != 0 {
		var specialchar string
		// 编辑特殊字符
		for i := 0; i < len(specialChars); {
			specialchar += specialChars[i : i+1]
			i += 2
		}
		special = stringx.SpecialCheck(value.Special, specialchar)
	}
	loggerx.InfoLog(c, ActionValidSpecial, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionValidSpecial)),
		Data:    special,
	})
}

// UpdateFieldsValidation 验证更新字段是否有流程
// @Summary 验证更新字段是否有流程
// @description 调用srv中的服务,验证更新字段是否有流程
// @Tags Validation
// @Accept json
// @Produce  json
// @Param d_id path string true "台账ID"
// @Param m_id path string true "映射ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /validation/datastores/{d_id}/mappings/{m_id} [get]
func (a *Validation) UpdateFieldsValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpdateFieldsValidation, fmt.Sprintf("Process UpdateFieldsValidation:%s", loggerx.MsgProcessStarted))

	type Req struct {
		UpdateFields []string `json:"update_fields"`
	}

	var req Req
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionItemUniqueValidation, err)
		return
	}

	// PATH参数
	datastoreId := c.Param("id")
	// 共通参数
	db := sessionx.GetUserCustomer(c)
	groupID := sessionx.GetUserGroup(c)
	appId := sessionx.GetCurrentApp(c)

	// 获取用户流程情报(默认必须是有效的流程)
	userWorkflows := wfx.GetUserWorkflow(db, groupID, appId, datastoreId, "update")

	result := false

	// 映射没有删除,去除用户的删除流程
	update := containerx.New()
LP:
	for _, uwf := range userWorkflows {
		if len(uwf.Params["fields"]) == 0 {
			result = true
			break LP
		} else {
			fields := uwf.Params["fields"]
			update.AddAll(strings.Split(fields, ",")...)
		}
	}

	if result {
		loggerx.InfoLog(c, ActionUpdateFieldsValidation, fmt.Sprintf("Process UpdateFieldsValidation:%s", loggerx.MsgProcessEnded))
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionUpdateFieldsValidation)),
			Data:    result,
		})
		return
	}

	fields := update.ToList()

LP1:
	for _, upf := range req.UpdateFields {
		for _, wpf := range fields {
			if upf == wpf {
				result = true
				break LP1
			}
		}
	}

	loggerx.InfoLog(c, ActionUpdateFieldsValidation, fmt.Sprintf("Process UpdateFieldsValidation:%s", loggerx.MsgProcessEnded))
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionUpdateFieldsValidation)),
		Data:    result,
	})
}
