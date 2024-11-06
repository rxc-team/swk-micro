package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/option"
)

// Option 选项
type Option struct{}

// log出力
const (
	OptionProcessName      = "Option"
	ActionFindOption       = "FindOption"
	ActionFindOptions      = "FindOptions"
	ActionFindOptionLabels = "FindOptionLabels"
)

// FindOptions 获取所有选项
// @Summary 获取所有选项
// @description 调用srv中的option服务，获取所有选项
// @Tags Option
// @Accept json
// @Security JWT
// @Produce  json
// @Param name query string false "选项名"
// @Param memo query string false "选项memo"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /options [get]
func (u *Option) FindOptions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOptions, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionsRequest
	// 从query中获取参数
	req.OptionName = c.Query("name")
	req.OptionMemo = c.Query("memo")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.InvalidatedIn = c.Query("invalidated_in")
	req.Database = sessionx.GetUserCustomer(c)

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
// @Summary 获取所有选项数据
// @description 调用srv中的option服务，获取所有选项数据
// @Tags Option
// @Accept json
// @Security JWT
// @Produce  json
// @Param o_id path string true "选项ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /options/{o_id} [get]
func (u *Option) FindOptionLabels(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOptionLabels, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionLabelsRequest
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

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
// @Summary 获取选项
// @description 调用srv中的option服务，通过ID获取选项
// @Tags Option
// @Accept json
// @Security JWT
// @Produce  json
// @Param o_id path string true "选项ID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /options/{o_id} [get]
func (u *Option) FindOption(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindOption, loggerx.MsgProcessStarted)

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionRequest
	// 从path中获取参数
	req.OptionId = c.Param("o_id")
	// 从共通中获取参数
	req.AppId = sessionx.GetCurrentApp(c)
	req.Invalid = c.Query("invalidated_in")
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
