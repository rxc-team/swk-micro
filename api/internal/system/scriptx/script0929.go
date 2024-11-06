package scriptx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/aclx"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// 自动更新用户的权限关系
type Script0929 struct{}

func (s *Script0929) Run() error {
	go updateUserAndRoleRelation()
	return nil
}

func updateUserAndRoleRelation() {
	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomersRequest
	response, err := customerService.FindCustomers(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("Script0929", err.Error())
		return
	}

	for _, cs := range response.GetCustomers() {

		userService := user.NewUserService("manage", client.DefaultClient)

		var ureq user.FindUsersRequest
		ureq.Domain = cs.GetDomain()
		ureq.Database = cs.GetCustomerId()

		uResp, err := userService.FindUsers(context.TODO(), &ureq)
		if err != nil {
			loggerx.ErrorLog("Script0929", err.Error())
			return
		}

		// 更新用户和角色的关系
		acl := aclx.GetCasbin()
		for _, u := range uResp.GetUsers() {
			// 先删除该用户的所有角色关系
			acl.DeleteRolesForUser(u.UserId)
			// 循环添加当前用户和角色以及app之间的关联
			for _, roleID := range u.GetRoles() {
				for _, appID := range u.GetApps() {
					// 添加当前用户和角色以及app之间的关联
					acl.AddGroupingPolicy(u.GetUserId(), roleID, appID)
				}
			}
		}

		roleService := role.NewRoleService("manage", client.DefaultClient)

		var rreq role.FindRolesRequest
		// 从共通中获取参数
		rreq.Domain = cs.GetDomain()
		rreq.Database = cs.GetCustomerId()

		rRes, err := roleService.FindRoles(context.TODO(), &rreq)
		if err != nil {
			loggerx.ErrorLog("Script0929", err.Error())
			return
		}

		for _, r := range rRes.GetRoles() {
			pmService := permission.NewPermissionService("manage", client.DefaultClient)
			var req permission.FindPermissionsRequest
			req.RoleId = r.RoleId
			req.Database = cs.GetCustomerId()
			response, err := pmService.FindPermissions(context.TODO(), &req)
			if err != nil {
				loggerx.ErrorLog("Script0929", err.Error())
				return
			}
			acl.DeletePermissionsForUser(r.GetRoleId())
			// 设置权限
			addPermission(response.GetPermission(), r.GetRoleId())
		}
	}
}

// 设置权限
func addPermission(ps []*permission.Permission, id string) {
	acl := aclx.GetCasbin()
	acl.DeletePermissionsForUser(id)
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
