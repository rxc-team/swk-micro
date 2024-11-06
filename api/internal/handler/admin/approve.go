package admin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/approve"
)

// Approve Approve
type Approve struct{}

// log出力
const (
	ApproveProcessName     = "Approve"
	ActionFindApproveCount = "FindApproveCount"
)

// FindApproveCount 获取台账中的所有临时数据
// @Router /approves [get]
func (i *Approve) FindApproveCount(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindApproveCount, loggerx.MsgProcessStarted)

	db := sessionx.GetUserCustomer(c)
	wfID := c.Query("wf_id")
	status, _ := strconv.ParseInt(c.Query("status"), 0, 64)

	var req approve.CountRequest
	req.WfId = wfID
	req.Status = status
	req.Database = db

	tplService := approve.NewApproveService("database", client.DefaultClient)

	response, err := tplService.FindCount(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindApproveCount, err)
		return
	}

	loggerx.InfoLog(c, ActionFindApproveCount, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ApproveProcessName, ActionFindApproveCount)),
		Data: gin.H{
			"total": response.GetTotal(),
		},
	})

}
