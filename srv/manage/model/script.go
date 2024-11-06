package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/manage/proto/script"
	"rxcsoft.cn/pit3/srv/manage/utils"

	database "rxcsoft.cn/utils/mongo"
)

const (
	// ScriptsCollection script collection
	ScriptsCollection = "scripts"
)

// Script Script信息
type Script struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	ScriptId      string             `json:"script_id" bson:"script_id"`
	ScriptName    string             `json:"script_name" bson:"script_name"`
	ScriptDesc    string             `json:"script_desc" bson:"script_desc"`
	ScriptType    string             `json:"script_type" bson:"script_type"`
	ScriptData    string             `json:"script_data" bson:"script_data"`
	ScriptFunc    string             `json:"script_func" bson:"script_func"`
	ScriptVersion string             `json:"script_version" bson:"script_version"`
	RunLogs       []string           `json:"run_logs" bson:"run_logs"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy     string             `json:"created_by" bson:"created_by"`
	RanAt         time.Time          `json:"ran_at" bson:"ran_at"`
	RanBy         string             `json:"ran_by" bson:"ran_by"`
}

// ToProto 转换为proto数据
func (u *Script) ToProto() *script.ScriptJob {
	return &script.ScriptJob{
		ScriptId:      u.ScriptId,
		ScriptName:    u.ScriptName,
		ScriptDesc:    u.ScriptDesc,
		ScriptType:    u.ScriptType,
		ScriptData:    u.ScriptData,
		ScriptFunc:    u.ScriptFunc,
		ScriptVersion: u.ScriptVersion,
		RunLogs:       u.RunLogs,
		CreatedAt:     u.CreatedAt.String(),
		CreatedBy:     u.CreatedBy,
		RanAt:         u.RanAt.String(),
		RanBy:         u.RanBy,
	}
}

// FindScriptJobs 查找多个Script记录
func FindScriptJobs(ctx context.Context, db, scriptType, scriptVersion, ranBy string) (u []Script, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScriptsCollection)

	query := bson.M{}

	// Script类型
	if scriptType != "" {
		query["script_type"] = scriptType
	}

	// Script版本
	if scriptVersion != "" {
		query["script_version"] = scriptVersion
	}

	// Script执行者
	if ranBy != "" {
		query["ran_by"] = ranBy
	}

	queryJSON, err := json.Marshal(query)
	utils.DebugLog("FindScriptJobs", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Script

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindScriptJobs", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindScriptJobs", err.Error())
		return nil, err
	}

	return result, nil
}

// FindScriptJob 通过ScriptID,查找单个Script记录
func FindScriptJob(ctx context.Context, db, scriptID string) (u Script, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScriptsCollection)

	query := bson.M{
		"script_id": scriptID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindScriptByID", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Script
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindScriptByID", err.Error())
		return result, err
	}

	return result, nil
}

// AddScriptJob 添加单个Script记录
func AddScriptJob(ctx context.Context, db string, u *Script) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScriptsCollection)
	// ctxTmp, cancel := context.WithTimeout(ctx, 120*time.Second)
	// defer cancel()
	u.ID = primitive.NewObjectID()
	u.RunLogs = []string{}

	queryJSON, _ := json.Marshal(u)
	utils.DebugLog("AddScript", fmt.Sprintf("Script: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, u)
	if err != nil {
		utils.ErrorLog("error AddScript", err.Error())
		return err
	}

	return nil
}

// ModifyScriptJob 更新Script的信息
func ModifyScriptJob(ctx context.Context, db string, u *Script) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScriptsCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	query := bson.M{
		"script_id": u.ScriptId,
	}

	change := bson.M{}

	// Script名称不为空的场合
	if u.ScriptName != "" {
		change["script_name"] = u.ScriptName
	}
	// Script登描述不为空的场合
	if u.ScriptDesc != "" {
		change["script_desc"] = u.ScriptDesc
	}
	// Script数据不为空的场合
	if u.ScriptData != "" {
		change["script_data"] = u.ScriptData
	}
	// Script函数不为空的场合
	if u.ScriptFunc != "" {
		change["script_func"] = u.ScriptFunc
	}
	// Script版本不为空的场合
	if u.ScriptVersion != "" {
		change["script_version"] = u.ScriptVersion
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyScriptJob", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyScriptJob", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyScriptJob", err.Error())
		return err
	}

	return nil
}

// StartScriptJob 开始执行脚本
func StartScriptJob(ctx context.Context, db, scriptId, userId string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScriptsCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	query := bson.M{
		"script_id": scriptId,
	}

	change := bson.M{
		"ran_by": userId,
		"ran_at": time.Now(),
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("StartScriptJob", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("StartScriptJob", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error StartScriptJob", err.Error())
		return err
	}

	return nil
}

// AddScriptLog 开始执行脚本
func AddScriptLog(ctx context.Context, db, scriptId, runLog string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScriptsCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	query := bson.M{
		"script_id": scriptId,
	}

	change := bson.M{
		"run_logs": runLog,
	}

	update := bson.M{"$push": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("StartScriptJob", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("StartScriptJob", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error StartScriptJob", err.Error())
		return err
	}

	return nil
}

// DeleteDuplicateAndAddIndex 删除重复Script记录并创建"script_id"的索引
func DeleteDuplicateAndAddIndex(ctx context.Context, db string, scriptIds []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScriptsCollection)
	for _, scriptID := range scriptIds {
		findQuery := bson.M{
			"script_id": scriptID,
		}

		cur, err := c.Find(ctx, findQuery)
		if err != nil {
			utils.ErrorLog("error DeleteDuplicateAndAddIndex", err.Error())
		}
		defer cur.Close(ctx)
		tmpflag := 0
		for cur.Next(ctx) {
			var result Script
			if tmpflag > 1 {
				err = cur.Decode(&result)
				if err != nil {
					utils.ErrorLog("error DeleteDuplicateAndAddIndex", err.Error())
					return err
				}
				deleteQuery := bson.M{
					"_id": result.ID,
				}
				queryJSON, _ := json.Marshal(deleteQuery)
				utils.DebugLog("DeleteDuplicateAndAddIndex", fmt.Sprintf("query: [ %s ]", queryJSON))
				_, err = c.DeleteOne(ctx, deleteQuery)
				if err != nil {
					utils.ErrorLog("error DeleteDuplicateAndAddIndex", err.Error())
					return err
				}
			}
			tmpflag++
		}
	}
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "script_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	// 判断script中是否存在 script_id 索引
	script, err := IndexExits(ctx, c, "script_id")
	if err != nil {
		utils.ErrorLog("DeleteDuplicateAndAddIndex", err.Error())
		return err
	}
	if !script {
		// 添加script表 script_id 唯一索引
		if _, err := c.Indexes().CreateOne(ctx, index); err != nil {
			utils.ErrorLog("DeleteDuplicateAndAddIndex", err.Error())
			return err
		}
	}
	return nil
}
