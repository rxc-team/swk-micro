package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/goinggo/mapstructure"
	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

type ItemMap map[string]*Value

type (
	// Item 台账的数据
	Item struct {
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
		LabelTime   time.Time          `json:"label_time" bson:"label_time"`
		Status      string             `json:"status" bson:"status"`
	}

	// ResultItem 台账的数据
	ResultItem struct {
		Docs  []*Item `json:"docs" bson:"docs"`
		Total int64   `json:"total" bson:"total"`
	}

	// Value 字段的值
	Value struct {
		DataType string      `json:"data_type,omitempty" bson:"data_type"`
		Value    interface{} `json:"value,omitempty" bson:"value"`
	}

	// ItemParam 查询单条记录
	ItemParam struct {
		IsOrigin    bool
		ItemID      string
		DatastoreID string
		Owners      []string
	}

	// RishiritsuParam 查询单条记录
	RishiritsuParam struct {
		DatastoreID string
		LeaseStymd  string
		LeaseKikan  string
	}

	// ItemsParam 分页查询多条记录
	ItemsParam struct {
		AppID         string
		DatastoreID   string
		ConditionList []*Condition
		ConditionType string
		PageIndex     int64
		PageSize      int64
		Sorts         []*SortItem
		SortValue     string
		IsOrigin      bool
		ShowLookup    bool
		Owners        []string
	}
	// DeleteItemsParam 删除多条数据记录
	DeleteItemsParam struct {
		AppID         string
		DatastoreID   string
		ConditionList []*Condition
		ConditionType string
		UserID        string
	}
	// CountParam 查询件数
	CountParam struct {
		AppID         string
		DatastoreID   string
		ConditionList []*Condition
		ConditionType string
		Owners        []string
	}
	// KaraCountParam 查询空值件数
	KaraCountParam struct {
		AppID       string
		DatastoreID string
		FieldID     string
		FieldType   string
		Owners      []string
	}

	// OwnersParam 变更查询的多条记录的所有者
	OwnersParam struct {
		AppID         string
		DatastoreID   string
		ConditionList []*Condition
		ConditionType string
		Owner         string
		Writer        string
		OldOwners     []string
	}
	// OwnerParam 变更查询的单条记录的所有者
	OwnerParam struct {
		AppID       string
		DatastoreID string
		ItemID      string
		Owner       string
		Writer      string
	}
	// ItemUpdateParam 单条台账记录更新传入的参数
	ItemUpdateParam struct {
		ItemID      string
		AppID       string
		DatastoreID string
		ItemMap     map[string]*Value
		UpdatedAt   time.Time
		UpdatedBy   string
		Owners      []string
		Lang        string
		Domain      string
	}

	// ChangeData 导入的数据
	ChangeData struct {
		Query  map[string]*item.Value
		Change map[string]*Value
		Index  int64
		ItemId string
	}

	// ImportCheckParam 导入盘点的参数
	ImportCheckParam struct {
		AppID        string
		DatastoreID  string
		ImportDatas  []*ChangeData
		Writer       string
		Owners       []string
		UpdateOwners []string
	}
)

// File 文件类型
type File struct {
	URL  string `json:"url" bson:"url"`
	Name string `json:"name" bson:"name"`
}

// TmpParam 从临时数据中读取数据，并插入对应台账中
type TmpParam struct {
	DB            string
	TemplateID    string
	APIKey        string
	UserID        string
	Datastores    []Datastore
	Owners        []string
	Keiyakuno     string
	Leasekaishacd string
	Bunruicd      string
	Segmentcd     string
	FileMap       map[string][]Field
}

// AttachParam 上传附件数据参数
type AttachParam struct {
	DB      string
	Items   []*Item
	DsMap   map[string]string
	FileMap map[string][]Field
	Owners  []string
}

// 定义相关的结构体
type FieldRule struct {
	DownloadName    string            `json:"download_name" bson:"download_name"`
	FieldId         string            `json:"field_id" bson:"field_id"`
	FieldConditions []*FieldCondition `json:"field_conditions" bson:"field_conditions"`
	SettingMethod   string            `json:"setting_method" bson:"setting_method"`
	FieldType       string            `json:"field_type" bson:"field_type"`
	DatastoreId     string            `json:"datastore_id" bson:"datastore_id"`
	Format          string            `json:"format" bson:"format"`
}

type FieldCondition struct {
	ConditionID   string        `json:"condition_id" bson:"condition_id"`
	ConditionName string        `json:"condition_name" bson:"condition_name"`
	FieldGroups   []*FieldGroup `json:"field_groups" bson:"field_groups"`
	ThenValue     string        `json:"then_value" bson:"then_value"`
	ElseValue     string        `json:"else_value" bson:"else_value"`
}

type FieldGroup struct {
	GroupID    string      `json:"group_id" bson:"group_id"`
	GroupName  string      `json:"group_name" bson:"group_name"`
	Type       string      `json:"type" bson:"type"`               // and or
	SwitchType string      `json:"switch_type" bson:"switch_type"` // and or
	FieldCons  []*FieldCon `json:"field_cons" bson:"field_cons"`
}

type FieldCon struct {
	ConField    string `json:"con_field" bson:"con_field"`
	ConOperator string `json:"con_operator" bson:"con_operator"`
	ConValue    string `json:"con_value" bson:"con_value"`
}

// ToProto 转换为proto数据
func (i *Item) ToProto() *item.Item {
	items := make(map[string]*item.Value, len(i.ItemMap))
	for key, it := range i.ItemMap {
		if key == "sakuseidate" || key == "kakuteidate" {
			defaultTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
			if defaultTime.Format("2006-01-02") != it.Value.(primitive.DateTime).Time().Format("2006-01-02") {
				items[key] = &item.Value{
					DataType: "date",
					Value:    it.Value.(primitive.DateTime).Time().Format("2006-01-02 15:04:05"),
				}
			} else {
				items[key] = &item.Value{
					DataType: "date",
					Value:    "",
				}
			}
			continue
		}
		dataType := it.DataType
		items[key] = &item.Value{
			DataType: dataType,
			Value:    GetValueFromModel(it),
		}
	}
	return &item.Item{
		ItemId:      i.ItemID,
		AppId:       i.AppID,
		DatastoreId: i.DatastoreID,
		Items:       items,
		Owners:      i.Owners,
		CheckType:   i.CheckType,
		CheckStatus: i.CheckStatus,
		CreatedAt:   i.CreatedAt.String(),
		CreatedBy:   i.CreatedBy,
		CheckedAt:   i.CheckedAt.String(),
		CheckedBy:   i.CheckedBy,
		UpdatedAt:   i.UpdatedAt.String(),
		UpdatedBy:   i.UpdatedBy,
		LabelTime:   i.LabelTime.String(),
		Status:      i.Status,
	}
}

// FindItems 通过appID获取Datastore信息
func FindItems(db string, params ItemsParam) (*ResultItem, error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(params.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       params.AppID,
		"datastore_id": params.DatastoreID,
	}

	sortItem, err := searchBefore(ctx, c, db, query, params)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return nil, err
	}

	if params.IsOrigin {
		return getItemsWithoutLookup(ctx, c, db, query, sortItem, params)
	}

	return getItemsWithLookup(ctx, c, db, query, sortItem, params)
}

// DownloadItems 下载台账数据
func DownloadItems(db string, params ItemsParam, stream item.ItemService_DownloadStream) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(params.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       params.AppID,
		"datastore_id": params.DatastoreID,
	}

	sortItem, err := searchBefore(ctx, c, db, query, params)
	if err != nil {
		utils.ErrorLog("DownloadItems", err.Error())
		return err
	}

	ds, err := getDatastore(db, params.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return err
	}

	fields, err := getFields(db, params.DatastoreID)
	if err != nil {
		utils.ErrorLog("DownloadItems", err.Error())
		return err
	}

	skip := (params.PageIndex - 1) * params.PageSize
	limit := params.PageSize

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	pipe = append(pipe, bson.M{
		"$sort": sortItem,
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

			continue
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

			continue
		}

		// 函数字段，重新拼接
		if f.FieldType == "function" {
			var formula bson.M
			err := json.Unmarshal([]byte(f.Formula), &formula)
			if err != nil {
				utils.ErrorLog("DownloadItems", err.Error())
				return err
			}

			if len(formula) > 0 {
				project["items."+f.FieldID+".value"] = formula
			} else {
				project["items."+f.FieldID+".value"] = ""
			}

			// 当前数据本身
			project["items."+f.FieldID+".data_type"] = f.ReturnType
			continue
		}

		project["items."+f.FieldID] = "$items." + f.FieldID
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)
	opt.SetBatchSize(int32(params.PageSize))

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("DownloadItems", err.Error())
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var it Item
		err := cur.Decode(&it)
		if err != nil {
			utils.ErrorLog("DownloadItems", err.Error())
			return err
		}

		if err := stream.Send(&item.DownloadResponse{Item: it.ToProto()}); err != nil {
			utils.ErrorLog("DownloadItems", err.Error())
			return err
		}
	}

	return nil
}

// FindAndModifyFile 通过appID获取Datastore信息
func FindAndModifyFile(db string, param ItemsParam, stream item.ItemService_FindAndModifyFileStream) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(param.DatastoreID))
	ctx := context.Background()

	// 如果没有文件字段，直接返回
	if !hasFileField(db, param.AppID, param.DatastoreID) {
		return nil
	}

	query := bson.M{
		"app_id":       param.AppID,
		"datastore_id": param.DatastoreID,
	}

	cur, err := c.Find(ctx, query)
	if err != nil {
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
			utils.ErrorLog("FindItems", err.Error())
			return err
		}

		items := bson.M{}
		for key, data := range dt.ItemMap {
			var files []File
			newFiles := make([]File, 0)
			if data.DataType == "file" {
				err := json.Unmarshal([]byte(data.Value.(string)), &files)
				if err != nil {
					utils.ErrorLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, param.AppID, param.DatastoreID, dt.ItemID, err))
					return err
				}
				for _, file := range files {

					if strings.Contains(file.URL, "/datastore_"+param.DatastoreID+"/") {
						// 编辑新的数据
						newFiles = append(newFiles, File{
							Name: file.Name,
							URL:  file.URL,
						})
						continue
					}

					// 生成新的数据的路径
					urlList := strings.Split(file.URL, "/")
					prefixUrl := strings.Join(urlList[0:4], "/")
					suffixUrl := strings.Join(urlList[6:], "/")
					newUrl := path.Join(prefixUrl, "/app_"+param.AppID+"/data", "/datastore_"+param.DatastoreID+"/", suffixUrl)

					// 编辑新的数据
					newFiles = append(newFiles, File{
						Name: file.Name,
						URL:  newUrl,
					})

					// 生成copy的路径
					old := strings.Join(urlList[3:], "/")
					new := path.Join("public", "/app_"+param.AppID+"/data", "/datastore_"+param.DatastoreID+"/", suffixUrl)

					// 发送变更前后的文件路径
					err := stream.Send(&item.FindResponse{
						OldUrl: old,
						NewUrl: new,
					})

					if err != nil {
						utils.ErrorLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, param.AppID, param.DatastoreID, dt.ItemID, err))
						return err
					}
				}
				value, err := json.Marshal(newFiles)
				if err != nil {
					utils.ErrorLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, param.AppID, param.DatastoreID, dt.ItemID, err))
					return err
				}
				items["items."+key+".value"] = string(value)
			}
		}

		// 如果没有台账数据，将不进行更新处理
		if len(items) > 0 {
			current++
			update := bson.M{
				"$set": items,
			}
			objectID, err := primitive.ObjectIDFromHex(dt.ItemID)
			if err != nil {
				utils.ErrorLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, param.AppID, param.DatastoreID, dt.ItemID, err))
				return err
			}

			upCxModel := mongo.NewUpdateOneModel()
			upCxModel.SetFilter(bson.M{
				"_id": objectID,
			})
			upCxModel.SetUpdate(update)
			upCxModel.SetUpsert(false)
			cxModels = append(cxModels, upCxModel)
		}

		if current%2000 == 0 {
			result, err := c.BulkWrite(ctx, cxModels)
			if err != nil {
				utils.ErrorLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s error: %v ", db, param.AppID, param.DatastoreID, err))
				return err
			}
			utils.InfoLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s update: %d ", db, param.AppID, param.DatastoreID, result.ModifiedCount))
			cxModels = cxModels[:0]
		}
	}

	if len(cxModels) > 0 {
		result, err := c.BulkWrite(ctx, cxModels)
		if err != nil {
			utils.ErrorLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s error: %v ", db, param.AppID, param.DatastoreID, err))
			return err
		}
		utils.InfoLog("ModifyFileItems", fmt.Sprintf("customer:%s app:%s datastore: %s  update: %d ", db, param.AppID, param.DatastoreID, result.ModifiedCount))
	}

	return nil
}

