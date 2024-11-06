package webui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/userx"
	"rxcsoft.cn/pit3/api/internal/common/slicex"
	"rxcsoft.cn/pit3/api/internal/common/transferx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/pit3/srv/workflow/proto/example"
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/proto/process"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// Approve Approve
type Approve struct{}

// log出力
const (
	ApproveProcessName       = "Approve"
	ActionFindApproveItems   = "FindApproveItems"
	ActionApproveLogDownload = "ApproveLogDownload"
	ActionFindApproveItem    = "FindApproveItem"
	ActionDeleteApproveItem  = "DeleteApproveItem"
)

// FindApproveItems 获取台账中的所有临时数据
// @Router /approves [POST]
func (i *Approve) FindApproveItems(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindApproveItems, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	userID := sessionx.GetAuthUserID(c)
	wfID := c.Query("wf_id")

	var req approve.ItemsRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionFindItems, err)
		return
	}

	req.WfId = wfID
	req.UserId = userID
	req.Database = db

	tplService := approve.NewApproveService("database", client.DefaultClient)

	response, err := tplService.FindItems(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveItems, err)
		return
	}

	loggerx.InfoLog(c, ActionFindApproveItems, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ApproveProcessName, ActionFindApproveItems)),
		Data: gin.H{
			"items_list": response.GetItems(),
			"total":      response.GetTotal(),
		},
	})

}

