package admin

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/aclx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
)

// Role 角色
type Role struct{}

// log出力
const (
	RoleProcessName             = "Role"
	ActionFindRoles             = "FindRoles"
	ActionFindRolesWithResource = "FindRolesWithResource"
	ActionFindRole              = "FindRole"
	ActionFindUserActions       = "FindUserActions"
	ActionAddRole               = "AddRole"
	ActionModifyRole            = "ModifyRole"
	ActionDeleteRole            = "DeleteRole"
	ActionDeleteSelectRoles     = "DeleteSelectRoles"
	ActionHardDeleteRoles       = "HardDeleteRoles"
	ActionRecoverSelectRoles    = "RecoverSelectRoles"
	ActionWhitelistClear        = "WhitelistClear"
)

// FindRoles 获取所有角色
// @Router /roles [get]
func (u *Role) FindRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindRoles, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.FindRolesRequest
	// 从query中获取参数
	req.RoleId = c.Query("role_id")
	req.RoleName = c.Query("role_name")
	req.Description = c.Query("description")
	req.InvalidatedIn = c.Query("invalidated_in")
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.FindRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindRoles, err)
		return
	}

	loggerx.InfoLog(c, ActionFindRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionFindRoles)),
		Data:    response.GetRoles(),
	})
}

// FindRole 获取角色
// @Router /roles/{role_id} [get]
func (u *Role) FindRole(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindRole, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.FindRoleRequest
	req.RoleId = c.Param("role_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := roleService.FindRole(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindRole, err)
		return
	}

	loggerx.InfoLog(c, ActionFindRole, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionFindRole)),
		Data:    response.GetRole(),
	})
}

// FindRole 获取角色
// @Router /roles/{role_id}/actions [get]
func (u *Role) FindRoleActions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindRole, loggerx.MsgProcessStarted)

	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	roles := []string{c.Param("role_id")}

	permissionType := c.Query("permission_type")
	actionType := c.Query("action_type")

	var req permission.FindActionsRequest
	req.RoleId = roles
	req.PermissionType = permissionType

	if permissionType == "app" {
		req.AppId = sessionx.GetCurrentApp(c)
	}

	req.ActionType = actionType

	req.Database = sessionx.GetUserCustomer(c)
	response, err := pmService.FindActions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindFolders, err)
		return
	}

	loggerx.InfoLog(c, ActionFindRole, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionFindRole)),
		Data:    response.GetActions(),
	})
}

// AddRole 添加角色
// @Router /roles [post]
func (u *Role) AddRole(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddRole, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.AddRoleRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddRole, err)
		return
	}

	appID := sessionx.GetCurrentApp(c)
	for _, p := range req.GetPermissions() {
		if p.PermissionType == "app" {
			p.AppId = appID
		}
	}

	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.AddRole(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddRole, err)
		return
	}

	// 设置权限
	aclx.SetRoleCasbin(response.GetRoleId(), req.GetPermissions())

	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["profile_name"] = req.GetRoleName()
	loggerx.ProcessLog(c, ActionAddRole, msg.L023, params)

	loggerx.SuccessLog(c, ActionAddRole, fmt.Sprintf("Role[%s] create Success", response.GetRoleId()))

	loggerx.InfoLog(c, ActionAddRole, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionAddRole)),
		Data:    response,
	})
}

