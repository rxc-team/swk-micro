package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/errors"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
)

// Field 字段
type Field struct{}

// log出力使用
const (
	FieldProcessName            = "Field"
	ActionFindAppFields         = "FindAppFields"
	ActionFindFields            = "FindFields"
	ActionFindField             = "FindField"
	ActionVerifyFunc            = "VerifyFunc"
	ActionAddField              = "AddField"
	ActionBlukAddField          = "BlukAddField"
	ActionModifyField           = "ModifyField"
	ActionDeleteField           = "DeleteField"
	ActionDeleteDatastoreFields = "DeleteDatastoreFields"
	ActionDeleteSelectFields    = "DeleteSelectFields"
	ActionHardDeleteFields      = "HardDeleteFields"
	ActionRecoverSelectFields   = "RecoverSelectFields"
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

// FindField 获取字段
// @Router /fields/{f_id} [get]
func (f *Field) FindField(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindField, loggerx.MsgProcessStarted)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.FieldRequest
	req.FieldId = c.Param("f_id")
	req.DatastoreId = c.Query("datastore_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := fieldService.FindField(context.TODO(), &req)
	if err != nil {
		er := errors.Parse(err.Error())
		if er.GetDetail() == mongo.ErrNoDocuments.Error() {
			c.JSON(200, httpx.Response{
				Status:  0,
				Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionFindField)),
				Data:    nil,
			})
			return
		}
		httpx.GinHTTPError(c, ActionFindField, err)
		return
	}

	loggerx.InfoLog(c, ActionFindField, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionFindField)),
		Data:    response.GetField(),
	})
}

// VerifyFunc 验证函数是否正确
// @Router /func/verify [post]
func (f *Field) VerifyFunc(c *gin.Context) {
	loggerx.InfoLog(c, ActionVerifyFunc, loggerx.MsgProcessStarted)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.VerifyFuncRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionVerifyFunc, err)
		return
	}

	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, _ := fieldService.VerifyFunc(context.TODO(), &req)

	loggerx.InfoLog(c, ActionFindField, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionFindField)),
		Data:    response,
	})
}

// AddField 添加字段
// @Router /datastores/{d_id}/fields [post]
func (f *Field) AddField(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddField, loggerx.MsgProcessStarted)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.AddRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddField, err)
		return
	}
	// 从path中获取
	req.DatastoreId = c.Param("d_id")
	// 从共通中获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := fieldService.AddField(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddField, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddField, fmt.Sprintf("Field[%s] Create Success", response.GetFieldId()))

	// 添加字段成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
	params["field_name"] = req.GetFieldName()     // 新规的时候取传入参数
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	// 通过台账Id查询台账名
	var freq datastore.DatastoreRequest
	freq.DatastoreId = c.Param("d_id")
	freq.Database = sessionx.GetUserCustomer(c)

	fresponse, err := datastoreService.FindDatastore(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddField, err)
		return
	}
	params["datastore_name"] = "{{" + fresponse.GetDatastore().GetDatastoreName() + "}}"
	params["api_key"] = fresponse.GetDatastore().GetApiKey()
	params["field_id"] = response.GetFieldId()

	loggerx.ProcessLog(c, ActionAddField, msg.L049, params)

	// 添加多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    sessionx.GetCurrentApp(c),
		Type:     "fields",
		Key:      req.GetDatastoreId() + "_" + response.GetFieldId(),
		Value:    req.GetFieldName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddField, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddField, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, sessionx.GetUserDomain(c))

	code := "I_009"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "refresh",
		Code:    code,
		Content: "添加字段成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + req.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionAddField, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionAddField)),
		Data:    response,
	})
}

