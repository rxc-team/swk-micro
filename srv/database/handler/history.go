/*
 * @Author: RXC 呉見華
 * @Date: 2019-12-16 10:44:16
 * @LastEditTime: 2020-05-15 14:55:03
 * @LastEditors: RXC 陈辉宇
 * @Description:
 * @FilePath: /database/handler/history.go
 * @
 */

package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/datahistory"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// History 台账数据
type History struct{}

// log出力使用
const (
	HistoryProcessName        = "History"
	ActionFindHistories       = "FindHistories"
	ActionFindLastHistories   = "FindLastHistories"
	ActionFindHistory         = "FindHistory"
	ActionCreateHistoryIndex  = "CreateHistoryIndex"
	ActionHardDeleteHistories = "HardDeleteHistories"
)

// FindHistories 获取多条履历数据
func (i *History) FindHistories(ctx context.Context, req *datahistory.HistoriesRequest, rsp *datahistory.HistoriesResponse) error {
	utils.InfoLog(ActionFindHistories, utils.MsgProcessStarted)

	params := model.FindParam{
		ItemID:        req.GetItemId(),
		HistoryType:   req.GetHistoryType(),
		DatastoreID:   req.GetDatastoreId(),
		PageIndex:     req.GetPageIndex(),
		PageSize:      req.GetPageSize(),
		FieldID:       req.GetFieldId(),
		CreatedAtFrom: req.GetCreatedAtFrom(),
		CreatedAtTo:   req.GetCreatedAtTo(),
		OldValue:      req.GetOldValue(),
		NewValue:      req.GetNewValue(),
		FieldList:     req.GetFieldList(),
	}

	histories, total, err := model.FindHistories(ctx, req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindHistories, err.Error())
		return err
	}

	res := &datahistory.HistoriesResponse{}
	for _, it := range histories {
		res.Histories = append(res.Histories, it.ToProto())
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindHistories, utils.MsgProcessEnded)
	return nil
}

// Download 下载履历
func (i *History) Download(ctx context.Context, req *datahistory.DownloadRequest, stream datahistory.HistoryService_DownloadStream) error {
	utils.InfoLog(ActionFindHistories, utils.MsgProcessStarted)

	params := model.FindParam{
		ItemID:        req.GetItemId(),
		HistoryType:   req.GetHistoryType(),
		DatastoreID:   req.GetDatastoreId(),
		FieldID:       req.GetFieldId(),
		CreatedAtFrom: req.GetCreatedAtFrom(),
		CreatedAtTo:   req.GetCreatedAtTo(),
		OldValue:      req.GetOldValue(),
		NewValue:      req.GetNewValue(),
		FieldList:     req.GetFieldList(),
	}

	err := model.DownloadHistories(ctx, req.GetDatabase(), params, stream)
	if err != nil {
		utils.ErrorLog(ActionFindHistories, err.Error())
		return err
	}

	utils.InfoLog(ActionFindHistories, utils.MsgProcessEnded)
	return nil
}

// FindHistoryCount 获取履历件数
func (i *History) FindHistoryCount(ctx context.Context, req *datahistory.CountRequest, rsp *datahistory.CountResponse) error {
	utils.InfoLog(ActionFindHistories, utils.MsgProcessStarted)

	params := model.FindParam{
		ItemID:        req.GetItemId(),
		HistoryType:   req.GetHistoryType(),
		DatastoreID:   req.GetDatastoreId(),
		FieldID:       req.GetFieldId(),
		CreatedAtFrom: req.GetCreatedAtFrom(),
		CreatedAtTo:   req.GetCreatedAtTo(),
		OldValue:      req.GetOldValue(),
		NewValue:      req.GetNewValue(),
		FieldList:     req.GetFieldList(),
	}

	total, err := model.FindHistoryCount(ctx, req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindHistories, err.Error())
		return err
	}

	rsp.Total = total

	utils.InfoLog(ActionFindHistories, utils.MsgProcessEnded)
	return nil
}

// FindLastHistories 通过ID获取履历数据
func (i *History) FindLastHistories(ctx context.Context, req *datahistory.LastRequest, rsp *datahistory.LastResponse) error {
	utils.InfoLog(ActionFindLastHistories, utils.MsgProcessStarted)

	total, histories, err := model.FindLastHistories(ctx, req.GetDatabase(), req.GetDatastoreId(), req.GetItemId(), req.GetFieldList())
	if err != nil {
		utils.ErrorLog(ActionFindLastHistories, err.Error())
		return err
	}

	res := &datahistory.LastResponse{}
	for _, it := range histories {
		res.HistoryList = append(res.HistoryList, it.ToProto())
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindLastHistories, utils.MsgProcessEnded)
	return nil
}

// FindHistory 通过ID获取履历数据
func (i *History) FindHistory(ctx context.Context, req *datahistory.HistoryRequest, rsp *datahistory.HistoryResponse) error {
	utils.InfoLog(ActionFindHistory, utils.MsgProcessStarted)

	res, err := model.FindHistory(ctx, req.GetDatabase(), req.GetHistoryId(), req.GetFieldList())
	if err != nil {
		utils.ErrorLog(ActionFindHistory, err.Error())
		return err
	}

	rsp.History = res.ToProto()

	utils.InfoLog(ActionFindHistory, utils.MsgProcessEnded)
	return nil
}

// CreateHistoryIndex 创建history索引
func (i *History) CreateIndex(ctx context.Context, req *datahistory.CreateIndexRequest, rsp *datahistory.CreateIndexResponse) error {
	utils.InfoLog(ActionCreateHistoryIndex, utils.MsgProcessStarted)

	err := model.CreateIndex(req.GetCustomerId())
	if err != nil {
		utils.ErrorLog(ActionCreateHistoryIndex, err.Error())
		return err
	}

	utils.InfoLog(ActionCreateHistoryIndex, utils.MsgProcessEnded)
	return nil
}
