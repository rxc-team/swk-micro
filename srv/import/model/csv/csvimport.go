package csv

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/client"
	"github.com/spf13/cast"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/import/common/containerx"
	"rxcsoft.cn/pit3/srv/import/common/filex"
	"rxcsoft.cn/pit3/srv/import/common/floatx"
	"rxcsoft.cn/pit3/srv/import/common/langx"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/common/storex"
	"rxcsoft.cn/pit3/srv/import/common/timestamp"
	"rxcsoft.cn/pit3/srv/import/model"
	"rxcsoft.cn/pit3/srv/import/system/sessionx"
	"rxcsoft.cn/pit3/srv/import/system/wfx"
	"rxcsoft.cn/pit3/srv/import/system/wsx"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
	storagecli "rxcsoft.cn/utils/storage/client"
	"rxcsoft.cn/utils/timex"
)

const DefaultEmptyStr = "#N/A"

// Import 文件导入并上传
func Import(base Params, file FileParams) {
	// 获取传入变量
	jobID := base.JobId
	action := base.Action
	encoding := base.Encoding
	zipCharset := base.ZipCharset
	userID := base.UserId
	owners := base.Owners
	roles := base.Roles
	lang := base.Lang
	domain := base.Domain
	datastoreID := base.DatastoreId
	db := base.Database
	groupID := base.GroupId
	appID := base.AppId
	updateOwners := base.UpdateOwners
	store := storex.NewRedisStore(600)
	uploadID := appID + "upload"

	// 发送消息 开始读取数据
	model.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "アップロードされたファイルを取得します",
		CurrentStep: "get-file",
		Database:    db,
	}, userID)

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "csvファイルの読み取りに失敗しました",
			CurrentStep: "get-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	// 自动删除minio中的临时文件
	defer func() {
		if len(file.FilePath) > 0 {
			os.Remove(file.FilePath)
			minioClient.DeleteObject(file.FilePath)
		}
		if len(file.ZipFilePath) > 0 {
			os.Remove(file.ZipFilePath)
			minioClient.DeleteObject(file.ZipFilePath)
		}

		// 最后删除public文件夹
		os.Remove("public/app_" + appID + "/temp")
		os.Remove("public/app_" + appID)
	}()

	// 获取文件
	err = model.GetFile(domain, appID, file.FilePath)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "アップロードしたファイルの取得に失敗しました",
			CurrentStep: "get-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	err = model.GetFile(domain, appID, file.ZipFilePath)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "アップロードしたファイルの取得に失敗しました",
			CurrentStep: "get-file",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	// 发送消息 开始读取数据
	model.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "CSVファイル読み込み中",
		CurrentStep: "read-file",
		Database:    db,
	}, userID)

	// 解压缩图片文件
	fileMap := map[string]string{}
	timestamp := timestamp.Timestamp()
	// 解压后图片文件
	if len(file.ZipFilePath) > 0 {
		// 发送消息 开始读取zip文件数据
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "リソースファイルを解凍します。",
			CurrentStep: "read-file",
			Database:    db,
		}, userID)

		// 设置临时解压缩用文件夹
		dst := "temp/zip_" + timestamp + "/"
		// 创建临时文件夹
		filex.Mkdir(dst)
		fileMap, err = filex.UnZipFile(file.ZipFilePath, dst, zipCharset)
		if err != nil {
			store.Set(uploadID, "")
			loggerx.ErrorLog("readCsvFileAndImport", err.Error())
			// 編輯錯誤日誌文件
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "ファイル解凍に失敗しました。",
				CurrentStep: "read-file",
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

	// 删除解压缩临时文件
	defer func() {
		os.Remove(filepath.Dir(file.ZipFilePath))
		var dir = "temp"
		for _, f := range fileMap {
			os.Remove(f)
			dir = filepath.Dir(f)
		}
		os.Remove(dir)
	}()

	// 发送消息 开始读取数据
	model.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "依存データの取得",
		CurrentStep: "check-data",
		Database:    db,
	}, userID)

	// 获取当前app的语言数据
	langData := model.GetLanguageData(db, lang, domain)
	// 获取台账的所有字段
	allFields := model.GetFields(db, datastoreID, appID, roles, true)
	// 获取台账的所有字段
	opList := model.GetOptions(db, appID)

	// 去掉自动採番和函数后的字段
	var fieldList []*field.Field

	var allUsers []*user.User

	for _, fl := range allFields {
		if fl.GetFieldType() == "user" {
			if len(allUsers) == 0 {
				allUsers = model.GetUsers(db, appID, domain)
			}
			fieldList = append(fieldList, fl)
			continue
		}

		// 去掉autonum和function字段
		if fl.GetFieldType() != "autonum" || fl.GetFieldType() != "function" {
			fieldList = append(fieldList, fl)
		}
	}

	cfg, err := model.GetConfig(db, appID)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "台帳情報を取得できませんでした",
			CurrentStep: "check-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		return
	}

	// 获取台账信息
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoreRequest
	// 从共通获取
	req.Database = db
	req.DatastoreId = datastoreID

	response, err := datastoreService.FindDatastore(context.TODO(), &req)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "台帳情報を取得できませんでした",
			CurrentStep: "check-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	ds := response.GetDatastore()

	// 用户组合法可选项
	gpMap := make(map[string]string)

	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	var preq permission.FindActionsRequest
	preq.RoleId = roles
	preq.PermissionType = "app"
	preq.AppId = appID
	preq.ActionType = "datastore"
	preq.ObjectId = datastoreID
	preq.Database = db
	pResp, err := pmService.FindActions(context.TODO(), &preq)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "台帳のアクション情報の取得に失敗しました",
			CurrentStep: "check-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	// 获取数据上传流
	itemService := item.NewItemService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
	}

	stream, err := itemService.ImportItem(context.Background(), opss)
	if err != nil {
		store.Set(uploadID, "")
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据查询错误
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルアップロードの初期化に失敗しました",
			CurrentStep: "check-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	// 上传meta信息
	err = stream.Send(&item.ImportRequest{
		Status: item.SendStatus_SECTION,
		Request: &item.ImportRequest_Meta{
			Meta: &item.ImportMetaData{
				Key:          "",
				AppId:        appID,
				DatastoreId:  datastoreID,
				Writer:       userID,
				Owners:       owners,
				UpdateOwners: updateOwners,
				Database:     db,
			},
		},
	})

	if err != nil {
		store.Set(uploadID, "")
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

		// 发送消息 数据查询错误
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルアップロードメタ送信に失敗しました",
			CurrentStep: "check-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)

		return
	}

	cparam := checkParam{
		db:          db,
		action:      action,
		apiKey:      ds.ApiKey,
		domain:      domain,
		groupID:     groupID,
		datastoreID: datastoreID,
		appID:       appID,
		userID:      userID,
		allFields:   fieldList,
		options:     opList,
		allUsers:    allUsers,
		langData:    langData,
		fileMap:     fileMap,
		roles:       roles,
		owners:      owners,
		relations:   ds.GetRelations(),
		lang:        lang,
		jobID:       jobID,
		wfID:        base.WfId,
		encoding:    encoding,
		gpMap:       gpMap,
		actions:     pResp.GetActions(),
		specialchar: model.GetSpecialChar(cfg.Special),
		emptyChange: base.EmptyChange,
	}

	// excel文件判断,若是则转换为csv
	if file.FilePath[strings.LastIndex(file.FilePath, ".")+1:] != "csv" {
		cparam.encoding = "utf-8"
		file.FilePath, err = model.ExcelToCsv(file.FilePath)
		if err != nil {
			store.Set(uploadID, "")
			loggerx.ErrorLog("readCsvFileAndImport", err.Error())
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "EXCELファイルの読み取りに失敗しました。",
				CurrentStep: "check-data",
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

	// 发送消息 开始进行数据验证（特殊的类型和标题等验证）并生成数据
	model.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     msg.GetMsg(lang, msg.Info, msg.I014, langx.GetLangValue(langData, langx.GetDatastoreKey(appID, datastoreID), langx.DefaultResult)),
		CurrentStep: "check-data",
		Database:    db,
	}, userID)

	items, wfID, err := readFile(cparam, file.FilePath)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		return
	}

	// 如果执行成功
	var errorList []string
	total := int64(len(items))
	var inserted int64 = 0
	var updated int64 = 0

	if len(wfID) == 0 {
		// 验证数据
		go func() {
			// 开始导入
			for _, data := range items {
				err := stream.Send(&item.ImportRequest{
					Status: item.SendStatus_SECTION,
					Request: &item.ImportRequest_Data{
						Data: &data,
					},
				})
				if err == io.EOF {
					return
				}
				if err != nil {
					store.Set(uploadID, "")
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 数据验证错误，停止上传
					model.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     "ファイルのアップロード中にエラーが発生しました。",
						CurrentStep: "upload",
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

			err := stream.Send(&item.ImportRequest{
				Status: item.SendStatus_COMPLETE,
				Request: &item.ImportRequest_Data{
					Data: nil,
				},
			})

			if err != nil {
				store.Set(uploadID, "")
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 数据验证错误，停止上传
				model.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "ファイルのアップロード中にエラーが発生しました。",
					CurrentStep: "upload",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}
		}()

		for {
			result, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				store.Set(uploadID, "")
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 数据验证错误，停止上传
				model.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "ファイルのアップロード中にエラーが発生しました。",
					CurrentStep: "upload",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}

			if result.Status == item.Status_FAILED {
				// 如果有失败的情况发生，将停止继续发送
				// cancel()
				for _, e := range result.GetResult().GetErrors() {
					eMsg := "第{0}〜{1}行目でエラーが発生しました。エラー内容：{2}"
					fieldErrorMsg := "第{0}行目でエラーが発生しました。フィールド名：[{1}]、エラー内容：{2}"
					noFieldErrorMsg := "第{0}行目でエラーが発生しました。エラー内容：{1}"
					if len(e.FieldId) == 0 {
						if e.CurrentLine != 0 {
							es, _ := msg.Format(noFieldErrorMsg, cast.ToString(e.CurrentLine), e.ErrorMsg)
							errorList = append(errorList, es)
						} else {
							es, _ := msg.Format(eMsg, cast.ToString(e.FirstLine), cast.ToString(e.LastLine), e.ErrorMsg)
							errorList = append(errorList, es)
						}
					} else {
						es, _ := msg.Format(fieldErrorMsg, cast.ToString(e.CurrentLine), langx.GetLangValue(langData, langx.GetFieldKey(appID, datastoreID, e.FieldId), langx.DefaultResult), e.ErrorMsg)
						errorList = append(errorList, es)
					}
				}

				continue
			}

			if result.Status == item.Status_SUCCESS {

				inserted = inserted + result.Result.Insert
				updated = updated + result.Result.Modify
				importMsg, _ := json.Marshal(map[string]interface{}{
					"total":    total,
					"inserted": inserted,
					"updated":  updated,
				})

				progress := (inserted + updated) / total * 100

				// 发送消息 收集上传结果
				model.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     string(importMsg),
					CurrentStep: "upload",
					Progress:    int64(progress),
					Insert:      int64(inserted),
					Update:      int64(updated),
					Total:       total,
					Database:    db,
				}, userID)

				continue
			}
		}

		if len(errorList) > 0 {
			path := filex.WriteAndSaveFile(domain, appID, errorList)
			// 发送消息 部分数据上传成功
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     msg.GetMsg(lang, msg.Error, msg.E012, langx.GetLangValue(langData, langx.GetDatastoreKey(appID, datastoreID), langx.DefaultResult)),
				CurrentStep: "upload",
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

			return

		}

		if (inserted + updated) != int64(len(items)) {
			// 发送消息 部分数据上传成功
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "データの一部が正常にインポートされませんでした。",
				CurrentStep: "end",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database:    db,
			}, userID)
			return
		}
		// 发送消息 全部上传成功
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     msg.GetMsg(lang, msg.Info, msg.I018, langx.GetLangValue(langData, langx.GetDatastoreKey(appID, datastoreID), langx.DefaultResult)),
			CurrentStep: "end",
			Progress:    100,
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)

		code := "I_015"
		params := wsx.MessageParam{
			Sender:  "SYSTEM",
			Domain:  domain,
			MsgType: "normal",
			Code:    code,
			Object:  "apps." + appID + ".datastores." + datastoreID,
			Link:    "/datastores/" + datastoreID + "/list",
			Content: "导入数据成功，请刷新浏览器获取最新数据！",
			Status:  "unread",
		}
		wsx.SendToCurrentAndParentGroup(params, db, groupID)
		return
	}

	// 需要审批处理的场合
	var approveErrorList []*item.Error
	approveService := approve.NewApproveService("database", client.DefaultClient)

	// 获取当前流程定义
	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var wreq workflow.WorkflowRequest
	wreq.WfId = wfID
	wreq.Database = db

	fResp, err := workflowService.FindWorkflow(context.TODO(), &wreq)
	if err != nil {
		store.Set(uploadID, "")
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルのアップロード中にエラーが発生しました。",
			CurrentStep: "upload",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userID)
		return
	}

	for _, data := range items {
		im := data.Items
		var aReq approve.AddRequest
		id := im.Items["id"].Value
		line := cast.ToInt64(im.Items["index"].Value)
		row := im.GetItems()
		delete(row, "index")
		delete(row, "id")

		itemMap := map[string]*approve.Value{}
		for key, it := range row {
			itemMap[key] = &approve.Value{
				DataType: it.GetDataType(),
				Value:    it.GetValue(),
			}
		}
		aReq.Items = itemMap
		aReq.Current = itemMap

		itemService := item.NewItemService("database", client.DefaultClient)

		if len(id) > 0 {

			var iReq item.ItemRequest
			iReq.DatastoreId = datastoreID
			iReq.ItemId = id
			iReq.Database = db
			iReq.IsOrigin = true
			iReq.Owners = owners

			iResp, err := itemService.FindItem(context.TODO(), &iReq)
			if err != nil {
				store.Set(uploadID, "")
				approveErrorList = append(approveErrorList, &item.Error{
					CurrentLine: line,
					ErrorMsg:    err.Error(),
				})
				continue
			}

			history := map[string]*approve.Value{}
			items := iResp.GetItem().GetItems()

			params := fResp.GetWorkflow().GetParams()
			fs := params["fields"]
			if len(fs) > 0 {
				fields := strings.Split(fs, ",")
				fieldMap := map[string]string{}
				for _, f := range fields {
					fieldMap[f] = f
				}

				for key, it := range items {
					if _, exist := fieldMap[key]; exist {
						history[key] = &approve.Value{
							DataType: it.GetDataType(),
							Value:    it.GetValue(),
						}
					}
				}
			} else {
				for key, it := range items {
					history[key] = &approve.Value{
						DataType: it.GetDataType(),
						Value:    it.GetValue(),
					}
				}
			}

			aReq.History = history
		}

		aReq.DatastoreId = datastoreID
		aReq.AppId = appID
		aReq.Writer = userID
		aReq.ItemId = id
		aReq.Database = db
		aReq.Domain = domain
		aReq.LangCd = lang
		// 开启流程
		approve := new(wfx.Approve)

		// 添加流程实例
		exID, err := approve.AddExample(db, wfID, userID)
		if err != nil {
			store.Set(uploadID, "")
			approveErrorList = append(approveErrorList, &item.Error{
				CurrentLine: line,
				ErrorMsg:    err.Error(),
			})
			continue
		}

		aReq.ExampleId = exID
		_, err = approveService.AddItem(context.TODO(), &aReq)
		if err != nil {
			store.Set(uploadID, "")
			approveErrorList = append(approveErrorList, &item.Error{
				CurrentLine: line,
				ErrorMsg:    err.Error(),
			})
			continue
		}

		if len(id) > 0 {
			// 数据状态转换成待审批状态
			var statusReq item.StatusRequest
			statusReq.AppId = appID
			statusReq.DatastoreId = datastoreID
			statusReq.ItemId = id
			statusReq.Database = db
			statusReq.Writer = userID
			statusReq.Status = "2"

			_, err = itemService.ChangeStatus(context.TODO(), &statusReq)
			if err != nil {
				store.Set(uploadID, "")
				approveErrorList = append(approveErrorList, &item.Error{
					CurrentLine: line,
					ErrorMsg:    err.Error(),
				})
				continue
			}
		}

		// 流程开始启动
		err = approve.StartExampleInstance(db, wfID, userID, exID, domain)
		if err != nil {
			store.Set(uploadID, "")
			approveErrorList = append(approveErrorList, &item.Error{
				CurrentLine: line,
				ErrorMsg:    err.Error(),
			})
			continue
		}

		inserted++
	}

	if len(approveErrorList) > 0 {
		for _, e := range approveErrorList {
			eMsg := "第{0}〜{1}行目でエラーが発生しました。エラー内容：{2}"
			fieldErrorMsg := "第{0}行目でエラーが発生しました。フィールド名：[{1}]、エラー内容：{2}"
			noFieldErrorMsg := "第{0}行目でエラーが発生しました。エラー内容：{1}"
			if len(e.FieldId) == 0 {
				if e.CurrentLine != 0 {
					es, _ := msg.Format(noFieldErrorMsg, cast.ToString(e.CurrentLine), e.ErrorMsg)
					errorList = append(errorList, es)
				} else {
					es, _ := msg.Format(eMsg, cast.ToString(e.FirstLine), cast.ToString(e.LastLine), e.ErrorMsg)
					errorList = append(errorList, es)
				}
			} else {
				es, _ := msg.Format(fieldErrorMsg, cast.ToString(e.CurrentLine), langx.GetLangValue(langData, langx.GetFieldKey(appID, datastoreID, e.FieldId), langx.DefaultResult), e.ErrorMsg)
				errorList = append(errorList, es)
			}
		}
	}

	if len(errorList) > 0 {
		path := filex.WriteAndSaveFile(domain, appID, errorList)
		// 发送消息 部分数据上传成功
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     msg.GetMsg(lang, msg.Info, msg.I017, langx.GetLangValue(langData, langx.GetDatastoreKey(appID, datastoreID), langx.DefaultResult)),
			CurrentStep: "end",
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database: db,
		}, userID)

		return

	}

	if (inserted + updated) != int64(len(items)) {
		if len(wfID) > 0 {
			// 发送消息 所有数据都进入审批流程状态
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "データは承認状態になりました。",
				CurrentStep: "end",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database:    db,
			}, userID)
		} else {
			// 发送消息 部分数据上传成功
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "データの一部が正常にインポートされませんでした。",
				CurrentStep: "end",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database:    db,
			}, userID)
		}
		return
	}

	// 发送消息 全部上传成功
	model.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     msg.GetMsg(lang, msg.Info, msg.I018, langx.GetLangValue(langData, langx.GetDatastoreKey(appID, datastoreID), langx.DefaultResult)),
		CurrentStep: "end",
		Progress:    100,
		EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
		Database:    db,
	}, userID)

	code := "I_015"
	params := wsx.MessageParam{
		Sender:  "SYSTEM",
		Domain:  domain,
		MsgType: "normal",
		Code:    code,
		Object:  "apps." + appID + ".datastores." + datastoreID,
		Link:    "/datastores/" + datastoreID + "/list",
		Content: "导入数据成功，请刷新浏览器获取最新数据！",
		Status:  "unread",
	}
	wsx.SendToCurrentAndParentGroup(params, db, groupID)
}

