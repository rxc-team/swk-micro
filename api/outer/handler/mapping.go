package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/spf13/cast"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"rxcsoft.cn/pit3/api/outer/common/excelx"
	"rxcsoft.cn/pit3/api/outer/common/filex"
	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/outer/common/logic/langx"
	"rxcsoft.cn/pit3/api/outer/common/logic/mappingx"
	"rxcsoft.cn/pit3/api/outer/common/typesx"
	"rxcsoft.cn/pit3/api/outer/system/jobx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/import/proto/upload"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// Mapping Mapping
type Mapping struct{}

// log出力
const (
	MappingProcessName    = "Mapping"
	ActionMappingUpload   = "MappingUpload"
	ActionMappingDownload = "MappingDownload"
)

// AddImportItem 映射导入数据
// @Router /datastores/{d_id}/upload [post]
func (i *Mapping) MappingUpload(c *gin.Context) {
	loggerx.InfoLog(c, ActionMappingUpload, loggerx.MsgProcessStarted)

	appID := sessionx.GetCurrentApp(c)
	datastoreID := c.Param("d_id")
	jobID := c.PostForm("job_id")
	domain := sessionx.GetUserDomain(c)
	// lang := sessionx.GetCurrentLanguage(c)
	userID := sessionx.GetAuthUserID(c)
	db := sessionx.GetUserCustomer(c)
	appRoot := "app_" + appID
	// 时间戳
	timestamp := time.Now().Format("20060102150405")

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "csv file import(mapping)",
		Origin:       "apps." + appID + ".datastores." + datastoreID,
		UserId:       userID,
		ShowProgress: true,
		// Message:      i18n.Tr(lang, "job.J_014"),
		Message:     "create a job",
		TaskType:    "ds-csv-import",
		Steps:       []string{"start", "data-ready", "build-check-data", "upload", "end"},
		CurrentStep: "start",
		Database:    db,
		AppId:       appID,
	})

	// 超级域名
	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionMappingUpload, err)
		return
	}

	// 获取上传的文件
	files, err := c.FormFile("file")
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionMappingUpload, err)
		return
	}

	// 文件类型检查
	if !filex.CheckSupport("csv", files.Header.Get("content-type")) {
		path := filex.WriteAndSaveFile(domain, appID, []string{fmt.Sprintf("the csv file type [%v] is not supported", files.Header.Get("content-type"))})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionMappingUpload, errors.New("このファイルタイプのアップロードはサポートされていません"))
		return
	}
	// 文件大小检查
	if !filex.CheckSize(domain, "csv", files.Size) {
		path := filex.WriteAndSaveFile(domain, appID, []string{"the csv file ファイルサイズが設定サイズを超えています"})
		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionMappingUpload, errors.New("ファイルサイズが設定サイズを超えています"))
		return
	}

	fo, err := files.Open()
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionMappingUpload, err)
		return
	}

	filePath := path.Join(appRoot, "temp", "temp_"+timestamp+files.Filename)
	file, err := minioClient.SavePublicObject(fo, filePath, files.Header.Get("content-type"))
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionMappingUpload, err)
		return
	}

	// 文件开始上传
	err = mappingUpload(c, file.Name)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_053"),
			Message:     "An error occurred while uploading the file",
			CurrentStep: "data-ready",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		httpx.GinHTTPError(c, ActionMappingUpload, err)
		return
	}

	loggerx.InfoLog(c, ActionMappingUpload, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, MappingProcessName, ActionMappingUpload)),
		Data:    gin.H{},
	})
}

