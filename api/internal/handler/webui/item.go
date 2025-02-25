package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"

	"rxcsoft.cn/pit3/api/internal/common/csvx"
	"rxcsoft.cn/pit3/api/internal/common/excelx"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/poolx"
	"rxcsoft.cn/pit3/api/internal/common/storex"
	"rxcsoft.cn/pit3/api/internal/common/transferx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
)

// Item Item
type Item struct{}

// log出力
const (
	ItemProcessName            = "Item"
	ActionFindItems            = "FindItems"
	ActionFindItem             = "FindItem"
	ActionFindRishiritsu       = "FindRishiritsu"
	ActionAddItem              = "AddItem"
	ActionMutilAddItem         = "MutilAddItem"
	ActionDownload             = "Download"
	ActionImportItem           = "ImportItem"
	ActionImportCsvItem        = "ImportCsvItem"
	ActionImportCheckItems     = "ImportCheckItems"
	ActionModifyItem           = "ModifyItem"
	ActionModifyContract       = "ModifyContract"
	ActionMutilModifyItem      = "MutilModifyItem"
	ActionChangeDebt           = "ChangeDebt"
	ActionContractExpire       = "ContractExpire"
	ActionTerminateContract    = "TerminateContract"
	ActionInventoryItem        = "InventoryItem"
	ActionMutilInventoryItem   = "MutilInventoryItem"
	ActionResetInventoryItems  = "ResetInventoryItems"
	ActionChangeOwners         = "ChangeOwners"
	ActionChangeSelectOwners   = "ChangeSelectOwners"
	ActionChangeItemOwner      = "ChangeItemOwner"
	ActionChangeLabelTime      = "ChangeLabelTime"
	ActionDeleteItem           = "DeleteItem"
	ActionDeleteDatastoreItems = "DeleteDatastoreItems"
	ActionDeleteSelectedItems  = "DeleteSelectedDatastoreItems"
	ActionGetFields            = "getFields"
	ActionGetAllUser           = "getAllUser"
	ActionGetAppLanguage       = "getAppLanguage"
	ActionImportZipCsvItem     = "ImportZipCsvItem"
)

// FindItems 获取台账中的所有数据
// @Router /datastores/{d_id}/items [get]
func (i *Item) FindItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindItems, loggerx.MsgProcessStarted)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var req item.ItemsRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionFindItems, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("d_id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	accesskey := c.QueryArray("access_key")
	if len(accesskey) > 0 {
		req.Owners = accesskey
	} else {
		req.Owners = sessionx.GetUserAccessKeys(c, req.DatastoreId, "R")
	}
	req.Database = sessionx.GetUserCustomer(c)

	response, err := itemService.FindItems(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindItems, err)
		return
	}

	var res []map[string]interface{}

	for _, item := range response.GetItems() {
		it := make(map[string]interface{})
		it["item_id"] = item.ItemId
		it["app_id"] = item.AppId
		it["datastore_id"] = item.DatastoreId
		it["owners"] = item.Owners
		it["created_at"] = item.CreatedAt
		it["created_by"] = item.CreatedBy
		it["updated_at"] = item.UpdatedAt
		it["updated_by"] = item.UpdatedBy
		it["checked_at"] = item.CheckedAt
		it["checked_by"] = item.CheckedBy
		it["label_time"] = item.LabelTime

		itemMap := make(map[string]interface{})
		for key, value := range item.GetItems() {
			itemMap[key] = value
		}
		it["items"] = itemMap

		res = append(res, it)
	}

	loggerx.InfoLog(c, ActionFindItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionFindItems)),
		Data: gin.H{
			"total":      response.GetTotal(),
			"items_list": res,
		},
	})
}

// FindItem 通过database_Id和item_id获取数据
// @Router /database/{d_id}/items/{id} [get]
func (i *Item) FindItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindItem, loggerx.MsgProcessStarted)

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.ItemRequest
	req.DatastoreId = c.Param("d_id")
	req.ItemId = c.Param("i_id")

	isOrigin := c.Query("is_origin")
	if isOrigin == "true" {
		req.IsOrigin = true
	} else {
		req.IsOrigin = false
	}
	req.Database = sessionx.GetUserCustomer(c)
	req.Owners = sessionx.GetUserAccessKeys(c, req.DatastoreId, "R")

	response, err := itemService.FindItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindItem, err)
		return
	}

	res := make(map[string]interface{})
	res["item_id"] = response.GetItem().ItemId
	res["app_id"] = response.GetItem().AppId
	res["datastore_id"] = response.GetItem().DatastoreId
	res["owners"] = response.GetItem().Owners
	res["created_at"] = response.GetItem().CreatedAt
	res["created_by"] = response.GetItem().CreatedBy
	res["updated_at"] = response.GetItem().UpdatedAt
	res["updated_by"] = response.GetItem().UpdatedBy
	res["checked_at"] = response.GetItem().CheckedAt
	res["checked_by"] = response.GetItem().CheckedBy
	res["check_type"] = response.GetItem().CheckType
	res["check_status"] = response.GetItem().CheckStatus
	res["label_time"] = response.GetItem().LabelTime
	res["status"] = response.GetItem().Status

	itemMap := make(map[string]interface{})
	for key, value := range response.GetItem().GetItems() {
		itemMap[key] = transferx.TransferData(value)
	}
	res["items"] = itemMap

	loggerx.InfoLog(c, ActionFindItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionFindItem)),
		Data:    res,
	})
}

