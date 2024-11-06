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
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/pit3/srv/journal/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// JournalCollection journal collection
	JournalCollection = "journals"
)

type (
	// Journal 分录
	Journal struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		JournalID   string             `json:"journal_id" bson:"journal_id"`
		JournalName string             `json:"journal_name" bson:"journal_name"`
		AppID       string             `json:"app_id" bson:"app_id"`
		Patterns    []*Pattern         `json:"patterns" bson:"patterns"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	}
	// Journal Pattern
	Pattern struct {
		PatternID   string      `json:"pattern_id" bson:"pattern_id"`
		PatternName string      `json:"pattern_name" bson:"pattern_name"`
		Subjects    []*JSubject `json:"subjects" bson:"subjects"`
	}
	// Journal JSubject
	JSubject struct {
		SubjectKey      string `json:"subject_key" bson:"subject_key"`
		LendingDivision string `json:"lending_division" bson:"lending_division"`
		ChangeFlag      string `json:"change_flag" bson:"change_flag"`
		DefaultName     string `json:"default_name" bson:"default_name"`
		SubjectName     string `json:"subject_name" bson:"subject_name"`
		AmountName      string `json:"amount_name" bson:"amount_name"`
		AmountField     string `json:"amount_field" bson:"amount_field"`
	}
)

// ToProto 转换为proto数据
func (w *Journal) ToProto() *journal.Journal {

	var patterns []*journal.Pattern

	for _, pt := range w.Patterns {
		patterns = append(patterns, pt.ToProto())
	}

	return &journal.Journal{
		JournalId:   w.JournalID,
		JournalName: w.JournalName,
		Patterns:    patterns,
		AppId:       w.AppID,
		CreatedAt:   w.CreatedAt.String(),
		CreatedBy:   w.CreatedBy,
		UpdatedAt:   w.UpdatedAt.String(),
		UpdatedBy:   w.UpdatedBy,
	}
}

// ToProto 转换为proto数据
func (w *Pattern) ToProto() *journal.Pattern {

	var subjects []*journal.Subject

	for _, sb := range w.Subjects {
		subjects = append(subjects, sb.ToProto())
	}

	return &journal.Pattern{
		PatternId:   w.PatternID,
		PatternName: w.PatternName,
		Subjects:    subjects,
	}
}

// ToProto 转换为proto数据
func (w *JSubject) ToProto() *journal.Subject {
	return &journal.Subject{
		SubjectKey:      w.SubjectKey,
		LendingDivision: w.LendingDivision,
		ChangeFlag:      w.ChangeFlag,
		DefaultName:     w.DefaultName,
		SubjectName:     w.SubjectName,
		AmountName:      w.AmountName,
		AmountField:     w.AmountField,
	}
}

// FindJournals 获取APP下的当前分类的分录
func FindJournals(db, appId string) (items []Journal, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(JournalCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appId,
	}

	var result []Journal

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	journals, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindJournals", err.Error())
		return nil, err
	}
	defer journals.Close(ctx)
	for journals.Next(ctx) {
		var exp Journal
		err := journals.Decode(&exp)
		if err != nil {
			utils.ErrorLog("error FindJournals", err.Error())
			return nil, err
		}
		result = append(result, exp)
	}

	return result, nil
}

// FindJournal 获取分录
func FindJournal(db, appID, journalID string) (items Journal, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(JournalCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":     appID,
		"journal_id": journalID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindJournal", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Journal

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindJournal", err.Error())
		return result, err
	}

	return result, nil
}

// ImportJournal 导入分录数据
func ImportJournal(db string, journals []*Journal) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(JournalCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var insertModels []mongo.WriteModel
	for _, item := range journals {
		item.ID = primitive.NewObjectID()
		insertCxModel := mongo.NewInsertOneModel()
		insertCxModel.SetDocument(item)
		insertModels = append(insertModels, insertCxModel)
	}

	if len(insertModels) > 0 {
		result, err := c.BulkWrite(ctx, insertModels)
		if err != nil {
			bke, ok := err.(mongo.BulkWriteException)
			if !ok {
				utils.ErrorLog("error ImportJournal", err.Error())
				return err
			}
			errInfo := bke.WriteErrors[0]
			utils.ErrorLog("error ImportJournal", errInfo.Error())
			return errInfo
		}
		log.Infof("ImportJournal add result %v", result)
	}

	return nil
}

type JournalParam struct {
	JournalID       string
	AppID           string
	PatternID       string
	SubjectKey      string
	LendingDivision string
	ChangeFlag      string
	SubjectName     string
	AmountName      string
	AmountField     string
}

// ModifyJournal 更新流程实例数据
func ModifyJournal(db, writer string, param JournalParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(JournalCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id":     param.AppID,
		"journal_id": param.JournalID,
	}

	change := bson.M{

		"updated_at": time.Now(),
		"updated_by": writer,
	}

	if len(param.SubjectName) > 0 {
		change["patterns.$[outer].subjects.$[inner].subject_name"] = param.SubjectName
	} else {
		change["patterns.$[outer].subjects.$[inner].subject_name"] = ""
	}
	if len(param.AmountName) > 0 {
		change["patterns.$[outer].subjects.$[inner].amount_name"] = param.AmountName
	}
	if len(param.AmountField) > 0 {
		change["patterns.$[outer].subjects.$[inner].amount_field"] = param.AmountField
	}
	if len(param.LendingDivision) > 0 {
		change["patterns.$[outer].subjects.$[inner].lending_division"] = param.LendingDivision
	}
	if len(param.ChangeFlag) > 0 {
		change["patterns.$[outer].subjects.$[inner].change_flag"] = param.ChangeFlag
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyJournal", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyJournal", fmt.Sprintf("update: [ %s ]", updateJSON))

	opt := options.Update()
	opt.SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"outer.subjects": bson.M{
					"$ne": nil,
				},
				"outer.pattern_id": param.PatternID,
			},
			bson.M{
				"inner.subject_key": param.SubjectKey,
			},
		},
	})

	_, err = c.UpdateOne(ctx, query, update, opt)
	if err != nil {
		utils.ErrorLog("error ModifyJournal", err.Error())
		return err
	}

	return nil
}