// ModifyField 更新字段
// @Router /fields/{f_id} [put]
func (f *Field) ModifyField(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyField, loggerx.MsgProcessStarted)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.ModifyRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyField, err)
		return
	}
	// 从path中获取参数
	req.FieldId = c.Param("f_id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)
	// 如果是布局显示和宽度显示的更新的处理
	if req.GetIsDisplaySetting() == "true" {
		response, err := fieldService.ModifyField(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyField, err)
			return
		}
		loggerx.SuccessLog(c, ActionModifyField, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyField))
		loggerx.InfoLog(c, ActionModifyField, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionModifyField)),
			Data:    response,
		})
		return
	}

	// 变更字段前查询字段信息
	var freq field.FieldRequest
	freq.FieldId = c.Param("f_id")
	freq.Database = sessionx.GetUserCustomer(c)
	freq.DatastoreId = req.GetDatastoreId()
	fresponse, err := fieldService.FindField(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyField, err)
		return
	}
	fieldInfo := fresponse.GetField()
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	var dreq datastore.DatastoreRequest
	dreq.DatastoreId = fieldInfo.GetDatastoreId()
	dreq.Database = sessionx.GetUserCustomer(c)

	dresponse, err := datastoreService.FindDatastore(context.TODO(), &dreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyField, err)
		return
	}
	response, err := fieldService.ModifyField(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyField, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyField, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyField))
	// 变更成功后，比较变更的结果，记录日志
	// 比较字段的必须入力是否变更
	isRequired := "false"
	if fieldInfo.GetIsRequired() {
		isRequired = "true"
	}
	if isRequired != req.GetIsRequired() && req.GetIsRequired() == "false" {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		loggerx.ProcessLog(c, ActionModifyField, msg.L053, params)
	}
	// 比较字段的唯一性验证是否变更
	unique := "false"
	if fieldInfo.GetUnique() {
		unique = "true"
	}
	if unique != req.GetUnique() && req.GetUnique() == "false" {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		loggerx.ProcessLog(c, ActionModifyField, msg.L054, params)
	}
	// 比较字段的位数是否变更（text，textarea）
	min := fmt.Sprint(fieldInfo.GetMinValue())
	max := fmt.Sprint(fieldInfo.GetMaxValue())
	if req.GetMinValue() != "" && req.GetMaxValue() != "" && (min != req.GetMinValue() || max != req.GetMaxValue()) {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		params["min"] = req.GetMinValue()
		params["max"] = req.GetMaxValue()
		loggerx.ProcessLog(c, ActionModifyField, msg.L055, params)
	}
	// 比较字段的选项组是否变更（option字段）
	optionID := fieldInfo.GetOptionId()
	if fieldInfo.GetFieldType() == "options" && req.GetOptionId() != optionID && len(req.GetOptionId()) > 0 {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		optionService := option.NewOptionService("database", client.DefaultClient)

		var oreq option.FindOptionRequest
		oreq.OptionId = req.GetOptionId()
		oreq.Invalid = "true" // 包含软删除的数据
		oreq.AppId = sessionx.GetCurrentApp(c)
		oreq.Database = sessionx.GetUserCustomer(c)

		oresponse, err := optionService.FindOption(context.TODO(), &oreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyField, err)
			return
		}
		if len(oresponse.GetOptions()) > 0 {
			params["option_gourp_name"] = "{{" + oresponse.GetOptions()[0].GetOptionName() + "}}"
			params["option_id"] = oresponse.GetOptions()[0].GetOptionId()
		} else {
			params["option_gourp_name"] = ""
			params["option_id"] = ""
		}
		loggerx.ProcessLog(c, ActionModifyField, msg.L056, params)
	}
	/// 比较字段的用户组是否变更（user类型）
	userGroup := fieldInfo.GetUserGroupId()
	if fieldInfo.GetFieldType() == "user" && req.GetUserGroupId() != userGroup && len(req.GetUserGroupId()) > 0 {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		groupService := group.NewGroupService("manage", client.DefaultClient)

		var greq group.FindGroupRequest
		greq.GroupId = req.GetUserGroupId()
		greq.Database = sessionx.GetUserCustomer(c)
		gresponse, err := groupService.FindGroup(context.TODO(), &greq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyField, err)
			return
		}
		params["group_name"] = "{{" + gresponse.GetGroup().GetGroupName() + "}}"
		loggerx.ProcessLog(c, ActionModifyField, msg.L057, params)
	}
	// 比较字段关联的台账的字段是否变更（关联台账字段）
	lookupField := fieldInfo.GetLookupFieldId()
	lookupDatastoreID := fieldInfo.GetLookupDatastoreId()
	if fieldInfo.GetFieldType() == "lookup" && (lookupField != req.GetLookupFieldId() || lookupDatastoreID != req.GetLookupDatastoreId()) && len(req.GetLookupFieldId()) > 0 && len(req.GetLookupDatastoreId()) > 0 {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		// 查询关联台账，字段信息
		var lookupfreq field.FieldRequest
		lookupfreq.FieldId = req.GetLookupFieldId()
		lookupfreq.DatastoreId = req.GetLookupDatastoreId()
		lookupfreq.Database = sessionx.GetUserCustomer(c)
		lookupfresponse, err := fieldService.FindField(context.TODO(), &lookupfreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyField, err)
			return
		}
		lookupfieldInfo := lookupfresponse.GetField()
		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
		var lookupdreq datastore.DatastoreRequest
		lookupdreq.DatastoreId = lookupfieldInfo.GetDatastoreId()
		lookupdreq.Database = sessionx.GetUserCustomer(c)

		lookupdresponse, err := datastoreService.FindDatastore(context.TODO(), &lookupdreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyField, err)
			return
		}
		params["field_name1"] = "{{" + lookupfieldInfo.GetFieldName() + "}}"
		params["field_id1"] = lookupfieldInfo.GetFieldId()
		params["datastore_name1"] = "{{" + lookupdresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key1"] = lookupdresponse.GetDatastore().GetApiKey()
		loggerx.ProcessLog(c, ActionModifyField, msg.L058, params)
	}
	// 比较字段的显示位数，前缀是否变更（自增字段）
	digit := fmt.Sprint(fieldInfo.GetDisplayDigits())
	prefix := fieldInfo.GetPrefix()
	if fieldInfo.GetFieldType() == "autonum" && (digit != req.GetDisplayDigits() || prefix != req.GetPrefix()) && len(req.GetDisplayDigits()) > 0 {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		params["display_digit"] = req.GetDisplayDigits()
		params["prefix"] = req.GetPrefix()
		loggerx.ProcessLog(c, ActionModifyField, msg.L059, params)
	}
	// 比较字段的返回形式，公式是否变更（函数字段）
	returnForm := fieldInfo.GetReturnType()
	formula := fieldInfo.GetFormula()
	if fieldInfo.GetFieldType() == "function" && (returnForm != req.GetReturnType() || formula != req.GetFormula()) && len(req.GetReturnType()) > 0 && len(req.GetFormula()) > 0 {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fieldInfo.GetFieldName() + "}}"
		params["field_id"] = fieldInfo.GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		params["type_name"] = fieldInfo.GetFieldType()
		params["return_form"] = req.GetReturnType()
		// 查询函数的公式里字段的信息
		// var formulafreq field.FieldRequest
		// formulafreq.FieldId = fieldId //从formula里取fieldId如果有的话
		// formulafreq.Database = sessionx.GetUserCustomer(c)
		// formulafreq.DatastoreId = req.GetDatastoreId()
		// formulafresponse, err := fieldService.FindField(context.TODO(), &formulafreq)
		// if err == nil {
		// 	lookupfieldInfo := formulafresponse.GetField()
		// }
		params["formula"] = req.GetFormula()
		loggerx.ProcessLog(c, ActionModifyField, msg.L060, params)
	}

	if len(req.GetFieldName()) > 0 {
		// 添加多语言数据
		langService := language.NewLanguageService("global", client.DefaultClient)

		languageReq := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    sessionx.GetCurrentApp(c),
			Type:     "fields",
			Key:      req.GetDatastoreId() + "_" + req.GetFieldId(),
			Value:    req.GetFieldName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}

		_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyField, err)
			return
		}
		loggerx.SuccessLog(c, ActionModifyField, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))
		// 通知刷新多语言数据
		langx.RefreshLanguage(req.Writer, sessionx.GetUserDomain(c))
	}

	code := "I_010"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "refresh",
		Code:    code,
		Content: "更新字段成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + req.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionModifyField, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionModifyField)),
		Data:    response,
	})
}

