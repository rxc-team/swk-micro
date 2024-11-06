package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

// MappingImport 导入更新台账数据
func MappingUpload(ctx context.Context, stream item.ItemService_MappingUploadStream) error {
	// 所有字段
	var allFields []Field
	// 必须字段
	var requriedFields []Field
	// 自动採番字段
	var autoFields []Field

	// 传入的部分参数
	var meta *item.MappingMetaData
	var dataList []*ChangeData
	var current int64 = 0

	defer stream.Close()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("EEEEEEEEEEE")
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

				// 执行数据处理操作
				err := dataHandler(ctx, meta, dataList, autoFields, requriedFields, allFields, stream)
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

			// 获取字段信息
			// 获取所有字段
			params := FindFieldsParam{
				AppID:       m.GetAppId(),
				DatastoreID: m.GetDatastoreId(),
			}

			allFields, err = FindFields(m.GetDatabase(), &params)
			if err != nil {
				utils.ErrorLog("MappingImport", err.Error())
				return errors.New("not found datastore fields")
			}

			for _, f := range allFields {
				if f.IsRequired {
					requriedFields = append(requriedFields, f)
				}
				if f.FieldType == "autonum" {
					autoFields = append(autoFields, f)
				}
			}

			// 直接进入下一次循环
			continue
		}

		data := req.GetData()
		// 如果data不等于空，则说明传入的是data
		if data != nil {
			current++

			// 添加值到对应的
			changes := make(map[string]*Value, len(data.GetChange()))
			for key, item := range data.GetChange() {
				changes[key] = &Value{
					DataType: item.DataType,
					Value:    GetValueFromProto(item),
				}
			}

			dataList = append(dataList, &ChangeData{
				Change: changes,
				Query:  data.GetQuery(),
				Index:  data.GetIndex(),
			})
		}

		if current%500 == 0 {
			// 如果没有设置metadata，将直接返回
			if meta == nil {
				return errors.New("not set meta data")
			}

			// 执行数据处理操作
			err := dataHandler(ctx, meta, dataList, autoFields, requriedFields, allFields, stream)
			if err != nil {
				return err
			}

			dataList = dataList[:0]
		}
	}

	return nil
}

