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
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/pit3/srv/task/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// TaskCollection task collection
	TaskCollection = "tasks"
)

type (
	// Task 任务
	Task struct {
		ID           primitive.ObjectID `json:"id" bson:"_id"`
		JobID        string             `json:"job_id" bson:"job_id"`
		ScheduleID   string             `json:"schedule_id" bson:"schedule_id"`
		JobName      string             `json:"job_name" bson:"job_name"`
		Origin       string             `json:"origin" bson:"origin"`
		UserID       string             `json:"user_id" bson:"user_id"`
		ShowProgress bool               `json:"show_progress" bson:"show_progress"`
		Progress     int64              `json:"progress" bson:"progress"`
		StartTime    string             `json:"start_time" bson:"start_time"`
		EndTime      string             `json:"end_time" bson:"end_time"`
		Message      string             `json:"message" bson:"message"`
		File         File               `json:"file" bson:"file"`
		ErrorFile    File               `json:"error_file" bson:"error_file"`
		TaskType     string             `json:"task_type" bson:"task_type"`
		Steps        []string           `json:"steps" bson:"steps"`
		CurrentStep  string             `json:"current_step" bson:"current_step"`
		AppID        string             `json:"app_id" bson:"app_id"`
		Insert       int64              `json:"insert" bson:"insert"`
		Update       int64              `json:"update" bson:"update"`
		Total        int64              `json:"total" bson:"total"`
	}
	// File 文件
	File struct {
		URL  string `json:"url" bson:"url"`
		Name string `json:"name" bson:"name"`
	}
)

// ToProto 转换为proto数据
func (t *Task) ToProto() *task.Task {
	return &task.Task{
		JobId:        t.JobID,
		ScheduleId:   t.ScheduleID,
		JobName:      t.JobName,
		Origin:       t.Origin,
		UserId:       t.UserID,
		ShowProgress: t.ShowProgress,
		Progress:     t.Progress,
		StartTime:    t.StartTime,
		EndTime:      t.EndTime,
		Message:      t.Message,
		File:         t.File.ToProto(),
		ErrorFile:    t.ErrorFile.ToProto(),
		TaskType:     t.TaskType,
		Steps:        t.Steps,
		CurrentStep:  t.CurrentStep,
		AppId:        t.AppID,
		Insert:       t.Insert,
		Update:       t.Update,
		Total:        t.Total,
	}
}

// ToProto 转换为proto数据
func (f *File) ToProto() *task.File {
	return &task.File{
		Url:  f.URL,
		Name: f.Name,
	}
}

// FindTasks 获取任务数据
func FindTasks(db, userID, scheduleID, appID string, pageIndex, pageSize int64) (items []Task, total int64, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(TaskCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"user_id": userID,
		"app_id":  appID,
	}

	if scheduleID != "" {
		query["schedule_id"] = scheduleID
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindTasks", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Task
	// 取总件数
	count, err := c.CountDocuments(ctx, query)
	if err != nil {
		return result, 0, err
	}

	if pageIndex != 0 && pageSize != 0 {
		opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}}).SetSkip((pageIndex - 1) * pageSize).SetLimit(pageSize)
		tasts, err := c.Find(ctx, query, opts)
		if err != nil {
			utils.ErrorLog("error FindTasks", err.Error())
			return nil, 0, err
		}
		defer tasts.Close(ctx)
		for tasts.Next(ctx) {
			var tas Task
			err := tasts.Decode(&tas)
			if err != nil {
				utils.ErrorLog("error FindTasks", err.Error())
				return nil, 0, err
			}
			result = append(result, tas)
		}
		return result, count, nil
	}

	opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}})
	tasts, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindTasks", err.Error())
		return nil, 0, err
	}
	defer tasts.Close(ctx)
	for tasts.Next(ctx) {
		var tas Task
		err := tasts.Decode(&tas)
		if err != nil {
			utils.ErrorLog("error FindTasks", err.Error())
			return nil, 0, err
		}
		result = append(result, tas)
	}

	return result, count, nil
}

// FindTask 获取任务数据
func FindTask(db, jobID string) (items Task, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(TaskCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"job_id": jobID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindTask", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Task

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindTask", err.Error())
		return result, err
	}

	return result, nil
}

