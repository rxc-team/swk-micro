package sessionx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/access"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// 根据父节点ID获取子节点信息
func getGroupsByPID(pID string, groups []*group.Group) (child []*group.Group) {
	var children []*group.Group
	for _, g := range groups {
		if g.ParentGroupId == pID {
			children = append(children, g)
		}
	}

	return children
}

func GetAccessKeys(db, userID, datastoreId, action string) []string {

	userInfo, err := getUserInfo(db, userID)
	if err != nil {
		loggerx.ErrorLog("GetUserAccessKeys", err.Error())
		return nil
	}

	owner := getGroupAccessKey(db, userInfo.Group)

	accessService := access.NewAccessService("manage", client.DefaultClient)

	var req access.FindUserAccesssRequest
	req.RoleId = userInfo.Roles
	req.GroupId = userInfo.Group
	req.AppId = userInfo.CurrentApp
	req.Owner = owner
	req.DatastoreId = datastoreId
	req.Action = action
	req.Database = db

	response, err := accessService.FindUserAccess(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("GetUserAccessKeys", err.Error())
		return nil
	}

	if len(response.GetAccessKeys()) == 0 {
		return []string{owner}
	}

	return response.GetAccessKeys()
}

// 获取用户信息
func getUserInfo(db, userID string) (userInfo *user.User, err error) {
	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUserRequest
	req.Type = 0
	req.UserId = userID
	req.Database = db
	response, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getUserInfo", err.Error())
		return nil, err
	}

	return response.GetUser(), nil
}

// getGroupAccessKey 获取Group的许可
func getGroupAccessKey(db, groupID string) (accessKey string) {

	groupService := group.NewGroupService("manage", client.DefaultClient)

	var groupReq group.FindGroupAccessRequest
	groupReq.GroupId = groupID
	groupReq.Database = db
	response, err := groupService.FindGroupAccess(context.TODO(), &groupReq)
	if err != nil {
		loggerx.ErrorLog("getGroupAccessKey", err.Error())
		return ""
	}

	return response.GetAccessKey()
}