func dataHandler(ctx context.Context, meta *item.MappingMetaData, dataList []*ChangeData, autoFields []Field, requriedFields []Field, allFields []Field, stream item.ItemService_MappingUploadStream) error {
	client := database.New()

	c := client.Database(database.GetDBName(meta.Database)).Collection(GetItemCollectionName(meta.GetDatastoreId()))

	firstLine := dataList[0].Index
	lastLine := dataList[len(dataList)-1].Index

	// 返回错误
	var importErrors []*item.Error
	// 执行任务
	var cxModels []mongo.WriteModel

	callback := func(sc mongo.SessionContext) (interface{}, error) {
		var result *mongo.BulkWriteResult

		hs := NewHistory(meta.GetDatabase(), meta.GetWriter(), meta.GetDatastoreId(), meta.GetLangCd(), meta.GetDomain(), sc, allFields)

		// 新规作成的场合
		if meta.MappingType == "insert" {
			step := len(dataList)
			autoList := make(map[string][]string)

			for _, f := range autoFields {
				list, err := autoNumListWithSession(sc, meta.Database, &f, step)
				if err != nil {
					if err.Error() != "(WriteConflict) WriteConflict" {
						utils.ErrorLog("MappingImport", err.Error())
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine: firstLine,
							LastLine:  lastLine,
							FieldId:   f.FieldID,
							FieldName: f.FieldName,
							ErrorMsg:  err.Error(),
						})
					}

					return nil, err
				} else {
					autoList[f.FieldID] = list
				}
			}

			for in, d := range dataList {
				// 获取当前行号
				line := d.Index
				// 数据
				dataItem := Item{
					AppID:       meta.GetAppId(),
					DatastoreID: meta.GetDatastoreId(),
					ItemMap:     d.Change,
					Owners:      meta.Owners,
					CreatedAt:   time.Now(),
					CreatedBy:   meta.Writer,
					UpdatedAt:   time.Now(),
					UpdatedBy:   meta.Writer,
				}
				dataItem.ID = primitive.NewObjectID()
				dataItem.ItemID = dataItem.ID.Hex()
				dataItem.Status = "1"
				dataItem.CheckStatus = "0"

				err := hs.Add(cast.ToString(in+1), dataItem.ID.Hex(), nil)
				if err != nil {
					utils.ErrorLog("MappingImport", err.Error())
					// 返回错误信息
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				// 没有必须字段的情况下直接插入数据
				if len(requriedFields) == 0 {
					for _, f := range allFields {
						if f.FieldType == "autonum" {
							nums := autoList[f.FieldID]
							dataItem.ItemMap[f.FieldID] = &Value{
								DataType: "autonum",
								Value:    nums[in],
							}
							continue
						}
						//  添加空数据
						addEmptyData(dataItem.ItemMap, f)
					}

					queryJSON, _ := json.Marshal(dataItem)
					utils.DebugLog("MappingImport", fmt.Sprintf("item: [ %s ]", queryJSON))

					insertCxModel := mongo.NewInsertOneModel()
					insertCxModel.SetDocument(dataItem)
					cxModels = append(cxModels, insertCxModel)

					err = hs.Compare(cast.ToString(in+1), dataItem.ItemMap)
					if err != nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							CurrentLine: line,
							LastLine:    lastLine,
							ErrorMsg:    err.Error(),
						})
						return nil, err
					}

					continue
				}

				// 有必须字段的情况下，先判断是否必须字段是否有值
				for _, f := range allFields {
					if f.IsRequired {
						if value, ok := dataItem.ItemMap[f.FieldID]; ok {
							if isEmptyValue(value) {
								importErrors = append(importErrors, &item.Error{
									FirstLine:   firstLine,
									CurrentLine: line,
									LastLine:    lastLine,
									FieldId:     f.FieldID,
									FieldName:   f.FieldName,
									ErrorMsg:    "このフィールドは必須フィールドです",
								})
							}
						} else {
							importErrors = append(importErrors, &item.Error{
								FirstLine:   firstLine,
								CurrentLine: line,
								LastLine:    lastLine,
								FieldId:     f.FieldID,
								FieldName:   f.FieldName,
								ErrorMsg:    "このフィールドは必須フィールドです",
							})
						}

						continue
					}

					if f.FieldType == "autonum" {
						nums := autoList[f.FieldID]
						dataItem.ItemMap[f.FieldID] = &Value{
							DataType: "autonum",
							Value:    nums[in],
						}
						continue
					}
					//  添加空数据
					addEmptyData(dataItem.ItemMap, f)
				}

				if len(importErrors) > 0 {
					return nil, errors.New("field is required")
				}

				queryJSON, _ := json.Marshal(dataItem)
				utils.DebugLog("MappingImport", fmt.Sprintf("item: [ %s ]", queryJSON))

				insertCxModel := mongo.NewInsertOneModel()
				insertCxModel.SetDocument(dataItem)
				cxModels = append(cxModels, insertCxModel)

				err = hs.Compare(cast.ToString(in+1), dataItem.ItemMap)
				if err != nil {
					// 返回错误信息
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				continue
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

							utils.ErrorLog("MappingImport", err.Error())
							return nil, err
						}
						errInfo := bke.WriteErrors[0]
						em := errInfo.Message
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

						utils.ErrorLog("MappingImport", errInfo.Message)
						return nil, errInfo
					}

					utils.ErrorLog("MappingImport", err.Error())
					return nil, err
				}

				result = res
			}

			err := hs.Commit()
			if err != nil {
				// 返回错误信息
				importErrors = append(importErrors, &item.Error{
					FirstLine: firstLine,
					LastLine:  lastLine,
					ErrorMsg:  err.Error(),
				})
				return nil, err
			}

			err = stream.Send(&item.MappingUploadResponse{
				Status: item.Status_SUCCESS,
				Result: &item.ImportResult{
					Insert: result.InsertedCount,
					Modify: result.ModifiedCount,
				},
			})

			if err != nil {
				// 返回错误信息
				importErrors = append(importErrors, &item.Error{
					FirstLine: firstLine,
					LastLine:  lastLine,
					ErrorMsg:  err.Error(),
				})
				return nil, err
			}

			return nil, nil
		}

		// 新规更新的场合
		// 循环更新或插入数据
		for in, d := range dataList {
			// 获取当前行号
			line := d.Index
			query := bson.M{
				"owners": bson.M{
					"$in": meta.UpdateOwners,
				},
			}
			for key, value := range d.Query {
				query["items."+key+".value"] = GetValueFromProto(value)
			}

			itemList := findItems(meta.Database, meta.GetDatastoreId(), query)

			// 新规更新的场合,没有找到对应数据
			if meta.MappingType == "upsert" && len(itemList) == 0 {
				// 新规更新的新规时
				itemMapData := d.Change
				// 合并query和change
				for key, value := range d.Query {
					itemMapData[key] = &Value{
						DataType: value.DataType,
						Value:    value.Value,
					}
				}
				// 数据
				dataItem := Item{
					AppID:       meta.GetAppId(),
					DatastoreID: meta.GetDatastoreId(),
					ItemMap:     itemMapData,
					Owners:      meta.Owners,
					CreatedAt:   time.Now(),
					CreatedBy:   meta.Writer,
					UpdatedAt:   time.Now(),
					UpdatedBy:   meta.Writer,
				}
				dataItem.ID = primitive.NewObjectID()
				dataItem.ItemID = dataItem.ID.Hex()
				dataItem.Status = "1"
				dataItem.CheckStatus = "0"

				err := hs.Add(cast.ToString(in+1), dataItem.ID.Hex(), nil)
				if err != nil {
					utils.ErrorLog("MappingImport", err.Error())
					// 返回错误信息
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				// 若无必须字段则直接插入
				if len(requriedFields) == 0 {

					for _, f := range allFields {
						if f.FieldType == "autonum" {
							num, err := autoNum(sc, meta.Database, f)
							if err != nil {
								if err.Error() != "(WriteConflict) WriteConflict" {
									utils.ErrorLog("MappingImport", err.Error())
									// 返回错误信息
									importErrors = append(importErrors, &item.Error{
										FirstLine:   firstLine,
										CurrentLine: line,
										LastLine:    lastLine,
										FieldId:     f.FieldID,
										FieldName:   f.FieldName,
										ErrorMsg:    err.Error(),
									})
								}

								return nil, err
							} else {
								dataItem.ItemMap[f.FieldID] = &Value{
									DataType: "autonum",
									Value:    num,
								}
							}

							continue
						}

						//  添加空数据
						addEmptyData(dataItem.ItemMap, f)
					}

					// 必须字段检查NG,返回错误
					if len(importErrors) > 0 {
						return nil, errors.New("field has error")
					}

					queryJSON, _ := json.Marshal(dataItem)
					utils.DebugLog("MappingImport", fmt.Sprintf("item: [ %s ]", queryJSON))

					insertCxModel := mongo.NewInsertOneModel()
					insertCxModel.SetDocument(dataItem)
					cxModels = append(cxModels, insertCxModel)

					err = hs.Compare(cast.ToString(in+1), dataItem.ItemMap)
					if err != nil {
						// 返回错误信息
						importErrors = append(importErrors, &item.Error{
							FirstLine:   firstLine,
							CurrentLine: line,
							LastLine:    lastLine,
							ErrorMsg:    err.Error(),
						})
						return nil, err
					}

					continue
				}

				// 有必须字段的情况下，先判断是否必须字段是否有值
				for _, f := range allFields {
					if f.IsRequired {
						if value, ok := dataItem.ItemMap[f.FieldID]; ok {
							if isEmptyValue(value) {
								importErrors = append(importErrors, &item.Error{
									FirstLine:   firstLine,
									CurrentLine: line,
									LastLine:    lastLine,
									FieldId:     f.FieldID,
									FieldName:   f.FieldName,
									ErrorMsg:    "このフィールドは必須フィールドです",
								})
							}
						} else {
							importErrors = append(importErrors, &item.Error{
								FirstLine:   firstLine,
								CurrentLine: line,
								LastLine:    lastLine,
								FieldId:     f.FieldID,
								FieldName:   f.FieldName,
								ErrorMsg:    "このフィールドは必須フィールドです",
							})
						}

						continue
					}

					if f.FieldType == "autonum" {
						num, err := autoNum(sc, meta.Database, f)
						if err != nil {
							if err.Error() != "(WriteConflict) WriteConflict" {
								utils.ErrorLog("MappingImport", err.Error())
								// 返回错误信息
								importErrors = append(importErrors, &item.Error{
									FirstLine:   firstLine,
									CurrentLine: line,
									LastLine:    lastLine,
									FieldId:     f.FieldID,
									FieldName:   f.FieldName,
									ErrorMsg:    err.Error(),
								})
							}

							return nil, err
						} else {
							dataItem.ItemMap[f.FieldID] = &Value{
								DataType: "autonum",
								Value:    num,
							}
						}

						continue
					}
					//  添加空数据
					addEmptyData(dataItem.ItemMap, f)
				}

				// 必须字段检查NG,返回错误
				if len(importErrors) > 0 {
					return nil, errors.New("field has error")
				}

				queryJSON, _ := json.Marshal(dataItem)
				utils.DebugLog("MappingImport", fmt.Sprintf("item: [ %s ]", queryJSON))

				insertCxModel := mongo.NewInsertOneModel()
				insertCxModel.SetDocument(dataItem)
				cxModels = append(cxModels, insertCxModel)

				err = hs.Compare(cast.ToString(in+1), dataItem.ItemMap)
				if err != nil {
					// 返回错误信息
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				continue
			}

			// 单更新时找不到更新对象数据error
			if meta.MappingType == "update" && len(itemList) == 0 {
				// 没有找到对应的数据
				importErrors = append(importErrors, &item.Error{
					FirstLine:   firstLine,
					CurrentLine: line,
					LastLine:    lastLine,
					ErrorMsg:    fmt.Sprintf("行 %d はエラーです,この台帳には該当するデータはありませんでした。", line),
				})
				utils.ErrorLog("MappingImport", fmt.Sprintf("行 %d はエラーです,この台帳には該当するデータはありませんでした。", line))
				return nil, fmt.Errorf("行 %d はエラーです,この台帳には該当するデータはありませんでした。", line)
			}

			// 更新的场合，更新类型是error的情况
			if meta.UpdateType == "error" {
				if len(itemList) > 1 {
					// 超过的情形，返回错误信息
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    fmt.Sprintf("行 %d はエラーです,複数のデータが見つかったため、更新処理は行われませんでした。", line),
					})
					utils.ErrorLog("MappingImport", fmt.Sprintf("行 %d はエラーです,複数のデータが見つかったため、更新処理は行われませんでした。", line))
					return nil, fmt.Errorf("行 %d はエラーです,複数のデータが見つかったため、更新処理は行われませんでした。", line)
				}

				oldItem := itemList[0]

				err := hs.Add(cast.ToString(in+1), oldItem.ItemID, oldItem.ItemMap)
				if err != nil {
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				change := bson.M{
					"updated_at": time.Now(),
					"updated_by": meta.Writer,
				}

				// 自增字段不更新
				for _, f := range allFields {
					if f.FieldType == "autonum" {
						if oldItem != nil {
							delete(oldItem.ItemMap, f.FieldID)
						}
						delete(d.Change, f.FieldID)
					}
					_, ok := d.Change[f.FieldID]
					// 需要进行自算的情况
					if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {

						if f.SelfCalculate == "add" {
							o := GetNumberValue(oldItem.ItemMap[f.FieldID])
							n := GetNumberValue(d.Change[f.FieldID])
							d.Change[f.FieldID].Value = o + n
							continue
						}
						if f.SelfCalculate == "sub" {
							o := GetNumberValue(oldItem.ItemMap[f.FieldID])
							n := GetNumberValue(d.Change[f.FieldID])
							d.Change[f.FieldID].Value = o - n
							continue
						}
					}
				}

				for key, value := range d.Change {
					change["items."+key] = value
				}

				update := bson.M{"$set": change}

				objectID, _ := primitive.ObjectIDFromHex(oldItem.ItemID)
				upCxModel := mongo.NewUpdateOneModel()
				upCxModel.SetFilter(bson.M{"_id": objectID})
				upCxModel.SetUpdate(update)
				upCxModel.SetUpsert(false)
				cxModels = append(cxModels, upCxModel)

				err = hs.Compare(cast.ToString(in+1), d.Change)
				if err != nil {
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				continue
			}

			// 更新的场合，更新类型是update-one的情况
			if meta.UpdateType == "update-one" {

				oldItem := itemList[0]

				err := hs.Add(cast.ToString(in+1), oldItem.ItemID, oldItem.ItemMap)
				if err != nil {
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				change := bson.M{
					"updated_at": time.Now(),
					"updated_by": meta.Writer,
				}

				// 自增字段不更新
				for _, f := range allFields {
					if f.FieldType == "autonum" {
						if oldItem != nil {
							delete(oldItem.ItemMap, f.FieldID)
						}
						delete(d.Change, f.FieldID)
					}
					_, ok := d.Change[f.FieldID]
					// 需要进行自算的情况
					if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
						if f.SelfCalculate == "add" {
							o := GetNumberValue(oldItem.ItemMap[f.FieldID])
							n := GetNumberValue(d.Change[f.FieldID])
							d.Change[f.FieldID].Value = o + n
							continue
						}
						if f.SelfCalculate == "sub" {
							o := GetNumberValue(oldItem.ItemMap[f.FieldID])
							n := GetNumberValue(d.Change[f.FieldID])
							d.Change[f.FieldID].Value = o - n
							continue
						}
					}
				}

				for key, value := range d.Change {
					change["items."+key] = value
				}

				update := bson.M{"$set": change}

				objectID, _ := primitive.ObjectIDFromHex(oldItem.ItemID)
				upCxModel := mongo.NewUpdateOneModel()
				upCxModel.SetFilter(bson.M{"_id": objectID})
				upCxModel.SetUpdate(update)
				upCxModel.SetUpsert(false)
				cxModels = append(cxModels, upCxModel)

				err = hs.Compare(cast.ToString(in+1), d.Change)
				if err != nil {
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				continue
			}

			// 更新的场合，更新类型是update-many的情况
			for k, oldItem := range itemList {

				index := strings.Builder{}
				index.WriteString("many_")
				index.WriteString(cast.ToString(in + 1))
				index.WriteString("_")
				index.WriteString(cast.ToString(k + 1))

				err := hs.Add(index.String(), oldItem.ItemID, oldItem.ItemMap)
				if err != nil {
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}

				// 自增字段不更新
				for _, f := range allFields {
					if f.FieldType == "autonum" {
						if oldItem != nil {
							delete(oldItem.ItemMap, f.FieldID)
						}
						delete(d.Change, f.FieldID)
					}

					_, ok := d.Change[f.FieldID]
					// 需要进行自算的情况
					if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
						if f.SelfCalculate == "add" {
							o := GetNumberValue(oldItem.ItemMap[f.FieldID])
							n := GetNumberValue(d.Change[f.FieldID])
							d.Change[f.FieldID].Value = o + n
							continue
						}
						if f.SelfCalculate == "sub" {
							o := GetNumberValue(oldItem.ItemMap[f.FieldID])
							n := GetNumberValue(d.Change[f.FieldID])
							d.Change[f.FieldID].Value = o - n
							continue
						}
					}
				}

				change := bson.M{
					"updated_at": time.Now(),
					"updated_by": meta.Writer,
				}

				for key, value := range d.Change {
					change["items."+key] = value
				}

				update := bson.M{"$set": change}

				objectID, _ := primitive.ObjectIDFromHex(oldItem.ItemID)
				upCxModel := mongo.NewUpdateOneModel()
				upCxModel.SetFilter(bson.M{"_id": objectID})
				upCxModel.SetUpdate(update)
				upCxModel.SetUpsert(false)
				cxModels = append(cxModels, upCxModel)

				err = hs.Compare(index.String(), d.Change)
				if err != nil {
					importErrors = append(importErrors, &item.Error{
						FirstLine:   firstLine,
						CurrentLine: line,
						LastLine:    lastLine,
						ErrorMsg:    err.Error(),
					})
					return nil, err
				}
			}

			continue
		}

		if len(cxModels) > 0 {
			res, err := c.BulkWrite(ctx, cxModels)
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

						utils.ErrorLog("MappingImport", err.Error())
						return nil, err
					}
					errInfo := bke.WriteErrors[0]
					em := errInfo.Message
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

					utils.ErrorLog("MappingImport", errInfo.Message)
					return nil, errInfo
				}

				utils.ErrorLog("MappingImport", err.Error())
				return nil, err
			}

			result = res
		}

		err := hs.Commit()
		if err != nil {
			// 返回错误信息
			importErrors = append(importErrors, &item.Error{
				FirstLine: firstLine,
				LastLine:  lastLine,
				ErrorMsg:  err.Error(),
			})
			return nil, err
		}

		err = stream.Send(&item.MappingUploadResponse{
			Status: item.Status_SUCCESS,
			Result: &item.ImportResult{
				Insert: result.InsertedCount,
				Modify: result.ModifiedCount,
			},
		})

		if err != nil {
			// 返回错误信息
			importErrors = append(importErrors, &item.Error{
				FirstLine: firstLine,
				LastLine:  lastLine,
				ErrorMsg:  err.Error(),
			})
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
		utils.ErrorLog("MappingImport", err.Error())
		// 返回错误信息
		importErrors = append(importErrors, &item.Error{
			FirstLine: firstLine,
			LastLine:  lastLine,
			ErrorMsg:  err.Error(),
		})

		err := stream.Send(&item.MappingUploadResponse{
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
		err := stream.Send(&item.MappingUploadResponse{
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
