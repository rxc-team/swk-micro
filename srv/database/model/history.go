package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/database/proto/datahistory"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
	"rxcsoft.cn/utils/timex"
)

const (
	// HistoriesCollection histories collection
	HistoriesCollection      = "data_histories"
	FieldHistoriesCollection = "field_histories"
)

type (
	// History 台账按操作为单位的履历的数据
	History struct {
		ID           primitive.ObjectID `json:"id" bson:"_id"`
		HistoryID    string             `json:"history_id" bson:"history_id"`
		HistoryType  string             `json:"history_type" bson:"history_type"`
		DatastoreID  string             `json:"datastore_id" bson:"datastore_id"`
		ItemID       string             `json:"item_id" bson:"item_id"`
		FixedItems   ItemMap            `json:"fixed_items" bson:"fixed_items"`
		Changes      []*Change          `json:"changes" bson:"changes"`
		TotalChanges int64              `json:"total_changes" bson:"total_changes"`
		CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy    string             `json:"created_by" bson:"created_by"`
	}
	// FieldHistory 台账按字段为单位的履历的数据
	FieldHistory struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		HistoryID   string             `json:"history_id" bson:"history_id"`
		HistoryType string             `json:"history_type" bson:"history_type"`
		DatastoreID string             `json:"datastore_id" bson:"datastore_id"`
		ItemID      string             `json:"item_id" bson:"item_id"`
		FieldID     string             `json:"field_id" bson:"field_id"`
		LocalName   string             `json:"local_name" bson:"local_name"`
		FieldName   string             `json:"field_name" bson:"field_name"`
		OldValue    string             `json:"old_value" bson:"old_value"`
		NewValue    string             `json:"new_value" bson:"new_value"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
	}

	// DownloadHistory 台账按字段为单位的履历的数据
	DownloadHistory struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		HistoryID   string             `json:"history_id" bson:"history_id"`
		HistoryType string             `json:"history_type" bson:"history_type"`
		DatastoreID string             `json:"datastore_id" bson:"datastore_id"`
		ItemID      string             `json:"item_id" bson:"item_id"`
		FieldID     string             `json:"field_id" bson:"field_id"`
		LocalName   string             `json:"local_name" bson:"local_name"`
		FieldName   string             `json:"field_name" bson:"field_name"`
		OldValue    string             `json:"old_value" bson:"old_value"`
		NewValue    string             `json:"new_value" bson:"new_value"`
		FixedItems  ItemMap            `json:"fixed_items" bson:"fixed_items"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
	}

	Change struct {
		FieldID   string `json:"field_id" bson:"field_id"`
		LocalName string `json:"local_name" bson:"local_name"`
		FieldName string `json:"field_name" bson:"field_name"`
		OldValue  string `json:"old_value" bson:"old_value"`
		NewValue  string `json:"new_value" bson:"new_value"`
	}

	// FindParam 分页查询多条记录
	FindParam struct {
		ItemID        string
		DatastoreID   string
		FieldID       string
		CreatedAtFrom string
		CreatedAtTo   string
		HistoryType   string
		OldValue      string
		NewValue      string
		FieldList     []string
		PageIndex     int64
		PageSize      int64
	}
)

// ToProto 转换为proto数据
func (h *History) ToProto() *datahistory.History {
	changes := make([]*datahistory.Change, 0)
	for _, c := range h.Changes {
		change := &datahistory.Change{
			FieldId:   c.FieldID,
			FieldName: c.FieldName,
			LocalName: c.LocalName,
			OldValue:  c.OldValue,
			NewValue:  c.NewValue,
		}
		if change.OldValue == "0001-01-01" {
			change.OldValue = ""
		}
		if change.NewValue == "0001-01-01" {
			change.NewValue = ""
		}
		changes = append(changes, change)
	}
	fixedItems := make(map[string]*datahistory.FixedValue, len(h.FixedItems))
	for key, it := range h.FixedItems {
		item := &datahistory.FixedValue{
			DataType: it.DataType,
			Value:    GetValueFromModel(it),
		}
		if item.DataType == "date" && item.Value == "0001-01-01" {
			item.Value = ""
		}

		fixedItems[key] = item
	}
	return &datahistory.History{
		HistoryId:    h.HistoryID,
		HistoryType:  h.HistoryType,
		DatastoreId:  h.DatastoreID,
		ItemId:       h.ItemID,
		FixedItems:   fixedItems,
		Changes:      changes,
		TotalChanges: h.TotalChanges,
		CreatedAt:    h.CreatedAt.String(),
		CreatedBy:    h.CreatedBy,
	}
}

