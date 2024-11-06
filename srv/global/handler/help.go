package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/help"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Help 帮助文档
type Help struct{}

// log出力使用
const (
	ActionFindHelps   = "FindHelps"
	ActionFindTags    = "FindTags"
	ActionFindHelp    = "FindHelp"
	ActionAddHelp     = "AddHelp"
	ActionModifyHelp  = "ModifyHelp"
	ActionDeleteHelp  = "DeleteHelp"
	ActionDeleteHelps = "DeleteHelps"
)

// FindHelp 获取单个帮助文档
func (t *Help) FindHelp(ctx context.Context, req *help.FindHelpRequest, rsp *help.FindHelpResponse) error {
	utils.InfoLog(ActionFindHelp, utils.MsgProcessStarted)

	res, err := model.FindHelp(req.GetDatabase(), req.GetHelpId())
	if err != nil {
		utils.ErrorLog(ActionFindHelp, err.Error())
		return err
	}

	rsp.Help = res.ToProto()

	utils.InfoLog(ActionFindHelp, utils.MsgProcessEnded)
	return nil
}

// FindHelps 获取多个帮助文档
func (t *Help) FindHelps(ctx context.Context, req *help.FindHelpsRequest, rsp *help.FindHelpsResponse) error {
	utils.InfoLog(ActionFindHelps, utils.MsgProcessStarted)

	helpList, err := model.FindHelps(req.GetDatabase(), req.GetTitle(), req.GetType(), req.GetTag(), req.GetLangCd())
	if err != nil {
		utils.ErrorLog(ActionFindHelps, err.Error())
		return err
	}

	res := &help.FindHelpsResponse{}

	for _, help := range helpList {
		res.Helps = append(res.Helps, help.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindHelps, utils.MsgProcessEnded)
	return nil
}

// FindTags 获取所有不重复帮助文档标签
func (t *Help) FindTags(ctx context.Context, req *help.FindTagsRequest, rsp *help.FindTagsResponse) error {
	utils.InfoLog(ActionFindTags, utils.MsgProcessStarted)

	tags, err := model.FindTags(req.GetDatabase())
	if err != nil {
		utils.ErrorLog(ActionFindTags, err.Error())
		return err
	}

	rsp.Tags = tags

	utils.InfoLog(ActionFindTags, utils.MsgProcessEnded)
	return nil
}

// AddHelp 添加帮助文档
func (t *Help) AddHelp(ctx context.Context, req *help.AddHelpRequest, rsp *help.AddHelpResponse) error {
	utils.InfoLog(ActionAddHelp, utils.MsgProcessStarted)

	params := model.Help{
		HelpID:    req.GetHelpId(),
		Title:     req.GetTitle(),
		Type:      req.GetType(),
		Content:   req.GetContent(),
		Images:    req.GetImages(),
		Tags:      req.GetTags(),
		LangCd:    req.GetLangCd(),
		CreatedAt: time.Now(),
		CreatedBy: req.GetWriter(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	id, err := model.AddHelp(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddHelp, err.Error())
		return err
	}

	rsp.HelpId = id

	utils.InfoLog(ActionAddHelp, utils.MsgProcessEnded)
	return nil
}

// ModifyHelp 更新帮助文档
func (t *Help) ModifyHelp(ctx context.Context, req *help.ModifyHelpRequest, rsp *help.ModifyHelpResponse) error {
	utils.InfoLog(ActionModifyHelp, utils.MsgProcessStarted)

	params := model.Help{
		HelpID:    req.GetHelpId(),
		Title:     req.GetTitle(),
		Type:      req.GetType(),
		Content:   req.GetContent(),
		Images:    req.GetImages(),
		Tags:      req.GetTags(),
		LangCd:    req.GetLangCd(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	err := model.ModifyHelp(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyHelp, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyHelp, utils.MsgProcessEnded)
	return nil
}

// DeleteHelp 硬删除帮助文档
func (t *Help) DeleteHelp(ctx context.Context, req *help.DeleteHelpRequest, rsp *help.DeleteHelpResponse) error {
	utils.InfoLog(ActionDeleteHelp, utils.MsgProcessStarted)

	err := model.DeleteHelp(req.GetDatabase(), req.GetHelpId())
	if err != nil {
		utils.ErrorLog(ActionDeleteHelp, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteHelp, utils.MsgProcessEnded)
	return nil
}

// DeleteHelps 硬删除多个帮助文档
func (t *Help) DeleteHelps(ctx context.Context, req *help.DeleteHelpsRequest, rsp *help.DeleteHelpsResponse) error {
	utils.InfoLog(ActionDeleteHelps, utils.MsgProcessStarted)

	err := model.DeleteHelps(req.GetDatabase(), req.GetHelpIdList())
	if err != nil {
		utils.ErrorLog(ActionDeleteHelps, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteHelps, utils.MsgProcessEnded)
	return nil
}