// AddItem 添加台账数据
// @Router /datastores/{d_id}/items [post]
func (i *Item) AddItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddItem, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)

	itemService := item.NewItemService("database", client.DefaultClient)

	var nReq item.AddRequest
	// 从body中获取参数
	if err := c.BindJSON(&nReq); err != nil {
		httpx.GinHTTPError(c, ActionAddItem, err)
		return
	}
	// 从path中获取参数
	nReq.DatastoreId = datastore
	// 从共通中获取参数
	nReq.AppId = appID
	nReq.Owners = sessionx.GetUserOwner(c)
	nReq.Writer = userID
	nReq.LangCd = sessionx.GetCurrentLanguage(c)
	nReq.Domain = domain
	nReq.Database = db

	response, err := itemService.AddItem(context.TODO(), &nReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddItem, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

	code := "I_014"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "normal",
		Code:    code,
		Link:    "/datastores/" + nReq.GetDatastoreId() + "/list",
		Content: "添加数据成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + nReq.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionAddItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionAddItem)),
		Data:    response,
	})
}

// ModifyItem 更新台账一条数据
// @Router /datastores/{d_id}/items/{i_id} [put]
func (i *Item) ModifyItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyItem, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	itemID := c.Param("i_id")
	wfID := c.Query("wf_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	owners := sessionx.GetUserAccessKeys(c, datastore, "W")

	if len(wfID) > 0 {
		itemService := item.NewItemService("database", client.DefaultClient)

		var iReq item.ItemRequest
		iReq.DatastoreId = datastore
		iReq.ItemId = itemID
		iReq.Database = db
		iReq.IsOrigin = true
		iReq.Owners = owners

		iResp, err := itemService.FindItem(context.TODO(), &iReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyItem, err)
			return
		}

		itemMap := map[string]*approve.Value{}
		items := iResp.GetItem().GetItems()

		for key, it := range items {
			if it.GetDataType() == "user" {
				var uList []string
				err := json.Unmarshal([]byte(it.GetValue()), &uList)
				if err != nil {
					itemMap[key] = &approve.Value{
						DataType: it.GetDataType(),
						Value:    "",
					}
				} else {
					itemMap[key] = &approve.Value{
						DataType: it.GetDataType(),
						Value:    strings.Join(uList, ","),
					}
				}
			} else if it.GetDataType() == "lookup" {
				if len(it.GetValue()) > 0 {
					result := strings.Split(it.GetValue(), " : ")
					itemMap[key] = &approve.Value{
						DataType: it.GetDataType(),
						Value:    result[0],
					}
				} else {
					itemMap[key] = &approve.Value{
						DataType: it.GetDataType(),
						Value:    "",
					}
				}
			} else {
				itemMap[key] = &approve.Value{
					DataType: it.GetDataType(),
					Value:    it.GetValue(),
				}
			}
		}

		approveService := approve.NewApproveService("database", client.DefaultClient)

		var req approve.AddRequest
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionModifyItem, err)
			return
		}
		req.Current = req.Items
		req.ItemId = itemID
		req.History = itemMap
		req.DatastoreId = datastore
		req.AppId = appID
		req.Writer = userID
		req.Database = db
		req.Domain = domain
		req.LangCd = sessionx.GetCurrentLanguage(c)

		response, err := approveService.AddItem(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyItem, err)
			return
		}

		// 数据状态转换成待审批状态
		var statusReq item.StatusRequest
		statusReq.AppId = appID
		statusReq.DatastoreId = datastore
		statusReq.ItemId = itemID
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "2"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyItem, err)
			return
		}

		loggerx.SuccessLog(c, ActionModifyItem, fmt.Sprintf("Item[%s] Update Success", response.GetItemId()))

		loggerx.InfoLog(c, ActionAddItem, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionAddItem)),
			Data:    response,
		})
		c.Abort()
		return
	}
	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.ModifyRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyItem, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = datastore
	req.ItemId = itemID
	// 从共通中获取参数
	req.AppId = appID
	req.Writer = userID
	req.Owners = owners
	req.LangCd = sessionx.GetCurrentLanguage(c)
	req.Domain = domain
	req.Database = db

	response, err := itemService.ModifyItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyItem, fmt.Sprintf("item[%s] update success", req.GetItemId()))

	code := "I_016"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "normal",
		Code:    code,
		Link:    "/datastores/" + req.GetDatastoreId() + "/list",
		Content: "更新数据成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + req.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionModifyItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionModifyItem)),
		Data:    response,
	})
}

