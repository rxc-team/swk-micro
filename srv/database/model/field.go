package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// FieldsCollection fields collection
	FieldsCollection = "fields"
)

// 结构体
type (
	//Field 字段
	Field struct {
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
		SelfCalculate     string             `json:"self_calculate" bson:"self_calculate"`
		CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy         string             `json:"created_by" bson:"created_by"`
		UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy         string             `json:"updated_by" bson:"updated_by"`
		DeletedAt         time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy         string             `json:"deleted_by" bson:"deleted_by"`
	}
	// ModifyFieldParam 更改字段的条件参数
	ModifyFieldParam struct {
		FieldID           string
		AppID             string
		DatastoreID       string
		FieldName         string
		FieldType         string
		IsRequired        string
		IsFixed           string
		IsImage           string
		IsCheckImage      string
		AsTitle           string
		Unique            string
		LookupAppID       string
		LookupDatastoreID string
		LookupFieldID     string
		UserGroupID       string
		OptionID          string
		Cols              string
		Rows              string
		X                 string
		Y                 string
		Width             string
		MinLength         string
		MaxLength         string
		MinValue          string
		MaxValue          string
		DisplayOrder      string
		DisplayDigits     string
		Precision         string
		Prefix            string
		ReturnType        string
		Formula           string
		SelfCalculate     string
		IsDisplaySetting  string
		Writer            string
	}
	// FindFieldsParam 查询多个字段的条件参数
	FindFieldsParam struct {
		AppID         string
		DatastoreID   string
		FieldName     string
		FieldType     string
		IsRequired    string
		IsFixed       string
		AsTitle       string
		InvalidatedIn string
	}
	// FindAppFieldsParam 查找APP中多个字段的条件参数
	FindAppFieldsParam struct {
		AppID             string
		FieldType         string
		LookUpDatastoreID string
		InvalidatedIn     string
	}
)

// ToProto 转换为proto数据
func (f *Field) ToProto() *field.Field {
	return &field.Field{
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
		Cols:              f.Cols,
		Rows:              f.Rows,
		X:                 f.X,
		Y:                 f.Y,
		MinLength:         f.MinLength,
		MaxLength:         f.MaxLength,
		MinValue:          f.MinValue,
		MaxValue:          f.MaxValue,
		Width:             f.Width,
		DisplayOrder:      f.DisplayOrder,
		DisplayDigits:     f.DisplayDigits,
		Precision:         f.Precision,
		Prefix:            f.Prefix,
		ReturnType:        f.ReturnType,
		Formula:           f.Formula,
		SelfCalculate:     f.SelfCalculate,
		AsTitle:           f.AsTitle,
		CreatedAt:         f.CreatedAt.String(),
		CreatedBy:         f.CreatedBy,
		UpdatedAt:         f.UpdatedAt.String(),
		UpdatedBy:         f.UpdatedBy,
		DeletedAt:         f.DeletedAt.String(),
		DeletedBy:         f.DeletedBy,
	}
}

// FindAppFields 查找APP中多个字段
func FindAppFields(db string, param *FindAppFieldsParam) (f []Field, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"app_id":     param.AppID,
	}

	// 是否包含无效数据
	if param.InvalidatedIn != "" {
		delete(query, "deleted_by")
	}

	// 字段类型不为空的场合
	if param.FieldType != "" {
		query["field_type"] = param.FieldType
	}

	// 关联字段的所属台账不为空的场合
	if param.LookUpDatastoreID != "" {
		query["lookup_datastore_id"] = param.LookUpDatastoreID
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindAppFields", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "datastore_id", Value: 1},
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: -1},
	}

	opts := options.Find().SetSort(sortItem)

	var result []Field
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("FindAppFields", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var fd Field
		err := cur.Decode(&fd)
		if err != nil {
			utils.ErrorLog("FindAppFields", err.Error())
			return nil, err
		}
		result = append(result, fd)
	}

	return result, nil
}