// DeleteSelectFields 删除选中的字段
// @Router /fields [delete]
func (f *Field) DeleteSelectFields(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSelectFields, loggerx.MsgProcessStarted)

	var req field.DeleteSelectFieldsRequest
	req.FieldIdList = c.QueryArray("field_id_list")
	req.DatastoreId = c.Query("datastore_id")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	fieldService := field.NewFieldService("database", client.DefaultClient)
	// 将字段无效化之前查询字段名、台账名、台账api_key
	fieldNameList := make(map[string]string)
	datastoreNameList := make(map[string]string)
	datastoreApiKeyList := make(map[string]string)
	for _, id := range req.GetFieldIdList() {
		var freq field.FieldRequest
		freq.FieldId = id
		freq.Database = sessionx.GetUserCustomer(c)
		freq.DatastoreId = req.GetDatastoreId()
		fresponse, err := fieldService.FindField(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteSelectFields, err)
			return
		}
		fieldNameList[id] = "{{" + fresponse.GetField().GetFieldName() + "}}"
		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
		var dreq datastore.DatastoreRequest
		dreq.DatastoreId = fresponse.GetField().GetDatastoreId()
		dreq.Database = sessionx.GetUserCustomer(c)

		dresponse, err := datastoreService.FindDatastore(context.TODO(), &dreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteSelectFields, err)
			return
		}
		datastoreNameList[id] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		datastoreApiKeyList[id] = dresponse.GetDatastore().GetApiKey()
	}
	response, err := fieldService.DeleteSelectFields(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectFields, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteSelectFields, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeleteSelectFields))
	// 将字段无效化成功后保存日志到DB
	for _, id := range req.GetFieldIdList() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = fieldNameList[id]
		params["field_id"] = id
		params["datastore_name"] = datastoreNameList[id]
		params["api_key"] = datastoreApiKeyList[id]
		loggerx.ProcessLog(c, ActionDeleteSelectFields, msg.L050, params)
	}

	code := "I_011"
	for _, id := range req.GetFieldIdList() {
		param := wsx.MessageParam{
			Sender:  "SYSTEM",
			Domain:  sessionx.GetUserDomain(c),
			MsgType: "refresh",
			Code:    code,
			Content: "无效化字段成功，请刷新浏览器获取最新数据！",
			Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + id,
			Status:  "unread",
		}
		wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))
	}

	loggerx.InfoLog(c, ActionDeleteSelectFields, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionDeleteSelectFields)),
		Data:    response,
	})
}

