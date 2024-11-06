package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Option 选项
type Option struct{}

// log出力使用
const (
	OptionProcessName           = "Option"
	ActionFindOption            = "FindOption"
	ActionFindOptions           = "FindOptions"
	ActionFindOptionLabels      = "FindOptionLabels"
	ActionFindOptionLabel       = "FindOptionLable"
	ActionAddOption             = "AddOption"
	ActionDeleteOption          = "DeleteOption"
	ActionDeleteSelectOptions   = "DeleteSelectOptions"
	ActionHardDeleteOptions     = "HardDeleteOptions"
	ActionDeleteOptionChild     = "DeleteOptionChild"
	ActionHardDeleteOptionChild = "HardDeleteOptionChild"
	ActionRecoverSelectOptions  = "RecoverSelectOptions"
	ActionRecoverOptionChild    = "RecoverOptionChild"
)

// FindOptions 获取所有选项
func (r *Option) FindOptions(ctx context.Context, req *option.FindOptionsRequest, rsp *option.FindOptionsResponse) error {
	utils.InfoLog(ActionFindOptions, utils.MsgProcessStarted)

	options, err := model.FindOptions(req.GetDatabase(), req.GetAppId(), req.GetOptionName(), req.GetOptionMemo(), req.GetInvalidatedIn())
	if err != nil {
		utils.ErrorLog(ActionFindOptions, err.Error())
		return err
	}

	res := &option.FindOptionsResponse{}
	for _, r := range options {
		res.Options = append(res.Options, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindOptions, utils.MsgProcessEnded)
	return nil
}

// FindOptionLabels 获取选项的所有值的数据
func (r *Option) FindOptionLabels(ctx context.Context, req *option.FindOptionLabelsRequest, rsp *option.FindOptionLabelsResponse) error {
	utils.InfoLog(ActionFindOptionLabels, utils.MsgProcessStarted)

	options, err := model.FindOptionLabels(req.GetDatabase(), req.GetAppId(), req.GetOptionName(), req.GetOptionMemo(), req.GetInvalidatedIn())
	if err != nil {
		utils.ErrorLog(ActionFindOptionLabels, err.Error())
		return err
	}

	res := &option.FindOptionLabelsResponse{}
	for _, r := range options {
		res.Options = append(res.Options, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindOptionLabels, utils.MsgProcessEnded)
	return nil
}

// FindOption 通过选项ID，选项值获取某个选项下的所有值的数据或者某个值的数据
func (r *Option) FindOption(ctx context.Context, req *option.FindOptionRequest, rsp *option.FindOptionResponse) error {
	utils.InfoLog(ActionFindOption, utils.MsgProcessStarted)

	options, err := model.FindOption(req.GetDatabase(), req.GetAppId(), req.GetOptionId(), req.GetInvalid())
	if err != nil {
		utils.ErrorLog(ActionFindOption, err.Error())
		return err
	}

	res := &option.FindOptionResponse{}
	for _, r := range options {
		res.Options = append(res.Options, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindOption, utils.MsgProcessEnded)
	return nil
}

// FindOptionLable 通过选项ID，选项值获取某个选项下的所有值的数据或者某个值的数据
func (r *Option) FindOptionLable(ctx context.Context, req *option.FindOptionLabelRequest, rsp *option.FindOptionLabelResponse) error {
	utils.InfoLog(ActionFindOptionLabel, utils.MsgProcessStarted)

	res, err := model.FindOptionLable(req.GetDatabase(), req.GetAppId(), req.GetOptionId(), req.GetOptionValue())
	if err != nil {
		utils.ErrorLog(ActionFindOptionLabel, err.Error())
		return err
	}

	rsp.Option = res.ToProto()

	utils.InfoLog(ActionFindOptionLabel, utils.MsgProcessEnded)
	return nil
}

// AddOption 添加选项
func (r *Option) AddOption(ctx context.Context, req *option.AddRequest, rsp *option.AddResponse) error {
	utils.InfoLog(ActionAddOption, utils.MsgProcessStarted)

	params := model.Option{
		OptionID:    req.GetOptionId(),
		OptionValue: req.GetOptionValue(),
		OptionLabel: req.GetOptionLabel(),
		OptionOrder: req.GetOptionOrder(),
		OptionName:  req.GetOptionName(),
		OptionMemo:  req.GetOptionMemo(),
		AppID:       req.GetAppId(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	id, option, err := model.AddOption(req.GetDatabase(), &params, req.GetIsNewOptionGroup())
	if err != nil {
		utils.ErrorLog(ActionAddOption, err.Error())
		return err
	}

	rsp.Id = id
	rsp.OptionId = option

	utils.InfoLog(ActionAddOption, utils.MsgProcessEnded)

	return nil
}

// DeleteOptionChild 删除某个选项下的某个值数据
func (r *Option) DeleteOptionChild(ctx context.Context, req *option.DeleteChildRequest, rsp *option.DeleteResponse) error {
	utils.InfoLog(ActionDeleteOptionChild, utils.MsgProcessStarted)

	err := model.DeleteOptionChild(req.GetDatabase(), req.GetAppId(), req.GetOptionId(), req.GetOptionValue(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteOptionChild, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteOptionChild, utils.MsgProcessEnded)
	return nil
}

// HardDeleteOptionChild 物理删除某个选项下的某个值数据
func (r *Option) HardDeleteOptionChild(ctx context.Context, req *option.HardDeleteChildRequest, rsp *option.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteOptionChild, utils.MsgProcessStarted)

	err := model.HardDeleteOptionChild(req.GetDatabase(), req.GetAppId(), req.GetOptionId(), req.GetOptionValue())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteOptionChild, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteOptionChild, utils.MsgProcessEnded)
	return nil
}

// DeleteOption 删除某个选项
func (r *Option) DeleteOption(ctx context.Context, req *option.DeleteOptionRequest, rsp *option.DeleteResponse) error {
	utils.InfoLog(ActionDeleteOption, utils.MsgProcessStarted)

	err := model.DeleteOption(req.GetDatabase(), req.GetAppId(), req.GetOptionId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteOption, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteOption, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectOptions 删除选中选项
func (r *Option) DeleteSelectOptions(ctx context.Context, req *option.DeleteSelectOptionsRequest, rsp *option.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectOptions, utils.MsgProcessStarted)

	err := model.DeleteSelectOptions(req.GetDatabase(), req.GetAppId(), req.GetOptionIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectOptions, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectOptions, utils.MsgProcessEnded)
	return nil
}

// HardDeleteOptions 物理删除选中选项
func (r *Option) HardDeleteOptions(ctx context.Context, req *option.HardDeleteOptionsRequest, rsp *option.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteOptions, utils.MsgProcessStarted)

	err := model.HardDeleteOptions(req.GetDatabase(), req.GetAppId(), req.GetOptionIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteOptions, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteOptions, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectOptions 恢复选中的选项
func (r *Option) RecoverSelectOptions(ctx context.Context, req *option.RecoverSelectOptionsRequest, rsp *option.RecoverSelectOptionsResponse) error {
	utils.InfoLog(ActionRecoverSelectOptions, utils.MsgProcessStarted)

	err := model.RecoverSelectOptions(req.GetDatabase(), req.GetAppId(), req.GetOptionIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectOptions, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectOptions, utils.MsgProcessEnded)
	return nil
}

// RecoverOptionChild 恢复某个选项组下的某个值数据
func (r *Option) RecoverOptionChild(ctx context.Context, req *option.RecoverChildRequest, rsp *option.RecoverChildResponse) error {
	utils.InfoLog(ActionRecoverOptionChild, utils.MsgProcessStarted)

	err := model.RecoverOptionChild(req.GetDatabase(), req.GetAppId(), req.GetOptionId(), req.GetOptionValue(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverOptionChild, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverOptionChild, utils.MsgProcessEnded)
	return nil
}
