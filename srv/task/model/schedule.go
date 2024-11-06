package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	"rxcsoft.cn/pit3/srv/task/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// ScheduleCollection schedule collection
	ScheduleCollection = "schedules"
)

type (
	// Schedule 任务
	Schedule struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		ScheduleID    string             `json:"schedule_id" bson:"schedule_id"`
		ScheduleName  string             `json:"schedule_name" bson:"schedule_name"`
		EntryID       int64              `json:"entry_id" bson:"entry_id"`
		Spec          string             `json:"spec" bson:"spec"`
		Multi         int64              `json:"multi" bson:"multi"`
		RetryTimes    int64              `json:"retry_times" bson:"retry_times"`
		RetryInterval int64              `json:"retry_interval" bson:"retry_interval"`
		StartTime     string             `json:"start_time" bson:"start_time"`
		EndTime       string             `json:"end_time" bson:"end_time"`
		ScheduleType  string             `json:"schedule_type" bson:"schedule_type"`
		RunNow        bool               `json:"run_now" bson:"run_now"`
		Status        string             `json:"status" bson:"status"`
		Params        map[string]string  `json:"params" bson:"params"`
		CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy     string             `json:"created_by" bson:"created_by"`
		UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy     string             `json:"updated_by" bson:"updated_by"`
	}

	// ModifyParam 更新参数
	ModifyParam struct {
		ScheduleID string
		EntryID    string
		Status     string
		Database   string
		Writer     string
	}
)

// ToProto 转换为proto数据
func (s *Schedule) ToProto() *schedule.Schedule {
	return &schedule.Schedule{
		ScheduleId:    s.ScheduleID,
		ScheduleName:  s.ScheduleName,
		EntryId:       s.EntryID,
		Spec:          s.Spec,
		Multi:         s.Multi,
		RetryTimes:    s.RetryTimes,
		RetryInterval: s.RetryInterval,
		StartTime:     s.StartTime,
		EndTime:       s.EndTime,
		ScheduleType:  s.ScheduleType,
		RunNow:        s.RunNow,
		Status:        s.Status,
		Params:        s.Params,
		CreatedAt:     s.CreatedAt.String(),
		CreatedBy:     s.CreatedBy,
		UpdatedAt:     s.UpdatedAt.String(),
		UpdatedBy:     s.UpdatedBy,
	}
}

// FindSchedules 获取任务计划数据
func FindSchedules(db, userID, scheduleType string, pageIndex, pageSize int64, runNow bool) (items []Schedule, total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScheduleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"run_now": runNow,
	}

	if userID != "" {
		query["created_by"] = userID
	}
	if scheduleType != "" {
		query["schedule_type"] = scheduleType
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindSchedules", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Schedule
	// 取总件数
	count, err := c.CountDocuments(ctx, query)
	if err != nil {
		return result, 0, err
	}

	if pageIndex != 0 && pageSize != 0 {
		opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}}).SetSkip((pageIndex - 1) * pageSize).SetLimit(pageSize)
		schedules, err := c.Find(ctx, query, opts)
		if err != nil {
			utils.ErrorLog("error FindSchedules", err.Error())
			return nil, 0, err
		}
		defer schedules.Close(ctx)
		for schedules.Next(ctx) {
			var schedule Schedule
			err := schedules.Decode(&schedule)
			if err != nil {
				utils.ErrorLog("error FindSchedules", err.Error())
				return nil, 0, err
			}
			result = append(result, schedule)
		}
		return result, count, nil
	}

	opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: -1}})
	schedules, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindSchedules", err.Error())
		return nil, 0, err
	}
	defer schedules.Close(ctx)
	for schedules.Next(ctx) {
		var schedule Schedule
		err := schedules.Decode(&schedule)
		if err != nil {
			utils.ErrorLog("error FindSchedules", err.Error())
			return nil, 0, err
		}
		result = append(result, schedule)
	}

	return result, count, nil
}

// FindSchedule 获取任务数据
func FindSchedule(db, scheduleID string) (items Schedule, err error) {
	client := database.New()
	opts := options.Collection()
	opts.SetReadPreference(readpref.Primary())
	c := client.Database(database.GetDBName(db)).Collection(ScheduleCollection, opts)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"schedule_id": scheduleID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindSchedule", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Schedule

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindSchedule", err.Error())
		return result, err
	}

	return result, nil
}