// hasFileField 判断台账字段中是否有文件字段
func hasFileField(db, appId, datastoreId string) bool {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FieldsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       appId,
		"datastore_id": datastoreId,
		"field_type":   "file",
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("hasFileField", fmt.Sprintf("query: [ %s ]", queryJSON))

	count, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("hasFileField", err.Error())
		return false
	}

	return count > 0
}

// 查询之前先创建索引
func searchBefore(ctx context.Context, c *mongo.Collection, db string, query bson.M, params ItemsParam) (bson.D, error) {
	indexKeys := []string{"app_id", "datastore_id"}
	indexMap := bson.M{
		"app_id":       1,
		"datastore_id": 1,
	}

	if len(params.Owners) > 0 {
		query["owners"] = bson.M{
			"$in": params.Owners,
		}
		indexMap["owners"] = 1
		indexKeys = append(indexKeys, "owners")
	}

	// 编辑 match 检索条件
	buildMatchAndSort(params.ConditionList, params.ConditionType, query, &indexKeys, indexMap)

	// 若未指定排序则采用台账默认排序
	if len(params.Sorts) == 0 {
		ds, err := FindDatastore(db, params.DatastoreID)
		if err != nil {
			utils.ErrorLog("searchBefore", err.Error())
			return nil, err
		}
		params.Sorts = ds.Sorts
	}

	// 排序
	sortItem := bson.D{}
	for _, sort := range params.Sorts {
		sortKey := "items." + sort.SortKey + ".value"
		if sort.SortValue == "ascend" {
			sortItem = append(sortItem, bson.E{Key: sortKey, Value: 1})
			indexMap[sortKey] = 1
			indexKeys = append(indexKeys, sortKey)
		} else {
			sortItem = append(sortItem, bson.E{Key: sortKey, Value: -1})
			indexMap[sortKey] = -1
			indexKeys = append(indexKeys, sortKey)
		}
	}

	sortItem = append(sortItem, bson.E{Key: "created_at", Value: -1})
	indexMap["created_at"] = -1
	indexKeys = append(indexKeys, "created_at")

	var indexs bson.D
	existMap := make(map[string]struct{})
	for _, key := range indexKeys {

		if _, exist := existMap[key]; !exist {
			existMap[key] = struct{}{}
			indexs = append(indexs, bson.E{
				Key: key, Value: indexMap[key],
			})
		}
	}

	index := mongo.IndexModel{
		Keys:    indexs,
		Options: options.Index().SetSparse(true).SetUnique(false),
	}

	// 删除多余的索引
	if err := DeleteLeastUsed(c); err != nil {
		utils.ErrorLog("searchBefore", err.Error())
		return nil, err
	}

	if len(indexs) < 31 {
		indexOpts := options.CreateIndexes().SetMaxTime(60 * time.Second)
		if _, err := c.Indexes().CreateOne(ctx, index, indexOpts); err != nil {
			utils.ErrorLog("searchBefore", err.Error())
			return nil, err
		}
	}

	return sortItem, nil
}

// getItemsWithLookup 使用关联查询
func getItemsWithLookup(ctx context.Context, c *mongo.Collection, db string, query bson.M, sortItem bson.D, params ItemsParam) (items *ResultItem, err error) {

	var result ResultItem

	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return nil, err
	}

	ds, err := getDatastore(db, params.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return nil, err
	}

	fields, err := getFields(db, params.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return nil, err
	}

	skip := (params.PageIndex - 1) * params.PageSize
	limit := params.PageSize

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	pipe = append(pipe, bson.M{
		"$sort": sortItem,
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
		"_id":          0,
		"item_id":      1,
		"app_id":       1,
		"datastore_id": 1,
		// "owners":       1,
		"check_type":   1,
		"check_status": 1,
		"created_at":   1,
		// "created_by":   1,
		"updated_at": 1,
		// "updated_by":   1,
		"checked_at": 1,
		// "checked_by":   1,
		"label_time": 1,
		"status":     1,
	}

	// 所有者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$in": []string{"$access_key", "$$owners"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "groups",
			"let": bson.M{
				"owners": "$owners",
			},
			"pipeline": pp,
			"as":       "relations_owners",
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})

		project["owners"] = "$relations_owners.group_name"
	}
	// 创建者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$eq": []string{"$user_id", "$$user"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "users",
			"let": bson.M{
				"user": "$created_by",
			},
			"pipeline": pp,
			"as":       "relations_created_by",
		}
		unwind := bson.M{
			"path":                       "$relations_created_by",
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})

		project["created_by"] = "$relations_created_by.user_name"
	}
	// 更新者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$eq": []string{"$user_id", "$$user"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "users",
			"let": bson.M{
				"user": "$updated_by",
			},
			"pipeline": pp,
			"as":       "relations_updated_by",
		}

		unwind := bson.M{
			"path":                       "$relations_updated_by",
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})

		project["updated_by"] = "$relations_updated_by.user_name"
	}
	// 盘点者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$eq": []string{"$user_id", "$$user"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "users",
			"let": bson.M{
				"user": "$checked_by",
			},
			"pipeline": pp,
			"as":       "relations_checked_by",
		}

		unwind := bson.M{
			"path":                       "$relations_checked_by",
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})

		project["checked_by"] = "$relations_checked_by.user_name"
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

			continue
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

			continue
		}

		// 函数字段，重新拼接
		if f.FieldType == "function" {
			var formula bson.M
			err := json.Unmarshal([]byte(f.Formula), &formula)
			if err != nil {
				utils.ErrorLog("FindItems", err.Error())
				return nil, err
			}

			if len(formula) > 0 {
				project["items."+f.FieldID+".value"] = formula
			} else {
				project["items."+f.FieldID+".value"] = ""
			}

			// 当前数据本身
			project["items."+f.FieldID+".data_type"] = f.ReturnType
			continue
		}

		project["items."+f.FieldID] = "$items." + f.FieldID
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)
	opt.SetBatchSize(int32(params.PageSize))

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item *Item
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindItems", err.Error())
			return nil, err
		}
		result.Docs = append(result.Docs, item)
	}

	result.Total = t

	return &result, nil
}

// getItemsWithoutLookup 不使用关联查询，只获取原始数据
func getItemsWithoutLookup(ctx context.Context, c *mongo.Collection, db string, query bson.M, sortItem bson.D, params ItemsParam) (items *ResultItem, err error) {

	var result ResultItem

	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return nil, err
	}

	skip := (params.PageIndex - 1) * params.PageSize
	limit := params.PageSize

	opt := options.Find()
	opt.SetSort(sortItem)
	if skip > 0 {
		opt.SetSkip(skip)
	}
	if limit > 0 {
		opt.SetLimit(limit)
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	cur, err := c.Find(ctx, query, opt)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item *Item
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindItems", err.Error())
			return nil, err
		}
		result.Docs = append(result.Docs, item)
	}

	result.Total = t

	return &result, nil
}

func buildMatch(conditionList []*Condition, conditionType string, query bson.M) {

	if len(conditionList) > 0 {
		if conditionType == "and" {
			// OR的场合
			and := []bson.M{}
			for _, condition := range conditionList {
				if condition.IsDynamic {

					switch condition.FieldType {
					case "text", "textarea", "autonum":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							})
						} else if condition.Operator == "<>" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							})
						} else {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							})
						}
					case "switch":
						// 只能是等于
						and = append(and, bson.M{
							"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
						})
					case "file":
						if condition.SearchValue == "true" {
							// 存在
							and = append(and, bson.M{
								"$and": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": true},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$ne": "[]"},
									},
								},
							})
						} else {
							// 不存在
							and = append(and, bson.M{
								"$or": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": false},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$eq": "[]"},
									},
								},
							})
						}
					case "lookup":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							})
						} else if condition.Operator == "<>" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							})
						} else {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							})
						}
					case "options":
						// 只能是in和=两种情况（默认是等于）
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$in": values},
								})
							}
						} else if condition.Operator == "<>" {
							value := condition.SearchValue
							// 不等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": value,
								},
							})
						} else {
							value := condition.SearchValue
							// 等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": value,
							})
						}
					case "user":
						// 只能是in和=两种情况（默认是等于）
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$in": values},
								})
							}
						} else if condition.Operator == "<>" {
							value := condition.SearchValue
							// 不等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": value,
								},
							})
						} else {
							value := condition.SearchValue
							// 等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": value,
							})
						}
					case "number", "date", "time":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])},
							})
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])},
							})
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gt": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == "<" {
								// 小于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lte": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == "<>" {
								// 不等于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$ne": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else {
								// 默认的[等于]的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
								})
							}
						}
					default:
					}
				} else {
					switch condition.FieldType {
					case "options", "type", "user":
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$in": values},
								})
							}
						} else if condition.Operator == "<>" {
							value := condition.SearchValue
							// 不等于
							and = append(and, bson.M{
								condition.FieldID: bson.M{
									"$ne": value,
								},
							})
						} else {
							value := condition.SearchValue
							// 等于
							and = append(and, bson.M{
								condition.FieldID: value,
							})
						}
					case "check":
						value := condition.SearchValue
						// 等于
						and = append(and, bson.M{
							condition.FieldID: value,
						})
					case "datetime":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							and = append(and, bson.M{
								condition.FieldID: bson.M{"$gte": getSearchValue(condition.FieldType, values[0])},
							})
							and = append(and, bson.M{
								condition.FieldID: bson.M{"$lt": getSearchValue(condition.FieldType, values[1])},
							})
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$gte": fullTime},
								})
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								value := getTime(condition.SearchValue)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$gt": value},
								})
							} else if condition.Operator == "<" {
								// 小于的场合
								value := getTime(condition.SearchValue)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$lt": value},
								})
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$lt": fullTime},
								})
							} else if condition.Operator == "<>" {
								// 不等于的场合
								value := getTime(condition.SearchValue)

								zeroTime := value
								fullTime := value.Add(time.Hour * 24)

								and = append(and, bson.M{
									"$or": []bson.M{
										{condition.FieldID: bson.M{"$gte": fullTime}},
										{condition.FieldID: bson.M{"$lt": zeroTime}},
									},
								})
							} else {
								// 默认的[等于]的场合
								value := getTime(condition.SearchValue)

								zeroTime := value
								fullTime := value.Add(time.Hour * 24)

								and = append(and, bson.M{
									condition.FieldID: bson.M{"$gte": zeroTime},
								})
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$lt": fullTime},
								})
							}
						}
					default:
						break
					}
				}
			}

			if len(and) > 0 {
				query["$and"] = and
			}
		} else {
			// OR的场合
			or := []bson.M{}
			for _, condition := range conditionList {
				if condition.IsDynamic {

					switch condition.FieldType {
					case "text", "textarea", "autonum":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							}
							or = append(or, q)
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							or = append(or, q)
						}
					case "switch":
						// 只能是等于
						q := bson.M{
							"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
						}
						or = append(or, q)
					case "file":
						if condition.SearchValue == "true" {
							// 存在
							q := bson.M{
								"$and": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": true},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$ne": "[]"},
									},
								},
							}
							or = append(or, q)
						} else {
							// 不存在
							q := bson.M{
								"$or": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": false},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$eq": "[]"},
									},
								},
							}
							or = append(or, q)
						}
					case "lookup":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							}
							or = append(or, q)
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							or = append(or, q)
						}
					case "options", "user":
						// 只能是in和=两种情况（默认是等于）
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$in": values},
								}
								or = append(or, q)
							}
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							// 等于
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							or = append(or, q)
						}
					case "number", "date", "time":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							q := bson.M{
								"$and": []bson.M{
									{"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
									{"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
								},
							}
							or = append(or, q)
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gt": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == "<" {
								// 小于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lte": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == "<>" {
								// 不等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$ne": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else {
								// 默认的[等于]的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
								}
								or = append(or, q)
							}
						}
					default:
					}
				} else {
					switch condition.FieldType {
					case "options", "type", "user":
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								q := bson.M{
									condition.FieldID: bson.M{"$in": values},
								}
								or = append(or, q)
							}
						} else if condition.Operator == "<>" {
							q := bson.M{
								condition.FieldID: bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							// 等于
							q := bson.M{
								condition.FieldID: condition.SearchValue,
							}
							or = append(or, q)
						}
					case "check":
						// 等于
						q := bson.M{
							condition.FieldID: condition.SearchValue,
						}
						or = append(or, q)
					case "datetime":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							q := bson.M{
								"$and": []bson.M{
									{condition.FieldID: bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
									{condition.FieldID: bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
								},
							}
							or = append(or, q)
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									condition.FieldID: bson.M{"$gte": fullTime},
								}
								or = append(or, q)
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								value := getTime(condition.SearchValue)
								q := bson.M{
									condition.FieldID: bson.M{"$gte": value},
								}
								or = append(or, q)
							} else if condition.Operator == "<" {
								// 小于的场合
								value := getTime(condition.SearchValue)
								q := bson.M{
									condition.FieldID: bson.M{"$lt": value},
								}
								or = append(or, q)
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									condition.FieldID: bson.M{"$lt": fullTime},
								}
								or = append(or, q)
							} else if condition.Operator == "<>" {
								// 不等于的场合
								value := getTime(condition.SearchValue)
								zeroTime := value
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									"$or": []bson.M{
										{condition.FieldID: bson.M{"$gte": fullTime}},
										{condition.FieldID: bson.M{"$lt": zeroTime}},
									},
								}
								or = append(or, q)
							} else {
								// 默认的[等于]的场合
								value := getTime(condition.SearchValue)
								zeroTime := value
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									"$and": []bson.M{
										{condition.FieldID: bson.M{"$gte": zeroTime}},
										{condition.FieldID: bson.M{"$lt": fullTime}},
									},
								}
								or = append(or, q)
							}
						}
					default:
					}
				}
			}

			if len(or) > 0 {
				query["$or"] = or
			}
		}
	}
}

