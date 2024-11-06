package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
)

// Datastore datastore
type Datastore struct{}

// log出力
const (
	DatastoreProcessName          = "Datastore"
	ActionFindDatastores          = "FindDatastores"
	ActionFindDatastore           = "FindDatastore"
	ActionFindDatastoreMapping    = "FindDatastoreMapping"
	ActionAddDatastore            = "AddDatastore"
	ActionAddDatastoreMapping     = "AddDatastoreMapping"
	ActionAddUniqueKey            = "AddUniqueKey"
	ActionAddRelation             = "AddRelation"
	ActionModifyDatastore         = "ModifyDatastore"
	ActionModifyDatastoreMenuSort = "ModifyDatastoreMenuSort"
	ActionModifyDatastoreMapping  = "ModifyDatastoreMapping"
	ActionDeleteDatastore         = "DeleteDatastore"
	ActionDeleteDatastoreMapping  = "DeleteDatastoreMapping"
	ActionDeleteUniqueKey         = "DeleteUniqueKey"
	ActionDeleteRelation          = "DeleteRelation"
	ActionDeleteSelectDatastores  = "DeleteSelectDatastores"
	ActionHardDeleteDatastores    = "HardDeleteDatastores"
)

// FindDatastores 获取app下所有台账
// @Router /datastores [get]
func (d *Datastore) FindDatastores(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDatastores, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoresRequest
	// 从query获取
	req.DatastoreName = c.Query("datastore_name")
	req.CanCheck = c.Query("can_check")
	req.ShowInMenu = c.Query("show_in_menu")
	// 从共通获取
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)
	if c.Query("app_id") != "" {
		req.AppId = c.Query("app_id")
	}

	response, err := datastoreService.FindDatastores(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastores, err)
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
		preq.Database = sessionx.GetUserCustomer(c)
		pResp, err := pmService.FindActions(context.TODO(), &preq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindDatastores, err)
			return
		}
		for _, act := range pResp.GetActions() {
			if act.ActionMap["read"] {
				set.Add(act.ObjectId)
			}
		}

		dsList := set.ToList()
		allDs := response.GetDatastores()
		var result []*datastore.Datastore
		for _, dsID := range dsList {
			f, err := findDatastore(dsID, allDs)
			if err == nil {
				result = append(result, f)
			}
		}

		loggerx.InfoLog(c, ActionFindDatastores, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastores)),
			Data:    result,
		})

		return
	}

	loggerx.InfoLog(c, ActionFindDatastores, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastores)),
		Data:    response.GetDatastores(),
	})
}

func findDatastore(dsID string, dsList []*datastore.Datastore) (r *datastore.Datastore, err error) {
	var reuslt *datastore.Datastore
	for _, d := range dsList {
		if d.GetDatastoreId() == dsID {
			reuslt = d
			break
		}
	}

	if reuslt == nil {
		return nil, fmt.Errorf("not found")
	}

	return reuslt, nil
}

// FindDatastore 通过DatastoreID获取台账信息
// @Router /datastores/{d_id} [get]
func (d *Datastore) FindDatastore(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoreRequest
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.FindDatastore(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastore, err)
		return
	}

	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastore)),
		Data:    response.GetDatastore(),
	})
}

// FindDatastoreRelations 通过DatastoreID获取台账信息关系
// @Router /datastores/{d_id}/realtion [get]
func (d *Datastore) FindDatastoreRelations(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	datastoreId := c.Param("d_id")
	db := sessionx.GetUserCustomer(c)
	appId := sessionx.GetCurrentApp(c)

	var req datastore.DatastoreRequest
	// 从path获取
	req.DatastoreId = datastoreId
	req.Database = db

	response, err := datastoreService.FindDatastore(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastore, err)
		return
	}

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var freq field.FieldsRequest
	// 从path获取
	freq.DatastoreId = datastoreId
	// 从共通获取
	freq.AppId = appId
	freq.Database = db

	cfResp, err := fieldService.FindFields(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindFields, err)
		return
	}

	var relations []*datastore.RelationItem

	for _, relation := range response.GetDatastore().GetRelations() {
		var dreq datastore.DatastoreRequest
		// 从path获取
		dreq.DatastoreId = relation.GetDatastoreId()
		dreq.Database = db

		dResp, err := datastoreService.FindDatastore(context.TODO(), &dreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindDatastore, err)
			return
		}

		fieldService := field.NewFieldService("database", client.DefaultClient)

		var freq field.FieldsRequest
		// 从path获取
		freq.DatastoreId = relation.GetDatastoreId()
		// 从共通获取
		freq.AppId = appId
		freq.Database = db

		fResp, err := fieldService.FindFields(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindFields, err)
			return
		}

		fields := make(map[string]string)

		for f, rf := range relation.GetFields() {
			var key = f
			var value = rf
			for _, fs := range fResp.GetFields() {
				if f == fs.FieldId {
					key = fs.FieldName
				}
			}
			for _, fs := range cfResp.GetFields() {
				if rf == fs.FieldId {
					value = fs.FieldName
				}
			}

			fields[key] = value

		}

		relations = append(relations, &datastore.RelationItem{
			RelationId:  relation.GetRelationId(),
			DatastoreId: dResp.GetDatastore().GetDatastoreName(),
			Fields:      fields,
		})

	}

	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastore)),
		Data:    relations,
	})
}

