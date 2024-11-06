package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
)

// Group 组
type Group struct{}

// log出力
const (
	GroupProcessName = "Group"
	ActionFindGroups = "FindGroups"
	ActionFindGroup  = "FindGroup"
)

// FindGroups 获取所有组
// @Summary 获取所有组
// @description 调用srv中的group服务，获取所有的组
// @Tags Group
// @Accept json
// @Security JWT
// @Produce  json
// @Param group_name query string false "组名"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
// @Summary 获取组
// @description 调用srv中的group服务，通过ID获取组相关信息
// @Tags Group
// @Accept json
// @Security JWT
// @Produce  json
// @Param group_id path string true "组ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