// ChangeOwners 更新所有者
// @Router /datastores/{d_id}/items [patch]
func (i *Item) ChangeOwners(c *gin.Context) {
	loggerx.InfoLog(c, ActionChangeOwners, loggerx.MsgProcessStarted)

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.OwnersRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionChangeOwners, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("d_id")
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := itemService.ChangeOwners(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionChangeOwners, err)
		return
	}
	loggerx.SuccessLog(c, ActionChangeOwners, fmt.Sprintf("Datastore[%s] ChangeOwners success", req.GetDatastoreId()))

	params := make(map[string]string)

	groupService := group.NewGroupService("manage", client.DefaultClient)

	// 获取变更前group的名称
	var oReq group.FindGroupRequest
	oReq.GroupId = req.GetNewOwner()
	oReq.Database = sessionx.GetUserCustomer(c)
	oResponse, err := groupService.FindGroup(context.TODO(), &oReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionChangeOwners, err)
		return
	}

	oGroupInfo := oResponse.GetGroup()
	// 获取变更后group的名称
	var nReq group.FindGroupRequest
	nReq.GroupId = req.GetNewOwner()
	nReq.Database = sessionx.GetUserCustomer(c)
	nResponse, err := groupService.FindGroup(context.TODO(), &nReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionChangeOwners, err)
		return
	}

	nGroupInfo := nResponse.GetGroup()

	// 获取台账信息
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var dReq datastore.DatastoreRequest
	// 从path获取
	dReq.DatastoreId = c.Param("d_id")
	dReq.Database = sessionx.GetUserCustomer(c)

	dResponse, err := datastoreService.FindDatastore(context.TODO(), &dReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastore, err)
		return
	}

	params["user_name"] = sessionx.GetUserName(c)
	params["group_a"] = "{{" + oGroupInfo.GetGroupName() + "}}"
	params["group_b"] = "{{" + nGroupInfo.GetGroupName() + "}}"
	params["datastore_name"] = "{{" + dResponse.GetDatastore().GetDatastoreName() + "}}"
	params["api_key"] = dResponse.GetDatastore().GetApiKey()

	loggerx.ProcessLog(c, ActionChangeOwners, msg.L008, params)

	loggerx.InfoLog(c, ActionChangeOwners, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionChangeOwners)),
		Data:    response,
	})
}

// ChangeSelectOwners 变更检索到的数据的所有者
// @Router /datastores/{d_id}/items/owners [post]
func (i *Item) ChangeSelectOwners(c *gin.Context) {
	loggerx.InfoLog(c, ActionChangeSelectOwners, loggerx.MsgProcessStarted)
	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.SelectOwnersRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionChangeSelectOwners, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("d_id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)
	req.OldOwners = sessionx.GetUserAccessKeys(c, req.DatastoreId, "R")

	_, err := itemService.ChangeSelectOwners(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionChangeSelectOwners, err)
		return
	}

	loggerx.InfoLog(c, ActionChangeSelectOwners, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionChangeSelectOwners)),
		Data:    nil,
	})
}

// ChangeItemOwner 更新当前itemid条件下的数据的所有者
// @Router /datastores/{d_id}/{i_id}/items/owner [post]
func (i *Item) ChangeItemOwner(c *gin.Context) {
	loggerx.InfoLog(c, ActionChangeItemOwner, loggerx.MsgProcessStarted)
	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.ItemOwnerRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionChangeItemOwner, err)
		return
	}
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	_, err := itemService.ChangeItemOwner(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionChangeItemOwner, err)
		return
	}

	loggerx.InfoLog(c, ActionChangeItemOwner, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionChangeItemOwner)),
		Data:    nil,
	})
}

// ResetInventoryItems 盘点台账盘点数据盘点状态重置
// @Router /apps/{app_id}/inventory/reset [patch]
func (i *Item) ResetInventoryItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionResetInventoryItems, loggerx.MsgProcessStarted)

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.ResetInventoryItemsRequest

	// 从path中获取参数
	req.AppId = c.Param("app_id")
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := itemService.ResetInventoryItems(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionResetInventoryItems, err)
		return
	}
	loggerx.SuccessLog(c, ActionResetInventoryItems, fmt.Sprintf("APP[%s] Inventory Reset success", req.GetAppId()))

	loggerx.InfoLog(c, ActionResetInventoryItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionResetInventoryItems)),
		Data:    response,
	})
}

// ChangeLabelTime 修改标签出力时间
// @Router /changeLabel/datastores/{d_id}/items [put]
func (i *Item) ChangeLabelTime(c *gin.Context) {
	loggerx.InfoLog(c, ActionChangeLabelTime, loggerx.MsgProcessStarted)

	var req item.LabelTimeRequest
	req.DatastoreId = c.Param("d_id")
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionChangeLabelTime, err)
		return
	}
	req.Database = sessionx.GetUserCustomer(c)

	itemService := item.NewItemService("database", client.DefaultClient)
	response, err := itemService.ChangeLabelTime(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionChangeLabelTime, err)
		return
	}
	loggerx.SuccessLog(c, ActionChangeLabelTime, fmt.Sprintf("Items[%s] ChangeLabelTime Success", req.GetItemIdList()))

	loggerx.InfoLog(c, ActionChangeLabelTime, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionChangeLabelTime)),
		Data:    response,
	})
}

