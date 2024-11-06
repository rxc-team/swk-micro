package dev

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datahistory"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// Customer 顾客
type Customer struct{}

// log出力
const (
	CustomerProcessName          = "Customer"
	ActionFindCustomers          = "FindCustomers"
	ActionFindCustomer           = "FindCustomer"
	ActionAddCustomer            = "AddCustomer"
	ActionModifyCustomer         = "ModifyCustomer"
	ActionModifyCustomerAdmin    = "ModifyCustomerAdmin"
	ActionDeleteCustomer         = "DeleteCustomer"
	ActionDeleteSelectCustomers  = "DeleteSelectCustomers"
	ActionHardDeleteCustomers    = "HardDeleteCustomers"
	ActionRecoverSelectCustomers = "RecoverSelectCustomers"
)

// FindCustomers 查找多个顾客记录
// @Router /customers [get]
func (u *Customer) FindCustomers(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindCustomers, loggerx.MsgProcessStarted)

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomersRequest
	req.CustomerName = c.Query("customer_name")
	req.InvalidatedIn = c.Query("invalidated_in")
	response, err := customerService.FindCustomers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindCustomers, err)
		return
	}

	var result []map[string]interface{}
	for _, item := range response.GetCustomers() {
		var apps []string
		appService := app.NewAppService("manage", client.DefaultClient)

		loggerx.InfoLog(c, ActionFindCustomers, fmt.Sprintf("Process FindApps:%s", loggerx.MsgProcessStarted))
		var appReq app.FindAppsRequest
		appReq.Domain = item.GetDomain()
		appReq.Database = item.GetCustomerId()
		appRes, err := appService.FindApps(context.TODO(), &appReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindCustomers, err)
			return
		}
		loggerx.InfoLog(c, ActionFindCustomers, fmt.Sprintf("Process FindApps:%s", loggerx.MsgProcessEnded))

		for _, v := range appRes.GetApps() {
			apps = append(apps, v.GetAppName())
		}

		comp := map[string]interface{}{}

		comp["apps"] = apps
		comp["customer_id"] = item.GetCustomerId()
		comp["customer_name"] = item.GetCustomerName()
		comp["domain"] = item.GetDomain()
		comp["logo"] = item.GetCustomerLogo()
		comp["created_at"] = item.GetCreatedAt()
		comp["created_by"] = item.GetCreatedBy()
		comp["updated_at"] = item.GetUpdatedAt()
		comp["updated_by"] = item.GetUpdatedBy()
		comp["deleted_at"] = item.GetDeletedAt()
		comp["deleted_by"] = item.GetDeletedBy()

		result = append(result, comp)
	}

	loggerx.InfoLog(c, ActionFindCustomers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionFindCustomers)),
		Data:    result,
	})
}

// FindCustomer 查找单个顾客记录
// @Router /customers/{customer_id} [get]
func (u *Customer) FindCustomer(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindCustomer, loggerx.MsgProcessStarted)

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	section := c.Query("section")

	if section == "domain" {
		var req customer.FindCustomerByDomainRequest
		req.Domain = c.Param("customer_id")
		response, err := customerService.FindCustomerByDomain(context.TODO(), &req)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindCustomer, err)
			return
		}

		loggerx.InfoLog(c, ActionFindCustomer, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionFindCustomer)),
			Data:    response.GetCustomer(),
		})
		return
	}

	var req customer.FindCustomerRequest
	req.CustomerId = c.Param("customer_id")
	response, err := customerService.FindCustomer(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindCustomer, err)
		return
	}

	loggerx.InfoLog(c, ActionFindCustomer, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionFindCustomer)),
		Data:    response.GetCustomer(),
	})
}

