package webui

import (
	"context"

	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/srv/database/proto/datastore"
)

// 获取台账信息
func getDatastoreInfo(db, appID, ds string) (*datastore.Datastore, error) {
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoreKeyRequest
	// 从path获取
	req.ApiKey = ds
	req.AppId = appID
	req.Database = db
	response, err := datastoreService.FindDatastoreByKey(context.TODO(), &req)
	if err != nil {
		return nil, err
	}
	return response.GetDatastore(), nil
}
