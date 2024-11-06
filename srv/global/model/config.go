package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	config "rxcsoft.cn/pit3/srv/global/proto/mail-config"
	"rxcsoft.cn/pit3/srv/global/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	ConfigCollection = "mail_configs"
)

// Config 邮件配置
type Config struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	ConfigID  string             `json:"config_id" bson:"config_id"`
	Mail      string             `json:"mail" bson:"mail"`
	Password  string             `json:"password" bson:"password"`
	Host      string             `json:"host" bson:"host"`
	Port      string             `json:"port" bson:"port"`
	Ssl       string             `json:"ssl" bson:"ssl"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy string             `json:"created_by" bson:"created_by"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy string             `json:"updated_by" bson:"updated_by"`
}

// ToProto 转换为proto数据
func (mc *Config) ToProto() *config.Config {
	return &config.Config{
		ConfigId:  mc.ConfigID,
		Mail:      mc.Mail,
		Password:  mc.Password,
		Host:      mc.Host,
		Port:      mc.Port,
		Ssl:       mc.Ssl,
		CreatedAt: mc.CreatedAt.String(),
		CreatedBy: mc.CreatedBy,
		UpdatedAt: mc.UpdatedAt.String(),
		UpdatedBy: mc.UpdatedBy,
	}
}

// FindConfigs 获取邮件配置集合
func FindConfigs(db string) (mc []Config, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ConfigCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	var result []Config
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindConfigs", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var mc Config
		err := cur.Decode(&mc)
		if err != nil {
			utils.ErrorLog("error FindConfigs", err.Error())
			return nil, err
		}
		result = append(result, mc)
	}
	return result, nil
}

// FindConfig 获取邮件配置
func FindConfig(db string) (mc Config, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ConfigCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindConfig", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Config
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindConfig", err.Error())
		return result, err
	}
	return result, nil
}

// AddConfig 添加邮件配置
func AddConfig(db string, mc *Config) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ConfigCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mc.ID = primitive.NewObjectID()
	mc.ConfigID = mc.ID.Hex()

	queryJSON, _ := json.Marshal(mc)
	utils.DebugLog("AddConfig", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, mc)
	if err != nil {
		utils.ErrorLog("error AddConfig", err.Error())
		return mc.ConfigID, err
	}

	return mc.ConfigID, nil
}

// ModifyConfig 更新邮件配置
func ModifyConfig(db string, mc *Config) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ConfigCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(mc.ConfigID)
	if err != nil {
		utils.ErrorLog("error ModifyConfig", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{
		"ssl":        mc.Ssl,
		"mail":       mc.Mail,
		"password":   mc.Password,
		"host":       mc.Host,
		"port":       mc.Port,
		"updated_at": mc.UpdatedAt,
		"updated_by": mc.UpdatedBy,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyConfig", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyConfig", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyConfig", err.Error())
		return err
	}
	return nil
}
