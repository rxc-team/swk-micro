package dev

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
	AllowProcessName  = "Allow"
	AllowFindAllows   = "FindAllows"
	AllowFindAllow    = "FindAllow"
	AllowAddAllow     = "AddAllow"
	AllowModifyAllow  = "ModifyAllow"
	AllowDeleteAllow  = "DeleteAllow"
	AllowDeleteAllows = "DeleteAllows"
)

// FindAllows 获取所有许可操作
// @Router /allows [get]
func (f *Allow) FindAllows(c *gin.Context) {
	loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessStarted)

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var req allow.FindAllowsRequest
	// 从query获取
	req.AllowType = c.Query("allow_type")
	req.ObjectType = c.Query("object_type")

	response, err := allowService.FindAllows(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, AllowFindAllows, err)
		return
	}

	actionService := action.NewActionService("manage", client.DefaultClient)

	var areq action.FindActionsRequest
	aResp, err := actionService.FindActions(context.TODO(), &areq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindActions, err)
		return
	}

	allowsData := response.GetAllows()
	lang := sessionx.GetCurrentLanguage(c)

	actions := aResp.GetActions()

	for _, a := range allowsData {
		for _, x := range a.GetActions() {
			for _, y := range actions {
				if a.AllowType == y.ActionObject && x.ApiKey == y.ActionKey && x.GroupKey == y.ActionGroup {
					switch lang {
					case "zh-CN":
						x.ActionName = y.ActionName["zh_CN"]
					case "ja-JP":
						x.ActionName = y.ActionName["ja_JP"]
					case "en-US":
						x.ActionName = y.ActionName["en_US"]
					case "th-TH":
						x.ActionName = y.ActionName["th_TH"]
					default:
						x.ActionName = y.ActionName["ja_JP"]
					}
				}
			}
		}
	}

	loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowFindAllows)),
		Data:    allowsData,
	})
}

// FindLevelAllows 获取所有许可操作
// @Router /allows [get]
func (f *Allow) FindLevelAllows(c *gin.Context) {
	loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessStarted)

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
		httpx.GinHTTPError(c, LevelFindLevel, err)
		return
	}

	if len(levelResp.GetLevel().GetAllows()) == 0 {
		loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowFindAllows)),
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
		httpx.GinHTTPError(c, AllowFindAllows, err)
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
	lang := sessionx.GetCurrentLanguage(c)

	actions := aResp.GetActions()

	for _, a := range allowsData {
		for _, x := range a.GetActions() {
			for _, y := range actions {
				if a.AllowType == y.ActionObject && x.ApiKey == y.ActionKey && x.GroupKey == y.ActionGroup {
					switch lang {
					case "zh-CN":
						x.ActionName = y.ActionName["zh_CN"]
					case "ja-JP":
						x.ActionName = y.ActionName["ja_JP"]
					case "en-US":
						x.ActionName = y.ActionName["en_US"]
					case "th-TH":
						x.ActionName = y.ActionName["th_TH"]
					default:
						x.ActionName = y.ActionName["ja_JP"]
					}
				}
			}
		}
	}

	loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowFindAllows)),
		Data:    allowsData,
	})
}

// FindLevelAllows 获取所有许可操作
// @Router /allows [get]
func (f *Allow) CheckAllow(c *gin.Context) {
	loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessStarted)

	allowType := c.Query("allow_type")

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
		httpx.GinHTTPError(c, LevelFindLevel, err)
		return
	}

	if len(levelResp.GetLevel().GetAllows()) == 0 {
		loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowFindAllows)),
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
		httpx.GinHTTPError(c, AllowFindAllows, err)
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

	loggerx.InfoLog(c, AllowFindAllows, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowFindAllows)),
		Data:    result,
	})
}

// FindAllow 获取许可操作
// @Router /allows/{allow_id} [get]
func (f *Allow) FindAllow(c *gin.Context) {
	loggerx.InfoLog(c, AllowFindAllow, loggerx.MsgProcessStarted)

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var req allow.FindAllowRequest
	req.AllowId = c.Param("allow_id")
	response, err := allowService.FindAllow(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, AllowFindAllow, err)
		return
	}

	actionService := action.NewActionService("manage", client.DefaultClient)

	var areq action.FindActionsRequest
	aResp, err := actionService.FindActions(context.TODO(), &areq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindActions, err)
		return
	}

	allowData := response.GetAllow()
	lang := sessionx.GetCurrentLanguage(c)

	actions := aResp.GetActions()

	for _, x := range allowData.GetActions() {
		for _, y := range actions {
			if x.ApiKey == y.ActionKey && x.GroupKey == y.ActionGroup {
				switch lang {
				case "zh-CN":
					x.ActionName = y.ActionName["zh_CN"]
				case "ja-JP":
					x.ActionName = y.ActionName["ja_JP"]
				case "en-US":
					x.ActionName = y.ActionName["en_US"]
				case "th-TH":
					x.ActionName = y.ActionName["th_TH"]
				default:
					x.ActionName = y.ActionName["ja_JP"]
				}
			}
		}
	}

	loggerx.InfoLog(c, AllowFindAllow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowFindAllow)),
		Data:    allowData,
	})
}

// AddAllow 添加许可操作
// @Router /allows [post]
func (f *Allow) AddAllow(c *gin.Context) {
	loggerx.InfoLog(c, AllowAddAllow, loggerx.MsgProcessStarted)

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var req allow.AddAllowRequest
	// 从body中获取
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, AllowAddAllow, err)
		return
	}
	// 从共通中获取
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := allowService.AddAllow(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, AllowAddAllow, err)
		return
	}
	loggerx.SuccessLog(c, AllowAddAllow, fmt.Sprintf("Allow[%s] Create Success", response.GetAllowId()))

	loggerx.InfoLog(c, AllowAddAllow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowAddAllow)),
		Data:    response,
	})
}

// ModifyAllow 更新许可操作
// @Router /allows/{allow_id} [put]
func (f *Allow) ModifyAllow(c *gin.Context) {
	loggerx.InfoLog(c, AllowModifyAllow, loggerx.MsgProcessStarted)

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var req allow.ModifyAllowRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, AllowModifyAllow, err)
		return
	}
	// 从path中获取参数
	req.AllowId = c.Param("allow_id")
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := allowService.ModifyAllow(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, AllowModifyAllow, err)
		return
	}
	loggerx.SuccessLog(c, AllowModifyAllow, fmt.Sprintf(loggerx.MsgProcesSucceed, AllowModifyAllow))

	loggerx.InfoLog(c, AllowModifyAllow, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowModifyAllow)),
		Data:    response,
	})
}

// DeleteAllows 硬删除多个许可
// @Router /allows [delete]
func (f *Allow) DeleteAllows(c *gin.Context) {
	loggerx.InfoLog(c, AllowDeleteAllows, loggerx.MsgProcessStarted)

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var req allow.DeleteAllowsRequest
	req.AllowIds = c.QueryArray("allow_ids")

	response, err := allowService.DeleteAllows(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, AllowDeleteAllows, err)
		return
	}
	loggerx.SuccessLog(c, AllowDeleteAllows, fmt.Sprintf(loggerx.MsgProcesSucceed, AllowDeleteAllows))

	loggerx.InfoLog(c, AllowDeleteAllows, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, AllowProcessName, AllowDeleteAllows)),
		Data:    response,
	})
}