// HardDeleteFields 物理删除选中字段
// @Router /phydel/datastores/{d_id}/fields [delete]
func (f *Field) HardDeleteFields(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteFields, loggerx.MsgProcessStarted)
	db := sessionx.GetUserCustomer(c)
	appId := sessionx.GetCurrentApp(c)
	datastoreId := c.Param("d_id")
	domain := sessionx.GetUserDomain(c)

	lang := sessionx.GetCurrentLanguage(c)

	langData := langx.GetLanguageData(db, lang, domain)

	var req field.HardDeleteFieldsRequest
	req.FieldIdList = c.QueryArray("field_id_list")
	req.DatastoreId = datastoreId
	req.Database = db

	fieldService := field.NewFieldService("database", client.DefaultClient)
	// 将字段删除之前查询字段名、台账名、台账api_key
	fieldNameList := make(map[string]string)
	datastoreNameList := make(map[string]string)
	datastoreApiKeyList := make(map[string]string)
	// 获取台账的unique_field关系
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	var uniqueReq datastore.DatastoreRequest
	uniqueReq.DatastoreId = datastoreId
	uniqueReq.Database = db
	uniqueRes, err := datastoreService.FindDatastore(context.TODO(), &uniqueReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteFields, err)
		return
	}
	datastoreUniqueFields := uniqueRes.GetDatastore().GetUniqueFields()
	for _, uniqueField := range datastoreUniqueFields {
		isDel := false
		for _, id := range req.GetFieldIdList() {
			isUnique := strings.Index(uniqueField, id)
			if isUnique > -1 {
				isDel = true
				break
			}
		}
		if isDel {
			var delUniqueReq datastore.DeleteUniqueRequest
			delUniqueReq.AppId = appId
			delUniqueReq.DatastoreId = datastoreId
			delUniqueReq.Database = db
			delUniqueReq.UniqueFields = uniqueField
			_, err := datastoreService.DeleteUniqueKey(c.Request.Context(), &delUniqueReq)
			if err != nil {
				httpx.GinHTTPError(c, ActionHardDeleteFields, err)
				return
			}
		}
	}
	// 硬删除的文件字段类型字段
	fileFieldIDs := []string{}
	for _, id := range req.GetFieldIdList() {
		var freq field.FieldRequest
		freq.FieldId = id
		freq.Database = db
		freq.DatastoreId = req.GetDatastoreId()
		fresponse, err := fieldService.FindField(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteFields, err)
			return
		}
		fieldNameList[id] = langx.GetLangValue(langData, fresponse.GetField().GetFieldName(), langx.DefaultResult)

		// 文件字段类型字段累计
		if fresponse.GetField().GetFieldType() == "file" {
			fileFieldIDs = append(fileFieldIDs, fresponse.GetField().GetFieldId())
		}

		var dreq datastore.DatastoreRequest
		dreq.DatastoreId = fresponse.GetField().GetDatastoreId()
		dreq.Database = db
		dresponse, err := datastoreService.FindDatastore(context.TODO(), &dreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteFields, err)
			return
		}
		datastoreNameList[id] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		datastoreApiKeyList[id] = dresponse.GetDatastore().GetApiKey()
	}

	// opss
	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
	}

	// 若有文件类型字段则累计其minio文件
	delFileList := []string{}
	if len(fileFieldIDs) > 0 {
		// 查找台账数据grpc
		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)
		itemService := item.NewItemService("database", ct)

		// 数据总量获取
		cReq := item.CountRequest{
			AppId:         appId,
			DatastoreId:   datastoreId,
			ConditionType: "and",
			Database:      db,
		}
		cResp, err := itemService.FindCount(context.TODO(), &cReq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteFields, err)
			return
		}

		// 每次2000为一组数据
		total := float64(cResp.GetTotal())
		count := math.Ceil(total / 2000)
	LP:
		for i := int64(0); i < int64(count); i++ {
			var req item.ItemsRequest
			req.DatastoreId = datastoreId
			req.AppId = appId
			req.PageIndex = i + 1
			req.PageSize = 2000
			req.Database = db
			req.IsOrigin = true

			response, err := itemService.FindItems(context.TODO(), &req, opss)
			if err != nil {
				httpx.GinHTTPError(c, ActionHardDeleteFields, err)
				return
			}

			if len(response.GetItems()) == 0 {
				break LP
			}

			for _, dt := range response.GetItems() {
				for _, id := range fileFieldIDs {
					if value, ext := dt.Items[id]; ext {
						var result []typesx.FileValue
						json.Unmarshal([]byte(value.GetValue()), &result)
						for _, file := range result {
							delFileList = append(delFileList, file.URL)
						}
					}
				}
			}
		}
	}

	response, err := fieldService.HardDeleteFields(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteFields, err)
		return
	}
	loggerx.SuccessLog(c, ActionHardDeleteFields, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionHardDeleteFields))

	// 根据上文累计minio文件删除冗余minio文件
	if len(delFileList) > 0 {
		go filex.DeletePublicDataFiles(domain, appId, delFileList)
	}

	for _, f := range req.FieldIdList {
		// 将字段删除成功后保存日志到DB
		// 编辑参数保存日志到DB
		fname := strings.Builder{}
		fname.WriteString(fieldNameList[f])
		fname.WriteString("(")
		fname.WriteString(sessionx.GetCurrentLanguage(c))
		fname.WriteString(")")
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = fname.String()
		params["field_id"] = f
		params["datastore_name"] = datastoreNameList[f]
		params["api_key"] = datastoreApiKeyList[f]
		loggerx.ProcessLog(c, ActionHardDeleteFields, msg.L051, params)

		// 删除多语言数据
		langx.DeleteAppLanguageData(db, domain, appId, "fields", datastoreId+"_"+f)
		loggerx.SuccessLog(c, ActionHardDeleteFields, fmt.Sprintf(loggerx.MsgProcesSucceed, "DeleteAppLanguageData"))
	}

	code := "I_012"
	for _, id := range req.GetFieldIdList() {
		param := wsx.MessageParam{
			Sender:  "SYSTEM",
			Domain:  sessionx.GetUserDomain(c),
			MsgType: "refresh",
			Code:    code,
			Content: "删除字段成功，请刷新浏览器获取最新数据！",
			Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + id,
			Status:  "unread",
		}
		wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))
	}
	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionHardDeleteFields, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionHardDeleteFields)),
		Data:    response,
	})
}

