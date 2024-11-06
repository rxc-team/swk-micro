package sessionx

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/access"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// log出力
const (
	UserProcessName           = "User"
	ActionGetAuthUserID       = "GetAuthUserID"
	ActionGetCurrentApp       = "GetCurrentApp"
	ActionGetUserAccessKeys   = "GetUserAccessKeys"
	ActionGetRelatedGroups    = "GetRelatedGroups"
	ActionGetUserUpAccessKeys = "GetUserUpAccessKeys"
	ActionGetCurrentApps      = "GetCurrentApps"
	ActionGetUserInfo         = "getUserInfo"
	ActionGetSuperAdmin       = "GetSuperAdmin"
	ActionGetSuperDomain      = "GetSuperDomain"
)

// log出力
const (
	defaultDomain    = "proship.co.jp"
	defaultDomainEnv = "DEFAULT_DOMAIN"
)

// GetAuthUserID 获取当前用户的ID
func GetAuthUserID(c *gin.Context) (id string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return ""
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return ""
	}

	if len(user.UserId) > 0 {
		return user.UserId
	}

	return ""
}

// GetUserDomain 获取当前用户的域
func GetUserDomain(c *gin.Context) (domain string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return ""
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return ""
	}

	if len(user.Domain) > 0 {
		return user.Domain
	}

	return ""
}

// GetUserCustomer 获取当前用户的顾客
func GetUserCustomer(c *gin.Context) (cs string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return ""
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return ""
	}

	if len(user.CustomerId) > 0 {
		return user.CustomerId
	}

	return ""
}

// GetUserEmail 获取当前用户邮箱
func GetUserEmail(c *gin.Context) (cs string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return ""
	}

	user, exist := userInfo.(*user.User)
	if !exist {
		return ""
	}

	if len(user.Email) > 0 {
		return user.Email
	}

	return ""
}

// GetUserGroup 获取当前用户的组
func GetUserGroup(c *gin.Context) (group string) {

	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return ""
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return ""
	}

	return user.GetGroup()
}

// GetCurrentLanguage 获取当前用户的语言
func GetCurrentLanguage(c *gin.Context) (langCd string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return "ja-JP"
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return "ja-JP"
	}

	return user.GetLanguage()
}

// GetCurrentTimezone 获取当前用户的时区
func GetCurrentTimezone(c *gin.Context) (langCd string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return "Asia/Tokyo"
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return "Asia/Tokyo"
	}

	return user.GetTimezone()
}

// GetUserRoles 获取当前用户角色
func GetUserRoles(c *gin.Context) (roles []string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return nil
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return nil
	}

	return user.GetRoles()
}

func GetUserAccessKeys(c *gin.Context, datastoreId, action string) []string {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return nil
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return nil
	}

	owner := getGroupAccessKey(user.CustomerId, user.Group)

	accessService := access.NewAccessService("manage", client.DefaultClient)

	var req access.FindUserAccesssRequest
	req.RoleId = user.Roles
	req.GroupId = user.Group
	req.AppId = user.CurrentApp
	req.Owner = owner
	req.DatastoreId = datastoreId
	req.Action = action
	req.Database = user.CustomerId

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

func GetUserOwner(c *gin.Context) []string {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return nil
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return nil
	}

	owner := getGroupAccessKey(user.CustomerId, user.Group)

	return []string{owner}
}

// GetRelatedGroups 获取传入用户组&其关联下属用户组ID
func GetRelatedGroups(c *gin.Context, inGroupID string, domain string) (outGroupIDs []string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return nil
	}
	user, exist := userInfo.(*user.User)
	if !exist {
		return nil
	}

	// 所有用户组ID
	var groupIDs []string
	// 添加自己的传入用户组ID
	groupIDs = append(groupIDs, inGroupID)

	// 获取子group的ID
	groupService := group.NewGroupService("manage", client.DefaultClient)

	var groupReq group.FindGroupsRequest
	groupReq.Database = user.CustomerId
	groupReq.Domain = domain
	response, err := groupService.FindGroups(context.TODO(), &groupReq)
	if err != nil {
		return []string{}
	}

	groupIDs = append(groupIDs, getNodeGroupIDs(inGroupID, response.Groups)...)

	return groupIDs
}

// 递归获取子节点的用户组ID
func getNodeGroupIDs(pID string, groups []*group.Group) (groupIDs []string) {

	// 根据父节点ID获取子节点信息
	children := getGroupsByPID(pID, groups)

	if len(children) == 0 {
		return groupIDs
	}

	for _, g := range children {
		groupIDs = append(groupIDs, g.GetGroupId())
		groupIDs = append(groupIDs, getNodeGroupIDs(g.GetGroupId(), groups)...)
	}
	return groupIDs
}

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

// GetCurrentApp 获取当前用户的APP
func GetCurrentApp(c *gin.Context) (app string) {

	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return ""
	}
	u, exist := userInfo.(*user.User)
	if !exist {
		return ""
	}

	if u.CustomerId == "system" {
		return "system"
	}

	if u.GetCurrentApp() == "" {
		userService := user.NewUserService("manage", client.DefaultClient)

		update := user.ModifyUserRequest{
			UserId:     u.UserId,
			CurrentApp: u.GetApps()[0],
			Database:   u.CustomerId,
		}

		_, err := userService.ModifyUser(context.TODO(), &update)
		if err != nil {
			loggerx.ErrorLog("GetCurrentApp", err.Error())
			return ""
		}

		return u.GetApps()[0]
	}

	return u.GetCurrentApp()
}

// GetUserApps 获取当前用户的所有app
func GetUserApps(c *gin.Context) (apps []string) {
	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return nil
	}
	u, exist := userInfo.(*user.User)
	if !exist {
		return nil
	}

	return u.GetApps()
}

// GetUserName 获取当前用户的用户名
func GetUserName(c *gin.Context) (name string) {

	//根据上下文获取载荷userInfo
	userInfo, exit := c.Get("userInfo")
	if !exit {
		return ""
	}
	u, exist := userInfo.(*user.User)
	if !exist {
		return ""
	}

	return u.GetUserName()
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

// GetSuperAdmin 获取超级管理员信息
func GetSuperAdmin(c *gin.Context) (id string, name string) {
	userService := user.NewUserService("manage", client.DefaultClient)

	req := user.FindDefaultUserRequest{
		UserType: 2,
	}

	superUser, err := userService.FindDefaultUser(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("GetSuperAdmin", err.Error())
		return "", ""
	}

	return superUser.GetUser().GetUserId(), superUser.GetUser().GetUserName()
}

// GetSuperDomain 获取超级domain信息
func GetSuperDomain() (dom string) {
	domain := os.Getenv(defaultDomainEnv)
	if len(domain) == 0 {
		domain = defaultDomain
	}
	return domain
}
