package logger

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

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

// AddLogger 添加日志
func AddLogger(l *Logger) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(LoggersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	l.ID = primitive.NewObjectID()

	_, err = c.InsertOne(ctx, l)
	if err != nil {
		return err
	}
	return nil
}
