/*
 * @Description: common
 * @Author: RXC 廖云江
 * @Date: 2019-09-26 14:30:05
 * @LastEditors: RXC 陈辉宇
 * @LastEditTime: 2020-10-29 13:31:08
 */

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/report/utils"
	database "rxcsoft.cn/utils/mongo"
)

// 变量名
const (
	// ItemCollection item_ collection
	ItemCollection = "item_"

	// DataStoresCollection data_stores collection
	DataStoresCollection = "data_stores"
)

// File 文件类型
type File struct {
	URL  string `json:"url" bson:"url"`
	Name string `json:"name" bson:"name"`
}

type Field struct {
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	FieldID           string             `json:"field_id" bson:"field_id"`
	AppID             string             `json:"app_id" bson:"app_id"`
	DatastoreID       string             `json:"datastore_id" bson:"datastore_id"`
	FieldName         string             `json:"field_name" bson:"field_name"`
	FieldType         string             `json:"field_type" bson:"field_type"`
	IsRequired        bool               `json:"is_required" bson:"is_required"`
	IsFixed           bool               `json:"is_fixed" bson:"is_fixed"`
	IsImage           bool               `json:"is_image" bson:"is_image"`
	IsCheckImage      bool               `json:"is_check_image" bson:"is_check_image"`
	Unique            bool               `json:"unique" bson:"unique"`
	LookupAppID       string             `json:"lookup_app_id" bson:"lookup_app_id"`
	LookupDatastoreID string             `json:"lookup_datastore_id" bson:"lookup_datastore_id"`
	LookupFieldID     string             `json:"lookup_field_id" bson:"lookup_field_id"`
	UserGroupID       string             `json:"user_group_id" bson:"user_group_id"`
	OptionID          string             `json:"option_id" bson:"option_id"`
	Cols              int64              `json:"cols" bson:"cols"`
	Rows              int64              `json:"rows" bson:"rows"`
	X                 int64              `json:"x" bson:"x"`
	Y                 int64              `json:"y" bson:"y"`
	MinLength         int64              `json:"min_length" bson:"min_length"`
	MaxLength         int64              `json:"max_length" bson:"max_length"`
	MinValue          int64              `json:"min_value" bson:"min_value"`
	MaxValue          int64              `json:"max_value" bson:"max_value"`
	AsTitle           bool               `json:"as_title" bson:"as_title"`
	Width             int64              `json:"width" bson:"width"`
	DisplayOrder      int64              `json:"display_order" bson:"display_order"`
	DisplayDigits     int64              `json:"display_digits" bson:"display_digits"`
	Precision         int64              `json:"precision" bson:"precision"`
	Prefix            string             `json:"prefix" bson:"prefix"`
	ReturnType        string             `json:"return_type" bson:"return_type"`
	Formula           string             `json:"formula" bson:"formula"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy         string             `json:"created_by" bson:"created_by"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy         string             `json:"updated_by" bson:"updated_by"`
	DeletedAt         time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy         string             `json:"deleted_by" bson:"deleted_by"`
}

// ResultItem 台账的数据
type ResultItem struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	ItemID      string             `json:"item_id" bson:"item_id"`
	AppID       string             `json:"app_id" bson:"app_id"`
	DatastoreID string             `json:"datastore_id" bson:"datastore_id"`
	ItemMap     ItemMap            `json:"items" bson:"items"`
	Owners      []string           `json:"owners" bson:"owners"`
	CheckType   string             `json:"check_type" bson:"check_type"`
	CheckStatus string             `json:"check_status" bson:"check_status"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy   string             `json:"created_by" bson:"created_by"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	CheckedAt   time.Time          `json:"checked_at" bson:"checked_at"`
	CheckedBy   string             `json:"checked_by" bson:"checked_by"`
	DeletedAt   time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy   string             `json:"deleted_by" bson:"deleted_by"`
	LabelTime   time.Time          `json:"label_time" bson:"label_time"`
	Status      string             `json:"status" bson:"status"`
}

// Datastore Datastore信息
type Datastore struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Relations []*RelationItem    `json:"relations" bson:"relations"`
}

// RelationItem 排序の条件
type RelationItem struct {
	RelationId  string            `json:"relation_id" bson:"relation_id"`
	DatastoreId string            `json:"datastore_id" bson:"datastore_id"`
	Fields      map[string]string `json:"fields" bson:"fields"`
}

// GetItemCollectionName 获取item的集合的名称
func GetItemCollectionName(datastoreID string) string {
	return ItemCollection + datastoreID
}

// GetValueString 获取对应的数据类型的数据
func GetValueString(value *Value) (v string) {
	switch value.DataType {
	case "text", "textarea", "options":
		if value.Value == nil {
			return ""
		}
		return value.Value.(string)
	case "number":
		switch value.Value.(type) {
		case int:
			return strconv.FormatFloat(float64(value.Value.(int)), 'f', -1, 64)
		case float64:
			return strconv.FormatFloat(value.Value.(float64), 'f', -1, 64)
		default:
			return strconv.FormatFloat(0.0, 'f', -1, 64)
		}
	case "autonum":
		return value.Value.(string)
	case "date":
		return value.Value.(time.Time).Format("2006-01-02")
	case "time":
		return value.Value.(string)
	case "switch":
		return strconv.FormatBool(value.Value.(bool))
	case "file":
		jsonBytes, _ := json.Marshal(value.Value)
		return string(jsonBytes)
	case "user":
		users := value.Value.([]interface{})
		var strArr []string
		for _, user := range users {
			strArr = append(strArr, user.(string))
		}
		return strings.Join(strArr, ",")
	case "lookup":
		if value.Value == nil {
			return ""
		}
		return value.Value.(string)
	default:
		jsonBytes, _ := json.Marshal(value.Value)
		return string(jsonBytes)
	}
}

