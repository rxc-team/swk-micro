package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/leasex"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wfx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
)

// log出力
const (
	LeaseProcessName        = "Lease"
	ActionGeneratePay       = "GeneratePay"
	ActionComputeLeaserepay = "ComputeLeaserepay"
)

// ModifyContract 契约情报变更
// @Router /datastores/{d_id}/items/{i_id}/contract [put]
func (i *Item) ModifyContract(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyContract, loggerx.MsgProcessStarted)

	datastoreID := c.Param("d_id")
	itemID := c.Param("i_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	groupID := sessionx.GetUserGroup(c)

	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var dsreq datastore.DatastoresRequest
	// 从共通获取
	dsreq.Database = db
	dsreq.AppId = appID

	dresp, err := datastoreService.FindDatastores(context.TODO(), &dsreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}

	dsMap := make(map[string]string)

	for _, ds := range dresp.GetDatastores() {
		dsMap[ds.ApiKey] = ds.GetDatastoreId()
	}

	var ireq item.ItemRequest
	ireq.DatastoreId = datastoreID
	ireq.ItemId = itemID
	ireq.Database = db
	ireq.IsOrigin = true
	ireq.Owners = sessionx.GetUserAccessKeys(c, datastoreID, "R")

	resp, err := itemService.FindItem(context.TODO(), &ireq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}
	keiyakuno := resp.Item.Items["keiyakuno"].Value

	// 利息台账债务变更前数据和偿还台账债务变更前数据取得
	var payData []typesx.Payment
	var leaseData []typesx.Lease
	var repayData []typesx.RePayment
	// 检索条件
	var conditions []*item.Condition
	conditions = append(conditions, &item.Condition{
		FieldId:       "keiyakuno",
		FieldType:     "lookup",
		SearchValue:   keiyakuno,
		Operator:      "=",
		IsDynamic:     true,
		ConditionType: "",
	})
	// 利息表排序条件
	var sorts []*item.SortItem
	sorts = append(sorts, &item.SortItem{
		SortKey:   "paymentymd",
		SortValue: "ascend",
	})
	// 偿还表排序条件
	var ssorts []*item.SortItem
	ssorts = append(ssorts, &item.SortItem{
		SortKey:   "syokyakuymd",
		SortValue: "ascend",
	})

	payAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["paymentStatus"], "R")

	// 债务变更前的支付台账数据取得
	var preq item.ItemsRequest
	preq.ConditionList = conditions
	preq.ConditionType = "and"
	preq.Sorts = sorts
	preq.DatastoreId = dsMap["paymentStatus"]
	// 共通参数
	preq.AppId = appID
	preq.Owners = payAccessKeys
	preq.Database = db
	preq.IsOrigin = true

	// 数据取得
	pResp, err := itemService.FindItems(context.TODO(), &preq, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}
	// 数据编辑到payData
	for _, it := range pResp.GetItems() {
		paymentcount, err := strconv.Atoi(it.Items["paymentcount"].GetValue())
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		paymentleasefee, err := strconv.ParseFloat(it.Items["paymentleasefee"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		paymentleasefeehendo, err := strconv.ParseFloat(it.Items["paymentleasefeehendo"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		incentives, err := strconv.ParseFloat(it.Items["incentives"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		sonotafee, err := strconv.ParseFloat(it.Items["sonotafee"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		kaiyakuson, err := strconv.ParseFloat(it.Items["kaiyakuson"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		fixed := false
		if it.Items["fixed"].GetValue() != "" {
			fixed, err = strconv.ParseBool(it.Items["fixed"].GetValue())
			if err != nil {
				httpx.GinHTTPError(c, ActionModifyContract, err)
				return
			}
		}
		paymentType := it.Items["paymentType"].GetValue()
		paymentymd := it.Items["paymentymd"].GetValue()
		pay := typesx.Payment{
			Paymentcount:         paymentcount,
			PaymentType:          paymentType,
			Paymentymd:           paymentymd,
			Paymentleasefee:      paymentleasefee,
			Paymentleasefeehendo: paymentleasefeehendo,
			Incentives:           incentives,
			Sonotafee:            sonotafee,
			Kaiyakuson:           kaiyakuson,
			Fixed:                fixed,
		}
		payData = append(payData, pay)
	}

	leaseAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["paymentInterest"], "R")

	// 利息台账债务变更前数据取得
	var lreq item.ItemsRequest
	lreq.ConditionList = conditions
	lreq.ConditionType = "and"
	lreq.Sorts = sorts
	lreq.DatastoreId = dsMap["paymentInterest"]
	// 共通参数
	lreq.AppId = appID
	lreq.Owners = leaseAccessKeys
	lreq.Database = db
	lreq.IsOrigin = true
	// 数据取得处理
	lResp, err := itemService.FindItems(context.TODO(), &lreq, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}
	// 数据编辑到leaseData
	for _, it := range lResp.GetItems() {
		interest, err := strconv.ParseFloat(it.Items["interest"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		repayment, err := strconv.ParseFloat(it.Items["repayment"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		balance, err := strconv.ParseFloat(it.Items["balance"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		present, err := strconv.ParseFloat(it.Items["present"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		paymentymd := it.Items["paymentymd"].GetValue()
		lease := typesx.Lease{
			Interest:   interest,
			Repayment:  repayment,
			Balance:    balance,
			Present:    present,
			Paymentymd: paymentymd,
		}
		leaseData = append(leaseData, lease)
	}

	repayAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["repayment"], "R")

	// 偿还台账债务变更前数据取得
	var rreq item.ItemsRequest
	rreq.ConditionList = conditions
	rreq.ConditionType = "and"
	rreq.Sorts = ssorts
	// path参数
	rreq.DatastoreId = dsMap["repayment"]
	// 共通参数
	rreq.AppId = appID
	rreq.Owners = repayAccessKeys
	rreq.Database = db
	rreq.IsOrigin = true
	// 数据取得处理
	rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}
	// 数据编辑到repayData
	for _, it := range rResp.GetItems() {
		endboka, err := strconv.ParseFloat(it.Items["endboka"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		boka, err := strconv.ParseFloat(it.Items["boka"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		syokyaku, err := strconv.ParseFloat(it.Items["syokyaku"].GetValue(), 64)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}
		syokyakuymd := it.Items["syokyakuymd"].GetValue()
		syokyakukbn := it.Items["syokyakukbn"].GetValue()
		rePayment := typesx.RePayment{
			Endboka:     endboka,
			Boka:        boka,
			Syokyaku:    syokyaku,
			Syokyakuymd: syokyakuymd,
			Syokyakukbn: syokyakukbn,
		}
		repayData = append(repayData, rePayment)
	}

	wks := wfx.GetUserWorkflow(db, groupID, appID, datastoreID, "info-change")
	if len(wks) > 0 {

		var iReq item.ItemRequest
		iReq.DatastoreId = datastoreID
		iReq.ItemId = itemID
		iReq.Database = db
		iReq.IsOrigin = true
		iReq.Owners = sessionx.GetUserAccessKeys(c, datastoreID, "W")

		iResp, err := itemService.FindItem(context.TODO(), &iReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}

		var req approve.AddRequest
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}

		// 插入履历数据
		henkouymd := req.Items["henkouymd"].GetValue()

		result, err := leasex.ChangeCompute(db, appID, henkouymd, userID, payData, leaseData, repayData, dsMap, true)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
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

		itemMap["template_id"] = &approve.Value{
			DataType: "text",
			Value:    result.TemplateID,
		}

		approveService := approve.NewApproveService("database", client.DefaultClient)

		req.ItemId = itemID
		req.History = itemMap
		req.Current = req.Items
		req.DatastoreId = datastoreID
		req.AppId = appID
		req.Writer = userID
		req.Database = db
		req.Domain = domain
		req.LangCd = sessionx.GetCurrentLanguage(c)
		// 开启流程
		approve := new(wfx.Approve)
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

		var statusReq item.StatusRequest
		statusReq.AppId = appID
		statusReq.DatastoreId = datastoreID
		statusReq.ItemId = itemID
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "2"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyContract, err)
			return
		}

		loggerx.InfoLog(c, ActionAddItem, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionAddItem)),
			Data:    response,
		})
		c.Abort()
		return
	}

	var req item.ModifyContractRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}

	// 插入履历数据
	henkouymd := req.Items["henkouymd"].GetValue()

	result, err := leasex.ChangeCompute(db, appID, henkouymd, userID, payData, leaseData, repayData, dsMap, true)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}

	req.Items["template_id"] = &item.Value{
		DataType: "text",
		Value:    result.TemplateID,
	}

	// 从path中获取参数
	req.DatastoreId = datastoreID
	req.ItemId = itemID
	// 从共通中获取参数
	req.AppId = appID
	req.Writer = userID
	req.LangCd = sessionx.GetCurrentLanguage(c)
	req.Domain = domain
	req.Database = db
	req.Owners = sessionx.GetUserAccessKeys(c, datastoreID, "W")

	response, err := itemService.ModifyContract(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyContract, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyContract, fmt.Sprintf("item[%s] update success", req.GetItemId()))

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

	loggerx.InfoLog(c, ActionModifyContract, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionModifyContract)),
		Data:    response,
	})
}

// TerminateContract 中途解约
// @Router /datastores/{d_id}/items/{i_id}/terminate [put]
func (i *Item) TerminateContract(c *gin.Context) {
	loggerx.InfoLog(c, ActionTerminateContract, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	itemID := c.Param("i_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	owners := sessionx.GetUserAccessKeys(c, datastore, "W")
	groupID := sessionx.GetUserGroup(c)

	wks := wfx.GetUserWorkflow(db, groupID, appID, datastore, "midway-cancel")
	if len(wks) > 0 {
		itemService := item.NewItemService("database", client.DefaultClient)

		var iReq item.ItemRequest
		iReq.DatastoreId = datastore
		iReq.ItemId = itemID
		iReq.Database = db
		iReq.IsOrigin = true
		iReq.Owners = sessionx.GetUserAccessKeys(c, datastore, "W")

		iResp, err := itemService.FindItem(context.TODO(), &iReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
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
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}
		req.ItemId = itemID
		req.Current = req.Items
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
		exID, err := approve.AddExample(db, wks[0].GetWfId(), userID)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}
		req.ExampleId = exID
		response, err := approveService.AddItem(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}
		// 流程开始启动
		err = approve.StartExampleInstance(db, wks[0].GetWfId(), userID, exID, domain)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}

		loggerx.SuccessLog(c, ActionTerminateContract, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

		var statusReq item.StatusRequest
		statusReq.AppId = appID
		statusReq.DatastoreId = datastore
		statusReq.ItemId = itemID
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "2"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}

		loggerx.InfoLog(c, ActionTerminateContract, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionChangeDebt)),
			Data:    response,
		})
		c.Abort()
		return
	}
	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.TerminateContractRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionTerminateContract, err)
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

	response, err := itemService.TerminateContract(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionTerminateContract, err)
		return
	}
	loggerx.SuccessLog(c, ActionTerminateContract, fmt.Sprintf("item[%s] update success", req.GetItemId()))

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

	loggerx.InfoLog(c, ActionTerminateContract, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionTerminateContract)),
		Data:    response,
	})
}

