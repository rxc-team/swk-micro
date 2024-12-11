package journalx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	merrors "github.com/micro/go-micro/v2/errors"
	"github.com/yidane/formula"
	"go.mongodb.org/mongo-driver/mongo"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/configx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/global/proto/sequence"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/pit3/srv/journal/proto/subject"
	"rxcsoft.cn/pit3/srv/task/proto/task"
)

type ItemData []map[string]*item.Value

type ImportData []*item.ListItems

type SubData map[string]string

type InsertParam struct {
	db          string
	jobID       string
	domain      string
	lang        string
	shiwakeno   string
	handleMonth string
	appID       string
	datastoreID string
	userID      string
	owners      []string
	dsMap       map[string]string
	jouData     *journal.Journal
	asSubMap    map[string]SubData
	jouDataMap  map[string]*journal.Journal
}

func genSequenceKey(app string) string {
	seq := strings.Builder{}
	seq.WriteString(app)
	seq.WriteString("_")
	seq.WriteString("shiwakeno")

	return seq.String()
}

func getSeq(db, key string) (int64, error) {
	seqService := sequence.NewSequenceService("global", client.DefaultClient)

	var req sequence.FindSequenceRequest
	req.SequenceKey = key
	req.Database = db

	response, err := seqService.FindSequence(context.TODO(), &req)
	if err != nil {
		er := merrors.Parse(err.Error())
		if er.GetDetail() == mongo.ErrNoDocuments.Error() {

			// 创建seq
			var add sequence.AddRequest
			add.SequenceKey = key
			add.StartValue = 1
			add.Database = db

			resp, err := seqService.AddSequence(context.TODO(), &add)
			if err != nil {
				return 0, err
			}

			return resp.GetSequence(), nil
		}
		return 0, err
	}

	return response.GetSequence(), nil
}

func genShiwakeno(db, app string) (string, error) {
	key := genSequenceKey(app)

	seq, err := getSeq(db, key)
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%013d", seq), nil

}

// getDatastoreMap 获取台账apikey和datastore_id的map
func getDatastoreMap(db, appID string) (dsMap map[string]string, err error) {
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoresRequest
	// 从共通获取
	req.Database = db
	req.AppId = appID

	response, err := datastoreService.FindDatastores(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getDatastoreMap", err.Error())
		return
	}

	dsMap = make(map[string]string)

	for _, ds := range response.GetDatastores() {
		dsMap[ds.ApiKey] = ds.GetDatastoreId()
	}

	return
}

// buildLeaseHistory 获取所有履历数据（根据对接区分查询）
func buildLeaseHistory(p InsertParam) (err error) {

	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}
	conditions := []*item.Condition{}
	conditions = append(conditions, &item.Condition{
		FieldId:     "dockkbn",
		FieldType:   "options",
		SearchValue: "undo",
		Operator:    "=",
		IsDynamic:   true,
	})

	accesskeys := sessionx.GetAccessKeys(p.db, p.userID, p.datastoreID, "R")

	// 先获取总的件数
	cReq := item.CountRequest{
		AppId:         p.appID,
		DatastoreId:   p.datastoreID,
		ConditionList: conditions,
		ConditionType: "and",
		Owners:        accesskeys,
		Database:      p.db,
	}

	countResponse, err := itemService.FindCount(context.TODO(), &cReq, opss)
	if err != nil {
		loggerx.ErrorLog("getLeaseHistory", err.Error())
		return
	}

	// 根据总的件数分批下载数据
	// 每次2000为一组数据
	total := float64(countResponse.GetTotal())
	count := math.Ceil(total / 500)

	for index := int64(0); index < int64(count); index++ {

		var req item.ItemsRequest
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "no",
			SortValue: "ascend",
		})
		sorts = append(sorts, &item.SortItem{
			SortKey:   "zengokbn",
			SortValue: "ascend",
		})
		req.Sorts = sorts
		req.ConditionList = conditions
		req.ConditionType = "and"
		req.DatastoreId = p.datastoreID
		req.PageIndex = index + 1
		req.PageSize = 500
		req.AppId = p.appID
		req.Owners = accesskeys
		req.IsOrigin = true
		req.Database = p.db

		response, err := itemService.FindItems(context.TODO(), &req, opss)
		if err != nil {
			loggerx.ErrorLog("getLeaseHistory", err.Error())
			return err
		}

		items := response.GetItems()

		var hsNo string

		var data ItemData
		itemList := make(map[string]ItemData)

		for index, item := range items {
			itemMap := item.Items
			// 获取履历番号
			no := itemMap["no"].Value
			// 第一条时
			if index == 0 {
				// 前次履历番号设定
				hsNo = no
				// 当前数据导入
				data = append(data, itemMap)
				// 如果当前数据就是最后一条的场合
				if len(items) == 1 {
					itemList[no] = data
				}
				// 继续循环
				continue
			}

			// 如果是最后一条的场合
			if index == (len(items) - 1) {
				// 同一条契约的数据
				if hsNo == no {
					// 当前数据导入
					data = append(data, itemMap)
					// 数据导入map中
					itemList[no] = data
				} else {
					// 之前的数据导入map中
					itemList[hsNo] = data
					// 清除data的数据
					data = nil
					// 当前数据导入
					data = append(data, itemMap)
					// 数据导入map中
					itemList[no] = data
				}
				continue
			}

			// 其他条的数据
			// 如果上一条履历的番号和当前的番号相同的场合，则继续加入数据
			if hsNo == no {
				// 前次履历番号设定
				hsNo = no
				// 当前数据导入
				data = append(data, itemMap)
				// 继续循环
				continue
			}

			// 如果上一条履历的番号和当前的番号不相同的场合
			// 之前的数据导入map中
			itemList[hsNo] = data
			// 清除data的数据
			data = nil
			// 当前数据导入
			data = append(data, itemMap)
			// 前次履历番号设定
			hsNo = no
		}

		its, err := genShiwakeData(p, itemList)
		if err != nil {
			loggerx.ErrorLog("getLeaseHistory", err.Error())
			return err
		}

		result, err := importData(p, its)
		if err != nil {
			loggerx.ErrorLog("getLeaseHistory", err.Error())
			return err
		}

		loggerx.DebugLog("getLeaseHistory", fmt.Sprintf("result %v", result))
	}

	return nil
}

// getJournal 获取分录数据
func getJournal(db, appID, journalID string) (j *journal.Journal, err error) {
	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.JournalRequest
	req.JournalId = journalID
	req.AppId = appID
	req.Database = db
	response, err := journalService.FindJournal(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getJournal", err.Error())
		return
	}

	j = response.GetJournal()

	return
}

// 获取科目的数据
func getSubjectMap(db, appID, datastoreID string, accesskeys []string) (asSubMap map[string]SubData, err error) {
	// 获取默认的科目
	subjectService := subject.NewSubjectService("journal", client.DefaultClient)

	var req subject.SubjectsRequest
	// 从query获取
	req.Database = db
	req.AppId = appID

	response, err := subjectService.FindSubjects(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getSubjectMap", err.Error())
		return
	}

	defSubjects := response.GetSubjects()

	defSubMap := make(SubData)
	for _, sub := range defSubjects {
		if len(sub.SubjectName) > 0 {
			defSubMap[sub.SubjectKey] = sub.SubjectName
		} else {
			defSubMap[sub.SubjectKey] = sub.DefaultName
		}
	}

	assets, err := getAssetsList(db, appID, datastoreID, accesskeys)
	if err != nil {
		loggerx.ErrorLog("getSubjectMap", err.Error())
		return
	}

	type AssetsSubject struct {
		Subjects SubData
		AsType   string
		Error    error
	}

	subChan := make(chan AssetsSubject, len(assets))

	for _, asType := range assets {
		go func(def SubData, aType string, ch chan AssetsSubject) {
			var req subject.SubjectsRequest
			// 从query获取
			req.AssetsType = aType
			req.Database = db
			req.AppId = appID

			response, err := subjectService.FindSubjects(context.TODO(), &req)
			if err != nil {
				loggerx.ErrorLog("getSubjectMap", err.Error())
				ch <- AssetsSubject{
					Error: err,
				}
				return
			}

			subjects := response.GetSubjects()

			subMap := make(SubData)

			if len(subjects) > 0 {
				for _, sub := range subjects {
					if len(sub.SubjectName) == 0 {
						if val, ok := def[sub.SubjectKey]; ok && len(val) > 0 {
							subMap[sub.SubjectKey] = val
						} else {
							ch <- AssetsSubject{
								Error: fmt.Errorf("default subject not found value, key:[%s], defaultName:[%s]", sub.SubjectKey, sub.DefaultName),
							}
							return
						}
					} else {
						subMap[sub.SubjectKey] = sub.SubjectName
					}
				}
			} else {
				subMap = defSubMap
			}

			ch <- AssetsSubject{
				Subjects: subMap,
				AsType:   aType,
			}
		}(defSubMap, asType, subChan)
	}

	asSubMap = make(map[string]SubData)
	asSubMap[""] = defSubMap
	count := 1
	for ch := range subChan {
		if ch.Error != nil {
			err = ch.Error
			loggerx.ErrorLog("getSubjectMap", ch.Error.Error())
			return
		}

		asSubMap[ch.AsType] = ch.Subjects
		if count == len(assets) {
			close(subChan)
		}
		count++
	}

	return
}

