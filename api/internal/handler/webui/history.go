package webui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/spf13/cast"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/userx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datahistory"
)

// History History
type History struct{}

// FindParam FindParam
type FindParam struct {
	ItemId        string `form:"item_id"`
	HistoryType   string `form:"history_type"`
	FieldId       string `form:"field_id"`
	CreatedAtFrom string `form:"created_at_from"`
	CreatedAtTo   string `form:"created_at_to"`
	OldValue      string `form:"old_value"`
	NewValue      string `form:"new_value"`
}

// log出力
const (
	HistoryProcessName        = "History"
	ActionFindHistories       = "FindHistories"
	ActionFindHistory         = "FindHistory"
	ActionHardDeleteHistories = "HardDeleteHistories"
	ActionDownloadHistories   = "DownloadHistories"
)

// FindHistories 获取所有履历数据
// @Router histories [get]
func (i *History) FindHistories(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindHistories, loggerx.MsgProcessStarted)

	var p FindParam
	if err := c.BindQuery(&p); err != nil {
		httpx.GinHTTPError(c, ActionFindHistories, err)
		return
	}

	appID := sessionx.GetCurrentApp(c)
	did := c.Param("d_id")
	roles := sessionx.GetUserRoles(c)
	db := sessionx.GetUserCustomer(c)

	// 获取当前台账的字段数据
	fields := fieldx.GetFields(db, did, appID, roles, true, false)

	// 获取对应的字段ID
	var fieldList []string
	for _, f := range fields {
		fieldList = append(fieldList, f.FieldId)
	}

	historyService := datahistory.NewHistoryService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
	}

	var req datahistory.HistoriesRequest
	// 从path中获取参数
	req.DatastoreId = did
	req.ItemId = p.ItemId
	req.HistoryType = p.HistoryType
	req.FieldId = p.FieldId
	req.CreatedAtFrom = p.CreatedAtFrom
	req.CreatedAtTo = p.CreatedAtTo
	req.OldValue = p.OldValue
	req.NewValue = p.NewValue
	req.PageIndex = cast.ToInt64(c.Query("page_index"))
	req.PageSize = cast.ToInt64(c.Query("page_size"))
	req.FieldList = fieldList
	req.Database = sessionx.GetUserCustomer(c)

	response, err := historyService.FindHistories(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindHistories, err)
		return
	}

	loggerx.InfoLog(c, ActionFindHistories, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, HistoryProcessName, ActionFindHistories)),
		Data: gin.H{
			"total":     response.GetTotal(),
			"histories": response.GetHistories(),
		},
	})
}

// FindLastHistories 获取最新的10条履历数据
// @Router histories [get]
func (i *History) FindLastHistories(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindHistories, loggerx.MsgProcessStarted)

	appID := sessionx.GetCurrentApp(c)
	did := c.Param("d_id")
	roles := sessionx.GetUserRoles(c)
	db := sessionx.GetUserCustomer(c)

	// 获取当前台账的字段数据
	fields := fieldx.GetFields(db, did, appID, roles, true, false)

	// 获取对应的字段ID
	var fieldList []string
	for _, f := range fields {
		fieldList = append(fieldList, f.FieldId)
	}

	historyService := datahistory.NewHistoryService("database", client.DefaultClient)

	var req datahistory.LastRequest
	// 从path中获取参数
	req.DatastoreId = did
	req.ItemId = c.Query("item_id")
	req.FieldList = fieldList
	req.Database = sessionx.GetUserCustomer(c)

	response, err := historyService.FindLastHistories(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindHistories, err)
		return
	}

	loggerx.InfoLog(c, ActionFindHistories, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, HistoryProcessName, ActionFindHistories)),
		Data:    response,
	})
}

// FindHistory 获取一条数据
// @Router histories [get]
func (i *History) FindHistory(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindHistory, loggerx.MsgProcessStarted)

	appID := sessionx.GetCurrentApp(c)
	did := c.Param("d_id")
	hid := c.Param("h_id")
	roles := sessionx.GetUserRoles(c)
	db := sessionx.GetUserCustomer(c)

	// 获取当前台账的字段数据
	fields := fieldx.GetFields(db, did, appID, roles, true, false)

	// 获取对应的字段ID
	var fieldList []string
	for _, f := range fields {
		fieldList = append(fieldList, f.FieldId)
	}

	historyService := datahistory.NewHistoryService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
	}

	var req datahistory.HistoryRequest
	// 从path中获取参数
	req.HistoryId = hid
	req.FieldList = fieldList
	req.Database = sessionx.GetUserCustomer(c)

	response, err := historyService.FindHistory(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindHistory, err)
		return
	}

	loggerx.InfoLog(c, ActionFindHistory, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, HistoryProcessName, ActionFindHistory)),
		Data:    response.GetHistory(),
	})
}

