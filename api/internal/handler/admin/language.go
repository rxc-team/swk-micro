package admin

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/language"
)

// Language 语言
type Language struct{}

// log出力
const (
	LanguageProcessName           = "Language"
	ActionFindLanguages           = "FindLanguages"
	ActionFindLanguage            = "FindLanguage"
	ActionAddLanguage             = "AddLanguage"
	ActionAddLanguageData         = "AddLanguageData"
	ActionAddGroupLanguageData    = "AddGroupLanguageData"
	ActionAddCommonData           = "AddCommonData"
	ActionAddManyLanData          = "AddManyLanData"
	ActionAddAppLanguageData      = "AddAppLanguageData"
	ActionDeleteAppLanguageData   = "DeleteAppLanguageData"
	ActionDeleteGroupLanguageData = "DeleteGroupLanguageData"
	ActionDeleteLanguageData      = "DeleteLanguageData"
)

// FindLanguage 获取语言数据
// @Router /languages/search [get]
func (l *Language) FindLanguage(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindLanguage, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.LangCd = c.Query("lang_cd")

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
	}

	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := languageService.FindLanguage(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindLanguage, err)
		return
	}
	// // 将空值更改为默认值
	for _, value := range response.GetApps() {
		if value.GetDashboards() == nil {
			value.Dashboards = make(map[string]string)
		}
		if value.GetDatastores() == nil {
			value.Datastores = make(map[string]string)
		}
		if value.GetFields() == nil {
			value.Fields = make(map[string]string)
		}
		if value.GetMappings() == nil {
			value.Mappings = make(map[string]string)
		}
		if value.GetOptions() == nil {
			value.Options = make(map[string]string)
		}
		if value.GetQueries() == nil {
			value.Queries = make(map[string]string)
		}
		if value.GetReports() == nil {
			value.Reports = make(map[string]string)
		}
		if value.GetStatuses() == nil {
			value.Statuses = make(map[string]string)
		}
		if value.GetWorkflows() == nil {
			value.Workflows = make(map[string]string)
		}
	}

	if response.GetCommon().GetGroups() == nil {
		response.GetCommon().Groups = make(map[string]string)
	}
	if response.GetCommon().GetWorkflows() == nil {
		response.GetCommon().Workflows = make(map[string]string)
	}

	loggerx.InfoLog(c, ActionFindLanguage, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionFindLanguage)),
		Data:    response,
	})

}

// AddLanguageData 添加APP语言数据
// @Router /apps/{a_id}/languages/{lang_cd} [post]
func (l *Language) AddLanguageData(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddLanguageData, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.AddLanguageDataRequest

	// 从body获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddLanguageData, err)
		return
	}

	req.AppId = c.Param("a_id")
	req.LangCd = c.Param("lang_cd")
	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(req.GetAppId())
	key.WriteString(".app_name")

	oName := langx.GetLangData(req.GetDatabase(), req.GetDomain(), req.GetLangCd(), key.String())

	response, err := languageService.AddLanguageData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddLanguageData, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddLanguageData, fmt.Sprintf("Language[%s] Add Success", req.GetLangCd()))

	if req.GetAppName() != oName {
		params := make(map[string]string)

		params["user_name"] = sessionx.GetUserName(c)
		params["object_name"] = "{{" + key.String() + "}}"
		params["language"] = req.GetLangCd()
		params["translation"] = req.GetAppName()

		loggerx.ProcessLog(c, ActionAddLanguageData, msg.L017, params)
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, req.Domain)

	loggerx.InfoLog(c, ActionAddLanguageData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionAddLanguageData)),
		Data:    response,
	})
}

