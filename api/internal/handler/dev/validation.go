package dev

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/help"
	types "rxcsoft.cn/pit3/srv/global/proto/help-type"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/storage/proto/file"
)

// Validation 验证
type Validation struct{}

// log出力
const (
	ValidationProcessName          = "Validation"
	ActionPasswordValidation       = "PasswordValidation"
	ActionUniqueValidation         = "UniqueValidation"
	ActionRoleNameUinqueValidation = "RoleNameUinqueValidation"
	ActionUserUinqueValidation     = "UserUinqueValidation"
	ActionCustomerDuplicated       = "CustomerUinqueValidation"
	ActionFileNameDuplicated       = "FileNameUinqueValidation"
	ActionHelpNameDuplicated       = "HelpNameUinqueValidation"
	ActionTypeNameDuplicated       = "TypeNameUinqueValidation"
	ActionBackUpNameDuplicated     = "BackUpNameUinqueValidation"
)

// PasswordValidation 验证密码
// @Router /validation/password [post]
func (a *Validation) PasswordValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionPasswordValidation, fmt.Sprintf("Process Login:%s", loggerx.MsgProcessStarted))

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.LoginRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionPasswordValidation, err)
		return
	}
	// 更换当前的密码为md5加密后的密码
	req.Password = cryptox.GenerateMd5Password(req.GetPassword(), req.GetEmail())

	_, err := userService.Login(context.TODO(), &req)

	var result bool
	if err != nil {
		loggerx.FailureLog(c, ActionPasswordValidation, fmt.Sprintf("Verification error: [%v]", err))
		result = false
	} else {
		loggerx.SuccessLog(c, ActionPasswordValidation, "Verify password success")
		result = true
	}

	loggerx.InfoLog(c, ActionPasswordValidation, fmt.Sprintf("Process Login:%s", loggerx.MsgProcessEnded))
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionPasswordValidation)),
		Data:    result,
	})
}

