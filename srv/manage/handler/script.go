package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/script"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Script Script信息
type Script struct{}

// log出力使用
const (
	ScriptProcessName                = "Script"
	ActionFindScriptJobs             = "FindScriptJobs"
	ActionFindScriptJob              = "FindScriptJob"
	ActionDeleteDuplicateAndAddIndex = "DeleteDuplicateAndAddIndex"
	ActionAddScriptJob               = "AddScriptJob"
	ActionModifyScriptJob            = "ModifyScriptJob"
	ActionStartScriptJob             = "StartScriptJob"
	ActionAddScriptLog               = "AddScriptLog"
)

// FindScriptJobs 通过Script通知邮件查询返回Script信息
func (u *Script) FindScriptJobs(ctx context.Context, req *script.FindScriptJobsRequest, rsp *script.FindScriptJobsResponse) error {
	utils.InfoLog(ActionFindScriptJobs, utils.MsgProcessStarted)

	scripts, err := model.FindScriptJobs(ctx, req.GetDatabase(), req.GetScriptType(), req.GetScriptVersion(), req.GetRanBy())
	if err != nil {
		utils.ErrorLog(ActionFindScriptJobs, err.Error())
		return err
	}

	res := &script.FindScriptJobsResponse{}
	for _, u := range scripts {
		res.ScriptJobs = append(res.ScriptJobs, u.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindScriptJobs, utils.MsgProcessEnded)
	return nil
}

// FindScriptJob 查找单个Script记录
func (u *Script) FindScriptJob(ctx context.Context, req *script.FindScriptJobRequest, rsp *script.FindScriptJobResponse) error {
	utils.InfoLog(ActionFindScriptJob, utils.MsgProcessStarted)

	// FindScriptJob 通过ScriptID,查找单个Script记录
	res, err := model.FindScriptJob(ctx, req.GetDatabase(), req.GetScriptId())
	if err != nil {
		utils.ErrorLog(ActionFindScriptJob, err.Error())
		return err
	}
	rsp.ScriptJob = res.ToProto()

	utils.InfoLog(ActionFindScriptJob, utils.MsgProcessEnded)
	return nil
}

// AddScriptJob 添加单个Script记录
func (u *Script) AddScriptJob(ctx context.Context, req *script.AddRequest, rsp *script.AddResponse) error {
	utils.InfoLog(ActionAddScriptJob, utils.MsgProcessStarted)

	params := model.Script{
		ScriptId:      req.GetScriptId(),
		ScriptName:    req.GetScriptName(),
		ScriptDesc:    req.GetScriptDesc(),
		ScriptType:    req.GetScriptType(),
		ScriptData:    req.GetScriptData(),
		ScriptFunc:    req.GetScriptFunc(),
		ScriptVersion: req.GetScriptVersion(),
		CreatedAt:     time.Now(),
		CreatedBy:     req.GetWriter(),
	}

	createdDate := req.GetCreatedAt()
	if len(createdDate) > 0 {
		createdAt, err := time.Parse("2006-01-02", createdDate)
		if err != nil {
			utils.ErrorLog(ActionAddScriptJob, err.Error())
			return err
		}

		params.CreatedAt = createdAt
	}

	err := model.AddScriptJob(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddScriptJob, err.Error())
		return err
	}

	utils.InfoLog(ActionAddScriptJob, utils.MsgProcessEnded)

	return nil
}

// ModifyScriptJob 更新Script的信息
func (u *Script) ModifyScriptJob(ctx context.Context, req *script.ModifyRequest, rsp *script.ModifyResponse) error {
	utils.InfoLog(ActionModifyScriptJob, utils.MsgProcessStarted)

	params := model.Script{
		ScriptName:    req.GetScriptName(),
		ScriptDesc:    req.GetScriptDesc(),
		ScriptData:    req.GetScriptData(),
		ScriptFunc:    req.GetScriptFunc(),
		ScriptVersion: req.GetScriptVersion(),
	}

	err := model.ModifyScriptJob(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyScriptJob, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyScriptJob, utils.MsgProcessEnded)
	return nil
}

// StartScriptJob 更新Script的信息
func (u *Script) StartScriptJob(ctx context.Context, req *script.StartRequest, rsp *script.StartResponse) error {
	utils.InfoLog(ActionStartScriptJob, utils.MsgProcessStarted)

	err := model.StartScriptJob(ctx, req.GetDatabase(), req.GetScriptId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionStartScriptJob, err.Error())
		return err
	}

	utils.InfoLog(ActionStartScriptJob, utils.MsgProcessEnded)
	return nil
}

// AddScriptLog 更新Script的信息
func (u *Script) AddScriptLog(ctx context.Context, req *script.AddScriptLogRequest, rsp *script.AddScriptLogResponse) error {
	utils.InfoLog(ActionAddScriptLog, utils.MsgProcessStarted)

	err := model.AddScriptLog(ctx, req.GetDatabase(), req.GetScriptId(), req.GetRunLog())
	if err != nil {
		utils.ErrorLog(ActionAddScriptLog, err.Error())
		return err
	}

	utils.InfoLog(ActionAddScriptLog, utils.MsgProcessEnded)
	return nil
}

// DeleteDuplicateAndAddIndex 删除重复Script记录并创建"script_id"的索引
func (u *Script) DeleteDuplicateAndAddIndex(ctx context.Context, req *script.DeleteScriptsRequest, rsp *script.DeleteScriptsResponse) error {
	utils.InfoLog(ActionDeleteDuplicateAndAddIndex, utils.MsgProcessStarted)

	// DeleteDuplicateAndAddIndex 删除重复Script记录并创建"script_id"的索引
	err := model.DeleteDuplicateAndAddIndex(ctx, req.GetDatabase(), req.GetScriptIds())
	if err != nil {
		utils.ErrorLog(ActionDeleteDuplicateAndAddIndex, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteDuplicateAndAddIndex, utils.MsgProcessEnded)
	return nil
}
