package admin

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
)

// Dashboard 仪表盘
type Dashboard struct{}

// log出力
const (
	DashboardProcessName          = "Dashboard"
	ActionFindDashboards          = "FindDashboards"
	ActionFindDashboard           = "FindDashboard"
	ActionFindDashboardData       = "FindDashboardData"
	ActionAddDashboard            = "AddDashboard"
	ActionModifyDashboard         = "ModifyDashboard"
	ActionDeleteDashboard         = "DeleteDashboard"
	ActionDeleteSelectDashboards  = "DeleteSelectDashboards"
	ActionHardDeleteDashboards    = "HardDeleteDashboards"
	ActionRecoverSelectDashboards = "RecoverSelectDashboards"
)

// FindDashboards 获取所属公司所属APP[所属报表]下所有仪表盘情报
// @Router /dashboards [get]
func (r *Dashboard) FindDashboards(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDashboards, loggerx.MsgProcessStarted)

	dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)

	var req dashboard.FindDashboardsRequest

	// Query数据
	req.ReportId = c.Query("report_id")

	// 共通数据
	req.Domain = sessionx.GetUserDomain(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := dashboardService.FindDashboards(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDashboards, err)
		return
	}

	needRole := c.Query("needRole")
	if needRole == "true" {
		roles := sessionx.GetUserRoles(c)
		set := containerx.New()

		pmService := permission.NewPermissionService("manage", client.DefaultClient)

		var preq permission.FindActionsRequest
		preq.RoleId = roles
		preq.PermissionType = "app"
		preq.AppId = sessionx.GetCurrentApp(c)
		preq.ActionType = "report"
		preq.Database = sessionx.GetUserCustomer(c)
		pResp, err := pmService.FindActions(context.TODO(), &preq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindFolders, err)
			return
		}
		for _, act := range pResp.GetActions() {
			set.AddAll(act.ObjectId)
		}

		rpList := set.ToList()
		allDs := response.GetDashboards()
		var result []*dashboard.Dashboard
		for _, reportID := range rpList {
			f, err := findDashboard(reportID, allDs)
			if err == nil {
				result = append(result, f...)
			}
		}
		// 字段排序
		sort.Sort(typesx.DashboardList(result))

		loggerx.InfoLog(c, ActionFindDashboards, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastores)),
			Data:    result,
		})

		return
	}

	loggerx.InfoLog(c, ActionFindDashboards, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DashboardProcessName, ActionFindDashboards)),
		Data:    response.GetDashboards(),
	})
}

func findDashboard(reportID string, dashList []*dashboard.Dashboard) (r []*dashboard.Dashboard, err error) {
	var reuslt []*dashboard.Dashboard
	for _, r := range dashList {
		if r.GetReportId() == reportID {
			reuslt = append(reuslt, r)
		}
	}

	if reuslt == nil {
		return nil, fmt.Errorf("not found")
	}

	return reuslt, nil
}

// FindDashboard 通过仪表盘ID获取单个仪表盘情报
// @Router /dashboards/{dash_id} [get]
func (r *Dashboard) FindDashboard(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDashboard, loggerx.MsgProcessStarted)

	dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)

	var req dashboard.FindDashboardRequest
	req.DashboardId = c.Param("dash_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := dashboardService.FindDashboard(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDashboard, err)
		return
	}

	loggerx.InfoLog(c, ActionFindDashboard, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DashboardProcessName, ActionFindDashboard)),
		Data:    response.GetDashboard(),
	})
}

// AddDashboard 添加单个仪表盘情报
// @Router /dashboards [post]
func (r *Dashboard) AddDashboard(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddDashboard, loggerx.MsgProcessStarted)

	dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)

	var req dashboard.AddDashboardRequest

	// Body数据
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddDashboard, err)
		return
	}

	// 共通数据
	req.Domain = sessionx.GetUserDomain(c)
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := dashboardService.AddDashboard(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDashboard, err)
		return
	}
	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["dashboard_name"] = req.GetDashboardName()
	loggerx.ProcessLog(c, ActionAddDashboard, msg.L071, params)

	loggerx.SuccessLog(c, ActionAddDashboard, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddDashboard"))

	// 添加多语言数据
	langService := language.NewLanguageService("global", client.DefaultClient)

	languageReq := language.AddAppLanguageDataRequest{
		Domain:   sessionx.GetUserDomain(c),
		LangCd:   sessionx.GetCurrentLanguage(c),
		AppId:    sessionx.GetCurrentApp(c),
		Type:     "dashboards",
		Key:      response.GetDashboardId(),
		Value:    req.GetDashboardName(),
		Writer:   sessionx.GetAuthUserID(c),
		Database: sessionx.GetUserCustomer(c),
	}

	_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddDashboard, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddDashboard, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, req.Domain)

	loggerx.InfoLog(c, ActionAddDashboard, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, DashboardProcessName, ActionAddDashboard)),
		Data:    response,
	})
}

