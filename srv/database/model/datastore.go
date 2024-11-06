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

	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// DataStoresCollection data_stores collection
	DataStoresCollection = "data_stores"
)

type (
	// Datastore Datastore信息
	Datastore struct {
		ID                  primitive.ObjectID `json:"id" bson:"_id"`
		DatastoreID         string             `json:"datastore_id" bson:"datastore_id"`
		AppID               string             `json:"app_id" bson:"app_id"`
		DatastoreName       string             `json:"datastore_name" bson:"datastore_name"`
		ApiKey              string             `json:"api_key" bson:"api_key"`
		CanCheck            bool               `json:"can_check" bson:"can_check"`
		ShowInMenu          bool               `json:"show_in_menu" bson:"show_in_menu"`
		NoStatus            bool               `json:"no_status" bson:"no_status"`
		Encoding            string             `json:"encoding" bson:"encoding"`
		Mappings            []*MappingConf     `json:"mappings" bson:"mappings"`
		Sorts               []*SortItem        `json:"sorts" bson:"sorts"`
		ScanFields          []string           `json:"scan_fields" bson:"scan_fields"`
		ScanFieldsConnector string             `json:"scan_fields_connector" bson:"scan_fields_connector"`
		PrintField1         string             `json:"print_field1" bson:"print_field1"`
		PrintField2         string             `json:"print_field2" bson:"print_field2"`
		PrintField3         string             `json:"print_field3" bson:"print_field3"`
		DisplayOrder        int64              `json:"display_order" bson:"display_order"`
		UniqueFields        []string           `json:"unique_fields" bson:"unique_fields"`
		Relations           []*RelationItem    `json:"relations" bson:"relations"`
		CreatedAt           time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy           string             `json:"created_by" bson:"created_by"`
		UpdatedAt           time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy           string             `json:"updated_by" bson:"updated_by"`
		DeletedAt           time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy           string             `json:"deleted_by" bson:"deleted_by"`
	}

	// RelationItem 排序の条件
	RelationItem struct {
		RelationId  string            `json:"relation_id" bson:"relation_id"`
		DatastoreId string            `json:"datastore_id" bson:"datastore_id"`
		Fields      map[string]string `json:"fields" bson:"fields"`
	}

	// MappingConf 映射规则
	MappingConf struct {
		MappingID     string         `json:"mapping_id" bson:"mapping_id"`
		MappingName   string         `json:"mapping_name" bson:"mapping_name"`
		MappingType   string         `json:"mapping_type" bson:"mapping_type"`
		UpdateType    string         `json:"update_type" bson:"update_type"`
		SeparatorChar string         `json:"separator_char" bson:"separator_char"`
		BreakChar     string         `json:"break_char" bson:"break_char"`
		ApplyType     string         `json:"apply_type" bson:"apply_type"`
		LineBreakCode string         `json:"line_break_code" bson:"line_break_code"`
		CharEncoding  string         `json:"char_encoding" bson:"char_encoding"`
		MappingRule   []*MappingRule `json:"mapping_rule" bson:"mapping_rule"`
	}

	// MappingRule 映射规则
	MappingRule struct {
		FromKey      string `json:"from_key" bson:"from_key"`
		ToKey        string `json:"to_key" bson:"to_key"`
		IsRequired   bool   `json:"is_required" bson:"is_required"`     // 必须
		Exist        bool   `json:"exist" bson:"exist"`                 // 存在
		Special      bool   `json:"special" bson:"special"`             // 禁則文字(正则表达式)
		DefaultValue string `json:"default_value" bson:"default_value"` // 禁則文字(正则表达式)
		PrimaryKey   bool   `json:"primary_key" bson:"primary_key"`     // 是否作为主键
		CheckChange  bool   `json:"check_change" bson:"check_change"`   // 是否作为主键
		Precision    int64  `json:"precision" bson:"precision"`         // 小数位数
		ShowOrder    int64  `json:"show_order" bson:"show_order"`       // 表示顺
		DataType     string `json:"data_type" bson:"data_type"`
		Format       string `json:"format" bson:"format"`   //逗号分割 当前格式-转换格式（日期）
		Replace      string `json:"replace" bson:"replace"` // 逗号分割 需要替换字符-变换字段（文本）
	}
)

