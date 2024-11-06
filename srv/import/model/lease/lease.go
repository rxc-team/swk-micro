package lease

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/import/common/filex"
	"rxcsoft.cn/pit3/srv/import/common/floatx"
	"rxcsoft.cn/pit3/srv/import/common/langx"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/common/storex"
	"rxcsoft.cn/pit3/srv/import/model"
	"rxcsoft.cn/pit3/srv/import/system/sessionx"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
	"rxcsoft.cn/utils/timex"
)

// Import 文件导入并上传
func Import(base Params, file FileParams) {
	// 获取传入变量
	jobID := base.JobId
	action := base.Action
	encoding := base.Encoding
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
	firstMonth := base.FirstMonth
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
		if len(file.PayFilePath) > 0 {
			os.Remove(file.PayFilePath)
			minioClient.DeleteObject(file.PayFilePath)
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
	err = model.GetFile(domain, appID, file.PayFilePath)
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

	var payData map[string]PayData
	// 读取支付文件数据
	if len(file.PayFilePath) > 0 {
		// excel文件判断,若是则转换为csv
		if file.PayFilePath[strings.LastIndex(file.PayFilePath, ".")+1:] != "csv" {
			encoding = "utf-8"
			file.PayFilePath, err = model.ExcelToCsv(file.PayFilePath)
			if err != nil {
				store.Set(uploadID, "")
				loggerx.ErrorLog("readCsvFileAndImport", err.Error())
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 数据验证错误，停止上传
				model.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     "支払いファイルのEXCELファイルの読み取りに失敗しました。",
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

		pay, err := buildPayData(file.PayFilePath)
		if err != nil {
			store.Set(uploadID, "")
			loggerx.ErrorLog("readCsvFileAndImport", err.Error())
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})

			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "支払いファイルの読み取りに失敗しました。",
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
		payData = pay
	}

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

	var req datastore.DatastoresRequest
	// 从共通获取
	req.Database = db
	req.AppId = appID

	response, err := datastoreService.FindDatastores(context.TODO(), &req)
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

	dsMap := make(map[string]string)
	var ds *datastore.Datastore
	for _, d := range response.GetDatastores() {
		if d.DatastoreId == datastoreID {
			ds = d
		}
		dsMap[d.ApiKey] = d.GetDatastoreId()
	}

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

	cparam := checkParam{
		db:          db,
		action:      action,
		domain:      domain,
		groupID:     groupID,
		datastoreID: datastoreID,
		appID:       appID,
		userID:      userID,
		handleMonth: cfg.GetSyoriYm(),
		beginMonth:  cfg.GetKishuYm(),
		smallAmount: cfg.GetMinorBaseAmount(),
		shortPeriod: cfg.GetShortLeases(),
		allFields:   fieldList,
		options:     opList,
		allUsers:    allUsers,
		langData:    langData,
		roles:       roles,
		owners:      owners,
		payData:     payData,
		relations:   ds.GetRelations(),
		lang:        lang,
		jobID:       jobID,
		encoding:    encoding,
		dsMap:       dsMap,
		gpMap:       gpMap,
		actions:     pResp.GetActions(),
		specialchar: model.GetSpecialChar(cfg.Special),
		firstMonth:  firstMonth,
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

	items, err := readFile(cparam, file.FilePath)
	if err != nil {
		store.Set(uploadID, "")
		loggerx.ErrorLog("readCsvFileAndImport", err.Error())
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

	// 如果执行成功
	var errorList []string
	total := int64(len(items))
	var inserted int64 = 0
	var updated int64 = 0

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

	// code := "I_015"
	// params := wsx.MessageParam{
	// 	Sender:  "SYSTEM",
	// 	Domain:  domain,
	// 	MsgType: "normal",
	// 	Code:    code,
	// 	Object:  "apps." + appID + ".datastores." + datastoreID,
	// 	Link:    "/datastores/" + datastoreID + "/list",
	// 	Content: "导入数据成功，请刷新浏览器获取最新数据！",
	// 	Status:  "unread",
	// }
	// wsx.SendToCurrentAndParentGroup(params, db, groupID)
	return
}

// readFile 读取文件
func readFile(p checkParam, filePath string) (data []item.ImportData, e error) {

	// eConflict := "第{0}行目でエラーが発生しました。エラー内容：更新処理と追加処理を同時に使用することはできません。"

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

		return data, err
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

			return data, err
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

			return data, errors.New(errmsg)
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

				return data, errors.New("csvヘッダー行に空白の列名があります。修正してください。")
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
					return data, errors.New("グループ情報取得処理でエラーが発生した")
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
			// id := row[0]

			// if p.action == "contract-insert" && len(id) > 0 {
			// 	es, _ := msg.Format(eConflict, cast.ToString(index+1))
			// 	errorList = append(errorList, es)
			// }

			// if p.action == "debt-change" || p.action == "info-change" || p.action == "midway-cancel" || p.action == "contract-expire" {
			// 	if len(id) == 0 {
			// 		es, _ := msg.Format(eConflict, cast.ToString(index+1))
			// 		errorList = append(errorList, es)
			// 	}
			// }

			// 将当前的数据添加到切分的数组中
			indexRow := append([]string{cast.ToString(index - 1)}, row...)

			p.fileData = indexRow

			// 数据验证并生成数据
			it, attachItems, checkErrors := checkAndBuildItems(p)
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
				Items:       it,
				AttachItems: attachItems,
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

		return data, errors.New("data has error")
	}

	return
}

// checkAndBuildItems 生成台账数据
func checkAndBuildItems(p checkParam) (list *item.ListItems, attachItems attachData, result []*item.Error) {

	headers := p.headerData
	data := p.fileData

	line := cast.ToInt64(data[0])

	appLangData := p.langData.Apps[p.appID]
	groupLangData := p.langData.Common.Groups

	var checkDataExistError []*item.Error

	// 必须属性字段
	var requiredFields []string

	if p.action == "contract-insert" {
		// 机能必须项目
		actionNeedFields := []string{
			//"keiyakuno", // 契約番号
			// "shisannm",        // 資産名称
			// "torihikikbn",     // 取引判定区分
			// "leasekaishacd", // リース会社
			// "assetlife",       // 耐用年数
			"leasestymd",      // リース開始日
			"leasekikan",      // リース期間
			"paymentstymd",    // 初回支払日
			"paymentcycle",    // 支払サイクル
			"paymentday",      // 支払日
			"paymentcounts",   // 支払回数
			"paymentleasefee", // 支払リース料
			// "rishiritsu",      // 追加借入利率
			//添加部分
			"sykshisankeisan", //使用権資産の計算方法*
			"biko1",           //備考１
			// "hkkjitenzan",     //比較開始時点の残存リース料*
			"firstleasefee", //初回リース料
			"finalleasefee", //最終回リース料
			// 以下是值可以为空但列必须存在的字段
			"keiyakuymd",         // 契約年月日
			"extentionOption",    // 延長リース期間
			"bunruicd",           // 分類コード
			"keiyakunm",          // 契約名称
			"segmentcd",          // 管理部門
			"initialDirectCosts", // 当初直接費用
			"restorationCosts",   // 原状回復コスト
			"field_viw",          //セグメント01
			"field_22c",          //セグメント02
			"field_1av",          //セグメント03
			"field_206",          //セグメント04
			"field_14l",          //セグメント05
			"field_7p3",          //任意マスタ01
			"field_248",          //任意マスタ02
			"field_3k7",          //任意マスタ03
			"field_1vg",          //任意マスタ04
			"field_5fj",          //任意マスタ05
			"field_20h",          //任意項目01
			"field_2h1",          //任意項目02
			"field_qi4",          //任意項目03
			"field_1ck",          //任意項目04
			"field_u1q",          //任意項目05
			// "incentivesAtOrPrior",     // インセンティブ
			// "cancellationrightoption", // 解約行使権オプション
			// "kaiyakuymd",              // 解約年月
			// "optionToPurchase",        // 購入オプション行使価額
			// "paymentsAtOrPrior",       // 前払リース料
		}
		requiredFields = actionNeedFields
	} else if p.action == "debt-change" {
		// 债务变更的场合,追加相应必须字段
		actionNeedFields := []string{
			"henkouymd",     // 变更年月日
			"leasekikan",    // リース期間
			"percentage",    // 百分比
			"paymentcounts", // 支払回数
			"rishiritsu",    // 追加借入利率
		}
		requiredFields = actionNeedFields
	} else if p.action == "info-change" {
		// 情报变更的场合,追加相应必须字段
		actionNeedFields := []string{
			"henkouymd", // 变更年月日
			"bunruicd",  // 分類コード
			"keiyakunm", // 契約名称
			"shisannm",  // 資産名称
			"segmentcd", // 管理部門
		}
		requiredFields = actionNeedFields
	} else if p.action == "midway-cancel" {
		// 中途解约的场合,追加相应必须字段
		actionNeedFields := []string{
			"kaiyakuymd", // 解约年月日
		}
		requiredFields = actionNeedFields
	} else if p.action == "contract-expire" {
		// 满了的场合,追加相应必须字段
		actionNeedFields := []string{
			"henkouymd",         // 变更年月日
			"expiresyokyakukbn", // リース満了償却区分
		}
		requiredFields = actionNeedFields
	}

	for _, f := range requiredFields {
		exist := false
	LP:
		for _, h := range headers {
			if f == h {
				exist = true
				break LP
			}
		}

		if !exist {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				FieldId:     f,
				ErrorMsg:    "このフィールドは必須であり、csvファイルに対応するデータ列はありません",
			})
		}
	}

	if len(checkDataExistError) > 0 {
		return nil, nil, checkDataExistError
	}

	// 契约番号唯一性检查用
	keiyakunoMap := make(map[string]struct{})
	// 更新番号唯一性检查用
	updateIdMap := make(map[string]struct{})

	itemService := item.NewItemService("database", client.DefaultClient)

	cols := make(map[string]*item.Value, len(data))
	for index, col := range data {
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
		// // 第二列为ID的场合
		// if index == 1 {
		// 	cols[field] = &item.Value{
		// 		DataType: "text",
		// 		Value:    col,
		// 	}
		// 	continue
		// }
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
			return nil, nil, checkDataExistError
		}

		// 如果是契约登录的场合，必须check判断
		if p.action == "contract-insert" {
			// 必须check
			switch field {
			case
				//"keiyakuno", // 契約番号
				// "leasekaishacd",   // リース会社
				// "shisannm",        // 資産名称
				// "torihikikbn",     // 取引判定区分
				// "assetlife",       // 耐用年数
				// "keiyakuymd",      // 契約年月日
				// "bunruicd",        // 分類コード
				// "keiyakunm",       // 契約名称
				// "segmentcd",       // 管理部門
				"leasestymd",      // リース開始日
				"leasekikan",      // リース期間
				"paymentcycle",    // 支払サイクル
				"paymentday",      // 支払日
				"paymentcounts",   // 支払回数
				"paymentleasefee", // 支払リース料
				// "rishiritsu",      // 追加借入利率
				"sykshisankeisan", //使用権資産の計算方法
				"paymentstymd":    // 初回支払日
				// "hkkjitenzan":     //比較開始時点の残存リース料
				if len(col) == 0 {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
					})
				}
			}
		}
		if p.action == "debt-change" {
			switch field {
			case
				"henkouymd",     // 变更年月日
				"leasekikan",    // リース期間
				"percentage",    // 百分比
				"paymentcounts", // 支払回数
				"rishiritsu":    // 追加借入利率
				if len(col) == 0 {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
					})
				}
			}
		}
		if p.action == "info-change" {
			switch field {
			case
				"henkouymd", // 变更年月日
				"bunruicd",  // 分類コード
				"keiyakunm", // 契約名称
				"shisannm",  // 資産名称
				"segmentcd": // 管理部門
				if len(col) == 0 {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
					})
				}
			}
		}
		if p.action == "midway-cancel" {
			if field == "kaiyakuymd" && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
			}
		}
		if p.action == "contract-expire" {
			switch field {
			case
				"henkouymd",         // 变更年月日
				"expiresyokyakukbn": // リース満了償却区分
				if len(col) == 0 {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
					})
				}
			}
		}

		// 契约必须检查已完成,下文行数据解析无需再检查
		fieldInfo.IsRequired = false

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
				// value := model.GetOptionValue(group, col, appLangData)
				value := col
				// if len(value) == 0 {
				// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
				// 		ErrorMsg: "[" + col + "]" + "有効なオプションが存在しません。",
				// 	})
				// }
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
			// 图片字段直接跳过，不更新。
			continue
		}
		if fieldInfo.FieldType == "text" || fieldInfo.FieldType == "textarea" {
			if fieldInfo.IsRequired && len(col) == 0 {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
					ErrorMsg: "このフィールドは必須であり、データを空にすることはできません",
				})
				continue
			}
			//针对契约番号做一个格式check
			if fieldInfo.FieldId == "keiyakuno" {
				if len(col) >= 4 && col[:4] == "auto" {
					checkDataExistError = append(checkDataExistError, &item.Error{
						CurrentLine: line,
						FieldId:     fieldInfo.FieldId,
						ErrorMsg:    "契約番号に「auto」始まりの番号は入力できません",
					})
				}
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
				if field == "rishiritsu" {
					cols[field] = &item.Value{
						DataType: fieldInfo.FieldType,
						Value:    "null",
					}
				} else {
					cols[field] = &item.Value{
						DataType: fieldInfo.FieldType,
						Value:    "0",
					}
				}
			} else {
				numValue, err := cast.ToFloat64E(col)
				if err != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
						ErrorMsg: "[" + col + "]" + "は数値ではありません。",
					})
				} else {
					if numValue < cast.ToFloat64(fieldInfo.MinValue) {
						checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: fieldInfo.FieldId,
							ErrorMsg: "[" + col + "]=>" + "フィールドの大きさが" + cast.ToString(fieldInfo.MinValue) + "以上でなければなりません。",
						})
					} else if numValue > cast.ToFloat64(fieldInfo.MaxValue) {
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
			continue
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

		if cols["residualValue"].GetValue() == "" {
			cols["residualValue"] = &item.Value{
				DataType: "number",
				Value:    "0",
			}
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
			return nil, nil, checkDataExistError
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
			return nil, nil, checkDataExistError
		}

		if response.Total == 0 {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("%v関連アイテムのデータは存在しません。", keys)})
			return nil, nil, checkDataExistError
		}
	}

	var oldItem *item.Item

	if len(cols["id"].GetValue()) > 0 {
		itemID := cols["id"].GetValue()
		itemService := item.NewItemService("database", client.DefaultClient)

		var req item.ItemRequest
		req.DatastoreId = p.datastoreID
		req.ItemId = itemID
		req.Database = p.db
		req.IsOrigin = true
		req.Owners = sessionx.GetAccessKeys(p.db, p.userID, p.datastoreID, "W")

		resp, err := itemService.FindItem(context.TODO(), &req)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "データが存在しないか、データを変更する権限がありません"})
			return nil, nil, checkDataExistError
		}

		oldItem = resp.GetItem()

		// 审批状态check
		status := oldItem.Status
		if nonAdmitCheck(status) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "データはプロセス承認中であり、更新できません。"})
			return nil, nil, checkDataExistError
		}
	}

	rowMap := make(map[string]*item.Value, len(data))

	// 如果是契约登录的场合，关联条件check,计算并将结果登录到临时台账
	if p.action == "contract-insert" {
		itemID := primitive.NewObjectID()
		// 获取契约番号
		keiyakuno := cols["keiyakuno"].GetValue()
		sequenceName := p.datastoreID + "_keiyakuno_auto"
		if keiyakuno == "" {
			num, err := model.GetNextSequenceValue(context.TODO(), p.db, sequenceName)
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{ErrorMsg: err.Error()})
				return nil, nil, checkDataExistError
			}
			keiyakuno = fmt.Sprintf("%s%0*d", "auto_", 10, num)
			cols["keiyakuno"] = &item.Value{
				DataType: "text",
				Value:    keiyakuno,
			}
		}

		// 契约番号唯一性检查(导入文件中的契约番号不可重复)
		if _, exist := keiyakunoMap[keiyakuno]; exist {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("「契約番号」が重複している, 契約番号: %s", keiyakuno)})
		} else {
			keiyakunoMap[keiyakuno] = struct{}{}
		}

		// 关联check1 [「リース開始年月日」は「契約年月日」以降の日付を入力してください。]
		leasestymd := cols["leasestymd"].GetValue()
		keiyakuymd := cols["keiyakuymd"].GetValue()
		if leasestCheck(leasestymd, keiyakuymd) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "「リース開始年月日」は「契約年月日」以降の日付を入力してください。"})
			return nil, nil, checkDataExistError
		}
		// 关联check2 [初回支払はリース開始日以降の日付を入力してください]
		paymentstymd := cols["paymentstymd"].GetValue()
		if paystCheck(paymentstymd, leasestymd) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "「初回支払年月日」は「リース開始年月日」以降の日付を入力してください"})
			return nil, nil, checkDataExistError
		}
		// 关联check3 [残価保証額] 和 [購入オプション行使価額] 只有一个有值
		residualValue := cols["residualValue"].GetValue()
		optionToPurchase := cols["optionToPurchase"].GetValue()
		if optionExclusiveCheck(residualValue, optionToPurchase) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "[残価保証額]と[購入オプション行使価額] 1つだけに価値があります"})
			return nil, nil, checkDataExistError
		}
		// 关联check4 [前払リース料]と[リース・インセンティブ(前払)]の合計を負にすることはできません
		// paymentsAtOrPriorValue := cols["paymentsAtOrPrior"].GetValue()
		// incentivesAtOrPriorValue := cols["incentivesAtOrPrior"].GetValue()
		// if prepaidCheck(paymentsAtOrPriorValue, incentivesAtOrPriorValue) {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "[前払リース料]と[リース・インセンティブ(前払)]の合計を負にすることはできません"})
		// 	return nil, nil, checkDataExistError
		// }
		// 获取支付数据
		paymentst, err := time.Parse("2006-01-02", cols["paymentstymd"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentstymd",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		paymentcycle, err := cast.ToIntE(cols["paymentcycle"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentcycle",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		if paymentcycle < 1 || paymentcycle > 24 {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentcycle",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		paymentday, err := cast.ToIntE(cols["paymentday"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentday",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		if paymentday < 1 || paymentday > 31 {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentday",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		paymentcounts, err := cast.ToIntE(cols["paymentcounts"].GetValue())
		if err != nil {

			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentcounts",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		residual, err := cast.ToFloat64E(cols["residualValue"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "residualValue",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		paymentleasefee, err := cast.ToFloat64E(cols["paymentleasefee"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentleasefee",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		firstleasefee, err := cast.ToFloat64E(cols["firstleasefee"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "firstleasefee",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		finalleasefee, err := cast.ToFloat64E(cols["finalleasefee"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "finalleasefee",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		// 跳过check，以免上传报错
		// optionTo, err := cast.ToFloat64E(cols["optionToPurchase"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "optionToPurchase",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		extentionOption, err := cast.ToIntE(cols["extentionOption"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "extentionOption",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		leasest, err := time.Parse("2006-01-02", cols["leasestymd"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "leasestymd",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		// 跳过check，以免上传报错
		// cancellationrightoption, err := cast.ToBoolE(cols["cancellationrightoption"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "cancellationrightoption",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		// paymentsAtOrPrior, err := cast.ToFloat64E(cols["paymentsAtOrPrior"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentsAtOrPrior",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		// incentivesAtOrPrior, err := cast.ToFloat64E(cols["incentivesAtOrPrior"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "incentivesAtOrPrior",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		initialDirectCosts, err := cast.ToFloat64E(cols["initialDirectCosts"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "initialDirectCosts",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		restorationCosts, err := cast.ToFloat64E(cols["restorationCosts"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "restorationCosts",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }
		leasekikan, err := cast.ToIntE(cols["leasekikan"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "leasekikan",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		torihikikbn := cols["torihikikbn"].GetValue()
		// leasekaishacd := cols["leasekaishacd"].GetValue()
		sykshisankeisan := cols["sykshisankeisan"].GetValue()
		if sykshisankeisan == "1" || sykshisankeisan == "2" {
		} else {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "sykshisankeisan",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		var rishiritsu float64
		if cols["rishiritsu"].GetValue() == "null" {

			leasestymdTime, _ := time.Parse("2006-01-02", cols["leasestymd"].GetValue())
			leasestymdStr := leasestymdTime.Format("Mon Jan 02 2006 15:04:05 MST 0900")

			datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

			var reqDid datastore.DatastoreKeyRequest
			// 从path获取
			reqDid.ApiKey = "ds_rishiritsu"
			reqDid.AppId = p.appID
			reqDid.Database = p.db
			responseDid, err := datastoreService.FindDatastoreByKey(context.TODO(), &reqDid)
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "rishiritsu",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			itemService := item.NewItemService("database", client.DefaultClient)

			var req item.RishiritsuRequest
			req.DatastoreId = responseDid.GetDatastore().DatastoreId
			req.Leasekikan = cols["leasekikan"].GetValue()
			req.Leasestymd = leasestymdStr
			req.Database = p.db

			response, err := itemService.FindRishiritsu(context.TODO(), &req)
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "rishiritsu",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			for key, value := range response.GetItem().GetItems() {
				if key == "rishiritsu" {
					rishiritsu = cast.ToFloat64(value.GetValue())
				}
			}

			cols["rishiritsu"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rishiritsu),
			}

		} else {
			rishiritsu, err = cast.ToFloat64E(cols["rishiritsu"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "rishiritsu",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
		}
		// assetlife, err := cast.ToIntE(cols["assetlife"].GetValue())
		// if err != nil {
		// 	checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "assetlife",
		// 		ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
		// 	})
		// 	return nil, nil, checkDataExistError
		// }

		var payData []Payment

		if value, exist := p.payData[keiyakuno]; exist {
			payData = value
		} else {
			q := PayParam{
				Paymentstymd:     paymentst,
				Paymentcycle:     paymentcycle,
				Paymentday:       paymentday,
				Paymentcounts:    paymentcounts,
				ResidualValue:    residual,
				Paymentleasefee:  paymentleasefee,
				OptionToPurchase: 0, //换成默认值，以免上传报错
				// Leasekaishacd:    leasekaishacd,
				Firstleasefee: firstleasefee,
				Finalleasefee: finalleasefee,
				Keiyakuno:     keiyakuno,
			}
			pay, err := generatePay(q)
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
				return nil, nil, checkDataExistError
			}
			payData = pay
		}

		leaseType := shortOrMinorJudge(p.db, p.appID, p.smallAmount, p.shortPeriod, leasekikan, extentionOption, payData)
		cols["lease_type"] = &item.Value{
			DataType: "options",
			Value:    leaseType,
		}

		expireymd, err := getExpireymd(leasestymd, leasekikan, extentionOption)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
			return nil, nil, checkDataExistError
		}
		cols["leaseexpireymd"] = &item.Value{
			DataType: "date",
			Value:    expireymd,
		}

		// 契约状态
		if !expireCheck(expireymd, p.handleMonth) {
			cols["status"] = &item.Value{
				DataType: "options",
				Value:    "complete",
			}
		} else {
			cols["status"] = &item.Value{
				DataType: "options",
				Value:    "normal",
			}
		}

		if leaseType != "normal_lease" {
			result, err := insertPay(p.db, p.appID, p.userID, p.dsMap, payData)
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			attachItems = append(attachItems, result.attachItems...)
		} else {
			seq, err := uuid.NewUUID()
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			lrp := LRParam{
				ResidualValue:           residual,
				Rishiritsu:              rishiritsu,
				Leasestymd:              leasest,
				CancellationRightOption: true, //换成默认值，以免上传报错
				Leasekikan:              leasekikan,
				ExtentionOption:         extentionOption,
				PaymentsAtOrPrior:       0, //换成默认值，以免上传报错
				IncentivesAtOrPrior:     0, //换成默认值，以免上传报错
				InitialDirectCosts:      initialDirectCosts,
				RestorationCosts:        restorationCosts,
				Assetlife:               0, //换成默认值，以免上传报错
				Torihikikbn:             torihikikbn,
				Payments:                payData,
				DsMap:                   p.dsMap,
				HandleMonth:             p.handleMonth,
				BeginMonth:              p.beginMonth,
				FirstMonth:              p.firstMonth,
				// Leasekaishacd:           leasekaishacd,
				Sykshisankeisan: sykshisankeisan,
				Item: &item.Item{
					ItemId:      itemID.Hex(),
					AppId:       p.appID,
					DatastoreId: p.datastoreID,
					Items:       cols,
				},
				seq: seq.String(),
			}
			result, err := compute(p.db, p.appID, p.userID, lrp)
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			cols["hkkjitenzan"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(result.Hkkjitenzan),
			}
			cols["sonnekigaku"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(result.Sonnekigaku),
			}
			cols["kisyuboka"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(result.KiSyuBoka),
			}
			attachItems = append(attachItems, result.attachItems...)
		}

		cols["id"] = &item.Value{
			DataType: "text",
			Value:    itemID.Hex(),
		}

		// 重新设置kaiyakuymd的值为空,防止误写入
		cols["kaiyakuymd"] = &item.Value{
			DataType: "date",
			Value:    "",
		}

		// 追加处理用字段
		cols["action"] = &item.Value{
			DataType: "text",
			Value:    p.action,
		}

		rowMap = cols
	}
	if p.action == "debt-change" {
		// 获取契约番号
		id := cols["id"].GetValue()
		// 契约番号唯一性检查(导入文件中的契约番号不可重复)
		if _, exist := updateIdMap[id]; exist {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("「ID」が重複している, ID: %s", id)})

			return nil, nil, checkDataExistError
		} else {
			updateIdMap[id] = struct{}{}
		}

		// 契约状态check
		KeiyakuStatus := oldItem.Items["status"].GetValue()
		if keiyakuStatusCheck2(KeiyakuStatus) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "契約状況が[解約済]、[契約満了]の場合、データを更新できません。"})
			return nil, nil, checkDataExistError
		}
		// check [残価保証額] 和 [購入オプション行使価額] 只有一个有值
		if residualValue, exist1 := cols["residualValue"]; exist1 {
			if optionToPurchase, exist2 := cols["optionToPurchase"]; exist2 {
				if optionExclusiveCheck(residualValue.GetValue(), optionToPurchase.GetValue()) {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "[残価保証額]と[購入オプション行使価額] 1つだけに価値があります"})
					return nil, nil, checkDataExistError
				}
			} else {
				if optionExclusiveCheck(residualValue.GetValue(), oldItem.Items["optionToPurchase"].GetValue()) {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "[購入オプション行使価額]にはすでに値があり、[残価保証額]の値を入力できません"})
					return nil, nil, checkDataExistError
				}
			}
		} else {
			if optionToPurchase, exist := cols["optionToPurchase"]; exist {
				if optionExclusiveCheck(oldItem.Items["residualValue"].GetValue(), optionToPurchase.GetValue()) {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "[残価保証額]にはすでに値があり、[購入オプション行使価額]の値を入力できません"})
					return nil, nil, checkDataExistError
				}
			}
		}
		// check 变更年月日
		henkouymd := cols["henkouymd"].GetValue()
		leasestymd := oldItem.Items["leasestymd"].GetValue()
		lastHenkouymd := oldItem.Items["henkouymd"].GetValue()
		if henkouCheck(henkouymd, lastHenkouymd, leasestymd, p.handleMonth, p.beginMonth) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "変更年月日の範囲が正しくありません"})
			return nil, nil, checkDataExistError
		}

		// check 解约年月日
		kaiyakuymd := ""
		if k, kExist := cols["kaiyakuymd"]; kExist {
			kaiyakuymd = k.GetValue()
			if kaiyakuymd != "" {
				leaseexpireymd := oldItem.Items["leaseexpireymd"].GetValue()
				if miraiKaiyakuCheck(kaiyakuymd, leaseexpireymd, p.handleMonth) {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "解约年月日の範囲が正しくありません"})
					return nil, nil, checkDataExistError
				}
			}
		}

		keiyakuno := oldItem.Items["keiyakuno"].GetValue()

		var oldPayData []Payment
		var leaseData []Lease
		var repayData []RePayment

		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "lookup",
			SearchValue:   keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "paymentymd",
			SortValue: "ascend",
		})
		// 偿还表排序
		var ssorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "syokyakuymd",
			SortValue: "ascend",
		})

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		payAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["paymentStatus"], "R")

		// 获取旧的支付结果
		var preq item.ItemsRequest
		preq.ConditionList = conditions
		preq.ConditionType = "and"
		preq.Sorts = sorts
		preq.DatastoreId = p.dsMap["paymentStatus"]
		// 从共通中获取参数
		preq.AppId = p.appID
		preq.Owners = payAccessKeys
		preq.Database = p.db
		preq.IsOrigin = true

		pResp, err := itemService.FindItems(context.TODO(), &preq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}

		for _, it := range pResp.GetItems() {
			paymentcount, err := cast.ToIntE(it.Items["paymentcount"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentcount",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			paymentleasefee, err := cast.ToFloat64E(it.Items["paymentleasefee"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentleasefee",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			paymentleasefeehendo, err := cast.ToFloat64E(it.Items["paymentleasefeehendo"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentleasefeehendo",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			incentives, err := cast.ToFloat64E(it.Items["incentives"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "incentives",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			sonotafee, err := cast.ToFloat64E(it.Items["sonotafee"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "sonotafee",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			kaiyakuson, err := cast.ToFloat64E(it.Items["kaiyakuson"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "kaiyakuson",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			fixed := false
			if it.Items["fixed"].GetValue() != "" {
				fixed, err = cast.ToBoolE(it.Items["fixed"].GetValue())
				if err != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err)})
					return nil, nil, checkDataExistError
				}
			}

			paymentType := it.Items["paymentType"].GetValue()
			paymentymd := it.Items["paymentymd"].GetValue()

			pay := Payment{
				Paymentcount:         paymentcount,
				PaymentType:          paymentType,
				Paymentymd:           paymentymd,
				Paymentleasefee:      paymentleasefee,
				Paymentleasefeehendo: paymentleasefeehendo,
				Incentives:           incentives,
				Sonotafee:            sonotafee,
				Kaiyakuson:           kaiyakuson,
				Fixed:                fixed,
			}

			oldPayData = append(oldPayData, pay)
		}

		interestAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["paymentInterest"], "R")

		// 获取旧的利息结果
		var lreq item.ItemsRequest
		lreq.ConditionList = conditions
		lreq.ConditionType = "and"
		lreq.Sorts = sorts
		lreq.DatastoreId = p.dsMap["paymentInterest"]
		// 从共通中获取参数
		lreq.AppId = p.appID
		lreq.Owners = interestAccessKeys
		lreq.Database = p.db
		lreq.IsOrigin = true

		lResp, err := itemService.FindItems(context.TODO(), &lreq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}

		for _, it := range lResp.GetItems() {
			interest, err := cast.ToFloat64E(it.Items["interest"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "interest",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			repayment, err := cast.ToFloat64E(it.Items["repayment"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "repayment",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			balance, err := cast.ToFloat64E(it.Items["balance"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "balance",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			present, err := cast.ToFloat64E(it.Items["present"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "present",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			paymentymd := it.Items["paymentymd"].GetValue()

			lease := Lease{
				Interest:   interest,
				Repayment:  repayment,
				Balance:    balance,
				Present:    present,
				Paymentymd: paymentymd,
			}

			leaseData = append(leaseData, lease)
		}

		repayAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["repayment"], "R")

		// 获取旧的偿还结果
		var rreq item.ItemsRequest
		rreq.ConditionList = conditions
		rreq.ConditionType = "and"
		rreq.Sorts = ssorts
		// 从path中获取参数
		rreq.DatastoreId = p.dsMap["repayment"]
		// 从共通中获取参数
		rreq.AppId = p.appID
		rreq.Owners = repayAccessKeys
		rreq.Database = p.db
		rreq.IsOrigin = true

		rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}

		for _, it := range rResp.GetItems() {
			endboka, err := cast.ToFloat64E(it.Items["endboka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "endboka",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			boka, err := cast.ToFloat64E(it.Items["boka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "boka",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			syokyaku, err := cast.ToFloat64E(it.Items["syokyaku"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "syokyaku",
					ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			syokyakuymd := it.Items["syokyakuymd"].GetValue()
			syokyakukbn := it.Items["syokyakukbn"].GetValue()

			RePayment := RePayment{
				Endboka:     endboka,
				Boka:        boka,
				Syokyaku:    syokyaku,
				Syokyakuymd: syokyakuymd,
				Syokyakukbn: syokyakukbn,
			}

			repayData = append(repayData, RePayment)
		}

		// 从变更前的数据中获取既定值(不可变更)
		paymentst, err := time.Parse("2006-01-02", oldItem.Items["paymentstymd"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}
		paymentcycle, err := cast.ToIntE(oldItem.Items["paymentcycle"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentcycle",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		paymentday, err := cast.ToIntE(oldItem.Items["paymentday"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentday",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		paymentleasefee, err := cast.ToFloat64E(oldItem.Items["paymentleasefee"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentleasefee",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		torihikikbn := oldItem.Items["torihikikbn"].GetValue()
		assetlife, err := cast.ToIntE(oldItem.Items["assetlife"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "assetlife",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		cancellationrightoption, err := cast.ToBoolE(oldItem.Items["cancellationrightoption"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "cancellationrightoption",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		// 从变更数据中获取，如果获取不到，则从变更前的数据中获取既定值(可变更非必须)
		var residual float64
		if val, ok := cols["residualValue"]; ok {
			num, err := cast.ToFloat64E(val.GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "residualValue",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			residual = num
		} else {
			num, err := cast.ToFloat64E(oldItem.Items["residualValue"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "residualValue",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			residual = num
		}
		var optionTo float64
		if val, ok := cols["optionToPurchase"]; ok {
			num, err := cast.ToFloat64E(val.GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "optionToPurchase",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			optionTo = num
		} else {
			num, err := cast.ToFloat64E(oldItem.Items["optionToPurchase"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "optionToPurchase",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			optionTo = num
		}
		var extentionOption int
		if val, ok := cols["extentionOption"]; ok {
			op, err := cast.ToIntE(val.GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "extentionOption",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			extentionOption = op
		} else {
			op, err := cast.ToIntE(oldItem.Items["extentionOption"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "extentionOption",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			extentionOption = op
		}

		// 从变更数据获取值(可变更必须)
		rishiritsu, err := cast.ToFloat64E(cols["rishiritsu"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "rishiritsu",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		kiSyuBoka, err := cast.ToFloat64E(oldItem.Items["kisyuboka"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
			return nil, nil, checkDataExistError
		}
		percentage, err := cast.ToFloat64E(cols["percentage"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "percentage",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		paymentcounts, err := cast.ToIntE(cols["paymentcounts"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentcounts",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}
		leasekikan, err := cast.ToIntE(cols["leasekikan"].GetValue())
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "leasekikan",
				ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		expireymd, err := getExpireymd(leasestymd, leasekikan, extentionOption)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
			return nil, nil, checkDataExistError
		}
		cols["leaseexpireymd"] = &item.Value{
			DataType: "date",
			Value:    expireymd,
		}

		// 未来解约的情形,参数修正
		if kaiyakuymd != "" && cancellationrightoption {
			// 恢复此情形已经不可变更项目
			rishiritsu, err = cast.ToFloat64E(cols["rishiritsu"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			residual, err = cast.ToFloat64E(oldItem.Items["residualValue"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			extentionOption, err = cast.ToIntE(oldItem.Items["extentionOption"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			optionTo, err = cast.ToFloat64E(oldItem.Items["optionToPurchase"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			paymentcounts, err = cast.ToIntE(cols["paymentcounts"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
			leasekikan, err = cast.ToIntE(cols["leasekikan"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
				return nil, nil, checkDataExistError
			}
		}

		var payData []Payment

		if value, exist := p.payData[keiyakuno]; exist {
			payData = value
		} else {
			q := PayParam{
				Paymentstymd:     paymentst,
				Paymentcycle:     paymentcycle,
				Paymentday:       paymentday,
				Paymentcounts:    paymentcounts,
				ResidualValue:    residual,
				Paymentleasefee:  paymentleasefee,
				OptionToPurchase: optionTo,
			}

			pay, err := generatePay(q)
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
				return nil, nil, checkDataExistError
			}
			payData = pay
		}

		// 只保留可以变更的字段(必须字段)
		rowMap["id"] = cols["id"]                         // ID
		rowMap["index"] = cols["index"]                   // index
		rowMap["henkouymd"] = cols["henkouymd"]           // 変更年月日
		rowMap["leasekikan"] = cols["leasekikan"]         // リース期間
		rowMap["percentage"] = cols["percentage"]         // パーセンテージ
		rowMap["paymentcounts"] = cols["paymentcounts"]   // 支払回数
		rowMap["rishiritsu"] = cols["rishiritsu"]         // 追加借入利子率
		rowMap["leaseexpireymd"] = cols["leaseexpireymd"] // 追加借入利子率
		// 只保留可以变更的字段(非必须字段)
		if val, ok := cols["extentionOption"]; ok {
			// 延長オプション期間
			rowMap["extentionOption"] = val
		}
		if val, ok := cols["residualValue"]; ok {
			// 残価保証額
			rowMap["residualValue"] = val
		}
		if val, ok := cols["optionToPurchase"]; ok {
			// 購入オプション行使価額
			rowMap["optionToPurchase"] = val
		}
		if val, ok := cols["kaiyakuymd"]; ok {
			// 解約年月日
			rowMap["kaiyakuymd"] = val
		}
		seq, err := uuid.NewUUID()
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
			return nil, nil, checkDataExistError
		}

		// 未来解约的情形,参数修正
		if kaiyakuymd != "" {
			// 期间百分比算出
			lstym, err := time.Parse("2006-01", leasestymd[0:7])
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
				return nil, nil, checkDataExistError
			}
			lstyear, lstmonth, _ := lstym.Date()
			kykym, err := time.Parse("2006-01", kaiyakuymd[0:7])
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
				return nil, nil, checkDataExistError
			}
			kykyear, kykmonth, _ := kykym.Date()
			// 确定未来解约时实际确定租赁期间
			realkikan := kykyear*12 + int(kykmonth) - lstyear*12 - int(lstmonth) + 1

			if cancellationrightoption {
				// 确定未来解约时剩余期间百分比
				percentage = float64(realkikan / (leasekikan + extentionOption))
			} else {
				percentage = 1
			}
			// 期间重定
			leasekikan = realkikan

			// 未来解约的情形,参数修正支付数据整理
			var kPayData []Payment
			kaiyakuym, _ := time.Parse("2006-01", kaiyakuymd[0:7])
			for _, pay := range payData {
				paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
				if err != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
					return nil, nil, checkDataExistError
				}
				if paymentym.Equal(kaiyakuym) || paymentym.Before(kaiyakuym) {
					kPayData = append(kPayData, pay)

				}
			}
			payData = kPayData
		}

		lrp := DebtParam{
			Kaiyakuymd:              kaiyakuymd,
			Henkouymd:               henkouymd,
			Leasestymd:              leasestymd,
			Leasekikan:              leasekikan,
			CancellationRightOption: cancellationrightoption,
			ExtentionOption:         extentionOption,
			Keiyakuno:               keiyakuno,
			Rishiritsu:              rishiritsu,
			ResidualValue:           residual,
			Assetlife:               assetlife,
			Torihikikbn:             torihikikbn,
			Percentage:              percentage,
			Payments:                payData,
			DsMap:                   p.dsMap,
			Item:                    oldItem,
			HandleMonth:             p.handleMonth,
			BeginMonth:              p.beginMonth,
			Change:                  rowMap,
			seq:                     seq.String(),
		}
		result, err := debtCompute(p.db, p.appID, p.userID, kiSyuBoka, oldPayData, leaseData, repayData, lrp)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
			return nil, nil, checkDataExistError
		}

		attachItems = append(attachItems, result.attachItems...)

		rowMap["kisyuboka"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(result.KiSyuBoka),
		}

		// 追加处理用字段
		rowMap["action"] = &item.Value{
			DataType: "text",
			Value:    "debt-change",
		}

		// 未来解约导致支付回数
		if kaiyakuymd != "" && cancellationrightoption {
			rowMap["paymentcounts"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(len(payData)),
			}
		}
	}
	if p.action == "midway-cancel" {
		// 获取契约番号
		id := cols["id"].GetValue()
		// 契约番号唯一性检查(导入文件中的契约番号不可重复)
		if _, exist := updateIdMap[id]; exist {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("「ID」が重複している, ID: %s", id)})
		} else {
			updateIdMap[id] = struct{}{}
		}
		// 契约状态check
		KeiyakuStatus := oldItem.Items["status"].GetValue()
		if keiyakuStatusCheck(KeiyakuStatus) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "契約状況が[解約済]、[契約満了]の場合、データを更新できません。"})
			return nil, nil, checkDataExistError
		}
		// check 解约年月日
		kaiyakuymd := cols["kaiyakuymd"].GetValue()
		leasestymd := oldItem.Items["leasestymd"].GetValue()
		lastHenkouymd := oldItem.Items["henkouymd"].GetValue()
		if kaiyakuCheck(kaiyakuymd, lastHenkouymd, leasestymd, p.handleMonth, p.beginMonth) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "解約年月日の範囲が正しくありません"})
			return nil, nil, checkDataExistError
		}

		keiyakuno := oldItem.Items["keiyakuno"].GetValue()

		// ******获取相关台账数据******
		var payData []Payment
		var leaseData []Lease
		var repayData []RePayment

		// 共通检索条件
		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "lookup",
			SearchValue:   keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		// 支付表和利息表排序
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "paymentymd",
			SortValue: "ascend",
		})
		// 偿还表排序
		var ssorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "syokyakuymd",
			SortValue: "ascend",
		})

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		payAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["paymentStatus"], "R")

		// 获取旧的支付结果
		var preq item.ItemsRequest
		preq.ConditionList = conditions
		preq.ConditionType = "and"
		preq.Sorts = sorts
		preq.DatastoreId = p.dsMap["paymentStatus"]
		// 从共通中获取参数
		preq.AppId = p.appID
		preq.Owners = payAccessKeys
		preq.Database = p.db
		preq.IsOrigin = true

		pResp, err := itemService.FindItems(context.TODO(), &preq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}

		for _, it := range pResp.GetItems() {
			paymentcount, err := cast.ToIntE(it.Items["paymentcount"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentcount",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			paymentleasefee, err := cast.ToFloat64E(it.Items["paymentleasefee"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentleasefee",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			paymentleasefeehendo, err := cast.ToFloat64E(it.Items["paymentleasefeehendo"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "paymentleasefeehendo",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			incentives, err := cast.ToFloat64E(it.Items["incentives"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "incentives",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			sonotafee, err := cast.ToFloat64E(it.Items["sonotafee"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "sonotafee",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			kaiyakuson, err := cast.ToFloat64E(it.Items["kaiyakuson"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "kaiyakuson",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			fixed := false
			if it.Items["fixed"].GetValue() != "" {
				fixed, err = cast.ToBoolE(it.Items["fixed"].GetValue())
				if err != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
					return nil, nil, checkDataExistError
				}
			}

			paymentType := it.Items["paymentType"].GetValue()
			paymentymd := it.Items["paymentymd"].GetValue()

			pay := Payment{
				Paymentcount:         paymentcount,
				PaymentType:          paymentType,
				Paymentymd:           paymentymd,
				Paymentleasefee:      paymentleasefee,
				Paymentleasefeehendo: paymentleasefeehendo,
				Incentives:           incentives,
				Sonotafee:            sonotafee,
				Kaiyakuson:           kaiyakuson,
				Fixed:                fixed,
			}

			payData = append(payData, pay)
		}

		interestAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["paymentInterest"], "R")

		// 获取旧的利息结果
		var lreq item.ItemsRequest
		lreq.ConditionList = conditions
		lreq.ConditionType = "and"
		lreq.Sorts = sorts
		lreq.DatastoreId = p.dsMap["paymentInterest"]
		// 从共通中获取参数
		lreq.AppId = p.appID
		lreq.Owners = interestAccessKeys
		lreq.Database = p.db
		lreq.IsOrigin = true

		lResp, err := itemService.FindItems(context.TODO(), &lreq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}

		for _, it := range lResp.GetItems() {
			interest, err := cast.ToFloat64E(it.Items["interest"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "interest",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			repayment, err := cast.ToFloat64E(it.Items["repayment"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "repayment",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			balance, err := cast.ToFloat64E(it.Items["balance"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "balance",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			present, err := cast.ToFloat64E(it.Items["present"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "present",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			paymentymd := it.Items["paymentymd"].GetValue()

			lease := Lease{
				Interest:   interest,
				Repayment:  repayment,
				Balance:    balance,
				Present:    present,
				Paymentymd: paymentymd,
			}

			leaseData = append(leaseData, lease)
		}

		repayAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["repayment"], "R")

		// 获取旧的偿还结果
		var rreq item.ItemsRequest
		rreq.ConditionList = conditions
		rreq.ConditionType = "and"
		rreq.Sorts = ssorts
		// 从path中获取参数
		rreq.DatastoreId = p.dsMap["repayment"]
		// 从共通中获取参数
		rreq.AppId = p.appID
		rreq.Owners = repayAccessKeys
		rreq.Database = p.db
		rreq.IsOrigin = true

		rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}

		for _, it := range rResp.GetItems() {
			endboka, err := cast.ToFloat64E(it.Items["endboka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "endboka",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			boka, err := cast.ToFloat64E(it.Items["boka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "boka",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			syokyaku, err := cast.ToFloat64E(it.Items["syokyaku"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, FieldId: "syokyaku",
					ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			syokyakuymd := it.Items["syokyakuymd"].GetValue()
			syokyakukbn := it.Items["syokyakukbn"].GetValue()

			RePayment := RePayment{
				Endboka:     endboka,
				Boka:        boka,
				Syokyaku:    syokyaku,
				Syokyakuymd: syokyakuymd,
				Syokyakukbn: syokyakukbn,
			}

			repayData = append(repayData, RePayment)
		}

		// 追加契约状态
		cols["status"] = &item.Value{
			DataType: "options",
			Value:    "cancel",
		}

		// 只保留可以变更的字段(必须字段)
		rowMap["id"] = cols["id"]                 // ID
		rowMap["index"] = cols["index"]           // index
		rowMap["status"] = cols["status"]         // 契约状态
		rowMap["kaiyakuymd"] = cols["kaiyakuymd"] // 解約年月日
		// 只保留可以变更的字段(非必须字段)
		if val, ok := cols["kaiyakusongaikin"]; ok {
			// 解約損害金
			rowMap["kaiyakusongaikin"] = val
		}

		seq, err := uuid.NewUUID()
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err)})
			return nil, nil, checkDataExistError
		}

		req := CancelParam{
			Kaiyakuymd:  kaiyakuymd,
			Keiyakuno:   keiyakuno,
			DsMap:       p.dsMap,
			Item:        oldItem,
			Change:      rowMap,
			HandleMonth: p.handleMonth,
			seq:         seq.String(),
		}

		result, err := cancelCompute(p.db, p.appID, p.userID, payData, leaseData, repayData, req)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err)})
			return nil, nil, checkDataExistError
		}

		attachItems = append(attachItems, result.attachItems...)

		// 追加处理用字段
		rowMap["action"] = &item.Value{
			DataType: "text",
			Value:    "midway-cancel",
		}
	}
	if p.action == "contract-expire" {
		// 获取契约番号
		id := cols["id"].GetValue()
		// 契约番号唯一性检查(导入文件中的契约番号不可重复)
		if _, exist := updateIdMap[id]; exist {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: fmt.Sprintf("「ID」が重複している, ID: %s", id)})
		} else {
			updateIdMap[id] = struct{}{}
		}
		// 契约状态check
		KeiyakuStatus := oldItem.Items["status"].GetValue()
		if keiyakuStatusCheck(KeiyakuStatus) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "契約状況が[解約済]、[契約満了]の場合、データを更新できません。"})
			return nil, nil, checkDataExistError
		}
		// 未满了状态check
		Leaseexpireymd := oldItem.Items["leaseexpireymd"].GetValue()
		if expireCheck(Leaseexpireymd, p.handleMonth) {
			checkDataExistError = append(checkDataExistError, &item.Error{CurrentLine: line, ErrorMsg: "データは契約満了な状態に達していません。"})
			return nil, nil, checkDataExistError
		}
		// check 满了年月日
		henkouymd := cols["henkouymd"].GetValue()
		if expireDayCheck(henkouymd, p.handleMonth, Leaseexpireymd) {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    "满了年月日の範囲が正しくありません",
			})
			return nil, nil, checkDataExistError
		}

		keiyakuno := oldItem.Items["keiyakuno"].GetValue()

		// ******获取相关台账数据******
		var repayData []RePayment
		// 共通检索条件
		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "lookup",
			SearchValue:   keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		// 偿还表排序
		var ssorts []*item.SortItem
		ssorts = append(ssorts, &item.SortItem{
			SortKey:   "syokyakuymd",
			SortValue: "ascend",
		})

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		repayAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["repayment"], "R")

		// 获取旧的偿还结果
		var rreq item.ItemsRequest
		rreq.ConditionList = conditions
		rreq.ConditionType = "and"
		rreq.Sorts = ssorts
		// 从path中获取参数
		rreq.DatastoreId = p.dsMap["repayment"]
		// 从共通中获取参数
		rreq.AppId = p.appID
		rreq.Owners = repayAccessKeys
		rreq.Database = p.db
		rreq.IsOrigin = true

		rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		for _, it := range rResp.GetItems() {
			endboka, err := cast.ToFloat64E(it.Items["endboka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "endboka",
					ErrorMsg:    fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			boka, err := cast.ToFloat64E(it.Items["boka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "boka",
					ErrorMsg:    fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			syokyaku, err := cast.ToFloat64E(it.Items["syokyaku"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "syokyaku",
					ErrorMsg:    fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			syokyakuymd := it.Items["syokyakuymd"].GetValue()
			syokyakukbn := it.Items["syokyakukbn"].GetValue()

			RePayment := RePayment{
				Endboka:     endboka,
				Boka:        boka,
				Syokyaku:    syokyaku,
				Syokyakuymd: syokyakuymd,
				Syokyakukbn: syokyakukbn,
			}

			repayData = append(repayData, RePayment)
		}
		// 契约状态
		cols["status"] = &item.Value{
			DataType: "options",
			Value:    "complete",
		}

		// リース満了償却区分
		expiresyokyakukbn := cols["expiresyokyakukbn"].GetValue()
		// 取引判定区分
		torihikikbn := oldItem.Items["torihikikbn"].GetValue()

		// 只保留可以变更的字段(必须字段)
		rowMap["id"] = cols["id"]                               // ID
		rowMap["index"] = cols["index"]                         // index
		rowMap["henkouymd"] = cols["henkouymd"]                 // 変更年月日
		rowMap["status"] = cols["status"]                       // 契约状态
		rowMap["expiresyokyakukbn"] = cols["expiresyokyakukbn"] // リース満了償却区分

		seq, err := uuid.NewUUID()
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		req := ExpireParam{
			Henkouymd:         henkouymd,
			Keiyakuno:         keiyakuno,
			Torihikikbn:       torihikikbn,
			Expiresyokyakukbn: expiresyokyakukbn,
			DsMap:             p.dsMap,
			Item:              oldItem,
			HandleMonth:       p.handleMonth,
			Change:            rowMap,
			seq:               seq.String(),
		}

		result, err := expireCompute(p.db, p.appID, p.userID, repayData, req)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		attachItems = append(attachItems, result.attachItems...)

		// 追加处理用字段
		if result.Leftgaku > 0 {
			rowMap["hasChange"] = &item.Value{
				DataType: "text",
				Value:    "1",
			}
		} else {
			rowMap["hasChange"] = &item.Value{
				DataType: "text",
				Value:    "0",
			}
		}

		rowMap["action"] = &item.Value{
			DataType: "text",
			Value:    "contract-expire",
		}
	}
	if p.action == "info-change" {
		// 获取契约番号
		id := cols["id"].GetValue()
		// 契约番号唯一性检查(导入文件中的契约番号不可重复)
		if _, exist := updateIdMap[id]; exist {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("「ID」が重複している, ID: %s", id),
			})
		} else {
			updateIdMap[id] = struct{}{}
		}
		// 契约状态check
		KeiyakuStatus := oldItem.Items["status"].GetValue()
		if keiyakuStatusCheck2(KeiyakuStatus) {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    "契約状況が[解約済]、[契約満了]の場合、データを更新できません。",
			})
			return nil, nil, checkDataExistError
		}
		// check 变更年月日
		henkouymdValue := cols["henkouymd"]
		leasestymdValue := oldItem.Items["leasestymd"]
		lastHenkouymd := oldItem.Items["henkouymd"].GetValue()
		if henkouCheck(henkouymdValue.GetValue(), lastHenkouymd, leasestymdValue.GetValue(), p.handleMonth, p.beginMonth) {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    "変更年月日の範囲が正しくありません",
			})
			return nil, nil, checkDataExistError
		}

		keiyakuno := oldItem.Items["keiyakuno"].GetValue()

		// ******获取相关台账数据******
		var payData []Payment
		var leaseData []Lease
		var repayData []RePayment

		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "keiyakuno",
			FieldType:     "lookup",
			SearchValue:   keiyakuno,
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "paymentymd",
			SortValue: "ascend",
		})
		// 偿还表排序
		var ssorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "syokyakuymd",
			SortValue: "ascend",
		})

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		payAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["paymentStatus"], "R")

		// 获取旧的支付结果
		var preq item.ItemsRequest
		preq.ConditionList = conditions
		preq.ConditionType = "and"
		preq.Sorts = sorts
		preq.DatastoreId = p.dsMap["paymentStatus"]
		// 从共通中获取参数
		preq.AppId = p.appID
		preq.Owners = payAccessKeys
		preq.Database = p.db
		preq.IsOrigin = true

		pResp, err := itemService.FindItems(context.TODO(), &preq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		for _, it := range pResp.GetItems() {
			paymentcount, err := cast.ToIntE(it.Items["paymentcount"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "paymentcount",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			paymentleasefee, err := cast.ToFloat64E(it.Items["paymentleasefee"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "paymentleasefee",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			paymentleasefeehendo, err := cast.ToFloat64E(it.Items["paymentleasefeehendo"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "paymentleasefeehendo",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			incentives, err := cast.ToFloat64E(it.Items["incentives"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "incentives",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			sonotafee, err := cast.ToFloat64E(it.Items["sonotafee"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "sonotafee",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			kaiyakuson, err := cast.ToFloat64E(it.Items["kaiyakuson"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "kaiyakuson",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			fixed := false
			if it.Items["fixed"].GetValue() != "" {
				fixed, err = cast.ToBoolE(it.Items["fixed"].GetValue())
				if err != nil {
					checkDataExistError = append(checkDataExistError, &item.Error{
						CurrentLine: line,
						ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
					})
					return nil, nil, checkDataExistError
				}
			}

			paymentType := it.Items["paymentType"].GetValue()
			paymentymd := it.Items["paymentymd"].GetValue()

			pay := Payment{
				Paymentcount:         paymentcount,
				PaymentType:          paymentType,
				Paymentymd:           paymentymd,
				Paymentleasefee:      paymentleasefee,
				Paymentleasefeehendo: paymentleasefeehendo,
				Incentives:           incentives,
				Sonotafee:            sonotafee,
				Kaiyakuson:           kaiyakuson,
				Fixed:                fixed,
			}

			payData = append(payData, pay)
		}

		interestAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["paymentInterest"], "R")

		// 获取旧的利息结果
		var lreq item.ItemsRequest
		lreq.ConditionList = conditions
		lreq.ConditionType = "and"
		lreq.Sorts = sorts
		lreq.DatastoreId = p.dsMap["paymentInterest"]
		// 从共通中获取参数
		lreq.AppId = p.appID
		lreq.Owners = interestAccessKeys
		lreq.Database = p.db
		lreq.IsOrigin = true

		lResp, err := itemService.FindItems(context.TODO(), &lreq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		for _, it := range lResp.GetItems() {
			interest, err := cast.ToFloat64E(it.Items["interest"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "interest",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			repayment, err := cast.ToFloat64E(it.Items["repayment"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "repayment",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			balance, err := cast.ToFloat64E(it.Items["balance"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "balance",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			present, err := cast.ToFloat64E(it.Items["present"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "present",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				continue
			}

			paymentymd := it.Items["paymentymd"].GetValue()

			lease := Lease{
				Interest:   interest,
				Repayment:  repayment,
				Balance:    balance,
				Present:    present,
				Paymentymd: paymentymd,
			}

			leaseData = append(leaseData, lease)
		}

		repayAccessKeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["repayment"], "R")

		// 获取旧的偿还结果
		var rreq item.ItemsRequest
		rreq.ConditionList = conditions
		rreq.ConditionType = "and"
		rreq.Sorts = ssorts
		// 从path中获取参数
		rreq.DatastoreId = p.dsMap["repayment"]
		// 从共通中获取参数
		rreq.AppId = p.appID
		rreq.Owners = repayAccessKeys
		rreq.Database = p.db
		rreq.IsOrigin = true

		rResp, err := itemService.FindItems(context.TODO(), &rreq, opss)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		for _, it := range rResp.GetItems() {
			endboka, err := cast.ToFloat64E(it.Items["endboka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "endboka",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			boka, err := cast.ToFloat64E(it.Items["boka"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "boka",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}
			syokyaku, err := cast.ToFloat64E(it.Items["syokyaku"].GetValue())
			if err != nil {
				checkDataExistError = append(checkDataExistError, &item.Error{
					CurrentLine: line,
					FieldId:     "syokyaku",
					ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容: %v", err),
				})
				return nil, nil, checkDataExistError
			}

			syokyakuymd := it.Items["syokyakuymd"].GetValue()
			syokyakukbn := it.Items["syokyakukbn"].GetValue()

			RePayment := RePayment{
				Endboka:     endboka,
				Boka:        boka,
				Syokyaku:    syokyaku,
				Syokyakuymd: syokyakuymd,
				Syokyakukbn: syokyakukbn,
			}

			repayData = append(repayData, RePayment)
		}

		// 删除不可以变更的字段
		delete(cols, "keiyakuno") // 契約番号
		// delete(cols, "leasekaishacd")           // リース会社
		delete(cols, "keiyakuymd")              // 契約年月日
		delete(cols, "leasestymd")              // リース開始日
		delete(cols, "leasekikan")              // リース期間
		delete(cols, "extentionOption")         // 延長リース期間
		delete(cols, "assetlife")               // 耐用年数
		delete(cols, "torihikikbn")             //  取引判定区分
		delete(cols, "paymentstymd")            // 初回支払日
		delete(cols, "paymentcycle")            // 支払サイクル
		delete(cols, "paymentday")              // 支払日
		delete(cols, "paymentcounts")           // 支払回数
		delete(cols, "paymentleasefee")         // 支払リース料
		delete(cols, "residualValue")           // 残価保証額
		delete(cols, "rishiritsu")              // 追加借入利率
		delete(cols, "paymentsAtOrPrior")       // 前払リース料
		delete(cols, "incentivesAtOrPrior")     // リース・インセンティブ（前払）
		delete(cols, "initialDirectCosts")      // 当初直接費用
		delete(cols, "restorationCosts")        // 原状回復コスト
		delete(cols, "usecancellationoption")   // 解約オプション行使
		delete(cols, "kaiyakuymd")              // 解約年月日
		delete(cols, "status")                  // 契約状態
		delete(cols, "expiresyokyakukbn")       // リース満了償却区分
		delete(cols, "leaseexpireymd")          // リース満了年月日
		delete(cols, "lease_type")              // リースタイプ
		delete(cols, "kaiyakusongaikin")        // 解約損害金
		delete(cols, "cancellationrightoption") // 解約行使権オプション
		delete(cols, "percentage")              // パーセンテージ

		rowMap = cols

		seq, err := uuid.NewUUID()
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("エラーが発生した計算結果で発生し、エラーの内容：: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		req := ChangeParam{
			Henkouymd:   henkouymdValue.GetValue(),
			DsMap:       p.dsMap,
			Item:        oldItem,
			Change:      rowMap,
			HandleMonth: p.handleMonth,
			seq:         seq.String(),
		}

		result, err := changeCompute(p.db, p.appID, p.userID, payData, leaseData, repayData, req)
		if err != nil {
			checkDataExistError = append(checkDataExistError, &item.Error{
				CurrentLine: line,
				ErrorMsg:    fmt.Sprintf("支払いテーブルの生成にエラーがあり、エラーの内容: %v", err),
			})
			return nil, nil, checkDataExistError
		}

		attachItems = append(attachItems, result.attachItems...)

		// 追加处理用字段
		rowMap["action"] = &item.Value{
			DataType: "text",
			Value:    "info-change",
		}
	}

	list = &item.ListItems{
		Items: rowMap,
	}

	return list, attachItems, checkDataExistError
}
