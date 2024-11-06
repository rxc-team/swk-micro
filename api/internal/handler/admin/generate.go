package admin

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/generate"
	"rxcsoft.cn/pit3/srv/global/proto/language"
)

// Generate 生成
type Generate struct{}

// log出力
const (
	GenerateProcessName = "Generate"
	ActionUpload        = "Upload"
)

// Upload 第一步① 上传csv数据
// @Router /upload [get]
func (u *Generate) Upload(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	charEncoding := c.PostForm("encoding")
	comma := c.PostForm("comma")
	// comment := c.PostForm("comment")

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
		httpx.GinHTTPError(c, ActionUpload, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "csv", files.Size) {
		httpx.GinHTTPError(c, ActionUpload, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	// 读取csv文件
	fs, err := files.Open()
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	var r *csv.Reader
	// UTF-8格式的场合，直接读取
	if charEncoding == "UTF-8" {
		r = csv.NewReader(fs)
	} else {
		// ShiftJIS格式的场合，先转换为uft-8，再读取
		utfReader := transform.NewReader(fs, japanese.ShiftJIS.NewDecoder())
		r = csv.NewReader(utfReader)
	}
	r.LazyQuotes = true

	if comma == "," {
		r.Comma = 44 // 逗号
	} else {
		r.Comma = 9 // 制表符
	}

	var fileData [][]string

	// 针对大文件，一行一行的读取文件
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			httpx.GinHTTPError(c, ActionUpload, err)
			return
		}
		if err == io.EOF {
			break
		}
		fileData = append(fileData, row)
	}

	var items []*generate.Item
	header := fileData[0]
	// 去除utf-8 withbom的前缀
	header[0] = strings.Replace(header[0], "\ufeff", "", 1)
	for _, row := range fileData[1:] {

		if len(row) != len(header) {
			httpx.GinHTTPError(c, ActionUpload, fmt.Errorf("csv数据不正"))
			return
		}

		item := generate.Item{}
		item.AppId = appId
		item.UserId = userId

		item.ItemMap = make(map[string]string)

		for i, col := range row {
			item.ItemMap[header[i]] = col
		}

		items = append(items, &item)
	}

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	var req generate.UploadRequest
	req.Items = items
	req.Database = db

	_, err = genService.UploadData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	var aReq generate.AddRequest
	aReq.AppId = appId
	aReq.UserId = userId
	aReq.Database = db

	_, err = genService.AddGenerateConfig(context.TODO(), &aReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    gin.H{},
	})
}

// FindRowData 第一步② 获取上传的数据
// @Router /row/data [get]
func (u *Generate) FindRowData(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	var req generate.RowRequest
	req.AppId = appId
	req.UserId = userId
	// 从query获取
	index := c.Query("page_index")
	size := c.Query("page_size")
	pageIndex, _ := strconv.ParseInt(index, 10, 64)
	pageSize, _ := strconv.ParseInt(size, 10, 64)
	req.PageIndex = pageIndex
	req.PageSize = pageSize
	req.Database = db

	response, err := genService.FindRowData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	if len(response.Items) == 0 {

		loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
			Data: gin.H{
				"header": []string{},
				"data":   []string{},
				"total":  100,
			},
		})
		return
	}

	first := response.Items[0]

	header := make([]string, 0, len(first.ItemMap))
	for k := range first.ItemMap {
		header = append(header, k)
	}

	var data [][]string

	for _, v := range response.GetItems() {
		var cols []string
		for _, key := range header {
			cols = append(cols, v.ItemMap[key])
		}

		data = append(data, cols)
	}

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data: gin.H{
			"header": header,
			"data":   data,
			"total":  100,
		},
	})
}

// DatabaseSet 第二步 台账信息设置
// @Router /database/set [get]
func (u *Generate) DatabaseSet(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	var req generate.ModifyRequest

	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyField, err)
		return
	}

	req.AppId = appId
	req.UserId = userId
	req.Step = 1
	req.Database = db

	_, err := genService.ModifyGenerateConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    gin.H{},
	})
}

