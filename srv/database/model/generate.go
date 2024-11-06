package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/database/proto/generate"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// GenConfigsCollection 导入配置 collection
	GenConfigsCollection = "gen_config"
	GenDatasCollection   = "gen_data"
)

type (
	// GenerateConfig 导入配置信息
	GenerateConfig struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		AppID         string             `json:"app_id" bson:"app_id"`
		UserID        string             `json:"user_id" bson:"user_id"`
		DatastoreID   string             `json:"datastore_id" bson:"datastore_id"`
		DatastoreName string             `json:"datastore_name" bson:"datastore_name"`
		ApiKey        string             `json:"api_key" bson:"api_key"`
		CanCheck      bool               `json:"can_check" bson:"can_check"`
		Step          int64              `json:"step" bson:"step"`
		MappingID     string             `json:"mapping_id" bson:"mapping_id"`
		Fields        []*GField          `json:"fields" bson:"fields"`
	}

	//GField 字段
	GField struct {
		FieldID           string `json:"field_id" bson:"field_id"`
		AppID             string `json:"app_id" bson:"app_id"`
		DatastoreID       string `json:"datastore_id" bson:"datastore_id"`
		FieldName         string `json:"field_name" bson:"field_name"`
		FieldType         string `json:"field_type" bson:"field_type"`
		IsRequired        bool   `json:"is_required" bson:"is_required"`
		IsFixed           bool   `json:"is_fixed" bson:"is_fixed"`
		IsImage           bool   `json:"is_image" bson:"is_image"`
		IsCheckImage      bool   `json:"is_check_image" bson:"is_check_image"`
		Unique            bool   `json:"unique" bson:"unique"`
		LookupAppID       string `json:"lookup_app_id" bson:"lookup_app_id"`
		LookupDatastoreID string `json:"lookup_datastore_id" bson:"lookup_datastore_id"`
		LookupFieldID     string `json:"lookup_field_id" bson:"lookup_field_id"`
		UserGroupID       string `json:"user_group_id" bson:"user_group_id"`
		OptionID          string `json:"option_id" bson:"option_id"`
		MinLength         int64  `json:"min_length" bson:"min_length"`
		MaxLength         int64  `json:"max_length" bson:"max_length"`
		MinValue          int64  `json:"min_value" bson:"min_value"`
		MaxValue          int64  `json:"max_value" bson:"max_value"`
		AsTitle           bool   `json:"as_title" bson:"as_title"`
		DisplayOrder      int64  `json:"display_order" bson:"display_order"`
		DisplayDigits     int64  `json:"display_digits" bson:"display_digits"`
		Precision         int64  `json:"precision" bson:"precision"`
		Prefix            string `json:"prefix" bson:"prefix"`
		ReturnType        string `json:"return_type" bson:"return_type"`
		Formula           string `json:"formula" bson:"formula"`
		CsvHeader         string `json:"csv_header" bson:"csv_header"`
		CanChange         bool   `json:"can_change" bson:"can_change"`
		IsEmptyLine       bool   `json:"is_empty_line" bson:"is_empty_line"`
		CheckErrors       string `json:"check_errors" bson:"check_errors"`
	}

	GItem struct {
		AppID   string            `json:"app_id" bson:"app_id"`
		UserID  string            `json:"user_id" bson:"user_id"`
		ItemMap map[string]string `json:"item_map" bson:"item_map"`
	}
)

// ToProto 转换为proto数据
func (f *GField) ToProto() *generate.Field {
	return &generate.Field{
		FieldId:           f.FieldID,
		AppId:             f.AppID,
		DatastoreId:       f.DatastoreID,
		FieldName:         f.FieldName,
		FieldType:         f.FieldType,
		IsRequired:        f.IsRequired,
		IsFixed:           f.IsFixed,
		IsImage:           f.IsImage,
		IsCheckImage:      f.IsCheckImage,
		Unique:            f.Unique,
		LookupAppId:       f.LookupAppID,
		LookupDatastoreId: f.LookupDatastoreID,
		LookupFieldId:     f.LookupFieldID,
		UserGroupId:       f.UserGroupID,
		OptionId:          f.OptionID,
		MinLength:         f.MinLength,
		MaxLength:         f.MaxLength,
		MinValue:          f.MinValue,
		MaxValue:          f.MaxValue,
		DisplayOrder:      f.DisplayOrder,
		DisplayDigits:     f.DisplayDigits,
		Precision:         f.Precision,
		Prefix:            f.Prefix,
		ReturnType:        f.ReturnType,
		Formula:           f.Formula,
		AsTitle:           f.AsTitle,
		CsvHeader:         f.CsvHeader,
		CanChange:         f.CanChange,
		IsEmptyLine:       f.IsEmptyLine,
		CheckErrors:       f.CheckErrors,
	}
}

// ToProto 转换为proto数据
func (d *GenerateConfig) ToProto() *generate.GenerateConfig {

	var fields []*generate.Field
	for _, v := range d.Fields {
		fields = append(fields, v.ToProto())
	}
	return &generate.GenerateConfig{
		AppId:         d.AppID,
		DatastoreId:   d.DatastoreID,
		DatastoreName: d.DatastoreName,
		ApiKey:        d.ApiKey,
		CanCheck:      d.CanCheck,
		Step:          d.Step,
		MappingId:     d.MappingID,
		Fields:        fields,
	}
}

