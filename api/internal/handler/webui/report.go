package webui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"

	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/excelx"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/userx"
	"rxcsoft.cn/pit3/api/internal/common/poolx"
	"rxcsoft.cn/pit3/api/internal/common/storex"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/report/proto/coldata"
	"rxcsoft.cn/pit3/srv/report/proto/report"
)

// Report 报表
type Report struct{}

// log出力
const (
	ReportProcessName        = "Report"
	ActionFindReports        = "FindReports"
	ActionFindReport         = "FindReport"
	ActionFindReportData     = "FindReportData"
	ActionGenerateReportData = "GenerateReportData"
	ActionReportDownload     = "ReportDownload"
	ActionCreateColData      = "CreateColData"
)

// 格式为  Mon Jan 02 2006 15:04:05 GMT+0900 (日本標準時)
// GMT时间转换器
func DateHandle(date string) (int64, int64, error) {
	runestr := []rune(date)
	datalength := len(runestr) - 17
	if datalength > 0 {
		content := date[0:datalength]
		newdata, err := time.Parse("Mon Jan 02 2006 15:04:05", content)
		if err != nil {
			return 0, 0, err
		}
		month, err := strconv.ParseInt(fmt.Sprintf("%d", newdata.Month()), 10, 64)
		if err != nil {
			return 0, 0, err
		}
		newYear := int64(newdata.Year())
		newMonth := month
		return newYear, newMonth, nil
	}
	return 0, 0, nil
}

// SelectColData 通过契约番号，年月搜索总表数据
// @Router /colData [get]
func (u *Report) SelectColData(c *gin.Context) {
	loggerx.InfoLog(c, "SelectColData", loggerx.MsgProcessStarted)
	coldataService := coldata.NewColDataService("report", client.DefaultClient)

	var req coldata.SelectColDataRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, "FindColData", err)
		return
	}

	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)
	if req.Date != "" {
		req.Year, _ = strconv.ParseInt(req.Date[0:4], 10, 64)
		req.Month, _ = strconv.ParseInt(req.Date[5:7], 10, 64)
	} else {
		req.Year = 0
		req.Month = 0
	}

	response, err := coldataService.SelectColData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, "SelectColData", err)
		return
	}

	// loggerx.InfoLog(c, ActionFindUsers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindUsers)),
		Data:    response,
	})
}

// FindColData 获取总表全部数据
// @Router /colData [get]
func (u *Report) FindColDatas(c *gin.Context) {
	loggerx.InfoLog(c, "FindColData", loggerx.MsgProcessStarted)
	coldataService := coldata.NewColDataService("report", client.DefaultClient)

	var req coldata.FindColDatasRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionFindReportData, err)
		return
	}
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)

	response, err := coldataService.FindColDatas(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, "FindColData", err)
		return
	}

	loggerx.InfoLog(c, "FindColData", loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, userProcessName, ActionFindUsers)),
		Data:    response,
	})
}

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
	store := storex.NewRedisStore(600)
	val := store.Get(reportId, false)

	if len(val) == 0 {
		store.Set(reportId, userID)
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
			store.Set(reportId, "")
		}()

		loggerx.InfoLog(c, ActionGenerateReportData, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionGenerateReportData)),
			Data:    gin.H{},
		})
	} else {
		loggerx.InfoLog(c, ActionGenerateReportData, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionGenerateReportData)),
			Data: gin.H{
				"msg": "レポートは実行中です。複数回実行しないでください。",
			},
		})
	}

}

