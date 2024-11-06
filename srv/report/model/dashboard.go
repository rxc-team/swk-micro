/*
 * @Description:仪表盘（model）
 * @Author: RXC 廖云江
 * @Date: 2019-08-19 14:25:58
 * @LastEditors: RXC 廖云江
 * @LastEditTime: 2021-02-20 13:21:33
 */

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
	"rxcsoft.cn/pit3/srv/report/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// DashboardsCollection dashboards collection
	DashboardsCollection = "dashboards"
)

type (
	// Dashboard 仪表盘信息
	Dashboard struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		DashboardID   string             `json:"dashboard_id" bson:"dashboard_id"`
		DashboardName string             `json:"dashboard_name" bson:"dashboard_name"`
		Domain        string             `json:"domain" bson:"domain"`
		AppID         string             `json:"app_id" bson:"app_id"`
		ReportID      string             `json:"report_id" bson:"report_id"`
		DashboardType string             `json:"dashboard_type" bson:"dashboard_type"`
		XRange        []float32          `json:"x_range" bson:"x_range"`
		YRange        []float32          `json:"y_range" bson:"y_range"`
		TickType      string             `json:"tick_type" bson:"tick_type"`
		Ticks         []int64            `json:"ticks" bson:"ticks"`
		TickCount     int64              `json:"tick_count" bson:"tick_count"`
		GFieldID      string             `json:"g_field_id" bson:"g_field_id"`
		XFieldID      string             `json:"x_field_id" bson:"x_field_id"`
		YFieldID      string             `json:"y_field_id" bson:"y_field_id"`
		LimitInPlot   bool               `json:"limit_in_plot" bson:"limit_in_plot"`
		StepType      string             `json:"step_type" bson:"step_type"`
		IsStack       bool               `json:"is_stack" bson:"is_stack"`
		IsPercent     bool               `json:"is_percent" bson:"is_percent"`
		IsGroup       bool               `json:"is_group" bson:"is_group"`
		Smooth        bool               `json:"smooth" bson:"smooth"`
		MinBarWidth   float32            `json:"min_bar_width" bson:"min_bar_width"`
		MaxBarWidth   float32            `json:"max_bar_width" bson:"max_bar_width"`
		Radius        float32            `json:"radius" bson:"radius"`
		InnerRadius   float32            `json:"inner_radius" bson:"inner_radius"`
		StartAngle    float32            `json:"start_angle" bson:"start_angle"`
		EndAngle      float32            `json:"end_angle" bson:"end_angle"`
		Slider        Slider             `json:"slider" bson:"slider"`
		Scrollbar     Scrollbar          `json:"scrollbar" bson:"scrollbar"`
		CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy     string             `json:"created_by" bson:"created_by"`
		UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy     string             `json:"updated_by" bson:"updated_by"`
		DeletedAt     time.Time          `json:"deleted_at" bson:"deleted_at"`
		DeletedBy     string             `json:"deleted_by" bson:"deleted_by"`
	}

	// DashboardData 字段的值
	DashboardData struct {
		XValue string  `json:"x_value" bson:"x_value"`
		XName  string  `json:"x_name" bson:"x_name"`
		XType  string  `json:"x_type" bson:"x_type"`
		GValue string  `json:"g_value" bson:"g_value"`
		GType  string  `json:"g_type" bson:"g_type"`
		YValue float64 `json:"y_value" bson:"y_value"`
		YName  string  `json:"y_name" bson:"y_name"`
	}

	Slider struct {
		Start  float32 `json:"start" bson:"start"`
		End    float32 `json:"end" bson:"end"`
		Height float32 `json:"height" bson:"height"`
	}

	Scrollbar struct {
		Type         string  `json:"type" bson:"type"`
		Width        float32 `json:"width" bson:"width"`
		Height       float32 `json:"height" bson:"height"`
		CategorySize float32 `json:"category_size" bson:"category_size"`
	}
)

