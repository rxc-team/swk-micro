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
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	RolesCollection = "roles"
)

// Role 角色
type Role struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	RoleID      string             `json:"role_id" bson:"role_id"`
	RoleName    string             `json:"role_name" bson:"role_name"`
	Description string             `json:"description" bson:"description"`
	Domain      string             `json:"domain" bson:"domain"`
	IPSegments  []IPSegment        `json:"ip_segments" bson:"ip_segments"`
	Menus       []string           `json:"menus" bson:"menus"`
	RoleType    int32              `json:"role_type" bson:"role_type"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy   string             `json:"created_by" bson:"created_by"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	DeletedAt   time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy   string             `json:"deleted_by" bson:"deleted_by"`
}

// IPSegment IP白名单
type IPSegment struct {
	Start string `json:"start" bson:"start"`
	End   string `json:"end" bson:"end"`
}

// DispalyDatastore 角色控制显示的台账
type DispalyDatastore struct {
	DatastoreID string          `json:"datastore_id" bson:"datastore_id"`
	Fields      []string        `json:"fields" bson:"fields"`
	Actions     map[string]bool `json:"actions" bson:"actions"`
}

// DispalyFolder 角色控制显示的文档
type DispalyFolder struct {
	FolderID string `json:"folder_id" bson:"folder_id"`
	Read     bool   `json:"read" bson:"read"`
	Write    bool   `json:"write" bson:"write"`
	Delete   bool   `json:"delete" bson:"delete"`
}

// ToProto 转换为proto数据
func (r *Role) ToProto() *role.Role {
	var ips []*role.IPSegment
	for _, ip := range r.IPSegments {
		ips = append(ips, ip.ToProto())
	}

	return &role.Role{
		RoleId:      r.RoleID,
		RoleName:    r.RoleName,
		Description: r.Description,
		Domain:      r.Domain,
		IpSegments:  ips,
		Menus:       r.Menus,
		RoleType:    r.RoleType,
		CreatedAt:   r.CreatedAt.String(),
		CreatedBy:   r.CreatedBy,
		UpdatedAt:   r.UpdatedAt.String(),
		UpdatedBy:   r.UpdatedBy,
		DeletedAt:   r.DeletedAt.String(),
		DeletedBy:   r.DeletedBy,
	}
}

// ToProto 转换为proto数据
func (i *IPSegment) ToProto() *role.IPSegment {
	return &role.IPSegment{
		Start: i.Start,
		End:   i.End,
	}
}

// FindRoles 查找多个角色记录
func FindRoles(ctx context.Context, db, roleID, roleType, roleName, description, domain, invalidatedIn string) (r []Role, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RolesCollection)

	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
	}

	// 是否包含无效数据
	if invalidatedIn != "" {
		delete(query, "deleted_by")
	}

	// 角色ID不为空,做条件
	if roleID != "" {
		query["role_id"] = roleID
	}
	// 角色类型不为空,做条件
	if roleType != "" {
		switch roleType {
		case "0":
			query["role_type"] = 0
		case "1":
			query["role_type"] = 1
		case "2":
			query["role_type"] = 2
		}
	}

	// 角色名称不为空,做条件
	if roleName != "" {
		query["role_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(roleName), Options: "m"}}

	}

	// 角色描述不为空,做条件
	if description != "" {
		query["description"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(description), Options: "m"}}

	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindRoles", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Role
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindRoles", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var role Role
		err := cur.Decode(&role)
		if err != nil {
			utils.ErrorLog("error FindRoles", err.Error())
			return nil, err
		}
		result = append(result, role)
	}

	return result, nil
}

// FindRole 查找单个角色记录
func FindRole(ctx context.Context, db, roleID string) (r Role, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RolesCollection)

	var result Role
	objectID, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		utils.ErrorLog("error FindRole", err.Error())
		return result, err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindRole", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindRole", err.Error())
		return result, err
	}

	return result, nil
}

