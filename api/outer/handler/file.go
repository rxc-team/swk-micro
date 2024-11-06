package handler

import (
	"errors"
	"fmt"
	"path"

	"github.com/gin-gonic/gin"

	"rxcsoft.cn/pit3/api/outer/common/filex"
	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// File 文件
type File struct{}

// log出力
const (
	FileProcessName              = "File"
	ActionFindFile               = "FindFile"
	ActionAddDataFile            = "AddDataFile"
	ActionAddAvatarFile          = "AddAvatarFile"
	ActionDeletePublicHeaderFile = "DeletePublicHeaderFile"
	ActionDeletePublicDataFile   = "DeletePublicDataFile"
	ActionDeletePublicDataFiles  = "DeletePublicDataFiles"
)

// ItemUpload 文件类型字段的文件上传
// @Summary 文件类型字段的文件上传
// @description 调用srv中的file服务，文件类型字段的文件上传
// @Tags File
// @Accept json
// @Security JWT
// @Produce  json
// @Param file body file.AddRequest true "文件信息"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /upload [post]
func (u *File) ItemUpload(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddDataFile, loggerx.MsgProcessStarted)

	domain := sessionx.GetUserDomain(c)

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDataFile, err)
		return
	}

	// 文件类型检查
	fileType := "doc"
	if c.PostForm("is_pic") == "true" {
		fileType = "pic"
	}
	if !filex.CheckSupport(fileType, files.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionAddDataFile, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, fileType, files.Size) {
		httpx.GinHTTPError(c, ActionAddDataFile, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	appID := sessionx.GetCurrentApp(c)
	appRoot := "app_" + appID
	datastoreID := c.PostForm("d_id")
	datastoreUrl := "datastore_" + datastoreID

	f, err := files.Open()
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDataFile, err)
		return
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDataFile, err)
		return
	}
	dir := path.Join(appRoot, "data", datastoreUrl, files.Filename)
	result, err := minioClient.SavePublicObject(f, dir, files.Header.Get("content-type"))
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDataFile, err)
		return
	}
	// 判断顾客上传文件是否在设置的最大存储空间以内
	canUpload := filex.CheckCanUpload(domain, float64(result.Size))
	if canUpload {
		// 如果没有超出最大值，就对顾客的已使用大小进行累加
		err = filex.ModifyUsedSize(domain, float64(result.Size))
		if err != nil {
			httpx.GinHTTPError(c, ActionAddDataFile, err)
			return
		}
	} else {
		// 如果已达上限，则删除刚才上传的文件
		minioClient.DeleteObject(result.Name)
		httpx.GinHTTPError(c, ActionAddDataFile, errors.New("最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"))
		return
	}
	loggerx.SuccessLog(c, ActionAddDataFile, fmt.Sprintf(loggerx.MsgProcesSucceed, "SavePublicObject"))

	loggerx.InfoLog(c, ActionAddDataFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, FileProcessName, ActionAddDataFile)),
		Data: gin.H{
			"url": result.MediaLink,
		},
	})
}

// HeaderFileUpload 用户头像文件上传
// @Summary 文件上传
// @description 调用srv中的file服务，文件上传
// @Tags File
// @Accept json
// @Security JWT
// @Produce  json
// @Param file body file.AddRequest true "文件信息"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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

	loggerx.SuccessLog(c, ActionAddAvatarFile, fmt.Sprintf(loggerx.MsgProcesSucceed, "SavePublicObject"))

	loggerx.InfoLog(c, ActionAddAvatarFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, FileProcessName, ActionAddAvatarFile)),
		Data: gin.H{
			"url": result.MediaLink,
		},
	})
}

// DeletePublicHeaderFile 删除头像或LOGO文件
// @Summary 删除头像或LOGO文件
// @Tags File
// @Accept json
// @Security JWT
// @Produce  json
// @Param file_name query string true "文件对象名"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /public/header/file [delete]
func (u *File) DeletePublicHeaderFile(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeletePublicHeaderFile, loggerx.MsgProcessStarted)

	// 域名
	domain := sessionx.GetUserDomain(c)
	// 删除对象文件名
	delFileName := c.Query("file_name")

	d, f, err := filex.DeletePublicHeaderFile(domain, delFileName)
	if err != nil {
		loggerx.FailureLog(c, ActionDeletePublicHeaderFile, err.Error())
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

// DeletePublicDataFile 删除文件类型字段数据的文件
// @Summary 删除文件类型字段数据的文件
// @Tags File
// @Accept json
// @Security JWT
// @Produce  json
// @Param file_name query string true "文件对象名"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /public/data/file [delete]
func (u *File) DeletePublicDataFile(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeletePublicDataFile, loggerx.MsgProcessStarted)

	// 域名
	domain := sessionx.GetUserDomain(c)
	// appID
	appID := sessionx.GetCurrentApp(c)
	// 删除对象文件名
	delFileName := c.Query("file_name")

	d, f, err := filex.DeletePublicDataFile(domain, appID, delFileName)
	if err != nil {
		loggerx.FailureLog(c, ActionDeletePublicDataFile, err.Error())
	}
	loggerx.SuccessLog(c, ActionDeletePublicDataFile, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeletePublicDataFile))

	loggerx.InfoLog(c, ActionDeletePublicDataFile, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, FileProcessName, ActionDeletePublicDataFile)),
		Data: gin.H{
			"domain":        d,
			"del_file_name": f,
		},
	})
}

// DeletePublicDataFiles 删除多个文件类型字段数据的文件
// @Summary 删除多个文件类型字段数据的文件
// @Tags File
// @Accept json
// @Security JWT
// @Produce  json
// @Param file_name_list query []string true "文件对象名集合"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /public/data/files [delete]
func (u *File) DeletePublicDataFiles(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeletePublicDataFiles, loggerx.MsgProcessStarted)

	// 域名
	domain := sessionx.GetUserDomain(c)
	// appID
	appID := sessionx.GetCurrentApp(c)
	// 删除对象文件名集合
	delFileNameList := c.QueryArray("file_name_list")

	d, fs, err := filex.DeletePublicDataFiles(domain, appID, delFileNameList)
	if err != nil {
		loggerx.FailureLog(c, ActionDeletePublicDataFiles, err.Error())
	}
	loggerx.SuccessLog(c, ActionDeletePublicDataFiles, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeletePublicDataFiles))

	loggerx.InfoLog(c, ActionDeletePublicDataFiles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, FileProcessName, ActionDeletePublicDataFiles)),
		Data: gin.H{
			"domain":         d,
			"del_file_names": fs,
		},
	})
}
