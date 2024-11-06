package handler

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"

	"rxcsoft.cn/pit3/api/outer/common/filex"
	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/jobx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/api/outer/system/wfx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// Item Item
type Item struct{}

// log出力
const (
	ItemProcessName          = "Item"
	ActionFindItems          = "FindItems"
	ActionFindItem           = "FindItem"
	ActionAddItem            = "AddItem"
	ActionImportCsvItem      = "ImportCsvItem"
	ActionModifyItem         = "ModifyItem"
	ActionInventoryItem      = "InventoryItem"
	ActionMutilInventoryItem = "MutilInventoryItem"
	ActionDeleteItem         = "DeleteItem"
)

// FindItems 获取台账中的所有数据
// @Summary 获取台账中的所有数据
// @description 调用srv中的database服务，获取所有的数据
// @Tags Item
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "台账ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
		it["check_status"] = item.CheckStatus
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
// @Summary 通过database_Id和item_id获取数据
// @description 调用srv中的database服务，通过database_Id和item_id获取数据
// @Tags Database
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "台账ID"
// @Param id path string true "数据ID/字段ID"
// @Param type path string true "查询类型"
// @Param value query string false "字段值"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
	req.Owners = sessionx.GetUserAccessKeys(c, req.DatastoreId, "R")
	req.Database = sessionx.GetUserCustomer(c)

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
		itemMap[key] = value
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
// @Summary 添加台账数据
// @description 调用srv中的item服务，添加台账数据
// @Tags Item
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "DatastoreID"
// @Param item body item.AddRequest true "台账数据信息"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /datastores/{d_id}/items [post]
func (i *Item) AddItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddItem, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	groupID := sessionx.GetUserGroup(c)

	wks := wfx.GetUserWorkflow(db, groupID, appID, datastore, "insert")
	if len(wks) > 0 {
		approveService := approve.NewApproveService("database", client.DefaultClient)

		var req approve.AddRequest
		// 从body中获取参数
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionAddItem, err)
			return
		}
		req.Current = req.Items
		// 从共通中获取参数
		req.DatastoreId = datastore
		req.AppId = appID
		req.Writer = userID
		req.Database = db
		req.Domain = domain
		req.LangCd = sessionx.GetCurrentLanguage(c)
		// 开启流程
		approve := new(wfx.Approve)
		// 添加流程实例
		exID, err := approve.AddExample(db, wks[0].GetWfId(), userID)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddItem, err)
			return
		}
		req.ExampleId = exID
		response, err := approveService.AddItem(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddItem, err)
			return
		}
		// 流程开始启动
		err = approve.StartExampleInstance(db, wks[0].GetWfId(), userID, exID, domain)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddItem, err)
			return
		}
		loggerx.SuccessLog(c, ActionAddItem, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

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
	nReq.Database = db

	response, err := itemService.AddItem(context.TODO(), &nReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddItem, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

	loggerx.InfoLog(c, ActionAddItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionAddItem)),
		Data:    response,
	})
}

// ModifyItem 更新台账一条数据
// @Summary 更新台账一条数据
// @description 调用srv中的item服务，更新台账一条数据
// @Tags Item
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "DatastoreID"
// @Param i_id path string true "ItemID"
// @Param item body item.AddRequest true "台账数据信息"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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

	if len(wfID) == 0 {
		groupID := sessionx.GetUserGroup(c)
		wks := wfx.GetUserWorkflow(db, groupID, appID, datastore, "update")
		for _, wk := range wks {
			if len(wk.Params["fields"]) == 0 {
				wfID = wk.GetWfId()
			}
		}
	}

	if len(wfID) > 0 && wfx.CheckWfValid(db, wfID) {

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
			itemMap[key] = &approve.Value{
				DataType: it.GetDataType(),
				Value:    it.GetValue(),
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
		// 开启流程
		approve := new(wfx.Approve)
		// 添加流程实例
		exID, err := approve.AddExample(db, wfID, userID)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyItem, err)
			return
		}
		req.ExampleId = exID
		response, err := approveService.AddItem(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyItem, err)
			return
		}
		// 流程开始启动
		err = approve.StartExampleInstance(db, wfID, userID, exID, domain)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyItem, err)
			return
		}
		loggerx.SuccessLog(c, ActionModifyItem, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

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

		loggerx.InfoLog(c, ActionModifyItem, loggerx.MsgProcessEnded)
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
	req.Database = db

	response, err := itemService.ModifyItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyItem, fmt.Sprintf("item[%s] update success", req.GetItemId()))

	loggerx.InfoLog(c, ActionModifyItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionModifyItem)),
		Data:    response,
	})
}

