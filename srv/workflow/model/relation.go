package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"rxcsoft.cn/pit3/srv/workflow/proto/relation"
	"rxcsoft.cn/pit3/srv/workflow/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// RelationCollection schedule collection
	RelationCollection = "wf_relations"
)

type (
	// Relation 流程设置
	Relation struct {
		ID         primitive.ObjectID `json:"id" bson:"_id"`
		AppId      string             `json:"app_id" bson:"app_id"`
		ObjectId   string             `json:"object_id" bson:"object_id"`
		GroupId    string             `json:"group_id" bson:"group_id"`
		WorkflowId string             `json:"workflow_id" bson:"workflow_id"`
		Action     string             `json:"action" bson:"action"`
	}
)

// ToProto 转换为proto数据
func (f *Relation) ToProto() *relation.Relation {
	return &relation.Relation{
		AppId:      f.AppId,
		ObjectId:   f.ObjectId,
		GroupId:    f.GroupId,
		WorkflowId: f.WorkflowId,
		Action:     f.Action,
	}
}

// FindRelations 获取流程设置数据
func FindRelations(db, appId, objectId, groupId, workflowId, action string) (items []Relation, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RelationCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":    appId,
		"object_id": appId,
		"group_id":  appId,
	}

	if len(workflowId) > 0 {
		query["workflow_id"] = workflowId
	}

	if len(action) > 0 {
		query["action"] = action
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindRelation", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Relation

	fr, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindRelations", err.Error())
		return nil, err
	}
	defer fr.Close(ctx)
	for fr.Next(ctx) {
		var fo Relation
		err := fr.Decode(&fo)
		if err != nil {
			utils.ErrorLog("error FindRelations", err.Error())
			return nil, err
		}
		result = append(result, fo)
	}

	return result, nil
}

// AddRelation 添加流程设置数据
func AddRelation(db string, s *Relation) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RelationCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.ID = primitive.NewObjectID()

	_, err = c.InsertOne(ctx, s)
	if err != nil {
		utils.ErrorLog("error AddRelation", err.Error())
		return err
	}

	return nil
}

// DeleteRelation 删除流程设置数据
func DeleteRelation(db, appId, objectId, groupId, workflowId string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RelationCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appId,
	}

	if len(objectId) > 0 {
		query["object_id"] = objectId
	}
	if len(groupId) > 0 {
		query["group_id"] = groupId
	}
	if len(workflowId) > 0 {
		query["workflow_id"] = workflowId
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteRelation", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteRelation", err.Error())
		return err
	}
	return nil
}
