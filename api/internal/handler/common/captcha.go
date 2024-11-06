package common

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/mailx"
	"rxcsoft.cn/pit3/api/internal/common/storex"
	"rxcsoft.cn/pit3/lib/msg"
)

// Captcha 验证码
type Captcha struct{}

//CaptchaVerify json request body.
type CaptchaVerify struct {
	ID          string `json:"id"`
	VerifyValue string `json:"verify_value"`
}

// log出力
const (
	CaptchaProcessName  = "Captcha"
	ActionCreatCaptcha  = "CreatCaptcha"
	ActionVerifyCaptcha = "VerifyCaptcha"
)

// CreatCaptcha 创建一个验证码
// @Router /captcha [get]
func (cp *Captcha) CreatCaptcha(c *gin.Context) {
	loggerx.InfoLog(c, ActionCreatCaptcha, loggerx.MsgProcessStarted)

	var config base64Captcha.ConfigDigit
	config.CaptchaLen = 6
	config.Height = 64
	config.Width = 160
	config.MaxSkew = 0.5
	config.DotCount = 5

	store := storex.NewRedisStore(120)
	base64Captcha.SetCustomStore(store)

	captchaID, captcaInterfaceInstance := base64Captcha.GenerateCaptcha("", config)
	base64blob := base64Captcha.CaptchaWriteToBase64Encoding(captcaInterfaceInstance)

	loggerx.InfoLog(c, ActionCreatCaptcha, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, CaptchaProcessName, ActionCreatCaptcha)),
		Data: gin.H{
			"image":      base64blob,
			"captcha_id": captchaID,
		},
	})
}

// VerifyCaptcha 校验验证码
// @Router /captcha [post]
func (cp *Captcha) VerifyCaptcha(c *gin.Context) {
	loggerx.InfoLog(c, ActionVerifyCaptcha, loggerx.MsgProcessStarted)

	var req CaptchaVerify
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionVerifyCaptcha, err)
		return
	}
	store := storex.NewRedisStore(120)
	base64Captcha.SetCustomStore(store)

	ok := base64Captcha.VerifyCaptchaAndIsClear(req.ID, req.VerifyValue, false)

	loggerx.InfoLog(c, ActionVerifyCaptcha, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, CaptchaProcessName, ActionVerifyCaptcha)),
		Data:    ok,
	})
}

// CreatSecondCaptcha 获取登录二次验证码
// @Router /login/captcha [get]
func (cp *Captcha) CreatSecondCaptcha(c *gin.Context) {
	loggerx.InfoLog(c, ActionCreatCaptcha, loggerx.MsgProcessStarted)

	// 参数
	db := c.Query("customer_id")
	email := c.Query("email")
	noticeEmail := c.Query("notice_email")

	// 生成随机密码
	captcha := cryptox.GenerateRandCaptcha()
	captchaID := primitive.NewObjectID().Hex()

	//设置过期时间
	store := storex.NewRedisStore(120)
	store.Set(captchaID, captcha)

	if len(noticeEmail) > 0 {
		// 发送验证码重置邮件
		// 定义收件人
		mailTo := []string{
			noticeEmail,
		}
		// 定义抄送人
		mailCcTo := []string{}
		// 邮件主题
		subject := "Proship user login authentication"
		// 邮件正文
		tpl := template.Must(template.ParseFiles("assets/html/captcha.html"))
		params := map[string]string{
			"email":   email,
			"captcha": captcha,
		}

		var out bytes.Buffer
		err := tpl.Execute(&out, params)
		if err != nil {
			httpx.GinHTTPError(c, ActionPasswordReset, err)
			return
		}

		err = mailx.SendMail(db, mailTo, mailCcTo, subject, out.String())
		if err != nil {
			httpx.GinHTTPError(c, ActionCreatCaptcha, err)
			return
		}
		loggerx.InfoLog(c, ActionCreatCaptcha, loggerx.MsgProcessEnded)
	}

	loggerx.InfoLog(c, ActionCreatCaptcha, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, PasswordProcessName, ActionCreatCaptcha)),
		Data: gin.H{
			"captcha_id": captchaID,
		},
	})
}

// VerifySecondCaptcha 检验二次验证
// @Router /login/VerifySecondCaptcha [Post]
func (cp *Captcha) VerifySecondCaptcha(c *gin.Context) {
	loggerx.InfoLog(c, ActionVerifyCaptcha, loggerx.MsgProcessStarted)

	// 参数
	var req CaptchaVerify
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionVerifyCaptcha, err)
		return
	}

	//Verify 验证数据
	store := storex.NewRedisStore(120)
	result := store.Verify(req.ID, req.VerifyValue, true)

	loggerx.InfoLog(c, ActionVerifyCaptcha, loggerx.MsgProcessStarted)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, CaptchaProcessName, ActionVerifyCaptcha)),
		Data:    result,
	})
}
