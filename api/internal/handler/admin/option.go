package admin

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// Option 选项
type Option struct{}

// log出力
const (
	OptionProcessName           = "Option"
	ActionFindOptions           = "FindOptions"
	ActionFindOption            = "FindOption"
	ActionFindOptionLabels      = "FindOptionLabels"
	ActionAddOption             = "AddOption"
	ActionDeleteOption          = "DeleteOption"
	ActionDeleteOptionChild     = "DeleteOptionChild"
	ActionHardDeleteOptionChild = "HardDeleteOptionChild"
	ActionDeleteSelectOptions   = "DeleteSelectOptions"
	ActionHardDeleteOptions     = "HardDeleteOptions"
	ActionRecoverSelectOptions  = "RecoverSelectOptions"
	ActionRecoverOptionChild    = "RecoverOptionChild"
	ActionOptionLogDownload     = "OptionLogDownload"
)

// FindOptions 获取所有选项
// @Router /options [get]
func (u *Option) FindOptions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOptions, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionsRequest
	// 从query中获取参数
	req.OptionName = c.Query("option_name")
	req.OptionMemo = c.Query("option_memo")
	req.InvalidatedIn = c.Query("invalidated_in")
	// 从共通中获取参数
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)
	if c.Query("app_id") != "" {
		req.AppId = c.Query("app_id")
	}

	response, err := optionService.FindOptions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindOptions, err)
		return
	}

	loggerx.InfoLog(c, ActionFindOptions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionFindOptions)),
		Data:    response.GetOptions(),
	})
}

// FindOptionLabels 获取所有选项数据
// @Router /options/{o_id} [get]
func (u *Option) FindOptionLabels(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOptionLabels, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionLabelsRequest
	// 从共通中获取参数
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)

	response, err := optionService.FindOptionLabels(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindOptionLabels, err)
		return
	}

	loggerx.InfoLog(c, ActionFindOptionLabels, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionFindOptionLabels)),
		Data:    response.GetOptions(),
	})
}

// FindOption 获取选项
// @Router /options/{o_id} [get]
func (u *Option) FindOption(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOption, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionRequest
	// 从path中获取参数
	req.OptionId = c.Param("o_id")
	req.Invalid = c.Query("invalidated_in")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := optionService.FindOption(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindOption, err)
		return
	}

	loggerx.InfoLog(c, ActionFindOption, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionFindOption)),
		Data:    response.GetOptions(),
	})
}

// AddOption 添加选项
// @Router /options [post]
func (u *Option) AddOption(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddOption, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.AddRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddOption, err)
		return
	}
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := optionService.AddOption(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddOption, err)
		return
	}

	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["option_gourp_name"] = req.GetOptionName()
	params["option_id"] = response.GetOptionId()
	params["option_name"] = req.GetOptionLabel()
	params["option_value"] = req.GetOptionValue()
	loggerx.ProcessLog(c, ActionAddOption, msg.L061, params)

	loggerx.SuccessLog(c, ActionAddOption, fmt.Sprintf("Option[%s] create Success", response.GetId()))

	if req.GetIsNewOptionGroup() {
		// 添加多语言数据
		langService := language.NewLanguageService("global", client.DefaultClient)

		languageReq := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    sessionx.GetCurrentApp(c),
			Type:     "options",
			Key:      response.GetOptionId(),
			Value:    req.GetOptionName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}

		_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddOption, err)
			return
		}
		loggerx.SuccessLog(c, ActionAddOption, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))
	}

	// 添加多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    sessionx.GetCurrentApp(c),
		Type:     "options",
		Key:      response.GetOptionId() + "_" + req.GetOptionValue(),
		Value:    req.GetOptionLabel(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddOption, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddOption, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, sessionx.GetUserDomain(c))

	loggerx.InfoLog(c, ActionAddOption, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionAddOption)),
		Data:    response,
	})
}

