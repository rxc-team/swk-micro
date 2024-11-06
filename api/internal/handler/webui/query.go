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
	"rxcsoft.cn/pit3/srv/database/proto/query"
)

// Query 快捷方式
type Query struct{}

// log出力
const (
	QueryProcessName          = "Query"
	ActionFindQueries         = "FindQueries"
	ActionFindQuery           = "FindQuery"
	ActionAddQuery            = "AddQuery"
	ActionModifyQuery         = "ModifyQuery"
	ActionDeleteQuery         = "DeleteQuery"
	ActionDeleteSelectQueries = "DeleteSelectQueries"
	ActionHardDeleteQueries   = "HardDeleteQueries"
)

// FindQueries 获取所有快捷方式
// @Router /querys [get]
func (u *Query) FindQueries(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindQueries, loggerx.MsgProcessStarted)

	queryService := query.NewQueryService("database", client.DefaultClient)

	var req query.FindQueriesRequest
	// 从query中获取参数
	req.DatastoreId = c.Query("d_id")
	req.QueryName = c.Query("q_name")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.UserId = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := queryService.FindQueries(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindQueries, err)
		return
	}

	loggerx.InfoLog(c, ActionFindQueries, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, QueryProcessName, ActionFindQueries)),
		Data:    response.GetQueryList(),
	})
}

// FindQuery 获取快捷方式
// @Router /querys/{q_id} [get]
func (u *Query) FindQuery(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindQuery, loggerx.MsgProcessStarted)

	queryService := query.NewQueryService("database", client.DefaultClient)

	var req query.FindQueryRequest
	// 从path中获取参数
	req.QueryId = c.Param("q_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := queryService.FindQuery(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindQuery, err)
		return
	}

	loggerx.InfoLog(c, ActionFindQuery, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, QueryProcessName, ActionFindQuery)),
		Data:    response.GetQuery(),
	})
}

// AddQuery 添加快捷方式
// @Router /querys [post]
func (u *Query) AddQuery(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddQuery, loggerx.MsgProcessStarted)

	queryService := query.NewQueryService("database", client.DefaultClient)

	var req query.AddRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddQuery, err)
		return
	}
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.UserId = sessionx.GetAuthUserID(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := queryService.AddQuery(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddQuery, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddQuery, fmt.Sprintf("Query[%s] create Success", response.GetQueryId()))

	loggerx.InfoLog(c, ActionAddQuery, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, QueryProcessName, ActionAddQuery)),
		Data:    response,
	})
}

// DeleteQuery 删除快捷方式
// @Router /querys/{q_id} [delete]
func (u *Query) DeleteQuery(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteQuery, loggerx.MsgProcessStarted)

	queryService := query.NewQueryService("database", client.DefaultClient)

	var req query.DeleteRequest
	// 从path中获取参数
	req.QueryId = c.Param("q_id")
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := queryService.DeleteQuery(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteQuery, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteQuery, fmt.Sprintf("Query[%s] delete Success", req.GetQueryId()))

	loggerx.InfoLog(c, ActionDeleteQuery, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, QueryProcessName, ActionDeleteQuery)),
		Data:    response,
	})
}
