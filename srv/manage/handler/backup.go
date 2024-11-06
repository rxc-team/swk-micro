package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Backup Backup信息
type Backup struct{}

// log出力使用
const (
	BackupProcessName = "Backup"

	ActionFindBackups       = "FindBackups"
	ActionFindBackup        = "FindBackup"
	ActionAddBackup         = "AddBackup"
	ActionHardDeleteBackups = "HardDeleteBackups"
)

// FindBackups 查找多个备份记录
func (b *Backup) FindBackups(ctx context.Context, req *backup.FindBackupsRequest, rsp *backup.FindBackupsResponse) error {
	utils.InfoLog(ActionFindBackups, utils.MsgProcessStarted)
	// 默认查询
	bks, err := model.FindBackups(ctx, req.GetDatabase(), req.GetCustomerId(), req.GetBackupName(), req.GetBackupType())
	if err != nil {
		utils.ErrorLog(ActionFindBackups, err.Error())
		return err
	}
	res := &backup.FindBackupsResponse{}
	for _, bk := range bks {
		res.Backups = append(res.Backups, bk.ToProto())
	}
	*rsp = *res

	utils.InfoLog(ActionFindBackups, utils.MsgProcessEnded)
	return nil
}

// FindBackup 通过备份ID查找单个备份记录
func (b *Backup) FindBackup(ctx context.Context, req *backup.FindBackupRequest, rsp *backup.FindBackupResponse) error {
	utils.InfoLog(ActionFindBackup, utils.MsgProcessStarted)

	res, err := model.FindBackup(ctx, req.GetDatabase(), req.GetBackupId())
	if err != nil {
		utils.ErrorLog(ActionFindBackup, err.Error())
		return err
	}

	rsp.Backup = res.ToProto()

	utils.InfoLog(ActionFindBackup, utils.MsgProcessEnded)
	return nil
}

// AddBackup 添加单个备份记录
func (b *Backup) AddBackup(ctx context.Context, req *backup.AddBackupRequest, rsp *backup.AddBackupResponse) error {
	utils.InfoLog(ActionAddBackup, utils.MsgProcessStarted)

	var copys []*model.CopyInfo
	for _, c := range req.GetCopyInfoList() {
		cp := model.CopyInfo{
			CopyType: c.GetCopyType(),
			Source:   c.GetSource(),
			Count:    c.GetCount(),
		}
		copys = append(copys, &cp)
	}

	params := model.Backup{
		BackupName:    req.GetBackupName(),
		CustomerID:    req.GetCustomerId(),
		BackupType:    req.GetBackupType(),
		AppID:         req.GetAppId(),
		AppType:       req.GetAppType(),
		HasData:       req.GetHasData(),
		Size:          req.GetSize(),
		CopyInfoList:  copys,
		FileName:      req.GetFileName(),
		FilePath:      req.GetFilePath(),
		CloudFileName: req.GetCloudFileName(),
		CloudFilePath: req.GetCloudFilePath(),
		CreatedAt:     time.Now(),
		CreatedBy:     req.GetWriter(),
	}

	id, err := model.AddBackup(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddBackup, err.Error())
		return err
	}

	rsp.BackupId = id

	utils.InfoLog(ActionAddBackup, utils.MsgProcessEnded)

	return nil
}

// HardDeleteBackups 物理删除选中的备份记录
func (b *Backup) HardDeleteBackups(ctx context.Context, req *backup.HardDeleteBackupsRequest, rsp *backup.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteBackups, utils.MsgProcessStarted)

	err := model.HardDeleteBackups(ctx, req.Database, req.GetBackupIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteBackups, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteBackups, utils.MsgProcessEnded)
	return nil
}
