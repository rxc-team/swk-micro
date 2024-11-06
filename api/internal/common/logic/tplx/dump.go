package tplx

import (
	"bufio"
	"context"
	"encoding/json"
	"math"
	"os"
	"path"
	"time"

	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/database/proto/print"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
	"rxcsoft.cn/pit3/srv/report/proto/report"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

func Dump(db, tplApp, tplDb, domain, timestamp, lang, userId string, req backup.AddBackupRequest) {
	// 超级域名
	superDomain := sessionx.GetSuperDomain()
	appId := "system"
	jobID := "job_" + time.Now().Format("20060102150405")
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "add template",
		Origin:       "-",
		UserId:       userId,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "add-template",
		Steps:        []string{"start", "get-data", "zip-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appId,
	})

	var copyInfos []*backup.CopyInfo

	// -----------------------------------------------APP情报----------------------------------------------------
	// 获取需要备份的app的数据信息
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "アプリデータを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	appService := app.NewAppService("manage", client.DefaultClient)

	var appReq app.FindAppRequest
	appReq.AppId = tplApp
	appReq.Database = tplDb
	appResponse, err := appService.FindApp(context.TODO(), &appReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 保存该数据到文件
	appJSON, err := json.Marshal(appResponse.GetApp())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	if string(appJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "apps")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(appJSON)}, timestamp, "apps")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// 设置copy信息
	copyInfos = append(copyInfos, &backup.CopyInfo{
		CopyType: "apps",
		Source:   appResponse.GetApp().GetAppName(),
		Count:    0,
	})

	// -----------------------------------------------APP下台账情报----------------------------------------------------
	// 获取当前app下的台账
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "データストアデータを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var dsReq datastore.DatastoresRequest
	dsReq.Database = tplDb
	dsReq.AppId = tplApp

	dsResponse, err := datastoreService.FindDatastores(context.TODO(), &dsReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 保存该数据到文件
	dsJSON, err := json.Marshal(dsResponse.GetDatastores())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	if string(dsJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "data_stores")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(dsJSON)}, timestamp, "data_stores")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// -----------------------------------------------APP下打印prints情报----------------------------------------------------
	// 获取当前app下的台账
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "印刷レイアウトデータを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	printService := print.NewPrintService("database", client.DefaultClient)

	var psReq print.FindPrintsRequest
	psReq.Database = tplDb
	psReq.AppId = tplApp

	psResponse, err := printService.FindPrints(context.TODO(), &psReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 保存该数据到文件
	psJSON, err := json.Marshal(psResponse.GetPrints())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	if string(psJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "prints")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(psJSON)}, timestamp, "prints")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// -----------------------------------------------APP下字段情报----------------------------------------------------
	// 获取当前app下的所有字段
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "フィールドデータを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	fieldService := field.NewFieldService("database", client.DefaultClient)

	var fsReq field.AppFieldsRequest
	fsReq.AppId = tplApp
	fsReq.Database = tplDb
	fsReq.InvalidatedIn = "true"

	fsResponse, err := fieldService.FindAppFields(context.TODO(), &fsReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	// 保存该数据到文件
	fsJSON, err := json.Marshal(fsResponse.GetFields())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	if string(fsJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "fields")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(fsJSON)}, timestamp, "fields")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// 设置copy信息
	for _, fs := range fsResponse.GetFields() {
		copyInfos = append(copyInfos, &backup.CopyInfo{
			CopyType: "fields",
			Source:   fs.GetFieldName(),
			Count:    0,
		})
	}

	// -----------------------------------------------APP下多语言情报----------------------------------------------------
	// 获取当前app下的语言数据
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "言語データを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var lgReq language.FindLanguagesRequest
	lgReq.Domain = domain
	lgReq.Database = tplDb

	lgResponse, err := languageService.FindLanguages(context.TODO(), &lgReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 保存该数据到文件
	lgJSON, err := json.Marshal(lgResponse.GetLanguageList())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	if string(lgJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "languages")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(lgJSON)}, timestamp, "languages")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// 设置copy信息
	copyInfos = append(copyInfos, &backup.CopyInfo{
		CopyType: "languages",
		Source:   domain,
		Count:    0,
	})

	// -----------------------------------------------APP下报表情报----------------------------------------------------
	// 获取app下的报表配置
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "レポートデータを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	reportService := report.NewReportService("report", client.DefaultClient)

	var rpReq report.FindReportsRequest
	rpReq.Domain = domain
	rpReq.AppId = tplApp
	rpReq.Database = tplDb

	rpResponse, err := reportService.FindReports(context.TODO(), &rpReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	// 保存该数据到文件
	rpJSON, err := json.Marshal(rpResponse.GetReports())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	if string(rpJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "reports")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(rpJSON)}, timestamp, "reports")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// 设置copy信息
	for _, rp := range rpResponse.GetReports() {
		copyInfos = append(copyInfos, &backup.CopyInfo{
			CopyType: "reports",
			Source:   rp.GetReportName(),
			Count:    0,
		})
	}

	// -----------------------------------------------APP下仪表盘情报----------------------------------------------------
	// 获取app下的仪表盘配置
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "ダッシュボードデータを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)

	var dashReq dashboard.FindDashboardsRequest
	dashReq.Domain = domain
	dashReq.AppId = tplApp
	dashReq.Database = tplDb

	dashResponse, err := dashboardService.FindDashboards(context.TODO(), &dashReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 保存该数据到文件
	dashJSON, err := json.Marshal(dashResponse.GetDashboards())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	if string(dashJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "dashboards")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(dashJSON)}, timestamp, "dashboards")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// 设置copy信息
	for _, dash := range dashResponse.GetDashboards() {
		copyInfos = append(copyInfos, &backup.CopyInfo{
			CopyType: "dashboards",
			Source:   dash.GetDashboardName(),
			Count:    0,
		})
	}

	// -----------------------------------------------APP下选项情报----------------------------------------------------
	// 获取app下的选项配置
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "オプションデータを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var opReq option.FindOptionLabelsRequest
	// 从共通中获取参数
	opReq.Database = tplDb
	opReq.AppId = tplApp

	opResponse, err := optionService.FindOptionLabels(context.TODO(), &opReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 保存该数据到文件
	opJSON, err := json.Marshal(opResponse.GetOptions())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	if string(opJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "options")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(opJSON)}, timestamp, "options")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// 设置copy信息
	for _, op := range opResponse.GetOptions() {
		copyInfos = append(copyInfos, &backup.CopyInfo{
			CopyType: "options",
			Source:   op.GetOptionLabel(),
			Count:    0,
		})
	}

	// -----------------------------------------------APP下権限----------------------------------------------------
	// 获取app下的権限
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "権限データを取得します",
		CurrentStep: "get-data",
		Database:    db,
	}, userId)

	roleService := role.NewRoleService("manage", client.DefaultClient)

	var roReq role.FindRolesRequest
	roReq.RoleName = "SYSTEM"
	roReq.Domain = domain
	roReq.Database = tplDb

	roResponse, err := roleService.FindRoles(context.TODO(), &roReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	var roleID []string
	for _, role := range roResponse.GetRoles() {
		if role.RoleName == "SYSTEM" {
			roleID = append(roleID, role.RoleId)
		}
	}

	permissionService := permission.NewPermissionService("manage", client.DefaultClient)

	var peReq permission.FindActionsRequest
	// 从共通中获取参数
	peReq.Database = tplDb
	peReq.AppId = tplApp
	peReq.RoleId = roleID
	peReq.PermissionType = "app"

	peResponse, err := permissionService.FindActions(context.TODO(), &peReq)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 保存该数据到文件
	peJSON, err := json.Marshal(peResponse.GetActions())
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	if string(peJSON) == "null" {
		_, err := filex.WriteAndSaveLocalFile([]string{}, timestamp, "permissions")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	} else {
		_, err := filex.WriteAndSaveLocalFile([]string{string(peJSON)}, timestamp, "permissions")
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
			// 发送消息-获取数据失败,终止任务
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
			}, userId)
			return
		}
	}

	// 设置copy信息
	for _, pe := range peResponse.GetActions() {
		copyInfos = append(copyInfos, &backup.CopyInfo{
			CopyType: "permissions",
			Source:   pe.GetObjectId(),
			Count:    0,
		})
	}

	// -----------------------------------------------APP下数据情报----------------------------------------------------
	// 判断是否需要包含数据
	if req.GetHasData() {
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "データストアのデータを取得します",
			CurrentStep: "get-data",
			Database:    db,
		}, userId)
		for _, ds := range dsResponse.GetDatastores() {

			total, err := dumpData(tplDb, tplApp, ds.DatastoreId, timestamp)
			if err != nil {
				loggerx.ErrorLog("dump", err.Error())
				path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
				// 发送消息-获取数据失败,终止任务
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
				}, userId)
				return
			}

			copyInfos = append(copyInfos, &backup.CopyInfo{
				CopyType: "data_stores",
				Source:   ds.GetDatastoreName(),
				Count:    total,
			})
		}
	} else {
		// 设置copy信息
		for _, ds := range dsResponse.GetDatastores() {
			copyInfos = append(copyInfos, &backup.CopyInfo{
				CopyType: "data_stores",
				Source:   ds.GetDatastoreName(),
				Count:    0,
			})
		}
	}

	// ----------------------------压缩各情报数据文件&保存压缩后文件到minio服务器-------------------------

	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "圧縮データファイル",
		CurrentStep: "zip-file",
		Database:    db,
	}, userId)

	dir := "backups/" + timestamp
	zipFileName := "backups/backups_" + timestamp + ".zip"
	// 压缩文件
	filex.ZipBackups(dir, zipFileName)

	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "ファイルをストレージに保存",
		CurrentStep: "save-file",
		Database:    db,
	}, userId)

	// 保存压缩文件到minio服务器
	fo, err := os.Open(zipFileName)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	defer fo.Close()
	appRoot := "app_template"
	// 超级域名
	minioClient, err := storagecli.NewClient(superDomain)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}
	filePath := path.Join(appRoot, "zip", "backups_"+timestamp+".zip")
	file, err := minioClient.SavePublicObject(fo, filePath, "application/x-zip-compressed")
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
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
		}, userId)
		return
	}

	// 删除压缩文件
	fo.Close()
	os.RemoveAll(zipFileName)
	os.RemoveAll(dir)

	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     "テンプレートレコードをdbに追加します",
		CurrentStep: "add-record",
		Database:    db,
	}, userId)

	// 设置文件信息
	req.FileName = file.Name
	req.Size = file.Size
	req.FilePath = file.MediaLink

	// 设置copy信息
	req.CopyInfoList = copyInfos

	backupService := backup.NewBackupService("manage", client.DefaultClient)

	_, err = backupService.AddBackup(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		path := filex.WriteAndSaveFile(superDomain, appId, []string{err.Error()})
		// 发送消息-获取数据失败,终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     err.Error(),
			CurrentStep: "add-record",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: db,
		}, userId)
		return
	}

	// 发送消息 写入保存文件成功，返回下载路径，任务结束
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       jobID,
		Message:     i18n.Tr(lang, "job.J_028"),
		CurrentStep: "end",
		EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
		Database:    db,
	}, userId)
}

