package handler

import (
	"context"
	"errors"
	"time"

	"github.com/micro/go-micro/v2/client"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
)

// Item 台账数据
type Item struct{}

// log出力使用
const (
	ItemProcessName             = "Item"
	ActionFindItems             = "FindItems"
	ActionFindItem              = "FindItem"
	ActionFindRishiritsu        = "FindRishiritsu"
	ActionFindKaraCount         = "FindKaraCount"
	ActionAddItem               = "AddItem"
	ActionMutilAddItem          = "MutilAddItem"
	ActionImportItem            = "ImportItem"
	ActionImportCheckItem       = "ImportCheckItem"
	ActionMappingImport         = "MappingImport"
	ActionModifyItem            = "ModifyItem"
	ActionDeleteItem            = "DeleteItem"
	ActionDeleteDatastoreItems  = "DeleteDatastoreItems"
	ActionDeleteItems           = "DeleteItems"
	ActionChangeOwners          = "ChangeOwners"
	ActionChangeItemOwner       = "ChangeItemOwner"
	ActionChangeStatus          = "ChangeStatus"
	ActionConfimItem            = "ConfimItem"
	ActionGenerateItem          = "GenerateItem"
	ActionGenerateShoukyakuItem = "GenerateShoukyakuItem"
	ActionSwkDownload           = "SwkDownload"
)

// FindItems 获取台账下的所有数据
func (i *Item) FindItems(ctx context.Context, req *item.ItemsRequest, rsp *item.ItemsResponse) error {
	utils.InfoLog(ActionFindItems, utils.MsgProcessStarted)

	var conditions []*model.Condition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.Condition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}
	var sorts []*model.SortItem
	for _, sort := range req.GetSorts() {
		sorts = append(sorts, &model.SortItem{
			SortKey:   sort.GetSortKey(),
			SortValue: sort.GetSortValue(),
		})
	}

	params := model.ItemsParam{
		AppID:         req.GetAppId(),
		DatastoreID:   req.GetDatastoreId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		PageIndex:     req.GetPageIndex(),
		PageSize:      req.GetPageSize(),
		Sorts:         sorts,
		IsOrigin:      req.GetIsOrigin(),
		ShowLookup:    req.GetShowLookup(),
		Owners:        req.GetOwners(),
	}

	result, err := model.FindItems(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindItems, err.Error())
		return err
	}

	res := &item.ItemsResponse{}
	for _, it := range result.Docs {
		res.Items = append(res.Items, it.ToProto())
	}

	res.Total = result.Total

	*rsp = *res

	utils.InfoLog(ActionFindItems, utils.MsgProcessEnded)
	return nil
}

// FindItems 获取台账下的所有数据
func (i *Item) Download(ctx context.Context, req *item.DownloadRequest, stream item.ItemService_DownloadStream) error {
	utils.InfoLog(ActionFindItems, utils.MsgProcessStarted)

	var conditions []*model.Condition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.Condition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}
	var sorts []*model.SortItem
	for _, sort := range req.GetSorts() {
		sorts = append(sorts, &model.SortItem{
			SortKey:   sort.GetSortKey(),
			SortValue: sort.GetSortValue(),
		})
	}

	params := model.ItemsParam{
		AppID:         req.GetAppId(),
		DatastoreID:   req.GetDatastoreId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		Sorts:         sorts,
		Owners:        req.GetOwners(),
	}

	err := model.DownloadItems(req.GetDatabase(), params, stream)
	if err != nil {
		utils.ErrorLog(ActionFindItems, err.Error())
		return err
	}

	utils.InfoLog(ActionFindItems, utils.MsgProcessEnded)

	return nil
}

// FindAndModifyFile 获取台账并更新file数据
func (i *Item) FindAndModifyFile(ctx context.Context, req *item.FindRequest, stream item.ItemService_FindAndModifyFileStream) error {
	utils.InfoLog(ActionFindItems, utils.MsgProcessStarted)
	params := model.ItemsParam{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
	}

	err := model.FindAndModifyFile(req.GetDatabase(), params, stream)
	if err != nil {
		utils.ErrorLog("FindAndModifyFile", req.AppId+"_"+req.DatastoreId+","+err.Error())
		return err
	}

	return nil
}

