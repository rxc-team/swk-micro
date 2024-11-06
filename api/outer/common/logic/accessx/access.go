package accessx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// Access 权限
type Access struct {
	Database string
	UserID   string
}

// GetAccessKeys 获取用户的许可
func (a *Access) GetAccessKeys() (accessKeys []string) {

	db := a.Database
	userID := a.UserID

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUserRequest
	req.Type = 0
	req.UserId = userID
	req.Database = db
	response, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		return []string{}
	}

	userInfo := response.GetUser()

	// 所有节点的keys
	var keys []string

	// 添加自己的key
	accessKey := a.getGroupAccessKey(db, userInfo.GetGroup())
	keys = append(keys, accessKey)

	return keys
}

// getGroupAccessKey 获取Group的许可
func (a *Access) getGroupAccessKey(db, groupID string) (accessKey string) {

	groupService := group.NewGroupService("manage", client.DefaultClient)

	var groupReq group.FindGroupAccessRequest
	groupReq.GroupId = groupID
	groupReq.Database = db
	response, err := groupService.FindGroupAccess(context.TODO(), &groupReq)
	if err != nil {
		return ""
	}

	return response.GetAccessKey()
}