// FindDatastoreByKey 通过ApiKey获取台账信息
// @Router /datastores/key/{api_key}/ [get]
func (d *Datastore) FindDatastoreByKey(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoreKeyRequest
	// 从path获取
	req.ApiKey = c.Param("api_key")
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := datastoreService.FindDatastoreByKey(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastore, err)
		return
	}

	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastore)),
		Data:    response.GetDatastore(),
	})
}

// FindDatastoreMapping 通过ID获取台账映射信息
// @Router /datastores/{d_id}/mappings/{m_id} [get]
func (d *Datastore) FindDatastoreMapping(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDatastoreMapping, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.MappingRequest
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	req.MappingId = c.Param("m_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.FindDatastoreMapping(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastoreMapping, err)
		return
	}

	loggerx.InfoLog(c, ActionFindDatastoreMapping, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastoreMapping)),
		Data:    response.GetMapping(),
	})
}

// AddDatastore 添加台账
// @Router /datastores [post]
func (d *Datastore) AddDatastore(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddDatastore, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.AddRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddDatastore, err)
		return
	}

	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.AddDatastore(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDatastore, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddDatastore, fmt.Sprintf("Datastore[%s] Create Success", response.GetDatastoreId()))

	// 添加台账成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)     // 取共通用户名
	params["datastore_name"] = req.GetDatastoreName() // 新规的时候取传入参数
	params["api_key"] = req.GetApiKey()

	loggerx.ProcessLog(c, ActionAddDatastore, msg.L045, params)

	// 添加多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    sessionx.GetCurrentApp(c),
		Type:     "datastores",
		Key:      response.GetDatastoreId(),
		Value:    req.GetDatastoreName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDatastore, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddDatastore, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))

	code := "I_005"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "refresh",
		Code:    code,
		Content: "台账添加成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + response.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, sessionx.GetUserDomain(c))

	loggerx.InfoLog(c, ActionAddDatastore, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionAddDatastore)),
		Data:    response,
	})
}

// AddDatastoreMapping 添加台账映射关系
// @Router /datastores/{d_id}/mappings [post]
func (d *Datastore) AddDatastoreMapping(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddDatastoreMapping, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.AddMappingRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddDatastoreMapping, err)
		return
	}

	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.DatastoreId = c.Param("d_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.AddDatastoreMapping(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDatastoreMapping, err)

		return
	}
	loggerx.SuccessLog(c, ActionAddDatastoreMapping, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddDatastoreMapping))

	// 添加多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    sessionx.GetCurrentApp(c),
		Type:     "mappings",
		Key:      req.GetDatastoreId() + "_" + response.GetMappingId(),
		Value:    req.GetMappingName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDatastoreMapping, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddDatastoreMapping, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), sessionx.GetUserDomain(c))

	loggerx.InfoLog(c, ActionAddDatastoreMapping, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionAddDatastoreMapping)),
		Data:    response,
	})
}

// AddUniqueKey 添加台账组合唯一属性
// @Router /datastores/{d_id}/unique [post]
func (d *Datastore) AddUniqueKey(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddUniqueKey, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Second * 60
		o.DialTimeout = time.Second * 60
	}

	var req datastore.AddUniqueRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddUniqueKey, err)
		return
	}

	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.DatastoreId = c.Param("d_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.AddUniqueKey(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddUniqueKey, err)

		return
	}
	loggerx.SuccessLog(c, ActionAddUniqueKey, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddUniqueKey))

	loggerx.InfoLog(c, ActionAddUniqueKey, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionAddUniqueKey)),
		Data:    response,
	})
}

