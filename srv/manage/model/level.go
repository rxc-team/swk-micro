package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/manage/proto/level"
	"rxcsoft.cn/pit3/srv/manage/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	LevelsCollection = "levels"
)

// Level 授权等级
type Level struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	LevelID   string             `json:"level_id" bson:"level_id"`
	LevelName string             `json:"level_name" bson:"level_name"`
	Allows    []string           `json:"allows" bson:"allows"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy string             `json:"created_by" bson:"created_by"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy string             `json:"updated_by" bson:"updated_by"`
}

// ToProto 转换为proto数据
func (m *Level) ToProto() *level.Level {
	return &level.Level{
		LevelId:   m.LevelID,
		LevelName: m.LevelName,
		Allows:    m.Allows,
		CreatedAt: m.CreatedAt.String(),
		CreatedBy: m.CreatedBy,
		UpdatedAt: m.UpdatedAt.String(),
		UpdatedBy: m.UpdatedBy,
	}
}

// FindLevels 查找多个授权等级记录
func FindLevels(ctx context.Context) (m []Level, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(LevelsCollection)

	query := bson.M{}

	var result []Level

	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindLevels", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindLevels", err.Error())
		return nil, err
	}

	return result, nil

}

// FindLevel 查找单个授权等级记录
func FindLevel(ctx context.Context, levelID string) (cus Level, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(LevelsCollection)

	var result Level

	query := bson.M{
		"level_id": levelID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindLevel", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindLevel", err.Error())
		return result, err
	}

	return result, nil
}

// AddLevel 添加单个授权等级记录
func AddLevel(ctx context.Context, m *Level) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(LevelsCollection)

	m.ID = primitive.NewObjectID()
	m.LevelID = m.ID.Hex()

	queryJSON, _ := json.Marshal(m)
	utils.DebugLog("AddLevel", fmt.Sprintf("Level: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, m)
	if err != nil {
		utils.ErrorLog("error AddLevel", err.Error())
		return "", err
	}

	return m.LevelID, nil
}

// ModifyLevel 修改单个授权等级记录
func ModifyLevel(ctx context.Context, levelKey, levelName, userID string, allows []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(LevelsCollection)

	query := bson.M{
		"level_id": levelKey,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": userID,
	}

	if len(levelName) > 0 {
		change["level_name"] = levelName
	}

	if len(allows) > 0 {
		change["allows"] = allows
	} else {
		change["allows"] = []string{}
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyLevel", fmt.Sprintf("Level: [ %s ]", queryJSON))
	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyLevel", fmt.Sprintf("Level: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyLevel", err.Error())
		return err
	}

	return nil
}

// DeleteLevel 硬删除单个授权等级记录
func DeleteLevel(ctx context.Context, levelID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(LevelsCollection)

	query := bson.M{
		"level_id": levelID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteLevel", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteLevel", err.Error())
		return err
	}

	return nil
}

// DeleteLevels 硬删除多个授权等级记录
func DeleteLevels(ctx context.Context, levelIds []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName("system")).Collection(LevelsCollection)

	for _, levelId := range levelIds {
		query := bson.M{
			"level_id": levelId,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("DeleteLevels", fmt.Sprintf("query: [ %s ]", queryJSON))

		_, err = c.DeleteOne(ctx, query)
		if err != nil {
			utils.ErrorLog("error DeleteLevels", err.Error())
			return err
		}
	}

	return nil
}