// ToProto 转换为proto数据
func (d *Datastore) ToProto() *datastore.Datastore {
	var mappings []*datastore.MappingConf
	for _, m := range d.Mappings {
		mappings = append(mappings, m.ToProto())
	}

	var sorts []*datastore.SortItem
	for _, s := range d.Sorts {
		sorts = append(sorts, s.ToProto())
	}
	var relations []*datastore.RelationItem
	for _, s := range d.Relations {
		relations = append(relations, s.ToProto())
	}

	return &datastore.Datastore{
		AppId:               d.AppID,
		DatastoreId:         d.DatastoreID,
		DatastoreName:       d.DatastoreName,
		ApiKey:              d.ApiKey,
		CanCheck:            d.CanCheck,
		ShowInMenu:          d.ShowInMenu,
		NoStatus:            d.NoStatus,
		Encoding:            d.Encoding,
		Mappings:            mappings,
		Sorts:               sorts,
		ScanFields:          d.ScanFields,
		ScanFieldsConnector: d.ScanFieldsConnector,
		PrintField1:         d.PrintField1,
		PrintField2:         d.PrintField2,
		PrintField3:         d.PrintField3,
		UniqueFields:        d.UniqueFields,
		DisplayOrder:        d.DisplayOrder,
		Relations:           relations,
		CreatedAt:           d.CreatedAt.String(),
		CreatedBy:           d.CreatedBy,
		UpdatedAt:           d.UpdatedAt.String(),
		UpdatedBy:           d.UpdatedBy,
		DeletedAt:           d.DeletedAt.String(),
		DeletedBy:           d.DeletedBy,
	}
}

// ToProto 转换为proto数据
func (s *SortItem) ToProto() *datastore.SortItem {
	return &datastore.SortItem{
		SortKey:   s.SortKey,
		SortValue: s.SortValue,
	}
}

// ToProto 转换为proto数据
func (s *RelationItem) ToProto() *datastore.RelationItem {
	return &datastore.RelationItem{
		RelationId:  s.RelationId,
		DatastoreId: s.DatastoreId,
		Fields:      s.Fields,
	}
}

// ToProto 转换为proto数据
func (m *MappingConf) ToProto() *datastore.MappingConf {

	var rules []*datastore.MappingRule
	for _, r := range m.MappingRule {
		rules = append(rules, r.ToProto())
	}

	return &datastore.MappingConf{
		MappingId:     m.MappingID,
		MappingName:   m.MappingName,
		MappingType:   m.MappingType,
		UpdateType:    m.UpdateType,
		SeparatorChar: m.SeparatorChar,
		BreakChar:     m.BreakChar,
		ApplyType:     m.ApplyType,
		LineBreakCode: m.LineBreakCode,
		CharEncoding:  m.CharEncoding,
		MappingRule:   rules,
	}
}

// ToProto 转换为proto数据
func (m *MappingRule) ToProto() *datastore.MappingRule {
	return &datastore.MappingRule{
		FromKey:      m.FromKey,
		ToKey:        m.ToKey,
		DataType:     m.DataType,
		CheckChange:  m.CheckChange,
		IsRequired:   m.IsRequired,
		Exist:        m.Exist,
		Special:      m.Special,
		DefaultValue: m.DefaultValue,
		Format:       m.Format,
		Precision:    m.Precision,
		ShowOrder:    m.ShowOrder,
		Replace:      m.Replace,
		PrimaryKey:   m.PrimaryKey,
	}
}

// FindDatastores 通过appID获取Datastore信息
func FindDatastores(db, appID, dsName, canCheck, showInMenu string) (ds []Datastore, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"app_id":     appID,
	}

	// 文件名不为空的场合，添加到查询条件中
	if dsName != "" {
		query["datastore_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(dsName), Options: "m"}}
	}

	// check不为空的场合，添加到查询条件中
	if canCheck != "" {
		result, err := strconv.ParseBool(canCheck)
		if err == nil {
			query["can_check"] = result
		}
	}

	// inMenu不为空的场合，添加到查询条件中
	if showInMenu != "" {
		result, err := strconv.ParseBool(showInMenu)
		if err == nil {
			query["show_in_menu"] = result
		}
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDatastores", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Datastore
	sortItem := bson.D{
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("FindDatastores", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var das Datastore
		err := cur.Decode(&das)
		if err != nil {
			utils.ErrorLog("FindDatastores", err.Error())
			return nil, err
		}
		result = append(result, das)
	}

	return result, nil
}

// FindDatastore 通过ID获取Datastore信息
func FindDatastore(db, id string) (ds Datastore, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Datastore

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by":   "",
		"datastore_id": id,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDatastore", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("FindDatastore", err.Error())
		return result, err
	}

	return result, nil
}

// FindDatastoreByKey 通过apiKey获取Datastore信息
func FindDatastoreByKey(db, app, apiKey string) (ds Datastore, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Datastore

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"app_id":     app,
		"api_key":    apiKey,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDatastore", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("FindDatastore", err.Error())
		return result, err
	}

	return result, nil
}

