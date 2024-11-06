package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"rxcsoft.cn/pit3/srv/database/proto/print"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// PrintsCollection prints collection
	PrintsCollection = "prints"
)

// 结构体
type (
	// Print 台账打印设置
	Print struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		AppID       string             `json:"app_id" bson:"app_id"`
		DatastoreID string             `json:"datastore_id" bson:"datastore_id"`
		Page        string             `json:"page" bson:"page"`
		Orientation string             `json:"orientation" bson:"orientation"`
		CheckField  string             `json:"check_field" bson:"check_field"`
		TitleWidth  int64              `json:"title_width" bson:"title_width"`
		Fields      []*PrintField      `json:"fields" bson:"fields"`
		ShowSign    bool               `json:"show_sign" bson:"show_sign"`
		SignName1   string             `json:"sign_name1" bson:"sign_name1"`
		SignName2   string             `json:"sign_name2" bson:"sign_name2"`
		ShowSystem  bool               `json:"show_system" bson:"show_system"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	}

	// PrintField 打印用字段情报
	PrintField struct {
		FieldID      string `json:"field_id" bson:"field_id"`
		FieldName    string `json:"field_name" bson:"field_name"`
		FieldType    string `json:"field_type" bson:"field_type"`
		IsImage      bool   `json:"is_image" bson:"is_image"`
		IsCheckImage bool   `json:"is_check_image" bson:"is_check_image"`
		AsTitle      bool   `json:"as_title" bson:"as_title"`
		Cols         int64  `json:"cols" bson:"cols"`
		Rows         int64  `json:"rows" bson:"rows"`
		X            int64  `json:"x" bson:"x"`
		Y            int64  `json:"y" bson:"y"`
		Width        int64  `json:"width" bson:"width"`
		Precision    int64  `json:"precision" bson:"precision"`
	}
	// PrintModifyParam 修改台账打印设置参数
	PrintModifyParam struct {
		AppID       string        `json:"app_id" bson:"app_id"`
		DatastoreID string        `json:"datastore_id" bson:"datastore_id"`
		Page        string        `json:"page" bson:"page"`
		Orientation string        `json:"orientation" bson:"orientation"`
		TitleWidth  int64         `json:"title_width" bson:"title_width"`
		CheckField  string        `json:"check_field" bson:"check_field"`
		Fields      []*PrintField `json:"fields" bson:"fields"`
		ShowSign    string        `json:"show_sign" bson:"show_sign"`
		SignName1   string        `json:"sign_name1" bson:"sign_name1"`
		SignName2   string        `json:"sign_name2" bson:"sign_name2"`
		ShowSystem  string        `json:"show_system" bson:"show_system"`
		UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string        `json:"updated_by" bson:"updated_by"`
	}
)

// ToProto 转换为proto数据
func (pf *PrintField) ToProto() *print.PrintField {
	return &print.PrintField{
		FieldId:   pf.FieldID,
		FieldName: pf.FieldName,
		FieldType: pf.FieldType,
		IsImage:   pf.IsImage,
		AsTitle:   pf.AsTitle,
		Cols:      pf.Cols,
		Rows:      pf.Rows,
		X:         pf.X,
		Y:         pf.Y,
		Width:     pf.Width,
		Precision: pf.Precision,
	}
}

// ToProto 转换为proto数据
func (p *Print) ToProto() *print.Print {
	var fs []*print.PrintField
	for _, f := range p.Fields {
		fs = append(fs, f.ToProto())
	}

	return &print.Print{
		AppId:       p.AppID,
		DatastoreId: p.DatastoreID,
		Page:        p.Page,
		Orientation: p.Orientation,
		CheckField:  p.CheckField,
		TitleWidth:  p.TitleWidth,
		Fields:      fs,
		ShowSign:    p.ShowSign,
		SignName1:   p.SignName1,
		SignName2:   p.SignName2,
		ShowSystem:  p.ShowSystem,
		CreatedAt:   p.CreatedAt.String(),
		CreatedBy:   p.CreatedBy,
		UpdatedAt:   p.UpdatedAt.String(),
		UpdatedBy:   p.UpdatedBy,
	}
}

// FindPrints 获取台账打印设置
func FindPrints(db, appId, DatastoreId string) (ps []Print, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(PrintsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	// appId不为空的场合，添加到查询条件中
	if appId != "" {
		query["app_id"] = appId
	}

	// DatastoreId不为空的场合，添加到查询条件中
	if DatastoreId != "" {
		query["datastore_id"] = DatastoreId
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindPrints", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Print
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("FindPrints", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var pr Print
		err := cur.Decode(&pr)
		if err != nil {
			utils.ErrorLog("FindPrints", err.Error())
			return nil, err
		}
		result = append(result, pr)
	}

	return result, nil
}

// FindPrint 通过AppID和台账ID获取台账打印设置
func FindPrint(db, appId, DatastoreId string) (p Print, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(PrintsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Print

	query := bson.M{
		"app_id":       appId,
		"datastore_id": DatastoreId,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindPrint", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			return result, mongo.ErrNoDocuments
		}
		utils.ErrorLog("FindPrint", err.Error())
		return result, err
	}

	return result, nil
}

// AddPrint 添加台账打印设置
func AddPrint(db string, p *Print) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(PrintsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p.ID = primitive.NewObjectID()

	queryJSON, _ := json.Marshal(p)
	utils.DebugLog("AddPrint", fmt.Sprintf("Print: [ %s ]", queryJSON))

	if _, err = c.InsertOne(ctx, p); err != nil {
		utils.ErrorLog("AddPrint", err.Error())
		return err
	}

	return nil
}

// ModifyPrint 修改台账打印设置
func ModifyPrint(db string, p *PrintModifyParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(PrintsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       p.AppID,
		"datastore_id": p.DatastoreID,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": p.UpdatedBy,
	}

	// 打印字段不为空的场合
	if len(p.Fields) != 0 {
		change["fields"] = p.Fields
	}

	// 是否显示签名不为空的场合
	if p.ShowSign != "" {
		result, err := strconv.ParseBool(p.ShowSign)
		if err == nil {
			change["show_sign"] = result
		}
	}

	// 打印纸张类型，A4，A3等
	change["page"] = p.Page
	// 打印方向，L 横屏 P 竖屏
	change["orientation"] = p.Orientation
	// 盘点图片字段
	change["check_field"] = p.CheckField
	// 标题宽度
	change["title_width"] = p.TitleWidth

	// 签名1
	change["sign_name1"] = p.SignName1

	// 签名2
	change["sign_name2"] = p.SignName2

	// 是否显示系统情报不为空的场合
	if p.ShowSystem != "" {
		result, err := strconv.ParseBool(p.ShowSystem)
		if err == nil {
			change["show_system"] = result
		}
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyPrint", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyPrint", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err = c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("ModifyPrint", err.Error())
		return err
	}

	return nil
}

// HardDeletePrints 物理删除台账打印设置
func HardDeletePrints(db, appId, DatastoreId string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(PrintsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appId,
	}

	if DatastoreId != "" {
		query["datastore_id"] = DatastoreId
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("HardDeletePrints", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err := c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("HardDeletePrints", err.Error())
		return err
	}

	return nil
}
