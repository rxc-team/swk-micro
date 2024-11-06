/*
 * @Description:报表（handler）
 * @Author: RXC 廖云江
 * @Date: 2019-08-19 14:25:57
 * @LastEditors: RXC 陳平
 * @LastEditTime: 2020-06-10 09:43:25
 */

package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/report/model"
	"rxcsoft.cn/pit3/srv/report/proto/report"
	"rxcsoft.cn/pit3/srv/report/utils"
)

// Report 报表
type Report struct{}

// log出力使用
const (
	ReportProcessName = "Report"

	ActionFindReports          = "FindReports"
	ActionFindReport           = "FindReport"
	ActionFindReportData       = "FindReportData"
	ActionDownloadReportData   = "DownloadReportData"
	ActionGenerateReportData   = "GenerateReportData"
	ActionAddReport            = "AddReport"
	ActionModifyReport         = "ModifyReport"
	ActionDeleteReport         = "DeleteReport"
	ActionDeleteSelectReports  = "DeleteSelectReports"
	ActionHardDeleteReports    = "HardDeleteReports"
	ActionRecoverSelectReports = "RecoverSelectReports"
)

// FindReports 获取所属公司所属APP下所有报表情报
func (r *Report) FindReports(ctx context.Context, req *report.FindReportsRequest, rsp *report.FindReportsResponse) error {
	utils.InfoLog(ActionFindReports, utils.MsgProcessStarted)

	reports, err := model.FindReports(req.GetDatabase(), req.GetDatastoreId(), req.GetDomain(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionFindReports, err.Error())
		return err
	}

	res := &report.FindReportsResponse{}
	for _, report := range reports {
		res.Reports = append(res.Reports, report.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindReports, utils.MsgProcessEnded)
	return nil
}

// FindReport 通过报表ID获取单个报表情报
func (r *Report) FindReport(ctx context.Context, req *report.FindReportRequest, rsp *report.FindReportResponse) error {
	utils.InfoLog(ActionFindReport, utils.MsgProcessStarted)

	res, err := model.FindReport(req.GetDatabase(), req.GetReportId())
	if err != nil {
		utils.ErrorLog(ActionFindReport, err.Error())
		return err
	}

	rsp.Report = res.ToProto()

	utils.InfoLog(ActionFindReport, utils.MsgProcessEnded)
	return nil
}

// FindReportData 通过报表ID获取报表数据
func (r *Report) FindReportData(ctx context.Context, req *report.FindReportDataRequest, rsp *report.FindReportDataResponse) error {
	utils.InfoLog(ActionFindReportData, utils.MsgProcessStarted)

	var conditions []*model.ReportCondition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.ReportCondition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}

	params := model.ReportParam{
		ReportID:      req.GetReportId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		PageIndex:     req.GetPageIndex(),
		PageSize:      req.GetPageSize(),
		Owners:        req.GetOwners(),
	}

	reportDataInfo, err := model.FindReportData(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionFindReportData, err.Error())
		return err
	}

	fields := make(map[string]*report.FieldInfo, len(reportDataInfo.Fields))
	for key, value := range reportDataInfo.Fields {
		fields[key] = value.ToProto()
	}

	var itemResult []*report.ReportData
	for _, data := range reportDataInfo.ReportData {
		itemResult = append(itemResult, data.ToProto())
	}

	rsp.Fields = fields
	rsp.ItemData = itemResult
	rsp.Total = reportDataInfo.Total
	rsp.ReportName = reportDataInfo.ReportInfo.ReportName

	utils.InfoLog(ActionFindReportData, utils.MsgProcessEnded)
	return nil
}

// Download 通过报表ID获取报表数据
func (r *Report) Download(ctx context.Context, req *report.DownloadRequest, stream report.ReportService_DownloadStream) error {
	utils.InfoLog(ActionFindReportData, utils.MsgProcessStarted)

	var conditions []*model.ReportCondition
	for _, condition := range req.GetConditionList() {
		conditions = append(conditions, &model.ReportCondition{
			FieldID:       condition.GetFieldId(),
			FieldType:     condition.GetFieldType(),
			SearchValue:   condition.GetSearchValue(),
			Operator:      condition.GetOperator(),
			IsDynamic:     condition.GetIsDynamic(),
			ConditionType: condition.GetConditionType(),
		})
	}

	params := model.ReportParam{
		ReportID:      req.GetReportId(),
		ConditionType: req.GetConditionType(),
		ConditionList: conditions,
		PageIndex:     req.GetPageIndex(),
		PageSize:      req.GetPageSize(),
		Owners:        req.GetOwners(),
	}

	err := model.DownloadReportData(req.GetDatabase(), params, stream)
	if err != nil {
		utils.ErrorLog(ActionDownloadReportData, err.Error())
		return err
	}

	utils.InfoLog(ActionDownloadReportData, utils.MsgProcessEnded)
	return nil
}

