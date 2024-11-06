package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/task/proto/history"
	"rxcsoft.cn/pit3/srv/task/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// HistoriesCollection histories collection
	HistoriesCollection = "task_histories"
)

type (
	// History 台账的数据
	History struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		JobID         string             `json:"job_id" bson:"job_id"`
		ScheduleID    string             `json:"schedule_id" bson:"schedule_id"`
		JobName       string             `json:"job_name" bson:"job_name"`
		Origin        string             `json:"origin" bson:"origin"`
		UserID        string             `json:"user_id" bson:"user_id"`
		Progress      int64              `json:"progress" bson:"progress"`
		StartTime     string             `json:"start_time" bson:"start_time"`
		EndTime       string             `json:"end_time" bson:"end_time"`
		Message       []Message          `json:"message" bson:"message"`
		ErrorFilePath string             `json:"error_file_path" bson:"error_file_path"`
		FilePath      string             `json:"file_path" bson:"file_path"`
		CurrentStep   string             `json:"current_step" bson:"current_step"`
		Steps         []string           `json:"steps" bson:"steps"`
		TaskType      string             `json:"task_type" bson:"task_type"`
		AppID         string             `json:"app_id" bson:"app_id"`
	}
	// Download 台账的数据
	Download struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		JobID         string             `json:"job_id" bson:"job_id"`
		ScheduleID    string             `json:"schedule_id" bson:"schedule_id"`
		JobName       string             `json:"job_name" bson:"job_name"`
		Origin        string             `json:"origin" bson:"origin"`
		UserID        string             `json:"user_id" bson:"user_id"`
		Progress      int64              `json:"progress" bson:"progress"`
		StartTime     string             `json:"start_time" bson:"start_time"`
		EndTime       string             `json:"end_time" bson:"end_time"`
		Message       string             `json:"message" bson:"message"`
		ErrorFilePath string             `json:"error_file_path" bson:"error_file_path"`
		FilePath      string             `json:"file_path" bson:"file_path"`
		CurrentStep   string             `json:"current_step" bson:"current_step"`
		Steps         []string           `json:"steps" bson:"steps"`
		TaskType      string             `json:"task_type" bson:"task_type"`
		AppID         string             `json:"app_id" bson:"app_id"`
	}

	Message struct {
		StartTime string `json:"start_time" bson:"start_time"`
		Message   string `json:"message" bson:"message"`
	}
)

// ToProto 转换为proto数据
func (h *Download) ToProto() *history.Download {
	return &history.Download{
		JobId:         h.JobID,
		ScheduleId:    h.ScheduleID,
		JobName:       h.JobName,
		Origin:        h.Origin,
		UserId:        h.UserID,
		Progress:      h.Progress,
		StartTime:     h.StartTime,
		EndTime:       h.EndTime,
		Message:       h.Message,
		CurrentStep:   h.CurrentStep,
		Steps:         h.Steps,
		ErrorFilePath: h.ErrorFilePath,
		FilePath:      h.FilePath,
		TaskType:      h.TaskType,
		AppId:         h.AppID,
	}
}

// ToProto 转换为proto数据
func (h *History) ToProto() *history.History {
	var messages []*history.Message
	for _, m := range h.Message {
		messages = append(messages, m.ToProto())
	}

	return &history.History{
		JobId:         h.JobID,
		ScheduleId:    h.ScheduleID,
		JobName:       h.JobName,
		Origin:        h.Origin,
		UserId:        h.UserID,
		Progress:      h.Progress,
		StartTime:     h.StartTime[:19],
		EndTime:       h.EndTime,
		Message:       messages,
		CurrentStep:   h.CurrentStep,
		Steps:         h.Steps,
		ErrorFilePath: h.ErrorFilePath,
		FilePath:      h.FilePath,
		TaskType:      h.TaskType,
		AppId:         h.AppID,
	}
}

// ToProto 转换为proto数据
func (h *Message) ToProto() *history.Message {
	return &history.Message{
		StartTime: h.StartTime[:19],
		Message:   h.Message,
	}
}

// DownloadHistories 下载履历数据
func DownloadHistories(db, userID, appID, scheduleID, jobID string) (items []Download, total int64, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(HistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id": appID,
	}

	if userID != "" {
		query["user_id"] = userID
	}

	if scheduleID != "" {
		query["schedule_id"] = scheduleID
	}

	if jobID != "" {
		query["job_id"] = jobID
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindHistories", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Download
	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		return result, 0, err
	}

	// 聚合查询
	opts := options.Find().SetSort(bson.D{{Key: "job_id", Value: -1}})
	cur, err := c.Find(ctx, query, opts)

	if err != nil {
		utils.ErrorLog("error FindHistories", err.Error())
		return result, 0, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var d Download
		err := cur.Decode(&d)
		if err != nil {
			utils.ErrorLog("error FindHistories", err.Error())
			return result, 0, err
		}
		result = append(result, d)
	}
	return result, t, nil
}

