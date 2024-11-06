package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/proto/template"
	"rxcsoft.cn/pit3/srv/database/utils"

	"rxcsoft.cn/pit3/srv/database/model"
)

// Template 临时台账数据
type Template struct{}

// log出力使用
const (
	TemplateProcessName        = "Template"
	ActionFindTemplateItems    = "FindTemplateItems"
	ActionMutilAddTemplateItem = "MutilAddTemplateItem"
	ActionDeleteTemplateItems  = "DeleteTemplateItems"
)

// FindTemplateItems 查找记录
func (i *Template) FindTemplateItems(ctx context.Context, req *template.ItemsRequest, rsp *template.ItemsResponse) error {
	utils.InfoLog(ActionFindTemplateItems, utils.MsgProcessStarted)

	items, total, err := model.FindTemplateItems(req.GetDatabase(), req.GetTemplateId(), req.GetDatastoreKey(), req.GetCollection(), req.GetPageSize(), req.GetPageIndex())
	if err != nil {
		utils.ErrorLog(ActionFindTemplateItems, err.Error())
		return err
	}

	res := &template.ItemsResponse{}
	for _, it := range items {
		res.Items = append(res.Items, it.ToProto(true))
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindTemplateItems, utils.MsgProcessEnded)
	return nil
}

// MutilAddTemplateItem 批量添加数据(契约台账数据新规审批时，把支付，试算，偿却信息添加进来)
func (i *Template) MutilAddTemplateItem(ctx context.Context, req *template.MutilAddRequest, rsp *template.MutilAddResponse) error {
	utils.InfoLog(ActionMutilAddTemplateItem, utils.MsgProcessStarted)

	itemList := make([]*model.TemplateItem, 0)
	for _, listItems := range req.GetData() {
		items := make(map[string]*model.Value, len(listItems.Items))
		for key, item := range listItems.Items {
			items[key] = &model.Value{
				DataType: item.DataType,
				Value:    model.GetTemplateValueFromProto(item),
			}
		}

		params := model.TemplateItem{
			AppID:        req.GetAppId(),
			DatastoreID:  listItems.GetDatastoreId(),
			ItemMap:      items,
			TemplateID:   listItems.GetTemplateId(),
			DatastoreKey: listItems.GetDatastoreKey(),
			CreatedAt:    time.Now(),
			CreatedBy:    req.GetWriter(),
		}
		itemList = append(itemList, &params)
	}

	err := model.MutilAddTemplateItem(req.GetDatabase(), req.GetCollection(), itemList)
	if err != nil {
		utils.ErrorLog(ActionMutilAddTemplateItem, err.Error())
		return err
	}

	utils.InfoLog(ActionMutilAddTemplateItem, utils.MsgProcessEnded)

	return nil
}

// DeleteTemplateItems 删除与契约相关的支付、试算等记录
func (i *Template) DeleteTemplateItems(ctx context.Context, req *template.DeleteRequest, rsp *template.DeleteResponse) error {
	utils.InfoLog(ActionDeleteTemplateItems, utils.MsgProcessStarted)

	err := model.DeleteTemplateItems(req.GetDatabase(), req.GetCollection(), req.GetTemplateId())
	if err != nil {
		utils.ErrorLog(ActionDeleteTemplateItems, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteTemplateItems, utils.MsgProcessEnded)
	return nil
}