func dumpData(db, appId, datastoreId, timestamp string) (int64, error) {

	fileName := "items_" + datastoreId

	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	cReq := item.CountRequest{
		AppId:         appId,
		DatastoreId:   datastoreId,
		ConditionType: "and",
		Database:      db,
	}

	cResp, err := itemService.FindCount(context.TODO(), &cReq, opss)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		return 0, err
	}

	// 每次2000为一组数据
	total := float64(cResp.GetTotal())
	count := math.Ceil(total / 2000)

	dir := "backups/" + timestamp + "/"

	filex.Mkdir(dir)

	filePath := dir + fileName + ".txt"

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		loggerx.ErrorLog("dump", err.Error())
		return 0, err
	}
	//及时关闭file句柄
	defer file.Close()

	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)

LP:
	for i := int64(0); i < int64(count); i++ {

		// 从path中获取参数
		var req item.ItemsRequest
		req.DatastoreId = datastoreId
		req.AppId = appId
		req.PageIndex = i + 1
		req.PageSize = 2000
		req.Database = db
		req.IsOrigin = true

		response, err := itemService.FindItems(context.TODO(), &req, opss)
		if err != nil {
			loggerx.ErrorLog("dump", err.Error())
			return 0, err
		}

		if len(response.GetItems()) == 0 {
			break LP
		}

		for j, dt := range response.GetItems() {
			data, err := json.Marshal(dt)
			if err != nil {
				loggerx.ErrorLog("dump", err.Error())
				return 0, err
			}

			write.WriteString(string(data))

			if !(i == int64(count)-1 && j == int(response.GetTotal()-1)) {
				write.WriteString("\n")
			}
		}
		write.Flush()
	}

	file.Close()

	return cResp.GetTotal(), nil
}
