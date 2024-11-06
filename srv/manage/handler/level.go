package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/level"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Level 授权等级
type Level struct{}

// log出力使用
const (
	LevelProcessName = "Level"

	LevelFindLevels   = "FindLevels"
	LevelFindLevel    = "FindLevel"
	LevelAddLevel     = "AddLevel"
	LevelModifyLevel  = "ModifyLevel"
	LevelDeleteLevel  = "DeleteLevel"
	LevelDeleteLevels = "DeleteLevels"
)

// FindLevels 查找多个授权等级记录
func (r *Level) FindLevels(ctx context.Context, req *level.FindLevelsRequest, rsp *level.FindLevelsResponse) error {
	utils.InfoLog(LevelFindLevels, utils.MsgProcessStarted)

	levels, err := model.FindLevels(ctx)
	if err != nil {
		utils.ErrorLog(LevelFindLevels, err.Error())
		return err
	}

	res := &level.FindLevelsResponse{}
	for _, r := range levels {
		res.Levels = append(res.Levels, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(LevelFindLevels, utils.MsgProcessEnded)
	return nil
}

// FindLevel 查找单个授权等级记录
func (r *Level) FindLevel(ctx context.Context, req *level.FindLevelRequest, rsp *level.FindLevelResponse) error {
	utils.InfoLog(LevelFindLevel, utils.MsgProcessStarted)

	res, err := model.FindLevel(ctx, req.GetLevelId())
	if err != nil {
		utils.ErrorLog(LevelFindLevel, err.Error())
		return err
	}

	rsp.Level = res.ToProto()

	utils.InfoLog(LevelFindLevel, utils.MsgProcessEnded)
	return nil
}

// AddLevel 添加单个授权等级记录
func (r *Level) AddLevel(ctx context.Context, req *level.AddLevelRequest, rsp *level.AddLevelResponse) error {
	utils.InfoLog(LevelAddLevel, utils.MsgProcessStarted)

	params := model.Level{
		LevelName: req.GetLevelName(),
		Allows:    req.GetAllows(),
		CreatedAt: time.Now(),
		CreatedBy: req.GetWriter(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	id, err := model.AddLevel(ctx, &params)
	if err != nil {
		utils.ErrorLog(LevelAddLevel, err.Error())
		return err
	}

	rsp.LevelId = id

	utils.InfoLog(LevelAddLevel, utils.MsgProcessEnded)
	return nil
}

// ModifyLevel 更新授权等级的信息
func (r *Level) ModifyLevel(ctx context.Context, req *level.ModifyLevelRequest, rsp *level.ModifyLevelResponse) error {
	utils.InfoLog(LevelModifyLevel, utils.MsgProcessStarted)

	err := model.ModifyLevel(ctx, req.GetLevelId(), req.GetLevelName(), req.GetWriter(), req.GetAllows())
	if err != nil {
		utils.ErrorLog(LevelModifyLevel, err.Error())
		return err
	}

	utils.InfoLog(LevelModifyLevel, utils.MsgProcessEnded)
	return nil
}

// DeleteLevel 硬删除单个授权等级记录
func (r *Level) DeleteLevel(ctx context.Context, req *level.DeleteLevelRequest, rsp *level.DeleteLevelResponse) error {
	utils.InfoLog(LevelDeleteLevel, utils.MsgProcessStarted)

	err := model.DeleteLevel(ctx, req.GetLevelId())
	if err != nil {
		utils.ErrorLog(LevelDeleteLevel, err.Error())
		return err
	}

	utils.InfoLog(LevelDeleteLevel, utils.MsgProcessEnded)
	return nil
}

// DeleteLevels 硬删除多个授权等级记录
func (r *Level) DeleteLevels(ctx context.Context, req *level.DeleteLevelsRequest, rsp *level.DeleteLevelsResponse) error {
	utils.InfoLog(LevelDeleteLevels, utils.MsgProcessStarted)

	err := model.DeleteLevels(ctx, req.GetLevelIds())
	if err != nil {
		utils.ErrorLog(LevelDeleteLevels, err.Error())
		return err
	}

	utils.InfoLog(LevelDeleteLevels, utils.MsgProcessEnded)
	return nil
}