// ApproveLogDownload 审批日志下载
// @Router /approves/log/download [POST]
func (i *Approve) ApproveLogDownload(c *gin.Context) {
	loggerx.InfoLog(c, ActionApproveLogDownload, loggerx.MsgProcessStarted)

	// 参数收集
	db := sessionx.GetUserCustomer(c)
	userID := sessionx.GetAuthUserID(c)
	appID := sessionx.GetCurrentApp(c)
	roles := sessionx.GetUserRoles(c)
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	jobID := "job_" + time.Now().Format("20060102150405")
	wfID := c.Query("wf_id")

	// body参数
	var req approve.ItemsRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionApproveLogDownload, err)
		return
	}

	// 参数补齐
	req.WfId = wfID
	req.UserId = userID
	req.Database = db

	// 审批日志下载
	go func() {

		// 创建审批日志下载任务
		jobx.CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      "approve log download",
			Origin:       "apps." + appID + ".datastores." + req.DatastoreId,
			UserId:       userID,
			ShowProgress: false,
			Message:      i18n.Tr(lang, "job.J_014"),
			TaskType:     "al-csv-download",
			Steps:        []string{"start", "get-data", "build-data", "write-to-file", "save-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		// 发送消息 开始获取数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "承認データを取得します",
			CurrentStep: "get-data",
			Database:    db,
		}, userID)

		// 获取流程信息
		workflowService := workflow.NewWfService("workflow", client.DefaultClient)
		var reqW workflow.WorkflowRequest
		reqW.WfId = wfID
		reqW.Database = db
		resWorkflow, err := workflowService.FindWorkflow(context.TODO(), &reqW)
		if err != nil {
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
		}
		workAction := resWorkflow.Workflow.Params["action"]
		workFields := strings.Split(resWorkflow.Workflow.Params["fields"], ",")

		tplService := approve.NewApproveService("database", client.DefaultClient)
		// 获取审批数据
		response, err := tplService.FindItems(context.TODO(), &req, opss)
		if err != nil {
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
		}

		// 发送消息 数据获取完成，开始编辑
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ビルド承認データ",
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		// 获取当前台账的字段数据
		fields := fieldx.GetFields(db, req.DatastoreId, appID, roles, true, false)
		// 排序
		sort.Sort(typesx.FieldList(fields))
		// 获取当前app的语言数据
		langData := langx.GetLanguageData(db, lang, domain)
		// 获取用户情报
		allUsers := userx.GetAllUser(db, appID, domain)

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "一時ファイルへのデータの読み取りと書き込み",
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		// 编辑数据
		timestamp := time.Now().Format("20060102150405")

		// 2000一组编辑文件数据
		total := float64(response.GetTotal())
		count := math.Ceil(total / 2000)

		// 文件头编辑
		var header []string
		// 循环台账动态字段(元数据部情报头)
		for _, field := range fields {
			header = append(header, langx.GetLangValue(langData, langx.GetFieldKey(field.AppId, field.DatastoreId, field.FieldId), langx.DefaultResult))
		}
		// 变化数据部头编辑
		if workAction == "update" {
			for i := 0; i < len(workFields); i++ {
				stri := strconv.FormatInt(int64(i+1), 10)
				header = append(header, i18n.Tr(lang, "fixed.F_006")+"("+stri+")")
				header = append(header, i18n.Tr(lang, "fixed.F_007")+"("+stri+")")
				header = append(header, i18n.Tr(lang, "fixed.F_008")+"("+stri+")")
			}
		}

		// 固定头部编辑
		header = append(header, i18n.Tr(lang, "fixed.F_009"))
		header = append(header, i18n.Tr(lang, "fixed.F_010"))
		header = append(header, i18n.Tr(lang, "fixed.F_011"))
		header = append(header, i18n.Tr(lang, "fixed.F_012"))
		header = append(header, i18n.Tr(lang, "fixed.F_013"))

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

		// 2000一组取写数据
		for i := int64(0); i < int64(count); i++ {

			req.PageIndex = i + 1
			req.PageSize = 2000

			res, err := tplService.FindItems(context.TODO(), &req, opss)
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

			// 无数据返回空头文件
			if len(res.GetItems()) == 0 {
				return
			}

			// 数据编辑
			var items [][]string
			// 循环审批数据表该流程数据
			for _, dt := range res.GetItems() {
				// 固定数据
				var fixedRow []string
				// 对象元数据取得
				var originItems map[string]*approve.Value
				// 变更数据
				var changedRow []string
				// 获取变更对象变更后数据
				changedAItems := dt.GetItems()
				// 获取变更对象变更前数据
				changedBItems := dt.GetHistory()

				if workAction == "update" {
					originItems = changedBItems
				} else {
					originItems = changedAItems
				}

				// 循环字段记录对象元数据和审批对象字段记录变更情报
				for _, field := range fields {
					// 记录对象元数据
					if it, ok := originItems[field.FieldId]; ok {
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
							var userStrList []string
							json.Unmarshal([]byte(it.GetValue()), &userStrList)
							var users []string
							for _, u := range userStrList {
								users = append(users, userx.TranUser(u, allUsers))
							}

							value = strings.Join(users, ",")
						case "options":
							value = langx.GetLangValue(langData, langx.GetOptionKey(field.AppId, field.OptionId, it.GetValue()), langx.DefaultResult)
						case "lookup":
							value = it.GetValue()
						}
						fixedRow = append(fixedRow, value)
					} else {
						fixedRow = append(fixedRow, "")
					}

					// 循环审批对象字段记录变更情报
					if workAction == "update" {
						if slicex.IsExist(workFields, field.FieldId) {
							// 变更字段名称

							changedRow = append(changedRow, langx.GetLangValue(langData, langx.GetFieldKey(field.AppId, field.DatastoreId, field.FieldId), langx.DefaultResult))
							// 变更字段名称变更前值
							if it, ok := changedBItems[field.FieldId]; ok {
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
									var userStrList []string
									json.Unmarshal([]byte(it.GetValue()), &userStrList)
									var users []string
									for _, u := range userStrList {
										users = append(users, userx.TranUser(u, allUsers))
									}

									value = strings.Join(users, ",")
								case "options":
									value = langx.GetLangValue(langData, langx.GetOptionKey(field.AppId, field.OptionId, it.GetValue()), langx.DefaultResult)
								case "lookup":
									value = it.GetValue()
								}
								changedRow = append(changedRow, value)
							} else {
								changedRow = append(changedRow, "")
							}
							// 变更字段名称变更后值
							if it, ok := changedAItems[field.FieldId]; ok {
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
									var userStrList []string
									json.Unmarshal([]byte(it.GetValue()), &userStrList)
									var users []string
									for _, u := range userStrList {
										users = append(users, userx.TranUser(u, allUsers))
									}

									value = strings.Join(users, ",")
								case "options":
									value = langx.GetLangValue(langData, langx.GetOptionKey(field.AppId, field.OptionId, it.GetValue()), langx.DefaultResult)
								case "lookup":
									value = it.GetValue()
								}
								changedRow = append(changedRow, value)
							} else {
								changedRow = append(changedRow, "")
							}
						}
					}
				}

				// 获取审批历程情报
				proceeService := process.NewProcessService("workflow", client.DefaultClient)
				var pReq process.ProcessesRequest
				pReq.ExId = dt.GetExampleId()
				pReq.Database = db
				pResp, err := proceeService.FindProcesses(context.TODO(), &pReq)
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

				// 申请者情报
				var row []string
				row = append(row, fixedRow...)
				row = append(row, changedRow...)

				// 流程操作
				wValue := ""
				if workAction == "new" {
					wValue = i18n.Tr(lang, "fixed.F_014")
				} else if workAction == "update" {
					wValue = i18n.Tr(lang, "fixed.F_015")
				} else {
					wValue = i18n.Tr(lang, "fixed.F_015")
				}
				row = append(row, wValue)
				// 审批
				aValue := i18n.Tr(lang, "fixed.F_018")
				row = append(row, aValue)

				// 申请者
				row = append(row, userx.TranUser(dt.GetApplicant(), allUsers))
				// 申请时间
				row = append(row, dt.GetCreatedAt())
				// 备考

				// 备考
				aValue = i18n.Tr(lang, "fixed.F_020")
				row = append(row, aValue)

				items = append(items, row)

				// 循环审批历程情报，变成行
				for _, ps := range pResp.GetProcesses() {
					if ps.GetStatus() == 0 {
						continue
					}
					if ps.GetComment() != "Approved by other approvers" {
						var row []string
						row = append(row, fixedRow...)
						row = append(row, changedRow...)

						// 流程操作
						wValue := ""
						if workAction == "new" {
							wValue = i18n.Tr(lang, "fixed.F_014")
						} else if workAction == "update" {
							wValue = i18n.Tr(lang, "fixed.F_015")
						} else {
							wValue = i18n.Tr(lang, "fixed.F_015")
						}
						row = append(row, wValue)
						// 审批
						aValue := ""
						if ps.GetStatus() == 1 {
							aValue = i18n.Tr(lang, "fixed.F_017")
						} else {
							aValue = i18n.Tr(lang, "fixed.F_019")
						}
						row = append(row, aValue)
						// 审批者
						row = append(row, userx.TranUser(ps.GetUserId(), allUsers))
						// 审批时间
						row = append(row, ps.GetCreatedAt())
						// 备考
						row = append(row, ps.GetComment())

						items = append(items, row)
					}
				}
			}

			writer.WriteAll(items)

			writer.Flush() // 此时才会将缓冲区数据写入

			current += len(response.GetItems())

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
		}

		f.Close()

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "一時ファイルからファイルをマージ",
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルをファイルサーバーに保存",
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
	loggerx.InfoLog(c, ActionApproveLogDownload, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, HistoryProcessName, ActionApproveLogDownload)),
		Data:    gin.H{},
	})
}

