package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/global/proto/question"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/report/proto/report"
	"rxcsoft.cn/pit3/srv/storage/proto/file"
	"rxcsoft.cn/pit3/srv/storage/proto/folder"
)

// Validation 验证
type Validation struct{}

// log出力
const (
	ValidationProcessName               = "Validation"
	ActionPasswordValidation            = "PasswordValidation"
	ActionUniqueValidation              = "UniqueValidation"
	ActionFieldRelationValidation       = "FieldRelationValidation"
	ActionValueUniqueValidation         = "ValueUniqueValidation"
	ActionOptionValueValidation         = "OptionValueValidation"
	ActionCustomerNameValidation        = "CustomerNameValidation"
	ActionDatastoreApiKeyValidation     = "DatastoreApiKeyValidation"
	ActionFieldIDUinqueValidation       = "FieldIDUinqueValidation"
	ActionFileNameUinqueValidation      = "FileNameUinqueValidation"
	ActionFolderNameUinqueValidation    = "FolderNameUinqueValidation"
	ActionGroupNameUinqueValidation     = "GroupNameUinqueValidation"
	ActionQuestionTitleUinqueValidation = "QuestionTitleUinqueValidation"
	ActionRoleNameUinqueValidation      = "RoleNameUinqueValidation"
	ActionScheduleNameUinqueValidation  = "ScheduleNameUinqueValidation"
	ActionUserUinqueValidation          = "UserUinqueValidation"
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
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	res, err := languageService.FindLanguage(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUniqueValidation, err)
		return
	}

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

