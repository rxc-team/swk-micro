package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

type (
	// Value 字段的值
	CopyValue struct {
		DataType string      `json:"data_type,omitempty" bson:"data_type"`
		Value    interface{} `json:"value,omitempty" bson:"value"`
	}
)

// CopyItems 复制台账数据
func CopyItems(db string, fromApp, toApp, fromDs, toDs string, withFile bool) (err error) {
	client := database.New()
	// 设置读写分离的级别
	opts := options.Collection()
	opts.SetReadPreference(readpref.Primary())

	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(fromDs), opts)
	ct := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(toDs), opts)
	cf := client.Database(database.GetDBName(db)).Collection("fields", opts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       fromApp,
		"datastore_id": fromDs,
	}

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	project := bson.M{
		"_id":          1,
		"item_id":      1,
		"app_id":       toApp,
		"datastore_id": toDs,
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

	if !withFile {
		cur, err := cf.Find(ctx, query)
		if err != nil {
			utils.ErrorLog("CopyItems", err.Error())
			return err
		}
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var f *Field
			err := cur.Decode(&f)
			if err != nil {
				utils.ErrorLog("CopyItems", err.Error())
				return err
			}
			if f.FieldType == "file" {
				project["items."+f.FieldID+".data_type"] = "file"
				project["items."+f.FieldID+".value"] = "[]"
				continue
			}
			project["items."+f.FieldID] = 1

		}
		pipe = append(pipe, bson.M{
			"$project": project,
		})

		pipe = append(pipe, bson.M{
			"$out": GetItemCollectionName(toDs),
		})

		opt := options.Aggregate()
		opt.SetAllowDiskUse(true)

		_, err = c.Aggregate(ctx, pipe, opt)
		if err != nil {
			utils.ErrorLog("CopyItems", err.Error())
			return err
		}
		// 只复制不含文件字段的数据
		return nil
	}

	project["items"] = 1
	pipe = append(pipe, bson.M{
		"$project": project,
	})

	pipe = append(pipe, bson.M{
		"$out": GetItemCollectionName(toDs),
	})

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	_, err = c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("CopyItems", err.Error())
		return err
	}

	// 更新复制后的数据
	newQuery := bson.M{
		"app_id":       toApp,
		"datastore_id": toDs,
	}
	cur, err := ct.Find(ctx, newQuery)
	if err != nil {
		utils.ErrorLog("CopyItems", err.Error())
		return err
	}
	defer cur.Close(ctx)

	var current int64 = 1
	var cxModels []mongo.WriteModel
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var dt *Item
		err := cur.Decode(&dt)
		if err != nil {
			utils.ErrorLog("CopyItems", err.Error())
			return err
		}

		items := bson.M{}
		for key, data := range dt.ItemMap {
			var files []File
			newFiles := make([]File, 0)
			if data.DataType == "file" {
				err := json.Unmarshal([]byte(data.Value.(string)), &files)
				if err != nil {
					utils.ErrorLog("CopyItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, toApp, toDs, dt.ItemID, err))
					return err
				}
				for _, file := range files {

					tmpUrl := strings.Replace(file.URL, "app_"+fromApp, "app_"+toApp, 1)
					newUrl := strings.Replace(tmpUrl, "datastore_"+fromDs, "datastore_"+toDs, 1)

					// 编辑新的数据
					newFiles = append(newFiles, File{
						Name: file.Name,
						URL:  newUrl,
					})
				}
				value, err := json.Marshal(newFiles)
				if err != nil {
					utils.ErrorLog("CopyItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, toApp, toDs, dt.ItemID, err))
					return err
				}
				items["items."+key+".value"] = string(value)
			} else {
				continue
			}

		}
		// 是否更新判断
		if len(items) == 0 {
			// 没有文件字段数据,无需更新,跳入下一条数据
			continue
		} else {
			// 有文件字段数据,需要更新,更新对象加1
			current++
		}

		update := bson.M{
			"$set": items,
		}
		objectID, err := primitive.ObjectIDFromHex(dt.ItemID)
		if err != nil {
			utils.ErrorLog("CopyItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, toApp, toDs, dt.ItemID, err))
			return err
		}

		upCxModel := mongo.NewUpdateOneModel()
		upCxModel.SetFilter(bson.M{
			"_id": objectID,
		})
		upCxModel.SetUpdate(update)
		upCxModel.SetUpsert(false)
		cxModels = append(cxModels, upCxModel)

		if current%2000 == 0 {
			result, err := ct.BulkWrite(ctx, cxModels)
			if err != nil {
				utils.ErrorLog("CopyItems", fmt.Sprintf("customer:%s app:%s datastore: %s error: %v ", db, toApp, toDs, err))
				return err
			}
			utils.InfoLog("CopyItems", fmt.Sprintf("customer:%s app:%s datastore: %s update: %d ", db, toApp, toDs, result.ModifiedCount))
			cxModels = cxModels[:0]
		}

	}
	if len(cxModels) > 0 {
		result, err := ct.BulkWrite(ctx, cxModels)
		if err != nil {
			utils.ErrorLog("CopyItems", fmt.Sprintf("customer:%s app:%s datastore: %s error: %v ", db, toApp, toDs, err))
			return err
		}
		utils.InfoLog("CopyItems", fmt.Sprintf("customer:%s app:%s datastore: %s  update: %d ", db, toApp, toDs, result.ModifiedCount))
	}

	return nil
}

// BulkAddItems 批量添加数据
func BulkAddItems(db, datastoreID string, items []mongo.WriteModel) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if len(items) > 0 {
		result, err := c.BulkWrite(ctx, items)
		if err != nil {
			utils.ErrorLog("MutilInventoryItem", err.Error())
			return err
		}

		// TODO 回滚不可用
		if int64(len(items)) != result.ModifiedCount {
			log.Warnf("error MutilInventoryItem with image %v", "not completely update!")
		}
	}

	return nil
}
