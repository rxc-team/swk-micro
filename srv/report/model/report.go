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

	"rxcsoft.cn/pit3/srv/report/proto/report"
	"rxcsoft.cn/pit3/srv/report/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// ReportsCollection reports collection
	ReportsCollection = "reports"

	// LanguagesCollection languages collection
	LanguagesCollection = "languages"

	// RolesCollection roles collection
	RolesCollection = "roles"
)

const MaxIndexCount = 52 // default max 64 x %82

type IndexUsage struct {
	Name  string `bson:"name"`
	Usage int64  `bson:"usage"`
}

type ItemMap map[string]*Value

type (
	// Report 报表信息
	Report struct {
		ID               primitive.ObjectID `json:"id" bson:"_id"`
		Domain           string             `json:"domain" bson:"domain"`
		AppID            string             `json:"app_id" bson:"app_id"`
		DatastoreID      string             `json:"datastore_id" bson:"datastore_id"`
		ReportID         string             `json:"report_id" bson:"report_id"`
		ReportName       string             `json:"report_name" bson:"report_name"`
		DisplayOrder     int64              `json:"display_order" bson:"display_order"`
		IsUseGroup       bool               `json:"is_use_group" bson:"is_use_group"`
		ReportConditions []*ReportCondition `json:"report_conditions,omitempty" bson:"report_conditions"`
		ConditionType    string             `json:"condition_type" bson:"condition_type"`
		GroupInfo        *GroupInfo         `json:"group_info" bson:"group_info"`
		SelectKeyInfos   []*KeyInfo         `json:"select_key_infos,omitempty" bson:"select_key_infos"`
		CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy        string             `json:"created_by" bson:"created_by"`
		UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy        string             `json:"updated_by" bson:"updated_by"`
		DeletedAt        time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy        string             `json:"deleted_by" bson:"deleted_by"`
	}

	// ReportCondition 报表检索条件
	ReportCondition struct {
		FieldID       string `json:"field_id" bson:"field_id"`
		FieldType     string `json:"field_type" bson:"field_type"`
		SearchValue   string `json:"search_value" bson:"search_value"`
		Operator      string `json:"operator" bson:"operator"`
		IsDynamic     bool   `json:"is_dynamic" bson:"is_dynamic"`
		ConditionType string `json:"condition_type" bson:"condition_type"`
	}

	// ReportParam 报表数据检索参数
	ReportParam struct {
		ReportID      string
		ConditionList []*ReportCondition
		ConditionType string
		PageIndex     int64
		PageSize      int64
		Owners        []string
	}

	// GroupInfo Group情报
	GroupInfo struct {
		GroupKeys []*KeyInfo  `json:"group_keys,omitempty" bson:"group_keys"`
		AggreKeys []*AggreKey `json:"aggre_keys,omitempty" bson:"aggre_keys"`
		ShowCount bool        `json:"show_count" bson:"show_count"`
	}

	// KeyInfo 字段情报
	KeyInfo struct {
		IsLookup    bool   `json:"is_lookup" bson:"is_lookup"`
		FieldID     string `json:"field_id" bson:"field_id"`
		DatastoreID string `json:"datastore_id" bson:"datastore_id"`
		DataType    string `json:"data_type" bson:"data_type"`
		OptionID    string `json:"option_id" bson:"option_id"`
		AliasName   string `json:"alias_name" bson:"alias_name"`
		Sort        string `json:"sort" bson:"sort"`
		IsDynamic   bool   `json:"is_dynamic" bson:"is_dynamic"`
		Unique      bool   `json:"unique" bson:"unique"`
		Order       int64  `json:"order" bson:"order"`
	}

	// FieldInfo 字段情报
	FieldInfo struct {
		DataType    string `json:"data_type" bson:"data_type"`
		AliasName   string `json:"alias_name" bson:"alias_name"`
		DatastoreID string `json:"datastore_id" bson:"datastore_id"`
		IsDynamic   bool   `json:"is_dynamic" bson:"is_dynamic"`
		OptionID    string `json:"option_id" bson:"option_id"`
		Unique      bool   `json:"unique" bson:"unique"`
		Order       int64  `json:"order" bson:"order"`
	}

	// AggreKey 聚合字段情报
	AggreKey struct {
		IsLookup    bool   `json:"is_lookup" bson:"is_lookup"`
		FieldID     string `json:"field_id" bson:"field_id"`
		AggreType   string `json:"aggre_type" bson:"aggre_type"`
		DataType    string `json:"data_type" bson:"data_type"`
		DatastoreID string `json:"datastore_id" bson:"datastore_id"`
		OptionID    string `json:"option_id" bson:"option_id"`
		AliasName   string `json:"alias_name" bson:"alias_name"`
		Sort        string `json:"sort" bson:"sort"`
		Order       int64  `json:"order" bson:"order"`
	}

	// ReportData 报表数据
	ReportData struct {
		Items       ItemMap   `json:"items" bson:"items"`
		Count       int64     `json:"count" bson:"count"`
		ItemID      string    `json:"item_id" bson:"item_id"`
		CheckType   string    `json:"check_type" bson:"check_type"`
		CheckStatus string    `json:"check_status" bson:"check_status"`
		CreatedAt   time.Time `json:"created_at" bson:"created_at"`
		CreatedBy   string    `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string    `json:"updated_by" bson:"updated_by"`
		CheckedAt   time.Time `json:"checked_at" bson:"checked_at"`
		CheckedBy   string    `json:"checked_by" bson:"checked_by"`
		UpdateTime  time.Time `json:"update_time" bson:"update_time"`
		LabelTime   time.Time `json:"label_time" bson:"label_time"`
	}

	// Value 字段的值
	Value struct {
		DataType string      `json:"data_type,omitempty" bson:"data_type"`
		Value    interface{} `json:"value,omitempty" bson:"value"`
	}

	// Config 顾客配置情报
	Config struct {
		Special         string `json:"special" bson:"special"`
		CheckStartDate  string `json:"check_start_date" bson:"check_start_date"`
		SyoriYm         string `json:"syori_ym" bson:"syori_ym"`
		ShortLeases     string `json:"short_leases" bson:"short_leases"`
		KishuYm         string `json:"kishu_ym" bson:"kishu_ym"`
		MinorBaseAmount string `json:"minor_base_amount" bson:"minor_base_amount"`
	}

	TotalResult struct {
		Total int64 `bson:"total"`
	}
)

// ToProto 转换为proto数据(报表数据)
func (r *ReportData) ToProto() *report.ReportData {
	items := make(map[string]*report.Value, len(r.Items))
	for key, value := range r.Items {
		items[key] = &report.Value{
			DataType: value.DataType,
			Value:    GetValue(value),
		}
	}
	return &report.ReportData{
		Items:       items,
		ItemId:      r.ItemID,
		Count:       r.Count,
		CheckType:   r.CheckType,
		CheckStatus: r.CheckStatus,
		CreatedAt:   r.CreatedAt.String(),
		CreatedBy:   r.CreatedBy,
		CheckedAt:   r.CheckedAt.String(),
		CheckedBy:   r.CheckedBy,
		UpdatedAt:   r.UpdatedAt.String(),
		UpdateTime:  r.UpdateTime.String(),
		LabelTime:   r.LabelTime.String(),
		UpdatedBy:   r.UpdatedBy,
	}
}

// ToProto 转换为proto数据(报表检索条件)
func (r *ReportCondition) ToProto() *report.ReportCondition {
	return &report.ReportCondition{
		FieldId:       r.FieldID,
		FieldType:     r.FieldType,
		SearchValue:   r.SearchValue,
		Operator:      r.Operator,
		IsDynamic:     r.IsDynamic,
		ConditionType: r.ConditionType,
	}
}

// ToProto 转换为proto数据(字段情报)
func (k *KeyInfo) ToProto() *report.KeyInfo {
	return &report.KeyInfo{
		IsLookup:    k.IsLookup,
		FieldId:     k.FieldID,
		DatastoreId: k.DatastoreID,
		DataType:    k.DataType,
		AliasName:   k.AliasName,
		OptionId:    k.OptionID,
		Sort:        k.Sort,
		IsDynamic:   k.IsDynamic,
		Unique:      k.Unique,
		Order:       k.Order,
	}
}

// ToProto 转换为proto数据(字段情报)
func (f *FieldInfo) ToProto() *report.FieldInfo {
	return &report.FieldInfo{
		DataType:    f.DataType,
		AliasName:   f.AliasName,
		DatastoreId: f.DatastoreID,
		IsDynamic:   f.IsDynamic,
		OptionId:    f.OptionID,
		Unique:      f.Unique,
		Order:       f.Order,
	}
}

// ToProto 转换为proto数据(聚合字段情报)
func (a *AggreKey) ToProto() *report.AggreKey {
	return &report.AggreKey{
		IsLookup:    a.IsLookup,
		FieldId:     a.FieldID,
		AggreType:   a.AggreType,
		DatastoreId: a.DatastoreID,
		DataType:    a.DataType,
		OptionId:    a.OptionID,
		AliasName:   a.AliasName,
		Sort:        a.Sort,
		Order:       a.Order,
	}
}

// ToProto 转换为proto数据(Group情报)
func (g *GroupInfo) ToProto() *report.GroupInfo {

	groupkeys := make([]*report.KeyInfo, len(g.GroupKeys))
	for index, ch := range g.GroupKeys {
		groupkeys[index] = ch.ToProto()
	}

	aggrekeys := make([]*report.AggreKey, len(g.AggreKeys))
	for index, ch := range g.AggreKeys {
		aggrekeys[index] = ch.ToProto()
	}

	return &report.GroupInfo{
		GroupKeys: groupkeys,
		AggreKeys: aggrekeys,
		ShowCount: g.ShowCount,
	}
}

// ToProto 转换为proto数据(报表)
func (r *Report) ToProto() *report.Report {

	reportconditions := make([]*report.ReportCondition, len(r.ReportConditions))
	for index, ch := range r.ReportConditions {
		reportconditions[index] = ch.ToProto()
	}

	if r.IsUseGroup {
		return &report.Report{
			Domain:           r.Domain,
			AppId:            r.AppID,
			DatastoreId:      r.DatastoreID,
			ReportId:         r.ReportID,
			ReportName:       r.ReportName,
			DisplayOrder:     r.DisplayOrder,
			IsUseGroup:       r.IsUseGroup,
			ReportConditions: reportconditions,
			ConditionType:    r.ConditionType,
			GroupInfo:        r.GroupInfo.ToProto(),
			CreatedAt:        r.CreatedAt.String(),
			CreatedBy:        r.CreatedBy,
			UpdatedAt:        r.UpdatedAt.String(),
			UpdatedBy:        r.UpdatedBy,
			DeletedAt:        r.DeletedAt.String(),
			DeletedBy:        r.DeletedBy,
		}
	}

	selectkeyinfos := make([]*report.KeyInfo, len(r.SelectKeyInfos))
	for index, ch := range r.SelectKeyInfos {
		selectkeyinfos[index] = ch.ToProto()
	}

	return &report.Report{
		Domain:           r.Domain,
		AppId:            r.AppID,
		DatastoreId:      r.DatastoreID,
		ReportId:         r.ReportID,
		ReportName:       r.ReportName,
		DisplayOrder:     r.DisplayOrder,
		IsUseGroup:       r.IsUseGroup,
		ReportConditions: reportconditions,
		ConditionType:    r.ConditionType,
		SelectKeyInfos:   selectkeyinfos,
		CreatedAt:        r.CreatedAt.String(),
		CreatedBy:        r.CreatedBy,
		UpdatedAt:        r.UpdatedAt.String(),
		UpdatedBy:        r.UpdatedBy,
		DeletedAt:        r.DeletedAt.String(),
		DeletedBy:        r.DeletedBy,
	}
}

// ToProto 转换为proto数据
func (v *Value) ToProto() *report.Value {
	return &report.Value{
		DataType: v.DataType,
		Value:    GetValueString(v),
	}
}

// FindReports 获取所属公司所属APP下所有报表情报
func FindReports(db, datastoreID, domain, appid string) (r []Report, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
		"app_id":     appid,
	}
	// 台账ID不为空,做条件
	if datastoreID != "" {
		query["datastore_id"] = datastoreID
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindReports", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "display_order", Value: 1},
	}

	opt := options.Find()
	opt.SetSort(sortItem)

	var result []Report
	reports, err := c.Find(ctx, query, opt)
	if err != nil {
		utils.ErrorLog("error FindReports", err.Error())
		return nil, err
	}
	defer reports.Close(ctx)
	for reports.Next(ctx) {
		var rep Report
		err := reports.Decode(&rep)
		if err != nil {
			utils.ErrorLog("error FindReports", err.Error())
			return nil, err
		}
		result = append(result, rep)
	}

	return result, nil
}

// FindDatastoreReports 获取台账下所有报表情报
func FindDatastoreReports(db, domain, appid, datastoreID string) (r []Report, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"datastore_id": datastoreID,
		"domain":       domain,
		"app_id":       appid,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDatastoreReports", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Report
	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)
	reports, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindDatastoreReports", err.Error())
		return nil, err
	}
	defer reports.Close(ctx)
	for reports.Next(ctx) {
		var rep Report
		err := reports.Decode(&rep)
		if err != nil {
			utils.ErrorLog("error FindDatastoreReports", err.Error())
			return nil, err
		}
		result = append(result, rep)
	}

	return result, nil
}

// FindReport 通过报表ID获取单个报表情报
func FindReport(db, id string) (r Report, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Report
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorLog("error FindReport", err.Error())
		return result, err
	}
	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"_id":        objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindReport", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindReport", err.Error())
		return result, err
	}

	return result, nil
}

// AddReport 添加单个报表情报
func AddReport(db string, r *Report) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r.ID = primitive.NewObjectID()
	r.ReportID = r.ID.Hex()
	r.ReportName = GetReportNameKey(r.AppID, r.ReportID)

	if r.ReportConditions == nil {
		r.ReportConditions = make([]*ReportCondition, 0)
	}
	if r.SelectKeyInfos == nil {
		r.SelectKeyInfos = make([]*KeyInfo, 0)
	}
	if r.GroupInfo == nil {
		r.GroupInfo = &GroupInfo{
			AggreKeys: make([]*AggreKey, 0),
			GroupKeys: make([]*KeyInfo, 0),
			ShowCount: false,
		}

	} else {
		if r.GroupInfo.AggreKeys == nil {
			r.GroupInfo.AggreKeys = make([]*AggreKey, 0)
		}
		if r.GroupInfo.GroupKeys == nil {
			r.GroupInfo.GroupKeys = make([]*KeyInfo, 0)
		}
	}

	queryJSON, _ := json.Marshal(r)
	utils.DebugLog("AddReport", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, r)
	if err != nil {
		utils.ErrorLog("error AddReport", err.Error())
		return "", err
	}

	go func() {
		err := GenerateReportData(db, r.ReportID)
		if err != nil {
			utils.ErrorLog("error AddReport", err.Error())
		}
	}()

	return r.ReportID, nil
}

// ModifyReq 更新报表数据结构体
type ModifyReq struct {
	DatastoreID      string
	ReportID         string
	ReportName       string
	DisplayOrder     string
	IsUseGroup       string
	ReportConditions []*ReportCondition
	ConditionType    string
	GroupInfo        *GroupInfo
	SelectKeyInfos   []*KeyInfo
	Writer           string
}

// ModifyReport 更新单个报表情报
func ModifyReport(db string, r *ModifyReq) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(r.ReportID)
	if err != nil {
		utils.ErrorLog("error ModifyReport", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": r.Writer,
	}

	// 台账ID不为空的场合
	if r.DatastoreID != "" {
		change["datastore_id"] = r.DatastoreID
	}

	// 报表显示顺序不为空的场合
	if r.DisplayOrder != "" {
		result, err := strconv.ParseInt(r.DisplayOrder, 10, 64)
		if err == nil {
			change["display_order"] = result
		}
	}

	// 报表是否使用分组不为空的场合
	if r.IsUseGroup != "" {
		result, err := strconv.ParseBool(r.IsUseGroup)
		if err == nil {
			change["is_use_group"] = result
		}
	}

	// 报表检索条件数组不为空的场合
	if len(r.ReportConditions) > 0 {
		change["report_conditions"] = r.ReportConditions
	} else {
		change["report_conditions"] = []bson.M{}
	}
	// 报表检索条件数组不为空的场合
	if len(r.ConditionType) > 0 {
		change["condition_type"] = r.ConditionType
	}

	if r.IsUseGroup == "true" {
		// Group情报不为空的场合
		if r.GroupInfo != nil {
			change["group_info"] = r.GroupInfo
		} else {
			change["group_info"] = &GroupInfo{
				AggreKeys: make([]*AggreKey, 0),
				GroupKeys: make([]*KeyInfo, 0),
				ShowCount: false,
			}
		}

		change["select_key_infos"] = make([]*KeyInfo, 0)
	} else {
		// 出力字段情报数组不为空的场合
		if len(r.SelectKeyInfos) > 0 {
			change["select_key_infos"] = r.SelectKeyInfos
		} else {
			change["select_key_infos"] = make([]*KeyInfo, 0)
		}

		change["group_info"] = &GroupInfo{
			AggreKeys: make([]*AggreKey, 0),
			GroupKeys: make([]*KeyInfo, 0),
			ShowCount: false,
		}
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyReport", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyReport", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyReport", err.Error())
		return err
	}

	go func() {
		err := GenerateReportData(db, r.ReportID)
		if err != nil {
			utils.ErrorLog("error ModifyReport", err.Error())
		}
	}()

	return nil
}

// DeleteReport 软删除单个报表情报
func DeleteReport(db, reportid, userid string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(reportid)
	if err != nil {
		utils.ErrorLog("error DeleteReport", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userid,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteReport", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteReport", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteReport ", err.Error())
		return err
	}

	return nil
}

// DeleteSelectReports 软删除多个报表情报
func DeleteSelectReports(db string, reportidlist []string, userid string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectReports", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectReports", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, reportid := range reportidlist {
			objectID, err := primitive.ObjectIDFromHex(reportid)
			if err != nil {
				utils.ErrorLog("error DeleteSelectReports", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userid,
			}}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSelectReports", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectReports", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error DeleteSelectReports", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectReports", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectReports", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// HardDeleteReports 物理删除多个报表情报
func HardDeleteReports(db string, reportidlist []string, domain string, appID string, writer string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	dc := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	lc := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	rc := client.Database(database.GetDBName(db)).Collection(RolesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteReports", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteReports", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, reportid := range reportidlist {
			// 删除报表数据
			objectID, err := primitive.ObjectIDFromHex(reportid)
			if err != nil {
				utils.ErrorLog("error HardDeleteReports", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteReports", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteReports", err.Error())
				return err
			}
			// 删除报表多语言
			lrquery := bson.M{
				"domain": domain,
			}

			setRKey := strings.Builder{}
			setRKey.WriteString("apps.")
			setRKey.WriteString(appID)
			setRKey.WriteString(".")
			setRKey.WriteString("reports")
			setRKey.WriteString(".")
			setRKey.WriteString(reportid)

			lrupdate := bson.M{
				"$unset": bson.M{
					setRKey.String(): 1,
				},
				"$set": bson.M{
					"updated_at": time.Now(),
					"updated_by": writer,
				},
			}

			lrqueryJSON, _ := json.Marshal(lrquery)
			utils.DebugLog("HardDeleteReports", fmt.Sprintf("query: [ %s ]", lrqueryJSON))

			lrupdateSON, _ := json.Marshal(lrupdate)
			utils.DebugLog("HardDeleteReports", fmt.Sprintf("update: [ %s ]", lrupdateSON))

			_, err = lc.UpdateMany(ctx, lrquery, lrupdate)
			if err != nil {
				utils.ErrorLog("error HardDeleteReports", err.Error())
				return err
			}

			// 根据reportID查询所有dashboard
			dquery := bson.M{
				"report_id": reportid,
			}

			dqueryJSON, _ := json.Marshal(dquery)
			utils.DebugLog("HardDeleteReports", fmt.Sprintf("query: [ %s ]", dqueryJSON))

			dashboards, err := dc.Find(ctx, dquery)
			if err != nil {
				utils.ErrorLog("error HardDeleteReports", err.Error())
				return err
			}
			defer dashboards.Close(ctx)
			for dashboards.Next(ctx) {
				var dd Dashboard
				err := dashboards.Decode(&dd)
				if err != nil {
					utils.ErrorLog("error HardDeleteReports", err.Error())
					return err
				}
				// 删除dashboard多语言
				ldquery := bson.M{
					"domain": domain,
				}

				setDKey := strings.Builder{}
				setDKey.WriteString("apps.")
				setDKey.WriteString(appID)
				setDKey.WriteString(".")
				setDKey.WriteString("dashboards")
				setDKey.WriteString(".")
				setDKey.WriteString(dd.DashboardID)

				ldupdate := bson.M{
					"$unset": bson.M{
						setDKey.String(): 1,
					},
					"$set": bson.M{
						"updated_at": time.Now(),
						"updated_by": writer,
					},
				}

				ldqueryJSON, _ := json.Marshal(ldquery)
				utils.DebugLog("HardDeleteReports", fmt.Sprintf("query: [ %s ]", ldqueryJSON))

				ldupdateSON, _ := json.Marshal(ldupdate)
				utils.DebugLog("HardDeleteReports", fmt.Sprintf("update: [ %s ]", ldupdateSON))

				_, err = lc.UpdateMany(ctx, ldquery, ldupdate)
				if err != nil {
					utils.ErrorLog("error HardDeleteReports", err.Error())
					return err
				}
				// 删除dashboard
				objectID, err := primitive.ObjectIDFromHex(dd.DashboardID)
				if err != nil {
					utils.ErrorLog("error HardDeleteReports", err.Error())
					return err
				}
				query := bson.M{
					"_id": objectID,
				}
				queryJSON, _ := json.Marshal(query)
				utils.DebugLog("HardDeleteReports", fmt.Sprintf("query: [ %s ]", queryJSON))

				_, err = dc.DeleteOne(sc, query)
				if err != nil {
					utils.ErrorLog("error HardDeleteReports", err.Error())
					return err
				}
			}
			// 删除角色中该报表的数据
			rquery := bson.M{}
			rupdate := bson.M{
				"$set": bson.M{
					"updated_at": time.Now(),
					"updated_by": writer,
				},
				"$pull": bson.M{
					"reports": reportid,
				},
			}
			rqueryJSON, _ := json.Marshal(rquery)
			utils.DebugLog("HardDeleteReports", fmt.Sprintf("query: [ %s ]", rqueryJSON))

			rupdateSON, _ := json.Marshal(rupdate)
			utils.DebugLog("HardDeleteReports", fmt.Sprintf("update: [ %s ]", rupdateSON))

			_, err = rc.UpdateMany(sc, rquery, rupdate)
			if err != nil {
				utils.ErrorLog("error HardDeleteReports", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteReports", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteReports", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// ReportDataInfo 报表数据信息
type ReportDataInfo struct {
	ReportInfo Report
	ReportData []*ReportData
	Fields     map[string]*FieldInfo
	Total      int64
}

// FindReportData 通过报表ID获取报表数据情报
func FindReportData(db string, params ReportParam) (reportDataInfo *ReportDataInfo, err error) {

	// 获取报表设置信息
	reportInfo, err := FindReport(db, params.ReportID)
	if err != nil {
		return nil, err
	}

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("report_" + params.ReportID)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var fieldInfos map[string]*FieldInfo

	fields, err := getAllFields(db, reportInfo.AppID, reportInfo.DatastoreID)
	if err != nil {
		utils.ErrorLog("error FindReportData", err.Error())
		return nil, err
	}

	match, indexKeys, indexMap := buildReportMatch(params.ConditionType, params.ConditionList, params.Owners)

	pipe := []bson.M{
		{
			"$match": match,
		},
	}

	var result []*ReportData
	skip := (params.PageIndex - 1) * params.PageSize
	limit := params.PageSize

	// 使用集计的场合
	if reportInfo.IsUseGroup {
		querys, fs := buildGroupFields(reportInfo.AppID, reportInfo.GroupInfo, fields, false)
		// 排序
		sortQuery := querys[len(querys)-1]
		pipe = append(pipe, sortQuery)
		// 限制
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
		// 其他
		otherQuery := querys[:len(querys)-1]
		pipe = append(pipe, otherQuery...)
		fieldInfos = fs

		// 索引
		sorts := sortQuery["$sort"].(bson.D)
		for _, sort := range sorts {
			indexMap[sort.Key] = sort.Value
			indexKeys = append(indexKeys, sort.Key)
		}

		var indexs bson.D
		existMap := make(map[string]struct{})
		for _, key := range indexKeys {
			if _, exist := existMap[key]; !exist {
				existMap[key] = struct{}{}
				indexs = append(indexs, bson.E{
					Key: key, Value: indexMap[key],
				})
			}
		}
		index := mongo.IndexModel{
			Keys:    indexs,
			Options: options.Index().SetSparse(true).SetUnique(false),
		}
		// 删除多余的索引
		if err := DeleteLeastUsed(c); err != nil {
			utils.ErrorLog("FindReportData", err.Error())
			return nil, err
		}
		if len(indexs) < 31 {
			indexOpts := options.CreateIndexes().SetMaxTime(60 * time.Second)
			if _, err := c.Indexes().CreateOne(ctx, index, indexOpts); err != nil {
				utils.ErrorLog("FindReportData", err.Error())
				return nil, err
			}
		}
	} else {
		querys, fs := buildSelectFields(reportInfo.AppID, reportInfo.SelectKeyInfos, fields)
		// 排序
		sortQuery := querys[len(querys)-1]
		pipe = append(pipe, sortQuery)
		// 限制
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
		// 其他
		otherQuery := querys[:len(querys)-1]
		pipe = append(pipe, otherQuery...)
		fieldInfos = fs

		// 索引
		sorts := sortQuery["$sort"].(bson.D)
		for _, sort := range sorts {
			indexMap[sort.Key] = sort.Value
			indexKeys = append(indexKeys, sort.Key)
		}

		var indexs bson.D
		existMap := make(map[string]struct{})
		for _, key := range indexKeys {
			if _, exist := existMap[key]; !exist {
				existMap[key] = struct{}{}
				indexs = append(indexs, bson.E{
					Key: key, Value: indexMap[key],
				})
			}
		}
		index := mongo.IndexModel{
			Keys:    indexs,
			Options: options.Index().SetSparse(true).SetUnique(false),
		}
		// 删除多余的索引
		if err := DeleteLeastUsed(c); err != nil {
			utils.ErrorLog("FindReportData", err.Error())
			return nil, err
		}
		if len(indexs) < 31 {
			indexOpts := options.CreateIndexes().SetMaxTime(60 * time.Second)
			if _, err := c.Indexes().CreateOne(ctx, index, indexOpts); err != nil {
				utils.ErrorLog("FindReportData", err.Error())
				return nil, err
			}
		}
	}

	total := make(chan int64)

	go func() {
		getCount(db, reportInfo, match, total)
	}()

	pipeJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindReportData", fmt.Sprintf("pipe: [ %s ]", pipeJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("error FindReportData", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var rep ReportData
		err := cur.Decode(&rep)
		if err != nil {
			utils.ErrorLog("error FindReportData", err.Error())
			return nil, err
		}
		result = append(result, &rep)
	}

	reportData := ReportDataInfo{
		ReportInfo: reportInfo,
		ReportData: result,
		Fields:     fieldInfos,
		Total:      <-total,
	}

	return &reportData, nil
}

// DownloadReportData 通过报表ID下载报表数据
func DownloadReportData(db string, params ReportParam, stream report.ReportService_DownloadStream) (err error) {

	// 获取报表设置信息
	reportInfo, err := FindReport(db, params.ReportID)
	if err != nil {
		return err
	}

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("report_" + params.ReportID)
	ctx, cancel := context.WithTimeout(context.Background(), 30*60*time.Second)
	defer cancel()

	fields, err := getAllFields(db, reportInfo.AppID, reportInfo.DatastoreID)
	if err != nil {
		utils.ErrorLog("error DownloadReportData", err.Error())
		return err
	}

	match, indexKeys, indexMap := buildReportMatch(params.ConditionType, params.ConditionList, params.Owners)

	pipe := []bson.M{
		{
			"$match": match,
		},
	}

	// skip := (params.PageIndex - 1) * params.PageSize
	// limit := params.PageSize

	// 使用集计的场合
	if reportInfo.IsUseGroup {
		querys, _ := buildGroupFields(reportInfo.AppID, reportInfo.GroupInfo, fields, false)
		// 排序
		sortQuery := querys[len(querys)-1]
		pipe = append(pipe, sortQuery)

		// 其他
		otherQuery := querys[:len(querys)-1]
		pipe = append(pipe, otherQuery...)

		// 索引
		sorts := sortQuery["$sort"].(bson.D)
		for _, sort := range sorts {
			indexMap[sort.Key] = sort.Value
			indexKeys = append(indexKeys, sort.Key)
		}

		var indexs bson.D
		existMap := make(map[string]struct{})
		for _, key := range indexKeys {
			if _, exist := existMap[key]; !exist {
				existMap[key] = struct{}{}
				indexs = append(indexs, bson.E{
					Key: key, Value: indexMap[key],
				})
			}
		}
		index := mongo.IndexModel{
			Keys:    indexs,
			Options: options.Index().SetSparse(true).SetUnique(false),
		}
		// 删除多余的索引
		if err := DeleteLeastUsed(c); err != nil {
			utils.ErrorLog("DownloadReportData", err.Error())
			return err
		}
		if len(indexs) < 31 {
			indexOpts := options.CreateIndexes().SetMaxTime(60 * time.Second)
			if _, err := c.Indexes().CreateOne(ctx, index, indexOpts); err != nil {
				utils.ErrorLog("DownloadReportData", err.Error())
				return err
			}
		}
	} else {
		querys, _ := buildSelectFields(reportInfo.AppID, reportInfo.SelectKeyInfos, fields)
		// 排序
		sortQuery := querys[len(querys)-1]
		pipe = append(pipe, sortQuery)
		// 限制
		// if skip > 0 {
		// 	pipe = append(pipe, bson.M{
		// 		"$skip": skip,
		// 	})
		// }
		// if limit > 0 {
		// 	pipe = append(pipe, bson.M{
		// 		"$limit": limit,
		// 	})
		// }
		// 其他
		otherQuery := querys[:len(querys)-1]
		pipe = append(pipe, otherQuery...)

		// 索引
		indexs := sortQuery["$sort"].(bson.D)
		index := mongo.IndexModel{
			Keys:    indexs,
			Options: options.Index().SetSparse(true).SetUnique(false),
		}
		// 删除多余的索引
		if err := DeleteLeastUsed(c); err != nil {
			utils.ErrorLog("DownloadReportData", err.Error())
			return err
		}
		if len(indexs) < 31 {
			indexOpts := options.CreateIndexes().SetMaxTime(60 * time.Second)
			if _, err := c.Indexes().CreateOne(ctx, index, indexOpts); err != nil {
				utils.ErrorLog("DownloadReportData", err.Error())
				return err
			}
		}
	}

	pipeJSON, _ := json.Marshal(pipe)
	utils.DebugLog("DownloadReportData", fmt.Sprintf("pipe: [ %s ]", pipeJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)
	opt.SetBatchSize(int32(100))

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("error DownloadReportData", err.Error())
		return err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var rep ReportData
		err := cur.Decode(&rep)
		if err != nil {
			utils.ErrorLog("error DownloadReportData", err.Error())
			return err
		}
		if err := stream.Send(&report.DownloadResponse{ItemData: rep.ToProto()}); err != nil {
			utils.ErrorLog("DownloadReportData", err.Error())
			return err
		}
	}
	return nil
}

// GenerateReportData  生成报表的数据
func GenerateReportData(db, reportId string) (err error) {

	// 获取报表设置信息
	reportInfo, err := FindReport(db, reportId)
	if err != nil {
		utils.ErrorLog("error GenerateReportData", err.Error())
		return err
	}
	handleMonth := ""
	config, err := findConfig(db, reportInfo.AppID)
	if err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			handleMonth = ""
		}
	}

	handleMonth = config.SyoriYm

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(reportInfo.DatastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fields, err := getFields(db, reportInfo.AppID, reportInfo.DatastoreID)
	if err != nil {
		utils.ErrorLog("error GenerateReportData", err.Error())
		return err
	}

	pipe := []bson.M{}

	match := buildMatch(handleMonth, reportInfo.ConditionType, reportInfo.ReportConditions)

	pipe = append(pipe, match)

	project := bson.M{
		"_id":          1,
		"item_id":      1,
		"app_id":       1,
		"datastore_id": 1,
		"owners":       1,
		"check_type":   1,
		"check_status": 1,
		"created_at":   1,
		"created_by":   1,
		"updated_at":   1,
		"updated_by":   1,
		"checked_at":   1,
		"checked_by":   1,
		"label_time":   1,
		"status":       1,
	}

	ds, err := getDatastore(db, reportInfo.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindItems", err.Error())
		return err
	}

	// 关联台账
	for _, relation := range ds.Relations {
		let := bson.M{}
		and := make([]bson.M, 0)
		for relationKey, localKey := range relation.Fields {
			let[localKey] = "$items." + localKey + ".value"

			and = append(and, bson.M{
				"$eq": []interface{}{"$items." + relationKey + ".value", "$$" + localKey},
			})
		}

		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": and,
					},
				},
			},
		}

		lookup := bson.M{
			"from":     "item_" + relation.DatastoreId,
			"let":      let,
			"pipeline": pp,
			"as":       relation.RelationId,
		}

		unwind := bson.M{
			"path":                       "$" + relation.RelationId,
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})
		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})

		// 关联的数据
		project["relations."+relation.RelationId] = "$" + relation.RelationId + ".items"
	}

	for _, f := range fields {
		// 函数字段，重新拼接
		if f.FieldType == "function" {
			var formula bson.M
			err := json.Unmarshal([]byte(f.Formula), &formula)
			if err != nil {
				utils.ErrorLog("GenerateReportData", err.Error())
				return err
			}

			if len(formula) > 0 {
				project["items."+f.FieldID+".value"] = formula
			} else {
				project["items."+f.FieldID+".value"] = ""
			}

			// 当前数据本身
			project["items."+f.FieldID+".data_type"] = f.ReturnType
			continue
		}

		// 当前数据本身
		project["items."+f.FieldID] = "$items." + f.FieldID
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	// 使用集计的场合
	if reportInfo.IsUseGroup {

		groupQuery := buildGroup(reportInfo.GroupInfo)
		pipe = append(pipe, groupQuery...)

		pipe = append(pipe, bson.M{
			"$out": "report_" + reportInfo.ReportID,
		})

		queryJSON, _ := json.Marshal(pipe)
		utils.DebugLog("GenerateReportData", fmt.Sprintf("pipe: [ %s ]", queryJSON))

		opts := options.Aggregate().SetAllowDiskUse(true)
		_, err := c.Aggregate(ctx, pipe, opts)
		if err != nil {
			utils.ErrorLog("error GenerateReportData", err.Error())
			return err
		}

	} else {
		if len(reportInfo.SelectKeyInfos) > 0 {

			groupQuery := buildSelect(reportInfo.SelectKeyInfos)
			pipe = append(pipe, groupQuery...)

			pipe = append(pipe, bson.M{
				"$out": "report_" + reportInfo.ReportID,
			})

			queryJSON, _ := json.Marshal(pipe)
			utils.DebugLog("GenerateReportData", fmt.Sprintf("pipe: [ %s ]", queryJSON))

			opts := options.Aggregate().SetAllowDiskUse(true)
			_, err := c.Aggregate(ctx, pipe, opts)
			if err != nil {
				utils.ErrorLog("error GenerateReportData", err.Error())
				return err
			}
		}
	}

	return nil
}

// getCount 获取报表的总件数
func getCount(db string, reportInfo Report, match bson.M, total chan int64) error {

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("report_" + reportInfo.ReportID)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// if reportInfo.IsUseGroup {
	// 	pipe := []bson.M{{
	// 		"$match": match,
	// 	}}

	// 	group := buildGroupCount(reportInfo.AppID, reportInfo.GroupInfo)

	// 	pipe = append(pipe, group...)

	// 	queryJSON, _ := json.Marshal(pipe)
	// 	utils.DebugLog("GenerateReportData", fmt.Sprintf("pipe: [ %s ]", queryJSON))

	// 	var result TotalResult

	// 	opt := options.Aggregate()
	// 	opt.SetAllowDiskUse(true)

	// 	cur, err := c.Aggregate(ctx, pipe, opt)
	// 	if err != nil {
	// 		utils.ErrorLog("getCount", err.Error())
	// 		return err
	// 	}
	// 	defer cur.Close(ctx)
	// 	for cur.Next(ctx) {
	// 		var rep TotalResult
	// 		err := cur.Decode(&rep)
	// 		if err != nil {
	// 			utils.ErrorLog("getCount", err.Error())
	// 			return err
	// 		}
	// 		result = rep
	// 	}

	// 	total <- result.Total

	// 	return nil
	// }

	to, err := c.CountDocuments(ctx, match)
	if err != nil {
		utils.ErrorLog("getCount", err.Error())
		return err
	}

	total <- to

	return nil
}

// FindCount 获取报表的总件数
func FindCount(db, id string, owners []string) (map[string]*FieldInfo, int64, error) {

	// 获取报表设置信息
	reportInfo, err := FindReport(db, id)
	if err != nil {
		utils.ErrorLog("FindCount", err.Error())
		return nil, 0, err
	}

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("report_" + id)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fields, err := getAllFields(db, reportInfo.AppID, reportInfo.DatastoreID)
	if err != nil {
		utils.ErrorLog("FindCount", err.Error())
		return nil, 0, err
	}

	if reportInfo.IsUseGroup {
		pipe, fieldInfos := buildGroupFields(reportInfo.AppID, reportInfo.GroupInfo, fields, true)
		type TotalResult struct {
			Total int64 `bson:"total"`
		}

		var result TotalResult

		opt := options.Aggregate()
		opt.SetAllowDiskUse(true)

		cur, err := c.Aggregate(ctx, pipe, opt)
		if err != nil {
			utils.ErrorLog("FindCount", err.Error())
			return nil, 0, err
		}
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var rep TotalResult
			err := cur.Decode(&rep)
			if err != nil {
				utils.ErrorLog("FindCount", err.Error())
				return nil, 0, err
			}
			result = rep
		}

		return fieldInfos, result.Total, nil
	}

	query := bson.M{
		"owners": bson.M{
			"$in": owners,
		},
	}

	total, err := c.CountDocuments(ctx, query)
	if err != nil {
		utils.ErrorLog("FindCount", err.Error())
		return nil, 0, err
	}

	_, fieldInfos := buildSelectFields(reportInfo.AppID, reportInfo.SelectKeyInfos, fields)
	return fieldInfos, total, nil
}

func buildMatch(handleMonth, conditionType string, conditions []*ReportCondition) (result bson.M) {
	// 默认过滤掉被软删除的数据
	query := bson.M{}

	if len(conditions) > 0 {
		if conditionType == "and" {
			// AND的场合
			and := []bson.M{}
			for _, condition := range conditions {
				if condition.IsDynamic {
					switch condition.FieldType {
					case "text", "textarea", "autonum":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							}
							and = append(and, q)
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							and = append(and, q)
						} else {
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							and = append(and, q)
						}
					case "switch":
						// 只能是等于
						q := bson.M{
							"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
						}
						and = append(and, q)
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
							and = append(and, q)
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
							and = append(and, q)
						}
					case "lookup":
						// 只能是like和=两种情况（默认是等于）
						if condition.Operator == "like" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(condition.SearchValue), Options: "m"}},
							}
							and = append(and, q)
						} else if condition.Operator == "<>" {
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							and = append(and, q)
						} else {
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							and = append(and, q)
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
								and = append(and, q)
							}
						} else if condition.Operator == "<>" {
							// 不等于
							q := bson.M{
								"items." + condition.FieldID + ".value": bson.M{
									"$ne": condition.SearchValue,
								},
							}
							and = append(and, q)
						} else {
							// 等于
							q := bson.M{
								"items." + condition.FieldID + ".value": condition.SearchValue,
							}
							and = append(and, q)
						}
					case "number", "time":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							q := bson.M{
								"$and": []bson.M{
									{"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
									{"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
								},
							}
							and = append(and, q)
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gt": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								and = append(and, q)
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								and = append(and, q)
							} else if condition.Operator == "<" {
								// 小于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								and = append(and, q)
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$lte": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								and = append(and, q)
							} else if condition.Operator == "<>" {
								// 不等于的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": bson.M{"$ne": getSearchValue(condition.FieldType, condition.SearchValue)},
								}
								and = append(and, q)
							} else {
								// 默认的[等于]的场合
								q := bson.M{
									"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
								}
								and = append(and, q)
							}
						}
					case "date":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							q := bson.M{
								"$and": []bson.M{
									{"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
									{"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
								},
							}
							and = append(and, q)
						} else {
							// 如果传入值是固定参数【handleMonth】时，获取系统处理月度，然后比较
							if condition.SearchValue == "handleMonth" {
								if len(handleMonth) > 0 {
									value, _ := time.Parse("2006-01", handleMonth)
									if condition.Operator == ">" {
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gt": fullTime}}
										and = append(and, q)
									} else if condition.Operator == ">=" {
										// 大于等于的场合
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gte": value}}
										and = append(and, q)
									} else if condition.Operator == "<" {
										// 小于的场合
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": value}}
										and = append(and, q)
									} else if condition.Operator == "<=" {
										// 小于等于的场合
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": fullTime}}
										and = append(and, q)
									} else if condition.Operator == "<>" {
										// 不等于的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$or": []bson.M{
											{"items." + condition.FieldID + ".value": bson.M{"$gte": fullTime}},
											{"items." + condition.FieldID + ".value": bson.M{"$lt": zeroTime}},
										}}

										and = append(and, q)
									} else {
										// 默认的[等于]的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$and": []bson.M{
											{"items." + condition.FieldID + ".value": bson.M{"$gte": zeroTime}},
											{"items." + condition.FieldID + ".value": bson.M{"$lt": fullTime}},
										}}

										and = append(and, q)
									}
								}
								// 如果传入值是固定参数【now】时，获取系统当前日期，然后比较
							} else if condition.SearchValue == "now" {
								value, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
								if condition.Operator == ">" {
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gt": value}}
									and = append(and, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gte": value}}
									and = append(and, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": value}}
									and = append(and, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lte": value}}
									and = append(and, q)
								} else if condition.Operator == "<>" {
									// 不等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$ne": value}}
									and = append(and, q)
								} else {
									// 默认的[等于]的场合
									q := bson.M{"items." + condition.FieldID + ".value": value}
									and = append(and, q)
								}
							} else {
								// 可以是=,>,<,>=,<=,<>(默认是等于)
								// 大于的场合
								if condition.Operator == ">" {
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gt": getTime(condition.SearchValue)}}
									and = append(and, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gte": getTime(condition.SearchValue)}}
									and = append(and, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": getTime(condition.SearchValue)}}
									and = append(and, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lte": getTime(condition.SearchValue)}}
									and = append(and, q)
								} else if condition.Operator == "<>" {
									// 不等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$ne": getTime(condition.SearchValue)}}
									and = append(and, q)
								} else {
									// 默认的[等于]的场合
									q := bson.M{"items." + condition.FieldID + ".value": getTime(condition.SearchValue)}
									and = append(and, q)
								}
							}
						}
					default:
					}
				} else {
					switch condition.FieldType {
					case "options", "type", "user":
						if condition.Operator == "in" {
							values := strings.Split(condition.SearchValue, ",")
							if len(values) > 0 {
								// IN
								q := bson.M{
									condition.FieldID: bson.M{"$in": values},
								}
								and = append(and, q)
							}
						} else if condition.Operator == "<>" {
							// 不等于
							q := bson.M{
								condition.FieldID: bson.M{
									"$ne": condition.SearchValue,
								},
							}
							and = append(and, q)
						} else {
							values := strings.Split(condition.SearchValue, ",")
							// 等于
							q := bson.M{
								condition.FieldID: values,
							}
							and = append(and, q)
						}
					case "check":
						// 等于
						q := bson.M{
							condition.FieldID: condition.SearchValue,
						}
						and = append(and, q)
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
							and = append(and, q)
						} else {
							if condition.SearchValue == "handleMonth" {
								// 如果传入值是固定参数【handleMonth】时，获取系统处理月度，然后比较
								if len(handleMonth) > 0 {
									value, _ := time.Parse("2006-01", handleMonth)
									if condition.Operator == ">" {
										// 大于的场合
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{condition.FieldID: bson.M{"$gt": fullTime}}
										and = append(and, q)
									} else if condition.Operator == ">=" {
										// 大于等于的场合
										q := bson.M{condition.FieldID: bson.M{"$gte": value}}
										and = append(and, q)
									} else if condition.Operator == "<" {
										// 小于的场合
										q := bson.M{condition.FieldID: bson.M{"$lt": value}}
										and = append(and, q)
									} else if condition.Operator == "<=" {
										// 小于等于的场合
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{condition.FieldID: bson.M{"$lte": fullTime}}
										and = append(and, q)
									} else if condition.Operator == "<>" {
										// 不等于的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$or": []bson.M{
											{condition.FieldID: bson.M{"$gte": fullTime}},
											{condition.FieldID: bson.M{"$lt": zeroTime}},
										}}

										and = append(and, q)
									} else {
										// 默认的[等于]的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$and": []bson.M{
											{condition.FieldID: bson.M{"$gte": zeroTime}},
											{condition.FieldID: bson.M{"$lt": fullTime}},
										}}
										and = append(and, q)
									}
								}
								// 如果传入值是固定参数【now】时，获取系统当前日期，然后比较
							} else if condition.SearchValue == "now" {
								value, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
								if condition.Operator == ">" {
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										condition.FieldID: bson.M{"$gte": fullTime},
									}
									and = append(and, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									q := bson.M{
										condition.FieldID: bson.M{"$gte": value},
									}
									and = append(and, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									q := bson.M{
										condition.FieldID: bson.M{"$lt": value},
									}
									and = append(and, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										condition.FieldID: bson.M{"$lte": fullTime},
									}
									and = append(and, q)
								} else if condition.Operator == "<>" {
									// 不等于的场合
									zeroTime := value
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										"$or": []bson.M{
											{condition.FieldID: bson.M{"$gte": fullTime}},
											{condition.FieldID: bson.M{"$lt": zeroTime}},
										},
									}
									and = append(and, q)
								} else {
									// 默认的[等于]的场合
									zeroTime := value
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										"$and": []bson.M{
											{condition.FieldID: bson.M{"$gte": zeroTime}},
											{condition.FieldID: bson.M{"$lt": fullTime}},
										},
									}
									and = append(and, q)
								}
							} else {
								// 可以是=,>,<,>=,<=,<>(默认是等于)
								// 大于的场合
								if condition.Operator == ">" {
									value := getTime(condition.SearchValue)
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										condition.FieldID: bson.M{"$gte": fullTime},
									}
									and = append(and, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									value := getTime(condition.SearchValue)
									q := bson.M{
										condition.FieldID: bson.M{"$gte": value},
									}
									and = append(and, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									value := getTime(condition.SearchValue)
									q := bson.M{
										condition.FieldID: bson.M{"$lt": value},
									}
									and = append(and, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									value := getTime(condition.SearchValue)
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										condition.FieldID: bson.M{"$lt": fullTime},
									}
									and = append(and, q)
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
									and = append(and, q)
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
									and = append(and, q)
								}
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
			for _, condition := range conditions {
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
							// 不等于
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
					case "number", "time":
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
					case "date":
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
							if condition.SearchValue == "handleMonth" {
								if len(handleMonth) > 0 {
									value, _ := time.Parse("2006-01", handleMonth)
									if condition.Operator == ">" {
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gt": fullTime}}
										or = append(or, q)
									} else if condition.Operator == ">=" {
										// 大于等于的场合
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gte": value}}
										or = append(or, q)
									} else if condition.Operator == "<" {
										// 小于的场合
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": value}}
										or = append(or, q)
									} else if condition.Operator == "<=" {
										// 小于等于的场合
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": fullTime}}
										or = append(or, q)
									} else if condition.Operator == "<>" {
										// 不等于的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$or": []bson.M{
											{"items." + condition.FieldID + ".value": bson.M{"$gte": fullTime}},
											{"items." + condition.FieldID + ".value": bson.M{"$lt": zeroTime}},
										}}

										or = append(or, q)
									} else {
										// 默认的[等于]的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$and": []bson.M{
											{"items." + condition.FieldID + ".value": bson.M{"$gte": zeroTime}},
											{"items." + condition.FieldID + ".value": bson.M{"$lt": fullTime}},
										}}
										or = append(or, q)
									}
								}
								// 如果传入值是固定参数【now】时，获取系统当前日期，然后比较
							} else if condition.SearchValue == "now" {
								value, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
								if condition.Operator == ">" {
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gt": value}}
									or = append(or, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gte": value}}
									or = append(or, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": value}}
									or = append(or, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lte": value}}
									or = append(or, q)
								} else if condition.Operator == "<>" {
									// 不等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$ne": value}}
									or = append(or, q)
								} else {
									// 默认的[等于]的场合
									q := bson.M{"items." + condition.FieldID + ".value": value}
									or = append(or, q)
								}
							} else {
								// 可以是=,>,<,>=,<=,<>(默认是等于)
								// 大于的场合
								if condition.Operator == ">" {
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gt": getTime(condition.SearchValue)}}
									or = append(or, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$gte": getTime(condition.SearchValue)}}
									or = append(or, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lt": getTime(condition.SearchValue)}}
									or = append(or, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$lte": getTime(condition.SearchValue)}}
									or = append(or, q)
								} else if condition.Operator == "<>" {
									// 不等于的场合
									q := bson.M{"items." + condition.FieldID + ".value": bson.M{"$ne": getTime(condition.SearchValue)}}
									or = append(or, q)
								} else {
									// 默认的[等于]的场合
									q := bson.M{"items." + condition.FieldID + ".value": getTime(condition.SearchValue)}
									or = append(or, q)
								}
							}
						}
					default:
					}
				} else {
					switch condition.FieldType {
					case "options", "type", "user":
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
							// 不等于
							q := bson.M{
								condition.FieldID: bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							values := strings.Split(condition.SearchValue, ",")
							// 等于
							q := bson.M{
								condition.FieldID: values,
							}
							or = append(or, q)
						}
					case "check":
						// 等于
						q := bson.M{
							condition.FieldID: condition.SearchValue,
						}
						or = append(or, q)
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
							if condition.SearchValue == "handleMonth" {
								// 如果传入值是固定参数【handleMonth】时，获取系统处理月度，然后比较
								if len(handleMonth) > 0 {
									value, _ := time.Parse("2006-01", handleMonth)
									if condition.Operator == ">" {
										// 大于的场合
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{condition.FieldID: bson.M{"$gt": fullTime}}
										or = append(or, q)
									} else if condition.Operator == ">=" {
										// 大于等于的场合
										q := bson.M{condition.FieldID: bson.M{"$gte": value}}
										or = append(or, q)
									} else if condition.Operator == "<" {
										// 小于的场合
										q := bson.M{condition.FieldID: bson.M{"$lt": value}}
										or = append(or, q)
									} else if condition.Operator == "<=" {
										// 小于等于的场合
										fullTime := value.AddDate(0, 1, 0)
										q := bson.M{condition.FieldID: bson.M{"$lte": fullTime}}
										or = append(or, q)
									} else if condition.Operator == "<>" {
										// 不等于的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$or": []bson.M{
											{condition.FieldID: bson.M{"$gte": fullTime}},
											{condition.FieldID: bson.M{"$lt": zeroTime}},
										}}

										or = append(or, q)
									} else {
										// 默认的[等于]的场合
										zeroTime := value
										fullTime := value.AddDate(0, 1, 0)

										q := bson.M{"$and": []bson.M{
											{condition.FieldID: bson.M{"$gte": zeroTime}},
											{condition.FieldID: bson.M{"$lt": fullTime}},
										}}
										or = append(or, q)
									}
								}
								// 如果传入值是固定参数【now】时，获取系统当前日期，然后比较
							} else if condition.SearchValue == "now" {
								value, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
								// 大于的场合
								if condition.Operator == ">" {
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										condition.FieldID: bson.M{"$gte": fullTime},
									}
									or = append(or, q)
								} else if condition.Operator == ">=" {
									// 大于等于的场合
									q := bson.M{
										condition.FieldID: bson.M{"$gte": value},
									}
									or = append(or, q)
								} else if condition.Operator == "<" {
									// 小于的场合
									q := bson.M{
										condition.FieldID: bson.M{"$lt": value},
									}
									or = append(or, q)
								} else if condition.Operator == "<=" {
									// 小于等于的场合
									fullTime := value.Add(time.Hour * 24)
									q := bson.M{
										condition.FieldID: bson.M{"$lte": fullTime},
									}
									or = append(or, q)
								} else if condition.Operator == "<>" {
									// 不等于的场合
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
						}
					default:
						break
					}
				}
			}

			if len(or) > 0 {
				query["$or"] = or
			}
		}
	}

	return bson.M{
		"$match": query,
	}
}

func buildGroup(group *GroupInfo) (result []bson.M) {
	pipe := []bson.M{}
	groupQuery := bson.M{}
	projectQuery := bson.M{}

	idQuery := bson.M{}
	if len(group.GroupKeys) > 0 {
		for _, groupKey := range group.GroupKeys {
			// 分组字段
			if groupKey.IsDynamic {
				if groupKey.IsLookup {
					// 以点分割字段 ID
					ps := strings.Split(groupKey.FieldID, "#")
					// 第一段为当前台账的关联key
					relationKey := ps[0]
					// 第二段为关联台账的其他字段
					fieldId := ps[1]

					idQuery[groupKey.FieldID] = "$relations." + relationKey + "." + fieldId + ".value"

					projectQuery["items."+groupKey.FieldID] = bson.M{
						"value":     "$_id." + groupKey.FieldID,
						"data_type": groupKey.DataType,
					}
				} else {
					idQuery[groupKey.FieldID] = "$items." + groupKey.FieldID
					projectQuery["items."+groupKey.FieldID] = "$_id." + groupKey.FieldID
				}

			} else {
				idQuery[groupKey.FieldID] = "$" + groupKey.FieldID
				projectQuery[groupKey.FieldID] = "$_id." + groupKey.FieldID

			}
		}

		idQuery["owners"] = "$owners"
		projectQuery["owners"] = "$_id.owners"

		groupQuery["_id"] = idQuery
	}

	if len(group.AggreKeys) > 0 {
		for _, aggreKey := range group.AggreKeys {
			if aggreKey.IsLookup {
				// 以点分割字段 ID
				ps := strings.Split(aggreKey.FieldID, "#")
				// 第一段为当前台账的关联key
				relationKey := ps[0]
				// 第二段为关联台账的其他字段
				fieldId := ps[1]

				switch aggreKey.AggreType {
				case "sum":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$sum": "$relations." + relationKey + "." + fieldId + ".value",
					}
				case "avg":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$avg": "$relations." + relationKey + "." + fieldId + ".value",
					}
				case "max":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$max": "$relations." + relationKey + "." + fieldId + ".value",
					}
				case "min":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$min": "$relations." + relationKey + "." + fieldId + ".value",
					}
				}
			} else {
				switch aggreKey.AggreType {
				case "sum":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$sum": "$items." + aggreKey.FieldID + ".value",
					}
				case "avg":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$avg": "$items." + aggreKey.FieldID + ".value",
					}
				case "max":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$max": "$items." + aggreKey.FieldID + ".value",
					}
				case "min":
					groupQuery[aggreKey.FieldID] = bson.M{
						"$min": "$items." + aggreKey.FieldID + ".value",
					}
				}
			}
			// 集计字段
			projectQuery["items."+aggreKey.FieldID] = bson.M{
				"value":     "$" + aggreKey.FieldID,
				"data_type": "number",
			}
		}
	}

	if group.ShowCount {
		groupQuery["count"] = bson.M{
			"$sum": 1,
		}
		projectQuery["count"] = 1
	}

	projectQuery["update_time"] = primitive.NewDateTimeFromTime(time.Now())

	pipe = append(pipe, bson.M{
		"$group": groupQuery,
	})
	pipe = append(pipe, bson.M{
		"$project": projectQuery,
	})

	return pipe
}

func buildGroupFields(appId string, group *GroupInfo, fields map[string][]*Field, getCount bool) ([]bson.M, map[string]*FieldInfo) {
	pipe := []bson.M{}
	groupQuery := bson.M{}
	sortQuery := bson.D{}
	defaultSortQuery := bson.D{}
	keys := make(map[string]*FieldInfo)

	lookupProject := bson.M{
		"count":       1,
		"update_time": "$_id.update_time",
	}

	idQuery := bson.M{
		"update_time": "$update_time",
	}
	// 将所有 lookup 的字段选出，并去掉重复的
	for _, groupKey := range group.GroupKeys {
		keys[groupKey.FieldID] = &FieldInfo{
			DataType:    groupKey.DataType,
			AliasName:   groupKey.AliasName,
			DatastoreID: groupKey.DatastoreID,
			OptionID:    groupKey.OptionID,
			IsDynamic:   groupKey.IsDynamic,
			Order:       groupKey.Order,
			Unique:      groupKey.Unique,
		}

		if groupKey.IsDynamic {

			if groupKey.DataType == "user" {
				pp := []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$in": []string{"$user_id", "$$user"},
									},
								},
							},
						},
					},
				}

				lookup := bson.M{
					"from": "users",
					"let": bson.M{
						"user": "$items." + groupKey.FieldID + ".value",
					},
					"pipeline": pp,
					"as":       "relations_" + groupKey.FieldID,
				}

				pipe = append(pipe, bson.M{
					"$lookup": lookup,
				})

				idQuery[groupKey.FieldID] = "$relations_" + groupKey.FieldID + ".user_name"
				lookupProject["items."+groupKey.FieldID+".value"] = "$_id." + groupKey.FieldID
				lookupProject["items."+groupKey.FieldID+".data_type"] = groupKey.DataType

				// 排序
				if len(groupKey.Sort) > 0 {
					if groupKey.Sort == "ascend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + groupKey.FieldID + ".value", Value: 1})
					}
					if groupKey.Sort == "dscend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + groupKey.FieldID + ".value", Value: -1})
					}
				}

				continue
			}

			if groupKey.DataType == "options" {
				pp := []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$eq": []string{"$app_id", "$$app_id"},
									},
									{
										"$eq": []string{"$option_id", "$$option_id"},
									},
									{
										"$eq": []string{"$option_value", "$$option_value"},
									},
								},
							},
						},
					},
				}

				lookup := bson.M{
					"from": "options",
					"let": bson.M{
						"app_id":       appId,
						"option_id":    groupKey.OptionID,
						"option_value": "$items." + groupKey.FieldID + ".value",
					},
					"pipeline": pp,
					"as":       "relations_" + groupKey.FieldID,
				}

				unwind := bson.M{
					"path":                       "$relations_" + groupKey.FieldID,
					"preserveNullAndEmptyArrays": true,
				}

				pipe = append(pipe, bson.M{
					"$lookup": lookup,
				})
				pipe = append(pipe, bson.M{
					"$unwind": unwind,
				})

				idQuery[groupKey.FieldID] = "$relations_" + groupKey.FieldID + ".option_label"
				lookupProject["items."+groupKey.FieldID+".value"] = "$_id." + groupKey.FieldID
				lookupProject["items."+groupKey.FieldID+".data_type"] = groupKey.DataType

				// 排序
				if len(groupKey.Sort) > 0 {
					if groupKey.Sort == "ascend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + groupKey.FieldID + ".value", Value: 1})
					}
					if groupKey.Sort == "dscend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + groupKey.FieldID + ".value", Value: -1})
					}
				}

				continue
			}

			// 排序
			if len(groupKey.Sort) > 0 {
				if groupKey.Sort == "ascend" {
					sortQuery = append(sortQuery, bson.E{Key: "items." + groupKey.FieldID + ".value", Value: 1})
				}
				if groupKey.Sort == "dscend" {
					sortQuery = append(sortQuery, bson.E{Key: "items." + groupKey.FieldID + ".value", Value: -1})
				}
			}

			idQuery[groupKey.FieldID] = "$items." + groupKey.FieldID
			lookupProject["items."+groupKey.FieldID] = "$_id." + groupKey.FieldID

			continue
		}

		idQuery[groupKey.FieldID] = "$" + groupKey.FieldID

		lookupProject[groupKey.FieldID] = "$_id." + groupKey.FieldID

		if len(groupKey.Sort) > 0 {
			if groupKey.Sort == "ascend" {
				sortQuery = append(sortQuery, bson.E{Key: groupKey.FieldID, Value: 1})
			}
			if groupKey.Sort == "dscend" {
				sortQuery = append(sortQuery, bson.E{Key: groupKey.FieldID, Value: -1})
			}
		}
	}

	groupQuery["_id"] = idQuery

	for _, aggreKey := range group.AggreKeys {

		keys[aggreKey.FieldID] = &FieldInfo{
			DataType:    aggreKey.DataType,
			DatastoreID: aggreKey.DatastoreID,
			AliasName:   aggreKey.AliasName,
			OptionID:    aggreKey.OptionID,
			IsDynamic:   true,
			Order:       aggreKey.Order,
			Unique:      false,
		}

		switch aggreKey.AggreType {
		case "sum":
			groupQuery[aggreKey.FieldID] = bson.M{
				"$sum": "$items." + aggreKey.FieldID + ".value",
			}
		case "avg":
			groupQuery[aggreKey.FieldID] = bson.M{
				"$avg": "$items." + aggreKey.FieldID + ".value",
			}
		case "max":
			groupQuery[aggreKey.FieldID] = bson.M{
				"$max": "$items." + aggreKey.FieldID + ".value",
			}
		case "min":
			groupQuery[aggreKey.FieldID] = bson.M{
				"$min": "$items." + aggreKey.FieldID + ".value",
			}
		}

		// 集计字段
		lookupProject["items."+aggreKey.FieldID] = bson.M{
			"value":     "$" + aggreKey.FieldID,
			"data_type": "number",
		}

		if len(aggreKey.Sort) > 0 {
			if aggreKey.Sort == "ascend" {
				sortQuery = append(sortQuery, bson.E{Key: "items." + aggreKey.FieldID + ".value", Value: 1})
			}
			if aggreKey.Sort == "dscend" {
				sortQuery = append(sortQuery, bson.E{Key: "items." + aggreKey.FieldID + ".value", Value: -1})
			}
		}
	}

	if group.ShowCount {
		keys["count"] = &FieldInfo{
			DataType:  "number",
			AliasName: "count",
			IsDynamic: false,
			Order:     10000,
		}

		groupQuery["count"] = bson.M{
			"$sum": "$count",
		}
		lookupProject["count"] = 1
	}

	pipe = append(pipe, bson.M{
		"$group": groupQuery,
	})

	if getCount {
		pipe = append(pipe, bson.M{
			"$count": "total",
		})

		return pipe, keys
	}

	pipe = append(pipe, bson.M{
		"$project": lookupProject,
	})

	if len(sortQuery) > 0 {
		pipe = append(pipe, bson.M{"$sort": sortQuery})
	} else {
		for _, groupKey := range group.GroupKeys {
			if groupKey.DataType != "user" {
				defaultSortQuery = append(defaultSortQuery, bson.E{Key: "items." + groupKey.FieldID + ".value", Value: 1})
			}
		}
		if len(defaultSortQuery) == 0 {
			defaultSortQuery = append(defaultSortQuery, bson.E{Key: "update_time", Value: 1})
		}
		pipe = append(pipe, bson.M{"$sort": defaultSortQuery})
	}

	return pipe, keys
}

// func buildGroupCount(appId string, group *GroupInfo) []bson.M {
// 	pipe := []bson.M{}
// 	groupQuery := bson.M{}

// 	idQuery := bson.M{
// 		"update_time": "$update_time",
// 	}
// 	// 将所有 lookup 的字段选出，并去掉重复的
// 	for _, groupKey := range group.GroupKeys {
// 		if groupKey.IsDynamic {

// 			idQuery[groupKey.FieldID] = "$items." + groupKey.FieldID

// 			continue
// 		}

// 		idQuery[groupKey.FieldID] = "$" + groupKey.FieldID
// 	}

// 	groupQuery["_id"] = idQuery

// 	for _, aggreKey := range group.AggreKeys {
// 		switch aggreKey.AggreType {
// 		case "sum":
// 			groupQuery[aggreKey.FieldID] = bson.M{
// 				"$sum": "$items." + aggreKey.FieldID + ".value",
// 			}
// 		case "avg":
// 			groupQuery[aggreKey.FieldID] = bson.M{
// 				"$avg": "$items." + aggreKey.FieldID + ".value",
// 			}
// 		case "max":
// 			groupQuery[aggreKey.FieldID] = bson.M{
// 				"$max": "$items." + aggreKey.FieldID + ".value",
// 			}
// 		case "min":
// 			groupQuery[aggreKey.FieldID] = bson.M{
// 				"$min": "$items." + aggreKey.FieldID + ".value",
// 			}
// 		}
// 	}

// 	if group.ShowCount {
// 		groupQuery["count"] = bson.M{
// 			"$sum": "$count",
// 		}
// 	}

// 	pipe = append(pipe, bson.M{
// 		"$group": groupQuery,
// 	})

// 	pipe = append(pipe, bson.M{
// 		"$count": "total",
// 	})

// 	return pipe
// }

func buildSelect(selectKeys []*KeyInfo) []bson.M {
	projectQuery := bson.M{}
	pipe := []bson.M{}

	for _, key := range selectKeys {
		// 分组字段
		if key.IsDynamic {
			if key.IsLookup {

				// 以点分割字段 ID
				ps := strings.Split(key.FieldID, "#")
				// 第一段为当前台账的关联key
				relationKey := ps[0]
				// 第二段为关联台账的其他字段
				fieldId := ps[1]

				projectQuery["items."+key.FieldID] = bson.M{
					"value":     "$relations." + relationKey + "." + fieldId + ".value",
					"data_type": key.DataType,
				}
			} else {
				projectQuery["items."+key.FieldID] = "$items." + key.FieldID
			}
		} else {
			projectQuery[key.FieldID] = "$" + key.FieldID
		}
	}

	projectQuery["owners"] = "$owners"
	projectQuery["item_id"] = "$item_id"

	projectQuery["update_time"] = primitive.NewDateTimeFromTime(time.Now())

	pipe = append(pipe, bson.M{
		"$project": projectQuery,
	})

	return pipe
}

func buildSelectFields(appId string, selectKeys []*KeyInfo, fields map[string][]*Field) ([]bson.M, map[string]*FieldInfo) {
	pipe := []bson.M{}
	sortQuery := bson.D{}
	defaultSortQuery := bson.D{}
	keys := make(map[string]*FieldInfo)

	lookupProject := bson.M{
		"count":       1,
		"update_time": 1,
		"item_id":     1,
	}
	for _, key := range selectKeys {

		keys[key.FieldID] = &FieldInfo{
			DataType:    key.DataType,
			AliasName:   key.AliasName,
			DatastoreID: key.DatastoreID,
			IsDynamic:   key.IsDynamic,
			OptionID:    key.OptionID,
			Order:       key.Order,
			Unique:      key.Unique,
		}

		if key.IsDynamic {
			if key.DataType == "user" {
				pp := []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$in": []string{"$user_id", "$$user"},
									},
								},
							},
						},
					},
				}

				lookup := bson.M{
					"from": "users",
					"let": bson.M{
						"user": "$items." + key.FieldID + ".value",
					},
					"pipeline": pp,
					"as":       "relations_" + key.FieldID,
				}

				pipe = append(pipe, bson.M{
					"$lookup": lookup,
				})

				lookupProject["items."+key.FieldID+".value"] = "$relations_" + key.FieldID + ".user_name"
				lookupProject["items."+key.FieldID+".data_type"] = "user"

				if len(key.Sort) > 0 {
					if key.Sort == "ascend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + key.FieldID + ".value", Value: 1})
					}
					if key.Sort == "dscend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + key.FieldID + ".value", Value: -1})
					}
				}

				continue
			}

			if key.DataType == "options" {
				pp := []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$eq": []string{"$app_id", "$$app_id"},
									},
									{
										"$eq": []string{"$option_id", "$$option_id"},
									},
									{
										"$eq": []string{"$option_value", "$$option_value"},
									},
								},
							},
						},
					},
				}

				lookup := bson.M{
					"from": "options",
					"let": bson.M{
						"app_id":       appId,
						"option_id":    key.OptionID,
						"option_value": "$items." + key.FieldID + ".value",
					},
					"pipeline": pp,
					"as":       "relations_" + key.FieldID,
				}

				unwind := bson.M{
					"path":                       "$relations_" + key.FieldID,
					"preserveNullAndEmptyArrays": true,
				}

				pipe = append(pipe, bson.M{
					"$lookup": lookup,
				})
				pipe = append(pipe, bson.M{
					"$unwind": unwind,
				})

				lookupProject["items."+key.FieldID+".value"] = "$relations_" + key.FieldID + ".option_label"
				lookupProject["items."+key.FieldID+".data_type"] = "options"

				if len(key.Sort) > 0 {
					if key.Sort == "ascend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + key.FieldID + ".value", Value: 1})
					}
					if key.Sort == "dscend" {
						sortQuery = append(sortQuery, bson.E{Key: "items." + key.FieldID + ".value", Value: -1})
					}
				}

				continue
			}

			lookupProject["items."+key.FieldID] = "$items." + key.FieldID

			if len(key.Sort) > 0 {
				if key.Sort == "ascend" {
					sortQuery = append(sortQuery, bson.E{Key: "items." + key.FieldID + ".value", Value: 1})
				}
				if key.Sort == "dscend" {
					sortQuery = append(sortQuery, bson.E{Key: "items." + key.FieldID + ".value", Value: -1})
				}
			}

			continue
		}

		lookupProject[key.FieldID] = "$" + key.FieldID

		if len(key.Sort) > 0 {
			if key.Sort == "ascend" {
				sortQuery = append(sortQuery, bson.E{Key: key.FieldID, Value: 1})
			}
			if key.Sort == "dscend" {
				sortQuery = append(sortQuery, bson.E{Key: key.FieldID, Value: -1})
			}
		}

		continue
	}

	pipe = append(pipe, bson.M{
		"$project": lookupProject,
	})

	if len(sortQuery) > 0 {
		pipe = append(pipe, bson.M{
			"$sort": sortQuery,
		})
	} else {
		defaultSortQuery = append(defaultSortQuery, bson.E{Key: "item_id", Value: 1})
		pipe = append(pipe, bson.M{
			"$sort": defaultSortQuery,
		})
	}

	return pipe, keys
}

// RecoverSelectReports 恢复选中报表情报
func RecoverSelectReports(db string, reportidlist []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ReportsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectReports", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectReports", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, reportid := range reportidlist {
			objectID, err := primitive.ObjectIDFromHex(reportid)
			if err != nil {
				utils.ErrorLog("error RecoverSelectReports", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"updated_at": time.Now(),
				"updated_by": userID,
				"deleted_by": "",
			}}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("RecoverSelectReports", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectReports", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectReports", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectReports", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectReports", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

func buildReportMatch(conditionType string, conditions []*ReportCondition, owners []string) (result bson.M, keys []string, keymaps bson.M) {
	indexKeys := []string{"owners"}
	indexMap := bson.M{
		"owners": 1,
	}

	query := bson.M{
		"owners": bson.M{"$in": owners},
	}

	if len(conditions) > 0 {
		if conditionType == "and" {
			// OR的场合
			and := []bson.M{}
			for _, condition := range conditions {
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
						indexMap["items."+condition.FieldID+".value"] = 1
						indexKeys = append(indexKeys, "items."+condition.FieldID+".value")
					case "switch":
						// 只能是等于
						and = append(and, bson.M{
							"items." + condition.FieldID + ".value": getSearchValue(condition.FieldType, condition.SearchValue),
						})
						indexMap["items."+condition.FieldID+".value"] = 1
						indexKeys = append(indexKeys, "items."+condition.FieldID+".value")
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
						indexMap["items."+condition.FieldID+".value"] = 1
						indexKeys = append(indexKeys, "items."+condition.FieldID+".value")
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
						indexMap["items."+condition.FieldID+".value"] = 1
						indexKeys = append(indexKeys, "items."+condition.FieldID+".value")
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
					case "number", "time":
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
						indexMap["items."+condition.FieldID+".value"] = 1
						indexKeys = append(indexKeys, "items."+condition.FieldID+".value")
					case "date":
						// 1 表示两个数据之间（searchValue用“~”隔开）,前>=,后小于<
						if condition.ConditionType == "1" {
							values := strings.Split(condition.SearchValue, "~")
							query["$and"] = []bson.M{
								{"items." + condition.FieldID + ".value": bson.M{"$gte": getSearchValue(condition.FieldType, values[0])}},
								{"items." + condition.FieldID + ".value": bson.M{"$lt": getSearchValue(condition.FieldType, values[1])}},
							}
						} else {
							// 可以是=,>,<,>=,<=,<>(默认是等于)
							// 大于的场合
							if condition.Operator == ">" {
								value := getTime(condition.SearchValue)
								query["items."+condition.FieldID+".value"] = bson.M{"$gt": value}
							} else if condition.Operator == ">=" {
								// 大于等于的场合
								value := getTime(condition.SearchValue)
								query["items."+condition.FieldID+".value"] = bson.M{"$gte": value}
							} else if condition.Operator == "<" {
								// 小于的场合
								value := getTime(condition.SearchValue)
								query["items."+condition.FieldID+".value"] = bson.M{"$lt": value}
							} else if condition.Operator == "<=" {
								// 小于等于的场合
								value := getTime(condition.SearchValue)
								query["items."+condition.FieldID+".value"] = bson.M{"$lte": value}
							} else if condition.Operator == "<>" {
								// 不等于的场合
								value := getTime(condition.SearchValue)

								query["items."+condition.FieldID+".value"] = bson.M{"$ne": value}
							} else {
								// 默认的[等于]的场合
								value := getTime(condition.SearchValue)
								query["items."+condition.FieldID+".value"] = value
							}
						}
						indexMap["items."+condition.FieldID+".value"] = 1
						indexKeys = append(indexKeys, "items."+condition.FieldID+".value")
					default:
					}
				} else {
					switch condition.FieldType {
					case "options", "type", "user":
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
						indexMap[condition.FieldID] = 1
						indexKeys = append(indexKeys, condition.FieldID)
					case "check":
						value := condition.SearchValue
						// 等于
						query[condition.FieldID] = value
						indexMap[condition.FieldID] = 1
						indexKeys = append(indexKeys, condition.FieldID)
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
						indexMap[condition.FieldID] = 1
						indexKeys = append(indexKeys, condition.FieldID)
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
			for _, condition := range conditions {
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
							// 不等于
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
					case "number", "time":
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
					case "date":
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
					case "options", "type", "user":
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
							// 不等于
							q := bson.M{
								condition.FieldID: bson.M{
									"$ne": condition.SearchValue,
								},
							}
							or = append(or, q)
						} else {
							values := strings.Split(condition.SearchValue, ",")
							// 等于
							q := bson.M{
								condition.FieldID: values,
							}
							or = append(or, q)
						}
					case "check":
						values := strings.Split(condition.SearchValue, ",")
						// 等于
						q := bson.M{
							condition.FieldID: values,
						}
						or = append(or, q)
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

	return query, indexKeys, indexMap
}

func DeleteLeastUsed(c *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	us, err := GetIndexUsage(c)
	if err != nil {
		utils.ErrorLog("DeleteLeastUsed", err.Error())
		return nil
	}

	// 当前索引未超过最大值的情况下，不需要删除索引
	if len(us) < MaxIndexCount {
		return nil
	}

	for i := MaxIndexCount; i < len(us)-1; i++ {
		if _, err := c.Indexes().DropOne(ctx, us[i].Name); err != nil {
			utils.ErrorLog("DeleteLeastUsed", err.Error())
			return err
		}
	}

	return nil
}

func GetIndexUsage(c *mongo.Collection) (u []*IndexUsage, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	query := []bson.M{
		{"$indexStats": bson.M{}},
		{"$project": bson.M{
			"name":  1,
			"usage": "$accesses.ops",
		}},
		{"$sort": bson.M{"usage": -1}},
	}

	var result []*IndexUsage
	cur, err := c.Aggregate(ctx, query)
	if err != nil {
		utils.ErrorLog("GetIndexUsage", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("GetIndexUsage", err.Error())
		return nil, err
	}

	return result, nil
}