// AddRelation 添加台账组合唯一属性
// @Router /datastores/{d_id}/relation [post]
func (d *Datastore) AddRelation(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddRelation, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.AddRelationRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddRelation, err)
		return
	}

	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.DatastoreId = c.Param("d_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.AddRelation(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddRelation, err)

		return
	}
	loggerx.SuccessLog(c, ActionAddRelation, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddRelation))

	loggerx.InfoLog(c, ActionAddRelation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionAddRelation)),
		Data:    response,
	})
}

// ModifyDatastore 更新台账
// @Router /datastores/{d_id} [PUT]
func (d *Datastore) ModifyDatastore(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyDatastore, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	// 变更前查询台账信息
	var freq datastore.DatastoreRequest
	freq.DatastoreId = c.Param("d_id")
	freq.Database = sessionx.GetUserCustomer(c)
	fresponse, err := datastoreService.FindDatastore(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyDatastore, err)
		return
	}
	datastoreInfo := fresponse.GetDatastore()

	var req datastore.ModifyRequest
	// 从body获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyDatastore, err)
		return
	}
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	// 从共通获取
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.ModifyDatastore(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyDatastore, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyDatastore, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyDatastore))
	// 变更成功后，比较变更的结果，记录日志
	// 比较是否是盘点台账
	canCheck := "false"
	if datastoreInfo.GetCanCheck() {
		canCheck = "true"
	}
	if canCheck != req.GetCanCheck() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["datastore_name"] = "{{" + datastoreInfo.GetDatastoreName() + "}}"
		params["api_key"] = datastoreInfo.GetApiKey()
		if req.GetCanCheck() == "true" {
			// 台账被设置为盘点台账的日志
			loggerx.ProcessLog(c, ActionModifyDatastore, msg.L047, params)
		} else {
			// 台账被设置为非盘点台账的日志
			loggerx.ProcessLog(c, ActionModifyDatastore, msg.L048, params)
		}
	}

	if req.GetDatastoreName() != "" {
		// 添加多语言数据
		langService := language.NewLanguageService("global", client.DefaultClient)

		languageReq := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    sessionx.GetCurrentApp(c),
			Type:     "datastores",
			Key:      req.GetDatastoreId(),
			Value:    req.GetDatastoreName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}

		_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyDatastore, err)
			return
		}
		loggerx.SuccessLog(c, ActionModifyDatastore, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))
		// 通知刷新多语言数据
		langx.RefreshLanguage(req.Writer, sessionx.GetUserDomain(c))
	}

	code := "I_006"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "refresh",
		Code:    code,
		Content: "台账更新成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + req.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionModifyDatastore, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionModifyDatastore)),
		Data:    response,
	})
}

// ModifyDatastoreSort 更新台账菜单排序
// @Router /datastores/sort [PUT]
func (d *Datastore) ModifyDatastoreSort(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyDatastoreMenuSort, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	var req datastore.MenuSortRequest
	// 从body获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyDatastoreMenuSort, err)
		return
	}

	req.Db = sessionx.GetUserCustomer(c)

	_, err := datastoreService.ModifyDatastoreMenuSort(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyDatastoreMenuSort, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyDatastoreMenuSort, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyDatastoreMenuSort))

	loggerx.InfoLog(c, ActionModifyDatastoreMenuSort, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionModifyDatastoreMenuSort)),
		Data:    req.DatastoresSort,
	})
}

// ModifyDatastoreMapping 更新台账映射
// @Router /datastores/{d_id}/mappings/{m_id} [PUT]
func (d *Datastore) ModifyDatastoreMapping(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyDatastoreMapping, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.ModifyMappingRequest
	// 从body获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyDatastoreMapping, err)
		return
	}
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	req.MappingId = c.Param("m_id")
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.ModifyDatastoreMapping(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyDatastoreMapping, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyDatastoreMapping, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyDatastoreMapping))

	if req.GetMappingName() != "" {
		// 添加多语言数据
		langService := language.NewLanguageService("global", client.DefaultClient)

		languageReq := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    sessionx.GetCurrentApp(c),
			Type:     "mappings",
			Key:      req.GetDatastoreId() + "_" + req.GetMappingId(),
			Value:    req.GetMappingName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}

		_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyDatastoreMapping, err)
			return
		}
		loggerx.SuccessLog(c, ActionModifyDatastoreMapping, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))
		// 通知刷新多语言数据
		langx.RefreshLanguage(sessionx.GetAuthUserID(c), sessionx.GetUserDomain(c))
	}

	loggerx.InfoLog(c, ActionModifyDatastoreMapping, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionModifyDatastoreMapping)),
		Data:    response,
	})
}