// AddAppLanguageData 添加app中的数据语言
// @Router languages/{lang_cd} [post]
func (l *Language) AddAppLanguageData(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddAppLanguageData, loggerx.MsgProcessStarted)
	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.AddAppLanguageDataRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddAppLanguageData, err)
		return
	}

	req.Domain = sessionx.GetUserDomain(c)
	req.LangCd = c.Param("lang_cd")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(req.GetAppId())
	key.WriteString(".")
	key.WriteString(req.GetType())
	key.WriteString(".")
	key.WriteString(req.GetKey())

	oName := langx.GetLangData(req.GetDatabase(), req.GetDomain(), req.GetLangCd(), key.String())

	response, err := languageService.AddAppLanguageData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddAppLanguageData, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddAppLanguageData, fmt.Sprintf("AppLanguage[%s] Add Success", req.GetAppId()))

	if req.GetValue() != oName {
		params := make(map[string]string)

		params["user_name"] = sessionx.GetUserName(c)
		params["object_name"] = "{{" + key.String() + "}}"
		params["language"] = req.GetLangCd()
		params["translation"] = req.GetValue()

		loggerx.ProcessLog(c, ActionAddAppLanguageData, msg.L017, params)
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, req.Domain)

	loggerx.InfoLog(c, ActionAddAppLanguageData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionAddLanguage)),
		Data:    response,
	})
}

// AddCommonData 添加Common语言数据
// @Router common/languages/{lang_cd} [post]
func (l *Language) AddCommonData(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddCommonData, loggerx.MsgProcessStarted)
	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.AddCommonDataRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddCommonData, err)
		return
	}

	req.Domain = sessionx.GetUserDomain(c)
	req.LangCd = c.Param("lang_cd")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	key := strings.Builder{}
	key.WriteString("common.")
	key.WriteString(req.GetType())
	key.WriteString(".")
	key.WriteString(req.GetKey())

	oName := langx.GetLangData(req.GetDatabase(), req.GetDomain(), req.GetLangCd(), key.String())

	response, err := languageService.AddCommonData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddCommonData, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddCommonData, fmt.Sprintf("Commonlanguage[%s-%s] Add Success", req.GetType(), req.GetKey()))

	if req.GetValue() != oName {
		params := make(map[string]string)

		params["user_name"] = sessionx.GetUserName(c)
		params["object_name"] = "{{" + key.String() + "}}"
		params["language"] = req.GetLangCd()
		params["translation"] = req.GetValue()

		loggerx.ProcessLog(c, ActionAddCommonData, msg.L017, params)
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, req.Domain)

	loggerx.InfoLog(c, ActionAddCommonData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionAddCommonData)),
		Data:    response,
	})
}

