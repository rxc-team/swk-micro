package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goinggo/mapstructure"
	"github.com/micro/go-micro/v2/client"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	database "rxcsoft.cn/utils/mongo"
	"rxcsoft.cn/utils/timex"
)

type HistoryServer struct {
	db    string
	uid   string
	did   string
	datas map[string]*Data
	sc    mongo.SessionContext
	fMap  map[string]Field
	uMap  map[string]string
	lang  *language.Language
}

type Data struct {
	itemId  string   // 履历对应数据ID
	before  ItemMap  // 变更前
	after   ItemMap  // 变更后
	hs      ItemMap  // 履历
	hsType  string   // 履历类型，根据before和after判断得出
	changes []Change // 变更点
}

// hs 台账按操作为单位的履历的数据
type hs struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	HistoryID   string             `json:"history_id" bson:"history_id"`
	HistoryType string             `json:"history_type" bson:"history_type"`
	DatastoreID string             `json:"datastore_id" bson:"datastore_id"`
	ItemID      string             `json:"item_id" bson:"item_id"`
	FixedItems  ItemMap            `json:"fixed_items" bson:"fixed_items"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy   string             `json:"created_by" bson:"created_by"`
}

func NewHistory(db, uid, did, langCd, domain string, sc mongo.SessionContext, fields []Field) *HistoryServer {

	hs := &HistoryServer{
		db:    db,
		uid:   uid,
		did:   did,
		datas: make(map[string]*Data),
		sc:    sc,
		lang:  utils.GetLanguageData(db, langCd, domain),
	}

	fMap := make(map[string]Field)
	for _, fd := range fields {
		fMap[fd.FieldID] = fd
	}

	hs.fMap = fMap

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest
	// 从query中获取参数
	req.InvalidatedIn = "true"
	req.Domain = domain
	req.Database = db

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		return hs
	}

	uMap := make(map[string]string)
	for _, u := range response.Users {
		uMap[u.UserId] = u.UserName
	}

	hs.uMap = uMap

	return hs
}

// 通过传入的数据，itemId，获取对应的变更前的值
func (h *HistoryServer) Add(index string, itemId string, data ItemMap) error {

	// 为nil的场合，设置为空map
	if data == nil {
		data = make(ItemMap)
	}

	// 取before的值
	h.datas[index] = &Data{
		itemId: itemId,
		before: data,
	}
	return nil
}

// 通过传入的数据，itemId，获取对应的变更后的值，做比较
func (h *HistoryServer) Compare(index string, change ItemMap) error {

	data := h.datas[index]

	// 为nil的场合，设置为空map
	if change == nil {
		change = make(ItemMap)
	}

	// 取after的值
	data.after = copyMap(change)

	changes := make([]Change, 0)

	// 删除的场合
	if len(change) == 0 {
		data.hsType = "delete"

		data.hs = copyMap(data.before)

		changes = append(changes, Change{
			FieldID:   "",
			FieldName: "",
			LocalName: "",
			OldValue:  "",
			NewValue:  "",
		})

		data.changes = changes
		return nil
	}

	data.hs = copyMap(data.before)

	// 比较变更点
	for field, n := range change {

		// 更新值
		data.hs[field] = n

		// 判断当前字段是否存在，不存在的场合，说明传入有误
		fieldInfo, exist := h.fMap[field]
		if !exist {
			continue
		}

		// 旧数据不存在的场合
		if len(data.before) == 0 {

			// 新规的场合
			data.hsType = "insert"

			v := value(n, fieldInfo, h.uMap, h.lang)
			if len(v) > 0 {
				changes = append(changes, Change{
					FieldID:   field,
					FieldName: fieldInfo.FieldName,
					LocalName: utils.GetLangValue(h.lang, fieldInfo.FieldName, ""),
					OldValue:  "",
					NewValue:  v,
				})
			}

			continue
		}

		// 更新的场合
		data.hsType = "update"

		// 比较数据变更
		if o, ok := data.before[field]; ok {

			// 如果旧数据存在，切值不相等的场合，作为change内容传入
			if !hsCompare(n, o) {
				changes = append(changes, Change{
					FieldID:   field,
					FieldName: fieldInfo.FieldName,
					LocalName: utils.GetLangValue(h.lang, fieldInfo.FieldName, ""),
					OldValue:  value(o, fieldInfo, h.uMap, h.lang),
					NewValue:  value(n, fieldInfo, h.uMap, h.lang),
				})
			}

			continue

		}

		// 旧数据中字段不存在的场合
		changes = append(changes, Change{
			FieldID:   field,
			FieldName: fieldInfo.FieldName,
			LocalName: utils.GetLangValue(h.lang, fieldInfo.FieldName, ""),
			OldValue:  "",
			NewValue:  value(n, fieldInfo, h.uMap, h.lang),
		})
	}

	for k, v := range data.hs {

		// 判断当前字段是否存在，不存在的场合，说明传入有误
		fieldInfo, exist := h.fMap[k]
		if exist {
			if v.DataType == "options" {
				v1 := cast.ToString(v.Value)
				key := utils.GetOptionKey(fieldInfo.AppID, fieldInfo.OptionID, v1)
				v1 = utils.GetLangValue(h.lang, key, "")
				data.hs[k] = &Value{
					DataType: v.DataType,
					Value:    v1,
				}
			}
			if v.DataType == "user" {
				jsonBytes, err := json.Marshal(v.Value)
				if err != nil {
					continue
				}
				var vList []string
				err = json.Unmarshal(jsonBytes, &vList)
				if err != nil {
					continue
				}

				var uList []string
				for _, uid := range vList {
					uList = append(uList, h.uMap[uid])
				}

				data.hs[k] = &Value{
					DataType: v.DataType,
					Value:    uList,
				}
			}
		}

	}

	// 有变更的值，存储起来
	data.changes = changes
	return nil
}

// 写入相关的履历数据
func (h *HistoryServer) Commit() error {
	client := database.New()

	client.Database(database.GetDBName(h.db)).CreateCollection(context.TODO(), HistoriesCollection)
	client.Database(database.GetDBName(h.db)).CreateCollection(context.TODO(), FieldHistoriesCollection)

	c := client.Database(database.GetDBName(h.db)).Collection(HistoriesCollection)
	fc := client.Database(database.GetDBName(h.db)).Collection(FieldHistoriesCollection)

	var dhs []mongo.WriteModel
	var fhs []mongo.WriteModel
	// 写入数据
	for index, data := range h.datas {

		now := time.Now()

		id := primitive.NewObjectID()
		hid := timex.Timestamp() + index

		// 将change和after存入数据库
		history := hs{
			ID:          id,
			HistoryID:   hid,
			HistoryType: data.hsType,
			DatastoreID: h.did,
			ItemID:      data.itemId,
			FixedItems:  data.hs,
			CreatedAt:   now,
			CreatedBy:   h.uid,
		}

		hsData := mongo.NewInsertOneModel()
		hsData.SetDocument(history)
		dhs = append(dhs, hsData)

		// 字段变更履历
		for _, c := range data.changes {
			fh := FieldHistory{
				ID:          primitive.NewObjectID(),
				HistoryID:   hid,
				HistoryType: data.hsType,
				DatastoreID: h.did,
				ItemID:      data.itemId,
				FieldID:     c.FieldID,
				FieldName:   c.FieldName,
				LocalName:   c.LocalName,
				OldValue:    c.OldValue,
				NewValue:    c.NewValue,
				CreatedAt:   now,
				CreatedBy:   h.uid,
			}

			fsData := mongo.NewInsertOneModel()
			fsData.SetDocument(fh)
			fhs = append(fhs, fsData)
		}
	}

	// 写入数据
	if len(dhs) > 0 {
		hs, err := c.BulkWrite(h.sc, dhs)
		if err != nil {
			utils.ErrorLog("Commit", err.Error())
			return err
		}
		fmt.Printf("add data history %v \n", hs)
	}
	if len(fhs) > 0 {
		fs, err := fc.BulkWrite(h.sc, fhs)
		if err != nil {
			utils.ErrorLog("Commit", err.Error())
			return err
		}
		fmt.Printf("add field history %v \n", fs)
	}
	return nil
}

// 相等为true，不等为false
func hsCompare(newValue *Value, oldValue *Value) bool {
	switch newValue.DataType {
	case "text", "textarea", "time", "options", "lookup":
		n := cast.ToString(newValue.Value)
		o := cast.ToString(oldValue.Value)

		return n == o
	case "number":
		n := cast.ToFloat64(newValue.Value)
		o := cast.ToFloat64(oldValue.Value)

		return n == o
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
			return true
		}
		return false
	case "switch":
		n := cast.ToBool(newValue.Value)
		o := cast.ToBool(oldValue.Value)

		return n == o
	case "user":
		n := cast.ToStringSlice(newValue.Value)

		var o []string
		err := mapstructure.Decode(oldValue.Value, &o)
		if err != nil {
			return false
		}

		return stringSliceEqual(n, o)
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

		return fileSliceEqual(new, old)
	default:
		return true
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func value(value *Value, field Field, uMap map[string]string, lang *language.Language) (v string) {
	switch value.DataType {
	case "text", "textarea", "number", "autonum", "time", "switch", "function", "lookup":
		return cast.ToString(value.Value)
	case "options":
		v := cast.ToString(value.Value)
		key := utils.GetOptionKey(field.AppID, field.OptionID, v)
		v = utils.GetLangValue(lang, key, "")
		return v
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
	case "user":

		jsonBytes, err := json.Marshal(value.Value)
		if err != nil {
			return ""
		}
		var v []string
		err = json.Unmarshal(jsonBytes, &v)
		if err != nil {
			return ""
		}

		var uList []string
		for _, uid := range v {
			uList = append(uList, uMap[uid])
		}

		return strings.Join(uList, ",")
	case "file":
		var fs []File
		err := json.Unmarshal([]byte(value.Value.(string)), &fs)
		if err != nil {
			return ""
		}

		var names []string
		for _, f := range fs {
			names = append(names, f.Name)
		}

		return strings.Join(names, ",")
	default:
		jsonBytes, _ := json.Marshal(value.Value)
		return string(jsonBytes)
	}
}

func copyMap(m map[string]*Value) map[string]*Value {
	result := make(map[string]*Value, len(m))
	for k, v := range m {
		result[k] = v
	}

	return result
}
