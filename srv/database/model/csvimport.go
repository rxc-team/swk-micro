package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

func getFieldMap(db, appID string) (map[string][]Field, error) {
	param := &FindAppFieldsParam{
		AppID:         appID,
		InvalidatedIn: "true",
	}
	fields, err := FindAppFields(db, param)
	if err != nil {
		utils.ErrorLog("getFieldMap", err.Error())
		return nil, err
	}
	var ds string
	var fs []Field
	result := make(map[string][]Field)
	for index, f := range fields {
		if index == 0 {
			ds = f.DatastoreID
			fs = append(fs, f)

			if len(fields) == 1 {
				result[ds] = fs
			}
			continue
		}

		if len(fields)-1 == index {
			if ds == f.DatastoreID {
				fs = append(fs, f)
				result[ds] = fs
			} else {
				result[ds] = fs
				fs = nil
				ds = f.DatastoreID
				fs = append(fs, f)
				result[ds] = fs
			}
			continue
		}

		if ds == f.DatastoreID {
			fs = append(fs, f)
			continue
		}

		result[ds] = fs
		fs = nil
		ds = f.DatastoreID
		fs = append(fs, f)
	}

	return result, nil
}

func insertAttachData(client *mongo.Client, sc mongo.SessionContext, p AttachParam) error {

	data := make(map[string][]*Item)

	for _, it := range p.Items {
		it.ID = primitive.NewObjectID()
		it.ItemID = it.ID.Hex()
		it.Status = "1"
		it.Owners = p.Owners
		data[it.DatastoreID] = append(data[it.DatastoreID], it)
	}

	for datastoreID, items := range data {
		c := client.Database(database.GetDBName(p.DB)).Collection(GetItemCollectionName(datastoreID))

		step := len(items)
		autoList := make(map[string][]string)

		allFields := p.FileMap[datastoreID]
		for _, f := range allFields {
			if f.FieldType == "autonum" {
				list, err := autoNumListWithSession(sc, p.DB, &f, step)
				if err != nil {
					utils.ErrorLog("insertAttachData", err.Error())
					return err
				}
				autoList[f.FieldID] = list
			}

		}

		var insertData []interface{}
		for index, it := range items {
			for _, f := range allFields {
				if f.FieldType == "autonum" {
					nums := autoList[f.FieldID]
					it.ItemMap[f.FieldID] = &Value{
						DataType: "autonum",
						Value:    nums[index],
					}
				}

				addEmptyData(it.ItemMap, f)
			}

			insertData = append(insertData, it)
		}
		_, err := c.InsertMany(sc, insertData)
		if err != nil {
			utils.ErrorLog("insertAttachData", err.Error())
			return err
		}
	}

	return nil
}