func buildMatchAndSort(conditionList []*Condition, conditionType string, query bson.M, indexKeys *[]string, indexMap bson.M) {

	if len(conditionList) > 0 {
		if conditionType == "and" {
			// AND的场合
			and := []bson.M{}
			for _, condition := range conditionList {
				if condition.IsDynamic {
					switch condition.FieldType {
					case "text", "textarea", "autonum":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							})
						} else if condition.Operator == "<>" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							})
						} else {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							})
						}
						indexMap["items."+condition.FieldID+".value"] = 1
						*indexKeys = append(*indexKeys, "items."+condition.FieldID+".value")
					case "switch":
						// 只能是等于
						and = append(and, bson.M{
							"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
						})
						indexMap["items."+condition.FieldID+".value"] = 1
						*indexKeys = append(*indexKeys, "items."+condition.FieldID+".value")
					case "file":
						if condition.SearchValue == "true" {
							// 存在
							and = append(and, bson.M{
								"$and": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": true},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$ne": "[]"},
									},
								},
							})
						} else {
							// 不存在
							and = append(and, bson.M{
								"$or": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": false},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$eq": "[]"},
									},
								},
							})
						}
					case "lookup":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							})
						} else if condition.Operator == "<>" {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							})
						} else {
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							})
						}
						indexMap["items."+condition.FieldID+".value"] = 1
						*indexKeys = append(*indexKeys, "items."+condition.FieldID+".value")
					case "options":
						// 只能是in和=两种情况（默认是等于）
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$in": values},
								})
							}
						} else if condition.Operator == "<>" {
							value := condition.SearchValue
							// 不等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": value,
								},
							})
						} else {
							value := condition.SearchValue
							// 等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": value,
							})
						}
						indexMap["items."+condition.FieldID+".value"] = 1
						*indexKeys = append(*indexKeys, "items."+condition.FieldID+".value")
					case "user":
						// 只能是in和=两种情况（默认是等于）
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$in": values},
								})
							}
						} else if condition.Operator == "<>" {
							value := condition.SearchValue
							// 不等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": value,
								},
							})
						} else {
							value := condition.SearchValue
							// 等于
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": value,
							})
						}
					case "number", "date", "time":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])},
							})
							and = append(and, bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])},
							})
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gt": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == "<" {
								// 小于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lte": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else if condition.Operator == "<>" {
								// 不等于的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$ne": getSearchValue(condition.FieldType, condition.SearchValue)},
								})
							} else {
								// 默认的[等于]的场合
								and = append(and, bson.M{
									"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
								})
							}
						}
						indexMap["items."+condition.FieldID+".value"] = 1
						*indexKeys = append(*indexKeys, "items."+condition.FieldID+".value")
					default:
					}
				} else {
					switch condition.FieldType {
					case "options", "type", "user", "group":
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$in": values},
								})
							}
						} else if condition.Operator == "<>" {
							value := condition.SearchValue
							// 不等于
							and = append(and, bson.M{
								condition.FieldID: bson.M{
									"$ne": value,
								},
							})
						} else {
							value := condition.SearchValue
							// 等于
							and = append(and, bson.M{
								condition.FieldID: value,
							})
						}
						indexMap[condition.FieldID] = 1
						*indexKeys = append(*indexKeys, condition.FieldID)
					case "check":
						value := condition.SearchValue
						// 等于
						and = append(and, bson.M{
							condition.FieldID: value,
						})
						indexMap[condition.FieldID] = 1
						*indexKeys = append(*indexKeys, condition.FieldID)
					case "datetime":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							and = append(and, bson.M{
								condition.FieldID: bson.M{"$gte": getSearchValue(condition.FieldType, values[0])},
							})
							and = append(and, bson.M{
								condition.FieldID: bson.M{"$lt": getSearchValue(condition.FieldType, values[1])},
							})
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$gte": fullTime},
								})
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								value := getTime(condition.SearchValue)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$gt": value},
								})
							} else if condition.Operator == "<" {
								// 小于的场合
								value := getTime(condition.SearchValue)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$lt": value},
								})
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$lt": fullTime},
								})
							} else if condition.Operator == "<>" {
								// 不等于的场合
								value := getTime(condition.SearchValue)

								zeroTime := value
								fullTime := value.Add(time.Hour * 24)

								and = append(and, bson.M{
									"$or": []bson.M{
										{condition.FieldID: bson.M{"$gte": fullTime}},
										{condition.FieldID: bson.M{"$lt": zeroTime}},
									},
								})
							} else {
								// 默认的[等于]的场合
								value := getTime(condition.SearchValue)

								zeroTime := value
								fullTime := value.Add(time.Hour * 24)

								and = append(and, bson.M{
									condition.FieldID: bson.M{"$gte": zeroTime},
								})
								and = append(and, bson.M{
									condition.FieldID: bson.M{"$lt": fullTime},
								})
							}
						}
						indexMap[condition.FieldID] = 1
						*indexKeys = append(*indexKeys, condition.FieldID)
					default:
						break
					}
				}
			}

			if len(and) > 0 {
				query["$and"] = and
			}
		} else {
			// OR的场合
			or := []bson.M{}
			for _, condition := range conditionList {
				if condition.IsDynamic {

					switch condition.FieldType {
					case "text", "textarea", "autonum":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							}
							or = append(or, q)
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							or = append(or, q)
						}
					case "switch":
						// 只能是等于
						q := bson.M{
							"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
						}
						or = append(or, q)
					case "file":
						if condition.SearchValue == "true" {
							// 存在
							q := bson.M{
								"$and": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": true},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$ne": "[]"},
									},
								},
							}
							or = append(or, q)
						} else {
							// 不存在
							q := bson.M{
								"$or": []bson.M{
									{
										"items." + condition.FieldID + ".value": bson.M{"$exists": false},
									},
									{
										"items." + condition.FieldID + ".value": bson.M{"$eq": "[]"},
									},
								},
							}
							or = append(or, q)
						}
					case "lookup":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							}
							or = append(or, q)
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							or = append(or, q)
						}
					case "options", "user":
						// 只能是in和=两种情况（默认是等于）
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$in": values},
								}
								or = append(or, q)
							}
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							// 等于
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							or = append(or, q)
						}
					case "number", "date", "time":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							q := bson.M{
								"$and": []bson.M{
									{"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
									{"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
								},
							}
							or = append(or, q)
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gt": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == "<" {
								// 小于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lte": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else if condition.Operator == "<>" {
								// 不等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$ne": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								or = append(or, q)
							} else {
								// 默认的[等于]的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
								}
								or = append(or, q)
							}
						}
					default:
					}
				} else {
					switch condition.FieldType {
					case "options", "type", "user", "group":
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								q := bson.M{
									condition.FieldID: bson.M{"$in": values},
								}
								or = append(or, q)
							}
						} else if condition.Operator == "<>" {
							q := bson.M{
								condition.FieldID: bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							// 等于
							q := bson.M{
								condition.FieldID: condition.SearchValue,
							}
							or = append(or, q)
						}
					case "check":
						// 等于
						q := bson.M{
							condition.FieldID: condition.SearchValue,
						}
						or = append(or, q)
					case "datetime":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							q := bson.M{
								"$and": []bson.M{
									{condition.FieldID: bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
									{condition.FieldID: bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
								},
							}
							or = append(or, q)
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									condition.FieldID: bson.M{"$gte": fullTime},
								}
								or = append(or, q)
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								value := getTime(condition.SearchValue)
								q := bson.M{
									condition.FieldID: bson.M{"$gte": value},
								}
								or = append(or, q)
							} else if condition.Operator == "<" {
								// 小于的场合
								value := getTime(condition.SearchValue)
								q := bson.M{
									condition.FieldID: bson.M{"$lt": value},
								}
								or = append(or, q)
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								value := getTime(condition.SearchValue)
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									condition.FieldID: bson.M{"$lt": fullTime},
								}
								or = append(or, q)
							} else if condition.Operator == "<>" {
								// 不等于的场合
								value := getTime(condition.SearchValue)
								zeroTime := value
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									"$or": []bson.M{
										{condition.FieldID: bson.M{"$gte": fullTime}},
										{condition.FieldID: bson.M{"$lt": zeroTime}},
									},
								}
								or = append(or, q)
							} else {
								// 默认的[等于]的场合
								value := getTime(condition.SearchValue)
								zeroTime := value
								fullTime := value.Add(time.Hour * 24)
								q := bson.M{
									"$and": []bson.M{
										{condition.FieldID: bson.M{"$gte": zeroTime}},
										{condition.FieldID: bson.M{"$lt": fullTime}},
									},
								}
								or = append(or, q)
							}
						}
					default:
					}
				}
			}

			if len(or) > 0 {
				query["$or"] = or
			}
		}
	}
}

// FindKaraCount 获取台账唯一字段空值总件数
func FindKaraCount(db string, params KaraCountParam) (total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(params.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       params.AppID,
		"datastore_id": params.DatastoreID,
	}

	if len(params.Owners) > 0 {
		query["owners"] = bson.M{"$in": params.Owners}
	}

	// 空项
	or := []bson.M{}
	q := bson.M{
		"items." + params.FieldID + ".value": nil,
	}
	or = append(or, q)

	// 空值
	switch params.FieldType {
	case "number":
		q := bson.M{
			"items." + params.FieldID + ".value": 0.0,
		}
		or = append(or, q)
	case "switch":
		// 无空值状态
		break
	case "user", "file":
		q := bson.M{
			"items." + params.FieldID + ".value": getSearchValue(params.FieldType, "[]"),
		}
		or = append(or, q)
	default:
		q := bson.M{
			"items." + params.FieldID + ".value": getSearchValue(params.FieldType, ""),
		}
		or = append(or, q)
	}

	query["$or"] = or

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindKaraCount", fmt.Sprintf("query: [ %s ]", queryJSON))

	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		return 0, err
	}

	return t, nil
}