// FindHistories 获取履历数据
func FindHistories(db, userID, appID, scheduleID, jobID string, pageIndex, pageSize int64) (items []History, total int64, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(HistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	match := bson.M{
		"app_id": appID,
	}

	if userID != "" {
		match["user_id"] = userID
	}

	if scheduleID != "" {
		match["schedule_id"] = scheduleID
	}

	if jobID != "" {
		match["job_id"] = jobID
	}

	var result []History
	// 取总件数
	t := make(chan int64, 1)
	go func() {
		getTotal(db, userID, appID, scheduleID, jobID, t)
	}()
	// 聚合查询

	pipe := []bson.M{
		{"$match": match},
	}

	pipe = append(pipe, bson.M{
		"$sort": bson.M{
			"job_id":     1,
			"start_time": 1,
		},
	})

	limit := pageSize
	skip := (pageIndex - 1) * pageSize

	group := bson.M{
		"_id": bson.M{
			"job_id": "$job_id",
		},
		"app_id": bson.M{
			"$last": "$app_id",
		},
		"job_name": bson.M{
			"$last": "$job_name",
		},
		"origin": bson.M{
			"$last": "$origin",
		},
		"user_id": bson.M{
			"$last": "$user_id",
		},
		"task_type": bson.M{
			"$last": "$task_type",
		},
		"progress": bson.M{
			"$last": "$progress",
		},
		"start_time": bson.M{
			"$first": "$start_time",
		},
		"end_time": bson.M{
			"$last": "$end_time",
		},
		"error_file_path": bson.M{
			"$last": "$error_file_path",
		},
		"file_path": bson.M{
			"$last": "$file_path",
		},
		"message": bson.M{
			"$push": bson.M{
				"start_time": "$start_time",
				"message":    "$message",
			},
		},
		"current_step": bson.M{
			"$last": "$current_step",
		},
		"steps": bson.M{
			"$last": "$steps",
		},
	}

	pipe = append(pipe, bson.M{
		"$group": group,
	})

	pipe = append(pipe, bson.M{
		"$sort": bson.M{
			"start_time": -1,
		},
	})

	if skip > 0 {
		pipe = append(pipe, bson.M{
			"$skip": skip,
		})
	}
	if limit > 0 {
		pipe = append(pipe, bson.M{
			"$limit": limit,
		})
	}

	project := bson.M{
		"_id":             0,
		"job_id":          "$_id.job_id",
		"app_id":          "$app_id",
		"job_name":        "$job_name",
		"origin":          "$origin",
		"task_type":       "$task_type",
		"user_id":         "$user_id",
		"progress":        "$progress",
		"start_time":      "$start_time",
		"end_time":        "$end_time",
		"error_file_path": "$error_file_path",
		"file_path":       "$file_path",
		"message":         "$message",
		"current_step":    "$current_step",
		"steps":           "$steps",
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindHistories", fmt.Sprintf("query: [ %s ]", queryJSON))

	opts := options.Aggregate().SetAllowDiskUse(true)
	cur, err := c.Aggregate(ctx, pipe, opts)
	if err != nil {
		utils.ErrorLog("error FindHistories", err.Error())
		return result, 0, err
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var h History
		err := cur.Decode(&h)
		if err != nil {
			utils.ErrorLog("error FindHistories", err.Error())
			return result, 0, err
		}
		result = append(result, h)
	}
	return result, <-t, nil
}

func getTotal(db, userID, appID, scheduleID, jobID string, total chan int64) {
	client := database.New()
	c := client.Database(database.Db).Collection(HistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	match := bson.M{
		"app_id": appID,
	}

	if userID != "" {
		match["user_id"] = userID
	}

	if scheduleID != "" {
		match["schedule_id"] = scheduleID
	}

	if jobID != "" {
		match["job_id"] = jobID
	}

	query := []bson.M{
		{"$match": match},
	}

	group := []bson.M{
		{"$group": bson.M{
			"_id": bson.M{
				"job_id": "$job_id",
			},
		}},
		{"$count": "total"},
	}

	query = append(query, group...)

	type Result struct {
		Total int64 `bson:"total"`
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("getTotal", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Result
	// 取总件数
	cur, err := c.Aggregate(ctx, query)
	if err != nil {
		total <- 0
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		err := cur.Decode(&result)
		if err != nil {
			utils.ErrorLog("error getTotal", err.Error())
			return
		}
	}

	total <- result.Total
}

// AddHistory 添加任务履历数据
func AddHistory(db string, h *Download) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(HistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	h.ID = primitive.NewObjectID()

	queryJSON, _ := json.Marshal(h)
	utils.DebugLog("AddHistory", fmt.Sprintf("History: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, h)
	if err != nil {
		utils.ErrorLog("error AddHistory", err.Error())
		return err
	}
	return nil
}