// CreateColData 创建总表
// @Router /:rp_id/create [post]
func (u *Report) CreateColData(c *gin.Context) {
	loggerx.InfoLog(c, ActionCreateColData, loggerx.MsgProcessStarted)

	var req coldata.CreateColDataRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionCreateColData, err)
		return
	}
	items := req.Items
	jobID := "job_" + time.Now().Format("20060102150405")
	userID := sessionx.GetAuthUserID(c)
	appID := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	store := storex.NewRedisStore(600)
	val := store.Get(appID, false)

	if len(val) == 0 {
		store.Set(appID, userID)
		go func() {
			taskData := task.AddRequest{
				JobId:        jobID,
				JobName:      "Create summary table",
				Origin:       "統合試算結果",
				UserId:       userID,
				ShowProgress: false,
				Message:      i18n.Tr(lang, "job.J_014"),
				TaskType:     "Create-summary-table",
				Steps:        []string{"start", "summary-table", "end"},
				CurrentStep:  "start",
				Database:     db,
				AppId:        appID,
			}

			jobx.CreateTask(taskData)

			// 发送消息 开始编辑数据
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_012"),
				CurrentStep: "summary-table",
				Database:    db,
			}, userID)

			var opss client.CallOption = func(o *client.CallOptions) {
				o.RequestTimeout = time.Hour * 1
				o.DialTimeout = time.Hour * 1
			}

			colDataService := coldata.NewColDataService("report", client.DefaultClient)

			var req coldata.CreateColDataRequest
			req.Database = db
			req.Items = items

			_, err := colDataService.CreateColData(context.TODO(), &req, opss)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "summary-table",
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
			store.Set(appID, "")
		}()

		loggerx.InfoLog(c, ActionCreateColData, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionCreateColData)),
			Data:    gin.H{},
		})
	} else {
		loggerx.InfoLog(c, ActionCreateColData, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionCreateColData)),
			Data: gin.H{
				"msg": "レポートは実行中です。複数回実行しないでください。",
			},
		})
	}

}

