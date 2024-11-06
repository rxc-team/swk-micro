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
	"rxcsoft.cn/pit3/srv/global/proto/language"
)

// Language 语言
type Language struct{}

// log出力
const (
	LanguageProcessName = "Language"
	ActionFindLanguage  = "FindLanguage"
)

// FindLanguage 获取语言数据
// @Summary 获取语言数据
// @description 调用srv中的language服务，获取语言数据
// @Tags Language
// @Accept json
// @Security JWT
// @Produce  json
// @Param lang_cd path string true "语言的Code"
// @Param domain query string true "域"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /languages/search [get]
func (l *Language) FindLanguage(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindLanguage, loggerx.MsgProcessStarted)

	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.LangCd = c.Query("lang_cd")
	req.Domain = sessionx.GetUserDomain(c)
	req.Database = sessionx.GetUserCustomer(c)
	response, err := languageService.FindLanguage(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindLanguage, err)
		return
	}

	loggerx.InfoLog(c, ActionFindLanguage, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, LanguageProcessName, ActionFindLanguage)),
		Data:    response,
	})
}