// UniqueValidation 验证多语言项目的唯一性
// @Router /validation/unique [post]
func (a *Validation) UniqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionUniqueValidation, loggerx.MsgProcessStarted)

	type UniqueReq struct {
		ObjectKey string `json:"object_key"`
		Name      string `json:"name"`
		Domain    string `json:"domain"`
		Prefix    string `json:"prefix"`
		ChangeId  string `json:"change_id"`
	}

	var uReq UniqueReq

	// 从body中获取参数
	if err := c.BindJSON(&uReq); err != nil {
		httpx.GinHTTPError(c, ActionPasswordValidation, err)
		return
	}

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.LangCd = sessionx.GetCurrentLanguage(c)
	if len(uReq.Domain) > 0 {
		loggerx.InfoLog(c, ActionUniqueValidation, fmt.Sprintf("Process FindCustomerByDomain:%s", loggerx.MsgProcessStarted))
		customerService := customer.NewCustomerService("manage", client.DefaultClient)
		var cReq customer.FindCustomerByDomainRequest
		cReq.Domain = uReq.Domain
		cRes, err := customerService.FindCustomerByDomain(context.TODO(), &cReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindLanguage, err)
			return
		}
		loggerx.InfoLog(c, ActionUniqueValidation, fmt.Sprintf("Process FindCustomerByDomain:%s", loggerx.MsgProcessEnded))

		req.Domain = uReq.Domain
		req.Database = cRes.GetCustomer().GetCustomerId()
	} else {
		req.Domain = sessionx.GetUserDomain(c)
		req.Database = sessionx.GetUserCustomer(c)
	}

	loggerx.InfoLog(c, ActionUniqueValidation, fmt.Sprintf("Process FindLanguage:%s", loggerx.MsgProcessStarted))
	res, err := languageService.FindLanguage(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUniqueValidation, err)
		return
	}
	loggerx.InfoLog(c, ActionUniqueValidation, fmt.Sprintf("Process FindLanguage:%s", loggerx.MsgProcessEnded))

	result := false

	if strings.HasPrefix(uReq.ObjectKey, "common.") {
		commonSubproject := uReq.ObjectKey[7:]
		switch commonSubproject {
		case "groups":
			for key, val := range res.GetCommon().GetGroups() {
				if val == uReq.Name && key != uReq.ChangeId {
					result = true
					break
				}
			}
		default:
			break
		}
	} else {
		if uReq.ObjectKey == "appName" {
			for key, app := range res.GetApps() {
				if app.AppName == uReq.Name && key != uReq.ChangeId {
					result = true
					break
				}
			}
		} else {
			// app为单位的内容
			appId := sessionx.GetCurrentApp(c)

			for appID, app := range res.GetApps() {
				if appID == appId {
					switch uReq.ObjectKey {
					case "datastores":
						for key, val := range app.Datastores {
							if val == uReq.Name && key != uReq.ChangeId {
								result = true
								break
							}
						}
					case "fields":
						for key, val := range app.Fields {
							if strings.Contains(key, uReq.Prefix) {
								fieldId := strings.Replace(key, uReq.Prefix+"_", "", 1)
								if len(fieldId) > 0 && val == uReq.Name && fieldId != uReq.ChangeId {
									result = true
									break
								}
							}
						}
					case "workflows":
						for key, val := range app.Workflows {
							// menu的场合
							if len(uReq.Prefix) > 0 {
								wmList := strings.Split(key, "_")
								if len(wmList) == 2 && wmList[0] == uReq.Prefix && val == uReq.Name && wmList[1] != uReq.ChangeId {
									result = true
									break
								}
							} else {
								if val == uReq.Name && key != uReq.ChangeId {
									result = true
									break
								}
							}
						}
					case "mappings":
						for key, val := range app.Mappings {
							mpList := strings.Split(key, "_")
							if len(mpList) == 2 && mpList[0] == uReq.Prefix && val == uReq.Name && mpList[1] != uReq.ChangeId {
								result = true
								break
							}
						}
					case "reports":
						for key, val := range app.Reports {
							if val == uReq.Name && key != uReq.ChangeId {
								result = true
								break
							}
						}
					case "dashboards":
						for key, val := range app.Dashboards {
							if val == uReq.Name && key != uReq.ChangeId {
								result = true
								break
							}
						}
					case "options":
						for key, val := range app.Options {
							// 选项组名
							if len(uReq.Prefix) > 0 {
								if val == uReq.Name && key != uReq.ChangeId {
									result = true
									break
								}
							} else {
								optVal := strings.Replace(key, uReq.Prefix+"_", "", -1)
								if len(optVal) > 0 && val == uReq.Name && optVal != uReq.ChangeId {
									result = true
									break
								}
							}

						}
					default:
						break
					}
				}
			}
		}
	}

	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionUniqueValidation)),
		Data:    result,
	})
}

// RoleNameValidation 验证角色名称唯一性
// @Router /validation/rolename [post]
func (a *Validation) RoleNameValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionRoleNameUinqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionRoleNameUinqueValidation, err)
		return
	}
	roleService := role.NewRoleService("manage", client.DefaultClient)

	var req role.FindRolesRequest
	req.InvalidatedIn = "true"
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := roleService.FindRoles(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionRoleNameUinqueValidation)),
			Data:    valid,
		})
		return
	}
	for _, role := range response.GetRoles() {
		if role.GetRoleName() == param.Name && role.GetRoleId() != param.ID {
			valid = false
			break
		}
	}

	loggerx.InfoLog(c, ActionRoleNameUinqueValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionRoleNameUinqueValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证用户名称或登录ID唯一性
// @Router /validation/username [post]
func (a *Validation) UserDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionUserUinqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Value string `json:"value"`
		ID    string `json:"id"`
		Type  string `json:"type"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionUserUinqueValidation, err)
		return
	}

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest
	req.InvalidatedIn = "true"
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionUserUinqueValidation)),
			Data:    valid,
		})
		return
	}

	// 验证用户名
	if param.Type == "name" {
		for _, user := range response.GetUsers() {
			if user.GetUserName() == param.Value && user.GetUserId() != param.ID {
				valid = false
				break
			}
		}
	}
	// 验证登录ID
	if param.Type == "email" {
		for _, user := range response.GetUsers() {
			if user.GetEmail() == param.Value+"@"+req.Domain && user.GetUserId() != param.ID {
				valid = false
				break
			}
		}
	}

	loggerx.InfoLog(c, ActionUserUinqueValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionUserUinqueValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证客户名称、域名唯一性
// @Router /validation/customer [post]
func (a *Validation) CustomerDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionCustomerDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Value string `json:"value"`
		ID    string `json:"id"`
		Type  string `json:"type"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionCustomerDuplicated, err)
		return
	}

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomersRequest
	req.InvalidatedIn = "true"
	response, err := customerService.FindCustomers(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionCustomerDuplicated)),
			Data:    valid,
		})
		return
	}
	if param.Type == "name" {
		for _, customer := range response.GetCustomers() {
			if customer.GetCustomerName() == param.Value && customer.GetCustomerId() != param.ID {
				valid = false
				break
			}
		}
	}
	if param.Type == "domain" {
		for _, customer := range response.GetCustomers() {
			if customer.GetDomain() == param.Value && customer.GetCustomerId() != param.ID {
				valid = false
				break
			}
		}
	}

	loggerx.InfoLog(c, ActionCustomerDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionCustomerDuplicated)),
		Data:    valid,
	})
}

