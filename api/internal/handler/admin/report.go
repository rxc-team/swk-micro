package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/task/proto/task"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/report/proto/report"
)

// Report 报表
type Report struct{}

// log出力
const (
	ReportProcessName          = "Report"
	ActionFindReports          = "FindReports"
	ActionFindReport           = "FindReport"
	ActionFindReportData       = "FindReportData"
	ActionGenerateReportData   = "GenerateReportData"
	ActionReportDownload       = "ReportDownload"
	ActionAddReport            = "AddReport"
	ActionModifyReport         = "ModifyReport"
	ActionDeleteReport         = "DeleteReport"
	ActionDeleteSelectReports  = "DeleteSelectReports"
	ActionHardDeleteReports    = "HardDeleteReports"
	ActionRecoverSelectReports = "RecoverSelectReports"
)

// FindReports 获取所属公司所属APP下所有报表情报
// @Router /reports [get]
func (r *Report) FindReports(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindReports, loggerx.MsgProcessStarted)

	reportService := report.NewReportService("report", client.DefaultClient)

	var req report.FindReportsRequest

	// 获取检索条件参数
	req.DatastoreId = c.Query("datastore_id")

	// 共通数据
	req.Domain = sessionx.GetUserDomain(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := reportService.FindReports(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindReports, err)
		return
	}

	needRole := c.Query("needRole")
	if needRole == "true" {

		pmService := permission.NewPermissionService("manage", client.DefaultClient)

		var preq permission.FindActionsRequest
		preq.RoleId = sessionx.GetUserRoles(c)
		preq.PermissionType = "app"
		preq.AppId = sessionx.GetCurrentApp(c)
		preq.ActionType = "report"
		preq.Database = sessionx.GetUserCustomer(c)
		pResp, err := pmService.FindActions(context.TODO(), &preq)
		if err != nil {
			httpx.GinHTTPError(c, ActionCheckAction, err)
			return
		}
		set := containerx.New()
		for _, act := range pResp.GetActions() {
			if act.ActionMap["read"] {
				set.Add(act.ObjectId)
			}
		}

		rpList := set.ToList()
		allDs := response.GetReports()
		var result []*report.Report
		for _, reportID := range rpList {
			f, err := findReport(reportID, allDs)
			if err == nil {
				result = append(result, f)
			}
		}

		loggerx.InfoLog(c, ActionFindReports, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastores)),
			Data:    result,
		})

		return
	}

	loggerx.InfoLog(c, ActionFindReports, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionFindReports)),
		Data:    response.GetReports(),
	})
}

func findReport(reportID string, reportList []*report.Report) (r *report.Report, err error) {
	var reuslt *report.Report
	for _, r := range reportList {
		if r.GetReportId() == reportID {
			reuslt = r
			break
		}
	}

	if reuslt == nil {
		return nil, fmt.Errorf("not found")
	}

	return reuslt, nil
}

// FindReport 通过报表ID获取单个报表情报
// @Router /reports/{rp_id} [get]
func (r *Report) FindReport(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindReport, loggerx.MsgProcessStarted)

	reportService := report.NewReportService("report", client.DefaultClient)

	var req report.FindReportRequest
	req.ReportId = c.Param("rp_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := reportService.FindReport(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindReport, err)
		return
	}

	loggerx.InfoLog(c, ActionFindReport, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionFindReport)),
		Data:    response.GetReport(),
	})
}

// FindReportData 通过报表ID获取单个报表数据
// @Router /reports/{rp_id}/data [post]
func (r *Report) FindReportData(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindReportData, loggerx.MsgProcessStarted)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	reportService := report.NewReportService("report", client.DefaultClient)

	var freq report.FindReportRequest
	freq.ReportId = c.Param("rp_id")
	freq.Database = sessionx.GetUserCustomer(c)

	fresp, err := reportService.FindReport(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindReport, err)
		return
	}

	var req report.FindReportDataRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionFindReportData, err)
		return
	}
	// 从path中获取参数
	req.ReportId = c.Param("rp_id")
	// 从共通中获取参数
	req.Owners = sessionx.GetUserAccessKeys(c, fresp.GetReport().GetDatastoreId(), "R")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := reportService.FindReportData(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindReportData, err)
		return
	}

	loggerx.InfoLog(c, ActionFindReportData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionFindReportData)),
		Data:    response,
	})
}

// GenerateReportData 生成报表数据
// @Router /reports/{rp_id}/data [post]
func (r *Report) GenerateReportData(c *gin.Context) {
	loggerx.InfoLog(c, ActionGenerateReportData, loggerx.MsgProcessStarted)

	jobID := "job_" + time.Now().Format("20060102150405")
	userID := sessionx.GetAuthUserID(c)
	appID := sessionx.GetCurrentApp(c)
	reportId := c.Param("rp_id")
	db := sessionx.GetUserCustomer(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)

	go func() {
		taskData := task.AddRequest{
			JobId:        jobID,
			JobName:      "generate report data",
			Origin:       "apps." + appID + ".reports." + reportId,
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(lang, "job.J_014"),
			TaskType:     "generate-report-data",
			Steps:        []string{"start", "generate-data", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		}

		jobx.CreateTask(taskData)

		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_012"),
			CurrentStep: "generate-data",
			Database:    db,
		}, userID)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		reportService := report.NewReportService("report", client.DefaultClient)

		var req report.GenerateReportDataRequest
		req.ReportId = reportId
		req.Database = db

		_, err := reportService.GenerateReportData(context.TODO(), &req, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "generate-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)
	}()

	loggerx.InfoLog(c, ActionGenerateReportData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionGenerateReportData)),
		Data:    gin.H{},
	})
}