func ImportItem(ctx context.Context, stream item.ItemService_ImportItemStream) error {

	fieldMap := make(map[string][]Field)
	dsMap := make(map[string]string)

	var meta *item.ImportMetaData
	var dataList []*Item
	var attachItems []*Item
	var oldItemList []primitive.ObjectID
	var current int64 = 0

	defer stream.Close()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		status := req.GetStatus()
		if status == item.SendStatus_COMPLETE {
			if len(dataList) > 0 {
				// 如果没有设置metadata，将直接返回
				if meta == nil {
					return errors.New("not set meta data")
				}

				// 获取变更前的条件
				query := bson.M{
					"_id": bson.M{
						"$in": oldItemList,
					},
					"owners": bson.M{
						"$in": meta.GetUpdateOwners(),
					},
				}

				// 获取变更前的记录
				oldItems := findItems(meta.GetDatabase(), meta.GetDatastoreId(), query)

				// 执行数据处理操作
				err := dataExec(ctx, meta, fieldMap, dsMap, dataList, attachItems, oldItems, stream)
				if err != nil {
					return err
				}
			}

			break
		}

		// 判断传入的类型
		m := req.GetMeta()
		// 如果m不等于空，则说明传入的是m
		if m != nil {
			// 设置meta的值
			meta = m

			// 根据所有台账，获取所有字段数据
			fm, err := getFieldMap(meta.GetDatabase(), meta.GetAppId())
			if err != nil {
				return err
			}

			fieldMap = fm

			// 获取所有台账
			dsList, e := FindDatastores(meta.GetDatabase(), meta.GetAppId(), "", "", "")
			if e != nil {
				utils.ErrorLog("ImportItem", e.Error())
				return err
			}

			for _, d := range dsList {
				dsMap[d.ApiKey] = d.DatastoreID
			}

			// 直接进入下一次循环
			continue
		}

		data := req.GetData()
		// 如果data不等于空，则说明传入的是data
		if data != nil {
			current++

			// 读取一条数据
			items := make(map[string]*Value, len(data.GetItems().Items))
			itemId := ""
			for key, item := range data.GetItems().GetItems() {
				if key != "id" {
					items[key] = &Value{
						DataType: item.DataType,
						Value:    GetValueFromProto(item),
					}
				} else {
					itemId = GetValueFromProto(item).(string)
				}
			}

			if len(itemId) > 0 {
				objectID, _ := primitive.ObjectIDFromHex(itemId)
				oldItemList = append(oldItemList, objectID)
			}

			dataList = append(dataList, &Item{
				ItemID:      itemId,
				AppID:       meta.GetAppId(),
				DatastoreID: meta.GetDatastoreId(),
				ItemMap:     items,
				CreatedAt:   time.Now(),
				CreatedBy:   meta.GetWriter(),
				UpdatedAt:   time.Now(),
				UpdatedBy:   meta.GetWriter(),
			})

			// 读取一条数据的附加数据
			for _, it := range data.GetAttachItems() {

				items := make(map[string]*Value, len(it.Items))
				for key, item := range it.GetItems() {
					items[key] = &Value{
						DataType: item.DataType,
						Value:    GetValueFromProto(item),
					}
				}

				attachItems = append(attachItems, &Item{
					AppID:       meta.GetAppId(),
					DatastoreID: it.GetDatastoreId(),
					ItemMap:     items,
					CreatedAt:   time.Now(),
					CreatedBy:   meta.GetWriter(),
					UpdatedAt:   time.Now(),
					UpdatedBy:   meta.GetWriter(),
				})
			}

		}

		if current%500 == 0 {
			// 如果没有设置metadata，将直接返回
			if meta == nil {
				return errors.New("not set meta data")
			}

			// 获取变更前的条件
			query := bson.M{
				"_id": bson.M{
					"$in": oldItemList,
				},
				"owners": bson.M{
					"$in": meta.GetUpdateOwners(),
				},
			}

			// 获取变更前的记录
			oldItems := findItems(meta.GetDatabase(), meta.GetDatastoreId(), query)

			// 执行数据处理操作
			err := dataExec(ctx, meta, fieldMap, dsMap, dataList, attachItems, oldItems, stream)
			if err != nil {
				return err
			}

			dataList = dataList[:0]
			attachItems = attachItems[:0]
			oldItemList = oldItemList[:0]
		}
	}

	return nil
}

