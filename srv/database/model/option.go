package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/database/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	OptionsCollection = "options"
)

// Max 最大顺
type Max struct {
	ID       string `json:"id" bson:"_id"`
	MaxValue string `json:"max_value" bson:"max_value"`
}

type (
	// Option 选项
	Option struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		OptionID    string             `json:"option_id" bson:"option_id"`
		OptionValue string             `json:"option_value" bson:"option_value"`
		OptionLabel string             `json:"option_label" bson:"option_label"`
		OptionOrder int32              `json:"option_order" bson:"option_order"`
		OptionName  string             `json:"option_name" bson:"option_name"`
		OptionMemo  string             `json:"option_memo" bson:"option_memo"`
		AppID       string             `json:"app_id" bson:"app_id"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
		DeletedAt   time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy   string             `json:"deleted_by" bson:"deleted_by"`
	}
)

// ToProto 转换为proto数据
func (o *Option) ToProto() *option.Option {
	return &option.Option{
		OptionId:    o.OptionID,
		OptionValue: o.OptionValue,
		OptionLabel: o.OptionLabel,
		OptionOrder: o.OptionOrder,
		OptionName:  o.OptionName,
		OptionMemo:  o.OptionMemo,
		AppId:       o.AppID,
		CreatedAt:   o.CreatedAt.String(),
		CreatedBy:   o.CreatedBy,
		UpdatedAt:   o.UpdatedAt.String(),
		UpdatedBy:   o.UpdatedBy,
		DeletedAt:   o.DeletedAt.String(),
		DeletedBy:   o.DeletedBy,
	}
}

// FindOptions 获取所有选项
func FindOptions(db, appID, optionName, optionMemeo, invalidatedIn string) (r []Option, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 通过名称模糊查询
	nameRegex := bson.M{"$match": bson.M{
		"option_name": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(optionName), Options: "m"}},
	}}
	// 通过memeo模糊查询
	memoRegex := bson.M{"$match": bson.M{
		"option_memo": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(optionMemeo), Options: "m"}},
	}}

	match := bson.M{
		"deleted_by": "",
		"app_id":     appID,
	}

	if invalidatedIn != "" {
		delete(match, "deleted_by")
	}

	query := []bson.M{
		{"$match": match},
	}

	// 文件名不为空的场合，添加到查询条件中
	if optionName != "" {
		query = append(query, nameRegex)
	}
	if optionMemeo != "" {
		query = append(query, memoRegex)
	}

	group := []bson.M{
		{"$group": bson.M{
			"_id": bson.M{
				"option_id": "$option_id",
			},
			"option_name": bson.M{
				"$first": "$option_name",
			},
			"option_memo": bson.M{
				"$first": "$option_memo",
			},
			"created_at": bson.M{
				"$first": "$created_at",
			},
			"created_by": bson.M{
				"$first": "$created_by",
			},
			"updated_at": bson.M{
				"$last": "$updated_at",
			},
			"updated_by": bson.M{
				"$last": "$updated_by",
			},
			"deleted_at": bson.M{
				"$last": "$deleted_at",
			},
			"deleted_by": bson.M{
				"$last": "$deleted_by",
			},
		}},
		{"$project": bson.M{
			"_id":         0,
			"option_id":   "$_id.option_id",
			"option_name": "$option_name",
			"option_memo": "$option_memo",
			"created_at":  "$created_at",
			"created_by":  "$created_by",
			"updated_at":  "$updated_at",
			"updated_by":  "$updated_by",
			"deleted_at":  "$deleted_at",
			"deleted_by":  "$deleted_by",
		}},
		{
			"$sort": bson.M{
				"option_id": 1,
			},
		},
	}

	query = append(query, group...)

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindOptions", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Option
	cur, err := c.Aggregate(ctx, query)
	if err != nil {
		utils.ErrorLog("FindOptions", err.Error())
		return result, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var opt Option
		err := cur.Decode(&opt)
		if err != nil {
			utils.ErrorLog("FindOptions", err.Error())
			return result, err
		}
		result = append(result, opt)
	}

	return result, nil
}

// FindOptionLabels 获取全部的选项数据
func FindOptionLabels(db, appID, optionName, optionMemeo, invalidatedIn string) (r []Option, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appID,
	}
	if optionName != "" {
		query["option_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(optionName), Options: "m"}}
	}
	if optionMemeo != "" {
		query["option_memo"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(optionMemeo), Options: "m"}}
	}
	if invalidatedIn == "" {
		query["deleted_by"] = ""
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindOptionLabels", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Option
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("FindOptionLabels", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var opt Option
		err := cur.Decode(&opt)
		if err != nil {
			utils.ErrorLog("FindOptionLabels", err.Error())
			return nil, err
		}
		result = append(result, opt)
	}

	return result, nil
}

// FindOptionLable 通过选项ID获取一组数据
func FindOptionLable(db, appID, optionID, optionValue string) (r Option, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"app_id":       appID,
		"option_id":    optionID,
		"option_value": optionValue,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindOptionLable", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Option

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("FindOptionLable", err.Error())
		return result, err
	}
	return result, nil
}

// FindOption 通过选项ID获取一组数据
func FindOption(db, appID, optionID, invalid string) (r []Option, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"option_id":  optionID,
		"app_id":     appID,
	}

	if invalid != "" {
		delete(query, "deleted_by")
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindOption", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Option
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("FindOption", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var opt Option
		err := cur.Decode(&opt)
		if err != nil {
			utils.ErrorLog("FindOption", err.Error())
			return nil, err
		}
		result = append(result, opt)
	}

	return result, nil
}

// AddOption 添加选项
func AddOption(db string, o *Option, isNew bool) (id, option string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	o.ID = primitive.NewObjectID()
	if isNew {
		seq, err := getOptionOrder(db, o.AppID)
		if err != nil {
			utils.ErrorLog("AddOption", err.Error())
			return "", "", err
		}
		if seq > 999999 {
			return "", "", errors.New("SequenceValue too large")
		}
		o.OptionID = fmt.Sprintf("%06d", seq)
	}
	o.OptionName = GetOptionNameKey(o.AppID, o.OptionID)
	o.OptionLabel = GetOptionLabelNameKey(o.AppID, o.OptionID, o.OptionValue)

	queryJSON, _ := json.Marshal(o)
	utils.DebugLog("AddOption", fmt.Sprintf("Option: [ %s ]", queryJSON))

	if _, err := c.InsertOne(ctx, o); err != nil {
		utils.ErrorLog("AddOption", err.Error())
		return "", "", err
	}

	return o.ID.Hex(), o.OptionID, nil
}

// DeleteOptionChild 删除某个选项组下的某个值数据
func DeleteOptionChild(db, appID, optionID, optionValue, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appID,
		"option_id":    optionID,
		"option_value": optionValue,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteOptionChild", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteOptionChild", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err := c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("DeleteOptionChild", err.Error())
		return err
	}

	return nil
}

// HardDeleteOptionChild 物理删除某个选项组下的某个值数据
func HardDeleteOptionChild(db, appID, optionID, optionValue string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appID,
		"option_id":    optionID,
		"option_value": optionValue,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("HardDeleteOptionChild", fmt.Sprintf("query: [ %s ]", queryJSON))

	if _, err := c.DeleteOne(ctx, query); err != nil {
		utils.ErrorLog("HardDeleteOptionChild", err.Error())
		return err
	}

	return nil
}

// DeleteOption 删除某个选项
func DeleteOption(db, appID, optionID, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":    appID,
		"option_id": optionID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteOption", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteOption", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err := c.UpdateMany(ctx, query, update); err != nil {
		utils.ErrorLog("DeleteOption", err.Error())
		return err
	}

	return nil
}

// DeleteAppOption APP下的所有选项
func DeleteAppOption(db, appID, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteAppOption", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteAppOption", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err := c.UpdateMany(ctx, query, update); err != nil {
		utils.ErrorLog("DeleteAppOption", err.Error())
		return err
	}

	return nil
}

// DeleteSelectOptions 删除选中选项
func DeleteSelectOptions(db, appID string, optionIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteSelectOptions", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteSelectOptions", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, optionID := range optionIDList {
			query := bson.M{
				"app_id":    appID,
				"option_id": optionID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSelectOptions", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectOptions", fmt.Sprintf("update: [ %s ]", updateJSON))

			_, err = c.UpdateMany(sc, query, update)
			if err != nil {
				utils.ErrorLog("DeleteSelectOptions", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteSelectOptions", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteSelectOptions", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// HardDeleteOptions 物理删除选中选项
func HardDeleteOptions(db, appID string, optionIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("HardDeleteOptions", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("HardDeleteOptions", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, optionID := range optionIDList {
			query := bson.M{
				"app_id":    appID,
				"option_id": optionID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteOptions", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteMany(sc, query)
			if err != nil {
				utils.ErrorLog("HardDeleteOptions", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("HardDeleteOptions", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("HardDeleteOptions", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// RecoverSelectOptions 恢复选中的选项
func RecoverSelectOptions(db, appID string, optionIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("RecoverSelectOptions", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("RecoverSelectOptions", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, optionID := range optionIDList {
			query := bson.M{
				"app_id":    appID,
				"option_id": optionID,
			}

			update := bson.M{"$set": bson.M{
				"updated_at": time.Now(),
				"updated_by": userID,
				"deleted_by": "",
			}}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("RecoverSelectOptions", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateJSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectOptions", fmt.Sprintf("update: [ %s ]", updateJSON))

			_, err = c.UpdateMany(sc, query, update)
			if err != nil {
				utils.ErrorLog("RecoverSelectOptions", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("RecoverSelectOptions", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("RecoverSelectOptions", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}

// RecoverOptionChild 恢复某个选项组下的某个值数据
func RecoverOptionChild(db, appID, optionID, optionValue, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(OptionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":       appID,
		"option_id":    optionID,
		"option_value": optionValue,
	}

	update := bson.M{"$set": bson.M{
		"updated_at": time.Now(),
		"updated_by": userID,
		"deleted_by": "",
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("RecoverOptionChild", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("RecoverOptionChild", fmt.Sprintf("update: [ %s ]", updateJSON))

	if _, err := c.UpdateOne(ctx, query, update); err != nil {
		utils.ErrorLog("RecoverOptionChild", err.Error())
		return err
	}

	return nil
}
