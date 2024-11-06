package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/global/model"
	types "rxcsoft.cn/pit3/srv/global/proto/help-type"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Type 帮助文档类型
type Type struct{}

// log出力使用
const (
	ActionFindTypes   = "FindTypes"
	ActionFindType    = "FindType"
	ActionAddType     = "AddType"
	ActionModifyType  = "ModifyType"
	ActionDeleteType  = "DeleteType"
	ActionDeleteTypes = "DeleteTypes"
)

// FindType 获取单个帮助文档类型
func (t *Type) FindType(ctx context.Context, req *types.FindTypeRequest, rsp *types.FindTypeResponse) error {
	utils.InfoLog(ActionFindType, utils.MsgProcessStarted)

	res, err := model.FindType(req.GetDatabase(), req.GetTypeId())
	if err != nil {
		utils.ErrorLog(ActionFindType, err.Error())
		return err
	}

	rsp.Type = res.ToProto()

	utils.InfoLog(ActionFindType, utils.MsgProcessEnded)
	return nil
}

// FindTypes 获取多个帮助文档类型
func (t *Type) FindTypes(ctx context.Context, req *types.FindTypesRequest, rsp *types.FindTypesResponse) error {
	utils.InfoLog(ActionFindTypes, utils.MsgProcessStarted)

	typeList, err := model.FindTypes(req.GetDatabase(), req.GetTypeName(), req.GetShow(), req.GetLangCd())
	if err != nil {
		utils.ErrorLog(ActionFindTypes, err.Error())
		return err
	}

	res := &types.FindTypesResponse{}

	for _, t := range typeList {
		res.Types = append(res.Types, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindTypes, utils.MsgProcessEnded)
	return nil
}

// AddType 添加帮助文档类型
func (t *Type) AddType(ctx context.Context, req *types.AddTypeRequest, rsp *types.AddTypeResponse) error {
	utils.InfoLog(ActionAddType, utils.MsgProcessStarted)

	params := model.Type{
		TypeID:    req.GetTypeId(),
		TypeName:  req.GetTypeName(),
		Show:      req.GetShow(),
		Icon:      req.GetIcon(),
		LangCd:    req.GetLangCd(),
		CreatedAt: time.Now(),
		CreatedBy: req.GetWriter(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	id, err := model.AddType(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddType, err.Error())
		return err
	}

	rsp.TypeId = id

	utils.InfoLog(ActionAddType, utils.MsgProcessEnded)
	return nil
}

// ModifyType 更新帮助文档类型
func (t *Type) ModifyType(ctx context.Context, req *types.ModifyTypeRequest, rsp *types.ModifyTypeResponse) error {
	utils.InfoLog(ActionModifyType, utils.MsgProcessStarted)

	params := model.Type{
		TypeID:    req.GetTypeId(),
		TypeName:  req.GetTypeName(),
		Show:      req.GetShow(),
		Icon:      req.GetIcon(),
		LangCd:    req.GetLangCd(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	err := model.ModifyType(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyType, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyType, utils.MsgProcessEnded)
	return nil
}

// DeleteType 硬删除帮助文档类型
func (t *Type) DeleteType(ctx context.Context, req *types.DeleteTypeRequest, rsp *types.DeleteTypeResponse) error {
	utils.InfoLog(ActionDeleteType, utils.MsgProcessStarted)

	err := model.DeleteType(req.GetDatabase(), req.GetTypeId())
	if err != nil {
		utils.ErrorLog(ActionDeleteType, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteType, utils.MsgProcessEnded)
	return nil
}

// DeleteTypes 硬删除多个帮助文档类型
func (t *Type) DeleteTypes(ctx context.Context, req *types.DeleteTypesRequest, rsp *types.DeleteTypesResponse) error {
	utils.InfoLog(ActionDeleteTypes, utils.MsgProcessStarted)

	err := model.DeleteTypes(req.GetDatabase(), req.GetTypeIdList())
	if err != nil {
		utils.ErrorLog(ActionDeleteTypes, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteTypes, utils.MsgProcessEnded)
	return nil
}