// 获取资产类型数据
func getAssetsList(db, appID, datastoreID string, accesskeys []string) (assets []string, err error) {
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	var req item.ItemsRequest
	var sorts []*item.SortItem
	sorts = append(sorts, &item.SortItem{
		SortKey:   "assets_class_id",
		SortValue: "ascend",
	})
	req.Sorts = sorts
	req.ConditionType = "and"
	req.DatastoreId = datastoreID
	req.AppId = appID
	req.Owners = accesskeys
	req.Database = db
	req.IsOrigin = true

	response, err := itemService.FindItems(context.TODO(), &req, opss)
	if err != nil {
		loggerx.ErrorLog("getAssetsList", err.Error())
		return
	}

	for _, item := range response.GetItems() {
		itemMap := item.Items
		// 获取履历番号
		assetsID := itemMap["assets_class_id"].Value
		assets = append(assets, assetsID)
	}

	return
}

// GenAddAndSubData 生成增减分录的数据
func GenAddAndSubData(domain, db, appID, userID, lang string, owners []string) (r *item.ImportResult, err error) {
	jobID := "job_" + time.Now().Format("20060102150405")
	//获取当前的language
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "Inc And Dec Journa",
		Origin:       "-",
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "journal",
		Steps:        []string{"start", "collect-data", "delete-old-data", "gen-data", "modify_status", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	go func() {
		// 发送消息 收集数据情报
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_002"),
			CurrentStep: "collect-data",
			Database:    db,
		}, userID)

		// 获取台账map
		dsMap, err := getDatastoreMap(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 收集数据情报失败 终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取处理月度
		cfg, err := configx.GetConfigVal(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 收集数据情报失败 终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}
		handleMonth := cfg.GetSyoriYm()

		// 获取所有分录数据
		jouDataMap := make(map[string]*journal.Journal)

		journalService := journal.NewJournalService("journal", client.DefaultClient)

		var journalReq journal.JournalsRequest
		journalReq.Database = db
		journalReq.AppId = appID

		journalResponse, err := journalService.FindJournals(context.TODO(), &journalReq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		for _, journal := range journalResponse.Journals {
			jouDataMap[journal.JournalId] = journal
		}

		// 获取所有分类的科目的数据
		assetDs := dsMap["assets"]
		assetAccesskeys := sessionx.GetAccessKeys(db, userID, assetDs, "R")
		asSubMap, err := getSubjectMap(db, appID, assetDs, assetAccesskeys)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取分录番号
		shiwakeno, err := genShiwakeno(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 删除旧数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_017"),
			CurrentStep: "delete-old-data",
			Database:    db,
		}, userID)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		itemService := item.NewItemService("database", client.DefaultClient)

		// 删除之前的分录数据
		var delreq item.DeleteItemsRequest
		delreq.DatastoreId = dsMap["shiwake"]
		delreq.AppId = appID
		delreq.UserId = userID
		delreq.Database = db
		delreq.ConditionType = "and"

		defaultTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "shiwaketype",
			FieldType:     "text",
			SearchValue:   "1",
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		conditions = append(conditions, &item.Condition{
			FieldId:       "kakuteidate",
			FieldType:     "date",
			SearchValue:   defaultTime.Format(time.RFC3339),
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		delreq.ConditionList = conditions

		_, err = itemService.DeleteItems(context.TODO(), &delreq, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "delete-old-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 数据上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_020"),
			CurrentStep: "gen-data",
			Database:    db,
		}, userID)

		// 通过增减履历数据生成分录data
		param := InsertParam{
			db:          db,
			jobID:       jobID,
			domain:      domain,
			lang:        lang,
			shiwakeno:   shiwakeno,
			handleMonth: handleMonth,
			appID:       appID,
			datastoreID: dsMap["zougenrireki"],
			userID:      userID,
			owners:      owners,
			dsMap:       dsMap,
			asSubMap:    asSubMap,
			jouDataMap:  jouDataMap,
		}

		//  生成数据
		err = buildObtainData(param)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "gen-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 任务成功结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)

	}()

	return r, nil
}

// buildObtainData 生成增减履历分录数据
func buildObtainData(p InsertParam) (e error) {
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	handleDate, err := time.Parse("2006-01", p.handleMonth)
	if err != nil {
		loggerx.ErrorLog("getObtainData", err.Error())
		return err
	}

	lastDay := getMonthLastDay(handleDate)
	defaultTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

	conditions := []*item.Condition{}
	conditions = append(conditions, &item.Condition{
		FieldId:     "keijoudate",
		FieldType:   "date",
		SearchValue: p.handleMonth + "-01",
		Operator:    ">=",
		IsDynamic:   true,
	})

	conditions = append(conditions, &item.Condition{
		FieldId:     "keijoudate",
		FieldType:   "date",
		SearchValue: p.handleMonth + "-" + lastDay,
		Operator:    "<=",
		IsDynamic:   true,
	})

	conditions = append(conditions, &item.Condition{
		FieldId:     "kakuteidate",
		FieldType:   "date",
		SearchValue: defaultTime.Format(time.RFC3339),
		Operator:    "=",
		IsDynamic:   true,
	})

	accesskeys := sessionx.GetAccessKeys(p.db, p.userID, p.datastoreID, "R")

	// 先获取总的件数
	cReq := item.CountRequest{
		AppId:         p.appID,
		DatastoreId:   p.datastoreID,
		ConditionList: conditions,
		ConditionType: "and",
		Owners:        accesskeys,
		Database:      p.db,
	}

	countResponse, err := itemService.FindCount(context.TODO(), &cReq, opss)
	if err != nil {
		loggerx.ErrorLog("getObtainData", err.Error())
		return err
	}

	// 根据总的件数分批下载数据
	// 每次2000为一组数据
	total := float64(countResponse.GetTotal())
	count := math.Ceil(total / 500)

	for index := int64(0); index < int64(count); index++ {

		var req item.ItemsRequest
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "shisanbangouoya.value",
			SortValue: "ascend",
		})
		sorts = append(sorts, &item.SortItem{
			SortKey:   "shisanbangoueda.value",
			SortValue: "ascend",
		})
		req.Sorts = sorts
		req.ConditionList = conditions
		req.ConditionType = "and"
		req.DatastoreId = p.datastoreID
		req.PageIndex = index + 1
		req.PageSize = 500
		req.AppId = p.appID
		req.Owners = accesskeys
		req.Database = p.db
		req.IsOrigin = true

		itemResp, err := itemService.FindItems(context.TODO(), &req, opss)
		if err != nil {
			loggerx.ErrorLog("getObtainData", err.Error())
			return err
		}

		// 分录数据编辑
		var items ImportData
		index := 1
		for count, obtainItem := range itemResp.GetItems() {
			var pattern *journal.Pattern
			if obtainItem.Items["setteikubun"].GetValue() == "固定資産取得" {
				pattern = getPattern("04001", p.jouDataMap["04"])
			}
			if obtainItem.Items["setteikubun"].GetValue() == "固定資産移動" {
				pattern = getPattern("05001", p.jouDataMap["05"])
			}
			branchCount := 1
			for line, sub := range pattern.GetSubjects() {
				expression := formula.NewExpression(sub.AmountField)
				params := getParam(sub.AmountField)
				for _, pm := range params {
					it, ok := obtainItem.Items[pm]
					if !ok {
						it = &item.Value{
							DataType: "number",
							Value:    "0",
						}
					}
					val, err := strconv.ParseFloat(it.GetValue(), 64)
					if err != nil {
						loggerx.ErrorLog("getObtainData", err.Error())
						return err
					}
					expression.AddParameter(pm, val)
				}

				result, err := expression.Evaluate()
				if err != nil {
					loggerx.ErrorLog("getObtainData", err.Error())
					return err
				}

				fv, err := result.Float64()
				if err != nil {
					loggerx.ErrorLog("getObtainData", err.Error())
					return err
				}

				if fv == 0.0 {
					continue
				}

				assetsType := obtainItem.Items["bunruicd"].GetValue()
				subMap := p.asSubMap[assetsType]

				// 创建登录数据
				itemsData := copyMap(obtainItem.Items)

				itemsData["keiyakuno"] = &item.Value{
					DataType: "lookup",
					Value:    obtainItem.Items["kaishacd"].GetValue(),
				}
				itemsData["koushinbangouoya"] = &item.Value{
					DataType: "text",
					Value:    obtainItem.Items["koushinbangouoya"].GetValue(),
				}
				itemsData["koushinbangoueda"] = &item.Value{
					DataType: "text",
					Value:    obtainItem.Items["koushinbangoueda"].GetValue(),
				}
				itemsData["shiwakeno"] = &item.Value{
					DataType: "text",
					Value:    p.shiwakeno,
				}
				itemsData["shiwakeymd"] = &item.Value{
					DataType: "date",
					Value:    time.Now().Format("2006-01-02"),
				}
				itemsData["shiwakeym"] = &item.Value{
					DataType: "text",
					Value:    p.handleMonth,
				}
				itemsData["partten"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternId,
				}
				itemsData["lineno"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(line + 1),
				}
				itemsData["taishakukubun"] = &item.Value{
					DataType: "text",
					Value:    sub.LendingDivision,
				}
				itemsData["kanjokamoku"] = &item.Value{
					DataType: "text",
					Value:    subMap[sub.GetSubjectKey()],
				}
				itemsData["shiwakekingaku"] = &item.Value{
					DataType: "number",
					Value:    result.String(),
				}
				itemsData["shiwakeaggno_parent"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(count + 1),
				}
				itemsData["shiwakeaggno_branch"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(branchCount),
				}
				itemsData["shiwaketype"] = &item.Value{
					DataType: "text",
					Value:    "1",
				}
				itemsData["remark"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternName,
				}
				itemsData["index"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(index),
				}

				its := &item.ListItems{
					Items: itemsData,
				}

				items = append(items, its)

				index++
				branchCount++
			}
		}

		response, err := importData(p, items)
		if err != nil {
			loggerx.ErrorLog("getObtainData", err.Error())
			return err
		}
		fmt.Printf("%+v", response)
	}

	// 发送消息 数据状态修改
	jobx.ModifyTask(task.ModifyRequest{
		JobId:       p.jobID,
		Message:     "仕訳状態更新",
		CurrentStep: "modify_status",
		Database:    p.db,
	}, p.userID)

	var req item.JournalRequest
	req.DatastoreId = p.datastoreID
	req.Database = p.db
	req.StartDate = p.handleMonth + "-01"
	req.LastDate = p.handleMonth + "-" + lastDay

	_, err = itemService.GenerateItem(context.TODO(), &req, opss)
	if err != nil {
		path := filex.WriteAndSaveFile(p.domain, p.appID, []string{err.Error()})
		// 发送消息 获取数据失败，终止任务
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       p.jobID,
			Message:     err.Error(),
			CurrentStep: "modify_status",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			ErrorFile: &task.File{
				Url:  path.MediaLink,
				Name: path.Name,
			},
			Database: p.db,
		}, p.userID)
		return
	}

	return nil
}

