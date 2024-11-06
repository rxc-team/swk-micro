package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// log出力使用
const (
	ActionFindUnApproveItems = "FindUnApproveItems"
)

// FindUnApproveItems 查询台账未审批数据件数
func (i *Item) FindUnApproveItems(ctx context.Context, req *item.UnApproveItemsRequest, rsp *item.UnApproveItemsResponse) error {
	utils.InfoLog(ActionFindUnApproveItems, utils.MsgProcessStarted)

	res, err := model.FindUnApproveItems(req.GetDatabase(), req.GetStatus(), req.GetDatastoreId())
	if err != nil {
		utils.ErrorLog(ActionFindUnApproveItems, err.Error())
		return err
	}

	rsp.Total = res

	utils.InfoLog(ActionFindUnApproveItems, utils.MsgProcessEnded)
	return nil
}