// DeleteItem 删除台账数据
// @Router /datastores/{d_id}/items/{i_id} [delete]
func (i *Item) DeleteItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteItem, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	itemID := c.Param("i_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	owners := sessionx.GetUserAccessKeys(c, datastore, "D")

	itemService := item.NewItemService("database", client.DefaultClient)

	// 获取minio文件
	var iReq item.ItemRequest
	iReq.DatastoreId = datastore
	iReq.ItemId = itemID
	iReq.Database = db
	iReq.IsOrigin = true
	iReq.Owners = owners
	iResp, err := itemService.FindItem(context.TODO(), &iReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteItem, err)
		return
	}
	delFileList := []string{}
	for _, k := range iResp.GetItem().GetItems() {
		if k.DataType == "file" {
			var result []typesx.FileValue
			json.Unmarshal([]byte(k.GetValue()), &result)
			for _, m := range result {
				if m.URL != "" {
					delFileList = append(delFileList, m.URL)
				}
			}
		}
	}

	var req item.DeleteRequest
	// 从path中获取参数
	req.DatastoreId = datastore
	req.ItemId = itemID
	// 从共通中获取参数
	req.Writer = userID
	req.Owners = owners
	req.LangCd = sessionx.GetCurrentLanguage(c)
	req.Domain = domain
	req.Database = db

	response, err := itemService.DeleteItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteItem, fmt.Sprintf("item[%s] delete success", req.GetItemId()))

	// 删除minio文件
	_, _, err = filex.DeletePublicDataFiles(domain, appID, delFileList)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteItem, err)
		return
	}

	code := "I_017"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "normal",
		Code:    code,
		Link:    "/datastores/" + req.GetDatastoreId() + "/list",
		Content: "无效化数据成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + req.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionDeleteItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDeleteItem)),
		Data:    response,
	})
}

// DeleteDatastoreItems 删除台账数据所有数据
// @Router /clear/datastores/{d_id}/items [delete]
func (i *Item) DeleteDatastoreItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	datastoreId := c.Param("d_id")
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)

	// 查找台账数据grpc
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)
	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	var req item.DeleteDatastoreItemsRequest
	// 从path中获取参数
	req.DatastoreId = datastoreId
	// 从共通中获取参数
	req.Writer = userID
	req.Database = db

	response, err := itemService.DeleteDatastoreItems(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteDatastoreItems, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteDatastoreItems, fmt.Sprintf("Datastore[%s] all data delete success", req.GetDatastoreId()))

	// 根据上文累计的冗余minio文件删除冗余minio文件
	go filex.DeleteDatastoreFiles(domain, appID, datastoreId)

	code := "I_018"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "normal",
		Code:    code,
		Link:    "/datastores/" + req.GetDatastoreId() + "/list",
		Content: "删除数据成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + req.GetDatastoreId(),
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDeleteDatastoreItems)),
		Data:    response,
	})
}

