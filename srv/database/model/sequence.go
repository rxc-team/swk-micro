package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// SequencesCollection sequences collection
	SequencesCollection = "sequences"
)

// autoNum
func autoNum(sc mongo.SessionContext, db string, f Field) (string, error) {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(f.DatastoreID)
	seq.WriteString("_fields_")
	seq.WriteString(f.FieldID)
	seq.WriteString("_auto")

	seqName := seq.String()

	num, err := getSeqWithSession(sc, db, seqName, 1)
	if err != nil {
		utils.ErrorLog("autoNum", err.Error())
		return "", err
	}

	over := autoNumCheck(f.DisplayDigits, num)
	if over {
		// 如果超过了，重置为原始值
		err := setSeq(db, seqName, num-1)
		if err != nil {
			utils.ErrorLog("autoNum", err.Error())
			return "", err
		}
		return "", errors.New("[自動採番]フィールドが設定された最大桁数を超える")
	}

	return fmt.Sprintf("%s%0*d", f.Prefix, f.DisplayDigits, num), nil
}

func keiyakunoAuto(sc mongo.SessionContext, db string, DatastoreID string) (string, error) {
	seqName := DatastoreID + "_keiyakuno_auto"

	result := createKeiyakunoSeq(db, seqName)
	if result != nil {
		utils.ErrorLog("createKeiyakunoSeq", result.Error())
		return "", result
	}

	num, err := getSeqWithSession(sc, db, seqName, 1)
	if err != nil {
		utils.ErrorLog("autoNum", err.Error())
		return "", err
	}

	over := autoNumCheck(10, num)
	if over {
		// 如果超过了，重置为原始值
		err := setSeq(db, seqName, num-1)
		if err != nil {
			utils.ErrorLog("autoNum", err.Error())
			return "", err
		}
		return "", errors.New("[自動採番]フィールドが設定された最大桁数を超える")
	}
	return fmt.Sprintf("%s%0*d", "auto_", 10, num), nil
}

func createKeiyakunoSeq(db, seqName string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var s Sequence
	s.ID = seqName
	s.SequenceValue = 0

	var existing Sequence
	query := bson.M{"_id": seqName}
	err = c.FindOne(ctx, query).Decode(&existing)
	if err == nil {
		// 已存在采番内容
		return nil
	} else {
		if _, err = c.InsertOne(ctx, s); err != nil {
			utils.ErrorLog("createSeq", err.Error())
			return nil
		}
	}
	return nil
}

// resetAutoNum
func resetAutoNum(db string, f Field) error {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(f.DatastoreID)
	seq.WriteString("_fields_")
	seq.WriteString(f.FieldID)
	seq.WriteString("_auto")

	seqName := seq.String()

	err := setSeq(db, seqName, 0)
	if err != nil {
		return err
	}

	return nil
}

// autoNumList
func autoNumList(db string, f *Field, step int) ([]string, error) {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(f.DatastoreID)
	seq.WriteString("_fields_")
	seq.WriteString(f.FieldID)
	seq.WriteString("_auto")

	seqName := seq.String()

	num, err := getSeq(db, seqName, step)
	if err != nil {
		utils.ErrorLog("autoNum", err.Error())
		return nil, err
	}

	over := autoNumCheck(f.DisplayDigits, num)
	if over {
		// 如果超过了，重置为原始值
		err := setSeq(db, seqName, num-int64(step))
		if err != nil {
			utils.ErrorLog("autoNum", err.Error())
			return nil, err
		}
		return nil, errors.New("[自動採番]フィールドが設定された最大桁数を超える")
	}

	var seqList []string

	for i := num - int64(step) + 1; i <= num; i++ {
		seqList = append(seqList, fmt.Sprintf("%s%0*d", f.Prefix, f.DisplayDigits, i))
	}

	return seqList, nil
}

// autoNumList
func autoNumListWithSession(sc mongo.SessionContext, db string, f *Field, step int) ([]string, error) {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(f.DatastoreID)
	seq.WriteString("_fields_")
	seq.WriteString(f.FieldID)
	seq.WriteString("_auto")

	seqName := seq.String()

	num, err := getSeqWithSession(sc, db, seqName, step)
	if err != nil {
		// utils.ErrorLog("autoNum", err.Error())
		return nil, err
	}

	over := autoNumCheck(f.DisplayDigits, num)
	if over {
		// 如果超过了，重置为原始值
		err := setSeqWithSession(sc, db, seqName, num-int64(step))
		if err != nil {
			// utils.ErrorLog("autoNum", err.Error())
			return nil, err
		}
		return nil, errors.New("[自動採番]フィールドが設定された最大桁数を超える")
	}

	var seqList []string

	for i := num - int64(step) + 1; i <= num; i++ {
		seqList = append(seqList, fmt.Sprintf("%s%0*d", f.Prefix, f.DisplayDigits, i))
	}

	return seqList, nil
}

// autoNumCheck 验证自增字段值是否已超过该字段表示位数
func autoNumCheck(displayDigits, value int64) (isOver bool) {
	// 获取自增字段情报
	if displayDigits != 0 {
		if int64(len(strconv.FormatInt(value, 10))) > displayDigits {
			return true
		}
	}
	return false
}

