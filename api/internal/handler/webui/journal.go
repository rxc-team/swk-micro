package webui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/configx"
	"rxcsoft.cn/pit3/api/internal/common/logic/journalx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// Journal 分录
type Journal struct{}

// log出力使用
const (
	JournalProcessName        = "Journal"
	ActionFindJournals        = "FindJournals"
	ActionFindJournal         = "FindJournal"
	ActionJournalCompute      = "JournalCompute"
	ActionImportJournal       = "ImportJournal"
	ActionModifyJournal       = "ModifyJournal"
	ActionJournalConfim       = "JournalConfim"
	ActionFindSakuseiData     = "FindSakuseiData"
	ActionAddDownloadSetting  = "AddDownloadSetting"
	ActionFindDownloadSetting = "FindDownloadSetting"
	ActionSwkDownload         = "SwkDownload"
)

// FindJournals 获取当前用户的所有分录
// @Router /journals [get]
func (f *Journal) FindJournals(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindJournals, loggerx.MsgProcessStarted)

	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.JournalsRequest
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)

	response, err := journalService.FindJournals(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindJournals, err)
		return
	}

	loggerx.InfoLog(c, ActionFindJournals, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionFindJournals)),
		Data:    response.GetJournals(),
	})
}

// ImportJournal 导入分录
// @Router /journals [post]
func (f *Journal) ImportJournal(c *gin.Context) {
	loggerx.InfoLog(c, ActionImportJournal, loggerx.MsgProcessStarted)
	domain := sessionx.GetUserDomain(c)

	timestamp := time.Now().Format("20060102150405")

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		httpx.GinHTTPError(c, ActionImportJournal, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionImportJournal, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "csv", files.Size) {
		httpx.GinHTTPError(c, ActionImportJournal, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	filename := "temp/journal" + "_" + timestamp + "_" + files.Filename
	if err := c.SaveUploadedFile(files, filename); err != nil {
		// 删除临时文件
		os.Remove(filename)
		httpx.GinHTTPError(c, ActionImportJournal, err)
		return
	}

	var journals []*journal.Journal

	// 读取journals数据
	err = filex.ReadFile(filename, &journals)
	if err != nil {
		// 删除临时文件
		os.Remove(filename)
		httpx.GinHTTPError(c, ActionImportJournal, err)
		return
	}

	// 删除临时文件
	os.Remove(filename)

	appID := sessionx.GetCurrentApp(c)

	for _, j := range journals {
		j.AppId = appID
	}

	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.ImportRequest
	// 从共通中获取
	req.Journals = journals
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := journalService.ImportJournal(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionImportJournal, err)
		return
	}
	loggerx.SuccessLog(c, ActionImportJournal, "Journal import success")

	loggerx.InfoLog(c, ActionImportJournal, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionImportJournal)),
		Data:    response,
	})
}

// ModifyJournal 修改分录
// @Router /journals/{j_id} [put]
func (f *Journal) ModifyJournal(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyJournal, loggerx.MsgProcessStarted)

	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.ModifyRequest

	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyJournal, err)
		return
	}

	// 从共通中获取
	req.JournalId = c.Param("j_id")
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := journalService.ModifyJournal(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyJournal, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyJournal, "Journal update success")

	loggerx.InfoLog(c, ActionModifyJournal, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionModifyJournal)),
		Data:    response,
	})
}

// JournalCompute 分录计算
// @Router /journals/{j_id} [get]
func (f *Journal) JournalCompute(c *gin.Context) {
	loggerx.InfoLog(c, ActionJournalCompute, loggerx.MsgProcessStarted)

	// 契约处理类型区分
	section := c.Query("section")
	domain := sessionx.GetUserDomain(c)
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	lang := sessionx.GetCurrentLanguage(c)
	owners := sessionx.GetUserOwner(c)

	if section == "repay" {
		result, err := journalx.GenRepayData(domain, db, appID, userID, lang, owners)
		if err != nil {
			httpx.GinHTTPError(c, ActionJournalCompute, err)
			return
		}
		loggerx.InfoLog(c, ActionJournalCompute, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionJournalCompute)),
			Data:    result,
		})
		return
	}
	if section == "pay" {
		result, err := journalx.GenPayData(domain, db, appID, userID, lang, owners)
		if err != nil {
			httpx.GinHTTPError(c, ActionJournalCompute, err)
			return
		}
		loggerx.InfoLog(c, ActionJournalCompute, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionJournalCompute)),
			Data:    result,
		})
		return
	}

	result, err := journalx.GenAddAndSubData(domain, db, appID, userID, lang, owners)
	if err != nil {
		httpx.GinHTTPError(c, ActionJournalCompute, err)
		return
	}
	loggerx.InfoLog(c, ActionJournalCompute, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionJournalCompute)),
		Data:    result,
	})
}

