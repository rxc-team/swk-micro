package dev

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/storage/proto/file"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// File 文件
type File struct{}

// log出力
const (
	FileProcessName              = "File"
	ActionFindFiles              = "FindFiles"
	ActionAddAvatarFile          = "AddAvatarFile"
	ActionAddFile                = "AddFile"
	ActionDownloadFile           = "DownloadFile"
	ActionHardDeleteFile         = "HardDeleteFile"
	ActionDeletePublicHeaderFile = "DeletePublicHeaderFile"
)

// FindFiles 获取所有文件
// @Router /files [get]
func (u *File) FindFiles(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindFiles, loggerx.MsgProcessStarted)

	fileService := file.NewFileService("storage", client.DefaultClient)

	var req file.FindFilesRequest
	folder := c.Param("fo_id")
	if folder == "public" {
		req.Type = 1
		req.FileName = c.Query("file_name")
		req.ContentType = c.Query("content_type")
	} else if folder == "company" {
		req.Type = 2
		req.Domain = sessionx.GetUserDomain(c)
		req.FileName = c.Query("file_name")
		req.ContentType = c.Query("content_type")
	} else if folder == "user" {
		req.Type = 3
		req.UserId = sessionx.GetAuthUserID(c)
		req.Domain = sessionx.GetUserDomain(c)
		req.FileName = c.Query("file_name")
		req.ContentType = c.Query("content_type")
	} else {
		req.Type = 0
		req.FolderId = folder
		req.FileName = c.Query("file_name")
		req.ContentType = c.Query("content_type")
		req.Domain = sessionx.GetUserDomain(c)
	}

	req.Database = sessionx.GetUserCustomer(c)

	if folder == "public" {
		req.Database = "system"
	}

	response, err := fileService.FindFiles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindFiles, err)
		return
	}

	loggerx.InfoLog(c, ActionFindFiles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FileProcessName, ActionFindFiles)),
		Data:    response.GetFileList(),
	})
}

// HeaderFileUpload 用户头像文件上传
// @Router /header/upload [post]
func (u *File) HeaderFileUpload(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddAvatarFile, loggerx.MsgProcessStarted)

	domain := sessionx.GetUserDomain(c)
	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		httpx.GinHTTPError(c, ActionAddAvatarFile, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("pic", files.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionAddAvatarFile, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "pic", files.Size) {
		httpx.GinHTTPError(c, ActionAddAvatarFile, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	f, err := files.Open()
	if err != nil {
		httpx.GinHTTPError(c, ActionAddAvatarFile, err)
		return
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddAvatarFile, err)
		return
	}

	dir := path.Join("header", files.Filename)
	result, err := minioClient.SavePublicObject(f, dir, files.Header.Get("content-type"))
	if err != nil {
		httpx.GinHTTPError(c, ActionAddAvatarFile, err)
		return
	}
	// 判断顾客上传文件是否在设置的最大存储空间以内
	canUpload := filex.CheckCanUpload(domain, float64(result.Size))
	if canUpload {
		// 如果没有超出最大值，就对顾客的已使用大小进行累加
		err = filex.ModifyUsedSize(domain, float64(result.Size))
		if err != nil {
			httpx.GinHTTPError(c, ActionAddAvatarFile, err)
			return
		}
	} else {
		// 如果已达上限，则删除刚才上传的文件
		minioClient.DeleteObject(result.Name)
		httpx.GinHTTPError(c, ActionAddAvatarFile, errors.New("最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"))
		return
	}

	loggerx.InfoLog(c, ActionAddAvatarFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, FileProcessName, ActionAddAvatarFile)),
		Data: gin.H{
			"url": result.MediaLink,
		},
	})
}

// UploadDev DEV端文档管理机能文件夹文档上传
// @Router /upload [post]
func (u *File) UploadDev(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddFile, loggerx.MsgProcessStarted)

	domain := sessionx.GetUserDomain(c)
	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		httpx.GinHTTPError(c, ActionAddFile, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("doc", files.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionAddFile, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "doc", files.Size) {
		httpx.GinHTTPError(c, ActionAddFile, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	f, err := files.Open()
	if err != nil {
		httpx.GinHTTPError(c, ActionAddFile, err)
		return
	}

	var req file.AddRequest
	// 获取上传文件类型
	folderId := c.Param("fo_id")
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddFile, err)
		return
	}

	// DEV端只能上传公共和公司文件夹
	if folderId != "public" && folderId != "company" {
		c.JSON(403, gin.H{
			"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
		})
		c.Abort()
		return
	}

	// 公开的话保存到公共路径下
	if folderId == "public" {
		dir := path.Join("document", files.Filename)
		result, err := minioClient.SaveObject(f, dir, files.Header.Get("content-type"))
		if err != nil {
			httpx.GinHTTPError(c, ActionAddFile, err)
			return
		}
		req.FolderId = "public"
		req.FileName = c.PostForm("file_name")
		req.ContentType = result.ContentType
		req.FilePath = result.MediaLink
		req.FileSize = result.Size
		req.Owners = []string{sessionx.GetAuthUserID(c)}
		req.ObjectName = result.Name
	} else if folderId == "company" {
		dir := path.Join("company", sessionx.GetUserEmail(c), files.Filename)
		result, err := minioClient.SaveObject(f, dir, files.Header.Get("content-type"))
		if err != nil {
			httpx.GinHTTPError(c, ActionAddFile, err)
			return
		}
		req.FolderId = "company"
		req.FileName = c.PostForm("file_name")
		req.ContentType = result.ContentType
		req.FilePath = result.MediaLink
		req.FileSize = result.Size
		req.Owners = []string{sessionx.GetUserDomain(c)}
		req.ObjectName = result.Name
	}

	fileService := file.NewFileService("storage", client.DefaultClient)

	// 获取当前文件的域
	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)
	// 保存上传信息到mongo中
	response, err := fileService.AddFile(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddFile, err)
		return
	}

	// 上传文件成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
	params["file_name"] = req.GetFileName()       // 新规的时候取传入参数

	loggerx.ProcessLog(c, ActionAddFile, msg.L002, params)

	loggerx.InfoLog(c, ActionAddFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, FileProcessName, ActionAddFile)),
		Data: gin.H{
			"url":     req.FilePath,
			"file_id": response.GetFileId(),
		},
	})
}

