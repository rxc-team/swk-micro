package model

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/manage/utils"
	database "rxcsoft.cn/utils/mongo"
)

var log = logrus.New()

const (
	// SequencesCollection sequences collection
	SequencesCollection = "sequences"
	// ItemCollection item_ collection
	ItemCollection = "item_"
)

// Sequence 序列集合
type Sequence struct {
	ID            string `json:"id" bson:"_id"`
	SequenceValue int32  `json:"sequence_value" bson:"sequence_value"`
}

// GetAppDisplayOrderSequenceName 获取台账内字段表示顺序列名
func GetAppDisplayOrderSequenceName(customerID string) string {
	return "app_" + customerID + "_displayorder"
}

// GetItemCollectionName 获取item的集合的名称
func GetItemCollectionName(datastoreID string) string {
	return ItemCollection + datastoreID
}

// GetNextSequenceValue 获取下个序列值
func GetNextSequenceValue(ctx context.Context, db, sequenceName string) (num int32, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)

	query := bson.M{
		"_id": sequenceName,
	}

	update := bson.M{
		"$inc": bson.M{"sequence_value": 1},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(1)

	var result Sequence

	if err := c.FindOneAndUpdate(ctx, query, update, opts).Decode(&result); err != nil {
		utils.ErrorLog("error getNextSequenceValue", err.Error())
		return 0, err
	}
	return result.SequenceValue, nil
}

// CreateSequence 创建序列
func CreateSequence(ctx context.Context, db, sequenceName string, startValue int32) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)

	var s Sequence
	s.ID = sequenceName
	s.SequenceValue = startValue

	_, err = c.InsertOne(ctx, s)
	if err != nil {
		utils.ErrorLog("error CreateSequence", err.Error())
		return err
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