// AddCustomer 添加单个顾客记录
// @Router /customers [post]
func (u *Customer) AddCustomer(c *gin.Context) {
	loggerx.InfoLog(c, ActionAddCustomer, loggerx.MsgProcessStarted)

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.AddCustomerRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, err)
		return
	}

	req.Writer = sessionx.GetAuthUserID(c)

	response, err := customerService.AddCustomer(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddCustomer"))

	// TODO 语言添加
	languageService := language.NewLanguageService("global", client.DefaultClient)

	// 中文
	var langZhReq language.AddLanguageRequest
	langZhReq.Domain = req.GetDomain()
	langZhReq.LangCd = "zh-CN"
	langZhReq.Text = "简体中文"
	langZhReq.Abbr = "🇨🇳"
	langZhReq.Writer = sessionx.GetAuthUserID(c)
	langZhReq.Database = response.CustomerId

	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage zh:%s", loggerx.MsgProcessStarted))
	_, zher := languageService.AddLanguage(context.TODO(), &langZhReq)
	if zher != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, zher)
		return
	}
	loggerx.SuccessLog(c, ActionAddCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddLanguage zh"))
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage zh:%s", loggerx.MsgProcessEnded))
	// 英语
	var langEnReq language.AddLanguageRequest
	langEnReq.Domain = req.GetDomain()
	langEnReq.LangCd = "en-US"
	langEnReq.Text = "English"
	langEnReq.Abbr = "🇬🇧"
	langEnReq.Writer = sessionx.GetAuthUserID(c)
	langEnReq.Database = response.CustomerId

	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage en:%s", loggerx.MsgProcessStarted))
	_, ener := languageService.AddLanguage(context.TODO(), &langEnReq)
	if ener != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, ener)
		return
	}
	loggerx.SuccessLog(c, ActionAddCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddLanguage en"))
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage en:%s", loggerx.MsgProcessEnded))

	// 英语
	var langThReq language.AddLanguageRequest
	langThReq.Domain = req.GetDomain()
	langThReq.LangCd = "th-TH"
	langThReq.Text = "ไทย"
	langThReq.Abbr = "🇹🇭"
	langThReq.Writer = sessionx.GetAuthUserID(c)
	langThReq.Database = response.CustomerId

	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage th:%s", loggerx.MsgProcessStarted))
	_, ther := languageService.AddLanguage(context.TODO(), &langThReq)
	if ener != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, ther)
		return
	}
	loggerx.SuccessLog(c, ActionAddCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddLanguage th"))
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage th:%s", loggerx.MsgProcessEnded))
	// 日语
	var langJpReq language.AddLanguageRequest
	langJpReq.Domain = req.GetDomain()
	langJpReq.LangCd = "ja-JP"
	langJpReq.Text = "日本語"
	langJpReq.Abbr = "🇯🇵"
	langJpReq.Writer = sessionx.GetAuthUserID(c)
	langJpReq.Database = response.CustomerId

	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage ja:%s", loggerx.MsgProcessStarted))
	_, jper := languageService.AddLanguage(context.TODO(), &langJpReq)
	if jper != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, jper)
		return
	}
	loggerx.SuccessLog(c, ActionAddCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "AddLanguage ja"))
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process AddLanguage ja:%s", loggerx.MsgProcessEnded))

	// 创建history索引
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process CreateHistoryIndex :%s", loggerx.MsgProcessStarted))
	historyService := datahistory.NewHistoryService("database", client.DefaultClient)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	var ireq datahistory.CreateIndexRequest
	ireq.CustomerId = response.CustomerId

	_, ierr := historyService.CreateIndex(context.TODO(), &ireq, opss)
	if ierr != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, ierr)
		return
	}
	loggerx.SuccessLog(c, ActionAddCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "CreateHistoryIndex"))
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process CreateHistoryIndex :%s", loggerx.MsgProcessEnded))

	// 创建顾客文件桶
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process CreateFileBucket :%s", loggerx.MsgProcessStarted))
	_, err = storagecli.NewClient(req.GetDomain())
	if err != nil {
		httpx.GinHTTPError(c, ActionAddCustomer, err)
		return
	}
	loggerx.SuccessLog(c, ActionAddCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "CreateFileBucket"))
	loggerx.InfoLog(c, ActionAddCustomer, fmt.Sprintf("Process CreateFileBucket :%s", loggerx.MsgProcessEnded))

	loggerx.InfoLog(c, ActionAddCustomer, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionAddCustomer)),
		Data:    response,
	})
}