// Download 文件下载
// @Router /download/files/{file_id} [get]
func (u *File) Download(c *gin.Context) {
	loggerx.InfoLog(c, ActionDownloadFile, loggerx.MsgProcessStarted)

	// 获取文件
	fileService := file.NewFileService("storage", client.DefaultClient)
	var req file.FindFileRequest
	req.FileId = c.Param("file_id")
	req.Database = c.Query("database")
	domain := "proship.co.jp"
	if req.Database == "" {
		req.Database = sessionx.GetUserCustomer(c)
		domain = sessionx.GetUserDomain(c)
	}
	response, err := fileService.FindFile(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadFile, err)
		return
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadFile, err)
		return
	}

	object, err := minioClient.GetObject(response.GetFile().GetObjectName())
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadFile, err)
		return
	}
	var result []byte

	buffer := make([]byte, 1024)
	for {
		n, err := object.Read(buffer)
		result = append(result, buffer[:n]...)
		if err == io.EOF {
			break
		}
	}

	loggerx.InfoLog(c, ActionDownloadFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, FileProcessName, ActionDownloadFile)),
		Data: gin.H{
			"file_data": result,
			"file":      response.GetFile(),
		},
	})
}

// HardDeleteFile 删除某个文件夹中的某个文件
// @Router /folders/{fo_id}/files/{file_id} [delete]
func (u *File) HardDeleteFile(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteFile, loggerx.MsgProcessStarted)

	fileService := file.NewFileService("storage", client.DefaultClient)

	fileID := c.Param("file_id")
	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)

	// 查询删除文件名
	var freq file.FindFileRequest
	freq.FileId = fileID
	freq.Database = db
	fresponse, err := fileService.FindFile(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteFile, err)
		return
	}

	// 删除文件记录
	var req file.HardDeleteRequest
	req.FileId = fileID
	req.Database = db
	response, err := fileService.HardDeleteFile(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteFile, err)
		return
	}
	// 删除文件MINIO记录
	filex.DeleteFile(domain, fresponse.File.ObjectName)
	loggerx.SuccessLog(c, ActionHardDeleteFile, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionHardDeleteFile))

	// 删除文件成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
	params["file_name"] = fresponse.GetFile().GetFileName()

	loggerx.ProcessLog(c, ActionHardDeleteFile, msg.L001, params)

	loggerx.InfoLog(c, ActionHardDeleteFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, FileProcessName, ActionHardDeleteFile)),
		Data:    response,
	})
}

// DeletePublicHeaderFile 删除头像或LOGO文件
// @Router /public/header/file [delete]
func (u *File) DeletePublicHeaderFile(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeletePublicHeaderFile, loggerx.MsgProcessStarted)

	// 域名
	domain := sessionx.GetUserDomain(c)
	// 删除对象文件名
	delFileName := c.Query("file_name")
	delObj, err := filex.GetMinioHeaderInfo(domain, delFileName)
	if err != nil {
		loggerx.FailureLog(c, ActionDeletePublicHeaderFile, err.Error())
	}

	d, f, err := filex.DeletePublicHeaderFile(domain, delFileName)
	if err != nil {
		loggerx.FailureLog(c, ActionDeletePublicHeaderFile, err.Error())
	}
	//更新客户minio中的已使用的内存
	if domain != "proship.co.jp" {
		err = filex.ModifyUsedSize(domain, -float64(delObj.Size))
		if err != nil {
			httpx.GinHTTPError(c, ActionDeletePublicHeaderFile, err)
			return
		}
	}

	loggerx.SuccessLog(c, ActionDeletePublicHeaderFile, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeletePublicHeaderFile))

	loggerx.InfoLog(c, ActionDeletePublicHeaderFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, FileProcessName, ActionDeletePublicHeaderFile)),
		Data: gin.H{
			"domain":        d,
			"del_file_name": f,
		},
	})
}