// DeleteAllDatastoreItems 删除台账数据所有数据
// @Router /clear/datastores/clearAll [delete]
func (i *Item) DeleteAllDatastoreItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	jobID := "job_" + time.Now().Format("20060102150405")
	lang := sessionx.GetCurrentLanguage(c)
	datastoreID := c.QueryArray("datastore_id")
	store := storex.NewRedisStore(600)
	clearID := appID + "clear"
	uploadID := appID + "upload"
	valUpload := store.Get(uploadID, false)

	if len(valUpload) == 0 {
		store.Set(clearID, "clear")
		go func() {
			// 创建任务
			jobx.CreateTask(task.AddRequest{
				JobId:        jobID,
				JobName:      "DeleteDatastoreItems",
				Origin:       "Delete Datastore",
				UserId:       userID,
				ShowProgress: false,
				Message:      i18n.Tr(lang, "job.J_014"),
				TaskType:     "Delete-Datastore",
				Steps:        []string{"start", "clear", "delete-Datastore", "delete-keiyakudaicho", "end"},
				CurrentStep:  "start",
				Database:     db,
				AppId:        appID,
			})

			// 发送消息 数据准备
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "クリア処理中",
				CurrentStep: "clear",
				Database:    db,
			}, userID)

			wg := sync.WaitGroup{}
			wg.Add(3)

			go func() {
				defer wg.Done()
				// 查找台账数据grpc
				ct := grpc.NewClient(
					grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
				)

				itemService := item.NewItemService("database", ct)
				var opss client.CallOption = func(o *client.CallOptions) {
					o.RequestTimeout = time.Hour * 1
					o.DialTimeout = time.Hour * 1
				}

				var req item.DeleteDatastoreItemsRequest
				// 从path中获取参数
				req.DatastoreId = datastoreID[1]
				// 从共通中获取参数
				req.Writer = userID
				req.Database = db

				_, err := itemService.DeleteDatastoreItems(context.TODO(), &req, opss)
				if err != nil {
					httpx.GinHTTPError(c, ActionDeleteDatastoreItems, err)
					return
				}

				// 根据上文累计的冗余minio文件删除冗余minio文件
				go filex.DeleteDatastoreFiles(domain, appID, datastoreID[1])

				loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessEnded)

				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "支払状況削除しました",
					CurrentStep: "delete-Datastore",
					Database:    db,
				}, userID)
			}()

			go func() {
				defer wg.Done()
				// 查找台账数据grpc
				ct := grpc.NewClient(
					grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
				)

				itemService := item.NewItemService("database", ct)
				var opss client.CallOption = func(o *client.CallOptions) {
					o.RequestTimeout = time.Hour * 1
					o.DialTimeout = time.Hour * 1
				}

				var req item.DeleteDatastoreItemsRequest
				// 从path中获取参数
				req.DatastoreId = datastoreID[2]
				// 从共通中获取参数
				req.Writer = userID
				req.Database = db

				_, err := itemService.DeleteDatastoreItems(context.TODO(), &req, opss)
				if err != nil {
					httpx.GinHTTPError(c, ActionDeleteDatastoreItems, err)
					return
				}

				// 根据上文累计的冗余minio文件删除冗余minio文件
				go filex.DeleteDatastoreFiles(domain, appID, datastoreID[2])

				loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessEnded)

				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "試算結果削除しました",
					CurrentStep: "delete-Datastore",
					Database:    db,
				}, userID)
			}()

			go func() {
				defer wg.Done()
				// 查找台账数据grpc
				ct := grpc.NewClient(
					grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
				)

				itemService := item.NewItemService("database", ct)
				var opss client.CallOption = func(o *client.CallOptions) {
					o.RequestTimeout = time.Hour * 1
					o.DialTimeout = time.Hour * 1
				}

				var req item.DeleteDatastoreItemsRequest
				// 从path中获取参数
				req.DatastoreId = datastoreID[3]
				// 从共通中获取参数
				req.Writer = userID
				req.Database = db

				_, err := itemService.DeleteDatastoreItems(context.TODO(), &req, opss)
				if err != nil {
					httpx.GinHTTPError(c, ActionDeleteDatastoreItems, err)
					return
				}

				// 根据上文累计的冗余minio文件删除冗余minio文件
				go filex.DeleteDatastoreFiles(domain, appID, datastoreID[3])

				loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessEnded)

				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "償却状況削除しました",
					CurrentStep: "delete-Datastore",
					Database:    db,
				}, userID)
			}()

			wg.Wait()

			// 查找台账数据grpc
			ct := grpc.NewClient(
				grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
			)

			itemService := item.NewItemService("database", ct)
			var opss client.CallOption = func(o *client.CallOptions) {
				o.RequestTimeout = time.Hour * 1
				o.DialTimeout = time.Hour * 1
			}

			var req item.DeleteDatastoreItemsRequest
			// 从path中获取参数
			req.DatastoreId = datastoreID[0]
			// 从共通中获取参数
			req.Writer = userID
			req.Database = db

			_, err := itemService.DeleteDatastoreItems(context.TODO(), &req, opss)
			if err != nil {
				httpx.GinHTTPError(c, ActionDeleteDatastoreItems, err)
				return
			}

			// 根据上文累计的冗余minio文件删除冗余minio文件
			go filex.DeleteDatastoreFiles(domain, appID, datastoreID[0])

			loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessEnded)

			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "契約情報削除しました",
				CurrentStep: "delete-keiyakudaicho",
				Database:    db,
			}, userID)

			// 发送消息 删除成功，任务结束
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_028"),
				CurrentStep: "end",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database:    db,
			}, userID)
			store.Set(clearID, "")
		}()

		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDeleteDatastoreItems)),
			Data:    gin.H{},
		})
	} else {
		loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDeleteDatastoreItems)),
			Data: gin.H{
				"msg": "アップロード処理中のため、クリア処理が出来ません。",
			},
		})
	}
}

// DeleteSelectedItems 删除台账数据所有数据
// @Router /clear/datastores/{d_id}/items/selected [delete]
func (i *Item) DeleteSelectedItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteDatastoreItems, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	datastoreId := c.Param("d_id")
	itemIDSet := c.QueryArray("itemelected")
	appID := sessionx.GetCurrentApp(c)
	domain := sessionx.GetUserDomain(c)
	// 查找台账数据grpc
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)
	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	var selectItemReq item.SelectedItemsRequest
	selectItemReq.AppId = appID
	selectItemReq.DatastoreId = datastoreId
	selectItemReq.Database = db
	selectItemReq.LangCd = sessionx.GetCurrentLanguage(c)
	selectItemReq.Domain = domain
	selectItemReq.ItemIdList = itemIDSet
	stream, err := itemService.DeleteSelectItems(context.TODO(), &selectItemReq, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectedItems, err)
		return
	}
	for {
		selectUrl, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				loggerx.ErrorLog("Script0324", err.Error())
				return
			}
		}
		// 删除图片
		index := strings.Index(selectUrl.GetDeleteUrl(), domain)
		if index != -1 {
			fileName := selectUrl.GetDeleteUrl()[index+len(domain)+1:]
			filex.DeleteFile(domain, fileName)
		}
	}
	stream.Close()

	loggerx.SuccessLog(c, ActionDeleteSelectedItems, fmt.Sprintf("Datastore[%s] all data delete success", datastoreId))

	code := "I_018"
	param := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  sessionx.GetUserDomain(c),
		MsgType: "normal",
		Code:    code,
		Link:    "/datastores/" + datastoreId + "/list",
		Content: "删除数据成功，请刷新浏览器获取最新数据！",
		Object:  "apps." + sessionx.GetCurrentApp(c) + ".datastores." + datastoreId,
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(param, sessionx.GetUserCustomer(c), sessionx.GetUserGroup(c))

	loggerx.InfoLog(c, ActionDeleteSelectedItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDeleteSelectedItems)),
	})
}

