package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"rxcsoft.cn/pit3/srv/database/proto/approve"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// ApproveCollection 审批集合
	ApproveCollection = "approves"
)

type (
	// ApproveItem 审批台账的数据
	ApproveItem struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		ItemID        string             `json:"item_id" bson:"item_id"`
		AppID         string             `json:"app_id" bson:"app_id"`
		DatastoreID   string             `json:"datastore_id" bson:"datastore_id"`
		ItemMap       ItemMap            `json:"items" bson:"items"`
		History       ItemMap            `json:"history" bson:"history"`
		Current       ItemMap            `json:"current" bson:"current"`
		ExampleID     string             `json:"example_id" bson:"example_id"`
		Applicant     string             `json:"applicant" bson:"applicant"`
		Approver      string             `json:"approver" bson:"approver"`
		ApproveStatus int64              `json:"approve_status" bson:"approve_status"`
		Node          Node               `json:"node" bson:"node"`
		CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy     string             `json:"created_by" bson:"created_by"`
		DeletedAt     time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy     string             `json:"deleted_by" bson:"deleted_by"`
	}

	ApproveInsertItem struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		ItemID        string             `json:"item_id" bson:"item_id"`
		AppID         string             `json:"app_id" bson:"app_id"`
		DatastoreID   string             `json:"datastore_id" bson:"datastore_id"`
		ItemMap       ItemMap            `json:"items" bson:"items"`
		History       ItemMap            `json:"history" bson:"history"`
		Current       ItemMap            `json:"current" bson:"current"`
		ExampleID     string             `json:"example_id" bson:"example_id"`
		Applicant     string             `json:"applicant" bson:"applicant"`
		Approver      string             `json:"approver" bson:"approver"`
		ApproveStatus int64              `json:"approve_status" bson:"approve_status"`
		CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy     string             `json:"created_by" bson:"created_by"`
		DeletedAt     time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy     string             `json:"deleted_by" bson:"deleted_by"`
	}

	// Node 流程节点
	Node struct {
		NodeID      string   `json:"node_id" bson:"node_id"`
		NodeName    string   `json:"node_name" bson:"node_name"`
		NodeType    string   `json:"node_type" bson:"node_type"`
		PrevNode    string   `json:"prev_node" bson:"prev_node"`
		NextNode    string   `json:"next_node" bson:"next_node"`
		Assignees   []string `json:"assignees" bson:"assignees"`
		ActType     string   `json:"act_type" bson:"act_type"`
		NodeGroupId string   `json:"node_group_id" bson:"node_group_id"`
	}

	// ApproveItemsParam 分页查询多条记录
	ApproveItemsParam struct {
		DatastoreID   string
		ConditionList []*Condition
		ConditionType string
		SearchType    string
		PageIndex     int64
		PageSize      int64
		Status        int64
		UserId        string
	}
)

// ToProto 转换为proto数据
func (t *ApproveItem) ToProto(showItem bool) *approve.ApproveItem {
	items := make(map[string]*approve.Value, len(t.ItemMap))
	for key, it := range t.ItemMap {
		dataType := it.DataType
		if it.Value != "##missing##" {
			item := &approve.Value{
				DataType: dataType,
				Value:    GetApproveValueString(it),
			}
			if item.DataType == "date" && item.Value == "0001-01-01" {
				item.Value = ""
			}
			items[key] = item
		}
	}
	hs := make(map[string]*approve.Value, len(t.History))
	for key, it := range t.History {
		dataType := it.DataType

		if it.Value != "##missing##" {
			history := &approve.Value{
				DataType: dataType,
				Value:    GetApproveValueString(it),
			}
			if history.DataType == "date" && history.Value == "0001-01-01" {
				history.Value = ""
			}
			hs[key] = history
		}
	}
	current := make(map[string]*approve.Value, len(t.Current))
	for key, it := range t.Current {
		dataType := it.DataType

		if it.Value != "##missing##" {
			approve := &approve.Value{
				DataType: dataType,
				Value:    GetApproveValueString(it),
			}
			if approve.DataType == "date" && approve.Value == "0001-01-01" {
				approve.Value = ""
			}
			current[key] = approve
		}
	}
	return &approve.ApproveItem{
		ItemId:        t.ItemID,
		AppId:         t.AppID,
		DatastoreId:   t.DatastoreID,
		Items:         items,
		History:       hs,
		Current:       current,
		ExampleId:     t.ExampleID,
		Applicant:     t.Applicant,
		Approver:      t.Approver,
		ApproveStatus: t.ApproveStatus,
		Node:          t.Node.ToProto(),
		CreatedAt:     t.CreatedAt.String(),
		CreatedBy:     t.CreatedBy,
		DeletedAt:     t.DeletedAt.String(),
		DeletedBy:     t.DeletedBy,
	}
}