// DeleteOptionChild 删除选项值数据
// @Router /options/{o_id}/values/{value} [delete]
func (u *Option) DeleteOptionChild(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteOptionChild, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var reqF option.FindOptionLabelRequest
	reqF.OptionId = c.Param("o_id")
	reqF.OptionValue = c.Param("value")
	// 从共通中获取参数
	reqF.Database = sessionx.GetUserCustomer(c)
	reqF.AppId = sessionx.GetCurrentApp(c)

	result, err := optionService.FindOptionLable(context.TODO(), &reqF)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteOptionChild, err)
		return
	}

	var req option.DeleteChildRequest
	// 从path中获取参数
	req.OptionId = c.Param("o_id")
	req.OptionValue = c.Param("value")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := optionService.DeleteOptionChild(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteOptionChild, err)
		return
	}

	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["option_gourp_name"] = "{{" + result.GetOption().OptionName + "}}"
	params["option_id"] = result.GetOption().GetOptionId()
	params["option_name"] = "{{" + result.GetOption().OptionLabel + "}}"
	params["option_value"] = result.GetOption().GetOptionValue()
	loggerx.ProcessLog(c, ActionDeleteOptionChild, msg.L065, params)

	loggerx.SuccessLog(c, ActionDeleteOptionChild, fmt.Sprintf("Optionvalue[%s] delete Success", req.GetOptionValue()))

	loggerx.InfoLog(c, ActionDeleteOptionChild, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionDeleteOptionChild)),
		Data:    response,
	})
}

// HardDeleteOptionChild 物理删除选项值数据(子选项)
// @Router /phydel/options/{o_id}/values/{value} [delete]
func (u *Option) HardDeleteOptionChild(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteOptionChild, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)

	var reqF option.FindOptionLabelRequest
	reqF.OptionId = c.Param("o_id")
	reqF.OptionValue = c.Param("value")
	// 从共通中获取参数
	reqF.Database = sessionx.GetUserCustomer(c)
	reqF.AppId = sessionx.GetCurrentApp(c)

	result, err := optionService.FindOptionLable(context.TODO(), &reqF)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteOptionChild, err)
		return
	}

	var req option.HardDeleteChildRequest
	// 从path中获取参数
	req.OptionId = c.Param("o_id")
	req.OptionValue = c.Param("value")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := optionService.HardDeleteOptionChild(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteOptionChild, err)
		return
	}

	optionLable := strings.Builder{}
	optionLable.WriteString(langx.GetLangData(db, domain, lang, result.GetOption().OptionLabel))
	optionLable.WriteString("(")
	optionLable.WriteString(sessionx.GetCurrentLanguage(c))
	optionLable.WriteString(")")
	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["option_gourp_name"] = "{{" + result.GetOption().OptionName + "}}"
	params["option_id"] = result.GetOption().GetOptionId()
	params["option_name"] = optionLable.String()
	params["option_value"] = result.GetOption().GetOptionValue()
	loggerx.ProcessLog(c, ActionHardDeleteOptionChild, msg.L066, params)

	loggerx.SuccessLog(c, ActionHardDeleteOptionChild, fmt.Sprintf("Optionvalue[%s] physically delete Success", req.GetOptionValue()))
	appId := sessionx.GetCurrentApp(c)
	// 删除多语言数据
	langx.DeleteAppLanguageData(db, domain, appId, "options", req.GetOptionId()+"_"+req.GetOptionValue())

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionHardDeleteOptionChild, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionHardDeleteOptionChild)),
		Data:    response,
	})
}

