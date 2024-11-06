package admin

import (
	"context"
	"encoding/json"
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
	AllowProcessName      = "Allow"
	ActionFindLevelAllows = "FindLevelAllows"
)

// FindLevelAllows 获取所有许可操作
// @Router /allows [get]
func (f *Allow) FindLevelAllows(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindLevelAllows, loggerx.MsgProcessStarted)

	// 获取顾客信息
	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomerRequest
	req.CustomerId = sessionx.GetUserCustomer(c)
	response, err := customerService.FindCustomer(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindCustomer, err)
		return
	}
	// 通过顾客信息获取顾客的授权等级信息
	levelService := level.NewLevelService("manage", client.DefaultClient)

	var lreq level.FindLevelRequest
	lreq.LevelId = response.GetCustomer().GetLevel()
	levelResp, err := levelService.FindLevel(context.TODO(), &lreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindLevelAllows, err)
		return
	}

	if len(levelResp.GetLevel().GetAllows()) == 0 {
		loggerx.InfoLog(c, ActionFindLevelAllows, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, ActionFindLevelAllows)),
			Data:    nil,
		})
		return
	}

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var alreq allow.FindLevelAllowsRequest
	// 从query获取
	alreq.AllowList = levelResp.GetLevel().GetAllows()

	allowResp, err := allowService.FindLevelAllows(context.TODO(), &alreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindLevelAllows, err)
		return
	}

	actionService := action.NewActionService("manage", client.DefaultClient)

	var areq action.FindActionsRequest
	aResp, err := actionService.FindActions(context.TODO(), &areq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindLevelAllows, err)
		return
	}

	allowsData := allowResp.GetAllows()

	actions := aResp.GetActions()

	for _, a := range allowsData {
		for _, x := range a.GetActions() {
			for _, y := range actions {
				if a.AllowType == y.ActionObject && x.ApiKey == y.ActionKey && x.GroupKey == y.ActionGroup {
					actionNmae, err := json.Marshal(y.ActionName)
					if err != nil {
						httpx.GinHTTPError(c, ActionFindLevelAllows, err)
						return
					}
					x.ActionName = string(actionNmae)
				}
			}
		}
	}

	loggerx.InfoLog(c, ActionFindLevelAllows, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, ActionFindLevelAllows)),
		Data:    allowsData,
	})
}