// FindFields 获取所有的字段
func FindFields(db string, param *FindFieldsParam) (f []Field, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by":   "",
		"datastore_id": param.DatastoreID,
	}

	// 是否包含无效数据
	if param.AppID != "" {
		query["app_id"] = param.AppID
	}

	// 是否包含无效数据
	if param.InvalidatedIn == "true" {
		delete(query, "deleted_by")
	}

	// 字段名不为空的场合
	if param.FieldName != "" {
		query["field_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(param.FieldName), Options: "m"}}
	}

	// 字段类型不为空的场合
	if param.FieldType != "" {
		query["field_type"] = param.FieldType
	}

	// 字段类型不为空的场合
	if param.IsRequired != "" {
		result, err := strconv.ParseBool(param.IsRequired)
		if err == nil {
			query["is_required"] = result
		}
	}

	// 是否固定字段不为空的场合
	if param.IsFixed != "" {
		result, err := strconv.ParseBool(param.IsFixed)
		if err == nil {
			query["is_fixed"] = result
		}
	}

	// 不为空的场合
	if param.AsTitle != "" {
		result, err := strconv.ParseBool(param.AsTitle)
		if err == nil {
			query["as_title"] = result
		}
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindFields", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "display_order", Value: 1},
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)

	var result []Field
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("FindFields", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var fd Field
		err := cur.Decode(&fd)
		if err != nil {
			utils.ErrorLog("FindFields", err.Error())
			return nil, err
		}
		result = append(result, fd)
	}

	return result, nil
}

// FindField 通过ID获取字段信息
func FindField(db, did, fid string) (f Field, err error) {
	client := database.New()
	opts := options.Collection()
	opts.SetReadPreference(readpref.Primary())
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection, opts)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Field

	query := bson.M{
		"field_id":     fid,
		"datastore_id": did,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindField", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			return result, mongo.ErrNoDocuments
		}
		utils.ErrorLog("FindField", err.Error())
		return result, err
	}

	return result, nil
}

// AddField 添加字段
func AddField(db string, f *Field) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	f.ID = primitive.NewObjectID()
	if len(f.FieldID) == 0 {
		f.FieldID = f.ID.Hex()
	}
	f.FieldName = GetFieldNameKey(f.AppID, f.DatastoreID, f.FieldID)

	order, err := getFieldsOrder(db, f.DatastoreID)
	if err != nil {
		utils.ErrorLog("AddField", err.Error())
		return "", err
	}
	f.DisplayOrder = order

	// TODO 添加字段对应的语言到集合中
	/* 	switch f.FieldType {
	   	case "textarea":
	   		f.Rows = 3
	   	case "user":
	   		f.Rows = 2
	   	case "file":
	   		f.Rows = 2
	   	default:
	   		f.Rows = 2
	   	} */

	queryJSON, _ := json.Marshal(f)
	utils.DebugLog("AddField", fmt.Sprintf("Field: [ %s ]", queryJSON))

	if _, err = c.InsertOne(ctx, f); err != nil {
		utils.ErrorLog("AddField", err.Error())
		return "", err
	}

	// 创建自增字段序列
	if f.FieldType == "autonum" {
		err := createAutoSeq(db, f.DatastoreID, f.FieldID)
		if err != nil {
			utils.ErrorLog("AddField", err.Error())
			return "", err
		}

		// 自增字段应该是唯一性字段
		err = AddUniqueKey(db, f.AppID, f.DatastoreID, f.FieldID)
		if err != nil {
			utils.ErrorLog("AddField", err.Error())
			return "", err
		}
	}

	// 添加索引
	if f.Unique {
		go func() {
			// 唯一性字段添加到台账的唯一组合字段关系中
			err := AddUniqueKey(db, f.AppID, f.DatastoreID, f.FieldID)
			if err != nil {
				utils.ErrorLog("AddField", err.Error())
				return
			}
			err = addUniqueIndexToItems(db, f.DatastoreID, f.FieldID, f.FieldType)
			if err != nil {
				utils.ErrorLog("AddField", err.Error())
				return
			}
		}()
	}

	// 同步数据
	go func() {
		if err := addSync(db, f); err != nil {
			utils.ErrorLog("AddField", err.Error())
			return
		}
	}()

	return f.FieldID, nil
}