// DeleteSelectOptions 删除选中选项（无效化）
// @Router /options [delete]
func (u *Option) DeleteSelectOptions(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSelectOptions, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)
	var req option.DeleteSelectOptionsRequest
	req.OptionIdList = c.QueryArray("option_id_list")
	req.Writer = sessionx.GetAuthUserID(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	var optionNameList []string
	optionIdList := make(map[string]string)
	for _, id := range req.OptionIdList {
		var reqF option.FindOptionRequest
		// 从path中获取参数
		reqF.OptionId = id
		reqF.Invalid = ""
		// 从共通中获取参数
		reqF.AppId = sessionx.GetCurrentApp(c)
		reqF.Database = sessionx.GetUserCustomer(c)

		result, err := optionService.FindOption(context.TODO(), &reqF)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteSelectOptions, err)
			return
		}

		if len(result.GetOptions()) > 0 {
			optionName := result.GetOptions()[0].OptionName
			optionNameList = append(optionNameList, optionName)
			optionIdList[optionName] = id
		}

	}

	response, err := optionService.DeleteSelectOptions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectOptions, err)
		return
	}

	//处理log
	for _, optionName := range optionNameList {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["option_gourp_name"] = "{{" + optionName + "}}"
		params["option_id"] = optionIdList[optionName]
		loggerx.ProcessLog(c, ActionDeleteSelectOptions, msg.L063, params)
	}
	loggerx.SuccessLog(c, ActionDeleteSelectOptions, fmt.Sprintf("Options[%s] delete Success", req.GetOptionIdList()))

	loggerx.InfoLog(c, ActionDeleteSelectOptions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionDeleteSelectOptions)),
		Data:    response,
	})
}

// HardDeleteOptions 物理删除选中选项
// @Router /phydel/options [delete]
func (u *Option) HardDeleteOptions(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteOptions, loggerx.MsgProcessStarted)

	var req option.HardDeleteOptionsRequest
	req.OptionIdList = c.QueryArray("option_id_list")

	req.Database = sessionx.GetUserCustomer(c)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)
	appId := sessionx.GetCurrentApp(c)

	req.AppId = appId

	langData := langx.GetLanguageData(db, lang, domain)

	optionService := option.NewOptionService("database", client.DefaultClient)
	var optionNameList []string
	optionIdList := make(map[string]string)
	for _, optionID := range req.OptionIdList {
		var findReq option.FindOptionRequest
		findReq.AppId = appId
		findReq.OptionId = optionID
		findReq.Invalid = "true"
		findReq.Database = sessionx.GetUserCustomer(c)
		response, err := optionService.FindOption(context.TODO(), &findReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteOptions, err)
			return
		}

		if len(response.GetOptions()) > 0 {
			optionName := langx.GetLangValue(langData, response.GetOptions()[0].OptionName, langx.DefaultResult)
			optionNameList = append(optionNameList, optionName)
			optionIdList[optionName] = optionID
		}

		for _, item := range response.GetOptions() {
			// 删除多语言数据
			langx.DeleteAppLanguageData(db, domain, appId, "options", optionID+"_"+item.GetOptionValue())
		}

		// 删除多语言数据
		langx.DeleteAppLanguageData(db, domain, appId, "options", optionID)
	}

	response, err := optionService.HardDeleteOptions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteOptions, err)
		return
	}

	for _, name := range optionNameList {
		optionName := strings.Builder{}
		optionName.WriteString(name)
		optionName.WriteString("(")
		optionName.WriteString(sessionx.GetCurrentLanguage(c))
		optionName.WriteString(")")
		//处理log
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["option_gourp_name"] = optionName.String()
		params["option_id"] = optionIdList[name]
		loggerx.ProcessLog(c, ActionHardDeleteOptions, msg.L062, params)
	}

	loggerx.SuccessLog(c, ActionHardDeleteOptions, fmt.Sprintf("Options[%s] physically delete Success", req.GetOptionIdList()))

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionHardDeleteOptions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionHardDeleteOptions)),
		Data:    response,
	})
}

