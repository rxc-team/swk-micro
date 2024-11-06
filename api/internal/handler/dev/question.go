package dev

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/errors"

	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/configx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/mailx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/global/proto/question"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// Question 问题
type Question struct{}

// log出力
const (
	QuestionProcessName   = "Question"
	ActionFindQuestions   = "FindQuestions"
	ActionFindQuestion    = "FindQuestion"
	ActionAddQuestion     = "AddQuestion"
	ActionModifyQuestion  = "ModifyQuestion"
	ActionDeleteQuestion  = "DeleteQuestion"
	ActionDeleteQuestions = "DeleteQuestions"
	SentQuestionEmail     = "SentQuestionEmail"
)

// FindQuestion 获取单个问题
// @Router /questions/{question_id} [get]
func (t *Question) FindQuestion(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindQuestion, loggerx.MsgProcessStarted)

	questionService := question.NewQuestionService("global", client.DefaultClient)

	var req question.FindQuestionRequest
	req.QuestionId = c.Param("question_id")

	response, err := questionService.FindQuestion(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindQuestion, err)
		return
	}

	loggerx.InfoLog(c, ActionFindQuestion, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, QuestionProcessName, ActionFindQuestion)),
		Data:    response.GetQuestion(),
	})
}

// FindQuestions 获取多个问题
// @Router /questions [get]
func (t *Question) FindQuestions(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindQuestions, loggerx.MsgProcessStarted)

	questionService := question.NewQuestionService("global", client.DefaultClient)

	var req question.FindQuestionsRequest
	req.Title = c.Query("title")
	req.Type = c.Query("type")
	req.Function = c.Query("function")
	req.Status = c.Query("status")
	if sessionx.GetUserDomain(c) == sessionx.GetSuperDomain() {
		req.Domain = c.Query("domain")
	} else {
		req.Domain = sessionx.GetUserDomain(c)
	}

	response, err := questionService.FindQuestions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindQuestions, err)
		return
	}

	loggerx.InfoLog(c, ActionFindQuestions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, QuestionProcessName, ActionFindQuestions)),
		Data:    response.GetQuestions(),
	})
}

// AddQuestion 添加问题
// @Router /questions [post]
func (t *Question) AddQuestion(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddQuestion, loggerx.MsgProcessStarted)

	questionService := question.NewQuestionService("global", client.DefaultClient)

	db := sessionx.GetUserCustomer(c)
	domain := sessionx.GetUserDomain(c)
	lang := sessionx.GetCurrentLanguage(c)

	var req question.AddQuestionRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddQuestion, err)
		return
	}

	req.Domain = sessionx.GetUserDomain(c)
	req.Writer = sessionx.GetAuthUserID(c)
	req.ResponderId, req.ResponderName = sessionx.GetSuperAdmin(c)

	// 问题定位共通情报取得
	customerID := sessionx.GetUserCustomer(c)
	userID := sessionx.GetAuthUserID(c)
	appID := sessionx.GetCurrentApp(c)
	// 处理月度取得
	cfg, err := configx.GetConfigVal(customerID, appID)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddQuestion, err)
		return
	}
	syoriym := cfg.GetSyoriYm()
	// 客户情报取得
	customerName := ""
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	var cReq customer.FindCustomerRequest
	cReq.CustomerId = customerID
	cResponse, err := customerService.FindCustomer(context.TODO(), &cReq)
	if err == nil {
		customerName = cResponse.Customer.CustomerName
	}
	// 客户的用户情报取得
	userName := ""
	userType := ""
	userService := user.NewUserService("manage", client.DefaultClient)
	var uReq user.FindUserRequest
	uReq.UserId = userID
	uReq.Database = customerID
	uResponse, err := userService.FindUser(context.TODO(), &uReq)
	if err == nil {
		userName = uResponse.User.UserName
		userType = strconv.FormatInt(int64(uResponse.User.UserType), 10)
	}

	// 问题定位情报编辑
	locates := strings.Builder{}
	// 用户公司名称
	locates.WriteString("Customer:")
	locates.WriteString(customerName)
	locates.WriteString(";")
	// 用户公司域名
	locates.WriteString("Domain:")
	locates.WriteString(sessionx.GetUserDomain(c))
	locates.WriteString(";")
	// 提问用户
	locates.WriteString("User:")
	locates.WriteString(userName)
	locates.WriteString(";")
	// 提问用户类型
	locates.WriteString("UserType:")
	locates.WriteString(userType)
	locates.WriteString(";")
	// 应用名称
	locates.WriteString("App:")
	locates.WriteString(langx.GetLangData(db, domain, lang, langx.GetAppKey(appID)))
	locates.WriteString(";")
	// 处理月度
	locates.WriteString("SyoriYm:")
	locates.WriteString(syoriym)
	locates.WriteString(";")
	// 用户语言
	locates.WriteString("Language:")
	locates.WriteString(sessionx.GetCurrentLanguage(c))
	locates.WriteString(";")
	// 用户时区
	locates.WriteString("Timezone:")
	locates.WriteString(sessionx.GetCurrentTimezone(c))
	locates.WriteString(";")

	req.Locations = locates.String() + req.Locations

	response, err := questionService.AddQuestion(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddQuestion, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddQuestion, fmt.Sprintf("Question[%s] create Success", response.GetQuestionId()))

	param := wsx.MessageParam{
		Sender:    req.QuestionerName,
		Recipient: req.ResponderId,
		Domain:    req.Domain,
		MsgType:   "qa",
		Link:      "/question/edit/" + response.QuestionId,
		Content:   req.GetTitle(),
		Status:    "unread",
	}
	wsx.SendToUser(param)
	// 向系统管理员发送客户添加问题的邮件
	var userReq user.FindUserRequest
	userReq.UserId = param.Recipient
	userReq.Database = "system"
	userResponse, err := userService.FindUser(context.TODO(), &userReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUser, err)
		return
	}
	// 定义主题和消息
	subject := "Customers have new question to deal with"
	message := "お客様には解決すべき新しい問題があります。"
	err = mailx.SendQuestionEmail(db, subject, message, userResponse.GetUser().GetNoticeEmail())
	if err != nil {
		httpx.GinHTTPError(c, SentQuestionEmail, err)
		return
	}
	loggerx.InfoLog(c, ActionAddQuestion, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, QuestionProcessName, ActionAddQuestion)),
		Data:    response,
	})
}