func genShiwakeData(p InsertParam, hsData map[string]ItemData) (it ImportData, e error) {
	// 分录数据
	var items ImportData
	index := 1
	count := 1
	for no, hs := range hsData {
		// 契约登录的场合
		if len(hs) == 1 {
			pattern := getPattern("01001", p.jouData)
			itemMap := hs[0]
			itemMap["historyno"] = &item.Value{
				DataType: "text",
				Value:    no,
			}
			branchCount := 1
			for line, sub := range pattern.GetSubjects() {
				expression := formula.NewExpression(sub.AmountField)
				params := getParam(sub.AmountField)
				for _, pm := range params {
					it, ok := itemMap[pm]
					if !ok {
						it = &item.Value{
							DataType: "number",
							Value:    "0",
						}
					}
					val, err := strconv.ParseFloat(it.GetValue(), 64)
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}
					expression.AddParameter(pm, val)
				}

				result, err := expression.Evaluate()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				fv, err := result.Float64()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				if fv == 0.0 {
					continue
				}

				keiyakuno := itemMap["keiyakuno"].GetValue()
				assetsType := itemMap["bunruicd"].GetValue()
				subMap := p.asSubMap[assetsType]

				// 创建登录数据
				itemsData := copyMap(itemMap)
				itemsData["shiwakeno"] = &item.Value{
					DataType: "text",
					Value:    p.shiwakeno,
				}
				itemsData["shiwakeymd"] = &item.Value{
					DataType: "date",
					Value:    time.Now().Format("2006-01-02"),
				}
				itemsData["shiwakeym"] = &item.Value{
					DataType: "text",
					Value:    p.handleMonth,
				}
				itemsData["partten"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternId,
				}
				itemsData["lineno"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(line + 1),
				}
				itemsData["taishakukubun"] = &item.Value{
					DataType: "text",
					Value:    sub.LendingDivision,
				}
				itemsData["kanjokamoku"] = &item.Value{
					DataType: "text",
					Value:    subMap[sub.GetSubjectKey()],
				}
				itemsData["shiwakekingaku"] = &item.Value{
					DataType: "number",
					Value:    result.String(),
				}
				itemsData["shiwakeaggno_parent"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(count),
				}
				itemsData["shiwakeaggno_branch"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(branchCount),
				}
				itemsData["shiwaketype"] = &item.Value{
					DataType: "text",
					Value:    "1",
				}
				itemsData["remark"] = &item.Value{
					DataType: "text",
					Value:    keiyakuno + "_" + pattern.PatternName,
				}
				itemsData["index"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(index),
				}

				its := &item.ListItems{
					Items: itemsData,
				}

				items = append(items, its)

				index++
				branchCount++
			}

			count++
			continue
		}
		// 变更后履历
		var newItemMap map[string]*item.Value
		// 变更前履历
		var oldItemMap map[string]*item.Value
		for _, h := range hs {
			zengokbn := h["zengokbn"].GetValue()
			if zengokbn == "before" {
				oldItemMap = h
			} else {
				newItemMap = h
			}
		}

		// 获取操作区分
		actkbn := newItemMap["actkbn"].GetValue()

		// 如果操作是情报变更的场合
		if actkbn == "infoalter" {
			// 获取分类
			oldBunrui := oldItemMap["bunruicd"].GetValue()
			newBunrui := newItemMap["bunruicd"].GetValue()

			// 获取部门
			oldSegment := oldItemMap["segmentcd"].GetValue()
			newSegment := newItemMap["segmentcd"].GetValue()
			// 当部门或者分类没有发生变更的情况，不出力分录数据
			if oldBunrui == newBunrui && oldSegment == newSegment {
				continue
			}

			pattern := getPattern("01002", p.jouData)

			var itemMap map[string]*item.Value

			branchCount := 1
			for line, sub := range pattern.GetSubjects() {
				// LendingDivision = "1" == 借方
				// ChangeFlag = "new" == 使用变更后的数据newItemMap
				// LendingDivision = "2" == 贷方
				// ChangeFlag = "old" ==  使用变更前的数据oldItemMap
				if sub.ChangeFlag == "old" {
					itemMap = copyMap(oldItemMap)
				} else {
					// 借方 => 变更后
					itemMap = copyMap(newItemMap)
				}
				itemMap["historyno"] = &item.Value{
					DataType: "text",
					Value:    no,
				}

				expression := formula.NewExpression(sub.AmountField)

				params := getParam(sub.AmountField)
				for _, pm := range params {
					it, ok := newItemMap[pm]
					if !ok {
						it = &item.Value{
							DataType: "number",
							Value:    "0",
						}
					}
					val, err := strconv.ParseFloat(it.GetValue(), 64)
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}
					expression.AddParameter(pm, val)
				}

				result, err := expression.Evaluate()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				fv, err := result.Float64()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				if fv == 0.0 {
					continue
				}

				keiyakuno := itemMap["keiyakuno"].GetValue()
				assetsType := itemMap["bunruicd"].GetValue()
				subMap := p.asSubMap[assetsType]

				// 创建登录数据
				itemsData := copyMap(itemMap)
				itemsData["shiwakeno"] = &item.Value{
					DataType: "text",
					Value:    p.shiwakeno,
				}
				itemsData["shiwakeymd"] = &item.Value{
					DataType: "date",
					Value:    time.Now().Format("2006-01-02"),
				}
				itemsData["shiwakeym"] = &item.Value{
					DataType: "text",
					Value:    p.handleMonth,
				}
				itemsData["partten"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternId,
				}
				itemsData["lineno"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(line + 1),
				}
				itemsData["taishakukubun"] = &item.Value{
					DataType: "text",
					Value:    sub.LendingDivision,
				}
				itemsData["kanjokamoku"] = &item.Value{
					DataType: "text",
					Value:    subMap[sub.GetSubjectKey()],
				}
				itemsData["shiwakekingaku"] = &item.Value{
					DataType: "number",
					Value:    result.String(),
				}
				itemsData["shiwakeaggno_parent"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(count),
				}
				itemsData["shiwakeaggno_branch"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(branchCount),
				}
				itemsData["shiwaketype"] = &item.Value{
					DataType: "text",
					Value:    "1",
				}
				itemsData["remark"] = &item.Value{
					DataType: "text",
					Value:    keiyakuno + "_" + pattern.PatternName,
				}
				itemsData["index"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(index),
				}

				its := &item.ListItems{
					Items: itemsData,
				}

				items = append(items, its)

				index++
				branchCount++
			}
			count++
			continue
		}
		// 如果操作是债务变更的场合
		if actkbn == "debtchange" {
			kaiyakuymd := newItemMap["kaiyakuymd"].GetValue()
			// 未来解约的场合
			if len(kaiyakuymd) > 0 && kaiyakuymd != "0001-01-01" {
				cancellationrightoption := newItemMap["cancellationrightoption"].GetValue()
				// 如果行使权存在的情况下，按比例减少处理。
				if cancellationrightoption == "true" {
					// 比例减少的场合
					pattern := getPattern("01006", p.jouData)

					var itemMap map[string]*item.Value

					branchCount := 1
					for line, sub := range pattern.GetSubjects() {
						// LendingDivision = "1" == 借方
						// ChangeFlag = "new" == 使用变更后的数据newItemMap
						// LendingDivision = "2" == 贷方
						// ChangeFlag = "old" ==  使用变更前的数据oldItemMap
						if sub.ChangeFlag == "old" {
							itemMap = copyMap(oldItemMap)
						} else {
							// 借方 => 变更后
							itemMap = copyMap(newItemMap)
						}
						itemMap["historyno"] = &item.Value{
							DataType: "text",
							Value:    no,
						}

						expression := formula.NewExpression(sub.AmountField)

						params := getParam(sub.AmountField)
						for _, pm := range params {
							it, ok := newItemMap[pm]
							if !ok {
								it = &item.Value{
									DataType: "number",
									Value:    "0",
								}
							}
							val, err := strconv.ParseFloat(it.GetValue(), 64)
							if err != nil {
								loggerx.ErrorLog("insertData", err.Error())
								return nil, err
							}
							expression.AddParameter(pm, val)
						}

						result, err := expression.Evaluate()
						if err != nil {
							loggerx.ErrorLog("insertData", err.Error())
							return nil, err
						}

						fv, err := result.Float64()
						if err != nil {
							loggerx.ErrorLog("insertData", err.Error())
							return nil, err
						}

						if fv == 0.0 {
							continue
						}

						keiyakuno := itemMap["keiyakuno"].GetValue()
						assetsType := itemMap["bunruicd"].GetValue()
						subMap := p.asSubMap[assetsType]

						// 创建登录数据
						itemsData := copyMap(itemMap)
						itemsData["shiwakeno"] = &item.Value{
							DataType: "text",
							Value:    p.shiwakeno,
						}
						itemsData["shiwakeymd"] = &item.Value{
							DataType: "date",
							Value:    time.Now().Format("2006-01-02"),
						}
						itemsData["shiwakeym"] = &item.Value{
							DataType: "text",
							Value:    p.handleMonth,
						}
						itemsData["partten"] = &item.Value{
							DataType: "text",
							Value:    pattern.PatternId,
						}
						itemsData["lineno"] = &item.Value{
							DataType: "number",
							Value:    strconv.Itoa(line + 1),
						}
						itemsData["taishakukubun"] = &item.Value{
							DataType: "text",
							Value:    sub.LendingDivision,
						}
						itemsData["kanjokamoku"] = &item.Value{
							DataType: "text",
							Value:    subMap[sub.GetSubjectKey()],
						}
						itemsData["shiwakekingaku"] = &item.Value{
							DataType: "number",
							Value:    result.String(),
						}
						itemsData["shiwakeaggno_parent"] = &item.Value{
							DataType: "text",
							Value:    strconv.Itoa(count),
						}
						itemsData["shiwakeaggno_branch"] = &item.Value{
							DataType: "text",
							Value:    strconv.Itoa(branchCount),
						}
						itemsData["shiwaketype"] = &item.Value{
							DataType: "text",
							Value:    "1",
						}
						itemsData["remark"] = &item.Value{
							DataType: "text",
							Value:    keiyakuno + "_" + pattern.PatternName,
						}
						itemsData["index"] = &item.Value{
							DataType: "number",
							Value:    strconv.Itoa(index),
						}

						its := &item.ListItems{
							Items: itemsData,
						}

						items = append(items, its)

						index++
						branchCount++
					}
					count++
					continue
				}

				// 没有行使权的场合，按照普通债务变更处理。
				pattern := getPattern("01007", p.jouData)

				var itemMap map[string]*item.Value

				branchCount := 1
				for line, sub := range pattern.GetSubjects() {
					// LendingDivision = "1" == 借方
					// ChangeFlag = "new" == 使用变更后的数据newItemMap
					// LendingDivision = "2" == 贷方
					// ChangeFlag = "old" ==  使用变更前的数据oldItemMap
					if sub.ChangeFlag == "old" {
						itemMap = copyMap(oldItemMap)
					} else {
						// 借方 => 变更后
						itemMap = copyMap(newItemMap)
					}
					itemMap["historyno"] = &item.Value{
						DataType: "text",
						Value:    no,
					}

					expression := formula.NewExpression(sub.AmountField)

					params := getParam(sub.AmountField)
					for _, pm := range params {
						it, ok := newItemMap[pm]
						if !ok {
							it = &item.Value{
								DataType: "number",
								Value:    "0",
							}
						}
						val, err := strconv.ParseFloat(it.GetValue(), 64)
						if err != nil {
							loggerx.ErrorLog("insertData", err.Error())
							return nil, err
						}
						expression.AddParameter(pm, val)
					}

					result, err := expression.Evaluate()
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}

					fv, err := result.Float64()
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}

					if fv == 0.0 {
						continue
					}

					keiyakuno := itemMap["keiyakuno"].GetValue()
					assetsType := itemMap["bunruicd"].GetValue()
					subMap := p.asSubMap[assetsType]

					// 创建登录数据
					itemsData := copyMap(itemMap)
					itemsData["shiwakeno"] = &item.Value{
						DataType: "text",
						Value:    p.shiwakeno,
					}
					itemsData["shiwakeymd"] = &item.Value{
						DataType: "date",
						Value:    time.Now().Format("2006-01-02"),
					}
					itemsData["shiwakeym"] = &item.Value{
						DataType: "text",
						Value:    p.handleMonth,
					}
					itemsData["partten"] = &item.Value{
						DataType: "text",
						Value:    pattern.PatternId,
					}
					itemsData["lineno"] = &item.Value{
						DataType: "number",
						Value:    strconv.Itoa(line + 1),
					}
					itemsData["taishakukubun"] = &item.Value{
						DataType: "text",
						Value:    sub.LendingDivision,
					}
					itemsData["kanjokamoku"] = &item.Value{
						DataType: "text",
						Value:    subMap[sub.GetSubjectKey()],
					}
					itemsData["shiwakekingaku"] = &item.Value{
						DataType: "number",
						Value:    result.String(),
					}
					itemsData["shiwakeaggno_parent"] = &item.Value{
						DataType: "text",
						Value:    strconv.Itoa(count),
					}
					itemsData["shiwakeaggno_branch"] = &item.Value{
						DataType: "text",
						Value:    strconv.Itoa(branchCount),
					}
					itemsData["shiwaketype"] = &item.Value{
						DataType: "text",
						Value:    "1",
					}
					itemsData["remark"] = &item.Value{
						DataType: "text",
						Value:    keiyakuno + "_" + pattern.PatternName,
					}
					itemsData["index"] = &item.Value{
						DataType: "number",
						Value:    strconv.Itoa(index),
					}

					its := &item.ListItems{
						Items: itemsData,
					}

					items = append(items, its)

					index++
					branchCount++
				}
				count++
				continue

			}

			// 获取比率
			oldPercentage := oldItemMap["percentage"].GetValue()
			newPercentage := newItemMap["percentage"].GetValue()

			// 比率未发生变更的场合，即普通的债务变更
			if oldPercentage == newPercentage {

				pattern := getPattern("01003", p.jouData)

				var itemMap map[string]*item.Value

				branchCount := 1
				for line, sub := range pattern.GetSubjects() {
					// LendingDivision = "1" == 借方
					// ChangeFlag = "new" == 使用变更后的数据newItemMap
					// LendingDivision = "2" == 贷方
					// ChangeFlag = "old" ==  使用变更前的数据oldItemMap
					if sub.ChangeFlag == "old" {
						itemMap = copyMap(oldItemMap)
					} else {
						// 借方 => 变更后
						itemMap = copyMap(newItemMap)
					}
					itemMap["historyno"] = &item.Value{
						DataType: "text",
						Value:    no,
					}

					expression := formula.NewExpression(sub.AmountField)

					params := getParam(sub.AmountField)
					for _, pm := range params {
						it, ok := newItemMap[pm]
						if !ok {
							it = &item.Value{
								DataType: "number",
								Value:    "0",
							}
						}
						val, err := strconv.ParseFloat(it.GetValue(), 64)
						if err != nil {
							loggerx.ErrorLog("insertData", err.Error())
							return nil, err
						}
						expression.AddParameter(pm, val)
					}

					result, err := expression.Evaluate()
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}

					fv, err := result.Float64()
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}

					if fv == 0.0 {
						continue
					}

					keiyakuno := itemMap["keiyakuno"].GetValue()
					assetsType := itemMap["bunruicd"].GetValue()
					subMap := p.asSubMap[assetsType]

					// 创建登录数据
					itemsData := copyMap(itemMap)
					itemsData["shiwakeno"] = &item.Value{
						DataType: "text",
						Value:    p.shiwakeno,
					}
					itemsData["shiwakeymd"] = &item.Value{
						DataType: "date",
						Value:    time.Now().Format("2006-01-02"),
					}
					itemsData["shiwakeym"] = &item.Value{
						DataType: "text",
						Value:    p.handleMonth,
					}
					itemsData["partten"] = &item.Value{
						DataType: "text",
						Value:    pattern.PatternId,
					}
					itemsData["lineno"] = &item.Value{
						DataType: "number",
						Value:    strconv.Itoa(line + 1),
					}
					itemsData["taishakukubun"] = &item.Value{
						DataType: "text",
						Value:    sub.LendingDivision,
					}
					itemsData["kanjokamoku"] = &item.Value{
						DataType: "text",
						Value:    subMap[sub.GetSubjectKey()],
					}
					itemsData["shiwakekingaku"] = &item.Value{
						DataType: "number",
						Value:    result.String(),
					}
					itemsData["shiwakeaggno_parent"] = &item.Value{
						DataType: "text",
						Value:    strconv.Itoa(count),
					}
					itemsData["shiwakeaggno_branch"] = &item.Value{
						DataType: "text",
						Value:    strconv.Itoa(branchCount),
					}
					itemsData["shiwaketype"] = &item.Value{
						DataType: "text",
						Value:    "1",
					}
					itemsData["remark"] = &item.Value{
						DataType: "text",
						Value:    keiyakuno + "_" + pattern.PatternName,
					}
					itemsData["index"] = &item.Value{
						DataType: "number",
						Value:    strconv.Itoa(index),
					}

					its := &item.ListItems{
						Items: itemsData,
					}

					items = append(items, its)

					index++
					branchCount++
				}
				count++
				continue
			}

			// 比例减少的场合
			pattern := getPattern("01005", p.jouData)

			var itemMap map[string]*item.Value

			branchCount := 1
			for line, sub := range pattern.GetSubjects() {
				// LendingDivision = "1" == 借方
				// ChangeFlag = "new" == 使用变更后的数据newItemMap
				// LendingDivision = "2" == 贷方
				// ChangeFlag = "old" ==  使用变更前的数据oldItemMap
				if sub.ChangeFlag == "old" {
					itemMap = copyMap(oldItemMap)
				} else {
					// 借方 => 变更后
					itemMap = copyMap(newItemMap)
				}
				itemMap["historyno"] = &item.Value{
					DataType: "text",
					Value:    no,
				}

				expression := formula.NewExpression(sub.AmountField)

				params := getParam(sub.AmountField)
				for _, pm := range params {
					it, ok := newItemMap[pm]
					if !ok {
						it = &item.Value{
							DataType: "number",
							Value:    "0",
						}
					}
					val, err := strconv.ParseFloat(it.GetValue(), 64)
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}
					expression.AddParameter(pm, val)
				}

				result, err := expression.Evaluate()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				fv, err := result.Float64()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				if fv == 0.0 {
					continue
				}

				keiyakuno := itemMap["keiyakuno"].GetValue()
				assetsType := itemMap["bunruicd"].GetValue()
				subMap := p.asSubMap[assetsType]

				// 创建登录数据
				itemsData := copyMap(itemMap)
				itemsData["shiwakeno"] = &item.Value{
					DataType: "text",
					Value:    p.shiwakeno,
				}
				itemsData["shiwakeymd"] = &item.Value{
					DataType: "date",
					Value:    time.Now().Format("2006-01-02"),
				}
				itemsData["shiwakeym"] = &item.Value{
					DataType: "text",
					Value:    p.handleMonth,
				}
				itemsData["partten"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternId,
				}
				itemsData["lineno"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(line + 1),
				}
				itemsData["taishakukubun"] = &item.Value{
					DataType: "text",
					Value:    sub.LendingDivision,
				}
				itemsData["kanjokamoku"] = &item.Value{
					DataType: "text",
					Value:    subMap[sub.GetSubjectKey()],
				}
				itemsData["shiwakekingaku"] = &item.Value{
					DataType: "number",
					Value:    result.String(),
				}
				itemsData["shiwakeaggno_parent"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(count),
				}
				itemsData["shiwakeaggno_branch"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(branchCount),
				}
				itemsData["shiwaketype"] = &item.Value{
					DataType: "text",
					Value:    "1",
				}
				itemsData["remark"] = &item.Value{
					DataType: "text",
					Value:    keiyakuno + "_" + pattern.PatternName,
				}
				itemsData["index"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(index),
				}

				its := &item.ListItems{
					Items: itemsData,
				}

				items = append(items, its)

				index++
				branchCount++
			}
			count++
			continue
		}
		// 如果操作是中途解约的场合
		if actkbn == "midcancel" {
			// 比例减少的场合
			pattern := getPattern("01008", p.jouData)

			var itemMap map[string]*item.Value

			branchCount := 1
			for line, sub := range pattern.GetSubjects() {
				// LendingDivision = "1" == 借方
				// ChangeFlag = "new" == 使用变更后的数据newItemMap
				// LendingDivision = "2" == 贷方
				// ChangeFlag = "old" ==  使用变更前的数据oldItemMap
				if sub.ChangeFlag == "old" {
					itemMap = copyMap(oldItemMap)
				} else {
					// 借方 => 变更后
					itemMap = copyMap(newItemMap)
				}
				itemMap["historyno"] = &item.Value{
					DataType: "text",
					Value:    no,
				}

				expression := formula.NewExpression(sub.AmountField)

				params := getParam(sub.AmountField)
				for _, pm := range params {
					it, ok := newItemMap[pm]
					if !ok {
						it = &item.Value{
							DataType: "number",
							Value:    "0",
						}
					}
					val, err := strconv.ParseFloat(it.GetValue(), 64)
					if err != nil {
						loggerx.ErrorLog("insertData", err.Error())
						return nil, err
					}
					expression.AddParameter(pm, val)
				}

				result, err := expression.Evaluate()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				fv, err := result.Float64()
				if err != nil {
					loggerx.ErrorLog("insertData", err.Error())
					return nil, err
				}

				if fv == 0.0 {
					continue
				}

				keiyakuno := itemMap["keiyakuno"].GetValue()
				assetsType := itemMap["bunruicd"].GetValue()
				subMap := p.asSubMap[assetsType]

				// 创建登录数据
				itemsData := copyMap(itemMap)
				itemsData["shiwakeno"] = &item.Value{
					DataType: "text",
					Value:    p.shiwakeno,
				}
				itemsData["shiwakeymd"] = &item.Value{
					DataType: "date",
					Value:    time.Now().Format("2006-01-02"),
				}
				itemsData["shiwakeym"] = &item.Value{
					DataType: "text",
					Value:    p.handleMonth,
				}
				itemsData["partten"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternId,
				}
				itemsData["lineno"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(line + 1),
				}
				itemsData["taishakukubun"] = &item.Value{
					DataType: "text",
					Value:    sub.LendingDivision,
				}
				itemsData["kanjokamoku"] = &item.Value{
					DataType: "text",
					Value:    subMap[sub.GetSubjectKey()],
				}
				itemsData["shiwakekingaku"] = &item.Value{
					DataType: "number",
					Value:    result.String(),
				}
				itemsData["shiwakeaggno_parent"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(count),
				}
				itemsData["shiwakeaggno_branch"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(branchCount),
				}
				itemsData["shiwaketype"] = &item.Value{
					DataType: "text",
					Value:    "1",
				}
				itemsData["remark"] = &item.Value{
					DataType: "text",
					Value:    keiyakuno + "_" + pattern.PatternName,
				}
				itemsData["index"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(index),
				}

				its := &item.ListItems{
					Items: itemsData,
				}

				items = append(items, its)

				index++
				branchCount++
			}
			count++
			continue
		}
	}
	return items, nil
}

