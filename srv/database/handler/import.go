package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/feed"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Import 台账导入
type Import struct{}

// log出力使用
const (
	ImportProcessName      = "Import"
	ActionFindImports      = "FindImports"
	ActionFindImportItems  = "FindImportItems"
	ActionAddImportItem    = "AddImportItem"
	ActionDeleteImportItem = "DeleteImportItem"
)

// FindImports 查找当前用户导入的记录
func (f *Import) FindImports(ctx context.Context, req *feed.ImportsRequest, rsp *feed.ImportsResponse) error {
	utils.InfoLog(ActionFindImports, utils.MsgProcessStarted)

	param := model.Request{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		JobID:       req.GetJobId(),
		Writer:      req.GetWriter(),
	}

	items, err := model.FindImports(req.GetDatabase(), param)
	if err != nil {
		utils.ErrorLog(ActionFindImports, err.Error())
		return err
	}

	res := &feed.ImportsResponse{}
	for _, i := range items {
		res.Items = append(res.Items, i.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindImports, utils.MsgProcessEnded)
	return nil
}

// FindImportItems 查找当前用户导入的数据
func (f *Import) FindImportItems(ctx context.Context, req *feed.ImportItemsRequest, rsp *feed.ImportItemsResponse) error {
	utils.InfoLog(ActionFindImportItems, utils.MsgProcessStarted)

	param := model.Request{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		JobID:       req.GetJobId(),
		Writer:      req.GetWriter(),
		MappingID:   req.GetMappingId(),
	}

	items, err := model.FindImportItems(req.GetDatabase(), param)
	if err != nil {
		utils.ErrorLog(ActionFindImportItems, err.Error())
		return err
	}

	res := &feed.ImportItemsResponse{}
	for _, i := range items {
		res.Items = append(res.Items, i.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindImportItems, utils.MsgProcessEnded)
	return nil
}

// AddImportItem 导入数据
func (f *Import) AddImportItem(ctx context.Context, req *feed.AddRequest, rsp *feed.AddResponse) error {
	utils.InfoLog(ActionAddImportItem, utils.MsgProcessStarted)

	var params []*model.ImItem
	for _, it := range req.Items {
		params = append(params, &model.ImItem{
			AppID:       it.AppId,
			DatastoreID: it.DatastoreId,
			JobID:       it.JobId,
			MappingID:   it.MappingId,
			Items:       it.Items,
			CreatedAt:   time.Now(),
			CreatedBy:   req.Writer,
		})
	}

	err := model.AddImportItem(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionAddImportItem, err.Error())
		return err
	}

	utils.InfoLog(ActionAddImportItem, utils.MsgProcessEnded)
	return nil
}

// DeleteImportItem 删除数据
func (f *Import) DeleteImportItem(ctx context.Context, req *feed.DeleteRequest, rsp *feed.DeleteResponse) error {
	utils.InfoLog(ActionDeleteImportItem, utils.MsgProcessStarted)

	param := model.Request{
		JobID: req.GetJobId(),
	}

	err := model.DeleteImportItem(req.GetDatabase(), param)
	if err != nil {
		utils.ErrorLog(ActionDeleteImportItem, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteImportItem, utils.MsgProcessEnded)
	return nil
}