// HistoryDownload 获取履历所有数据,以csv文件的方式下载
// @Router /datastores/{d_id}/items [get]
func (i *History) HistoryDownload(c *gin.Context) {
	loggerx.InfoLog(c, ActionDownloadHistories, loggerx.MsgProcessStarted)
	var p FindParam
	if err := c.BindQuery(&p); err != nil {
		httpx.GinHTTPError(c, ActionDownloadHistories, err)
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

		// 创建任务
		jobx.CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      "history file download",
			Origin:       "apps." + appID + ".datastores." + datastoreID,
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(lang, "job.J_014"),
			TaskType:     "hs-csv-download",
			Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})

		historyService := datahistory.NewHistoryService("database", client.DefaultClient)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
		}

		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		// 获取当前台账的字段数据
		fields := fieldx.GetFields(db, datastoreID, appID, roles, true, false)

		// 获取对应的字段ID
		var fieldList []string
		for _, f := range fields {
			fieldList = append(fieldList, f.FieldId)
		}

		cReq := datahistory.CountRequest{
			DatastoreId:   datastoreID,
			ItemId:        p.ItemId,
			HistoryType:   p.HistoryType,
			FieldId:       p.FieldId,
			CreatedAtFrom: p.CreatedAtFrom,
			CreatedAtTo:   p.CreatedAtTo,
			OldValue:      p.OldValue,
			NewValue:      p.NewValue,
			Database:      db,
			FieldList:     fieldList,
		}

		count, err := historyService.FindHistoryCount(context.TODO(), &cReq, opss)
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

				return
			}
		}

		dReq := datahistory.DownloadRequest{
			DatastoreId:   datastoreID,
			ItemId:        p.ItemId,
			HistoryType:   p.HistoryType,
			FieldId:       p.FieldId,
			CreatedAtFrom: p.CreatedAtFrom,
			CreatedAtTo:   p.CreatedAtTo,
			OldValue:      p.OldValue,
			NewValue:      p.NewValue,
			Database:      db,
			FieldList:     fieldList,
		}

		stream, err := historyService.Download(context.TODO(), &dReq, opss)
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

				return
			}
		}

		// 获取当前app的语言数据
		langData := langx.GetLanguageData(db, lang, domain)

		// 排序
		sort.Sort(typesx.FieldList(fields))

		allUsers := userx.GetAllUser(db, appID, domain)

		timestamp := time.Now().Format("20060102150405")

		// 每次2000为一组数据
		total := float64(count.GetTotal())

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_033"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		var header []string

		// 循环字段
		for _, field := range fields {
			header = append(header, langx.GetLangValue(langData, langx.GetFieldKey(field.AppId, field.DatastoreId, field.FieldId), langx.DefaultResult))
		}

		header = append(header, i18n.Tr(lang, "fixed.F_026"))
		header = append(header, i18n.Tr(lang, "fixed.F_006"))
		header = append(header, i18n.Tr(lang, "fixed.F_007"))
		header = append(header, i18n.Tr(lang, "fixed.F_008"))
		header = append(header, i18n.Tr(lang, "fixed.F_027"))
		header = append(header, i18n.Tr(lang, "fixed.F_012"))
		header = append(header, i18n.Tr(lang, "fixed.F_011"))

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
				// 获取固定数据
				fixedItems := dt.GetFixedItems()
				// 固定数据
				var fixedRow []string

				exist := false

				// 循环字段
				for _, field := range fields {

					if dt.FieldId == field.FieldId {
						exist = true
					}

					// 获取固定字段数据
					if it, ok := fixedItems[field.FieldId]; ok {
						value := ""
						switch it.GetDataType() {
						case "text", "textarea", "number", "time", "switch":
							value = it.GetValue()
						case "date":
							if strings.HasPrefix(it.GetValue(), "0001-01-01") {
								value = ""
							} else {
								value = it.GetValue()
							}
						case "user":
							value = it.GetValue()
						case "options":
							value = it.GetValue()
						case "lookup":
							value = it.GetValue()
						}
						fixedRow = append(fixedRow, value)
					} else {
						fixedRow = append(fixedRow, "")
					}
				}

				var row []string
				row = append(row, fixedRow...)
				// 履历ID
				row = append(row, dt.HistoryId)
				// 如果存在该字段，字段名称
				if exist {
					row = append(row, langx.GetLangValue(langData, dt.FieldName, langx.DefaultResult))
				} else {
					row = append(row, dt.LocalName+"(Deprecated)")
				}
				// 变更前后
				row = append(row, dt.OldValue)
				row = append(row, dt.NewValue)
				// 变更类型

				if dt.GetHistoryType() == "insert" {
					row = append(row, i18n.Tr(lang, "fixed.F_014"))
				} else if dt.GetHistoryType() == "delete" {
					row = append(row, i18n.Tr(lang, "fixed.F_016"))
				} else {
					row = append(row, i18n.Tr(lang, "fixed.F_015"))
				}
				// 变更日期和变更者
				row = append(row, dt.GetCreatedAt())
				row = append(row, userx.TranUser(dt.GetCreatedBy(), allUsers))

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
		filePath := path.Join(appRoot, "csv", "history_"+timestamp+".csv")
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
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, HistoryProcessName, ActionDownloadHistories)),
		Data:    gin.H{},
	})

}