// RecoverSelectOptions 恢复选中的选项(选项组)
// @Router /recover/options [PUT]
func (u *Option) RecoverSelectOptions(c *gin.Context) {
	loggerx.InfoLog(c, ActionRecoverSelectOptions, loggerx.MsgProcessStarted)
	optionService := option.NewOptionService("database", client.DefaultClient)
	var req option.RecoverSelectOptionsRequest

	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectOptions, err)
		return
	}
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	var optionNameList []string
	optionIdList := make(map[string]string)
	for _, id := range req.OptionIdList {
		var reqF option.FindOptionRequest
		// 从path中获取参数
		reqF.OptionId = id
		reqF.Invalid = "invalidatedIn"
		// 从共通中获取参数
		reqF.AppId = sessionx.GetCurrentApp(c)
		reqF.Database = sessionx.GetUserCustomer(c)

		result, err := optionService.FindOption(context.TODO(), &reqF)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteSelectOptions, err)
			return
		}

		if len(result.GetOptions()) > 0 {
			optionName := result.GetOptions()[0].OptionName
			optionNameList = append(optionNameList, optionName)
			optionIdList[optionName] = id
		}
	}

	response, err := optionService.RecoverSelectOptions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectOptions, err)
		return
	}
	for _, optionName := range optionNameList {
		//处理log
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["option_gourp_name"] = "{{" + optionName + "}}"
		params["option_id"] = optionIdList[optionName]
		loggerx.ProcessLog(c, ActionRecoverSelectOptions, msg.L064, params)
	}
	loggerx.SuccessLog(c, ActionRecoverSelectOptions, fmt.Sprintf("Options[%s] recover Success", req.GetOptionIdList()))

	loggerx.InfoLog(c, ActionRecoverSelectOptions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I013, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionRecoverSelectOptions)),
		Data:    response,
	})
}

// RecoverOptionChild 恢复某个选项组下的某个值数据
// @Router /recover/options/{o_id}/values/{value} [delete]
func (u *Option) RecoverOptionChild(c *gin.Context) {
	loggerx.InfoLog(c, ActionRecoverOptionChild, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var reqF option.FindOptionLabelRequest
	reqF.OptionId = c.Param("o_id")
	reqF.OptionValue = c.Param("value")
	// 从共通中获取参数
	reqF.Database = sessionx.GetUserCustomer(c)
	reqF.AppId = sessionx.GetCurrentApp(c)

	result, err := optionService.FindOptionLable(context.TODO(), &reqF)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverOptionChild, err)
		return
	}

	var req option.RecoverChildRequest
	// 从path中获取参数
	req.OptionId = c.Param("o_id")
	req.OptionValue = c.Param("value")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := optionService.RecoverOptionChild(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverOptionChild, err)
		return
	}

	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["option_gourp_name"] = "{{" + result.GetOption().OptionName + "}}"
	params["option_id"] = result.GetOption().GetOptionId()
	params["option_name"] = "{{" + result.GetOption().OptionLabel + "}}"
	params["option_value"] = result.GetOption().GetOptionValue()
	loggerx.ProcessLog(c, ActionDeleteOptionChild, msg.L067, params)

	loggerx.SuccessLog(c, ActionRecoverOptionChild, fmt.Sprintf("OptionValue[%s] Recover Success", req.GetOptionValue()))

	loggerx.InfoLog(c, ActionRecoverOptionChild, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionRecoverOptionChild)),
		Data:    response,
	})
}

