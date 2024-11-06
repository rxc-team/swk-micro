package webui

import (
	"context"
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
)

// Field 字段
type Field struct{}

// log出力使用
const (
	FieldProcessName    = "Field"
	ActionFindAppFields = "FindAppFields"
	ActionFindFields    = "FindFields"
)

// FindAppFields 查找APP中多个字段
// @Router /app/fields [get]
func (f *Field) FindAppFields(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindAppFields, loggerx.MsgProcessStarted)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.AppFieldsRequest
	// Query中获取
	req.AppId = c.Query("app_id")
	if req.AppId == "" {
		// 从共通获取
		req.AppId = sessionx.GetCurrentApp(c)
	}
	req.FieldType = c.Query("field_type")
	req.LookupDatastoreId = c.Query("lookup_datastore_id")
	req.InvalidatedIn = c.Query("invalidated_in")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := fieldService.FindAppFields(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindAppFields, err)
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
			httpx.GinHTTPError(c, ActionFindFolders, err)
			return
		}
		set := containerx.New()
		for _, act := range pResp.GetActions() {
			for _, field := range act.GetFields() {
				set.Add(act.GetObjectId() + "#" + field)
			}
		}

		dsfieldList := set.ToList()
		allFields := response.GetFields()
		var result []*field.Field
		for _, dsfieldID := range dsfieldList {
			f, err := fieldx.FindDatastoreField(dsfieldID, allFields)
			if err == nil {
				result = append(result, f)
			}
		}

		loggerx.InfoLog(c, ActionFindAppFields, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionFindAppFields)),
			Data:    result,
		})

		return
	}

	loggerx.InfoLog(c, ActionFindAppFields, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionFindAppFields)),
		Data:    response.GetFields(),
	})
}

// FindFields 获取所有字段
// @Router /datastores/{d_id}/fields [get]
func (f *Field) FindFields(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindFields, loggerx.MsgProcessStarted)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.FieldsRequest
	// 从query获取
	req.FieldName = c.Query("field_name")
	req.FieldType = c.Query("field_type")
	req.IsRequired = c.Query("is_required")
	req.IsFixed = c.Query("is_fixed")
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
		set := containerx.New()

		pmService := permission.NewPermissionService("manage", client.DefaultClient)

		var preq permission.FindActionsRequest
		preq.RoleId = roles
		preq.PermissionType = "app"
		preq.AppId = sessionx.GetCurrentApp(c)
		preq.ActionType = "datastore"
		preq.ObjectId = req.GetDatastoreId()
		preq.Database = sessionx.GetUserCustomer(c)
		pResp, err := pmService.FindActions(context.TODO(), &preq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindFolders, err)
			return
		}
		for _, act := range pResp.GetActions() {
			if act.ObjectId == req.DatastoreId {
				set.AddAll(act.Fields...)
			}
		}

		fieldList := set.ToList()
		allFields := response.GetFields()
		var result []*field.Field
		for _, fieldID := range fieldList {
			f, err := fieldx.FindField(fieldID, allFields)
			if err == nil {
				result = append(result, f)
			}
		}

		// 字段排序
		sort.Sort(typesx.FieldList(result))

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