// FindCount 获取台账数据的件数
func (i *Item) FindCount(ctx context.Context, req *item.CountRequest, rsp *item.CountResponse) error {
	utils.InfoLog(ActionFindItems, utils.MsgProcessStarted)

	var conditions []*model.Condition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.Condition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}

	params := model.CountParam{
		AppID:         req.GetAppId(),
		DatastoreID:   req.GetDatastoreId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		Owners:        req.GetOwners(),
	}

	total, err := model.FindCount(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindItems, err.Error())
		return err
	}

	rsp.Total = total

	utils.InfoLog(ActionFindItems, utils.MsgProcessEnded)
	return nil
}

// FindKaraCount 获取台账唯一字段空值总件数
func (i *Item) FindKaraCount(ctx context.Context, req *item.KaraCountRequest, rsp *item.KaraCountResponse) error {
	utils.InfoLog(ActionFindKaraCount, utils.MsgProcessStarted)

	params := model.KaraCountParam{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		FieldID:     req.GetFieldId(),
		FieldType:   req.GetFieldType(),
		Owners:      req.GetOwners(),
	}

	total, err := model.FindKaraCount(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindKaraCount, err.Error())
		return err
	}

	rsp.Total = total

	utils.InfoLog(ActionFindKaraCount, utils.MsgProcessEnded)
	return nil
}

// FindItem 通过ID获取数据
func (i *Item) FindItem(ctx context.Context, req *item.ItemRequest, rsp *item.ItemResponse) error {
	utils.InfoLog(ActionFindItem, utils.MsgProcessStarted)

	param := model.ItemParam{
		ItemID:      req.GetItemId(),
		DatastoreID: req.GetDatastoreId(),
		IsOrigin:    req.GetIsOrigin(),
		Owners:      req.GetOwners(),
	}

	res, err := model.FindItem(req.GetDatabase(), &param)
	if err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			return errors.New("データが存在しないか、データの権限がありません")
		}
		utils.ErrorLog(ActionFindItem, err.Error())
		return err
	}

	rsp.Item = res.ToProto()

	utils.InfoLog(ActionFindItem, utils.MsgProcessEnded)
	return nil
}

// FindRishiritsu 通过租赁日期，租赁期间获取数据
func (i *Item) FindRishiritsu(ctx context.Context, req *item.RishiritsuRequest, rsp *item.RishiritsuResponse) error {
	utils.InfoLog(ActionFindRishiritsu, utils.MsgProcessStarted)

	param := model.RishiritsuParam{
		DatastoreID: req.GetDatastoreId(),
		LeaseStymd:  req.GetLeasestymd(),
		LeaseKikan:  req.GetLeasekikan(),
	}

	res, err := model.FindRishiritsu(req.GetDatabase(), &param)
	if err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			return errors.New("利子率マスタから利子率を取得できませんでした、追加借入利子率を手入力してください")
		}
		utils.ErrorLog(ActionFindRishiritsu, err.Error())
		return err
	}

	rsp.Item = res.ToProto()

	utils.InfoLog(ActionFindRishiritsu, utils.MsgProcessEnded)
	return nil
}