func getParam(f string) []string {
	comp := regexp.MustCompile(`\[([^]]+)\]`)
	//利用自匹配获取正则表达式里括号[]中的匹配内容
	submatchs := comp.FindAllStringSubmatch(f, -1)

	var result []string
	for _, match := range submatchs {
		result = append(result, match[1])
	}

	return result
}

func getPattern(pid string, j *journal.Journal) *journal.Pattern {
	for _, p := range j.GetPatterns() {
		if p.PatternId == pid {
			return p
		}
	}
	return nil
}

func copyMap(m map[string]*item.Value) map[string]*item.Value {
	result := make(map[string]*item.Value, len(m))
	for k, v := range m {
		result[k] = v
	}

	return result
}

// GenPayData 生成支付分录的数据
func GenPayData(domain, db, appID, userID, lang string, owners []string) (r *item.ImportResult, err error) {
	jobID := "job_" + time.Now().Format("20060102150405")
	//获取当前language
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "Payment Journal",
		Origin:       "-",
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "journal",
		Steps:        []string{"start", "collect-data", "gen-data", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	go func() {
		// 发送消息 收集数据情报
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "すべての編集データを収集します",
			CurrentStep: "collect-data",
			Database:    db,
		}, userID)

		// 获取台账map
		dsMap, err := getDatastoreMap(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 收集数据情报失败 终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取处理月度
		cfg, err := configx.GetConfigVal(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 收集数据情报失败 终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}
		handleMonth := cfg.GetSyoriYm()

		// 获取分录数据
		jouData, err := getJournal(db, appID, "03")
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取所有分类的科目的数据
		assetDs := dsMap["assets"]
		assetAccesskeys := sessionx.GetAccessKeys(db, userID, assetDs, "R")
		asSubMap, err := getSubjectMap(db, appID, assetDs, assetAccesskeys)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取分录番号
		shiwakeno, err := genShiwakeno(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 数据上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_021"),
			CurrentStep: "gen-data",
			Database:    db,
		}, userID)

		// 通过支付数据生成分录data
		param := InsertParam{
			db:          db,
			jobID:       jobID,
			domain:      domain,
			lang:        lang,
			shiwakeno:   shiwakeno,
			handleMonth: handleMonth,
			appID:       appID,
			datastoreID: dsMap["paymentInterest"],
			userID:      userID,
			owners:      owners,
			dsMap:       dsMap,
			jouData:     jouData,
			asSubMap:    asSubMap,
		}

		err = buildPayData(param)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "gen-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 任务成功结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)
	}()

	return r, nil
}

