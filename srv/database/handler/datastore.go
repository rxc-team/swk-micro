package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Datastore 台账
type Datastore struct{}

// log出力使用
const (
	DatastoreProcessName          = "Datastore"
	ActionFindDatastores          = "FindDatastores"
	ActionFindDatastore           = "FindDatastore"
	ActionFindDatastoreMapping    = "FindDatastoreMapping"
	ActionAddDatastore            = "AddDatastore"
	ActionAddDatastoreMapping     = "AddDatastoreMapping"
	ActionAddUniqueKey            = "AddUniqueKey"
	ActionAddRelation             = "AddRelation"
	ActionModifyDatastore         = "ModifyDatastore"
	ActionModifyDatastoreMenuSort = "ModifyDatastoreMenuSort"
	ActionModifyDatastoreMapping  = "ModifyDatastoreMapping"
	ActionDeleteDatastore         = "DeleteDatastore"
	ActionDeleteDatastoreMapping  = "DeleteDatastoreMapping"
	ActionDeleteUniqueKey         = "DeleteUniqueKey"
	ActionDeleteRelation          = "DeleteRelation"
	ActionDeleteSelectDatastores  = "DeleteSelectDatastores"
	ActionHardDeleteDatastores    = "HardDeleteDatastores"
)

// FindDatastores 获取app下所有台账
func (d *Datastore) FindDatastores(ctx context.Context, req *datastore.DatastoresRequest, rsp *datastore.DatastoresResponse) error {
	utils.InfoLog(ActionFindDatastores, utils.MsgProcessStarted)

	datastores, err := model.FindDatastores(req.GetDatabase(), req.GetAppId(), req.GetDatastoreName(), req.GetCanCheck(), req.GetShowInMenu())
	if err != nil {
		utils.ErrorLog(ActionFindDatastores, err.Error())
		return err
	}

	res := &datastore.DatastoresResponse{}
	for _, ds := range datastores {
		res.Datastores = append(res.Datastores, ds.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindDatastores, utils.MsgProcessEnded)
	return nil
}

// FindDatastore 通过ID获取台账信息
func (d *Datastore) FindDatastore(ctx context.Context, req *datastore.DatastoreRequest, rsp *datastore.DatastoreResponse) error {
	utils.InfoLog(ActionFindDatastore, utils.MsgProcessStarted)

	res, err := model.FindDatastore(req.GetDatabase(), req.GetDatastoreId())
	if err != nil {
		utils.ErrorLog(ActionFindDatastore, err.Error())
		return err
	}

	rsp.Datastore = res.ToProto()

	utils.InfoLog(ActionFindDatastore, utils.MsgProcessEnded)
	return nil
}

// FindDatastoreByKey 通过apiKey获取Datastore信息
func (d *Datastore) FindDatastoreByKey(ctx context.Context, req *datastore.DatastoreKeyRequest, rsp *datastore.DatastoreResponse) error {
	utils.InfoLog(ActionFindDatastore, utils.MsgProcessStarted)

	res, err := model.FindDatastoreByKey(req.GetDatabase(), req.GetAppId(), req.GetApiKey())
	if err != nil {
		utils.ErrorLog(ActionFindDatastore, err.Error())
		return err
	}

	rsp.Datastore = res.ToProto()

	utils.InfoLog(ActionFindDatastore, utils.MsgProcessEnded)
	return nil
}

// FindDatastoreMapping 通过ID获取台账映射信息
func (d *Datastore) FindDatastoreMapping(ctx context.Context, req *datastore.MappingRequest, rsp *datastore.MappingResponse) error {
	utils.InfoLog(ActionFindDatastoreMapping, utils.MsgProcessStarted)

	res, err := model.FindDatastoreMapping(req.GetDatabase(), req.GetDatastoreId(), req.GetMappingId())
	if err != nil {
		utils.ErrorLog(ActionFindDatastoreMapping, err.Error())
		return err
	}

	rsp.Mapping = res.ToProto()

	utils.InfoLog(ActionFindDatastoreMapping, utils.MsgProcessEnded)
	return nil
}

// AddDatastore 添加台账
func (d *Datastore) AddDatastore(ctx context.Context, req *datastore.AddRequest, rsp *datastore.AddResponse) error {
	utils.InfoLog(ActionAddDatastore, utils.MsgProcessStarted)

	var sorts []*model.SortItem
	for _, s := range req.GetSorts() {
		sort := &model.SortItem{
			SortKey:   s.SortKey,
			SortValue: s.SortValue,
		}
		sorts = append(sorts, sort)
	}

	params := model.Datastore{
		AppID:               req.GetAppId(),
		DatastoreName:       req.GetDatastoreName(),
		ApiKey:              req.GetApiKey(),
		CanCheck:            req.GetCanCheck(),
		ShowInMenu:          req.GetShowInMenu(),
		NoStatus:            req.GetNoStatus(),
		Encoding:            req.GetEncoding(),
		Sorts:               sorts,
		ScanFields:          req.GetScanFields(),
		ScanFieldsConnector: req.GetScanFieldsConnector(),
		PrintField1:         req.GetPrintField1(),
		PrintField2:         req.GetPrintField2(),
		PrintField3:         req.GetPrintField3(),
		DisplayOrder:        req.GetDisplayOrder(),
		UniqueFields:        req.GetUniqueFields(),
		CreatedAt:           time.Now(),
		CreatedBy:           req.GetWriter(),
		UpdatedAt:           time.Now(),
		UpdatedBy:           req.GetWriter(),
	}

	id, err := model.AddDatastore(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddDatastore, err.Error())
		return err
	}

	rsp.DatastoreId = id

	utils.InfoLog(ActionAddDatastore, utils.MsgProcessEnded)

	return nil
}