// ToProto 转换为proto数据
func (r *Dashboard) ToProto() *dashboard.Dashboard {

	return &dashboard.Dashboard{
		DashboardId:   r.DashboardID,
		DashboardName: r.DashboardName,
		Domain:        r.Domain,
		AppId:         r.AppID,
		ReportId:      r.ReportID,
		DashboardType: r.DashboardType,
		XRange:        r.XRange,
		YRange:        r.YRange,
		TickType:      r.TickType,
		Ticks:         r.Ticks,
		TickCount:     r.TickCount,
		GFieldId:      r.GFieldID,
		XFieldId:      r.XFieldID,
		YFieldId:      r.YFieldID,
		LimitInPlot:   r.LimitInPlot,
		StepType:      r.StepType,
		IsStack:       r.IsStack,
		IsPercent:     r.IsPercent,
		IsGroup:       r.IsGroup,
		Smooth:        r.Smooth,
		MinBarWidth:   r.MinBarWidth,
		MaxBarWidth:   r.MaxBarWidth,
		Radius:        r.Radius,
		InnerRadius:   r.InnerRadius,
		StartAngle:    r.StartAngle,
		EndAngle:      r.EndAngle,
		Slider:        r.Slider.ToProto(),
		Scrollbar:     r.Scrollbar.ToProto(),
		CreatedAt:     r.CreatedAt.String(),
		CreatedBy:     r.CreatedBy,
		UpdatedAt:     r.UpdatedAt.String(),
		UpdatedBy:     r.UpdatedBy,
		DeletedAt:     r.DeletedAt.String(),
		DeletedBy:     r.DeletedBy,
	}
}

// ToProto 转换为proto数据(仪表盘数据)
func (r *Slider) ToProto() *dashboard.Slider {

	return &dashboard.Slider{
		Start:  r.Start,
		End:    r.End,
		Height: r.Height,
	}
}

// ToProto 转换为proto数据(仪表盘数据)
func (r *Scrollbar) ToProto() *dashboard.Scrollbar {

	return &dashboard.Scrollbar{
		Type:         r.Type,
		Width:        r.Width,
		Height:       r.Height,
		CategorySize: r.CategorySize,
	}
}

// ToProto 转换为proto数据(仪表盘数据)
func (r *DashboardData) ToProto() *dashboard.DashboardData {

	dashboard := &dashboard.DashboardData{
		XValue: r.XValue,
		XName:  r.XName,
		XType:  r.XType,
		GValue: r.GValue,
		GType:  r.GType,
		YValue: r.YValue,
		YName:  r.YName,
	}
	if dashboard.XValue == "0001-01-01" {
		dashboard.XValue = ""
	}
	if dashboard.GValue == "0001-01-01" {
		dashboard.GValue = ""
	}
	return dashboard
}

// FindDashboards 获取所属公司所属APP[所属报表]下所有仪表盘情报
func FindDashboards(db, domain, appid, reportid string) (r []Dashboard, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
		"app_id":     appid,
	}

	if reportid != "" {
		query["report_id"] = reportid
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDashboards", fmt.Sprintf("query: [ %s ]", queryJSON))

	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}

	opt := options.Find()
	opt.SetSort(sortItem)

	var result []Dashboard
	dashboards, err := c.Find(ctx, query, opt)
	if err != nil {
		utils.ErrorLog("error FindDashboards", err.Error())
		return nil, err
	}
	defer dashboards.Close(ctx)
	for dashboards.Next(ctx) {
		var dd Dashboard
		err := dashboards.Decode(&dd)
		if err != nil {
			utils.ErrorLog("error FindDashboards", err.Error())
			return nil, err
		}
		result = append(result, dd)
	}

	return result, nil
}

// FindDashboard 通过仪表盘ID获取单个仪表盘情报
func FindDashboard(db, id string) (r Dashboard, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Dashboard
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorLog("error FindDashboard", err.Error())
		return result, err
	}
	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"_id":        objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDashboard", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindDashboard", err.Error())
		return result, err
	}

	return result, nil
}