// ModifyRole 更新角色
// @Router /roles/{role_id} [put]
func (u *Role) ModifyRole(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyRole, loggerx.MsgProcessStarted)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	//变更前数据
	var reqF role.FindRoleRequest
	reqF.RoleId = c.Param("role_id")
	reqF.Database = sessionx.GetUserCustomer(c)
	oldRole, err := roleService.FindRole(context.TODO(), &reqF)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyRole, err)
		return
	}

	var req role.ModifyRoleRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyRole, err)
		return
	}

	appID := sessionx.GetCurrentApp(c)
	for _, p := range req.GetPermissions() {
		if p.PermissionType == "app" {
			p.AppId = appID
		}
	}

	// 从path中获取参数
	req.RoleId = c.Param("role_id")
	// 从body中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.ModifyRole(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyRole, err)
		return
	}

	aclx.SetRoleCasbin(req.GetRoleId(), req.GetPermissions())

	wg := sync.WaitGroup{}
	wg.Add(1)

	// IP白名单
	go func() {
		defer wg.Done()
		oIps := oldRole.GetRole().GetIpSegments()
		nIps := req.GetIpSegments()

		if len(oIps) != len(nIps) {

			if len(nIps) == 0 {
				//处理log
				params := make(map[string]string)
				params["user_name"] = sessionx.GetUserName(c)
				params["profile_name"] = req.GetRoleName()
				params["ip_segments"] = ""
				//处理log
				loggerx.ProcessLog(c, ActionModifyRole, msg.L074, params)
			} else {
				ipList := make([]string, len(nIps))
				for i, ip := range nIps {
					ipList[i] = ip.GetStart() + "-" + ip.GetEnd()
				}

				//处理log
				params := make(map[string]string)
				params["user_name"] = sessionx.GetUserName(c)
				params["profile_name"] = req.GetRoleName()
				params["ip_segments"] = strings.Join(ipList, ",")
				//处理log
				loggerx.ProcessLog(c, ActionModifyRole, msg.L074, params)
			}

		} else {
			hasChange := false
		NL:
			for _, nip := range nIps {
				exist := false
			OL:
				for _, oip := range oIps {
					if nip.GetStart() == oip.GetStart() && nip.GetEnd() == oip.GetEnd() {
						exist = true
						break OL
					}
				}

				if !exist {
					hasChange = true
					break NL
				}
			}

			if hasChange {
				ipList := make([]string, len(nIps))
				for i, ip := range nIps {
					ipList[i] = ip.GetStart() + "-" + ip.GetEnd()
				}

				//处理log
				params := make(map[string]string)
				params["user_name"] = sessionx.GetUserName(c)
				params["profile_name"] = req.GetRoleName()
				params["ip_segments"] = strings.Join(ipList, ",")
				//处理log
				loggerx.ProcessLog(c, ActionModifyRole, msg.L074, params)
			}
		}
	}()

	// //台账
	// go func() {
	// 	defer wg.Done()
	// 	for _, data := range req.GetDatastores() {
	// 		var oldDatastore *role.DispalyDatastore
	// 		var datastoreName string

	// 		var datastoreID string
	// 		if strings.HasSuffix(data.GetDatastoreId(), "_empty") {
	// 			datastoreID = strings.Replace(data.GetDatastoreId(), "_empty", "", 1)
	// 		} else {
	// 			datastoreID = data.GetDatastoreId()
	// 		}

	// 		for _, oldDatastoreInfo := range oldRole.GetRole().Datastores {
	// 			if oldDatastoreInfo.GetDatastoreId() == datastoreID {
	// 				oldDatastore = oldDatastoreInfo
	// 			}
	// 		}

	// 		//台账机能
	// 		if oldDatastore != nil {
	// 			//查找台账名称
	// 			datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	// 			var reqF datastore.DatastoreRequest
	// 			// 从path获取
	// 			reqF.DatastoreId = oldDatastore.GetDatastoreId()
	// 			reqF.Database = sessionx.GetUserCustomer(c)

	// 			datastoreInfo, err := datastoreService.FindDatastore(context.TODO(), &reqF)
	// 			if err != nil {
	// 				httpx.GinHTTPError(c, ActionModifyRole, err)
	// 				return
	// 			}
	// 			datastoreName = datastoreInfo.GetDatastore().DatastoreName

	// 			actions := oldDatastore.GetActions()
	// 			for key, value := range data.GetActions() {
	// 				if actions[key] != value {
	// 					//处理log
	// 					params := make(map[string]string)
	// 					params["user_name"] = sessionx.GetUserName(c)
	// 					params["profile_name"] = req.GetRoleName()
	// 					params["datastore_name"] = "{{" + datastoreName + "}}"
	// 					params["permitted_name"] = key
	// 					if value {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L027, params)
	// 					} else {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L028, params)
	// 					}
	// 				}
	// 			}

	// 			//台账字段
	// 			oldFieldMap := make(map[string]struct{})
	// 			newFieldMap := make(map[string]struct{})
	// 			for _, oldFieldID := range oldDatastore.GetFields() {
	// 				oldFieldMap[oldFieldID] = struct{}{}
	// 			}
	// 			for _, fieldID := range data.GetFields() {
	// 				newFieldMap[fieldID] = struct{}{}
	// 				if _, ok := oldFieldMap[fieldID]; !ok {
	// 					//查找字段名称
	// 					fieldService := field.NewFieldService("database", client.DefaultClient)
	// 					var reqF field.FieldRequest
	// 					reqF.FieldId = fieldID
	// 					reqF.DatastoreId = datastoreID
	// 					reqF.Database = sessionx.GetUserCustomer(c)
	// 					fieldInfo, err := fieldService.FindField(context.TODO(), &reqF)
	// 					if err != nil {
	// 						httpx.GinHTTPError(c, ActionModifyRole, err)
	// 						return
	// 					}
	// 					fieldName := fieldInfo.GetField().FieldName

	// 					//处理log
	// 					params := make(map[string]string)
	// 					params["user_name"] = sessionx.GetUserName(c)
	// 					params["profile_name"] = req.GetRoleName()
	// 					params["datastore_name"] = "{{" + datastoreName + "}}"
	// 					params["field_name"] = "{{" + fieldName + "}}"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L029, params)
	// 				}
	// 			}

	// 			for _, oldFieldID := range oldDatastore.GetFields() {
	// 				if _, ok := newFieldMap[oldFieldID]; !ok {
	// 					//查找字段名称
	// 					fieldService := field.NewFieldService("database", client.DefaultClient)
	// 					var reqF field.FieldRequest
	// 					reqF.FieldId = oldFieldID
	// 					reqF.DatastoreId = datastoreID
	// 					reqF.Database = sessionx.GetUserCustomer(c)
	// 					fieldInfo, err := fieldService.FindField(context.TODO(), &reqF)
	// 					if err != nil {
	// 						httpx.GinHTTPError(c, ActionModifyRole, err)
	// 						return
	// 					}
	// 					fieldName := fieldInfo.GetField().FieldName

	// 					//处理log
	// 					params := make(map[string]string)
	// 					params["user_name"] = sessionx.GetUserName(c)
	// 					params["profile_name"] = req.GetRoleName()
	// 					params["datastore_name"] = "{{" + datastoreName + "}}"
	// 					params["field_name"] = "{{" + fieldName + "}}"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L030, params)
	// 				}
	// 			}
	// 		} else {
	// 			//查找台账名称
	// 			datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	// 			var reqF datastore.DatastoreRequest
	// 			// 从path获取
	// 			reqF.DatastoreId = datastoreID
	// 			reqF.Database = sessionx.GetUserCustomer(c)

	// 			datastoreInfo, err := datastoreService.FindDatastore(context.TODO(), &reqF)
	// 			if err != nil {
	// 				httpx.GinHTTPError(c, ActionModifyRole, err)
	// 				return
	// 			}
	// 			datastoreName = datastoreInfo.GetDatastore().DatastoreName
	// 			// 台账机能
	// 			for key, value := range data.GetActions() {
	// 				//处理log
	// 				params := make(map[string]string)
	// 				params["user_name"] = sessionx.GetUserName(c)
	// 				params["profile_name"] = req.GetRoleName()
	// 				params["datastore_name"] = "{{" + datastoreName + "}}"
	// 				params["permitted_name"] = key
	// 				if value {
	// 					//处理log
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L027, params)
	// 				}
	// 			}

	// 			//台账字段
	// 			for _, fieldID := range data.GetFields() {
	// 				//查找字段名称
	// 				fieldService := field.NewFieldService("database", client.DefaultClient)
	// 				var reqF field.FieldRequest
	// 				reqF.FieldId = fieldID
	// 				reqF.DatastoreId = datastoreID
	// 				reqF.Database = sessionx.GetUserCustomer(c)
	// 				fieldInfo, err := fieldService.FindField(context.TODO(), &reqF)
	// 				if err != nil {
	// 					httpx.GinHTTPError(c, ActionModifyRole, err)
	// 					return
	// 				}
	// 				fieldName := fieldInfo.GetField().FieldName
	// 				//处理log
	// 				params := make(map[string]string)
	// 				params["user_name"] = sessionx.GetUserName(c)
	// 				params["profile_name"] = req.GetRoleName()
	// 				params["datastore_name"] = "{{" + datastoreName + "}}"
	// 				params["field_name"] = "{{" + fieldName + "}}"
	// 				loggerx.ProcessLog(c, ActionModifyRole, msg.L029, params)

	// 			}
	// 		}
	// 	}
	// }()

	// //报表
	// go func() {
	// 	wg.Done()
	// 	oldReportMap := make(map[string]struct{})
	// 	newReportMap := make(map[string]struct{})
	// 	for _, oldReportID := range oldRole.GetRole().Reports {
	// 		oldReportMap[oldReportID] = struct{}{}
	// 	}

	// 	//原数据不存在新数据的报表
	// 	for _, reportID := range req.GetReports() {
	// 		newReportMap[reportID] = struct{}{}
	// 		if _, ok := oldReportMap[reportID]; !ok {
	// 			//查找报表名称
	// 			reportService := report.NewReportService("report", client.DefaultClient)
	// 			var reqF report.FindReportRequest
	// 			reqF.ReportId = reportID
	// 			reqF.Database = sessionx.GetUserCustomer(c)

	// 			reportInfo, err := reportService.FindReport(context.TODO(), &reqF)
	// 			if err != nil {
	// 				httpx.GinHTTPError(c, ActionModifyRole, err)
	// 				return
	// 			}
	// 			reportName := reportInfo.GetReport().ReportName
	// 			//处理log
	// 			params := make(map[string]string)
	// 			params["user_name"] = sessionx.GetUserName(c)
	// 			params["profile_name"] = req.GetRoleName()
	// 			params["report_name"] = "{{" + reportName + "}}"
	// 			loggerx.ProcessLog(c, ActionModifyRole, msg.L031, params)
	// 		}
	// 	}
	// 	///新数据不存在原数据的报表
	// 	for _, oldReportID := range oldRole.GetRole().Reports {
	// 		if _, ok := newReportMap[oldReportID]; !ok {
	// 			//查找报表名称
	// 			reportService := report.NewReportService("report", client.DefaultClient)
	// 			var reqF report.FindReportRequest
	// 			reqF.ReportId = oldReportID
	// 			reqF.Database = sessionx.GetUserCustomer(c)

	// 			reportInfo, err := reportService.FindReport(context.TODO(), &reqF)
	// 			if err != nil {
	// 				httpx.GinHTTPError(c, ActionModifyRole, err)
	// 				return
	// 			}
	// 			reportName := reportInfo.GetReport().ReportName
	// 			//处理log
	// 			params := make(map[string]string)
	// 			params["user_name"] = sessionx.GetUserName(c)
	// 			params["profile_name"] = req.GetRoleName()
	// 			params["report_name"] = "{{" + reportName + "}}"
	// 			loggerx.ProcessLog(c, ActionModifyRole, msg.L032, params)
	// 		}
	// 	}
	// }()

	// //文件操作
	// go func() {
	// 	wg.Done()
	// 	oldFolderIDMap := make(map[string]struct{})
	// 	for _, oldFolderInfo := range oldRole.GetRole().Folders {
	// 		oldFolderIDMap[oldFolderInfo.GetFolderId()] = struct{}{}
	// 	}

	// 	//文件权限变更判断
	// 	var oldFolder *role.DispalyFolder
	// 	for _, oldFolderInfo := range oldRole.GetRole().Folders {
	// 		for _, data := range req.GetFolders() {
	// 			if oldFolderInfo.GetFolderId() == data.GetFolderId() {
	// 				oldFolder = oldFolderInfo
	// 			}
	// 			var folderName string
	// 			isFolderFlg := false
	// 			if _, ok := oldFolderIDMap[data.GetFolderId()]; ok {
	// 				isFolderFlg = true
	// 				//查找文件夹名称
	// 				folderService := folder.NewFolderService("storage", client.DefaultClient)
	// 				var reqF folder.FindFolderRequest
	// 				reqF.FolderId = data.GetFolderId()
	// 				reqF.Database = sessionx.GetUserCustomer(c)
	// 				folderInfo, err := folderService.FindFolder(context.TODO(), &reqF)
	// 				if err != nil {
	// 					httpx.GinHTTPError(c, ActionModifyRole, err)
	// 					return
	// 				}
	// 				folderName = folderInfo.GetFolder().FolderName
	// 			}
	// 			if isFolderFlg {
	// 				params := make(map[string]string)
	// 				params["user_name"] = sessionx.GetUserName(c)
	// 				params["profile_name"] = req.GetRoleName()
	// 				params["folder_name"] = folderName
	// 				//文件读取权限
	// 				if data.GetRead() != oldFolder.GetRead() {
	// 					params["authority_name"] = "read"
	// 					if data.GetRead() {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L033, params)
	// 					} else {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L034, params)
	// 					}

	// 				}
	// 				//文件写入权限
	// 				if data.GetWrite() != oldFolder.GetWrite() {
	// 					params["authority_name"] = "write"
	// 					if data.GetWrite() {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L033, params)
	// 					} else {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L034, params)
	// 					}
	// 				}
	// 				//文件删除权限
	// 				if data.GetDelete() != oldFolder.GetDelete() {
	// 					params["authority_name"] = "delete"
	// 					if data.GetDelete() {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L033, params)
	// 					} else {
	// 						//处理log
	// 						loggerx.ProcessLog(c, ActionModifyRole, msg.L034, params)
	// 					}
	// 				}
	// 			}
	// 			if !isFolderFlg {
	// 				params := make(map[string]string)
	// 				params["user_name"] = sessionx.GetUserName(c)
	// 				params["profile_name"] = req.GetRoleName()
	// 				params["folder_name"] = folderName
	// 				//文件读取权限
	// 				if data.GetRead() {
	// 					//处理log
	// 					params["authority_name"] = "read"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L033, params)
	// 				} else {
	// 					//处理log
	// 					params["authority_name"] = "read"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L034, params)
	// 				}

	// 				//文件写入权限
	// 				if data.GetWrite() {
	// 					//处理log
	// 					params["authority_name"] = "write"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L033, params)
	// 				} else {
	// 					//处理log
	// 					params["authority_name"] = "write"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L034, params)
	// 				}
	// 				//文件删除权限
	// 				if data.GetDelete() {
	// 					//处理log
	// 					params["authority_name"] = "delete"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L033, params)
	// 				} else {
	// 					//处理log
	// 					params["authority_name"] = "delete"
	// 					loggerx.ProcessLog(c, ActionModifyRole, msg.L034, params)
	// 				}
	// 			}
	// 		}
	// 	}
	// }()

	wg.Wait()

	loggerx.SuccessLog(c, ActionModifyRole, fmt.Sprintf("Role[%s] update Success", req.GetRoleId()))

	loggerx.InfoLog(c, ActionModifyRole, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionModifyRole)),
		Data:    response,
	})
}