// FindSakuseiData 查找分录做成数据
// @Router /journals/findSakuseiData [get]
func (f *Journal) FindSakuseiData(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindSakuseiData, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)

	// 通过apikey获取增减履历台账情报
	rirekiDate, err := getDatastoreInfo(db, appID, "zougenrireki")
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSakuseiData, err)
		return
	}

	// 通过apikey获取偿却台账情报
	repaymentDate, err := getDatastoreInfo(db, appID, "repayment")
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSakuseiData, err)
		return
	}

	// 获取处理月度
	cfg, err := configx.GetConfigVal(db, appID)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSakuseiData, err)
		return
	}
	handleMonth := cfg.GetSyoriYm()

	itemService := item.NewItemService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	handleDate, err := time.Parse("2006-01", handleMonth)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSakuseiData, err)
		return
	}

	lastDay := getMonthLastDay(handleDate)
	defaultTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

	rirekiConditions := []*item.Condition{}
	rirekiConditions = append(rirekiConditions, &item.Condition{
		FieldId:     "keijoudate",
		FieldType:   "date",
		SearchValue: handleMonth + "-01",
		Operator:    ">=",
		IsDynamic:   true,
	})

	rirekiConditions = append(rirekiConditions, &item.Condition{
		FieldId:     "keijoudate",
		FieldType:   "date",
		SearchValue: handleMonth + "-" + lastDay,
		Operator:    "<=",
		IsDynamic:   true,
	})

	rirekiConditions = append(rirekiConditions, &item.Condition{
		FieldId:     "sakuseidate",
		FieldType:   "date",
		SearchValue: defaultTime.Format(time.RFC3339),
		Operator:    "<>",
		IsDynamic:   true,
	})

	rirekiConditions = append(rirekiConditions, &item.Condition{
		FieldId:     "kakuteidate",
		FieldType:   "date",
		SearchValue: defaultTime.Format(time.RFC3339),
		Operator:    "=",
		IsDynamic:   true,
	})

	// 获取总的件数
	rirekiReq := item.CountRequest{
		AppId:         appID,
		DatastoreId:   rirekiDate.DatastoreId,
		ConditionList: rirekiConditions,
		ConditionType: "and",
		Database:      db,
	}

	rirekiCountResponse, err := itemService.FindCount(context.TODO(), &rirekiReq, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSakuseiData, err)
		return
	}

	repaymentConditions := []*item.Condition{}
	repaymentConditions = append(repaymentConditions, &item.Condition{
		FieldId:     "syokyakuymd",
		FieldType:   "date",
		SearchValue: handleMonth + "-01",
		Operator:    ">=",
		IsDynamic:   true,
	})

	repaymentConditions = append(repaymentConditions, &item.Condition{
		FieldId:     "syokyakuymd",
		FieldType:   "date",
		SearchValue: handleMonth + "-" + lastDay,
		Operator:    "<=",
		IsDynamic:   true,
	})

	repaymentConditions = append(repaymentConditions, &item.Condition{
		FieldId:     "sakuseidate",
		FieldType:   "date",
		SearchValue: defaultTime.Format(time.RFC3339),
		Operator:    "<>",
		IsDynamic:   true,
	})

	repaymentConditions = append(repaymentConditions, &item.Condition{
		FieldId:     "kakuteidate",
		FieldType:   "date",
		SearchValue: defaultTime.Format(time.RFC3339),
		Operator:    "=",
		IsDynamic:   true,
	})

	// 获取总的件数
	repaymentReq := item.CountRequest{
		AppId:         appID,
		DatastoreId:   repaymentDate.DatastoreId,
		ConditionList: repaymentConditions,
		ConditionType: "and",
		Database:      db,
	}

	repaymentCountResponse, err := itemService.FindCount(context.TODO(), &repaymentReq, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSakuseiData, err)
		return
	}

	loggerx.InfoLog(c, ActionFindSakuseiData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionJournalCompute)),
		Data: gin.H{
			"rirekiTotal":    rirekiCountResponse.GetTotal(),
			"repaymentTotal": repaymentCountResponse.GetTotal(),
		},
	})

}