// ModifyDashboard 更新单个仪表盘情报
// @Router /dashboards/{dash_id} [PUT]
func (r *Dashboard) ModifyDashboard(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyDashboard, loggerx.MsgProcessStarted)
	dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)

	var req dashboard.ModifyDashboardRequest
	// Body数据
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyDashboard, err)
		return
	}
	// Path数据
	req.DashboardId = c.Param("dash_id")
	// 共通数据
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := dashboardService.ModifyDashboard(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyDashboard, err)
		return
	}
	//处理log
	params := make(map[string]string)
	params["user_name"] = sessionx.GetUserName(c)
	params["dashboard_name"] = req.GetDashboardName()
	loggerx.ProcessLog(c, ActionAddDashboard, msg.L073, params)

	loggerx.SuccessLog(c, ActionModifyDashboard, fmt.Sprintf(loggerx.MsgProcesSucceed, "ModifyDashboard"))

	if req.GetDashboardName() != "" {
		// 添加多语言数据
		langService := language.NewLanguageService("global", client.DefaultClient)

		languageReq := language.AddAppLanguageDataRequest{
			Domain:   sessionx.GetUserDomain(c),
			LangCd:   sessionx.GetCurrentLanguage(c),
			AppId:    sessionx.GetCurrentApp(c),
			Type:     "dashboards",
			Key:      req.GetDashboardId(),
			Value:    req.GetDashboardName(),
			Writer:   sessionx.GetAuthUserID(c),
			Database: sessionx.GetUserCustomer(c),
		}

		_, err = langService.AddAppLanguageData(context.TODO(), &languageReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionModifyDashboard, err)
			return
		}
		loggerx.SuccessLog(c, ActionModifyDashboard, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddAppLanguageData"))
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(req.Writer, sessionx.GetUserDomain(c))

	loggerx.InfoLog(c, ActionModifyDashboard, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, DashboardProcessName, ActionModifyDashboard)),
		Data:    response,
	})
}

// HardDeleteDashboards 物理删除多个仪表盘情报
// @Router /phydel/dashboards [delete]
func (r *Dashboard) HardDeleteDashboards(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteDashboards, loggerx.MsgProcessStarted)

	dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)
	appId := sessionx.GetCurrentApp(c)
	langData := langx.GetLanguageData(db, lang, domain)

	var req dashboard.HardDeleteDashboardsRequest
	// Query数据
	req.DashboardIdList = c.QueryArray("dashboard_id_list")
	req.Database = sessionx.GetUserCustomer(c)
	var dashboardNameList []string
	for _, id := range req.DashboardIdList {
		var reqF dashboard.FindDashboardRequest
		reqF.DashboardId = id
		reqF.Database = sessionx.GetUserCustomer(c)

		result, err := dashboardService.FindDashboard(context.TODO(), &reqF)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteDashboards, err)
			return
		}
		dashboardName := langx.GetLangValue(langData, result.GetDashboard().DashboardName, langx.DefaultResult)
		dashboardNameList = append(dashboardNameList, dashboardName)
	}

	response, err := dashboardService.HardDeleteDashboards(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteDashboards, err)
		return
	}

	for _, name := range dashboardNameList {
		dashboardName := strings.Builder{}
		dashboardName.WriteString(name)
		dashboardName.WriteString("(")
		dashboardName.WriteString(sessionx.GetCurrentLanguage(c))
		dashboardName.WriteString(")")
		//处理log
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["dashboard_name"] = dashboardName.String()
		loggerx.ProcessLog(c, ActionHardDeleteDashboards, msg.L072, params)
	}

	loggerx.SuccessLog(c, ActionHardDeleteDashboards, fmt.Sprintf(loggerx.MsgProcesSucceed, "HardDeleteDashboards"))

	for _, d := range req.GetDashboardIdList() {
		// 删除app语言
		langx.DeleteAppLanguageData(db, domain, appId, "dashboard", d)
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(sessionx.GetAuthUserID(c), domain)

	loggerx.InfoLog(c, ActionHardDeleteDashboards, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, DashboardProcessName, ActionHardDeleteDashboards)),
		Data:    response,
	})
}
