package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	timeconv "github.com/Andrew-M-C/go.timeconv"
	"github.com/micro/go-micro/v2/client"
	"github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	database "rxcsoft.cn/utils/mongo"
)

var log = logrus.New()

const (
	// TimeFormat 日期格式化format
	TimeFormat = "2006-01-02 03:04:05"
	// DateFormat 日期格式化format
	DateFormat = "2006-01-02"
)

// 结构体
type (
	// Condition ショットカードの条件
	Condition struct {
		FieldID       string `json:"field_id" bson:"field_id"`
		FieldType     string `json:"field_type" bson:"field_type"`
		SearchValue   string `json:"search_value" bson:"search_value"`
		Operator      string `json:"operator" bson:"operator"`
		IsDynamic     bool   `json:"is_dynamic" bson:"is_dynamic"`
		ConditionType string `json:"condition_type" bson:"condition_type"`
	}
	// SortItem 排序の条件
	SortItem struct {
		SortKey   string `json:"sort_key" bson:"sort_key"`
		SortValue string `json:"sort_value" bson:"sort_value"`
	}

	// ImportResult 导入结果
	ImportResult struct {
		Insert int64    `json:"insert"`
		Modify int64    `json:"modify"`
		Errors []*Error `json:"errors"`
	}

	// Error 行错误
	Error struct {
		FirstLine   int64  `json:"first_line"`
		LastLine    int64  `json:"last_line"`
		CurrentLine int64  `json:"current_line"`
		FieldID     string `json:"field_id"`
		FieldName   string `json:"field_name"`
		ErrorMsg    string `json:"error_msg"`
	}
	// Sequence 序列集合
	Sequence struct {
		ID            string `json:"id" bson:"_id"`
		SequenceValue int64  `json:"sequence_value" bson:"sequence_value"`
	}
)

// ToProto 转换为proto数据
func (i *ImportResult) ToProto() *item.ImportResult {

	var errors []*item.Error
	for _, er := range i.Errors {
		errors = append(errors, er.ToProto())
	}

	return &item.ImportResult{
		Insert: i.Insert,
		Modify: i.Modify,
		Errors: errors,
	}
}

// ToProto 转换为proto数据
func (i *Error) ToProto() *item.Error {
	return &item.Error{
		FirstLine:   i.FirstLine,
		LastLine:    i.LastLine,
		CurrentLine: i.CurrentLine,
		FieldId:     i.FieldID,
		FieldName:   i.FieldName,
		ErrorMsg:    i.ErrorMsg,
	}
}

// 变量名
const (
	// ItemCollection item_ collection
	ItemCollection = "item_"
	// FieldPrefix 字段名前缀
	FieldPrefix = "database.field.field_"
	// DatastorePrefix 台账名前缀
	DatastorePrefix = "database.datastore.datastore_"
	// SequenceDatastorePrefix 台账字段表示顺序列名前缀
	SequenceDatastorePrefix = "datastore_"
	// DisplayOrder 表示顺
	DisplayOrder = "_displayorder"
	// Fields 字段集合
	Fields = "_fields_"
	// Auto 自增字段
	Auto = "_auto"
)

var ErrInvalidIndexValue = errors.New("invalid index value")

// GetItemCollectionName 获取item的集合的名称
func GetItemCollectionName(datastoreID string) string {
	return ItemCollection + datastoreID
}

// GetFieldNameKey 获取字段名的前缀
func GetFieldNameKey(appID, datastoreID, fieldID string) string {
	return "apps." + appID + ".fields." + datastoreID + "_" + fieldID
}

// GetDatastoreNameKey 获取台账名的前缀
func GetDatastoreNameKey(appID, optionID string) string {
	return "apps." + appID + ".datastores." + optionID
}

// GetMappingNameKey 获取台账映射的前缀
func GetMappingNameKey(appID, datastoreID, mappingID string) string {
	return "apps." + appID + ".mappings." + datastoreID + "_" + mappingID
}

// GetOptionNameKey 获取选择组名的前缀
func GetOptionNameKey(appID, optionID string) string {
	return "apps." + appID + ".options." + optionID
}