// AddTask 添加任务数据
func AddTask(db string, t *Task) (id string, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(TaskCollection)
	hc := client.Database(database.Db).Collection(HistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建集合，防止出错。
	client.Database(database.Db).CreateCollection(ctx, TaskCollection)
	client.Database(database.Db).CreateCollection(ctx, HistoriesCollection)

	query := bson.M{
		"job_id": t.JobID,
	}

	var result Task

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {

			t.ID = primitive.NewObjectID()
			t.StartTime = time.Now().UTC().Format("2006-01-02 15:04:05.000000")

			_, err = c.InsertOne(ctx, t)
			if err != nil {
				utils.ErrorLog("AddTask", err.Error())
				return "", err
			}

			h := Download{
				ID:          primitive.NewObjectID(),
				JobID:       t.JobID,
				ScheduleID:  t.ScheduleID,
				JobName:     t.JobName,
				Origin:      t.Origin,
				UserID:      t.UserID,
				Progress:    t.Progress,
				StartTime:   time.Now().UTC().Format("2006-01-02 15:04:05.000000"),
				Message:     t.Message,
				TaskType:    t.TaskType,
				CurrentStep: t.CurrentStep,
				Steps:       t.Steps,
				AppID:       t.AppID,
			}

			queryJSON, _ := json.Marshal(h)
			utils.DebugLog("AddHistory", fmt.Sprintf("History: [ %s ]", queryJSON))

			_, err = hc.InsertOne(ctx, h)
			if err != nil {
				utils.ErrorLog("AddTask", err.Error())
				return "", err
			}

			return t.ID.Hex(), nil
		}
	}
	return "", nil
}

// ModifyTask 更新任务数据
func ModifyTask(db string, t *Task) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(TaskCollection)
	hc := client.Database(database.Db).Collection(HistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"job_id": t.JobID,
	}

	change := bson.M{}
	if t.Progress != 0 {
		change["progress"] = t.Progress
	}
	if t.EndTime != "" {
		change["end_time"] = t.EndTime
	}
	if t.Message != "" {
		change["message"] = t.Message
	}
	if t.File.URL != "" {
		change["file"] = t.File
	}
	if t.ErrorFile.URL != "" {
		change["error_file"] = t.ErrorFile
	}
	if t.CurrentStep != "" {
		change["current_step"] = t.CurrentStep
	}
	if t.Total != 0 {
		change["total"] = t.Total
	}
	if t.Insert != 0 {
		change["insert"] = t.Insert
	}
	if t.Update != 0 {
		change["update"] = t.Update
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyTask", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyTask", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("ModifyTask", err.Error())
		return err
	}

	tk, err := FindTask(db, t.JobID)
	if err != nil {
		utils.ErrorLog("ModifyTask", err.Error())
		return err
	}

	h := Download{
		ID:            primitive.NewObjectID(),
		JobID:         tk.JobID,
		ScheduleID:    tk.ScheduleID,
		JobName:       tk.JobName,
		Origin:        tk.Origin,
		TaskType:      tk.TaskType,
		UserID:        tk.UserID,
		Progress:      t.Progress,
		StartTime:     time.Now().UTC().Format("2006-01-02 15:04:05.000000"),
		EndTime:       t.EndTime,
		Message:       t.Message,
		ErrorFilePath: t.ErrorFile.URL,
		FilePath:      t.File.URL,
		CurrentStep:   t.CurrentStep,
		Steps:         tk.Steps,
		AppID:         tk.AppID,
	}

	queryJSON1, _ := json.Marshal(h)
	utils.DebugLog("AddHistory", fmt.Sprintf("History: [ %s ]", queryJSON1))

	_, err = hc.InsertOne(ctx, h)
	if err != nil {
		utils.ErrorLog("ModifyTask", err.Error())
		return err
	}

	return nil
}

// DeleteTask 删除任务数据
func DeleteTask(appID, userID, jobID string) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(TaskCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.DebugLog("DeleteTask", fmt.Sprintf("job_id: [ %s ]", jobID))
	_, err = c.DeleteOne(ctx, bson.M{"app_id": appID, "user_id": userID, "job_id": jobID})
	if err != nil {
		utils.ErrorLog("error DeleteTask", err.Error())
		return err
	}

	return nil
}