// FindDatastoreMapping 通过ID获取Datastore映射信息
func FindDatastoreMapping(db, id, mappingID string) (mp *MappingConf, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Datastore

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorLog("FindDatastoreMapping", err.Error())
		return nil, err
	}

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"_id":                 objectID,
		"mappings.mapping_id": mappingID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDatastoreMapping", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.FindOne().SetProjection(bson.M{
		"_id":        0,
		"mappings.$": 1,
	})

	if err := c.FindOne(ctx, query, opt).Decode(&result); err != nil {
		utils.ErrorLog("FindDatastoreMapping", err.Error())
		return nil, err
	}

	return result.Mappings[0], nil
}

// AddDatastore 添加Datastore信息
func AddDatastore(db string, ds *Datastore) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ds.ID = primitive.NewObjectID()
	ds.DatastoreID = ds.ID.Hex()
	ds.DatastoreName = GetDatastoreNameKey(ds.AppID, ds.DatastoreID)
	if len(ds.Mappings) == 0 {
		ds.Mappings = make([]*MappingConf, 0)
	}
	if ds.ApiKey == "" {
		ds.ApiKey = ds.DatastoreID
	}

	if len(ds.Sorts) == 0 {
		ds.Sorts = make([]*SortItem, 0)
	}

	if len(ds.ScanFields) == 0 {
		ds.ScanFields = make([]string, 0)
	}

	if len(ds.UniqueFields) == 0 {
		ds.UniqueFields = make([]string, 0)
	}
	if len(ds.Relations) == 0 {
		ds.Relations = make([]*RelationItem, 0)
	}

	queryJSON, _ := json.Marshal(ds)
	utils.DebugLog("AddDatastore", fmt.Sprintf("Datastore: [ %s ]", queryJSON))

	if _, err = c.InsertOne(ctx, ds); err != nil {
		utils.ErrorLog("AddDatastore", err.Error())
		return "", err
	}

	if err := createDatastoreIndex(db, ds.DatastoreID, ds.ApiKey, ds.UniqueFields); err != nil {
		utils.ErrorLog("AddDatastore", err.Error())
		return "", err
	}

	if err := createFieldsOrder(db, ds.DatastoreID); err != nil {
		utils.ErrorLog("AddDatastore", err.Error())
		return "", err
	}

	return ds.DatastoreID, nil
}

// createDatastoreIndex 创建台账的数据的索引
func createDatastoreIndex(db, datastoreID, apiKey string, uniqueFields []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client.Database(database.GetDBName(db)).CreateCollection(ctx, GetItemCollectionName(datastoreID))

	models := []mongo.IndexModel{
		{
			Keys:    bson.M{"item_id": 1},
			Options: options.Index().SetSparse(true).SetUnique(false),
		},
		{
			Keys:    bson.M{"datastore_id": 1},
			Options: options.Index().SetSparse(true).SetUnique(false),
		},
		{
			Keys:    bson.M{"app_id": 1},
			Options: options.Index().SetSparse(true).SetUnique(false),
		},
		{
			Keys:    bson.M{"status": 1},
			Options: options.Index().SetSparse(true).SetUnique(false),
		},
	}

	// 复合主键
	for _, uf := range uniqueFields {
		if len(uf) > 0 {
			ufs := strings.Split(uf, ",")
			var keys bson.D
			for _, key := range ufs {
				keys = append(keys, bson.E{Key: key, Value: 1})
			}

			models = append(models, mongo.IndexModel{
				Keys:    keys,
				Options: options.Index().SetSparse(true).SetUnique(false),
			})
		}
	}

	// if apiKey == "rireki" {
	// 	models = append(models, mongo.IndexModel{
	// 		Keys: bson.D{
	// 			{Key: "items.no.value", Value: 1},
	// 			{Key: "items.zengokbn.value", Value: 1},
	// 			{Key: "created_at", Value: 1},
	// 		},
	// 		Options: options.Index().SetSparse(true).SetUnique(false),
	// 	})
	// }
	// if apiKey == "repayment" {
	// 	models = append(models, mongo.IndexModel{
	// 		Keys: bson.D{
	// 			{Key: "items.keiyakuno.value", Value: 1},
	// 			{Key: "created_at", Value: 1},
	// 		},
	// 		Options: options.Index().SetSparse(true).SetUnique(false),
	// 	})
	// }
	// if apiKey == "paymentInterest" {
	// 	models = append(models, mongo.IndexModel{
	// 		Keys: bson.D{
	// 			{Key: "items.keiyakuno.value", Value: 1},
	// 			{Key: "created_at", Value: 1},
	// 		},
	// 		Options: options.Index().SetSparse(true).SetUnique(false),
	// 	})
	// }

	opts := options.CreateIndexes().SetMaxTime(60 * time.Second)
	if _, err := c.Indexes().CreateMany(ctx, models, opts); err != nil {
		utils.ErrorLog("createDatastoreIndex", err.Error())
		return err
	}

	return nil
}