// buildPayData 获取并编辑当前月度的支付数据（根据处理月度查询），生成分录数据
func buildPayData(p InsertParam) (e error) {

	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	handleDate, err := time.Parse("2006-01", p.handleMonth)
	if err != nil {
		loggerx.ErrorLog("getPayData", err.Error())
		return err
	}

	lastDay := getMonthLastDay(handleDate)

	var req item.ItemsRequest
	conditions := []*item.Condition{}
	conditions = append(conditions, &item.Condition{
		FieldId:     "paymentymd",
		FieldType:   "date",
		SearchValue: p.handleMonth + "-01",
		Operator:    ">=",
		IsDynamic:   true,
	})

	conditions = append(conditions, &item.Condition{
		FieldId:     "paymentymd",
		FieldType:   "date",
		SearchValue: p.handleMonth + "-" + lastDay,
		Operator:    "<=",
		IsDynamic:   true,
	})

	accesskeys := sessionx.GetAccessKeys(p.db, p.userID, p.datastoreID, "R")

	// 先获取总的件数
	cReq := item.CountRequest{
		AppId:         p.appID,
		DatastoreId:   p.datastoreID,
		ConditionList: conditions,
		ConditionType: "and",
		Owners:        accesskeys,
		Database:      p.db,
	}

	countResponse, err := itemService.FindCount(context.TODO(), &cReq, opss)
	if err != nil {
		loggerx.ErrorLog("getPayData", err.Error())
		return err
	}

	// 根据总的件数分批下载数据
	// 每次2000为一组数据
	total := float64(countResponse.GetTotal())
	count := math.Ceil(total / 500)

	for index := int64(0); index < int64(count); index++ {
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "keiyakuno.value",
			SortValue: "ascend",
		})
		req.Sorts = sorts
		req.ConditionList = conditions
		req.ConditionType = "and"
		req.DatastoreId = p.datastoreID
		req.AppId = p.appID
		req.PageIndex = index + 1
		req.PageSize = 500
		req.Owners = accesskeys
		req.Database = p.db
		req.IsOrigin = true

		itemResp, err := itemService.FindItems(context.TODO(), &req, opss)
		if err != nil {
			loggerx.ErrorLog("getPayData", err.Error())
			return err
		}

		var data ItemData

		for _, item := range itemResp.GetItems() {
			itemMap := item.Items
			data = append(data, itemMap)
		}

		// 分录数据
		var items ImportData
		index := 1
		for count, payItem := range data {
			// 契约登录的场合
			pattern := getPattern("03001", p.jouData)
			keiyakuno := payItem["keiyakuno"].GetValue()
			keiyakuAccesskeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["keiyakudaicho"], "R")
			itemMap, err := getKeiyakuData(p.db, p.appID, p.dsMap["keiyakudaicho"], keiyakuno, keiyakuAccesskeys)
			if err != nil {
				loggerx.ErrorLog("getPayData", err.Error())
				return err
			}

			branchCount := 1
			for line, sub := range pattern.GetSubjects() {
				expression := formula.NewExpression(sub.AmountField)
				params := getParam(sub.AmountField)
				for _, pm := range params {
					it, ok := payItem[pm]
					if !ok {
						it = &item.Value{
							DataType: "number",
							Value:    "0",
						}
					}
					val, err := strconv.ParseFloat(it.GetValue(), 64)
					if err != nil {
						loggerx.ErrorLog("getPayData", err.Error())
						return err
					}
					expression.AddParameter(pm, val)
				}

				result, err := expression.Evaluate()
				if err != nil {
					loggerx.ErrorLog("getPayData", err.Error())
					return err
				}

				fv, err := result.Float64()
				if err != nil {
					loggerx.ErrorLog("getPayData", err.Error())
					return err
				}

				if fv == 0.0 {
					continue
				}

				assetsType := itemMap["bunruicd"].GetValue()
				subMap := p.asSubMap[assetsType]

				// 创建登录数据
				itemsData := copyMap(itemMap)

				itemsData["keiyakuno"] = &item.Value{
					DataType: "lookup",
					Value:    keiyakuno,
				}

				itemsData["shiwakeno"] = &item.Value{
					DataType: "text",
					Value:    p.shiwakeno,
				}
				itemsData["shiwakeymd"] = &item.Value{
					DataType: "date",
					Value:    time.Now().Format("2006-01-02"),
				}
				itemsData["shiwakeym"] = &item.Value{
					DataType: "text",
					Value:    p.handleMonth,
				}
				itemsData["partten"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternId,
				}
				itemsData["lineno"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(line + 1),
				}
				itemsData["taishakukubun"] = &item.Value{
					DataType: "text",
					Value:    sub.LendingDivision,
				}
				itemsData["kanjokamoku"] = &item.Value{
					DataType: "text",
					Value:    subMap[sub.GetSubjectKey()],
				}
				itemsData["shiwakekingaku"] = &item.Value{
					DataType: "number",
					Value:    result.String(),
				}
				itemsData["shiwakeaggno_parent"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(count + 1),
				}
				itemsData["shiwakeaggno_branch"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(branchCount),
				}
				itemsData["shiwaketype"] = &item.Value{
					DataType: "text",
					Value:    "3",
				}
				itemsData["remark"] = &item.Value{
					DataType: "text",
					Value:    "支払_" + p.handleMonth,
				}
				itemsData["index"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(index),
				}

				its := &item.ListItems{
					Items: itemsData,
				}

				items = append(items, its)

				index++
				branchCount++
			}
		}

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var delreq item.DeleteItemsRequest
		delreq.DatastoreId = p.dsMap["shiwake"]
		delreq.AppId = p.appID
		delreq.UserId = p.userID
		delreq.Database = p.db
		delreq.ConditionType = "and"

		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "shiwaketype",
			FieldType:     "text",
			SearchValue:   "3",
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		delreq.ConditionList = conditions

		_, err = itemService.DeleteItems(context.TODO(), &delreq, opss)
		if err != nil {
			loggerx.ErrorLog("getPayData", err.Error())
			return err
		}

		response, err := importData(p, items)
		if err != nil {
			loggerx.ErrorLog("getRepaymentData", err.Error())
			return err
		}
		fmt.Printf("%+v", response)
	}

	return nil
}