// UniqueValidation 验证选项值的唯一性
// @Router /validation/option [post]
func (a *Validation) OptionValueUinqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionOptionValueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	}
	var valid bool = true

	optionService := option.NewOptionService("database", client.DefaultClient)
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionOptionValueValidation, err)
		return
	}

	value := param.Value
	var req option.FindOptionRequest
	// 从path中获取参数
	req.OptionId = param.ID
	req.Invalid = "true"
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	optionRes, err := optionService.FindOption(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionOptionValueValidation)),
			Data:    valid,
		})
		return
	}
	for _, op := range optionRes.GetOptions() {
		if op.OptionValue == value {
			valid = false
			break
		}
	}

	loggerx.InfoLog(c, ActionOptionValueValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionOptionValueValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证客户名称唯一性
// @Router /validation/customer [post]
func (a *Validation) CustomerNameUinqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionUniqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionCustomerNameValidation, err)
		return
	}

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomersRequest
	req.InvalidatedIn = "true"
	response, err := customerService.FindCustomers(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionCustomerNameValidation)),
			Data:    valid,
		})
		return
	}
	for _, item := range response.GetCustomers() {
		if item.GetCustomerName() == param.Name && item.GetCustomerId() != param.ID {
			valid = false
			break
		}
	}

	loggerx.InfoLog(c, ActionCustomerNameValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionCustomerNameValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证台账apiKey唯一性
// @Router /validation/datastoreApiKey [post]
func (a *Validation) DatastoreApiKeyUinqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionUniqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		ApiKey string `json:"api_key"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionDatastoreApiKeyValidation, err)
		return
	}

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoresRequest
	// 从共通获取
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)

	response, err := datastoreService.FindDatastores(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionDatastoreApiKeyValidation)),
			Data:    valid,
		})
		return
	}
	for _, datastore := range response.GetDatastores() {
		if datastore.GetApiKey() == param.ApiKey {
			valid = false
			break
		}
	}
	loggerx.InfoLog(c, ActionDatastoreApiKeyValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionDatastoreApiKeyValidation)),
		Data:    valid,
	})
}

// UniqueValidation 字段ID唯一性检查
// @Router /validation/Field [post]
func (a *Validation) FieldIDUinqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionFieldIDUinqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		DatastoreID string `json:"datastore_id"`
		FieldID     string `json:"field_id"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionFieldIDUinqueValidation, err)
		return
	}

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.FieldRequest
	req.FieldId = param.FieldID
	req.DatastoreId = param.DatastoreID
	req.Database = sessionx.GetUserCustomer(c)
	response, err := fieldService.FindField(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFieldIDUinqueValidation)),
			Data:    valid,
		})
		return
	}
	if response.GetField().GetFieldId() == param.FieldID {
		valid = false
	}

	loggerx.InfoLog(c, ActionFieldIDUinqueValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFieldIDUinqueValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证文件名称唯一性
// @Router /validation/filename [post]
func (a *Validation) FileNameUinqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionFileNameUinqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionFileNameUinqueValidation, err)
		return
	}
	fileService := file.NewFileService("storage", client.DefaultClient)

	var req file.FindFilesRequest
	folder := param.Type
	if folder == "company" {
		req.Type = 2
		req.Domain = sessionx.GetUserDomain(c)
	} else if folder == "user" {
		req.Type = 3
		req.UserId = sessionx.GetAuthUserID(c)
		req.Domain = sessionx.GetUserDomain(c)
	} else {
		req.Type = 0
		req.FolderId = folder
		req.Domain = sessionx.GetUserDomain(c)
	}

	req.Database = sessionx.GetUserCustomer(c)

	response, err := fileService.FindFiles(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFileNameUinqueValidation)),
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

	loggerx.InfoLog(c, ActionFileNameUinqueValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFileNameUinqueValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证文件夹名称唯一性
// @Router /validation/foldername [post]
func (a *Validation) FolderNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionFolderNameUinqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Name string `json:"name"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionFolderNameUinqueValidation, err)
		return
	}
	folderService := folder.NewFolderService("storage", client.DefaultClient)

	var req folder.FindFoldersRequest
	// 从共通中获取参数
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := folderService.FindFolders(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFolderNameUinqueValidation)),
			Data:    valid,
		})
		return
	}
	for _, folder := range response.GetFolderList() {
		if folder.FolderName == param.Name {
			valid = false
			break
		}
	}

	loggerx.InfoLog(c, ActionFolderNameUinqueValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFolderNameUinqueValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证问题标题唯一性
// @Router /validation/groupname [post]
func (a *Validation) QuestionTitleDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionQuestionTitleUinqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Title string `json:"title"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionQuestionTitleUinqueValidation, err)
		return
	}
	questionService := question.NewQuestionService("global", client.DefaultClient)

	var req question.FindQuestionsRequest
	req.Domain = sessionx.GetUserDomain(c)

	response, err := questionService.FindQuestions(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionQuestionTitleUinqueValidation)),
			Data:    valid,
		})
		return
	}
	for _, question := range response.GetQuestions() {
		if question.Title == param.Title {
			valid = false
			break
		}
	}
	loggerx.InfoLog(c, ActionQuestionTitleUinqueValidation, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionQuestionTitleUinqueValidation)),
		Data:    valid,
	})
}

