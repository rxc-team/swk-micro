package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/database/proto/template"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// TemplateCollection template_items collection
	TemplateCollection = "tpl_"
)

type (
	// TemplateItem 临时台账的数据
	TemplateItem struct {
		ID           primitive.ObjectID `json:"id" bson:"_id"`
		ItemID       string             `json:"item_id" bson:"item_id"`
		AppID        string             `json:"app_id" bson:"app_id"`
		DatastoreID  string             `json:"datastore_id" bson:"datastore_id"`
		ItemMap      ItemMap            `json:"items" bson:"items"`
		TemplateID   string             `json:"template_id" bson:"template_id"`
		DatastoreKey string             `json:"datastore_key" bson:"datastore_key"`
		CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy    string             `json:"created_by" bson:"created_by"`
	}
)

func genTplCollectionName(userID string) string {
	return TemplateCollection + userID
}

// ToProto 转换为proto数据
func (t *TemplateItem) ToProto(showItem bool) *template.TemplateItem {
	items := make(map[string]*template.Value, len(t.ItemMap))
	for key, it := range t.ItemMap {
		dataType := it.DataType
		items[key] = &template.Value{
			DataType: dataType,
			Value:    GetValueFromModel(it),
		}
	}
	return &template.TemplateItem{
		ItemId:       t.ItemID,
		AppId:        t.AppID,
		DatastoreId:  t.DatastoreID,
		Items:        items,
		TemplateId:   t.TemplateID,
		DatastoreKey: t.DatastoreKey,
		CreatedAt:    t.CreatedAt.String(),
		CreatedBy:    t.CreatedBy,
	}
}

// GetTemplateValueString 获取值
func GetTemplateValueString(it interface{}, showItem bool) string {
	itemMap := it.(primitive.M)
	dataType := itemMap["data_type"].(string)
	switch dataType {
	case "text", "textarea", "options":
		return itemMap["value"].(string)
	case "number":
		switch itemMap["value"].(type) {
		case int:
			return strconv.FormatFloat(float64(itemMap["value"].(int)), 'f', -1, 64)
		case int32:
			return strconv.FormatInt(int64(itemMap["value"].(int32)), 10)
		case int64:
			return strconv.FormatInt(itemMap["value"].(int64), 10)
		case float64:
			return strconv.FormatFloat(itemMap["value"].(float64), 'f', -1, 64)
		default:
			return strconv.FormatFloat(0.0, 'f', -1, 64)
		}
	case "autonum":
		return itemMap["value"].(string)
	case "date":
		switch itemMap["value"].(type) {
		case primitive.DateTime:
			return itemMap["value"].(primitive.DateTime).Time().Format("2006-01-02")
		case time.Time:
			return itemMap["value"].(time.Time).Format("2006-01-02")
		default:
			return itemMap["value"].(time.Time).Format("2006-01-02")
		}
	case "time":
		return itemMap["value"].(string)
	case "switch":
		return strconv.FormatBool(itemMap["value"].(bool))
	case "user":
		jsonBytes, _ := json.Marshal(itemMap["value"])
		return string(jsonBytes)
	case "file":
		jsonBytes, _ := json.Marshal(itemMap["value"])
		return string(jsonBytes)
	case "lookup":
		return itemMap["value"].(string)
	default:
		jsonBytes, _ := json.Marshal(itemMap["value"])
		return string(jsonBytes)
	}
}

// GetTemplateDataValue 获取对应的数据类型的数据
func GetTemplateDataValue(value *template.Value) (v interface{}) {
	switch value.DataType {
	case "text", "textarea":
		return value.GetValue()
	case "number":
		result, _ := strconv.ParseFloat(value.GetValue(), 64)
		return result
	case "autonum":
		return value.GetValue()
	case "date":
		if len(value.GetValue()) == 0 {
			date, _ := time.Parse("2006-01-02", "0001-01-01")
			return date
		}
		date, _ := time.Parse("2006-01-02", value.GetValue())
		return date
	case "time":
		return value.GetValue()
	case "switch":
		result, _ := strconv.ParseBool(value.GetValue())
		return result
	case "user":
		result := strings.Split(value.GetValue(), ",")
		return result
	case "file":
		var result []File
		json.Unmarshal([]byte(value.GetValue()), &result)
		return result
	case "options":
		return value.GetValue()
	case "lookup":
		return value.GetValue()
	}

	return nil
}

// FindTemplateItems 查找记录
func FindTemplateItems(db, tmpID, datastoreKey, collection string, pageSize, pageIndex int64) (items []TemplateItem, total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"template_id":   tmpID,
		"datastore_key": datastoreKey,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindTemplateItems", fmt.Sprintf("query: [ %s ]", queryJSON))
	var result []TemplateItem

	// 取总件数
	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindTemplateItems", err.Error())
		return result, 0, err
	}

	if pageSize != 0 && pageIndex != 0 {
		skip := (pageIndex - 1) * pageSize
		limit := pageSize

		opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetSkip(skip).SetLimit(limit)
		cur, err := c.Find(ctx, query, opts)
		if err != nil {
			utils.ErrorLog("FindTemplateItems", err.Error())
			return result, 0, err
		}
		defer cur.Close(ctx)

		err = cur.All(ctx, &result)
		if err != nil {
			utils.ErrorLog("FindTemplateItems", err.Error())
			return result, 0, err
		}
		return result, t, nil
	}

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("FindTemplateItems", err.Error())
		return result, 0, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var apItem TemplateItem
		err := cur.Decode(&apItem)
		if err != nil {
			utils.ErrorLog("FindTemplateItems", err.Error())
			return result, 0, err
		}
		result = append(result, apItem)
	}

	return result, t, nil
}

// MutilAddTemplateItem 批量添加数据(契约台账数据新规审批时，把支付，试算，偿却信息添加进来)
func MutilAddTemplateItem(db, collection string, items []*TemplateItem) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var cxModels []mongo.WriteModel

	for _, i := range items {
		i.ID = primitive.NewObjectID()
		i.ItemID = i.ID.Hex()

		insertCxModel := mongo.NewInsertOneModel()
		insertCxModel.SetDocument(i)
		cxModels = append(cxModels, insertCxModel)
	}

	if len(cxModels) > 0 {
		_, err = c.BulkWrite(ctx, cxModels)
		if err != nil {
			bke, ok := err.(mongo.BulkWriteException)
			if !ok {
				utils.ErrorLog("MutilAddTemplateItem", err.Error())
				return err
			}
			errInfo := bke.WriteErrors[0]
			utils.ErrorLog("MutilAddTemplateItem", errInfo.Message)
			return errors.New(errInfo.Message)
		}
	}

	return nil
}

// DeleteTemplateItems 删除记录
func DeleteTemplateItems(db, collection, tmpID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"template_id": tmpID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteTemplateItems", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err := c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("DeleteTemplateItems", err.Error())
		return err
	}
	return nil
}
