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
	"rxcsoft.cn/pit3/srv/global/proto/logger"
	"rxcsoft.cn/pit3/srv/global/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	LoggersCollection = "loggers"
)

// Logger 日志
type Logger struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	AppName   string             `json:"app_name" bson:"app_name"`
	UserID    string             `json:"user_id" bson:"user_id"`
	Domain    string             `json:"domain" bson:"domain"`
	LogType   string             `json:"log_type" bson:"log_type"`
	ProcessID string             `json:"process_id" bson:"process_id"`
	ClientIP  string             `json:"client_ip" bson:"client_ip"`
	Source    string             `json:"source" bson:"source"`
	Msg       string             `json:"msg" bson:"msg"`
	Time      time.Time          `json:"time" bson:"time"`
	Level     string             `json:"level" bson:"level"`
	Params    map[string]string  `json:"params" bson:"params"`
}

// ToProto 转换为proto数据
func (l *Logger) ToProto() *logger.Logger {
	return &logger.Logger{
		AppName:   l.AppName,
		UserId:    l.UserID,
		Domain:    l.Domain,
		LogType:   l.LogType,
		ProcessId: l.ProcessID,
		ClientIp:  l.ClientIP,
		Source:    l.Source,
		Msg:       l.Msg,
		Time:      l.Time.Format("2006-01-02 15:04:05.000000"),
		Level:     l.Level,
		Params:    l.Params,
	}
}

// FindLoggers 获取日志
func FindLoggers(appName, userID, domain, logType, clientIP, level, startTime, endTime string, pageIndex, pageSize int64) (l []Logger, total int64, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(LoggersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	startTimeTmp := startTime + " 00:00:00.000000"
	newStartTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTimeTmp, time.Local)
	endTimeTmp := endTime + " 23:59:59.999999"
	newEndTime, err := time.ParseInLocation("2006-01-02 15:04:05", endTimeTmp, time.Local)

	query := bson.M{}

	// app_name非空
	if appName != "" {
		query["app_name"] = appName
	}
	// user_id非空
	if userID != "" {
		query["user_id"] = userID
	}
	// log_type非空
	if logType != "" {
		query["log_type"] = logType
	}
	// client_ip非空
	if clientIP != "" {
		query["client_ip"] = clientIP
	}
	// domain非空
	if domain != "" {
		query["domain"] = domain
	}
	// level非空
	if level != "" {
		query["level"] = level
	}
	// 开始时间非空
	if startTime != "" && endTime != "" {
		query["$and"] = []bson.M{
			{
				"time": bson.M{
					"$gte": newStartTime,
				},
			},
			{
				"time": bson.M{
					"$lte": newEndTime,
				},
			},
		}
	} else if startTime != "" {
		query["time"] = bson.M{"$gte": newStartTime}
	} else if endTime != "" {
		query["time"] = bson.M{"$lte": newEndTime}
	}

	var result []Logger
	// 取总件数
	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		return result, 0, err
	}

	if t > 10000 {
		t = 10000
	}

	// 聚合查询
	if pageIndex != 0 && pageSize != 0 {
		opts := options.Find().SetSort(bson.D{{Key: "time", Value: -1}}).SetSkip((pageIndex - 1) * pageSize).SetLimit(pageSize)
		cur, err := c.Find(ctx, query, opts)
		if err != nil {
			utils.ErrorLog("error FindLoggers", err.Error())
			return nil, 0, err
		}
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var lo Logger
			err := cur.Decode(&lo)
			if err != nil {
				utils.ErrorLog("error FindLoggers", err.Error())
				return nil, 0, err
			}
			result = append(result, lo)
		}

		return result, t, nil
	}
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindLoggers", fmt.Sprintf("query: [ %s ]", queryJSON))
	opts := options.Find().SetSort(bson.D{{Key: "time", Value: -1}})
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindLoggers", err.Error())
		return nil, 0, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var lo Logger
		err := cur.Decode(&lo)
		if err != nil {
			utils.ErrorLog("error FindLoggers", err.Error())
			return nil, 0, err
		}
		result = append(result, lo)
	}

	return result, t, nil
}

// AddLogger 添加日志
func AddLogger(l *Logger) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(LoggersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	l.ID = primitive.NewObjectID()

	_, err = c.InsertOne(ctx, l)
	if err != nil {
		utils.ErrorLog("error AddLogger", err.Error())
		return err
	}
	return nil
}

// CreateLoggerIndex 创建日志索引
func CreateLoggerIndex() (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(LoggersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models := []mongo.IndexModel{
		{
			Keys:    bson.M{"time": 1},
			Options: options.Index().SetBackground(true).SetSparse(false).SetUnique(false).SetExpireAfterSeconds(60 * 60 * 24 * 365),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	if _, err := c.Indexes().CreateMany(ctx, models, opts); err != nil {
		utils.ErrorLog("error add time index", err.Error())
		return err
	}

	return nil
}