// ToProto 转换为proto数据
func (h *FieldHistory) ToProto() *datahistory.FieldHistory {
	history := &datahistory.FieldHistory{
		HistoryId:   h.HistoryID,
		HistoryType: h.HistoryType,
		DatastoreId: h.DatastoreID,
		ItemId:      h.ItemID,
		FieldId:     h.FieldID,
		LocalName:   h.LocalName,
		FieldName:   h.FieldName,
		OldValue:    h.OldValue,
		NewValue:    h.NewValue,
		CreatedAt:   h.CreatedAt.String(),
		CreatedBy:   h.CreatedBy,
	}
	if history.OldValue == "0001-01-01" {
		history.OldValue = ""
	}
	if history.NewValue == "0001-01-01" {
		history.NewValue = ""
	}
	return history
}

// ToProto 转换为proto数据
func (v *DownloadHistory) ToProto() *datahistory.DownloadHistory {

	fixedItems := make(map[string]*datahistory.FixedValue, len(v.FixedItems))
	for key, it := range v.FixedItems {
		item := &datahistory.FixedValue{
			DataType: it.DataType,
			Value:    GetValueFromModel(it),
		}
		if item.DataType == "date" && item.Value == "0001-01-01" {
			item.Value = ""
		}
		fixedItems[key] = item
	}

	history := &datahistory.DownloadHistory{
		HistoryId:   v.HistoryID,
		HistoryType: v.HistoryType,
		DatastoreId: v.DatastoreID,
		ItemId:      v.ItemID,
		FieldId:     v.FieldID,
		LocalName:   v.LocalName,
		FieldName:   v.FieldName,
		OldValue:    v.OldValue,
		NewValue:    v.NewValue,
		FixedItems:  fixedItems,
		CreatedAt:   v.CreatedAt.String(),
		CreatedBy:   v.CreatedBy,
	}
	if history.OldValue == "0001-01-01" {
		history.OldValue = ""
	}
	if history.NewValue == "0001-01-01" {
		history.NewValue = ""
	}
	return history
}

// FindHistories 通过appID获取Datastore信息
func FindHistories(ctx context.Context, db string, params FindParam) (items []FieldHistory, total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldHistoriesCollection)

	query := bson.M{
		"field_id": bson.M{
			"$in": params.FieldList,
		},
	}

	// 台账ID不为空的场合，添加到查询条件中
	if params.DatastoreID != "" {
		query["datastore_id"] = params.DatastoreID
	}
	// ItemID不为空的场合，添加到查询条件中
	if params.ItemID != "" {
		query["item_id"] = params.ItemID
	}
	// 履历类型不为空的场合，添加到查询条件中
	if params.HistoryType != "" {
		query["history_type"] = params.HistoryType
	}

	// 字段不为空的场合，添加到查询条件中
	if params.FieldID != "" {
		query["field_id"] = params.FieldID
	}
	// 变更前
	if params.OldValue != "" {
		query["old_value"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(params.OldValue), Options: "m"}}
	}
	// 变更后
	if params.NewValue != "" {
		query["new_value"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(params.NewValue), Options: "m"}}
	}

	// 时间不为空的场合，添加到查询条件中
	if params.CreatedAtFrom != "" && params.CreatedAtTo != "" {

		from, _ := time.Parse(DateFormat, params.CreatedAtFrom)
		to, _ := time.Parse(DateFormat, params.CreatedAtTo)

		query["$and"] = []bson.M{
			{
				"created_at": bson.M{
					"$gte": from,
				},
			},
			{
				"created_at": bson.M{
					"$lt": to,
				},
			},
		}
	}

	var result []FieldHistory
	// 取总件数
	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindHistories", err.Error())
		return result, 0, err
	}

	// 聚合查询
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindHistories", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "history_id", Value: -1},
		{Key: "created_at", Value: -1},
	}

	opt := options.Find()
	opt.SetSort(sortItem)

	skip := (params.PageIndex - 1) * params.PageSize
	limit := params.PageSize

	if skip > 0 {
		opt.SetSkip(skip)
	}

	if limit > 0 {
		opt.SetLimit(limit)
	}

	cur, err := c.Find(ctx, query, opt)
	if err != nil {
		utils.ErrorLog("FindHistories", err.Error())
		return result, 0, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var his FieldHistory
		err := cur.Decode(&his)
		if err != nil {
			utils.ErrorLog("FindHistories", err.Error())
			return result, 0, err
		}
		result = append(result, his)
	}

	return result, t, nil
}