// ModifyCustomer 修改单个顾客记录
// @Router /customers/{customer_id} [put]
func (u *Customer) ModifyCustomer(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyCustomer, loggerx.MsgProcessStarted)

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	// 变更前查询顾客信息
	var freq customer.FindCustomerRequest
	freq.CustomerId = c.Param("customer_id")
	fresponse, err := customerService.FindCustomer(context.TODO(), &freq)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyCustomer, err)
		return
	}
	customerInfo := fresponse.GetCustomer()

	var req customer.ModifyCustomerRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionModifyCustomer, err)
		return
	}

	req.CustomerId = c.Param("customer_id")
	req.Writer = sessionx.GetAuthUserID(c)

	response, err := customerService.ModifyCustomer(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionModifyCustomer, err)
		return
	}
	loggerx.SuccessLog(c, ActionModifyCustomer, fmt.Sprintf(loggerx.MsgProcesSucceed, "ModifyCustomer"))

	// 变更成功后，比较变更的结果，记录日志
	// 比较顾客名称
	name := customerInfo.GetCustomerName()
	if name != req.GetCustomerName() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		params["company_name"] = req.GetCustomerName()

		loggerx.ProcessLog(c, ActionModifyCustomer, msg.L009, params)
	}

	// 比较二次验证是否变更
	secondCheck := customerInfo.GetSecondCheck()
	if strconv.FormatBool(secondCheck) != req.GetSecondCheck() {
		params := make(map[string]string)
		params["user_name"] = sessionx.GetUserName(c)
		if secondCheck {
			// 关闭了登录二次验证的日志
			loggerx.ProcessLog(c, ActionModifyCustomer, msg.L011, params)
		} else {
			// 打开了登录二次验证的日志
			loggerx.ProcessLog(c, ActionModifyCustomer, msg.L010, params)
		}
	}

	loggerx.InfoLog(c, ActionModifyCustomer, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionModifyCustomer)),
		Data:    response,
	})
}

// DeleteSelectCustomers 删除选中顾客记录
// @Router /customers [delete]
func (u *Customer) DeleteSelectCustomers(c *gin.Context) {
	loggerx.InfoLog(c, ActionDeleteSelectCustomers, loggerx.MsgProcessStarted)

	var req customer.DeleteSelectCustomersRequest
	req.CustomerIdList = c.QueryArray("customer_id_list")
	req.Writer = sessionx.GetAuthUserID(c)

	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	response, err := customerService.DeleteSelectCustomers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionDeleteSelectCustomers, err)
		return
	}
	loggerx.SuccessLog(c, ActionDeleteSelectCustomers, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionDeleteSelectCustomers))

	loggerx.InfoLog(c, ActionDeleteSelectCustomers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionDeleteSelectCustomers)),
		Data:    response,
	})
}

// HardDeleteCustomers 物理删除客户
// @Router /phydel/customers [delete]
func (u *Customer) HardDeleteCustomers(c *gin.Context) {
	loggerx.InfoLog(c, ActionHardDeleteCustomers, loggerx.MsgProcessStarted)
	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.HardDeleteCustomersRequest
	req.CustomerIdList = c.QueryArray("customer_id_list")
	var domains []string
	for _, cid := range req.GetCustomerIdList() {
		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}
		var req customer.FindCustomerRequest
		req.CustomerId = cid
		response, err := customerService.FindCustomer(context.TODO(), &req, opss)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteCustomers, err)
			return
		}
		domains = append(domains, response.GetCustomer().GetDomain())
	}

	response, err := customerService.HardDeleteCustomers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionHardDeleteCustomers, err)
		return
	}
	loggerx.SuccessLog(c, ActionHardDeleteCustomers, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionHardDeleteCustomers))

	// 删除顾客对应的文件
	for _, domain := range domains {
		minioClient, err := storagecli.NewClient(domain)
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteCustomers, err)
			return
		}

		err = minioClient.DeleteBucket()
		if err != nil {
			httpx.GinHTTPError(c, ActionHardDeleteCustomers, err)
			return
		}
	}

	loggerx.InfoLog(c, ActionHardDeleteCustomers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I006, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionHardDeleteCustomers)),
		Data:    response,
	})
}

// RecoverSelectCustomers 恢复选中顾客记录
// @Router /recover/customers [PUT]
func (u *Customer) RecoverSelectCustomers(c *gin.Context) {
	loggerx.InfoLog(c, ActionRecoverSelectCustomers, loggerx.MsgProcessStarted)

	var req customer.RecoverSelectCustomersRequest

	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectCustomers, err)
		return
	}
	// 从共通中获取参数
	req.Writer = sessionx.GetAuthUserID(c)

	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	response, err := customerService.RecoverSelectCustomers(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionRecoverSelectCustomers, err)
		return
	}
	loggerx.SuccessLog(c, ActionRecoverSelectCustomers, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionRecoverSelectCustomers))

	loggerx.InfoLog(c, ActionRecoverSelectCustomers, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I013, fmt.Sprintf(httpx.Temp, CustomerProcessName, ActionRecoverSelectCustomers)),
		Data:    response,
	})
}