// UniqueValidation 验证文件名称唯一性
// @Router /validation/filename [post]
func (a *Validation) FileNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionFileNameDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Name   string `json:"name"`
		Folder string `json:"folder"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionFileNameDuplicated, err)
		return
	}
	fileService := file.NewFileService("storage", client.DefaultClient)

	var req file.FindFilesRequest
	folder := param.Folder
	if folder == "public" {
		req.Type = 1
	} else if folder == "company" {
		req.Type = 2
		req.Domain = sessionx.GetUserDomain(c)
	}

	if folder == "public" {
		req.Database = "system"
	}
	req.Database = sessionx.GetUserCustomer(c)

	response, err := fileService.FindFiles(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFileNameDuplicated)),
			Data:    valid,
		})
		return
	}
	for _, file := range response.GetFileList() {
		index := strings.LastIndex(file.GetFileName(), ".")
		fileName := file.GetFileName()[0:index]
		if fileName == param.Name {
			valid = false
			break
		}
	}

	loggerx.InfoLog(c, ActionFileNameDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFileNameDuplicated)),
		Data:    valid,
	})
}

// UniqueValidation 验证帮助文章名称唯一性
// @Router /validation/helpname [post]
func (a *Validation) HelpNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionHelpNameDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionHelpNameDuplicated, err)
		return
	}

	helpService := help.NewHelpService("global", client.DefaultClient)

	var req help.FindHelpsRequest
	req.Database = "system"

	response, err := helpService.FindHelps(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionHelpNameDuplicated)),
			Data:    valid,
		})
		return
	}
	for _, help := range response.GetHelps() {
		if help.GetTitle() == param.Name && help.GetHelpId() != param.ID {
			valid = false
			break
		}
	}

	loggerx.InfoLog(c, ActionHelpNameDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionHelpNameDuplicated)),
		Data:    valid,
	})
}

// UniqueValidation 验证类别名称唯一性
// @Router /validation/typename [post]
func (a *Validation) TypeNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionTypeNameDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionTypeNameDuplicated, err)
		return
	}

	typeService := types.NewTypeService("global", client.DefaultClient)

	var req types.FindTypesRequest
	req.Database = "system"

	response, err := typeService.FindTypes(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionTypeNameDuplicated)),
			Data:    valid,
		})
		return
	}

	for _, t := range response.GetTypes() {
		if t.GetTypeName() == param.Name && t.GetTypeId() != param.ID {
			valid = false
			break
		}
	}

	loggerx.InfoLog(c, ActionTypeNameDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionTypeNameDuplicated)),
		Data:    valid,
	})
}

// UniqueValidation 验证模板名称唯一性
// @Router /validation/backupname [post]
func (a *Validation) BackUpNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionBackUpNameDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Name string `json:"name"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionBackUpNameDuplicated, err)
		return
	}

	backupService := backup.NewBackupService("manage", client.DefaultClient)

	var req backup.FindBackupsRequest
	req.BackupType = "template"
	req.Database = sessionx.GetUserCustomer(c)
	response, err := backupService.FindBackups(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionBackUpNameDuplicated)),
			Data:    valid,
		})
		return
	}
	if len(response.GetBackups()) > 0 {
		for _, backup := range response.GetBackups() {
			if backup.BackupName == param.Name {
				valid = false
			}
		}
	}

	loggerx.InfoLog(c, ActionBackUpNameDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionBackUpNameDuplicated)),
		Data:    valid,
	})
}