// InventoryItem 扫描更新台账一条数据
// @Summary 扫描更新台账一条数据
// @description 调用srv中的item服务，扫描更新台账一条数据
// @Tags Item
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "DatastoreID"
// @Param i_id path string true "ItemID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /datastores/{d_id}/items/{i_id} [patch]
func (i *Item) InventoryItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionInventoryItem, loggerx.MsgProcessStarted)

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.InventoryItemRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionInventoryItem, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("d_id")
	req.ItemId = c.Param("i_id")
	// 从共通中获取参数
	if req.AppId == "" {
		// 从共通获取
		req.AppId = sessionx.GetCurrentApp(c)
	}
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := itemService.InventoryItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionInventoryItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionInventoryItem, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionInventoryItem))

	loggerx.InfoLog(c, ActionInventoryItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionInventoryItem)),
		Data:    response,
	})
}

// MutilInventoryItem 扫描更新多条台账数据
// @Summary 扫描更新多条台账数据
// @description 调用srv中的item服务，扫描更新多条台账数据
// @Tags Item
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "DatastoreID"
// @Param i_id path string true "ItemID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /datastores/{d_id}/items/{i_id}/mutil [patch]
func (i *Item) MutilInventoryItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionMutilInventoryItem, loggerx.MsgProcessStarted)

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.MutilInventoryItemRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionMutilInventoryItem, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("d_id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := itemService.MutilInventoryItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionMutilInventoryItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionMutilInventoryItem, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionMutilInventoryItem))

	loggerx.InfoLog(c, ActionMutilInventoryItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionMutilInventoryItem)),
		Data:    response,
	})
}

// DeleteItem 删除台账数据
// @Summary 删除台账数据
// @description 调用srv中的item服务，删除台账数据
// @Tags Item
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "台账ID"
// @Param i_id path string true "数据ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /datastores/{d_id}/items/{i_id} [delete]
func (i *Item) DeleteItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteItem, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	itemID := c.Param("i_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	groupID := sessionx.GetUserGroup(c)
	owners := sessionx.GetUserAccessKeys(c, datastore, "D")
	// 获取当前用户的group信息
	wks := wfx.GetUserWorkflow(db, groupID, appID, datastore, "delete")
	if len(wks) > 0 {

		itemService := item.NewItemService("database", client.DefaultClient)

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

		items := iResp.GetItem().GetItems()

		itemMap := map[string]*approve.Value{}
		for key, it := range items {
			itemMap[key] = &approve.Value{
				DataType: it.GetDataType(),
				Value:    it.GetValue(),
			}
		}

		approveService := approve.NewApproveService("database", client.DefaultClient)

		var req approve.AddRequest
		req.ItemId = itemID
		req.Items = itemMap
		req.Current = req.Items
		req.DatastoreId = datastore
		req.AppId = appID
		req.Writer = userID
		req.Database = db
		req.Domain = domain
		req.LangCd = sessionx.GetCurrentLanguage(c)
		// 开启流程
		approve := new(wfx.Approve)
		// 添加流程实例
		exID, err := approve.AddExample(db, wks[0].GetWfId(), userID)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteItem, err)
			return
		}
		req.ExampleId = exID
		response, err := approveService.AddItem(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteItem, err)
			return
		}

		var statusReq item.StatusRequest
		statusReq.AppId = appID
		statusReq.DatastoreId = datastore
		statusReq.ItemId = itemID
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "2"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteItem, err)
			return
		}

		// 流程开始启动
		err = approve.StartExampleInstance(db, wks[0].GetWfId(), userID, exID, domain)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteItem, err)
			return
		}
		loggerx.SuccessLog(c, ActionDeleteItem, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

		loggerx.InfoLog(c, ActionDeleteItem, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDeleteItem)),
			Data:    response,
		})
		c.Abort()
		return
	}

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.DeleteRequest
	// 从path中获取参数
	req.DatastoreId = datastore
	req.ItemId = itemID
	// 从共通中获取参数
	req.Owners = owners
	req.Writer = userID
	req.Database = db

	response, err := itemService.DeleteItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteItem, fmt.Sprintf("item[%s] delete success", req.GetItemId()))

	loggerx.InfoLog(c, ActionDeleteItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionDeleteItem)),
		Data:    response,
	})
}

