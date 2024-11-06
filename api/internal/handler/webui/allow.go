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
	"rxcsoft.cn/pit3/srv/manage/proto/action"
	"rxcsoft.cn/pit3/srv/manage/proto/allow"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/level"
)

// Allow 许可操作
type Allow struct{}

// log出力使用
const (
	AllowProcessName = "Allow"
	AllowCheckAllow  = "CheckAllow"
)

// FindLevelAllows 获取所有许可操作
// @Router /allows [get]
func (f *Allow) CheckAllow(c *gin.Context) {
	loggerx.InfoLog(c, AllowCheckAllow, loggerx.MsgProcessStarted)

	allowType := c.Query("allow_type")

	// 获取顾客信息
	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomerRequest
	req.CustomerId = sessionx.GetUserCustomer(c)
	response, err := customerService.FindCustomer(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, AllowCheckAllow, err)
		return
	}
	// 通过顾客信息获取顾客的授权等级信息
	levelService := level.NewLevelService("manage", client.DefaultClient)

	var lreq level.FindLevelRequest
	lreq.LevelId = response.GetCustomer().GetLevel()
	levelResp, err := levelService.FindLevel(context.TODO(), &lreq)
	if err != nil {
		httpx.GinHTTPError(c, AllowCheckAllow, err)
		return
	}

	if len(levelResp.GetLevel().GetAllows()) == 0 {
		loggerx.InfoLog(c, AllowCheckAllow, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowCheckAllow)),
			Data:    false,
		})
		return
	}

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var alreq allow.FindLevelAllowsRequest
	// 从query获取
	alreq.AllowList = levelResp.GetLevel().GetAllows()

	allowResp, err := allowService.FindLevelAllows(context.TODO(), &alreq)
	if err != nil {
		httpx.GinHTTPError(c, AllowCheckAllow, err)
		return
	}

	actionService := action.NewActionService("manage", client.DefaultClient)

	var areq action.FindActionsRequest
	aResp, err := actionService.FindActions(context.TODO(), &areq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindActions, err)
		return
	}

	allowsData := allowResp.GetAllows()
	actions := aResp.GetActions()
	result := false

	for _, a := range allowsData {
		if a.AllowType == allowType {
			for _, x := range a.GetActions() {
				for _, y := range actions {
					if a.AllowType == y.ActionObject && x.ApiKey == y.ActionKey && x.ApiKey == "read" && x.GroupKey == y.ActionGroup {
						result = true
					}
				}
			}
		}

	}

	loggerx.InfoLog(c, AllowCheckAllow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowCheckAllow)),
		Data:    result,
	})
}
