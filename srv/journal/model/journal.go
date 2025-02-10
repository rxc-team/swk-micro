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

	//Field规则
	FieldConf struct {
		AppId         string       `json:"app_id" bson:"app_id"`
		LayoutName    string       `json:"layout_name" bson:"layout_name"`
		CharEncoding  string       `json:"char_encoding" bson:"char_encoding"`
		HeaderRow     string       `json:"header_row" bson:"header_row"`
		SeparatorChar string       `json:"separator_char" bson:"separator_char"`
		LineBreaks    string       `json:"line_breaks" bson:"line_breaks"`
		FixedLength   bool         `json:"fixed_length" bson:"fixed_length"`
		NumberItems   int64        `json:"number_items" bson:"number_items"`
		ValidFlag     string       `json:"valid_flag" bson:"valid_flag"`
		FieldRule     []*FieldRule `json:"field_rule" bson:"field_rule"`
	}

	// FieldRule规则
	FieldRule struct {
		DownloadName      string            `json:"download_name" bson:"download_name"`
		FieldId           string            `json:"field_id" bson:"field_id"`
		FieldConditions   []*FieldCondition `json:"field_conditions" bson:"field_conditions"`
		SettingMethod     string            `json:"setting_method" bson:"setting_method"`
		FieldType         string            `json:"field_type" bson:"field_type"`
		DatastoreId       string            `json:"datastore_id" bson:"datastore_id"`
		Format            string            `json:"format" bson:"format"`
		EditContent       string            `json:"edit_content" bson:"edit_content"`
		ElseValue         string            `json:"else_value" bson:"else_value"`
		ElseType          string            `json:"else_type" bson:"else_type"`
		ElseValueDataType string            `json:"else_value_data_type" bson:"else_value_data_type"`
	}

	FieldCondition struct {
		ConditionID       string         `json:"condition_id" bson:"condition_id"`
		ConditionName     string         `json:"condition_name" bson:"condition_name"`
		FieldGroups       []*FieldGroup  `json:"field_groups" bson:"field_groups"`
		ThenValue         string         `json:"then_value" bson:"then_value"`
		ElseValue         string         `json:"else_value" bson:"else_value"`
		ThenType          string         `json:"then_type" bson:"then_type"`
		ElseType          string         `json:"else_type" bson:"else_type"`
		ThenCustomType    string         `json:"then_custom_type" bson:"then_custom_type"`
		ElseCustomType    string         `json:"else_custom_type" bson:"else_custom_type"`
		ThenCustomFields  []*CustomField `json:"then_custom_fields" bson:"then_custom_fields"`
		ElseCustomFields  []*CustomField `json:"elsen_custom_fields" bson:"else_custom_fields"`
		ThenValueDataType string         `json:"then_value_data_type" bson:"then_value_data_type"`
		ElseValueDataType string         `json:"else_value_data_type" bson:"else_value_data_type"`
	}

	// Journal Group
	FieldGroup struct {
		GroupID    string      `json:"group_id" bson:"group_id"`
		GroupName  string      `json:"group_name" bson:"group_name"`
		Type       string      `json:"type" bson:"type"`
		SwitchType string      `json:"switch_type" bson:"switch_type"`
		FieldCons  []*FieldCon `json:"field_cons" bson:"field_cons"`
	}
	// Journal Con
	FieldCon struct {
		ConID       string `json:"con_id" bson:"con_id"`
		ConName     string `json:"con_name" bson:"con_name"`
		ConField    string `json:"con_field" bson:"con_field"`
		ConOperator string `json:"con_operator" bson:"con_operator"`
		ConValue    string `json:"con_value" bson:"con_value"`
		ConDataType string `json:"con_data_type" bson:"con_data_type"`
	}

	// CustomField
	CustomField struct {
		CustomFieldType     string `json:"custom_field_type" bson:"custom_field_type"`
		CustomFieldValue    string `json:"custom_field_value" bson:"custom_field_value"`
		CustomFieldDataType string `json:"custom_field_data_type" bson:"custom_field_data_type"`
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

// ToProto 转换为proto数据
func (w *FieldCondition) ToProto() *journal.FieldCondition {
	var fieldGroups []*journal.FieldGroup

	for _, group := range w.FieldGroups {
		fieldGroups = append(fieldGroups, group.ToProto())
	}

	var thenCustomFields []*journal.CustomField

	for _, fields := range w.ThenCustomFields {
		thenCustomFields = append(thenCustomFields, fields.ToProto())
	}

	var elseCustomFields []*journal.CustomField

	for _, fields := range w.ElseCustomFields {
		elseCustomFields = append(elseCustomFields, fields.ToProto())
	}

	return &journal.FieldCondition{
		ConditionId:       w.ConditionID,
		ConditionName:     w.ConditionName,
		FieldGroups:       fieldGroups,
		ThenValue:         w.ThenValue,
		ElseValue:         w.ElseValue,
		ThenType:          w.ThenType,
		ElseType:          w.ElseType,
		ThenCustomType:    w.ThenCustomType,
		ElseCustomType:    w.ElseCustomType,
		ThenCustomFields:  thenCustomFields,
		ElseCustomFields:  elseCustomFields,
		ThenValueDataType: w.ThenValueDataType,
		ElseValueDataType: w.ElseValueDataType,
	}
}

// ToProto 转换为proto数据
func (w *FieldGroup) ToProto() *journal.FieldGroup {

	var cons []*journal.FieldCon

	for _, con := range w.FieldCons {
		cons = append(cons, con.ToProto())
	}

	return &journal.FieldGroup{
		GroupId:    w.GroupID,
		GroupName:  w.GroupName,
		Type:       w.Type,
		SwitchType: w.SwitchType,
		FieldCons:  cons,
	}
}

// ToProto 转换为proto数据
func (w *FieldCon) ToProto() *journal.FieldCon {
	return &journal.FieldCon{
		ConId:       w.ConID,
		ConName:     w.ConName,
		ConField:    w.ConField,
		ConOperator: w.ConOperator,
		ConValue:    w.ConValue,
		ConDataType: w.ConDataType,
	}
}

// ToProto 转换为proto数据
func (w *CustomField) ToProto() *journal.CustomField {
	return &journal.CustomField{
		CustomFieldType:     w.CustomFieldType,
		CustomFieldValue:    w.CustomFieldValue,
		CustomFieldDataType: w.CustomFieldDataType,
	}
}

// ToProto 转换为 proto 数据
func (m *FieldConf) ToProto() *journal.FindDownloadSettingResponse {
	// 创建 FieldRule 的 proto 列表
	var fieldRules []*journal.FieldRule
	for _, r := range m.FieldRule {
		fieldRules = append(fieldRules, r.ToProto())
	}

	return &journal.FindDownloadSettingResponse{
		AppId:         m.AppId,
		LayoutName:    m.LayoutName,
		CharEncoding:  m.CharEncoding,
		HeaderRow:     m.HeaderRow,
		SeparatorChar: m.SeparatorChar,
		LineBreaks:    m.LineBreaks,
		FixedLength:   m.FixedLength,
		NumberItems:   m.NumberItems,
		ValidFlag:     m.ValidFlag,
		FieldRule:     fieldRules,
	}
}

// ToProto 转换为 proto 数据
func (f *FieldRule) ToProto() *journal.FieldRule {
	var fieldConditions []*journal.FieldCondition
	for _, r := range f.FieldConditions {
		fieldConditions = append(fieldConditions, r.ToProto())
	}

	return &journal.FieldRule{
		DownloadName:      f.DownloadName,
		FieldId:           f.FieldId,
		FieldConditions:   fieldConditions,
		SettingMethod:     f.SettingMethod,
		FieldType:         f.FieldType,
		DatastoreId:       f.DatastoreId,
		Format:            f.Format,
		EditContent:       f.EditContent,
		ElseValue:         f.ElseValue,
		ElseType:          f.ElseType,
		ElseValueDataType: f.ElseValueDataType,
	}
}

// ConvertToProto 将 []FieldConf 切片转换为 Protobuf 格式的 FindDownloadSettingsResponse
func ConvertToProto(fieldConfs []FieldConf) *journal.FindDownloadSettingsResponse {
	protoResponse := &journal.FindDownloadSettingsResponse{}

	// 遍历 FieldConf 切片，并将每个 FieldConf 转换为 Protobuf 格式
	for _, fc := range fieldConfs {
		protoResponse.FieldConf = append(protoResponse.FieldConf, fc.ToProto())
	}

	return protoResponse
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

// 添加分录下载设定
func AddDownloadSetting(db string, appID string, fd FieldConf) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("journals_download")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appID,
	}

	_, err = c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("AddDownloadSetting", err.Error())
		return err
	}

	if _, err = c.InsertOne(ctx, fd); err != nil {
		utils.ErrorLog("AddDownloadSetting", err.Error())
		return err
	}

	return nil
}

// 查询分录下载设定
func FindDownloadSetting(db string, appID string) (fd FieldConf, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("journals_download")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appID,
	}
	var result FieldConf

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			// 不返回错误
			return result, nil
		}
		utils.ErrorLog("FindDownloadSetting", err.Error())
	}

	return result, nil
}

// 查询所有分录下载设定
func FindDownloadSettings(db string, appID string) (fd []FieldConf, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("journals_download")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appID,
	}

	// 定义一个切片用于存储查询结果
	var results []FieldConf

	// 使用 Find 查询所有符合条件的文档
	cursor, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("FindDownloadSetting", err.Error())
		return nil, err
	}

	// 确保在函数返回之前关闭游标
	defer cursor.Close(ctx)

	// 将游标中的所有数据解码到 results 切片中
	if err := cursor.All(ctx, &results); err != nil {
		utils.ErrorLog("FindDownloadSetting", err.Error())
		return nil, err
	}

	// 如果没有找到任何数据，返回空切片
	if len(results) == 0 {
		return nil, nil
	}

	return results, nil
}
