package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	AppsCollection       = "apps"
	DataStoresCollection = "data_stores"
)

// Config 顾客配置情报
type Config struct {
	AppID string `json:"app_id" bson:"app_id"`
	Value string `json:"value" bson:"value"`
}

// App 应用程序
type App struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	AppID        string             `json:"app_id" bson:"app_id"`
	AppName      string             `json:"app_name" bson:"app_name"`
	AppType      string             `json:"app_type" bson:"app_type"`
	DisplayOrder int64              `json:"display_order" bson:"display_order"`
	TemplateID   string             `json:"template_id" bson:"template_id"`
	Domain       string             `json:"domain" bson:"domain"`
	IsTrial      bool               `json:"is_trial" bson:"is_trial"`
	StartTime    string             `json:"start_time" bson:"start_time"`
	EndTime      string             `json:"end_time" bson:"end_time"`
	CopyFrom     string             `json:"copy_from" bson:"copy_from"`
	FollowApp    string             `json:"follow_app" bson:"follow_app"`
	Remarks      string             `json:"remarks" bson:"remarks"`
	Configs      Configs            `json:"configs" bson:"configs"`
	SwkControl   bool               `json:"swk_control" bson:"swk_control"`
	ConfimMethod string             `json:"confim_method" bson:"confim_method"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy    string             `json:"created_by" bson:"created_by"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy    string             `json:"updated_by" bson:"updated_by"`
	DeletedAt    time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy    string             `json:"deleted_by" bson:"deleted_by"`
}
type Configs struct {
	Special         string `json:"special" bson:"special"`
	CheckStartDate  string `json:"check_start_date" bson:"check_start_date"`
	SyoriYm         string `json:"syori_ym" bson:"syori_ym"`
	ShortLeases     string `json:"short_leases" bson:"short_leases"`
	KishuYm         string `json:"kishu_ym" bson:"kishu_ym"`
	MinorBaseAmount string `json:"minor_base_amount" bson:"minor_base_amount"`
}

// Max 最大顺
type Max struct {
	ID       string `json:"id" bson:"_id"`
	MaxValue int64  `json:"max_value" bson:"max_value"`
}

// ToProto 转换为proto数据
func (a *App) ToProto() *app.App {
	apps := app.App{
		AppId:        a.AppID,
		AppName:      a.AppName,
		AppType:      a.AppType,
		DisplayOrder: a.DisplayOrder,
		TemplateId:   a.TemplateID,
		Domain:       a.Domain,
		IsTrial:      a.IsTrial,
		StartTime:    a.StartTime,
		EndTime:      a.EndTime,
		CopyFrom:     a.CopyFrom,
		FollowApp:    a.FollowApp,
		Remarks:      a.Remarks,
		SwkControl:   a.SwkControl,
		ConfimMethod: a.ConfimMethod,
		CreatedAt:    a.CreatedAt.String(),
		CreatedBy:    a.CreatedBy,
		UpdatedAt:    a.UpdatedAt.String(),
		UpdatedBy:    a.UpdatedBy,
		DeletedAt:    a.DeletedAt.String(),
		DeletedBy:    a.DeletedBy,
	}
	config := app.Configs{
		Special:         a.Configs.Special,
		CheckStartDate:  a.Configs.CheckStartDate,
		KishuYm:         a.Configs.KishuYm,
		SyoriYm:         a.Configs.SyoriYm,
		ShortLeases:     a.Configs.ShortLeases,
		MinorBaseAmount: a.Configs.MinorBaseAmount,
	}
	apps.Configs = &config
	return &apps
}

