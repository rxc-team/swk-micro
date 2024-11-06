package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/global/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	SequencesCollection = "sequences"
)

// Sequence 採番
type Sequence struct {
	ID            string `json:"id" bson:"_id"`
	SequenceValue int64  `json:"sequence_value" bson:"sequence_value"`
}

// FindSequence 获取单个採番
func FindSequence(db, sequenceKey string) (q int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"_id": sequenceKey,
	}

	change := bson.M{
		"$inc": bson.M{
			"sequence_value": 1,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(1)

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindSequence", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Sequence

	if err := c.FindOneAndUpdate(ctx, query, change, opts).Decode(&result); err != nil {
		utils.ErrorLog("error FindSequence", err.Error())
		return result.SequenceValue, err
	}

	return result.SequenceValue, nil
}

// AddSequence 添加採番
func AddSequence(db, key string, startValue int64) (seq int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var s Sequence
	s.ID = key
	s.SequenceValue = startValue
	queryJSON, _ := json.Marshal(s)
	utils.DebugLog("AddSequence", fmt.Sprintf("Sequence: [ %s ]", queryJSON))

	if _, err = c.InsertOne(ctx, s); err != nil {
		utils.ErrorLog("error AddSequence", err.Error())
		return 0, err
	}

	return startValue, nil
}
