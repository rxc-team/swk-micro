package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/check"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// CheckHistory 台账数据
type CheckHistory struct{}

// log出力使用
const (
	ActionFindCheckHistories    = "FindHistories"
	ActionFindCheckHistoryCount = "FindHistoryCount"
	ActionFindCheckHistory      = "FindHistory"
	ActionDeleteCheckHistories  = "HardDeleteHistories"
)

// FindHistories 获取多条履历数据
func (i *CheckHistory) FindHistories(ctx context.Context, req *check.HistoriesRequest, rsp *check.HistoriesResponse) error {
	utils.InfoLog(ActionFindCheckHistories, utils.MsgProcessStarted)

	params := model.CheckSearchParam{
		ItemId:         req.GetItemId(),
		DatastoreId:    req.GetDatastoreId(),
		CheckType:      req.GetCheckType(),
		CheckStartDate: req.GetCheckStartDate(),
		CheckedAtFrom:  req.GetCheckedAtFrom(),
		CheckedAtTo:    req.GetCheckedAtTo(),
		CheckedBy:      req.GetCheckedBy(),
		DisplayFields:  req.GetDisplayFields(),
		PageIndex:      req.GetPageIndex(),
		PageSize:       req.GetPageSize(),
	}

	histories, total, err := model.FindCheckHistories(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindCheckHistories, err.Error())
		return err
	}

	res := &check.HistoriesResponse{}
	for _, it := range histories {
		res.Histories = append(res.Histories, it.ToProto())
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindCheckHistories, utils.MsgProcessEnded)
	return nil
}

// Download 下载履历
func (i *CheckHistory) Download(ctx context.Context, req *check.DownloadRequest, stream check.CheckHistoryService_DownloadStream) error {
	utils.InfoLog(ActionFindCheckHistories, utils.MsgProcessStarted)

	params := model.CheckSearchParam{
		ItemId:         req.GetItemId(),
		DatastoreId:    req.GetDatastoreId(),
		CheckType:      req.GetCheckType(),
		CheckStartDate: req.GetCheckStartDate(),
		CheckedAtFrom:  req.GetCheckedAtFrom(),
		CheckedAtTo:    req.GetCheckedAtTo(),
		CheckedBy:      req.GetCheckedBy(),
		DisplayFields:  req.GetDisplayFields(),
	}

	err := model.DownloadCheckHistories(req.GetDatabase(), params, stream)
	if err != nil {
		utils.ErrorLog(ActionFindCheckHistories, err.Error())
		return err
	}

	utils.InfoLog(ActionFindCheckHistories, utils.MsgProcessEnded)
	return nil
}

// FindHistoryCount 获取履历件数
func (i *CheckHistory) FindHistoryCount(ctx context.Context, req *check.CountRequest, rsp *check.CountResponse) error {
	utils.InfoLog(ActionFindCheckHistoryCount, utils.MsgProcessStarted)

	params := model.CheckSearchParam{
		ItemId:         req.GetItemId(),
		DatastoreId:    req.GetDatastoreId(),
		CheckType:      req.GetCheckType(),
		CheckStartDate: req.GetCheckStartDate(),
		CheckedAtFrom:  req.GetCheckedAtFrom(),
		CheckedAtTo:    req.GetCheckedAtTo(),
		CheckedBy:      req.GetCheckedBy(),
	}

	total, err := model.FindCheckHistoryCount(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindCheckHistoryCount, err.Error())
		return err
	}

	rsp.Total = total

	utils.InfoLog(ActionFindCheckHistoryCount, utils.MsgProcessEnded)
	return nil
}

// HardDeleteHistories 物理删除选中台账履历数据
func (i *CheckHistory) DeleteHistories(ctx context.Context, req *check.DeleteRequest, rsp *check.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteHistories, utils.MsgProcessStarted)

	err := model.DeleteCheckHistories(req.GetDatabase(), req.GetCheckIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteHistories, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteHistories, utils.MsgProcessEnded)
	return nil
}
