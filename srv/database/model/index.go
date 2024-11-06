package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/database/utils"
)

const MaxIndexCount = 52 // default max 64 x %82

type IndexUsage struct {
	Name  string `bson:"name"`
	Usage int64  `bson:"usage"`
}

func GetIndexUsage(c *mongo.Collection) (u []*IndexUsage, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	query := []bson.M{
		{"$indexStats": bson.M{}},
		{"$project": bson.M{
			"name":  1,
			"usage": "$accesses.ops",
		}},
		{"$sort": bson.M{"usage": -1}},
	}

	var result []*IndexUsage
	cur, err := c.Aggregate(ctx, query)
	if err != nil {
		utils.ErrorLog("GetIndexUsage", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("GetIndexUsage", err.Error())
		return nil, err
	}

	return result, nil
}

func DeleteLeastUsed(c *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	us, err := GetIndexUsage(c)
	if err != nil {
		utils.ErrorLog("DeleteLeastUsed", err.Error())
		return nil
	}

	// 当前索引未超过最大值的情况下，不需要删除索引
	if len(us) < MaxIndexCount {
		return nil
	}

	for i := MaxIndexCount; i < len(us)-1; i++ {
		if _, err := c.Indexes().DropOne(ctx, us[i].Name); err != nil {
			utils.ErrorLog("DeleteLeastUsed", err.Error())
			return err
		}
	}

	return nil
}