// AddDashboard 添加单个仪表盘情报
func AddDashboard(db string, r *Dashboard) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r.ID = primitive.NewObjectID()
	r.DashboardID = r.ID.Hex()
	r.DashboardName = GetDashboardNameKey(r.AppID, r.DashboardID)

	queryJSON, _ := json.Marshal(r)
	utils.DebugLog("AddDashboard", fmt.Sprintf("Dashboard: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, r)
	if err != nil {
		utils.ErrorLog("error AddDashboard", err.Error())
		return "", err
	}

	return r.DashboardID, nil
}

// ModifyParams 更新仪表盘数据结构体
type ModifyParams struct {
	DashboardID   string
	DashboardName string
	DashboardType string
	ReportID      string
	ChartType     string
	XRange        []float32
	YRange        []float32
	TickType      string
	Ticks         []int64
	TickCount     int64
	GFieldID      string
	LimitInPlot   bool
	StepType      string
	IsStack       bool
	IsPercent     bool
	IsGroup       bool
	Smooth        bool
	MinBarWidth   float32
	MaxBarWidth   float32
	Radius        float32
	InnerRadius   float32
	StartAngle    float32
	EndAngle      float32
	Slider        Slider
	Scrollbar     Scrollbar
	XFieldID      string
	YFieldID      string
	Writer        string
}

// ModifyDashboard 更新单个仪表盘情报
func ModifyDashboard(db string, r *ModifyParams) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(r.DashboardID)
	if err != nil {
		utils.ErrorLog("error ModifyDashboard", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at":     time.Now(),
		"updated_by":     r.Writer,
		"report_id":      r.ReportID,
		"dashboard_type": r.DashboardType,
		"chart_type":     r.ChartType,
		"x_range":        r.XRange,
		"y_range":        r.YRange,
		"tick_type":      r.TickType,
		"ticks":          r.Ticks,
		"tick_count":     r.TickCount,
		"g_field_id":     r.GFieldID,
		"x_field_id":     r.XFieldID,
		"y_field_id":     r.YFieldID,

		"limit_in_plot": r.LimitInPlot,
		"step_type":     r.StepType,
		"is_stack":      r.IsStack,
		"is_percent":    r.IsPercent,
		"is_group":      r.IsGroup,
		"smooth":        r.Smooth,
		"min_bar_width": r.MinBarWidth,
		"max_bar_width": r.MaxBarWidth,
		"radius":        r.Radius,
		"inner_radius":  r.InnerRadius,
		"start_angle":   r.StartAngle,
		"end_angle":     r.EndAngle,
		"slider":        r.Slider,
		"scrollbar":     r.Scrollbar,
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyDashboard", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyDashboard", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyDashboard", err.Error())
		return err
	}

	return nil
}

// DeleteDashboard 软删除单个仪表盘情报
func DeleteDashboard(db, dashboardid, userid string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(dashboardid)
	if err != nil {
		utils.ErrorLog("error DeleteDashboard", err.Error())
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
	utils.DebugLog("DeleteDashboard", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteDashboard", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteDashboard ", err.Error())
		return err
	}

	return nil
}

// DeleteSelectDashboards 软删除多个仪表盘情报
func DeleteSelectDashboards(db string, dashboardidlist []string, userid string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectDashboards", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectDashboards", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, dashboardid := range dashboardidlist {
			objectID, err := primitive.ObjectIDFromHex(dashboardid)
			if err != nil {
				utils.ErrorLog("error DeleteSelectDashboards", err.Error())
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
			utils.DebugLog("DeleteSelectDashboards", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectDashboards", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error DeleteSelectDashboards", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectDashboards", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectDashboards", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// HardDeleteDashboards 物理删除多个仪表盘情报
func HardDeleteDashboards(db string, dashboardidlist []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteDashboards", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteDashboards", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, dashboardid := range dashboardidlist {
			objectID, err := primitive.ObjectIDFromHex(dashboardid)
			if err != nil {
				utils.ErrorLog("error HardDeleteDashboards", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteDashboards", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteDashboards", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteDashboards", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteDashboards", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// HardDeleteReportDashboards 物理删除多个仪表盘情报
func HardDeleteReportDashboards(db, reportID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"report_id": reportID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("HardDeleteReportDashboards", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err := c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("error HardDeleteReportDashboards", err.Error())
		return err
	}

	return nil
}

// DashboardDataInfo 仪表板数据信息
type DashboardDataInfo struct {
	DashboardInfo Dashboard
	DashboardData []*DashboardData
}

// FindDashboardData 通过仪表盘ID获取仪表盘数据情报
func FindDashboardData(db, id string, owners []string) (dashboard *DashboardDataInfo, err error) {

	// 通过仪表盘ID获取仪表盘设置信息
	dashboardInfo, err := FindDashboard(db, id)
	if err != nil {
		return nil, err
	}

	// 通过仪表盘设置信息中的报表ID获取报表数据情报
	utils.DebugLog("FindDashboardData", fmt.Sprintf("ReportID: [ %s ]", dashboardInfo.ReportID))

	params := ReportParam{
		ReportID:      dashboardInfo.ReportID,
		ConditionType: "and",
		PageIndex:     0,
		PageSize:      0,
		Owners:        owners,
	}
	reportDataInfo, err := FindReportData(db, params)
	if err != nil {
		return nil, err
	}

	var result []*DashboardData
	// 仪表盘数据情报数据编辑
	for _, reportdata := range reportDataInfo.ReportData {

		var item DashboardData

		switch dashboardInfo.XFieldID {
		case "check_type":
			item.XValue = reportdata.CheckType
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		case "check_status":
			item.XValue = reportdata.CheckStatus
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		case "created_at":
			item.XValue = reportdata.CreatedAt.UTC().Format("2006-01-02")
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		case "created_by":
			item.XValue = reportdata.CreatedBy
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		case "updated_at":
			item.XValue = reportdata.UpdatedAt.UTC().Format("2006-01-02")
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		case "updated_by":
			item.XValue = reportdata.UpdatedBy
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		case "checked_at":
			item.XValue = reportdata.CheckedAt.UTC().Format("2006-01-02")
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		case "checked_by":
			item.XValue = reportdata.CheckedBy
			item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
			item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
		default:
			xvalue, xok := reportdata.Items[dashboardInfo.XFieldID]
			if xok {
				item.XValue = GetValue(xvalue)
				item.XType = xvalue.DataType
				item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
			} else {
				item.XValue = "-"
				item.XType = reportDataInfo.Fields[dashboardInfo.XFieldID].DataType
				item.XName = reportDataInfo.Fields[dashboardInfo.XFieldID].AliasName
			}
		}

		if dashboardInfo.YFieldID == "count" {
			item.YValue = float64(reportdata.Count)
			item.YName = "count"
		} else {
			yvalue, yok := reportdata.Items[dashboardInfo.YFieldID]
			if yok {
				itemMap := yvalue
				switch itemMap.Value.(type) {
				case int:
					item.YValue = float64(itemMap.Value.(int))
					item.YName = reportDataInfo.Fields[dashboardInfo.YFieldID].AliasName
				case float64:
					item.YValue = itemMap.Value.(float64)
					item.YName = reportDataInfo.Fields[dashboardInfo.YFieldID].AliasName
				default:
					item.YValue = 0.0
					item.YName = reportDataInfo.Fields[dashboardInfo.YFieldID].AliasName
				}
			} else {
				item.YValue = 0.0
				item.YName = reportDataInfo.Fields[dashboardInfo.YFieldID].AliasName
			}
		}

		if len(dashboardInfo.GFieldID) > 0 {
			switch dashboardInfo.GFieldID {
			case "check_type":
				item.GValue = reportdata.CheckType
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			case "check_status":
				item.GValue = reportdata.CheckStatus
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			case "created_at":
				item.GValue = reportdata.CreatedAt.UTC().Format("2006-01-02")
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			case "created_by":
				item.GValue = reportdata.CreatedBy
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			case "updated_at":
				item.GValue = reportdata.UpdatedAt.UTC().Format("2006-01-02")
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			case "updated_by":
				item.GValue = reportdata.UpdatedBy
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			case "checked_at":
				item.GValue = reportdata.CheckedAt.UTC().Format("2006-01-02")
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			case "checked_by":
				item.GValue = reportdata.CheckedBy
				item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
			default:
				gvalue, xok := reportdata.Items[dashboardInfo.GFieldID]
				if xok {
					item.GValue = GetValue(gvalue)
					item.GType = gvalue.DataType
				} else {
					item.GValue = "-"
					item.GType = reportDataInfo.Fields[dashboardInfo.GFieldID].DataType
				}
			}
		}

		result = append(result, &item)
	}

	// 对该x进行排序
	sort.Sort(DashboardDataList(result))

	return &DashboardDataInfo{
		DashboardInfo: dashboardInfo,
		DashboardData: result,
	}, nil
}

// DashboardDataList 数据排序排序
type DashboardDataList []*DashboardData

//排序规则：按displayOrder排序（由小到大）
func (list DashboardDataList) Len() int {
	return len(list)
}

func (list DashboardDataList) Less(i, j int) bool {
	return list[i].XValue < list[j].XValue
}

func (list DashboardDataList) Swap(i, j int) {
	var temp *DashboardData = list[i]
	list[i] = list[j]
	list[j] = temp
}

// RecoverSelectDashboards 恢复选中仪表盘情报
func RecoverSelectDashboards(db string, dashboardidlist []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(DashboardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectDashboards", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectDashboards", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, dashboardid := range dashboardidlist {
			objectID, err := primitive.ObjectIDFromHex(dashboardid)
			if err != nil {
				utils.ErrorLog("error RecoverSelectDashboards", err.Error())
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
			utils.DebugLog("RecoverSelectDashboards", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectDashboards", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectDashboards", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectDashboards", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectDashboards", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}