// createAutoSeq 创建自动採番字段序列
func createAutoSeq(db, datastoreId, fieldId string) (err error) {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(datastoreId)
	seq.WriteString("_fields_")
	seq.WriteString(fieldId)
	seq.WriteString("_auto")

	seqName := seq.String()

	if err := createSeq(db, seqName); err != nil {
		utils.ErrorLog("createAutoSeq", err.Error())
		return err
	}

	return nil
}

// deleteAutoSeq 创建自动採番字段序列
func deleteAutoSeq(db, datastoreId, fieldId string) (err error) {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(datastoreId)
	seq.WriteString("_fields_")
	seq.WriteString(fieldId)
	seq.WriteString("_auto")

	seqName := seq.String()

	if err := delSeq(db, seqName); err != nil {
		utils.ErrorLog("deleteAutoSeq", err.Error())
		return err
	}

	return nil
}

// createFieldsOrder 创建字段排序seq
func createFieldsOrder(db, datastoreID string) error {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(datastoreID)
	seq.WriteString("_fields__displayorder")

	seqName := seq.String()

	err := createSeq(db, seqName)

	if err != nil {
		return err
	}

	return nil
}

// getFieldsOrder 字段排序
func getFieldsOrder(db, datastoreID string) (int64, error) {
	seq := strings.Builder{}
	seq.WriteString("datastore_")
	seq.WriteString(datastoreID)
	seq.WriteString("_fields__displayorder")

	seqName := seq.String()

	num, err := getSeq(db, seqName, 1)
	if err != nil {
		utils.ErrorLog("getFieldsOrder", err.Error())
		return 0, err
	}

	return num, nil
}

// getOptionOrder 字段排序
func getOptionOrder(db, appId string) (int64, error) {
	seq := strings.Builder{}
	seq.WriteString("option_")
	seq.WriteString(appId)

	seqName := seq.String()

	num, err := getSeq(db, seqName, 1)
	if err != nil {
		utils.ErrorLog("getFieldsOrder", err.Error())
		return 0, err
	}

	return num, nil
}

// createSeq 创建seq，默认值是0
func createSeq(db, seqName string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var s Sequence
	s.ID = seqName
	s.SequenceValue = 0

	if _, err = c.InsertOne(ctx, s); err != nil {
		utils.ErrorLog("createSeq", err.Error())
		return err
	}

	return nil
}

// getSeq 获取序列值,给定序列的step，默认step是1
func getSeq(db, seqName string, step int) (num int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"_id": seqName,
	}

	change := bson.M{
		"$inc": bson.M{
			"sequence_value": step,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(1)
	var result Sequence
	if err := c.FindOneAndUpdate(ctx, query, change, opts).Decode(&result); err != nil {
		utils.ErrorLog("getSeq", err.Error())
		return result.SequenceValue, err
	}

	return result.SequenceValue, nil
}

// getSeqWithSession 获取序列值,给定序列的step，默认step是1
func getSeqWithSession(sc mongo.SessionContext, db, seqName string, step int) (num int64, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)

	query := bson.M{
		"_id": seqName,
	}

	change := bson.M{
		"$inc": bson.M{
			"sequence_value": step,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(1)
	var result Sequence
	if err := c.FindOneAndUpdate(sc, query, change, opts).Decode(&result); err != nil {
		// utils.ErrorLog("getSeq", err.Error())
		return result.SequenceValue, err
	}

	return result.SequenceValue, nil
}

// setSeq 重新赋值
func setSeq(db, seqName string, value int64) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"_id": seqName,
	}

	change := bson.M{
		"$set": bson.M{
			"sequence_value": value,
		},
	}

	if _, err := c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("getSeq", err.Error())
		return err
	}

	return nil
}

// setSeqWithSession 重新赋值
func setSeqWithSession(sc mongo.SessionContext, db, seqName string, value int64) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)

	query := bson.M{
		"_id": seqName,
	}

	change := bson.M{
		"$set": bson.M{
			"sequence_value": value,
		},
	}

	if _, err := c.UpdateOne(sc, query, change); err != nil {
		// utils.ErrorLog("getSeq", err.Error())
		return err
	}

	return nil
}

// delSeq 删除对应seq
func delSeq(db, seqName string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"_id": seqName,
	}

	if _, err := c.DeleteOne(ctx, query); err != nil {
		utils.ErrorLog("delSeq", err.Error())
		return err
	}

	return nil
}

// SetSequenceValue 直接设置值
func SetSequenceValue(db, sequenceName string, value int64) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(SequencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"_id": sequenceName,
	}

	change := bson.M{
		"$set": bson.M{
			"sequence_value": value,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("SetSequenceValue", fmt.Sprintf("query: [ %s ]", queryJSON))

	if _, err = c.UpdateOne(ctx, query, change); err != nil {
		utils.ErrorLog("SetSequenceValue", err.Error())
		return err
	}

	return nil
}
