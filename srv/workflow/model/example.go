package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/workflow/proto/example"
	"rxcsoft.cn/pit3/srv/workflow/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// ExampleCollection schedule collection
	ExampleCollection = "wf_examples"
)

type (
	// Example 流程实例
	Example struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		ExampleID   string             `json:"ex_id" bson:"ex_id"`
		WorkflowID  string             `json:"wf_id" bson:"wf_id"`
		ExampleName string             `json:"ex_name" bson:"ex_name"`
		UserID      string             `json:"user_id" bson:"user_id"`
		Status      int64              `json:"status" bson:"status"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	}
)

// ToProto 转换为proto数据
func (w *Example) ToProto() *example.Example {
	return &example.Example{
		ExId:      w.ExampleID,
		WfId:      w.WorkflowID,
		ExName:    w.ExampleName,
		UserId:    w.UserID,
		Status:    w.Status,
		CreatedAt: w.CreatedAt.String(),
		CreatedBy: w.CreatedBy,
		UpdatedAt: w.UpdatedAt.String(),
		UpdatedBy: w.UpdatedBy,
	}
}

// FindExamples 获取流程实例数据
func FindExamples(db, wfID string) (items []Example, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ExampleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"wf_id": wfID,
	}

	var result []Example

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	examples, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindExamples", err.Error())
		return nil, err
	}
	defer examples.Close(ctx)
	for examples.Next(ctx) {
		var exp Example
		err := examples.Decode(&exp)
		if err != nil {
			utils.ErrorLog("error FindExamples", err.Error())
			return nil, err
		}
		result = append(result, exp)
	}

	return result, nil
}

// FindExample 获取流程实例数据
func FindExample(db, exampleID string) (items Example, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ExampleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"ex_id": exampleID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindExample", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Example

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindExample", err.Error())
		return result, err
	}

	return result, nil
}

// AddExample 添加流程实例数据
func AddExample(db string, s *Example) (scheduleID string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ExampleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.ID = primitive.NewObjectID()
	s.ExampleID = s.ID.Hex()

	_, err = c.InsertOne(ctx, s)
	if err != nil {
		utils.ErrorLog("error AddExample", err.Error())
		return "", err
	}

	return s.ExampleID, nil
}

// ModifyExample 更新流程实例数据
func ModifyExample(db, exID, status, writer string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ExampleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(exID)
	if err != nil {
		utils.ErrorLog("error ModifyExample", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	st, _ := strconv.Atoi(status)
	change := bson.M{
		"status":     st,
		"updated_at": time.Now(),
		"updated_by": writer,
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyExample", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyExample", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyExample", err.Error())
		return err
	}

	return nil
}

// DeleteExample 删除流程实例数据
func DeleteExample(db string, exID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ExampleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(exID)
	if err != nil {
		utils.ErrorLog("error DeleteExample", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteExample", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteExample", err.Error())
		return err
	}
	return nil
}
