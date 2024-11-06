package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/common/logic/langx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/api/outer/system/wfx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/language"
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
// @Summary 获取单个工作流程
// @description 调用srv中的workflow服务，获取单个工作流程
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param workflow_id path string true "工作流程ID"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
// @Summary 获取多个工作流程
// @description 调用srv中的workflow服务，获取多个工作流程
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param title query string false "标题"
// @Param type query string false "类型"
// @Param tag query string false "标签"
// @Param is_dev query string false "dev区分"
// @Param lang_cd query string false "登录语言代号"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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

// AddWorkflow 添加工作流程
// @Summary 添加工作流程
// @description 调用srv中的workflow服务，添加工作流程
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param workflow body workflow.AddWorkflowRequest true "工作流程信息"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /workflows [post]
func (t *Workflow) AddWorkflow(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddWorkflow, loggerx.MsgProcessStarted)

	type Request struct {
		Workflow workflow.AddRequest `json:"workflow"`
		Nodes    []node.AddRequest   `json:"nodes"`
	}

	user := sessionx.GetAuthUserID(c)
	db := sessionx.GetUserCustomer(c)
	appId := sessionx.GetCurrentApp(c)

	var req Request
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddWorkflow, err)
		return
	}

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)
	wReq := req.Workflow
	wReq.AppId = appId
	wReq.Writer = user
	wReq.Database = db

	response, err := workflowService.AddWorkflow(context.TODO(), &wReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddWorkflow, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddWorkflow, fmt.Sprintf("Workflow[%s] create success", response.GetWfId()))

	// 添加流程对应的语言
	loggerx.InfoLog(c, ActionAddWorkflow, fmt.Sprintf("Process AddAppLanguageData:%s", loggerx.MsgProcessStarted))
	languageService := language.NewLanguageService("global", client.DefaultClient)
	langParams := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    req.Workflow.AppId,
		Type:     "workflows",
		Key:      response.GetWfId(),
		Value:    req.Workflow.GetWfName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}
	_, err = languageService.AddAppLanguageData(context.TODO(), &langParams)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddWorkflow, err)
		return
	}

	// 添加菜单对应的语言
	menuLangParams := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    req.Workflow.AppId,
		Type:     "workflows",
		Key:      "menu_" + response.GetWfId(),
		Value:    req.Workflow.GetMenuName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = languageService.AddAppLanguageData(context.TODO(), &menuLangParams)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddWorkflow, err)
		return
	}
	loggerx.InfoLog(c, ActionAddWorkflow, fmt.Sprintf("Process AddAppLanguageData:%s", loggerx.MsgProcessEnded))

	nodeService := node.NewNodeService("workflow", client.DefaultClient)
	for _, nReq := range req.Nodes {
		nReq.WfId = response.GetWfId()
		nReq.NodeType = "1"
		nReq.Writer = user
		nReq.Database = db

		nResp, err := nodeService.AddNode(context.TODO(), &nReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddWorkflow, err)
			return
		}
		loggerx.SuccessLog(c, ActionAddWorkflow, fmt.Sprintf("Workflow node [%s] create success", nResp.GetNodeId()))
	}

	// 添加工作流程成功后保存日志到DB
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
	params["workflow_name"] = wReq.GetWfName()    // 新规的时候取传入参数

	loggerx.ProcessLog(c, ActionAddWorkflow, msg.L018, params)

	// 通知刷新多语言数据
	langx.RefreshLanguage(user, sessionx.GetUserDomain(c))

	loggerx.InfoLog(c, ActionAddWorkflow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionAddWorkflow)),
		Data:    response,
	})
}

