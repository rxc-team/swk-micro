package handler

import (
	"context"
	"fmt"
	"time"

	"rxcsoft.cn/pit3/srv/storage/model"
	"rxcsoft.cn/pit3/srv/storage/proto/file"
	"rxcsoft.cn/pit3/srv/storage/utils"
)

// File 文件信息
type File struct{}

// log出力使用
const (
	Temp        = "%v.%v"
	TimeTemp    = "%v"
	ProcessName = "File"

	ActionFindFiles          = "FindFiles"
	ActionFindFile           = "FindFile"
	ActionAddFile            = "AddFile"
	ActionDeleteFile         = "DeleteFile"
	ActionHardDeleteFile     = "HardDeleteFile"
	ActionDeleteSelectFiles  = "DeleteSelectFiles"
	ActionHardDeleteFiles    = "HardDeleteFiles"
	ActionDeleteFolderFile   = "DeleteFolderFile"
	ActionRecoverSelectFiles = "RecoverSelectFiles"
	ActionRecoverFolderFiles = "RecoverFolderFiles"
)

// FindFiles 获取所有文件
func (u *File) FindFiles(ctx context.Context, req *file.FindFilesRequest, rsp *file.FindFilesResponse) error {
	utils.InfoLog(ActionFindFiles, utils.MsgProcessStarted)

	var files []model.File
	var err error

	utils.DebugLog(ActionFindFiles, fmt.Sprintf("Type: [ %d ]", req.Type))
	// 查询公开文件的场合
	if req.Type == 1 {
		files, err = model.FindPublicFiles(req.GetDatabase(), req.GetFileName(), req.GetContentType())
	} else if req.Type == 2 {
		// 查询公司文件的场合
		files, err = model.FindCompanyFiles(req.GetDatabase(), req.GetDomain(), req.GetFileName(), req.GetContentType())
	} else if req.GetType() == 3 {
		// 查询个人文件的场合
		files, err = model.FindUserFiles(req.GetDatabase(), req.GetUserId(), req.GetDomain(), req.GetFileName(), req.GetContentType())
	} else {
		// 查询普通文件夹文件的场合
		files, err = model.FindFiles(req.GetDatabase(), req.GetFolderId(), req.GetDomain(), req.GetFileName(), req.GetContentType())
	}

	// 出现错误的场合，直接返回错误
	if err != nil {
		utils.ErrorLog(ActionFindFiles, err.Error())
		return err
	}

	res := &file.FindFilesResponse{}
	for _, u := range files {
		res.FileList = append(res.FileList, u.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindFiles, utils.MsgProcessEnded)
	return nil
}

// FindFile 通过ID获取文件
func (u *File) FindFile(ctx context.Context, req *file.FindFileRequest, rsp *file.FindFileResponse) error {
	utils.InfoLog(ActionFindFile, utils.MsgProcessStarted)

	res, err := model.FindFile(req.GetDatabase(), req.GetFileId())
	if err != nil {
		utils.ErrorLog(ActionFindFile, err.Error())
		return err
	}

	rsp.File = res.ToProto()

	utils.InfoLog(ActionFindFile, utils.MsgProcessEnded)
	return nil
}

// AddFile 添加文件
func (u *File) AddFile(ctx context.Context, req *file.AddRequest, rsp *file.AddResponse) error {
	utils.InfoLog(ActionAddFile, utils.MsgProcessStarted)

	params := model.File{
		FolderID:    req.GetFolderId(),
		FileName:    req.GetFileName(),
		ObjectName:  req.GetObjectName(),
		FilePath:    req.GetFilePath(),
		FileSize:    req.GetFileSize(),
		ContentType: req.GetContentType(),
		Owners:      req.GetOwners(),
		Domain:      req.GetDomain(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	id, err := model.AddFile(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddFile, err.Error())
		return err
	}

	rsp.FileId = id

	utils.InfoLog(ActionAddFile, utils.MsgProcessEnded)

	return nil
}

// DeleteFile 删除单个文件
func (u *File) DeleteFile(ctx context.Context, req *file.DeleteRequest, rsp *file.DeleteResponse) error {
	utils.InfoLog(ActionDeleteFile, utils.MsgProcessStarted)

	err := model.DeleteFile(req.GetDatabase(), req.GetFileId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteFile, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteFile, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectFiles 删除多个文件
func (u *File) DeleteSelectFiles(ctx context.Context, req *file.DeleteSelectFilesRequest, rsp *file.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSelectFiles, utils.MsgProcessStarted)

	_, err := model.DeleteSelectFiles(req.GetDatabase(), req.GetFileIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectFiles, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectFiles, utils.MsgProcessEnded)
	return nil
}

// HardDeleteFile 硬删除单个文件
func (u *File) HardDeleteFile(ctx context.Context, req *file.HardDeleteRequest, rsp *file.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteFile, utils.MsgProcessStarted)

	err := model.HardDeleteFile(req.GetDatabase(), req.GetFileId())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteFile, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteFile, utils.MsgProcessEnded)
	return nil
}

// HardDeleteFiles 物理删除多个文件
func (u *File) HardDeleteFiles(ctx context.Context, req *file.HardDeleteFilesRequest, rsp *file.DeleteResponse) error {
	utils.InfoLog(ActionHardDeleteFiles, utils.MsgProcessStarted)

	_, err := model.HardDeleteFiles(req.GetDatabase(), req.GetFileIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteFiles, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteFiles, utils.MsgProcessEnded)
	return nil
}

// DeleteFolderFile 删除文件夹文件
func (u *File) DeleteFolderFile(ctx context.Context, req *file.DeleteFolderRequest, rsp *file.DeleteResponse) error {
	utils.InfoLog(ActionDeleteFolderFile, utils.MsgProcessStarted)

	_, err := model.DeleteFolderFile(req.GetDatabase(), req.GetFolderId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteFolderFile, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteFolderFile, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectFiles 恢复选中文件
func (u *File) RecoverSelectFiles(ctx context.Context, req *file.RecoverSelectFilesRequest, rsp *file.RecoverSelectFilesResponse) error {
	utils.InfoLog(ActionRecoverSelectFiles, utils.MsgProcessStarted)

	_, err := model.RecoverSelectFiles(req.GetDatabase(), req.GetFileIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectFiles, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectFiles, utils.MsgProcessEnded)
	return nil
}

// RecoverFolderFiles 恢复文件夹文件
func (u *File) RecoverFolderFiles(ctx context.Context, req *file.RecoverFolderFilesRequest, rsp *file.RecoverFolderFilesResponse) error {
	utils.InfoLog(ActionRecoverFolderFiles, utils.MsgProcessStarted)

	_, err := model.RecoverFolderFiles(req.GetDatabase(), req.GetFolderId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverFolderFiles, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverFolderFiles, utils.MsgProcessEnded)
	return nil
}
