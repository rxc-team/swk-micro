package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
)

// Item Item
type Item struct{}

// log出力
const (
	ItemProcessName           = "Item"
	ActionFindItems           = "FindItems"
	ActionFindUnApproveItems  = "FindUnApproveItems"
	ActionChangeOwners        = "ChangeOwners"
	ActionResetInventoryItems = "ResetInventoryItems"
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

// FindUnApproveItems 通过database_Id和status获取未审批数据件数
// @Router /database/{d_id}/items/{id} [get]
func (i *Item) FindUnApproveItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUnApproveItems, loggerx.MsgProcessStarted)

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.UnApproveItemsRequest
	req.DatastoreId = c.Param("d_id")
	req.Status = c.Query("status")
	req.Database = sessionx.GetUserCustomer(c)

	res, err := itemService.FindUnApproveItems(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUnApproveItems, err)
		return
	}

	loggerx.InfoLog(c, ActionFindUnApproveItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionFindUnApproveItems)),
		Data:    res.GetTotal(),
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
		httpx.GinHTTPError(c, ActionModifyGroup, err)
		return
	}

	oGroupInfo := oResponse.GetGroup()
	// 获取变更后group的名称
	var nReq group.FindGroupRequest
	nReq.GroupId = req.GetNewOwner()
	nReq.Database = sessionx.GetUserCustomer(c)
	nResponse, err := groupService.FindGroup(context.TODO(), &nReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyGroup, err)
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
	loggerx.ProcessLog(c, ActionModifyGroup, msg.L008, params)

	loggerx.InfoLog(c, ActionChangeOwners, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionChangeOwners)),
		Data:    response,
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
