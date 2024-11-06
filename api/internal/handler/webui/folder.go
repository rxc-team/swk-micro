package webui

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/storage/proto/folder"
)

// Folder 文件夹
type Folder struct{}

// log出力
const (
	FolderProcessName          = "Folder"
	ActionFindFolders          = "FindFolders"
	ActionFindFolder           = "FindFolder"
	ActionAddFolder            = "AddFolder"
	ActionDeleteFolder         = "DeleteFolder"
	ActionDeleteSelectFolders  = "DeleteSelectFolders"
	ActionHardDeleteFolders    = "HardDeleteFolders"
	ActionRecoverSelectFolders = "RecoverSelectFolders"
)

// FindFolders 获取所有文件夹
// @Router /folders [get]
func (u *Folder) FindFolders(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindFolders, loggerx.MsgProcessStarted)

	folderService := folder.NewFolderService("storage", client.DefaultClient)

	var req folder.FindFoldersRequest
	// 从query中获取参数
	req.FolderName = c.Query("folder_name")
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := folderService.FindFolders(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindFolders, err)
		return
	}
	folders := make([]map[string]interface{}, len(response.GetFolderList()))

	roles := sessionx.GetUserRoles(c)

	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	var preq permission.FindActionsRequest
	preq.RoleId = roles
	preq.PermissionType = "common"
	preq.ActionType = "folder"
	preq.Database = sessionx.GetUserCustomer(c)
	pResp, err := pmService.FindActions(context.TODO(), &preq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindFolders, err)
		return
	}

	// 文件夹列表
	folderList := response.GetFolderList()

	for index, folder := range folderList {
		if len(folders[index]) == 7 {
			for _, act := range pResp.GetActions() {
				if act.GetObjectId() == folder.GetFolderId() {
					if act.ActionMap["read"] {
						folders[index]["read"] = true
					}
					if act.ActionMap["write"] {
						folders[index]["write"] = true
					}
					if act.ActionMap["delete"] {
						folders[index]["delete"] = true
					}
					break
				}
			}
		} else {
			folders[index] = make(map[string]interface{}, 7)
			folders[index]["folder_id"] = folder.GetFolderId()
			folders[index]["folder_name"] = folder.GetFolderName()
			folders[index]["folder_dir"] = folder.GetFolderDir()
			folders[index]["domain"] = folder.GetDomain()
			// 默认操作权限都为flase
			folders[index]["read"] = false
			folders[index]["write"] = false
			folders[index]["delete"] = false

			for _, act := range pResp.GetActions() {
				if act.GetObjectId() == folder.GetFolderId() {
					if act.ActionMap["read"] {
						folders[index]["read"] = true
					}
					if act.ActionMap["write"] {
						folders[index]["write"] = true
					}
					if act.ActionMap["delete"] {
						folders[index]["delete"] = true
					}
					break
				}
			}
		}

	}

	loggerx.InfoLog(c, ActionFindFolders, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastores)),
		Data:    folders,
	})
}

// AddFolder 文件夹添加
// @Router /folders [post]
func (u *Folder) AddFolder(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddFolder, loggerx.MsgProcessStarted)

	folderService := folder.NewFolderService("storage", client.DefaultClient)

	var req folder.AddRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddFolder, err)
		return
	}
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := folderService.AddFolder(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddFolder, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddFolder, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddFolder))

	loggerx.InfoLog(c, ActionAddFolder, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, FolderProcessName, ActionAddFolder)),
		Data: gin.H{
			"folder_id": response.GetFolderId(),
		},
	})
}

// HardDeleteFolders 物理删除多个文件夹
// @Router /phydel/folders [delete]
func (u *Folder) HardDeleteFolders(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteFolders, loggerx.MsgProcessStarted)

	var req folder.HardDeleteFoldersRequest
	req.FolderIdList = c.QueryArray("folder_id_list")
	req.Database = sessionx.GetUserCustomer(c)

	folderService := folder.NewFolderService("storage", client.DefaultClient)
	response, err := folderService.HardDeleteFolders(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteFolders, err)
		return
	}
	loggerx.SuccessLog(c, ActionHardDeleteFolders, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionHardDeleteFolders))

	loggerx.InfoLog(c, ActionHardDeleteFolders, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, FolderProcessName, ActionHardDeleteFolders)),
		Data:    response,
	})
}
