package webui

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/containerx"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/report/proto/dashboard"
	"rxcsoft.cn/pit3/srv/report/proto/report"
)

// Dashboard 仪表盘
type Dashboard struct{}

// log出力
const (
	DashboardProcessName    = "Dashboard"
	ActionFindDashboards    = "FindDashboards"
	ActionFindDashboardData = "FindDashboardData"
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

// FindDashboardData 通过仪表盘ID获取仪表盘数据情报
// @Router /dashboards/{dash_id}/data [get]
func (r *Dashboard) FindDashboardData(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDashboardData, loggerx.MsgProcessStarted)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Hour * 1
		o.DialTimeout = time.Hour * 1
	}

	dashboardId := c.Param("dash_id")
	db := sessionx.GetUserCustomer(c)

	dashboardService := dashboard.NewDashboardService("report", client.DefaultClient)

	var dreq dashboard.FindDashboardRequest
	dreq.DashboardId = dashboardId
	dreq.Database = db

	dResp, err := dashboardService.FindDashboard(context.TODO(), &dreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDashboardData, err)
		return
	}

	reportService := report.NewReportService("report", client.DefaultClient)

	var freq report.FindReportRequest
	freq.ReportId = dResp.GetDashboard().GetReportId()
	freq.Database = db

	fresp, err := reportService.FindReport(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindReport, err)
		return
	}

	datastoreId := fresp.GetReport().GetDatastoreId()

	accessKeys := sessionx.GetUserAccessKeys(c, datastoreId, "R")

	var req dashboard.FindDashboardDataRequest
	req.DashboardId = dashboardId
	req.Owners = accessKeys
	req.Database = db

	response, err := dashboardService.FindDashboardData(context.TODO(), &req, opss)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDashboardData, err)
		return
	}

	loggerx.InfoLog(c, ActionFindDashboardData, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DashboardProcessName, ActionFindDashboardData)),
		Data:    response,
	})
}
