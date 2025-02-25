package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/model/mapping"
	"rxcsoft.cn/pit3/srv/import/proto/upload"
)

// Upload 上传服务
type Upload struct{}

// log出力使用
const (
	UploadProcessName     = "Upload"
	ActionCSVUpload       = "CSVUpload"
	ActionInventoryUpload = "InventoryUpload"
	ActionMappingUpload   = "MappingUpload"
)

// InventoryUpload 读取上传文件批量盘点
func (f *Upload) MappingUpload(ctx context.Context, req *upload.MappingRequest, rsp *upload.MappingResponse) error {
	loggerx.InfoLog(ActionMappingUpload, loggerx.MsgProcessStarted)

	base := mapping.Params{
		MappingID:    req.BaseParams.GetMappingId(),
		JobId:        req.GetBaseParams().GetJobId(),
		UserId:       req.GetBaseParams().GetUserId(),
		AppId:        req.GetBaseParams().GetAppId(),
		Lang:         req.GetBaseParams().GetLang(),
		Domain:       req.GetBaseParams().GetDomain(),
		DatastoreId:  req.GetBaseParams().GetDatastoreId(),
		UpdateOwners: req.GetBaseParams().GetAccessKeys(),
		Owners:       req.GetBaseParams().GetOwners(),
		Roles:        req.GetBaseParams().GetRoles(),
		Database:     req.GetBaseParams().GetDatabase(),
		EmptyChange:  req.GetBaseParams().GetEmptyChange(),
	}

	go func() {
		mapping.Import(base, req.GetFilePath())
	}()

	loggerx.InfoLog(ActionMappingUpload, loggerx.MsgProcessEnded)
	return nil
}