// DeleteSelectRoles 删除选中角色
// @Router /roles [delete]
func (u *Role) DeleteSelectRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSelectRoles, loggerx.MsgProcessStarted)
	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.DeleteSelectRolesRequest
	req.RoleIdList = c.QueryArray("role_id_list")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	var roleNameList []string
	for _, id := range req.RoleIdList {
		var reqF role.FindRoleRequest
		reqF.RoleId = id
		reqF.Database = sessionx.GetUserCustomer(c)
		result, err := roleService.FindRole(context.TODO(), &reqF)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteSelectRoles, err)
			return
		}
		roleName := result.GetRole().RoleName
		roleNameList = append(roleNameList, roleName)
	}

	response, err := roleService.DeleteSelectRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectRoles, err)
		return
	}

	//处理log
	for _, roleName := range roleNameList {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["profile_name"] = roleName
		loggerx.ProcessLog(c, ActionDeleteSelectRoles, msg.L025, params)
	}

	loggerx.SuccessLog(c, ActionDeleteSelectRoles, fmt.Sprintf("Roles[%s] delete Success", req.GetRoleIdList()))

	loggerx.InfoLog(c, ActionDeleteSelectRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionDeleteSelectRoles)),
		Data:    response,
	})
}

// HardDeleteRoles 物理删除选中角色
// @Router /phydel/roles [delete]
func (u *Role) HardDeleteRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteRoles, loggerx.MsgProcessStarted)
	roleService := role.NewRoleService("manage", client.DefaultClient)
	var req role.HardDeleteRolesRequest
	req.RoleIdList = c.QueryArray("role_id_list")
	req.Database = sessionx.GetUserCustomer(c)

	var roleNameList []string
	for _, id := range req.RoleIdList {
		var reqF role.FindRoleRequest
		reqF.RoleId = id
		reqF.Database = sessionx.GetUserCustomer(c)
		result, err := roleService.FindRole(context.TODO(), &reqF)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteRoles, err)
			return
		}
		roleName := result.GetRole().RoleName
		roleNameList = append(roleNameList, roleName)
	}

	response, err := roleService.HardDeleteRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteRoles, err)
		return
	}

	// 删除
	for _, id := range req.RoleIdList {
		aclx.ClearRoleCasbin(id)
	}

	//处理log
	for _, roleName := range roleNameList {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["profile_name"] = roleName
		loggerx.ProcessLog(c, ActionHardDeleteRoles, msg.L024, params)
	}
	loggerx.SuccessLog(c, ActionHardDeleteRoles, fmt.Sprintf("Roles[%s] physically delete Success", req.GetRoleIdList()))

	loggerx.InfoLog(c, ActionHardDeleteRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionHardDeleteRoles)),
		Data:    response,
	})
}