// JournalConfim 分录确定
// @Router /journals/confim [get]
func (f *Journal) JournalConfim(c *gin.Context) {
	loggerx.InfoLog(c, ActionJournalConfim, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	jobID := "job_" + time.Now().Format("20060102150405")
	var datastoreIDs []string

	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "Journal Confim",
		Origin:       "-",
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "journal",
		Steps:        []string{"start", "collect-data", "journal-confim", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	// 发送消息 收集数据情报
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     i18n.Tr(lang, "job.J_002"),
		CurrentStep: "collect-data",
		Database:    db,
	}, userID)

	// 获取台账map
	dsMap, err := getDatastoreMap(db, appID)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 收集数据情报失败 终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "collect-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		return
	}
	datastoreIDs = append(datastoreIDs, dsMap["zougenrireki"])
	datastoreIDs = append(datastoreIDs, dsMap["repayment"])
	datastoreIDs = append(datastoreIDs, dsMap["shiwake"])

	// 获取处理月度
	cfg, err := configx.GetConfigVal(db, appID)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 收集数据情报失败 终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "collect-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		return
	}
	handleMonth := cfg.GetSyoriYm()

	handleDate, err := time.Parse("2006-01", handleMonth)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 收集数据情报失败 终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "collect-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		return
	}
	lastDay := getMonthLastDay(handleDate)

	itemService := item.NewItemService("database", client.DefaultClient)

	for _, datastoreID := range datastoreIDs {
		var req item.JournalRequest
		req.DatastoreId = datastoreID
		req.Database = db
		req.StartDate = handleMonth + "-01"
		req.LastDate = handleMonth + "-" + lastDay

		_, err = itemService.ConfimItem(context.TODO(), &req)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "journal-confim",
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

	// 发送消息 任务成功结束
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     i18n.Tr(lang, "job.J_028"),
		CurrentStep: "end",
		EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
		Database:    db,
	}, userID)

	loggerx.InfoLog(c, ActionJournalConfim, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionJournalConfim)),
		Data:    nil,
	})
}

// AddDatastoreMapping 添加分录下载设置
// @Router /download/setting[post]
func (f *Journal) AddDownloadSetting(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddDownloadSetting, loggerx.MsgProcessStarted)

	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.AddDownloadSettingRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddDownloadSetting, err)
		return
	}
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := journalService.AddDownloadSetting(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDownloadSetting, err)

		return
	}
	loggerx.SuccessLog(c, ActionAddDownloadSetting, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddDownloadSetting))

	loggerx.InfoLog(c, ActionAddDownloadSetting, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionAddDownloadSetting)),
		Data:    response,
	})
}

// FindDownloadSetting 查询分录下载设置
// @Router download/find[GET]
func (f *Journal) FindDownloadSetting(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDownloadSetting, loggerx.MsgProcessStarted)

	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.FindDownloadSettingRequest

	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := journalService.FindDownloadSetting(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDownloadSetting, err)

		return
	}
	loggerx.SuccessLog(c, ActionFindDownloadSetting, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionFindDownloadSetting))

	loggerx.InfoLog(c, ActionFindDownloadSetting, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDownloadSetting)),
		Data:    response,
	})
}

