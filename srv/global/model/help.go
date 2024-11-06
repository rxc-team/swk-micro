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
	"rxcsoft.cn/pit3/srv/global/proto/help"
	"rxcsoft.cn/pit3/srv/global/utils"
	"rxcsoft.cn/utils/helpers"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	HelpsCollection = "helps"
)

// Help 帮助文档
type Help struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	HelpID    string             `json:"help_id" bson:"help_id"`
	Title     string             `json:"title" bson:"title"`
	Type      string             `json:"type" bson:"type"`
	Content   string             `json:"content" bson:"content"`
	Images    []string           `json:"images" bson:"images"`
	Tags      []string           `json:"tags" bson:"tags"`
	LangCd    string             `json:"lang_cd" bson:"lang_cd"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy string             `json:"created_by" bson:"created_by"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy string             `json:"updated_by" bson:"updated_by"`
}

// ToProto 转换为proto数据
func (h *Help) ToProto() *help.Help {
	return &help.Help{
		HelpId:    h.HelpID,
		Title:     h.Title,
		Type:      h.Type,
		Content:   h.Content,
		Images:    h.Images,
		Tags:      h.Tags,
		LangCd:    h.LangCd,
		CreatedAt: h.CreatedAt.String(),
		CreatedBy: h.CreatedBy,
		UpdatedAt: h.UpdatedAt.String(),
		UpdatedBy: h.UpdatedBy,
	}
}

// FindHelp 获取单个帮助文档
func FindHelp(db, helpID string) (h Help, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HelpsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Help
	objectID, err := primitive.ObjectIDFromHex(helpID)
	if err != nil {
		utils.ErrorLog("error FindHelp", err.Error())
		return result, err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindHelp", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindHelp", err.Error())
		return result, err
	}
	return result, nil
}

// FindHelps 获取多个帮助文档
func FindHelps(db, title, helpType, tag, langCd string) (h []Help, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HelpsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	// 帮助文档标题非空
	if title != "" {
		query["title"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(title), Options: "m"}}
	}
	// 帮助文档类型非空
	if helpType != "" {
		query["type"] = helpType
	}
	// 帮助文档标签非空
	if tag != "" {
		query["tags"] = bson.M{"$in": []string{tag}}
	}
	// 登录语言代号非空
	if langCd != "" {
		query["lang_cd"] = langCd
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindHelps", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Help
	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindHelps", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var h Help
		err := cur.Decode(&h)
		if err != nil {
			utils.ErrorLog("error FindHelps", err.Error())
			return nil, err
		}
		result = append(result, h)
	}

	return result, nil
}

// FindTags 获取所有不重复帮助文档标签
func FindTags(db string) (tags []string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HelpsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result []string

	query := bson.M{}

	cur, err := c.Distinct(ctx, "tags", query)
	if err != nil {
		utils.ErrorLog("error FindTags", err.Error())
		return nil, err
	}
	for _, c := range cur {
		result = append(result, c.(string))
	}

	return result, nil
}

// AddHelp 添加帮助文档
func AddHelp(db string, h *Help) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HelpsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if len(h.Images) == 0 {
		h.Images = make([]string, 0)
	}
	if len(h.Tags) == 0 {
		h.Tags = make([]string, 0)
	}
	h.ID = primitive.NewObjectID()
	h.HelpID = h.ID.Hex()

	queryJSON, _ := json.Marshal(h)
	utils.DebugLog("AddHelp", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, h)
	if err != nil {
		utils.ErrorLog("error AddHelp", err.Error())
		return h.HelpID, err
	}
	return h.HelpID, nil
}

// ModifyHelp 更新帮助文档
func ModifyHelp(db string, h *Help) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HelpsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(h.HelpID)
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": h.UpdatedAt,
		"updated_by": h.UpdatedBy,
	}

	// 帮助文档标题不为空的场合
	if h.Title != "" {
		change["title"] = h.Title
	}
	// 帮助文档类型不为空的场合
	if h.Type != "" {
		change["type"] = h.Type
	}
	// 帮助文档内容不为空的场合
	if h.Content != "" {
		change["content"] = h.Content
	}
	// 帮助文档图片不为空的场合
	if len(h.Images) > 0 {
		change["images"] = h.Images
	} else {
		change["images"] = []string{}
	}
	// 帮助文档标签不为空的场合
	if len(h.Tags) > 0 {
		change["tags"] = h.Tags
	}
	// 登录语言代号不为空的场合
	if h.LangCd != "" {
		change["lang_cd"] = h.LangCd
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyHelp", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyHelp", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyHelp", err.Error())
		return err
	}

	return nil
}

// DeleteHelp 硬删除帮助文档
func DeleteHelp(db, helpID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HelpsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(helpID)
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteHelp", fmt.Sprintf("update: [ %s ]", queryJSON))

	_, err := c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteHelp", err.Error())
		return err
	}
	return nil
}

// DeleteHelps 硬删除多个帮助文档
func DeleteHelps(db string, helpIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(HelpsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteHelps", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteHelps", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, helpID := range helpIDList {
			objectID, err := primitive.ObjectIDFromHex(helpID)
			if err != nil {
				utils.ErrorLog("error DeleteHelps", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteHelps", fmt.Sprintf("update: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error DeleteHelps", err.Error())
				return err
			}

		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteHelps", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteCustomers", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}