// VerifyFunc 验证函数是否正确
func VerifyFunc(db, appID, datastoreID, fa, returnType string) (bool, map[string]string, error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       appID,
		"datastore_id": datastoreID,
	}

	var result []Item

	ds, err := getDatastore(db, datastoreID)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return false, nil, err
	}

	fields, err := getFields(db, datastoreID)
	if err != nil {
		utils.ErrorLog("VerifyFunc", err.Error())
		return false, nil, err
	}

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	pipe = append(pipe, bson.M{
		"$limit": 1,
	})

	project := bson.M{
		"_id":          0,
		"item_id":      1,
		"app_id":       1,
		"datastore_id": 1,
		"owners":       1,
		"check_type":   1,
		"check_status": 1,
		"created_at":   1,
		"created_by":   1,
		"updated_at":   1,
		"updated_by":   1,
		"checked_at":   1,
		"checked_by":   1,
		"label_time":   1,
		"status":       1,
	}

	// 关联台账
	for _, relation := range ds.Relations {
		let := bson.M{}
		and := make([]bson.M, 0)
		for relationKey, localKey := range relation.Fields {
			let[localKey] = "$items." + localKey + ".value"

			and = append(and, bson.M{
				"$eq": []interface{}{"$items." + relationKey + ".value", "$$" + localKey},
			})
		}

		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": and,
					},
				},
			},
		}

		lookup := bson.M{
			"from":     "item_" + relation.DatastoreId,
			"let":      let,
			"pipeline": pp,
			"as":       relation.RelationId,
		}

		unwind := bson.M{
			"path":                       "$" + relation.RelationId,
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})
	}

	for _, f := range fields {
		project["items."+f.FieldID] = "$items." + f.FieldID
	}

	// 函数字段，重新拼接
	var formula bson.M
	err = json.Unmarshal([]byte(fa), &formula)
	if err != nil {
		utils.ErrorLog("aaa", err.Error())
		return false, nil, err
	}

	if len(formula) > 0 {
		project["items.text#001.value"] = formula
	} else {
		project["items.text#001.value"] = ""
	}

	project["items.text#001.data_type"] = returnType

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("VerifyFunc", err.Error())
		return false, nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("VerifyFunc", err.Error())
		return false, nil, err
	}

	for _, i := range result {
		for key, it := range i.ItemMap {
			if key == "text#001" {
				err := CheckValueFromModel(it.Value, returnType)
				if err != nil {
					utils.ErrorLog("VerifyFunc", err.Error())
					return false, nil, err
				}
			}
		}
	}

	return true, nil, err
}

// addSync 创建字段时数据同步
func addSync(db string, f *Field) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(f.DatastoreID))
	ctx := context.Background()

	if f.Unique {
		return nil
	}

	if f.FieldType == "autonum" {
		// 默认过滤掉被软删除的数据
		query := bson.D{
			{Key: "app_id", Value: f.AppID},
			{Key: "datastore_id", Value: f.DatastoreID},
		}

		total, err := c.CountDocuments(ctx, query)
		if err != nil {
			utils.ErrorLog("addSync", err.Error())
			return err
		}

		seqList, err := autoNumList(db, f, int(total))
		if err != nil {
			utils.ErrorLog("addSync", err.Error())
			return err
		}

		go func() {

			cur, err := c.Find(ctx, query)
			if err != nil {
				utils.ErrorLog("addSync", err.Error())
				return
			}

			defer cur.Close(ctx)

			index := 0

			for cur.Next(ctx) {
				var item Item
				err := cur.Decode(&item)
				if err != nil {
					utils.ErrorLog("addSync", err.Error())
					return
				}

				data := &Value{
					DataType: f.FieldType,
					Value:    seqList[index],
				}

				change := bson.M{
					"updated_at": time.Now(),
					"updated_by": f.CreatedBy,
				}

				change["items."+f.FieldID] = data
				update := bson.M{"$set": change}

				_, err = c.UpdateByID(ctx, item.ID, update)
				if err != nil {
					utils.ErrorLog("addSync", err.Error())
					return
				}

				index++
			}
		}()

		return nil
	}

	data := &Value{
		DataType: f.FieldType,
		Value:    "",
	}

	switch f.FieldType {
	case "text", "textarea", "options", "function", "lookup", "time":
		data.Value = ""
	case "number":
		data.Value = 0.0
	case "date":
		data.Value = time.Time{}
	case "switch":
		data.Value = false
	case "user":
		data.Value = []string{}
	case "file":
		data.Value = "[]"
	}
	uq := bson.M{
		"app_id":       f.AppID,
		"datastore_id": f.DatastoreID,
	}
	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": f.CreatedBy,
	}

	change["items."+f.FieldID] = data
	update := bson.M{"$set": change}

	_, err := c.UpdateMany(ctx, uq, update)
	if err != nil {
		utils.ErrorLog("addSync", err.Error())
		return err
	}

	return nil
}

// addUniqueIndexToItems 添加字段唯一性索引
func addUniqueIndexToItems(db, datastoreID, field, fieldType string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	key := "items." + field + ".value"

	dropKey := "items." + field + ".value_1"
	c.Indexes().DropOne(ctx, dropKey)

	index := mongo.IndexModel{
		Keys:    bson.D{{Key: key, Value: 1}},
		Options: options.Index().SetSparse(true).SetUnique(true),
	}

	_, e := c.Indexes().CreateOne(ctx, index)
	if e != nil {
		utils.ErrorLog("addUniqueIndexToItems", e.Error())
		return e
	}

	return nil
}