// DownloadOptions 下载所有选项数据
// @Router /options/{o_id} [get]
func (u *Option) DownloadCSVOptions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOptionLabels, loggerx.MsgProcessStarted)
	// 参数收集
	userID := sessionx.GetAuthUserID(c)
	appID := sessionx.GetCurrentApp(c)

	db := sessionx.GetUserCustomer(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	langData := langx.GetLanguageData(db, lang, domain)

	jobID := "job_" + time.Now().Format("20060102150405")
	go func() {
		// 创建审批日志下载任务
		jobx.CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      "option data download",
			Origin:       "page.schedule.optionOrigin",
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(lang, "job.J_014"),
			TaskType:     "op-csv-download",
			Steps:        []string{"start", "get-data", "build-data", "write-to-file", "save-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})
		// 发送消息 开始获取数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "オプションデータを取得します",
			CurrentStep: "get-data",
			Database:    db,
		}, userID)

		optionService := option.NewOptionService("database", client.DefaultClient)

		var req option.FindOptionLabelsRequest
		// 从共通中获取参数
		req.Database = sessionx.GetUserCustomer(c)
		req.AppId = sessionx.GetCurrentApp(c)
		req.InvalidatedIn = c.Query("invalidatedIn")
		req.OptionMemo = c.Query("option_memo")
		req.OptionName = c.Query("option_name")

		optionRes, err := optionService.FindOptionLabels(context.TODO(), &req)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "get-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 数据获取完成，开始编辑
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ビルドオプションデータ",
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		// 文件头编辑
		var header []string
		header = append(header, i18n.Tr(lang, "fixed.F_001"))
		header = append(header, i18n.Tr(lang, "fixed.F_002"))
		header = append(header, i18n.Tr(lang, "fixed.F_003"))
		header = append(header, i18n.Tr(lang, "fixed.F_004"))
		header = append(header, i18n.Tr(lang, "fixed.F_005"))

		headers := append([][]string{}, header)

		// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
		headers[0][0] = "\xEF\xBB\xBF" + headers[0][0]

		filex.Mkdir("temp/")
		// 编辑数据
		timestamp := time.Now().Format("20060102150405")

		// 写入文件到本地
		filename := "temp/tmp" + "_" + timestamp + ".csv"
		f, err := os.Create(filename)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "write-to-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		writer := csv.NewWriter(f)
		err = writer.WriteAll(headers)
		if err != nil {
			if err.Error() == "encoding: rune not supported by encoding." {
				path := filex.WriteAndSaveFile(domain, appID, []string{"現在のタイトルには、日本語の[shift-jis]エンコード以外の文字が含まれており、実行を続行できません。"})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "write-to-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "write-to-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		writer.Flush() // 此时才会将缓冲区中的header数据写入

		// 向文件中写入option数据
		var items [][]string
		for _, option := range optionRes.GetOptions() {
			// 设置csv行
			var itemData []string
			itemData = append(itemData, option.GetOptionId())
			itemData = append(itemData, langx.GetLangValue(langData, option.GetOptionName(), langx.DefaultResult))
			itemData = append(itemData, option.GetOptionValue())
			itemData = append(itemData, langx.GetLangValue(langData, option.GetOptionLabel(), langx.DefaultResult))
			itemData = append(itemData, option.GetOptionMemo())
			// 添加行
			items = append(items, itemData)
		}
		writer.WriteAll(items)

		writer.Flush() // 此时才会将缓冲区的option数据写入
		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "一時ファイルからファイルをマージ",
			CurrentStep: "save-file",
			Database:    db,
		}, userID)
		f.Close()

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルをファイルサーバーに保存",
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		// 上传文件到minio服务器
		fo, err := os.Open(filename)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		defer func() {
			fo.Close()
			os.Remove(filename)
		}()

		// 写入文件到 minio
		minioClient, err := storagecli.NewClient(domain)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		appRoot := "app_" + appID
		filePath := path.Join(appRoot, "csv", "option"+timestamp+".csv")
		path, err := minioClient.SavePublicObject(fo, filePath, "text/csv")
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		// 判断顾客上传文件是否在设置的最大存储空间以内
		canUpload := filex.CheckCanUpload(domain, float64(path.Size))
		if canUpload {
			// 如果没有超出最大值，就对顾客的已使用大小进行累加
			err = filex.ModifyUsedSize(domain, float64(path.Size))
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
		} else {
			// 如果已达上限，则删除刚才上传的文件
			minioClient.DeleteObject(path.Name)
			path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "最大ストレージ容量に達しました。ファイルのアップロードに失敗しました",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			File: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database: db,
		}, userID)
	}()

	// 设置下载的文件名
	loggerx.InfoLog(c, ActionOptionLogDownload, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, "Option", ActionOptionLogDownload)),
		Data:    gin.H{},
	})
}