// FindAppsByIds 根据APPID数组查询
func FindAppsByIds(ctx context.Context, db, domain string, appIDlist []string) (a []App, err error) {
	// _, sc := database.BeginMongo()
	// c := sc.DB(database.GetDBName(db)).C(AppsCollection)
	// defer sc.Close()

	// var result []App

	// query := bson.M{
	// 	"deleted_by": "",
	// 	"domain":     domain,
	// 	"app_id":     bson.M{"$in": appIDlist}}

	// queryJSON, _ := json.Marshal(query)
	// utils.DebugLog("FindAppListApps", "FindAppListApps", fmt.Sprintf("query: [ %s ]", queryJSON))

	// if err := c.Find(query).Sort("display_order").All(&result); err != nil {
	// 	utils.ErrorLog("error FindSystemApps", err.Error())
	// 	return result, err
	// }
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	var result []App

	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
		"app_id":     bson.M{"$in": appIDlist}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindAppListApps", fmt.Sprintf("query: [ %s ]", queryJSON))
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindAppListApps", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var app App
		err := cur.Decode(&app)
		if err != nil {
			utils.ErrorLog("error FindAppListApps", err.Error())
			return nil, err
		}
		result = append(result, app)
	}

	return result, nil
}

// FindApps 默认查询
func FindApps(ctx context.Context, db, domain, appName, invalidatedIn, isTrial, startTime, endTime, copyFrom string) (a []App, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	var result []App
	// 默认不包含无效数据
	query := bson.M{
		"deleted_by": "",
	}

	if domain != "" {
		query["domain"] = domain
	}

	// 是否包含无效数据
	if invalidatedIn != "" {
		delete(query, "deleted_by")
	}

	// APP名称不为空
	if appName != "" {
		query["app_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(appName), Options: "m"}}
	}

	// 试用FLG不为空
	if isTrial != "" {
		ok, err := strconv.ParseBool(isTrial)
		if err != nil {
			ok = false
		}
		query["is_trial"] = ok
	}
	// 正式使用开始日不为空
	if startTime != "" {
		query["start_time"] = bson.M{"$gte": startTime}
	}

	// 正式使用开始日不为空
	if endTime != "" {
		query["end_time"] = bson.M{"$lte": endTime}
	}

	// 复制元不为空
	if copyFrom != "" {
		query["copy_from"] = copyFrom
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindApps", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindApps", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var app App
		err := cur.Decode(&app)
		if err != nil {
			utils.ErrorLog("error FindApps", err.Error())
			return nil, err
		}
		result = append(result, app)
	}
	return result, nil
}

// FindApp 通过APPID查找单个APP记录
func FindApp(ctx context.Context, db, appID string) (a App, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	var result App
	objectID, err := primitive.ObjectIDFromHex(appID)
	if err != nil {
		utils.ErrorLog("error FindCustomer", err.Error())
		return result, err
	}

	query := bson.M{
		"deleted_by": "",
		"_id":        objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindApp", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindApp", err.Error())
		return result, err
	}
	return result, nil
}