// AddDatastoreMapping 添加台账映射
func (d *Datastore) AddDatastoreMapping(ctx context.Context, req *datastore.AddMappingRequest, rsp *datastore.AddMappingResponse) error {
	utils.InfoLog(ActionAddDatastoreMapping, utils.MsgProcessStarted)

	var rules []*model.MappingRule
	for _, r := range req.GetMappingRule() {
		rule := &model.MappingRule{
			FromKey:      r.FromKey,
			ToKey:        r.ToKey,
			IsRequired:   r.IsRequired,
			Exist:        r.Exist,
			Special:      r.Special,
			Format:       r.Format,
			Replace:      r.Replace,
			DataType:     r.DataType,
			CheckChange:  r.CheckChange,
			DefaultValue: r.DefaultValue,
			PrimaryKey:   r.PrimaryKey,
			Precision:    r.Precision,
			ShowOrder:    r.ShowOrder,
		}
		rules = append(rules, rule)
	}

	params := model.MappingConf{
		MappingName:   req.MappingName,
		MappingType:   req.MappingType,
		UpdateType:    req.UpdateType,
		ApplyType:     req.ApplyType,
		SeparatorChar: req.SeparatorChar,
		BreakChar:     req.BreakChar,
		LineBreakCode: req.LineBreakCode,
		CharEncoding:  req.CharEncoding,
		MappingRule:   rules,
	}

	id, err := model.AddMapping(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), params)
	if err != nil {
		utils.ErrorLog(ActionAddDatastoreMapping, err.Error())
		return err
	}

	rsp.MappingId = id

	utils.InfoLog(ActionAddDatastoreMapping, utils.MsgProcessEnded)

	return nil
}

// AddUniqueKey 添加台账唯一组合索引
func (d *Datastore) AddUniqueKey(ctx context.Context, req *datastore.AddUniqueRequest, rsp *datastore.AddUniqueResponse) error {
	utils.InfoLog(ActionAddUniqueKey, utils.MsgProcessStarted)

	err := model.AddUniqueKey(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), req.GetUniqueFields())
	if err != nil {
		utils.ErrorLog(ActionAddUniqueKey, err.Error())
		return err
	}

	utils.InfoLog(ActionAddUniqueKey, utils.MsgProcessEnded)

	return nil
}