// Download 通过报表ID获取单个报表数据
// @Router /reports/{rp_id}/data [get]
func (r *Report) Download(c *gin.Context) {

	type ReportDownloadConditions struct {
		ConditionList []*report.Condition `json:"condition_list"`
		ConditionType string              `json:"condition_type"`
	}

	loggerx.InfoLog(c, ActionReportDownload, loggerx.MsgProcessStarted)

	jobID := c.Query("job_id")
	fileType := c.Query("file_type")
	reportID := c.Param("rp_id")
	userID := sessionx.GetAuthUserID(c)
	domain := sessionx.GetUserDomain(c)
	appID := sessionx.GetCurrentApp(c)
	langCd := sessionx.GetCurrentLanguage(c)
	db := sessionx.GetUserCustomer(c)
	appRoot := "app_" + appID

	// 从body中获取参数
	var downConditions ReportDownloadConditions
	if err := c.BindJSON(&downConditions); err != nil {
		httpx.GinHTTPError(c, ActionDownload, err)
		return
	}

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "report file download",
		Origin:       "apps." + appID + ".reports." + reportID,
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(langCd, "job.J_014"),
		TaskType:     "rp-csv-download",
		Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})
	sp, err := poolx.NewSystemPool()
	if err != nil {
		httpx.GinHTTPError(c, ActionDownload, err)
		return
	}

	// 发送消息 开始编辑数据
	if sp.Free() == 0 {
		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(langCd, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)
	}

	syncRun := func() {
		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		reportService := report.NewReportService("report", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}
		freq := report.FindReportRequest{
			ReportId: reportID,
			Database: db,
		}

		fresp, err := reportService.FindReport(context.TODO(), &freq)
		if err != nil {
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "build-data",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
		}

		accessKeys := sessionx.GetAccessKeys(db, userID, fresp.GetReport().GetDatastoreId(), "R")
		cReq := report.CountRequest{
			ReportId: reportID,
			Owners:   accessKeys,
			Database: db,
		}

		cResp, err := reportService.FindCount(context.TODO(), &cReq, opss)
		if err != nil {
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "build-data",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
		}

		var reportFields []*typesx.FieldInfo
		for fieldID, field := range cResp.GetFields() {
			if field.GetIsDynamic() {
				reportFields = append(reportFields, &typesx.FieldInfo{
					FieldID:     fieldID,
					DataType:    field.GetDataType(),
					AliasName:   field.GetAliasName(),
					DatastoreID: field.GetDatastoreId(),
					IsDynamic:   field.GetIsDynamic(),
					Order:       field.GetOrder(),
				})
			} else {
				reportFields = append(reportFields, &typesx.FieldInfo{
					FieldID:     fieldID,
					DataType:    field.GetDataType(),
					AliasName:   field.GetAliasName(),
					DatastoreID: field.GetDatastoreId(),
					IsDynamic:   field.GetIsDynamic(),
					Order:       field.GetOrder(),
				})
			}
		}

		// 获取当前app的语言数据
		langData := langx.GetLanguageData(db, langCd, domain)

		var allUsers []*user.User
		// 排序
		sort.Sort(typesx.FieldInfoList(reportFields))

		var headers []string

		fixedFields := make(map[string]*typesx.FixedField, 7)

		fixedFields["created_at"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_028"),
		}
		fixedFields["created_by"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_029"),
		}
		fixedFields["updated_at"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_030"),
		}
		fixedFields["updated_by"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_031"),
		}
		fixedFields["checked_at"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_032"),
		}
		fixedFields["checked_by"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_033"),
		}
		fixedFields["check_type"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_034"),
		}
		fixedFields["check_status"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_035"),
		}
		fixedFields["label_time"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_036"),
		}
		fixedFields["count"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_037"),
		}
		fixedFields["checkOver"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_038"),
		}
		fixedFields["checkWait"] = &typesx.FixedField{
			FieldName: i18n.Tr(langCd, "fixed.F_039"),
		}

		for _, fl := range reportFields {

			if fl.IsDynamic {
				headers = append(headers, fl.AliasName)
			} else {
				headers = append(headers, fixedFields[fl.FieldID].FieldName)
			}

			if fl.DataType == "user" && allUsers == nil && len(allUsers) == 0 {
				allUsers = userx.GetAllUser(db, appID, domain)
			}
		}

		timestamp := time.Now().Format("20060102150405")

		total := float64(cResp.GetTotal())

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(langCd, "job.J_033"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		headerData := append([][]string{}, headers)
		var req report.DownloadRequest
		req.ConditionList = downConditions.ConditionList
		req.ConditionType = downConditions.ConditionType
		req.ReportId = reportID
		req.Owners = accessKeys
		req.Database = db

		stream, err := reportService.Download(context.TODO(), &req, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "write-to-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// Excel文件下载
		if fileType == "xlsx" {
			excelFile := excelize.NewFile()
			// 创建一个工作表
			index := excelFile.NewSheet("Sheet1")
			// 设置工作簿的默认工作表
			excelFile.SetActiveSheet(index)

			// 标题写入
			for i, rows := range headerData {
				for j, v := range rows {
					y := excelx.GetAxisY(j+1) + strconv.Itoa(i+1)
					excelFile.SetCellValue("Sheet1", y, v)
				}
			}
			var current int = 0
			// 数据写入

			var items [][]string
			line := 0
			// 设置csv行
			for {
				dt, err := stream.Recv()
				if err == io.EOF {
					// 当前结束了，但是items还有数据
					if len(items) > 0 {

						// 返回消息
						result := make(map[string]interface{})

						result["total"] = total
						result["current"] = current

						message, _ := json.Marshal(result)

						// 发送消息 写入条数
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     string(message),
							CurrentStep: "write-to-file",
							Database:    db,
						}, userID)

						// 写入数据
						for k, rows := range items {
							for j, v := range rows {
								y := excelx.GetAxisY(j+1) + strconv.Itoa(line*500+k+3)
								excelFile.SetCellValue("Sheet1", y, v)
							}
						}
					}
					break
				}
				var row []string
				for _, field := range reportFields {
					if field.IsDynamic {
						if value, ok := dt.ItemData.Items[field.FieldID]; ok {
							switch value.DataType {
							case "text", "textarea", "number", "time", "switch", "datetime":
								row = append(row, value.Value)
							case "autonum":
								row = append(row, value.Value)
							case "date":
								if value.GetValue() == "0001-01-01" {
									row = append(row, "")
								} else {
									row = append(row, value.GetValue())
								}
							case "lookup":
								row = append(row, value.Value)
							case "options":
								row = append(row, langx.GetLangValue(langData, value.GetValue(), langx.DefaultResult))
							case "user":
								row = append(row, value.GetValue())
							case "file":
								var files []typesx.FileValue
								json.Unmarshal([]byte(value.GetValue()), &files)
								var fileStrList []string
								for _, f := range files {
									fileStrList = append(fileStrList, f.Name)
								}
								row = append(row, strings.Join(fileStrList, ","))
							}
						} else {
							row = append(row, "")
						}
					} else {
						switch field.FieldID {
						case "created_at":
							if strings.HasPrefix(dt.GetItemData().GetCreatedAt(), "0001-01-01") {
								row = append(row, "")
							} else {
								row = append(row, dt.GetItemData().GetCreatedAt())
							}
						case "created_by":
							row = append(row, userx.TranUser(dt.GetItemData().GetCreatedBy(), allUsers))
						case "updated_at":
							if strings.HasPrefix(dt.GetItemData().GetUpdatedAt(), "0001-01-01") {
								row = append(row, "")
							} else {
								row = append(row, dt.GetItemData().GetUpdatedAt())
							}
						case "updated_by":
							row = append(row, userx.TranUser(dt.GetItemData().GetUpdatedBy(), allUsers))
						case "checked_at":
							if strings.HasPrefix(dt.GetItemData().GetCheckedAt(), "0001-01-01") {
								row = append(row, "")
							} else {
								row = append(row, dt.GetItemData().GetCheckedAt())
							}
						case "checked_by":
							row = append(row, userx.TranUser(dt.GetItemData().GetCheckedBy(), allUsers))
						case "check_type":
							row = append(row, dt.GetItemData().GetCheckType())
						case "check_status":
							status := "checkWait"
							if dt.GetItemData().GetCheckStatus() == "1" {
								status = "checkOver"
							}
							row = append(row, fixedFields[status].FieldName)
						}
					}
				}

				if dt.GetItemData().GetCount() != 0 {
					row = append(row, strconv.FormatInt(dt.GetItemData().GetCount(), 10))
				}

				items = append(items, row)
				current++
			}

			for _, rows := range items {
				for j, v := range rows {
					y := excelx.GetAxisY(j+1) + strconv.Itoa(int((current + 2)))
					excelFile.SetCellValue("Sheet1", y, v)
				}
			}
			outFile := "text.xlsx"

			if err := excelFile.SaveAs(outFile); err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			fo, err := os.Open(outFile)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			defer func() {
				fo.Close()
				os.Remove(outFile)
			}()

			// 写入文件到 minio
			minioClient, err := storagecli.NewClient(domain)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			dir := path.Join(appRoot, "excel", "report_"+timestamp+".xlsx")
			path, err := minioClient.SavePublicObject(fo, dir, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			// 判断顾客上传文件是否在设置的最大存储空间以内
			canUpload := filex.CheckCanUpload(domain, float64(path.Size))
			if canUpload {
				// 如果没有超出最大值，就对顾客的已使用大小进行累加
				err = filex.ModifyUsedSize(domain, float64(path.Size))
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 保存文件失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "save-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)
					return
				}
			} else {
				// 如果已达上限，则删除刚才上传的文件
				minioClient.DeleteObject(path.Name)
				path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     i18n.Tr(langCd, "job.J_007"),
					CurrentStep: "save-file",
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
				Message:     i18n.Tr(langCd, "job.J_028"),
				CurrentStep: "end",
				File: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

		} else {
			// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
			headerData[0][0] = "\xEF\xBB\xBF" + headerData[0][0]

			filex.Mkdir("temp/")

			// 写入文件到本地
			filename := "temp/tmp" + "_" + timestamp + ".csv"
			f, err := os.Create(filename)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "write-to-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			writer := csv.NewWriter(f)
			writer.WriteAll(headerData)

			writer.Flush() // 此时才会将缓冲区数据写入

			var current int = 0

			var items [][]string
			for {
				dt, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						// 当前结束了，但是items还有数据
						if len(items) > 0 {

							// 返回消息
							result := make(map[string]interface{})

							result["total"] = total
							result["current"] = current

							message, _ := json.Marshal(result)

							// 发送消息 写入条数
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     string(message),
								CurrentStep: "write-to-file",
								Database:    db,
							}, userID)

							// 写入数据
							err = writer.WriteAll(items)
							if err != nil {
								if err.Error() == "encoding: rune not supported by encoding." {
									path := filex.WriteAndSaveFile(domain, appID, []string{"現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"})
									// 发送消息 获取数据失败，终止任务
									jobx.ModifyTask(task.ModifyRequest{
										JobId:       jobID,
										Message:     err.Error(),
										CurrentStep: "write-to-file",
										EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
										ErrorFile: &task.File{
											Url:  path.MediaLink,
											Name: path.Name,
										},
										Database: db,
									}, userID)

									return
								}

								path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
								// 发送消息 获取数据失败，终止任务
								jobx.ModifyTask(task.ModifyRequest{
									JobId:       jobID,
									Message:     err.Error(),
									CurrentStep: "write-to-file",
									EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
									ErrorFile: &task.File{
										Url:  path.MediaLink,
										Name: path.Name,
									},
									Database: db,
								}, userID)

								return
							}

							// 缓冲区数据写入
							writer.Flush()
						}
						break
					}
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "write-to-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}

				var row []string
				for _, field := range reportFields {
					if field.IsDynamic {
						if value, ok := dt.GetItemData().Items[field.FieldID]; ok {
							switch value.DataType {
							case "text", "textarea", "number", "time", "switch", "datetime":
								row = append(row, value.Value)
							case "autonum":
								row = append(row, value.Value)
							case "date":
								if value.GetValue() == "0001-01-01" {
									row = append(row, "")
								} else {
									row = append(row, value.GetValue())
								}
							case "lookup":
								row = append(row, value.Value)
							case "options":
								row = append(row, langx.GetLangValue(langData, value.GetValue(), langx.DefaultResult))
							case "user":
								row = append(row, value.GetValue())
							case "file":
								var files []typesx.FileValue
								json.Unmarshal([]byte(value.GetValue()), &files)
								var fileStrList []string
								for _, f := range files {
									fileStrList = append(fileStrList, f.Name)
								}
								row = append(row, strings.Join(fileStrList, ","))
							}
						} else {
							row = append(row, "")
						}
					} else {
						switch field.FieldID {
						case "created_at":
							if strings.HasPrefix(dt.GetItemData().GetCreatedAt(), "0001-01-01") {
								row = append(row, "")
							} else {
								row = append(row, dt.GetItemData().GetCreatedAt())
							}
						case "created_by":
							row = append(row, userx.TranUser(dt.GetItemData().GetCreatedBy(), allUsers))
						case "updated_at":
							if strings.HasPrefix(dt.GetItemData().GetUpdatedAt(), "0001-01-01") {
								row = append(row, "")
							} else {
								row = append(row, dt.GetItemData().GetUpdatedAt())
							}
						case "updated_by":
							row = append(row, userx.TranUser(dt.GetItemData().GetUpdatedBy(), allUsers))
						case "checked_at":
							if strings.HasPrefix(dt.GetItemData().GetCheckedAt(), "0001-01-01") {
								row = append(row, "")
							} else {
								row = append(row, dt.GetItemData().GetCheckedAt())
							}
						case "checked_by":
							row = append(row, userx.TranUser(dt.GetItemData().GetCheckedBy(), allUsers))
						case "check_type":
							row = append(row, dt.GetItemData().GetCheckType())
						case "check_status":
							status := "checkWait"
							if dt.GetItemData().GetCheckStatus() == "1" {
								status = "checkOver"
							}
							row = append(row, fixedFields[status].FieldName)

						}
					}
				}

				if dt.GetItemData().GetCount() != 0 {
					row = append(row, strconv.FormatInt(dt.GetItemData().GetCount(), 10))
				}

				items = append(items, row)
				current++
				if current%500 == 0 {
					// 返回消息
					result := make(map[string]interface{})

					result["total"] = total
					result["current"] = current

					message, _ := json.Marshal(result)

					// 发送消息 写入条数
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     string(message),
						CurrentStep: "write-to-file",
						Database:    db,
					}, userID)

					// 写入数据
					err = writer.WriteAll(items)
					if err != nil {
						if err.Error() == "encoding: rune not supported by encoding." {
							path := filex.WriteAndSaveFile(domain, appID, []string{"現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "write-to-file",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: db,
							}, userID)

							return
						}

						path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     err.Error(),
							CurrentStep: "write-to-file",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: db,
						}, userID)

						return
					}

					// 缓冲区数据写入
					writer.Flush()

					// 清空items
					items = items[:0]
				}
			}

			defer stream.Close()
			defer f.Close()

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(langCd, "job.J_029"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(langCd, "job.J_043"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			fo, err := os.Open(filename)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			defer func() {
				fo.Close()
				os.Remove(filename)
			}()

			// 写入文件到 minio
			minioClient, err := storagecli.NewClient(domain)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			dir := path.Join(appRoot, "csv", "report_"+timestamp+".csv")
			path, err := minioClient.SavePublicObject(fo, dir, "text/csv")
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			// 判断顾客上传文件是否在设置的最大存储空间以内
			canUpload := filex.CheckCanUpload(domain, float64(path.Size))
			if canUpload {
				// 如果没有超出最大值，就对顾客的已使用大小进行累加
				err = filex.ModifyUsedSize(domain, float64(path.Size))
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 保存文件失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "save-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
			} else {
				// 如果已达上限，则删除刚才上传的文件
				minioClient.DeleteObject(path.Name)
				path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     i18n.Tr(langCd, "job.J_007"),
					CurrentStep: "save-file",
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
				Message:     i18n.Tr(langCd, "job.J_028"),
				CurrentStep: "end",
				File: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

		}
	}

	err = sp.Submit(syncRun)
	if err != nil {
		httpx.GinHTTPError(c, ActionDownload, err)
		return
	}

	loggerx.InfoLog(c, ActionReportDownload, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ReportProcessName, ActionReportDownload)),
		Data:    gin.H{},
	})
}

