package webui

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wfx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
)

// Workflow 工作流程
type Workflow struct{}

// log出力
const (
	WorkflowProcessName     = "Workflow"
	ActionFindWorkflows     = "FindWorkflows"
	ActionFindWorkflow      = "FindWorkflow"
	ActionAddWorkflow       = "AddWorkflow"
	ActionModifyWorkflow    = "ModifyWorkflow"
	ActionDeleteWorkflow    = "DeleteWorkflow"
	ActionFindActions       = "FindActions"
	ActionDismiss           = "Dismiss"
	ActionAdmit             = "Admit"
	ActionFindUserWorkflows = "FindUserWorkflows"
)

// FindWorkflow 获取单个工作流程
// @Router /workflows/{workflow_id} [get]
func (t *Workflow) FindWorkflow(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindWorkflow, loggerx.MsgProcessStarted)

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var req workflow.WorkflowRequest
	req.WfId = c.Param("wf_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := workflowService.FindWorkflow(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindWorkflow, err)
		return
	}

	nodeService := node.NewNodeService("workflow", client.DefaultClient)

	var nReq node.NodesRequest
	nReq.WfId = c.Param("wf_id")
	nReq.Database = sessionx.GetUserCustomer(c)

	nResp, err := nodeService.FindNodes(context.TODO(), &nReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindWorkflow, err)
		return
	}

	loggerx.InfoLog(c, ActionFindWorkflow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionFindWorkflow)),
		Data: gin.H{
			"workflow": response.GetWorkflow(),
			"nodes":    nResp.GetNodes(),
		},
	})
}

// FindWorkflows 获取多个工作流程
// @Router /workflows [get]
func (t *Workflow) FindWorkflows(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindWorkflows, loggerx.MsgProcessStarted)

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var req workflow.WorkflowsRequest
	req.AppId = sessionx.GetCurrentApp(c)
	req.IsValid = c.Query("is_valid")
	req.ObjectId = c.Query("datastore")
	req.Action = c.Query("action")
	req.GroupId = c.Query("group")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := workflowService.FindWorkflows(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindWorkflows, err)
		return
	}

	loggerx.InfoLog(c, ActionFindWorkflows, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionFindWorkflows)),
		Data:    response.GetWorkflows(),
	})
}

// Admit 承认
// @Router /workflows/{workflow_id} [delete]
func (t *Workflow) Admit(c *gin.Context) {
	loggerx.InfoLog(c, ActionAdmit, loggerx.MsgProcessStarted)

	type Request struct {
		ExampleID string `json:"ex_id"`
		Comment   string `json:"comment"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAdmit, err)
		return
	}

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	userID := sessionx.GetAuthUserID(c)

	approve := new(wfx.Approve)
	err := approve.Admit(db, req.ExampleID, userID, domain, req.Comment)
	if err != nil {
		httpx.GinHTTPError(c, ActionAdmit, err)
		return
	}
	// loggerx.SuccessLog(c, ActionDeleteWorkflow,  fmt.Sprintf("Workflow [%s] nodes delete success", fReq.GetWfId()))

	loggerx.InfoLog(c, ActionAdmit, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionAdmit)),
		Data:    nil,
	})
}

// Dismiss 拒绝
// @Router /workflows/{workflow_id} [delete]
func (t *Workflow) Dismiss(c *gin.Context) {
	loggerx.InfoLog(c, ActionDismiss, loggerx.MsgProcessStarted)

	type Request struct {
		ExampleID string `json:"ex_id"`
		Comment   string `json:"comment"`
	}

	var req Request

	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionDismiss, err)
		return
	}

	db := sessionx.GetUserCustomer(c)
	userID := sessionx.GetAuthUserID(c)

	approve := new(wfx.Approve)
	err := approve.Dismiss(db, req.ExampleID, userID, req.Comment)
	if err != nil {
		httpx.GinHTTPError(c, ActionDismiss, err)
		return
	}

	loggerx.InfoLog(c, ActionDismiss, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionDismiss)),
		Data:    nil,
	})
}

// FindUserWorkflows 查找当前台账需要流程的操作
// @Router /workflows/{workflow_id} [delete]
func (t *Workflow) FindUserWorkflows(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindUserWorkflows, loggerx.MsgProcessStarted)

	datastore := c.Query("datastore")
	action := c.Query("action")
	groupID := sessionx.GetUserGroup(c)
	db := sessionx.GetUserCustomer(c)
	appId := sessionx.GetCurrentApp(c)

	result := wfx.GetUserWorkflow(db, groupID, appId, datastore, action)

	loggerx.InfoLog(c, ActionFindActions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionFindActions)),
		Data:    result,
	})
}
