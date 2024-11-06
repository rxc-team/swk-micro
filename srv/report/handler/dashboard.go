/*
 * @Description:仪表盘（handler）
 * @Author: RXC 廖云江
 * @Date: 2019-08-19 14:25:57
 * @LastEditors: RXC 陳平
 * @LastEditTime: 2020-06-12 11:01:52
 */

package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/report/model"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
	"rxcsoft.cn/pit3/srv/report/utils"
)

// Dashboard 仪表盘
type Dashboard struct{}

// log出力使用
const (
	DashboardProcessName = "Dashboard"

	ActionFindDashboards          = "FindDashboards"
	ActionFindDashboard           = "FindDashboard"
	ActionFindDashboardData       = "FindDashboardData"
	ActionAddDashboard            = "AddDashboard"
	ActionModifyDashboard         = "ModifyDashboard"
	ActionDeleteDashboard         = "DeleteDashboard"
	ActionDeleteSelectDashboards  = "DeleteSelectDashboards"
	ActionHardDeleteDashboards    = "HardDeleteDashboards"
	ActionRecoverSelectDashboards = "RecoverSelectDashboards"
)

// FindDashboards 获取所属公司所属APP[所属报表]下所有仪表盘情报
func (r *Dashboard) FindDashboards(ctx context.Context, req *dashboard.FindDashboardsRequest, rsp *dashboard.FindDashboardsResponse) error {
	utils.InfoLog(ActionFindDashboards, utils.MsgProcessStarted)

	dashboards, err := model.FindDashboards(req.GetDatabase(), req.GetDomain(), req.GetAppId(), req.GetReportId())
	if err != nil {
		utils.ErrorLog(ActionFindDashboards, err.Error())
		return err
	}

	res := &dashboard.FindDashboardsResponse{}
	for _, dashboard := range dashboards {
		res.Dashboards = append(res.Dashboards, dashboard.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindDashboards, utils.MsgProcessEnded)
	return nil
}

// FindDashboard 通过仪表盘ID获取单个仪表盘情报
func (r *Dashboard) FindDashboard(ctx context.Context, req *dashboard.FindDashboardRequest, rsp *dashboard.FindDashboardResponse) error {
	utils.InfoLog(ActionFindDashboard, utils.MsgProcessStarted)

	res, err := model.FindDashboard(req.GetDatabase(), req.GetDashboardId())
	if err != nil {
		utils.ErrorLog(ActionFindDashboard, err.Error())
		return err
	}

	rsp.Dashboard = res.ToProto()

	utils.InfoLog(ActionFindDashboard, utils.MsgProcessEnded)
	return nil
}

// FindDashboardData 通过仪表盘ID获取仪表盘数据情报
func (r *Dashboard) FindDashboardData(ctx context.Context, req *dashboard.FindDashboardDataRequest, rsp *dashboard.FindDashboardDataResponse) error {
	utils.InfoLog(ActionFindDashboardData, utils.MsgProcessStarted)

	dashboardDatas, err := model.FindDashboardData(req.GetDatabase(), req.GetDashboardId(), req.GetOwners())
	if err != nil {
		utils.ErrorLog(ActionFindDashboardData, err.Error())
		return err
	}
	var datas []*dashboard.DashboardData
	for _, dashboardData := range dashboardDatas.DashboardData {
		datas = append(datas, dashboardData.ToProto())
	}

	rsp.DashboardDatas = datas
	rsp.DashboardType = dashboardDatas.DashboardInfo.DashboardType
	rsp.DashboardName = dashboardDatas.DashboardInfo.DashboardName

	utils.InfoLog(ActionFindDashboardData, utils.MsgProcessEnded)
	return nil
}

// AddDashboard 添加单个仪表盘情报
func (r *Dashboard) AddDashboard(ctx context.Context, req *dashboard.AddDashboardRequest, rsp *dashboard.AddDashboardResponse) error {
	utils.InfoLog(ActionAddDashboard, utils.MsgProcessStarted)

	params := toModelOfAddDash(req)
	id, err := model.AddDashboard(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionAddDashboard, err.Error())
		return err
	}

	rsp.DashboardId = id

	utils.InfoLog(ActionAddDashboard, utils.MsgProcessEnded)

	return nil
}

// ModifyDashboard 更新单个仪表盘情报
func (r *Dashboard) ModifyDashboard(ctx context.Context, req *dashboard.ModifyDashboardRequest, rsp *dashboard.ModifyDashboardResponse) error {
	utils.InfoLog(ActionModifyDashboard, utils.MsgProcessStarted)

	params := toModelOfModifyDash(req)
	err := model.ModifyDashboard(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionModifyDashboard, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyDashboard, utils.MsgProcessEnded)
	return nil
}

