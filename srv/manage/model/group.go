package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	GroupsCollection = "groups"
)

// Group 组织
type Group struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	GroupID       string             `json:"group_id" bson:"group_id"`
	ParentGroupID string             `json:"parent_group_id" bson:"parent_group_id"`
	GroupName     string             `json:"group_name" bson:"group_name"`
	DisplayOrder  int64              `json:"display_order" bson:"display_order"`
	AccessKey     string             `json:"access_key" bson:"access_key"`
	Domain        string             `json:"domain" bson:"domain"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy     string             `json:"created_by" bson:"created_by"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy     string             `json:"updated_by" bson:"updated_by"`
	DeletedAt     time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy     string             `json:"deleted_by" bson:"deleted_by"`
}

// ToProto 转换为proto数据
func (g *Group) ToProto() *group.Group {

	return &group.Group{
		GroupId:       g.GroupID,
		ParentGroupId: g.ParentGroupID,
		GroupName:     g.GroupName,
		DisplayOrder:  g.DisplayOrder,
		AccessKey:     g.AccessKey,
		Domain:        g.Domain,
		CreatedAt:     g.CreatedAt.String(),
		CreatedBy:     g.CreatedBy,
		UpdatedAt:     g.UpdatedAt.String(),
		UpdatedBy:     g.UpdatedBy,
		DeletedAt:     g.DeletedAt.String(),
		DeletedBy:     g.DeletedBy,
	}
}

// FindGroups 查找多个Group记录
func FindGroups(ctx context.Context, db, domain, groupName string) (r []Group, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GroupsCollection)

	var result []Group

	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
	}

	if groupName != "" {
		query["group_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(groupName), Options: "m"}}
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindGroups", fmt.Sprintf("query: [ %s ]", queryJSON))

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindGroups", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var gro Group
		err := cur.Decode(&gro)
		if err != nil {
			utils.ErrorLog("error FindGroups", err.Error())
			return nil, err
		}
		result = append(result, gro)
	}

	return result, nil
}

// FindGroup 查找单个Group记录
func FindGroup(ctx context.Context, db, groupID string) (r Group, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GroupsCollection)

	var result Group
	objectID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		utils.ErrorLog("error FindGroup", err.Error())
		return result, err
	}

	query := bson.M{
		"deleted_by": "",
		"$or": []bson.M{
			{"_id": objectID},
			{"access_key": groupID},
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindGroup", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindGroup", err.Error())
		return result, err
	}

	return result, nil
}

// FindGroupAccess 查找Group的权限记录
func FindGroupAccess(ctx context.Context, db, groupID string) (accessKey string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GroupsCollection)

	var result Group
	objectID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		utils.ErrorLog("error FindGroup", err.Error())
		return "", err
	}

	query := bson.M{
		"deleted_by": "",
		"_id":        objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindGroupAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindGroupAccess", err.Error())
		return "", err
	}

	return result.AccessKey, nil
}

// AddGroup 添加单个Group记录
func AddGroup(ctx context.Context, db string, g *Group) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GroupsCollection)

	if g.ParentGroupID == "root" {
		err := CreateSequence(ctx, db, g.Domain, 1)
		if err != nil {
			return "", err
		}

		g.DisplayOrder = 1
	} else {
		// 查询最大顺
		maxval, err := GetNextSequenceValue(ctx, db, g.Domain)
		if err != nil {
			return "", err
		}
		// 编辑最大顺
		g.DisplayOrder = int64(maxval)
	}

	g.ID = primitive.NewObjectID()
	g.GroupID = g.ID.Hex()
	g.AccessKey = primitive.NewObjectIDFromTimestamp(time.Now()).Hex()
	g.GroupName = "common.groups." + g.GroupID

	queryJSON, _ := json.Marshal(g)
	utils.DebugLog("AddGroup", fmt.Sprintf("Group: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, g)
	if err != nil {
		utils.ErrorLog("error AddGroup", err.Error())
		return "", err
	}
	return g.GroupID, nil
}

// ModifyGroup 更新组的信息
func ModifyGroup(ctx context.Context, db string, g *Group) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GroupsCollection)

	objectID, err := primitive.ObjectIDFromHex(g.GroupID)
	if err != nil {
		utils.ErrorLog("error ModifyGroup", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"group_name": "common.groups." + g.GroupID,
		"updated_at": g.UpdatedAt,
		"updated_by": g.UpdatedBy,
	}

	if len(g.ParentGroupID) > 0 {
		change["parent_group_id"] = g.ParentGroupID
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyGroup", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyGroup", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyGroup", err.Error())
		return err
	}

	return nil
}

// HardDeleteGroups 物理删除选中Group
func HardDeleteGroups(ctx context.Context, db string, groupIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GroupsCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteGroups", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteGroups", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, groupID := range groupIDList {
			objectID, err := primitive.ObjectIDFromHex(groupID)
			if err != nil {
				utils.ErrorLog("error HardDeleteGroups", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteGroups", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteGroups", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteGroups", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteGroups", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}