// Download 通过appID获取Datastore信息
func DownloadHistories(ctx context.Context, db string, params FindParam, stream datahistory.HistoryService_DownloadStream) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldHistoriesCollection)

	query := bson.M{
		"field_id": bson.M{
			"$in": params.FieldList,
		},
	}

	// 台账ID不为空的场合，添加到查询条件中
	if params.DatastoreID != "" {
		query["datastore_id"] = params.DatastoreID
	}
	// ItemID不为空的场合，添加到查询条件中
	if params.ItemID != "" {
		query["item_id"] = params.ItemID
	}
	// 履历类型不为空的场合，添加到查询条件中
	if params.HistoryType != "" {
		query["history_type"] = params.HistoryType
	}

	// 字段不为空的场合，添加到查询条件中
	if params.FieldID != "" {
		query["field_id"] = params.FieldID
	}
	// 变更前
	if params.OldValue != "" {
		query["old_value"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(params.OldValue), Options: "m"}}
	}
	// 变更后
	if params.NewValue != "" {
		query["new_value"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(params.NewValue), Options: "m"}}
	}

	// 时间不为空的场合，添加到查询条件中
	if params.CreatedAtFrom != "" && params.CreatedAtTo != "" {

		from, _ := time.Parse(DateFormat, params.CreatedAtFrom)
		to, _ := time.Parse(DateFormat, params.CreatedAtTo)

		query["$and"] = []bson.M{
			{
				"created_at": bson.M{
					"$gte": from,
				},
			},
			{
				"created_at": bson.M{
					"$lt": to,
				},
			},
		}
	}

	pipe := mongo.Pipeline{}
	pipe = append(pipe, bson.D{
		{Key: "$match", Value: query},
	})

	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}

	pipe = append(pipe, bson.D{
		{Key: "$sort", Value: sortItem},
	})

	pp := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$history_id", "$$id"},
						},
					},
				},
			},
		},
	}

	lookup := bson.M{
		"from": "data_histories",
		"let": bson.M{
			"id": "$history_id",
		},
		"pipeline": pp,
		"as":       "history_detail",
	}

	pipe = append(pipe, bson.D{
		{Key: "$lookup", Value: lookup},
	})

	unwind := bson.M{
		"path":                       "$history_detail",
		"preserveNullAndEmptyArrays": true,
	}

	pipe = append(pipe, bson.D{
		{Key: "$unwind", Value: unwind},
	})

	project := bson.M{
		"_id":          1,
		"history_id":   1,
		"history_type": 1,
		"datastore_id": 1,
		"item_id":      1,
		"field_id":     1,
		"local_name":   1,
		"field_name":   1,
		"old_value":    1,
		"new_value":    1,
		"checked_at":   1,
		"fixed_items":  "$history_detail.fixed_items",
		"created_at":   1,
		"created_by":   1,
	}

	pipe = append(pipe, bson.D{
		{Key: "$project", Value: project},
	})

	// 聚合查询
	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("Download", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("Download", err.Error())
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var his DownloadHistory
		err := cur.Decode(&his)
		if err != nil {
			utils.ErrorLog("Download", err.Error())
			return err
		}

		if err := stream.Send(&datahistory.DownloadResponse{History: his.ToProto()}); err != nil {
			utils.ErrorLog("Download", err.Error())
			return err
		}
	}

	return nil
}