// FindCount 获取台账数据件数
func FindCount(db string, params CountParam) (total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(params.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       params.AppID,
		"datastore_id": params.DatastoreID,
	}

	if len(params.Owners) > 0 {
		query["owners"] = bson.M{"$in": params.Owners}
	}

	buildMatch(params.ConditionList, params.ConditionType, query)

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		return 0, err
	}

	return t, nil
}

// FindItem 通过ID获取数据信息
func FindItem(db string, p *ItemParam) (item Item, err error) {

	if p.IsOrigin {
		return getItem(db, p.ItemID, p.DatastoreID, p.Owners)
	}

	return getItemWithLookup(db, p)
}

// getItemWithLookup 通过ID获取数据信息
func getItemWithLookup(db string, p *ItemParam) (item Item, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(p.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Item
	var query bson.M

	objectID, err := primitive.ObjectIDFromHex(p.ItemID)
	if err != nil {
		utils.ErrorLog("FindItem", err.Error())
		return result, err
	}
	// 默认过滤掉被软删除的数据
	query = bson.M{
		"_id": objectID,
		"owners": bson.M{
			"$in": p.Owners,
		},
	}

	ds, err := getDatastore(db, p.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return result, err
	}

	fields, err := getFields(db, p.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindItem", err.Error())
		return result, err
	}

	skip := 0
	limit := 1

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	pipe = append(pipe, bson.M{
		"$skip": skip,
	})
	pipe = append(pipe, bson.M{
		"$limit": limit,
	})

	project := bson.M{
		"_id":          1,
		"item_id":      1,
		"app_id":       1,
		"datastore_id": 1,
		// "owners":       1,
		"check_type":   1,
		"check_status": 1,
		"created_at":   1,
		// "created_by":   1,
		"updated_at": 1,
		// "updated_by":   1,
		"checked_at": 1,
		// "checked_by":   1,
		"label_time": 1,
		"status":     1,
	}

	// 所有者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$in": []string{"$access_key", "$$owners"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "groups",
			"let": bson.M{
				"owners": "$owners",
			},
			"pipeline": pp,
			"as":       "relations_owners",
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})

		project["owners"] = "$relations_owners.group_name"
	}
	// 创建者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$eq": []string{"$user_id", "$$user"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "users",
			"let": bson.M{
				"user": "$created_by",
			},
			"pipeline": pp,
			"as":       "relations_created_by",
		}
		unwind := bson.M{
			"path":                       "$relations_created_by",
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})

		project["created_by"] = "$relations_created_by.user_name"
	}
	// 更新者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$eq": []string{"$user_id", "$$user"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "users",
			"let": bson.M{
				"user": "$updated_by",
			},
			"pipeline": pp,
			"as":       "relations_updated_by",
		}

		unwind := bson.M{
			"path":                       "$relations_updated_by",
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})

		project["updated_by"] = "$relations_updated_by.user_name"
	}
	// 盘点者
	{
		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$eq": []string{"$user_id", "$$user"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from": "users",
			"let": bson.M{
				"user": "$checked_by",
			},
			"pipeline": pp,
			"as":       "relations_checked_by",
		}

		unwind := bson.M{
			"path":                       "$relations_checked_by",
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})

		project["checked_by"] = "$relations_checked_by.user_name"
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
			project["items."+f.FieldID+".data_type"] = "$items." + f.FieldID + ".data_type"

			continue
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
			project["items."+f.FieldID+".data_type"] = "$items." + f.FieldID + ".data_type"

			continue
		}

		// 函数字段，重新拼接
		if f.FieldType == "function" {
			var formula bson.M
			err := json.Unmarshal([]byte(f.Formula), &formula)
			if err != nil {
				utils.ErrorLog("FindItem", err.Error())
				return result, err
			}

			if len(formula) > 0 {
				project["items."+f.FieldID+".value"] = formula
			} else {
				project["items."+f.FieldID+".value"] = ""
			}

			// 当前数据本身
			project["items."+f.FieldID+".data_type"] = f.ReturnType
			continue
		}

		project["items."+f.FieldID] = "$items." + f.FieldID
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
		utils.ErrorLog("FindItem", err.Error())
		return result, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item Item
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindItem", err.Error())
			return result, err
		}

		result = item
	}

	if len(result.ItemID) == 0 {
		return result, mongo.ErrNoDocuments
	}

	return result, nil
}

// getItem 通过ID获取数据信息
func getItem(db, itemId, datastoreId string, owners []string) (item Item, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreId))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Item
	var query bson.M

	objectID, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		utils.ErrorLog("FindItem", err.Error())
		return result, err
	}
	// 默认过滤掉被软删除的数据
	query = bson.M{
		"_id": objectID,
	}

	if len(owners) > 0 {
		query["owners"] = bson.M{
			"$in": owners,
		}
	}

	ds, err := getDatastore(db, datastoreId)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return result, err
	}

	fields, err := getFields(db, datastoreId)
	if err != nil {
		utils.ErrorLog("FindItem", err.Error())
		return result, err
	}

	skip := 0
	limit := 1

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	pipe = append(pipe, bson.M{
		"$skip": skip,
	})
	pipe = append(pipe, bson.M{
		"$limit": limit,
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
		// 函数字段，重新拼接
		if f.FieldType == "function" {
			var formula bson.M
			err := json.Unmarshal([]byte(f.Formula), &formula)
			if err != nil {
				utils.ErrorLog("FindItem", err.Error())
				return result, err
			}

			if len(formula) > 0 {
				project["items."+f.FieldID+".value"] = formula
			} else {
				project["items."+f.FieldID+".value"] = ""
			}

			// 当前数据本身
			project["items."+f.FieldID+".data_type"] = f.ReturnType
			continue
		}

		project["items."+f.FieldID] = "$items." + f.FieldID
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
		utils.ErrorLog("FindItem", err.Error())
		return result, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item Item
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindItem", err.Error())
			return result, err
		}

		result = item
	}

	if len(result.ItemID) == 0 {
		return result, mongo.ErrNoDocuments
	}

	return result, nil
}

// FindRishiritsu 获取数据信息(利子率)
func FindRishiritsu(db string, p *RishiritsuParam) (item Item, err error) {

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(p.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	leaseStymd := p.LeaseStymd
	// 去掉时区名称部分
	leaseStymd = strings.Split(leaseStymd, " (")[0]
	layout := "Mon Jan 02 2006 15:04:05 MST 0900"
	// 解析时间字符串
	t, err := time.Parse(layout, leaseStymd)
	if err != nil {
		utils.ErrorLog("FindRishiritsu", err.Error())
		return
	}
	// 转换为 ISO 8601 格式
	isoTime := t.UTC().Format(time.RFC3339)
	newLeaseStymd := isoTime
	newLayout := "2006-01-02T15:04:05Z"
	// 将字符串转换为 time.Time 类型
	newTime, err := time.Parse(newLayout, newLeaseStymd)
	if err != nil {
		utils.ErrorLog("FindRishiritsu", err.Error())
		return
	}

	leaseKikan := p.LeaseKikan
	// 将字符串转换为 float64
	floatLeaseKikan, err := strconv.ParseFloat(leaseKikan, 64)
	if err != nil {
		utils.ErrorLog("FindRishiritsu", err.Error())
		return
	}

	var resultTime Item
	pipeTime := []bson.M{}

	addFieldsTime := bson.M{
		"$addFields": bson.M{
			"timeDiff": bson.M{
				"$abs": bson.M{
					"$subtract": bson.A{"$items.baseym.value", newTime},
				},
			},
		},
	}

	queryTime := bson.M{
		"$match": bson.M{
			"items.baseym.value": bson.M{
				"$lte": newTime,
			},
		},
	}

	sortTime := bson.M{
		"$sort": bson.M{
			"timeDiff": 1,
		},
	}

	limitTime := bson.M{
		"$limit": 1,
	}

	pipeTime = append(pipeTime, addFieldsTime)
	pipeTime = append(pipeTime, queryTime)
	pipeTime = append(pipeTime, sortTime)
	pipeTime = append(pipeTime, limitTime)

	queryJSONTime, _ := json.Marshal(pipeTime)
	utils.DebugLog("FindRishiritsu", fmt.Sprintf("query: [ %s ]", queryJSONTime))

	optTime := options.Aggregate()
	optTime.SetAllowDiskUse(true)

	curTime, err := c.Aggregate(ctx, pipeTime, optTime)
	if err != nil {
		utils.ErrorLog("FindRishiritsu", err.Error())
		return resultTime, err
	}
	defer curTime.Close(ctx)

	for curTime.Next(ctx) {
		var item Item
		err := curTime.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindRishiritsu", err.Error())
			return resultTime, err
		}

		resultTime = item
	}

	if len(resultTime.ItemID) == 0 {
		return resultTime, mongo.ErrNoDocuments
	}

	var result Item
	pipe := []bson.M{}

	addFields := bson.M{
		"$addFields": bson.M{
			"valueDiff": bson.M{
				"$abs": bson.M{
					"$subtract": bson.A{"$items.leaseperiod.value", floatLeaseKikan},
				},
			},
		},
	}

	query := bson.M{
		"$match": bson.M{
			"items.baseym.value": resultTime.ItemMap["baseym"].Value,
			"items.leaseperiod.value": bson.M{
				"$gte": floatLeaseKikan,
			},
		},
	}

	sort := bson.M{
		"$sort": bson.M{
			"valueDiff": 1,
		},
	}

	limit := bson.M{
		"$limit": 1,
	}

	pipe = append(pipe, addFields)
	pipe = append(pipe, query)
	pipe = append(pipe, sort)
	pipe = append(pipe, limit)

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindRishiritsu", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("FindRishiritsu", err.Error())
		return result, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item Item
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindRishiritsu", err.Error())
			return result, err
		}

		result = item
	}

	if len(result.ItemID) == 0 {
		return result, mongo.ErrNoDocuments
	}

	return result, nil
}

// AddCopyItem 复制台账数据
func AddCopyItem(db, collection string, i *Item) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(i.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	i.ID = primitive.NewObjectID()
	i.ItemID = i.ID.Hex()

	if _, err = c.InsertOne(ctx, i); err != nil {
		utils.ErrorLog("AddItem", err.Error())
		return "", err
	}

	return i.ItemID, nil
}

