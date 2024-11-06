package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/workflow/model"
	"rxcsoft.cn/pit3/srv/workflow/proto/relation"
	"rxcsoft.cn/pit3/srv/workflow/utils"
)

// Relation 流程定义
type Relation struct{}

// log出力使用
const (
	RelationProcessName = "Relation"

	ActionFindRelations  = "FindRelations"
	ActionAddRelation    = "AddRelation"
	ActionDeleteRelation = "DeleteRelation"
)

// FindRelations 通过JobID获取流程定义
func (f *Relation) FindRelations(ctx context.Context, req *relation.RelationsRequest, rsp *relation.RelationsResponse) error {
	utils.InfoLog(ActionFindRelations, utils.MsgProcessStarted)

	response, err := model.FindRelations(req.GetDatabase(), req.GetAppId(), req.GetObjectId(), req.GetGroupId(), req.GetWorkflowId(), req.GetAction())
	if err != nil {
		utils.ErrorLog(ActionFindRelations, err.Error())
		return err
	}

	res := &relation.RelationsResponse{}
	for _, f := range response {
		res.Relations = append(res.Relations, f.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindRelations, utils.MsgProcessEnded)
	return nil
}

// AddRelation 添加流程定义
func (f *Relation) AddRelation(ctx context.Context, req *relation.AddRequest, rsp *relation.AddResponse) error {
	utils.InfoLog(ActionAddRelation, utils.MsgProcessStarted)

	param := model.Relation{
		AppId:      req.GetAppId(),
		ObjectId:   req.GetObjectId(),
		GroupId:    req.GetGroupId(),
		WorkflowId: req.GetWorkflowId(),
		Action:     req.GetAction(),
	}

	err := model.AddRelation(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionAddRelation, err.Error())
		return err
	}

	utils.InfoLog(ActionAddRelation, utils.MsgProcessEnded)

	return nil
}

// DeleteRelation 删除流程定义
func (f *Relation) DeleteRelation(ctx context.Context, req *relation.DeleteRequest, rsp *relation.DeleteResponse) error {
	utils.InfoLog(ActionDeleteRelation, utils.MsgProcessStarted)

	err := model.DeleteRelation(req.GetDatabase(), req.GetAppId(), req.GetObjectId(), req.GetGroupId(), req.GetWorkflowId())
	if err != nil {
		utils.ErrorLog(ActionDeleteRelation, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteRelation, utils.MsgProcessEnded)
	return nil
}