// ChangeDebt 债务变更
// @Router /datastores/{d_id}/items/{i_id}/debt [put]
func (i *Item) ChangeDebt(c *gin.Context) {
	loggerx.InfoLog(c, ActionChangeDebt, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	itemID := c.Param("i_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	owners := sessionx.GetUserOwner(c)

	groupID := sessionx.GetUserGroup(c)

	wks := wfx.GetUserWorkflow(db, groupID, appID, datastore, "debt-change")
	if len(wks) > 0 {
		itemService := item.NewItemService("database", client.DefaultClient)

		var iReq item.ItemRequest
		iReq.DatastoreId = datastore
		iReq.ItemId = itemID
		iReq.Database = db
		iReq.IsOrigin = true
		iReq.Owners = sessionx.GetUserAccessKeys(c, datastore, "W")

		iResp, err := itemService.FindItem(context.TODO(), &iReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionChangeDebt, err)
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
			httpx.GinHTTPError(c, ActionChangeDebt, err)
			return
		}
		req.ItemId = itemID
		req.Current = req.Items
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
		exID, err := approve.AddExample(db, wks[0].GetWfId(), userID)
		if err != nil {
			httpx.GinHTTPError(c, ActionChangeDebt, err)
			return
		}
		req.ExampleId = exID
		response, err := approveService.AddItem(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionChangeDebt, err)
			return
		}
		// 流程开始启动
		err = approve.StartExampleInstance(db, wks[0].GetWfId(), userID, exID, domain)
		if err != nil {
			httpx.GinHTTPError(c, ActionChangeDebt, err)
			return
		}

		loggerx.SuccessLog(c, ActionChangeDebt, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

		var statusReq item.StatusRequest
		statusReq.AppId = appID
		statusReq.DatastoreId = datastore
		statusReq.ItemId = itemID
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "2"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionChangeDebt, err)
			return
		}

		loggerx.InfoLog(c, ActionChangeDebt, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionChangeDebt)),
			Data:    response,
		})
		c.Abort()
		return
	}
	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.ChangeDebtRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionChangeDebt, err)
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

	response, err := itemService.ChangeDebt(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionChangeDebt, err)
		return
	}
	loggerx.SuccessLog(c, ActionChangeDebt, fmt.Sprintf("item[%s] update success", req.GetItemId()))

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

	loggerx.InfoLog(c, ActionChangeDebt, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionChangeDebt)),
		Data:    response,
	})
}

