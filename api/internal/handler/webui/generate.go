package webui

// import (
// 	"context"
// 	"encoding/csv"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"math"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/kataras/i18n"
// 	"github.com/micro/go-micro/v2/client"
// 	"github.com/micro/go-micro/v2/client/grpc"
// 	"golang.org/x/text/encoding/japanese"
// 	"golang.org/x/text/transform"
// 	"rxcsoft.cn/pit3/api/internal/common/filex"
// 	"rxcsoft.cn/pit3/api/internal/common/floatx"
// 	"rxcsoft.cn/pit3/api/internal/common/httpx"
// 	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
// 	"rxcsoft.cn/pit3/api/internal/common/loggerx"
// 	"rxcsoft.cn/pit3/api/internal/system/sessionx"
// 	"rxcsoft.cn/pit3/api/internal/system/wsx"
// 	"rxcsoft.cn/pit3/lib/msg"
// 	"rxcsoft.cn/pit3/srv/database/proto/feed"
// 	"rxcsoft.cn/pit3/srv/database/proto/item"
// 	"rxcsoft.cn/pit3/srv/manage/proto/user"
// 	"rxcsoft.cn/pit3/srv/task/proto/task"
// )

// type Generate struct{}

// // AddImportItem 映射导入数据
// // @Summary 映射导入数据
// // @description 调用srv中的item服务,映射导入数据
// // @Tags Import
// // @Accept json
// // @Security JWT
// // @Produce  json
// // @Success 200 {object} handler.Response
// // @Param files body blob true "文件数据信息"
// // @Failure 401 {object} handler.ErrorResponse
// // @Failure 403 {object} handler.ErrorResponse
// // @Failure 500 {object} handler.ErrorResponse
// // @Router /datastores/{d_id}/upload [post]
// func (i *Generate) GenerateUpload(c *gin.Context) {
// 	loggerx.InfoLog(c, ActionMappingUpload, loggerx.MsgProcessStarted)

// 	ct := grpc.NewClient(
// 		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
// 	)

// 	importService := feed.NewImportService("database", ct)

// 	var req feed.AddRequest

// 	appID := sessionx.GetCurrentApp(c)
// 	datastoreID := c.Param("d_id")
// 	mappingID := "generate"
// 	jobID := c.PostForm("job_id")
// 	encoding := c.PostForm("encoding")
// 	comma := c.PostForm("comma")
// 	comment := c.PostForm("comment")
// 	userID := sessionx.GetAuthUserID(c)
// 	db := sessionx.GetUserCustomer(c)

// 	// 获取上传的文件
// 	files, err := c.FormFile("file")
// 	if err != nil {
// 		httpx.GinHTTPError(c, ActionMappingUpload, err)
// 		return
// 	}

// 	// 文件类型检查
// 	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
// 		httpx.GinHTTPError(c, ActionMappingUpload, errors.New("このファイルタイプのアップロードはサポートされていません"))
// 		return
// 	}
// 	// 文件大小检查
// 	if !filex.CheckSize("csv", files.Size) {
// 		httpx.GinHTTPError(c, ActionMappingUpload, errors.New("ファイルサイズが設定サイズを超えています"))
// 		return
// 	}

// 	// 读取文件
// 	fs, err := files.Open()
// 	if err != nil {
// 		httpx.GinHTTPError(c, ActionMappingUpload, err)
// 		return
// 	}
// 	defer fs.Close()

// 	var r *csv.Reader
// 	// UTF-8格式的场合，直接读取
// 	if encoding == "UTF-8" {
// 		r = csv.NewReader(fs)
// 	} else {
// 		// ShiftJIS格式的场合，先转换为uft-8，再读取
// 		utfReader := transform.NewReader(fs, japanese.ShiftJIS.NewDecoder())
// 		r = csv.NewReader(utfReader)
// 	}
// 	r.LazyQuotes = true

// 	if comma == "," {
// 		r.Comma = 44 // 逗号
// 	} else {
// 		r.Comma = 9 // 制表符
// 	}

// 	if comment == "double" {
// 		r.Comment = 34 // 双引号
// 	} else {
// 		r.Comment = 39 // 单引号
// 	}

// 	var fileData [][]string

// 	// 针对大文件，一行一行的读取文件
// 	for {
// 		row, err := r.Read()
// 		if err != nil && err != io.EOF {
// 			loggerx.FailureLog(c, ActionMappingUpload, err.Error())
// 			return
// 		}
// 		if err == io.EOF {
// 			break
// 		}
// 		fileData = append(fileData, row)
// 	}