// FindApproveItem 通过database_Id和item_id获取数据
// @Router /approves/{item_id} [get]
func (i *Approve) FindApproveItem(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindApproveItem, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	userID := sessionx.GetAuthUserID(c)
	appID := sessionx.GetCurrentApp(c)
	roles := sessionx.GetUserRoles(c)

	datastoreId := c.Query("datastore_id")

	tplService := approve.NewApproveService("database", client.DefaultClient)

	var req approve.ItemRequest
	req.ExampleId = c.Param("ex_id")
	req.DatastoreId = datastoreId
	req.Database = db

	response, err := tplService.FindItem(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveItem, err)
		return
	}

	status, err := getStatus(db, response.GetItem().GetExampleId(), userID)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveItem, err)
		return
	}

	if !status.CanShow {
		httpx.GinHTTPError(c, ActionFindApproveItem, errors.New("you do not have permission"))
		return
	}

	// 获取流程实例
	exService := example.NewExampleService("workflow", client.DefaultClient)

	var eReq example.ExampleRequest
	eReq.ExId = response.GetItem().GetExampleId()
	eReq.Database = db

	exResp, err := exService.FindExample(context.TODO(), &eReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveItem, err)
		return
	}

	nodeService := node.NewNodeService("workflow", client.DefaultClient)

	var nReq node.NodesRequest
	nReq.WfId = exResp.GetExample().GetWfId()
	nReq.Database = db

	nResp, err := nodeService.FindNodes(context.TODO(), &nReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveItem, err)
		return
	}

	// 查找该实例对应的进程
	proceeService := process.NewProcessService("workflow", client.DefaultClient)
	var pReq process.ProcessesRequest
	pReq.ExId = response.GetItem().GetExampleId()
	pReq.Database = db

	pResp, err := proceeService.FindProcesses(context.TODO(), &pReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveItem, err)
		return
	}

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var wreq workflow.WorkflowRequest
	wreq.WfId = exResp.GetExample().GetWfId()
	wreq.Database = db

	fResp, err := workflowService.FindWorkflow(context.TODO(), &wreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveItem, err)
		return
	}

	params := fResp.GetWorkflow().GetParams()
	fields := fieldx.GetFields(db, response.GetItem().GetDatastoreId(), appID, roles, true, true)
	showFields := []*field.Field{}
	if params["action"] == "update" {
		fs := params["fields"]
		if len(fs) > 0 {
			fieldList := strings.Split(fs, ",")
			fieldMap := map[string]string{}
			for _, f := range fieldList {
				fieldMap[f] = f
			}

			for _, f := range fields {
				if _, exist := fieldMap[f.FieldId]; exist {
					showFields = append(showFields, f)
				}
			}
		} else {
			showFields = fieldx.GetFields(db, response.GetItem().GetDatastoreId(), appID, roles, true, true)
		}
	} else {
		showFields = fieldx.GetFields(db, response.GetItem().GetDatastoreId(), appID, roles, true, true)
	}

	res := make(map[string]interface{})
	res["item_id"] = response.GetItem().GetItemId()
	res["app_id"] = response.GetItem().GetAppId()
	res["example_id"] = response.GetItem().GetExampleId()
	res["datastore_id"] = response.GetItem().GetDatastoreId()
	res["created_at"] = response.GetItem().GetCreatedAt()
	res["created_by"] = response.GetItem().GetCreatedBy()
	res["status"] = status
	res["change_fields"] = showFields
	res["fields"] = fields
	res["process"] = pResp.GetProcesses()
	res["nodes"] = nResp.GetNodes()

	itemMap := make(map[string]interface{})
	for key, value := range response.GetItem().GetCurrent() {
		itemMap[key] = transferx.TransferApprove(value)
	}
	res["items"] = itemMap
	history := make(map[string]interface{})
	for key, value := range response.GetItem().GetHistory() {
		history[key] = transferx.TransferApprove(value)
	}
	res["history"] = history

	loggerx.InfoLog(c, ActionFindApproveItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ApproveProcessName, ActionFindApproveItem)),
		Data:    res,
	})
}

