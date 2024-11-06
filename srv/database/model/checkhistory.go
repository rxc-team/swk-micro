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

	"rxcsoft.cn/pit3/srv/database/proto/check"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// CheckHistoryCollection histories collection
	CheckHistoryCollection = "check_histories"
)

type (
	// CheckHistory 台账的数据
	CheckHistory struct {
		ID             primitive.ObjectID `json:"id" bson:"_id"`
		CheckId        string             `json:"check_id" bson:"check_id"`
		ItemId         string             `json:"item_id" bson:"item_id"`
		DatastoreId    string             `json:"datastore_id" bson:"datastore_id"`
		ItemMap        ItemMap            `json:"items" bson:"items"`
		CheckType      string             `json:"check_type" bson:"check_type"`
		CheckStartDate string             `json:"check_start_date" bson:"check_start_date"`
		CheckedAt      time.Time          `json:"checked_at" bson:"checked_at"`
		CheckedBy      string             `json:"checked_by" bson:"checked_by"`
	}

	// CheckSearchParam 分页查询多条记录
	CheckSearchParam struct {
		DatastoreId    string
		ItemId         string
		CheckType      string
		CheckStartDate string
		CheckedAtFrom  string
		CheckedAtTo    string
		CheckedBy      string
		DisplayFields  []string
		PageIndex      int64
		PageSize       int64
	}
)

// ToProto 转换为proto数据
func (h *CheckHistory) ToProto() *check.CheckHistory {
	items := make(map[string]*check.Value, len(h.ItemMap))
	for key, it := range h.ItemMap {
		dataType := it.DataType
		item := &check.Value{
			DataType: dataType,
			Value:    GetValueFromModel(it),
		}
		if dataType == "date" && item.Value == "0001-01-01" {
			item.Value = ""
		}
		items[key] = item
	}
	return &check.CheckHistory{
		CheckId:        h.CheckId,
		ItemId:         h.ItemId,
		DatastoreId:    h.DatastoreId,
		Items:          items,
		CheckType:      h.CheckType,
		CheckStartDate: h.CheckStartDate,
		CheckedAt:      h.CheckedAt.String(),
		CheckedBy:      h.CheckedBy,
	}
}

// FindCheckHistories 通过appID获取Datastore信息
func FindCheckHistories(db string, params CheckSearchParam) (items []*CheckHistory, total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(CheckHistoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"datastore_id": params.DatastoreId,
	}

	// ItemID不为空的场合，添加到查询条件中
	if params.ItemId != "" {
		query["item_id"] = params.ItemId
	}
	// 履历类型不为空的场合，添加到查询条件中
	if params.CheckType != "" {
		query["check_type"] = params.CheckType
	}

	// 盘点开始日
	if params.CheckStartDate != "" {
		query["check_start_date"] = params.CheckStartDate
	}

	// 盘点人
	if params.CheckedBy != "" {
		query["checked_by"] = params.CheckedBy
	}

	// 时间不为空的场合，添加到查询条件中
	if params.CheckedAtFrom != "" && params.CheckedAtTo != "" {

		from, _ := time.Parse(DateFormat, params.CheckedAtFrom)
		to, _ := time.Parse(DateFormat, params.CheckedAtTo)

		query["$and"] = []bson.M{
			{
				"checked_at": bson.M{
					"$gte": from,
				},
			},
			{
				"checked_at": bson.M{
					"$lt": to,
				},
			},
		}
	}

	var result []*CheckHistory
	// 取总件数
	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindCheckHistories", err.Error())
		return result, 0, err
	}

	ds, err := getDatastore(db, params.DatastoreId)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return result, 0, err
	}

	fields, err := getFields(db, params.DatastoreId)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return result, 0, err
	}

	skip := (params.PageIndex - 1) * params.PageSize
	limit := params.PageSize

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	pipe = append(pipe, bson.M{
		"$sort": bson.M{
			"checked_at": -1,
		},
	})

	if skip > 0 {
		pipe = append(pipe, bson.M{
			"$skip": skip,
		})
	}
	if limit > 0 {
		pipe = append(pipe, bson.M{
			"$limit": limit,
		})
	}

	project := bson.M{
		"_id":              0,
		"check_id":         1,
		"item_id":          1,
		"datastore_id":     1,
		"check_type":       1,
		"check_start_date": 1,
		"checked_at":       1,
		"checked_by":       1,
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

	LP:
		for _, df := range params.DisplayFields {
			if df == f.FieldID {
				if f.FieldType == "user" {
					pp := []bson.M{
						{
							"$match": bson.M{
								"$expr": bson.M{
									"$and": []bson.M{
										{
											"$in": []string{"$user_id", "$$user"},
										},
									},
								},
							},
						},
					}

					lookup := bson.M{
						"from": "users",
						"let": bson.M{
							"user": "$items." + f.FieldID + ".value",
						},
						"pipeline": pp,
						"as":       "relations_" + f.FieldID,
					}

					pipe = append(pipe, bson.M{
						"$lookup": lookup,
					})

					project["items."+f.FieldID+".value"] = "$relations_" + f.FieldID + ".user_name"
					project["items."+f.FieldID+".data_type"] = "user"

					break LP
				}
				if f.FieldType == "options" {
					pp := []bson.M{
						{
							"$match": bson.M{
								"$expr": bson.M{
									"$and": []bson.M{
										{
											"$eq": []string{"$app_id", "$$app_id"},
										},
										{
											"$eq": []string{"$option_id", "$$option_id"},
										},
										{
											"$eq": []string{"$option_value", "$$option_value"},
										},
									},
								},
							},
						},
					}

					lookup := bson.M{
						"from": "options",
						"let": bson.M{
							"app_id":       f.AppID,
							"option_id":    f.OptionID,
							"option_value": "$items." + f.FieldID + ".value",
						},
						"pipeline": pp,
						"as":       "relations_" + f.FieldID,
					}

					unwind := bson.M{
						"path":                       "$relations_" + f.FieldID,
						"preserveNullAndEmptyArrays": true,
					}

					pipe = append(pipe, bson.M{
						"$lookup": lookup,
					})
					pipe = append(pipe, bson.M{
						"$unwind": unwind,
					})

					project["items."+f.FieldID+".value"] = "$relations_" + f.FieldID + ".option_label"
					project["items."+f.FieldID+".data_type"] = "options"

					break LP
				}

				// 函数字段，重新拼接
				if f.FieldType == "function" {
					var formula bson.M
					err := json.Unmarshal([]byte(f.Formula), &formula)
					if err != nil {
						utils.ErrorLog("FindItems", err.Error())
						return result, 0, err
					}

					if len(formula) > 0 {
						project["items."+f.FieldID+".value"] = formula
					} else {
						project["items."+f.FieldID+".value"] = ""
					}

					// 当前数据本身
					project["items."+f.FieldID+".data_type"] = f.ReturnType
					break LP
				}

				project["items."+f.FieldID] = "$items." + f.FieldID
				break LP
			}
		}

	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return result, 0, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item *CheckHistory
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindItems", err.Error())
			return result, 0, err
		}
		result = append(result, item)
	}

	return result, t, nil
}