// RecoverSelectRoles 恢复选中角色
// @Router /recover/roles [PUT]
func (u *Role) RecoverSelectRoles(c *gin.Context) {
	loggerx.InfoLog(c, ActionRecoverSelectRoles, loggerx.MsgProcessStarted)
	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.RecoverSelectRolesRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectRoles, err)
		return
	}

	db := sessionx.GetUserCustomer(c)
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = db

	var roleNameList []string
	for _, id := range req.RoleIdList {
		var reqF role.FindRoleRequest
		reqF.RoleId = id
		reqF.Database = db
		result, err := roleService.FindRole(context.TODO(), &reqF)
		if err != nil {
			httpx.GinHTTPError(c, ActionRecoverSelectRoles, err)
			return
		}
		roleName := result.GetRole().RoleName
		roleNameList = append(roleNameList, roleName)
	}

	response, err := roleService.RecoverSelectRoles(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectRoles, err)
		return
	}

	//处理log
	for _, roleName := range roleNameList {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["profile_name"] = roleName
		loggerx.ProcessLog(c, ActionRecoverSelectRoles, msg.L026, params)
	}
	loggerx.SuccessLog(c, ActionRecoverSelectRoles, fmt.Sprintf("Roles[%s] recover Success", req.GetRoleIdList()))

	loggerx.InfoLog(c, ActionRecoverSelectRoles, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I013, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionRecoverSelectRoles)),
		Data:    response,
	})
}

