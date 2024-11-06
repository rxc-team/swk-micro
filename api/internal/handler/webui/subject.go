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
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/journal/proto/subject"
)

// Subject 科目
type Subject struct{}

// log出力使用
const (
	SubjectProcessName  = "Subject"
	ActionFindSubjects  = "FindSubjects"
	ActionFindSubject   = "FindSubject"
	ActionImportSubject = "ImportSubject"
	ActionModifySubject = "ModifySubject"
	ActionDeleteSubject = "DeleteSubject"
)

// FindSubjects 获取当前用户的所有科目
// @Router /subjects [get]
func (f *Subject) FindSubjects(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindSubjects, loggerx.MsgProcessStarted)

	subjectService := subject.NewSubjectService("journal", client.DefaultClient)

	var req subject.SubjectsRequest
	// 从query获取
	assetstType := c.Query("assets_type")
	req.AssetsType = assetstType
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)

	response, err := subjectService.FindSubjects(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindSubjects, err)
		return
	}

	loggerx.InfoLog(c, ActionFindSubjects, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, SubjectProcessName, ActionFindSubjects)),
		Data:    response.GetSubjects(),
	})
}

// ImportSubject 添加科目
// @Router /subjects [post]
func (f *Subject) ImportSubject(c *gin.Context) {
	loggerx.InfoLog(c, ActionImportSubject, loggerx.MsgProcessStarted)
	subjectService := subject.NewSubjectService("journal", client.DefaultClient)

	domain := sessionx.GetUserDomain(c)
	timestamp := time.Now().Format("20060102150405")

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		httpx.GinHTTPError(c, ActionImportSubject, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionImportSubject, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "csv", files.Size) {
		httpx.GinHTTPError(c, ActionImportSubject, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	filename := "temp/subject" + "_" + timestamp + "_" + files.Filename
	if err := c.SaveUploadedFile(files, filename); err != nil {
		// 删除临时文件
		os.Remove(filename)
		httpx.GinHTTPError(c, ActionImportSubject, err)
		return
	}

	var subs []*subject.Subject

	// 读取journals数据
	err = filex.ReadFile(filename, &subs)
	if err != nil {
		// 删除临时文件
		os.Remove(filename)
		httpx.GinHTTPError(c, ActionImportSubject, err)
		return
	}

	// 删除临时文件
	os.Remove(filename)

	appID := sessionx.GetCurrentApp(c)

	for _, s := range subs {
		s.AppId = appID
	}

	var req subject.ImportRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionImportSubject, err)
		return
	}
	// 从共通中获取
	req.Subjects = subs
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := subjectService.ImportSubject(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionImportSubject, err)
		return
	}
	loggerx.SuccessLog(c, ActionImportSubject, "Subject import Success")

	loggerx.InfoLog(c, ActionImportSubject, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, SubjectProcessName, ActionImportSubject)),
		Data:    response,
	})
}

// ModifySubject 修改科目
// @Router /subjects/{s_key} [put]
func (f *Subject) ModifySubject(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifySubject, loggerx.MsgProcessStarted)

	journalService := subject.NewSubjectService("journal", client.DefaultClient)

	var req subject.ModifyRequest

	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifySubject, err)
		return
	}

	// 从共通中获取
	req.SubjectKey = c.Param("s_key")
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := journalService.ModifySubject(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifySubject, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifySubject, "Subject update success")

	loggerx.InfoLog(c, ActionModifySubject, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, JournalProcessName, ActionModifySubject)),
		Data:    response,
	})
}