// ToProto 转换为proto数据
func (n *Node) ToProto() *approve.Node {
	return &approve.Node{
		NodeId:      n.NodeID,
		NodeName:    n.NodeName,
		NodeType:    n.NodeType,
		PrevNode:    n.PrevNode,
		NextNode:    n.NextNode,
		Assignees:   n.Assignees,
		ActType:     n.ActType,
		NodeGroupId: n.NodeGroupId,
	}
}

// ToProto 转换为proto数据
func (v *Value) ToProto() *approve.Value {
	approve := &approve.Value{
		DataType: v.DataType,
		Value:    GetApproveValueString(v),
	}
	if approve.DataType == "date" && approve.Value == "0001-01-01" {
		approve.Value = ""
	}
	return approve
}

// GetApproveValueString 获取值
func GetApproveValueString(value *Value) (v string) {
	switch value.DataType {
	case "text", "textarea":
		if value.Value == nil {
			return ""
		}
		return value.Value.(string)
	case "options":
		if value.Value == nil {
			return ""
		}
		return value.Value.(string)
	case "number":
		if value.Value == nil {
			return "0"
		}
		switch value.Value.(type) {
		case int64:
			return strconv.FormatInt(value.Value.(int64), 10)
		case float64:
			return strconv.FormatFloat(value.Value.(float64), 'f', -1, 64)
		default:
			return strconv.FormatFloat(0.0, 'f', -1, 64)
		}
	case "autonum":
		if value.Value == nil {
			return ""
		}
		switch value.Value.(type) {
		case int64:
			return strconv.FormatInt(value.Value.(int64), 10)
		default:
			return value.Value.(string)
		}
	case "date":
		if value.Value == nil {
			return ""
		}
		switch value.Value.(type) {
		case primitive.DateTime:
			return value.Value.(primitive.DateTime).Time().Format("2006-01-02")
		case time.Time:
			return value.Value.(time.Time).Format("2006-01-02")
		default:
			return ""
		}
	case "time":
		if value.Value == nil {
			return ""
		}
		return value.Value.(string)
	case "switch":
		if value.Value == nil {
			return "false"
		}
		return strconv.FormatBool(value.Value.(bool))
	case "user":
		if value.Value == nil {
			return ""
		}
		jsonBytes, err := json.Marshal(value.Value)
		if err != nil {
			return ""
		}
		return string(jsonBytes)
	case "file":
		if value.Value == nil {
			return ""
		}
		return value.Value.(string)
	case "function":
		if value.Value == nil {
			return ""
		}
		switch value.Value.(type) {
		case int32:
			return strconv.FormatInt(int64(value.Value.(int32)), 10)
		case int64:
			return strconv.FormatInt(value.Value.(int64), 10)
		case float64:
			return strconv.FormatFloat(value.Value.(float64), 'f', -1, 64)
		default:
			return value.Value.(string)
		}
	case "lookup":
		if value.Value == nil {
			return ""
		}
		return value.Value.(string)
	default:
		jsonBytes, _ := json.Marshal(value.Value)
		return string(jsonBytes)
	}
}

// GetApproveDataValue 获取对应的数据类型的数据
func GetApproveDataValue(value *approve.Value, field Field, uMap map[string]string, lang *language.Language, needConst bool) (v interface{}) {
	switch value.DataType {
	case "text", "textarea":
		return value.GetValue()
	case "number":
		result, err := strconv.ParseFloat(value.GetValue(), 64)
		if err != nil {
			return 0
		}
		return result
	case "date":
		zone := time.Time{}
		if len(value.GetValue()) == 0 {
			return zone
		}
		date, err := time.Parse("2006-01-02", value.GetValue())
		if err != nil {
			return zone
		}
		return date
	case "time":
		return value.GetValue()
	case "switch":
		result, err := strconv.ParseBool(value.GetValue())
		if err != nil {
			return false
		}
		return result
	case "user":
		if len(value.GetValue()) == 0 {
			return []string{}
		}
		result := strings.Split(value.GetValue(), ",")
		if needConst {
			var uList []string
			for _, uid := range result {
				uList = append(uList, uMap[uid])
			}
			return uList
		}
		return result
	case "file":
		return value.GetValue()
	case "options":
		if needConst {
			return utils.GetLangValue(lang, utils.GetOptionKey(field.AppID, field.OptionID, value.GetValue()), "")
		}
		return value.GetValue()
	case "lookup":
		return value.GetValue()
	}

	return ""
}