// DeleteDatastoreMapping 删除台账映射
// @Router /datastores/{d_id}/mappings/{m_id} [delete]
func (d *Datastore) DeleteDatastoreMapping(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteDatastoreMapping, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DeleteMappingRequest
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	req.MappingId = c.Param("m_id")
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.DeleteDatastoreMapping(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteDatastoreMapping, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteDatastoreMapping, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeleteDatastoreMapping))

	// 删除多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.DeleteAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		AppId:    sessionx.GetCurrentApp(c),
		Type:     "mappings",
		Key:      req.GetDatastoreId() + "_" + req.GetMappingId(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = langService.DeleteAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteDatastoreMapping, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteDatastoreMapping, fmt.Sprintf(loggerx.MsgProcesSucceed, "DeleteAppLanguageData"))

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), sessionx.GetUserDomain(c))

	loggerx.InfoLog(c, ActionDeleteDatastoreMapping, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionDeleteDatastoreMapping)),
		Data:    response,
	})
}

// DeleteUniqueKey 删除台账组合唯一字段
// @Router /datastores/{d_id}/unique/{key} [delete]
func (d *Datastore) DeleteUniqueKey(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteUniqueKey, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DeleteUniqueRequest
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	req.UniqueFields = c.Param("fields")
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.DeleteUniqueKey(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteUniqueKey, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteUniqueKey, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeleteUniqueKey))

	loggerx.InfoLog(c, ActionDeleteUniqueKey, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionDeleteUniqueKey)),
		Data:    response,
	})
}

// DeleteRelation 删除台账组合唯一字段
// @Router /datastores/{d_id}/relation/{r_id} [delete]
func (d *Datastore) DeleteRelation(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteRelation, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DeleteRelationRequest
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	req.RelationId = c.Param("r_id")
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.DeleteRelation(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteRelation, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteRelation, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeleteRelation))

	loggerx.InfoLog(c, ActionDeleteRelation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionDeleteRelation)),
		Data:    response,
	})
}

// HardDeleteDatastores 物理删除多个台账
// @Router /phydel/datastores [delete]
func (d *Datastore) HardDeleteDatastores(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteDatastores, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.HardDeleteDatastoresRequest
	req.DatastoreIdList = c.QueryArray("datastore_id_list")
	db := sessionx.GetUserCustomer(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	appId := sessionx.GetCurrentApp(c)
	req.Database = db

	langData := langx.GetLanguageData(db, lang, domain)

	// 查询要删除的台账信息
	var dreq datastore.DatastoreRequest
	dreq.Database = sessionx.GetUserCustomer(c)
	// 删除前获取台账名称和api_key的集合(保存日志&发送消息用)
	var dsNames []string
	var apiKeys []string
	for _, ds := range req.DatastoreIdList {
		dsNames = append(dsNames, langx.GetLangValue(langData, langx.GetDatastoreKey(appId, ds), langx.DefaultResult))
		dreq.DatastoreId = ds
		dResponse, err := datastoreService.FindDatastore(context.TODO(), &dreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteDatastores, err)
			return
		}
		apiKeys = append(apiKeys, dResponse.GetDatastore().GetApiKey())
	}

	// 删除台账和台账多语言&台账下属各成员和各成员多语言
	response, err := datastoreService.HardDeleteDatastores(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteDatastores, err)
		return
	}
	loggerx.SuccessLog(c, ActionHardDeleteDatastores, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionHardDeleteDatastores))

	// 编辑台账名参数保存日志到DB
	dname := strings.Builder{}
	dname.WriteString(strings.Join(dsNames, ","))
	dname.WriteString("(")
	dname.WriteString(sessionx.GetCurrentLanguage(c))
	dname.WriteString(")")
	// 编辑台账api_key参数保存日志到DB
	apikey := strings.Builder{}
	apikey.WriteString(strings.Join(apiKeys, ","))
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["datastore_name"] = dname.String()
	params["api_key"] = apikey.String()
	loggerx.ProcessLog(c, ActionHardDeleteDatastores, msg.L046, params)

	// 删除app语言
	for _, ds := range req.DatastoreIdList {
		langx.DeleteAppLanguageData(db, domain, appId, "datastores", ds)
	}

	// 删除成功后向画面发送刷新页面提示消息
	code := "I_008"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "refresh",
		Code:    code,
		Content: "台账删除成功，请刷新浏览器获取最新数据！",
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionHardDeleteDatastores, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionHardDeleteDatastores)),
		Data:    response,
	})
}