// 	header := fileData[0]
// 	header[0] = strings.Replace(header[0], "\ufeff", "", 1)
// 	for _, items := range fileData[1:] {
// 		// 判断数据项目数是否越界(超过header)
// 		if len(items) > len(header) {
// 			httpx.GinHTTPError(c, ActionMappingUpload, errors.New("incorrect number of row data items"))
// 			return
// 		}
// 		var imt feed.ImportItem
// 		imt.AppId = appID
// 		imt.DatastoreId = datastoreID
// 		imt.JobId = jobID
// 		imt.MappingId = mappingID
// 		imt.Items = make(map[string]string)
// 		for index, it := range items {
// 			imt.Items[header[index]] = it
// 		}
// 		req.Items = append(req.Items, &imt)
// 	}

// 	req.Writer = userID
// 	req.Database = db

// 	var opss client.CallOption = func(o *client.CallOptions) {
// 		o.RequestTimeout = time.Hour * 1
// 		o.DialTimeout = time.Hour * 1
// 	}

// 	response, err := importService.AddImportItem(context.TODO(), &req, opss)
// 	if err != nil {
// 		httpx.GinHTTPError(c, ActionMappingUpload, err)
// 		return
// 	}
// 	loggerx.SuccessLog(c, ActionMappingUpload, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionMappingUpload))

// 	loggerx.InfoLog(c, ActionMappingUpload, loggerx.MsgProcessEnded)
// 	c.JSON(200, httpx.Response{
// 		Status:  0,
// 		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, MappingProcessName, ActionMappingUpload)),
// 		Data:    response,
// 	})
// }

// func FindGenerateData(c *gin.Context, db, appID, datastoreID, mappingID, jobID, domain, lang, userID string, owners, roles []string) {
// 	loggerx.InfoLog(c, ActionMappingUpload, fmt.Sprintf("Process importData:%s", loggerx.MsgProcessStarted))

// 	ct := grpc.NewClient(
// 		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
// 	)

// 	importService := feed.NewImportService("database", ct)
// 	itemService := item.NewItemService("database", ct)

// 	var req feed.ImportItemsRequest
// 	req.AppId = appID
// 	req.DatastoreId = datastoreID
// 	req.JobId = jobID
// 	req.MappingId = mappingID
// 	req.Writer = userID
// 	req.Database = db

// 	// 发送消息 开始读取数据
// 	jobx.ModifyTask(task.ModifyRequest{
// 		JobId:       jobID,
// 		Message:     i18n.Tr(lang, "job.J_024"),
// 		CurrentStep: "data-ready",
// 		Database:    db,
// 	})

// 	wsx.SendMsg("uploader", "data-ready", jobID, userID)

// 	loggerx.InfoLog(c, ActionMappingUpload, fmt.Sprintf("Process FindImportItems:%s", loggerx.MsgProcessStarted))

// 	var opss client.CallOption = func(o *client.CallOptions) {
// 		o.RequestTimeout = time.Hour * 1
// 		o.DialTimeout = time.Hour * 1
// 	}
// 	response, err := importService.FindImportItems(context.TODO(), &req, opss)
// 	if err != nil {
// 		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

// 		// 发送消息 数据查询错误
// 		jobx.ModifyTask(task.ModifyRequest{
// 			JobId:       jobID,
// 			Message:     i18n.Tr(lang, "job.J_006"),
// 			CurrentStep: "data-ready",
// 			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
// 			ErrorFile: &task.File{
// 				Url:  path.MediaLink,
// 				Name: path.Name,
// 			},
// 			Database: db,
// 		})

// 		wsx.SendMsg("uploader", "data-ready", jobID, userID)
// 		return
// 	}
// 	loggerx.InfoLog(c, ActionMappingUpload, fmt.Sprintf("Process FindImportItems:%s", loggerx.MsgProcessEnded))

// 	items := response.GetItems()

// 	fields := getFields(db, datastoreID, appID, roles, false, false)
// 	// 获取当前app的语言数据
// 	langData := getAppLanguage(db, appID, lang, domain)
// 	mappingInfo, e1 := getMappingInfo(db, datastoreID, mappingID)
// 	if e1 != nil {
// 		// 清空导入的数据
// 		clearTemp(db, jobID)

// 		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

// 		// 发送消息 映射情报获取错误
// 		jobx.ModifyTask(task.ModifyRequest{
// 			JobId:       jobID,
// 			Message:     i18n.Tr(lang, "job.J_005"),
// 			CurrentStep: "data-ready",
// 			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
// 			ErrorFile: &task.File{
// 				Url:  path.MediaLink,
// 				Name: path.Name,
// 			},
// 			Database: db,
// 		})

// 		wsx.SendMsg("uploader", "data-ready", jobID, userID)
// 		return
// 	}

// 	var userList []*user.User
// 	lookupItems := make(map[string][]*item.Item)

// 	for _, mp := range mappingInfo.MappingRule {
// 		if mp.DataType == "user" {
// 			if len(userList) == 0 {
// 				userList = getAllUser(c, db, appID, domain)
// 			}
// 		}