// getMonthLastDay  获取当前月份的最后一天
func getMonthLastDay(date time.Time) (day string) {
	// 年月日取得
	years := date.Year()
	month := date.Month()

	// 月末日取得
	lastday := 0
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			lastday = 30
		} else {
			lastday = 31
		}
	} else {
		if ((years%4) == 0 && (years%100) != 0) || (years%400) == 0 {
			lastday = 29
		} else {
			lastday = 28
		}
	}

	return strconv.Itoa(lastday)
}

// GenRepayData 生成偿还分录的数据
func GenRepayData(domain, db, appID, userID, lang string, owners []string) (r *item.ImportResult, err error) {
	jobID := "job_" + time.Now().Format("20060102150405")
	//获取当前的language
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "Depreciate Journal",
		Origin:       "-",
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "journal",
		Steps:        []string{"start", "collect-data", "gen-data", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	go func() {
		// 发送消息 收集数据情报
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_002"),
			CurrentStep: "collect-data",
			Database:    db,
		}, userID)

		// 获取台账map
		dsMap, err := getDatastoreMap(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 收集数据情报失败 终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取处理月度
		cfg, err := configx.GetConfigVal(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 收集数据情报失败 终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}
		handleMonth := cfg.GetSyoriYm()

		// 获取分录数据
		jouData, err := getJournal(db, appID, "02")
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取所有分类的科目的数据
		assetDs := dsMap["assets"]
		assetAccesskeys := sessionx.GetAccessKeys(db, userID, assetDs, "R")
		asSubMap, err := getSubjectMap(db, appID, assetDs, assetAccesskeys)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 获取分录番号
		shiwakeno, err := genShiwakeno(db, appID)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "collect-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 数据上传
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_022"),
			CurrentStep: "gen-data",
			Database:    db,
		}, userID)

		// 通过支付数据生成分录data
		param := InsertParam{
			db:          db,
			jobID:       jobID,
			lang:        lang,
			domain:      domain,
			shiwakeno:   shiwakeno,
			handleMonth: handleMonth,
			appID:       appID,
			datastoreID: dsMap["repayment"],
			userID:      userID,
			owners:      owners,
			dsMap:       dsMap,
			jouData:     jouData,
			asSubMap:    asSubMap,
		}

		//  生成数据
		err = buildRepaymentData(param)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "gen-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 任务成功结束
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_028"),
			CurrentStep: "end",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)
	}()

	return r, nil
}