// DeleteApprove 删除台账数据
// @Summary 删除台账数据
// @description 调用srv中的item服务，删除台账数据
// @Tags Approve
// @Accept json
// @Security JWT
// @Produce  json
// @Param item_id path string true "数据ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /approves/{item_id} [delete]
func (i *Approve) DeleteApprove(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteApproveItem, loggerx.MsgProcessStarted)

	tplService := approve.NewApproveService("database", client.DefaultClient)

	var req approve.DeleteRequest
	// 从path中获取参数
	req.Items = c.QueryArray("items")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := tplService.DeleteItems(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteApproveItem, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteApproveItem, fmt.Sprintf("DeleteApprove[%v] Success", req.GetItems()))

	loggerx.InfoLog(c, ActionDeleteApproveItem, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, ApproveProcessName, ActionDeleteApproveItem)),
		Data:    response,
	})
}

// getStatus 获取流程审批的状态
func getStatus(db, exID, userID string) (s *typesx.Status, e error) {

	// 获取流程实例
	exService := example.NewExampleService("workflow", client.DefaultClient)

	var req example.ExampleRequest
	req.ExId = exID
	req.Database = db

	exResp, err := exService.FindExample(context.TODO(), &req)
	if err != nil {
		return nil, err
	}

	status := exResp.GetExample().GetStatus()

	result := new(typesx.Status)

	// 设置审批状态
	result.ApproveStatus = status
	// 设置申请者
	result.Applicant = exResp.GetExample().GetUserId()
	// 申请者是当前用户
	if userID == exResp.GetExample().GetUserId() {
		result.CanShow = true
	}

	// 查找该实例对应的进程
	proceeService := process.NewProcessService("workflow", client.DefaultClient)
	var pReq process.ProcessesRequest
	pReq.ExId = exID
	pReq.Database = db

	pResp, err := proceeService.FindProcesses(context.TODO(), &pReq)
	if err != nil {
		return nil, err
	}
	// 审批中的状态
	if status == 1 {
		// 获取当前用户的进程
		for i := 0; i < len(pResp.GetProcesses()); i++ {
			p := pResp.Processes[i]
			// 审批人员是当前用户
			if userID == p.UserId {
				result.CanShow = true
			}
			if p.Status == 0 {
				result.CurrentNode = p.CurrentNode
			}
		}

		return result, nil
	}
	// 承认的状态
	if status == 2 {
		processList := pResp.GetProcesses()
		length := len(processList)
		getApprover := false
		// 获取当前用户的进程
		for i := 0; i < length; i++ {
			p := processList[i]
			if !getApprover {
				if p.GetComment() != "Approved by other approvers" {
					result.CurrentNode = p.CurrentNode
					result.Approver = p.UserId
					getApprover = true
				}
			}

			// 审批人员是当前用户
			if userID == p.UserId {
				result.CanShow = true
			}
		}

		return result, nil
	}

	// 却下的状态
	processList := pResp.GetProcesses()
	length := len(processList)
	// 获取当前用户的进程
	for i := 0; i < length; i++ {
		p := processList[i]
		// 审批人员是当前用户
		if userID == p.UserId {
			result.CanShow = true
		}

		if p.Status == 2 {
			result.CurrentNode = p.CurrentNode
			result.Approver = p.UserId
		}
	}

	return result, nil
}
