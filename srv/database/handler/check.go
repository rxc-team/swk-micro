package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// log出力使用
const (
	ActionInventoryItem       = "InventoryItem"
	ActionMutilInventoryItem  = "MutilInventoryItem"
	ActionChangeLabelTime     = "ChangeLabelTime"
	ActionResetInventoryItems = "ResetInventoryItems"
)

// InventoryItem 盘点台账一条数据
func (i *Item) InventoryItem(ctx context.Context, req *item.InventoryItemRequest, rsp *item.InventoryItemResponse) error {
	utils.InfoLog(ActionInventoryItem, utils.MsgProcessStarted)

	err := model.InventoryItem(req.GetDatabase(), req.GetWriter(), req.GetItemId(), req.GetDatastoreId(), req.GetAppId(), req.GetCheckType(), req.GetImage(), req.GetCheckField())
	if err != nil {
		utils.ErrorLog(ActionInventoryItem, err.Error())
		return err
	}

	utils.InfoLog(ActionInventoryItem, utils.MsgProcessEnded)
	return nil
}

// MutilInventoryItem 一次盘点多条台账数据
func (i *Item) MutilInventoryItem(ctx context.Context, req *item.MutilInventoryItemRequest, rsp *item.MutilInventoryItemResponse) error {
	utils.InfoLog(ActionMutilInventoryItem, utils.MsgProcessStarted)

	err := model.MutilInventoryItem(req.GetDatabase(), req.GetWriter(), req.GetDatastoreId(), req.GetAppId(), req.GetCheckType(), req.GetItemIdList())
	if err != nil {
		utils.ErrorLog(ActionMutilInventoryItem, err.Error())
		return err
	}

	utils.InfoLog(ActionMutilInventoryItem, utils.MsgProcessEnded)
	return nil
}

// ChangeLabelTime 修改标签出力时间
func (i *Item) ChangeLabelTime(ctx context.Context, req *item.LabelTimeRequest, rsp *item.LabelTimeResponse) error {
	utils.InfoLog(ActionChangeLabelTime, utils.MsgProcessStarted)

	err := model.ChangeLabelTime(req.GetDatabase(), req.GetItemIdList(), req.GetDatastoreId())
	if err != nil {
		utils.ErrorLog(ActionChangeLabelTime, err.Error())
		return err
	}

	utils.InfoLog(ActionChangeLabelTime, utils.MsgProcessEnded)
	return nil
}

// ResetInventoryItems 盘点台账盘点数据盘点状态重置
func (i *Item) ResetInventoryItems(ctx context.Context, req *item.ResetInventoryItemsRequest, rsp *item.ResetInventoryItemsResponse) error {
	utils.InfoLog(ActionResetInventoryItems, utils.MsgProcessStarted)

	err := model.ResetInventoryItems(req.GetDatabase(), req.GetWriter(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionResetInventoryItems, err.Error())
		return err
	}

	utils.InfoLog(ActionResetInventoryItems, utils.MsgProcessEnded)
	return nil
}