func AddUniqueKey(db, appId, datastoreId, uniqueFields string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ct := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreId))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appId,
		"datastore_id": datastoreId,
	}

	change := bson.M{
		"$addToSet": bson.M{
			"unique_fields": uniqueFields,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("AddUniqueKey", fmt.Sprintf("query: [ %s ]", queryJSON))

	changeJSON, _ := json.Marshal(change)
	utils.DebugLog("AddUniqueKey", fmt.Sprintf("change: [ %s ]", changeJSON))

	// 更新台账
	if _, err := c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("AddUniqueKey", err.Error())
		return err
	}

	// 后台添加索引，防止前台崩溃
	var index mongo.IndexModel

	// 复合主键
	if len(uniqueFields) > 0 {
		ufs := strings.Split(uniqueFields, ",")
		var keys bson.D
		for _, key := range ufs {
			keys = append(keys, bson.E{Key: "items." + key + ".value", Value: 1})
		}

		index = mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetSparse(true).SetUnique(true),
		}
	}

	// 添加唯一索引
	if _, err := ct.Indexes().CreateOne(ctx, index); err != nil {

		change1 := bson.M{
			"$pull": bson.M{
				"unique_fields": uniqueFields,
			},
		}

		// 更新台账,删除唯一属性
		if _, err := c.UpdateOne(ctx, query, change1); err != nil {
			utils.ErrorLog("AddUniqueKey", err.Error())
			return err
		}

		if mongo.IsDuplicateKeyError(err) {
			return errors.New("一意の制約を作成できませんでした。複合プロパティに既存のデータに重複するアイテムがあります")
		}

		utils.ErrorLog("AddUniqueKey", err.Error())
		return err
	}

	return nil
}

// AddRelation 添加关系
func AddRelation(db, appId, datastoreId string, relation RelationItem) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appId,
		"datastore_id": datastoreId,
	}

	// 生成唯一id
	if relation.RelationId == "" {
		relation.RelationId = primitive.NewObjectID().Hex()
	}

	change := bson.M{
		"$addToSet": bson.M{
			"relations": relation,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("AddRelation", fmt.Sprintf("query: [ %s ]", queryJSON))

	changeJSON, _ := json.Marshal(change)
	utils.DebugLog("AddRelation", fmt.Sprintf("change: [ %s ]", changeJSON))

	// 更新台账
	if _, err := c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("AddRelation", err.Error())
		return err
	}

	return nil
}

// AddMapping 添加映射信息
func AddMapping(db, appID, datastoreID string, mp MappingConf) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mp.MappingID = primitive.NewObjectID().Hex()
	mp.MappingName = GetMappingNameKey(appID, datastoreID, mp.MappingID)

	query := bson.M{
		"app_id":       appID,
		"datastore_id": datastoreID,
	}

	change := bson.M{
		"$addToSet": bson.M{
			"mappings": mp,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("AddMapping", fmt.Sprintf("query: [ %s ]", queryJSON))

	changeJSON, _ := json.Marshal(change)
	utils.DebugLog("AddMapping", fmt.Sprintf("change: [ %s ]", changeJSON))

	if _, err = c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("AddMapping", err.Error())
		return "", err
	}

	return mp.MappingID, nil
}

// ModifyReq 修改台账记录
type ModifyReq struct {
	DatastoreID         string
	DatastoreName       string
	ApiKey              string
	LookupDatastoreList []string
	Sorts               []*SortItem
	CanCheck            string
	ShowInMenu          string
	NoStatus            string
	Encoding            string
	Writer              string
	Owners              string
	ScanFields          []string
	ScanFieldsConnector string
	PrintField1         string
	PrintField2         string
	PrintField3         string
	DisplayOrder        int64
}