// GetValue 获取值
func GetValue(itemMap *Value) string {
	dataType := itemMap.DataType
	switch dataType {
	case "text", "textarea", "options", "lookup":
		if itemMap.Value != nil {
			return itemMap.Value.(string)
		}
		return ""
	case "number":
		value := itemMap.Value
		switch value.(type) {
		case int64:
			return strconv.FormatInt(value.(int64), 10)
		case float64:
			return strconv.FormatFloat(value.(float64), 'f', -1, 64)
		case string:
			return value.(string)
		default:
			return strconv.FormatFloat(0.0, 'f', -1, 64)
		}
	case "autonum":
		value := itemMap.Value
		switch value.(type) {
		case int64:
			return strconv.FormatInt(value.(int64), 10)
		default:
			return value.(string)
		}
	case "date":
		value := itemMap.Value
		switch value.(type) {
		case primitive.DateTime:
			return value.(primitive.DateTime).Time().Format("2006-01-02")
		case time.Time:
			return value.(time.Time).Format("2006-01-02")
		case string:
			return value.(string)
		default:
			return value.(time.Time).Format("2006-01-02")
		}
	case "time":
		return itemMap.Value.(string)
	case "switch":
		value := itemMap.Value
		switch value.(type) {
		case string:
			return value.(string)
		default:
			return strconv.FormatBool(value.(bool))
		}
	case "user":
		value := itemMap.Value
		switch value.(type) {
		case string:
			return value.(string)
		default:
			users := value.(primitive.A)
			var strArr []string
			for _, user := range users {
				strArr = append(strArr, user.(string))
			}
			return strings.Join(strArr, ",")
		}
	case "file":
		value := itemMap.Value
		jsonBytes, _ := json.Marshal(value)
		return string(jsonBytes)
	default:
		return itemMap.Value.(string)
	}
}

// getSearchValue 根据字段类型获取相应的值
func getSearchValue(dataType, value string) (v interface{}) {
	switch dataType {
	case "text", "textarea":
		return value
	case "number":
		result, _ := strconv.ParseFloat(value, 64)
		return result
	case "autonum":
		return value
	case "date":
		if len(value) == 0 {
			date, _ := time.Parse("2006-01-02", "0001-01-01")
			return date
		}
		date, _ := time.Parse("2006-01-02", value)
		return date
	case "datetime":
		if len(value) == 0 {
			date, _ := time.Parse("2006-01-02", "0001-01-01")
			return date
		}
		date, _ := time.Parse("2006-01-02", value)
		return date
	case "time":
		return value
	case "switch":
		result, _ := strconv.ParseBool(value)
		return result
	case "user":
		var result []string
		json.Unmarshal([]byte(value), &result)
		return result
	case "file":
		var result []File
		json.Unmarshal([]byte(value), &result)
		return result
	case "options":
		return value
	case "lookup":
		return value
	}

	return nil
}

func getTime(value string) time.Time {
	if len(value) == 0 {
		date, _ := time.Parse("2006-01-02", "0001-01-01")
		return date
	}
	date, _ := time.Parse("2006-01-02", value)

	return date
}

// findConfig 查找顾客配置情报
func findConfig(db, appID string) (cs *Config, e error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("apps")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认检索所有数据
	query := bson.M{
		"app_id": appID,
	}

	var result model.App

	err := c.FindOne(ctx, query).Decode(&result)
	if err != nil {
		utils.ErrorLog("error FindConfigs", err.Error())
		return nil, err
	}

	return &Config{
		Special:         result.Configs.Special,
		SyoriYm:         result.Configs.SyoriYm,
		ShortLeases:     result.Configs.ShortLeases,
		CheckStartDate:  result.Configs.CheckStartDate,
		KishuYm:         result.Configs.KishuYm,
		MinorBaseAmount: result.Configs.MinorBaseAmount,
	}, nil
}

func getAllFields(db, appId, datastoreId string) (map[string][]*Field, error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("fields")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by":   "",
		"app_id":       appId,
		"datastore_id": datastoreId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("getRelationFields", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)

	fieldsMap := make(map[string][]*Field)
	var result []*Field
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("getRelationFields", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var f Field
		err := cur.Decode(&f)
		if err != nil {
			utils.ErrorLog("FindItems", err.Error())
		}

		if f.FieldType == "lookup" {
			if _, ok := fieldsMap[f.LookupDatastoreID]; !ok {
				fs, err := getFields(db, f.AppID, f.LookupDatastoreID)
				if err != nil {
					utils.ErrorLog("FindItems", err.Error())
				}

				fieldsMap[f.LookupDatastoreID] = fs
			}
		}

		result = append(result, &f)
	}

	fieldsMap[datastoreId] = result

	return fieldsMap, nil
}

func getFields(db, appId, datastoreId string) (fields []*Field, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("fields")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by":   "",
		"app_id":       appId,
		"datastore_id": datastoreId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("getRelationFields", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)

	var result []*Field
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("getRelationFields", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("getRelationFields", err.Error())
		return nil, err
	}

	return result, nil
}

func getDatastore(db, datastoreId string) (d *Datastore, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by":   "",
		"datastore_id": datastoreId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("getFields", fmt.Sprintf("query: [ %s ]", queryJSON))

	var ds Datastore
	err = c.FindOne(ctx, query).Decode(&ds)
	if err != nil {
		utils.ErrorLog("getFields", err.Error())
		return nil, err
	}

	return &ds, nil
}
