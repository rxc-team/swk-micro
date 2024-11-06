package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/manage/proto/access"
	"rxcsoft.cn/pit3/srv/manage/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	AccessCollection = "access"
)

// Access 权限
type Access struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id"`
	AccessID  string              `json:"access_id" bson:"access_id"`
	RoleID    string              `json:"role_id" bson:"role_id"`
	GroupID   string              `json:"group_id" bson:"group_id"`
	Apps      map[string]*AppData `json:"apps" bson:"apps"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	CreatedBy string              `json:"created_by" bson:"created_by"`
	UpdatedAt time.Time           `json:"updated_at" bson:"updated_at"`
	UpdatedBy string              `json:"updated_by" bson:"updated_by"`
	DeletedAt time.Time           `json:"deleted_at" bson:"deleted_at"`
	DeletedBy string              `json:"deleted_by" bson:"deleted_by"`
}

// AppData app数据
type AppData struct {
	DataAccess map[string]*DataAccess `json:"data_access" bson:"data_access"`
}

// DataAccess 数据权限
type DataAccess struct {
	Actions []*DataAction `json:"actions" bson:"actions"`
}

// DataAction 权限操作
type DataAction struct {
	GroupID   string `json:"group_id" bson:"group_id"`
	AccessKey string `json:"access_key" bson:"access_key"`
	CanFind   bool   `json:"can_find" bson:"can_find"`
	CanUpdate bool   `json:"can_update" bson:"can_update"`
	CanDelete bool   `json:"can_delete" bson:"can_delete"`
}

// ToProto 转换为proto数据
func (r *Access) ToProto() *access.Access {
	apps := make(map[string]*access.AppData)
	for k, v := range r.Apps {
		apps[k] = v.ToProto()
	}

	return &access.Access{
		AccessId:  r.AccessID,
		RoleId:    r.RoleID,
		GroupId:   r.GroupID,
		Apps:      apps,
		CreatedAt: r.CreatedAt.String(),
		CreatedBy: r.CreatedBy,
		UpdatedAt: r.UpdatedAt.String(),
		UpdatedBy: r.UpdatedBy,
		DeletedAt: r.DeletedAt.String(),
		DeletedBy: r.DeletedBy,
	}
}

// ToProto 转换为proto数据
func (a *AppData) ToProto() *access.AppData {

	das := make(map[string]*access.DataAccess)
	for k, v := range a.DataAccess {
		das[k] = v.ToProto()
	}

	return &access.AppData{
		DataAccess: das,
	}
}

// ToProto 转换为proto数据
func (a *DataAccess) ToProto() *access.DataAccess {

	var actions []*access.DataAction
	for _, v := range a.Actions {
		actions = append(actions, v.ToProto())
	}

	return &access.DataAccess{
		Actions: actions,
	}
}

// ToProto 转换为proto数据
func (a *DataAction) ToProto() *access.DataAction {
	return &access.DataAction{
		GroupId:   a.GroupID,
		AccessKey: a.AccessKey,
		CanFind:   a.CanFind,
		CanUpdate: a.CanUpdate,
		CanDelete: a.CanDelete,
	}
}

// FindAccess 查找多个权限记录
func FindAccess(ctx context.Context, db, roleId, groupId string) (r []Access, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	query := bson.M{}

	if len(roleId) > 0 {
		query["role_id"] = roleId
	}

	if len(groupId) > 0 {
		query["group_id"] = groupId
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Access
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindAccess", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var access Access
		err := cur.Decode(&access)
		if err != nil {
			utils.ErrorLog("error FindAccess", err.Error())
			return nil, err
		}
		result = append(result, access)
	}

	return result, nil
}

// FindUserAccess 获取用户权限
func FindUserAccess(ctx context.Context, db, groupId, appId, datastoreId, owner, action string, roles []string) (r []string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	pipe := []bson.M{}

	match := bson.M{
		"role_id": bson.M{
			"$in": roles,
		},
		"group_id": groupId,
	}

	pipe = append(pipe, bson.M{
		"$match": match,
	})

	ownerAccess := DataAction{
		AccessKey: owner,
		CanFind:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	project := bson.M{
		"actions": bson.M{
			"$concatArrays": []interface{}{
				bson.M{
					"$cond": bson.M{
						"if": bson.M{"$eq": []interface{}{
							bson.M{
								"$type": "$apps." + appId + ".data_access." + datastoreId + ".actions",
							},
							"missing",
						}},
						"then": []interface{}{},
						"else": "$apps." + appId + ".data_access." + datastoreId + ".actions",
					},
				},
				[]interface{}{
					ownerAccess,
				},
			},
		},
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	unwind := bson.M{
		"path":                       "$actions",
		"preserveNullAndEmptyArrays": true,
	}

	pipe = append(pipe, bson.M{
		"$unwind": unwind,
	})

	// 读权限
	if action == "R" {
		match1 := bson.M{
			"actions.can_find": true,
		}

		pipe = append(pipe, bson.M{
			"$match": match1,
		})
	}

	// 更新权限
	if action == "W" {
		match1 := bson.M{
			"actions.can_update": true,
		}

		pipe = append(pipe, bson.M{
			"$match": match1,
		})
	}

	// 删除权限
	if action == "D" {
		match1 := bson.M{
			"actions.can_delete": true,
		}

		pipe = append(pipe, bson.M{
			"$match": match1,
		})
	}

	group := bson.M{
		"_id": bson.M{},
		"access_keys": bson.M{
			"$addToSet": "$actions.access_key",
		},
	}

	pipe = append(pipe, bson.M{
		"$group": group,
	})

	queryJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

	type Result struct {
		AccessKeys []string `json:"access_keys" bson:"access_keys"`
	}

	var result Result
	cur, err := c.Aggregate(ctx, pipe)
	if err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			return []string{}, nil
		}

		utils.ErrorLog("error FindAccess", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var access Result
		err := cur.Decode(&access)
		if err != nil {
			utils.ErrorLog("error FindAccess", err.Error())
			return nil, err
		}
		result = access
	}

	if len(result.AccessKeys) == 0 {
		return []string{}, nil
	}

	return result.AccessKeys, nil
}

// FindOneAccess 查找单个权限记录
func FindOneAccess(ctx context.Context, db, accessID string) (r Access, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	var result Access
	objectID, err := primitive.ObjectIDFromHex(accessID)
	if err != nil {
		utils.ErrorLog("error FindAccess", err.Error())
		return result, err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindAccess", err.Error())
		return result, err
	}

	return result, nil
}

// AddAccess 添加单个权限记录
func AddAccess(ctx context.Context, db string, r *Access) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	query := bson.M{
		"role_id":  r.RoleID,
		"group_id": r.GroupID,
	}

	var result Access

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {

		if err.Error() == mongo.ErrNoDocuments.Error() {
			r.ID = primitive.NewObjectID()
			r.AccessID = r.ID.Hex()

			queryJSON, _ := json.Marshal(r)
			utils.DebugLog("AddAccess", fmt.Sprintf("Access: [ %s ]", queryJSON))
			// 插入权限表数据
			_, err = c.InsertOne(ctx, r)
			if err != nil {
				utils.ErrorLog("error AddAccess", err.Error())
				return "", err
			}

			return r.AccessID, nil
		}

		utils.ErrorLog("error FindAccess", err.Error())
		return "", err
	}

	objectID, e := primitive.ObjectIDFromHex(result.AccessID)
	if e != nil {
		utils.ErrorLog("error ModifyAccess", e.Error())
		return "", e
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": r.UpdatedBy,
	}

	for k, v := range r.Apps {
		change["apps."+k] = v
	}

	update := bson.M{
		"$set": change,
	}

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyAccess", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateByID(ctx, objectID, update)
	if err != nil {
		utils.ErrorLog("error ModifyAccess", err.Error())
		return result.AccessID, err
	}

	return r.AccessID, nil
}

// AddDataAction 更新权限的信息
func AddDataAction(ctx context.Context, db, accessId, appId, datastoreId, writer string, action *DataAction) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	objectID, err := primitive.ObjectIDFromHex(accessId)
	if err != nil {
		utils.ErrorLog("error ModifyAccess", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	// change := bson.M{
	// 	"updated_at": time.Now(),
	// 	"updated_by": writer,
	// }

	update := bson.M{
		"$addtoset": bson.M{
			"apps." + appId + ".data_access." + datastoreId + ".actions": action,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyAccess", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyAccess", err.Error())
		return err
	}

	return nil
}

// DeleteDataAction 删除单个权限
func DeleteDataAction(ctx context.Context, db, accessId, appId, datastoreId, groupId, writer string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	objectID, err := primitive.ObjectIDFromHex(accessId)
	if err != nil {
		utils.ErrorLog("error DeleteAccess", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{
		"$pull": bson.M{
			"apps." + appId + ".data_access." + datastoreId + ".actions.$.group_id": groupId,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteAccess", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteAccess", err.Error())
		return err
	}

	return nil
}

// DeleteSelectAccess 无效化选择的数据
func DeleteSelectAccess(ctx context.Context, db, userID string, accessList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectAccess", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectAccess", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, accessID := range accessList {
			objectID, err := primitive.ObjectIDFromHex(accessID)
			if err != nil {
				utils.ErrorLog("error DeleteSelectAccess", err.Error())
				return err
			}
			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}
			queryJSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.UpdateByID(sc, objectID, update)
			if err != nil {
				utils.ErrorLog("error DeleteSelectAccess", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectAccess", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectAccess", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// HardDeleteAccess 物理删除选中权限
func HardDeleteAccess(ctx context.Context, db string, accessList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteAccess", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteAccess", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, accessID := range accessList {
			objectID, err := primitive.ObjectIDFromHex(accessID)
			if err != nil {
				utils.ErrorLog("error HardDeleteAccess", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteAccess", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteAccess", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteAccess", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// RecoverSelectAccess 恢复选中权限
func RecoverSelectAccess(ctx context.Context, db string, accessList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(AccessCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectAccess", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectAccess", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, accessID := range accessList {
			objectID, err := primitive.ObjectIDFromHex(accessID)
			if err != nil {
				utils.ErrorLog("error RecoverSelectAccess", err.Error())
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
			utils.DebugLog("RecoverSelectAccess", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectAccess", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectAccess", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectAccess", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectAccess", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}
