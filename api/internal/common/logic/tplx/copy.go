package tplx

import (
	"context"
	"fmt"
	"time"

	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	merrors "github.com/micro/go-micro/v2/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/aclx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/srv/database/proto/copy"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/database/proto/print"
	"rxcsoft.cn/pit3/srv/database/proto/query"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
	"rxcsoft.cn/pit3/srv/report/proto/report"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
)

// CopyParams
type CopyParams struct {
	WithData     bool
	WithFile     bool
	JobID        string
	AppID        string
	DB           string
	Domain       string
	CurrentAppID string
	UserID       string
	Lang         string
	Roles        []string
}

// CopyApp 复制app
func CopyApp(params CopyParams) error {
	go func() {
		// 发送消息-获取复制元APP情报
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       params.JobID,
			Message:     i18n.Tr(params.Lang, "job.J_023"),
			CurrentStep: "get-app-data",
			Database:    params.DB,
		}, params.UserID)
		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 5
			o.DialTimeout = time.Minute * 5
		}
		// 获取语言数据
		languageService := language.NewLanguageService("global", client.DefaultClient)
		var lgReq language.FindLanguagesRequest
		lgReq.Domain = params.Domain
		lgReq.Database = params.DB
		lgResponse, err := languageService.FindLanguages(context.TODO(), &lgReq, opss)
		if err != nil {
			loggerx.ErrorLog("copy", err.Error())
			path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       params.JobID,
				Message:     err.Error(),
				CurrentStep: "get-app-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: params.DB,
			}, params.UserID)
			return
		}
		langData := lgResponse.GetLanguageList()

		// 获取APP下属选项
		optionService := option.NewOptionService("database", client.DefaultClient)
		var opReq option.FindOptionLabelsRequest
		opReq.Database = params.DB
		opReq.AppId = params.AppID
		opResponse, err := optionService.FindOptionLabels(context.TODO(), &opReq, opss)
		if err != nil {
			loggerx.ErrorLog("copy", err.Error())
			path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       params.JobID,
				Message:     err.Error(),
				CurrentStep: "get-app-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: params.DB,
			}, params.UserID)
			return
		}

		// 发送消息-恢复选项数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       params.JobID,
			Message:     i18n.Tr(params.Lang, "job.J_039"),
			CurrentStep: "restore",
			Database:    params.DB,
		}, params.UserID)

		// 恢复选项
		maxOptionID := "000000"
		for _, op := range opResponse.GetOptions() {
			optionService := option.NewOptionService("database", client.DefaultClient)
			var req option.AddRequest
			req.OptionId = op.GetOptionId()
			req.OptionMemo = op.GetOptionMemo()
			req.OptionOrder = op.GetOptionOrder()
			req.OptionValue = op.GetOptionValue()
			req.AppId = params.CurrentAppID
			req.Writer = op.GetCreatedBy()
			req.Database = params.DB

			// 设置最大选项组ID值
			if op.GetOptionId() > maxOptionID {
				maxOptionID = op.GetOptionId()
			}

			_, err := optionService.AddOption(context.TODO(), &req, opss)
			if err != nil {
				loggerx.ErrorLog("copy", err.Error())
				path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       params.JobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: params.DB,
				}, params.UserID)
				return
			}

			// 恢复对应的语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[params.AppID]
				if exist {
					name, ok := appLang.Options[op.GetOptionId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   params.Domain,
							LangCd:   lang.LangCd,
							AppId:    params.CurrentAppID,
							Type:     "options",
							Key:      op.GetOptionId(),
							Value:    name,
							Writer:   op.GetCreatedBy(),
							Database: params.DB,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("copy", err.Error())
							path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       params.JobID,
								Message:     err.Error(),
								CurrentStep: "restore",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: params.DB,
							}, params.UserID)
							return
						}
					}
					labelName, ok := appLang.Options[op.GetOptionId()+"_"+op.GetOptionValue()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   params.Domain,
							LangCd:   lang.LangCd,
							AppId:    params.CurrentAppID,
							Type:     "options",
							Key:      op.GetOptionId() + "_" + op.GetOptionValue(),
							Value:    labelName,
							Writer:   op.GetCreatedBy(),
							Database: params.DB,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("copy", err.Error())
							path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       params.JobID,
								Message:     err.Error(),
								CurrentStep: "restore",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: params.DB,
							}, params.UserID)
							return
						}
					}
				}
			}
		}
		if len(opResponse.GetOptions()) > 0 {
			// 设置序列值
			err = setOptionSequenceValue(params.DB, params.CurrentAppID, maxOptionID)
			if err != nil {
				fmt.Printf("error: %v \r\n", err)
			}
		}

		// 发送消息 获取模板文件信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       params.JobID,
			Message:     i18n.Tr(params.Lang, "job.J_038"),
			CurrentStep: "restore",
			Database:    params.DB,
		}, params.UserID)

		// 获取APP下属台账
		datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
		var dsReq datastore.DatastoresRequest
		dsReq.Database = params.DB
		dsReq.AppId = params.AppID
		dsResponse, err := datastoreService.FindDatastores(context.TODO(), &dsReq, opss)
		if err != nil {
			loggerx.ErrorLog("copy", err.Error())
			path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       params.JobID,
				Message:     err.Error(),
				CurrentStep: "restore",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: params.DB,
			}, params.UserID)
			return
		}

		// 获取APP下属字段
		fieldService := field.NewFieldService("database", client.DefaultClient)
		var fsReq field.AppFieldsRequest
		fsReq.AppId = params.AppID
		fsReq.Database = params.DB
		fsReq.InvalidatedIn = "true"
		fsResponse, err := fieldService.FindAppFields(context.TODO(), &fsReq, opss)
		if err != nil {
			loggerx.ErrorLog("copy", err.Error())
			path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       params.JobID,
				Message:     err.Error(),
				CurrentStep: "restore",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: params.DB,
			}, params.UserID)
			return
		}

		// 台账复制前后映射
		dsMap := map[string]string{}

		printService := print.NewPrintService("database", client.DefaultClient)

		// 恢复台账情报
		for _, ds := range dsResponse.GetDatastores() {
			var req datastore.AddRequest
			req.AppId = params.CurrentAppID
			req.ApiKey = ds.GetApiKey()
			req.CanCheck = ds.GetCanCheck()
			req.ShowInMenu = ds.GetShowInMenu()
			req.NoStatus = ds.GetNoStatus()
			req.Encoding = ds.GetEncoding()
			req.Writer = ds.GetCreatedBy()
			req.Database = params.DB
			req.DatastoreName = ds.GetDatastoreName()
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
			// 恢复台账的unique_fields
			req.UniqueFields = ds.GetUniqueFields()

			dsRes, err := datastoreService.AddDatastore(context.TODO(), &req, opss)
			if err != nil {
				loggerx.ErrorLog("copy", err.Error())
				path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       params.JobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: params.DB,
				}, params.UserID)
				return
			}

			// 台账复制前后映射编辑
			dsMap[ds.GetDatastoreId()] = dsRes.GetDatastoreId()

			// 恢复台账名称多语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[params.AppID]
				if exist {
					name, ok := appLang.Datastores[ds.GetDatastoreId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   params.Domain,
							LangCd:   lang.LangCd,
							AppId:    params.CurrentAppID,
							Type:     "datastores",
							Key:      dsRes.GetDatastoreId(),
							Value:    name,
							Writer:   ds.GetCreatedBy(),
							Database: params.DB,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("copy", err.Error())
							path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       params.JobID,
								Message:     err.Error(),
								CurrentStep: "restore",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: params.DB,
							}, params.UserID)
							return
						}
					}
				}
			}

			// 恢复台账映射关系
			var reqMap datastore.AddMappingRequest
			// 映射所属情报
			reqMap.AppId = params.CurrentAppID
			reqMap.Database = params.DB
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
				res, err := datastoreService.AddDatastoreMapping(context.TODO(), &reqMap, opss)
				if err != nil {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}

				// 恢复映射对应的多语言
				for _, lang := range langData {
					appLang, exist := lang.GetApps()[params.AppID]
					if exist {
						name, ok := appLang.Mappings[ds.GetDatastoreId()+"_"+mapInfo.GetMappingId()]
						if ok {
							langReq := language.AddAppLanguageDataRequest{
								Domain:   params.Domain,
								LangCd:   lang.LangCd,
								AppId:    params.CurrentAppID,
								Type:     "mappings",
								Key:      reqMap.DatastoreId + "_" + res.MappingId,
								Value:    name,
								Writer:   ds.GetCreatedBy(),
								Database: params.DB,
							}
							err := addAppLangData(langReq)
							if err != nil {
								loggerx.ErrorLog("copy", err.Error())
								path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
								// 发送消息 获取数据失败，终止任务
								jobx.ModifyTask(task.ModifyRequest{
									JobId:       params.JobID,
									Message:     err.Error(),
									CurrentStep: "restore",
									EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
									ErrorFile: &task.File{
										Url:  path.MediaLink,
										Name: path.Name,
									},
									Database: params.DB,
								}, params.UserID)
								return
							}
						}
					}
				}
			}

			// 恢复台账打印设置
			var reqP print.FindPrintRequest
			reqP.AppId = params.AppID
			reqP.Database = params.DB
			reqP.DatastoreId = ds.GetDatastoreId()
			resP, err := printService.FindPrint(context.TODO(), &reqP, opss)
			if err != nil {
				er := merrors.Parse(err.Error())
				if er.GetDetail() != mongo.ErrNoDocuments.Error() {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
			}
			if resP != nil && resP.Print != nil {
				var req print.AddPrintRequest
				req.AppId = params.CurrentAppID
				req.Writer = resP.Print.GetCreatedBy()
				req.Database = params.DB
				req.DatastoreId = dsRes.GetDatastoreId()
				req.Page = resP.Print.Page
				req.Orientation = resP.Print.Orientation
				req.TitleWidth = resP.Print.TitleWidth
				req.ShowSign = resP.Print.ShowSign
				req.SignName1 = resP.Print.SignName1
				req.SignName2 = resP.Print.SignName2
				req.ShowSystem = resP.Print.ShowSystem
				req.CheckField = resP.Print.CheckField
				req.Fields = resP.Print.Fields
				_, err := printService.AddPrint(context.TODO(), &req)
				if err != nil {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
			}
			// 恢复台账快捷方式
			queryService := query.NewQueryService("database", client.DefaultClient)
			var reqQuery query.FindQueriesRequest
			// 取到快捷方式的初始值
			reqQuery.DatastoreId = ds.GetDatastoreId()
			reqQuery.AppId = params.AppID
			reqQuery.Database = params.DB
			resQuery, err := queryService.FindQueries(context.TODO(), &reqQuery, opss)
			if err != nil {
				er := merrors.Parse(err.Error())
				if er.GetDetail() != mongo.ErrNoDocuments.Error() {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
			}
			// 恢复快捷方式
			if len(resQuery.GetQueryList()) != 0 {
				var reqQue query.AddRequest
				reqQue.Database = params.DB
				reqQue.DatastoreId = dsRes.GetDatastoreId()
				for _, que := range resQuery.GetQueryList() {
					queryService := query.NewQueryService("database", client.DefaultClient)
					reqQue.Writer = que.GetCreatedBy()
					reqQue.UserId = que.GetUserId()
					reqQue.AppId = params.CurrentAppID
					reqQue.ConditionType = que.GetConditionType()
					reqQue.Conditions = que.GetConditions()
					reqQue.Description = que.GetDescription()
					reqQue.Fields = que.GetFields()
					reqQue.QueryName = que.GetQueryName()
					_, err := queryService.AddQuery(context.TODO(), &reqQue, opss)
					if err != nil {
						loggerx.ErrorLog("copy", err.Error())
						path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       params.JobID,
							Message:     err.Error(),
							CurrentStep: "restore",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: params.DB,
						}, params.UserID)
						return
					}
				}
			}
		}

		// 恢复台账其他信息
		for _, ds := range dsResponse.GetDatastores() {
			currentDatastoreID := dsMap[ds.DatastoreId]
			// 恢复台账对应的字段
			for _, f := range fsResponse.GetFields() {
				if f.GetDatastoreId() == ds.DatastoreId {
					fieldService := field.NewFieldService("database", client.DefaultClient)
					req := field.AddRequest{
						AppId:         params.CurrentAppID,
						DatastoreId:   currentDatastoreID,
						FieldType:     f.GetFieldType(),
						FieldId:       f.GetFieldId(),
						IsRequired:    f.GetIsRequired(),
						IsImage:       f.GetIsImage(),
						AsTitle:       f.GetAsTitle(),
						Unique:        f.GetUnique(),
						LookupAppId:   params.CurrentAppID,
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
						Writer:        f.GetCreatedBy(),
						Database:      params.DB,
					}
					if d, exist := dsMap[f.GetLookupDatastoreId()]; exist {
						req.LookupDatastoreId = d
					}

					fsRes, err := fieldService.AddField(context.TODO(), &req, opss)
					if err != nil {
						loggerx.ErrorLog("restore", err.Error())
						path := filex.WriteAndSaveFile(params.Domain, "system", []string{err.Error()})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       params.JobID,
							Message:     err.Error(),
							CurrentStep: "restore",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: params.DB,
						}, params.UserID)
						return
					}
					// 添加字段对应的语言
					for _, lang := range langData {
						appLang, exist := lang.GetApps()[params.AppID]
						if exist {
							key := ds.GetDatastoreId() + "_" + f.GetFieldId()
							name, ok := appLang.Fields[key]
							if ok {
								langReq := language.AddAppLanguageDataRequest{
									Domain:   params.Domain,
									LangCd:   lang.LangCd,
									AppId:    params.CurrentAppID,
									Type:     "fields",
									Key:      currentDatastoreID + "_" + fsRes.GetFieldId(),
									Value:    name,
									Writer:   f.GetCreatedBy(),
									Database: params.DB,
								}
								err := addAppLangData(langReq)
								if err != nil {
									loggerx.ErrorLog("restore", err.Error())
									path := filex.WriteAndSaveFile(params.Domain, "system", []string{err.Error()})
									// 发送消息 获取数据失败，终止任务
									jobx.ModifyTask(task.ModifyRequest{
										JobId:       params.JobID,
										Message:     err.Error(),
										CurrentStep: "restore",
										EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
										ErrorFile: &task.File{
											Url:  path.MediaLink,
											Name: path.Name,
										},
										Database: params.DB,
									}, params.UserID)
									return
								}
							}
						}
					}
				}
			}
			// 恢复台账的relations
			if len(ds.GetRelations()) > 0 {
				datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
				var relationReq datastore.AddRelationRequest
				relationReq.AppId = params.CurrentAppID
				relationReq.DatastoreId = dsMap[ds.GetDatastoreId()]
				relationReq.Database = params.DB
				var relItem datastore.RelationItem
				for _, relation := range ds.GetRelations() {
					relItem.Fields = relation.GetFields()
					relItem.DatastoreId = dsMap[relation.GetDatastoreId()]
					relItem.RelationId = relation.GetRelationId()
					relationReq.Relation = &relItem
					_, err := datastoreService.AddRelation(context.TODO(), &relationReq)
					if err != nil {
						loggerx.ErrorLog("copy", err.Error())
						path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
						// 发送消息-获取数据失败,终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       params.JobID,
							Message:     err.Error(),
							CurrentStep: "restore",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: params.DB,
						}, params.UserID)
						return
					}
				}
			}

			if params.WithData {
				// 恢复台账数据
				copyService := copy.NewCopyService("database", client.DefaultClient)

				var opss client.CallOption = func(o *client.CallOptions) {
					o.RequestTimeout = time.Hour * 1
					o.DialTimeout = time.Hour * 1
				}

				cReq := copy.CopyItemsRequest{
					AppId:           params.AppID,
					DatastoreId:     ds.GetDatastoreId(),
					CopyAppId:       params.CurrentAppID,
					CopyDatastoreId: currentDatastoreID,
					Database:        params.DB,
					WithFile:        params.WithFile,
				}

				_, err := copyService.CopyItems(context.TODO(), &cReq, opss)
				if err != nil {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息-获取数据失败,终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
			}

		}
		// 复制台账文件
		if params.WithData && params.WithFile {
			// 恢复台账文件
			dsParam := filex.CopyFile{
				Domain:       params.Domain,
				OldApp:       params.AppID,
				NewApp:       params.CurrentAppID,
				DatastoreMap: dsMap,
			}
			filex.CopyMinioFile(dsParam)
		}

		// 恢复APP下流程情报
		workflowService := workflow.NewWfService("workflow", client.DefaultClient)
		var req workflow.WorkflowsRequest
		req.AppId = params.AppID
		req.Database = params.DB
		oldwfs, err := workflowService.FindWorkflows(context.TODO(), &req, opss)
		if err != nil {
			er := merrors.Parse(err.Error())
			if er.GetDetail() != mongo.ErrNoDocuments.Error() {
				loggerx.ErrorLog("copy", err.Error())
				path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
				// 发送消息-获取数据失败,终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       params.JobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: params.DB,
				}, params.UserID)
				return
			}
		}
		if oldwfs != nil && len(oldwfs.Workflows) > 0 {
			for _, wf := range oldwfs.Workflows {
				// 流程情报
				var wReq workflow.AddRequest
				wReq.AppId = params.CurrentAppID
				wReq.Writer = wf.CreatedBy
				wReq.Database = params.DB
				wReq.IsValid = wf.IsValid
				wReq.GroupId = wf.GroupId
				wReq.AcceptOrDismiss = wf.AcceptOrDismiss
				wReq.WorkflowType = wf.WorkflowType
				newParams := wf.Params
				newParams["datastore"] = dsMap[wf.Params["datastore"]]
				wReq.Params = newParams
				// 恢复流程情报
				resWf, err := workflowService.AddWorkflow(context.TODO(), &wReq, opss)
				if err != nil {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息-获取数据失败,终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
				// 节点情报
				nodeService := node.NewNodeService("workflow", client.DefaultClient)
				var nReq node.NodesRequest
				nReq.WfId = wf.WfId
				nReq.Database = params.DB
				nResp, err := nodeService.FindNodes(context.TODO(), &nReq, opss)
				if err != nil {
					er := merrors.Parse(err.Error())
					if er.GetDetail() != mongo.ErrNoDocuments.Error() {
						loggerx.ErrorLog("copy", err.Error())
						path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
						// 发送消息-获取数据失败,终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       params.JobID,
							Message:     err.Error(),
							CurrentStep: "restore",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: params.DB,
						}, params.UserID)
						return
					}
				}
				// 恢复节点情报
				if nResp != nil && len(nResp.Nodes) > 0 {
					for _, n := range nResp.Nodes {
						var nReq node.AddRequest
						nReq.NodeId = n.NodeId
						nReq.NodeName = n.NodeName
						nReq.WfId = resWf.GetWfId()
						nReq.NodeType = n.NodeType
						nReq.PrevNode = n.PrevNode
						nReq.NextNode = n.NextNode
						nReq.Assignees = n.Assignees
						nReq.ActType = n.ActType
						nReq.NodeGroupId = n.NodeGroupId
						nReq.Database = params.DB
						nReq.Writer = n.CreatedBy
						_, err := nodeService.AddNode(context.TODO(), &nReq, opss)
						if err != nil {
							loggerx.ErrorLog("copy", err.Error())
							path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
							// 发送消息-获取数据失败,终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       params.JobID,
								Message:     err.Error(),
								CurrentStep: "restore",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: params.DB,
							}, params.UserID)
							return
						}
					}
				}

				// 恢复流程对应的语言
				languageService := language.NewLanguageService("global", client.DefaultClient)
				for _, lang := range langData {
					// 流程名称多语言
					appLang, ok := lang.GetApps()[params.AppID]
					if ok {
						wName, exist := appLang.GetWorkflows()[wf.WfId]
						if exist {
							langParams := language.AddAppLanguageDataRequest{
								Domain:   params.Domain,
								LangCd:   lang.LangCd,
								AppId:    params.CurrentAppID,
								Type:     "workflows",
								Key:      resWf.GetWfId(),
								Value:    wName,
								Writer:   wf.GetCreatedBy(),
								Database: params.DB,
							}
							_, err = languageService.AddAppLanguageData(context.TODO(), &langParams)
							if err != nil {
								loggerx.ErrorLog("copy", err.Error())
								path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
								// 发送消息-获取数据失败,终止任务
								jobx.ModifyTask(task.ModifyRequest{
									JobId:       params.JobID,
									Message:     err.Error(),
									CurrentStep: "restore",
									EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
									ErrorFile: &task.File{
										Url:  path.MediaLink,
										Name: path.Name,
									},
									Database: params.DB,
								}, params.UserID)
								return
							}
						}

						// 流程菜单名称多语言
						wmName, exist := appLang.GetWorkflows()["menu_"+wf.WfId]
						if exist {
							menuLangParams := language.AddAppLanguageDataRequest{
								Domain:   params.Domain,
								LangCd:   lang.LangCd,
								AppId:    params.CurrentAppID,
								Type:     "workflows",
								Key:      "menu_" + resWf.GetWfId(),
								Value:    wmName,
								Writer:   wf.GetCreatedBy(),
								Database: params.DB,
							}
							_, err = languageService.AddAppLanguageData(context.TODO(), &menuLangParams)
							if err != nil {
								loggerx.ErrorLog("copy", err.Error())
								path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
								// 发送消息-获取数据失败,终止任务
								jobx.ModifyTask(task.ModifyRequest{
									JobId:       params.JobID,
									Message:     err.Error(),
									CurrentStep: "restore",
									EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
									ErrorFile: &task.File{
										Url:  path.MediaLink,
										Name: path.Name,
									},
									Database: params.DB,
								}, params.UserID)
								return
							}
						}
					}
				}
			}
		}

		// 发送消息 获取报表信息
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       params.JobID,
			Message:     i18n.Tr(params.Lang, "job.J_040"),
			CurrentStep: "restore",
			Database:    params.DB,
		}, params.UserID)

		// 获取报表数据
		reportService := report.NewReportService("report", client.DefaultClient)
		var rpReq report.FindReportsRequest
		rpReq.Domain = params.Domain
		rpReq.AppId = params.AppID
		rpReq.Database = params.DB
		rpResponse, err := reportService.FindReports(context.TODO(), &rpReq, opss)
		if err != nil {
			loggerx.ErrorLog("copy", err.Error())
			path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       params.JobID,
				Message:     err.Error(),
				CurrentStep: "restore",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: params.DB,
			}, params.UserID)
			return
		}

		rpMap := map[string]string{}
		// 恢复对应的报表
		for _, r := range rpResponse.GetReports() {
			reportService := report.NewReportService("report", client.DefaultClient)
			var rpReq report.AddReportRequest
			rpReq.Domain = params.Domain
			rpReq.AppId = params.CurrentAppID
			rpReq.DatastoreId = dsMap[r.GetDatastoreId()]
			rpReq.Writer = r.GetCreatedBy()
			rpReq.Database = params.DB
			rpReq.IsUseGroup = r.GetIsUseGroup()
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
						AliasName:   gk.GetAliasName(),
						OptionId:    gk.GetOptionId(),
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
					FieldId:     cd.GetFieldId(),
					FieldType:   cd.GetFieldType(),
					SearchValue: cd.GetSearchValue(),
					Operator:    cd.GetOperator(),
					//ConditionType: cd.GetConditionType(),
					IsDynamic: cd.GetIsDynamic(),
				})
			}
			rpReq.ConditionType = r.ConditionType
			rpReq.ReportConditions = cdList

			reRes, err := reportService.AddReport(context.TODO(), &rpReq, opss)
			if err != nil {
				loggerx.ErrorLog("copy", err.Error())
				path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       params.JobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: params.DB,
				}, params.UserID)
				return
			}

			rpMap[r.GetReportId()] = reRes.GetReportId()

			// 添加报表多语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[params.AppID]
				if exist {
					name, ok := appLang.Reports[r.GetReportId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   params.Domain,
							LangCd:   lang.LangCd,
							AppId:    params.CurrentAppID,
							Type:     "reports",
							Key:      reRes.GetReportId(),
							Value:    name,
							Writer:   r.GetCreatedBy(),
							Database: params.DB,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("copy", err.Error())
							path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       params.JobID,
								Message:     err.Error(),
								CurrentStep: "restore",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: params.DB,
							}, params.UserID)
							return
						}
					}
				}
			}
		}
		// 发送消息-恢复角色数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       params.JobID,
			Message:     i18n.Tr(params.Lang, "job.J_036"),
			CurrentStep: "restore",
			Database:    params.DB,
		}, params.UserID)

		pmservice := permission.NewPermissionService("manage", client.DefaultClient)

		for _, roleId := range params.Roles {
			// 角色许可
			var pmList []*role.Permission

			var preq permission.FindPermissionsRequest
			preq.RoleId = roleId
			preq.Database = params.DB

			permList, err := pmservice.FindPermissions(context.TODO(), &preq)
			if err != nil {
				loggerx.ErrorLog("copy", err.Error())
				path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
				// 发送消息-获取数据失败,终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       params.JobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: params.DB,
				}, params.UserID)
				return
			}

			// 将原有的action权限添加进去。
			for _, p := range permList.Permission {

				var acts []*role.Action
				for _, a := range p.Actions {
					acts = append(acts, &role.Action{
						ObjectId:  a.ObjectId,
						Fields:    a.Fields,
						ActionMap: a.ActionMap,
					})
				}

				pmList = append(pmList, &role.Permission{
					PermissionId:   p.PermissionId,
					RoleId:         p.RoleId,
					PermissionType: p.PermissionType,
					AppId:          p.AppId,
					ActionType:     p.ActionType,
					Actions:        acts,
					CreatedAt:      p.CreatedAt,
					CreatedBy:      p.CreatedBy,
					UpdatedAt:      p.UpdatedAt,
					UpdatedBy:      p.UpdatedBy,
				})
			}

			{
				// 台账
				var reqPM permission.FindActionsRequest
				reqPM.AppId = params.AppID
				reqPM.Database = params.DB
				reqPM.PermissionType = "app"
				reqPM.RoleId = []string{roleId}

				reqPM.ActionType = "datastore"
				pmActions, err := pmservice.FindActions(context.TODO(), &reqPM)
				if err != nil {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息-获取数据失败,终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
				// 台账角色权限点
				var actions []*role.Action
				for _, action := range pmActions.GetActions() {
					actions = append(actions, &role.Action{
						ObjectId:  dsMap[action.GetObjectId()],
						Fields:    action.GetFields(),
						ActionMap: action.GetActionMap(), // TODO 需要获取权限
					})
				}

				// 添加台账和字段的role信息
				pmList = append(pmList, &role.Permission{
					RoleId:         roleId,
					PermissionType: "app",
					AppId:          params.CurrentAppID,
					ActionType:     "datastore",
					Actions:        actions,
				})
			}
			{
				// 报表
				var reqPM permission.FindActionsRequest
				reqPM.AppId = params.AppID
				reqPM.Database = params.DB
				reqPM.PermissionType = "app"
				reqPM.RoleId = []string{roleId}
				reqPM.ActionType = "report"

				pmActions, err := pmservice.FindActions(context.TODO(), &reqPM)
				if err != nil {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息-获取数据失败,终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
				// 台账角色权限点
				var actions []*role.Action
				for _, action := range pmActions.GetActions() {
					actions = append(actions, &role.Action{
						ObjectId:  rpMap[action.GetObjectId()],
						Fields:    action.GetFields(),
						ActionMap: action.GetActionMap(), // TODO 需要获取权限
					})
				}
				pmList = append(pmList, &role.Permission{
					RoleId:         roleId,
					PermissionType: "app",
					AppId:          params.CurrentAppID,
					ActionType:     "report",
					Actions:        actions,
				})
			}

			{
				// 文档
				var reqPM permission.FindActionsRequest
				reqPM.Database = params.DB
				reqPM.PermissionType = "common"
				reqPM.RoleId = []string{roleId}

				reqPM.ActionType = "folder"
				pmActions, err := pmservice.FindActions(context.TODO(), &reqPM)
				if err != nil {
					loggerx.ErrorLog("copy", err.Error())
					path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
					// 发送消息-获取数据失败,终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       params.JobID,
						Message:     err.Error(),
						CurrentStep: "restore",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: params.DB,
					}, params.UserID)
					return
				}
				// 台账角色权限点
				var actions []*role.Action
				for _, action := range pmActions.GetActions() {
					actions = append(actions, &role.Action{
						ObjectId:  dsMap[action.GetObjectId()],
						Fields:    action.GetFields(),
						ActionMap: action.GetActionMap(), // TODO 需要获取权限
					})
				}

				// 添加台账和字段的role信息
				pmList = append(pmList, &role.Permission{
					RoleId:         roleId,
					PermissionType: "common",
					ActionType:     "folder",
					Actions:        actions,
				})
			}

			// 恢复role信息
			roleService := role.NewRoleService("manage", client.DefaultClient)
			// 角色更新参数
			var roleReq role.ModifyRoleRequest
			roleReq.RoleId = roleId
			roleReq.Database = params.DB
			roleReq.Permissions = pmList
			_, err = roleService.ModifyRole(context.TODO(), &roleReq, opss)
			if err != nil {
				loggerx.ErrorLog("copy", err.Error())
				path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       params.JobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: params.DB,
				}, params.UserID)
				return
			}

			aclx.SetRoleCasbin(roleReq.GetRoleId(), pmList)

			// 通知刷新多语言数据
			langx.RefreshLanguage(params.UserID, params.Domain)
		}

		// 发送消息-恢复仪表盘
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       params.JobID,
			Message:     i18n.Tr(params.Lang, "job.J_037"),
			CurrentStep: "restore",
			Database:    params.DB,
		}, params.UserID)

		// 获取仪表盘
		dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)
		var dashReq dashboard.FindDashboardsRequest
		dashReq.Domain = params.Domain
		dashReq.AppId = params.AppID
		dashReq.Database = params.DB
		dashResponse, err := dashboardService.FindDashboards(context.TODO(), &dashReq, opss)
		if err != nil {
			loggerx.ErrorLog("copy", err.Error())
			path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       params.JobID,
				Message:     err.Error(),
				CurrentStep: "restore",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: params.DB,
			}, params.UserID)
			return
		}

		// 恢复仪表盘
		for _, dash := range dashResponse.GetDashboards() {
			dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)
			var req dashboard.AddDashboardRequest
			req.DashboardName = dash.GetDashboardName()
			req.Domain = params.Domain
			req.AppId = params.CurrentAppID
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
			req.Database = params.DB
			dashRes, err := dashboardService.AddDashboard(context.TODO(), &req, opss)
			if err != nil {
				loggerx.ErrorLog("copy", err.Error())
				path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       params.JobID,
					Message:     err.Error(),
					CurrentStep: "restore",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: params.DB,
				}, params.UserID)
				return
			}

			// 添加仪表盘对应的语言
			for _, lang := range langData {
				appLang, exist := lang.GetApps()[params.AppID]
				if exist {
					name, ok := appLang.Dashboards[dash.GetDashboardId()]
					if ok {
						langReq := language.AddAppLanguageDataRequest{
							Domain:   params.Domain,
							LangCd:   lang.LangCd,
							AppId:    params.CurrentAppID,
							Type:     "dashboards",
							Key:      dashRes.GetDashboardId(),
							Value:    name,
							Writer:   dash.GetCreatedBy(),
							Database: params.DB,
						}
						err := addAppLangData(langReq)
						if err != nil {
							loggerx.ErrorLog("copy", err.Error())
							path := filex.WriteAndSaveFile(params.Domain, params.AppID, []string{err.Error()})
							// 发送消息 获取数据失败，终止任务
							jobx.ModifyTask(task.ModifyRequest{
								JobId:       params.JobID,
								Message:     err.Error(),
								CurrentStep: "restore",
								EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
								ErrorFile: &task.File{
									Url:  path.MediaLink,
									Name: path.Name,
								},
								Database: params.DB,
							}, params.UserID)
							return
						}
					}
				}
			}
		}

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       params.JobID,
			Message:     i18n.Tr(params.Lang, "job.J_028"),
			CurrentStep: "end",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    params.DB,
		}, params.UserID)

	}()
	return nil
}
