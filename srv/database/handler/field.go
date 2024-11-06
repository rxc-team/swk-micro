package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Field 台账字段
type Field struct{}

// log出力使用
const (
	FieldProcessName            = "Field"
	ActionFindAppFields         = "FindAppFields"
	ActionFindFields            = "FindFields"
	ActionFindField             = "FindField"
	ActionSetSequenceValue      = "SetSequenceValue"
	ActionVerifyFunc            = "VerifyFunc"
	ActionAddField              = "AddField"
	ActionBlukAddField          = "BlukAddField"
	ActionModifyField           = "ModifyField"
	ActionDeleteField           = "DeleteField"
	ActionDeleteDatastoreFields = "DeleteDatastoreFields"
	ActionDeleteSelectFields    = "DeleteSelectFields"
	ActionHardDeleteFields      = "HardDeleteFields"
	ActionRecoverSelectFields   = "RecoverSelectFields"
)

// FindAppFields 查找APP中多个字段
func (f *Field) FindAppFields(ctx context.Context, req *field.AppFieldsRequest, rsp *field.AppFieldsResponse) error {
	utils.InfoLog(ActionFindAppFields, utils.MsgProcessStarted)

	param := model.FindAppFieldsParam{
		AppID:             req.GetAppId(),
		FieldType:         req.GetFieldType(),
		LookUpDatastoreID: req.GetLookupDatastoreId(),
		InvalidatedIn:     req.GetInvalidatedIn(),
	}

	fields, err := model.FindAppFields(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionFindAppFields, err.Error())
		return err
	}

	res := &field.AppFieldsResponse{}
	for _, f := range fields {
		res.Fields = append(res.Fields, f.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindAppFields, utils.MsgProcessEnded)
	return nil
}

// FindFields 获取台账下的所有字段
func (f *Field) FindFields(ctx context.Context, req *field.FieldsRequest, rsp *field.FieldsResponse) error {
	utils.InfoLog(ActionFindFields, utils.MsgProcessStarted)

	param := model.FindFieldsParam{
		AppID:         req.GetAppId(),
		DatastoreID:   req.GetDatastoreId(),
		FieldName:     req.GetFieldName(),
		FieldType:     req.GetFieldType(),
		IsRequired:    req.GetIsRequired(),
		IsFixed:       req.GetIsFixed(),
		AsTitle:       req.GetAsTitle(),
		InvalidatedIn: req.GetInvalidatedIn(),
	}

	fields, err := model.FindFields(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionFindFields, err.Error())
		return err
	}

	res := &field.FieldsResponse{}
	for _, f := range fields {
		res.Fields = append(res.Fields, f.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindFields, utils.MsgProcessEnded)
	return nil
}

// FindField 通过ID获取用户
func (f *Field) FindField(ctx context.Context, req *field.FieldRequest, rsp *field.FieldResponse) error {
	utils.InfoLog(ActionFindField, utils.MsgProcessStarted)

	res, err := model.FindField(req.GetDatabase(), req.GetDatastoreId(), req.GetFieldId())
	if err != nil {
		utils.ErrorLog(ActionFindField, err.Error())
		return err
	}

	rsp.Field = res.ToProto()

	utils.InfoLog(ActionFindField, utils.MsgProcessEnded)
	return nil
}

// VerifyFunc 验证函数是否正确
func (f *Field) VerifyFunc(ctx context.Context, req *field.VerifyFuncRequest, rsp *field.VerifyFuncResponse) error {
	utils.InfoLog(ActionFindField, utils.MsgProcessStarted)

	res, params, err := model.VerifyFunc(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), req.GetFormula(), req.GetReturnType())
	rsp.Result = res
	if err != nil {
		rsp.Error = err.Error()
	}
	rsp.Params = params

	utils.InfoLog(ActionFindField, utils.MsgProcessEnded)
	return nil
}

// SetSequenceValue 设置序列值
func (f *Field) SetSequenceValue(ctx context.Context, req *field.SetSequenceValueRequest, rsp *field.SetSequenceValueResponse) error {
	utils.InfoLog(ActionSetSequenceValue, utils.MsgProcessStarted)

	// 设置序列值
	err := model.SetSequenceValue(req.GetDatabase(), req.GetSequenceName(), req.GetSequenceValue())
	if err != nil {
		utils.ErrorLog(ActionSetSequenceValue, err.Error())
		return err
	}

	utils.InfoLog(ActionSetSequenceValue, utils.MsgProcessEnded)
	return nil
}