// AddUniqueKey 添加台账关系
func (d *Datastore) AddRelation(ctx context.Context, req *datastore.AddRelationRequest, rsp *datastore.AddRelationResponse) error {
	utils.InfoLog(ActionAddRelation, utils.MsgProcessStarted)

	relation := model.RelationItem{
		RelationId:  req.GetRelation().GetRelationId(),
		DatastoreId: req.GetRelation().GetDatastoreId(),
		Fields:      req.GetRelation().GetFields(),
	}

	err := model.AddRelation(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), relation)
	if err != nil {
		utils.ErrorLog(ActionAddRelation, err.Error())
		return err
	}

	utils.InfoLog(ActionAddRelation, utils.MsgProcessEnded)

	return nil
}

// ModifyDatastore 更新台账的字段
func (d *Datastore) ModifyDatastore(ctx context.Context, req *datastore.ModifyRequest, rsp *datastore.ModifyResponse) error {
	utils.InfoLog(ActionModifyDatastore, utils.MsgProcessStarted)
	var sorts []*model.SortItem
	for _, s := range req.GetSorts() {
		sort := &model.SortItem{
			SortKey:   s.SortKey,
			SortValue: s.SortValue,
		}
		sorts = append(sorts, sort)
	}

	params := model.ModifyReq{
		DatastoreID:         req.GetDatastoreId(),
		DatastoreName:       req.GetDatastoreName(),
		ApiKey:              req.GetApiKey(),
		CanCheck:            req.GetCanCheck(),
		Sorts:               sorts,
		ShowInMenu:          req.GetShowInMenu(),
		NoStatus:            req.GetNoStatus(),
		Encoding:            req.GetEncoding(),
		ScanFields:          req.GetScanFields(),
		ScanFieldsConnector: req.GetScanFieldsConnector(),
		PrintField1:         req.GetPrintField1(),
		PrintField2:         req.GetPrintField2(),
		PrintField3:         req.GetPrintField3(),
		Writer:              req.GetWriter(),
	}

	err := model.ModifyDatastore(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyDatastore, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyDatastore, utils.MsgProcessEnded)
	return nil
}

// ModifyDatastoreMenuSort 更新台账的字段
func (d *Datastore) ModifyDatastoreMenuSort(ctx context.Context, req *datastore.MenuSortRequest, rsp *datastore.MenuSortResponse) error {
	utils.InfoLog(ActionModifyDatastoreMenuSort, utils.MsgProcessStarted)

	for _, sort := range req.DatastoresSort {
		params := model.ModifyReq{
			DatastoreID:  sort.GetDatastoreId(),
			ApiKey:       sort.GetApiKey(),
			DisplayOrder: sort.GetDisplayOrder(),
		}

		err := model.ModifyDatastore(req.GetDb(), &params)
		if err != nil {
			utils.ErrorLog(ActionModifyDatastoreMenuSort, err.Error())
			return err
		}
	}

	utils.InfoLog(ActionModifyDatastoreMenuSort, utils.MsgProcessEnded)
	return nil
}