func dataExec(ctx context.Context, meta *item.ImportMetaData, fieldMap map[string][]Field, dsMap map[string]string, dataList []*Item, attachItems []*Item, oldItems []*Item, stream item.ItemService_ImportItemStream) error {

	client := database.New()
	c := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(meta.GetDatastoreId()))

	var importErrors []*item.Error
	var cxModels []mongo.WriteModel

	// 当前行号
	firstLine := int64(dataList[0].ItemMap["index"].Value.(float64))
	// 获取当前最后行号
	lastLine := int64(dataList[len(dataList)-1].ItemMap["index"].Value.(float64))

	// 获取插入的条数
	insert := len(dataList) - len(oldItems)
	autoList := make(map[string][]string)

	callback := func(sc mongo.SessionContext) (interface{}, error) {

		// hs := NewHistory(meta.Database, meta.Writer, meta.DatastoreId, meta.LangCd, meta.Domain, sc, fieldMap[meta.DatastoreId])

		var result *mongo.BulkWriteResult

		for _, f := range fieldMap[meta.DatastoreId] {
			if f.FieldType == "autonum" {
				list, err := autoNumListWithSession(sc, meta.GetDatabase(), &f, insert)
				if err != nil {
					if err.Error() != "(WriteConflict) WriteConflict" {
						utils.ErrorLog("ImportItem", err.Error())
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine: firstLine,
							LastLine:  lastLine,
							ErrorMsg:  err.Error(),
						})
					}

					return nil, err

				}

				autoList[f.FieldID] = list
			}
		}

		for index, it := range dataList {
			// 获取当前行号
			line := int64(it.ItemMap["index"].Value.(float64))
			delete(it.ItemMap, "index")

			// 判断itemid是否传入
			if it.ItemID == "" {
				// 没有找到必须字段的情况下，直接插入数据
				it.ID = primitive.NewObjectID()
				it.ItemID = it.ID.Hex()
				it.Status = "1"
				it.CheckStatus = "0"
				it.Owners = meta.GetOwners()
				if owner, ok := it.ItemMap["owner"]; ok {
					it.Owners = []string{owner.Value.(string)}
					delete(it.ItemMap, "owner")
				}

				// 删除临时数据
				delete(it.ItemMap, "action")

				for _, f := range fieldMap[meta.GetDatastoreId()] {
					if f.FieldType == "autonum" {
						nums := autoList[f.FieldID]
						it.ItemMap[f.FieldID] = &Value{
							DataType: "autonum",
							Value:    nums[index],
						}
						continue
					}

					if f.FieldID == "sakuseidate" {
						it.ItemMap[f.FieldID] = &Value{
							DataType: "date",
							Value:    time.Now(),
						}
						continue
					}

					//  添加空数据
					addEmptyData(it.ItemMap, f)
				}

				// err := hs.Add(cast.ToString(index+1), it.ItemID, nil)
				// if err != nil {
				// 	utils.ErrorLog("MappingImport", err.Error())
				// 	// 返回错误信息
				// 	importErrors = append(importErrors, &item.Error{
				// 		FirstLine:   firstLine,
				// 		CurrentLine: line,
				// 		LastLine:    lastLine,
				// 		ErrorMsg:    err.Error(),
				// 	})
				// 	return nil, err
				// }

				insertCxModel := mongo.NewInsertOneModel()
				insertCxModel.SetDocument(it)
				cxModels = append(cxModels, insertCxModel)

				// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
				// if err != nil {
				// 	// 返回错误信息
				// 	importErrors = append(importErrors, &item.Error{
				// 		FirstLine:   firstLine,
				// 		CurrentLine: line,
				// 		LastLine:    lastLine,
				// 		ErrorMsg:    err.Error(),
				// 	})
				// 	return nil, err
				// }

				continue
			} else {
				// action
				action := "update"
				if val, exist := it.ItemMap["action"]; exist {
					action = val.Value.(string)
					// 删除临时数据ID
					delete(it.ItemMap, "action")
				}
				if action == "update" {
					// 查询变更前的数据
					oldItem := getOldItem(oldItems, it.ItemID)

					if oldItem == nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							LastLine:    lastLine,
							CurrentLine: line,
							ErrorMsg:    "データが存在しないか、データを変更する権限がありません",
						})

						continue
					}

					// err := hs.Add(cast.ToString(index+1), it.ItemID, oldItem.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 更新条件
					objectID, _ := primitive.ObjectIDFromHex(it.ItemID)
					query := bson.M{
						"_id": objectID,
					}

					change := bson.M{
						"updated_at": it.UpdatedAt,
						"updated_by": it.UpdatedBy,
					}

					delete(it.ItemMap, meta.GetKey())

					if owner, ok := it.ItemMap["owner"]; ok {
						change["owners"] = []string{owner.Value.(string)}
						delete(it.ItemMap, "owner")
					}

					// 自增字段不更新
					for _, f := range fieldMap[meta.GetDatastoreId()] {
						if f.FieldType == "autonum" {
							if oldItem != nil {
								delete(oldItem.ItemMap, f.FieldID)
							}
							delete(it.ItemMap, f.FieldID)
						}
						_, ok := it.ItemMap[f.FieldID]
						// 需要进行自算的情况
						if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
							if f.SelfCalculate == "add" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o + n
								continue
							}
							if f.SelfCalculate == "sub" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o - n
								continue
							}
						}
					}

					for k, v := range it.ItemMap {
						change["items."+k] = v
					}

					update := bson.M{"$set": change}
					upCxModel := mongo.NewUpdateOneModel()
					upCxModel.SetFilter(query)
					upCxModel.SetUpdate(update)
					upCxModel.SetUpsert(false)
					cxModels = append(cxModels, upCxModel)

					// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					continue
				}
				if action == "image" {
					// 查询变更前的数据
					oldItem := getOldItem(oldItems, it.ItemID)

					if oldItem == nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							LastLine:    lastLine,
							CurrentLine: line,
							ErrorMsg:    "データが存在しないか、データを変更する権限がありません",
						})

						continue
					}

					// err := hs.Add(cast.ToString(index+1), it.ItemID, oldItem.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 更新条件
					objectID, _ := primitive.ObjectIDFromHex(it.ItemID)
					query := bson.M{
						"_id": objectID,
					}

					change := bson.M{
						"updated_at": it.UpdatedAt,
						"updated_by": it.UpdatedBy,
					}

					delete(it.ItemMap, meta.GetKey())

					if owner, ok := it.ItemMap["owner"]; ok {
						change["owners"] = []string{owner.Value.(string)}
						delete(it.ItemMap, "owner")
					}

					// 自增字段不更新
					for _, f := range fieldMap[meta.GetDatastoreId()] {
						if f.FieldType == "autonum" {
							if oldItem != nil {
								delete(oldItem.ItemMap, f.FieldID)
							}
							delete(it.ItemMap, f.FieldID)
						}
					}

					for key, value := range it.ItemMap {
						if oldItem != nil {
							if ovalue, ok := oldItem.ItemMap[key]; ok {
								var new []File
								err := json.Unmarshal([]byte(value.Value.(string)), &new)
								if err != nil {
									continue
								}
								var old []File
								err = json.Unmarshal([]byte(ovalue.Value.(string)), &old)
								if err != nil {
									continue
								}

								old = append(old, new...)

								fs, err := json.Marshal(old)
								if err != nil {
									continue
								}

								it.ItemMap[key] = &Value{
									DataType: value.DataType,
									Value:    string(fs),
								}

								change["items."+key] = Value{
									DataType: value.DataType,
									Value:    string(fs),
								}
							} else {
								var new []File
								err := json.Unmarshal([]byte(value.Value.(string)), &new)
								if err != nil {
									continue
								}
								var old []File

								old = append(old, new...)

								fs, err := json.Marshal(old)
								if err != nil {
									continue
								}

								change["items."+key] = Value{
									DataType: value.DataType,
									Value:    string(fs),
								}
							}
							continue
						}

						var new []File
						err := json.Unmarshal([]byte(value.Value.(string)), &new)
						if err != nil {
							continue
						}
						var old []File

						old = append(old, new...)

						fs, err := json.Marshal(old)
						if err != nil {
							continue
						}

						change["items."+key] = Value{
							DataType: value.DataType,
							Value:    string(fs),
						}

					}

					update := bson.M{"$set": change}
					upCxModel := mongo.NewUpdateOneModel()
					upCxModel.SetFilter(query)
					upCxModel.SetUpdate(update)
					upCxModel.SetUpsert(false)
					cxModels = append(cxModels, upCxModel)

					// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					continue
				}
				// 判断是否是契约登录
				if action == "contract-insert" {
					// 没有找到必须字段的情况下，直接插入数据
					oid, err := primitive.ObjectIDFromHex(it.ItemID)
					if err != nil {
						utils.ErrorLog("ImportItem", err.Error())
						return nil, err
					}
					it.ID = oid
					it.Status = "1"
					it.CheckStatus = "0"
					it.Owners = meta.GetOwners()
					if owner, ok := it.ItemMap["owner"]; ok {
						it.Owners = []string{owner.Value.(string)}
						delete(it.ItemMap, "owner")
					}

					// err = hs.Add(cast.ToString(index+1), it.ItemID, nil)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 自增字段更新
					for _, f := range fieldMap[meta.GetDatastoreId()] {
						if f.FieldType == "autonum" {
							nums := autoList[f.FieldID]
							it.ItemMap[f.FieldID] = &Value{
								DataType: "autonum",
								Value:    nums[index],
							}
						}
					}

					insertCxModel := mongo.NewInsertOneModel()
					insertCxModel.SetDocument(it)
					cxModels = append(cxModels, insertCxModel)

					// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					continue
				}
				if action == "info-change" {
					// 变更前契约情报取得
					oldItem := getOldItem(oldItems, it.ItemID)
					if oldItem == nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							LastLine:    lastLine,
							CurrentLine: line,
							ErrorMsg:    "データが存在しないか、データを変更する権限がありません",
						})

						continue
					}

					// err := hs.Add(cast.ToString(index+1), it.ItemID, oldItem.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 自增字段不更新
					for _, f := range fieldMap[meta.GetDatastoreId()] {
						if f.FieldType == "autonum" {
							if oldItem != nil {
								delete(oldItem.ItemMap, f.FieldID)
							}
							delete(it.ItemMap, f.FieldID)
						}
						_, ok := it.ItemMap[f.FieldID]
						// 需要进行自算的情况
						if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
							if f.SelfCalculate == "add" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o + n
								continue
							}
							if f.SelfCalculate == "sub" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o - n
								continue
							}
						}
					}

					change := bson.M{
						"updated_at": it.UpdatedAt,
						"updated_by": it.UpdatedBy,
					}

					if owner, ok := it.ItemMap["owner"]; ok {
						change["owners"] = []string{owner.Value.(string)}
						delete(it.ItemMap, "owner")
					}

					// 循环契约情报数据对比变更
					for key, value := range it.ItemMap {
						change["items."+key] = value
					}

					// 契约情报变更参数编辑
					update := bson.M{"$set": change}
					objectID, e := primitive.ObjectIDFromHex(it.ItemID)
					if e != nil {
						utils.ErrorLog("ImportItem", e.Error())
						return nil, e
					}
					query := bson.M{
						"_id": objectID,
					}

					// 第一步：更新契约台账情报
					upCxModel := mongo.NewUpdateOneModel()
					upCxModel.SetFilter(query)
					upCxModel.SetUpdate(update)
					upCxModel.SetUpsert(false)
					cxModels = append(cxModels, upCxModel)

					// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					continue
				}
				if action == "debt-change" {
					// 支付，利息，偿还表取得
					dsPay := dsMap["paymentStatus"]
					dsInterest := dsMap["paymentInterest"]
					dsRepay := dsMap["repayment"]

					cpay := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(dsPay))
					cinter := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(dsInterest))
					crepay := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(dsRepay))

					// 变更前契约情报取得
					oldItem := getOldItem(oldItems, it.ItemID)
					if oldItem == nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							LastLine:    lastLine,
							CurrentLine: line,
							ErrorMsg:    "データが存在しないか、データを変更する権限がありません",
						})

						continue
					}

					// err := hs.Add(cast.ToString(index+1), it.ItemID, oldItem.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 自增字段不更新
					for _, f := range fieldMap[meta.GetDatastoreId()] {
						if f.FieldType == "autonum" {
							if oldItem != nil {
								delete(oldItem.ItemMap, f.FieldID)
							}
							delete(it.ItemMap, f.FieldID)
						}
						_, ok := it.ItemMap[f.FieldID]
						// 需要进行自算的情况
						if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
							if f.SelfCalculate == "add" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o + n
								continue
							}
							if f.SelfCalculate == "sub" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o - n
								continue
							}
						}
					}

					// 变更箇所数记录用
					changeCount := 0

					// 变更情报编辑
					change := bson.M{
						"updated_at": it.UpdatedAt,
						"updated_by": it.UpdatedBy,
					}

					if owner, ok := it.ItemMap["owner"]; ok {
						change["owners"] = []string{owner.Value.(string)}
						delete(it.ItemMap, "owner")
					}

					// 循环契约情报数据对比变更
					for key, value := range it.ItemMap {
						// 项目[百分比]不做更新
						if key != "percentage" {
							change["items."+key] = value
						}

						// 对比前后数据值
						if _, ok := oldItem.ItemMap[key]; ok {
							// 该项数据历史存在,判断历史与当前是否变更
							if compare(value, oldItem.ItemMap[key]) {
								changeCount++
							}
						} else {
							// 该项数据历史不存在,判断当前是否为空
							if value.Value == "" || value.Value == "[]" {
								continue
							}

							changeCount++
						}
					}

					// 契约情报变更参数编辑
					update := bson.M{"$set": change}
					objectID, e := primitive.ObjectIDFromHex(it.ItemID)
					if e != nil {
						utils.ErrorLog("ImportItem", e.Error())
						return nil, e
					}
					query := bson.M{
						"_id": objectID,
					}

					// 第一步：更新契约台账情报
					upCxModel := mongo.NewUpdateOneModel()
					upCxModel.SetFilter(query)
					upCxModel.SetUpdate(update)
					upCxModel.SetUpsert(false)
					cxModels = append(cxModels, upCxModel)

					// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 第二步：若字段有变更,添加契约历史情报和新旧履历情报
					if changeCount > 0 {
						// 追加履历特有的数据
						keiyakuno := GetValueFromModel(oldItem.ItemMap["keiyakuno"])

						/* ******************契约更新后根据契约番号删除以前的支付，利息，偿还的数据************* */
						querydel := bson.M{
							"items.keiyakuno.value": keiyakuno,
						}

						if _, err := cpay.DeleteMany(sc, querydel); err != nil {
							utils.ErrorLog("ImportItem", err.Error())
							return nil, err
						}

						if _, err := cinter.DeleteMany(sc, querydel); err != nil {
							utils.ErrorLog("ImportItem", err.Error())
							return nil, err
						}

						if _, err := crepay.DeleteMany(sc, querydel); err != nil {
							utils.ErrorLog("ImportItem", err.Error())
							return nil, err
						}

					}

					continue
				}
				if action == "midway-cancel" {
					// 支付，利息，偿还表取得
					dsPay := dsMap["paymentStatus"]
					dsInterest := dsMap["paymentInterest"]
					dsRepay := dsMap["repayment"]

					cpay := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(dsPay))
					cinter := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(dsInterest))
					crepay := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(dsRepay))

					// 变更前契约情报取得
					oldItem := getOldItem(oldItems, it.ItemID)
					if oldItem == nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							LastLine:    lastLine,
							CurrentLine: line,
							ErrorMsg:    "データが存在しないか、データを変更する権限がありません",
						})

						continue
					}

					// err := hs.Add(cast.ToString(index+1), it.ItemID, oldItem.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 自增字段不更新
					for _, f := range fieldMap[meta.GetDatastoreId()] {
						if f.FieldType == "autonum" {
							if oldItem != nil {
								delete(oldItem.ItemMap, f.FieldID)
							}
							delete(it.ItemMap, f.FieldID)
						}
						_, ok := it.ItemMap[f.FieldID]
						// 需要进行自算的情况
						if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
							if f.SelfCalculate == "add" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o + n
								continue
							}
							if f.SelfCalculate == "sub" {
								o := GetNumberValue(oldItem.ItemMap[f.FieldID])
								n := GetNumberValue(it.ItemMap[f.FieldID])
								it.ItemMap[f.FieldID].Value = o - n
								continue
							}
						}
					}

					// 变更箇所数记录用
					changeCount := 0

					// 追加契约状态
					it.ItemMap["status"] = &Value{
						DataType: "options",
						Value:    "cancel",
					}

					// 契约变更情报编辑
					change := bson.M{
						"updated_at": it.UpdatedAt,
						"updated_by": it.UpdatedBy,
					}

					if owner, ok := it.ItemMap["owner"]; ok {
						change["owners"] = []string{owner.Value.(string)}
						delete(it.ItemMap, "owner")
					}

					// 循环契约情报数据对比变更
					for key, value := range it.ItemMap {
						change["items."+key] = value
						// 对比前后数据值
						if _, ok := oldItem.ItemMap[key]; ok {
							// 该项数据历史存在,判断历史与当前是否变更
							if compare(value, oldItem.ItemMap[key]) {
								changeCount++
							}
						} else {
							// 该项数据历史不存在,判断当前是否为空
							if value.Value == "" || value.Value == "[]" {
								continue
							}

							changeCount++
						}
					}

					// 契约情报变更参数编辑
					update := bson.M{"$set": change}
					objectID, e := primitive.ObjectIDFromHex(it.ItemID)
					if e != nil {
						utils.ErrorLog("ImportItem", e.Error())
						return nil, e
					}
					query := bson.M{
						"_id": objectID,
					}

					// 第一步：更新契约台账情报
					upCxModel := mongo.NewUpdateOneModel()
					upCxModel.SetFilter(query)
					upCxModel.SetUpdate(update)
					upCxModel.SetUpsert(false)
					cxModels = append(cxModels, upCxModel)

					// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 第二步：若字段有变更,添加契约历史情报和新旧履历情报
					if changeCount > 0 {
						// 将契约番号变成lookup类型
						keiyakuno := GetValueFromModel(oldItem.ItemMap["keiyakuno"])

						/* ******************契约更新后根据契约番号删除以前的支付，利息，偿还的数据************* */
						querydel := bson.M{
							"items.keiyakuno.value": keiyakuno,
						}

						if _, err := cpay.DeleteMany(sc, querydel); err != nil {
							utils.ErrorLog("ImportItem", err.Error())
							return nil, err
						}

						if _, err := cinter.DeleteMany(sc, querydel); err != nil {
							utils.ErrorLog("ImportItem", err.Error())
							return nil, err
						}

						if _, err := crepay.DeleteMany(sc, querydel); err != nil {
							utils.ErrorLog("ImportItem", err.Error())
							return nil, err
						}
					}

					continue
				}
				if action == "contract-expire" {
					// 取出是否需要更新偿还台账数据的flag
					hasChange := it.ItemMap["hasChange"].Value
					delete(it.ItemMap, "hasChange")

					// 变更前契约情报取得
					oldItem := getOldItem(oldItems, it.ItemID)
					if oldItem == nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							LastLine:    lastLine,
							CurrentLine: line,
							ErrorMsg:    "データが存在しないか、データを変更する権限がありません",
						})

						continue
					}

					// err := hs.Add(cast.ToString(index+1), it.ItemID, oldItem.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					// 变更箇所数记录用
					changeCount := 0

					// 契约变更情报编辑
					change := bson.M{
						"updated_at": it.UpdatedAt,
						"updated_by": it.UpdatedBy,
					}

					if owner, ok := it.ItemMap["owner"]; ok {
						change["owners"] = []string{owner.Value.(string)}
						delete(it.ItemMap, "owner")
					}

					// 循环契约情报数据对比变更
					for key, value := range it.ItemMap {
						change["items."+key] = value
						// 对比前后数据值
						if _, ok := oldItem.ItemMap[key]; ok {
							// 该项数据历史存在,判断历史与当前是否变更
							if compare(value, oldItem.ItemMap[key]) {
								changeCount++
							}
						} else {
							// 该项数据历史不存在,判断当前是否为空
							if value.Value == "" || value.Value == "[]" {
								continue
							}
							changeCount++
						}
					}

					// 契约情报变更参数编辑
					update := bson.M{"$set": change}
					objectID, e := primitive.ObjectIDFromHex(it.ItemID)
					if e != nil {
						utils.ErrorLog("ImportItem", e.Error())
						return nil, e
					}
					query := bson.M{
						"_id": objectID,
					}

					// 更新契约台账情报
					// 第一步：更新契约台账情报
					upCxModel := mongo.NewUpdateOneModel()
					upCxModel.SetFilter(query)
					upCxModel.SetUpdate(update)
					upCxModel.SetUpsert(false)
					cxModels = append(cxModels, upCxModel)

					// err = hs.Compare(cast.ToString(index+1), it.ItemMap)
					// if err != nil {
					// 	importErrors = append(importErrors, &item.Error{
					// 		FirstLine:   firstLine,
					// 		CurrentLine: line,
					// 		LastLine:    lastLine,
					// 		ErrorMsg:    err.Error(),
					// 	})
					// 	return nil, err
					// }

					if changeCount > 0 {
						if hasChange == "1" {
							keiyakuno := GetValueFromModel(oldItem.ItemMap["keiyakuno"])
							dsRepay := dsMap["repayment"]
							crepay := client.Database(database.GetDBName(meta.GetDatabase())).Collection(GetItemCollectionName(dsRepay))
							/* ******************契约更新后根据契约番号删除以前的偿还的数据************* */
							querydel := bson.M{
								"items.keiyakuno.value": keiyakuno,
							}
							if _, err := crepay.DeleteMany(sc, querydel); err != nil {
								utils.ErrorLog("ImportItem", err.Error())
								return nil, err
							}
						}
					}
					continue
				}
			}
		}

		if len(cxModels) > 0 {
			res, err := c.BulkWrite(sc, cxModels)
			if err != nil {
				isDuplicate := mongo.IsDuplicateKeyError(err)
				if isDuplicate {
					bke, ok := err.(mongo.BulkWriteException)
					if !ok {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine: firstLine,
							LastLine:  lastLine,
							ErrorMsg:  err.Error(),
						})

						utils.ErrorLog("ImportItem", err.Error())
						return nil, err
					}
					errInfo := bke.WriteErrors[0]
					em := errInfo.Message
					// 判断是適用開始年月日和リース期間的组合
					if strings.Contains(em, "items.baseym.value") && strings.Contains(em, "items.leaseperiod.value") {
						// 使用正则表达式匹配
						re := regexp.MustCompile(`items\.baseym\.value:\s*new Date\((\d+)\),\s*items\.leaseperiod\.value:\s*(\d+\.\d+)`)
						matches := re.FindStringSubmatch(em[strings.LastIndex(em, "dup key"):])

						if len(matches) > 0 {
							baseymValue := matches[1]
							leaseperiodValue := matches[2]
							num, _ := strconv.ParseInt(baseymValue, 10, 64)
							baseymDate := time.Unix(num/1000, 0).Format("2006-01-02")

							// 返回错误信息
							importErrors = append(importErrors, &item.Error{
								FirstLine: firstLine,
								LastLine:  lastLine,
								ErrorMsg:  fmt.Sprintf("プライマリキーの重複エラー、重複値は[%s][%s]です。", "適用開始年月日:"+baseymDate, "リース期間:"+leaseperiodValue),
							})

							utils.ErrorLog("ImportItem", errInfo.Message)
							return nil, errInfo
						}
					}
					// 利子率コード重复报错
					if strings.Contains(em, "items.ritsucode.value") {
						// 使用正则表达式匹配
						re := regexp.MustCompile(`items\.ritsucode\.value:\s*"([^"]+)"`)
						matches := re.FindStringSubmatch(em[strings.LastIndex(em, "dup key"):])

						if len(matches) > 0 {
							ritsucodeValue := matches[1]

							// 返回错误信息
							importErrors = append(importErrors, &item.Error{
								FirstLine: firstLine,
								LastLine:  lastLine,
								ErrorMsg:  fmt.Sprintf("「利子率コード」が重複している, 重複値は[%s]です。", ritsucodeValue),
							})

							utils.ErrorLog("ImportItem", errInfo.Message)
							return nil, errInfo
						}
					}
					// 契约番号重复报错
					if strings.Contains(em, "items.keiyakuno.value") {
						// 使用正则表达式匹配
						re := regexp.MustCompile(`items\.keiyakuno\.value:\s*"([^"]+)"`)
						matches := re.FindStringSubmatch(em[strings.LastIndex(em, "dup key"):])

						if len(matches) > 0 {
							keiyakunoValue := matches[1]

							// 返回错误信息
							importErrors = append(importErrors, &item.Error{
								FirstLine: firstLine,
								LastLine:  lastLine,
								ErrorMsg:  fmt.Sprintf("「契約番号」が重複している, 重複値は[%s]です。", keiyakunoValue),
							})

							utils.ErrorLog("ImportItem", errInfo.Message)
							return nil, errInfo
						}
					}
					values := utils.FieldMatch(`"([^\"]+)"`, em[strings.LastIndex(em, "dup key"):])
					for i, v := range values {
						values[i] = strings.Trim(v, `"`)
					}
					fields := utils.FieldMatch(`field_[0-9a-z]{3}`, em[strings.LastIndex(em, "dup key"):])
					// 返回错误信息
					importErrors = append(importErrors, &item.Error{
						FirstLine: firstLine,
						LastLine:  lastLine,
						ErrorMsg:  fmt.Sprintf("プライマリキーの重複エラー、API-KEY[%s],重複値は[%s]です。", strings.Join(fields, ","), strings.Join(values, ",")),
					})

					utils.ErrorLog("ImportItem", errInfo.Message)
					return nil, errInfo
				}

				utils.ErrorLog("ImportItem", err.Error())
				return nil, err
			}

			result = res
		}

		// 提交履历
		// err := hs.Commit()
		// if err != nil {
		// 	utils.ErrorLog("ImportItem", err.Error())
		// 	// 返回错误信息
		// 	importErrors = append(importErrors, &item.Error{
		// 		FirstLine: firstLine,
		// 		LastLine:  lastLine,
		// 		ErrorMsg:  err.Error(),
		// 	})
		// 	return nil, err
		// }

		// 插入附加数据
		params := AttachParam{
			DB:      meta.GetDatabase(),
			DsMap:   dsMap,
			FileMap: fieldMap,
			Items:   attachItems,
			Owners:  meta.GetOwners(),
		}
		err := insertAttachData(client, sc, params)
		if err != nil {
			if err.Error() != "(WriteConflict) WriteConflict" {
				utils.ErrorLog("ImportItem", err.Error())
				// 返回错误信息
				importErrors = append(importErrors, &item.Error{
					FirstLine: firstLine,
					LastLine:  lastLine,
					ErrorMsg:  err.Error(),
				})
			}
			return nil, err
		}

		err = stream.Send(&item.ImportResponse{
			Status: item.Status_SUCCESS,
			Result: &item.ImportResult{
				Insert: result.InsertedCount,
				Modify: result.ModifiedCount,
			},
		})

		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	opts := &options.SessionOptions{}
	// 提交时间改为5分钟
	commitTime := 5 * time.Minute
	opts.SetDefaultMaxCommitTime(&commitTime)
	opts.SetDefaultReadConcern(readconcern.Snapshot())

	session, err := client.StartSession(opts)
	if err != nil {
		utils.ErrorLog("ImportItem", err.Error())
		// 返回错误信息
		importErrors = append(importErrors, &item.Error{
			FirstLine: firstLine,
			LastLine:  lastLine,
			ErrorMsg:  err.Error(),
		})

		err := stream.Send(&item.ImportResponse{
			Status: item.Status_FAILED,
			Result: &item.ImportResult{
				Errors: importErrors,
			},
		})

		if err != nil {
			return err
		}
	}

	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		utils.ErrorLog("ImportItem", err.Error())
		// 返回错误信息
		err := stream.Send(&item.ImportResponse{
			Status: item.Status_FAILED,
			Result: &item.ImportResult{
				Errors: importErrors,
			},
		})

		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