// AddApp 添加单个APP记录
func AddApp(ctx context.Context, db string, a *App) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	// 添加app排序的seq
	seq := GetAppDisplayOrderSequenceName(db)
	max, err := GetNextSequenceValue(ctx, db, seq)
	if err != nil {
		utils.ErrorLog("error AddApp", err.Error())
		return "", err
	}

	// 编辑最大顺
	a.DisplayOrder = int64(max)

	// 编辑ID
	a.ID = primitive.NewObjectID()
	a.AppID = a.ID.Hex()
	a.AppName = "apps." + a.AppID + ".app_name"

	queryJSON, _ := json.Marshal(a)
	utils.DebugLog("AddApp", fmt.Sprintf("App: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, a)
	if err != nil {
		utils.ErrorLog("error AddApp", err.Error())
		return "", err
	}

	// 添加app排序的seq
	if err := CreateSequence(ctx, db, "option_"+a.AppID, 0); err != nil {
		utils.ErrorLog("error CreateSequence", err.Error())
		return "", err
	}

	return a.AppID, nil
}

// ModifyApp 修改单个APP记录
func ModifyApp(ctx context.Context, db string, a *App) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	objectID, err := primitive.ObjectIDFromHex(a.AppID)
	if err != nil {
		utils.ErrorLog("error ModifyApp", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": a.UpdatedAt,
		"updated_by": a.UpdatedBy,
		"is_trial":   a.IsTrial,
		"start_time": a.StartTime,
		"end_time":   a.EndTime,
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyApp", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyApp", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyApp", err.Error())
		return err
	}
	return nil
}

// ModifyAppConfigs 修改单个APP记录
func ModifyAppConfigs(ctx context.Context, db, appID string, config Configs) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	objectID, err := primitive.ObjectIDFromHex(appID)
	if err != nil {
		utils.ErrorLog("error ModifyAppConfigs", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"configs": config,
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyAppConfigs", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyAppConfigs", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyAppConfigs", err.Error())
		return err
	}
	return nil
}

// ModifyAppSort APP排序修改
func ModifyAppSort(ctx context.Context, db string, applist []*App) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error ModifyAppSort", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error ModifyAppSort", err.Error())
		return 0, err
	}
	count = 0
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, app := range applist {
			objectID, err := primitive.ObjectIDFromHex(app.AppID)
			if err != nil {
				utils.ErrorLog("error ModifyAppSort", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"display_order": app.DisplayOrder,
				"updated_at":    app.UpdatedAt,
				"updated_by":    app.UpdatedBy,
			}}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("ModifyAppSort", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("ModifyAppSort", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			count++
			if err != nil {
				utils.ErrorLog("error ModifyAppSort", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error ModifyAppSort", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error ModifyAppSort", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	return count, nil
}

// DeleteApp 删除单个APP记录
func DeleteApp(ctx context.Context, db, appID, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	objectID, err := primitive.ObjectIDFromHex(appID)
	if err != nil {
		utils.ErrorLog("error DeleteApp", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteApp", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteApp", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteApp", err.Error())
		return err
	}

	return nil
}

// DeleteSelectApps 删除选中的APP记录
func DeleteSelectApps(ctx context.Context, db string, appIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectApps", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectApps", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, appID := range appIDList {
			objectID, err := primitive.ObjectIDFromHex(appID)
			if err != nil {
				utils.ErrorLog("error DeleteSelectApps", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSelectApps", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectApps", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error DeleteSelectApps", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectApps", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectApps", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// HardDeleteApps 物理删除选中的APP记录
func HardDeleteApps(ctx context.Context, db string, appIDList []string) error {
	client := database.New()
	appCollection := client.Database(database.GetDBName(db)).Collection(AppsCollection)
	datastoreCollection := client.Database(database.GetDBName(db)).Collection("data_stores")
	configsCollection := client.Database(database.GetDBName(db)).Collection("configs")
	fieldCollection := client.Database(database.GetDBName(db)).Collection("fields")
	historyCollection := client.Database(database.GetDBName(db)).Collection("histories")
	checkHistoryCollection := client.Database(database.GetDBName(db)).Collection("check_histories")
	taskCollection := client.Database(database.Db).Collection("tasks")
	taskHistoryCollection := client.Database(database.Db).Collection("task_histories")
	reportCollection := client.Database(database.GetDBName(db)).Collection("reports")
	dashboardCollection := client.Database(database.GetDBName(db)).Collection("dashboards")
	languageCollection := client.Database(database.GetDBName(db)).Collection("languages")
	userCollection := client.Database(database.GetDBName(db)).Collection("users")
	roleCollection := client.Database(database.GetDBName(db)).Collection("roles")
	groupCollection := client.Database(database.GetDBName(db)).Collection("groups")
	formCollection := client.Database(database.GetDBName(db)).Collection("wf_form")
	nodeCollection := client.Database(database.GetDBName(db)).Collection("wf_node")
	workflowCollection := client.Database(database.GetDBName(db)).Collection("wf_workflows")
	exampleCollection := client.Database(database.GetDBName(db)).Collection("wf_examples")
	peocessCollection := client.Database(database.GetDBName(db)).Collection("wf_process")
	seqCollection := client.Database(database.GetDBName(db)).Collection("sequences")
	queryCollection := client.Database(database.GetDBName(db)).Collection("queries")
	scheduleCollection := client.Database(database.GetDBName(db)).Collection("schedules")
	optionCollection := client.Database(database.GetDBName(db)).Collection("options")
	journalCollection := client.Database(database.GetDBName(db)).Collection("journals")
	printsCollection := client.Database(database.GetDBName(db)).Collection("prints")
	cAC := client.Database(database.GetDBName(db)).Collection("access")
	cPM := client.Database(database.GetDBName(db)).Collection("permissions")

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteApps", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteApps", err.Error())
		return err
	}

	type (
		Datastore struct {
			DatastoreID string `json:"datastore_id" bson:"datastore_id"`
		}

		Field struct {
			FieldID     string `json:"field_id" bson:"field_id"`
			DatastoreID string `json:"datastore_id" bson:"datastore_id"`
		}

		Report struct {
			ReportID string `json:"report_id" bson:"report_id"`
		}

		WorkFlow struct {
			WorkFlowID string `json:"wf_id" bson:"wf_id"`
		}

		Example struct {
			ExampleID string `json:"ex_id" bson:"ex_id"`
		}
	)

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, appID := range appIDList {
			// 删除对应APP
			objectID, err := primitive.ObjectIDFromHex(appID)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteApps", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = appCollection.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除app下的所有台账
			q1 := bson.M{
				"app_id": appID,
			}
			queryJSON1, _ := json.Marshal(q1)
			utils.DebugLog("HardDeleteApps", fmt.Sprintf("query: [ %s ]", queryJSON1))

			// 查询到所有的台账备用
			var dsList []Datastore

			cur, err := datastoreCollection.Find(ctx, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}
			defer cur.Close(ctx)

			err = cur.All(ctx, &dsList)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}
			// 删除APP下的台账
			_, err = datastoreCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除APP下的所有选项
			_, err = optionCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除对应的任务
			q7 := bson.M{
				"params.app_id": appID,
			}

			_, err = scheduleCollection.DeleteMany(sc, q7)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			//删除总表
			for _, ds := range appIDList {
				d := client.Database(database.GetDBName(db)).Collection("summary_" + ds)
				err := d.Drop(ctx)
				if err != nil {
					utils.ErrorLog("error Delete summary", err.Error())
					return err
				}
			}

			// 循环台账删除数据
			for _, ds := range dsList {
				d := client.Database(database.GetDBName(db)).Collection("item_" + ds.DatastoreID)
				err := d.Drop(ctx)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}

				// 删除角色台账配置信息
				update2 := bson.M{
					"$pull": bson.M{
						"datastores": bson.M{
							"datastore_id": ds.DatastoreID,
						},
					},
				}
				_, err = roleCollection.UpdateMany(sc, bson.M{}, update2)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
				// 删除App下的access
				udAc := bson.M{
					"$unset": bson.M{
						"apps." + appID: "",
					},
				}
				_, err = cAC.UpdateMany(ctx, bson.M{}, udAc)
				if err != nil {
					utils.ErrorLog("HardDeleteApps", err.Error())
					return err
				}

				// 删除App下的permissions
				quPm := bson.M{
					"app_id": appID,
				}
				_, err = cPM.DeleteMany(ctx, quPm)
				if err != nil {
					utils.ErrorLog("HardDeleteApps", err.Error())
					return err
				}

				// 删除用户组中流程的信息
				update3 := bson.M{
					"$unset": bson.M{
						"workflow." + ds.DatastoreID: "",
					},
				}
				_, err = groupCollection.UpdateMany(sc, bson.M{}, update3)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}

				// 删除所有流程的信息
				q2 := bson.M{
					"params.datastore": ds.DatastoreID,
				}

				// 查找流程信息备用
				var wfList []WorkFlow

				cur2, err := workflowCollection.Find(ctx, q2)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
				defer cur2.Close(ctx)

				err = cur2.All(ctx, &wfList)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}

				// 删除台账下的流程等
				for _, wf := range wfList {
					q3 := bson.M{
						"wf_id": wf.WorkFlowID,
					}

					// 删除form
					_, err = formCollection.DeleteMany(sc, q3)
					if err != nil {
						utils.ErrorLog("error HardDeleteApps", err.Error())
						return err
					}
					// 删除node
					_, err = nodeCollection.DeleteMany(sc, q3)
					if err != nil {
						utils.ErrorLog("error HardDeleteApps", err.Error())
						return err
					}
					// 删除workflow
					_, err = workflowCollection.DeleteMany(sc, q3)
					if err != nil {
						utils.ErrorLog("error HardDeleteApps", err.Error())
						return err
					}
					// 删除example
					// 查找example信息备用
					var exList []Example

					cur3, err := exampleCollection.Find(ctx, q3)
					if err != nil {
						utils.ErrorLog("error HardDeleteApps", err.Error())
						return err
					}
					defer cur3.Close(ctx)

					err = cur3.All(ctx, &exList)
					if err != nil {
						utils.ErrorLog("error HardDeleteApps", err.Error())
						return err
					}

					_, err = exampleCollection.DeleteMany(sc, q3)
					if err != nil {
						utils.ErrorLog("error HardDeleteApps", err.Error())
						return err
					}
					// 删除process
					for _, ex := range exList {
						q4 := bson.M{
							"ex_id": ex.ExampleID,
						}
						_, err = peocessCollection.DeleteMany(sc, q4)
						if err != nil {
							utils.ErrorLog("error HardDeleteApps", err.Error())
							return err
						}
					}
				}

				// 删除对应的seq
				q5 := bson.M{
					"_id": "datastore_" + ds.DatastoreID + "_fields__displayorder",
				}

				_, err = seqCollection.DeleteOne(sc, q5)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
				// 删除对应seq中的option
				qo := bson.M{
					"_id": "option_" + appID,
				}
				_, err = seqCollection.DeleteOne(sc, qo)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
				// 删除对应的query
				q6 := bson.M{
					"datastore_id": ds.DatastoreID,
				}

				_, err = queryCollection.DeleteMany(sc, q6)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
				// 删除盘点记录
				_, err = checkHistoryCollection.DeleteMany(sc, q6)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
			}

			// 删除APP下的字段
			// 查询到autonum的字段备用
			var afList []Field

			q8 := bson.M{
				"app_id":     appID,
				"field_type": "autonum",
			}

			cur6, err := fieldCollection.Find(ctx, q8)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}
			defer cur6.Close(ctx)

			err = cur6.All(ctx, &afList)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			_, err = fieldCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除APP下的履历
			_, err = historyCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除APP下的任务task
			_, err = taskCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除APP下的任务履历
			_, err = taskHistoryCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除APP下的报表
			// 查询到所有的台账备用
			var rpList []Report
			cur1, err := reportCollection.Find(ctx, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}
			defer cur1.Close(ctx)

			err = cur1.All(ctx, &rpList)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			_, err = reportCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 循环删除角色对应的报表配置
			for _, rp := range rpList {
				// 删除角色台账配置信息
				update2 := bson.M{
					"$pull": bson.M{
						"reports": rp.ReportID,
					},
				}
				_, err = roleCollection.UpdateMany(sc, bson.M{}, update2)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}

				// 删除app下的所有报表数据
				d := client.Database(database.GetDBName(db)).Collection("report_" + rp.ReportID)
				err = d.Drop(ctx)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
			}

			// 删除APP下的仪表盘
			_, err = dashboardCollection.DeleteMany(sc, q1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除app的语言数据
			update := bson.M{
				"$unset": bson.M{
					"apps." + appID: "",
				},
			}
			_, err = languageCollection.UpdateMany(sc, bson.M{}, update)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除用户中的app信息
			update1 := bson.M{
				"$pull": bson.M{
					"apps": appID,
				},
			}
			_, err = userCollection.UpdateMany(sc, bson.M{}, update1)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除对应的seq
			q6 := bson.M{
				"_id": "app_" + appID + "_displayorder",
			}

			_, err = seqCollection.DeleteOne(sc, q6)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除autonum对应的seq
			for _, f := range afList {
				q7 := bson.M{
					"_id": "datastore_" + f.DatastoreID + "_fields_" + f.FieldID + "_auto",
				}

				_, err = seqCollection.DeleteOne(sc, q7)
				if err != nil {
					utils.ErrorLog("error HardDeleteApps", err.Error())
					return err
				}
			}

			// 删除对应的仕訳
			qj := bson.M{
				"app_id": appID,
			}
			_, err = journalCollection.DeleteMany(sc, qj)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除该应用对应的配置
			qcon := bson.M{
				"app_id": appID,
			}
			_, err = configsCollection.DeleteMany(sc, qcon)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}

			// 删除该应用对应的所有台账的打印设置
			qprint := bson.M{
				"app_id": appID,
			}
			_, err = printsCollection.DeleteMany(sc, qprint)
			if err != nil {
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteApps", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteApps", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// RecoverSelectApps 恢复选中的APP记录
func RecoverSelectApps(ctx context.Context, db string, appIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectApps", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectApps", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, appID := range appIDList {
			objectID, err := primitive.ObjectIDFromHex(appID)
			if err != nil {
				utils.ErrorLog("error RecoverSelectApps", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"updated_at": time.Now(),
				"updated_by": userID,
				"deleted_by": "",
			}}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("RecoverSelectApps", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectApps", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectApps", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectApps", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectApps", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// NextMonth 下一月度处理
func NextMonth(ctx context.Context, db string, conf Config) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)
	cd := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)

	// Datastore Datastore信息
	type Datastore struct {
		DatastoreID string `json:"datastore_id" bson:"datastore_id"`
	}

	// 契约台账履历表取得
	var result Datastore
	query := bson.M{
		"deleted_by": "",
		"app_id":     conf.AppID,
		"api_key":    "rireki",
	}
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("NextMonth", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := cd.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("NextMonth", err.Error())
		return err
	}

	cr := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(result.DatastoreID))

	// 开启事务,更新契约履历和数据
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error NextMonth", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error NextMonth", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// 更新配置表数据
		query := bson.M{
			"app_id": conf.AppID,
		}
		config := bson.M{}
		config["configs.syori_ym"] = conf.Value
		update := bson.M{
			"$set": config,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("NextMonth", fmt.Sprintf("query: [ %s ]", queryJSON))

		updateSON, _ := json.Marshal(update)
		utils.DebugLog("NextMonth", fmt.Sprintf("update: [ %s ]", updateSON))

		_, err = c.UpdateOne(sc, query, update)
		if err != nil {
			utils.ErrorLog("error NextMonth", err.Error())
			return err
		}

		// 更新契约履历表数据
		query1 := bson.M{
			"items.dockkbn": bson.M{
				"data_type": "options",
				"value":     "undo",
			},
		}
		update1 := bson.M{
			"$set": bson.M{
				"items.dockkbn": bson.M{
					"data_type": "options",
					"value":     "done",
				},
			},
		}

		queryJSON1, _ := json.Marshal(query1)
		utils.DebugLog("NextMonth", fmt.Sprintf("query: [ %s ]", queryJSON1))

		updateSON1, _ := json.Marshal(update1)
		utils.DebugLog("NextMonth", fmt.Sprintf("update: [ %s ]", updateSON1))

		_, err := cr.UpdateMany(sc, query1, update1)
		if err != nil {
			utils.ErrorLog("error NextMonth", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error NextMonth", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error NextMonth", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// ModifySwkSetting 更新基本设定
func ModifySwkSetting(ctx context.Context, db string, appID string, handleMonth string, swkControl bool, confimMethod string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AppsCollection)

	query := bson.M{
		"app_id": appID,
	}

	update := bson.M{
		"$set": bson.M{
			"configs.syori_ym": handleMonth,
			"swk_control":      swkControl,
			"confim_method":    confimMethod,
		},
	}

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("ModifySwkSetting", err.Error())
		return err
	}

	return nil
}