// ModifyDatastoreMapping 修改台账映射
func (d *Datastore) ModifyDatastoreMapping(ctx context.Context, req *datastore.ModifyMappingRequest, rsp *datastore.ModifyMappingResponse) error {
	utils.InfoLog(ActionModifyDatastoreMapping, utils.MsgProcessStarted)

	var rules []*model.MappingRule
	for _, r := range req.GetMappingRule() {
		rule := &model.MappingRule{
			FromKey:      r.FromKey,
			ToKey:        r.ToKey,
			IsRequired:   r.IsRequired,
			Exist:        r.Exist,
			Special:      r.Special,
			Format:       r.Format,
			Replace:      r.Replace,
			DataType:     r.DataType,
			CheckChange:  r.CheckChange,
			DefaultValue: r.DefaultValue,
			PrimaryKey:   r.PrimaryKey,
			Precision:    r.Precision,
			ShowOrder:    r.ShowOrder,
		}
		if req.GetApplyType() == "datastore" {
			rule.CheckChange = false
		}
		rules = append(rules, rule)
	}

	params := model.MappingConf{
		MappingID:     req.MappingId,
		MappingType:   req.MappingType,
		UpdateType:    req.UpdateType,
		MappingName:   req.MappingName,
		ApplyType:     req.ApplyType,
		SeparatorChar: req.SeparatorChar,
		BreakChar:     req.BreakChar,
		LineBreakCode: req.LineBreakCode,
		CharEncoding:  req.CharEncoding,
		MappingRule:   rules,
	}

	err := model.ModifyMapping(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), params)
	if err != nil {
		utils.ErrorLog(ActionModifyDatastoreMapping, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyDatastoreMapping, utils.MsgProcessEnded)

	return nil
}

// DeleteDatastore 删除单个台账
func (d *Datastore) DeleteDatastore(ctx context.Context, req *datastore.DeleteRequest, rsp *datastore.DeleteResponse) error {
	utils.InfoLog(ActionDeleteDatastore, utils.MsgProcessStarted)

	err := model.DeleteDatastore(req.GetDatabase(), req.GetDatastoreId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteDatastore, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteDatastore, utils.MsgProcessEnded)
	return nil
}

// DeleteDatastoreMapping 删除台账映射
func (d *Datastore) DeleteDatastoreMapping(ctx context.Context, req *datastore.DeleteMappingRequest, rsp *datastore.DeleteResponse) error {
	utils.InfoLog(ActionDeleteDatastoreMapping, utils.MsgProcessStarted)

	err := model.DeleteMapping(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), req.GetMappingId())
	if err != nil {
		utils.ErrorLog(ActionDeleteDatastoreMapping, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteDatastoreMapping, utils.MsgProcessEnded)

	return nil
}

// DeleteUniqueKey 删除台账唯一组合索引
func (d *Datastore) DeleteUniqueKey(ctx context.Context, req *datastore.DeleteUniqueRequest, rsp *datastore.DeleteUniqueResponse) error {
	utils.InfoLog(ActionDeleteUniqueKey, utils.MsgProcessStarted)

	err := model.DeleteUniqueKey(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), req.GetUniqueFields())
	if err != nil {
		utils.ErrorLog(ActionDeleteUniqueKey, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteUniqueKey, utils.MsgProcessEnded)

	return nil
}

// DeleteRelation 删除台账关系
func (d *Datastore) DeleteRelation(ctx context.Context, req *datastore.DeleteRelationRequest, rsp *datastore.DeleteRelationResponse) error {
	utils.InfoLog(ActionDeleteRelation, utils.MsgProcessStarted)

	err := model.DeleteRelation(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), req.GetRelationId())
	if err != nil {
		utils.ErrorLog(ActionDeleteRelation, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteRelation, utils.MsgProcessEnded)

	return nil
}

// DeleteSelectDatastores 删除多个台账
func (d *Datastore) DeleteSelectDatastores(ctx context.Context, req *datastore.DeleteSelectRequest, rsp *datastore.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectDatastores, utils.MsgProcessStarted)
	err := model.DeleteSelectDatastores(req.GetDatabase(), req.GetDatastoreIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectDatastores, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectDatastores, utils.MsgProcessEnded)
	return nil
}

// HardDeleteDatastores 物理删除多个台账
func (d *Datastore) HardDeleteDatastores(ctx context.Context, req *datastore.HardDeleteDatastoresRequest, rsp *datastore.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteDatastores, utils.MsgProcessStarted)

	err := model.HardDeleteDatastores(req.GetDatabase(), req.GetDatastoreIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteDatastores, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteDatastores, utils.MsgProcessEnded)
	return nil
}