// Download 获取台账中的所有数据,以csv文件的方式下载
// @Router /datastores/{d_id}/items [get]
func (i *Mapping) MappingDownload(c *gin.Context) {

	type (
		// DownloadRequest 下载
		DownloadRequest struct {
			ItemCondition item.ItemsRequest `json:"item_condition" bson:"item_condition"`
		}
	)

	loggerx.InfoLog(c, ActionMappingDownload, loggerx.MsgProcessStarted)

	jobID := c.Query("job_id")
	mappingID := c.Query("mapping_id")
	datastoreID := c.Param("d_id")
	appID := sessionx.GetCurrentApp(c)
	owners := sessionx.GetUserAccessKeys(c, datastoreID, "R")
	userID := sessionx.GetAuthUserID(c)
	roles := sessionx.GetUserRoles(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	db := sessionx.GetUserCustomer(c)
	encoding := "utf-8"
	fileType := "csv"
	appRoot := "app_" + appID

	// 从body中获取参数
	var request DownloadRequest
	if err := c.BindJSON(&request); err != nil {
		httpx.GinHTTPError(c, ActionMappingDownload, err)
		return
	}

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "datastore file download(mapping)",
		Origin:       "apps." + appID + ".datastores." + datastoreID,
		UserId:       userID,
		ShowProgress: false,
		Message:      "create a job",
		TaskType:     "ds-csv-download",
		Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	// 发送消息 开始编辑数据
	jobx.ModifyTask(task.ModifyRequest{
		JobId: jobID,
		// Message:     i18n.Tr(lang, "job.J_012"),
		Message:     "build data",
		CurrentStep: "build-data",
		Database:    db,
	}, userID)

	cReq := item.CountRequest{
		AppId:         appID,
		DatastoreId:   datastoreID,
		ConditionList: request.ItemCondition.ConditionList,
		ConditionType: request.ItemCondition.ConditionType,
		Owners:        owners,
		Database:      db,
	}

	cResp, err := itemService.FindCount(context.TODO(), &cReq, opss)
	if err != nil {
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionMappingDownload, err)
			return
		}
	}

	dReq := item.DownloadRequest{
		AppId:         appID,
		DatastoreId:   datastoreID,
		ConditionList: request.ItemCondition.ConditionList,
		ConditionType: request.ItemCondition.ConditionType,
		Owners:        owners,
		Database:      db,
	}

	stream, err := itemService.Download(context.TODO(), &dReq, opss)
	if err != nil {
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionMappingDownload, err)
			return
		}
	}

	// 获取当前台账的字段数据
	var fields []*typesx.DownloadField

	mappingInfo, err := mappingx.GetMappingInfo(db, datastoreID, mappingID)
	if err != nil {
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 获取数据失败，终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "build-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		httpx.GinHTTPError(c, ActionMappingDownload, err)
		return
	}

	allFields := fieldx.GetFields(db, datastoreID, appID, roles, false, false)

	for _, rule := range mappingInfo.MappingRule {
		if rule.FromKey == "" && rule.ToKey != "" {
			fields = append(fields, &typesx.DownloadField{
				FieldName:    rule.ToKey,
				FieldType:    "text",
				FieldId:      "#",
				Prefix:       rule.DefaultValue,
				DisplayOrder: rule.ShowOrder,
			})
		} else {
			for _, f := range allFields {
				if rule.FromKey == f.FieldId {
					fields = append(fields, &typesx.DownloadField{
						FieldId:       f.FieldId,
						FieldName:     rule.ToKey,
						FieldType:     f.FieldType,
						IsImage:       f.IsImage,
						AsTitle:       f.AsTitle,
						DisplayOrder:  rule.ShowOrder,
						DisplayDigits: f.DisplayDigits,
						Precision:     f.Precision,
						Prefix:        f.Prefix,
						Format:        rule.Format,
					})
				}
			}
		}
	}

	// 排序
	sort.Sort(typesx.DownloadFields(fields))
	// 获取当前app的语言数据
	langData := langx.GetLanguageData(db, lang, domain)

	timestamp := time.Now().Format("20060102150405")

	// 每次2000为一组数据
	total := cResp.GetTotal()

	// 发送消息 数据编辑完成，开始写入文件
	jobx.ModifyTask(task.ModifyRequest{
		JobId: jobID,
		// Message:     i18n.Tr(lang, "job.J_033"),
		Message:     "read and write data to temp file",
		CurrentStep: "write-to-file",
		Database:    db,
	}, userID)

	// 设定csv头部
	var header []string
	var headers [][]string
	for _, fl := range fields {
		header = append(header, fl.FieldName)
	}
	headers = append(headers, header)

	// Excel文件下载
	if fileType == "xlsx" {
		excelFile := excelize.NewFile()
		// 创建一个工作表
		index := excelFile.NewSheet("Sheet1")
		// 设置工作簿的默认工作表
		excelFile.SetActiveSheet(index)

		for i, rows := range headers {
			for j, v := range rows {
				y := excelx.GetAxisY(j+1) + strconv.Itoa(i+1)
				excelFile.SetCellValue("Sheet1", y, v)
			}
		}

		var current int = 0
		var items [][]string
		line := 0

		for {
			it, err := stream.Recv()
			if err == io.EOF {
				// 当前结束了，但是items还有数据
				if len(items) > 0 {

					// 返回消息
					result := make(map[string]interface{})

					result["total"] = total
					result["current"] = current

					message, _ := json.Marshal(result)

					// 发送消息 写入条数
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     string(message),
						CurrentStep: "write-to-file",
						Database:    db,
					}, userID)

					// 写入数据
					for k, rows := range items {
						for j, v := range rows {
							y := excelx.GetAxisY(j+1) + strconv.Itoa(int((line*500 + k + 3)))
							excelFile.SetCellValue("Sheet1", y, v)
						}
					}
				}
				break
			}

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

				httpx.GinHTTPError(c, ActionMappingDownload, err)
				return
			}
			current++
			dt := it.GetItem()

			{
				// 设置csv行
				var itemData []string
				for _, fl := range fields {
					// 使用默认的值的情况
					if fl.FieldId == "#" {
						itemData = append(itemData, fl.Prefix)
					} else {
						itemMap := dt.GetItems()
						if value, ok := itemMap[fl.FieldId]; ok {
							result := ""
							switch value.DataType {
							case "text", "textarea", "number", "time", "switch":
								result = value.GetValue()
							case "autonum":
								result = value.GetValue()
							case "lookup":
								result = value.GetValue()
							case "options":
								result = value.GetValue()
							case "date":
								if value.GetValue() == "0001-01-01" {
									result = ""
								} else {
									if len(fl.Format) > 0 {
										date, err := time.Parse("2006-01-02", value.GetValue())
										if err != nil {
											result = ""
										} else {

											result = date.Format(fl.Format)
										}
									} else {
										result = value.GetValue()
									}
								}
							case "user":
								var userStrList []string
								json.Unmarshal([]byte(value.GetValue()), &userStrList)
								result = strings.Join(userStrList, ",")
							case "file":
								var files []typesx.FileValue
								json.Unmarshal([]byte(value.GetValue()), &files)
								var fileStrList []string
								for _, f := range files {
									fileStrList = append(fileStrList, f.Name)
								}
								result = strings.Join(fileStrList, ",")
							default:
								break
							}

							itemData = append(itemData, result)
						} else {
							itemData = append(itemData, "")
						}
					}
				}
				// 添加行
				items = append(items, itemData)
			}

			if current%500 == 0 {
				// 返回消息
				result := make(map[string]interface{})

				result["total"] = total
				result["current"] = current

				message, _ := json.Marshal(result)

				// 发送消息 写入条数
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     string(message),
					CurrentStep: "write-to-file",
					Database:    db,
				}, userID)

				// 写入数据
				for k, rows := range items {
					for j, v := range rows {
						y := excelx.GetAxisY(j+1) + strconv.Itoa(int((line*500 + k + 3)))
						excelFile.SetCellValue("Sheet1", y, v)
					}
				}

				// 清空items
				items = items[:0]

				line++
			}
		}

		defer stream.Close()

		outFile := "text.xlsx"

		if err := excelFile.SaveAs(outFile); err != nil {
			fmt.Println(err)
		}

		fo, err := os.Open(outFile)
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

			httpx.GinHTTPError(c, ActionMappingDownload, err)
			return
		}
		defer fo.Close()

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

			httpx.GinHTTPError(c, ActionMappingDownload, err)
			return
		}
		filePath := path.Join(appRoot, "excel", "datastore_"+timestamp+".xlsx")
		path, err := minioClient.SavePublicObject(fo, filePath, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
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

			httpx.GinHTTPError(c, ActionMappingDownload, err)
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

				httpx.GinHTTPError(c, ActionMappingDownload, err)
				return
			}
		} else {
			// 如果已达上限，则删除刚才上传的文件
			minioClient.DeleteObject(path.Name)
			path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId: jobID,
				// Message:     i18n.Tr(lang, "job.J_007"),
				Message:     "Maximum storage space has been reached, file upload failed",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionMappingDownload, errors.New("最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"))
			return
		}

		os.Remove(outFile)

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_028"),
			Message:     "job execution success",
			CurrentStep: "end",
			File: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database: db,
		}, userID)

	} else {
		var writer *csv.Writer

		filex.Mkdir("temp/")

		// 写入文件到本地
		filename := "temp/tmp" + "_" + timestamp + "_header" + ".csv"
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

			httpx.GinHTTPError(c, ActionMappingDownload, err)
			return
		}

		if encoding == "sjis" {
			converter := transform.NewWriter(f, japanese.ShiftJIS.NewEncoder())
			writer = csv.NewWriter(converter)
		} else {
			writer = csv.NewWriter(f)
			// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
			headers[0][0] = "\xEF\xBB\xBF" + headers[0][0]
		}

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

				httpx.GinHTTPError(c, ActionMappingDownload, errors.New("現在のタイトルには、日本語の[shift-jis]エンコード以外の文字が含まれており、実行を続行できません。"))
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
		writer.Flush() // 此时才会将缓冲区数据写入

		var current int = 0
		var items [][]string

		for {
			it, err := stream.Recv()
			if err == io.EOF {
				// 当前结束了，但是items还有数据
				if len(items) > 0 {

					// 返回消息
					result := make(map[string]interface{})

					result["total"] = total
					result["current"] = current

					message, _ := json.Marshal(result)

					// 发送消息 写入条数
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     string(message),
						CurrentStep: "write-to-file",
						Database:    db,
					}, userID)

					// 写入数据
					err = writer.WriteAll(items)
					if err != nil {
						if err.Error() == "encoding: rune not supported by encoding." {
							path := filex.WriteAndSaveFile(domain, appID, []string{"現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"})
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

							httpx.GinHTTPError(c, ActionMappingDownload, errors.New("現在のタイトルには、日本語の[shift-jis]エンコード以外の文字が含まれており、実行を続行できません。"))
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

						httpx.GinHTTPError(c, ActionMappingDownload, err)
						return
					}

					// 缓冲区数据写入
					writer.Flush()
				}
				break
			}

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

				httpx.GinHTTPError(c, ActionMappingDownload, err)
				return
			}
			current++
			dt := it.GetItem()

			{
				// 设置csv行
				var itemData []string
				for _, fl := range fields {
					// 使用默认的值的情况
					if fl.FieldId == "#" {
						itemData = append(itemData, fl.Prefix)
					} else {
						itemMap := dt.GetItems()
						if value, ok := itemMap[fl.FieldId]; ok {
							result := ""
							switch value.DataType {
							case "text", "textarea", "number", "time", "switch":
								result = value.GetValue()
							case "autonum":
								result = value.GetValue()
							case "lookup":
								result = value.GetValue()
							case "options":
								result = langx.GetLangValue(langData, value.GetValue(), langx.DefaultResult)
							case "date":
								if value.GetValue() == "0001-01-01" {
									result = ""
								} else {
									if len(fl.Format) > 0 {
										date, err := time.Parse("2006-01-02", value.GetValue())
										if err != nil {
											result = ""
										} else {

											result = date.Format(fl.Format)
										}
									} else {
										result = value.GetValue()
									}
								}
							case "user":
								var userStrList []string
								json.Unmarshal([]byte(value.GetValue()), &userStrList)
								result = strings.Join(userStrList, ",")
							case "file":
								var files []typesx.FileValue
								json.Unmarshal([]byte(value.GetValue()), &files)
								var fileStrList []string
								for _, f := range files {
									fileStrList = append(fileStrList, f.Name)
								}
								result = strings.Join(fileStrList, ",")
							default:
								break
							}

							itemData = append(itemData, result)
						} else {
							itemData = append(itemData, "")
						}
					}
				}
				// 添加行
				items = append(items, itemData)

			}

			if current%500 == 0 {
				// 返回消息
				result := make(map[string]interface{})

				result["total"] = total
				result["current"] = current

				message, _ := json.Marshal(result)

				// 发送消息 写入条数
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     string(message),
					CurrentStep: "write-to-file",
					Database:    db,
				}, userID)

				// 写入数据
				// 写入数据
				err = writer.WriteAll(items)
				if err != nil {
					if err.Error() == "encoding: rune not supported by encoding." {
						path := filex.WriteAndSaveFile(domain, appID, []string{"現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"})
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

						httpx.GinHTTPError(c, ActionMappingDownload, errors.New("現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"))
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

					httpx.GinHTTPError(c, ActionMappingDownload, err)
					return
				}

				// 缓冲区数据写入
				writer.Flush()

				// 清空items
				items = items[:0]
			}
		}
		defer stream.Close()
		defer f.Close()

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_029"),
			Message:     "merge file from temp file",
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_043"),
			Message:     "save file to file server",
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

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

			httpx.GinHTTPError(c, ActionMappingDownload, err)
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

			httpx.GinHTTPError(c, ActionMappingDownload, err)
			return
		}
		filePath := path.Join(appRoot, "csv", "datastore_"+timestamp+".csv")
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

			httpx.GinHTTPError(c, ActionMappingDownload, err)
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

				httpx.GinHTTPError(c, ActionMappingDownload, err)
				return
			}
		} else {
			// 如果已达上限，则删除刚才上传的文件
			minioClient.DeleteObject(path.Name)
			path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
			// 发送消息 保存文件失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId: jobID,
				// Message:     i18n.Tr(lang, "job.J_007"),
				Message:     "Maximum storage space has been reached, file upload failed",
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			httpx.GinHTTPError(c, ActionMappingDownload, errors.New("最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"))
			return
		}

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId: jobID,
			// Message:     i18n.Tr(lang, "job.J_028"),
			Message:     "job execution success",
			CurrentStep: "end",
			File: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database: db,
		}, userID)
	}

	loggerx.InfoLog(c, ActionMappingDownload, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, MappingProcessName, ActionMappingDownload)),
		Data:    gin.H{},
	})
}

// mappingUpload 开始导入
func mappingUpload(c *gin.Context, filePath string) error {

	base := upload.MappingParams{
		MappingId:   c.PostForm("mapping_id"),
		JobId:       c.PostForm("job_id"),
		EmptyChange: cast.ToBool(c.PostForm("empty_change")),
		UserId:      sessionx.GetAuthUserID(c),
		AppId:       sessionx.GetCurrentApp(c),
		Lang:        sessionx.GetCurrentLanguage(c),
		Domain:      sessionx.GetUserDomain(c),
		DatastoreId: c.Param("d_id"),
		AccessKeys:  sessionx.GetUserAccessKeys(c, c.Param("d_id"), "W"),
		Owners:      sessionx.GetUserOwner(c),
		Roles:       sessionx.GetUserRoles(c),
		Database:    sessionx.GetUserCustomer(c),
	}

	uploadService := upload.NewUploadService("import", client.DefaultClient)

	var req upload.MappingRequest
	req.BaseParams = &base
	req.FilePath = filePath

	_, err := uploadService.MappingUpload(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("csvUpload", err.Error())
		return err
	}

	return nil
}
