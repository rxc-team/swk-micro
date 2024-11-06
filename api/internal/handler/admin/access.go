package admin

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/access"
)

// Access Access
type Access struct{}

// log出力
const (
	AccessProcessName         = "Access"
	ActionFindAccess          = "FindAccess"
	ActionFindOneAccess       = "FindOneAccess"
	ActionAddAccess           = "AddAccess"
	ActionAddDataAction       = "AddDataAction"
	ActionDeleteDataAction    = "DeleteDataAction"
	ActionDeleteSelectAccess  = "DeleteSelectAccess"
	ActionHardDeleteAccess    = "HardDeleteAccess"
	ActionRecoverSelectAccess = "RecoverSelectAccess"
)

// FindAccess 获取所有Access
// @Router /access [get]
func (u *Access) FindAccess(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindAccess, loggerx.MsgProcessStarted)

	accessService := access.NewAccessService("manage", client.DefaultClient)

	var req access.FindAccessRequest
	// 从query中获取参数
	req.RoleId = c.Query("role_id")
	req.GroupId = c.Query("group_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := accessService.FindAccess(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindAccess, err)
		return
	}

	loggerx.InfoLog(c, ActionFindAccess, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    response.GetAccessList(),
	})
}

// FindAccess 获取Access
// @Router /access/{access_id} [get]
func (u *Access) FindOneAccess(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindAccess, loggerx.MsgProcessStarted)

	accessService := access.NewAccessService("manage", client.DefaultClient)

	var req access.FindOneAccessRequest
	req.AccessId = c.Param("access_id")
	req.Database = sessionx.GetUserCustomer(c)
	response, err := accessService.FindOneAccess(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindAccess, err)
		return
	}

	loggerx.InfoLog(c, ActionFindAccess, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionFindAccess)),
		Data:    response.GetAccess(),
	})
}

// AddAccess 添加Access
// @Router /access [post]
func (u *Access) AddAccess(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddAccess, loggerx.MsgProcessStarted)

	accessService := access.NewAccessService("manage", client.DefaultClient)

	var req access.AddAccessRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddAccess, err)
		return
	}

	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := accessService.AddAccess(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddAccess, err)
		return
	}

	loggerx.SuccessLog(c, ActionAddAccess, fmt.Sprintf("Access[%s] create Success", response.GetAccessId()))

	loggerx.InfoLog(c, ActionAddAccess, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionAddAccess)),
		Data:    response,
	})
}

// DeleteSelectAccess 删除选中Access
// @Router /access [delete]
func (u *Access) DeleteSelectAccess(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSelectAccess, loggerx.MsgProcessStarted)
	accessService := access.NewAccessService("manage", client.DefaultClient)

	var req access.DeleteSelectAccessRequest
	req.AccessList = c.QueryArray("access_list")
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := accessService.DeleteSelectAccess(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectAccess, err)
		return
	}

	loggerx.SuccessLog(c, ActionDeleteSelectAccess, fmt.Sprintf("Access[%s] delete Success", req.GetAccessList()))

	loggerx.InfoLog(c, ActionDeleteSelectAccess, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionDeleteSelectAccess)),
		Data:    response,
	})
}

// HardDeleteAccess 物理删除选中Access
// @Router /phydel/access [delete]
func (u *Access) HardDeleteAccess(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteAccess, loggerx.MsgProcessStarted)
	accessService := access.NewAccessService("manage", client.DefaultClient)
	var req access.HardDeleteAccessRequest
	req.AccessList = c.QueryArray("access_list")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := accessService.HardDeleteAccess(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteAccess, err)
		return
	}

	loggerx.SuccessLog(c, ActionHardDeleteAccess, fmt.Sprintf("Access[%s] physically delete Success", req.GetAccessList()))

	loggerx.InfoLog(c, ActionHardDeleteAccess, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, AccessProcessName, ActionHardDeleteAccess)),
		Data:    response,
	})
}