// AddItem 添加台账数据
func (i *Item) AddItem(ctx context.Context, req *item.AddRequest, rsp *item.AddResponse) error {
	utils.InfoLog(ActionAddItem, utils.MsgProcessStarted)

	items := make(map[string]*model.Value, len(req.Items))
	for key, item := range req.Items {
		items[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetValueFromProto(item),
		}
	}

	params := model.Item{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		ItemMap:     items,
		Owners:      req.GetOwners(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	id, err := model.AddItem(req.GetDatabase(), req.GetWriter(), req.GetLangCd(), req.GetDomain(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddItem, err.Error())
		return err
	}

	rsp.ItemId = id

	utils.InfoLog(ActionAddItem, utils.MsgProcessEnded)

	return nil
}

// MappingUpload 批量导入更新台账数据
func (i *Item) MappingUpload(ctx context.Context, stream item.ItemService_MappingUploadStream) error {
	utils.InfoLog(ActionMappingImport, utils.MsgProcessStarted)

	err := model.MappingUpload(ctx, stream)
	if err != nil {
		utils.ErrorLog(ActionMappingImport, err.Error())
		return err
	}

	utils.InfoLog(ActionMappingImport, utils.MsgProcessEnded)

	return nil
}

// ImportItem1 批量导入台账数据
func (i *Item) ImportItem(ctx context.Context, stream item.ItemService_ImportItemStream) error {
	utils.InfoLog(ActionImportItem, utils.MsgProcessStarted)

	err := model.ImportItem(ctx, stream)
	if err != nil {
		utils.ErrorLog(ActionImportItem, err.Error())
		return err
	}

	utils.InfoLog(ActionImportItem, utils.MsgProcessEnded)

	return nil
}

// ImportCheckItem 批量盘点
func (i *Item) ImportCheckItem(ctx context.Context, stream item.ItemService_ImportCheckItemStream) error {
	utils.InfoLog(ActionImportCheckItem, utils.MsgProcessStarted)

	err := model.ImportCheckItem(ctx, stream)
	if err != nil {
		utils.ErrorLog(ActionImportCheckItem, err.Error())
		return err
	}

	utils.InfoLog(ActionImportCheckItem, utils.MsgProcessEnded)

	return nil
}

// ModifyItem 更新台账一条数据
func (i *Item) ModifyItem(ctx context.Context, req *item.ModifyRequest, rsp *item.ModifyResponse) error {
	utils.InfoLog(ActionModifyItem, utils.MsgProcessStarted)

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

	err := model.ModifyItem(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyItem, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyItem, utils.MsgProcessEnded)
	return nil
}

// GenerateItem 更新台账数据作成
func (i *Item) GenerateItem(ctx context.Context, req *item.JournalRequest, rsp *item.JournalResponse) error {
	utils.InfoLog(ActionGenerateItem, utils.MsgProcessStarted)

	err := model.GenerateItem(req.GetDatabase(), req.GetDatastoreId(), req.GetStartDate(), req.GetLastDate())
	if err != nil {
		utils.ErrorLog(ActionGenerateItem, err.Error())
		return err
	}

	utils.InfoLog(ActionGenerateItem, utils.MsgProcessEnded)
	return nil
}

// GenerateShoukyakuItem 更新偿还台账数据作成
func (i *Item) GenerateShoukyakuItem(ctx context.Context, req *item.JournalRequest, rsp *item.JournalResponse) error {
	utils.InfoLog(ActionGenerateShoukyakuItem, utils.MsgProcessStarted)

	err := model.GenerateShoukyakuItem(req.GetDatabase(), req.GetDatastoreId(), req.GetStartDate(), req.GetLastDate())
	if err != nil {
		utils.ErrorLog(ActionGenerateShoukyakuItem, err.Error())
		return err
	}

	utils.InfoLog(ActionGenerateShoukyakuItem, utils.MsgProcessEnded)
	return nil
}

// ConfimItem 更新台账数据确认
func (i *Item) ConfimItem(ctx context.Context, req *item.JournalRequest, rsp *item.JournalResponse) error {
	utils.InfoLog(ActionConfimItem, utils.MsgProcessStarted)

	err := model.ConfimItem(req.GetDatabase(), req.GetDatastoreId(), req.GetStartDate(), req.GetLastDate())
	if err != nil {
		utils.ErrorLog(ActionConfimItem, err.Error())
		return err
	}

	utils.InfoLog(ActionConfimItem, utils.MsgProcessEnded)
	return nil
}

// DeleteItem 删除单个台账数据
func (i *Item) DeleteItem(ctx context.Context, req *item.DeleteRequest, rsp *item.DeleteResponse) error {
	utils.InfoLog(ActionDeleteItem, utils.MsgProcessStarted)

	err := model.DeleteItem(req.GetDatabase(), req.GetDatastoreId(), req.GetItemId(), req.GetWriter(), req.GetLangCd(), req.GetDomain(), req.GetOwners())
	if err != nil {
		utils.ErrorLog(ActionDeleteItem, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteItem, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectItems 删除选中的台账数据
func (i *Item) DeleteSelectItems(ctx context.Context, req *item.SelectedItemsRequest, stream item.ItemService_DeleteSelectItemsStream) error {
	utils.InfoLog(ActionDeleteItem, utils.MsgProcessStarted)

	err := model.DeleteSelectItems(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId(), req.GetItemIdList(), stream)
	if err != nil {
		utils.ErrorLog(ActionDeleteItem, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteItem, utils.MsgProcessEnded)
	return nil
}

// DeleteDatastoreItems 删除台账所有数据
func (i *Item) DeleteDatastoreItems(ctx context.Context, req *item.DeleteDatastoreItemsRequest, rsp *item.DeleteResponse) error {
	utils.InfoLog(ActionDeleteDatastoreItems, utils.MsgProcessStarted)

	err := model.DeleteDatastoreItems(req.GetDatabase(), req.GetDatastoreId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteDatastoreItems, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteDatastoreItems, utils.MsgProcessEnded)
	return nil
}

// DeleteItems 删除多条数据记录
func (i *Item) DeleteItems(ctx context.Context, req *item.DeleteItemsRequest, rsp *item.DeleteResponse) error {
	utils.InfoLog(ActionDeleteItems, utils.MsgProcessStarted)

	var conditions []*model.Condition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.Condition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}

	params := model.DeleteItemsParam{
		AppID:         req.GetAppId(),
		DatastoreID:   req.GetDatastoreId(),
		ConditionList: conditions,
		ConditionType: req.GetConditionType(),
		UserID:        req.GetUserId(),
	}

	err := model.DeleteItems(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionDeleteItems, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteItems, utils.MsgProcessEnded)
	return nil
}

// ChangeOwners 更新所有者
func (i *Item) ChangeOwners(ctx context.Context, req *item.OwnersRequest, rsp *item.OwnersResponse) error {
	utils.InfoLog(ActionChangeOwners, utils.MsgProcessStarted)

	err := model.ChangeOwners(req.GetDatabase(), req)
	if err != nil {
		utils.ErrorLog(ActionChangeOwners, err.Error())
		return err
	}

	utils.InfoLog(ActionChangeOwners, utils.MsgProcessEnded)
	return nil
}

// ChangeSelectOwners 通过检索条件更新所有者信息
func (i *Item) ChangeSelectOwners(ctx context.Context, req *item.SelectOwnersRequest, rsp *item.SelectOwnersResponse) error {
	utils.InfoLog(ActionChangeOwners, utils.MsgProcessStarted)

	var conditions []*model.Condition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.Condition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}

	params := model.OwnersParam{
		AppID:         req.GetAppId(),
		DatastoreID:   req.GetDatastoreId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		Owner:         req.GetOwner(),
		Writer:        req.GetWriter(),
		OldOwners:     req.GetOldOwners(),
	}

	err := model.ChangeSelectOwners(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionChangeOwners, err.Error())
		return err
	}

	utils.InfoLog(ActionChangeOwners, utils.MsgProcessEnded)
	return nil
}

// ChangeItemOwner 更新单条记录所属组织
func (i *Item) ChangeItemOwner(ctx context.Context, req *item.ItemOwnerRequest, rsp *item.ItemOwnerResponse) error {
	utils.InfoLog(ActionChangeItemOwner, utils.MsgProcessStarted)

	params := model.OwnerParam{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		ItemID:      req.GetItemId(),
		Owner:       req.GetOwner(),
		Writer:      req.GetWriter(),
	}

	err := model.ChangeItemOwner(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionChangeItemOwner, err.Error())
		return err
	}

	utils.InfoLog(ActionChangeItemOwner, utils.MsgProcessEnded)
	return nil
}

// ChangeStatus 更新状态
func (i *Item) ChangeStatus(ctx context.Context, req *item.StatusRequest, rsp *item.StatusResponse) error {
	utils.InfoLog(ActionChangeStatus, utils.MsgProcessStarted)

	err := model.ChangeStatus(req.GetDatabase(), req)
	if err != nil {
		utils.ErrorLog(ActionChangeStatus, err.Error())
		return err
	}

	utils.InfoLog(ActionChangeStatus, utils.MsgProcessEnded)
	return nil
}

// FindItems 获取台账下的所有数据
func (i *Item) SwkDownload(ctx context.Context, req *item.DownloadRequest, stream item.ItemService_DownloadStream) error {
	utils.InfoLog(ActionSwkDownload, utils.MsgProcessStarted)

	var conditions []*model.Condition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.Condition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}
	var sorts []*model.SortItem
	for _, sort := range req.GetSorts() {
		sorts = append(sorts, &model.SortItem{
			SortKey:   sort.GetSortKey(),
			SortValue: sort.GetSortValue(),
		})
	}

	params := model.ItemsParam{
		AppID:         req.GetAppId(),
		DatastoreID:   req.GetDatastoreId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		Sorts:         sorts,
		Owners:        req.GetOwners(),
	}

	journalService := journal.NewJournalService("journal", client.DefaultClient)
	var journal journal.FindDownloadSettingRequest

	// 从共通获取
	journal.AppId = req.GetAppId()
	journal.Database = req.GetDatabase()

	downloadInfo, jerr := journalService.FindDownloadSetting(context.TODO(), &journal)
	if jerr != nil {
		utils.ErrorLog(ActionSwkDownload, jerr.Error())
		return jerr
	}

	err := model.SwkDownloadItems(req.GetDatabase(), params, stream, downloadInfo)
	if err != nil {
		utils.ErrorLog(ActionSwkDownload, err.Error())
		return err
	}

	utils.InfoLog(ActionSwkDownload, utils.MsgProcessEnded)

	return nil
}
