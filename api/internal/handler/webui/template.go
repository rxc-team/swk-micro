package webui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/template"
)

// Template Template
type Template struct{}

// log出力
const (
	TemplateProcessName       = "Template"
	ActionFindTemplateItems   = "ActionFindTemplateItems"
	ActionDeleteTemplateItems = "DeleteTemplateItems"
)

// FindTemplateItems 查询临时台账数据
// @Router /templates [get]
func (i *Template) FindTemplateItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindTemplateItems, loggerx.MsgProcessStarted)

	tplService := template.NewTemplateService("database", client.DefaultClient)

	var req template.ItemsRequest
	// 从query获取
	req.TemplateId = c.Query("template_id")
	req.PageIndex, _ = strconv.ParseInt(c.Query("page_index"), 0, 64)
	req.PageSize, _ = strconv.ParseInt(c.Query("page_size"), 0, 64)
	req.DatastoreKey = c.Query("datastore_key")
	req.Database = sessionx.GetUserCustomer(c)
	req.Collection = sessionx.GetAuthUserID(c)

	response, err := tplService.FindTemplateItems(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindTemplateItems, err)
		return
	}
	// 试算
	if req.GetDatastoreKey() == "paymentInterest" {
		var res []typesx.Lease
		for _, item := range response.GetItems() {
			res = append(res, toLease(item.GetItems()))
		}
		loggerx.InfoLog(c, ActionFindTemplateItems, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, TemplateProcessName, ActionFindTemplateItems)),
			Data: gin.H{
				"total": response.GetTotal(),
				"data":  res,
			},
		})
		return
	}
	// RePayment 偿还数据
	if req.GetDatastoreKey() == "repayment" {
		var res []typesx.RePayment
		for _, item := range response.GetItems() {
			res = append(res, toRePayment(item.GetItems()))
		}
		loggerx.InfoLog(c, ActionFindTemplateItems, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, TemplateProcessName, ActionFindTemplateItems)),
			Data: gin.H{
				"total": response.GetTotal(),
				"data":  res,
			},
		})
		return
	}

	var res []map[string]interface{}

	for _, item := range response.GetItems() {
		it := make(map[string]interface{})
		it["item_id"] = item.ItemId
		it["app_id"] = item.AppId
		it["datastore_id"] = item.DatastoreId
		it["created_at"] = item.CreatedAt
		it["created_by"] = item.CreatedBy
		it["template_id"] = item.TemplateId
		it["datastore_key"] = item.DatastoreKey

		itemMap := make(map[string]interface{})
		for key, value := range item.GetItems() {
			itemMap[key] = value
		}
		it["items"] = itemMap

		res = append(res, it)
	}

	loggerx.InfoLog(c, ActionFindTemplateItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, TemplateProcessName, ActionFindTemplateItems)),
		Data: gin.H{
			"total": response.GetTotal(),
			"data":  res,
		},
	})
}

// map转Lease对象
func toLease(itemMap map[string]*template.Value) (lease typesx.Lease) {
	interest := 0.0
	repayment := 0.0
	balance := 0.0
	present := 0.0
	paymentymd := ""
	if interestvalue, ok := itemMap["interest"]; ok {
		interest, _ = strconv.ParseFloat(interestvalue.Value, 64)
	}
	if repaymentvalue, ok := itemMap["repayment"]; ok {
		repayment, _ = strconv.ParseFloat(repaymentvalue.Value, 64)
	}
	if balancevalue, ok := itemMap["balance"]; ok {
		balance, _ = strconv.ParseFloat(balancevalue.Value, 64)
	}
	if presentvalue, ok := itemMap["present"]; ok {
		present, _ = strconv.ParseFloat(presentvalue.Value, 64)
	}
	if paymentymdvalue, ok := itemMap["paymentymd"]; ok {
		paymentymd = paymentymdvalue.Value
	}

	lease = typesx.Lease{
		Interest:   interest,
		Repayment:  repayment,
		Balance:    balance,
		Present:    present,
		Paymentymd: paymentymd,
	}
	return lease
}

// map转RePayment对象
func toRePayment(itemMap map[string]*template.Value) (rePayment typesx.RePayment) {
	endboka := 0.0
	boka := 0.0
	syokyaku := 0.0
	syokyakuymd := ""
	syokyakukbn := ""
	if endbokavalue, ok := itemMap["endboka"]; ok {
		endboka, _ = strconv.ParseFloat(endbokavalue.Value, 64)
	}
	if bokavalue, ok := itemMap["boka"]; ok {
		boka, _ = strconv.ParseFloat(bokavalue.Value, 64)
	}
	if syokyakuvalue, ok := itemMap["syokyaku"]; ok {
		syokyaku, _ = strconv.ParseFloat(syokyakuvalue.Value, 64)
	}
	if syokyakuymdvalue, ok := itemMap["syokyakuymd"]; ok {
		syokyakuymd = syokyakuymdvalue.Value
	}
	if syokyakukbnvalue, ok := itemMap["syokyakukbn"]; ok {
		syokyakukbn = syokyakukbnvalue.Value
	}
	rePayment = typesx.RePayment{
		Endboka:     endboka,
		Boka:        boka,
		Syokyaku:    syokyaku,
		Syokyakuymd: syokyakuymd,
		Syokyakukbn: syokyakukbn,
	}

	return rePayment
}

// DeleteTemplateItems 删除临时数据
// @Router /templates/{template_id} [delete]
func (i *Template) DeleteTemplateItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteTemplateItems, loggerx.MsgProcessStarted)

	tplService := template.NewTemplateService("database", client.DefaultClient)

	var req template.DeleteRequest
	// 从path中获取参数
	req.TemplateId = c.Param("template_id")
	req.Collection = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := tplService.DeleteTemplateItems(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteTemplateItems, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteTemplateItems, fmt.Sprintf("DeleteTemplate[%s] Success", req.GetTemplateId()))

	loggerx.InfoLog(c, ActionDeleteTemplateItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, TemplateProcessName, ActionDeleteTemplateItems)),
		Data:    response,
	})
}