// readFile 读取文件
func readFile(p checkParam, filePath string) (data []item.ImportData, wf string, e error) {

	eConflict := "第{0}行目でエラーが発生しました。エラー内容：更新処理と追加処理を同時に使用することはできません。"

	var errorList []string
	// 读取文件
	fs, err := os.Open(filePath)
	if err != nil {
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       p.jobID,
			Message:     "ファイルの読み取りに失敗しました。",
			CurrentStep: "check-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: p.db,
		}, p.userID)

		return data, p.wfID, err
	}

	defer fs.Close()

	var r *csv.Reader

	if p.encoding == "sjis" {
		converter := transform.NewReader(fs, japanese.ShiftJIS.NewDecoder())
		r = csv.NewReader(converter)
		r.LazyQuotes = true
	} else {
		r = csv.NewReader(fs)
		r.LazyQuotes = true
	}

	accesskeys := sessionx.GetAccessKeys(p.db, p.userID, p.datastoreID, "W")

	//针对大文件，一行一行的读取文件
	index := 0
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			loggerx.ErrorLog("readCsvFileAndImport", err.Error())
			path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})

			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       p.jobID,
				Message:     "ファイルの読み取りに失敗しました。",
				CurrentStep: "check-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: p.db,
			}, p.userID)

			return data, wf, err
		}
		if err == io.EOF {
			break
		}
		// 验证行数据是否只包含逗号，只有逗号的行不合法
		isValid, errmsg := filex.CheckRowDataValid(row, index)
		if !isValid {
			path := filex.WriteAndSaveFile(p.domain, p.appID, []string{errmsg})

			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       p.jobID,
				Message:     errmsg,
				CurrentStep: "check-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: p.db,
			}, p.userID)

			return data, wf, errors.New(errmsg)
		}

		if index == 1 {
			hasEmpty := false
			for _, h := range row {
				if h == "" || h == "　" {
					hasEmpty = true
				}
			}

			if hasEmpty {
				path := filex.WriteAndSaveFile(p.domain, p.appID, []string{"csvヘッダー行に空白の列名があります。修正してください。"})

				// 发送消息 数据验证错误，停止上传
				model.ModifyTask(task.ModifyRequest{
					JobId:       p.jobID,
					Message:     "csvファイルの形式が正しくありません。",
					CurrentStep: "check-data",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: p.db,
				}, p.userID)

				return data, p.wfID, errors.New("csvヘッダー行に空白の列名があります。修正してください。")
			}

			// 判断当前传入参数中是否有流程信息,并验证是否符合流程处理规则
			if len(p.wfID) == 0 {
				// 流程获取
				wks := wfx.GetUserWorkflow(p.db, p.groupID, p.appID, p.datastoreID, p.action)
				if len(wks) > 0 {
					if p.action == "update" {
						var fields []string
						for _, wk := range wks {
							if len(wk.Params["fields"]) == 0 {
								p.wfID = wk.GetWfId()
							} else {
								fs := strings.Split(wk.Params["fields"], ",")
								fields = append(fields, fs...)
							}
						}

						if len(p.wfID) == 0 {
							for _, f := range fields {
								exist := model.CheckExist(f, row[1:])
								if exist {
									path := filex.WriteAndSaveFile(p.domain, p.appID, []string{"ワークフローが必要です。"})
									// 发送消息 数据验证错误，停止上传
									model.ModifyTask(task.ModifyRequest{
										JobId:       p.jobID,
										Message:     "ワークフローが必要です。",
										CurrentStep: "check-data",
										EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
										ErrorFile: &task.File{
											Url:  path.MediaLink,
											Name: path.Name,
										},
										Database: p.db,
									}, p.userID)

									return data, wf, err
								}
							}
						}
					} else {
						p.wfID = wks[0].GetWfId()
					}
				}
			} else {
				// 获取当前流程定义
				workflowService := workflow.NewWfService("workflow", client.DefaultClient)

				var wreq workflow.WorkflowRequest
				wreq.WfId = p.wfID
				wreq.Database = p.db

				fResp, err := workflowService.FindWorkflow(context.TODO(), &wreq)
				if err != nil {
					path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})

					// 发送消息 数据验证错误，停止上传
					model.ModifyTask(task.ModifyRequest{
						JobId:       p.jobID,
						Message:     "ワークフローの取得に失敗しました。",
						CurrentStep: "check-data",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: p.db,
					}, p.userID)

					return data, p.wfID, err
				}

				if !fResp.GetWorkflow().GetIsValid() {
					path := filex.WriteAndSaveFile(p.domain, p.appID, []string{"ワークフローは無効になっています。もう一度選択して、もう一度お試しください"})

					// 发送消息 数据验证错误，停止上传
					model.ModifyTask(task.ModifyRequest{
						JobId:       p.jobID,
						Message:     "ワークフローは無効です。",
						CurrentStep: "check-data",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: p.db,
					}, p.userID)
					return data, p.wfID, err
				}

				params := fResp.GetWorkflow().GetParams()

				if p.action == "update" {
					fs := params["fields"]
					if len(fs) > 0 {
						fields := strings.Split(fs, ",")

						fieldList := row[1:]

						if len(fields) != len(fieldList) {
							path := filex.WriteAndSaveFile(p.domain, p.appID, []string{"ワークフローによって定義されたフィールドがcsvフィールドと一致しません。"})

							// 发送消息 数据验证错误，停止上传
							model.ModifyTask(task.ModifyRequest{
								JobId:       p.jobID,
								Message:     "ワークフローによって定義されたフィールドがcsvフィールドと一致しません。",
								CurrentStep: "check-data",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: p.db,
							}, p.userID)

							return data, p.wfID, errors.New("data has error")
						}

						eMsg := "ワークフロー更新フィールド[{0}]が存在しません"

						for _, f := range fields {
							exist := model.CheckExist(f, row)
							if !exist {
								es, _ := msg.Format(eMsg, f)
								errorList = append(errorList, es)
							}
						}

						if len(errorList) > 0 {
							path := filex.WriteAndSaveFile(p.domain, p.appID, errorList)

							// 发送消息 数据验证错误，停止上传
							model.ModifyTask(task.ModifyRequest{
								JobId:       p.jobID,
								Message:     "ワークフローで定義されたフィールドが存在しません。",
								CurrentStep: "check-data",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: p.db,
							}, p.userID)

							return data, p.wfID, errors.New("data has error")
						}
					}
				}
			}

			// 判断是否需要owner信息
			exist := model.CheckExist("owner", row)
			if exist {
				groupService := group.NewGroupService("manage", client.DefaultClient)

				var groupReq group.FindGroupsRequest
				groupReq.Domain = p.domain
				groupReq.Database = p.db

				resGroup, err := groupService.FindGroups(context.TODO(), &groupReq)
				if err != nil {
					path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})
					// 发送消息 数据验证错误，停止上传
					model.ModifyTask(task.ModifyRequest{
						JobId:       p.jobID,
						Message:     "グループ情報取得処理でエラーが発生した",
						CurrentStep: "check-data",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: p.db,
					}, p.userID)

					return data, p.wfID, errors.New("グループ情報取得処理でエラーが発生した")
				}

				for _, g := range model.GetValidGroupsByID(p.groupID, resGroup.Groups) {
					p.gpMap[g.GroupId] = g.AccessKey
				}
			}

			p.headerData = append([]string{"index"}, row...)
			index++
			continue
		}

		if index > 1 {
			id := row[0]

			rd := rowData{
				index: index - 1,
			}

			if (p.action == "insert") && len(id) > 0 {
				es, _ := msg.Format(eConflict, cast.ToString(index+1))
				errorList = append(errorList, es)
			}

			if p.action == "update" {
				if len(id) == 0 {
					es, _ := msg.Format(eConflict, cast.ToString(index+1))
					errorList = append(errorList, es)
					index++
					continue
				}

				itemService := item.NewItemService("database", client.DefaultClient)

				var req item.ItemRequest
				req.DatastoreId = p.datastoreID
				req.ItemId = id
				req.Database = p.db
				req.IsOrigin = true
				req.Owners = accesskeys

				resp, err := itemService.FindItem(context.TODO(), &req)
				if err != nil {
					errorList = append(errorList, "データが存在しないか、データを変更する権限がありません")
					index++
					continue
				}

				rd.item = resp.GetItem()
			}

			// 将当前的数据添加到切分的数组中
			indexRow := append([]string{cast.ToString(index - 1)}, row...)
			rd.data = indexRow

			p.fileData = rd

			// 数据验证并生成数据
			it, checkErrors := checkAndBuildItems(p)
			if len(checkErrors) > 0 {
				fieldErrorMsg := "第{0}行目でエラーが発生しました。フィールド名：[{1}]、エラー内容：{2}"
				noFieldErrorMsg := "第{0}行目でエラーが発生しました。エラー内容：{1}"

				for _, e := range checkErrors {
					if len(e.FieldId) == 0 {
						es, _ := msg.Format(noFieldErrorMsg, cast.ToString(e.CurrentLine), e.ErrorMsg)
						errorList = append(errorList, es)
					} else {
						es, _ := msg.Format(fieldErrorMsg, cast.ToString(e.CurrentLine), langx.GetLangValue(p.langData, langx.GetFieldKey(p.appID, p.datastoreID, e.FieldId), langx.DefaultResult), e.ErrorMsg)
						errorList = append(errorList, es)
					}
				}

				index++
				continue
			}

			typeErrors := model.CheckDataType(it)
			if len(typeErrors) > 0 {

				fieldErrorMsg := "第{0}行目でエラーが発生しました。フィールド名：[{1}]、エラー内容：{2}"

				for _, e := range typeErrors {
					es, _ := msg.Format(fieldErrorMsg, cast.ToString(e.CurrentLine), langx.GetLangValue(p.langData, langx.GetFieldKey(p.appID, p.datastoreID, e.FieldId), langx.DefaultResult), e.ErrorMsg)
					errorList = append(errorList, es)
				}

				index++
				continue
			}

			data = append(data, item.ImportData{
				Items: it,
			})
		}

		index++
	}

	if len(errorList) > 0 {
		path := filex.WriteAndSaveFile(p.domain, p.appID, errorList)

		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       p.jobID,
			Message:     "データ検証エラー",
			CurrentStep: "check-data",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: p.db,
		}, p.userID)

		return data, p.wfID, errors.New("data has error")
	}

	return
}

