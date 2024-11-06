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
	"rxcsoft.cn/pit3/srv/database/proto/option"
)

// Option 选项
type Option struct{}

// log出力
const (
	OptionProcessName      = "Option"
	ActionFindOptions      = "FindOptions"
	ActionFindOption       = "FindOption"
	ActionFindOptionLabels = "FindOptionLabels"
)

// FindOptions 获取所有选项
// @Router /options [get]
func (u *Option) FindOptions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOptions, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionsRequest
	// 从query中获取参数
	req.OptionName = c.Query("option_name")
	req.OptionMemo = c.Query("option_memo")
	req.InvalidatedIn = c.Query("invalidated_in")
	// 从共通中获取参数
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)
	if c.Query("app_id") != "" {
		req.AppId = c.Query("app_id")
	}

	response, err := optionService.FindOptions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindOptions, err)
		return
	}

	loggerx.InfoLog(c, ActionFindOptions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionFindOptions)),
		Data:    response.GetOptions(),
	})
}

// FindOptionLabels 获取所有选项数据
// @Router /options/{o_id} [get]
func (u *Option) FindOptionLabels(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOptionLabels, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionLabelsRequest
	// 从共通中获取参数
	req.Database = sessionx.GetUserCustomer(c)
	req.AppId = sessionx.GetCurrentApp(c)

	response, err := optionService.FindOptionLabels(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindOptionLabels, err)
		return
	}

	loggerx.InfoLog(c, ActionFindOptionLabels, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionFindOptionLabels)),
		Data:    response.GetOptions(),
	})
}

// FindOption 获取选项
// @Router /options/{o_id} [get]
func (u *Option) FindOption(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOption, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionRequest
	// 从path中获取参数
	req.OptionId = c.Param("o_id")
	req.Invalid = c.Query("invalidated_in")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := optionService.FindOption(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindOption, err)
		return
	}

	loggerx.InfoLog(c, ActionFindOption, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, OptionProcessName, ActionFindOption)),
		Data:    response.GetOptions(),
	})
}
