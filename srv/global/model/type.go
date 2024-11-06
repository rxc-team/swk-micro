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
	types "rxcsoft.cn/pit3/srv/global/proto/help-type"
	"rxcsoft.cn/pit3/srv/global/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	TypesCollection = "help_types"
)

// Type 帮助文档类型
type Type struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	TypeID    string             `json:"type_id" bson:"type_id"`
	TypeName  string             `json:"type_name" bson:"type_name"`
	Show      string             `json:"show" bson:"show"`
	Icon      string             `json:"icon" bson:"icon"`
	LangCd    string             `json:"lang_cd" bson:"lang_cd"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy string             `json:"created_by" bson:"created_by"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy string             `json:"updated_by" bson:"updated_by"`
}

// ToProto 转换为proto数据
func (t *Type) ToProto() *types.Type {
	return &types.Type{
		TypeId:    t.TypeID,
		TypeName:  t.TypeName,
		Show:      t.Show,
		Icon:      t.Icon,
		LangCd:    t.LangCd,
		CreatedAt: t.CreatedAt.String(),
		CreatedBy: t.CreatedBy,
		UpdatedAt: t.UpdatedAt.String(),
		UpdatedBy: t.UpdatedBy,
	}
}

// FindType 获取单个帮助文档类型
func FindType(db, typeID string) (t Type, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(TypesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Type
	objectID, err := primitive.ObjectIDFromHex(typeID)
	if err != nil {
		utils.ErrorLog("error FindType", err.Error())
		return result, err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindType", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindType", err.Error())
		return result, err
	}
	return result, nil
}

// FindTypes 获取多个帮助文档类型
func FindTypes(db, typeName, show, langCd string) (t []Type, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(TypesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	// 类型名称非空
	if typeName != "" {
		query["type_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(typeName), Options: "m"}}
	}
	// 是否显示在帮助概览画面非空
	if show != "" {
		query["show"] = show
	}
	// 登录语言代号非空
	if langCd != "" {
		query["lang_cd"] = langCd
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindTypes", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Type
	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindTypes", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var t Type
		err := cur.Decode(&t)
		if err != nil {
			utils.ErrorLog("error FindTypes", err.Error())
			return nil, err
		}
		result = append(result, t)
	}

	return result, nil
}

// AddType 添加帮助文档类型
func AddType(db string, t *Type) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(TypesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.ID = primitive.NewObjectID()
	t.TypeID = t.ID.Hex()

	queryJSON, _ := json.Marshal(t)
	utils.DebugLog("AddType", fmt.Sprintf("Type: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, t)
	if err != nil {
		utils.ErrorLog("error AddType", err.Error())
		return t.TypeID, err
	}

	return t.TypeID, nil
}

// ModifyType 更新帮助文档类型
func ModifyType(db string, t *Type) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(TypesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(t.TypeID)
	if err != nil {
		utils.ErrorLog("error ModifyType", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": t.UpdatedAt,
		"updated_by": t.UpdatedBy,
	}

	// 类型名称不为空的场合
	if t.TypeName != "" {
		change["type_name"] = t.TypeName
	}
	// 是否显示在帮助概览画面不为空的场合
	if t.Show != "" {
		change["show"] = t.Show
	}
	// 类型图标不为空的场合
	if t.Icon != "" {
		change["icon"] = t.Icon
	}
	// 登录语言代号不为空的场合
	if t.LangCd != "" {
		change["lang_cd"] = t.LangCd
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyType", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyType", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyType", err.Error())
		return err
	}

	return nil
}

// DeleteType 硬删除帮助文档类型
func DeleteType(db, typeID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(TypesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(typeID)
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteType", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err := c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteType", err.Error())
		return err
	}
	return nil
}

// DeleteTypes 硬删除多个帮助文档类型
func DeleteTypes(db string, typeIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(TypesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteTypes", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteTypes", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, typeID := range typeIDList {
			objectID, err := primitive.ObjectIDFromHex(typeID)
			if err != nil {
				utils.ErrorLog("error DeleteTypes", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteTypes", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(ctx, query)
			if err != nil {
				utils.ErrorLog("error DeleteTypes", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteTypes", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteTypes", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}