// FindHistoryCount 获取履历件数
func FindHistoryCount(ctx context.Context, db string, params FindParam) (total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldHistoriesCollection)

	query := bson.M{
		"field_id": bson.M{
			"$in": params.FieldList,
		},
	}

	// 台账ID不为空的场合，添加到查询条件中
	if params.DatastoreID != "" {
		query["datastore_id"] = params.DatastoreID
	}
	// ItemID不为空的场合，添加到查询条件中
	if params.ItemID != "" {
		query["item_id"] = params.ItemID
	}
	// 履历类型不为空的场合，添加到查询条件中
	if params.HistoryType != "" {
		query["history_type"] = params.HistoryType
	}

	// 字段不为空的场合，添加到查询条件中
	if params.FieldID != "" {
		query["field_id"] = params.FieldID
	}
	// 变更前
	if params.OldValue != "" {
		query["old_value"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(params.OldValue), Options: "m"}}
	}
	// 变更后
	if params.NewValue != "" {
		query["new_value"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(params.NewValue), Options: "m"}}
	}

	// 时间不为空的场合，添加到查询条件中
	if params.CreatedAtFrom != "" && params.CreatedAtTo != "" {

		from, _ := time.Parse(DateFormat, params.CreatedAtFrom)
		to, _ := time.Parse(DateFormat, params.CreatedAtTo)

		query["$and"] = []bson.M{
			{
				"created_at": bson.M{
					"$gte": from,
				},
			},
			{
				"created_at": bson.M{
					"$lt": to,
				},
			},
		}
	}

	// 取总件数
	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindHistoryCount", err.Error())
		return 0, err
	}

	return t, nil
}

// FindLastHistories 查找最新的10条记录
func FindLastHistories(ctx context.Context, db, datastoreId, itemId string, fieldList []string) (t int64, hs []*History, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HistoriesCollection)

	query := bson.M{
		"datastore_id": datastoreId,
		"item_id":      itemId,
	}

	total, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindLastHistories", err.Error())
		return 0, nil, err
	}

	pipe := mongo.Pipeline{}
	pipe = append(pipe, bson.D{
		{Key: "$match", Value: query},
	})

	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}

	pipe = append(pipe, bson.D{
		{Key: "$sort", Value: sortItem},
	})

	pipe = append(pipe, bson.D{
		{Key: "$limit", Value: 10},
	})

	pp := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$history_id", "$$id"},
						},
						{
							"$in": []string{"$field_id", "$$fs"},
						},
					},
				},
			},
		},
	}

	lookup := bson.M{
		"from": "field_histories",
		"let": bson.M{
			"id": "$history_id",
			"fs": fieldList,
		},
		"pipeline": pp,
		"as":       "changes",
	}
	pp1 := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$history_id", "$$id"},
						},
					},
				},
			},
		},
	}

	lookup1 := bson.M{
		"from": "field_histories",
		"let": bson.M{
			"id": "$history_id",
		},
		"pipeline": pp1,
		"as":       "total_changes",
	}

	pipe = append(pipe, bson.D{
		{Key: "$lookup", Value: lookup},
	})
	pipe = append(pipe, bson.D{
		{Key: "$lookup", Value: lookup1},
	})

	project := bson.M{
		"_id":          1,
		"history_id":   1,
		"history_type": 1,
		"datastore_id": 1,
		"item_id":      1,
		"fixed_items":  1,
		"changes":      1,
		"total_changes": bson.M{
			"$size": "$total_changes",
		},
		"created_at": 1,
		"created_by": 1,
	}

	pipe = append(pipe, bson.D{
		{Key: "$project", Value: project},
	})

	var result []*History

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindLastHistories", fmt.Sprintf("query: [ %s ]", queryJSON))

	cur, err := c.Aggregate(ctx, pipe)
	if err != nil {
		utils.ErrorLog("FindLastHistories", err.Error())
		return total, nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &result); err != nil {
		utils.ErrorLog("FindLastHistories", err.Error())
		return total, nil, err
	}

	return total, result, nil
}

