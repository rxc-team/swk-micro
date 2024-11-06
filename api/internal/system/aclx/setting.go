package aclx

import (
	"github.com/micro/go-micro/v2/broker"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/utils/mq"
)

func SetRoleCasbin(roleId string, permissions []*role.Permission) {
	acl := GetCasbin()
	acl.DeletePermissionsForUser(roleId)
	// 设置权限
	if len(permissions) > 0 {
		addPermission(permissions, roleId)
	}

	// 通知重新加载权限
	bk := mq.NewBroker()

	err := bk.Publish("acl.refresh", &broker.Message{})
	if err != nil {
		loggerx.ErrorLog("setRoleCasbin", err.Error())
		return
	}
}

func ClearRoleCasbin(roleId string) {
	// 添加任务
	acl := GetCasbin()
	acl.DeletePermissionsForUser(roleId)

	// 通知重新加载权限
	bk := mq.NewBroker()

	err := bk.Publish("acl.refresh", &broker.Message{})
	if err != nil {
		loggerx.ErrorLog("clearRoleCasbin", err.Error())
		return
	}
}

func SetUserCasbin(userId string, roles []string, apps []string) {
	// 添加任务
	acl := GetCasbin()
	// 先删除该用户的所有角色关系
	acl.DeleteRolesForUser(userId)
	// 循环添加当前用户和角色以及app之间的关联
	for _, roleID := range roles {
		for _, appID := range apps {
			// 添加当前用户和角色以及app之间的关联
			acl.AddGroupingPolicy(userId, roleID, appID)
		}
	}

	bk := mq.NewBroker()

	err := bk.Publish("acl.refresh", &broker.Message{})
	if err != nil {
		loggerx.ErrorLog("setUserCasbin", err.Error())
		return
	}
}

func ClearUserCasbin(UserId string) {
	// 添加任务
	acl := GetCasbin()
	// 先删除该用户的所有角色关系
	acl.DeleteRolesForUser(UserId)

	bk := mq.NewBroker()

	err := bk.Publish("acl.refresh", &broker.Message{})
	if err != nil {
		loggerx.ErrorLog("clearUserCasbin", err.Error())
		return
	}
}

// 设置权限
func addPermission(ps []*role.Permission, id string) {
	acl := GetCasbin()
	for _, p := range ps {
		if p.ActionType == "datastore" {
			for _, act := range p.Actions {
				for k, v := range act.ActionMap {
					if v {
						methods := datastoreActionMap[k]
						if len(methods) > 0 {
							for _, m := range methods {
								acl.AddPermissionForUser(id, m.Path, m.Method)
							}
						}
					}

				}
			}
		}
		if p.ActionType == "report" {
			for _, act := range p.Actions {
				for k, v := range act.ActionMap {
					if v {
						methods := reportActionMap[k]
						if len(methods) > 0 {
							for _, m := range methods {
								acl.AddPermissionForUser(id, m.Path, m.Method)
							}
						}
					}

				}
			}
		}
		if p.ActionType == "folder" {
			for _, act := range p.Actions {
				for k, v := range act.ActionMap {
					if v {
						methods := docActionMap[k]
						if len(methods) > 0 {
							for _, m := range methods {
								acl.AddPermissionForUser(id, m.Path, m.Method)
							}
						}
					}

				}
			}
		}
	}
}
