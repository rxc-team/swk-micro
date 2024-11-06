package filex

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
)

// CheckCanUpload 判断已上传文件是否已超过顾客的最大存储空间
func CheckCanUpload(domain string, fileSize float64) (canUpload bool) {
	// 如果是超级管理员用户的情况下，不教验大小。
	if domain == "proship.co.jp" {
		return true
	}

	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	var req customer.FindCustomerByDomainRequest
	req.Domain = domain
	response, err := customerService.FindCustomerByDomain(context.TODO(), &req)
	if err != nil {
		return false
	}

	maxSize := response.GetCustomer().GetMaxSize() * 1024 * 1024 * 1024
	usedSize := response.GetCustomer().GetUsedSize() + fileSize

	return maxSize >= usedSize
}

// ModifyUsedSize 上传文件成功后，修改顾客的已使用存储空间的大小
func ModifyUsedSize(domain string, fileSize float64) (err error) {

	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.ModifyUsedSizeRequest

	req.Domain = domain
	req.UsedSize = fileSize

	_, err = customerService.ModifyUsedSize(context.TODO(), &req)
	if err != nil {
		return err
	}

	return nil
}
