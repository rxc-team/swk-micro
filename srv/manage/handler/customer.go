package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Customer 顾客
type Customer struct{}

// log出力使用
const (
	CustomerProcessName = "Customer"

	ActionFindCustomers            = "FindCustomers"
	ActionFindCustomer             = "FindCustomer"
	ActionFindCustomerByDomain     = "FindCustomerByDomain"
	ActionAddCustomer              = "AddCustomer"
	ActionModifyCustomer           = "ModifyCustomer"
	ActionModifyUsedSizeOfCustomer = "ModifyUsedSizeOfCustomer"
	ActionDeleteCustomer           = "DeleteCustomer"
	ActionDeleteSelectCustomers    = "DeleteSelectCustomers"
	ActionHardDeleteCustomers      = "HardDeleteCustomers"
	ActionRecoverSelectCustomers   = "RecoverSelectCustomers"
)

// FindCustomers 查找多个顾客记录
func (g *Customer) FindCustomers(ctx context.Context, req *customer.FindCustomersRequest, rsp *customer.FindCustomersResponse) error {
	utils.InfoLog(ActionFindCustomers, utils.MsgProcessStarted)

	customers, err := model.FindCustomers(ctx, req.CustomerName, req.Domain, req.InvalidatedIn)
	if err != nil {
		utils.ErrorLog(ActionFindCustomers, err.Error())
		return err
	}

	res := &customer.FindCustomersResponse{}
	for _, r := range customers {
		res.Customers = append(res.Customers, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindCustomers, utils.MsgProcessEnded)
	return nil
}

// FindCustomer 查找单个顾客记录
func (g *Customer) FindCustomer(ctx context.Context, req *customer.FindCustomerRequest, rsp *customer.FindCustomerResponse) error {
	utils.InfoLog(ActionFindCustomer, utils.MsgProcessStarted)

	res, err := model.FindCustomer(ctx, req.CustomerId)
	if err != nil {
		utils.ErrorLog(ActionFindCustomer, err.Error())
		return err
	}

	rsp.Customer = res.ToProto()

	utils.InfoLog(ActionFindCustomer, utils.MsgProcessEnded)
	return nil
}

// FindCustomerByDomain 通过域名查找单个顾客记录
func (g *Customer) FindCustomerByDomain(ctx context.Context, req *customer.FindCustomerByDomainRequest, rsp *customer.FindCustomerByDomainResponse) error {
	utils.InfoLog(ActionFindCustomerByDomain, utils.MsgProcessStarted)

	res, err := model.FindCustomerByDomain(ctx, req.Domain)
	if err != nil {
		utils.ErrorLog(ActionFindCustomerByDomain, err.Error())
		return err
	}

	rsp.Customer = res.ToProto()

	utils.InfoLog(ActionFindCustomerByDomain, utils.MsgProcessEnded)
	return nil
}

// AddCustomer 添加单个顾客记录
func (g *Customer) AddCustomer(ctx context.Context, req *customer.AddCustomerRequest, rsp *customer.AddCustomerResponse) error {
	utils.InfoLog(ActionAddCustomer, utils.MsgProcessStarted)

	params := model.Customer{
		CustomerName:     req.GetCustomerName(),
		SecondCheck:      req.GetSecondCheck(),
		CustomerLogo:     req.GetCustomerLogo(),
		Domain:           req.GetDomain(),
		DefaultUser:      req.GetDefaultUser(),
		DefaultUserEmail: req.GetDefaultUserEmail(),
		DefaultTimezone:  req.GetDefaultTimezone(),
		DefaultLanguage:  req.GetDefaultLanguage(),
		MaxUsers:         req.GetMaxUsers(),
		UsedUsers:        req.GetUsedUsers(),
		MaxSize:          req.GetMaxSize(),
		MaxDataSize:      req.GetMaxDataSize(),
		Level:            req.GetLevel(),
		UploadFileSize:   req.GetUploadFileSize(),
		CreatedAt:        time.Now(),
		CreatedBy:        req.GetWriter(),
		UpdatedAt:        time.Now(),
		UpdatedBy:        req.GetWriter(),
	}

	id, err := model.AddCustomer(ctx, &params)
	if err != nil {
		utils.ErrorLog(ActionAddCustomer, err.Error())
		return err
	}

	rsp.CustomerId = id

	utils.InfoLog(ActionAddCustomer, utils.MsgProcessEnded)
	return nil
}

// ModifyCustomer 修改单个顾客记录
func (g *Customer) ModifyCustomer(ctx context.Context, req *customer.ModifyCustomerRequest, rsp *customer.ModifyCustomerResponse) error {
	utils.InfoLog(ActionModifyCustomer, utils.MsgProcessStarted)

	params := model.UpdateParam{
		CustomerID:       req.GetCustomerId(),
		CustomerName:     req.GetCustomerName(),
		DefaultUserEmail: req.GetDefaultUserEmail(),
		MaxSize:          req.GetMaxSize(),
		SecondCheck:      req.GetSecondCheck(),
		CustomerLogo:     req.GetCustomerLogo(),
		MaxUsers:         req.GetMaxUsers(),
		MaxDataSize:      req.GetMaxDataSize(),
		Level:            req.GetLevel(),
		UploadFileSize:   req.GetUploadFileSize(),
	}

	err := model.ModifyCustomer(ctx, &params, req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionModifyCustomer, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyCustomer, utils.MsgProcessEnded)
	return nil
}

// ModifyUsedSizeOfCustomer 修改顾客的已使用的存储空间大小
func (g *Customer) ModifyUsedUsers(ctx context.Context, req *customer.ModifyUsedUsersRequest, rsp *customer.ModifyUsedUsersResponse) error {
	utils.InfoLog(ActionModifyUsedSizeOfCustomer, utils.MsgProcessStarted)

	err := model.ModifyUsedUsers(ctx, req.GetCustomerId(), req.GetUsedUsers())
	if err != nil {
		utils.ErrorLog(ActionModifyUsedSizeOfCustomer, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyUsedSizeOfCustomer, utils.MsgProcessEnded)
	return nil
}

// ModifyUsedSizeOfCustomer 修改顾客的已使用的存储空间大小
func (g *Customer) ModifyUsedSize(ctx context.Context, req *customer.ModifyUsedSizeRequest, rsp *customer.ModifyUsedSizeResponse) error {
	utils.InfoLog(ActionModifyUsedSizeOfCustomer, utils.MsgProcessStarted)

	err := model.ModifyUsedSizeOfCustomer(ctx, req.GetDomain(), req.GetUsedSize())
	if err != nil {
		utils.ErrorLog(ActionModifyUsedSizeOfCustomer, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyUsedSizeOfCustomer, utils.MsgProcessEnded)
	return nil
}

// DeleteCustomer 删除单个顾客记录
func (g *Customer) DeleteCustomer(ctx context.Context, req *customer.DeleteCustomerRequest, rsp *customer.DeleteCustomerResponse) error {
	utils.InfoLog(ActionDeleteCustomer, utils.MsgProcessStarted)

	err := model.DeleteCustomer(ctx, req.GetCustomerId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteCustomer, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteCustomer, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectCustomers 删除多个顾客记录
func (g *Customer) DeleteSelectCustomers(ctx context.Context, req *customer.DeleteSelectCustomersRequest, rsp *customer.DeleteSelectCustomersResponse) error {
	utils.InfoLog(ActionDeleteSelectCustomers, utils.MsgProcessStarted)

	err := model.DeleteSelectCustomers(ctx, req.GetCustomerIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectCustomers, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectCustomers, utils.MsgProcessEnded)
	return nil
}

// HardDeleteCustomers 物理删除选中客户
func (g *Customer) HardDeleteCustomers(ctx context.Context, req *customer.HardDeleteCustomersRequest, rsp *customer.HardDeleteCustomersResponse) error {
	utils.InfoLog(ActionHardDeleteCustomers, utils.MsgProcessStarted)

	err := model.HardDeleteCustomers(ctx, req.GetCustomerIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteCustomers, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteCustomers, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectCustomers 恢复选中顾客记录
func (g *Customer) RecoverSelectCustomers(ctx context.Context, req *customer.RecoverSelectCustomersRequest, rsp *customer.RecoverSelectCustomersResponse) error {
	utils.InfoLog(ActionRecoverSelectCustomers, utils.MsgProcessStarted)

	err := model.RecoverSelectCustomers(ctx, req.GetCustomerIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectCustomers, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectCustomers, utils.MsgProcessEnded)
	return nil
}
