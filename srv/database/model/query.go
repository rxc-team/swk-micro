package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"rxcsoft.cn/pit3/srv/database/proto/query"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	QueriesCollection = "queries"
)

type (
	// Query 快捷方式
	Query struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		QueryID       string             `json:"query_id" bson:"query_id"`
		UserID        string             `json:"user_id" bson:"user_id"`
		DatastoreID   string             `json:"datastore_id" bson:"datastore_id"`
		AppID         string             `json:"app_id" bson:"app_id"`
		QueryName     string             `json:"query_name" bson:"query_name"`
		Description   string             `json:"description" bson:"description"`
		ConditionType string             `json:"condition_type" bson:"condition_type"`
		Conditions    []*Condition       `json:"conditions,omitempty" bson:"conditions"`
		Fields        []string           `json:"fields" bson:"fields"`
		CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy     string             `json:"created_by" bson:"created_by"`
		UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy     string             `json:"updated_by" bson:"updated_by"`
		DeletedAt     time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy     string             `json:"deleted_by" bson:"deleted_by"`
	}
)

// ToProto 转换为proto数据
func (q *Query) ToProto() *query.Query {
	var conditions []*query.Condition

	for _, ch := range q.Conditions {
		conditions = append(conditions, ch.ToProto())
	}

	return &query.Query{
		QueryId:       q.QueryID,
		UserId:        q.UserID,
		DatastoreId:   q.DatastoreID,
		AppId:         q.AppID,
		QueryName:     q.QueryName,
		Description:   q.Description,
		ConditionType: q.ConditionType,
		Conditions:    conditions,
		Fields:        q.Fields,
		CreatedAt:     q.CreatedAt.String(),
		CreatedBy:     q.CreatedBy,
		UpdatedAt:     q.UpdatedAt.String(),
		UpdatedBy:     q.UpdatedBy,
		DeletedAt:     q.DeletedAt.String(),
		DeletedBy:     q.DeletedBy,
	}
}

// ToProto 转换为proto数据
func (c *Condition) ToProto() *query.Condition {
	return &query.Condition{
		FieldId:       c.FieldID,
		FieldType:     c.FieldType,
		SearchValue:   c.SearchValue,
		Operator:      c.Operator,
		ConditionType: c.ConditionType,
		IsDynamic:     c.IsDynamic,
	}
}

// FindQueries 获取所有的快捷方式
func FindQueries(db, userID, appID, datastoreID, queryName string) (q []Query, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(QueriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
	}

	// 快捷方式所属用户不为空的场合，添加到查询条件中
	if userID != "" {
		query["user_id"] = userID
	}

	// 快捷方式所属APP不为空的场合，添加到查询条件中
	if appID != "" {
		query["app_id"] = appID
	}

	// 快捷方式所属台账不为空的场合，添加到查询条件中
	if datastoreID != "" {
		query["datastore_id"] = datastoreID
	}

	// 快捷方式名称不为空的场合，添加到查询条件中
	if queryName != "" {
		query["query_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(queryName), Options: "m"}}
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindQueries", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Query
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("FindQueries", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var que Query
		err := cur.Decode(&que)
		if err != nil {
			utils.ErrorLog("FindQueries", err.Error())
			return nil, err
		}
		result = append(result, que)
	}

	return result, nil
}

// FindQuery 通过ID获取快捷方式信息
func FindQuery(db, queryID string) (q Query, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(QueriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Query

	objectID, err := primitive.ObjectIDFromHex(queryID)
	if err != nil {
		utils.ErrorLog("FindQuery", err.Error())
		return result, err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindQuery", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("FindQuery", err.Error())
		return result, err
	}

	return result, nil
}

// AddQuery 添加快捷方式
func AddQuery(db string, r *Query) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(QueriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r.ID = primitive.NewObjectID()
	r.QueryID = r.ID.Hex()

	queryJSON, _ := json.Marshal(r)
	utils.DebugLog("AddQuery", fmt.Sprintf("Query: [ %s ]", queryJSON))

	if _, err = c.InsertOne(ctx, r); err != nil {
		utils.ErrorLog("FindQuery", err.Error())
		return "", err
	}

	return r.QueryID, nil
}

// DeleteQuery 删除快捷方式
func DeleteQuery(db, queryID, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(QueriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(queryID)
	if err != nil {
		utils.ErrorLog("DeleteQuery", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteQuery", fmt.Sprintf("query: [ %s ]", queryJSON))

	if _, err := c.DeleteOne(ctx, query); err != nil {
		utils.ErrorLog("DeleteQuery", err.Error())
		return err
	}

	return nil
}

// DeleteSelectQueries 删除多个query
func DeleteSelectQueries(db string, queryIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(QueriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteSelectQueries", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteSelectQueries", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, queryID := range queryIDList {
			objectID, err := primitive.ObjectIDFromHex(queryID)
			if err != nil {
				utils.ErrorLog("DeleteSelectQueries", err.Error())
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
			utils.DebugLog("DeleteSelectQueries", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectQueries", fmt.Sprintf("update: [ %s ]", updateJSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("DeleteSelectQueries", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteSelectQueries", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteSelectQueries", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// HardDeleteQueries 物理删除快捷方式
func HardDeleteQueries(db string, queryIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(QueriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("HardDeleteQueries", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("HardDeleteQueries", err.Error())
		return err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, queryID := range queryIDList {
			objectID, err := primitive.ObjectIDFromHex(queryID)
			if err != nil {
				utils.ErrorLog("HardDeleteQueries", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteQueries", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("HardDeleteQueries", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("HardDeleteQueries", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("HardDeleteQueries", err.Error())
		return err
	}

	session.EndSession(ctx)
	return nil
}