// FindApproveItems 获取审批的数据
func FindApproveItems(db, wfID string, param ApproveItemsParam) (items []ApproveItem, total int64, err error) {
	client := database.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 生成临时数据
	tmpCollection := "approves_" + wfID + "_" + strconv.Itoa(time.Now().UTC().Nanosecond())

	opts := options.Collection()
	opts.SetReadPreference(readpref.Primary())

	c := client.Database(database.GetDBName(db)).Collection(tmpCollection, opts)

	err = genTmpItems(db, wfID, param.UserId, tmpCollection, param.Status)
	if err != nil {
		utils.ErrorLog("FindApproveItems", err.Error())
		return nil, 0, err
	}

	query := bson.M{}

	buildApproveMatch(param.ConditionList, param.SearchType, param.ConditionType, query)

	// 排序
	sortItem := bson.D{
		bson.E{Key: "created_at", Value: -1},
	}

	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindApproveItems", err.Error())
		return nil, 0, err
	}

	fields, err := getFields(db, param.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindApproveItems", err.Error())
		return nil, 0, err
	}

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	pipe = append(pipe, bson.M{
		"$sort": sortItem,
	})

	skip := (param.PageIndex - 1) * param.PageSize
	limit := param.PageSize

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
		"_id":            1,
		"item_id":        1,
		"app_id":         1,
		"datastore_id":   1,
		"example_id":     "$ex_id",
		"approve_status": 1,
		"applicant":      1,
		"node":           1,
		"approver": bson.M{
			"$cond": bson.M{
				"if": bson.M{"$ne": []interface{}{
					"$last_process.status",
					0,
				}},
				"then": "$last_process.user_id",
				"else": "",
			},
		},
		"created_at": 1,
		"created_by": 1,
		"deleted_at": 1,
		"deleted_by": 1,
	}

	// 抽出项目编辑
	for _, f := range fields {
		// TODO 用户&选项类型数据已经固化不需要关联查询出名称
		// if f.FieldType == "user" {
		// 	pp := []bson.M{
		// 		{
		// 			"$match": bson.M{
		// 				"$expr": bson.M{
		// 					"$and": []bson.M{
		// 						{
		// 							"$in": []string{"$user_id", "$$user"},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	}

		// 	lookup := bson.M{
		// 		"from": "users",
		// 		"let": bson.M{
		// 			"user": bson.M{
		// 				"$cond": bson.M{
		// 					"if": bson.M{
		// 						"$isArray": []string{"$items." + f.FieldID + ".value"},
		// 					},
		// 					"then": "$items." + f.FieldID + ".value",
		// 					"else": []string{},
		// 				},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "relations_" + f.FieldID,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": lookup,
		// 	})

		// 	project["items."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$items." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$relations_" + f.FieldID + ".user_name",
		// 			},
		// 		},
		// 	}

		// 	hlookup := bson.M{
		// 		"from": "users",
		// 		"let": bson.M{
		// 			"user": bson.M{
		// 				"$cond": bson.M{
		// 					"if": bson.M{
		// 						"$isArray": []string{"$history." + f.FieldID + ".value"},
		// 					},
		// 					"then": "$history." + f.FieldID + ".value",
		// 					"else": []string{},
		// 				},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "history_relations_" + f.FieldID,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": hlookup,
		// 	})

		// 	project["history."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$history." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$history_relations_" + f.FieldID + ".user_name",
		// 			},
		// 		},
		// 	}

		// 	continue
		// }
		// if f.FieldType == "options" {
		// 	pp := []bson.M{
		// 		{
		// 			"$match": bson.M{
		// 				"$expr": bson.M{
		// 					"$and": []bson.M{
		// 						{
		// 							"$eq": []string{"$app_id", "$$app_id"},
		// 						},
		// 						{
		// 							"$eq": []string{"$option_id", "$$option_id"},
		// 						},
		// 						{
		// 							"$eq": []string{"$option_value", "$$option_value"},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	}

		// 	lookup := bson.M{
		// 		"from": "options",
		// 		"let": bson.M{
		// 			"app_id":    f.AppID,
		// 			"option_id": f.OptionID,
		// 			"option_value": bson.M{
		// 				"$ifNull": []string{"$items." + f.FieldID + ".value", ""},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "relations_" + f.FieldID,
		// 	}

		// 	unwind := bson.M{
		// 		"path":                       "$relations_" + f.FieldID,
		// 		"preserveNullAndEmptyArrays": true,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": lookup,
		// 	})
		// 	pipe = append(pipe, bson.M{
		// 		"$unwind": unwind,
		// 	})

		// 	project["items."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$items." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$relations_" + f.FieldID + ".option_label",
		// 			},
		// 		},
		// 	}

		// 	hlookup := bson.M{
		// 		"from": "options",
		// 		"let": bson.M{
		// 			"app_id":    f.AppID,
		// 			"option_id": f.OptionID,
		// 			"option_value": bson.M{
		// 				"$ifNull": []string{"$history." + f.FieldID + ".value", ""},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "history_relations_" + f.FieldID,
		// 	}

		// 	hunwind := bson.M{
		// 		"path":                       "$history_relations_" + f.FieldID,
		// 		"preserveNullAndEmptyArrays": true,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": hlookup,
		// 	})
		// 	pipe = append(pipe, bson.M{
		// 		"$unwind": hunwind,
		// 	})

		// 	project["history."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$history." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$history_relations_" + f.FieldID + ".option_label",
		// 			},
		// 		},
		// 	}

		// 	continue
		// }

		// TODO 审批中不需要函数字段内容
		if f.FieldType == "function" {
			// project["items."+f.FieldID+".value"] = ""
			// project["items."+f.FieldID+".data_type"] = f.ReturnType
			// project["history."+f.FieldID+".value"] = ""
			// project["history."+f.FieldID+".data_type"] = f.ReturnType
			continue
		}

		project["items."+f.FieldID] = bson.M{
			"$cond": bson.M{
				"if": bson.M{"$eq": []interface{}{
					bson.M{
						"$type": "$items." + f.FieldID + ".value",
					},
					"missing",
				}},
				"then": bson.M{
					"data_type": f.FieldType,
					"value":     "##missing##",
				},
				"else": bson.M{
					"data_type": f.FieldType,
					"value":     "$items." + f.FieldID + ".value",
				},
			},
		}
		project["history."+f.FieldID] = bson.M{
			"$cond": bson.M{
				"if": bson.M{"$eq": []interface{}{
					bson.M{
						"$type": "$history." + f.FieldID + ".value",
					},
					"missing",
				}},
				"then": bson.M{
					"data_type": f.FieldType,
					"value":     "##missing##",
				},
				"else": bson.M{
					"data_type": f.FieldType,
					"value":     "$history." + f.FieldID + ".value",
				},
			},
		}
		project["current."+f.FieldID] = bson.M{
			"$cond": bson.M{
				"if": bson.M{"$eq": []interface{}{
					bson.M{
						"$type": "$current." + f.FieldID + ".value",
					},
					"missing",
				}},
				"then": bson.M{
					"data_type": f.FieldType,
					"value":     "##missing##",
				},
				"else": bson.M{
					"data_type": f.FieldType,
					"value":     "$current." + f.FieldID + ".value",
				},
			},
		}
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	var result []ApproveItem

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("FindApproveItems", err.Error())
		return result, 0, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item ApproveItem
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindApproveItems", err.Error())
			return result, 0, err
		}

		result = append(result, item)
	}

	err = c.Drop(ctx)
	if err != nil {
		utils.ErrorLog("FindApproveItems", err.Error())
		return result, 0, err
	}

	return result, t, nil
}