// ModifyQuestion 更新问题
// @Router /questions/{question_id} [put]
func (t *Question) ModifyQuestion(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyQuestion, loggerx.MsgProcessStarted)

	questionService := question.NewQuestionService("global", client.DefaultClient)

	var req question.ModifyQuestionRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyQuestion, err)
		return
	}

	req.QuestionId = c.Param("question_id")

	if req.Postscript != nil {
		req.Postscript.Postscripter = sessionx.GetAuthUserID(c)
	}
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := questionService.ModifyQuestion(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyQuestion, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyQuestion, fmt.Sprintf("Question[%s] Update Success", req.GetQuestionId()))

	loggerx.InfoLog(c, ActionModifyQuestion, fmt.Sprintf("Process[%s]", loggerx.MsgProcessStarted))
	var reqF question.FindQuestionRequest
	reqF.QuestionId = req.QuestionId
	res, e := questionService.FindQuestion(context.TODO(), &reqF)
	if e != nil {
		httpx.GinHTTPError(c, ActionModifyQuestion, e)
		return
	}
	loggerx.InfoLog(c, ActionModifyQuestion, fmt.Sprintf("Process[%s]", loggerx.MsgProcessEnded))

	var sender = ""
	// var recipient = ""
	var recipientID = ""

	sender = req.GetPostscript().GetPostscripterName()
	// recipient = res.GetQuestion().GetQuestionerName()
	recipientID = res.GetQuestion().GetCreatedBy()

	param := wsx.MessageParam{
		Sender:    sender,
		Recipient: recipientID,
		Domain:    res.GetQuestion().GetDomain(),
		MsgType:   "qa",
		Link:      "/question/edit/" + req.QuestionId,
		Content:   res.GetQuestion().GetTitle(),
		Status:    "unread",
	}
	wsx.SendToUser(param)

	// 向系统管理员发送客户更新问题的邮件
	userService := user.NewUserService("manage", client.DefaultClient)
	// 客户情报取得
	customerName := ""
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	var cReq customer.FindCustomerByDomainRequest
	cReq.Domain = res.GetQuestion().GetDomain()
	cResponse, err := customerService.FindCustomerByDomain(context.TODO(), &cReq)
	if err == nil {
		customerName = cResponse.Customer.CustomerName
	}
	// 客户端用户情报取得
	var userReq user.FindUserRequest
	userReq.UserId = param.Recipient
	userReq.Database = cResponse.GetCustomer().GetCustomerId()
	userResponse, err := userService.FindUser(context.TODO(), &userReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindUser, err)
		return
	}
	// 根据问题状态更改邮件消息的内容
	db := sessionx.GetUserCustomer(c)
	if res.GetQuestion().GetStatus() == "open" {
		// 定义主题和消息
		subject := "Customers have new question to deal with"
		message := "システム管理者は、顧客 { " + customerName + " } のユーザー {" + res.GetQuestion().GetQuestionerName() + "} からの次の質問に回答しました: " + "\n\t<br/>" + "タイトル: " + res.GetQuestion().GetTitle() + "\n\t<br/>"
		// 发送邮件
		err := mailx.SendQuestionEmail(db, subject, message, userResponse.GetUser().GetNoticeEmail())
		if err != nil {
			httpx.GinHTTPError(c, SentQuestionEmail, err)
			return
		}
	} else {
		// 定义主题和消息
		subject := "Customer closed a question"
		message := "システム管理者は、顧客 { " + customerName + " } のユーザー {" + res.GetQuestion().GetQuestionerName() + "} に関する次の問題をクローズしました:" + "\n\t<br/>" + "タイトル: " + res.GetQuestion().GetTitle() + "\n\t<br/>"
		// 发送邮件
		err := mailx.SendQuestionEmail(db, subject, message, userResponse.GetUser().GetNoticeEmail())
		if err != nil {
			httpx.GinHTTPError(c, SentQuestionEmail, err)
			return
		}
	}

	loggerx.InfoLog(c, ActionModifyQuestion, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, QuestionProcessName, ActionModifyQuestion)),
		Data:    response,
	})
}

