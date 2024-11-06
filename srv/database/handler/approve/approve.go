package approve

import (
	"context"
	"time"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/pit3/srv/manage/proto/user"

	"rxcsoft.cn/pit3/srv/database/model"
)

// Approve 临时台账数据
type Approve struct{}

// log出力使用
const (
	ApproveProcessName = "Approve"
	ActionFindItems    = "FindItems"
	ActionFindCount    = "FindCount"
	ActionFindItem     = "FindItem"
	ActionAddItem      = "AddItem"
	ActionDeleteItems  = "DeleteItems"
)

// FindItems 获取审批数据
func (i *Approve) FindItems(ctx context.Context, req *approve.ItemsRequest, rsp *approve.ItemsResponse) error {
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

	params := model.ApproveItemsParam{
		DatastoreID:   req.GetDatastoreId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		PageIndex:     req.GetPageIndex(),
		PageSize:      req.GetPageSize(),
		SearchType:    req.GetSearchType(),
		UserId:        req.GetUserId(),
		Status:        req.GetStatus(),
	}

	items, total, err := model.FindApproveItems(req.GetDatabase(), req.GetWfId(), params)
	if err != nil {
		utils.ErrorLog(ActionFindItems, err.Error())
		return err
	}

	res := &approve.ItemsResponse{}
	for _, it := range items {
		res.Items = append(res.Items, it.ToProto(false))
	}

	res.Total = total

	*rsp = *res

	utils.InfoLog(ActionFindItems, utils.MsgProcessEnded)
	return nil
}

// FindCount 获取审批数据
func (i *Approve) FindCount(ctx context.Context, req *approve.CountRequest, rsp *approve.CountResponse) error {
	utils.InfoLog(ActionFindCount, utils.MsgProcessStarted)

	total, err := model.FindApproveCount(req.GetDatabase(), req.GetWfId(), req.GetStatus())
	if err != nil {
		utils.ErrorLog(ActionFindCount, err.Error())
		return err
	}

	rsp.Total = total

	utils.InfoLog(ActionFindCount, utils.MsgProcessEnded)
	return nil
}

// FindItem 通过ID获取数据
func (i *Approve) FindItem(ctx context.Context, req *approve.ItemRequest, rsp *approve.ItemResponse) error {
	utils.InfoLog(ActionFindItem, utils.MsgProcessStarted)

	res, err := model.FindApproveItem(req.GetDatabase(), req.GetExampleId(), req.GetDatastoreId())
	if err != nil {
		utils.ErrorLog(ActionFindItem, err.Error())
		return err
	}

	rsp.Item = res.ToProto(true)

	utils.InfoLog(ActionFindItem, utils.MsgProcessEnded)
	return nil
}

// AddItem 添加审批数据
func (i *Approve) AddItem(ctx context.Context, req *approve.AddRequest, rsp *approve.AddResponse) error {
	utils.InfoLog(ActionAddItem, utils.MsgProcessStarted)

	// 当前语言取得(选项翻译用)
	lang := utils.GetLanguageData(req.Database, req.LangCd, req.Domain)

	// 选项字段情报取得(选项翻译用)
	fps := model.FindFieldsParam{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		FieldType:   "options",
	}
	opFields, err := model.FindFields(req.GetDatabase(), &fps)
	if err != nil {
		utils.ErrorLog(ActionAddItem, err.Error())
		return err
	}

	// 用户情报取得(用户ID转用户名用)
	userService := user.NewUserService("manage", client.DefaultClient)
	var reqU user.FindUsersRequest
	reqU.InvalidatedIn = "true"
	reqU.Domain = req.Domain
	reqU.Database = req.Database
	response, err := userService.FindUsers(context.TODO(), &reqU)
	if err != nil {
		utils.ErrorLog(ActionAddItem, err.Error())
		return err
	}
	userMap := make(map[string]string)
	for _, u := range response.Users {
		userMap[u.UserId] = u.UserName
	}

	// 审批数据编辑
	items := make(map[string]*model.Value, len(req.GetItems()))
	for key, item := range req.GetItems() {
		items[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetApproveDataValue(item, getField(key, opFields), userMap, lang, false),
		}
	}
	hs := make(map[string]*model.Value, len(req.GetHistory()))
	for key, item := range req.GetHistory() {
		hs[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetApproveDataValue(item, getField(key, opFields), userMap, lang, true),
		}
	}
	current := make(map[string]*model.Value, len(req.GetCurrent()))
	for key, item := range req.GetCurrent() {
		current[key] = &model.Value{
			DataType: item.DataType,
			Value:    model.GetApproveDataValue(item, getField(key, opFields), userMap, lang, true),
		}
	}

	params := model.ApproveItem{
		ItemID:      req.GetItemId(),
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		ItemMap:     items,
		History:     hs,
		Current:     current,
		ExampleID:   req.GetExampleId(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
	}

	id, err := model.AddApprove(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddItem, err.Error())
		return err
	}

	rsp.ItemId = id

	utils.InfoLog(ActionAddItem, utils.MsgProcessEnded)

	return nil
}

// DeleteItems 删除审批数据
func (i *Approve) DeleteItems(ctx context.Context, req *approve.DeleteRequest, rsp *approve.DeleteResponse) error {
	utils.InfoLog(ActionDeleteItems, utils.MsgProcessStarted)

	err := model.DeleteApproveItems(req.GetDatabase(), req.GetItems())
	if err != nil {
		utils.ErrorLog(ActionDeleteItems, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteItems, utils.MsgProcessEnded)
	return nil
}

// getField 通过字段ID获取字段信息
func getField(id string, fs []model.Field) (f model.Field) {
	var res model.Field
	for _, f := range fs {
		if f.FieldID == id {
			return f
		}
	}
	return res
}