// DownloadCheckHistories 通过appID获取Datastore信息
func DownloadCheckHistories(db string, params CheckSearchParam, stream check.CheckHistoryService_DownloadStream) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(CheckHistoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	query := bson.M{
		"datastore_id": params.DatastoreId,
	}

	// ItemID不为空的场合，添加到查询条件中
	if params.ItemId != "" {
		query["item_id"] = params.ItemId
	}
	// 履历类型不为空的场合，添加到查询条件中
	if params.CheckType != "" {
		query["check_type"] = params.CheckType
	}

	// 盘点开始日
	if params.CheckStartDate != "" {
		query["check_start_date"] = params.CheckStartDate
	}

	// 盘点人
	if params.CheckedBy != "" {
		query["checked_by"] = params.CheckedBy
	}

	// 时间不为空的场合，添加到查询条件中
	if params.CheckedAtFrom != "" && params.CheckedAtTo != "" {

		from, _ := time.Parse(DateFormat, params.CheckedAtFrom)
		to, _ := time.Parse(DateFormat, params.CheckedAtTo)

		query["$and"] = []bson.M{
			{
				"checked_at": bson.M{
					"$gte": from,
				},
			},
			{
				"checked_at": bson.M{
					"$lt": to,
				},
			},
		}
	}

	// 聚合查询
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DownloadHistories", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "checked_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)

	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("DownloadHistories", err.Error())
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var his CheckHistory
		err := cur.Decode(&his)
		if err != nil {
			utils.ErrorLog("DownloadHistories", err.Error())
			return err
		}

		if err := stream.Send(&check.DownloadResponse{History: his.ToProto()}); err != nil {
			utils.ErrorLog("DownloadHistories", err.Error())
			return err
		}
	}

	return nil
}

// FindCheckHistoryCount 获取履历件数
func FindCheckHistoryCount(db string, params CheckSearchParam) (total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(CheckHistoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"datastore_id": params.DatastoreId,
	}

	// ItemID不为空的场合，添加到查询条件中
	if params.ItemId != "" {
		query["item_id"] = params.ItemId
	}
	// 履历类型不为空的场合，添加到查询条件中
	if params.CheckType != "" {
		query["check_type"] = params.CheckType
	}

	// 盘点开始日
	if params.CheckStartDate != "" {
		query["check_start_date"] = params.CheckStartDate
	}

	// 盘点人
	if params.CheckedBy != "" {
		query["checked_by"] = params.CheckedBy
	}

	// 时间不为空的场合，添加到查询条件中
	if params.CheckedAtFrom != "" && params.CheckedAtTo != "" {

		from, _ := time.Parse(DateFormat, params.CheckedAtFrom)
		to, _ := time.Parse(DateFormat, params.CheckedAtTo)

		query["$and"] = []bson.M{
			{
				"checked_at": bson.M{
					"$gte": from,
				},
			},
			{
				"checked_at": bson.M{
					"$lt": to,
				},
			},
		}
	}

	// 取总件数
	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindCheckHistoryCount", err.Error())
		return 0, err
	}

	return t, nil
}

// DeleteCheckHistories 物理删除选中的数据
func DeleteCheckHistories(db string, hsList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(CheckHistoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteCheckHistories", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteCheckHistories", err.Error())
		return err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, hsID := range hsList {
			objectID, err := primitive.ObjectIDFromHex(hsID)
			if err != nil {
				utils.ErrorLog("DeleteCheckHistories", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteCheckHistories", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("DeleteCheckHistories", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteCheckHistories", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteCheckHistories", err.Error())
		return err
	}

	session.EndSession(ctx)
	return nil
}