// RecoverSelectFields 恢复选中的字段
// @Router /recover/fields [PUT]
func (f *Field) RecoverSelectFields(c *gin.Context) {
	loggerx.InfoLog(c, ActionRecoverSelectFields, loggerx.MsgProcessStarted)

	var req field.RecoverSelectFieldsRequest

	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectFields, err)
		return
	}
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	fieldService := field.NewFieldService("database", client.DefaultClient)
	response, err := fieldService.RecoverSelectFields(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectFields, err)
		return
	}
	loggerx.SuccessLog(c, ActionRecoverSelectFields, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionRecoverSelectFields))
	code := "I_013"
	// 将字段恢复成功后保存日志到DB
	for _, id := range req.GetFieldIdList() {
		// 查找字段名日志中使用
		var freq field.FieldRequest
		freq.FieldId = id
		freq.Database = sessionx.GetUserCustomer(c)
		freq.DatastoreId = req.GetDatastoreId()
		fresponse, err := fieldService.FindField(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionRecoverSelectFields, err)
			return
		}
		// 查找台账名日志中使用
		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
		var dreq datastore.DatastoreRequest
		dreq.DatastoreId = fresponse.GetField().GetDatastoreId()
		dreq.Database = sessionx.GetUserCustomer(c)

		dresponse, err := datastoreService.FindDatastore(context.TODO(), &dreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionRecoverSelectFields, err)
			return
		}
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["field_name"] = "{{" + fresponse.GetField().GetFieldName() + "}}"
		params["field_id"] = fresponse.GetField().GetFieldId()
		params["datastore_name"] = "{{" + dresponse.GetDatastore().GetDatastoreName() + "}}"
		params["api_key"] = dresponse.GetDatastore().GetApiKey()
		loggerx.ProcessLog(c, ActionRecoverSelectFields, msg.L052, params)

		param := wsx.MessageParam{
			Sender:  "SYSTEM",
			Domain:  sessionx.GetUserDomain(c),
			MsgType: "refresh",
			Code:    code,
			Content: "恢复字段成功，请刷新浏览器获取最新数据！",
			Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + id,
			Status:  "unread",
		}
		wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))
	}

	loggerx.InfoLog(c, ActionRecoverSelectFields, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I013, fmt.Sprintf(httpx.Temp, FieldProcessName, ActionRecoverSelectFields)),
		Data:    response,
	})
}