// checkAndBuildItems 生成台账数据
func checkAndBuildItems(p checkParam) (list *item.ListItems, result []*item.Error) {

	headers := p.headerData
	data := p.fileData

	line := cast.ToInt64(data.index)

	appLangData := p.langData.Apps[p.appID]
	groupLangData := p.langData.Common.Groups

	var checkDataExistError []*item.Error

	if p.action == "image" {

		set := containerx.New()
		for _, act := range p.actions {
			if act.ObjectId == p.datastoreID {
				set.AddAll(act.Fields...)
			}
		}

		var allFields []*field.Field
		for _, f := range p.allFields {
			if f.FieldType == "file" {
				allFields = append(allFields, f)
			}
		}

		fieldList := set.ToList()
		var fileFields []*field.Field
		for _, fieldID := range fieldList {
			f, err := model.FindField(fieldID, allFields)
			if err == nil {
				fileFields = append(fileFields, f)
			}
		}

		for _, h := range headers {
			if h != "id" && h != "index" {
				exist := false
			LP2:
				for _, f := range fileFields {
					if f.FieldId == h {
						exist = true
						break LP2
					}
				}

				if !exist {
					checkDataExistError = append(checkDataExistError, &item.Error{
						CurrentLine: line,
						FieldId:     h,
						ErrorMsg:    "このフィールドはファイルタイプフィールドではないため、アップロードできません。",
					})
				}
			}

		}

		if len(checkDataExistError) > 0 {
			return nil, checkDataExistError
		}
	} else {

		set := containerx.New()
		for _, act := range p.actions {
			if act.ObjectId == p.datastoreID {
				set.AddAll(act.Fields...)
			}
		}

		var allFields []*field.Field
		for _, f := range p.allFields {
			if f.IsRequired {
				allFields = append(allFields, f)
			}
		}

		fieldList := set.ToList()
		var requiredFields []*field.Field
		for _, fieldID := range fieldList {
			f, err := model.FindField(fieldID, allFields)
			if err == nil {
				requiredFields = append(requiredFields, f)
			}
		}

		if p.action == "insert" {
			for _, f := range requiredFields {
				exist := false
			LP1:
				for _, h := range headers {
					if f.FieldId == h {
						exist = true
						break LP1
					}
				}

				if !exist {
					checkDataExistError = append(checkDataExistError, &item.Error{
						CurrentLine: line,
						FieldId:     f.FieldId,
						ErrorMsg:    "このフィールドは必須であり、csvファイルに対応するデータ列はありません",
					})
				}
			}

			if len(checkDataExistError) > 0 {
				return nil, checkDataExistError
			}
		}
	}

	itemService := item.NewItemService("database", client.DefaultClient)

	cols := make(map[string]*item.Value, len(data.data))
	for index, col := range data.data {
		field := headers[index]
		// 第一列为行号的场合
		if index == 0 {
			// 设置值为
			cols[field] = &item.Value{
				DataType: "number",
				Value:    col,
			}
			continue
		}
		// 第二列为ID的场合
		if index == 1 {
			cols[field] = &item.Value{
				DataType: "text",
				Value:    col,
			}
			continue
		}
		// 所属组织情报列的场合
		if field == "owner" {
			if len(col) != 0 {
				if groupAccessKey, exist := p.gpMap[model.GetGroupValue(col, groupLangData)]; exist {
					cols[field] = &item.Value{
						DataType: "text",
						Value:    groupAccessKey,
					}
				} else {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "[" + col + "]" + "ユーザーグループが存在しないか、選択できません。"})
				}
			}
			continue
		}

		fieldInfo := model.CheckFieldExist(field, p.allFields)
		if fieldInfo == nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("[%s]このフィールドが見つかりません", field)})
			return nil, checkDataExistError
		}

		// 新规作成的场合
		if p.action == "insert" {
			// 如果当前值，等于默认空白更新内容，则直接替换为空白，然后继续进行后续判断
			if col == DefaultEmptyStr {
				col = ""
			}
		} else {
			// 如果空白表示不更新的情况下
			// 判断当前值，是否是空白，如果是空白，直接不更新该字段，跳出当前循环。
			if !p.emptyChange && len(col) == 0 {
				continue
			}

			// 如果当前值，等于默认空白更新内容，则直接替换为空白，然后继续进行后续判断
			if col == DefaultEmptyStr {
				col = ""
			}
		}

		if fieldInfo.FieldType == "user" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}

			if len(col) == 0 {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    strings.Join([]string{}, ","),
				}
			} else {
				users := strings.Split(col, ",")
				var userList []string
				for _, u := range users {
					un := model.ReTranUser(u, p.allUsers)
					if len(un) == 0 {
						checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
							ErrorMsg: "[" + u + "]" + "ユーザーが存在しません。",
						})
					}
					userList = append(userList, model.ReTranUser(u, p.allUsers))
				}

				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    strings.Join(userList, ","),
				}
			}

			continue
		}
		if fieldInfo.FieldType == "lookup" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}

			if len(col) == 0 {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    "",
				}
			} else {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    col,
				}
			}

			continue
		}
		if fieldInfo.FieldType == "options" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}

			if len(col) == 0 {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    "",
				}
			} else {
				group := fieldInfo.GetOptionId()
				value := model.GetOptionValue(group, col, appLangData)
				if len(value) == 0 {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]" + "有効なオプションが存在しません。",
					})
				}

				if !model.CheckOptionValid(group, value, p.options) {
					if len(value) == 0 {
						checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
							ErrorMsg: "[" + col + "]" + "有効なオプションが存在しません。",
						})
					}
				}

				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    value,
				}
			}
			continue
		}
		if fieldInfo.FieldType == "date" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}

			if len(col) == 0 {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    "",
				}
			} else {
				time, err := timex.ToTimeE(col)
				if err != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]" + "は有効な日付ではありません。",
					})
					continue
				}

				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    time.Format("2006-01-02"),
				}
			}
			continue
		}
		if fieldInfo.FieldType == "time" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}

			if len(col) == 0 {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    "",
				}
			} else {
				_, e := time.Parse("15:04:05", col)
				if e != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]" + "は有効な時間ではありません。",
					})
				}

				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    col,
				}
			}
			continue
		}
		if fieldInfo.FieldType == "file" {
			if len(col) == 0 {
				fvStr, _ := json.Marshal([]FileValue{})
				value := string(fvStr)
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    value,
				}
			} else {
				fileNames := strings.Split(col, ",")
				var fileList []*FileValue
				for _, fileName := range fileNames {

					localPath, exist := p.fileMap[fileName]
					if exist {
						fileLink, err := uploadFile(localPath, p.domain, p.appID, p.datastoreID)
						if err != nil {
							checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
								ErrorMsg: "ファイルのアップロードに失敗しました",
							})
							continue
						}

						file := &FileValue{
							Name: fileName,
							URL:  fileLink,
						}

						fileList = append(fileList, file)
					} else {
						checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
							ErrorMsg: "イメージ" + "[" + fileName + "]" + "がアップロードされていません。",
						})
					}
				}
				fvStr, _ := json.Marshal(fileList)
				value := string(fvStr)
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    value,
				}
			}
			continue
		}
		if fieldInfo.FieldType == "text" || fieldInfo.FieldType == "textarea" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}

			if len(col) == 0 {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    "",
				}
			} else {
				if model.CharCount(col) < int(fieldInfo.MinLength) {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]=>" + "フィールドの長さが" + cast.ToString(fieldInfo.MinLength) + "桁以上でなければなりません。",
					})
				}
				if model.CharCount(col) > int(fieldInfo.MaxLength) {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]=>" + "フィールドの長さが" + cast.ToString(fieldInfo.MaxLength) + "桁以下でなければなりません。",
					})
				}
				// 验证是否包含无效特殊字符
				if !model.SpecialCheck(col, p.specialchar) {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]=>" + "無効な特殊文字があります",
					})
				}

				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    col,
				}
			}
			continue
		}
		if fieldInfo.FieldType == "number" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}

			if len(col) == 0 {
				cols[field] = &item.Value{
					DataType: fieldInfo.FieldType,
					Value:    "0",
				}
			} else {
				numValue, err := cast.ToFloat64E(col)
				if err != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]" + "は数値ではありません。",
					})
				} else {
					if numValue < float64(fieldInfo.MinValue) {
						checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
							ErrorMsg: "[" + col + "]=>" + "フィールドの大きさが" + cast.ToString(fieldInfo.MinValue) + "以上でなければなりません。",
						})
					} else if numValue > float64(fieldInfo.MaxValue) {
						checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
							ErrorMsg: "[" + col + "]=>" + "フィールドの大きさが" + cast.ToString(fieldInfo.MaxValue) + "以下でなければなりません。",
						})
					} else {
						cols[field] = &item.Value{
							DataType: fieldInfo.FieldType,
							Value:    floatx.ToFixedString(numValue, fieldInfo.GetPrecision()),
						}
					}
				}
			}
		}
		// 其他场合
		if fieldInfo.IsRequired && len(col) == 0 {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
				ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
			})
			continue
		}

		cols[field] = &item.Value{
			DataType: fieldInfo.FieldType,
			Value:    col,
		}
	}

	// TODO redis查重
	// rds := redisx.New()

	// 检查关系是否存在
	for _, rat := range p.relations {

		var keys []string

		// TODO redis查重
		// existKey := strings.Builder{}
		// existKey.WriteString("item_")
		// existKey.WriteString(rat.GetDatastoreId())
		// existKey.WriteString("_")

		param := item.CountRequest{
			AppId:         p.appID,
			DatastoreId:   rat.DatastoreId,
			ConditionList: []*item.Condition{},
			ConditionType: "and",
			Database:      p.db,
		}

		// 第一步，判断当前关系字段是否存在于传入数据中
		var existCount = 0
		var emptyCount = 0
		for relationKey, localKey := range rat.Fields {
			name := appLangData.Fields[p.datastoreID+"_"+localKey]
			keys = append(keys, name)

			if val, ok := cols[localKey]; ok {
				if len(val.Value) == 0 {
					emptyCount++
				}
				// TODO redis查重
				// existKey.WriteString("items.")
				// existKey.WriteString(relationKey)
				// existKey.WriteString(".value_")
				// existKey.WriteString(cols[localKey].GetValue())

				param.ConditionList = append(param.ConditionList, &item.Condition{
					FieldId:     relationKey,
					FieldType:   "text",
					SearchValue: cols[localKey].GetValue(),
					Operator:    "=",
					IsDynamic:   true,
				})

				existCount++
			}
		}

		// 如果全部不存在,直接跳过
		if existCount == 0 {
			continue
		}

		// 如果部分存在
		if existCount < len(rat.Fields) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("%v関連アイテムには、対応するすべてのデータが渡されていません", keys)})
			return nil, checkDataExistError
		}

		// 如果当前关系的所有值都是空，则跳过检查
		if emptyCount == len(rat.Fields) {
			continue
		}

		// TODO redis查重
		// fmt.Println(existKey.String())

		// total, err := rds.Exists(context.TODO(), existKey.String()).Result()
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("%v関連アイテムのデータは存在しません。", keys)})
		// 	return nil, nil, checkDataExistError
		// }
		response, err := itemService.FindCount(context.TODO(), &param)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("%v関連アイテムのデータは存在しません。", keys)})
			return nil, checkDataExistError
		}

		if response.Total == 0 {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("%v関連アイテムのデータは存在しません。", keys)})
			return nil, checkDataExistError
		}
	}

	// 更新场合的ID必须检查
	if (p.action == "update" || p.action == "image") && len(cols["id"].GetValue()) == 0 {
		checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "更新の場合、データIDを空にすることはできません"})

		return nil, checkDataExistError
	}

	// 状态判断
	if data.item != nil {
		// 审批状态check
		status := data.item.Status
		if model.NonAdmitCheck(status) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "データはプロセス承認中であり、更新できません。"})
			return nil, checkDataExistError
		}
	}

	rowMap := cols
	// 追加处理用字段
	rowMap["action"] = &item.Value{
		DataType: "text",
		Value:    p.action,
	}

	list = &item.ListItems{
		Items: rowMap,
	}

	return list, checkDataExistError
}

// uploadFile 上传文件
func uploadFile(filePath string, domain, appID, datastoreID string) (mpath string, err error) {

	// 文件mime类型
	contentType := getContentType(filePath)

	fo, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer fo.Close()

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(filePath)

	appRoot := "app_" + appID
	datastoreRoot := "datastore_" + datastoreID
	fp := path.Join(appRoot, "data", datastoreRoot, fileName)
	// 文件上传minio文件服务器,获取路径
	result, err := minioClient.SavePublicObject(fo, fp, contentType)
	if err != nil {
		return "", err
	}
	// 判断顾客上传文件是否在设置的最大存储空间以内
	canUpload := filex.CheckCanUpload(domain, float64(result.Size))
	if canUpload {
		// 如果没有超出最大值，就对顾客的已使用大小进行累加
		err = filex.ModifyUsedSize(domain, float64(result.Size))
		if err != nil {
			return "", err
		}

		mpath = result.MediaLink
		return
	}

	// 如果已达上限，则删除刚才上传的文件
	minioClient.DeleteObject(result.Name)
	return "", errors.New("最大ストレージ容量に達しました。ファイルのアップロードに失敗しました")
}