// AddItem 添加台账数据
func AddItem(db, collection, lang, domain string, i *Item) (id string, err error) {
	if _, exist := i.ItemMap["copykbn"]; exist {
		itemId, copyerr := AddCopyItem(db, collection, i)
		return itemId, copyerr
	}

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(i.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	callback := func(sc mongo.SessionContext) (interface{}, error) {
		i.ID = primitive.NewObjectID()
		i.ItemID = i.ID.Hex()
		i.Status = "1"
		i.CheckStatus = "0"

		// 获取所有台账
		dsList, e := FindDatastores(db, i.AppID, "", "", "")
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				dsList = []Datastore{}
			} else {
				utils.ErrorLog("AddItem", err.Error())
				return "", err
			}
		}

		// 根据所有台账，获取所有字段数据
		fieldMap, err := getFieldMap(db, i.AppID)
		if err != nil {
			utils.ErrorLog("AddItem", err.Error())
			return "", err
		}

		dsMap := make(map[string]string)
		for _, d := range dsList {
			dsMap[d.ApiKey] = d.DatastoreID
		}
		//dsrireki := dsMap["rireki"]
		//cr := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsrireki))

		// 查找该台账的信息
		var ds Datastore
		for _, d := range dsList {
			if d.DatastoreID == i.DatastoreID {
				ds = d
			}
		}

		// 判断是否为契约台账
		if ds.ApiKey == "keiyakudaicho" {
			if i.ItemMap["keiyakuno"].Value == "" {
				num, err := keiyakunoAuto(sc, db, i.DatastoreID)
				if err != nil {
					return nil, err
				}
				i.ItemMap["keiyakuno"].Value = num
			}

			// 契约台账新规的场合添加契约百分比为1
			i.ItemMap["percentage"] = &Value{
				DataType: "number",
				Value:    1,
			}
			// 租赁满了日算出
			leasestymd := GetValueFromModel(i.ItemMap["leasestymd"])[:10]
			leasekikan := GetValueFromModel(i.ItemMap["leasekikan"])
			extentionOption := GetValueFromModel(i.ItemMap["extentionOption"])
			expireymd, err := GetExpireymd(leasestymd, leasekikan, extentionOption)
			hkkjitenzans := GetValueFromModel(i.ItemMap["hkkjitenzan"])
			sonnekigakus := GetValueFromModel(i.ItemMap["sonnekigaku"])
			hkkjitenzan, _ := strconv.ParseFloat(hkkjitenzans, 64)
			sonnekigaku, _ := strconv.ParseFloat(sonnekigakus, 64)

			if err != nil {
				utils.ErrorLog("AddItem", err.Error())
				return "", err
			}
			// 契约台账新规的场合添加租赁满了日
			i.ItemMap["leaseexpireymd"] = &Value{
				DataType: "date",
				Value:    expireymd,
			}
			i.ItemMap["hkkjitenzan"] = &Value{
				DataType: "number",
				Value:    hkkjitenzan,
			}
			i.ItemMap["sonnekigaku"] = &Value{
				DataType: "number",
				Value:    sonnekigaku,
			}

			// 处理月取得
			cfg, err := getConfig(db, i.AppID)
			if err != nil {
				utils.ErrorLog("AddItem", err.Error())
				return "", err
			}
			handleMonth := cfg.GetSyoriYm()
			if !ExpireCheck(expireymd.Format("2006-01-02"), handleMonth) {
				i.ItemMap["status"] = &Value{
					DataType: "options",
					Value:    "complete",
				}
			} else {
				i.ItemMap["status"] = &Value{
					DataType: "options",
					Value:    "normal",
				}
			}
		}

		fields := fieldMap[i.DatastoreID]

		for _, f := range fields {
			if f.FieldType == "autonum" {
				num, err := autoNum(sc, db, f)
				if err != nil {
					return nil, err
				}

				i.ItemMap[f.FieldID] = &Value{
					DataType: "autonum",
					Value:    num,
				}
				continue
			}

			//  添加空数据
			addEmptyData(i.ItemMap, f)
		}

		// 临时数据ID
		templateID := ""
		if val, exist := i.ItemMap["template_id"]; exist {
			templateID = val.Value.(string)
			// 删除临时数据ID
			delete(i.ItemMap, "template_id")
		}

		// hs := NewHistory(db, i.CreatedBy, i.DatastoreID, lang, domain, sc, fields)

		// err = hs.Add("1", i.ItemID, nil)
		// if err != nil {
		// 	utils.ErrorLog("AddItem", err.Error())
		// 	return nil, err
		// }

		// 插入数据
		queryJSON, _ := json.Marshal(i)
		utils.DebugLog("AddItem", fmt.Sprintf("item: [ %s ]", queryJSON))

		if _, err = c.InsertOne(sc, i); err != nil {
			utils.ErrorLog("AddItem", err.Error())
			return nil, err
		}

		// 判断是否为契约台账，并登录契约生成的数据
		if ds.ApiKey == "keiyakudaicho" {
			ct := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))

			rirekiSeq, err := uuid.NewUUID()
			if err != nil {
				utils.ErrorLog("AddItem", err.Error())
				return nil, err
			}

			// 变更后契约履历情报
			newItemMap := copyMap(i.ItemMap)
			// 契約履歴番号
			newItemMap["no"] = &Value{
				DataType: "text",
				Value:    rirekiSeq.String(),
			}
			// 操作区分编辑
			newItemMap["actkbn"] = &Value{
				DataType: "options",
				Value:    "",
			}
			// 修正区分编辑
			newItemMap["zengokbn"] = &Value{
				DataType: "options",
				Value:    "after",
			}
			// 对接区分编辑
			newItemMap["dockkbn"] = &Value{
				DataType: "options",
				Value:    "undo",
			}
			// 将契约番号变成lookup类型
			keiyakuno := i.ItemMap["keiyakuno"]

			keiyakuItem := keiyakuno.Value.(string)

			itMap := make(map[string]interface{})
			for key, val := range i.ItemMap {
				itMap[key] = val
			}

			newItemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}

			leaseType := i.ItemMap["lease_type"]
			leaseTypeVal := leaseType.Value.(string)

			if leaseTypeVal == "normal_lease" {
				// 查询临时表取得契约履历的下列数据
				queryTmp := bson.M{
					"template_id":   templateID,
					"datastore_key": "rireki",
				}
				var rirekiTmp TemplateItem

				if err := ct.FindOne(ctx, queryTmp).Decode(&rirekiTmp); err != nil {
					utils.ErrorLog("AddItem", err.Error())
					return nil, err
				}

				// リース料総額
				newItemMap["leaseTotal"] = &Value{
					DataType: "number",
					Value:    rirekiTmp.ItemMap["leaseTotal"].Value,
				}
				// リース料総額の現在価値
				newItemMap["presentTotal"] = &Value{
					DataType: "number",
					Value:    rirekiTmp.ItemMap["presentTotal"].Value,
				}
				// 処理月度の先月までの償却費の累計額
				newItemMap["preDepreciationTotal"] = &Value{
					DataType: "number",
					Value:    rirekiTmp.ItemMap["preDepreciationTotal"].Value,
				}
			}

			// 添加新契约履历情报
			/* var newRirekiItem Item
			newRirekiItem.ID = primitive.NewObjectID()
			newRirekiItem.ItemID = newRirekiItem.ID.Hex()
			newRirekiItem.AppID = i.AppID
			newRirekiItem.DatastoreID = dsrireki
			newRirekiItem.CreatedAt = i.UpdatedAt
			newRirekiItem.CreatedBy = i.UpdatedBy
			newRirekiItem.ItemMap = newItemMap
			newRirekiItem.Owners = i.Owners */

			/* if _, err := cr.InsertOne(sc, newRirekiItem); err != nil {
				utils.ErrorLog("AddItem", err.Error())
				return nil, err
			} */

			// 获取会社信息
			leasekaisha := i.ItemMap["leasekaishacd"]
			kaisyaItem := leasekaisha.Value.(string)

			// 获取分類コード信息
			bunruicd := i.ItemMap["bunruicd"]
			bunruicdItem := bunruicd.Value.(string)

			// 获取管理部門
			segmentcd := i.ItemMap["segmentcd"]
			segmentcdItem := segmentcd.Value.(string)

			// 取支付信息数据登录数据库 paymentStatus
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "paymentStatus",
				UserID:        i.CreatedBy,
				Owners:        i.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
				FileMap:       fieldMap,
			})
			if err != nil {
				utils.ErrorLog("AddItem", err.Error())
				return nil, err
			}

			if leaseTypeVal == "normal_lease" {
				// 取利息数据登录数据库 paymentInterest
				err = insertTempData(client, sc, TmpParam{
					DB:            db,
					TemplateID:    templateID,
					APIKey:        "paymentInterest",
					UserID:        i.CreatedBy,
					Owners:        i.Owners,
					Datastores:    dsList,
					Keiyakuno:     keiyakuItem,
					Leasekaishacd: kaisyaItem,
					Bunruicd:      bunruicdItem,
					Segmentcd:     segmentcdItem,
					FileMap:       fieldMap,
				})
				if err != nil {
					utils.ErrorLog("AddItem", err.Error())
					return nil, err
				}

				// 取偿还数据登录数据库 repayment
				err = insertTempData(client, sc, TmpParam{
					DB:            db,
					TemplateID:    templateID,
					APIKey:        "repayment",
					UserID:        i.CreatedBy,
					Owners:        i.Owners,
					Datastores:    dsList,
					Keiyakuno:     keiyakuItem,
					Leasekaishacd: kaisyaItem,
					Bunruicd:      bunruicdItem,
					Segmentcd:     segmentcdItem,
					FileMap:       fieldMap,
				})
				if err != nil {
					utils.ErrorLog("AddItem", err.Error())
					return nil, err
				}
			}

			// 删除临时数据
			tc := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
			query := bson.M{
				"template_id": templateID,
			}

			_, err = tc.DeleteMany(sc, query)
			if err != nil {
				utils.ErrorLog("AddItem", err.Error())
				return nil, err
			}
		}

		/* err = hs.Compare("1", i.ItemMap)
		if err != nil {
			utils.ErrorLog("AddItem", err.Error())
			return nil, err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("AddItem", err.Error())
			return nil, err
		}
		*/
		return nil, nil
	}

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("AddItem", err.Error())
		return "", err
	}

	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		isDuplicate := mongo.IsDuplicateKeyError(err)
		if isDuplicate {
			we, ok := err.(mongo.WriteException)
			if !ok {
				utils.ErrorLog("AddItem", err.Error())
				return "", err
			}
			errInfo := we.WriteErrors[0]
			em := errInfo.Message
			values := utils.FieldMatch(`"([^\"]+)"`, em[strings.LastIndex(em, "dup key"):])
			for i, v := range values {
				values[i] = strings.Trim(v, `"`)
			}
			fields := utils.FieldMatch(`field_[0-9a-z]{3}`, em[strings.LastIndex(em, "dup key"):])

			utils.ErrorLog("AddItem", errInfo.Message)
			return "", fmt.Errorf("プライマリキーの重複エラー、API-KEY[%s],重複値は[%s]です。", strings.Join(fields, ","), strings.Join(values, ","))
		}

		utils.ErrorLog("AddItem", err.Error())
		return "", err
	}

	return i.ItemID, nil
}

// ModifyItem 修改台账数据
func ModifyItem(db string, p *ItemUpdateParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(p.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	param := &FindFieldsParam{
		AppID:       p.AppID,
		DatastoreID: p.DatastoreID,
	}

	allFields, err := FindFields(db, param)
	if err != nil {
		utils.ErrorLog("ModifyItem", err.Error())
		if err.Error() != mongo.ErrNoDocuments.Error() {
			return err
		}
	}

	oldItem, e := getItem(db, p.ItemID, p.DatastoreID, p.Owners)
	if e != nil {
		if e.Error() == mongo.ErrNoDocuments.Error() {
			return errors.New("データが存在しないか、あなたまたはデータの申請者はデータを変更する権限がありません")
		}
		utils.ErrorLog("ModifyItem", e.Error())
		return e
	}

	callback := func(sc mongo.SessionContext) (interface{}, error) {
		// 自增字段不更新
		if len(allFields) > 0 {
			for _, f := range allFields {
				if f.FieldType == "autonum" {
					delete(oldItem.ItemMap, f.FieldID)
					delete(p.ItemMap, f.FieldID)
					continue
				}
				_, ok := p.ItemMap[f.FieldID]
				// 需要进行自算的情况
				if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
					if f.SelfCalculate == "add" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o + n
						continue
					}
					if f.SelfCalculate == "sub" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o - n
						continue
					}
				}
			}
		}

		hs := NewHistory(db, p.UpdatedBy, p.DatastoreID, p.Lang, p.Domain, sc, allFields)

		err := hs.Add("1", p.ItemID, oldItem.ItemMap)
		if err != nil {
			utils.ErrorLog("ModifyItem", err.Error())
			return nil, err
		}

		item := bson.M{
			"updated_at": p.UpdatedAt,
			"updated_by": p.UpdatedBy,
		}

		// 盘点状态初始化
		if oldItem.CheckStatus == "" {
			item["check_status"] = "0"
		}

		for k, v := range p.ItemMap {
			item["items."+k] = v
		}

		update := bson.M{"$set": item}

		objectID, err := primitive.ObjectIDFromHex(p.ItemID)
		if err != nil {
			utils.ErrorLog("ModifyItem", err.Error())
			return nil, err
		}

		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("ModifyItem", fmt.Sprintf("update: [ %s ]", updateJSON))

		if _, err := c.UpdateByID(ctx, objectID, update); err != nil {
			utils.ErrorLog("ModifyItem", err.Error())
			return nil, err
		}

		err = hs.Compare("1", p.ItemMap)
		if err != nil {
			utils.ErrorLog("ModifyItem", err.Error())
			return nil, err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("ModifyItem", err.Error())
			return nil, err
		}

		return nil, nil
	}

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("AddItem", err.Error())
		return err
	}

	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		utils.ErrorLog("AddItem", err.Error())
		return err
	}

	return nil
}