// DeleteDashboard 软删除单个仪表盘情报
func (r *Dashboard) DeleteDashboard(ctx context.Context, req *dashboard.DeleteDashboardRequest, rsp *dashboard.DeleteResponse) error {
	utils.InfoLog(ActionDeleteDashboard, utils.MsgProcessStarted)

	err := model.DeleteDashboard(req.GetDatabase(), req.GetDashboardId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteDashboard, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteDashboard, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectDashboards 软删除多个仪表盘情报
func (r *Dashboard) DeleteSelectDashboards(ctx context.Context, req *dashboard.DeleteSelectDashboardsRequest, rsp *dashboard.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectDashboards, utils.MsgProcessStarted)

	err := model.DeleteSelectDashboards(req.GetDatabase(), req.GetDashboardIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectDashboards, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectDashboards, utils.MsgProcessEnded)
	return nil
}

// HardDeleteDashboards 物理删除多个仪表盘情报
func (r *Dashboard) HardDeleteDashboards(ctx context.Context, req *dashboard.HardDeleteDashboardsRequest, rsp *dashboard.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteDashboards, utils.MsgProcessStarted)

	err := model.HardDeleteDashboards(req.GetDatabase(), req.GetDashboardIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteDashboards, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteDashboards, utils.MsgProcessEnded)
	return nil
}

// toModelOfAddDash 转换为model数据
func toModelOfAddDash(r *dashboard.AddDashboardRequest) *model.Dashboard {

	slider := model.Slider{
		Start:  r.Slider.Start,
		End:    r.Slider.End,
		Height: r.Slider.Height,
	}
	scrollbar := model.Scrollbar{
		Type:         r.Scrollbar.Type,
		Width:        r.Scrollbar.Width,
		Height:       r.Scrollbar.Height,
		CategorySize: r.Scrollbar.CategorySize,
	}

	return &model.Dashboard{
		DashboardName: r.GetDashboardName(),
		Domain:        r.GetDomain(),
		AppID:         r.GetAppId(),
		ReportID:      r.GetReportId(),
		DashboardType: r.GetDashboardType(),
		XRange:        r.GetXRange(),
		YRange:        r.GetYRange(),
		TickType:      r.GetTickType(),
		Ticks:         r.GetTicks(),
		TickCount:     r.GetTickCount(),
		GFieldID:      r.GetGFieldId(),
		XFieldID:      r.GetXFieldId(),
		YFieldID:      r.GetYFieldId(),
		LimitInPlot:   r.LimitInPlot,
		StepType:      r.StepType,
		IsStack:       r.IsStack,
		IsPercent:     r.IsPercent,
		IsGroup:       r.IsGroup,
		Smooth:        r.Smooth,
		MinBarWidth:   r.MinBarWidth,
		MaxBarWidth:   r.MaxBarWidth,
		Radius:        r.Radius,
		InnerRadius:   r.InnerRadius,
		StartAngle:    r.StartAngle,
		EndAngle:      r.EndAngle,
		Slider:        slider,
		Scrollbar:     scrollbar,
		CreatedAt:     time.Now(),
		CreatedBy:     r.GetWriter(),
		UpdatedAt:     time.Now(),
		UpdatedBy:     r.GetWriter(),
	}
}

// toModelOfModifyDash 转换为model数据
func toModelOfModifyDash(r *dashboard.ModifyDashboardRequest) *model.ModifyParams {

	slider := model.Slider{
		Start:  r.Slider.Start,
		End:    r.Slider.End,
		Height: r.Slider.Height,
	}
	scrollbar := model.Scrollbar{
		Type:         r.Scrollbar.Type,
		Width:        r.Scrollbar.Width,
		Height:       r.Scrollbar.Height,
		CategorySize: r.Scrollbar.CategorySize,
	}

	return &model.ModifyParams{
		DashboardID:   r.GetDashboardId(),
		DashboardName: r.GetDashboardName(),
		DashboardType: r.GetDashboardType(),
		XRange:        r.GetXRange(),
		YRange:        r.GetYRange(),
		TickType:      r.GetTickType(),
		Ticks:         r.GetTicks(),
		TickCount:     r.GetTickCount(),
		GFieldID:      r.GetGFieldId(),
		LimitInPlot:   r.LimitInPlot,
		StepType:      r.StepType,
		IsStack:       r.IsStack,
		IsPercent:     r.IsPercent,
		IsGroup:       r.IsGroup,
		Smooth:        r.Smooth,
		MinBarWidth:   r.MinBarWidth,
		MaxBarWidth:   r.MaxBarWidth,
		Radius:        r.Radius,
		InnerRadius:   r.InnerRadius,
		StartAngle:    r.StartAngle,
		EndAngle:      r.EndAngle,
		Slider:        slider,
		Scrollbar:     scrollbar,
		ReportID:      r.GetReportId(),
		XFieldID:      r.GetXFieldId(),
		YFieldID:      r.GetYFieldId(),
		Writer:        r.GetWriter(),
	}
}

// RecoverSelectDashboards 恢复选中仪表盘情报
func (r *Dashboard) RecoverSelectDashboards(ctx context.Context, req *dashboard.RecoverSelectDashboardsRequest, rsp *dashboard.RecoverSelectDashboardsResponse) error {
	utils.InfoLog(ActionRecoverSelectDashboards, utils.MsgProcessStarted)

	err := model.RecoverSelectDashboards(req.GetDatabase(), req.GetDashboardIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectDashboards, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectDashboards, utils.MsgProcessEnded)
	return nil
}
