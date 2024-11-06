package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/storage/model"
	"rxcsoft.cn/pit3/srv/storage/proto/folder"
	"rxcsoft.cn/pit3/srv/storage/utils"
)

// Folder 文件信息
type Folder struct{}

// log出力使用
const (
	FolderProcessName = "Folder"

	ActionFindFolders          = "FindFolders"
	ActionFindFolder           = "FindFolder"
	ActionAddFolder            = "AddFolder"
	ActionModifyFolder         = "ModifyFolder"
	ActionDeleteFolder         = "DeleteFolder"
	ActionDeleteSelectFolders  = "DeleteSelectFolders"
	ActionHardDeleteFolders    = "HardDeleteFolders"
	ActionRecoverSelectFolders = "RecoverSelectFolders"
)

// FindFolders 获取所有文件夹
func (u *Folder) FindFolders(ctx context.Context, req *folder.FindFoldersRequest, rsp *folder.FindFoldersResponse) error {
	utils.InfoLog(ActionFindFolders, utils.MsgProcessStarted)

	folders, err := model.FindFolders(req.GetDatabase(), req.GetDomain(), req.GetFolderName())
	if err != nil {
		utils.ErrorLog(ActionFindFolders, err.Error())
		return err
	}

	res := &folder.FindFoldersResponse{}
	for _, u := range folders {
		res.FolderList = append(res.FolderList, u.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindFolders, utils.MsgProcessEnded)
	return nil
}

// FindFolder 通过ID获取文件夹
func (u *Folder) FindFolder(ctx context.Context, req *folder.FindFolderRequest, rsp *folder.FindFolderResponse) error {
	utils.InfoLog(ActionFindFolder, utils.MsgProcessStarted)

	res, err := model.FindFolder(req.GetDatabase(), req.GetFolderId())
	if err != nil {
		utils.ErrorLog(ActionFindFolder, err.Error())
		return err
	}

	rsp.Folder = res.ToProto()

	utils.InfoLog(ActionFindFolder, utils.MsgProcessEnded)
	return nil
}

// AddFolder 添加文件夹
func (u *Folder) AddFolder(ctx context.Context, req *folder.AddRequest, rsp *folder.AddResponse) error {
	utils.InfoLog(ActionAddFolder, utils.MsgProcessStarted)

	params := model.Folder{
		FolderName: req.GetFolderName(),
		FolderDir:  req.GetFolderDir(),
		Domain:     req.GetDomain(),
		CreatedAt:  time.Now(),
		CreatedBy:  req.GetWriter(),
		UpdatedAt:  time.Now(),
		UpdatedBy:  req.GetWriter(),
	}

	id, err := model.AddFolder(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddFolder, err.Error())
		return err
	}

	rsp.FolderId = id

	utils.InfoLog(ActionAddFolder, utils.MsgProcessEnded)

	return nil
}

// ModifyFolder 更新文件夹
func (u *Folder) ModifyFolder(ctx context.Context, req *folder.ModifyRequest, rsp *folder.ModifyResponse) error {
	utils.InfoLog(ActionModifyFolder, utils.MsgProcessStarted)

	err := model.ModifyFolder(req.GetDatabase(), req.GetFolderId(), req.GetFolderName(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionModifyFolder, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyFolder, utils.MsgProcessEnded)

	return nil
}

// DeleteFolder 删除文件夹
func (u *Folder) DeleteFolder(ctx context.Context, req *folder.DeleteRequest, rsp *folder.DeleteResponse) error {
	utils.InfoLog(ActionDeleteFolder, utils.MsgProcessStarted)

	err := model.DeleteFolder(req.GetDatabase(), req.GetFolderId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteFolder, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteFolder, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectFolders 删除多个文件夹
func (u *Folder) DeleteSelectFolders(ctx context.Context, req *folder.DeleteSelectFoldersRequest, rsp *folder.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectFolders, utils.MsgProcessStarted)

	_, err := model.DeleteSelectFolders(req.GetDatabase(), req.GetFolderIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectFolders, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectFolders, utils.MsgProcessEnded)
	return nil
}

// HardDeleteFolders 物理删除多个文件夹
func (u *Folder) HardDeleteFolders(ctx context.Context, req *folder.HardDeleteFoldersRequest, rsp *folder.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteFolders, utils.MsgProcessStarted)

	_, err := model.HardDeleteFolders(req.GetDatabase(), req.GetFolderIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteFolders, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteFolders, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectFolders 恢复选中文件夹
func (u *Folder) RecoverSelectFolders(ctx context.Context, req *folder.RecoverSelectFoldersRequest, rsp *folder.RecoverSelectFoldersResponse) error {
	utils.InfoLog(ActionRecoverSelectFolders, utils.MsgProcessStarted)

	_, err := model.RecoverSelectFolders(req.GetDatabase(), req.GetFolderIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectFolders, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectFolders, utils.MsgProcessEnded)
	return nil
}