// dropUniqueIndexToItems 删除字段唯一性索引
func dropUniqueIndexToItems(db, datastoreID, field string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dropKey := "items." + field + ".value_1"
	if _, e := c.Indexes().DropOne(ctx, dropKey); e != nil {
		if !strings.HasPrefix(e.Error(), "(IndexNotFound) index not found with name") {
			utils.ErrorLog("dropUniqueIndexToItems", e.Error())
			return e
		}
	}

	return nil
}

// BlukAddField 批量添加字段
func BlukAddField(db string, fields []*Field) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("BlukAddField", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("BlukAddField", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, field := range fields {
			field.ID = primitive.NewObjectID()
			// if len(field.FieldID) == 0{
			// 	field.FieldID = field.ID.Hex()
			// }
			field.FieldID = field.ID.Hex()

			switch field.FieldType {
			case "textarea":
				field.Rows = 3
			case "user":
				field.Rows = 2
			case "file":
				field.Rows = 2
			}

			queryJSON, _ := json.Marshal(field)
			utils.DebugLog("BlukAddField", fmt.Sprintf("field: [ %s ]", queryJSON))

			_, err = c.InsertOne(sc, field)
			if err != nil {
				utils.ErrorLog("BlukAddField", err.Error())
				return err
			}

			// 添加索引
			if field.Unique {
				addUniqueIndexToItems(db, field.DatastoreID, field.FieldID, field.FieldType)
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("BlukAddField", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("BlukAddField", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// ModifyField 修改字段
func ModifyField(db string, p *ModifyFieldParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"field_id":     p.FieldID,
		"datastore_id": p.DatastoreID,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": p.Writer,
	}

	// 字段类型不为空的场合
	if p.IsRequired != "" {
		result, err := strconv.ParseBool(p.IsRequired)
		if err == nil {
			change["is_required"] = result
		}
	}

	// 是否固定字段不为空的场合
	if p.IsFixed != "" {
		result, err := strconv.ParseBool(p.IsFixed)
		if err == nil {
			change["is_fixed"] = result
		}
	}

	// 文件图片类型不为空的场合
	if p.IsImage != "" {
		result, err := strconv.ParseBool(p.IsImage)
		if err == nil {
			change["is_image"] = result
		}
	}

	// 文件盘点图片类型不为空的场合
	if p.IsCheckImage != "" {
		result, err := strconv.ParseBool(p.IsCheckImage)
		if err == nil {
			change["is_check_image"] = result
		}
	}

	// 字段作为标题不为空的场合
	if p.AsTitle != "" {
		result, err := strconv.ParseBool(p.AsTitle)
		if err == nil {
			change["as_title"] = result
		}
	}

	// 字段唯一不为空的场合
	if p.Unique != "" {
		result, err := strconv.ParseBool(p.Unique)
		if err == nil {
			change["unique"] = result
		}

		if result {
			addUniqueIndexToItems(db, p.DatastoreID, p.FieldID, p.FieldType)
		} else {
			dropUniqueIndexToItems(db, p.DatastoreID, p.FieldID)
		}
	}

	// 关联字段的所属APP不为空的场合
	if p.LookupAppID != "" {
		change["lookup_app_id"] = p.LookupAppID
	}
	// 关联字段的所属台账不为空的场合
	if p.LookupDatastoreID != "" {
		change["lookup_datastore_id"] = p.LookupDatastoreID
	}
	// 关联字段的ID不为空的场合
	if p.LookupFieldID != "" {
		change["lookup_field_id"] = p.LookupFieldID
	}
	if p.IsDisplaySetting != "true" {
		// 字段的序列表示前綴
		change["prefix"] = p.Prefix
	}

	// 用户组不为空的场合
	if p.UserGroupID != "" {
		change["user_group_id"] = p.UserGroupID
	}
	// 选项组不为空的场合
	if p.OptionID != "" {
		change["option_id"] = p.OptionID
	}
	// 返回类型不为空的场合
	if p.ReturnType != "" {
		change["return_type"] = p.ReturnType
	}
	// 公式不为空的场合
	if p.Formula != "" {
		change["formula"] = p.Formula
	}
	// 自算不为空的场合
	if p.SelfCalculate != "" {
		change["self_calculate"] = p.SelfCalculate
	}

	// 字段的位置不为空的场合
	if p.Cols != "" {
		result, err := strconv.ParseInt(p.Cols, 10, 64)
		if err == nil {
			change["cols"] = result
		}
	}
	if p.Rows != "" {
		result, err := strconv.ParseInt(p.Rows, 10, 64)
		if err == nil {
			change["rows"] = result
		}
	}
	if p.X != "" {
		result, err := strconv.ParseInt(p.X, 10, 64)
		if err == nil {
			change["x"] = result
		}
	}
	if p.Y != "" {
		result, err := strconv.ParseInt(p.Y, 10, 64)
		if err == nil {
			change["y"] = result
		}
	}
	if p.MinLength != "" {
		result, err := strconv.ParseInt(p.MinLength, 10, 64)
		if err == nil {
			change["min_length"] = result
		}
	}
	if p.MaxLength != "" {
		result, err := strconv.ParseInt(p.MaxLength, 10, 64)
		if err == nil {
			change["max_length"] = result
		}
	}
	if p.MinValue != "" {
		result, err := strconv.ParseInt(p.MinValue, 10, 64)
		if err == nil {
			change["min_value"] = result
		}
	}
	if p.MaxValue != "" {
		result, err := strconv.ParseInt(p.MaxValue, 10, 64)
		if err == nil {
			change["max_value"] = result
		}
	}

	// 字段的宽度不为空的场合
	if p.Width != "" {
		result, err := strconv.ParseInt(p.Width, 10, 64)
		if err == nil {
			change["width"] = result
		}
	}

	// 字段的表示顺不为空的场合
	if p.DisplayOrder != "" {
		result, err := strconv.ParseInt(p.DisplayOrder, 10, 64)
		if err == nil {
			change["display_order"] = result
		}
	}

	// 字段的表示位數
	if p.DisplayDigits != "" {
		result, err := strconv.ParseInt(p.DisplayDigits, 10, 64)
		if err == nil {
			change["display_digits"] = result
		}
	}

	// 字段的小数精度
	if p.Precision != "" {
		result, err := strconv.ParseInt(p.Precision, 10, 64)
		if err == nil {
			change["precision"] = result
		}
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyField", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyField", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err = c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("ModifyField", err.Error())
		return err
	}

	return nil
}

// DeleteField 删除单个字段
func DeleteField(db, datastoreID, fieldID, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"field_id":     fieldID,
		"datastore_id": datastoreID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteField", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteField", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err := c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("DeleteField", err.Error())
		return err
	}

	return nil
}

// DeleteDatastoreFields 物理删除该台账下的所有字段
func DeleteDatastoreFields(db, datastoreID, userID string) error {
	client := database.New()
	cf := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	cd := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteDatastoreFields", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteDatastoreFields", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// 删除台账字段
		query := bson.M{
			"datastore_id": datastoreID,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("DeleteDatastoreFields", fmt.Sprintf("query: [ %s ]", queryJSON))

		if _, err := cf.DeleteMany(ctx, query); err != nil {
			utils.ErrorLog("DeleteDatastoreFields", err.Error())
			return err
		}

		// 删除台账数据
		if err := cd.Drop(ctx); err != nil {
			utils.ErrorLog("DeleteDatastoreFields", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteDatastoreFields", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteDatastoreFields", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// DeleteSelectFields 删除选中的字段
func DeleteSelectFields(db, datastoreID string, fieldIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteSelectFields", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteSelectFields", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, fieldID := range fieldIDList {

			query := bson.M{
				"field_id":     fieldID,
				"datastore_id": datastoreID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSelectFields", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectFields", fmt.Sprintf("update: [ %s ]", updateJSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("DeleteSelectFields", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteSelectFields", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteSelectFields", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// HardDeleteFields 物理删除选中的字段
func HardDeleteFields(ctx context.Context, db, datastoreID string, fieldIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ic := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	dc := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)

	opts := &options.SessionOptions{}
	opts.SetDefaultReadPreference(readpref.Primary())
	opts.SetDefaultReadConcern(readconcern.Snapshot())

	session, err := client.StartSession(opts)
	if err != nil {
		utils.ErrorLog("HardDeleteFields", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("HardDeleteFields", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		ds, err := getDatastore(db, datastoreID)
		if err != nil {
			return err
		}

		for _, fieldID := range fieldIDList {

			query := bson.M{
				"field_id":     fieldID,
				"datastore_id": datastoreID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteFields", fmt.Sprintf("query: [ %s ]", queryJSON))

			var field Field
			err := c.FindOne(sc, query).Decode(&field)
			if err != nil {
				utils.ErrorLog("HardDeleteFields", err.Error())
				return err
			}

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("HardDeleteFields", err.Error())
				return err
			}

			// 删除台账数据
			query1 := bson.M{
				"datastore_id": datastoreID,
			}
			update1 := bson.M{
				"$unset": bson.M{"items." + fieldID: ""},
			}

			if _, err := ic.UpdateMany(sc, query1, update1); err != nil {
				utils.ErrorLog("HardDeleteFields", err.Error())
				return err
			}

			// 删除扫描字段设置和默认排序字段设置
			query2 := bson.M{
				"datastore_id": datastoreID,
			}
			update2 := bson.M{
				"$pull": bson.M{
					"scan_fields": fieldID,
					"sorts": bson.M{
						"sort_key": fieldID,
					},
				},
			}
			if _, e := dc.UpdateOne(sc, query2, update2); e != nil {
				utils.ErrorLog("HardDeleteFields", e.Error())
				return e
			}

			if field.FieldType == "autonum" {

				// // 删除uniquekey
				// err = deleteUniqueKey(sc, db, datastoreID, fieldID)
				// if err != nil {
				// 	utils.ErrorLog("HardDeleteFields", err.Error())
				// 	return err
				// }

				// 删除自动採番的seq
				if e := deleteAutoSeq(db, datastoreID, fieldID); e != nil {
					utils.ErrorLog("HardDeleteFields", e.Error())
					return e
				}
			}

			// if field.Unique {
			// 	// 删除uniquekey
			// 	err = deleteUniqueKey(sc, db, datastoreID, fieldID)
			// 	if err != nil {
			// 		utils.ErrorLog("HardDeleteFields", err.Error())
			// 		return err
			// 	}
			// }

			// 删除关联组合唯一性字段
			for _, f := range ds.UniqueFields {

				fields := strings.Split(f, ",")
				hasExist := false
				for _, uf := range fields {
					if uf == fieldID {
						hasExist = true
						break
					}
				}

				if hasExist {
					// 删除uniquekey
					err = deleteUniqueKey(sc, db, datastoreID, f)
					if err != nil {
						utils.ErrorLog("HardDeleteFields", err.Error())
						return err
					}
				}

			}

			for _, relation := range ds.Relations {
				for _, localKey := range relation.Fields {
					if localKey == fieldID {
						deleteRelation(sc, db, datastoreID, relation.RelationId)
					}
				}
			}

			// 删除mapping中的该字段
			for i := range ds.Mappings {

				// 删除扫描字段设置和默认排序字段设置
				query := bson.M{
					"datastore_id": datastoreID,
				}

				key := strings.Builder{}
				key.WriteString("mappings.")
				key.WriteString(cast.ToString(i))
				key.WriteString(".mapping_rule")

				update := bson.M{
					"$pull": bson.M{
						key.String(): bson.M{
							"from_key": "field_2",
						},
					},
				}
				if _, e := dc.UpdateOne(sc, query, update); e != nil {
					utils.ErrorLog("HardDeleteFields", e.Error())
					return e
				}
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("HardDeleteFields", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("HardDeleteFields", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

func deleteRelation(ctx mongo.SessionContext, db, datastoreId, relationId string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)

	query := bson.M{
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

func deleteUniqueKey(ctx mongo.SessionContext, db, datastoreId, uniqueFields string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DataStoresCollection)
	ct := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreId))

	query := bson.M{
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
		// 删除唯一索引
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

		if _, err := ct.Indexes().DropOne(context.TODO(), name); err != nil {

			if !strings.HasPrefix(err.Error(), "(IndexNotFound) index not found with name") {
				utils.ErrorLog("HardDeleteFields", err.Error())
				return
			}

			return
		}
	}()

	return nil
}

// RecoverSelectFields 恢复选中的字段
func RecoverSelectFields(db, datastoreID string, fieldIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("RecoverSelectFields", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("RecoverSelectFields", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, fieldID := range fieldIDList {

			query := bson.M{
				"field_id":     fieldID,
				"datastore_id": datastoreID,
			}

			update := bson.M{"$set": bson.M{
				"updated_at": time.Now(),
				"updated_by": userID,
				"deleted_by": "",
			}}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("RecoverSelectFields", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectFields", fmt.Sprintf("update: [ %s ]", updateJSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("RecoverSelectFields", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("RecoverSelectFields", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("RecoverSelectFields", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}