// GenerateItem
func GenerateItem(db, datastoreID, startDate, lastDate string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	startDay, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		utils.ErrorLog("Error parsing start date:", err.Error())
		return
	}

	lastDay, err := time.Parse("2006-01-02", lastDate)
	if err != nil {
		utils.ErrorLog("Error parsing last date:", err.Error())
		return
	}

	defaultTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

	query := bson.M{
		"$and": []bson.M{
			{"items.keijoudate.value": bson.M{"$gte": startDay}},
			{"items.keijoudate.value": bson.M{"$lte": lastDay}},
			{"items.kakuteidate.value": defaultTime},
		},
	}

	update := bson.M{"$set": bson.M{
		"items.sakuseidate.value": time.Now(),
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("GenerateItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("GenerateItem", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("GenerateItem", err.Error())
		return err
	}
	return
}

// GenerateShoukyakuItem
func GenerateShoukyakuItem(db, datastoreID, startDate, lastDate string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	startDay, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		utils.ErrorLog("Error parsing start date:", err.Error())
		return
	}

	lastDay, err := time.Parse("2006-01-02", lastDate)
	if err != nil {
		utils.ErrorLog("Error parsing last date:", err.Error())
		return
	}

	defaultTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

	query := bson.M{
		"$and": []bson.M{
			{"items.syokyakuymd.value": bson.M{"$gte": startDay}},
			{"items.syokyakuymd.value": bson.M{"$lte": lastDay}},
			{"items.kakuteidate.value": defaultTime},
		},
	}

	update := bson.M{"$set": bson.M{
		"items.sakuseidate.value": time.Now(),
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("GenerateShoukyakuItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("GenerateShoukyakuItem", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("GenerateShoukyakuItem", err.Error())
		return err
	}
	return
}

// ConfimItem
func ConfimItem(db, datastoreID, startDate, lastDate string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startDay, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		utils.ErrorLog("Error parsing start date:", err.Error())
		return
	}

	lastDay, err := time.Parse("2006-01-02", lastDate)
	if err != nil {
		utils.ErrorLog("Error parsing start date:", err.Error())
		return
	}

	query := bson.M{
		"$or": []bson.M{
			{
				"$and": []bson.M{
					{"items.keijoudate.value": bson.M{"$gte": startDay}},
					{"items.keijoudate.value": bson.M{"$lte": lastDay}},
				},
			},
			{
				"$and": []bson.M{
					{"items.syokyakuymd.value": bson.M{"$gte": startDay}},
					{"items.syokyakuymd.value": bson.M{"$lte": lastDay}},
				},
			},
		},
	}

	update := bson.M{"$set": bson.M{
		"items.kakuteidate.value": time.Now(),
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ConfimItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ConfimItem", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("ConfimItem", err.Error())
		return err
	}
	return
}

// ResetInventoryItems 盘点台账盘点数据盘点状态重置
func ResetInventoryItems(db, userID, appID string) (err error) {
	client := database.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 盘点台账取得
	ds, err := FindDatastores(db, appID, "", "true", "")
	if err != nil {
		utils.ErrorLog("ResetInventoryItems", err.Error())
		return err
	}

	// 盘点开始日付取得
	cfg, err := getConfig(db, appID)
	if err != nil {
		utils.ErrorLog("ResetInventoryItems", err.Error())
		return err
	}
	checkStartDate := cfg.GetCheckStartDate()
	startDate, err := time.Parse("2006-01-02", checkStartDate)
	if err != nil {
		utils.ErrorLog("ResetInventoryItems", err.Error())
		return err
	}

	// 循环重置
	for _, d := range ds {
		c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(d.DatastoreID))

		query := bson.M{
			"$and": []bson.M{
				{"checked_at": bson.M{"$gt": startDate}},
				{"checked_by": bson.M{"$ne": ""}},
			},
		}

		update := bson.M{"$set": bson.M{
			"check_status": "1",
			"updated_at":   time.Now(),
			"updated_by":   userID,
		}}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("ResetInventoryItems", fmt.Sprintf("query: [ %s ]", queryJSON))

		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("ResetInventoryItems", fmt.Sprintf("update: [ %s ]", updateJSON))

		_, err = c.UpdateMany(ctx, query, update)
		if err != nil {
			utils.ErrorLog("ResetInventoryItems", err.Error())
			return err
		}
		query1 := bson.M{
			"checked_at": bson.M{"$lt": startDate},
		}

		update1 := bson.M{"$set": bson.M{
			"check_status": "0",
			"updated_at":   time.Now(),
			"updated_by":   userID,
		}}

		queryJSON1, _ := json.Marshal(query1)
		utils.DebugLog("ResetInventoryItems", fmt.Sprintf("query: [ %s ]", queryJSON1))

		updateJSON1, _ := json.Marshal(update1)
		utils.DebugLog("ResetInventoryItems", fmt.Sprintf("update: [ %s ]", updateJSON1))

		_, err = c.UpdateMany(ctx, query1, update1)
		if err != nil {
			utils.ErrorLog("ResetInventoryItems", err.Error())
			return err
		}
	}

	return nil
}

// InventoryItem 单个盘点处理
func InventoryItem(db, userID, itemID, datastoreID, appID, checkType, image, checkField string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	hc := client.Database(database.GetDBName(db)).Collection(CheckHistoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, e := primitive.ObjectIDFromHex(itemID)
	if e != nil {
		utils.ErrorLog("InventoryItem", e.Error())
		return e
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("InventoryItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	oldItem, err := getItem(db, itemID, datastoreID, []string{})
	if err != nil {
		utils.ErrorLog("InventoryItem", err.Error())
		return err
	}

	// 处理月取得
	cfg, err := getConfig(db, appID)
	if err != nil {
		utils.ErrorLog("InventoryItem", err.Error())
		return err
	}
	// 获取盘点开始日付，判断是否更新盘点状态
	checkStartDate, err := time.ParseInLocation("2006-01-02", cfg.GetCheckStartDate(), time.Local)
	if err != nil {
		utils.ErrorLog("InventoryItem", err.Error())
		return err
	}
	nowTime := time.Now()

	if checkType == "Image" {
		var fs []File
		if value, ok := oldItem.ItemMap[checkField]; ok {
			oldValue := value.Value.(string)
			err = json.Unmarshal([]byte(oldValue), &fs)
			if err != nil {
				utils.ErrorLog("InventoryItem", err.Error())
				return err
			}
		}

		imageFile := File{
			Name: filepath.Base(image),
			URL:  image,
		}

		fs = append(fs, imageFile)

		value, err := json.Marshal(&fs)
		if err != nil {
			utils.ErrorLog("InventoryItem", err.Error())
			return err
		}
		var updateSet = bson.M{}
		if checkStartDate.Before(nowTime) {
			updateSet = bson.M{"$set": bson.M{
				"items." + checkField + ".value":     string(value),
				"items." + checkField + ".data_type": "file",
				"check_type":                         checkType,
				"check_status":                       "1",
				"checked_at":                         nowTime,
				"checked_by":                         userID,
				"updated_at":                         nowTime,
				"updated_by":                         userID,
			}}
		} else {
			updateSet = bson.M{"$set": bson.M{
				"items." + checkField + ".value":     string(value),
				"items." + checkField + ".data_type": "file",
				"check_type":                         checkType,
				"check_status":                       "0",
				"checked_at":                         nowTime,
				"checked_by":                         userID,
				"updated_at":                         nowTime,
				"updated_by":                         userID,
			}}

		}
		updateSetJSON, _ := json.Marshal(updateSet)
		utils.DebugLog("InventoryItem", fmt.Sprintf("update: [ %s ]", updateSetJSON))

		_, err = c.UpdateOne(ctx, query, updateSet)
		if err != nil {
			utils.ErrorLog("InventoryItem", err.Error())
			return err
		}
	} else {
		var update = bson.M{}
		if checkStartDate.Before(nowTime) {
			update = bson.M{"$set": bson.M{
				"check_type":   checkType,
				"check_status": "1",
				"checked_at":   nowTime,
				"checked_by":   userID,
				"updated_at":   nowTime,
				"updated_by":   userID,
			}}
		} else {
			update = bson.M{"$set": bson.M{
				"check_type":   checkType,
				"check_status": "0",
				"checked_at":   nowTime,
				"checked_by":   userID,
				"updated_at":   nowTime,
				"updated_by":   userID,
			}}
		}

		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("InventoryItem", fmt.Sprintf("update: [ %s ]", updateJSON))

		if _, err := c.UpdateOne(ctx, query, update); err != nil {
			utils.ErrorLog("InventoryItem", err.Error())
			return err
		}
	}

	h := &CheckHistory{
		ItemId:         itemID,
		DatastoreId:    datastoreID,
		ItemMap:        oldItem.ItemMap,
		CheckType:      checkType,
		CheckStartDate: cfg.GetCheckStartDate(),
		CheckedAt:      time.Now(),
		CheckedBy:      userID,
	}

	h.ID = primitive.NewObjectID()
	h.CheckId = h.ID.Hex()

	_, err = hc.InsertOne(ctx, h)
	if err != nil {
		utils.ErrorLog("InventoryItem", err.Error())
		return err
	}

	return nil
}

// MutilInventoryItem 多条数据盘点
func MutilInventoryItem(db, userID, datastoreID, appID, checkType string, itemIDList []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	hc := client.Database(database.GetDBName(db)).Collection(CheckHistoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var models []mongo.WriteModel
	var hsModels []mongo.WriteModel

	// 处理月取得
	cfg, err := getConfig(db, appID)
	if err != nil {
		utils.ErrorLog("MutilInventoryItem", err.Error())
		return err
	}
	checkStartDate, err := time.ParseInLocation("2006-01-02", cfg.GetCheckStartDate(), time.Local)
	if err != nil {
		utils.ErrorLog("MutilInventoryItem", err.Error())
		return err
	}
	nowTime := time.Now()
	if checkStartDate.Before(nowTime) {
		for _, itemID := range itemIDList {
			update := bson.M{"$set": bson.M{
				"check_type":   checkType,
				"check_status": "1",
				"checked_at":   nowTime,
				"checked_by":   userID,
				"updated_at":   nowTime,
				"updated_by":   userID,
			}}

			objectID, err := primitive.ObjectIDFromHex(itemID)
			if err != nil {
				utils.ErrorLog("MutilInventoryItem", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("MutilInventoryItem", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("MutilInventoryItem", fmt.Sprintf("update: [ %s ]", updateJSON))

			upModel := mongo.NewUpdateOneModel()
			upModel.SetFilter(query)
			upModel.SetUpdate(update)
			upModel.SetUpsert(false)

			models = append(models, upModel)

			oldItem, err := getItem(db, itemID, datastoreID, []string{})
			if err != nil {
				utils.ErrorLog("InventoryItem", err.Error())
				return err
			}

			h := &CheckHistory{
				ItemId:         itemID,
				DatastoreId:    datastoreID,
				ItemMap:        oldItem.ItemMap,
				CheckType:      checkType,
				CheckStartDate: cfg.GetCheckStartDate(),
				CheckedAt:      time.Now(),
				CheckedBy:      userID,
			}

			h.ID = primitive.NewObjectID()
			h.CheckId = h.ID.Hex()

			hs := mongo.NewInsertOneModel()
			hs.SetDocument(h)

			hsModels = append(hsModels, hs)
		}
	} else {
		for _, itemID := range itemIDList {
			update := bson.M{"$set": bson.M{
				"check_type":   checkType,
				"check_status": "0",
				"checked_at":   nowTime,
				"checked_by":   userID,
				"updated_at":   nowTime,
				"updated_by":   userID,
			}}

			objectID, err := primitive.ObjectIDFromHex(itemID)
			if err != nil {
				utils.ErrorLog("MutilInventoryItem", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("MutilInventoryItem", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("MutilInventoryItem", fmt.Sprintf("update: [ %s ]", updateJSON))

			upModel := mongo.NewUpdateOneModel()
			upModel.SetFilter(query)
			upModel.SetUpdate(update)
			upModel.SetUpsert(false)

			models = append(models, upModel)

			oldItem, err := getItem(db, itemID, datastoreID, []string{})
			if err != nil {
				utils.ErrorLog("InventoryItem", err.Error())
				return err
			}

			h := &CheckHistory{
				ItemId:         itemID,
				DatastoreId:    datastoreID,
				ItemMap:        oldItem.ItemMap,
				CheckType:      checkType,
				CheckStartDate: cfg.GetCheckStartDate(),
				CheckedAt:      time.Now(),
				CheckedBy:      userID,
			}

			h.ID = primitive.NewObjectID()
			h.CheckId = h.ID.Hex()

			hs := mongo.NewInsertOneModel()
			hs.SetDocument(h)

			hsModels = append(hsModels, hs)
		}
	}

	if len(models) > 0 {
		result, err := c.BulkWrite(ctx, models)
		if err != nil {
			utils.ErrorLog("MutilInventoryItem", err.Error())
			return err
		}

		// TODO 回滚不可用
		if int64(len(itemIDList)) != result.ModifiedCount {
			log.Warnf("error MutilInventoryItem with image %v", "not completely update!")
		}
	}
	if len(hsModels) > 0 {
		result, err := hc.BulkWrite(ctx, hsModels)
		if err != nil {
			utils.ErrorLog("MutilInventoryItem", err.Error())
			return err
		}

		// TODO 回滚不可用
		if int64(len(itemIDList)) != result.ModifiedCount {
			log.Warnf("error MutilInventoryItem with image %v", "not completely update!")
		}
	}

	return nil
}

// DeleteItem 删除台账数据
func DeleteItem(db, datastoreID, itemID, userID, lang, domain string, owners []string) error {
	// 通过台账ID获取台账情报
	dsInfo, err := FindDatastore(db, datastoreID)
	if err != nil {
		utils.ErrorLog("DeleteItem", err.Error())
		return err
	}

	appId := dsInfo.AppID

	// 通过台账情报的台账APIKEY判断台账属性
	if dsInfo.ApiKey == "keiyakudaicho" {
		// 删除租赁契约台账数据
		err = deleteContractItem(db, appId, datastoreID, itemID, userID, lang, domain, owners)
		if err != nil {
			utils.ErrorLog("DeleteItem", err.Error())
			return err
		}
	} else {
		// 删除普通台账数据
		err = deleteItem(db, datastoreID, itemID, userID, lang, domain, owners)
		if err != nil {
			utils.ErrorLog("DeleteItem", err.Error())
			return err
		}
	}

	return nil
}

// DeleteSelectItems 删除选中的台账数据
func DeleteSelectItems(db, appID, datastoreID string, itemID []string, stream item.ItemService_DeleteSelectItemsStream) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx := context.Background()

	query := bson.M{
		"app_id":       appID,
		"datastore_id": datastoreID,
	}
	query["item_id"] = bson.M{
		"$in": itemID,
	}

	cur, err := c.Find(ctx, query)
	if err != nil {
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
			utils.ErrorLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, appID, datastoreID, dt.ItemID, err))
			return err
		}

		for _, data := range dt.ItemMap {
			var files []File
			if data.DataType == "file" {
				err := json.Unmarshal([]byte(data.Value.(string)), &files)
				if err != nil {
					utils.ErrorLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, appID, datastoreID, dt.ItemID, err))
					return err
				}
				for _, file := range files {
					// 发送的文件路径
					err := stream.Send(&item.SelectedItemsResponse{
						DeleteUrl: file.URL,
					})
					if err != nil {
						utils.ErrorLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, appID, datastoreID, dt.ItemID, err))
						return err
					}
				}
			}
		}

		current++

		objectID, err := primitive.ObjectIDFromHex(dt.ItemID)
		if err != nil {
			utils.ErrorLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s itemId:%s error: %v ", db, appID, datastoreID, dt.ItemID, err))
			return err
		}

		deCxModel := mongo.NewDeleteOneModel()
		deCxModel.SetFilter(bson.M{
			"_id": objectID,
		})
		cxModels = append(cxModels, deCxModel)
		if current%2000 == 0 {
			result, err := c.BulkWrite(ctx, cxModels)
			if err != nil {
				utils.ErrorLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s error: %v ", db, appID, datastoreID, err))
				return err
			}
			utils.InfoLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s update: %d ", db, appID, datastoreID, result.ModifiedCount))
			cxModels = cxModels[:0]
		}
	}

	if len(cxModels) > 0 {
		result, err := c.BulkWrite(ctx, cxModels)
		if err != nil {
			utils.ErrorLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s error: %v ", db, appID, datastoreID, err))
			return err
		}
		utils.InfoLog("DeleteSelectItems", fmt.Sprintf("customer:%s app:%s datastore: %s  update: %d ", db, appID, datastoreID, result.ModifiedCount))
	}
	return nil
}

// DeleteItems 删除多条数据记录
func DeleteItems(db string, dps DeleteItemsParam) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dps.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       dps.AppID,
		"datastore_id": dps.DatastoreID,
		"created_by":   dps.UserID,
	}

	buildMatch(dps.ConditionList, dps.ConditionType, query)

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteItems", fmt.Sprintf("query: [ %s ]", queryJSON))

	// 删除多条数据记录
	if _, err := c.DeleteMany(ctx, query); err != nil {
		utils.ErrorLog("DeleteItems", err.Error())
		return err
	}

	return nil
}

// DeleteDatastoreItems 物理删除台账所有数据
func DeleteDatastoreItems(db, datastoreID, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	s := client.Database(database.GetDBName(db)).Collection("sequences")
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()

	// 通过台账ID获取台账情报
	dsInfo, err := FindDatastore(db, datastoreID)
	if err != nil {
		utils.ErrorLog("DeleteDatastoreItems", err.Error())
		return err
	}

	// 通过台账情报的台账APIKEY判断台账属性
	if dsInfo.ApiKey == "keiyakudaicho" {
		querys := bson.M{
			"_id": datastoreID + "_keiyakuno_auto",
		}

		update := bson.M{
			"$set": bson.M{
				"sequence_value": 0,
			},
		}

		//将契约番号自动采番表中清0
		_, err := s.UpdateOne(ctx, querys, update)
		if err != nil {
			utils.ErrorLog("Error updating document:", err.Error())
			return err
		}
	}

	query := bson.M{
		"datastore_id": datastoreID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteDatastoreItems", fmt.Sprintf("query: [ %s ]", queryJSON))

	if _, err := c.DeleteMany(ctx, query); err != nil {
		utils.ErrorLog("DeleteDatastoreItems", err.Error())
		return err
	}

	// 清空台账的履历数据
	/* h := client.Database(database.GetDBName(db)).Collection(HistoriesCollection)
	if _, err := h.DeleteMany(ctx, query); err != nil {
		utils.ErrorLog("DeleteDatastoreItems", err.Error())
		return err
	} */
	// 清空台账的字段的履历数据
	/* fh := client.Database(database.GetDBName(db)).Collection(FieldHistoriesCollection)
	if _, err := fh.DeleteMany(ctx, query); err != nil {
		utils.ErrorLog("DeleteDatastoreItems", err.Error())
		return err
	} */

	// 重置auto字段的seq
	param := &FindFieldsParam{
		DatastoreID: datastoreID,
		FieldType:   "autonum",
	}

	// 清空台账的自动入力字段的数据
	autoFields, err := FindFields(db, param)
	if err != nil {
		utils.ErrorLog("DeleteDatastoreItems", err.Error())
		if err.Error() != mongo.ErrNoDocuments.Error() {
			return err
		}
		return nil
	}

	for _, f := range autoFields {
		err := resetAutoNum(db, f)
		if err != nil {
			utils.ErrorLog("DeleteDatastoreItems", err.Error())
			return err
		}
	}

	// 清空盘点履历
	/* ch := client.Database(database.GetDBName(db)).Collection(CheckHistoryCollection)
	if _, err := ch.DeleteMany(ctx, query); err != nil {
		utils.ErrorLog("DeleteDatastoreItems", err.Error())
		return err
	} */

	return nil
}

// ChangeOwners 更新所有者
func ChangeOwners(db string, ch *item.OwnersRequest) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(ch.DatastoreId))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"owners": ch.GetOldOwner(),
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": ch.GetWriter(),
			"owners.$":   ch.GetNewOwner(),
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ChangeOwners", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ChangeOwners", fmt.Sprintf("update: [ %s ]", updateJSON))

	result, err := c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("ChangeOwners", err.Error())
		return err
	}

	log.Infof("change %v", result.ModifiedCount)

	return nil
}

