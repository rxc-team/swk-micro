package mappingx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
)

// 获取mapping信息
func GetMappingInfo(db, datastoreID, mappingID string) (*datastore.MappingConf, error) {
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.MappingRequest
	req.DatastoreId = datastoreID
	req.MappingId = mappingID
	req.Database = db

	response, err := datastoreService.FindDatastoreMapping(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getMappingInfo", err.Error())
		return nil, err
	}

	return response.GetMapping(), nil
}
