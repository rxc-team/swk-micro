package webui

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
)

// Group 组
type Group struct{}

// log出力
const (
	GroupProcessName = "Group"
	ActionFindGroups = "FindGroups"
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