// GetOptionLabelNameKey 获取选择名的前缀
func GetOptionLabelNameKey(appID, optionID, dsID string) string {
	return "apps." + appID + ".options." + optionID + "_" + dsID
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

// 字符串获取时间
func getTime(value string) time.Time {
	if len(value) == 0 {
		date, _ := time.Parse("2006-01-02", "0001-01-01")
		return date
	}
	date, _ := time.Parse("2006-01-02", value)

	return date
}

// IsUsed 该字段是否被功能函数字段引用
func IsUsed(fID string, fields []Field) (r bool) {
	for _, f := range fields {
		if f.Formula != "" {
			if strings.Contains(f.Formula, fID) {
				return true
			}
		}
	}

	return false
}

// GetExpireymd 计算租赁满了日
func GetExpireymd(leasestymd, leasekikan, extentionOption string) (value time.Time, err error) {
	var expireymd time.Time
	// 租赁开始日转换
	stymd, err := time.Parse("2006-01-02", leasestymd)
	if err != nil {
		return expireymd, err
	}
	lkikan, err := strconv.Atoi(leasekikan)
	if err != nil {
		return expireymd, err
	}
	ekikan, err := strconv.Atoi(extentionOption)
	if err != nil {
		return expireymd, err
	}
	// 租赁满了日算出
	expireymd = timeconv.AddDate(stymd, 0, lkikan+ekikan, 0)

	return expireymd, nil
}

// 满了检查
func ExpireCheck(leaseexpireymd, handleMondth string) bool {
	if len(leaseexpireymd) < 7 || len(handleMondth) < 7 {
		return true
	}
	if leaseexpireymd[:7] > handleMondth[:7] {
		return true
	}
	return false
}

// getConfig 获取用户配置情报
func getConfig(db, appID string) (cfg *app.Configs, err error) {
	configService := app.NewAppService("manage", client.DefaultClient)

	var req app.FindAppRequest
	req.AppId = appID
	req.Database = db

	response, err := configService.FindApp(context.TODO(), &req)
	if err != nil {
		return nil, err
	}

	return response.GetApp().GetConfigs(), nil
}

func getFields(db, datastoreId string) (fields []*Field, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by":   "",
		"datastore_id": datastoreId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("getFields", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)

	var result []*Field
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("getFields", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("getFields", err.Error())
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

func addEmptyData(itemMap ItemMap, f Field) {
	// autonum 不需要添加
	if _, exist := itemMap[f.FieldID]; !exist {
		switch f.FieldType {
		case "text", "textarea", "options", "function", "lookup", "time":
			itemMap[f.FieldID] = &Value{
				DataType: f.FieldType,
				Value:    "",
			}
		case "number":
			itemMap[f.FieldID] = &Value{
				DataType: f.FieldType,
				Value:    0.0,
			}
		case "date":
			itemMap[f.FieldID] = &Value{
				DataType: f.FieldType,
				Value:    time.Time{},
			}
		case "switch":
			itemMap[f.FieldID] = &Value{
				DataType: f.FieldType,
				Value:    false,
			}
		case "user":
			itemMap[f.FieldID] = &Value{
				DataType: f.FieldType,
				Value:    []string{},
			}
		case "file":
			itemMap[f.FieldID] = &Value{
				DataType: f.FieldType,
				Value:    "[]",
			}
		}
	}
}

func transformBsoncoreDocument(registry *bsoncodec.Registry, val interface{}, mapAllowed bool, paramName string) bsoncore.Document {
	if registry == nil {
		registry = bson.DefaultRegistry
	}
	if val == nil {
		return nil
	}
	if bs, ok := val.([]byte); ok {
		// Slight optimization so we'll just use MarshalBSON and not go through the codec machinery.
		val = bson.Raw(bs)
	}
	if !mapAllowed {
		refValue := reflect.ValueOf(val)
		if refValue.Kind() == reflect.Map && refValue.Len() > 1 {
			return nil
		}
	}

	// TODO(skriptble): Use a pool of these instead.
	buf := make([]byte, 0, 256)
	b, err := bson.MarshalAppendWithRegistry(registry, buf[:0], val)
	if err != nil {
		return nil
	}
	return b
}

func getOrGenerateIndexName(keySpecDocument bsoncore.Document, model mongo.IndexModel) (string, error) {
	if model.Options != nil && model.Options.Name != nil {
		return *model.Options.Name, nil
	}

	name := bytes.NewBufferString("")
	first := true

	elems, err := keySpecDocument.Elements()
	if err != nil {
		return "", err
	}
	for _, elem := range elems {
		if !first {
			_, err := name.WriteRune('_')
			if err != nil {
				return "", err
			}
		}

		_, err := name.WriteString(elem.Key())
		if err != nil {
			return "", err
		}

		_, err = name.WriteRune('_')
		if err != nil {
			return "", err
		}

		var value string

		bsonValue := elem.Value()
		switch bsonValue.Type {
		case bsontype.Int32:
			value = fmt.Sprintf("%d", bsonValue.Int32())
		case bsontype.Int64:
			value = fmt.Sprintf("%d", bsonValue.Int64())
		case bsontype.String:
			value = bsonValue.StringValue()
		default:
			return "", ErrInvalidIndexValue
		}

		_, err = name.WriteString(value)
		if err != nil {
			return "", err
		}

		first = false
	}

	return name.String(), nil
}
