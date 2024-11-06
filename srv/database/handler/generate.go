package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/generate"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Generate 台账
type Generate struct{}

// log出力使用
const (
	GenerateProcessName        = "Generate"
	ActionFindGenerateConfig   = "FindGenerateConfig"
	ActionAddGenerateConfig    = "AddGenerateConfig"
	ActionModifyGenerateConfig = "ModifyGenerateConfig"
	ActionUploadData           = "UploadData"
	ActionFindRowData          = "FindRowData"
	ActionFindColumnData       = "FindColumnData"
	ActionDeleteGenerateConfig = "DeleteGenerateConfig"
)

// FindGenerateConfig 获取app下所有台账
func (d *Generate) FindGenerateConfig(ctx context.Context, req *generate.FindRequest, rsp *generate.FindResponse) error {
	utils.InfoLog(ActionFindGenerateConfig, utils.MsgProcessStarted)

	result, err := model.FindGenerateConfig(req.GetDatabase(), req.GetAppId(), req.GetUserId())
	if err != nil {
		utils.ErrorLog(ActionFindGenerateConfig, err.Error())
		return err
	}

	rsp.GenConfig = result.ToProto()

	utils.InfoLog(ActionFindGenerateConfig, utils.MsgProcessEnded)
	return nil
}

// FindRowData 删除单个台账
func (d *Generate) FindRowData(ctx context.Context, req *generate.RowRequest, rsp *generate.RowResponse) error {
	utils.InfoLog(ActionFindRowData, utils.MsgProcessStarted)

	result, err := model.FindRowData(req.GetDatabase(), req.GetAppId(), req.GetUserId(), req.GetPageIndex(), req.GetPageSize())
	if err != nil {
		utils.ErrorLog(ActionFindRowData, err.Error())
		return err
	}

	var items []*generate.Item
	for _, v := range result {
		items = append(items, v.ToProto())
	}

	rsp.Items = items

	utils.InfoLog(ActionFindRowData, utils.MsgProcessEnded)
	return nil
}

// FindColumnData 删除单个台账
func (d *Generate) FindColumnData(ctx context.Context, req *generate.ColumnRequest, rsp *generate.ColumnResponse) error {
	utils.InfoLog(ActionFindColumnData, utils.MsgProcessStarted)

	result, err := model.FindColumnData(req.GetDatabase(), req.GetAppId(), req.GetUserId(), req.GetColumnName())
	if err != nil {
		utils.ErrorLog(ActionFindColumnData, err.Error())
		return err
	}

	rsp.Items = result

	utils.InfoLog(ActionFindColumnData, utils.MsgProcessEnded)
	return nil
}

// AddGenerateConfig 通过ID获取台账信息
func (d *Generate) AddGenerateConfig(ctx context.Context, req *generate.AddRequest, rsp *generate.AddResponse) error {
	utils.InfoLog(ActionAddGenerateConfig, utils.MsgProcessStarted)

	data := &model.GenerateConfig{
		AppID:  req.GetAppId(),
		UserID: req.GetUserId(),
	}

	err := model.AddGenerateConfig(req.GetDatabase(), data)
	if err != nil {
		utils.ErrorLog(ActionAddGenerateConfig, err.Error())
		return err
	}

	utils.InfoLog(ActionAddGenerateConfig, utils.MsgProcessEnded)
	return nil
}

// ModifyGenerateConfig 更新台账的字段
func (d *Generate) ModifyGenerateConfig(ctx context.Context, req *generate.ModifyRequest, rsp *generate.ModifyResponse) error {
	utils.InfoLog(ActionModifyGenerateConfig, utils.MsgProcessStarted)
	var fields []*model.GField
	for _, f := range req.GetFields() {
		fs := &model.GField{
			FieldID:           f.FieldId,
			AppID:             f.AppId,
			DatastoreID:       f.DatastoreId,
			FieldName:         f.FieldName,
			FieldType:         f.FieldType,
			IsRequired:        f.IsRequired,
			IsFixed:           f.IsFixed,
			IsImage:           f.IsImage,
			IsCheckImage:      f.IsCheckImage,
			Unique:            f.Unique,
			LookupAppID:       f.LookupAppId,
			LookupDatastoreID: f.LookupDatastoreId,
			LookupFieldID:     f.LookupFieldId,
			UserGroupID:       f.UserGroupId,
			OptionID:          f.OptionId,
			MinLength:         f.MinLength,
			MaxLength:         f.MaxLength,
			MinValue:          f.MinValue,
			MaxValue:          f.MaxValue,
			DisplayOrder:      f.DisplayOrder,
			DisplayDigits:     f.DisplayDigits,
			Precision:         f.Precision,
			Prefix:            f.Prefix,
			ReturnType:        f.ReturnType,
			Formula:           f.Formula,
			AsTitle:           f.AsTitle,
			CsvHeader:         f.CsvHeader,
			CanChange:         f.CanChange,
			IsEmptyLine:       f.IsEmptyLine,
			CheckErrors:       f.CheckErrors,
		}
		fields = append(fields, fs)
	}

	params := model.GModifyReq{
		DatastoreID:   req.GetDatastoreId(),
		DatastoreName: req.GetDatastoreName(),
		ApiKey:        req.GetApiKey(),
		CanCheck:      req.GetCanCheck(),
		Step:          req.GetStep(),
		MappingID:     req.GetMappingId(),
		Fields:        fields,
	}

	err := model.ModifyGenerateConfig(req.GetDatabase(), req.GetAppId(), req.GetUserId(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyGenerateConfig, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyGenerateConfig, utils.MsgProcessEnded)
	return nil
}

// UploadData 修改台账映射
func (d *Generate) UploadData(ctx context.Context, req *generate.UploadRequest, rsp *generate.UploadResponse) error {
	utils.InfoLog(ActionUploadData, utils.MsgProcessStarted)

	var items []*model.GItem
	for _, it := range req.GetItems() {
		item := &model.GItem{
			AppID:   it.AppId,
			UserID:  it.UserId,
			ItemMap: it.ItemMap,
		}
		items = append(items, item)
	}

	err := model.UploadData(req.GetDatabase(), items)
	if err != nil {
		utils.ErrorLog(ActionUploadData, err.Error())
		return err
	}

	utils.InfoLog(ActionUploadData, utils.MsgProcessEnded)

	return nil
}

// DeleteGenerateConfig 删除台账映射
func (d *Generate) DeleteGenerateConfig(ctx context.Context, req *generate.DeleteRequest, rsp *generate.DeleteResponse) error {
	utils.InfoLog(ActionDeleteGenerateConfig, utils.MsgProcessStarted)

	err := model.DeleteGenerateConfig(req.GetDatabase(), req.GetAppId(), req.GetUserId())
	if err != nil {
		utils.ErrorLog(ActionDeleteGenerateConfig, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteGenerateConfig, utils.MsgProcessEnded)

	return nil
}
