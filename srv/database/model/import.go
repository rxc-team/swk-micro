package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"rxcsoft.cn/pit3/srv/database/proto/feed"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// ImportCollection import collection
	ImportCollection = "import_tmps"
)

type (
	// ImItem 导入数据
	ImItem struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		AppID       string             `json:"app_id" bson:"app_id"`
		DatastoreID string             `json:"datastore_id" bson:"datastore_id"`
		MappingID   string             `json:"mapping_id" bson:"mapping_id"`
		JobID       string             `json:"job_id" bson:"job_id"`
		Items       map[string]string  `json:"items" bson:"items"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
	}
	// Request 请求参数
	Request struct {
		AppID       string `json:"app_id" bson:"app_id"`
		DatastoreID string `json:"datastore_id" bson:"datastore_id"`
		MappingID   string `json:"mapping_id" bson:"mapping_id"`
		JobID       string `json:"job_id" bson:"job_id"`
		Writer      string `json:"writer" bson:"writer"`
	}
)

// ToProto 转换为proto数据
func (i *ImItem) ToProto() *feed.ImportItem {
	return &feed.ImportItem{
		AppId:       i.AppID,
		DatastoreId: i.DatastoreID,
		MappingId:   i.MappingID,
		JobId:       i.JobID,
		Items:       i.Items,
		CreatedBy:   i.CreatedBy,
		CreatedAt:   i.CreatedAt.String(),
	}
}

// FindImports 查找导入的记录
func FindImports(db string, r Request) (items []*ImItem, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ImportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	match := bson.M{
		"app_id":       r.AppID,
		"datastore_id": r.DatastoreID,
		"job_id":       r.JobID,
		"created_by":   r.Writer,
	}

	query := []bson.M{
		{"$match": match},
	}

	group := []bson.M{
		{"$group": bson.M{
			"_id": bson.M{
				"app_id":       "$app_id",
				"datastore_id": "$datastore_id",
				"job_id":       "$job_id",
				"created_by":   "$created_by",
				"created_at":   "$created_at",
			},
			"mapping_id": bson.M{
				"$first": "$mapping_id",
			},
		}},
		{"$project": bson.M{
			"_id":          0,
			"app_id":       "$_id.app_id",
			"datastore_id": "$_id.datastore_id",
			"job_id":       "$job_id",
			"mapping_id":   "$mapping_id",
			"created_by":   "$_id.created_by",
			"created_at":   "$_id.created_at",
		}},
		{
			"$sort": bson.M{
				"created_at": 1,
			},
		},
	}

	query = append(query, group...)

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindImports", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []*ImItem
	cur, err := c.Aggregate(ctx, query)
	if err != nil {
		utils.ErrorLog("FindImports", err.Error())
		return result, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item ImItem
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindImports", err.Error())
			return result, err
		}
		result = append(result, &item)
	}

	return result, nil
}

// FindImportItems 查找导入的数据
func FindImportItems(db string, r Request) (items []*ImItem, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ImportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       r.AppID,
		"datastore_id": r.DatastoreID,
		"job_id":       r.JobID,
		"created_by":   r.Writer,
		"mapping_id":   r.MappingID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindImportItems", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []*ImItem
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("FindImportItems", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item ImItem
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindImportItems", err.Error())
			return nil, err
		}
		result = append(result, &item)
	}

	return result, nil
}

// AddImportItem 通过ID获取数据信息
func AddImportItem(db string, items []*ImItem) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ImportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var models []mongo.WriteModel

	for _, it := range items {
		it.ID = primitive.NewObjectID()

		insertModel := mongo.NewInsertOneModel()
		insertModel.SetDocument(it)

		models = append(models, insertModel)
	}

	if len(models) > 0 {
		rs, err := c.BulkWrite(ctx, models)
		if err != nil {
			utils.ErrorLog("AddImportItem", err.Error())
			return err
		}

		log.Debugf("add result %v", rs)
	}

	return nil
}

// DeleteImportItem 清空数据
func DeleteImportItem(db string, r Request) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ImportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"job_id": r.JobID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteImportItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	rs, err := c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("DeleteImportItem", err.Error())
		return err
	}

	log.Debugf("delete result %v", rs)

	return nil
}
