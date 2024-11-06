package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// log出力使用
const (
	ActionModifyContract    = "ModifyContract"
	ActionChangeDebt        = "ChangeDebt"
	ActionContractExpire    = "ContractExpire"
	ActionTerminateContract = "TerminateContract"
)

// ModifyContract 契约情报变更
func (i *Item) ModifyContract(ctx context.Context, req *item.ModifyContractRequest, rsp *item.ModifyContractResponse) error {
	utils.InfoLog(ActionModifyContract, utils.MsgProcessStarted)

	items := make(map[string]*model.Value, len(req.Items))
	for key, item := range req.Items {
		items[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetValueFromProto(item),
		}
	}

	params := model.ItemUpdateParam{
		AppID:       req.GetAppId(),
		ItemID:      req.GetItemId(),
		DatastoreID: req.GetDatastoreId(),
		ItemMap:     items,
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
		Owners:      req.GetOwners(),
		Lang:        req.GetLangCd(),
		Domain:      req.GetDomain(),
	}

	err := model.ModifyContract(req.GetDatabase(), req.GetWriter(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyContract, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyContract, utils.MsgProcessEnded)
	return nil
}

// ChangeDebt 债务变更
func (i *Item) ChangeDebt(ctx context.Context, req *item.ChangeDebtRequest, rsp *item.ChangeDebtResponse) error {
	utils.InfoLog(ActionChangeDebt, utils.MsgProcessStarted)

	items := make(map[string]*model.Value, len(req.Items))
	for key, item := range req.Items {
		items[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetValueFromProto(item),
		}
	}

	params := model.ItemUpdateParam{
		AppID:       req.GetAppId(),
		ItemID:      req.GetItemId(),
		DatastoreID: req.GetDatastoreId(),
		ItemMap:     items,
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
		Owners:      req.GetOwners(),
		Lang:        req.GetLangCd(),
		Domain:      req.GetDomain(),
	}

	err := model.ChangeDebt(req.GetDatabase(), req.GetWriter(), &params)
	if err != nil {
		utils.ErrorLog(ActionChangeDebt, err.Error())
		return err
	}

	utils.InfoLog(ActionChangeDebt, utils.MsgProcessEnded)
	return nil
}

// TerminateContract 中途解约
func (i *Item) TerminateContract(ctx context.Context, req *item.TerminateContractRequest, rsp *item.TerminateContractResponse) error {
	utils.InfoLog(ActionTerminateContract, utils.MsgProcessStarted)

	items := make(map[string]*model.Value, len(req.Items))
	for key, item := range req.Items {
		items[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetValueFromProto(item),
		}
	}

	params := model.ItemUpdateParam{
		AppID:       req.GetAppId(),
		ItemID:      req.GetItemId(),
		DatastoreID: req.GetDatastoreId(),
		ItemMap:     items,
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
		Owners:      req.GetOwners(),
		Lang:        req.GetLangCd(),
		Domain:      req.GetDomain(),
	}

	err := model.TerminateContract(req.GetDatabase(), req.GetWriter(), &params)
	if err != nil {
		utils.ErrorLog(ActionTerminateContract, err.Error())
		return err
	}

	utils.InfoLog(ActionTerminateContract, utils.MsgProcessEnded)
	return nil
}

// ContractExpire 契约满了
func (i *Item) ContractExpire(ctx context.Context, req *item.ContractExpireRequest, rsp *item.ContractExpireResponse) error {
	utils.InfoLog(ActionContractExpire, utils.MsgProcessStarted)

	items := make(map[string]*model.Value, len(req.Items))
	for key, item := range req.Items {
		items[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetValueFromProto(item),
		}
	}

	params := model.ItemUpdateParam{
		AppID:       req.GetAppId(),
		ItemID:      req.GetItemId(),
		DatastoreID: req.GetDatastoreId(),
		ItemMap:     items,
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
		Owners:      req.GetOwners(),
		Lang:        req.GetLangCd(),
		Domain:      req.GetDomain(),
	}

	err := model.ContractExpire(req.GetDatabase(), req.GetWriter(), &params)
	if err != nil {
		utils.ErrorLog(ActionContractExpire, err.Error())
		return err
	}

	utils.InfoLog(ActionContractExpire, utils.MsgProcessEnded)
	return nil
}