// ContractExpire 契约满了
// @Router /datastores/{d_id}/items/{i_id}/contractExpire [put]
func (i *Item) ContractExpire(c *gin.Context) {
	loggerx.InfoLog(c, ActionContractExpire, loggerx.MsgProcessStarted)

	datastore := c.Param("d_id")
	itemID := c.Param("i_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	groupID := sessionx.GetUserGroup(c)
	owners := sessionx.GetUserOwner(c)

	wks := wfx.GetUserWorkflow(db, groupID, appID, datastore, "contract-expire")
	if len(wks) > 0 {
		itemService := item.NewItemService("database", client.DefaultClient)

		var iReq item.ItemRequest
		iReq.DatastoreId = datastore
		iReq.ItemId = itemID
		iReq.Database = db
		iReq.IsOrigin = true
		iReq.Owners = sessionx.GetUserAccessKeys(c, datastore, "W")

		iResp, err := itemService.FindItem(context.TODO(), &iReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
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

		wfID := wks[0].GetWfId()

		approveService := approve.NewApproveService("database", client.DefaultClient)

		var req approve.AddRequest
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}
		req.ItemId = itemID
		req.Current = req.Items
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
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}
		req.ExampleId = exID
		response, err := approveService.AddItem(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}
		// 流程开始启动
		err = approve.StartExampleInstance(db, wfID, userID, exID, domain)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}

		loggerx.SuccessLog(c, ActionTerminateContract, fmt.Sprintf("Item[%s] Add Success", response.GetItemId()))

		var statusReq item.StatusRequest
		statusReq.AppId = appID
		statusReq.DatastoreId = datastore
		statusReq.ItemId = itemID
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "2"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionTerminateContract, err)
			return
		}

		loggerx.InfoLog(c, ActionTerminateContract, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionChangeDebt)),
			Data:    response,
		})
		c.Abort()
		return
	}

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.ContractExpireRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionTerminateContract, err)
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

	response, err := itemService.ContractExpire(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionContractExpire, err)
		return
	}
	loggerx.SuccessLog(c, ActionContractExpire, fmt.Sprintf("item[%s] update success", req.GetItemId()))

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

	loggerx.InfoLog(c, ActionContractExpire, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionContractExpire)),
		Data:    response,
	})
}