// genTmpItems 生成临时数据
func genTmpItems(db, wfID, userID, out string, status int64) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("wf_examples")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"wf_id": wfID,
	}

	pipe := []bson.M{
		{"$match": query},
	}

	// approves
	approvesPipe := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$example_id", "$$ex_id"},
						},
					},
				},
			},
		},
	}

	approvesLookup := bson.M{
		"from": "approves",
		"let": bson.M{
			"ex_id": "$ex_id",
		},
		"pipeline": approvesPipe,
		"as":       "item",
	}

	approvesUnwind := bson.M{
		"path":                       "$item",
		"preserveNullAndEmptyArrays": true,
	}

	pipe = append(pipe, bson.M{
		"$lookup": approvesLookup,
	})
	pipe = append(pipe, bson.M{
		"$unwind": approvesUnwind,
	})

	// wf_process
	processPipe := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$ex_id", "$$ex_id"},
						},
						{
							"$ne": []string{"$updated_by", "SYSTEM"},
						},
					},
				},
			},
		},
		{
			"$sort": bson.M{"updated_at": 1},
		},
	}

	processLookup := bson.M{
		"from": "wf_process",
		"let": bson.M{
			"ex_id": "$ex_id",
		},
		"pipeline": processPipe,
		"as":       "process",
	}

	pipe = append(pipe, bson.M{
		"$lookup": processLookup,
	})

	match := bson.M{
		"$or": []bson.M{
			{"user_id": userID},
			{"process.user_id": userID},
		},
	}

	if status != 0 {
		match["status"] = status
	}

	pipe = append(pipe, bson.M{
		"$match": match,
	})

	project := bson.M{
		"_id":            1,
		"ex_id":          1,
		"app_id":         "$item.app_id",
		"created_at":     "$item.created_at",
		"created_by":     "$item.created_by",
		"datastore_id":   "$item.datastore_id",
		"deleted_at":     "$item.deleted_at",
		"deleted_by":     "$item.deleted_by",
		"example_id":     1,
		"history":        "$item.history",
		"current":        "$item.current",
		"item_id":        "$item.item_id",
		"items":          "$item.items",
		"applicant":      "$user_id",
		"approve_status": "$status",
		"last_process": bson.M{
			"$slice": []interface{}{"$process", -1, 1},
		},
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	lastProcessUnwind := bson.M{
		"path":                       "$last_process",
		"preserveNullAndEmptyArrays": true,
	}

	pipe = append(pipe, bson.M{
		"$unwind": lastProcessUnwind,
	})

	// wf_node
	nodePipe := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$node_id", "$$node_id"},
						},
						{
							"$eq": []string{"$wf_id", "$$wf_id"},
						},
					},
				},
			},
		},
	}

	nodeLookup := bson.M{
		"from": "wf_node",
		"let": bson.M{
			"node_id": "$last_process.current_node",
			"wf_id":   wfID,
		},
		"pipeline": nodePipe,
		"as":       "node",
	}

	nodeUnwind := bson.M{
		"path":                       "$node",
		"preserveNullAndEmptyArrays": true,
	}

	pipe = append(pipe, bson.M{
		"$lookup": nodeLookup,
	})
	pipe = append(pipe, bson.M{
		"$unwind": nodeUnwind,
	})

	pipe = append(pipe, bson.M{
		"$out": out,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("GenTmpItems", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	_, err = c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("GenTmpItems", err.Error())
		return err
	}

	return nil
}

// FindApproveCount 获取件数
func FindApproveCount(db, wfID string, status int64) (total int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("wf_examples")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"wf_id":  wfID,
		"status": status,
	}

	t, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindApproveCount", err.Error())
		return 0, err
	}

	return t, nil
}