// FindDownloadSetting 查询分录下载设置
// @Router download/find[GET]
func (f *Journal) FindDownloadSettings(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDownloadSetting, loggerx.MsgProcessStarted)

	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.FindDownloadSettingsRequest

	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := journalService.FindDownloadSettings(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDownloadSetting, err)

		return
	}
	loggerx.SuccessLog(c, ActionFindDownloadSetting, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionFindDownloadSetting))

	loggerx.InfoLog(c, ActionFindDownloadSetting, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDownloadSetting)),
		Data:    response,
	})
}

// 分录下载
func (f *Journal) SwkDownload(c *gin.Context) {

	loggerx.InfoLog(c, ActionSwkDownload, loggerx.MsgProcessStarted)

	jobID := "job_" + time.Now().Format("20060102150405")
	appID := sessionx.GetCurrentApp(c)
	userID := sessionx.GetAuthUserID(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	db := sessionx.GetUserCustomer(c)
	appRoot := "app_" + appID

	go func() {

		// 创建任务
		jobx.CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      "download(swk)",
			Origin:       "仕訳台帳",
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(lang, "job.J_014"),
			TaskType:     "swk-download",
			Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
		var ds datastore.DatastoreKeyRequest

		ds.ApiKey = "shiwake"
		ds.AppId = appID
		ds.Database = db
		response, err := datastoreService.FindDatastoreByKey(context.TODO(), &ds)
		datastoreID := response.Datastore.DatastoreId
		owners := sessionx.GetUserAccessKeys(c, datastoreID, "R")
		if err != nil {
			httpx.GinHTTPError(c, ActionFindDatastore, err)
			return
		}

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		cReq := item.CountRequest{
			AppId:         appID,
			DatastoreId:   datastoreID,
			ConditionType: "and",
			Owners:        owners,
			Database:      db,
		}

		cResp, err := itemService.FindCount(context.TODO(), &cReq, opss)
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

		journalService := journal.NewJournalService("journal", client.DefaultClient)
		var req journal.FindDownloadSettingRequest

		// 从共通获取
		req.AppId = appID
		req.Database = db

		downloadInfo, err := journalService.FindDownloadSetting(context.TODO(), &req)
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
		encoding := downloadInfo.CharEncoding

		dReq := item.DownloadRequest{
			AppId:         appID,
			DatastoreId:   datastoreID,
			ConditionType: "and",
			Owners:        owners,
			Database:      db,
		}

		stream, err := itemService.SwkDownload(context.TODO(), &dReq, opss)
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

		// 获取当前台账的字段数据
		var fields []*typesx.DownloadField

		// 获取当前app的语言数据
		langData := langx.GetLanguageData(db, lang, domain)

		for _, rule := range downloadInfo.FieldRule {
			if rule.SettingMethod == "1" {
				fields = append(fields, &typesx.DownloadField{
					FieldName: rule.DownloadName,
					FieldType: "text",
					FieldId:   "#",
					Prefix:    rule.EditContent,
				})
			} else {
				fields = append(fields, &typesx.DownloadField{
					FieldName: rule.DownloadName,
					FieldType: rule.FieldType,
					FieldId:   rule.FieldId,
					Prefix:    "",
					Format:    rule.Format,
				})
			}
		}

		// 排序
		sort.Sort(typesx.DownloadFields(fields))

		timestamp := time.Now().Format("20060102150405")

		// 每次2000为一组数据
		total := cResp.GetTotal()

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_033"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		// 设定csv头部
		var header []string
		var headers [][]string
		for _, fl := range fields {
			header = append(header, fl.FieldName)
		}
		headers = append(headers, header)
		var writer *csv.Writer
		var delimiter string
		if downloadInfo.SeparatorChar == "separatorCharTab" {
			// 如果是Tab，则使用制表符
			delimiter = "\t"
		}

		filex.Mkdir("temp/")

		// 写入文件到本地
		filename := "temp/tmp" + "_" + timestamp + "_header" + ".csv"
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

		//是否有存在标题行
		if downloadInfo.HeaderRow == "exsit" {
			if encoding == "Shift-JIS" {
				converter := transform.NewWriter(f, japanese.ShiftJIS.NewEncoder())
				writer = csv.NewWriter(converter)
			} else {
				writer = csv.NewWriter(f)
				// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
				headers[0][0] = "\xEF\xBB\xBF" + headers[0][0]
			}
			if downloadInfo.SeparatorChar == "separatorCharTab" {
				writer.Comma = rune(delimiter[0])
			}
			err = writer.WriteAll(headers)
			if err != nil {
				if err.Error() == "encoding: rune not supported by encoding." {
					path := filex.WriteAndSaveFile(domain, appID, []string{"現在のタイトルには、日本語の[shift-jis]エンコード以外の文字が含まれており、実行を続行できません。"})
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
			writer.Flush() // 此时才会将缓冲区数据写入
		}
		writer = csv.NewWriter(f)

		var current int = 0
		var items [][]string

		for {
			it, err := stream.Recv()
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
					if downloadInfo.SeparatorChar == "separatorCharTab" {
						writer.Comma = rune(delimiter[0])
					}
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
			current++
			dt := it.GetItem()

			{
				// 设置csv行
				var itemData []string
				for _, fl := range fields {
					// 使用默认的值的情况
					if fl.FieldId == "#" {
						itemData = append(itemData, fl.Prefix)
					} else {
						itemMap := dt.GetItems()
						if value, ok := itemMap[fl.FieldId]; ok {
							result := ""
							switch value.DataType {
							case "text", "textarea", "number", "time", "switch":
								result = value.GetValue()
							case "autonum":
								result = value.GetValue()
							case "lookup":
								result = value.GetValue()
							case "options":
								result = langx.GetLangValue(langData, value.GetValue(), langx.DefaultResult)
							case "date":
								if value.GetValue() == "0001-01-01" {
									result = ""
								} else {
									if len(fl.Format) > 0 {
										date, err := time.Parse("2006-01-02", value.GetValue())
										if err != nil {
											result = ""
										} else {
											result = date.Format(fl.Format)
										}
									} else {
										result = value.GetValue()
									}
								}
							case "user":
								var userStrList []string
								json.Unmarshal([]byte(value.GetValue()), &userStrList)
								result = strings.Join(userStrList, ",")
							case "file":
								var files []typesx.FileValue
								json.Unmarshal([]byte(value.GetValue()), &files)
								var fileStrList []string
								for _, f := range files {
									fileStrList = append(fileStrList, f.Name)
								}
								result = strings.Join(fileStrList, ",")
							case "function":
								result = value.GetValue()
							default:
								break
							}

							itemData = append(itemData, result)
						} else {
							itemData = append(itemData, "")
						}
					}
				}
				// 添加行
				items = append(items, itemData)

			}

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
				if downloadInfo.SeparatorChar == "separatorCharTab" {
					writer.Comma = rune(delimiter[0])
				}
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
			Message:     i18n.Tr(lang, "job.J_029"),
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_043"),
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
		filePath := path.Join(appRoot, "csv", "datastore_"+timestamp+".csv")
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
				Message:     i18n.Tr(lang, "job.J_007"),
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
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			File: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database: db,
		}, userID)

	}()

	loggerx.InfoLog(c, ActionSwkDownload, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, MappingProcessName, ActionSwkDownload)),
		Data:    gin.H{},
	})
}

// getMonthLastDay  获取当前月份的最后一天
func getMonthLastDay(date time.Time) (day string) {
	// 年月日取得
	years := date.Year()
	month := date.Month()

	// 月末日取得
	lastday := 0
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			lastday = 30
		} else {
			lastday = 31
		}
	} else {
		if ((years%4) == 0 && (years%100) != 0) || (years%400) == 0 {
			lastday = 29
		} else {
			lastday = 28
		}
	}

	return strconv.Itoa(lastday)
}

// getDatastoreMap 获取台账apikey和datastore_id的map
func getDatastoreMap(db, appID string) (dsMap map[string]string, err error) {
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoresRequest
	// 从共通获取
	req.Database = db
	req.AppId = appID

	response, err := datastoreService.FindDatastores(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getDatastoreMap", err.Error())
		return
	}

	dsMap = make(map[string]string)

	for _, ds := range response.GetDatastores() {
		dsMap[ds.ApiKey] = ds.GetDatastoreId()
	}

	return
}