// AddSchedule 添加任务数据
func AddSchedule(db string, s *Schedule) (scheduleID string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScheduleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.ID = primitive.NewObjectID()
	s.ScheduleID = s.ID.Hex()

	_, err = c.InsertOne(ctx, s)
	if err != nil {
		utils.ErrorLog("error AddSchedule", err.Error())
		return "", err
	}

	return s.ScheduleID, nil
}

// ModifySchedule 更新任务数据
func ModifySchedule(m *ModifyParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(m.Database)).Collection(ScheduleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(m.ScheduleID)
	if err != nil {
		utils.ErrorLog("error ModifySchedule", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": m.Writer,
	}

	if m.EntryID != "" {
		id, _ := strconv.ParseInt(m.EntryID, 10, 64)
		change["entry_id"] = id
	}

	if m.Status != "" {
		change["status"] = m.Status
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifySchedule", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifySchedule", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifySchedule", err.Error())
		return err
	}

	return nil
}

// DeleteSchedule 删除任务数据
func DeleteSchedule(db string, scheduleIds []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScheduleCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSchedule", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSchedule", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, scheduleID := range scheduleIds {
			objectID, err := primitive.ObjectIDFromHex(scheduleID)
			if err != nil {
				utils.ErrorLog("error DeleteSchedule", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSchedule", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error DeleteSchedule", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSchedule", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSchedule", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// AddScheduleNameUniqueIndex 添加schedule_name唯一索引
func AddScheduleNameUniqueIndex(ctx context.Context, db, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ScheduleCollection)
	query := bson.M{}
	cursor, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error AddScheduleNameUniqueIndex", err.Error())
		return err
	}
	defer cursor.Close(ctx)
	uniqueMap := make(map[string]int64)
	for cursor.Next(ctx) {
		var schedule Schedule
		err = cursor.Decode(&schedule)
		if err != nil {
			utils.ErrorLog("error AddScheduleNameUniqueIndex", err.Error())
			return err
		}
		if schedule.ScheduleName == "db restore" {
			deleteSchedule := bson.M{
				"schedule_id": schedule.ScheduleID,
			}
			_, err = c.DeleteOne(ctx, deleteSchedule)
			if err != nil {
				utils.ErrorLog("error AddScheduleNameUniqueIndex", err.Error())
				return err
			}
			continue
		}

		if schedule.CreatedBy == userID {
			_, ok := uniqueMap[schedule.ScheduleType]
			if ok {
				deleteSchedule := bson.M{
					"schedule_id": schedule.ScheduleID,
				}
				_, err = c.DeleteOne(ctx, deleteSchedule)
				if err != nil {
					utils.ErrorLog("error AddScheduleNameUniqueIndex", err.Error())
					return err
				}
			} else {
				uniqueMap[schedule.ScheduleType] = 1
			}
			continue
		} else {
			deleteSchedule := bson.M{
				"schedule_id": schedule.ScheduleID,
			}
			_, err = c.DeleteOne(ctx, deleteSchedule)
			if err != nil {
				utils.ErrorLog("error AddScheduleNameUniqueIndex", err.Error())
				return err
			}
		}

	}
	// 添加schedule_name唯一索引
	Index := mongo.IndexModel{
		Keys:    bson.D{{Key: "schedule_name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	schedule, err := IndexExits(ctx, c, "schedule_name")
	if err != nil {
		utils.ErrorLog("error AddScheduleNameUniqueIndex", err.Error())
		return err
	}
	if !schedule {
		_, err = c.Indexes().CreateOne(ctx, Index)
		if err != nil {
			utils.ErrorLog("error AddScheduleNameUniqueIndex", err.Error())
			return err
		}
	}
	return nil
}

// 判断索引是否存在
func IndexExits(ctx context.Context, c *mongo.Collection, indexName string) (bool, error) {
	isExits := false
	index := indexName + "_1"
	var results []bson.M
	opts := options.ListIndexes().SetMaxTime(10 * time.Second)
	cursor, err := c.Indexes().List(ctx, opts)
	if err != nil {
		utils.ErrorLog("IndexExits", err.Error())
		return isExits, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &results)
	if err != nil {
		utils.ErrorLog("IndexExits", err.Error())
		return isExits, err
	}
	for _, result := range results {
		value, ok := result["name"]
		if ok && value == index {
			isExits = true
		}
		fmt.Printf("value is %v", value)
	}
	return isExits, nil
}
