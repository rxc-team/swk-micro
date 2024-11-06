package common

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// Mail 用户
type Mail struct{}

// log出力
const (
	userProcessName  = "User"
	ActionActiveMail = "ActiveMail"
)

// ActiveMail 激活用户邮箱
// @Router /active/mail [patch]
func (u *Mail) ActiveMail(c *gin.Context) {
	loggerx.InfoLog(c, ActionActiveMail, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	type ActiveMailParams struct {
		LoginID     string `json:"login_id"`
		NoticeEmail string `json:"notice_email"`
	}

	var params ActiveMailParams
	// 从body中获取参数
	if err := c.BindJSON(&params); err != nil {
		httpx.GinHTTPError(c, ActionActiveMail, err)
		return
	}

	// 变更前查询用户信息
	var freq user.EmailRequest
	freq.Email = params.LoginID
	fresponse, err := userService.FindUserByEmail(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionActiveMail, err)
		return
	}
	userInfo := fresponse.GetUser()

	email := params.NoticeEmail
	if len(email) == 0 {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionActiveMail)),
			Data:    false,
		})
		return
	}

	if email != userInfo.GetNoticeEmail() {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionActiveMail)),
			Data:    false,
		})
		return
	}

	if userInfo.GetNoticeEmailStatus() == "Verified" {
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionActiveMail)),
			Data:    true,
		})
		return
	}

	var req user.ModifyUserRequest
	req.UserId = userInfo.GetUserId()
	req.Writer = userInfo.GetUserId()
	req.Database = userInfo.GetCustomerId()

	req.NoticeEmailStatus = "Verified"

	_, err = userService.ModifyUser(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionActiveMail, err)
		return
	}

	loggerx.InfoLog(c, ActionActiveMail, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, userProcessName, ActionActiveMail)),
		Data:    true,
	})
}