// ColDataDownload 下载总表
// @Router report/downloadColData
func (u *Report) ColDataDownload(c *gin.Context) {
	loggerx.InfoLog(c, "ColDateDownload", loggerx.MsgProcessStarted)

	appID := sessionx.GetCurrentApp(c)
	jobID := "job_" + time.Now().Format("20060102150405")
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)
	userID := sessionx.GetAuthUserID(c)
	db := sessionx.GetUserCustomer(c)

	var req coldata.DownloadRequest
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Keiyakuno = c.Query("keiyakuno")
	date := c.Query("date")

	// 报表csv下载标题字段
	titlename := c.QueryArray("titlename")

	newYear, newMonth, err := DateHandle(date)
	if err != nil {
		httpx.GinHTTPError(c, "ColDateDownload", err)
		return
	}

	req.Year = newYear
	req.Month = newMonth

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "Download coldata",
		Origin:       "統合試算結果",
		UserId:       userID,
		ShowProgress: true,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "coldata-csv-download",
		Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	// 发送消息 数据准备
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "依存データを取得します",
		CurrentStep: "build-data",
		Database:    db,
	}, userID)

	//获取上传流
	coldataService := coldata.NewColDataService("report", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	stream, err := coldataService.Download(context.Background(), &req, opss)

	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据查询错误
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルアップロードの初期化に失敗しました",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	defer stream.Close()

	// 发送消息 开始读写数据
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "データの読み書きを開始します",
		CurrentStep: "write-to-file",
		Database:    db,
	}, userID)

	timestamp := time.Now().Format("20060102150405")

	// 标题字段
	header := []string{"契約管理番号", "契約年月日", "リース開始年月日", "リース期間", "リース満了年月日",
		"延長オプション期間", "契約名称", "備考1", "初回支払年月日", "支払サイクル", "支払日", "支払回数", "残価保証額",
		"追加借入利子率", "当初直接費用", "原状回復コスト", "使用権資産の計算方法", "管理部門", "分類コード", titlename[0],
		titlename[1], titlename[2], titlename[3], titlename[4], titlename[5], titlename[6], titlename[7],
		titlename[8], titlename[9], titlename[10], titlename[11], titlename[12], titlename[13], titlename[14],
		"適用開始時点の残存リース料", "利益剰余金", "年", "月", "支払リース料", "期首元本残高", "元本返済相当額",
		"元本残高（当月末）", "支払利息相当額", "使用権資産額期首簿価", "使用権資産償却費", "使用権資産当月末帳簿価格"}
	// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
	header[0] = "\xEF\xBB\xBF" + header[0]

	headers := append([][]string{}, header)

	filex.Mkdir("temp/")

	// 写入文件到本地
	filename := "temp/coldata" + "_" + timestamp + ".csv"
	f, err := os.Create(filename)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 获取数据失败，终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "write-to-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	writer := csv.NewWriter(f)
	writer.WriteAll(headers)

	writer.Flush() // 此时才会将缓冲区数据写入

	var items [][]string

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "write-to-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		var row []string

		//契約管理番号
		row = append(row, resp.ColDatas.GetKeiyakuno())
		//契約年月日
		if resp.ColDatas.GetKeiyakuymd()[0:10] == "0001-01-01" {
			row = append(row, "")
		} else {
			row = append(row, resp.ColDatas.GetKeiyakuymd()[0:10])
		}
		//リース開始年月日
		row = append(row, resp.ColDatas.GetLeasestymd()[0:10])
		//リース期間
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetLeasekikan()), 10))
		//リース満了年月日
		row = append(row, resp.ColDatas.GetLeaseexpireymd()[0:10])
		//延長オプション期間
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetExtentionoption()), 10))
		//契約名称
		row = append(row, resp.ColDatas.GetKeiyakunm())
		//備考1
		row = append(row, resp.ColDatas.GetBiko1())
		//初回支払年月日
		row = append(row, resp.ColDatas.GetPaymentstymd()[0:10])
		//支払サイクル
		if resp.ColDatas.GetPaymentcycle() == "1" {
			row = append(row, "毎月")
		} else {
			row = append(row, resp.ColDatas.GetPaymentcycle()+"ヶ月毎")
		}
		//支払日
		row = append(row, resp.ColDatas.GetPaymentday()+"日")
		//支払回数
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetPaymentcounts()), 10))
		//残価保証額
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetResidualvalue()), 10))
		//追加借入利子率
		row = append(row, resp.ColDatas.GetRishiritsu())
		//当初直接費用
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetInitialdirectcosts()), 10))
		//原状回復コスト
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetRestorationcosts()), 10))
		//使用権資産の計算方法
		if resp.ColDatas.GetSykshisankeisan() == "1" {
			row = append(row, "適用開始時点から計算")
		} else {
			row = append(row, "取得時点に遡って計算")
		}
		//管理部門
		row = append(row, resp.ColDatas.GetSegmentcd())
		//分類コード
		row = append(row, resp.ColDatas.GetBunruicd())
		//セグメント01
		row = append(row, resp.ColDatas.GetFieldViw())
		//セグメント02
		row = append(row, resp.ColDatas.GetField_22C())
		//セグメント03
		row = append(row, resp.ColDatas.GetField_1Av())
		//セグメント04
		row = append(row, resp.ColDatas.GetField_206())
		//セグメント05
		row = append(row, resp.ColDatas.GetField_14L())
		//任意マスタ01
		row = append(row, resp.ColDatas.GetField_7P3())
		//任意マスタ02
		row = append(row, resp.ColDatas.GetField_248())
		//任意マスタ03
		row = append(row, resp.ColDatas.GetField_3K7())
		//任意マスタ04
		row = append(row, resp.ColDatas.GetField_1Vg())
		//任意マスタ05
		row = append(row, resp.ColDatas.GetField_5Fj())
		//任意項目01
		row = append(row, resp.ColDatas.GetField_20H())
		//任意項目02
		row = append(row, resp.ColDatas.GetField_2H1())
		//任意項目03
		row = append(row, resp.ColDatas.GetFieldQi4())
		//任意項目04
		row = append(row, resp.ColDatas.GetField_1Ck())
		//任意項目05
		row = append(row, resp.ColDatas.GetFieldU1Q())
		//適用開始時点の残存リース料
		row = append(row, resp.ColDatas.GetHkkjitenzan())
		//利益剰余金
		row = append(row, resp.ColDatas.GetSonnekigaku())
		//年
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetYear()), 10))
		//月
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetMonth()), 10))
		//支払リース料
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetPaymentleasefee()), 10))
		//期首元本残高
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetFirstbalance()), 10))
		//元本返済相当額
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetRepayment()), 10))
		//当月末元本残高
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetBalance()), 10))
		//支払利息相当額
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetInterest()), 10))
		//使用権資産額期首簿価
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetBoka()), 10))
		//使用権資産償却費
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetSyokyaku()), 10))
		//使用権資産当月末帳簿価格
		row = append(row, strconv.FormatInt(int64(resp.ColDatas.GetEndboka()), 10))

		items = append(items, row)
	}

	defer f.Close()

	// 写入数据
	writer.WriteAll(items)
	writer.Flush() // 此时才会将缓冲区数据写入

	// 发送消息 写入文件成功，开始保存文档到文件服务器
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     i18n.Tr(lang, "job.J_029"),
		CurrentStep: "save-file",
		Database:    db,
	}, userID)

	fo, err := os.Open(filename)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 保存文件失败，终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	defer func() {
		fo.Close()
		os.Remove(filename)
	}()

	// 写入文件到 minio
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 保存文件失败，终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	appRoot := "app_" + appID
	filePath := path.Join(appRoot, "csv", "check_"+timestamp+".csv")
	path, err := minioClient.SavePublicObject(fo, filePath, "text/csv")
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 保存文件失败，终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "save-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	// 判断顾客上传文件是否在设置的最大存储空间以内
	// canUpload := filex.CheckCanUpload(domain, float64(path.Size))
	// if canUpload {
	// 	// 如果没有超出最大值，就对顾客的已使用大小进行累加
	// 	err = filex.ModifyUsedSize(domain, float64(path.Size))
	// 	if err != nil {
	// 		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
	// 		// 发送消息 保存文件失败，终止任务
	// 		jobx.ModifyTask(task.ModifyRequest{
	// 			JobId:       jobID,
	// 			Message:     err.Error(),
	// 			CurrentStep: "save-file",
	// 			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
	// 			ErrorFile: &task.File{
	// 				Url:  path.MediaLink,
	// 				Name: path.Name,
	// 			},
	// 			Database: db,
	// 		}, userID)
	// 		return
	// 	}
	// } else {
	// 	// 如果已达上限，则删除刚才上传的文件
	// 	minioClient.DeleteObject(path.Name)
	// 	path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
	// 	// 发送消息 保存文件失败，终止任务
	// 	jobx.ModifyTask(task.ModifyRequest{
	// 		JobId:       jobID,
	// 		Message:     i18n.Tr(lang, "job.J_007"),
	// 		CurrentStep: "save-file",
	// 		EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
	// 		ErrorFile: &task.File{
	// 			Url:  path.MediaLink,
	// 			Name: path.Name,
	// 		},
	// 		Database: db,
	// 	}, userID)
	// 	return
	// }

	// 发送消息 写入保存文件成功，返回下载路径，任务结束
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     i18n.Tr(lang, "job.J_028"),
		CurrentStep: "end",
		Progress:    100,
		File: &task.File{
			Url:  path.MediaLink,
			Name: path.Name,
		},
		EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
		Database: db,
	}, userID)

	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, "Coldata", "DownloadColDatas")),
		Data:    gin.H{},
	})

}