// AddField 添加台账字段
func (f *Field) AddField(ctx context.Context, req *field.AddRequest, rsp *field.AddResponse) error {
	utils.InfoLog(ActionAddField, utils.MsgProcessStarted)
	if req.GetCols() <= 0 {
		req.Cols = 1
	}
	if req.GetRows() <= 0 {
		req.Rows = 1
	}
	if req.GetX() <= 0 {
		req.X = 0
	}
	if req.GetY() <= 0 {
		req.Y = 0
	}
	if req.GetWidth() <= 0 {
		req.Width = 100
	}

	params := model.Field{
		AppID:             req.GetAppId(),
		DatastoreID:       req.GetDatastoreId(),
		FieldID:           req.GetFieldId(),
		FieldName:         req.GetFieldName(),
		FieldType:         req.GetFieldType(),
		IsRequired:        req.GetIsRequired(),
		IsFixed:           req.GetIsFixed(),
		IsImage:           req.GetIsImage(),
		IsCheckImage:      req.GetIsCheckImage(),
		LookupAppID:       req.GetLookupAppId(),
		LookupDatastoreID: req.GetLookupDatastoreId(),
		LookupFieldID:     req.GetLookupFieldId(),
		UserGroupID:       req.GetUserGroupId(),
		OptionID:          req.GetOptionId(),
		Cols:              req.GetCols(),
		Rows:              req.GetRows(),
		X:                 req.GetX(),
		Y:                 req.GetY(),
		Width:             req.GetWidth(),
		MinLength:         req.GetMinLength(),
		MaxLength:         req.GetMaxLength(),
		MinValue:          req.GetMinValue(),
		MaxValue:          req.GetMaxValue(),
		Unique:            req.GetUnique(),
		DisplayOrder:      req.GetDisplayOrder(),
		DisplayDigits:     req.GetDisplayDigits(),
		Prefix:            req.GetPrefix(),
		Precision:         req.GetPrecision(),
		ReturnType:        req.GetReturnType(),
		Formula:           req.GetFormula(),
		SelfCalculate:     req.GetSelfCalculate(),
		AsTitle:           req.GetAsTitle(),
		CreatedAt:         time.Now(),
		CreatedBy:         req.GetWriter(),
		UpdatedAt:         time.Now(),
		UpdatedBy:         req.GetWriter(),
	}

	id, err := model.AddField(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddField, err.Error())
		return err
	}

	rsp.FieldId = id

	utils.InfoLog(ActionAddField, utils.MsgProcessEnded)

	return nil
}

// BlukAddField 批量添加台账字段
func (f *Field) BlukAddField(ctx context.Context, req *field.BlukAddRequest, rsp *field.BlukAddResponse) error {
	utils.InfoLog(ActionBlukAddField, utils.MsgProcessStarted)

	var params []*model.Field

	for _, f := range req.GetFields() {
		params = append(params, &model.Field{
			AppID:             f.GetAppId(),
			DatastoreID:       f.GetDatastoreId(),
			FieldID:           f.GetFieldId(),
			FieldName:         f.GetFieldName(),
			FieldType:         f.GetFieldType(),
			IsRequired:        f.GetIsRequired(),
			IsFixed:           f.GetIsFixed(),
			IsImage:           f.GetIsImage(),
			IsCheckImage:      f.GetIsCheckImage(),
			LookupAppID:       f.GetLookupAppId(),
			LookupDatastoreID: f.GetLookupDatastoreId(),
			LookupFieldID:     f.GetLookupFieldId(),
			UserGroupID:       f.GetUserGroupId(),
			OptionID:          f.GetOptionId(),
			Cols:              1,
			Rows:              1,
			X:                 0,
			Y:                 0,
			Width:             100.0,
			MinLength:         f.GetMinLength(),
			MaxLength:         f.GetMaxLength(),
			MinValue:          f.GetMinValue(),
			MaxValue:          f.GetMaxValue(),
			Unique:            f.GetUnique(),
			DisplayOrder:      f.GetDisplayOrder(),
			DisplayDigits:     f.GetDisplayDigits(),
			Precision:         f.GetPrecision(),
			Prefix:            f.GetPrefix(),
			ReturnType:        f.GetReturnType(),
			Formula:           f.GetFormula(),
			SelfCalculate:     f.GetSelfCalculate(),
			AsTitle:           f.GetAsTitle(),
			CreatedAt:         time.Now(),
			CreatedBy:         f.GetWriter(),
			UpdatedAt:         time.Now(),
			UpdatedBy:         f.GetWriter(),
		})
	}

	err := model.BlukAddField(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionBlukAddField, err.Error())
		return err
	}

	utils.InfoLog(ActionBlukAddField, utils.MsgProcessEnded)
	return nil
}

