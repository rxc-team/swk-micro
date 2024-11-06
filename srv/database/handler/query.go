package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/query"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Query 快捷方式
type Query struct{}

// log出力使用
const (
	QueryProcessName          = "Query"
	ActionFindQueries         = "FindQueries"
	ActionFindQuery           = "FindQuery"
	ActionAddQuery            = "AddQuery"
	ActionModifyQuery         = "ModifyQuery"
	ActionDeleteQuery         = "DeleteQuery"
	ActionDeleteSelectQueries = "DeleteSelectQueries"
	ActionHardDeleteQueries   = "HardDeleteQueries"
)

// FindQueries 获取所有快捷方式
func (r *Query) FindQueries(ctx context.Context, req *query.FindQueriesRequest, rsp *query.FindQueriesResponse) error {
	utils.InfoLog(ActionFindQueries, utils.MsgProcessStarted)

	Queries, err := model.FindQueries(req.GetDatabase(), req.GetUserId(), req.GetAppId(), req.GetDatastoreId(), req.GetQueryName())
	if err != nil {
		utils.ErrorLog(ActionFindQueries, err.Error())
		return err
	}

	for _, r := range Queries {
		rsp.QueryList = append(rsp.QueryList, r.ToProto())
	}

	utils.InfoLog(ActionFindQueries, utils.MsgProcessEnded)
	return nil
}

// FindQuery 通过QueryID获取快捷方式
func (r *Query) FindQuery(ctx context.Context, req *query.FindQueryRequest, rsp *query.FindQueryResponse) error {
	utils.InfoLog(ActionFindQuery, utils.MsgProcessStarted)

	res, err := model.FindQuery(req.GetDatabase(), req.GetQueryId())
	if err != nil {
		utils.ErrorLog(ActionFindQuery, err.Error())
		return err
	}

	rsp.Query = res.ToProto()

	utils.InfoLog(ActionFindQuery, utils.MsgProcessEnded)
	return nil
}

// AddQuery 添加快捷方式
func (r *Query) AddQuery(ctx context.Context, req *query.AddRequest, rsp *query.AddResponse) error {
	utils.InfoLog(ActionAddQuery, utils.MsgProcessStarted)

	params := toModel(req)

	id, err := model.AddQuery(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionAddQuery, err.Error())
		return err
	}

	rsp.QueryId = id

	utils.InfoLog(ActionAddQuery, utils.MsgProcessEnded)
	return nil
}

// DeleteQuery 删除单个快捷方式
func (r *Query) DeleteQuery(ctx context.Context, req *query.DeleteRequest, rsp *query.DeleteResponse) error {
	utils.InfoLog(ActionDeleteQuery, utils.MsgProcessStarted)

	err := model.DeleteQuery(req.GetDatabase(), req.GetQueryId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteQuery, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteQuery, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectQueries 删除选中的快捷方式
func (r *Query) DeleteSelectQueries(ctx context.Context, req *query.DeleteSelectQueriesRequest, rsp *query.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectQueries, utils.MsgProcessStarted)

	err := model.DeleteSelectQueries(req.GetDatabase(), req.GetQueryIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectQueries, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectQueries, utils.MsgProcessEnded)
	return nil
}

// HardDeleteQueries 物理删除选中的快捷方式
func (r *Query) HardDeleteQueries(ctx context.Context, req *query.HardDeleteQueriesRequest, rsp *query.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteQueries, utils.MsgProcessStarted)

	err := model.HardDeleteQueries(req.GetDatabase(), req.GetQueryIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteQueries, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteQueries, utils.MsgProcessEnded)
	return nil
}

// toModel 转换为model数据
func toModel(q *query.AddRequest) *model.Query {
	var conditions []*model.Condition

	for _, ch := range q.Conditions {
		conditions = append(conditions, &model.Condition{
			FieldID:       ch.FieldId,
			FieldType:     ch.FieldType,
			SearchValue:   ch.SearchValue,
			Operator:      ch.Operator,
			IsDynamic:     ch.IsDynamic,
			ConditionType: ch.ConditionType,
		})
	}

	return &model.Query{
		UserID:        q.UserId,
		DatastoreID:   q.DatastoreId,
		AppID:         q.AppId,
		QueryName:     q.QueryName,
		Description:   q.Description,
		Conditions:    conditions,
		ConditionType: q.ConditionType,
		Fields:        q.Fields,
		CreatedAt:     time.Now(),
		CreatedBy:     q.Writer,
		UpdatedAt:     time.Now(),
		UpdatedBy:     q.Writer,
	}
}