// buildRepaymentData 获取当前月度的偿还数据（根据处理月度查询）,生成偿还分录数据
func buildRepaymentData(p InsertParam) (e error) {
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	handleDate, err := time.Parse("2006-01", p.handleMonth)
	if err != nil {
		loggerx.ErrorLog("getRepaymentData", err.Error())
		return err
	}

	lastDay := getMonthLastDay(handleDate)

	conditions := []*item.Condition{}
	conditions = append(conditions, &item.Condition{
		FieldId:     "syokyakuymd",
		FieldType:   "date",
		SearchValue: p.handleMonth + "-01",
		Operator:    ">=",
		IsDynamic:   true,
	})

	conditions = append(conditions, &item.Condition{
		FieldId:     "syokyakuymd",
		FieldType:   "date",
		SearchValue: p.handleMonth + "-" + lastDay,
		Operator:    "<=",
		IsDynamic:   true,
	})

	accesskeys := sessionx.GetAccessKeys(p.db, p.userID, p.datastoreID, "R")

	// 先获取总的件数
	cReq := item.CountRequest{
		AppId:         p.appID,
		DatastoreId:   p.datastoreID,
		ConditionList: conditions,
		ConditionType: "and",
		Owners:        accesskeys,
		Database:      p.db,
	}

	countResponse, err := itemService.FindCount(context.TODO(), &cReq, opss)
	if err != nil {
		loggerx.ErrorLog("getPayData", err.Error())
		return err
	}

	// 根据总的件数分批下载数据
	// 每次2000为一组数据
	total := float64(countResponse.GetTotal())
	count := math.Ceil(total / 500)

	for index := int64(0); index < int64(count); index++ {

		var req item.ItemsRequest
		var sorts []*item.SortItem
		sorts = append(sorts, &item.SortItem{
			SortKey:   "keiyakuno.value",
			SortValue: "ascend",
		})
		req.Sorts = sorts
		req.ConditionList = conditions
		req.ConditionType = "and"
		req.DatastoreId = p.datastoreID
		req.PageIndex = index + 1
		req.PageSize = 500
		req.AppId = p.appID
		req.Owners = accesskeys
		req.Database = p.db
		req.IsOrigin = true

		itemResp, err := itemService.FindItems(context.TODO(), &req, opss)
		if err != nil {
			loggerx.ErrorLog("getRepaymentData", err.Error())
			return err
		}

		var data ItemData

		for _, item := range itemResp.GetItems() {
			itemMap := item.Items
			data = append(data, itemMap)
		}

		// 分录数据编辑
		var items ImportData
		index := 1
		for count, repayItem := range data {
			// 契约登录的场合
			pattern := getPattern("02001", p.jouData)
			keiyakuno := repayItem["keiyakuno"].GetValue()
			keiyakuAccesskeys := sessionx.GetAccessKeys(p.db, p.userID, p.dsMap["keiyakudaicho"], "R")
			itemMap, err := getKeiyakuData(p.db, p.appID, p.dsMap["keiyakudaicho"], keiyakuno, keiyakuAccesskeys)
			if err != nil {
				loggerx.ErrorLog("getRepaymentData", err.Error())
				return err
			}

			branchCount := 1
			for line, sub := range pattern.GetSubjects() {
				expression := formula.NewExpression(sub.AmountField)
				params := getParam(sub.AmountField)
				for _, pm := range params {
					it, ok := repayItem[pm]
					if !ok {
						it = &item.Value{
							DataType: "number",
							Value:    "0",
						}
					}
					val, err := strconv.ParseFloat(it.GetValue(), 64)
					if err != nil {
						loggerx.ErrorLog("getRepaymentData", err.Error())
						return err
					}
					expression.AddParameter(pm, val)
				}

				result, err := expression.Evaluate()
				if err != nil {
					loggerx.ErrorLog("getRepaymentData", err.Error())
					return err
				}

				fv, err := result.Float64()
				if err != nil {
					loggerx.ErrorLog("getRepaymentData", err.Error())
					return err
				}

				if fv == 0.0 {
					continue
				}

				assetsType := itemMap["bunruicd"].GetValue()
				subMap := p.asSubMap[assetsType]

				// 创建登录数据
				itemsData := copyMap(itemMap)

				itemsData["keiyakuno"] = &item.Value{
					DataType: "lookup",
					Value:    keiyakuno,
				}

				itemsData["shiwakeno"] = &item.Value{
					DataType: "text",
					Value:    p.shiwakeno,
				}
				itemsData["shiwakeymd"] = &item.Value{
					DataType: "date",
					Value:    time.Now().Format("2006-01-02"),
				}
				itemsData["shiwakeym"] = &item.Value{
					DataType: "text",
					Value:    p.handleMonth,
				}
				itemsData["partten"] = &item.Value{
					DataType: "text",
					Value:    pattern.PatternId,
				}
				itemsData["lineno"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(line + 1),
				}
				itemsData["taishakukubun"] = &item.Value{
					DataType: "text",
					Value:    sub.LendingDivision,
				}
				itemsData["kanjokamoku"] = &item.Value{
					DataType: "text",
					Value:    subMap[sub.GetSubjectKey()],
				}
				itemsData["shiwakekingaku"] = &item.Value{
					DataType: "number",
					Value:    result.String(),
				}
				itemsData["shiwakeaggno_parent"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(count + 1),
				}
				itemsData["shiwakeaggno_branch"] = &item.Value{
					DataType: "text",
					Value:    strconv.Itoa(branchCount),
				}
				itemsData["shiwaketype"] = &item.Value{
					DataType: "text",
					Value:    "2",
				}
				itemsData["remark"] = &item.Value{
					DataType: "text",
					Value:    "償却_" + p.handleMonth,
				}
				itemsData["index"] = &item.Value{
					DataType: "number",
					Value:    strconv.Itoa(index),
				}

				its := &item.ListItems{
					Items: itemsData,
				}

				items = append(items, its)

				index++
				branchCount++
			}
		}

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var delreq item.DeleteItemsRequest
		delreq.DatastoreId = p.dsMap["shiwake"]
		delreq.AppId = p.appID
		delreq.UserId = p.userID
		delreq.Database = p.db
		delreq.ConditionType = "and"

		var conditions []*item.Condition
		conditions = append(conditions, &item.Condition{
			FieldId:       "shiwaketype",
			FieldType:     "text",
			SearchValue:   "2",
			Operator:      "=",
			IsDynamic:     true,
			ConditionType: "",
		})
		delreq.ConditionList = conditions

		_, err = itemService.DeleteItems(context.TODO(), &delreq, opss)
		if err != nil {
			loggerx.ErrorLog("getRepaymentData", err.Error())
			return err
		}

		response, err := importData(p, items)
		if err != nil {
			loggerx.ErrorLog("getRepaymentData", err.Error())
			return err
		}
		fmt.Printf("%+v", response)
	}

	return nil
}