// FiledSet 第三步 台账字段设置
// @Router /field/set [get]
func (u *Generate) FiledSet(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	var req generate.ModifyRequest

	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyField, err)
		return
	}

	req.AppId = appId
	req.UserId = userId
	req.Step = 2
	req.Database = db

	_, err := genService.ModifyGenerateConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    gin.H{},
	})
}

// FindColumnData 第三步② 获取列的数据进行校验等处理
// @Router /column/data [get]
func (u *Generate) FindColumnData(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	var req generate.ColumnRequest
	req.ColumnName = c.Query("col")
	req.AppId = appId
	req.UserId = userId
	req.Database = db

	response, err := genService.FindColumnData(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    response.GetItems(),
	})
}

// CreateDatabase 第四步 台账作成
// @Router /database/create [get]
func (u *Generate) CreateDatabase(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	// 获取导入的信息
	var req generate.FindRequest
	req.AppId = appId
	req.UserId = userId
	req.Database = db

	response, err := genService.FindGenerateConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	conf := response.GetGenConfig()

	// 添加台账
	if len(conf.GetDatastoreId()) == 0 {
		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

		var req datastore.AddRequest
		req.AppId = conf.AppId
		req.DatastoreName = conf.DatastoreName
		req.ApiKey = conf.ApiKey
		req.CanCheck = conf.CanCheck
		req.ShowInMenu = true
		req.NoStatus = true
		req.Encoding = "utf-8"
		req.Writer = userId
		req.Database = db

		// 从共通获取
		req.AppId = sessionx.GetCurrentApp(c)
		req.Writer = sessionx.GetAuthUserID(c)
		req.Database = sessionx.GetUserCustomer(c)

		response, err := datastoreService.AddDatastore(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddDatastore, err)
			return
		}

		conf.DatastoreId = response.GetDatastoreId()

		// 添加台账成功后保存日志到DB
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)     // 取共通用户名
		params["datastore_name"] = req.GetDatastoreName() // 新规的时候取传入参数
		params["api_key"] = req.GetApiKey()

		loggerx.ProcessLog(c, ActionAddDatastore, msg.L045, params)

		// 添加多语言数据
		langService := language.NewLanguageService("global", client.DefaultClient)

		languageReq := language.AddAppLanguageDataRequest{
			Domain:   domain,
			LangCd:   lang,
			AppId:    appId,
			Type:     "datastores",
			Key:      response.GetDatastoreId(),
			Value:    req.GetDatastoreName(),
			Writer:   userId,
			Database: db,
		}

		_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddDatastore, err)
			return
		}
	}

	// 添加字段
	if len(conf.GetFields()) > 0 {
		for _, f := range conf.GetFields() {
			if f.CanChange && !f.IsEmptyLine {
				fieldService := field.NewFieldService("database", client.DefaultClient)

				var req field.AddRequest
				req.AppId = appId
				req.DatastoreId = conf.DatastoreId
				req.FieldName = f.FieldName
				req.FieldType = f.FieldType
				req.FieldId = f.FieldId
				req.IsFixed = f.IsFixed
				req.IsRequired = f.IsRequired
				req.IsImage = f.IsImage
				req.IsCheckImage = f.IsCheckImage
				req.AsTitle = f.AsTitle
				req.Unique = f.Unique
				req.LookupAppId = appId
				req.LookupDatastoreId = f.LookupDatastoreId
				req.LookupFieldId = f.LookupFieldId
				req.UserGroupId = f.UserGroupId
				req.OptionId = f.OptionId
				req.MinLength = f.MinLength
				req.MaxLength = f.MaxLength
				req.MinValue = f.MinValue
				req.MaxValue = f.MaxValue
				req.DisplayOrder = f.DisplayOrder
				req.DisplayDigits = f.DisplayDigits
				req.Precision = f.Precision
				req.Prefix = f.Prefix
				req.ReturnType = f.ReturnType
				req.Formula = f.Formula
				req.Writer = userId
				req.Database = db

				response, err := fieldService.AddField(context.TODO(), &req)
				if err != nil {
					httpx.GinHTTPError(c, ActionAddField, err)
					return
				}

				// 添加字段成功后保存日志到DB
				params := make(map[string]string)
				params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
				params["field_name"] = req.GetFieldName()     // 新规的时候取传入参数
				params["datastore_name"] = conf.DatastoreName
				params["api_key"] = conf.ApiKey
				params["field_id"] = response.GetFieldId()

				loggerx.ProcessLog(c, ActionAddField, msg.L049, params)

				// 添加多语言数据
				langService := language.NewLanguageService("global", client.DefaultClient)

				languageReq := language.AddAppLanguageDataRequest{
					Domain:   domain,
					LangCd:   lang,
					AppId:    appId,
					Type:     "fields",
					Key:      req.GetDatastoreId() + "_" + response.GetFieldId(),
					Value:    req.GetFieldName(),
					Writer:   userId,
					Database: db,
				}

				_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
				if err != nil {
					httpx.GinHTTPError(c, ActionAddField, err)
					return
				}
			}
		}
	}

	var mReq generate.ModifyRequest

	mReq.AppId = appId
	mReq.UserId = userId
	mReq.DatastoreId = conf.DatastoreId // 更新台账ID
	mReq.Step = 3
	mReq.Database = db

	_, err = genService.ModifyGenerateConfig(context.TODO(), &mReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(userId, domain)

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    gin.H{},
	})
}