// GenerateReportData 生成报表数据
func (r *Report) GenerateReportData(ctx context.Context, req *report.GenerateReportDataRequest, rsp *report.GenerateReportDataResponse) error {
	utils.InfoLog(ActionGenerateReportData, utils.MsgProcessStarted)

	err := model.GenerateReportData(req.GetDatabase(), req.GetReportId())
	if err != nil {
		utils.ErrorLog(ActionGenerateReportData, err.Error())
		return err
	}

	utils.InfoLog(ActionGenerateReportData, utils.MsgProcessEnded)
	return nil
}

// FindCount 获取报表数据件数
func (r *Report) FindCount(ctx context.Context, req *report.CountRequest, rsp *report.CountResponse) error {
	utils.InfoLog(ActionFindReportData, utils.MsgProcessStarted)

	fs, total, err := model.FindCount(req.GetDatabase(), req.GetReportId(), req.GetOwners())
	if err != nil {
		utils.ErrorLog(ActionFindReportData, err.Error())
		return err
	}

	fields := make(map[string]*report.FieldInfo, len(fs))
	for key, value := range fs {
		fields[key] = value.ToProto()
	}

	rsp.Fields = fields
	rsp.Total = total

	utils.InfoLog(ActionFindReportData, utils.MsgProcessEnded)
	return nil
}

// AddReport 添加单个报表情报
func (r *Report) AddReport(ctx context.Context, req *report.AddReportRequest, rsp *report.AddReportResponse) error {
	utils.InfoLog(ActionAddReport, utils.MsgProcessStarted)

	params := toModelOfAdd(req)
	id, err := model.AddReport(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionAddReport, err.Error())
		return err
	}

	rsp.ReportId = id

	utils.InfoLog(ActionAddReport, utils.MsgProcessEnded)

	return nil
}

// ModifyReport 更新单个报表情报
func (r *Report) ModifyReport(ctx context.Context, req *report.ModifyReportRequest, rsp *report.ModifyReportResponse) error {
	utils.InfoLog(ActionModifyReport, utils.MsgProcessStarted)

	params := toModelOfModify(req)
	err := model.ModifyReport(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionModifyReport, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyReport, utils.MsgProcessEnded)
	return nil
}

// DeleteReport 软删除单个报表情报
func (r *Report) DeleteReport(ctx context.Context, req *report.DeleteReportRequest, rsp *report.DeleteResponse) error {
	utils.InfoLog(ActionDeleteReport, utils.MsgProcessStarted)

	err := model.DeleteReport(req.GetDatabase(), req.GetReportId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteReport, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteReport, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectReports 软删除多个报表情报
func (r *Report) DeleteSelectReports(ctx context.Context, req *report.DeleteSelectReportsRequest, rsp *report.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectReports, utils.MsgProcessStarted)

	err := model.DeleteSelectReports(req.GetDatabase(), req.GetReportIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectReports, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectReports, utils.MsgProcessEnded)
	return nil
}

// HardDeleteReports 物理删除多个报表情报
func (r *Report) HardDeleteReports(ctx context.Context, req *report.HardDeleteReportsRequest, rsp *report.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteReports, utils.MsgProcessStarted)

	err := model.HardDeleteReports(req.GetDatabase(), req.GetReportIdList(), req.GetDomain(), req.GetAppId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteReports, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteReports, utils.MsgProcessEnded)
	return nil
}

// toModelOfAdd 转换为model数据(添加报表情报数据)
func toModelOfAdd(r *report.AddReportRequest) *model.Report {

	var reportconditions []*model.ReportCondition
	for _, ch := range r.GetReportConditions() {
		reportconditions = append(reportconditions, toModelReportCondition(ch))
	}

	if r.GetIsUseGroup() {
		return &model.Report{
			Domain:           r.GetDomain(),
			AppID:            r.GetAppId(),
			DatastoreID:      r.GetDatastoreId(),
			ReportName:       r.GetReportName(),
			DisplayOrder:     r.GetDisplayOrder(),
			IsUseGroup:       r.GetIsUseGroup(),
			ReportConditions: reportconditions,
			ConditionType:    r.GetConditionType(),
			GroupInfo:        toModelGroupInfo(r.GetGroupInfo()),
			CreatedAt:        time.Now(),
			CreatedBy:        r.GetWriter(),
			UpdatedAt:        time.Now(),
			UpdatedBy:        r.GetWriter(),
		}
	}

	var selectkeyinfos []*model.KeyInfo
	for _, ch := range r.GetSelectKeyInfos() {
		selectkeyinfos = append(selectkeyinfos, toModelKeyInfo(ch))
	}

	return &model.Report{
		Domain:           r.GetDomain(),
		AppID:            r.GetAppId(),
		DatastoreID:      r.GetDatastoreId(),
		ReportName:       r.GetReportName(),
		DisplayOrder:     r.GetDisplayOrder(),
		IsUseGroup:       r.GetIsUseGroup(),
		ReportConditions: reportconditions,
		ConditionType:    r.GetConditionType(),
		SelectKeyInfos:   selectkeyinfos,
		CreatedAt:        time.Now(),
		CreatedBy:        r.GetWriter(),
		UpdatedAt:        time.Now(),
		UpdatedBy:        r.GetWriter(),
	}

}