// getKeiyakuData 获取租赁数据（根据keiyakuno查询）
func getKeiyakuData(db, appID, datastoreID, keiyakuno string, accesskeys []string) (d map[string]*item.Value, err error) {
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	var req item.ItemsRequest
	conditions := []*item.Condition{}
	conditions = append(conditions, &item.Condition{
		FieldId:     "keiyakuno",
		FieldType:   "text",
		SearchValue: keiyakuno,
		Operator:    "=",
		IsDynamic:   true,
	})
	req.ConditionList = conditions
	req.ConditionType = "and"
	req.DatastoreId = datastoreID
	req.AppId = appID
	req.Owners = accesskeys
	req.IsOrigin = true
	req.Database = db

	response, err := itemService.FindItems(context.TODO(), &req, opss)
	if err != nil {
		loggerx.ErrorLog("getKeiyakuData", err.Error())
		return
	}

	if response.GetTotal() == 0 {
		return nil, errors.New("not found data")
	}
	if response.GetTotal() > 1 {
		return nil, errors.New("found more data")
	}

	items := response.GetItems()

	return items[0].GetItems(), nil
}

// getKoushinbangouData 获取履历数据（根据koushinbangou查询）
func getKoushinbangouData(db, appID, datastoreID, koushinbangouoya, koushinbangoueda string, accesskeys []string) (d map[string]*item.Value, err error) {
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	defaultTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

	var req item.ItemsRequest
	conditions := []*item.Condition{}
	conditions = append(conditions, &item.Condition{
		FieldId:     "koushinbangouoya",
		FieldType:   "text",
		SearchValue: koushinbangouoya,
		Operator:    "=",
		IsDynamic:   true,
	})
	conditions = append(conditions, &item.Condition{
		FieldId:     "koushinbangoueda",
		FieldType:   "text",
		SearchValue: koushinbangoueda,
		Operator:    "=",
		IsDynamic:   true,
	})
	conditions = append(conditions, &item.Condition{
		FieldId:     "kakuteidate",
		FieldType:   "date",
		SearchValue: defaultTime.Format(time.RFC3339),
		Operator:    "=",
		IsDynamic:   true,
	})
	req.ConditionList = conditions
	req.ConditionType = "and"
	req.DatastoreId = datastoreID
	req.AppId = appID
	req.Owners = accesskeys
	req.IsOrigin = true
	req.Database = db

	response, err := itemService.FindItems(context.TODO(), &req, opss)
	if err != nil {
		loggerx.ErrorLog("getKoushinbangouData", err.Error())
		return
	}

	if response.GetTotal() == 0 {
		return nil, errors.New("not found data")
	}
	if response.GetTotal() > 1 {
		return nil, errors.New("found more data")
	}

	items := response.GetItems()

	return items[0].GetItems(), nil
}

// importData 批量导入数据
func importData(p InsertParam, items ImportData) (*item.ImportResult, error) {

	// 获取数据上传流
	itemService := item.NewItemService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
	}

	stream, err := itemService.ImportItem(context.Background(), opss)
	if err != nil {
		loggerx.ErrorLog("importData", err.Error())
		return nil, err
	}

	// 上传meta信息
	err = stream.Send(&item.ImportRequest{
		Status: item.SendStatus_SECTION,
		Request: &item.ImportRequest_Meta{
			Meta: &item.ImportMetaData{
				Key:         "",
				AppId:       p.appID,
				DatastoreId: p.dsMap["shiwake"],
				Writer:      p.userID,
				Owners:      p.owners,
				Database:    p.db,
			},
		},
	})

	if err != nil {
		loggerx.ErrorLog("importData", err.Error())
		return nil, err
	}

	langData := langx.GetLanguageData(p.db, p.lang, p.domain)

	var errorList []string
	var inserted int64 = 0
	var updated int64 = 0

	// 验证数据
	go func() {
		// 开始导入
		for _, data := range items {
			err := stream.Send(&item.ImportRequest{
				Status: item.SendStatus_SECTION,
				Request: &item.ImportRequest_Data{
					Data: &item.ImportData{
						Items: data,
					},
				},
			})
			if err == io.EOF {
				return
			}
			if err != nil {
				loggerx.ErrorLog("importData", err.Error())
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
			loggerx.ErrorLog("importData", err.Error())
			return
		}

	}()

	for {
		result, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			loggerx.ErrorLog("importData", err.Error())
			return nil, err
		}

		if result.Status == item.Status_FAILED {
			// 如果有失败的情况发生，将停止继续发送
			for _, e := range result.GetResult().GetErrors() {
				eMsg := "第{0}〜{1}行目でエラーが発生しました。エラー内容：{2}"
				fieldErrorMsg := "第{0}行目でエラーが発生しました。フィールド名：[{1}]、エラー内容：{2}"
				noFieldErrorMsg := "第{0}行目でエラーが発生しました。エラー内容：{1}"
				if len(e.FieldId) == 0 {
					if e.CurrentLine != 0 {
						es, _ := msg.Format(noFieldErrorMsg, strconv.FormatInt(e.CurrentLine, 10), e.ErrorMsg)
						errorList = append(errorList, es)
					} else {
						es, _ := msg.Format(eMsg, strconv.FormatInt(e.FirstLine, 10), strconv.FormatInt(e.LastLine, 10), e.ErrorMsg)
						errorList = append(errorList, es)
					}
				} else {
					es, _ := msg.Format(fieldErrorMsg, strconv.FormatInt(e.CurrentLine, 10), langx.GetLangValue(langData, langx.GetFieldKey(p.appID, p.datastoreID, e.FieldId), langx.DefaultResult), e.ErrorMsg)
					errorList = append(errorList, es)
				}
			}

			// 终止继续发送
			err := stream.Send(&item.ImportRequest{
				Status: item.SendStatus_COMPLETE,
				Request: &item.ImportRequest_Data{
					Data: nil,
				},
			})

			if err != nil {
				loggerx.ErrorLog("importData", err.Error())
				return nil, err
			}
			break
		}

		if result.Status == item.Status_SUCCESS {

			inserted = inserted + result.Result.Insert
			updated = updated + result.Result.Modify
			continue
		}
	}

	if len(errorList) > 0 {
		loggerx.ErrorLog("importData", fmt.Sprintf("%v", errorList))
		return nil, fmt.Errorf("%v", errorList)

	}

	return &item.ImportResult{
		Insert: inserted,
		Modify: updated,
	}, nil
}
