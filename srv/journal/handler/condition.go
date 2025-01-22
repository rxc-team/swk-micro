package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/journal/model"
	"rxcsoft.cn/pit3/srv/journal/proto/condition"
	"rxcsoft.cn/pit3/srv/journal/utils"
)

// Journal 仕訳
type Condition struct{}

// log出力使用
const (
	// JournalProcessName = "Journal"

	ActionAddCondition   = "AddCondition"
	ActionFindConditions = "FindConditions"
)

// 添加分录下载设定
func (f *Condition) AddCondition(ctx context.Context, req *condition.AddConditionRequest, rsp *condition.AddConditionResponse) error {

	var groups []*model.Group
	for _, g := range req.GetGroups() {
		var cons []*model.Con
		for _, con := range g.GetCons() {
			cons = append(cons, &model.Con{
				ConID:       con.ConId,
				ConName:     con.ConName,
				ConField:    con.ConField,
				ConOperator: con.ConOperator,
				ConValue:    con.ConValue,
			})
		}

		groups = append(groups, &model.Group{
			GroupID:    g.GroupId,
			GroupName:  g.GroupName,
			Type:       g.Type,
			SwitchType: g.SwitchType,
			Cons:       cons,
		})
	}

	params := model.Condition{
		ConditionID:   req.ConditionId,
		ConditionName: req.ConditionName,
		AppID:         req.AppId,
		Groups:        groups,
		ThenValue:     req.ThenValue,
		ElseValue:     req.ElseValue,
		CreatedAt:     time.Now(),
		CreatedBy:     "",
		UpdatedAt:     time.Now(),
		UpdatedBy:     "",
	}

	err := model.AddCondition(req.GetDatabase(), req.GetAppId(), params)
	if err != nil {
		utils.ErrorLog(ActionImportJournal, err.Error())
		return err
	}

	utils.InfoLog(ActionImportJournal, utils.MsgProcessEnded)

	return nil
}

// 查找分录下载设定
func (f *Condition) FindConditions(ctx context.Context, req *condition.FindConditionsRequest, rsp *condition.FindConditionsResponse) error {

	res, err := model.FindConditions(req.GetDatabase(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionFindConditions, err.Error())
		return err
	}

	// *rsp = *res.ToProto()
	*rsp = *model.ToProto(res)

	utils.InfoLog(ActionFindConditions, utils.MsgProcessEnded)

	return nil
}
