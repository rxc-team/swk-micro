package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/manage/proto/action"
	"rxcsoft.cn/pit3/srv/manage/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	ActionsCollection = "actions"
)

// Action 许可操作
type Action struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	ActionKey    string             `json:"action_key" bson:"action_key"`
	ActionName   map[string]string  `json:"action_name" bson:"action_name"`
	ActionObject string             `json:"action_object" bson:"action_object"`
	ActionGroup  string             `json:"action_group" bson:"action_group"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy    string             `json:"created_by" bson:"created_by"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy    string             `json:"updated_by" bson:"updated_by"`
}

// ActionDelParam 许可操作删除参数
type ActionDelParam struct {
	ActionKey    string `json:"action_key" bson:"action_key"`
	ActionObject string `json:"action_object" bson:"action_object"`
}

// ToProto 转换为proto数据
func (m *Action) ToProto() *action.Action {
	return &action.Action{
		ActionKey:    m.ActionKey,
		ActionName:   m.ActionName,
		ActionObject: m.ActionObject,
		ActionGroup:  m.ActionGroup,
		CreatedAt:    m.CreatedAt.String(),
		CreatedBy:    m.CreatedBy,
		UpdatedAt:    m.UpdatedAt.String(),
		UpdatedBy:    m.UpdatedBy,
	}
}

// FindActions 查找多个许可操作记录
func FindActions(ctx context.Context, actionGroup string) (m []Action, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(ActionsCollection)

	query := bson.M{}

	if len(actionGroup) > 0 {
		query["action_group"] = actionGroup
	}

	var result []Action
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindActions", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindActions", err.Error())
		return nil, err
	}

	return result, nil

}

// FindAction 查找单个许可操作记录
func FindAction(ctx context.Context, actionKey, actionObject string) (cus Action, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(ActionsCollection)

	var result Action

	query := bson.M{
		"action_key":    actionKey,
		"action_object": actionObject,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindAction", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindAction", err.Error())
		return result, err
	}

	return result, nil
}

// AddAction 添加单个许可操作记录
func AddAction(ctx context.Context, m *Action) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(ActionsCollection)

	m.ID = primitive.NewObjectID()

	queryJSON, _ := json.Marshal(m)
	utils.DebugLog("FindDeleteAction", fmt.Sprintf("Action: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, m)
	if err != nil {
		utils.ErrorLog("error AddAction", err.Error())
		return err
	}

	return nil
}

// ModifyAction 修改单个许可操作记录
func ModifyAction(ctx context.Context, p Action) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(ActionsCollection)

	query := bson.M{
		"action_key":    p.ActionKey,
		"action_object": p.ActionObject,
	}

	change := bson.M{
		"updated_at": p.UpdatedAt,
		"updated_by": p.UpdatedBy,
	}

	change["action_name"] = p.ActionName

	if len(p.ActionGroup) > 0 {
		change["action_group"] = p.ActionGroup
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyAction", fmt.Sprintf("Action: [ %s ]", queryJSON))
	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyAction", fmt.Sprintf("Action: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyAction", err.Error())
		return err
	}

	return nil
}

// DeleteAction 硬删除单个许可操作记录
func DeleteAction(ctx context.Context, actionKey, actionObject string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(ActionsCollection)

	query := bson.M{
		"action_key":    actionKey,
		"action_object": actionObject,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteAction", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteAction", err.Error())
		return err
	}

	return nil
}

// DeleteActions 硬删除多个许可操作记录
func DeleteActions(ctx context.Context, dels []*ActionDelParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(ActionsCollection)

	for _, del := range dels {
		query := bson.M{
			"action_key":    del.ActionKey,
			"action_object": del.ActionObject,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("DeleteActions", fmt.Sprintf("query: [ %s ]", queryJSON))

		_, err = c.DeleteOne(ctx, query)
		if err != nil {
			utils.ErrorLog("error DeleteActions", err.Error())
			return err
		}
	}

	return nil
}
