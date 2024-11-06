package wfx

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/template"
	"rxcsoft.cn/pit3/srv/import/common/accessx"
	"rxcsoft.cn/pit3/srv/import/system/sessionx"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
)

// DsHandler 台账数据
type DsHandler struct {
}

// Admit 承认处理
func (b *DsHandler) Admit(w *Work) (result string, err error) {

	wfID := w.WorkflowID
	exID := w.ExampleID
	userID := w.UserID
	db := w.Database

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var req workflow.WorkflowRequest
	req.WfId = wfID
	req.Database = db

	fResp, err := workflowService.FindWorkflow(context.TODO(), &req)
	if err != nil {
		return "", err
	}
	params := fResp.GetWorkflow().GetParams()

	action := params["action"]

	// 开启一个流程实例
	tplService := approve.NewApproveService("database", client.DefaultClient)

	var tReq approve.ItemRequest
	tReq.ExampleId = exID
	tReq.DatastoreId = params["datastore"]
	tReq.Database = db

	tResp, err := tplService.FindItem(context.TODO(), &tReq)
	if err != nil {
		return "", err
	}

	if action == "insert" {

		itemService := item.NewItemService("database", client.DefaultClient)

		var aReq item.AddRequest
		// 从body中获取参数
		items := map[string]*item.Value{}
		for key, it := range tResp.GetItem().GetItems() {
			if it.GetDataType() == "user" {
				var uList []string
				err := json.Unmarshal([]byte(it.GetValue()), &uList)
				if err != nil {
					items[key] = &item.Value{
						DataType: it.GetDataType(),
						Value:    "",
					}
				} else {
					items[key] = &item.Value{
						DataType: it.GetDataType(),
						Value:    strings.Join(uList, ","),
					}
				}
			} else {
				items[key] = &item.Value{
					DataType: it.GetDataType(),
					Value:    it.GetValue(),
				}
			}
		}

		access := accessx.New(db, userID)

		owners := access.GetAccessKeys()

		// 从path中获取参数
		aReq.DatastoreId = tResp.GetItem().GetDatastoreId()
		aReq.Items = items
		// 从共通中获取参数
		aReq.AppId = tResp.GetItem().GetAppId()
		aReq.Owners = owners
		aReq.Writer = userID
		aReq.Database = db

		_, err := itemService.AddItem(context.TODO(), &aReq)
		if err != nil {
			templateID := items["template_id"].GetValue()
			e1 := deleteTemplateItems(db, userID, templateID)
			if e1 != nil {
				return "", e1
			}
			return "", err
		}

		return "ok", nil
	}
	if action == "update" {

		itemService := item.NewItemService("database", client.DefaultClient)

		owners := sessionx.GetAccessKeys(db, userID, params["datastore"], "W")

		// 从body中获取参数
		items := map[string]*item.Value{}
		// 项目变更后数据
		for key, it := range tResp.GetItem().GetItems() {
			if it.GetDataType() == "user" {
				var uList []string
				err := json.Unmarshal([]byte(it.GetValue()), &uList)
				if err != nil {
					items[key] = &item.Value{
						DataType: it.GetDataType(),
						Value:    "",
					}
				} else {
					items[key] = &item.Value{
						DataType: it.GetDataType(),
						Value:    strings.Join(uList, ","),
					}
				}
			} else if it.GetDataType() == "lookup" {
				if len(it.GetValue()) > 0 {
					result := strings.Split(it.GetValue(), " : ")
					items[key] = &item.Value{
						DataType: it.GetDataType(),
						Value:    result[0],
					}
				} else {
					items[key] = &item.Value{
						DataType: it.GetDataType(),
						Value:    "",
					}
				}
			} else {
				items[key] = &item.Value{
					DataType: it.GetDataType(),
					Value:    it.GetValue(),
				}
			}
		}

		var mReq item.ModifyRequest
		// 从path中获取参数
		mReq.DatastoreId = tResp.GetItem().GetDatastoreId()
		mReq.Items = items
		mReq.ItemId = tResp.GetItem().GetItemId()
		// 从共通中获取参数
		mReq.AppId = tResp.GetItem().GetAppId()
		mReq.Writer = userID
		mReq.Owners = owners
		mReq.Database = db

		_, err := itemService.ModifyItem(context.TODO(), &mReq)
		if err != nil {
			return "fail", err
		}

		var statusReq item.StatusRequest
		statusReq.AppId = tResp.GetItem().GetAppId()
		statusReq.DatastoreId = tResp.GetItem().GetDatastoreId()
		statusReq.ItemId = tResp.GetItem().GetItemId()
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "1"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			return "fail", err
		}

		return "ok", nil
	}

	return "ok", nil
}

// 契约台账登录时处理支付，试算相关数据完成后删除对应数据
func deleteTemplateItems(db, collection, templateID string) (err error) {
	tplService := template.NewTemplateService("database", client.DefaultClient)
	var req template.DeleteRequest
	req.TemplateId = templateID
	req.Collection = collection
	req.Database = db

	_, err = tplService.DeleteTemplateItems(context.TODO(), &req)
	if err != nil {
		return err
	}
	return nil
}

// Dismiss 却下处理
func (b *DsHandler) Dismiss(w *Work) (result string, err error) {

	wfID := w.WorkflowID
	exID := w.ExampleID
	userID := w.UserID
	db := w.Database

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var req workflow.WorkflowRequest
	req.WfId = wfID
	req.Database = db

	fResp, err := workflowService.FindWorkflow(context.TODO(), &req)
	if err != nil {
		return "", err
	}

	params := fResp.GetWorkflow().GetParams()
	action := params["action"]

	// 开启一个流程实例
	tplService := approve.NewApproveService("database", client.DefaultClient)

	var tReq approve.ItemRequest
	tReq.ExampleId = exID
	tReq.DatastoreId = params["datastore"]
	tReq.Database = db

	tResp, err := tplService.FindItem(context.TODO(), &tReq)
	if err != nil {
		return "", err
	}

	if action == "update" {
		itemService := item.NewItemService("database", client.DefaultClient)
		var statusReq item.StatusRequest
		statusReq.AppId = tResp.GetItem().GetAppId()
		statusReq.DatastoreId = tResp.GetItem().GetDatastoreId()
		statusReq.ItemId = tResp.GetItem().GetItemId()
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "1"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			return "fail", err
		}

		return "ok", nil
	}
	if action == "delete" {

		itemService := item.NewItemService("database", client.DefaultClient)

		var statusReq item.StatusRequest
		statusReq.AppId = tResp.GetItem().GetAppId()
		statusReq.DatastoreId = tResp.GetItem().GetDatastoreId()
		statusReq.ItemId = tResp.GetItem().GetItemId()
		statusReq.Database = db
		statusReq.Writer = userID
		statusReq.Status = "1"

		_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
		if err != nil {
			return "fail", err
		}

		return "ok", nil
	}

	return "ok", nil
}