// ModifyDatastore 修改Datastore信息
func ModifyDatastore(db string, ds *ModifyReq) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(ds.DatastoreID)
	if err != nil {
		utils.ErrorLog("ModifyDatastore", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": ds.Writer,
	}

	// 排序字段
	if ds.Sorts == nil {
		ds.Sorts = make([]*SortItem, 0)
	}
	change["sorts"] = ds.Sorts

	// check不为空的场合
	if ds.CanCheck != "" {
		result, err := strconv.ParseBool(ds.CanCheck)
		if err == nil {
			change["can_check"] = result
			if result {
				// 扫描字段
				if ds.ScanFields == nil {
					ds.ScanFields = make([]string, 0)
				}
				change["scan_fields"] = ds.ScanFields
				// 扫描字段连接符
				change["scan_fields_connector"] = ds.ScanFieldsConnector

				change["print_field1"] = ds.PrintField1
				change["print_field2"] = ds.PrintField2
				change["print_field3"] = ds.PrintField3
			} else {
				// 扫描字段
				change["scan_fields"] = []string{}
				// 扫描字段连接符
				change["scan_fields_connector"] = ""

				change["print_field1"] = ""
				change["print_field2"] = ""
				change["print_field3"] = ""
			}
		}
	}
	// apiKey不为空的场合
	if ds.ApiKey != "" {
		change["api_key"] = ds.ApiKey
	}

	// inMenu不为空的场合
	if ds.ShowInMenu != "" {
		result, err := strconv.ParseBool(ds.ShowInMenu)
		if err == nil {
			change["show_in_menu"] = result
		}
	}

	// 使用工作流不为空的场合
	if ds.NoStatus != "" {
		result, err := strconv.ParseBool(ds.NoStatus)
		if err == nil {
			change["no_status"] = result
		}
	}

	// encoding不为空的场合
	if ds.Encoding != "" {
		change["encoding"] = ds.Encoding
	}

	// owners不为空的场合
	if len(ds.Owners) > 0 {
		change["owners"] = ds.Owners
	}

	// ScanFields不为空的场合
	if len(ds.ScanFields) > 0 {
		change["scan_fields"] = ds.ScanFields
	}
	// ScanFieldsConnector不为空的场合
	if len(ds.ScanFieldsConnector) > 0 {
		change["scan_fields_connector"] = ds.ScanFieldsConnector
	}
	// PrintField1不为空的场合
	if len(ds.PrintField1) > 0 {
		change["print_field1"] = ds.PrintField1
	}
	// PrintField2不为空的场合
	if len(ds.PrintField2) > 0 {
		change["print_field2"] = ds.PrintField2
	}
	// PrintField3不为空的场合
	if len(ds.PrintField3) > 0 {
		change["print_field3"] = ds.PrintField3
	}
	// PrintField3不为空的场合
	if ds.DisplayOrder > 0 {
		change["display_order"] = ds.DisplayOrder
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyDatastore", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyDatastore", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err = c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("ModifyDatastore", err.Error())
		return err
	}

	return nil
}

// ModifyMapping 修改映射信息
func ModifyMapping(db, appID, datastoreID string, mp MappingConf) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":              appID,
		"datastore_id":        datastoreID,
		"mappings.mapping_id": mp.MappingID,
	}

	change := bson.M{
		"$set": bson.M{
			"mappings.$.mapping_type":    mp.MappingType,
			"mappings.$.update_type":     mp.UpdateType,
			"mappings.$.separator_char":  mp.SeparatorChar,
			"mappings.$.break_char":      mp.BreakChar,
			"mappings.$.line_break_code": mp.LineBreakCode,
			"mappings.$.char_encoding":   mp.CharEncoding,
			"mappings.$.apply_type":      mp.ApplyType,
			"mappings.$.mapping_rule":    mp.MappingRule,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyMapping", fmt.Sprintf("query: [ %s ]", queryJSON))

	changeJSON, _ := json.Marshal(change)
	utils.DebugLog("ModifyMapping", fmt.Sprintf("change: [ %s ]", changeJSON))

	if _, err = c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("ModifyDatastore", err.Error())
		return err
	}

	return nil
}

