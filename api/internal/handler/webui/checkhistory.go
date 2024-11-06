package webui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/internal/common/logic/itemx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/userx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/check"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
)

// CheckHistory CheckHistory
type CheckHistory struct{}

// log出力
const (
	CheckHistoryProcessName      = "CheckHistory"
	ActionFindCheckHistories     = "FindCheckHistories"
	ActionFindCheckHistory       = "FindCheckHistory"
	ActionDeleteCheckHistories   = "DeleteCheckHistories"
	ActionDownloadCheckHistories = "DownloadCheckHistories"
)

// FindHistories 获取所有履历数据
// @Router histories [get]
func (i *CheckHistory) FindHistories(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindHistories, loggerx.MsgProcessStarted)

	checkService := check.NewCheckHistoryService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	datastoreID := c.Param("d_id")
	db := sessionx.GetUserCustomer(c)
	appID := sessionx.GetCurrentApp(c)
	roles := sessionx.GetUserRoles(c)

	var req check.HistoriesRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionFindHistories, err)
		return
	}

	// 从path中获取参数
	req.DatastoreId = datastoreID
	req.Database = db

	fs, err := getScanFields(db, datastoreID)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindHistories, err)
		return
	}

	req.DisplayFields = fs

	response, err := checkService.FindHistories(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindHistories, err)
		return
	}

	// 获取当前台账的字段数据
	fields := fieldx.GetFields(db, datastoreID, appID, roles, true, false)

	var displayFields []*field.Field

	for _, f := range fields {
	LP:
		for _, sf := range fs {
			if sf == f.FieldId {
				displayFields = append(displayFields, f)
				break LP
			}
		}
	}

	loggerx.InfoLog(c, ActionFindHistories, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, CheckHistoryProcessName, ActionFindHistories)),
		Data: gin.H{
			"total":     response.GetTotal(),
			"fields":    displayFields,
			"histories": response.GetHistories(),
		},
	})
}

