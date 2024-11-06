package handler

import (
	"context"
	"fmt"

	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/common/storex"
	"rxcsoft.cn/pit3/srv/import/model/csv"
	"rxcsoft.cn/pit3/srv/import/model/inventory"
	"rxcsoft.cn/pit3/srv/import/model/lease"
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

// CSVUpload 读取上传文件并导入
func (f *Upload) CSVUpload(ctx context.Context, req *upload.CSVRequest, rsp *upload.CSVResponse) error {
	loggerx.InfoLog(ActionCSVUpload, loggerx.MsgProcessStarted)

	action := req.GetBaseParams().GetAction()

	if action == "insert" || action == "update" || action == "image" {
		base := csv.Params{
			JobId:        req.GetBaseParams().GetJobId(),
			Action:       req.GetBaseParams().GetAction(),
			Encoding:     req.GetBaseParams().GetEncoding(),
			ZipCharset:   req.GetBaseParams().GetZipCharset(),
			UserId:       req.GetBaseParams().GetUserId(),
			AppId:        req.GetBaseParams().GetAppId(),
			Lang:         req.GetBaseParams().GetLang(),
			Domain:       req.GetBaseParams().GetDomain(),
			DatastoreId:  req.GetBaseParams().GetDatastoreId(),
			GroupId:      req.GetBaseParams().GetGroupId(),
			UpdateOwners: req.GetBaseParams().GetAccessKeys(),
			Owners:       req.GetBaseParams().GetOwners(),
			Roles:        req.GetBaseParams().GetRoles(),
			WfId:         req.GetBaseParams().GetWfId(),
			EmptyChange:  req.GetBaseParams().GetEmptyChange(),
			Database:     req.GetBaseParams().GetDatabase(),
		}
		file := csv.FileParams{
			FilePath:    req.GetFileParams().GetFilePath(),
			ZipFilePath: req.GetFileParams().GetZipFilePath(),
		}

		store := storex.NewRedisStore(600)
		uploadID := base.AppId + "upload"
		store.Set(uploadID, "upload")
		go func() {
			csv.Import(base, file)
			store.Set(uploadID, "")
		}()

		loggerx.InfoLog(ActionCSVUpload, loggerx.MsgProcessEnded)
		return nil
	}
	if action == "contract-insert" || action == "debt-change" || action == "info-change" || action == "midway-cancel" || action == "contract-expire" {
		base := lease.Params{
			JobId:        req.GetBaseParams().GetJobId(),
			Action:       req.GetBaseParams().GetAction(),
			Encoding:     req.GetBaseParams().GetEncoding(),
			UserId:       req.GetBaseParams().GetUserId(),
			AppId:        req.GetBaseParams().GetAppId(),
			Lang:         req.GetBaseParams().GetLang(),
			Domain:       req.GetBaseParams().GetDomain(),
			DatastoreId:  req.GetBaseParams().GetDatastoreId(),
			GroupId:      req.GetBaseParams().GetGroupId(),
			UpdateOwners: req.GetBaseParams().GetAccessKeys(),
			Owners:       req.GetBaseParams().GetOwners(),
			Roles:        req.GetBaseParams().GetRoles(),
			Database:     req.GetBaseParams().GetDatabase(),
			FirstMonth:   req.GetBaseParams().GetFirstMonth(),
		}
		file := lease.FileParams{
			FilePath:    req.GetFileParams().GetFilePath(),
			PayFilePath: req.GetFileParams().GetPayFilePath(),
		}

		store := storex.NewRedisStore(600)
		uploadID := base.AppId + "upload"
		store.Set(uploadID, "upload")
		go func() {
			lease.Import(base, file)
			store.Set(uploadID, "")
		}()

		loggerx.InfoLog(ActionCSVUpload, loggerx.MsgProcessEnded)
		return nil
	}

	loggerx.InfoLog(ActionCSVUpload, loggerx.MsgProcessEnded)
	return fmt.Errorf("action [%v] in not support", action)
}

// InventoryUpload 读取上传文件批量盘点
func (f *Upload) InventoryUpload(ctx context.Context, req *upload.InventoryRequest, rsp *upload.InventoryResponse) error {
	loggerx.InfoLog(ActionInventoryUpload, loggerx.MsgProcessStarted)

	base := inventory.Params{
		JobId:        req.GetBaseParams().GetJobId(),
		Encoding:     req.GetBaseParams().GetEncoding(),
		UserId:       req.GetBaseParams().GetUserId(),
		AppId:        req.GetBaseParams().GetAppId(),
		Lang:         req.GetBaseParams().GetLang(),
		Domain:       req.GetBaseParams().GetDomain(),
		DatastoreId:  req.GetBaseParams().GetDatastoreId(),
		GroupId:      req.GetBaseParams().GetGroupId(),
		UpdateOwners: req.GetBaseParams().GetAccessKeys(),
		Owners:       req.GetBaseParams().GetOwners(),
		Roles:        req.GetBaseParams().GetRoles(),
		Database:     req.GetBaseParams().GetDatabase(),
		MainKeys:     req.GetBaseParams().GetMainKeys(),
		CheckType:    req.GetBaseParams().GetCheckType(),
		CheckedAt:    req.GetBaseParams().GetCheckedAt(),
		CheckedBy:    req.GetBaseParams().GetCheckedBy(),
	}

	go func() {
		inventory.Import(base, req.GetFilePath())
	}()

	loggerx.InfoLog(ActionInventoryUpload, loggerx.MsgProcessEnded)
	return nil
}

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
