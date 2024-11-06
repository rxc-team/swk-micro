package inventory

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

	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/spf13/cast"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/import/common/filex"
	"rxcsoft.cn/pit3/srv/import/common/langx"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/model"
	"rxcsoft.cn/pit3/srv/import/system/wsx"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// Import 读取上传文件批量盘点
func Import(base Params, filePath string) {
	// 获取传入变量
	jobID := base.JobId
	encoding := base.Encoding
	userID := base.UserId
	owners := base.Owners
	currentAppID := base.AppId
	roles := base.Roles
	lang := base.Lang
	domain := base.Domain
	datastoreID := base.DatastoreId
	db := base.Database
	groupID := base.GroupId
	appID := base.AppId
	checkType := base.CheckType
	checkedAt := base.CheckedAt
	checkedBy := base.CheckedBy
	updateOwners := base.UpdateOwners

	mainKeys := make(map[string]struct{})
	for _, v := range base.MainKeys {
		mainKeys[v] = struct{}{}
	}

	// 发送消息 开始读取数据
	model.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "アップロードされたファイルを取得します",
		CurrentStep: "get-file",
		Database:    db,
	}, userID)

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
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
		if len(filePath) > 0 {
			os.Remove(filePath)
			minioClient.DeleteObject(filePath)
		}
		// 最后删除public文件夹
		os.Remove("public/app_" + appID)
	}()

	// 发送消息 开始读取数据
	model.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     i18n.Tr(lang, "job.J_001"),
		CurrentStep: "read-file",
		Database:    db,
	}, userID)

	// 从minio获取文件到本地临时文件夹备用
	err = model.GetFile(domain, appID, filePath)
	if err != nil {
		loggerx.ErrorLog("ReadCheckCsvFileAndImport", err.Error())
		// 編輯錯誤日誌文件
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_016"),
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
	allFields := model.GetFields(db, datastoreID, currentAppID, roles, true)
	// 获取用户信息
	allUsers := model.GetUsers(db, currentAppID, domain)

	// 获取数据上传流
	itemService := item.NewItemService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
	}

	stream, err := itemService.ImportCheckItem(context.Background(), opss)
	if err != nil {
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
	err = stream.Send(&item.ImportCheckRequest{
		Status: item.SendStatus_SECTION,
		Request: &item.ImportCheckRequest_Meta{
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

	// 读取上传文件内容,获取头部和数据部(2000一组)和数据合计
	param := checkReadParam{
		db:          db,
		appID:       appID,
		domain:      domain,
		datastoreID: datastoreID,
		jobID:       jobID,
		userID:      userID,
		groupID:     groupID,
		encoding:    encoding,
		lang:        lang,
		allFields:   allFields,
		allUsers:    allUsers,
		langData:    langData,
		mainKeys:    mainKeys,
		checkType:   checkType,
		checkedAt:   checkedAt,
		checkedBy:   checkedBy,
	}

	// excel文件判断,若是则转换为csv
	if filePath[strings.LastIndex(filePath, ".")+1:] != "csv" {
		param.encoding = "utf-8"
		filePath, err = model.ExcelToCsv(filePath)
		if err != nil {
			loggerx.ErrorLog("ReadCheckCsvFileAndImport", err.Error())
			// 編輯錯誤日誌文件
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "EXCELファイルの読み取りに失敗しました。",
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

	items, err := readCheckFile(param, filePath)
	if err != nil {
		loggerx.ErrorLog("ReadCheckCsvFileAndImport", err.Error())
		// 編輯錯誤日誌文件
		path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
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

	// 如果执行成功
	var errorList []string
	total := int64(len(items))
	var inserted int64 = 0
	var updated int64 = 0

	// 验证数据
	go func() {
		// 开始导入
		for _, data := range items {
			err := stream.Send(&item.ImportCheckRequest{
				Status: item.SendStatus_SECTION,
				Request: &item.ImportCheckRequest_Data{
					Data: data,
				},
			})
			if err == io.EOF {
				return
			}
			if err != nil {
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

		err := stream.Send(&item.ImportCheckRequest{
			Status: item.SendStatus_COMPLETE,
			Request: &item.ImportCheckRequest_Data{
				Data: nil,
			},
		})

		if err != nil {
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
}

// readCheckFile 读取盘点文件
func readCheckFile(p checkReadParam, filePath string) (data []*item.ChangeData, e error) {

	var errorList []string

	// 打开文件
	fs, err := os.Open(filePath)
	if err != nil {
		loggerx.ErrorLog("readCheckCsvFileAndImport", err.Error())
		path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})

		// 发送消息 数据验证错误，停止上传
		model.ModifyTask(task.ModifyRequest{
			JobId:       p.jobID,
			Message:     "ファイルの読み取りに失敗しました。",
			CurrentStep: "read-file",
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

	// 编码转换
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
		// 读取一行
		row, err := r.Read()
		// 读取错误
		if err != nil && err != io.EOF {
			loggerx.ErrorLog("readCheckCsvFileAndImport", err.Error())

			path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})

			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       p.jobID,
				Message:     "ファイルの読み取りに失敗しました。",
				CurrentStep: "read-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: p.db,
			}, p.userID)

			return data, err
		}
		// 读到末尾
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
		// 头部行数据
		if index == 0 {
			hasEmpty := false
			for _, h := range row {
				if h == "" || h == "　" {
					hasEmpty = true
				}
			}

			// 头部列空白判断
			if hasEmpty {
				path := filex.WriteAndSaveFile(p.domain, p.appID, []string{"csvヘッダー行に空白の列名があります。修正してください。"})

				// 发送消息 数据验证错误，停止上传
				model.ModifyTask(task.ModifyRequest{
					JobId:       p.jobID,
					Message:     "csvファイルの形式が正しくありません。",
					CurrentStep: "read-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: p.db,
				}, p.userID)

				return data, errors.New("csvヘッダー行に空白の列名があります。修正してください。")
			}

			row[0] = strings.Replace(row[0], "\ufeff", "", 1)
			for _, h := range row {
				hasExit := false
				// 获取对应标题的key
				for _, f := range p.allFields {
					name := langx.GetLangValue(p.langData, f.FieldName, "")

					if name == h {

						_, ok := p.mainKeys[f.GetFieldId()]
						if !ok {
							return data, fmt.Errorf(fmt.Sprintf("[%s]フィールドがメインキー設定にありません", h))
						}

						p.headerData = append(p.headerData, f)
						hasExit = true
					}

				}

				if !hasExit {
					return data, fmt.Errorf(fmt.Sprintf("[%s]このフィールドが見つかりません", h))
				}
			}

			index++
			continue
		}
		// 获取盘点开始日付
		isUpdate := true
		cfg, err := model.GetConfig(p.db, p.appID)
		if err != nil {
			loggerx.ErrorLog("readCheckCsvFileAndImport", err.Error())

			path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})

			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       p.jobID,
				Message:     "ファイルの読み取りに失敗しました。",
				CurrentStep: "read-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: p.db,
			}, p.userID)

			return data, err
		}
		checkStartDate, err := time.ParseInLocation("2006-01-02", cfg.GetCheckStartDate(), time.Local)
		if err != nil {
			loggerx.ErrorLog("readCheckCsvFileAndImport", err.Error())

			path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})

			// 发送消息 数据验证错误，停止上传
			model.ModifyTask(task.ModifyRequest{
				JobId:       p.jobID,
				Message:     "ファイルの読み取りに失敗しました。",
				CurrentStep: "read-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: p.db,
			}, p.userID)

			return data, err
		}
		nowTime := time.Now()
		if checkStartDate.After(nowTime) {
			isUpdate = false
		}

		// 数据部数据读取
		if index > 0 {
			// 设置值
			p.fileData = row
			// 设置行数
			p.index = index
			// 数据判断并作成
			it, checkErrors := checkCheckAndBuildItems(p, isUpdate)
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

			data = append(data, it)

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

// checkCheckAndBuildItems 生成台账数据
func checkCheckAndBuildItems(p checkReadParam, isUpdate bool) (list *item.ChangeData, result []*item.Error) {

	headers := p.headerData
	data := p.fileData

	var checkDataExistError []*item.Error

	currentLine := int64(p.index)
	query := make(map[string]*item.Value)
	change := make(map[string]*item.Value)
	for i, col := range data {
		field := headers[i]

		// 字段数据检查(常)
		v, e := requiredCheck(col, field, currentLine)
		if e != nil {
			checkDataExistError = append(checkDataExistError, e)
			continue
		}
		// 更新条件编辑
		fmt.Print(v)
		query[field.GetFieldId()] = v
	}

	// 默认盘点情报编辑
	change["check_type"] = &item.Value{
		DataType: "text",
		Value:    p.checkType,
	}
	change["checked_at"] = &item.Value{
		DataType: "date",
		Value:    p.checkedAt,
	}
	change["checked_by"] = &item.Value{
		DataType: "text",
		Value:    p.checkedBy,
	}
	// 固定盘点情报编辑
	if isUpdate {
		change["check_status"] = &item.Value{
			DataType: "text",
			Value:    "1",
		}
	} else {
		change["check_status"] = &item.Value{
			DataType: "text",
			Value:    "0",
		}
	}

	list = &item.ChangeData{
		Query:  query,
		Change: change,
		Index:  currentLine,
	}

	return list, checkDataExistError
}

// requiredCheck 必须检查
func requiredCheck(col string, fieldInfo *field.Field, currentLine int64) (*item.Value, *item.Error) {

	// 字段类型必须是text和autonum
	if fieldInfo.FieldType == "text" || fieldInfo.FieldType == "autonum" {
		if len(col) == 0 {
			return nil, &item.Error{
				CurrentLine: currentLine,
				FieldId:     fieldInfo.FieldId,
				ErrorMsg:    "このフィールドは必須であり、データを空にすることはできません",
			}
		}

		return &item.Value{
			DataType: fieldInfo.FieldType,
			Value:    col,
		}, nil

	}

	return nil, &item.Error{
		CurrentLine: currentLine,
		FieldId:     fieldInfo.FieldId,
		ErrorMsg:    "この現在のフィールドのタイプが設定と一致しません",
	}
}