// 		if mp.DataType == "lookup" {
// 			if lookupItems[mp.FromKey] == nil || len(lookupItems[mp.FromKey]) == 0 {
// 				for _, fl := range fields {
// 					if fl.FieldId == mp.FromKey {
// 						findAccessKeys := sessionx.GetUserAccessKeys(c, fl.GetLookupDatastoreId(), "R")
// 						itemList := itemx.GetLookupItems(db, fl.GetLookupDatastoreId(), fl.GetAppId(), findAccessKeys)
// 						lookupItems[fl.FieldId] = itemList
// 					}
// 				}
// 			}
// 		}
// 	}

// 	errorList := []string{}

// 	// 发送消息 开始进行数据上传（包括数据验证和上传错误）
// 	jobx.ModifyTask(task.ModifyRequest{
// 		JobId:       jobID,
// 		Message:     i18n.Tr(lang, "job.J_046"),
// 		CurrentStep: "build-check-data",
// 		Database:    db,
// 	})

// 	wsx.SendMsg("uploader", "build-check-data", jobID, userID)

// 	var importData []*item.ImportData
// Loop:
// 	for i, it := range items {
// 		index := i + 1
// 		query := make(map[string]*item.Value)
// 		change := make(map[string]*item.Value)
// 		for _, mp := range mappingInfo.MappingRule {

