package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/outer/common/containerx"
	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
)

// Field 字段
type Field struct{}

// log出力使用
const (
	FieldProcessName = "Field"
	ActionFindFields = "FindFields"
)

//FindFields 获取所有字段
// @Summary 获取所有字段
// @description 调用srv中的field服务，获取所有字段
// @Tags Field
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "台账ID"
// @Param field_name query string false "字段名"
// @Param field_type query string false "字段类型"
// @Param is_required query string false "是否必须入力"
// @Param as_title query string false "是否作为标题"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /datastores/{d_id}/fields [get]
func (f *Field) FindFields(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindFields, loggerx.MsgProcessStarted)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.FieldsRequest
	// 从query获取
	req.FieldName = c.Query("field_name")
	req.FieldType = c.Query("field_type")
	req.IsRequired = c.Query("is_required")
	req.AsTitle = c.Query("as_title")
	req.InvalidatedIn = c.Query("invalidated_in")
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := fieldService.FindFields(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindFields, err)
		return
	}
	needRole := c.Query("needRole")
	if needRole == "true" {
		roles := sessionx.GetUserRoles(c)

		pmService := permission.NewPermissionService("manage", client.DefaultClient)

		var preq permission.FindActionsRequest
		preq.RoleId = roles
		preq.PermissionType = "app"
		preq.AppId = sessionx.GetCurrentApp(c)
		preq.ActionType = "datastore"
		preq.Database = sessionx.GetUserCustomer(c)
		pResp, err := pmService.FindActions(context.TODO(), &preq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindFields, err)
			return
		}
		set := containerx.New()
		for _, act := range pResp.GetActions() {
			if act.ObjectId == req.DatastoreId {
				set.AddAll(act.Fields...)
			}
		}

		fieldList := set.ToList()
		allFields := response.GetFields()
		var result []*field.Field
		for _, fieldID := range fieldList {
			f, err := findField(fieldID, allFields)
			if err == nil {
				result = append(result, f)
			}
		}

		loggerx.InfoLog(c, ActionFindFields, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionFindFields)),
			Data:    result,
		})

		return
	}

	loggerx.InfoLog(c, ActionFindFields, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionFindFields)),
		Data:    response.GetFields(),
	})
}

func findField(fieldID string, fields []*field.Field) (r *field.Field, err error) {
	var reuslt *field.Field
	for _, f := range fields {
		if f.GetFieldId() == fieldID {
			reuslt = f
			break
		}
	}

	if reuslt == nil {
		return nil, fmt.Errorf("not found")
	}

	return reuslt, nil
}