// UniqueValidation 验证角色名称唯一性
// @Router /validation/rolename [post]
func (a *Validation) RoleNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionRoleNameUinqueValidation, loggerx.MsgProcessStarted)

	type ReqParam struct {
		RoleID   string `json:"role_id"`
		RoleName string `json:"role_name"`
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
		if role.GetRoleName() == param.RoleName && role.GetRoleId() != param.RoleID {
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

// UniqueValidation 验证用户名称、登录id唯一性
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

// ValueUniqueValidation 验证应用下属唯一性字段数值唯一性
// @Router /validation/apps/{app_id}/value/unique [post]
func (a *Validation) ValueUniqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionValueUniqueValidation, fmt.Sprintf("Process ValueUniqueValidation:%s", loggerx.MsgProcessStarted))

	result := []string{}
	// 共通数据
	db := sessionx.GetUserCustomer(c)
	appId := c.Param("app_id")

	// 获取app下属所有字段
	fieldService := field.NewFieldService("database", client.DefaultClient)
	var afsReq field.AppFieldsRequest
	afsReq.AppId = appId
	afsReq.Database = db
	afsReq.InvalidatedIn = "true"
	afsResponse, err := fieldService.FindAppFields(context.TODO(), &afsReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionValueUniqueValidation, err)
		return
	}
	// 判断app下属字段是否有唯一性字段
	uniqueDsFs := make(map[string][]*field.Field)
	for _, f := range afsResponse.GetFields() {
		if f.Unique {
			uniqueDsFs[f.DatastoreId] = append(uniqueDsFs[f.DatastoreId], f)
		}
	}
	// 唯一性字段存在则检查其数据是否有重复空值
	if len(uniqueDsFs) > 0 {
		// 查找台账数据grpc
		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)
		itemService := item.NewItemService("database", ct)
		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}
		// 循环app下属台账,检查数据
		for datastoreId, fs := range uniqueDsFs {
			for _, f := range fs {
				// 获取台账唯一字段空值总件数
				cReq := item.KaraCountRequest{
					AppId:       appId,
					DatastoreId: datastoreId,
					FieldId:     f.FieldId,
					FieldType:   f.FieldType,
					Database:    db,
				}
				cResp, err := itemService.FindKaraCount(context.TODO(), &cReq, opss)
				if err != nil {
					httpx.GinHTTPError(c, ActionValueUniqueValidation, err)
					return
				}
				if cResp.GetTotal() > 1 {
					key := appId + "." + datastoreId + "." + f.FieldId + "." + strconv.FormatInt(cResp.GetTotal(), 10)
					result = append(result, key)
					// break
				}
			}
			// if len(result) > 0 {
			// 	break
			// }
		}
	}

	loggerx.InfoLog(c, ActionValueUniqueValidation, fmt.Sprintf("Process ValueUniqueValidation:%s", loggerx.MsgProcessEnded))
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionValueUniqueValidation)),
		Data:    result,
	})
}

// FieldRelationValidation 验证台账字段是否被有引用的报表
// @Router /validation/datastores/{id}/fields/{f_id}/relation [get]
func (a *Validation) FieldRelationValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionFieldRelationValidation, fmt.Sprintf("Process FindReports:%s", loggerx.MsgProcessStarted))

	datastoreID := c.Param("id")
	fieldID := c.Param("f_id")
	// check报表中是否引用该台账
	reportService := report.NewReportService("report", client.DefaultClient)
	var req report.FindReportsRequest
	// 共通数据
	req.Domain = sessionx.GetUserDomain(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := reportService.FindReports(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFieldRelationValidation, err)
		return
	}

	var result []*map[string]string

	for _, report := range response.GetReports() {
		if report.GetDatastoreId() == datastoreID {
			// 判断报表条件是否使用该字段
			for _, con := range report.GetReportConditions() {
				if fieldID == con.GetFieldId() {
					result = append(result, &map[string]string{
						"report_id":   report.GetReportId(),
						"report_name": report.GetReportName(),
					})
				}
			}
			// 判断报表出力字段是否使用该字段
			if report.GetIsUseGroup() {
				groupFields := report.GetGroupInfo().GetGroupKeys()
				aggreFields := report.GetGroupInfo().GetAggreKeys()

				for _, field := range groupFields {
					if fieldID == field.GetFieldId() {
						result = append(result, &map[string]string{
							"report_id":   report.GetReportId(),
							"report_name": report.GetReportName(),
						})
					}
				}
				for _, field := range aggreFields {
					if fieldID == field.GetFieldId() {
						result = append(result, &map[string]string{
							"report_id":   report.GetReportId(),
							"report_name": report.GetReportName(),
						})
					}
				}
			} else {
				selectFields := report.GetSelectKeyInfos()

				for _, field := range selectFields {
					if fieldID == field.GetFieldId() {
						result = append(result, &map[string]string{
							"report_id":   report.GetReportId(),
							"report_name": report.GetReportName(),
						})
					}
				}
			}
		}
	}

	loggerx.InfoLog(c, ActionFieldRelationValidation, fmt.Sprintf("Process FindReports:%s", loggerx.MsgProcessEnded))
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFieldRelationValidation)),
		Data:    result,
	})
}