// 			switch mp.DataType {
// 			case "text", "textarea":
// 				if mp.PrimaryKey {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断值是否为空
// 						if value != "" {
// 							// 判断是否需要禁止文字check
// 							if mp.Special {
// 								if specialCheck(value) {
// 									query[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    value,
// 									}
// 								} else {
// 									// 报禁止文字错误
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には特殊文字があります。", index, mp.ToKey))
// 									break Loop
// 								}
// 							} else {
// 								query[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    value,
// 								}
// 							}
// 						} else {
// 							// 主键不能为空
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 							break Loop
// 						}
// 					} else {
// 						// 主键不能为空
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 						break Loop
// 					}
// 				} else {
// 					// 判断tokey是否有值，没有值直接使用默认值
// 					if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 						change[mp.FromKey] = &item.Value{
// 							DataType: mp.DataType,
// 							Value:    mp.DefaultValue,
// 						}
// 					} else {
// 						// 判断是否导入该字段的数据
// 						if value, ok := it.Items[mp.ToKey]; ok {
// 							// 判断是否需要必须check
// 							if mp.IsRequired {
// 								// 判断值是否为空
// 								if value != "" {
// 									// 判断是否需要禁止文字check
// 									if mp.Special {
// 										if specialCheck(value) {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    value,
// 											}
// 										} else {
// 											// 报禁止文字错误
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には特殊文字があります。", index, mp.ToKey))
// 											break Loop
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								} else {
// 									// 该字段是必须入力项目，一定要输入值
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 									break Loop
// 								}
// 							} else {
// 								// 判断值是否为空
// 								if value != "" {
// 									// 判断是否需要禁止文字check
// 									if mp.Special {
// 										if specialCheck(value) {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    value,
// 											}
// 										} else {
// 											// 报禁止文字错误
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には特殊文字があります。", index, mp.ToKey))
// 											break Loop
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								} else {
// 									if len(mp.DefaultValue) > 0 {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    mp.DefaultValue,
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								}
// 							}
// 						} else {
// 							// 判断是否需要必须check
// 							if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							} else {
// 								change[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    mp.DefaultValue,
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "number":
// 				if mp.PrimaryKey {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断值是否为空
// 						if value != "" {
// 							// 判断是否是数字类型
// 							_, e := strconv.ParseFloat(value, 64)
// 							if e != nil {
// 								fmt.Printf("number format has error : %v", e)
// 								// 不是正确的数字类型的数据。
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", index, mp.ToKey))
// 								break Loop
// 							} else {
// 								query[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    value,
// 								}
// 							}
// 						} else {
// 							// 主键不能为空
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 							break Loop
// 						}
// 					} else {
// 						// 主键不能为空
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 						break Loop
// 					}
// 				} else {
// 					// 判断tokey是否有值，没有值直接使用默认值
// 					if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 						// 判断是否是数字类型
// 						_, e := strconv.ParseInt(mp.DefaultValue, 10, 64)
// 						if e != nil {
// 							fmt.Printf("number format has error : %v", e)
// 							// 不是正确的数字类型的数据。
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", index, mp.ToKey))
// 							break Loop
// 						} else {
// 							change[mp.FromKey] = &item.Value{
// 								DataType: mp.DataType,
// 								Value:    mp.DefaultValue,
// 							}
// 						}
// 					} else {
// 						// 判断是否导入该字段的数据
// 						if value, ok := it.Items[mp.ToKey]; ok {
// 							// 判断是否需要必须check
// 							if mp.IsRequired {
// 								// 判断值是否为空
// 								if value != "" {
// 									// 判断是否是数字类型
// 									nv, e := strconv.ParseFloat(value, 64)
// 									if e != nil {
// 										fmt.Printf("number format has error : %v", e)
// 										// 不是正确的数字类型的数据。
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", index, mp.ToKey))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    floatx.ToFixedString(nv, GetPrecision(fields, mp.GetFromKey())),
// 										}
// 									}
// 								} else {
// 									// 该字段是必须入力项目，一定要输入值
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 									break Loop
// 								}
// 							} else {
// 								// 判断值是否为空
// 								if value != "" {
// 									// 判断是否是数字类型
// 									nv, e := strconv.ParseFloat(value, 64)
// 									if e != nil {
// 										fmt.Printf("number format has error : %v", e)
// 										// 不是正确的数字类型的数据。
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", index, mp.ToKey))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    floatx.ToFixedString(nv, GetPrecision(fields, mp.GetFromKey())),
// 										}
// 									}
// 								} else {
// 									if len(mp.DefaultValue) > 0 {
// 										// 判断是否是数字类型
// 										_, e := strconv.ParseInt(mp.DefaultValue, 10, 64)
// 										if e != nil {
// 											fmt.Printf("number format has error : %v", e)
// 											// 不是正确的数字类型的数据。
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", index, mp.ToKey))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    mp.DefaultValue,
// 											}
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								}
// 							}
// 						} else {
// 							// 判断是否需要必须check
// 							if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							} else {
// 								// 判断是否是数字类型
// 								_, e := strconv.ParseInt(mp.DefaultValue, 10, 64)
// 								if e != nil {
// 									fmt.Printf("number format has error : %v", e)
// 									// 不是正确的数字类型的数据。
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は数値ではありません。", index, mp.ToKey))
// 									break Loop
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    mp.DefaultValue,
// 									}
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "time":
// 				if mp.PrimaryKey {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断值是否为空
// 						if value != "" {
// 							// 判断是否是时间字符串
// 							_, e := time.Parse("15:04:05", value)
// 							if e != nil {
// 								fmt.Printf("time format has error : %v", e)
// 								// 不是正确的时间类型的数据。
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", index, mp.ToKey))
// 								break Loop
// 							} else {
// 								query[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    value,
// 								}
// 							}
// 						} else {
// 							// 主键不能为空
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 							break Loop
// 						}
// 					} else {
// 						// 主键不能为空
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 						break Loop
// 					}
// 				} else {
// 					// 判断tokey是否有值，没有值直接使用默认值
// 					if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 						// 判断是否是时间字符串
// 						_, e := time.Parse("15:04:05", mp.DefaultValue)
// 						if e != nil {
// 							fmt.Printf("time format has error : %v", e)
// 							// 不是正确的时间类型的数据。
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", index, mp.ToKey))
// 							break Loop
// 						} else {
// 							change[mp.FromKey] = &item.Value{
// 								DataType: mp.DataType,
// 								Value:    mp.DefaultValue,
// 							}
// 						}
// 					} else {
// 						// 判断是否导入该字段的数据
// 						if value, ok := it.Items[mp.ToKey]; ok {
// 							// 判断是否需要必须check
// 							if mp.IsRequired {
// 								// 判断值是否为空
// 								if value != "" {
// 									// 判断是否是时间字符串
// 									_, e := time.Parse("15:04:05", value)
// 									if e != nil {
// 										fmt.Printf("time format has error : %v", e)
// 										// 不是正确的时间类型的数据。
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", index, mp.ToKey))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								} else {
// 									// 该字段是必须入力项目，一定要输入值
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 									break Loop
// 								}
// 							} else {
// 								// 判断值是否为空
// 								if value != "" {
// 									// 判断是否是时间字符串
// 									_, e := time.Parse("15:04:05", value)
// 									if e != nil {
// 										fmt.Printf("time format has error : %v", e)
// 										// 不是正确的时间类型的数据。
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", index, mp.ToKey))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								} else {
// 									if len(mp.DefaultValue) > 0 {
// 										// 判断是否是时间字符串
// 										_, e := time.Parse("15:04:05", mp.DefaultValue)
// 										if e != nil {
// 											fmt.Printf("time format has error : %v", e)
// 											// 不是正确的时间类型的数据。
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", index, mp.ToKey))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    mp.DefaultValue,
// 											}
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								}
// 							}
// 						} else {
// 							// 判断是否需要必须check
// 							if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							} else {
// 								// 判断是否是时间字符串
// 								_, e := time.Parse("15:04:05", mp.DefaultValue)
// 								if e != nil {
// 									fmt.Printf("time format has error : %v", e)
// 									// 不是正确的时间类型的数据。
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な時間ではありません。", index, mp.ToKey))
// 									break Loop
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    mp.DefaultValue,
// 									}
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "date":
// 				if mp.PrimaryKey {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断值是否为空
// 						if value != "" {
// 							// 判断是否需要格式化日期
// 							if mp.Format != "" {
// 								ti, e := time.Parse(mp.Format, value)
// 								if e != nil {
// 									fmt.Printf("date format has error : %v", e)
// 									// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 									break Loop
// 								} else {
// 									query[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    ti.Format("2006-01-02"),
// 									}
// 								}
// 							} else {
// 								ti, e := time.Parse("2006-01-02", value)
// 								if e != nil {
// 									fmt.Printf("date format has error : %v", e)
// 									// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 									break Loop
// 								} else {
// 									query[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    ti.Format("2006-01-02"),
// 									}
// 								}
// 							}
// 						} else {
// 							// 主键不能为空
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 							break Loop
// 						}
// 					} else {
// 						// 主键不能为空
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] が主キーで、空にすることはできません。", index, mp.ToKey))
// 						break Loop
// 					}
// 				} else {
// 					// 判断tokey是否有值，没有值直接使用默认值
// 					if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 						ti, e := time.Parse("2006-01-02", mp.DefaultValue)
// 						if e != nil {
// 							fmt.Printf("date format has error : %v", e)
// 							// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 							break Loop
// 						} else {
// 							change[mp.FromKey] = &item.Value{
// 								DataType: mp.DataType,
// 								Value:    ti.Format("2006-01-02"),
// 							}
// 						}
// 					} else {
// 						// 判断是否导入该字段的数据
// 						if value, ok := it.Items[mp.ToKey]; ok {
// 							// 判断是否需要必须check
// 							if mp.IsRequired {
// 								// 判断值是否为空
// 								if value != "" {
// 									// 判断是否需要格式化日期
// 									if mp.Format != "" {
// 										ti, e := time.Parse(mp.Format, value)
// 										if e != nil {
// 											fmt.Printf("date format has error : %v", e)
// 											// 日期格式化出错
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    ti.Format("2006-01-02"),
// 											}
// 										}
// 									} else {
// 										ti, e := time.Parse("2006-01-02", value)
// 										if e != nil {
// 											fmt.Printf("date format has error : %v", e)
// 											// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    ti.Format("2006-01-02"),
// 											}
// 										}
// 									}
// 								} else {
// 									// 该字段是必须入力项目，一定要输入值
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 									break Loop
// 								}
// 							} else {
// 								if value != "" {
// 									// 判断是否需要格式化日期
// 									if mp.Format != "" {
// 										ti, e := time.Parse(mp.Format, value)
// 										if e != nil {
// 											fmt.Printf("date format has error : %v", e)
// 											// 日期格式化出错
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    ti.Format("2006-01-02"),
// 											}
// 										}
// 									} else {
// 										ti, e := time.Parse("2006-01-02", value)
// 										if e != nil {
// 											fmt.Printf("date format has error : %v", e)
// 											// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    ti.Format("2006-01-02"),
// 											}
// 										}
// 									}
// 								} else {
// 									if len(mp.DefaultValue) > 0 {
// 										ti, e := time.Parse("2006-01-02", mp.DefaultValue)
// 										if e != nil {
// 											fmt.Printf("date format has error : %v", e)
// 											// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    ti.Format("2006-01-02"),
// 											}
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    value,
// 										}
// 									}
// 								}
// 							}
// 						} else {
// 							// 判断是否需要必须check
// 							if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							} else {
// 								ti, e := time.Parse("2006-01-02", mp.DefaultValue)
// 								if e != nil {
// 									fmt.Printf("date format has error : %v", e)
// 									// 日期格式化出错，可能原因，不是日期类型的数据，或者当前的格式与对应的日期不匹配。
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] は有効な日付ではありません。", index, mp.ToKey))
// 									break Loop
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    ti.Format("2006-01-02"),
// 									}
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "switch":
// 				// 判断tokey是否有值，没有值直接使用默认值
// 				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 					_, err := strconv.ParseBool(mp.DefaultValue)
// 					if err != nil {
// 						// 存在check错误
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 						break Loop
// 					} else {
// 						change[mp.FromKey] = &item.Value{
// 							DataType: mp.DataType,
// 							Value:    mp.DefaultValue,
// 						}
// 					}
// 				} else {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断是否需要必须check
// 						if mp.IsRequired {
// 							// 判断值是否为空
// 							if value != "" {
// 								_, err := strconv.ParseBool(value)
// 								if err != nil {
// 									// 存在check错误
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", index, mp.ToKey, value))
// 									break Loop
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    value,
// 									}
// 								}
// 							} else {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							}
// 						} else {
// 							if len(mp.DefaultValue) > 0 {
// 								_, err := strconv.ParseBool(mp.DefaultValue)
// 								if err != nil {
// 									// 存在check错误
// 									errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 									break Loop
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    mp.DefaultValue,
// 									}
// 								}
// 							} else {
// 								change[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    value,
// 								}
// 							}
// 						}
// 					} else {
// 						// 判断是否需要必须check
// 						if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 							// 该字段是必须入力项目，一定要输入值
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 							break Loop
// 						} else {
// 							_, err := strconv.ParseBool(mp.DefaultValue)
// 							if err != nil {
// 								// 存在check错误
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な値 [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 								break Loop
// 							} else {
// 								change[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    mp.DefaultValue,
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "user":
// 				// 判断tokey是否有值，没有值直接使用默认值
// 				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 					uid := reTranUser(mp.DefaultValue, userList)
// 					if uid == "" {
// 						// 存在check错误
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 						break Loop
// 					} else {
// 						change[mp.FromKey] = &item.Value{
// 							DataType: mp.DataType,
// 							Value:    uid,
// 						}
// 					}
// 				} else {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断是否需要必须check
// 						if mp.IsRequired {
// 							// 判断值是否为空
// 							if value != "" {
// 								// 判断是单，还是多个用户
// 								inx := strings.LastIndex(value, ",")
// 								if inx == -1 {
// 									uid := reTranUser(value, userList)
// 									// 判断是否需要进行存在check
// 									if mp.Exist {
// 										if uid == "" {
// 											// 存在check错误
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", index, mp.ToKey, value))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    uid,
// 											}
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    uid,
// 										}
// 									}
// 								} else {
// 									users := strings.Split(value, ",")
// 									var uids []string
// 									for _, u := range users {
// 										uid := reTranUser(u, userList)
// 										// 判断是否需要进行存在check
// 										if mp.Exist {
// 											if uid == "" {
// 												// 存在check错误
// 												errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", index, mp.ToKey, u))
// 												break Loop
// 											} else {
// 												uids = append(uids, uid)
// 											}
// 										} else {
// 											uids = append(uids, uid)
// 										}
// 									}
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    strings.Join(uids, ","),
// 									}
// 								}
// 							} else {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							}
// 						} else {
// 							// 判断值是否为空
// 							if value != "" {
// 								// 判断是单，还是多个用户
// 								inx := strings.LastIndex(value, ",")
// 								if inx == -1 {
// 									uid := reTranUser(value, userList)
// 									// 判断是否需要进行存在check
// 									if mp.Exist {
// 										if uid == "" {
// 											// 存在check错误
// 											errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", index, mp.ToKey, value))
// 											break Loop
// 										} else {
// 											change[mp.FromKey] = &item.Value{
// 												DataType: mp.DataType,
// 												Value:    uid,
// 											}
// 										}
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    uid,
// 										}
// 									}
// 								} else {
// 									users := strings.Split(value, ",")
// 									var uids []string
// 									for _, u := range users {
// 										uid := reTranUser(u, userList)
// 										// 判断是否需要进行存在check
// 										if mp.Exist {
// 											if uid == "" {
// 												// 存在check错误
// 												errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", index, mp.ToKey, u))
// 												break Loop
// 											} else {
// 												uids = append(uids, uid)
// 											}
// 										} else {
// 											uids = append(uids, uid)
// 										}
// 									}
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    strings.Join(uids, ","),
// 									}
// 								}
// 							} else {
// 								if len(mp.DefaultValue) > 0 {
// 									uid := reTranUser(mp.DefaultValue, userList)
// 									if uid == "" {
// 										// 存在check错误
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    uid,
// 										}
// 									}
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    value,
// 									}
// 								}
// 							}
// 						}
// 					} else {
// 						// 判断是否需要必须check
// 						if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 							// 该字段是必须入力项目，一定要输入值
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 							break Loop
// 						} else {
// 							uid := reTranUser(mp.DefaultValue, userList)
// 							if uid == "" {
// 								// 存在check错误
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効なユーザ [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 								break Loop
// 							} else {
// 								change[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    uid,
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "options":
// 				// 判断tokey是否有值，没有值直接使用默认值
// 				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 					oid := getOptionValueByName(db, appID, langData, mp.DefaultValue)
// 					if oid == "" {
// 						// 存在check错误
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 						break Loop
// 					} else {
// 						change[mp.FromKey] = &item.Value{
// 							DataType: mp.DataType,
// 							Value:    oid,
// 						}
// 					}
// 				} else {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断是否需要必须check
// 						if mp.IsRequired {
// 							// 判断值是否为空
// 							if value != "" {
// 								// 获取选项ID
// 								oid := getOptionValueByName(db, appID, langData, value)
// 								// 判断是否需要进行存在check
// 								if mp.Exist {
// 									if oid == "" {
// 										// 存在check错误
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", index, mp.ToKey, value))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    oid,
// 										}
// 									}
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    oid,
// 									}
// 								}
// 							} else {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							}
// 						} else {
// 							// 判断值是否为空
// 							if value != "" {
// 								// 获取选项ID
// 								oid := getOptionValueByName(db, appID, langData, value)
// 								// 判断是否需要进行存在check
// 								if mp.Exist {
// 									if oid == "" {
// 										// 存在check错误
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", index, mp.ToKey, value))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    oid,
// 										}
// 									}
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    oid,
// 									}
// 								}
// 							} else {
// 								if len(mp.DefaultValue) > 0 {
// 									oid := getOptionValueByName(db, appID, langData, mp.DefaultValue)
// 									if oid == "" {
// 										// 存在check错误
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    oid,
// 										}
// 									}
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    value,
// 									}
// 								}
// 							}
// 						}
// 					} else {
// 						// 判断是否需要必须check
// 						if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 							// 该字段是必须入力项目，一定要输入值
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 							break Loop
// 						} else {
// 							oid := getOptionValueByName(db, appID, langData, mp.DefaultValue)
// 							if oid == "" {
// 								// 存在check错误
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。 [ %s ] には有効なオプション [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 								break Loop
// 							} else {
// 								change[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    oid,
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "lookup":
// 				// 判断tokey是否有值，没有值直接使用默认值
// 				if len(mp.ToKey) == 0 && len(mp.DefaultValue) > 0 {
// 					lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, mp.DefaultValue)
// 					if lv == "" {
// 						// 存在check错误
// 						errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 						break Loop
// 					} else {
// 						change[mp.FromKey] = &item.Value{
// 							DataType: mp.DataType,
// 							Value:    lv,
// 						}
// 					}
// 				} else {
// 					// 判断是否导入该字段的数据
// 					if value, ok := it.Items[mp.ToKey]; ok {
// 						// 判断是否需要必须check
// 						if mp.IsRequired {
// 							// 判断值是否为空
// 							if value != "" {
// 								// 获取关联数据
// 								lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, value)
// 								// 判断是否需要进行存在check
// 								if mp.Exist {
// 									if lv == "" {
// 										// 存在check错误
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", index, mp.ToKey, value))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    lv,
// 										}
// 									}
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    lv,
// 									}
// 								}
// 							} else {
// 								// 该字段是必须入力项目，一定要输入值
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 								break Loop
// 							}
// 						} else {
// 							// 判断值是否为空
// 							if value != "" {
// 								// 获取关联数据
// 								lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, value)
// 								// 判断是否需要进行存在check
// 								if mp.Exist {
// 									if lv == "" {
// 										// 存在check错误
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", index, mp.ToKey, value))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    lv,
// 										}
// 									}
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    lv,
// 									}
// 								}
// 							} else {
// 								if len(mp.DefaultValue) > 0 {
// 									lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, mp.DefaultValue)
// 									if lv == "" {
// 										// 存在check错误
// 										errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 										break Loop
// 									} else {
// 										change[mp.FromKey] = &item.Value{
// 											DataType: mp.DataType,
// 											Value:    lv,
// 										}
// 									}
// 								} else {
// 									change[mp.FromKey] = &item.Value{
// 										DataType: mp.DataType,
// 										Value:    "",
// 									}
// 								}

// 							}
// 						}
// 					} else {
// 						// 判断是否需要必须check
// 						if mp.IsRequired && len(mp.DefaultValue) == 0 {
// 							// 该字段是必须入力项目，一定要输入值
// 							errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] フィールドは必須入力です。", index, mp.ToKey))
// 							break Loop
// 						} else {
// 							lv := getLookupInfo(fields, lookupItems[mp.FromKey], mp.FromKey, mp.DefaultValue)
// 							if lv == "" {
// 								// 存在check错误
// 								errorList = append(errorList, fmt.Sprintf("行 %d はエラーです。[ %s ] には有効な関連項目 [ %s ] がありません。", index, mp.ToKey, mp.DefaultValue))
// 								break Loop
// 							} else {
// 								change[mp.FromKey] = &item.Value{
// 									DataType: mp.DataType,
// 									Value:    lv,
// 								}
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}

// 		importData = append(importData, &item.ImportData{
// 			Query:  query,
// 			Change: change,
// 			Index:  int64(index),
// 		})

// 	}

// 	if len(errorList) > 0 {
// 		path := filex.WriteAndSaveFile(domain, appID, errorList)

// 		// 发送消息 出现错误
// 		jobx.ModifyTask(task.ModifyRequest{
// 			JobId:       jobID,
// 			Message:     i18n.Tr(lang, "job.J_003"),
// 			CurrentStep: "build-check-data",
// 			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
// 			ErrorFile: &task.File{
// 				Url:  path.MediaLink,
// 				Name: path.Name,
// 			},
// 			Database: db,
// 		})

// 		wsx.SendMsg("uploader", "build-check-data", jobID, userID)
// 		// 清空导入的数据
// 		clearTemp(db, jobID)
// 		return
// 	}

// 	splitLength := float64(len(importData)) / 2000.0
// 	sp := math.Ceil(splitLength)

// 	datas := splitImportData(c, importData, int64(sp))

// 	var hasError int64
// 	// 总的件数
// 	totalCount := float64(len(importData))
// 	// 已经导入的件数
// 	var insertedCount float64
// 	var updatedCount float64

// 	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
// 	defer cancel()

// 	for _, dt := range datas {
// 		var req item.MappingImportRequest
// 		req.AppId = appID
// 		req.DatastoreId = datastoreID
// 		req.MappingType = mappingInfo.MappingType
// 		req.UpdateType = mappingInfo.UpdateType
// 		req.Datas = dt
// 		req.Database = db
// 		req.Writer = userID
// 		req.Owners = owners

// 		loggerx.InfoLog(c, ActionMappingUpload, fmt.Sprintf("Process MappingImport:%s", loggerx.MsgProcessStarted))
// 		res, _ := itemService.MappingImport(ctx, &req)
// 		loggerx.InfoLog(c, ActionMappingUpload, fmt.Sprintf("Process MappingImport:%s", loggerx.MsgProcessEnded))

// 		if len(res.GetResult().GetErrors()) > 0 {
// 			hasError++
// 			for _, e := range res.GetResult().GetErrors() {
// 				eMsg := "第{0}〜{1}行目でエラーが発生しました。エラー内容：{2}"
// 				fieldErrorMsg := "第{0}行目でエラーが発生しました。フィールド名：[{1}]、エラー内容：{2}"
// 				noFieldErrorMsg := "第{0}行目でエラーが発生しました。エラー内容：{1}"
// 				if len(e.FieldId) == 0 {
// 					if e.CurrentLine != 0 {
// 						es, _ := msg.Format(noFieldErrorMsg, strconv.FormatInt(e.CurrentLine, 10), e.ErrorMsg)
// 						errorList = append(errorList, es)
// 					} else {
// 						es, _ := msg.Format(eMsg, strconv.FormatInt(e.FirstLine, 10), strconv.FormatInt(e.LastLine, 10), e.ErrorMsg)
// 						errorList = append(errorList, es)
// 					}
// 				} else {
// 					es, _ := msg.Format(fieldErrorMsg, strconv.FormatInt(e.CurrentLine, 10), langx.Translates(langData, "fields", datastoreID+"_"+e.FieldId), e.ErrorMsg)
// 					errorList = append(errorList, es)
// 				}
// 			}
// 		}

// 		insertedCount += float64(res.GetResult().GetInsert())
// 		updatedCount += float64(res.GetResult().GetModify())

// 		importMsg, _ := json.Marshal(map[string]interface{}{
// 			"total":    totalCount,
// 			"inserted": insertedCount,
// 			"updated":  updatedCount,
// 		})

// 		progress := (insertedCount + updatedCount) / totalCount * 100

// 		// 发送消息 收集上传结果
// 		jobx.ModifyTask(task.ModifyRequest{
// 			JobId:       jobID,
// 			Message:     string(importMsg),
// 			CurrentStep: "upload",
// 			Progress:    int64(progress),
// 			Insert:      int64(insertedCount),
// 			Update:      int64(updatedCount),
// 			Total:       int64(totalCount),
// 			Database:    db,
// 		})

// 		wsx.SendMsg("uploader", "upload", jobID, userID)

// 	}

// 	if len(errorList) > 0 {
// 		path := filex.WriteAndSaveFile(domain, appID, errorList)

// 		// 发送消息 出现错误
// 		jobx.ModifyTask(task.ModifyRequest{
// 			JobId:       jobID,
// 			Message:     i18n.Tr(lang, "job.J_008"),
// 			CurrentStep: "upload",
// 			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
// 			ErrorFile: &task.File{
// 				Url:  path.MediaLink,
// 				Name: path.Name,
// 			},
// 			Database: db,
// 		})

// 		wsx.SendMsg("uploader", "upload", jobID, userID)

// 		// 清空导入的数据
// 		clearTemp(db, jobID)
// 		return
// 	}

// 	// 发送消息 全部上传成功
// 	jobx.ModifyTask(task.ModifyRequest{
// 		JobId:       jobID,
// 		Message:     i18n.Tr(lang, "job.J_009"),
// 		CurrentStep: "end",
// 		Progress:    100,
// 		EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
// 		Database:    db,
// 	})

// 	wsx.SendMsg("uploader", "end", jobID, userID)

// 	// 清空导入的数据
// 	clearTemp(db, jobID)

// 	loggerx.InfoLog(c, ActionMappingUpload, fmt.Sprintf("Process importData:%s", loggerx.MsgProcessEnded))
// }