// FindHistory 通过ID获取数据信息
func FindHistory(ctx context.Context, db, historyID string, fieldList []string) (hs *History, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HistoriesCollection)

	query := bson.M{
		"history_id": historyID,
	}

	pipe := mongo.Pipeline{}
	pipe = append(pipe, bson.D{
		{Key: "$match", Value: query},
	})

	pp := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$history_id", "$$id"},
						},
						{
							"$in": []string{"$field_id", "$$fs"},
						},
					},
				},
			},
		},
	}

	lookup := bson.M{
		"from": "field_histories",
		"let": bson.M{
			"id": "$history_id",
			"fs": fieldList,
		},
		"pipeline": pp,
		"as":       "changes",
	}

	pipe = append(pipe, bson.D{
		{Key: "$lookup", Value: lookup},
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindHistory", fmt.Sprintf("query: [ %s ]", queryJSON))

	cur, err := c.Aggregate(ctx, pipe)
	if err != nil {
		utils.ErrorLog("FindHistory", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	if cur.Next(ctx) {
		var his History

		if err := cur.Decode(&his); err != nil {
			utils.ErrorLog("FindHistory", err.Error())
			return nil, err
		}

		return &his, nil
	}

	return nil, fmt.Errorf("no found")
}

// AddHistory 添加台账数据
func AddHistory(db string, h *History) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HistoriesCollection)
	fc := client.Database(database.GetDBName(db)).Collection(FieldHistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	h.ID = primitive.NewObjectID()
	h.HistoryID = timex.Timestamp()

	// Hs 台账按操作为单位的履历的数据
	type Hs struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		HistoryID   string             `json:"history_id" bson:"history_id"`
		HistoryType string             `json:"history_type" bson:"history_type"`
		DatastoreID string             `json:"datastore_id" bson:"datastore_id"`
		ItemID      string             `json:"item_id" bson:"item_id"`
		FixedItems  ItemMap            `json:"fixed_items" bson:"fixed_items"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
	}

	hs := Hs{
		ID:          h.ID,
		HistoryID:   h.HistoryID,
		HistoryType: h.HistoryType,
		DatastoreID: h.DatastoreID,
		ItemID:      h.ItemID,
		FixedItems:  h.FixedItems,
		CreatedAt:   h.CreatedAt,
		CreatedBy:   h.CreatedBy,
	}

	queryJSON, _ := json.Marshal(hs)
	utils.DebugLog("AddHistory", fmt.Sprintf("history: [ %s ]", queryJSON))

	if _, err := c.InsertOne(ctx, hs); err != nil {
		utils.ErrorLog("AddHistory", err.Error())
		return err
	}

	// 字段变更履历
	for _, c := range h.Changes {
		fh := FieldHistory{
			ID:          h.ID,
			HistoryID:   h.HistoryID,
			HistoryType: h.HistoryType,
			DatastoreID: h.DatastoreID,
			ItemID:      h.ItemID,
			FieldID:     c.FieldID,
			FieldName:   c.FieldName,
			LocalName:   c.LocalName,
			OldValue:    c.OldValue,
			NewValue:    c.NewValue,
			CreatedAt:   h.CreatedAt,
			CreatedBy:   h.CreatedBy,
		}

		queryJSON, _ := json.Marshal(fh)
		utils.DebugLog("AddHistory", fmt.Sprintf("history: [ %s ]", queryJSON))

		if _, err := fc.InsertOne(ctx, fh); err != nil {
			utils.ErrorLog("AddHistory", err.Error())
			return err
		}
	}

	return nil
}

// CreateIndex 创建history索引
func CreateIndex(customerID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(customerID)).Collection(HistoriesCollection)
	fc := client.Database(database.GetDBName(customerID)).Collection(FieldHistoriesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hIndexs := []mongo.IndexModel{}

	// 单个索引，方便关联查询
	hIndexs = append(hIndexs, mongo.IndexModel{
		Keys: bson.D{
			{Key: "history_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	// 复合索引，方便普通查询
	hIndexs = append(hIndexs, mongo.IndexModel{
		Keys: bson.D{
			{Key: "datastore_id", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "item_id", Value: 1},
			{Key: "history_type", Value: 1},
		},
		Options: options.Index(),
	})

	if _, err := c.Indexes().CreateMany(ctx, hIndexs); err != nil {
		utils.ErrorLog("CreateIndex", err.Error())
		return err
	}

	fhIndexs := []mongo.IndexModel{}

	// 单个索引，方便关联查询
	fhIndexs = append(fhIndexs, mongo.IndexModel{
		Keys: bson.D{
			{Key: "history_id", Value: 1},
		},
		Options: options.Index(),
	})

	// 复合索引，方便普通查询
	fhIndexs = append(fhIndexs, mongo.IndexModel{
		Keys: bson.D{
			{Key: "datastore_id", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "field_id", Value: 1},
			{Key: "item_id", Value: 1},
			{Key: "history_type", Value: 1},
		},
		Options: options.Index(),
	})

	if _, err := fc.Indexes().CreateMany(ctx, fhIndexs); err != nil {
		utils.ErrorLog("CreateIndex", err.Error())
		return err
	}

	return nil
}