// DeleteQuestions 硬删除多个问题
// @Router /questions [delete]
func (t *Question) DeleteQuestions(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteQuestions, loggerx.MsgProcessStarted)

	questionService := question.NewQuestionService("global", client.DefaultClient)

	var req question.DeleteQuestionsRequest
	req.QuestionIdList = c.QueryArray("question_id_list")

	var reqFind question.FindQuestionRequest
	for _, qID := range req.QuestionIdList {
		reqFind.QuestionId = qID
		// 通过id查询出问题情报
		res, err := questionService.FindQuestion(context.TODO(), &reqFind)
		if err != nil {
			httpx.GinHTTPError(c, ActionDeleteQuestions, err)
			return
		}
		// 获取当前问题所属顾客的domain
		domain := res.GetQuestion().GetDomain()
		superdomain := sessionx.GetUserDomain(c)
		// 删除获取问题创建时的图片
		for _, imgName := range res.GetQuestion().GetImages() {
			// 删除问题的图片
			_, _, err := filex.DeletePublicHeaderFile(domain, imgName)
			if err != nil {
				// 若minio中的文件已被删除，让程序继续执行
				er := errors.Parse(err.Error())
				if er.GetDetail() != "The specified key does not exist." {
					httpx.GinHTTPError(c, ActionDeleteQuestions, err)
					return
				}
			}
		}
		// 删除问题回复的图片
		for _, postscriptsImg := range res.GetQuestion().GetPostscripts() {
			// 若存在追记图片，则删除
			if len(postscriptsImg.GetImages()) > 0 {
				if postscriptsImg.GetPostscripterName() == "SYSTEM" {
					_, _, err := filex.DeletePublicHeaderFile(superdomain, postscriptsImg.GetImages()[0])
					if err != nil {
						// 若minio中的文件已被删除，让程序继续执行
						er := errors.Parse(err.Error())
						if er.GetDetail() != "The specified key does not exist." {
							httpx.GinHTTPError(c, ActionDeleteQuestions, err)
							return
						}
					}
				} else {
					_, _, err := filex.DeletePublicHeaderFile(domain, postscriptsImg.GetImages()[0])
					if err != nil {
						// 若minio中的文件已被删除，让程序继续执行
						er := errors.Parse(err.Error())
						if er.GetDetail() != "The specified key does not exist." {
							httpx.GinHTTPError(c, ActionDeleteQuestions, err)
							return
						}
					}
				}
			}
		}
	}

	response, err := questionService.DeleteQuestions(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteQuestions, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteQuestions, fmt.Sprintf("Questions[%s] HardDelete Success", req.GetQuestionIdList()))

	loggerx.InfoLog(c, ActionDeleteQuestions, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, QuestionProcessName, ActionDeleteQuestions)),
		Data:    response,
	})
}
