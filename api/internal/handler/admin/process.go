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
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/proto/process"
)

// Process 进程
type Process struct{}

// log出力
const (
	ProcessName        = "Process"
	ActionFindsProcess = "FindsProcess"
	NodeName           = "Node"
	ActionFindsNode    = "FindsNode"
	FormName           = "Form"
	ActionFindForms    = "FindForms"
)

// FindsProcess 获取所有进程
// @Router /Process [get]
func (f *Process) FindsProcess(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindsProcess, loggerx.MsgProcessStarted)

	procesService := process.NewProcessService("workflow", client.DefaultClient)

	var req process.FindsProcessesRequest
	// 从path获取
	req.UserId = c.Param("user_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := procesService.FindsProcesses(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindsProcess, err)
		return
	}

	loggerx.InfoLog(c, ActionFindsProcess, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, ProcessName, ActionFindsProcess)),
		Data:    response.GetProcesses(),
	})
}

// FindNodes 获取所有节点
// @Router /Node [get]
func (f *Process) FindNodes(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindsNode, loggerx.MsgProcessStarted)

	nodeService := node.NewNodeService("workflow", client.DefaultClient)
	var nReq node.NodesRequest
	// 从path获取
	nReq.Database = sessionx.GetUserCustomer(c)

	response, err := nodeService.FindNodes(context.TODO(), &nReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindsNode, err)
		return
	}

	loggerx.InfoLog(c, ActionFindsNode, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, NodeName, ActionFindsProcess)),
		Data:    response.GetNodes(),
	})
}
