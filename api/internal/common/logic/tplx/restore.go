package tplx

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/aclx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/srv/database/proto/copy"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/database/proto/print"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/allow"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/level"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
	"rxcsoft.cn/pit3/srv/report/proto/report"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// restore 从app模板创建新app
func Restore(db, userID, jobID, roleID, domain, groupID, backupID, currentAppID, lang string, accessKeys []string) error {

	timestamp := time.Now().Format("20060102150405")

	go func() {

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_027"),
			CurrentStep: "get-template-file",
			Database:    db,
		}, userID)

		backupService := backup.NewBackupService("manage", client.DefaultClient)

		var req backup.FindBackupRequest
		req.BackupId = backupID
		req.Database = "system"
		response, err := backupService.FindBackup(context.TODO(), &req)
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "get-template-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_035"),
			CurrentStep: "read-file",
			Database:    db,
		}, userID)

		// 获取文件
		dir := "backups/"
		// 创建文件夹
		err = filex.Mkdir(dir)
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
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
		superDomain := sessionx.GetSuperDomain()
		minioClient, err := storagecli.NewClient(superDomain)
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
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
		object, err := minioClient.GetObject(response.GetBackup().GetFileName())
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
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
		var result []byte

		buffer := make([]byte, 1024)
		for {
			n, err := object.Read(buffer)
			result = append(result, buffer[:n]...)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					loggerx.ErrorLog("restore", err.Error())
					path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
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
			}
		}

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_044"),
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		name := dir + timestamp + ".zip"
		err = filex.SaveLocalFile(result, name)
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
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

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_048"),
			CurrentStep: "unzip-file",
			Database:    db,
		}, userID)

		// 解压文件
		zipDir, err := filex.UnZip(name, "", "utf-8")
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "unzip-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		err = os.Remove(name)
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "unzip-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_039"),
			CurrentStep: "restore",
			Database:    db,
		}, userID)

		// 读取app数据
		var appData app.App
		err = filex.ReadFile(zipDir+"/apps.json", &appData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		appID := appData.GetAppId()

		// 读取语言数据
		var langData []*language.Language
		err = filex.ReadFile(zipDir+"/languages.json", &langData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		// 读取台账数据
		var opData []*option.Option
		err = filex.ReadFile(zipDir+"/options.json", &opData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		// 恢复选项数据
		maxOptionID := "000000"
		for _, op := range opData {
			optionService := option.NewOptionService("database", client.DefaultClient)

			var req option.AddRequest
			req.OptionId = op.GetOptionId()
			req.OptionMemo = op.GetOptionMemo()
			req.OptionOrder = op.GetOptionOrder()
			req.OptionValue = op.GetOptionValue()
			req.AppId = currentAppID
			req.Writer = userID
			req.Database = db

			// 设置最大选项组ID值
			if op.GetOptionId() > maxOptionID {
				maxOptionID = op.GetOptionId()
			}

			_, err := optionService.AddOption(context.TODO(), &req, opss)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}

			// 恢复对应的语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[appID]
				if exist {
					name, ok := appLang.Options[op.GetOptionId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   domain,
							LangCd:   lang.LangCd,
							AppId:    currentAppID,
							Type:     "options",
							Key:      op.GetOptionId(),
							Value:    name,
							Writer:   userID,
							Database: db,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("restore", err.Error())
							path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "restore",
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
					labelName, ok := appLang.Options[op.GetOptionId()+"_"+op.GetOptionValue()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   domain,
							LangCd:   lang.LangCd,
							AppId:    currentAppID,
							Type:     "options",
							Key:      op.GetOptionId() + "_" + op.GetOptionValue(),
							Value:    labelName,
							Writer:   userID,
							Database: db,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("restore", err.Error())
							path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "restore",
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
				}
			}
		}
		if len(opData) > 0 {
			// 设置序列值
			err = setOptionSequenceValue(db, currentAppID, maxOptionID)
			if err != nil {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_038"),
			CurrentStep: "restore",
			Database:    db,
		}, userID)

		// 读取台账数据
		var dsData []*datastore.Datastore
		err = filex.ReadFile(zipDir+"/data_stores.json", &dsData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		// 读取権限数据
		var peData []*permission.Action
		err = filex.ReadFile(zipDir+"/permissions.json", &peData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		// 读取台账打印数据
		var prsData []*print.Print
		err = filex.ReadFile(zipDir+"/prints.json", &prsData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		// 读取字段数据
		var fsData []*field.Field
		err = filex.ReadFile(zipDir+"/fields.json", &fsData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

		// 台账复制前后映射
		dsMap := map[string]string{}
		dsActionsMap := map[string]map[string]bool{}
		// 台账角色权限点
		var dsActions []*role.Action
		// 角色许可
		var pmList []*role.Permission

		// 角色更新参数
		var roleReq role.ModifyRoleRequest
		roleReq.RoleId = roleID
		roleReq.Database = db

		printService := print.NewPrintService("database", client.DefaultClient)

		// 恢复台账情报
		for _, ds := range dsData {
			var req datastore.AddRequest
			req.AppId = currentAppID
			req.ApiKey = ds.GetApiKey()
			req.CanCheck = ds.GetCanCheck()
			req.ShowInMenu = ds.GetShowInMenu()
			req.NoStatus = ds.GetNoStatus()
			req.Encoding = ds.GetEncoding()
			req.Writer = userID
			req.Database = db
			// 恢复台账默认排序
			req.Sorts = ds.GetSorts()
			req.DisplayOrder = ds.DisplayOrder
			// 恢复台账扫描设置
			req.ScanFieldsConnector = ds.GetScanFieldsConnector()
			req.ScanFields = ds.GetScanFields()
			// 恢复台账二维码打印字段设置
			req.PrintField1 = ds.GetPrintField1()
			req.PrintField2 = ds.GetPrintField2()
			req.PrintField3 = ds.GetPrintField3()
			req.UniqueFields = ds.GetUniqueFields()

			dsRes, err := datastoreService.AddDatastore(context.TODO(), &req, opss)
			if err != nil {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}

			// 台账复制前后映射编辑
			dsMap[ds.GetDatastoreId()] = dsRes.GetDatastoreId()
			for _, pe := range peData {
				if pe.ObjectId == ds.GetDatastoreId() {
					dsActionsMap[dsRes.GetDatastoreId()] = setPermission(db, "datastore", ds.GetApiKey(), ds.GetCanCheck(), pe.ActionMap)
				}
			}

			// 恢复台账名称多语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[appID]
				if exist {
					name, ok := appLang.Datastores[ds.GetDatastoreId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   domain,
							LangCd:   lang.LangCd,
							AppId:    currentAppID,
							Type:     "datastores",
							Key:      dsRes.GetDatastoreId(),
							Value:    name,
							Writer:   userID,
							Database: db,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("restore", err.Error())
							path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "restore",
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
				}
			}

			// 恢复台账映射关系
			var reqMap datastore.AddMappingRequest
			// 映射所属情报
			reqMap.AppId = currentAppID
			reqMap.Database = db
			reqMap.DatastoreId = dsRes.GetDatastoreId()
			// 恢复映射情报
			for _, mapInfo := range ds.GetMappings() {
				reqMap.MappingType = mapInfo.MappingType
				reqMap.UpdateType = mapInfo.UpdateType
				reqMap.SeparatorChar = mapInfo.SeparatorChar
				reqMap.BreakChar = mapInfo.BreakChar
				reqMap.LineBreakCode = mapInfo.LineBreakCode
				reqMap.CharEncoding = mapInfo.CharEncoding
				reqMap.MappingRule = mapInfo.MappingRule
				reqMap.ApplyType = mapInfo.ApplyType
				res, err := datastoreService.AddDatastoreMapping(context.TODO(), &reqMap)
				if err != nil {
					loggerx.ErrorLog("restore", err.Error())
					path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)
					return
				}

				// 恢复映射对应的多语言
				for _, lang := range langData {
					appLang, exist := lang.GetApps()[appID]
					if exist {
						name, ok := appLang.Mappings[ds.GetDatastoreId()+"_"+mapInfo.GetMappingId()]
						if ok {
							langReq := language.AddAppLanguageDataRequest{
								Domain:   domain,
								LangCd:   lang.LangCd,
								AppId:    currentAppID,
								Type:     "mappings",
								Key:      reqMap.DatastoreId + "_" + res.MappingId,
								Value:    name,
								Writer:   userID,
								Database: db,
							}
							err := addAppLangData(langReq)
							if err != nil {
								loggerx.ErrorLog("restore", err.Error())
								path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
								// 发送消息 获取数据失败，终止任务
								jobx.ModifyTask(task.ModifyRequest{
									JobId:       jobID,
									Message:     err.Error(),
									CurrentStep: "restore",
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
					}
				}
			}

			// 恢复台账打印设置
			for i := 0; i < len(prsData); i++ {
				if prsData[i].AppId == appID && prsData[i].DatastoreId == ds.GetDatastoreId() {
					rp := prsData[i]
					var req print.AddPrintRequest
					req.AppId = currentAppID
					req.Writer = rp.GetCreatedBy()
					req.Database = db
					req.DatastoreId = dsRes.GetDatastoreId()
					req.Page = rp.Page
					req.Orientation = rp.Orientation
					req.TitleWidth = rp.TitleWidth
					req.ShowSign = rp.ShowSign
					req.SignName1 = rp.SignName1
					req.SignName2 = rp.SignName2
					req.ShowSystem = rp.ShowSystem
					req.CheckField = rp.CheckField
					req.Fields = rp.Fields
					_, err := printService.AddPrint(context.TODO(), &req)
					if err != nil {
						loggerx.ErrorLog("restore", err.Error())
						path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     err.Error(),
							CurrentStep: "restore",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: db,
						}, userID)
						return
					}
					break
				}
			}
		}

		// 恢复台账其他信息
		for _, ds := range dsData {
			currentDatastoreID := dsMap[ds.DatastoreId]
			var fields []string
			for _, pe := range peData {
				if pe.ObjectId == ds.DatastoreId {
					fields = append(fields, pe.Fields...)
				}
			}

			// 恢复台账对应的字段
			for _, f := range fsData {
				/* var fields []string */

				if f.GetDatastoreId() == ds.DatastoreId {
					fieldService := field.NewFieldService("database", client.DefaultClient)
					req := field.AddRequest{
						AppId:         currentAppID,
						DatastoreId:   currentDatastoreID,
						FieldType:     f.GetFieldType(),
						FieldId:       f.GetFieldId(),
						IsRequired:    f.GetIsRequired(),
						IsImage:       f.GetIsImage(),
						AsTitle:       f.GetAsTitle(),
						Unique:        f.GetUnique(),
						LookupAppId:   currentAppID,
						LookupFieldId: f.GetLookupFieldId(),
						UserGroupId:   f.GetUserGroupId(),
						OptionId:      f.GetOptionId(),
						MinLength:     f.GetMinLength(),
						MaxLength:     f.GetMaxLength(),
						MinValue:      f.GetMinValue(),
						MaxValue:      f.GetMaxValue(),
						Width:         f.GetWidth(),
						Cols:          f.GetCols(),
						Rows:          f.GetRows(),
						X:             f.GetX(),
						Y:             f.GetY(),
						DisplayOrder:  f.GetDisplayOrder(),
						DisplayDigits: f.GetDisplayDigits(),
						Prefix:        f.GetPrefix(),
						ReturnType:    f.GetReturnType(),
						Formula:       f.GetFormula(),
						IsCheckImage:  f.GetIsCheckImage(),
						Precision:     f.GetPrecision(),
						IsFixed:       f.GetIsFixed(),
						Writer:        userID,
						Database:      db,
					}
					if d, exist := dsMap[f.GetLookupDatastoreId()]; exist {
						req.LookupDatastoreId = d
					}
					if len(f.GetUserGroupId()) > 0 {
						req.UserGroupId = groupID
					}
					fsRes, err := fieldService.AddField(context.TODO(), &req, opss)
					if err != nil {
						loggerx.ErrorLog("restore", err.Error())
						path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     err.Error(),
							CurrentStep: "restore",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: db,
						}, userID)
						return
					}
					// 添加字段对应的语言
					for _, lang := range langData {
						appLang, exist := lang.GetApps()[appID]
						if exist {
							key := ds.GetDatastoreId() + "_" + f.GetFieldId()
							name, ok := appLang.Fields[key]
							if ok {
								langReq := language.AddAppLanguageDataRequest{
									Domain:   domain,
									LangCd:   lang.LangCd,
									AppId:    currentAppID,
									Type:     "fields",
									Key:      currentDatastoreID + "_" + fsRes.GetFieldId(),
									Value:    name,
									Writer:   userID,
									Database: db,
								}
								err := addAppLangData(langReq)
								if err != nil {
									loggerx.ErrorLog("restore", err.Error())
									path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
									// 发送消息 获取数据失败，终止任务
									jobx.ModifyTask(task.ModifyRequest{
										JobId:       jobID,
										Message:     err.Error(),
										CurrentStep: "restore",
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
						}
					}
					/* fields = append(fields, fsRes.GetFieldId()) */
				}
			}
			// 恢复台账的relations
			if len(ds.GetRelations()) > 0 {
				datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
				var relationReq datastore.AddRelationRequest
				relationReq.AppId = currentAppID
				relationReq.DatastoreId = dsMap[ds.GetDatastoreId()]
				relationReq.Database = db
				var relItem datastore.RelationItem
				for _, relation := range ds.GetRelations() {
					relItem.Fields = relation.GetFields()
					relItem.DatastoreId = dsMap[relation.GetDatastoreId()]
					relItem.RelationId = relation.GetRelationId()
					relationReq.Relation = &relItem
					_, err := datastoreService.AddRelation(context.TODO(), &relationReq)
					if err != nil {
						loggerx.ErrorLog("restore", err.Error())
						path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     err.Error(),
							CurrentStep: "restore",
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
			}

			// 若数据存在
			if response.GetBackup().GetHasData() {
				// 读取台账数据
				fi, err := os.Open(zipDir + "/items_" + ds.GetDatastoreId() + ".txt")
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}
				defer fi.Close()

				line := 0
				var items []*copy.BulkItem

				br := bufio.NewReader(fi)
				for {
					bytes, err := br.ReadBytes('\n')
					if err == io.EOF {
						if len(items) > 0 {

							splitLength := float64(len(items)) / 2000.0
							sp := math.Ceil(splitLength)

							datas := splitRestoreData(items, int64(sp))

							for _, its := range datas {
								// 插入数据库
								var bulk copy.BulkAddRequest
								bulk.AppId = currentAppID
								bulk.DatastoreId = currentDatastoreID
								bulk.Owners = accessKeys
								bulk.Writer = userID
								bulk.Items = its
								bulk.Database = db

								copyService := copy.NewCopyService("database", client.DefaultClient)
								_, err := copyService.BulkAddItems(context.TODO(), &bulk, opss)
								if err != nil {
									loggerx.ErrorLog("restore", err.Error())
									path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
									// 发送消息 获取数据失败，终止任务
									jobx.ModifyTask(task.ModifyRequest{
										JobId:       jobID,
										Message:     err.Error(),
										CurrentStep: "restore",
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

						}

						break
					}

					var it copy.BulkItem

					err = json.Unmarshal(bytes, &it)
					if err != nil {
						loggerx.ErrorLog("restore", err.Error())
						path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     err.Error(),
							CurrentStep: "restore",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: db,
						}, userID)
						return
					}

					itemMap := map[string]*copy.Value{}
					for fieldID, val := range it.GetItems() {
						// 用户数据默认是当前顾客的管理员用户
						if val.DataType == "user" {
							val.Value = userID
						}
						itemMap[fieldID] = val
					}

					it.Items = itemMap

					if line < 2000 {
						items = append(items, &it)
						line++
					}

					if line == 2000 {
						items = append(items, &it)
						// 插入数据库
						var bulk copy.BulkAddRequest
						bulk.AppId = currentAppID
						bulk.DatastoreId = currentDatastoreID
						bulk.Owners = accessKeys
						bulk.Writer = userID
						bulk.Items = items
						bulk.Database = db

						copyService := copy.NewCopyService("database", client.DefaultClient)
						_, err := copyService.BulkAddItems(context.TODO(), &bulk, opss)
						if err != nil {
							loggerx.ErrorLog("restore", err.Error())
							path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "restore",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: db,
							}, userID)
							return
						}

						items = items[:0]
						line = 0
					}
				}
			}

			dsActions = append(dsActions, &role.Action{
				ObjectId:  currentDatastoreID,
				Fields:    fields,
				ActionMap: dsActionsMap[currentDatastoreID], // TODO 需要获取权限
			})
		}

		// 恢复APP下流程情报
		// 获取不到模板组织情报故不复制流程情报 TODO

		// 添加台账和字段的role信息
		pmList = append(pmList, &role.Permission{
			RoleId:         roleID,
			PermissionType: "app",
			AppId:          currentAppID,
			ActionType:     "datastore",
			Actions:        dsActions,
		})

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_040"),
			CurrentStep: "restore",
			Database:    db,
		}, userID)

		// 读取报表数据
		var rpData []*report.Report
		err = filex.ReadFile(zipDir+"/reports.json", &rpData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		var rpActions []*role.Action

		rpMap := map[string]string{}
		// 恢复对应的报表
		for _, r := range rpData {
			reportService := report.NewReportService("report", client.DefaultClient)

			var rpReq report.AddReportRequest
			rpReq.Domain = domain
			rpReq.AppId = currentAppID
			rpReq.DatastoreId = dsMap[r.GetDatastoreId()]
			rpReq.Writer = userID
			rpReq.Database = db
			rpReq.IsUseGroup = r.GetIsUseGroup()
			rpReq.ConditionType = r.ConditionType

			if r.GetIsUseGroup() {
				gkf := &report.GroupInfo{
					ShowCount: r.GetGroupInfo().GetShowCount(),
				}
				var groupKeys []*report.KeyInfo
				for _, gk := range r.GetGroupInfo().GetGroupKeys() {
					groupKeys = append(groupKeys, &report.KeyInfo{
						IsLookup:    gk.GetIsLookup(),
						FieldId:     gk.GetFieldId(),
						DatastoreId: dsMap[gk.GetDatastoreId()],
						DataType:    gk.GetDataType(),
						AliasName:   gk.GetAliasName(),
						OptionId:    gk.GetOptionId(),
						Sort:        gk.GetSort(),
						IsDynamic:   gk.GetIsDynamic(),
						Unique:      gk.GetUnique(),
						Order:       gk.GetOrder(),
					})
				}
				gkf.GroupKeys = groupKeys
				var aggreKeys []*report.AggreKey
				for _, ak := range r.GetGroupInfo().GetAggreKeys() {
					aggreKeys = append(aggreKeys, &report.AggreKey{
						IsLookup:    ak.GetIsLookup(),
						FieldId:     ak.GetFieldId(),
						DatastoreId: dsMap[ak.GetDatastoreId()],
						DataType:    ak.GetDataType(),
						AliasName:   ak.GetAliasName(),
						OptionId:    ak.GetOptionId(),
						Sort:        ak.GetSort(),
						AggreType:   ak.GetAggreType(),
						Order:       ak.GetOrder(),
					})
				}
				gkf.AggreKeys = aggreKeys
				rpReq.GroupInfo = gkf
			} else {
				var selectKeys []*report.KeyInfo
				for _, gk := range r.GetSelectKeyInfos() {
					selectKeys = append(selectKeys, &report.KeyInfo{
						IsLookup:    gk.GetIsLookup(),
						FieldId:     gk.GetFieldId(),
						DatastoreId: dsMap[gk.GetDatastoreId()],
						DataType:    gk.GetDataType(),
						OptionId:    gk.GetOptionId(),
						AliasName:   gk.GetAliasName(),
						Sort:        gk.GetSort(),
						IsDynamic:   gk.GetIsDynamic(),
						Unique:      gk.GetUnique(),
						Order:       gk.GetOrder(),
					})
				}

				rpReq.SelectKeyInfos = selectKeys
			}
			var cdList []*report.ReportCondition
			for _, cd := range r.GetReportConditions() {
				cdList = append(cdList, &report.ReportCondition{
					FieldId:       cd.GetFieldId(),
					FieldType:     cd.GetFieldType(),
					SearchValue:   cd.GetSearchValue(),
					Operator:      cd.GetOperator(),
					ConditionType: cd.GetConditionType(),
					IsDynamic:     cd.GetIsDynamic(),
				})
			}
			rpReq.ReportConditions = cdList

			reRes, err := reportService.AddReport(context.TODO(), &rpReq, opss)
			if err != nil {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}
			rpMap[r.GetReportId()] = reRes.GetReportId()
			var rpPe map[string]bool
			for _, pe := range peData {
				if pe.ObjectId == r.ReportId {
					rpPe = pe.ActionMap
				}
			}

			// 添加role中报表的数据
			rpActions = append(rpActions, &role.Action{
				ObjectId: reRes.GetReportId(),
				ActionMap: map[string]bool{
					"read": rpPe["read"],
				},
			})

			// 添加报表多语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[appID]
				if exist {
					name, ok := appLang.Reports[r.GetReportId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   domain,
							LangCd:   lang.LangCd,
							AppId:    currentAppID,
							Type:     "reports",
							Key:      reRes.GetReportId(),
							Value:    name,
							Writer:   userID,
							Database: db,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("restore", err.Error())
							path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "restore",
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
				}
			}
		}

		pmList = append(pmList, &role.Permission{
			RoleId:         roleID,
			PermissionType: "app",
			AppId:          currentAppID,
			ActionType:     "report",
			Actions:        rpActions,
		})

		roleReq.Permissions = pmList

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_037"),
			CurrentStep: "restore",
			Database:    db,
		}, userID)

		// 读取仪表数据
		var dashData []*dashboard.Dashboard
		err = filex.ReadFile(zipDir+"/dashboards.json", &dashData)
		if err != nil {
			if err.Error() != "not found data" {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
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

		// 恢复仪表盘
		for _, dash := range dashData {
			dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)
			var req dashboard.AddDashboardRequest
			req.DashboardName = dash.GetDashboardName()
			req.Domain = domain
			req.AppId = currentAppID
			if r, exist := rpMap[dash.GetReportId()]; exist {
				req.ReportId = r
			}
			req.DashboardType = dash.GetDashboardType()
			req.XRange = dash.GetXRange()
			req.YRange = dash.GetYRange()
			req.TickType = dash.GetTickType()
			req.Ticks = dash.GetTicks()
			req.TickCount = dash.GetTickCount()
			req.GFieldId = dash.GetGFieldId()
			req.XFieldId = dash.GetXFieldId()
			req.YFieldId = dash.GetYFieldId()
			req.LimitInPlot = dash.GetLimitInPlot()
			req.StepType = dash.GetStepType()
			req.IsStack = dash.GetIsStack()
			req.IsPercent = dash.GetIsPercent()
			req.IsGroup = dash.GetIsGroup()
			req.Smooth = dash.GetSmooth()
			req.MinBarWidth = dash.GetMinBarWidth()
			req.MaxBarWidth = dash.GetMaxBarWidth()
			req.Radius = dash.GetRadius()
			req.InnerRadius = dash.GetInnerRadius()
			req.StartAngle = dash.GetStartAngle()
			req.EndAngle = dash.GetEndAngle()
			req.Slider = dash.GetSlider()
			req.Scrollbar = dash.GetScrollbar()
			req.Writer = dash.GetCreatedBy()
			req.Database = db

			dashRes, err := dashboardService.AddDashboard(context.TODO(), &req, opss)
			if err != nil {
				loggerx.ErrorLog("restore", err.Error())
				path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}

			// 添加字段对应的语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[appID]
				if exist {
					name, ok := appLang.Dashboards[dash.GetDashboardId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   domain,
							LangCd:   lang.LangCd,
							AppId:    currentAppID,
							Type:     "dashboards",
							Key:      dashRes.GetDashboardId(),
							Value:    name,
							Writer:   userID,
							Database: db,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("restore", err.Error())
							path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       jobID,
								Message:     err.Error(),
								CurrentStep: "restore",
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
				}
			}
		}

		err = os.RemoveAll(zipDir)
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "restore",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_036"),
			CurrentStep: "restore",
			Database:    db,
		}, userID)

		// 恢复role信息
		roleService := role.NewRoleService("manage", client.DefaultClient)
		_, err = roleService.ModifyRole(context.TODO(), &roleReq, opss)
		if err != nil {
			loggerx.ErrorLog("restore", err.Error())
			path := filex.WriteAndSaveFile(domain, "system", []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "restore",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		aclx.SetRoleCasbin(roleReq.GetRoleId(), roleReq.GetPermissions())

		// 通知刷新多语言数据
		langx.RefreshLanguage(userID, domain)

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)
	}()

	return nil
}

func setPermission(customerId, allowType, apiKey string, canCheck bool, peMap map[string]bool) (am map[string]bool) {

	objectType := "base"
	if allowType == "datastore" {
		objectType = getDatastoreType(apiKey, canCheck)
	}
	if allowType == "report" {
		objectType = "report"
	}
	if allowType == "folder" {
		objectType = "folder"
	}

	actions, err := getActions(customerId, allowType, objectType, peMap)
	if err != nil {
		return make(map[string]bool)
	}

	return actions
}

func getActions(customerId, allowType, objType string, peMap map[string]bool) (am map[string]bool, err error) {
	// 获取顾客信息
	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomerRequest
	req.CustomerId = customerId
	response, err := customerService.FindCustomer(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getLevelAllows", err.Error())
		return
	}
	// 通过顾客信息获取顾客的授权等级信息
	levelService := level.NewLevelService("manage", client.DefaultClient)

	var lreq level.FindLevelRequest
	lreq.LevelId = response.GetCustomer().GetLevel()
	levelResp, err := levelService.FindLevel(context.TODO(), &lreq)
	if err != nil {
		loggerx.ErrorLog("getLevelAllows", err.Error())
		return
	}

	if len(levelResp.GetLevel().GetAllows()) == 0 {
		return
	}

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var alreq allow.FindLevelAllowsRequest
	// 从query获取
	alreq.AllowList = levelResp.GetLevel().GetAllows()

	allowResp, err := allowService.FindLevelAllows(context.TODO(), &alreq)
	if err != nil {
		loggerx.ErrorLog("getLevelAllows", err.Error())
		return
	}

	actionMap := make(map[string]bool)

	for _, a := range allowResp.GetAllows() {
		if a.AllowType == allowType && a.ObjectType == objType {
			for _, action := range a.Actions {
				actionMap[action.ApiKey] = peMap[action.ApiKey]
			}
			return actionMap, nil
		}
	}

	return
}

func getDatastoreType(apiKey string, canCheck bool) (dsType string) {

	if canCheck {
		return "check"
	}
	if apiKey == "keiyakudaicho" {
		return "lease"
	}
	if apiKey == "paymentStatus" || apiKey == "repayment" || apiKey == "paymentInterest" || apiKey == "rireki" {
		return "lease_relation"
	}
	if apiKey == "shiwake" {
		return "journal"
	}

	return "base"
}

func addAppLangData(req language.AddAppLanguageDataRequest) error {
	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}
	languageService := language.NewLanguageService("global", client.DefaultClient)
	_, err := languageService.AddAppLanguageData(context.TODO(), &req, opss)
	if err != nil {
		return err
	}

	return nil
}

func setOptionSequenceValue(database, appID, value string) error {
	// 序列值
	intVal, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	// 更新序列值
	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.SetSequenceValueRequest
	req.Database = database
	req.SequenceName = "option_" + appID
	req.SequenceValue = intVal

	_, err = fieldService.SetSequenceValue(context.TODO(), &req)
	if err != nil {
		return err
	}

	return nil
}

// 分割数据
func splitRestoreData(arr []*copy.BulkItem, num int64) (segmens [][]*copy.BulkItem) {
	max := int64(len(arr))
	if num == 0 {
		segmens = append(segmens, arr)
		return
	}
	if max < num {
		segmens = append(segmens, arr)
		return
	}
	var step = max / num
	var beg int64
	var end int64
	for i := int64(0); i < num || end < max; i++ {
		beg = 0 + i*step
		end = beg + step

		if end > max {
			segmens = append(segmens, arr[beg:max])
		} else {
			segmens = append(segmens, arr[beg:end])
		}
	}
	return
}