// HistoryDownload 获取履历所有数据,以csv文件的方式下载
// @Router /datastores/{d_id}/items [get]
func (i *CheckHistory) HistoryDownload(c *gin.Context) {
	loggerx.InfoLog(c, ActionDownloadHistories, loggerx.MsgProcessStarted)

	var req check.HistoriesRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionFindHistories, err)
		return
	}

	jobID := c.Query("job_id")
	userID := sessionx.GetAuthUserID(c)
	datastoreID := c.Param("d_id")
	appID := sessionx.GetCurrentApp(c)

	roles := sessionx.GetUserRoles(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	db := sessionx.GetUserCustomer(c)

	go func() {

		taskData := task.AddRequest{
			JobId:        jobID,
			JobName:      "Download check history",
			Origin:       "apps." + appID + ".datastores." + datastoreID,
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(sessionx.GetCurrentLanguage(c), "job.J_014"),
			TaskType:     "check-hs-csv-download",
			Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		}

		jobx.CreateTask(taskData)

		checkService := check.NewCheckHistoryService("database", client.DefaultClient)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		cReq := check.CountRequest{
			DatastoreId:    datastoreID,
			ItemId:         req.GetItemId(),
			CheckType:      req.GetCheckType(),
			CheckStartDate: req.GetCheckStartDate(),
			CheckedAtFrom:  req.GetCheckedAtFrom(),
			CheckedAtTo:    req.GetCheckedAtTo(),
			CheckedBy:      req.GetCheckedBy(),
			Database:       db,
		}

		cResp, err := checkService.FindHistoryCount(context.TODO(), &cReq, opss)
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

			return
		}

		dReq := check.DownloadRequest{
			DatastoreId:    datastoreID,
			ItemId:         req.GetItemId(),
			CheckType:      req.GetCheckType(),
			CheckStartDate: req.GetCheckStartDate(),
			CheckedAtFrom:  req.GetCheckedAtFrom(),
			CheckedAtTo:    req.GetCheckedAtTo(),
			CheckedBy:      req.GetCheckedBy(),
			Database:       db,
		}

		stream, err := checkService.Download(context.TODO(), &dReq, opss)
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

			return
		}

		scanFields, err := getScanFields(db, datastoreID)
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

			return
		}

		req.DisplayFields = scanFields

		// 获取当前台账的字段数据
		fields := fieldx.GetFields(db, datastoreID, appID, roles, true, false)
		// 获取当前app的语言数据
		langData := langx.GetLanguageData(db, lang, domain)

		allUsers := userx.GetAllUser(db, appID, domain)
		lookupItems := map[string][]*item.Item{}

		var displayFields []*field.Field

		for _, fl := range fields {
		SL:
			for _, sf := range scanFields {
				if sf == fl.FieldId {
					if fl.GetFieldType() == "lookup" {
						findAccessKeys := sessionx.GetAccessKeys(db, userID, fl.GetLookupDatastoreId(), "R")
						itemList := itemx.GetLookupItems(db, fl.GetLookupDatastoreId(), fl.GetAppId(), findAccessKeys)
						lookupItems[fl.FieldId] = itemList
					}
					displayFields = append(displayFields, fl)
					break SL
				}
			}

		}

		timestamp := time.Now().Format("20060102150405")

		// 每次2000为一组数据
		total := cResp.GetTotal()

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_033"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		var header []string

		// 循环字段
		for _, field := range displayFields {
			header = append(header, langx.GetLangValue(langData, langx.GetFieldKey(field.AppId, field.DatastoreId, field.FieldId), langx.DefaultResult))
		}

		header = append(header, i18n.Tr(lang, "fixed.F_021"))
		header = append(header, i18n.Tr(lang, "fixed.F_022"))
		header = append(header, i18n.Tr(lang, "fixed.F_023"))
		header = append(header, i18n.Tr(lang, "fixed.F_024"))

		headers := append([][]string{}, header)

		// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
		headers[0][0] = "\xEF\xBB\xBF" + headers[0][0]

		filex.Mkdir("temp/")

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
		writer.WriteAll(headers)

		writer.Flush() // 此时才会将缓冲区数据写入

		var current int = 0
		var items [][]string

		for {
			hs, err := stream.Recv()
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
					writer.WriteAll(items)
					writer.Flush() // 此时才会将缓冲区数据写入
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

				return
			}
			current++
			dt := hs.GetHistory()

			{
				// 获取变更数据
				changedItems := dt.GetItems()
				// 变更数据
				var row []string

				// 循环字段
				for _, field := range displayFields {

					// 获取变更字段数据
					if it, ok := changedItems[field.FieldId]; ok {
						switch it.GetDataType() {
						case "text", "textarea", "time", "switch":
							row = append(row, it.GetValue())
						case "number":
							row = append(row, it.GetValue())
						case "date":
							if strings.HasPrefix(it.GetValue(), "0001-01-01") {
								row = append(row, "")
							} else {
								row = append(row, it.GetValue())
							}

						case "file":
							var fileList []typesx.FileValue
							json.Unmarshal([]byte(it.GetValue()), &fileList)
							var files []string
							for _, f := range fileList {
								files = append(files, f.Name)
							}

							row = append(row, strings.Join(files, ","))

						case "user":
							var userStrList []string
							json.Unmarshal([]byte(it.GetValue()), &userStrList)
							var users []string
							for _, u := range userStrList {
								users = append(users, userx.TranUser(u, allUsers))
							}

							row = append(row, strings.Join(users, ","))
						case "options":
							row = append(row, langx.GetLangValue(langData, langx.GetOptionKey(field.AppId, field.OptionId, it.GetValue()), langx.DefaultResult))
						case "lookup":
							row = append(row, it.GetValue())
						}
					}
				}

				// 盘点类型
				row = append(row, dt.GetCheckType())

				// 盘点开始日期
				row = append(row, dt.GetCheckStartDate())

				// 盘点日期
				row = append(row, dt.GetCheckedAt())

				// 盘点者
				row = append(row, userx.TranUser(dt.GetCheckedBy(), allUsers))

				items = append(items, row)
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
				writer.WriteAll(items)
				writer.Flush() // 此时才会将缓冲区数据写入

				// 清空items
				items = items[:0]
			}

		}

		defer stream.Close()
		defer f.Close()

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_029"),
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_043"),
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
		filePath := path.Join(appRoot, "csv", "check_"+timestamp+".csv")
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
				Message:     i18n.Tr(lang, "job.J_007"),
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
	loggerx.InfoLog(c, ActionDownloadHistories, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, CheckHistoryProcessName, ActionDownloadHistories)),
		Data:    gin.H{},
	})
}

// getScanFields 获取扫描字段
func getScanFields(db, datastoreID string) ([]string, error) {
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoreRequest
	// 从path获取
	req.DatastoreId = datastoreID
	req.Database = db

	response, err := datastoreService.FindDatastore(context.TODO(), &req)
	if err != nil {

		return nil, err
	}

	scanFields := response.GetDatastore().GetScanFields()
	field1 := response.GetDatastore().GetPrintField1()
	field2 := response.GetDatastore().GetPrintField2()
	field3 := response.GetDatastore().GetPrintField3()

	fs := containerx.New()
	fs.AddAll(scanFields...)
	if len(field1) > 0 {
		fs.Add(field1)
	}
	if len(field2) > 0 {
		fs.Add(field2)
	}
	if len(field3) > 0 {
		fs.Add(field3)
	}
	return fs.ToList(), nil
}