// ToProto 转换为proto数据
func (d *GItem) ToProto() *generate.Item {

	return &generate.Item{
		AppId:   d.AppID,
		UserId:  d.UserID,
		ItemMap: d.ItemMap,
	}
}

// FindGenerateConfig 查找导入配置信息
func FindGenerateConfig(db, appId, userId string) (*GenerateConfig, error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GenConfigsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":  appId,
		"user_id": userId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindGenerateConfig", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result GenerateConfig
	err := c.FindOne(ctx, query).Decode(&result)
	if err != nil {
		utils.ErrorLog("FindGenerateConfig", err.Error())
		return nil, err
	}

	return &result, nil
}

// FindRowData 分页查找全部数据
func FindRowData(db, appId, userId string, pageIndex, pageSize int64) ([]*GItem, error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GenDatasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":  appId,
		"user_id": userId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindRowData", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []*GItem
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("FindRowData", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("FindRowData", err.Error())
		return nil, err
	}

	return result, nil
}

// FindColumnData 查找某一列去重后的数据
func FindColumnData(db, appId, userId, coulmnName string) ([]string, error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GenDatasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":  appId,
		"user_id": userId,
	}

	pipe := []bson.M{
		{"$match": query},
	}

	group := bson.M{
		"_id": "$item_map." + coulmnName,
	}
	pipe = append(pipe, bson.M{"$group": group})

	project := bson.M{
		"column": "$_id",
	}
	pipe = append(pipe, bson.M{"$project": project})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindColumnData", fmt.Sprintf("query: [ %s ]", queryJSON))

	type Result struct {
		Column string `bson:"column"`
	}

	var result []*Result

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("FindColumnData", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("FindColumnData", err.Error())
		return nil, err
	}

	var list []string
	for _, v := range result {
		list = append(list, v.Column)
	}

	return list, nil
}

// AddGenerateConfig 添加导入配置信息
func AddGenerateConfig(db string, data *GenerateConfig) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GenConfigsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data.ID = primitive.NewObjectID()
	data.Step = 0
	data.Fields = make([]*GField, 0)

	queryJSON, _ := json.Marshal(data)
	utils.DebugLog("AddGenerateConfig", fmt.Sprintf("[ %s ]", queryJSON))

	if _, err = c.InsertOne(ctx, data); err != nil {
		utils.ErrorLog("AddGenerateConfig", err.Error())
		return err
	}

	return nil
}

// GModifyReq 导入配置信息
type GModifyReq struct {
	DatastoreID   string
	DatastoreName string
	ApiKey        string
	CanCheck      string
	MappingID     string
	Step          int64
	Fields        []*GField
}

// ModifyGenerateConfig 修改导入配置信息
func ModifyGenerateConfig(db, appId, userId string, data *GModifyReq) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GenConfigsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":  appId,
		"user_id": userId,
	}

	change := bson.M{
		"step": data.Step,
	}

	if len(data.DatastoreID) > 0 {
		change["datastore_id"] = data.DatastoreID
	}
	if len(data.DatastoreName) > 0 {
		change["datastore_name"] = data.DatastoreName
	}
	if len(data.ApiKey) > 0 {
		change["datastore_api_key"] = data.ApiKey
	}
	if len(data.CanCheck) > 0 {
		result, err := strconv.ParseBool(data.CanCheck)
		if err == nil {
			change["datastore_can_check"] = result
		}
	}

	if len(data.Fields) > 0 {
		change["fields"] = data.Fields
	}
	if len(data.MappingID) > 0 {
		change["mapping_id"] = data.MappingID
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyGenerateConfig", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyGenerateConfig", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err = c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("ModifyGenerateConfig", err.Error())
		return err
	}

	return nil
}

// UploadData 上传数据
func UploadData(db string, data []*GItem) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GenDatasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var insertData []interface{}
	for _, v := range data {
		insertData = append(insertData, v)
	}

	queryJSON, _ := json.Marshal(insertData)
	utils.DebugLog("UploadData", fmt.Sprintf("query: [ %s ]", queryJSON))

	if _, err := c.InsertMany(ctx, insertData); err != nil {
		utils.ErrorLog("UploadData", err.Error())
		return err
	}

	return nil
}

// DeleteGenerateConfig 删除导入配置信息
func DeleteGenerateConfig(db, appId, userId string) error {
	client := database.New()
	cg := client.Database(database.GetDBName(db)).Collection(GenConfigsCollection)
	cd := client.Database(database.GetDBName(db)).Collection(GenDatasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":  appId,
		"user_id": userId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteGenerateConfig", fmt.Sprintf("query: [ %s ]", queryJSON))

	if _, err := cg.DeleteOne(ctx, query); err != nil {
		utils.ErrorLog("DeleteGenerateConfig", err.Error())
		return err
	}
	if _, err := cd.DeleteMany(ctx, query); err != nil {
		utils.ErrorLog("DeleteGenerateConfig", err.Error())
		return err
	}

	return nil
}
