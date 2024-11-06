package admin

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
)

// Group 组
type Group struct{}

// log出力
const (
	GroupProcessName       = "Group"
	ActionFindGroups       = "FindGroups"
	ActionFindGroup        = "FindGroup"
	ActionModifyGroup      = "ModifyGroup"
	ActionHardDeleteGroups = "HardDeleteGroups"
)

// FindGroups 获取所有组
// @Router /groups [get]
func (u *Group) FindGroups(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindGroups, loggerx.MsgProcessStarted)

	groupService := group.NewGroupService("manage", client.DefaultClient)

	var req group.FindGroupsRequest
	// 从query中获取参数
	req.GroupName = c.Query("group_name")
	// 当前用户的domain
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := groupService.FindGroups(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindGroups, err)
		return
	}

	loggerx.InfoLog(c, ActionFindGroups, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, GroupProcessName, ActionFindGroups)),
		Data:    response.GetGroups(),
	})
}

// FindGroup 获取组
// @Router /groups/{group_id} [get]
func (u *Group) FindGroup(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindGroup, loggerx.MsgProcessStarted)

	groupService := group.NewGroupService("manage", client.DefaultClient)

	var req group.FindGroupRequest
	req.GroupId = c.Param("group_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := groupService.FindGroup(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindGroup, err)
		return
	}

	loggerx.InfoLog(c, ActionFindGroup, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, GroupProcessName, ActionFindGroup)),
		Data:    response.GetGroup(),
	})
}

// AddGroup 添加组
// @Router /groups [post]
func (u *Group) AddGroup(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddGroup, loggerx.MsgProcessStarted)

	groupService := group.NewGroupService("manage", client.DefaultClient)

	var req group.AddGroupRequest
	// 从body中获取参数资源
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddGroup, err)
		return
	}
	// 当前用户的domain
	req.Domain = sessionx.GetUserDomain(c)
	// 当前用户为创建者
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := groupService.AddGroup(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddGroup, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddGroup, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddGroup))

	// 添加用户组对应的语言
	loggerx.InfoLog(c, ActionAddGroup, fmt.Sprintf("Process AddCommonData:%s", loggerx.MsgProcessStarted))
	languageService := language.NewLanguageService("global", client.DefaultClient)
	langParams := language.AddCommonDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		Type:     "groups",
		Key:      response.GetGroupId(),
		Value:    req.GetGroupName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}
	_, err = languageService.AddCommonData(context.TODO(), &langParams)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddGroup, err)
		return
	}
	loggerx.InfoLog(c, ActionAddGroup, fmt.Sprintf("Process AddCommonData:%s", loggerx.MsgProcessEnded))

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, req.Domain)

	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["group_name"] = req.GetGroupName()

	loggerx.ProcessLog(c, ActionAddGroup, msg.L003, params)

	loggerx.InfoLog(c, ActionAddGroup, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, GroupProcessName, ActionAddGroup)),
		Data:    response,
	})
}

// ModifyGroup 更新组
// @Router /groups/{group_id} [put]
func (u *Group) ModifyGroup(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyGroup, loggerx.MsgProcessStarted)

	groupService := group.NewGroupService("manage", client.DefaultClient)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)

	// 查找变更的group的数据
	var fReq group.FindGroupRequest
	fReq.GroupId = c.Param("group_id")
	fReq.Database = sessionx.GetUserCustomer(c)
	fResponse, err := groupService.FindGroup(context.TODO(), &fReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyGroup, err)
		return
	}

	groupInfo := fResponse.GetGroup()

	var req group.ModifyGroupRequest
	// 从body中获取参数资源
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyGroup, err)
		return
	}
	// 从path中获取参数
	req.GroupId = c.Param("group_id")
	// 当前用户为更新者
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := groupService.ModifyGroup(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyGroup, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyGroup, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyGroup))

	// 变更成功后，比较变更的结果，记录日志
	// 比较group名称
	name := langx.GetLangData(db, domain, lang, groupInfo.GetGroupName())
	if name != req.GetGroupName() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["group_name"] = name
		params["group_name1"] = req.GetGroupName()

		loggerx.ProcessLog(c, ActionModifyGroup, msg.L005, params)

		// 变更用户组对应的语言
		loggerx.InfoLog(c, ActionModifyGroup, fmt.Sprintf("Process AddCommonData:%s", loggerx.MsgProcessStarted))
		languageService := language.NewLanguageService("global", client.DefaultClient)
		langParams := language.AddCommonDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			Type:     "groups",
			Key:      req.GetGroupId(),
			Value:    req.GetGroupName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}
		_, err = languageService.AddCommonData(context.TODO(), &langParams)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyGroup, err)
			return
		}
		loggerx.InfoLog(c, ActionModifyGroup, fmt.Sprintf("Process AddCommonData:%s", loggerx.MsgProcessEnded))

		// 通知刷新多语言数据
		langx.RefreshLanguage(req.Writer, domain)
	}

	// 比较親グループ
	if groupInfo.GetParentGroupId() != req.GetParentGroupId() {

		// 获取变更后的亲group的名称
		var fReq group.FindGroupRequest
		fReq.GroupId = req.GetParentGroupId()
		fReq.Database = sessionx.GetUserCustomer(c)
		fResponse, err := groupService.FindGroup(context.TODO(), &fReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyGroup, err)
			return
		}

		pGroupInfo := fResponse.GetGroup()

		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["group_name"] = "{{" + groupInfo.GetGroupName() + "}}"
		params["group_name1"] = "{{" + pGroupInfo.GetGroupName() + "}}"

		loggerx.ProcessLog(c, ActionModifyGroup, msg.L004, params)
	}

	loggerx.InfoLog(c, ActionModifyGroup, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, GroupProcessName, ActionModifyGroup)),
		Data:    response,
	})
}

// HardDeleteGroups 物理删除选中Group
// @Router /phydel/groups [delete]
func (u *Group) HardDeleteGroups(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteGroups, loggerx.MsgProcessStarted)

	var req group.HardDeleteGroupsRequest
	req.GroupIdList = c.QueryArray("group_id_list")
	req.Database = sessionx.GetUserCustomer(c)

	groupService := group.NewGroupService("manage", client.DefaultClient)
	response, err := groupService.HardDeleteGroups(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteGroups, err)
		return
	}
	loggerx.SuccessLog(c, ActionHardDeleteGroups, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionHardDeleteGroups))

	for _, f := range req.GetGroupIdList() {
		// 删除用户组对应的语言
		loggerx.InfoLog(c, ActionHardDeleteGroups, fmt.Sprintf("Process DeleteCommonData:%s", loggerx.MsgProcessStarted))
		languageService := language.NewLanguageService("global", client.DefaultClient)
		langParams := language.DeleteCommonDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			Type:     "groups",
			Key:      f,
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}
		_, err = languageService.DeleteCommonData(context.TODO(), &langParams)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteGroups, err)
			return
		}
		loggerx.InfoLog(c, ActionHardDeleteGroups, fmt.Sprintf("Process DeleteCommonData:%s", loggerx.MsgProcessEnded))
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), sessionx.GetUserDomain(c))

	loggerx.InfoLog(c, ActionHardDeleteGroups, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, GroupProcessName, ActionHardDeleteGroups)),
		Data:    response,
	})
}