// ChangeSelectOwners 通过检索条件更新所有者信息
func ChangeSelectOwners(db string, params OwnersParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(params.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       params.AppID,
		"datastore_id": params.DatastoreID,
	}

	if len(params.OldOwners) > 0 {
		query["owners"] = bson.M{"$in": params.OldOwners}
	}

	buildMatch(params.ConditionList, params.ConditionType, query)

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": params.Writer,
			"owners":     []string{params.Owner},
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ChangeSelectOwners", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ChangeSelectOwners", fmt.Sprintf("update: [ %s ]", updateJSON))

	// 更新处理
	_, err = c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("ChangeSelectOwners", err.Error())
		return err
	}

	return nil
}

// ChangeItemOwner 更新单条记录所属组织
func ChangeItemOwner(db string, params OwnerParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(params.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       params.AppID,
		"datastore_id": params.DatastoreID,
		"item_id":      params.ItemID,
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": params.Writer,
			"owners":     []string{params.Owner},
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ChangeItemOwner", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ChangeItemOwner", fmt.Sprintf("update: [ %s ]", updateJSON))

	// 更新处理
	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("ChangeItemOwner", err.Error())
		return err
	}

	return nil
}

// ChangeStatus 更新状态
func ChangeStatus(db string, ch *item.StatusRequest) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(ch.DatastoreId))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       ch.GetAppId(),
		"datastore_id": ch.GetDatastoreId(),
		"item_id":      ch.GetItemId(),
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": ch.GetWriter(),
			"status":     ch.GetStatus(),
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ChangeStatus", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ChangeStatus", fmt.Sprintf("update: [ %s ]", updateJSON))

	result, err := c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("ChangeStatus", err.Error())
		return err
	}

	log.Infof("change %v", result.ModifiedCount)

	return nil
}

// ChangeLabelTime 修改标签出力时间
func ChangeLabelTime(db string, itemIDList []string, datastoreID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var models []mongo.WriteModel

	for _, itemID := range itemIDList {
		objectID, err := primitive.ObjectIDFromHex(itemID)
		if err != nil {
			utils.ErrorLog("ChangeLabelTime", err.Error())
			return err
		}

		query := bson.M{
			"_id": objectID,
		}

		update := bson.M{"$set": bson.M{
			"label_time": time.Now(),
		}}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("ChangeLabelTime", fmt.Sprintf("query: [ %s ]", queryJSON))

		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("ChangeLabelTime", fmt.Sprintf("update: [ %s ]", updateJSON))

		upModel := mongo.NewUpdateOneModel()
		upModel.SetFilter(query)
		upModel.SetUpdate(update)
		upModel.SetUpsert(false)

		models = append(models, upModel)
	}
	if len(models) > 0 {
		result, err := c.BulkWrite(ctx, models)
		if err != nil {
			utils.ErrorLog("ChangeLabelTime", err.Error())
			return err
		}

		// TODO 回滚不可用
		if int64(len(itemIDList)) != result.ModifiedCount {
			log.Warnf("error ChangeLabelTime %v", "not completely updated!")
		}
	}

	return nil
}

