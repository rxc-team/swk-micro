package webui

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/stringx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/query"
	"rxcsoft.cn/pit3/srv/global/proto/question"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/storage/proto/file"
)

// Validation 验证
type Validation struct{}
type SpecialChar struct {
	Special string `json:"special"`
}

// log出力
const (
	ValidationProcessName         = "Validation"
	ActionPasswordValidation      = "PasswordValidation"
	ActionItemUniqueValidation    = "ItemUniqueValidation"
	ActionWorkflowExistValidation = "WorkflowExistValidation"
	ValidProcessName              = "Validator"
	ActionValidSpecial            = "ValidSpecial"
	ActionQueryNameDuplicated     = "QueryNameUinqueValidation"
	ActionFileNameDuplicated      = "FileNameUinqueValidation"
	ActionQuestionTitleDuplicated = "QuestionTitleUinqueValidation"
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

// ItemUniqueValidation 验证台账项目的唯一性
// @Router /validation/datastores/{id}/fields/{f_id}/relation [get]
func (a *Validation) ItemUniqueValidation(c *gin.Context) {
	loggerx.InfoLog(c, ActionItemUniqueValidation, fmt.Sprintf("Process FindItems:%s", loggerx.MsgProcessStarted))

	itemService := item.NewItemService("database", client.DefaultClient)

	var req item.CountRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionItemUniqueValidation, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := itemService.FindCount(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionItemUniqueValidation, err)
		return
	}

	result := true

	if response.GetTotal() > 0 {
		result = false
	}

	loggerx.InfoLog(c, ActionItemUniqueValidation, fmt.Sprintf("Process FindItems:%s", loggerx.MsgProcessEnded))
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionItemUniqueValidation)),
		Data:    result,
	})
}

// ValidSpecialChar 验证特殊字符是否合法
// @Router /validation/specialchar [get]
func (a *Validation) ValidSpecialChar(c *gin.Context) {
	loggerx.InfoLog(c, ActionValidSpecial, loggerx.MsgProcessStarted)
	// 获取公共参数
	db := sessionx.GetUserCustomer(c)
	appId := sessionx.GetCurrentApp(c)
	var value SpecialChar
	err := c.BindJSON(&value)
	if err != nil {
		httpx.GinHTTPError(c, ActionValidSpecial, err)
		return
	}

	appService := app.NewAppService("manage", client.DefaultClient)

	var req app.FindAppRequest
	req.AppId = appId
	req.Database = db
	response, err := appService.FindApp(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionValidSpecial, err)
		return
	}
	specialChars := response.GetApp().GetConfigs().GetSpecial()
	var special bool = true
	if len(specialChars) != 0 {
		var specialchar string
		// 编辑特殊字符
		for i := 0; i < len(specialChars); {
			specialchar += specialChars[i : i+1]
			i += 2
		}
		special = stringx.SpecialCheck(value.Special, specialchar)
	}
	loggerx.InfoLog(c, ActionValidSpecial, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ValidProcessName, ActionValidSpecial)),
		Data:    special,
	})
}

// UniqueValidation 验证快捷方式名称唯一性
// @Router /validation/queryname [post]
func (a *Validation) QueryNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionQueryNameDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Name string `json:"name"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionQueryNameDuplicated, err)
		return
	}

	queryService := query.NewQueryService("database", client.DefaultClient)

	var req query.FindQueriesRequest
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.UserId = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := queryService.FindQueries(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionQueryNameDuplicated)),
			Data:    valid,
		})
		return
	}
	if len(response.GetQueryList()) > 0 {
		for _, query := range response.GetQueryList() {
			if query.QueryName == param.Name {
				valid = false
				break
			}
		}
	}

	loggerx.InfoLog(c, ActionQueryNameDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionQueryNameDuplicated)),
		Data:    valid,
	})
}

// UniqueValidation 验证文件名称唯一性
// @Router /validation/filename [post]
func (a *Validation) FileNameDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionFileNameDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		FolderID string `json:"folder_id"`
		FileName string `json:"file_name"`
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
	folder := param.FolderID
	if folder == "user" {
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
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFileNameDuplicated)),
			Data:    valid,
		})
		return
	}
	if len(response.GetFileList()) > 0 {
		for _, file := range response.GetFileList() {
			index := strings.LastIndex(file.GetFileName(), ".")
			fileName := file.GetFileName()[0:index]
			if fileName == param.FileName {
				valid = false
				break
			}
		}
	}

	loggerx.InfoLog(c, ActionFileNameDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionFileNameDuplicated)),
		Data:    valid,
	})
}

// UniqueValidation 验证问题标题唯一性
// @Router /validation/querytitle [post]
func (a *Validation) QuestionTitleDuplicated(c *gin.Context) {
	loggerx.InfoLog(c, ActionQuestionTitleDuplicated, loggerx.MsgProcessStarted)

	type ReqParam struct {
		Title string `json:"title"`
	}
	var valid bool = true
	var param ReqParam
	err := c.BindJSON(&param)
	if err != nil {
		httpx.GinHTTPError(c, ActionQuestionTitleDuplicated, err)
		return
	}

	questionService := question.NewQuestionService("global", client.DefaultClient)

	var req question.FindQuestionsRequest
	req.Domain = sessionx.GetUserDomain(c)

	response, err := questionService.FindQuestions(context.TODO(), &req)
	if err != nil {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionQuestionTitleDuplicated)),
			Data:    valid,
		})
		return
	}
	if len(response.GetQuestions()) > 0 {
		for _, qu := range response.GetQuestions() {
			if qu.GetTitle() == param.Title {
				valid = false
				break
			}
		}
	}

	loggerx.InfoLog(c, ActionQuestionTitleDuplicated, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ValidationProcessName, ActionQuestionTitleDuplicated)),
		Data:    valid,
	})
}
