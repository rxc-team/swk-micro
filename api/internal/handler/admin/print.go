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
	"rxcsoft.cn/pit3/srv/database/proto/print"
)

// Print 台账打印设置
type Print struct{}

// log出力使用
const (
	PrintProcessName       = "Print"
	ActionFindPrint        = "FindPrint"
	ActionAddPrint         = "AddPrint"
	ActionModifyPrint      = "ModifyPrint"
	ActionHardDeletePrints = "HardDeletePrints"
)

// FindPrint 获取台账打印设置
// @Router /prints/{datastore_id} [get]
func (f *Print) FindPrint(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindPrint, loggerx.MsgProcessStarted)

	printService := print.NewPrintService("database", client.DefaultClient)

	var req print.FindPrintRequest

	// 从共通中获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)
	// 从path中获取
	req.DatastoreId = c.Param("datastore_id")

	response, err := printService.FindPrint(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindPrint, err)
		return
	}

	loggerx.InfoLog(c, ActionFindPrint, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, PrintProcessName, ActionFindPrint)),
		Data:    response.GetPrint(),
	})
}

// AddPrint 添加台账打印设置
// @Router /prints [post]
func (f *Print) AddPrint(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddPrint, loggerx.MsgProcessStarted)

	printService := print.NewPrintService("database", client.DefaultClient)

	var req print.AddPrintRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddPrint, err)
		return
	}
	// 从共通中获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := printService.AddPrint(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddPrint, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddPrint, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionAddPrint))

	loggerx.InfoLog(c, ActionAddPrint, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, PrintProcessName, ActionAddPrint)),
		Data:    response,
	})
}

// ModifyPrint 更新台账打印设置
// @Router /prints/{datastore_id} [put]
func (f *Print) ModifyPrint(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyPrint, loggerx.MsgProcessStarted)

	printService := print.NewPrintService("database", client.DefaultClient)

	var req print.ModifyPrintRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyPrint, err)
		return
	}
	// 从path中获取参数
	req.DatastoreId = c.Param("datastore_id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := printService.ModifyPrint(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyPrint, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyPrint, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionModifyPrint))

	loggerx.InfoLog(c, ActionModifyPrint, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PrintProcessName, ActionModifyPrint)),
		Data:    response,
	})
}
