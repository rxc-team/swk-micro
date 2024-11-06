package itemx

import (
	"context"
	"time"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/database/proto/item"
)

// 获取关联台账数据
func GetLookupItems(db, datastoreID, appID string, accessKey []string) []*item.Item {
	ct := grpc.NewClient(
		grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
	)

	itemService := item.NewItemService("database", ct)

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
		o.DialTimeout = time.Minute * 10
	}

	var req item.ItemsRequest
	req.DatastoreId = datastoreID
	req.ConditionType = "and"
	req.AppId = appID
	req.Owners = accessKey
	req.IsOrigin = true
	req.Database = db

	response, err := itemService.FindItems(context.TODO(), &req, opss)
	if err != nil {
		loggerx.ErrorLog("getLookupItems", err.Error())
		return nil
	}

	return response.GetItems()
}