// FindConfig 获取配置
// @Router /config [get]
func (u *Generate) FindConfig(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	// 获取导入的信息
	var req generate.FindRequest
	req.AppId = appId
	req.UserId = userId
	req.Database = db

	response, err := genService.FindGenerateConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    response.GetGenConfig(),
	})
}

// CreateMapping 第五步 映射作成
// @Router /mapping/create [get]
func (u *Generate) CreateMapping(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	// 获取导入的信息
	var req generate.FindRequest
	req.AppId = appId
	req.UserId = userId
	req.Database = db

	response, err := genService.FindGenerateConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	conf := response.GetGenConfig()

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var mmReq datastore.AddMappingRequest
	// 从body中获取
	if err := c.BindJSON(&mmReq); err != nil {
		httpx.GinHTTPError(c, ActionAddDatastoreMapping, err)
		return
	}

	// 从共通获取
	mmReq.AppId = appId
	mmReq.DatastoreId = conf.GetDatastoreId()
	mmReq.Database = db

	mmResp, err := datastoreService.AddDatastoreMapping(context.TODO(), &mmReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDatastoreMapping, err)
		return
	}

	// 添加多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.AddAppLanguageDataRequest{
		Domain:   domain,
		LangCd:   lang,
		AppId:    appId,
		Type:     "mappings",
		Key:      conf.GetDatastoreId() + "_" + mmResp.GetMappingId(),
		Value:    mmReq.GetMappingName(),
		Writer:   userId,
		Database: db,
	}

	_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDatastoreMapping, err)
		return
	}

	var mReq generate.ModifyRequest

	mReq.AppId = appId
	mReq.UserId = userId
	mReq.MappingId = mmResp.GetMappingId()
	mReq.Step = 4
	mReq.Database = db

	_, err = genService.ModifyGenerateConfig(context.TODO(), &mReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(userId, domain)

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    gin.H{},
	})
}

// Complete 第五步 完成(删除记录)
// @Router /complete [get]
func (u *Generate) Complete(c *gin.Context) {
	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessStarted)

	userId := sessionx.GetAuthUserID(c)
	appId := sessionx.GetCurrentApp(c)
	db := sessionx.GetUserCustomer(c)

	// 开始导入到数据库中
	genService := generate.NewGenerateService("database", client.DefaultClient)

	// 获取导入的信息
	var req generate.DeleteRequest
	req.AppId = appId
	req.UserId = userId
	req.Database = db

	_, err := genService.DeleteGenerateConfig(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionUpload, err)
		return
	}

	loggerx.InfoLog(c, ActionUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    gin.H{},
	})
}