// GeneratePay 生成支付数据(租赁系统用)
// @Router /generate/pay [post]
func (i *Item) GeneratePay(c *gin.Context) {
	loggerx.InfoLog(c, ActionGeneratePay, loggerx.MsgProcessStarted)

	// 从body中获取契约情报
	var req typesx.PayParam
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionGeneratePay, err)
		return
	}

	// 生成支付数据(租赁系统用)
	payData, err := leasex.GeneratePay(req)
	if err != nil {
		httpx.GinHTTPError(c, ActionGeneratePay, err)
		return
	}

	loggerx.SuccessLog(c, ActionGeneratePay, "generatePay success")

	loggerx.InfoLog(c, ActionGeneratePay, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionGeneratePay)),
		Data:    payData,
	})
}

// ComputeLeaserepay 计算利息和偿还数据(租赁系统用)
// @Router /compute/leaserepay [post]
func (i *Item) ComputeLeaserepay(c *gin.Context) {
	loggerx.InfoLog(c, ActionComputeLeaserepay, loggerx.MsgProcessStarted)

	// 共通参数取得
	db := sessionx.GetUserCustomer(c)
	userID := sessionx.GetAuthUserID(c)
	appID := sessionx.GetCurrentApp(c)
	// accessKey := sessionx.GetUserOwner(c)
	// 契约处理类型区分
	section := c.Query("section")

	// 契约新规的情形
	if len(section) == 0 {
		// 从body中获取契约情报
		var req typesx.LRParam
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

		var dsreq datastore.DatastoresRequest
		// 从共通获取
		dsreq.Database = db
		dsreq.AppId = appID

		response, err := datastoreService.FindDatastores(context.TODO(), &dsreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		dsMap := make(map[string]string)

		for _, ds := range response.GetDatastores() {
			dsMap[ds.ApiKey] = ds.GetDatastoreId()
		}

		req.DsMap = dsMap

		var tID string
		var kisyuBoka float64
		var hkkjitenzan float64
		var sonnekigaku float64
		leaseType := leasex.ShortOrMinorJudge(db, appID, req.Leasekikan, req.ExtentionOption, req.Payments)
		if leaseType != "normal_lease" {
			result, err := leasex.InsertPay(db, appID, userID, dsMap, req.Payments, true)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			tID = result.TemplateID
		} else {
			// 计算利息和偿还数据后返回临时数据ID(租赁系统用)
			result, err := leasex.Compute(db, appID, userID, req, true)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			tID = result.TemplateID
			kisyuBoka = result.KiSyuBoka
			hkkjitenzan = result.Hkkjitenzan
			sonnekigaku = result.Sonnekigaku
		}

		loggerx.InfoLog(c, ActionComputeLeaserepay, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionComputeLeaserepay)),
			Data: gin.H{
				"template_id": tID,
				"lease_type":  leaseType,
				"kisyuboka":   kisyuBoka,
				"hkkjitenzan": hkkjitenzan,
				"sonnekigaku": sonnekigaku,
			},
		})
		return
	}

	// 债务变更的情形
	if section == "debt" {
		// 从body中获取契约情报
		var req typesx.DebtParam
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

		var dsreq datastore.DatastoresRequest
		// 从共通获取
		dsreq.Database = db
		dsreq.AppId = appID

		response, err := datastoreService.FindDatastores(context.TODO(), &dsreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		dsMap := make(map[string]string)

		for _, ds := range response.GetDatastores() {
			dsMap[ds.ApiKey] = ds.GetDatastoreId()
		}

		req.DsMap = dsMap

		// 利息台账债务变更前数据和偿还台账债务变更前数据取得
		var kisyuBoka float64
		var payData []typesx.Payment
		var leaseData []typesx.Lease
		var repayData []typesx.RePayment
		// 检索条件
		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "lookup",
			SearchValue:   req.Keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		// 利息表排序条件
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "paymentymd",
			SortValue: "ascend",
		})
		// 偿还表排序条件
		var ssorts []*item.SortItem
		ssorts = append(ssorts, &item.SortItem{
			SortKey:   "syokyakuymd",
			SortValue: "ascend",
		})

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		// 检索条件
		var conditionList []*item.Condition
		conditionList = append(conditionList, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "text",
			SearchValue:   req.Keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})

		var kreq item.ItemsRequest
		kreq.ConditionList = conditionList
		kreq.ConditionType = "and"
		kreq.AppId = appID
		kreq.DatastoreId = dsMap["keiyakudaicho"]
		kreq.IsOrigin = true
		kreq.PageIndex = 1
		kreq.PageSize = 1
		kreq.Database = db
		kreq.Owners = sessionx.GetUserAccessKeys(c, dsMap["keiyakudaicho"], "R")

		// 债务变更前的契约台账数据取得
		kres, err := itemService.FindItems(context.TODO(), &kreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		keiyaItem := kres.Items[0]

		if value, exist := keiyaItem.GetItems()["kisyuboka"]; exist {
			val, _ := strconv.ParseFloat(value.GetValue(), 64)
			kisyuBoka = val
		}

		payAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["paymentStatus"], "R")

		// 债务变更前的支付台账数据取得
		var preq item.ItemsRequest
		preq.ConditionList = conditions
		preq.ConditionType = "and"
		preq.Sorts = sorts
		preq.DatastoreId = dsMap["paymentStatus"]
		// 共通参数
		preq.AppId = appID
		preq.Owners = payAccessKeys
		preq.Database = db
		preq.IsOrigin = true
		// 数据取得
		pResp, err := itemService.FindItems(context.TODO(), &preq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}
		// 数据编辑到payData
		for _, it := range pResp.GetItems() {
			paymentcount, err := strconv.Atoi(it.Items["paymentcount"].GetValue())
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			paymentleasefee, err := strconv.ParseFloat(it.Items["paymentleasefee"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			paymentleasefeehendo, err := strconv.ParseFloat(it.Items["paymentleasefeehendo"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			incentives, err := strconv.ParseFloat(it.Items["incentives"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			sonotafee, err := strconv.ParseFloat(it.Items["sonotafee"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			kaiyakuson, err := strconv.ParseFloat(it.Items["kaiyakuson"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			fixed := false
			if it.Items["fixed"].GetValue() != "" {
				fixed, err = strconv.ParseBool(it.Items["fixed"].GetValue())
				if err != nil {
					httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
					return
				}
			}
			paymentType := it.Items["paymentType"].GetValue()
			paymentymd := it.Items["paymentymd"].GetValue()
			pay := typesx.Payment{
				Paymentcount:         paymentcount,
				PaymentType:          paymentType,
				Paymentymd:           paymentymd,
				Paymentleasefee:      paymentleasefee,
				Paymentleasefeehendo: paymentleasefeehendo,
				Incentives:           incentives,
				Sonotafee:            sonotafee,
				Kaiyakuson:           kaiyakuson,
				Fixed:                fixed,
			}
			payData = append(payData, pay)
		}

		leaseAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["paymentInterest"], "R")

		// 利息台账债务变更前数据取得
		var lreq item.ItemsRequest
		lreq.ConditionList = conditions
		lreq.ConditionType = "and"
		lreq.Sorts = sorts
		lreq.DatastoreId = dsMap["paymentInterest"]
		// 共通参数
		lreq.AppId = appID
		lreq.Owners = leaseAccessKeys
		lreq.Database = db
		lreq.IsOrigin = true
		// 数据取得处理
		lResp, err := itemService.FindItems(context.TODO(), &lreq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}
		// 数据编辑到leaseData
		for _, it := range lResp.GetItems() {
			interest, err := strconv.ParseFloat(it.Items["interest"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			repayment, err := strconv.ParseFloat(it.Items["repayment"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			balance, err := strconv.ParseFloat(it.Items["balance"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			present, err := strconv.ParseFloat(it.Items["present"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			paymentymd := it.Items["paymentymd"].GetValue()
			lease := typesx.Lease{
				Interest:   interest,
				Repayment:  repayment,
				Balance:    balance,
				Present:    present,
				Paymentymd: paymentymd,
			}
			leaseData = append(leaseData, lease)
		}

		repayAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["repayment"], "R")

		// 偿还台账债务变更前数据取得
		var rreq item.ItemsRequest
		rreq.ConditionList = conditions
		rreq.ConditionType = "and"
		rreq.Sorts = ssorts
		// path参数
		rreq.DatastoreId = dsMap["repayment"]
		// 共通参数
		rreq.AppId = appID
		rreq.Owners = repayAccessKeys
		rreq.Database = db
		rreq.IsOrigin = true
		// 数据取得处理
		rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}
		// 数据编辑到repayData
		for _, it := range rResp.GetItems() {
			endboka, err := strconv.ParseFloat(it.Items["endboka"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			boka, err := strconv.ParseFloat(it.Items["boka"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			syokyaku, err := strconv.ParseFloat(it.Items["syokyaku"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			syokyakuymd := it.Items["syokyakuymd"].GetValue()
			syokyakukbn := it.Items["syokyakukbn"].GetValue()
			RePayment := typesx.RePayment{
				Endboka:     endboka,
				Boka:        boka,
				Syokyaku:    syokyaku,
				Syokyakuymd: syokyakuymd,
				Syokyakukbn: syokyakukbn,
			}
			repayData = append(repayData, RePayment)
		}

		// 债务变更计算后返回计算统计结果(租赁系统用)
		result, err := leasex.DebtCompute(db, appID, userID, kisyuBoka, payData, leaseData, repayData, req, true)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		loggerx.InfoLog(c, ActionComputeLeaserepay, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionComputeLeaserepay)),
			Data:    result,
		})
		return
	}

	// 中途解约的情形
	if section == "cancel" {
		// 从body中获取契约情报
		var req typesx.CancelParam
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

		var dsreq datastore.DatastoresRequest
		// 从共通获取
		dsreq.Database = db
		dsreq.AppId = appID

		response, err := datastoreService.FindDatastores(context.TODO(), &dsreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		dsMap := make(map[string]string)

		for _, ds := range response.GetDatastores() {
			dsMap[ds.ApiKey] = ds.GetDatastoreId()
		}

		req.DsMap = dsMap

		// 债务变更前的相关台账数据取得
		var payData []typesx.Payment
		var leaseData []typesx.Lease
		var repayData []typesx.RePayment
		// 检索条件
		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "lookup",
			SearchValue:   req.Keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		// 支付表和利息表排序
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "paymentymd",
			SortValue: "ascend",
		})
		// 偿还表排序
		var ssorts []*item.SortItem
		ssorts = append(ssorts, &item.SortItem{
			SortKey:   "syokyakuymd",
			SortValue: "ascend",
		})

		// 债务变更前的支付台账数据取得
		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		payAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["paymentStatus"], "R")

		var preq item.ItemsRequest
		preq.ConditionList = conditions
		preq.ConditionType = "and"
		preq.Sorts = sorts
		preq.DatastoreId = dsMap["paymentStatus"]
		// 共通参数
		preq.AppId = appID
		preq.Owners = payAccessKeys
		preq.Database = db
		preq.IsOrigin = true
		// 数据取得
		pResp, err := itemService.FindItems(context.TODO(), &preq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}
		// 数据编辑到payData
		for _, it := range pResp.GetItems() {
			paymentcount, err := strconv.Atoi(it.Items["paymentcount"].GetValue())
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			paymentleasefee, err := strconv.ParseFloat(it.Items["paymentleasefee"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			paymentleasefeehendo, err := strconv.ParseFloat(it.Items["paymentleasefeehendo"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			incentives, err := strconv.ParseFloat(it.Items["incentives"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			sonotafee, err := strconv.ParseFloat(it.Items["sonotafee"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			kaiyakuson, err := strconv.ParseFloat(it.Items["kaiyakuson"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			fixed := false
			if it.Items["fixed"].GetValue() != "" {
				fixed, err = strconv.ParseBool(it.Items["fixed"].GetValue())
				if err != nil {
					httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
					return
				}
			}
			paymentType := it.Items["paymentType"].GetValue()
			paymentymd := it.Items["paymentymd"].GetValue()
			pay := typesx.Payment{
				Paymentcount:         paymentcount,
				PaymentType:          paymentType,
				Paymentymd:           paymentymd,
				Paymentleasefee:      paymentleasefee,
				Paymentleasefeehendo: paymentleasefeehendo,
				Incentives:           incentives,
				Sonotafee:            sonotafee,
				Kaiyakuson:           kaiyakuson,
				Fixed:                fixed,
			}
			payData = append(payData, pay)
		}

		leaseAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["paymentInterest"], "R")

		// 债务变更前的利息台账数据取得
		var lreq item.ItemsRequest
		lreq.ConditionList = conditions
		lreq.ConditionType = "and"
		lreq.Sorts = sorts
		lreq.DatastoreId = dsMap["paymentInterest"]
		// 共通参数
		lreq.AppId = appID
		lreq.Owners = leaseAccessKeys
		lreq.Database = db
		lreq.IsOrigin = true
		// 数据取得
		lResp, err := itemService.FindItems(context.TODO(), &lreq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}
		// 数据编辑到leaseData
		for _, it := range lResp.GetItems() {
			interest, err := strconv.ParseFloat(it.Items["interest"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			repayment, err := strconv.ParseFloat(it.Items["repayment"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			balance, err := strconv.ParseFloat(it.Items["balance"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			present, err := strconv.ParseFloat(it.Items["present"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			paymentymd := it.Items["paymentymd"].GetValue()
			lease := typesx.Lease{
				Interest:   interest,
				Repayment:  repayment,
				Balance:    balance,
				Present:    present,
				Paymentymd: paymentymd,
			}
			leaseData = append(leaseData, lease)
		}

		repayAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["repayment"], "R")

		// 债务变更前的偿还台账数据取得
		var rreq item.ItemsRequest
		rreq.ConditionList = conditions
		rreq.ConditionType = "and"
		rreq.Sorts = ssorts
		// path参数
		rreq.DatastoreId = dsMap["repayment"]
		// 共通参数
		rreq.AppId = appID
		rreq.Owners = repayAccessKeys
		rreq.Database = db
		rreq.IsOrigin = true
		// 数据取得
		rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}
		// 数据编辑到repayData
		for _, it := range rResp.GetItems() {
			endboka, err := strconv.ParseFloat(it.Items["endboka"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			boka, err := strconv.ParseFloat(it.Items["boka"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			syokyaku, err := strconv.ParseFloat(it.Items["syokyaku"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			syokyakuymd := it.Items["syokyakuymd"].GetValue()
			syokyakukbn := it.Items["syokyakukbn"].GetValue()
			RePayment := typesx.RePayment{
				Endboka:     endboka,
				Boka:        boka,
				Syokyaku:    syokyaku,
				Syokyakuymd: syokyakuymd,
				Syokyakukbn: syokyakukbn,
			}
			repayData = append(repayData, RePayment)
		}

		// 中途解约计算处理后返回相关计算统计结果(租赁系统用)
		result, err := leasex.CancelCompute(db, appID, userID, payData, leaseData, repayData, req, true)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		loggerx.InfoLog(c, ActionComputeLeaserepay, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionComputeLeaserepay)),
			Data:    result,
		})
		return
	}

	// 契约满了的情形
	if section == "expire" {
		// 从body中获取契约情报
		var req typesx.ExpireParam
		if err := c.BindJSON(&req); err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

		var dsreq datastore.DatastoresRequest
		// 从共通获取
		dsreq.Database = db
		dsreq.AppId = appID

		response, err := datastoreService.FindDatastores(context.TODO(), &dsreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		dsMap := make(map[string]string)

		for _, ds := range response.GetDatastores() {
			dsMap[ds.ApiKey] = ds.GetDatastoreId()
		}

		req.DsMap = dsMap

		// 满了时偿却情报取得
		var repayData []typesx.RePayment
		// 检索条件
		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "lookup",
			SearchValue:   req.Keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		// 偿还表排序
		var ssorts []*item.SortItem
		ssorts = append(ssorts, &item.SortItem{
			SortKey:   "syokyakuymd",
			SortValue: "ascend",
		})
		// 满了时偿却情报取得
		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		repayAccessKeys := sessionx.GetAccessKeys(db, userID, dsMap["repayment"], "R")

		var rreq item.ItemsRequest
		rreq.ConditionList = conditions
		rreq.ConditionType = "and"
		rreq.Sorts = ssorts
		// path参数
		rreq.DatastoreId = dsMap["repayment"]
		// 共通参数
		rreq.AppId = appID
		rreq.Owners = repayAccessKeys
		rreq.Database = db
		rreq.IsOrigin = true
		// 数据取得
		rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}
		// 数据编辑到repayData
		for _, it := range rResp.GetItems() {
			endboka, err := strconv.ParseFloat(it.Items["endboka"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			boka, err := strconv.ParseFloat(it.Items["boka"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			syokyaku, err := strconv.ParseFloat(it.Items["syokyaku"].GetValue(), 64)
			if err != nil {
				httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
				return
			}
			syokyakuymd := it.Items["syokyakuymd"].GetValue()
			syokyakukbn := it.Items["syokyakukbn"].GetValue()
			RePayment := typesx.RePayment{
				Endboka:     endboka,
				Boka:        boka,
				Syokyaku:    syokyaku,
				Syokyakuymd: syokyakuymd,
				Syokyakukbn: syokyakukbn,
			}
			repayData = append(repayData, RePayment)
		}

		// 解约满了计算处理后返回相关计算统计结果(租赁系统用)
		result, err := leasex.ExpireCompute(db, appID, userID, repayData, req, true)
		if err != nil {
			httpx.GinHTTPError(c, ActionComputeLeaserepay, err)
			return
		}

		loggerx.InfoLog(c, ActionComputeLeaserepay, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LeaseProcessName, ActionComputeLeaserepay)),
			Data:    result,
		})
		return
	}
}