// FindUnApproveItems 查询台账未审批数据件数
func FindUnApproveItems(db string, status string, datastoreID string) (total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"status": status,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindUnApproveItems", fmt.Sprintf("query: [ %s ]", queryJSON))

	// 取总件数
	count, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindUnApproveItems", err.Error())
		return -1, err
	}

	// int转为int64
	// totalStr := strconv.Itoa(count)
	// t, _ := strconv.ParseInt(totalStr, 10, 64)

	return count, nil
}

func compare(newValue *Value, oldValue *Value) bool {
	switch newValue.DataType {
	case "text", "textarea", "time":
		if newValue.Value.(string) == oldValue.Value.(string) {
			return false
		}
		return true
	case "options":
		if newValue.Value == nil && oldValue.Value == nil {
			return false
		}

		if newValue.Value == nil || oldValue.Value == nil {
			return true
		}

		if newValue.Value.(string) == oldValue.Value.(string) {
			return false
		}
		return true
	case "number":
		new := ""
		old := ""
		switch newValue.Value.(type) {
		case int:
			new = strconv.FormatFloat(float64(newValue.Value.(int)), 'f', -1, 64)
		case int32:
			new = strconv.FormatFloat(float64(newValue.Value.(int32)), 'f', -1, 64)
		case int64:
			new = strconv.FormatFloat(float64(newValue.Value.(int64)), 'f', -1, 64)
		case float64:
			new = strconv.FormatFloat(float64(newValue.Value.(float64)), 'f', -1, 64)
		}
		switch oldValue.Value.(type) {
		case int:
			old = strconv.FormatFloat(float64(oldValue.Value.(int)), 'f', -1, 64)
		case int32:
			old = strconv.FormatFloat(float64(oldValue.Value.(int32)), 'f', -1, 64)
		case int64:
			old = strconv.FormatFloat(float64(oldValue.Value.(int64)), 'f', -1, 64)
		case float64:
			old = strconv.FormatFloat(float64(oldValue.Value.(float64)), 'f', -1, 64)
		}

		if new == old {
			return false
		}
		return true
	case "date":
		var newDate time.Time
		var oldDate time.Time
		switch newValue.Value.(type) {
		case primitive.DateTime:
			newDate = newValue.Value.(primitive.DateTime).Time()
		case time.Time:
			newDate = newValue.Value.(time.Time)
		default:
			newDate = newValue.Value.(time.Time)
		}
		switch oldValue.Value.(type) {
		case primitive.DateTime:
			oldDate = oldValue.Value.(primitive.DateTime).Time()
		case time.Time:
			oldDate = oldValue.Value.(time.Time)
		default:
			oldDate = oldValue.Value.(time.Time)
		}
		new := newDate.Format("2006-01-02")
		old := oldDate.Format("2006-01-02")
		if new == old {
			return false
		}
		return true
	case "switch":
		new := strconv.FormatBool(newValue.Value.(bool))
		old := strconv.FormatBool(oldValue.Value.(bool))
		if new == old {
			return false
		}
		return true
	case "user":
		new := newValue.Value.([]string)
		var old []string
		err := mapstructure.Decode(oldValue.Value, &old)
		if err != nil {
			return false
		}
		if reflect.DeepEqual(new, old) {
			return false
		}
		return true
	case "file":
		var new []File
		err := json.Unmarshal([]byte(newValue.Value.(string)), &new)
		if err != nil {
			return false
		}
		var old []File
		err = json.Unmarshal([]byte(oldValue.Value.(string)), &old)
		if err != nil {
			return false
		}

		return !fileSliceEqual(new, old)
	case "lookup":
		if newValue.Value == nil && oldValue.Value == nil {
			return false
		}

		if newValue.Value == nil || oldValue.Value == nil {
			return true
		}
		if newValue.Value.(string) == oldValue.Value.(string) {
			return false
		}
		return true
	default:
		return false
	}
}

func fileSliceEqual(a, b []File) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for _, ai := range a {
		exist := false

		for _, bj := range b {
			if ai.URL == bj.URL && ai.Name == bj.Name {
				exist = true
			}
		}

		if !exist {
			return false
		}
	}

	return true
}

func getOldItem(items []*Item, itemID string) *Item {
	var i *Item
Loop:
	for _, it := range items {
		if it.ItemID == itemID {
			i = it
			break Loop
		}
	}
	return i
}

// isEmptyValue 判断是否为空
func isEmptyValue(value *Value) (r bool) {
	switch value.DataType {
	case "text", "textarea", "options":
		if len(value.Value.(string)) == 0 {
			return true
		}
		return false
	case "number":
		return false
	case "autonum":
		if len(value.Value.(string)) == 0 {
			return true
		}
		return false
	case "date":
		var res time.Time
		switch value.Value.(type) {
		case primitive.DateTime:
			res = value.Value.(primitive.DateTime).Time()
		case time.Time:
			res = value.Value.(time.Time)
		default:
			res = value.Value.(time.Time)
		}
		if res.IsZero() {
			return true
		}
		return false
	case "time":
		if len(value.Value.(string)) == 0 {
			return true
		}
		return false
	case "switch":
		return false
	case "user":
		v := value.Value.([]string)
		if len(v) == 0 || (len(v) == 1 && v[0] == "") {
			return true
		}
		return false
	case "file":
		return false
	case "lookup":
		if len(value.Value.(string)) == 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func findItems(db, datastoreID string, query bson.M) []*Item {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result []*Item
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("findItems", err.Error())
		return nil
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item Item
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("findItems", err.Error())
			return nil
		}
		result = append(result, &item)
	}

	return result
}

// deleteItem 删除普通台账数据
func deleteItem(db, datastoreID, itemID, userID, lang, domain string, owners []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 开启事务
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("deleteItem", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("deleteItem", err.Error())
		return err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// 添加删除履历
		i, e := getItem(db, itemID, datastoreID, owners)
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				return errors.New("データが存在しないか、データを削除する権限がありません")
			}
			utils.ErrorLog("deleteItem", e.Error())
			return e
		}

		hs := NewHistory(db, userID, datastoreID, lang, domain, sc, nil)

		err := hs.Add("1", itemID, i.ItemMap)
		if err != nil {
			utils.ErrorLog("deleteItem", err.Error())
			return err
		}

		objectID, err := primitive.ObjectIDFromHex(itemID)
		if err != nil {
			utils.ErrorLog("deleteItem", err.Error())
			return err
		}

		query := bson.M{
			"_id": objectID,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("DeleteItem", fmt.Sprintf("query: [ %s ]", queryJSON))
		// 删除台账数据
		if _, err := c.DeleteOne(sc, query); err != nil {
			utils.ErrorLog("deleteItem", err.Error())
			return err
		}

		err = hs.Compare("1", nil)
		if err != nil {
			utils.ErrorLog("deleteItem", err.Error())
			return err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("deleteItem", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				utils.ErrorLog("deleteItem", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("deleteItem", err.Error())
		return err
	}

	session.EndSession(ctx)

	return nil
}

// getMonthLastDay  获取当前月份的最后一天
func getMonthLastDay(date time.Time) (day string) {
	// 年月日取得
	years := date.Year()
	month := date.Month()

	// 月末日取得
	lastday := 0
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			lastday = 30
		} else {
			lastday = 31
		}
	} else {
		if ((years%4) == 0 && (years%100) != 0) || (years%400) == 0 {
			lastday = 29
		} else {
			lastday = 28
		}
	}

	return strconv.Itoa(lastday)
}

// DownloadItems 下载台账数据
func SwkDownloadItems(db string, params ItemsParam, stream item.ItemService_DownloadStream, downloadInfo *journal.FindDownloadSettingResponse) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(params.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       params.AppID,
		"datastore_id": params.DatastoreID,
	}

	ds, err := getDatastore(db, params.DatastoreID)
	if err != nil {
		utils.ErrorLog("SwkDownloadItems", err.Error())
		return err
	}

	skip := (params.PageIndex - 1) * params.PageSize
	limit := params.PageSize

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

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

	for _, f := range downloadInfo.FieldRule {
		if f.SettingMethod != "1" && f.FieldType != "function" {
			project["items."+f.FieldId] = "$items." + f.FieldId
		}
		if f.FieldType == "function" {
			// 生成最终的嵌套 JSON
			finalJson := generateOptimizedJson(f.FieldConditions)

			jsonData, err := json.MarshalIndent(finalJson, "", "  ")
			if err != nil {
				log.Fatalf("Error marshaling JSON: %v", err)
			}
			var formula bson.M
			err = json.Unmarshal([]byte(string(jsonData)), &formula)
			if err != nil {
				utils.ErrorLog("DownloadItems", err.Error())
				return err
			}
			project["items."+f.FieldId+".value"] = formula
			project["items."+f.FieldId+".data_type"] = f.FieldType
			continue
		}

	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("SwkDownloadItems", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)
	opt.SetBatchSize(int32(params.PageSize))

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("SwkDownloadItems", err.Error())
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var it Item
		err := cur.Decode(&it)
		if err != nil {
			utils.ErrorLog("SwkDownloadItems", err.Error())
			return err
		}

		if err := stream.Send(&item.DownloadResponse{Item: it.ToProto()}); err != nil {
			utils.ErrorLog("SwkDownloadItems", err.Error())
			return err
		}
	}

	return nil
}

// 构建fieldCondition的条件
func buildOptimizedCondition(fieldCondition *journal.FieldCondition) map[string]interface{} {
	var conditions []map[string]interface{}

	// 遍历每个 fieldGroup，生成条件
	for _, group := range fieldCondition.FieldGroups {
		var groupConditions []map[string]interface{}
		for _, fieldCon := range group.FieldCons {
			// 根据 con_operator 的值判断是 eq 还是 ne
			var operator string
			if fieldCon.ConOperator == "{\"$eq\":[\"a\",\"b\"]}" {
				operator = "$eq"
			} else if fieldCon.ConOperator == "{\"$ne\":[\"a\",\"b\"]}" {
				operator = "$ne"
			}

			// 添加条件到 groupConditions 中
			groupConditions = append(groupConditions, map[string]interface{}{
				operator: []interface{}{
					"$items." + fieldCon.ConField + ".value", // 动态生成字段名
					fieldCon.ConValue,
				},
			})
		}

		// 用 $and 或 $or 将多个条件组合起来
		if group.Type == "and" {
			conditions = append(conditions, map[string]interface{}{"$and": groupConditions})
		} else if group.Type == "or" {
			conditions = append(conditions, map[string]interface{}{"$or": groupConditions})
		}
	}

	// 返回生成的条件结构
	return map[string]interface{}{
		"if":   conditions,                                      // 这里是所有条件的集合，按顺序连接
		"then": "$items." + fieldCondition.ThenValue + ".value", // 使用动态生成的字段值
	}
}

// 生成 JSON 结构
func generateOptimizedJson(fieldConditions []*journal.FieldCondition) map[string]interface{} {
	var previousElseValue interface{}

	// 用来保存生成的 branches
	var branches []map[string]interface{}

	// 遍历所有 fieldConditions，依次生成条件
	for i := len(fieldConditions) - 1; i >= 0; i-- { // 从后往前处理，确保正确的顺序
		condition := buildOptimizedCondition(fieldConditions[i])

		// 将前一个条件的 else 作为当前条件的 else
		if previousElseValue != nil {
			condition["else"] = previousElseValue
		} else {
			// 第一个条件的 else 就是当前的 elseValue
			condition["else"] = fieldConditions[i].ElseValue
		}

		// 更新 previousElseValue 为当前条件
		previousElseValue = condition

		// 将当前条件加入到 branches 中
		branches = append([]map[string]interface{}{
			{
				"case": condition["if"],
				"then": condition["then"],
			},
		}, branches...)
	}

	// 最终返回根节点的 $switch，指向第一条条件
	return map[string]interface{}{
		"$switch": map[string]interface{}{
			"branches": branches,
			"default":  "default_value", // 添加默认值字段及elsevalue
		},
	}
}
