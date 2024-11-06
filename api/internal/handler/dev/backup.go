package dev

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/tplx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
)

// Backup 备份
type Backup struct{}

type DumpResult struct {
	HasData      bool
	Size         int64
	CopyInfoList []*backup.CopyInfo
	FileName     string
	FilePath     string
}

// log出力
const (
	BackupProcessName               = "Backup"
	ActionFindBackups               = "FindBackups"
	ActionFindBackup                = "FindBackup"
	ActionAddBackup                 = "AddBackup"
	ActionHardDeleteBackups         = "HardDeleteBackups"
	ActionHardDeleteTemplateBackups = "HardDeleteTemplateBackups"
)

// FindBackups 查找多个备份记录
// @Router /backups [get]
func (b *Backup) FindBackups(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindBackups, loggerx.MsgProcessStarted)

	backupService := backup.NewBackupService("manage", client.DefaultClient)

	var req backup.FindBackupsRequest
	req.CustomerId = c.Query("customer_id")
	req.BackupName = c.Query("backup_name")
	req.BackupType = c.Query("backup_type")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := backupService.FindBackups(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindBackups, err)
		return
	}

	loggerx.InfoLog(c, ActionFindBackups, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, BackupProcessName, ActionFindBackups)),
		Data:    response.GetBackups(),
	})
}

// FindBackup 查找单个备份记录
// @Router /backups/{backup_id} [get]
func (b *Backup) FindBackup(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindBackup, loggerx.MsgProcessStarted)

	backupService := backup.NewBackupService("manage", client.DefaultClient)

	var req backup.FindBackupRequest
	req.BackupId = c.Param("backup_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := backupService.FindBackup(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindBackup, err)
		return
	}

	loggerx.InfoLog(c, ActionFindBackup, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, BackupProcessName, ActionFindBackup)),
		Data:    response.GetBackup(),
	})
}

// AddBackup 添加单个备份记录
// @Router /backups [post]
func (b *Backup) AddBackup(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddApp, loggerx.MsgProcessStarted)

	var req backup.AddBackupRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddApp, err)
		return
	}
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	// 获取数据(共通)
	loggerx.InfoLog(c, ActionAddApp, fmt.Sprintf("Process FindCustomer:%s", loggerx.MsgProcessStarted))

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var cReq customer.FindCustomerRequest
	cReq.CustomerId = req.GetCustomerId()
	cResponse, err := customerService.FindCustomer(context.TODO(), &cReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddApp, err)
		return
	}

	loggerx.InfoLog(c, ActionAddApp, fmt.Sprintf("Process FindCustomer:%s", loggerx.MsgProcessEnded))

	domain := cResponse.GetCustomer().GetDomain()
	tplDb := req.GetCustomerId()
	tplApp := req.GetAppId()

	db := sessionx.GetUserCustomer(c)
	lang := sessionx.GetCurrentLanguage(c)
	userId := sessionx.GetAuthUserID(c)

	timestamp := strconv.Itoa(time.Now().Nanosecond())

	go tplx.Dump(db, tplApp, tplDb, domain, timestamp, lang, userId, req)

	loggerx.SuccessLog(c, ActionAddApp, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddBackup"))
	loggerx.InfoLog(c, ActionAddApp, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, BackupProcessName, ActionAddBackup)),
		Data:    gin.H{},
	})
}

// HardDeleteTemplateBackups 物理删除备份
// @Router /phydel/backups [delete]
func (b *Backup) HardDeleteTemplateBackups(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteTemplateBackups, loggerx.MsgProcessStarted)

	backupService := backup.NewBackupService("manage", client.DefaultClient)

	var req backup.HardDeleteBackupsRequest
	req.BackupIdList = c.QueryArray("backup_id_list")
	req.Database = sessionx.GetUserCustomer(c)

	var filepaths []string
	for _, bid := range req.BackupIdList {
		var req backup.FindBackupRequest
		req.BackupId = bid
		req.Database = sessionx.GetUserCustomer(c)
		response, err := backupService.FindBackup(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteTemplateBackups, err)
			return
		}

		filepaths = append(filepaths, response.GetBackup().GetFileName())

		// dev端实际只删app模板备份
		if response.GetBackup().GetBackupType() != "template" {
			c.JSON(403, gin.H{
				"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
			})
			c.Abort()
			return
		}
	}

	response, err := backupService.HardDeleteBackups(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteTemplateBackups, err)
		return
	}
	domin := sessionx.GetUserDomain(c)
	for _, fileName := range filepaths {
		err := filex.DeleteMinioTemplateBackups(domin, fileName)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteTemplateBackups, err)
			return
		}
	}
	loggerx.SuccessLog(c, ActionHardDeleteTemplateBackups, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionHardDeleteTemplateBackups))

	loggerx.InfoLog(c, ActionHardDeleteTemplateBackups, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, BackupProcessName, ActionHardDeleteTemplateBackups)),
		Data:    response,
	})
}
