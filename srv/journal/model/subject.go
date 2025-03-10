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
	"rxcsoft.cn/pit3/srv/journal/proto/subject"
	"rxcsoft.cn/pit3/srv/journal/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// SubjectCollection subject collection
	SubjectCollection = "subjects"
)

type (
	// Subject 科目
	Subject struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		SubjectKey  string             `json:"subject_key" bson:"subject_key"`
		SubjectName string             `json:"subject_name" bson:"subject_name"`
		SubjectCd   string             `json:"subject_cd" bson:"subject_cd"`
		DefaultName string             `json:"default_name" bson:"default_name"`
		AssetsType  string             `json:"assets_type" bson:"assets_type"`
		AppID       string             `json:"app_id" bson:"app_id"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	}
	//检索条件
	Condition struct {
		FieldID     string `json:"field_id" bson:"field_id"`
		SearchValue string `json:"search_value" bson:"search_value"`
		Operator    string `json:"operator" bson:"operator"`
	}
)

// ToProto 转换为proto数据
func (w *Subject) ToProto() *subject.Subject {
	return &subject.Subject{
		SubjectKey:  w.SubjectKey,
		SubjectName: w.SubjectName,
		SubjectCd:   w.SubjectCd,
		DefaultName: w.DefaultName,
		AssetsType:  w.AssetsType,
		AppId:       w.AppID,
		CreatedAt:   w.CreatedAt.String(),
		CreatedBy:   w.CreatedBy,
		UpdatedAt:   w.UpdatedAt.String(),
		UpdatedBy:   w.UpdatedBy,
	}
}

// FindSubjects 获取APP下的当前分类的科目
func FindSubjects(db, appId, assetsType string) (items []Subject, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SubjectCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":      appId,
		"assets_type": assetsType,
	}

	var result []Subject

	opts := options.Find().SetSort(bson.D{{Key: "subject_key", Value: 1}})
	subjects, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindSubjects", err.Error())
		return nil, err
	}
	defer subjects.Close(ctx)
	for subjects.Next(ctx) {
		var exp Subject
		err := subjects.Decode(&exp)
		if err != nil {
			utils.ErrorLog("error FindSubjects", err.Error())
			return nil, err
		}
		result = append(result, exp)
	}

	return result, nil
}

// FindSubject 获取科目
func FindSubject(db, appId, assetsType, subjectKey string) (items Subject, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SubjectCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":      appId,
		"assets_type": assetsType,
		"subject_key": subjectKey,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindSubject", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Subject

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindSubject", err.Error())
		return result, err
	}

	return result, nil
}

// ImportSubject 导入科目数据
func ImportSubject(db string, subjects []*Subject) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SubjectCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error ImportSubject", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error ImportSubject", err.Error())
		return err
	}
	var insertModels []mongo.WriteModel
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, item := range subjects {
			item.ID = primitive.NewObjectID()
			insertCxModel := mongo.NewInsertOneModel()
			insertCxModel.SetDocument(item)
			insertModels = append(insertModels, insertCxModel)
		}

		if len(insertModels) > 0 {
			result, err := c.BulkWrite(sc, insertModels)
			if err != nil {
				bke, ok := err.(mongo.BulkWriteException)
				if !ok {
					utils.ErrorLog("error ImportSubject", err.Error())
					return err
				}
				errInfo := bke.WriteErrors[0]
				utils.ErrorLog("error ImportSubject", errInfo.Error())
				return errInfo
			}
			log.Infof("ImportSubject add result %v", result)
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error ImportSubject", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error ImportSubject", err.Error())
		return err
	}

	session.EndSession(ctx)

	return nil
}

// ModifySubject 更新流程实例数据
func ModifySubject(db, appId, assetsType, subjectKey, subjectName, SubjectCd, defaultName, writer string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SubjectCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":      appId,
		"subject_key": subjectKey,
	}

	if assetsType != "" {
		query["assets_type"] = assetsType
	} else {
		query["assets_type"] = ""
	}

	change := bson.M{
		"default_name":  defaultName,
		"subject_name":  subjectName,
		"shurui_type":   "",
		"segument_type": "",
		"subject_cd":    SubjectCd,
		"updated_at":    time.Now(),
		"updated_by":    writer,
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifySubject", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifySubject", fmt.Sprintf("update: [ %s ]", updateJSON))

	opt := options.Update()
	opt.SetUpsert(true)

	_, err = c.UpdateOne(ctx, query, update, opt)
	if err != nil {
		utils.ErrorLog("error ModifySubject", err.Error())
		return err
	}

	return nil
}

// DeleteSubject 删除流程实例数据
func DeleteSubject(db string, appId, assetsType string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SubjectCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":      appId,
		"assets_type": assetsType,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteSubject", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteSubject", err.Error())
		return err
	}
	return nil
}

// GetSubjects 查找科目一览
func GetSubjects(db string, appId, conditionType string, conditions []*Condition) (items []Subject, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SubjectCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建查询条件
	var query bson.M
	query = bson.M{"app_id": appId}
	if len(conditions) > 0 {
		// 如果有条件，继续构建查询
		if conditionType == "and" {
			var andConditions []bson.M
			for _, condition := range conditions {
				conditionQuery := bson.M{}
				switch condition.Operator {
				case "=":
					conditionQuery[condition.FieldID] = condition.SearchValue
				case "<>":
					conditionQuery[condition.FieldID] = bson.M{"$ne": condition.SearchValue}
				case "like":
					conditionQuery[condition.FieldID] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}}
				}
				andConditions = append(andConditions, conditionQuery)
			}
			query = bson.M{"$and": andConditions}
		} else if conditionType == "or" {
			var orConditions []bson.M
			for _, condition := range conditions {
				conditionQuery := bson.M{}
				switch condition.Operator {
				case "=":
					conditionQuery[condition.FieldID] = condition.SearchValue
				case "<>":
					conditionQuery[condition.FieldID] = bson.M{"$ne": condition.SearchValue}
				case "like":
					conditionQuery[condition.FieldID] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}}
				}
				orConditions = append(orConditions, conditionQuery)
			}
			query = bson.M{"$or": orConditions}
		}
	}

	// 执行查询
	cursor, err := c.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 创建一个map来处理重复的assets_type
	assetsTypeMap := make(map[string]bool)
	var result Subject

	// 遍历查询结果并填充返回项
	for cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		// 检查是否已存在相同的 assets_type
		if _, exists := assetsTypeMap[result.AssetsType]; !exists {
			items = append(items, result)
			assetsTypeMap[result.AssetsType] = true
		}
	}
	return items, nil
}