func buildApproveMatch(conditionList []*Condition, searchType string, conditionType string, query bson.M) {

	if searchType == "item" {
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
						case "user":
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
						case "user":
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
	} else {
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
									"history." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
								})
							} else if condition.Operator == "<>" {
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": bson.M{
										"$ne": condition.SearchValue,
									},
								})
							} else {
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": condition.SearchValue,
								})
							}
						case "switch":
							// 只能是等于
							and = append(and, bson.M{
								"history." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
							})
						case "file":
							if condition.SearchValue == "true" {
								// 存在
								and = append(and, bson.M{
									"$and": []bson.M{
										{
											"history." + condition.FieldID + ".value": bson.M{"$exists": true},
										},
										{
											"history." + condition.FieldID + ".value": bson.M{"$ne": "[]"},
										},
									},
								})
							} else {
								// 不存在
								and = append(and, bson.M{
									"$or": []bson.M{
										{
											"history." + condition.FieldID + ".value": bson.M{"$exists": false},
										},
										{
											"history." + condition.FieldID + ".value": bson.M{"$eq": "[]"},
										},
									},
								})
							}
						case "lookup":
							// 只能是like和=两种情况（默认是等于）
							if condition.Operator == "like" {
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
								})
							} else if condition.Operator == "<>" {
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": bson.M{
										"$ne": condition.SearchValue,
									},
								})
							} else {
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": condition.SearchValue,
								})
							}
						case "options":
							// 只能是in和=两种情况（默认是等于）
							if condition.Operator == "in" {
								values := strings.Split(condition.SearchValue, ",")
								if len(values) > 0 {
									// IN
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$in": values},
									})
								}
							} else if condition.Operator == "<>" {
								value := condition.SearchValue
								// 不等于
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": bson.M{
										"$ne": value,
									},
								})
							} else {
								value := condition.SearchValue
								// 等于
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": value,
								})
							}
						case "user":
							// 只能是in和=两种情况（默认是等于）
							if condition.Operator == "in" {
								values := strings.Split(condition.SearchValue, ",")
								if len(values) > 0 {
									// IN
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$in": values},
									})
								}
							} else if condition.Operator == "<>" {
								value := condition.SearchValue
								// 不等于
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": bson.M{
										"$ne": value,
									},
								})
							} else {
								value := condition.SearchValue
								// 等于
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": value,
								})
							}
						case "number", "date", "time":
							// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
							if condition.ConditionType == "1" {
								values := strings.Split(condition.SearchValue, "~")
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])},
								})
								and = append(and, bson.M{
									"history." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])},
								})
							} else {
								// 可以是=,>,<,>=,<=,<>(默认是等于)
								// 大于的场合
								if condition.Operator == ">" {
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$gt": getSearchValue(condition.FieldType, condition.SearchValue)},
									})
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, condition.SearchValue)},
									})
								} else if condition.Operator == "<" {
									// 小于的场合
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, condition.SearchValue)},
									})
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$lte": getSearchValue(condition.FieldType, condition.SearchValue)},
									})
								} else if condition.Operator == "<>" {
									// 不等于的场合
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$ne": getSearchValue(condition.FieldType, condition.SearchValue)},
									})
								} else {
									// 默认的[等于]的场合
									and = append(and, bson.M{
										"history." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
									})
								}
							}
						default:
						}
					} else {
						switch condition.FieldType {
						case "user":
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
									"history." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
								}
								or = append(or, q)
							} else if condition.Operator == "<>" {
								q := bson.M{
									"history." + condition.FieldID + ".value": bson.M{
										"$ne": condition.SearchValue,
									},
								}
								or = append(or, q)
							} else {
								q := bson.M{
									"history." + condition.FieldID + ".value": condition.SearchValue,
								}
								or = append(or, q)
							}
						case "switch":
							// 只能是等于
							q := bson.M{
								"history." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
							}
							or = append(or, q)
						case "file":
							if condition.SearchValue == "true" {
								// 存在
								q := bson.M{
									"$and": []bson.M{
										{
											"history." + condition.FieldID + ".value": bson.M{"$exists": true},
										},
										{
											"history." + condition.FieldID + ".value": bson.M{"$ne": "[]"},
										},
									},
								}
								or = append(or, q)
							} else {
								// 不存在
								q := bson.M{
									"$or": []bson.M{
										{
											"history." + condition.FieldID + ".value": bson.M{"$exists": false},
										},
										{
											"history." + condition.FieldID + ".value": bson.M{"$eq": "[]"},
										},
									},
								}
								or = append(or, q)
							}
						case "lookup":
							// 只能是like和=两种情况（默认是等于）
							if condition.Operator == "like" {
								q := bson.M{
									"history." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
								}
								or = append(or, q)
							} else if condition.Operator == "<>" {
								q := bson.M{
									"history." + condition.FieldID + ".value": bson.M{
										"$ne": condition.SearchValue,
									},
								}
								or = append(or, q)
							} else {
								q := bson.M{
									"history." + condition.FieldID + ".value": condition.SearchValue,
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
										"history." + condition.FieldID + ".value": bson.M{"$in": values},
									}
									or = append(or, q)
								}
							} else if condition.Operator == "<>" {
								q := bson.M{
									"history." + condition.FieldID + ".value": bson.M{
										"$ne": condition.SearchValue,
									},
								}
								or = append(or, q)
							} else {
								// 等于
								q := bson.M{
									"history." + condition.FieldID + ".value": condition.SearchValue,
								}
								or = append(or, q)
							}
						case "number", "date", "time":
							// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
							if condition.ConditionType == "1" {
								values := strings.Split(condition.SearchValue, "~")
								q := bson.M{
									"$and": []bson.M{
										{"history." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
										{"history." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
									},
								}
								or = append(or, q)
							} else {
								// 可以是=,>,<,>=,<=,<>(默认是等于)
								// 大于的场合
								if condition.Operator == ">" {
									q := bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$gt": getSearchValue(condition.FieldType, condition.SearchValue)},
									}
									or = append(or, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									q := bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, condition.SearchValue)},
									}
									or = append(or, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									q := bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, condition.SearchValue)},
									}
									or = append(or, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									q := bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$lte": getSearchValue(condition.FieldType, condition.SearchValue)},
									}
									or = append(or, q)
								} else if condition.Operator == "<>" {
									// 不等于的场合
									q := bson.M{
										"history." + condition.FieldID + ".value": bson.M{"$ne": getSearchValue(condition.FieldType, condition.SearchValue)},
									}
									or = append(or, q)
								} else {
									// 默认的[等于]的场合
									q := bson.M{
										"history." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
									}
									or = append(or, q)
								}
							}
						default:
						}
					} else {
						switch condition.FieldType {
						case "user":
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

}

// FindApproveItem 通过流程实例ID获取流程审批数据信息
func FindApproveItem(db, exId, datastoreId string) (item ApproveItem, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ApproveCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result ApproveItem

	query := bson.M{
		"deleted_by": "",
		"example_id": exId,
	}

	fields, err := getFields(db, datastoreId)
	if err != nil {
		utils.ErrorLog("FindApproveItem", err.Error())
		return result, err
	}

	pipe := []bson.M{
		{
			"$match": query,
		},
	}

	project := bson.M{
		"_id":          1,
		"item_id":      1,
		"app_id":       1,
		"datastore_id": 1,
		"example_id":   1,
		"created_at":   1,
		"created_by":   1,
		"deleted_at":   1,
		"deleted_by":   1,
	}

	for _, f := range fields {

		// if !IsOrigin {
		// TODO 用户&选项类型数据已经固化不需要关联查询出名称
		// if f.FieldType == "user" {
		// 	pp := []bson.M{
		// 		{
		// 			"$match": bson.M{
		// 				"$expr": bson.M{
		// 					"$and": []bson.M{
		// 						{
		// 							"$in": []string{"$user_id", "$$user"},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	}

		// 	lookup := bson.M{
		// 		"from": "users",
		// 		"let": bson.M{
		// 			"user": bson.M{
		// 				"$cond": bson.M{
		// 					"if": bson.M{
		// 						"$isArray": []string{"$items." + f.FieldID + ".value"},
		// 					},
		// 					"then": "$items." + f.FieldID + ".value",
		// 					"else": []string{},
		// 				},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "relations_" + f.FieldID,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": lookup,
		// 	})

		// 	project["items."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$items." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$relations_" + f.FieldID + ".user_name",
		// 			},
		// 		},
		// 	}

		// 	hlookup := bson.M{
		// 		"from": "users",
		// 		"let": bson.M{
		// 			"user": bson.M{
		// 				"$cond": bson.M{
		// 					"if": bson.M{
		// 						"$isArray": []string{"$history." + f.FieldID + ".value"},
		// 					},
		// 					"then": "$history." + f.FieldID + ".value",
		// 					"else": []string{},
		// 				},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "history_relations_" + f.FieldID,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": hlookup,
		// 	})

		// 	project["history."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$history." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$history_relations_" + f.FieldID + ".user_name",
		// 			},
		// 		},
		// 	}

		// 	continue
		// }
		// if f.FieldType == "options" {
		// 	pp := []bson.M{
		// 		{
		// 			"$match": bson.M{
		// 				"$expr": bson.M{
		// 					"$and": []bson.M{
		// 						{
		// 							"$eq": []string{"$app_id", "$$app_id"},
		// 						},
		// 						{
		// 							"$eq": []string{"$option_id", "$$option_id"},
		// 						},
		// 						{
		// 							"$eq": []string{"$option_value", "$$option_value"},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	}

		// 	lookup := bson.M{
		// 		"from": "options",
		// 		"let": bson.M{
		// 			"app_id":    f.AppID,
		// 			"option_id": f.OptionID,
		// 			"option_value": bson.M{
		// 				"$ifNull": []string{"$items." + f.FieldID + ".value", ""},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "relations_" + f.FieldID,
		// 	}

		// 	unwind := bson.M{
		// 		"path":                       "$relations_" + f.FieldID,
		// 		"preserveNullAndEmptyArrays": true,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": lookup,
		// 	})
		// 	pipe = append(pipe, bson.M{
		// 		"$unwind": unwind,
		// 	})

		// 	project["items."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$items." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$relations_" + f.FieldID + ".option_label",
		// 			},
		// 		},
		// 	}

		// 	hlookup := bson.M{
		// 		"from": "options",
		// 		"let": bson.M{
		// 			"app_id":    f.AppID,
		// 			"option_id": f.OptionID,
		// 			"option_value": bson.M{
		// 				"$ifNull": []string{"$history." + f.FieldID + ".value", ""},
		// 			},
		// 		},
		// 		"pipeline": pp,
		// 		"as":       "history_relations_" + f.FieldID,
		// 	}

		// 	hunwind := bson.M{
		// 		"path":                       "$history_relations_" + f.FieldID,
		// 		"preserveNullAndEmptyArrays": true,
		// 	}

		// 	pipe = append(pipe, bson.M{
		// 		"$lookup": hlookup,
		// 	})
		// 	pipe = append(pipe, bson.M{
		// 		"$unwind": hunwind,
		// 	})

		// 	project["history."+f.FieldID] = bson.M{
		// 		"$cond": bson.M{
		// 			"if": bson.M{
		// 				"$eq": []interface{}{
		// 					bson.M{
		// 						"$type": "$history." + f.FieldID + ".value",
		// 					},
		// 					"missing",
		// 				},
		// 			},
		// 			"then": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "##missing##",
		// 			},
		// 			"else": bson.M{
		// 				"data_type": f.FieldType,
		// 				"value":     "$history_relations_" + f.FieldID + ".option_label",
		// 			},
		// 		},
		// 	}

		// 	continue
		// }
		// }

		// TODO 审批中不需要函数字段内容
		if f.FieldType == "function" {
			// project["items."+f.FieldID+".value"] = ""
			// project["items."+f.FieldID+".data_type"] = f.ReturnType
			// project["history."+f.FieldID+".value"] = ""
			// project["history."+f.FieldID+".data_type"] = f.ReturnType
			continue
		}

		project["items."+f.FieldID] = bson.M{
			"$cond": bson.M{
				"if": bson.M{"$eq": []interface{}{
					bson.M{
						"$type": "$items." + f.FieldID + ".value",
					},
					"missing",
				}},
				"then": bson.M{
					"data_type": f.FieldType,
					"value":     "##missing##",
				},
				"else": bson.M{
					"data_type": f.FieldType,
					"value":     "$items." + f.FieldID + ".value",
				},
			},
		}

		project["history."+f.FieldID] = bson.M{
			"$cond": bson.M{
				"if": bson.M{"$eq": []interface{}{
					bson.M{
						"$type": "$history." + f.FieldID + ".value",
					},
					"missing",
				}},
				"then": bson.M{
					"data_type": f.FieldType,
					"value":     "##missing##",
				},
				"else": bson.M{
					"data_type": f.FieldType,
					"value":     "$history." + f.FieldID + ".value",
				},
			},
		}

		project["current."+f.FieldID] = bson.M{
			"$cond": bson.M{
				"if": bson.M{"$eq": []interface{}{
					bson.M{
						"$type": "$current." + f.FieldID + ".value",
					},
					"missing",
				}},
				"then": bson.M{
					"data_type": f.FieldType,
					"value":     "##missing##",
				},
				"else": bson.M{
					"data_type": f.FieldType,
					"value":     "$current." + f.FieldID + ".value",
				},
			},
		}
	}

	project["items.template_id.value"] = "$items.template_id.value"
	project["items.template_id.data_type"] = "$items.template_id.data_type"

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindItem", fmt.Sprintf("query: [ %s ]", queryJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("FindApproveItem", err.Error())
		return result, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var item ApproveItem
		err := cur.Decode(&item)
		if err != nil {
			utils.ErrorLog("FindApproveItem", err.Error())
			return result, err
		}

		result = item
	}

	return result, nil
}