// ModifyField 更新台账的字段
func (f *Field) ModifyField(ctx context.Context, req *field.ModifyRequest, rsp *field.ModifyResponse) error {
	utils.InfoLog(ActionModifyField, utils.MsgProcessStarted)

	params := model.ModifyFieldParam{
		FieldID:           req.GetFieldId(),
		AppID:             req.GetAppId(),
		DatastoreID:       req.GetDatastoreId(),
		FieldName:         req.GetFieldName(),
		FieldType:         req.GetFieldType(),
		IsRequired:        req.GetIsRequired(),
		IsFixed:           req.GetIsFixed(),
		IsImage:           req.GetIsImage(),
		IsCheckImage:      req.GetIsCheckImage(),
		AsTitle:           req.GetAsTitle(),
		Unique:            req.GetUnique(),
		LookupAppID:       req.GetLookupAppId(),
		LookupDatastoreID: req.GetLookupDatastoreId(),
		LookupFieldID:     req.GetLookupFieldId(),
		UserGroupID:       req.GetUserGroupId(),
		OptionID:          req.GetOptionId(),
		Cols:              req.GetCols(),
		Rows:              req.GetRows(),
		X:                 req.GetX(),
		Y:                 req.GetY(),
		Width:             req.GetWidth(),
		MinLength:         req.GetMinLength(),
		MaxLength:         req.GetMaxLength(),
		MinValue:          req.GetMinValue(),
		MaxValue:          req.GetMaxValue(),
		DisplayOrder:      req.GetDisplayOrder(),
		DisplayDigits:     req.GetDisplayDigits(),
		Precision:         req.GetPrecision(),
		Prefix:            req.GetPrefix(),
		ReturnType:        req.GetReturnType(),
		Formula:           req.GetFormula(),
		SelfCalculate:     req.GetSelfCalculate(),
		IsDisplaySetting:  req.GetIsDisplaySetting(),
		Writer:            req.GetWriter(),
	}

	err := model.ModifyField(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyField, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyField, utils.MsgProcessEnded)
	return nil
}

// DeleteField 删除台账字段
func (f *Field) DeleteField(ctx context.Context, req *field.DeleteRequest, rsp *field.DeleteResponse) error {
	utils.InfoLog(ActionDeleteField, utils.MsgProcessStarted)

	err := model.DeleteField(req.GetDatabase(), req.GetDatastoreId(), req.GetFieldId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteField, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteField, utils.MsgProcessEnded)
	return nil
}

// DeleteDatastoreFields 删除台账字段
func (f *Field) DeleteDatastoreFields(ctx context.Context, req *field.DeleteDatastoreFieldsRequest, rsp *field.DeleteResponse) error {
	utils.InfoLog(ActionDeleteDatastoreFields, utils.MsgProcessStarted)

	err := model.DeleteDatastoreFields(req.GetDatabase(), req.GetDatastoreId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteDatastoreFields, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteDatastoreFields, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectFields 删除选中台账字段
func (f *Field) DeleteSelectFields(ctx context.Context, req *field.DeleteSelectFieldsRequest, rsp *field.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectFields, utils.MsgProcessStarted)

	err := model.DeleteSelectFields(req.GetDatabase(), req.GetDatastoreId(), req.GetFieldIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectFields, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectFields, utils.MsgProcessEnded)
	return nil
}

// HardDeleteFields 物理删除选中台账字段
func (f *Field) HardDeleteFields(ctx context.Context, req *field.HardDeleteFieldsRequest, rsp *field.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteFields, utils.MsgProcessStarted)

	err := model.HardDeleteFields(ctx, req.GetDatabase(), req.GetDatastoreId(), req.GetFieldIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteFields, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteFields, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectFields 恢复选中的字段
func (f *Field) RecoverSelectFields(ctx context.Context, req *field.RecoverSelectFieldsRequest, rsp *field.RecoverSelectFieldsResponse) error {
	utils.InfoLog(ActionRecoverSelectFields, utils.MsgProcessStarted)

	err := model.RecoverSelectFields(req.GetDatabase(), req.GetDatastoreId(), req.GetFieldIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectFields, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectFields, utils.MsgProcessEnded)
	return nil
}