// ModifyWorkflow 更新工作流程
// @Summary 更新工作流程
// @description 调用srv中的workflow服务，更新工作流程
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param workflow_id path string true "工作流程ID"
// @Param body body workflow.ModifyWorkflowRequest true "工作流程属性"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /workflows/{workflow_id} [put]
func (t *Workflow) ModifyWorkflow(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyWorkflow, loggerx.MsgProcessStarted)

	var wReq workflow.ModifyRequest
	if err := c.BindJSON(&wReq); err != nil {
		httpx.GinHTTPError(c, ActionModifyWorkflow, err)
		return
	}

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)
	// 变更前查询工作流程信息
	var freq workflow.WorkflowRequest
	freq.WfId = c.Param("wf_id")
	freq.Database = sessionx.GetUserCustomer(c)

	fresponse, err := workflowService.FindWorkflow(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyWorkflow, err)
		return
	}
	workflowInfo := fresponse.GetWorkflow()

	wReq.WfId = c.Param("wf_id")
	wReq.Writer = sessionx.GetAuthUserID(c)
	wReq.Database = sessionx.GetUserCustomer(c)

	response, err := workflowService.ModifyWorkflow(context.TODO(), &wReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyWorkflow, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyWorkflow, fmt.Sprintf("Workflow[%s] update success", wReq.GetWfId()))

	// 变更成功后，比较变更的结果，记录日志
	// 比较workflow名称
	wfname := langx.GetLangData(db, domain, lang, workflowInfo.GetWfName())
	if wfname != wReq.GetWfName() {
		// 变更流程对应的语言
		loggerx.InfoLog(c, ActionAddWorkflow, fmt.Sprintf("Process AddAppLanguageData:%s", loggerx.MsgProcessStarted))
		languageService := language.NewLanguageService("global", client.DefaultClient)
		langParams := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    workflowInfo.AppId,
			Type:     "workflows",
			Key:      wReq.GetWfId(),
			Value:    wReq.GetWfName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}
		_, err = languageService.AddAppLanguageData(context.TODO(), &langParams)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddWorkflow, err)
			return
		}
		loggerx.InfoLog(c, ActionAddWorkflow, fmt.Sprintf("Process AddAppLanguageData:%s", loggerx.MsgProcessEnded))
	}
	// 比较菜单名称
	menuname := langx.GetLangData(db, domain, lang, workflowInfo.GetMenuName())
	if menuname != wReq.GetMenuName() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["workflow_name"] = "{{" + workflowInfo.GetWfName() + "}}"
		params["menu_name"] = wReq.GetMenuName()

		loggerx.ProcessLog(c, ActionModifyWorkflow, msg.L021, params)

		// 变更菜单对应的语言
		loggerx.InfoLog(c, ActionAddWorkflow, fmt.Sprintf("Process AddAppLanguageData:%s", loggerx.MsgProcessStarted))
		languageService := language.NewLanguageService("global", client.DefaultClient)
		langParams := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    workflowInfo.AppId,
			Type:     "workflows",
			Key:      "menu_" + wReq.GetWfId(),
			Value:    wReq.GetMenuName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}
		_, err = languageService.AddAppLanguageData(context.TODO(), &langParams)
		if err != nil {
			httpx.GinHTTPError(c, ActionAddWorkflow, err)
			return
		}
		loggerx.InfoLog(c, ActionAddWorkflow, fmt.Sprintf("Process AddAppLanguageData:%s", loggerx.MsgProcessEnded))
	}

	// 比较有效性是否变更
	isvalid := "false"
	if workflowInfo.GetIsValid() {
		isvalid = "true"
	}
	if isvalid != wReq.GetIsValid() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["workflow_name"] = "{{" + workflowInfo.GetWfName() + "}}"
		if wReq.GetIsValid() == "true" {
			// 工作流程设置为有效的日志
			loggerx.ProcessLog(c, ActionModifyWorkflow, msg.L019, params)
		} else {
			// 工作流程设置为无效的日志
			loggerx.ProcessLog(c, ActionModifyWorkflow, msg.L020, params)
		}
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionModifyWorkflow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionModifyWorkflow)),
		Data:    response,
	})
}

// DeleteWorkflow 硬删除工作流程
// @Summary 硬删除工作流程
// @description 调用srv中的workflow服务，硬删除工作流程
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param workflow_id path string true "工作流程ID"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /workflows/{workflow_id} [delete]
func (t *Workflow) DeleteWorkflow(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteWorkflow, loggerx.MsgProcessStarted)

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)

	langData := langx.GetLanguageData(db, domain, lang)

	workflows := c.QueryArray("workflows")
	var deleteWorks []*workflow.Workflow
	for _, wk := range workflows {
		// 删除前查询工作流程信息
		var freq workflow.WorkflowRequest
		freq.WfId = wk
		freq.Database = sessionx.GetUserCustomer(c)

		fresponse, err := workflowService.FindWorkflow(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteWorkflow, err)
			return
		}
		deleteWorks = append(deleteWorks, fresponse.GetWorkflow())
	}

	var req workflow.DeleteRequest
	req.Workflows = workflows
	req.Database = sessionx.GetUserCustomer(c)

	response, err := workflowService.DeleteWorkflow(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteWorkflow, err)
		return
	}

	loggerx.SuccessLog(c, ActionDeleteWorkflow, fmt.Sprintf("Workflow[%v] delete success", req.GetWorkflows()))

	for _, wf := range deleteWorks {
		// 删除工作流程后保存日志到DB
		rname := strings.Builder{}
		rname.WriteString(langx.GetLangValue(langData, wf.GetWfName(), langx.DefaultResult))
		rname.WriteString("(")
		rname.WriteString(sessionx.GetCurrentLanguage(c))
		rname.WriteString(")")
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c) // 取共通用户名
		params["workflow_name"] = rname.String()

		loggerx.ProcessLog(c, ActionDeleteWorkflow, msg.L022, params)

		// 删除流程对应的语言
		languageService := language.NewLanguageService("global", client.DefaultClient)
		langParams := language.DeleteAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			AppId:    wf.AppId,
			Type:     "workflows",
			Key:      wf.GetWfId(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}
		_, err = languageService.DeleteAppLanguageData(context.TODO(), &langParams)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteWorkflow, err)
			return
		}

		// 删除菜单对应的语言
		menuLangParams := language.DeleteAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			AppId:    wf.AppId,
			Type:     "workflows",
			Key:      "menu_" + wf.GetWfId(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}
		_, err = languageService.DeleteAppLanguageData(context.TODO(), &menuLangParams)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteWorkflow, err)
			return
		}
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionDeleteWorkflow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, WorkflowProcessName, ActionDeleteWorkflow)),
		Data:    response,
	})
}

// Admit 承认
// @Summary 承认
// @description 调用srv中的workflow服务，承认
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param workflow_id path string true "工作流程ID"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
// @Summary 拒绝
// @description 调用srv中的workflow服务，拒绝
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param workflow_id path string true "工作流程ID"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
// @Summary 查找当前台账需要流程的操作
// @description 调用srv中的workflow服务，查找当前台账需要流程的操作
// @Tags Workflow
// @Accept json
// @Security JWT
// @Produce  json
// @Param workflow_id path string true "工作流程ID"
// @success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
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