// AddApprove 添加台账数据
func AddApprove(db string, i *ApproveItem) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ApproveCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	newId := primitive.NewObjectID()
	itemId := i.ItemID
	if len(i.ItemID) == 0 {
		itemId = newId.Hex()
	}
	ini := ApproveInsertItem{
		ID:            newId,
		ItemID:        itemId,
		AppID:         i.AppID,
		DatastoreID:   i.DatastoreID,
		ItemMap:       i.ItemMap,
		History:       i.History,
		Current:       i.Current,
		ExampleID:     i.ExampleID,
		Applicant:     i.Applicant,
		Approver:      i.Approver,
		ApproveStatus: i.ApproveStatus,
		CreatedAt:     i.CreatedAt,
		CreatedBy:     i.CreatedBy,
		DeletedAt:     i.DeletedAt,
		DeletedBy:     i.DeletedBy,
	}

	queryJSON, _ := json.Marshal(ini)
	utils.DebugLog("AddApprove", fmt.Sprintf("ApproveItem: [ %s ]", queryJSON))

	if _, err = c.InsertOne(ctx, ini); err != nil {
		utils.ErrorLog("AddApprove", err.Error())
		return "", err
	}

	return ini.ItemID, nil
}

// DeleteApproveItems 删除审批数据
func DeleteApproveItems(db string, items []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ApproveCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteApproveItems", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteApproveItems", err.Error())
		return err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, id := range items {
			query := bson.M{
				"example_id": id,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteApproveItems", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("DeleteApproveItems", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteApproveItems", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteApproveItems", err.Error())
		return err
	}

	session.EndSession(ctx)
	return nil
}
