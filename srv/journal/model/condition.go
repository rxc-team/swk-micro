package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"rxcsoft.cn/pit3/srv/journal/proto/condition"
	"rxcsoft.cn/pit3/srv/journal/utils"
	database "rxcsoft.cn/utils/mongo"
)

type (
	// Journal 分录
	Condition struct {
		ID            primitive.ObjectID `json:"id" bson:"_id"`
		AppID         string             `json:"app_id" bson:"app_id"`
		ConditionID   string             `json:"condition_id" bson:"condition_id"`
		ConditionName string             `json:"condition_name" bson:"condition_name"`
		Groups        []*Group           `json:"groups" bson:"groups"`
		ThenValue     string             `json:"then_value" bson:"then_value"`
		ElseValue     string             `json:"else_value" bson:"else_value"`
		CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy     string             `json:"created_by" bson:"created_by"`
		UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy     string             `json:"updated_by" bson:"updated_by"`
	}
	// Journal Group
	Group struct {
		GroupID    string `json:"group_id" bson:"group_id"`
		GroupName  string `json:"group_name" bson:"group_name"`
		Type       string `json:"type" bson:"type"`
		SwitchType string `json:"switch_type" bson:"switch_type"`
		Cons       []*Con `json:"cons" bson:"cons"`
	}
	// Journal Con
	Con struct {
		ConID       string `json:"con_id" bson:"con_id"`
		ConName     string `json:"con_name" bson:"con_name"`
		ConField    string `json:"con_field" bson:"con_field"`
		ConOperator string `json:"con_operator" bson:"con_operator"`
		ConValue    string `json:"con_value" bson:"con_value"`
	}
)

// ToProto 转换为proto数据
func (w *Condition) ToProto() *condition.Condition {

	var groups []*condition.Group

	for _, pt := range w.Groups {
		groups = append(groups, pt.ToProto())
	}

	return &condition.Condition{
		ConditionId:   w.ConditionID,
		ConditionName: w.ConditionName,
		Groups:        groups,
		ThenValue:     w.ThenValue,
		ElseValue:     w.ElseValue,
		AppId:         w.AppID,
		CreatedAt:     w.CreatedAt.String(),
		CreatedBy:     w.CreatedBy,
		UpdatedAt:     w.UpdatedAt.String(),
		UpdatedBy:     w.UpdatedBy,
	}
}

// ToProto 转换为proto数据
func (w *Group) ToProto() *condition.Group {

	var cons []*condition.Con

	for _, con := range w.Cons {
		cons = append(cons, con.ToProto())
	}

	return &condition.Group{
		GroupId:    w.GroupID,
		GroupName:  w.GroupName,
		Type:       w.Type,
		SwitchType: w.SwitchType,
		Cons:       cons,
	}
}

// ToProto 转换为proto数据
func (w *Con) ToProto() *condition.Con {
	return &condition.Con{
		ConId:       w.ConID,
		ConName:     w.ConName,
		ConField:    w.ConField,
		ConOperator: w.ConOperator,
		ConValue:    w.ConValue,
	}
}

func ToProto(conditions []Condition) *condition.FindConditionsResponse {
	protoResponse := &condition.FindConditionsResponse{}

	// 遍历 FieldConf 切片，并将每个 FieldConf 转换为 Protobuf 格式
	for _, fc := range conditions {
		protoResponse.Conditions = append(protoResponse.Conditions, fc.ToProto())
	}

	return protoResponse
}

// 添加分录下载设定
func AddCondition(db string, appID string, condition Condition) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("conditions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建一个新的 ObjectID
	condition.ID = primitive.NewObjectID()

	if _, err = c.InsertOne(ctx, condition); err != nil {
		utils.ErrorLog("AddCondition", err.Error())
		return err
	}

	return nil
}

// 查询所有分录下载设定
func FindConditions(db string, appID string) (condition []Condition, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("conditions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appID,
	}

	// 定义一个切片用于存储查询结果
	var results []Condition

	// 使用 Find 查询所有符合条件的文档
	cursor, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("FindConditions", err.Error())
		return nil, err
	}

	// 确保在函数返回之前关闭游标
	defer cursor.Close(ctx)

	// 将游标中的所有数据解码到 results 切片中
	if err := cursor.All(ctx, &results); err != nil {
		utils.ErrorLog("FindConditions", err.Error())
		return nil, err
	}

	// 如果没有找到任何数据，返回空切片
	if len(results) == 0 {
		return nil, nil
	}

	return results, nil
}