// AddManyLanData 添加或更新多条语言数据
// @Router /import/csv [post]
func (l *Language) AddManyLanData(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddManyLanData, loggerx.MsgProcessStarted)

	domain := sessionx.GetUserDomain(c)

	languageService := language.NewLanguageService("global", client.DefaultClient)
	var req language.AddManyLanDataRequest

	req.Domain = domain
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	appID := sessionx.GetCurrentApp(c)

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		httpx.GinHTTPError(c, ActionAddManyLanData, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionAddManyLanData, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "csv", files.Size) {
		httpx.GinHTTPError(c, ActionAddManyLanData, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	// 读取文件
	var fileData [][]string
	fileName := files.Filename

	fs, err := files.Open()
	if err != nil {
		httpx.GinHTTPError(c, ActionAddManyLanData, err)
		return
	}
	defer fs.Close()

	if fileName[strings.LastIndex(fileName, ".")+1:] != "csv" {
		// 读取Excel文件
		excelFile, err := excelize.OpenReader(fs)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddManyLanData, err)
			return
		}
		// 设置读取对象sheet,默认Sheet1
		sheetDefault := "Sheet1"
		var sheetName, sheetFirst string
		for index, name := range excelFile.GetSheetMap() {
			if index == 1 {
				sheetFirst = name
			}
			if name == sheetDefault {
				sheetName = name
				break
			}
		}
		if sheetName == "" {
			sheetName = sheetFirst
		}
		// 读取行列内容
		fileData, err = excelFile.GetRows(sheetName)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddManyLanData, err)
			return
		}
	} else {
		// 读取csv文件
		var r *csv.Reader
		encoding := c.PostForm("encoding")
		// UTF-8格式的场合，直接读取
		if encoding == "utf-8" {
			r = csv.NewReader(fs)
		} else {
			// ShiftJIS格式的场合，先转换为uft-8，再读取
			utfReader := transform.NewReader(fs, japanese.ShiftJIS.NewDecoder())
			r = csv.NewReader(utfReader)
		}
		r.LazyQuotes = true

		// 针对大文件，一行一行的读取文件
		for {
			row, err := r.Read()
			if err != nil && err != io.EOF {
				loggerx.FailureLog(c, ActionAddManyLanData, err.Error())
			}
			if err == io.EOF {
				break
			}
			fileData = append(fileData, row)
		}
	}

	var zhlans []*language.LanData
	var enlans []*language.LanData
	var jalans []*language.LanData
	var thlans []*language.LanData

	itemIndexs := make(map[string]int)
	header := fileData[1]
	for index, it := range header {
		switch it {
		case "type_name":
			itemIndexs["type_name"] = index
		case "type_id":
			itemIndexs["type_id"] = index
		case "key":
			itemIndexs["key"] = index
		case "zh_value":
			itemIndexs["zh_value"] = index
		case "en_value":
			itemIndexs["en_value"] = index
		case "ja_value":
			itemIndexs["ja_value"] = index
		case "th_value":
			itemIndexs["th_value"] = index
		}
	}

	for _, lanItems := range fileData[2:] {
		zhlan := &language.LanData{
			Type:  lanItems[itemIndexs["type_id"]],
			AppId: appID,
			Key:   lanItems[itemIndexs["key"]],
			Value: lanItems[itemIndexs["zh_value"]],
		}
		enlan := &language.LanData{
			Type:  lanItems[itemIndexs["type_id"]],
			AppId: appID,
			Key:   lanItems[itemIndexs["key"]],
			Value: lanItems[itemIndexs["en_value"]],
		}
		jalan := &language.LanData{
			Type:  lanItems[itemIndexs["type_id"]],
			AppId: appID,
			Key:   lanItems[itemIndexs["key"]],
			Value: lanItems[itemIndexs["ja_value"]],
		}
		thlan := &language.LanData{
			Type:  lanItems[itemIndexs["type_id"]],
			AppId: appID,
			Key:   lanItems[itemIndexs["key"]],
			Value: lanItems[itemIndexs["th_value"]],
		}
		if lanItems[itemIndexs["zh_value"]] != "" {
			zhlans = append(zhlans, zhlan)
		}
		if lanItems[itemIndexs["en_value"]] != "" {
			enlans = append(enlans, enlan)
		}
		if lanItems[itemIndexs["ja_value"]] != "" {
			jalans = append(jalans, jalan)
		}
		if lanItems[itemIndexs["th_value"]] != "" {
			thlans = append(thlans, thlan)
		}
	}

	if _, ok := itemIndexs["zh_value"]; ok {
		req.LangCd = "zh-CN"
		req.Lans = zhlans
		_, err = languageService.AddManyLanData(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddManyLanData, err)
			return
		}
	}
	if _, ok := itemIndexs["en_value"]; ok {
		req.LangCd = "en-US"
		req.Lans = enlans
		_, err = languageService.AddManyLanData(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddManyLanData, err)
			return
		}
	}
	if _, ok := itemIndexs["ja_value"]; ok {
		req.LangCd = "ja-JP"
		req.Lans = jalans
		_, err = languageService.AddManyLanData(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddManyLanData, err)
			return
		}
	}
	if _, ok := itemIndexs["th_value"]; ok {
		req.LangCd = "th-TH"
		req.Lans = thlans
		_, err = languageService.AddManyLanData(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddManyLanData, err)
			return
		}
	}
	loggerx.SuccessLog(c, ActionAddManyLanData, "ManyLanData Add Success")

	// ProcessLog
	userName := sessionx.GetUserName(c)

	if _, ok := itemIndexs["zh_value"]; ok {
		ozhlans := langx.GetLansMap(req.Database, req.Domain, "zh-CN")
		for _, lan := range zhlans {
			key := langx.GetLanKey(lan)
			oName := langx.GetOldName(key, ozhlans)
			if lan.Value != oName {
				params := make(map[string]string)
				params["user_name"] = userName
				params["object_name"] = "{{" + lan.Key + "}}"
				params["language"] = "zh-CN"
				params["translation"] = lan.Value
				loggerx.ProcessLog(c, ActionAddManyLanData, msg.L017, params)
			}
		}
	}
	if _, ok := itemIndexs["en_value"]; ok {
		oenlans := langx.GetLansMap(req.Database, req.Domain, "en-US")
		for _, lan := range enlans {
			key := langx.GetLanKey(lan)
			oName := langx.GetOldName(key, oenlans)
			if lan.Value != oName {
				params := make(map[string]string)
				params["user_name"] = userName
				params["object_name"] = "{{" + lan.Key + "}}"
				params["language"] = "en-US"
				params["translation"] = lan.Value
				loggerx.ProcessLog(c, ActionAddManyLanData, msg.L017, params)
			}
		}
	}
	if _, ok := itemIndexs["ja_value"]; ok {
		ojalans := langx.GetLansMap(req.Database, req.Domain, "ja-JP")
		for _, lan := range jalans {
			key := langx.GetLanKey(lan)
			oName := langx.GetOldName(key, ojalans)
			if lan.Value != oName {
				params := make(map[string]string)
				params["user_name"] = userName
				params["object_name"] = "{{" + lan.Key + "}}"
				params["language"] = "ja-JP"
				params["translation"] = lan.Value
				loggerx.ProcessLog(c, ActionAddManyLanData, msg.L017, params)
			}
		}
	}
	if _, ok := itemIndexs["th_value"]; ok {
		othlans := langx.GetLansMap(req.Database, req.Domain, "th-TH")
		for _, lan := range thlans {
			key := langx.GetLanKey(lan)
			oName := langx.GetOldName(key, othlans)
			if lan.Value != oName {
				params := make(map[string]string)
				params["user_name"] = userName
				params["object_name"] = "{{" + lan.Key + "}}"
				params["language"] = "th-TH"
				params["translation"] = lan.Value
				loggerx.ProcessLog(c, ActionAddManyLanData, msg.L017, params)
			}
		}
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, req.Domain)

	loggerx.InfoLog(c, ActionAddManyLanData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionAddManyLanData)),
		Data:    nil,
	})
}

