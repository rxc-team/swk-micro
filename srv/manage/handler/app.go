package handler

import (
	"context"
	"strconv"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// App App信息
type App struct{}

// log出力使用
const (
	AppProcessName = "App"

	ActionFindApps          = "FindApps"
	ActionFindAppListApps   = "FindAppListApps"
	ActionFindApp           = "FindApp"
	ActionAddApp            = "AddApp"
	ActionModifyApp         = "ModifyApp"
	ActionModifyAppConfigs  = "ModifyAppConfigs"
	ActionModifyAppSort     = "ModifyAppSort"
	ActionDeleteApp         = "DeleteApp"
	ActionDeleteSelectApps  = "DeleteSelectApps"
	ActionHardDeleteApps    = "HardDeleteApps"
	ActionRecoverSelectApps = "RecoverSelectApps"
	ActionNextMonth         = "NextMonth"
)

// FindApps 查找多个APP记录
func (a *App) FindApps(ctx context.Context, req *app.FindAppsRequest, rsp *app.FindAppsResponse) error {
	utils.InfoLog(ActionFindApps, utils.MsgProcessStarted)

	// 默认查询
	apps, err := model.FindApps(ctx, req.Database, req.Domain, req.AppName, req.InvalidatedIn, req.IsTrial, req.StartTime, req.EndTime, req.CopyFrom)
	if err != nil {
		utils.ErrorLog(ActionFindApps, err.Error())
		return err
	}
	res := &app.FindAppsResponse{}
	for _, a := range apps {
		res.Apps = append(res.Apps, a.ToProto())
	}
	*rsp = *res

	utils.InfoLog(ActionFindApps, utils.MsgProcessEnded)
	return nil
}

// FindAppsByIds 根据APPID数组查询多个APP记录
func (a *App) FindAppsByIds(ctx context.Context, req *app.FindAppsByIdsRequest, rsp *app.FindAppsByIdsResponse) error {
	utils.InfoLog(ActionFindAppListApps, utils.MsgProcessStarted)

	apps, err := model.FindAppsByIds(ctx, req.Database, req.Domain, req.AppIdList)
	if err != nil {
		utils.ErrorLog(ActionFindAppListApps, err.Error())
		return err
	}

	res := &app.FindAppsByIdsResponse{}
	for _, a := range apps {
		res.Apps = append(res.Apps, a.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindAppListApps, utils.MsgProcessEnded)
	return nil
}

// FindApp 通过APPID查找单个APP记录
func (a *App) FindApp(ctx context.Context, req *app.FindAppRequest, rsp *app.FindAppResponse) error {
	utils.InfoLog(ActionFindApp, utils.MsgProcessStarted)

	res, err := model.FindApp(ctx, req.GetDatabase(), req.AppId)
	if err != nil {
		utils.ErrorLog(ActionFindApp, err.Error())
		return err
	}

	rsp.App = res.ToProto()

	utils.InfoLog(ActionFindApp, utils.MsgProcessEnded)
	return nil
}

// AddApp 添加单个APP记录
func (a *App) AddApp(ctx context.Context, req *app.AddAppRequest, rsp *app.AddAppResponse) error {
	utils.InfoLog(ActionAddApp, utils.MsgProcessStarted)

	params := model.App{
		AppName:      req.GetAppName(),
		AppType:      req.GetAppType(),
		DisplayOrder: req.GetDisplayOrder(),
		Domain:       req.GetDomain(),
		TemplateID:   req.GetTemplateId(),
		IsTrial:      req.GetIsTrial(),
		StartTime:    req.GetStartTime(),
		EndTime:      req.GetEndTime(),
		CopyFrom:     req.GetCopyFrom(),
		FollowApp:    req.GetFollowApp(),
		Remarks:      req.GetRemarks(),
		CreatedAt:    time.Now(),
		CreatedBy:    req.GetWriter(),
		UpdatedAt:    time.Now(),
		UpdatedBy:    req.GetWriter(),
	}
	params.Configs = model.Configs{
		Special:         req.GetConfigs().GetSpecial(),
		SyoriYm:         req.GetConfigs().GetSyoriYm(),
		ShortLeases:     req.GetConfigs().GetShortLeases(),
		KishuYm:         req.GetConfigs().GetKishuYm(),
		MinorBaseAmount: req.GetConfigs().GetMinorBaseAmount(),
		CheckStartDate:  req.GetConfigs().GetCheckStartDate(),
	}

	id, err := model.AddApp(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddApp, err.Error())
		return err
	}

	rsp.AppId = id

	utils.InfoLog(ActionAddApp, utils.MsgProcessEnded)

	return nil
}

// ModifyApp 修改单个APP记录
func (a *App) ModifyApp(ctx context.Context, req *app.ModifyAppRequest, rsp *app.ModifyAppResponse) error {
	utils.InfoLog(ActionModifyApp, utils.MsgProcessStarted)

	isTrial, err := strconv.ParseBool(req.GetIsTrial())
	if err != nil {
		isTrial = false
	}

	params := model.App{
		AppID:     req.GetAppId(),
		AppName:   req.GetAppName(),
		IsTrial:   isTrial,
		StartTime: req.GetStartTime(),
		EndTime:   req.GetEndTime(),
		Remarks:   req.GetRemarks(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	err = model.ModifyApp(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyApp, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyApp, utils.MsgProcessEnded)
	return nil
}

// ModifyAppConfigs 修改APP的configs
func (a *App) ModifyAppConfigs(ctx context.Context, req *app.ModifyConfigsRequest, rsp *app.ModifyConfigsResponse) error {
	utils.InfoLog(ActionModifyAppConfigs, utils.MsgProcessStarted)

	config := model.Configs{
		Special:         req.GetConfigs().GetSpecial(),
		SyoriYm:         req.GetConfigs().GetSyoriYm(),
		ShortLeases:     req.GetConfigs().GetShortLeases(),
		KishuYm:         req.GetConfigs().GetKishuYm(),
		MinorBaseAmount: req.GetConfigs().GetMinorBaseAmount(),
		CheckStartDate:  req.GetConfigs().GetCheckStartDate(),
	}
	err := model.ModifyAppConfigs(ctx, req.GetDatabase(), req.GetAppId(), config)
	if err != nil {
		utils.ErrorLog(ActionModifyApp, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyAppConfigs, utils.MsgProcessEnded)
	return nil
}

// ModifyAppSort APP排序修改
func (a *App) ModifyAppSort(ctx context.Context, req *app.ModifyAppSortRequest, rsp *app.ModifyAppSortResponse) error {
	utils.InfoLog(ActionModifyAppSort, utils.MsgProcessStarted)

	var params []*model.App

	for _, app := range req.AppList {
		param := model.App{
			AppID:        app.GetAppId(),
			DisplayOrder: app.GetDisplayOrder(),
			UpdatedAt:    time.Now(),
			UpdatedBy:    req.GetWriter(),
		}
		params = append(params, &param)
	}

	_, err := model.ModifyAppSort(ctx, req.Database, params)
	if err != nil {
		utils.ErrorLog(ActionModifyAppSort, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyAppSort, utils.MsgProcessEnded)
	return nil
}

// DeleteApp 删除单个APP记录
func (a *App) DeleteApp(ctx context.Context, req *app.DeleteAppRequest, rsp *app.DeleteAppResponse) error {
	utils.InfoLog(ActionDeleteApp, utils.MsgProcessStarted)

	err := model.DeleteApp(ctx, req.Database, req.GetAppId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteApp, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteApp, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectApps 删除选中的APP记录
func (a *App) DeleteSelectApps(ctx context.Context, req *app.DeleteSelectAppsRequest, rsp *app.DeleteSelectAppsResponse) error {
	utils.InfoLog(ActionDeleteSelectApps, utils.MsgProcessStarted)

	err := model.DeleteSelectApps(ctx, req.Database, req.GetAppIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectApps, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectApps, utils.MsgProcessEnded)
	return nil
}

// HardDeleteApps 物理删除选中的APP记录
func (a *App) HardDeleteApps(ctx context.Context, req *app.HardDeleteAppsRequest, rsp *app.HardDeleteAppsResponse) error {
	utils.InfoLog(ActionHardDeleteApps, utils.MsgProcessStarted)

	err := model.HardDeleteApps(ctx, req.Database, req.GetAppIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteApps, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteApps, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectApps 恢复选中的APP记录
func (a *App) RecoverSelectApps(ctx context.Context, req *app.RecoverSelectAppsRequest, rsp *app.RecoverSelectAppsResponse) error {
	utils.InfoLog(ActionRecoverSelectApps, utils.MsgProcessStarted)

	err := model.RecoverSelectApps(ctx, req.Database, req.GetAppIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectApps, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectApps, utils.MsgProcessEnded)
	return nil
}

// NextMonth 下一月度处理
func (a *App) NextMonth(ctx context.Context, req *app.NextMonthRequest, rsp *app.NextMonthResponse) error {
	utils.InfoLog(ActionNextMonth, utils.MsgProcessStarted)

	param := model.Config{
		AppID: req.GetAppId(),
		Value: req.GetValue(),
	}

	err := model.NextMonth(ctx, req.GetDatabase(), param)
	if err != nil {
		utils.ErrorLog(ActionNextMonth, err.Error())
		return err
	}

	utils.InfoLog(ActionNextMonth, utils.MsgProcessEnded)
	return nil
}
