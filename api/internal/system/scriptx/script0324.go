package scriptx

import (
	"context"
	"io"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/script"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// 修改minio和db中台账文件字段的文件的路径
type Script0324 struct{}

const ()

// FileValue 文件类型
type FileValue struct {
	URL  string `json:"url" bson:"url"`
	Name string `json:"name" bson:"name"`
}

func (s *Script0324) Run() error {
	go updateFileUrl()
	return nil
}

func updateFileUrl() {
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	// 获取顾客信息
	var customerReq customer.FindCustomersRequest
	customerRes, err := customerService.FindCustomers(context.TODO(), &customerReq)
	if err != nil {
		loggerx.ErrorLog("Script0324", err.Error())
		return
	}
	for _, customer := range customerRes.GetCustomers() {
		// 获取顾客app信息
		appService := app.NewAppService("manage", client.DefaultClient)
		var appReq app.FindAppsRequest
		appReq.Database = customer.GetCustomerId()
		appReq.Domain = customer.GetDomain()
		appRes, err := appService.FindApps(context.TODO(), &appReq)
		if err != nil {
			loggerx.ErrorLog("Script0324", err.Error())
			return
		}
		minioClient, err := storagecli.NewClient(appReq.Domain)
		if err != nil {
			loggerx.ErrorLog("Script0324", err.Error())
			return
		}
		// oldURL集合
		delSet := make(map[string]struct{})
		for _, app := range appRes.GetApps() {
			// 获取台账
			datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
			var datastoreReq datastore.DatastoresRequest
			datastoreReq.AppId = app.GetAppId()
			datastoreReq.Database = appReq.Database
			datastoreRes, err := datastoreService.FindDatastores(context.TODO(), &datastoreReq)
			if err != nil {
				loggerx.ErrorLog("Script0324", err.Error())
				return
			}

			for _, datastore := range datastoreRes.GetDatastores() {
				// 查找并更新数据
				itemService := item.NewItemService("database", client.DefaultClient)
				var fileReq item.FindRequest
				fileReq.AppId = app.GetAppId()
				fileReq.DatastoreId = datastore.GetDatastoreId()
				fileReq.Database = appReq.Database
				stream, err := itemService.FindAndModifyFile(context.TODO(), &fileReq)
				if err != nil {
					loggerx.ErrorLog("Script0324", err.Error())
					return
				}
				for {
					resp, err := stream.Recv()
					if err != nil {
						if err == io.EOF {
							break
						} else {
							loggerx.ErrorLog("Script0324", err.Error())
							return
						}
					}

					// 拷贝文件对象
					_, err = minioClient.CopyObject(resp.GetOldUrl(), resp.GetNewUrl())
					if err != nil {
						loggerx.ErrorLog("Script0324", err.Error())
						return
					}
					delSet[resp.GetOldUrl()] = struct{}{}
				}
				stream.Close()
			}
		}
		// 删除多余的文件
		for k := range delSet {
			err = minioClient.DeleteObject(k)
			if err != nil {
				loggerx.ErrorLog("Script0324", err.Error())
				return
			}
		}
	}
	// 发送消息，通知datapatch完成
	scriptService := script.NewScriptService("manage", client.DefaultClient)

	var scriptReq script.FindScriptJobRequest
	scriptReq.ScriptId = "0324"
	scriptReq.Database = "system"

	scriptRes, err := scriptService.FindScriptJob(context.TODO(), &scriptReq)
	if err != nil {
		loggerx.ErrorLog("Script0324", err.Error())
		return
	}
	param := wsx.MessageParam{
		Sender:    "system",
		Recipient: scriptRes.GetScriptJob().GetRanBy(),
		Domain:    "proship.co.jp",
		MsgType:   "datapatch",
		Content:   "datapatch successed",
		Status:    "unread",
	}
	wsx.SendToUser(param)

}
