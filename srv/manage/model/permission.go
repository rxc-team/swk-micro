package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	PermissionsCollection = "permissions"
)

// Permission 菜单
type Permission struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	PermissionId   string             `json:"permission_id" bson:"permission_id"`
	RoleId         string             `json:"role_id" bson:"role_id"`
	PermissionType string             `json:"permission_type" bson:"permission_type"`
	AppId          string             `json:"app_id" bson:"app_id"`
	ActionType     string             `json:"action_type" bson:"action_type"`
	Actions        []*PAction         `json:"actions" bson:"actions"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy      string             `json:"created_by" bson:"created_by"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy      string             `json:"updated_by" bson:"updated_by"`
}

// PAction 操作权限
type PAction struct {
	ObjectId  string          `json:"object_id" bson:"object_id"`
	Fields    []string        `json:"fields" bson:"fields"`
	ActionMap map[string]bool `json:"action_map" bson:"action_map"`
}

// ToProto 转换为proto数据
func (m *PAction) ToProto() *permission.Action {
	return &permission.Action{
		ObjectId:  m.ObjectId,
		Fields:    m.Fields,
		ActionMap: m.ActionMap,
	}
}

// ToProto 转换为proto数据
func (m *Permission) ToProto() *permission.Permission {

	var actions []*permission.Action
	for _, a := range m.Actions {
		actions = append(actions, &permission.Action{
			ObjectId:  a.ObjectId,
			Fields:    a.Fields,
			ActionMap: a.ActionMap,
		})
	}

	return &permission.Permission{
		PermissionId:   m.PermissionId,
		RoleId:         m.RoleId,
		PermissionType: m.PermissionType,
		AppId:          m.AppId,
		ActionType:     m.ActionType,
		Actions:        actions,
		CreatedAt:      m.CreatedAt.String(),
		CreatedBy:      m.CreatedBy,
		UpdatedAt:      m.UpdatedAt.String(),
		UpdatedBy:      m.UpdatedBy,
	}
}

// FindActions 查找权限的具体内容
func FindPActions(ctx context.Context, db, appID, permissionType, actionType, objectId string, roles []string) (m []*PAction, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(PermissionsCollection)

	query := bson.M{
		"role_id": bson.M{
			"$in": roles,
		},
	}

	if len(permissionType) > 0 {
		query["permission_type"] = permissionType

		if permissionType == "app" {
			query["app_id"] = appID
		}
	}

	if len(actionType) > 0 {
		query["action_type"] = actionType
	}

	if len(objectId) > 0 {
		query["actions.object_id"] = objectId
	}

	var result []Permission

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindPermissions", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindPermissions", err.Error())
		return nil, err
	}

	actMap := make(map[string]*PAction)

	for _, ps := range result {
		for _, act := range ps.Actions {
			if val, exist := actMap[act.ObjectId]; exist {
				fields := utils.New()
				fields.AddAll(val.Fields...)
				fields.AddAll(act.Fields...)

				aMap := make(map[string]bool)
				// 获取原有数据
				for k, v := range val.ActionMap {
					aMap[k] = v
				}

				// merge新数据
				for k, v := range act.ActionMap {
					if _, ok := aMap[k]; ok {
						if v {
							aMap[k] = v
						}
					} else {
						aMap[k] = v
					}
				}

				// 将结果添加到最终map中
				actMap[act.ObjectId] = &PAction{
					ObjectId:  act.ObjectId,
					Fields:    fields.ToList(),
					ActionMap: aMap,
				}
			} else {
				// 将结果添加到最终map中
				actMap[act.ObjectId] = &PAction{
					ObjectId:  act.ObjectId,
					Fields:    act.Fields,
					ActionMap: act.ActionMap,
				}
			}
		}
	}

	// 转换最终结果为数组
	var actions []*PAction
	for _, v := range actMap {
		actions = append(actions, v)
	}

	return actions, nil

}

// FindPermissions 查找权限
func FindPermissions(ctx context.Context, db, role string) (m []*Permission, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(PermissionsCollection)

	query := bson.M{
		"role_id": role,
	}

	var result []*Permission

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindPermissions", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindPermissions", err.Error())
		return nil, err
	}

	return result, nil

}