// AddRole 添加单个角色记录
func AddRole(ctx context.Context, db string, r *Role, p []*Permission) (id string, err error) {
	client := database.New()
	rc := client.Database(database.GetDBName(db)).Collection(RolesCollection)
	pc := client.Database(database.GetDBName(db)).Collection(PermissionsCollection)

	client.Database(database.GetDBName(db)).CreateCollection(ctx, RolesCollection)
	client.Database(database.GetDBName(db)).CreateCollection(ctx, PermissionsCollection)

	r.ID = primitive.NewObjectID()
	r.RoleID = r.ID.Hex()

	if len(r.Menus) == 0 {
		r.Menus = make([]string, 0)
	}

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error AddRole", err.Error())
		return "", err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error AddRole", err.Error())
		return "", err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		queryJSON, _ := json.Marshal(r)
		utils.DebugLog("AddRole", fmt.Sprintf("Role: [ %s ]", queryJSON))
		// 插入角色表数据
		_, err = rc.InsertOne(sc, r)
		if err != nil {
			utils.ErrorLog("error AddRole", err.Error())
			return err
		}

		var pList []interface{}

		for _, ps := range p {
			ps.ID = primitive.NewObjectID()
			ps.PermissionId = ps.ID.Hex()
			ps.RoleId = r.RoleID
			pList = append(pList, ps)
		}

		if len(pList) > 0 {
			// 插入权限表数据
			_, err := pc.InsertMany(sc, pList)
			if err != nil {
				utils.ErrorLog("error AddRole", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error AddRole", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error AddRole", err.Error())
		return "", err
	}
	session.EndSession(ctx)

	return r.RoleID, nil
}

// ModifyRole 更新角色的信息
func ModifyRole(ctx context.Context, db string, r *Role, p []*Permission) (err error) {
	client := database.New()
	rc := client.Database(database.GetDBName(db)).Collection(RolesCollection)
	pc := client.Database(database.GetDBName(db)).Collection(PermissionsCollection)

	client.Database(database.GetDBName(db)).CreateCollection(ctx, PermissionsCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error ModifyRole", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error ModifyRole", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		objectID, err := primitive.ObjectIDFromHex(r.RoleID)
		if err != nil {
			utils.ErrorLog("error ModifyRole", err.Error())
			return err
		}
		query := bson.M{
			"_id": objectID,
		}

		change := bson.M{
			"updated_at": r.UpdatedAt,
			"updated_by": r.UpdatedBy,
		}

		// 白名单的场合
		if len(r.IPSegments) > 0 {
			change["ip_segments"] = r.IPSegments
		} else {
			change["ip_segments"] = make([]IPSegment, 0)
		}
		// 菜单的场合
		if len(r.Menus) > 0 {
			change["menus"] = r.Menus
		} else {
			change["menus"] = make([]string, 0)
		}

		// 角色名称不为空的场合z
		if r.RoleName != "" {
			change["role_name"] = r.RoleName
		}

		// 角色描述不为空的场合
		if r.Description != "" {
			change["description"] = r.Description
		}

		// 角色类型不为空的场合
		if r.RoleType != 0 {
			change["role_type"] = r.Description
		}

		update := bson.D{
			{Key: "$set", Value: change},
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("ModifyRole", fmt.Sprintf("query: [ %s ]", queryJSON))

		updateSON, _ := json.Marshal(update)
		utils.DebugLog("ModifyRole", fmt.Sprintf("update: [ %s ]", updateSON))

		_, err = rc.UpdateOne(sc, query, update)
		if err != nil {
			utils.ErrorLog("error ModifyRole", err.Error())
			return err
		}

		if len(p) > 0 {
			// 插入权限表数据
			var pList []interface{}

			for _, ps := range p {
				ps.ID = primitive.NewObjectID()
				ps.PermissionId = ps.ID.Hex()

				// 删除原有的角色权限
				del := bson.M{
					"role_id":         r.RoleID,
					"permission_type": ps.PermissionType,
					"action_type":     ps.ActionType,
				}

				if ps.PermissionType == "app" {
					del["app_id"] = ps.AppId
				}

				_, err := pc.DeleteMany(sc, del)
				if err != nil {
					utils.ErrorLog("error ModifyRole", err.Error())
					return err
				}

				pList = append(pList, ps)
			}
			_, err = pc.InsertMany(sc, pList)
			if err != nil {
				utils.ErrorLog("error ModifyRole", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error ModifyRole", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error ModifyRole", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// DeleteRole 删除单个角色
func DeleteRole(ctx context.Context, db, roleID, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RolesCollection)

	objectID, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		utils.ErrorLog("error DeleteRole", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteRole", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteRole", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteRole", err.Error())
		return err
	}

	return nil
}

// DeleteSelectRoles 删除选中角色
func DeleteSelectRoles(ctx context.Context, db string, roleIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RolesCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectRoles", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectRoles", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, roleID := range roleIDList {
			objectID, err := primitive.ObjectIDFromHex(roleID)
			if err != nil {
				utils.ErrorLog("error DeleteSelectRoles", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSelectRoles", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectRoles", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error DeleteSelectRoles", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectRoles", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectRoles", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// HardDeleteRoles 物理删除选中角色
func HardDeleteRoles(ctx context.Context, db string, roleIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RolesCollection)
	pc := client.Database(database.GetDBName(db)).Collection(PermissionsCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteRoles", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteRoles", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, roleID := range roleIDList {
			objectID, err := primitive.ObjectIDFromHex(roleID)
			if err != nil {
				utils.ErrorLog("error HardDeleteRoles", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteRoles", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteRoles", err.Error())
				return err
			}

			// 删除原有的角色权限
			del := bson.M{
				"role_id": roleID,
			}

			_, err = pc.DeleteMany(sc, del)
			if err != nil {
				utils.ErrorLog("error ModifyRole", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteRoles", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteRoles", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// RecoverSelectRoles 恢复选中角色
func RecoverSelectRoles(ctx context.Context, db string, roleIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RolesCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectRoles", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectRoles", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, roleID := range roleIDList {
			objectID, err := primitive.ObjectIDFromHex(roleID)
			if err != nil {
				utils.ErrorLog("error RecoverSelectRoles", err.Error())
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
			utils.DebugLog("RecoverSelectRoles", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectRoles", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectRoles", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectRoles", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectRoles", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// WhitelistClear 清空白名单
func WhitelistClear(ctx context.Context, userID string, db string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(RolesCollection)

	query := bson.M{
		"role_type": 1,
	}

	update := bson.M{"$set": bson.M{
		"ip_segments": nil,
		"updated_at":  time.Now(),
		"updated_by":  userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("WhitelistClear", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("WhitelistClear", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err := c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error WhitelistClear", err.Error())
		return err
	}

	return nil
}