// DeleteDatastore 删除单个Datastore信息
func DeleteDatastore(db, datastoreID, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(datastoreID)
	if err != nil {
		utils.ErrorLog("DeleteDatastore", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteDatastore", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteDatastore", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err = c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("DeleteDatastore", err.Error())
		return err
	}

	return nil
}

// DeleteMapping 删除映射信息
func DeleteMapping(db, appID, datastoreID, mappingID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appID,
		"datastore_id": datastoreID,
	}

	change := bson.M{
		"$pull": bson.M{
			"mappings": bson.M{
				"mapping_id": mappingID,
			},
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteMapping", fmt.Sprintf("query: [ %s ]", queryJSON))

	changeJSON, _ := json.Marshal(change)
	utils.DebugLog("DeleteMapping", fmt.Sprintf("change: [ %s ]", changeJSON))

	if _, err = c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("DeleteMapping", err.Error())
		return err
	}

	return nil
}

// DeleteRelation 删除关系
func DeleteRelation(db, appId, datastoreId, relationId string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appId,
		"datastore_id": datastoreId,
	}

	change := bson.M{
		"$pull": bson.M{
			"relations": bson.M{
				"relation_id": relationId,
			},
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteRelation", fmt.Sprintf("query: [ %s ]", queryJSON))

	changeJSON, _ := json.Marshal(change)
	utils.DebugLog("DeleteRelation", fmt.Sprintf("change: [ %s ]", changeJSON))

	// 更新台账
	if _, err := c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("DeleteRelation", err.Error())
		return err
	}

	return nil
}

// DeleteUniqueKey 删除唯一性key
func DeleteUniqueKey(db, appId, datastoreId, uniqueFields string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ct := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreId))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appId,
		"datastore_id": datastoreId,
	}

	change := bson.M{
		"$pull": bson.M{
			"unique_fields": uniqueFields,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteUniqueKey", fmt.Sprintf("query: [ %s ]", queryJSON))

	changeJSON, _ := json.Marshal(change)
	utils.DebugLog("DeleteUniqueKey", fmt.Sprintf("change: [ %s ]", changeJSON))

	// 更新台账
	if _, err := c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("DeleteUniqueKey", err.Error())
		return err
	}

	go func() {
		// 后台删除索引
		var model mongo.IndexModel

		// 复合主键
		if len(uniqueFields) > 0 {
			ufs := strings.Split(uniqueFields, ",")
			var keys bson.D
			for _, key := range ufs {
				keys = append(keys, bson.E{Key: "items." + key + ".value", Value: 1})
			}

			model = mongo.IndexModel{
				Keys:    keys,
				Options: options.Index().SetSparse(true).SetUnique(true),
			}
		}

		keys := transformBsoncoreDocument(nil, model.Keys, false, "keys")
		if keys == nil {
			utils.ErrorLog("DeleteUniqueKey", "key invalid")
			return
		}

		name, err := getOrGenerateIndexName(keys, model)
		if err != nil {
			utils.ErrorLog("DeleteUniqueKey", err.Error())
			return
		}

		// 删除唯一索引
		if _, err := ct.Indexes().DropOne(context.TODO(), name); err != nil {
			utils.ErrorLog("DeleteUniqueKey", err.Error())
			return
		}
	}()

	return nil
}