// WhitelistClear 清空白名单
// @Router /role/whitelistClear/roles [PUT]
func (u *Role) WhitelistClear(c *gin.Context) {
	loggerx.InfoLog(c, ActionWhitelistClear, loggerx.MsgProcessStarted)

	var req role.WhitelistClearRequest

	req.Database = c.Query("database")
	req.Writer = sessionx.GetAuthUserID(c)

	roleService := role.NewRoleService("manage", client.DefaultClient)
	response, err := roleService.WhitelistClear(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionWhitelistClear, err)
		return
	}

	loggerx.SuccessLog(c, ActionWhitelistClear, fmt.Sprintf("customer[%s] administrator WhitelistClear Success", req.GetDatabase()))

	// 清空白名单成功后保存日志到DB
	var reqR role.FindRolesRequest
	reqR.Database = c.Query("database")
	reqR.RoleType = "1"
	reqR.Domain = c.Query("domain")
	res, err := roleService.FindRoles(context.TODO(), &reqR)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindRoles, err)
		return
	}
	arr := []string{}
	for _, ro := range res.GetRoles() {
		arr = append(arr, ro.GetRoleName())
	}
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["profile_name"] = strings.Join(arr, ",")
	params["ip_segments"] = ""
	loggerx.ProcessLog(c, ActionWhitelistClear, msg.L074, params)

	loggerx.InfoLog(c, ActionWhitelistClear, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, RoleProcessName, ActionWhitelistClear)),
		Data:    response,
	})
}