// ImportCsvItem 直接上传csv文件进行台账数据
// @Router /import/csv/datastores/{d_id}/items [post]
func (i *Item) ImportCsvItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionImportCsvItem, loggerx.MsgProcessStarted)

	jobID := c.PostForm("job_id")
	datastoreID := c.Param("d_id")
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	appRoot := "app_" + appID
	// lang := sessionx.GetCurrentLanguage(c)
	action := c.PostForm("action")

	// 时间戳
	timestamp := time.Now().Format("20060102150405")

	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "csv file import (" + action + ")",
		Origin:       "apps." + appID + ".datastores." + datastoreID,
		UserId:       userID,
		ShowProgress: true,
		// Message:      i18n.Tr(lang, "job.J_014"),
		Message:     "create a job",
		TaskType:    "ds-csv-import",
		Steps:       []string{"start", "save-file", "unzip-file", "read-file", "check-data", "upload", "end"},
		CurrentStep: "start",
		Database:    db,
		AppId:       appID,
	})

	var zipFilePath string
	var payFilePath string

	// 超级域名
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionImportCsvItem, err)
		return
	}

	// 支付数据文件
	payFile, err := c.FormFile("payFile")
	if err != nil {
		if err.Error() == "http: no such file" {
			payFilePath = ""
		} else {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId: jobID,
				// Message:     i18n.Tr(lang, "job.J_053"),
				Message:     "An error occurred while uploading the file",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionImportCsvItem, err)
			return
		}
	} else {
		// 文件类型检查
		if !filex.CheckSupport("csv", payFile.Header.Get("content-type")) {
			path := filex.WriteAndSaveFile(domain, appID, []string{fmt.Sprintf("the pay file type [%v] is not supported", payFile.Header.Get("content-type"))})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイルのアップロード中にエラーが発生しました。",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			httpx.GinHTTPError(c, ActionImportCsvItem, errors.New("このファイルタイプのアップロードはサポートされていません"))
			return
		}
		// 文件大小检查
		if !filex.CheckSize(domain, "csv", payFile.Size) {
			path := filex.WriteAndSaveFile(domain, appID, []string{"the pay file ファイルサイズが設定サイズを超えています"})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイルのアップロード中にエラーが発生しました。",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			httpx.GinHTTPError(c, ActionImportCsvItem, errors.New("ファイルサイズが設定サイズを超えています"))
			return
		}

		fo, err := payFile.Open()
		if err != nil {
			httpx.GinHTTPError(c, ActionImportCsvItem, err)
			return
		}

		filePath := path.Join(appRoot, "temp", "temp_"+timestamp+payFile.Filename)
		file, err := minioClient.SavePublicObject(fo, filePath, payFile.Header.Get("content-type"))
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId: jobID,
				// Message:     i18n.Tr(lang, "job.J_053"),
				Message:     "An error occurred while uploading the file",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionImportCsvItem, err)
			return
		}

		payFilePath = file.Name
	}

	// zip压缩文件
	zipFile, err := c.FormFile("zipFile")
	if err != nil {
		if err.Error() == "http: no such file" {
			zipFilePath = ""
		} else {
			// 处理出错删除minio已上传文件
			minioClient.DeleteObject(payFilePath)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId: jobID,
				// Message:     i18n.Tr(lang, "job.J_053"),
				Message:     "An error occurred while uploading the file",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionImportCsvItem, err)
			return
		}
	} else {
		// 文件类型检查
		if !filex.CheckSupport("zip", zipFile.Header.Get("content-type")) {
			// 处理出错删除minio已上传文件
			minioClient.DeleteObject(payFilePath)
			path := filex.WriteAndSaveFile(domain, appID, []string{fmt.Sprintf("the zip file type [%v] is not supported", zipFile.Header.Get("content-type"))})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイルのアップロード中にエラーが発生しました。",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			httpx.GinHTTPError(c, ActionImportCsvItem, errors.New("このファイルタイプのアップロードはサポートされていません"))
			return
		}
		// 文件大小检查
		if !filex.CheckSize(domain, "zip", zipFile.Size) {
			// 处理出错删除minio已上传文件
			minioClient.DeleteObject(payFilePath)
			path := filex.WriteAndSaveFile(domain, appID, []string{"the zip file maximum size of uploaded file is 1G"})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイルのアップロード中にエラーが発生しました。",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			httpx.GinHTTPError(c, ActionImportCsvItem, errors.New("maximum size of uploaded file is 1G"))
			return
		}

		fo, err := zipFile.Open()
		if err != nil {
			// 处理出错删除minio已上传文件
			minioClient.DeleteObject(payFilePath)
			httpx.GinHTTPError(c, ActionImportCsvItem, err)
			return
		}
		filePath := path.Join(appRoot, "temp", "temp_"+timestamp+zipFile.Filename)
		file, err := minioClient.SavePublicObject(fo, filePath, "application/x-zip-compressed")
		if err != nil {
			// 处理出错删除minio已上传文件
			minioClient.DeleteObject(payFilePath)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据验证错误，停止上传
			jobx.ModifyTask(task.ModifyRequest{
				JobId: jobID,
				// Message:     i18n.Tr(lang, "job.J_053"),
				Message:     "An error occurred while uploading the file",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionImportCsvItem, err)
			return
		}

		zipFilePath = file.Name
	}

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		// 处理出错删除minio已上传文件
		minioClient.DeleteObject(payFilePath)
		minioClient.DeleteObject(zipFilePath)
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionImportCsvItem, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
		// 处理出错删除minio已上传文件
		minioClient.DeleteObject(payFilePath)
		minioClient.DeleteObject(zipFilePath)
		path := filex.WriteAndSaveFile(domain, appID, []string{fmt.Sprintf("the csv file type [%v] is not supported", files.Header.Get("content-type"))})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionImportCsvItem, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "csv", files.Size) {
		// 处理出错删除minio已上传文件
		minioClient.DeleteObject(payFilePath)
		minioClient.DeleteObject(zipFilePath)
		path := filex.WriteAndSaveFile(domain, appID, []string{"the csv file ファイルサイズが設定サイズを超えています"})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionImportCsvItem, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	fo, err := files.Open()
	if err != nil {
		// 处理出错删除minio已上传文件
		minioClient.DeleteObject(payFilePath)
		minioClient.DeleteObject(zipFilePath)
		httpx.GinHTTPError(c, ActionImportCsvItem, err)
		return
	}

	filePath := path.Join(appRoot, "temp", "temp_"+timestamp+files.Filename)
	file, err := minioClient.SavePublicObject(fo, filePath, files.Header.Get("content-type"))
	if err != nil {
		// 处理出错删除minio已上传文件
		minioClient.DeleteObject(payFilePath)
		minioClient.DeleteObject(zipFilePath)
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionImportCsvItem, err)
		return
	}
	// 文件开始上传
	err = csvUpload(c, file.Name, zipFilePath, payFilePath)
	if err != nil {
		// 处理出错删除minio已上传文件
		minioClient.DeleteObject(payFilePath)
		minioClient.DeleteObject(zipFilePath)
		minioClient.DeleteObject(file.Name)
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionImportCsvItem, err)
		return
	}

	loggerx.InfoLog(c, ActionImportCsvItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ItemProcessName, ActionImportCsvItem)),
		Data:    nil,
	})
}