// Download 获取台账中的所有数据,以csv文件的方式下载
// @Router /datastores/{d_id}/items [get]
func (i *Item) Download(c *gin.Context) {

	type (
		// DownloadRequest 下载
		DownloadRequest struct {
			ItemCondition item.ItemsRequest `json:"item_condition" bson:"item_condition"`
		}
	)

	loggerx.InfoLog(c, ActionDownload, loggerx.MsgProcessStarted)

	jobID := c.Query("job_id")
	datastoreID := c.Param("d_id")
	appID := sessionx.GetCurrentApp(c)
	owners := sessionx.GetUserAccessKeys(c, datastoreID, "R")
	userID := sessionx.GetAuthUserID(c)
	roles := sessionx.GetUserRoles(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	db := sessionx.GetUserCustomer(c)
	encoding := "utf-8"
	fileType := "csv"
	appRoot := "app_" + appID

	// 从body中获取参数
	var request DownloadRequest
	if err := c.BindJSON(&request); err != nil {
		httpx.GinHTTPError(c, ActionDownload, err)
		return
	}

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "datastore file download",
		Origin:       "apps." + appID + ".datastores." + datastoreID,
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "ds-csv-download",
		Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	sp, err := poolx.NewSystemPool()
	if err != nil {
		httpx.GinHTTPError(c, ActionDownload, err)
		return
	}

	if sp.Free() == 0 {
		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "サーバービジー、処理待ちです",
			CurrentStep: "build-data",
			Database:    db,
		}, userID)
	}

	syncRun := func() {
		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		cReq := item.CountRequest{
			AppId:         appID,
			DatastoreId:   datastoreID,
			ConditionList: request.ItemCondition.ConditionList,
			ConditionType: request.ItemCondition.ConditionType,
			Owners:        owners,
			Database:      db,
		}

		cResp, err := itemService.FindCount(context.TODO(), &cReq, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		dReq := item.DownloadRequest{
			AppId:         appID,
			DatastoreId:   datastoreID,
			ConditionList: request.ItemCondition.ConditionList,
			ConditionType: request.ItemCondition.ConditionType,
			Owners:        owners,
			Database:      db,
		}

		stream, err := itemService.Download(context.TODO(), &dReq, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 获取当前台账的字段数据
		var fields []*typesx.DownloadField

		fieldList := fieldx.GetFields(db, datastoreID, appID, roles, false, false)
		for _, f := range fieldList {
			fields = append(fields, &typesx.DownloadField{
				FieldId:       f.FieldId,
				FieldName:     f.FieldName,
				FieldType:     f.FieldType,
				IsImage:       f.IsImage,
				AsTitle:       f.AsTitle,
				DisplayOrder:  f.DisplayOrder,
				DisplayDigits: f.DisplayDigits,
				Precision:     f.Precision,
				Prefix:        f.Prefix,
			})
		}

		// 排序
		sort.Sort(typesx.DownloadFields(fields))
		// 获取当前app的语言数据
		langData := langx.GetLanguageData(db, lang, domain)

		timestamp := time.Now().Format("20060102150405")

		// 每次2000为一组数据
		total := cResp.GetTotal()

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_033"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		// 设定csv头部
		var header []string
		var apiKeyheader []string
		var headers [][]string

		// 添加ID
		header = append(header, "ID")
		apiKeyheader = append(apiKeyheader, "id")
		for _, fl := range fields {
			header = append(header, langx.GetLangValue(langData, fl.FieldName, langx.DefaultResult))
			apiKeyheader = append(apiKeyheader, fl.FieldId)
		}

		headers = append(headers, header)
		headers = append(headers, apiKeyheader)

		// Excel文件下载
		if fileType == "xlsx" {

			excelFile := excelize.NewFile()
			// 创建一个工作表
			index := excelFile.NewSheet("Sheet1")
			// 设置工作簿的默认工作表
			excelFile.SetActiveSheet(index)

			for i, rows := range headers {
				for j, v := range rows {
					y := excelx.GetAxisY(j+1) + strconv.Itoa(i+1)
					excelFile.SetCellValue("Sheet1", y, v)
				}
			}

			var current int = 0
			var items [][]string
			line := 0

			for {
				it, err := stream.Recv()
				if err == io.EOF {
					// 当前结束了，但是items还有数据
					if len(items) > 0 {

						// 返回消息
						result := make(map[string]interface{})

						result["total"] = total
						result["current"] = current

						message, _ := json.Marshal(result)

						// 发送消息 写入条数
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     string(message),
							CurrentStep: "write-to-file",
							Database:    db,
						}, userID)

						// 写入数据
						for k, rows := range items {
							for j, v := range rows {
								y := excelx.GetAxisY(j+1) + strconv.Itoa(line*500+k+3)
								excelFile.SetCellValue("Sheet1", y, v)
							}
						}
					}
					break
				}

				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "write-to-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
				current++
				dt := it.GetItem()

				{
					// 设置csv行
					var itemData []string
					// 添加ID
					itemData = append(itemData, dt.GetItemId())
					for _, fl := range fields {
						// 使用默认的值的情况
						if fl.FieldId == "#" {
							itemData = append(itemData, fl.Prefix)
						} else {
							itemMap := dt.GetItems()
							if value, ok := itemMap[fl.FieldId]; ok {
								result := ""
								switch value.DataType {
								case "text", "textarea", "number", "time", "switch":
									result = value.GetValue()
								case "autonum":
									result = value.GetValue()
								case "lookup":
									result = value.GetValue()
								case "options":
									result = langx.GetLangValue(langData, value.GetValue(), langx.DefaultResult)
								case "date":
									if value.GetValue() == "0001-01-01" {
										result = ""
									} else {
										if len(fl.Format) > 0 {
											date, err := time.Parse("2006-01-02", value.GetValue())
											if err != nil {
												result = ""
											} else {

												result = date.Format(fl.Format)
											}
										} else {
											result = value.GetValue()
										}
									}
								case "user":
									var userStrList []string
									json.Unmarshal([]byte(value.GetValue()), &userStrList)
									result = strings.Join(userStrList, ",")
								case "file":
									var files []typesx.FileValue
									json.Unmarshal([]byte(value.GetValue()), &files)
									var fileStrList []string
									for _, f := range files {
										fileStrList = append(fileStrList, f.Name)
									}
									result = strings.Join(fileStrList, ",")
								default:
									break
								}

								itemData = append(itemData, result)
							} else {
								itemData = append(itemData, "")
							}
						}
					}
					// 添加行
					items = append(items, itemData)

				}

				if current%500 == 0 {
					// 返回消息
					result := make(map[string]interface{})

					result["total"] = total
					result["current"] = current

					message, _ := json.Marshal(result)

					// 发送消息 写入条数
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     string(message),
						CurrentStep: "write-to-file",
						Database:    db,
					}, userID)

					// 写入数据
					for k, rows := range items {
						for j, v := range rows {
							y := excelx.GetAxisY(j+1) + strconv.Itoa(line*500+k+3)
							excelFile.SetCellValue("Sheet1", y, v)
						}
					}

					// 清空items
					items = items[:0]

					line++
				}
			}

			defer stream.Close()

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_029"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			filex.Mkdir("temp/")
			// 写入文件到本地
			outFile := "temp/tmp" + "_" + timestamp + ".xlsx"

			if err := excelFile.SaveAs(outFile); err != nil {
				fmt.Println(err)
			}

			fo, err := os.Open(outFile)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			defer fo.Close()

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_043"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			// 写入文件到 minio
			minioClient, err := storagecli.NewClient(domain)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			filePath := path.Join(appRoot, "excel", "datastore_"+timestamp+".xlsx")
			path, err := minioClient.SavePublicObject(fo, filePath, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			// 判断顾客上传文件是否在设置的最大存储空间以内
			canUpload := filex.CheckCanUpload(domain, float64(path.Size))
			if canUpload {
				// 如果没有超出最大值，就对顾客的已使用大小进行累加
				err = filex.ModifyUsedSize(domain, float64(path.Size))
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 保存文件失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "save-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
			} else {
				// 如果已达上限，则删除刚才上传的文件
				minioClient.DeleteObject(path.Name)
				path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     i18n.Tr(lang, "job.J_007"),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			os.Remove(outFile)

			// 发送消息 写入保存文件成功，返回下载路径，任务结束
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_028"),
				CurrentStep: "end",
				File: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

		} else {

			var writer *csvx.SyncWriter

			filex.Mkdir("temp/")

			// 写入文件到本地
			filename := "temp/tmp" + "_" + timestamp + ".csv"
			f, err := os.Create(filename)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "write-to-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			if encoding == "sjis" {
				converter := transform.NewWriter(f, japanese.ShiftJIS.NewEncoder())
				writer = csvx.NewSyncWriter(converter)
			} else {
				writer = csvx.NewSyncWriter(f)
				// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
				headers[0][0] = "\xEF\xBB\xBF" + headers[0][0]
			}

			err = writer.WriteAll(headers)
			if err != nil {
				if err.Error() == "encoding: rune not supported by encoding." {
					path := filex.WriteAndSaveFile(domain, appID, []string{"現在のタイトルには、日本語の[shift-jis]エンコード以外の文字が含まれており、実行を続行できません。"})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "write-to-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}

				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "write-to-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			writer.Flush() // 此时才会将缓冲区数据写入

			var current int = 0
			var items [][]string

			for {
				it, err := stream.Recv()
				if err == io.EOF {
					// 当前结束了，但是items还有数据
					if len(items) > 0 {

						// 返回消息
						result := make(map[string]interface{})

						result["total"] = total
						result["current"] = current

						message, _ := json.Marshal(result)

						// 发送消息 写入条数
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     string(message),
							CurrentStep: "write-to-file",
							Database:    db,
						}, userID)

						// 写入数据
						err = writer.WriteAll(items)
						if err != nil {
							if err.Error() == "encoding: rune not supported by encoding." {
								path := filex.WriteAndSaveFile(domain, appID, []string{"現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"})
								// 发送消息 获取数据失败，终止任务
								jobx.ModifyTask(task.ModifyRequest{
									JobId:       jobID,
									Message:     err.Error(),
									CurrentStep: "write-to-file",
									EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
									ErrorFile: &task.File{
										Url:  path.MediaLink,
										Name: path.Name,
									},
									Database: db,
								}, userID)

								return
							}

							path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "write-to-file",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: db,
							}, userID)

							return
						}

						// 缓冲区数据写入
						writer.Flush()
					}
					break
				}

				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "write-to-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
				current++
				dt := it.GetItem()

				{
					// 设置csv行
					var itemData []string
					// 添加ID
					itemData = append(itemData, dt.GetItemId())
					for _, fl := range fields {
						// 使用默认的值的情况
						if fl.FieldId == "#" {
							itemData = append(itemData, fl.Prefix)
						} else {
							itemMap := dt.GetItems()
							if value, ok := itemMap[fl.FieldId]; ok {
								result := ""
								switch value.DataType {
								case "text", "textarea", "number", "time", "switch":
									result = value.GetValue()
								case "autonum":
									result = value.GetValue()
								case "lookup":
									result = value.GetValue()
								case "options":
									result = langx.GetLangValue(langData, value.GetValue(), langx.DefaultResult)
								case "date":
									if value.GetValue() == "0001-01-01" {
										result = ""
									} else {
										if len(fl.Format) > 0 {
											date, err := time.Parse("2006-01-02", value.GetValue())
											if err != nil {
												result = ""
											} else {

												result = date.Format(fl.Format)
											}
										} else {
											result = value.GetValue()
										}
									}
								case "user":
									var userStrList []string
									json.Unmarshal([]byte(value.GetValue()), &userStrList)
									result = strings.Join(userStrList, ",")
								case "file":
									var files []typesx.FileValue
									json.Unmarshal([]byte(value.GetValue()), &files)
									var fileStrList []string
									for _, f := range files {
										fileStrList = append(fileStrList, f.Name)
									}
									result = strings.Join(fileStrList, ",")
								default:
									break
								}

								itemData = append(itemData, result)
							} else {
								itemData = append(itemData, "")
							}
						}
					}
					// 添加行
					items = append(items, itemData)

				}

				if current%500 == 0 {
					// 返回消息
					result := make(map[string]interface{})

					result["total"] = total
					result["current"] = current

					message, _ := json.Marshal(result)

					// 发送消息 写入条数
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     string(message),
						CurrentStep: "write-to-file",
						Database:    db,
					}, userID)

					// 写入数据
					// 写入数据
					err = writer.WriteAll(items)
					if err != nil {
						if err.Error() == "encoding: rune not supported by encoding." {
							path := filex.WriteAndSaveFile(domain, appID, []string{"現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "write-to-file",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: db,
							}, userID)

							return
						}

						path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     err.Error(),
							CurrentStep: "write-to-file",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: db,
						}, userID)

						return
					}

					// 缓冲区数据写入
					writer.Flush()

					// 清空items
					items = items[:0]
				}
			}
			defer stream.Close()
			defer f.Close()

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_029"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_043"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			fo, err := os.Open(filename)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			defer func() {
				fo.Close()
				os.Remove(filename)
			}()

			// 写入文件到 minio
			minioClient, err := storagecli.NewClient(domain)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			filePath := path.Join(appRoot, "csv", "datastore_"+timestamp+".csv")
			path, err := minioClient.SavePublicObject(fo, filePath, "text/csv")
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			// 判断顾客上传文件是否在设置的最大存储空间以内
			canUpload := filex.CheckCanUpload(domain, float64(path.Size))
			if canUpload {
				// 如果没有超出最大值，就对顾客的已使用大小进行累加
				err = filex.ModifyUsedSize(domain, float64(path.Size))
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 保存文件失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "save-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
			} else {
				// 如果已达上限，则删除刚才上传的文件
				minioClient.DeleteObject(path.Name)
				path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     i18n.Tr(lang, "job.J_007"),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			// 发送消息 写入保存文件成功，返回下载路径，任务结束
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_028"),
				CurrentStep: "end",
				File: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

		}
	}

	err = sp.Submit(syncRun)
	if err != nil {
		httpx.GinHTTPError(c, ActionDownload, err)
		return
	}

	// 设置下载的文件名
	loggerx.InfoLog(c, ActionDownload, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDownload)),
		Data:    gin.H{},
	})
}
