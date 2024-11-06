package admin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
)

// Customer 顾客
type Customer struct{}

// log出力
const (
	CustomerProcessName  = "Customer"
	ActionFindCustomers  = "FindCustomers"
	ActionFindCustomer   = "FindCustomer"
	ActionModifyCustomer = "ModifyCustomer"
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

// ModifyCustomer 修改单个顾客记录
// @Router /customers/{customer_id} [put]
func (u *Customer) ModifyCustomer(c *gin.Context) {
	loggerx.InfoLog(c, ActionModifyCustomer, loggerx.MsgProcessStarted)

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	// 变更前查询顾客信息
	var freq customer.FindCustomerRequest
	freq.CustomerId = c.Param("customer_id")

	// admin只能改自己公司
	if freq.CustomerId != sessionx.GetUserCustomer(c) {
		c.JSON(403, gin.H{
			"message": msg.GetMsg("ja-JP", msg.Error, msg.E007),
		})
		c.Abort()
		return
	}

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
