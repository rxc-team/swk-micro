package webui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/journalx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
)

// Journal 分录
type Journal struct{}

// log出力使用
const (
	JournalProcessName   = "Journal"
	ActionFindJournals   = "FindJournals"
	ActionFindJournal    = "FindJournal"
	ActionJournalCompute = "JournalCompute"
	ActionImportJournal  = "ImportJournal"
	ActionModifyJournal  = "ModifyJournal"
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