// DeleteAppLanguageData 删除App中的语言数据
// @Router /languages/types/{type}/keys/{key} [delete]
func (l *Language) DeleteAppLanguageData(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteAppLanguageData, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.DeleteAppLanguageDataRequest

	req.Domain = sessionx.GetUserDomain(c)

	appID := c.Query("a_id")
	if len(appID) > 0 {
		req.AppId = appID
	} else {
		req.AppId = sessionx.GetCurrentApp(c)
	}

	req.Type = c.Param("type")
	req.Key = c.Param("key")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := languageService.DeleteAppLanguageData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteAppLanguageData, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteAppLanguageData, fmt.Sprintf("AppLanguage[%s] delete Success", req.GetAppId()))

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), req.Domain)

	loggerx.InfoLog(c, ActionDeleteAppLanguageData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionDeleteAppLanguageData)),
		Data:    response,
	})
}

// DeleteLanguageData 删除app语言数据
// @Router /apps/{a_id}/languages [delete]
func (l *Language) DeleteLanguageData(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteLanguageData, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.DeleteLanguageDataRequest

	req.Domain = sessionx.GetUserDomain(c)
	req.AppId = c.Param("a_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := languageService.DeleteLanguageData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteLanguageData, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteLanguageData, fmt.Sprintf("Domain[%s] Language delete Success", req.GetDomain()))

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), req.Domain)

	loggerx.InfoLog(c, ActionDeleteLanguageData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionDeleteLanguageData)),
		Data:    response,
	})
}