// toModelOfModify 转换为model数据(更新报表情报数据)
func toModelOfModify(r *report.ModifyReportRequest) *model.ModifyReq {

	var reportconditions []*model.ReportCondition
	for _, ch := range r.GetReportConditions() {
		reportconditions = append(reportconditions, toModelReportCondition(ch))
	}

	var selectkeyinfos []*model.KeyInfo
	for _, ch := range r.GetSelectKeyInfos() {
		selectkeyinfos = append(selectkeyinfos, toModelKeyInfo(ch))
	}

	return &model.ModifyReq{
		DatastoreID:      r.GetDatastoreId(),
		ReportID:         r.GetReportId(),
		ReportName:       r.GetReportName(),
		DisplayOrder:     r.GetDisplayOrder(),
		IsUseGroup:       r.GetIsUseGroup(),
		ReportConditions: reportconditions,
		ConditionType:    r.GetConditionType(),
		GroupInfo:        toModelGroupInfo(r.GetGroupInfo()),
		SelectKeyInfos:   selectkeyinfos,
		Writer:           r.GetWriter(),
	}
}

// toModelReportCondition 转换为model数据(报表检索条件)
func toModelReportCondition(r *report.ReportCondition) *model.ReportCondition {
	return &model.ReportCondition{
		FieldID:       r.FieldId,
		FieldType:     r.FieldType,
		SearchValue:   r.SearchValue,
		Operator:      r.Operator,
		IsDynamic:     r.IsDynamic,
		ConditionType: r.ConditionType,
	}
}

// toModelKeyInfo 转换为model数据(字段情报)
func toModelKeyInfo(r *report.KeyInfo) *model.KeyInfo {
	return &model.KeyInfo{
		IsLookup:    r.IsLookup,
		FieldID:     r.FieldId,
		DatastoreID: r.DatastoreId,
		DataType:    r.DataType,
		AliasName:   r.AliasName,
		Sort:        r.Sort,
		IsDynamic:   r.IsDynamic,
		Unique:      r.Unique,
		Order:       r.Order,
		OptionID:    r.OptionId,
	}
}

// toModelAggreKey 转换为model数据(聚合字段情报)
func toModelAggreKey(r *report.AggreKey) *model.AggreKey {
	return &model.AggreKey{
		IsLookup:    r.IsLookup,
		FieldID:     r.FieldId,
		AggreType:   r.AggreType,
		DataType:    r.DataType,
		DatastoreID: r.DatastoreId,
		AliasName:   r.AliasName,
		Sort:        r.Sort,
		Order:       r.Order,
		OptionID:    r.OptionId,
	}
}

// toModelGroupInfo 转换为model数据(Group情报)
func toModelGroupInfo(r *report.GroupInfo) *model.GroupInfo {

	var groupkeys []*model.KeyInfo
	for _, ch := range r.GroupKeys {
		groupkeys = append(groupkeys, toModelKeyInfo(ch))
	}

	var aggrekeys []*model.AggreKey
	for _, ch := range r.AggreKeys {
		aggrekeys = append(aggrekeys, toModelAggreKey(ch))
	}

	return &model.GroupInfo{
		GroupKeys: groupkeys,
		AggreKeys: aggrekeys,
		ShowCount: r.ShowCount,
	}
}

// RecoverSelectReports 恢复选中报表情报
func (r *Report) RecoverSelectReports(ctx context.Context, req *report.RecoverSelectReportsRequest, rsp *report.RecoverSelectReportsResponse) error {
	utils.InfoLog(ActionRecoverSelectReports, utils.MsgProcessStarted)

	err := model.RecoverSelectReports(req.GetDatabase(), req.GetReportIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectReports, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectReports, utils.MsgProcessEnded)
	return nil
}