// DeleteSelectDatastores 删除多个Datastore信息
func DeleteSelectDatastores(db string, datastoreIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteSelectDatastores", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteSelectDatastores", err.Error())
		return err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, datastoreID := range datastoreIDList {
			objectID, err := primitive.ObjectIDFromHex(datastoreID)
			if err != nil {
				utils.ErrorLog("DeleteSelectDatastores", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSelectDatastores", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectDatastores", fmt.Sprintf("update: [ %s ]", updateJSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("DeleteSelectDatastores", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteSelectDatastores", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteSelectDatastores", err.Error())
		return err
	}

	session.EndSession(ctx)
	return nil
}

// HardDeleteDatastores 物理删除多个Datastore信息
func HardDeleteDatastores(db string, datastoreIDList []string) error {
	client := database.New()
	cDS := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	cFD := client.Database(database.GetDBName(db)).Collection("fields")
	cHS := client.Database(database.GetDBName(db)).Collection("histories")
	cSD := client.Database(database.GetDBName(db)).Collection("schedules")
	cQR := client.Database(database.GetDBName(db)).Collection("queries")
	cRP := client.Database(database.GetDBName(db)).Collection("reports")
	cDR := client.Database(database.GetDBName(db)).Collection("dashboards")
	cLA := client.Database(database.GetDBName(db)).Collection("languages")
	cRO := client.Database(database.GetDBName(db)).Collection("roles")
	cGP := client.Database(database.GetDBName(db)).Collection("groups")
	cAC := client.Database(database.GetDBName(db)).Collection("access")
	cPM := client.Database(database.GetDBName(db)).Collection("permissions")
	cWF := client.Database(database.GetDBName(db)).Collection("wf_form")
	cWN := client.Database(database.GetDBName(db)).Collection("wf_node")
	cWS := client.Database(database.GetDBName(db)).Collection("wf_workflows")
	cWE := client.Database(database.GetDBName(db)).Collection("wf_examples")
	cWP := client.Database(database.GetDBName(db)).Collection("wf_process")
	cSE := client.Database(database.GetDBName(db)).Collection("sequences")
	cPT := client.Database(database.GetDBName(db)).Collection("prints")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("HardDeleteDatastores", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("HardDeleteDatastores", err.Error())
		return err
	}

	// 查询存放临时数据用
	type Field struct {
		FieldID     string `json:"field_id" bson:"field_id"`
		DatastoreID string `json:"datastore_id" bson:"datastore_id"`
		AppID       string `json:"app_id" bson:"app_id"`
	}
	type Report struct {
		ReportID string `json:"report_id" bson:"report_id"`
	}
	type Dashboard struct {
		DashboardID string `json:"dashboard_id" bson:"dashboard_id"`
	}
	type WorkFlow struct {
		WorkFlowID string `json:"wf_id" bson:"wf_id"`
	}
	type Example struct {
		ExampleID string `json:"ex_id" bson:"ex_id"`
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// 循环删除台账和台账多语言&台账下属各成员和各成员多语言
		for _, datastoreID := range datastoreIDList {
			// 获取台账信息备用
			dsInfo, err := FindDatastore(db, datastoreID)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 查询到台账下的所有字段备用
			var allfList []Field
			allfq := bson.M{
				"datastore_id": datastoreID,
			}
			allfcur, err := cFD.Find(ctx, allfq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			defer allfcur.Close(ctx)
			err = allfcur.All(ctx, &allfList)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账
			objectID, err := primitive.ObjectIDFromHex(datastoreID)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteDatastores", fmt.Sprintf("query: [ %s ]", queryJSON))
			_, err = cDS.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账数据
			d := client.Database(database.GetDBName(db)).Collection("item_" + datastoreID)
			err = d.Drop(ctx)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除角色台账配置信息
			rdupd := bson.M{
				"$pull": bson.M{
					"datastores": bson.M{
						"datastore_id": datastoreID,
					},
				},
			}
			_, err = cRO.UpdateMany(sc, bson.M{}, rdupd)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除用户组中台账相关流程的信息
			gdwupd := bson.M{
				"$unset": bson.M{
					"workflow." + datastoreID: "",
				},
			}
			_, err = cGP.UpdateMany(sc, bson.M{}, gdwupd)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 查找流程信息备用
			var wfList []WorkFlow
			wfsq := bson.M{
				"params.datastore": datastoreID,
			}
			wfscur, err := cWF.Find(ctx, wfsq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			defer wfscur.Close(ctx)
			err = wfscur.All(ctx, &wfList)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 循环删除台账下所有流程的信息
			for _, wf := range wfList {
				wfq := bson.M{
					"wf_id": wf.WorkFlowID,
				}
				// 删除form
				_, err = cWF.DeleteMany(sc, wfq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除node
				_, err = cWN.DeleteMany(sc, wfq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除workflow
				_, err = cWS.DeleteMany(sc, wfq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除workflow的语言
				langUp := bson.M{
					"$unset": bson.M{
						"apps." + dsInfo.AppID + ".workflows." + wf.WorkFlowID:      "",
						"apps." + dsInfo.AppID + ".workflows.menu_" + wf.WorkFlowID: "",
					},
				}
				_, err = cLA.UpdateMany(sc, bson.M{}, langUp)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 查找example信息备用
				var exList []Example
				wfcur, err := cWE.Find(ctx, wfq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				defer wfcur.Close(ctx)
				err = wfcur.All(ctx, &exList)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除example
				_, err = cWE.DeleteMany(sc, wfq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除process
				for _, ex := range exList {
					exq := bson.M{
						"ex_id": ex.ExampleID,
					}
					_, err = cWP.DeleteMany(sc, exq)
					if err != nil {
						utils.ErrorLog("HardDeleteDatastores", err.Error())
						return err
					}
				}
			}
			// 删除台账对应的字段排序用序列
			seq := bson.M{
				"_id": "datastore_" + datastoreID + "_fields__displayorder",
			}
			_, err = cSE.DeleteOne(sc, seq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 查询到台账下的自增字段备用
			var afList []Field
			autofsq := bson.M{
				"datastore_id": datastoreID,
				"field_type":   "autonum",
			}
			autofscur, err := cFD.Find(ctx, autofsq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			defer autofscur.Close(ctx)
			err = autofscur.All(ctx, &afList)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除自增字段对应的序列
			for _, f := range afList {
				autoseq := bson.M{
					"_id": "datastore_" + f.DatastoreID + "_fields_" + f.FieldID + "_auto",
				}
				_, err = cSE.DeleteOne(sc, autoseq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
			}
			// 删除台账下的字段
			dsq := bson.M{
				"datastore_id": datastoreID,
			}
			_, err = cFD.DeleteMany(sc, dsq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账下的履历
			_, err = cHS.DeleteMany(sc, dsq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账的access
			udAc := bson.M{
				"$unset": bson.M{
					"apps." + dsInfo.AppID + ".data_access." + datastoreID: "",
				},
			}
			_, err = cAC.UpdateMany(ctx, bson.M{}, udAc)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}

			// 删除台账的permissions
			quPm := bson.M{
				"actions": bson.M{
					"$not": bson.M{
						"$eq": nil,
					},
				},
				"action_type": "datastore",
			}
			udPm := bson.M{
				"$pull": bson.M{
					"actions": bson.M{
						"object_id": datastoreID,
					},
				},
			}
			_, err = cPM.UpdateMany(ctx, quPm, udPm)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}

			// 查询到台账的所有的报表备用
			var rpList []Report
			rpscur, err := cRP.Find(ctx, dsq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			defer rpscur.Close(ctx)
			err = rpscur.All(ctx, &rpList)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账下的报表
			_, err = cRP.DeleteMany(sc, dsq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			for _, rp := range rpList {
				// 删除台账下的报表的语言
				rpLangUp := bson.M{
					"$unset": bson.M{
						"apps." + dsInfo.AppID + ".reports." + rp.ReportID: "",
					},
				}
				_, err = cLA.UpdateMany(sc, bson.M{}, rpLangUp)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除角色报表配置信息
				rrupd := bson.M{
					"$pull": bson.M{
						"reports": rp.ReportID,
					},
				}
				_, err = cRO.UpdateMany(sc, bson.M{}, rrupd)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除报表的permissions
				quPm := bson.M{
					"actions": bson.M{
						"$not": bson.M{
							"$eq": nil,
						},
					},
					"action_type": "report",
				}
				udPm := bson.M{
					"$pull": bson.M{
						"actions": bson.M{
							"object_id": rp.ReportID,
						},
					},
				}
				_, err = cPM.UpdateMany(ctx, quPm, udPm)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}

				// 删除报表数据
				d := client.Database(database.GetDBName(db)).Collection("report_" + rp.ReportID)
				err = d.Drop(ctx)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}

				// 匹配报表下的仪表盘
				rpidq := bson.M{
					"report_id": rp.ReportID,
				}
				// 查询到台账的所有的仪表盘备用
				var dashList []Dashboard
				dashscur, err := cDR.Find(ctx, rpidq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				defer dashscur.Close(ctx)
				err = dashscur.All(ctx, &dashList)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除报表下的仪表盘
				_, err = cDR.DeleteMany(sc, rpidq)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
				// 删除报表下的仪表盘的语言
				for _, dash := range dashList {
					dashLangUp := bson.M{
						"$unset": bson.M{
							"apps." + dsInfo.AppID + ".dashboards." + dash.DashboardID: "",
						},
					}
					_, err = cLA.UpdateMany(sc, bson.M{}, dashLangUp)
					if err != nil {
						utils.ErrorLog("HardDeleteDatastores", err.Error())
						return err
					}
				}
			}
			// 删除以该台账为单位的同步任务计划
			scheq := bson.M{
				"schedule_type":       "data-sync",
				"params.datastore_id": datastoreID,
			}
			_, err = cSD.DeleteMany(sc, scheq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账的快捷方式
			queryq := bson.M{
				"datastore_id": datastoreID,
			}
			_, err = cQR.DeleteMany(sc, queryq)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账对应的语言
			dsLangUp := bson.M{
				"$unset": bson.M{
					"apps." + dsInfo.AppID + ".datastores." + datastoreID: "",
				},
			}
			_, err = cLA.UpdateMany(sc, bson.M{}, dsLangUp)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
			// 删除台账映射对应的语言
			for _, mp := range dsInfo.Mappings {
				mappingLangUp := bson.M{
					"$unset": bson.M{
						"apps." + dsInfo.AppID + ".mappings." + datastoreID + "_" + mp.MappingID: "",
					},
				}
				_, err = cLA.UpdateMany(sc, bson.M{}, mappingLangUp)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
			}
			// 删除台账字段对应的语言
			for _, field := range allfList {
				fieldLangUp := bson.M{
					"$unset": bson.M{
						"apps." + field.AppID + ".fields." + datastoreID + "_" + field.FieldID: "",
					},
				}
				_, err = cLA.UpdateMany(sc, bson.M{}, fieldLangUp)
				if err != nil {
					utils.ErrorLog("HardDeleteDatastores", err.Error())
					return err
				}
			}
			// 删除台账的打印设置
			dpt := bson.M{
				"datastore_id": datastoreID,
			}
			_, err = cPT.DeleteMany(sc, dpt)
			if err != nil {
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("HardDeleteDatastores", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("HardDeleteDatastores", err.Error())
		return err
	}

	session.EndSession(ctx)
	return nil
}
