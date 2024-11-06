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
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"rxcsoft.cn/pit3/srv/workflow/proto/process"
	"rxcsoft.cn/pit3/srv/workflow/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// ProcessCollection schedule collection
	ProcessCollection = "wf_process"
)

type (
	// Process 进程
	Process struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		ProcessID   string             `json:"pro_id" bson:"pro_id"`
		ExampleID   string             `json:"ex_id" bson:"ex_id"`
		CurrentNode string             `json:"current_node" bson:"current_node"`
		UserID      string             `json:"user_id" bson:"user_id"`
		ExpireDate  string             `json:"expire_date" bson:"expire_date"`
		Comment     string             `json:"comment" bson:"comment"`
		Status      int64              `json:"status" bson:"status"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	}
)

// ToProto 转换为proto数据
func (p *Process) ToProto() *process.Process {
	return &process.Process{
		ProId:       p.ProcessID,
		ExId:        p.ExampleID,
		CurrentNode: p.CurrentNode,
		UserId:      p.UserID,
		ExpireDate:  p.ExpireDate,
		Status:      p.Status,
		Comment:     p.Comment,
		CreatedAt:   p.CreatedAt.String(),
		CreatedBy:   p.CreatedBy,
		UpdatedAt:   p.UpdatedAt.String(),
		UpdatedBy:   p.UpdatedBy,
	}
}

// FindProcesses 获取进程数据
func FindProcesses(db, exID string) (items []Process, err error) {
	client := database.New()
	opt := options.Collection()
	opt.SetReadPreference(readpref.Primary())
	c := client.Database(database.GetDBName(db)).Collection(ProcessCollection, opt)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"ex_id": exID,
	}

	var result []Process
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	processes, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindProcesss", err.Error())
		return nil, err
	}
	defer processes.Close(ctx)
	for processes.Next(ctx) {
		var pcs Process
		err := processes.Decode(&pcs)
		if err != nil {
			utils.ErrorLog("error FindProcesss", err.Error())
			return nil, err
		}
		result = append(result, pcs)
	}

	return result, nil
}

// FindsProcesses 获取所有进程数据
func FindsProcesses(userID, db string) (items []Process, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ProcessCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"user_id": userID,
		"status":  0,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindsProcesss", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Process
	processes, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindsProcesss", err.Error())
		return nil, err
	}
	defer processes.Close(ctx)
	for processes.Next(ctx) {
		var pcs Process
		err := processes.Decode(&pcs)
		if err != nil {
			utils.ErrorLog("error FindsProcesss", err.Error())
			return nil, err
		}
		result = append(result, pcs)
	}
	return result, nil
}

// AddProcess 添加进程数据
func AddProcess(db string, s *Process) (scheduleID string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ProcessCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.ID = primitive.NewObjectID()
	s.ProcessID = s.ID.Hex()

	_, err = c.InsertOne(ctx, s)
	if err != nil {
		utils.ErrorLog("error AddProcess", err.Error())
		return "", err
	}

	return s.ProcessID, nil
}

// ModifyProcess 更新进程数据
func ModifyProcess(db, proID, status, comment, writer string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ProcessCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(proID)
	if err != nil {
		utils.ErrorLog("error ModifyProcess", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	st, _ := strconv.Atoi(status)
	change := bson.M{
		"status":     st,
		"comment":    comment,
		"updated_at": time.Now(),
		"updated_by": writer,
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyProcess", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyProcess", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyProcess", err.Error())
		return err
	}

	return nil
}

// DeleteProcess 删除进程数据
func DeleteProcess(db string, exID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ProcessCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"ex_id": exID,
	}
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteProcess", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteProcess", err.Error())
		return err
	}
	return nil
}
