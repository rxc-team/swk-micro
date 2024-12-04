package webui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/configx"
	"rxcsoft.cn/pit3/api/internal/common/logic/journalx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/pit3/srv/task/proto/task"
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

	conditions := []*item.Condition{}
	conditions = append(conditions, &item.Condition{
		FieldId:     "keijoudate",
		FieldType:   "date",
		SearchValue: handleMonth + "-01",
		Operator:    ">=",
		IsDynamic:   true,
	})

	conditions = append(conditions, &item.Condition{
		FieldId:     "keijoudate",
		FieldType:   "date",
		SearchValue: handleMonth + "-" + lastDay,
		Operator:    "<=",
		IsDynamic:   true,
	})

	conditions = append(conditions, &item.Condition{
		FieldId:     "journalstatus",
		FieldType:   "text",
		SearchValue: "作成",
		Operator:    "=",
		IsDynamic:   true,
	})

	// 获取总的件数
	cReq := item.CountRequest{
		AppId:         appID,
		DatastoreId:   rirekiDate.DatastoreId,
		ConditionList: conditions,
		ConditionType: "and",
		Database:      db,
	}

	countResponse, err := itemService.FindCount(context.TODO(), &cReq, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSakuseiData, err)
		return
	}

	loggerx.InfoLog(c, ActionFindSakuseiData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionJournalCompute)),
		Data:    countResponse,
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
