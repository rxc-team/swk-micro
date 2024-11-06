package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/task/model"
	"rxcsoft.cn/pit3/srv/task/proto/history"
	"rxcsoft.cn/pit3/srv/task/utils"
)

// History 台账数据
type History struct{}

// log出力使用
const (
	HistoryProcessName  = "TaskHistory"
	ActionFindHistories = "FindHistories"
)

// DownloadHistories 下载履历数据
func (i *History) DownloadHistories(ctx context.Context, req *history.DownloadRequest, rsp *history.DownloadResponse) error {
	utils.InfoLog(ActionFindHistories, utils.MsgProcessStarted)

	histories, total, err := model.DownloadHistories(req.GetDatabase(), req.GetUserId(), req.GetAppId(), req.GetScheduleId(), req.GetJobId())
	if err != nil {
		utils.ErrorLog(ActionFindHistories, err.Error())
		return err
	}

	res := &history.DownloadResponse{}
	for _, it := range histories {
		res.Histories = append(res.Histories, it.ToProto())
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindHistories, utils.MsgProcessEnded)
	return nil
}

// FindHistories 获取多条履历数据
func (i *History) FindHistories(ctx context.Context, req *history.HistoriesRequest, rsp *history.HistoriesResponse) error {
	utils.InfoLog(ActionFindHistories, utils.MsgProcessStarted)

	histories, total, err := model.FindHistories(req.GetDatabase(), req.GetUserId(), req.GetAppId(), req.GetScheduleId(), req.GetJobId(), req.GetPageIndex(), req.GetPageSize())
	if err != nil {
		utils.ErrorLog(ActionFindHistories, err.Error())
		return err
	}

	res := &history.HistoriesResponse{}
	for _, it := range histories {
		res.Histories = append(res.Histories, it.ToProto())
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindHistories, utils.MsgProcessEnded)
	return nil
}
