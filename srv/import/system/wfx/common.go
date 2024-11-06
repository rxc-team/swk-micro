package wfx

import (
	"context"
	"strings"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

func findApprovers(db, domain string, wkGoupId, nodeGroupId, userGroupId string, groups []*group.Group, assignees []string) []string {
	// 审批者
	var approvers []string

	// 如果当前节点的审批没有组织，那就从当前申请数据的组织到流程的组织直接去查询相关的角色
	if len(nodeGroupId) == 0 {
		// 开始，查询当前组织下有没有对应角色的人
		for _, ap := range assignees {
			as := strings.Split(ap, "_")
			// 如果审批者是 u 打头，则说明是特别审批者，直接添加进入即可
			if as[0] == "u" {
				approvers = append(approvers, as[1])
				continue
			}

			// 如果审批者是 r 打头，则说明是角色审批，目前的设置，角色只有一种。
			if as[0] == "r" {
				users := findRelatedUsers(db, domain, wkGoupId, userGroupId, as[1], groups)
				if len(users) > 0 {
					approvers = append(approvers, users...)
				}
			}

		}

		// 如果当前节点的审批有组织，那就从当前审批的组织直接去查询相关的角色
	} else {
		// 开始，查询当前组织下有没有对应角色的人
		for _, ap := range assignees {
			as := strings.Split(ap, "_")
			// 如果审批者是 u 打头，则说明是特别审批者，直接添加进入即可
			if as[0] == "u" {
				approvers = append(approvers, as[1])
				continue
			}

			// 如果审批者是 r 打头，则说明是角色审批，目前的设置，角色只有一种。
			if as[0] == "r" {
				users := findUsers(db, domain, nodeGroupId, as[1])
				if len(users) > 0 {
					approvers = append(approvers, users...)
				}
			}

		}
	}

	return approvers
}

func findRelatedUsers(db, domain, wkGoupId, groupId, role string, groups []*group.Group) (result []string) {

	currentGroup := findGroup(groupId, groups)

	if currentGroup == nil {
		return
	}

	// 如果当前组织的上级已经是当前的最后结果了
	if wkGoupId == currentGroup.GetParentGroupId() {
		users := findUsers(db, domain, groupId, role)

		if len(users) > 0 {
			result = append(result, users...)
		}
		return
	}

	users := findUsers(db, domain, groupId, role)
	if len(users) > 0 {
		result = append(result, users...)
		return
	}
	result = findRelatedUsers(db, domain, wkGoupId, currentGroup.GetParentGroupId(), role, groups)
	return
}

func findUsers(db, domain, group, role string) []string {
	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest
	// 从query中获取参数
	req.Group = group
	req.Role = role
	// 从共通中获取参数
	req.Domain = domain
	req.Database = db

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("findUsers", err.Error())
		return nil
	}

	var users []string
	for _, u := range response.GetUsers() {
		users = append(users, u.UserId)
	}

	return users
}

func findGroup(group string, groups []*group.Group) *group.Group {
	for _, g := range groups {
		if g.GroupId == group {
			return g
		}
	}
	return nil
}
