package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/manage/proto/allow"
	"rxcsoft.cn/pit3/srv/manage/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	AllowsCollection = "allows"
)

// Allow 许可操作
type Allow struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	AllowId    string             `json:"allow_id" bson:"allow_id"`
	AllowName  string             `json:"allow_name" bson:"allow_name"`
	AllowType  string             `json:"allow_type" bson:"allow_type"`
	ObjectType string             `json:"object_type" bson:"object_type"`
	Actions    []*AAction         `json:"actions" bson:"actions"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy  string             `json:"created_by" bson:"created_by"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy  string             `json:"updated_by" bson:"updated_by"`
}

// AAction 许可操作
type AAction struct {
	ApiKey     string `json:"api_key" bson:"api_key"`
	GroupKey   string `json:"group_key" bson:"group_key"`
	ActionName string `json:"action_name" bson:"action_name"`
}

// ToProto 转换为proto数据
func (m *Allow) ToProto() *allow.Allow {
	var actions []*allow.Action
	for _, a := range m.Actions {
		actions = append(actions, &allow.Action{
			ApiKey:     a.ApiKey,
			GroupKey:   a.GroupKey,
			ActionName: a.ActionName,
		})
	}

	return &allow.Allow{
		AllowId:    m.AllowId,
		AllowName:  m.AllowName,
		AllowType:  m.AllowType,
		ObjectType: m.ObjectType,
		Actions:    actions,
		CreatedAt:  m.CreatedAt.String(),
		CreatedBy:  m.CreatedBy,
		UpdatedAt:  m.UpdatedAt.String(),
		UpdatedBy:  m.UpdatedBy,
	}
}

// FindAllows 查找多个许可操作记录
func FindAllows(ctx context.Context, allowType, objectType string) (m []Allow, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(AllowsCollection)

	query := bson.M{}

	if len(allowType) > 0 {
		query["allow_type"] = allowType
	}
	if len(objectType) > 0 {
		query["object_type"] = objectType
	}

	var result []Allow
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindAllows", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindAllows", err.Error())
		return nil, err
	}

	return result, nil

}

// FindLevelAllows 查找多个许可操作记录
func FindLevelAllows(ctx context.Context, allows []string) (m []Allow, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(AllowsCollection)

	query := bson.M{
		"allow_id": bson.M{
			"$in": allows,
		},
	}

	var result []Allow

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindAllows", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindAllows", err.Error())
		return nil, err
	}

	return result, nil

}

// FindAllow 查找单个许可操作记录
func FindAllow(ctx context.Context, allowId string) (cus Allow, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(AllowsCollection)

	var result Allow

	query := bson.M{
		"allow_id": allowId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindAllow", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindAllow", err.Error())
		return result, err
	}

	return result, nil
}

// AddAllow 添加单个许可操作记录
func AddAllow(ctx context.Context, m *Allow) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(AllowsCollection)

	m.ID = primitive.NewObjectID()
	m.AllowId = m.ID.Hex()

	queryJSON, _ := json.Marshal(m)
	utils.DebugLog("FindDeleteAllow", fmt.Sprintf("Allow: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, m)
	if err != nil {
		utils.ErrorLog("error AddAllow", err.Error())
		return err
	}

	return nil
}

// ModifyAllow 修改单个许可操作记录
func ModifyAllow(ctx context.Context, allowId, allowName, allowType, objType, userID string, actions []AAction) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(AllowsCollection)

	query := bson.M{
		"allow_id": allowId,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": userID,
	}

	if len(allowName) > 0 {
		change["allow_name"] = allowName
	}
	if len(allowType) > 0 {
		change["allow_type"] = allowType
	}
	if len(objType) > 0 {
		change["object_type"] = objType
	}
	if len(actions) > 0 {
		change["actions"] = actions
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyAllow", fmt.Sprintf("Allow: [ %s ]", queryJSON))
	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyAllow", fmt.Sprintf("Allow: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyAllow", err.Error())
		return err
	}

	return nil
}

// DeleteAllow 硬删除单个许可操作记录
func DeleteAllow(ctx context.Context, allowId string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(AllowsCollection)

	query := bson.M{
		"allow_id": allowId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteAllow", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteAllow", err.Error())
		return err
	}

	return nil
}

// DeleteAllows 硬删除多个许可操作记录
func DeleteAllows(ctx context.Context, allowIds []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(AllowsCollection)

	for _, allowId := range allowIds {
		query := bson.M{
			"allow_id": allowId,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("DeleteAllows", fmt.Sprintf("query: [ %s ]", queryJSON))

		_, err = c.DeleteOne(ctx, query)
		if err != nil {
			utils.ErrorLog("error DeleteAllows", err.Error())
			return err
		}
	}

	return nil
}