// AddReport 添加单个报表情报
// @Router /reports [post]
func (r *Report) AddReport(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddReport, loggerx.MsgProcessStarted)

	reportService := report.NewReportService("report", client.DefaultClient)

	var req report.AddReportRequest

	// Body数据
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddReport, err)
		return
	}

	// 共通数据
	req.Domain = sessionx.GetUserDomain(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := reportService.AddReport(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddReport, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddReport, fmt.Sprintf("Report[%s] Create Success", response.GetReportId()))

	// 添加报表成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
	params["report_name"] = req.GetReportName()   // 新规的时候取传入参数

	loggerx.ProcessLog(c, ActionAddReport, msg.L068, params)

	// 添加多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    sessionx.GetCurrentApp(c),
		Type:     "reports",
		Key:      response.GetReportId(),
		Value:    req.GetReportName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddReport, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddReport, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, req.Domain)

	loggerx.InfoLog(c, ActionAddReport, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionAddReport)),
		Data:    response,
	})
}

// ModifyReport 更新单个报表情报
// @Router /reports/{rp_id} [PUT]
func (r *Report) ModifyReport(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyReport, loggerx.MsgProcessStarted)

	reportService := report.NewReportService("report", client.DefaultClient)

	// 变更前查询报表信息
	var freq report.FindReportRequest
	freq.ReportId = c.Param("rp_id")
	freq.Database = sessionx.GetUserCustomer(c)

	fresponse, err := reportService.FindReport(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyReport, err)
		return
	}

	var req report.ModifyReportRequest
	// Body数据
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyReport, err)
		return
	}
	// Path数据
	req.ReportId = c.Param("rp_id")
	// 共通数据
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := reportService.ModifyReport(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyReport, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyReport, fmt.Sprintf("Report[%s] update Success", req.GetReportId()))

	// 更新报表成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
	params["report_name"] = "{{" + fresponse.GetReport().GetReportName() + "}}"

	loggerx.ProcessLog(c, ActionModifyReport, msg.L070, params)

	if req.GetReportName() != "" {
		// 添加多语言数据
		langService := language.NewLanguageService("global", client.DefaultClient)

		languageReq := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    sessionx.GetCurrentApp(c),
			Type:     "reports",
			Key:      req.GetReportId(),
			Value:    req.GetReportName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}

		_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyReport, err)
			return
		}
		loggerx.SuccessLog(c, ActionModifyReport, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))
		// 通知刷新多语言数据
		langx.RefreshLanguage(req.Writer, req.Domain)
	}

	loggerx.InfoLog(c, ActionModifyReport, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionModifyReport)),
		Data:    response,
	})
}

// HardDeleteReports 物理删除多个报表情报
// @Router /phydel/reports [delete]
func (r *Report) HardDeleteReports(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteReports, loggerx.MsgProcessStarted)

	reportService := report.NewReportService("report", client.DefaultClient)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)
	appId := sessionx.GetCurrentApp(c)

	langData := langx.GetLanguageData(db, lang, domain)

	// 删除报表之前查询报表名
	reportNameList := make(map[string]string)
	for _, id := range c.QueryArray("report_id_list") {
		var freq report.FindReportRequest
		freq.ReportId = id
		freq.Database = sessionx.GetUserCustomer(c)

		fresponse, err := reportService.FindReport(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteReports, err)
			return
		}
		reportNameList[id] = langx.GetLangValue(langData, fresponse.GetReport().GetReportName(), langx.DefaultResult)
	}

	var req report.HardDeleteReportsRequest
	// Query数据
	req.ReportIdList = c.QueryArray("report_id_list")
	req.Database = sessionx.GetUserCustomer(c)
	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.AppId = sessionx.GetCurrentApp(c)

	response, err := reportService.HardDeleteReports(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteReports, err)
		return
	}
	loggerx.SuccessLog(c, ActionHardDeleteReports, fmt.Sprintf("Report[%s] physically delete Success", req.GetReportIdList()))

	for _, r := range req.GetReportIdList() {
		// 删除报表后保存日志到DB
		rname := strings.Builder{}
		rname.WriteString(reportNameList[r])
		rname.WriteString("(")
		rname.WriteString(sessionx.GetCurrentLanguage(c))
		rname.WriteString(")")
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["report_name"] = rname.String()

		langx.DeleteAppLanguageData(db, domain, appId, "reports", r)

		loggerx.ProcessLog(c, ActionHardDeleteReports, msg.L069, params)
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionHardDeleteReports, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionHardDeleteReports)),
		Data:    response,
	})
}
