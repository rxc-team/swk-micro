package handler

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/copy"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Copy 台账数据
type Copy struct{}

// log出力使用
const (
	CopyProcessName = "Copy"
	ActionCopyItems = "CopyItems"
)

// CopyItems 复制数据
func (i *Copy) CopyItems(ctx context.Context, req *copy.CopyItemsRequest, rsq *copy.CopyItemsResponse) error {
	utils.InfoLog(ActionCopyItems, utils.MsgProcessStarted)

	err := model.CopyItems(req.GetDatabase(), req.GetAppId(), req.GetCopyAppId(), req.GetDatastoreId(), req.GetCopyDatastoreId(), req.GetWithFile())
	if err != nil {
		utils.ErrorLog(ActionCopyItems, err.Error())
		return err
	}

	utils.InfoLog(ActionCopyItems, utils.MsgProcessEnded)

	return nil
}

// BulkAddItems 批量添加数据
func (i *Copy) BulkAddItems(ctx context.Context, req *copy.BulkAddRequest, rsp *copy.BulkAddResponse) error {
	utils.InfoLog(ActionCopyItems, utils.MsgProcessStarted)

	// var items []*model.Item
	var items []mongo.WriteModel

	for _, it := range req.GetItems() {
		itemMap := make(model.ItemMap)

		for k, v := range it.Items {
			itemMap[k] = &model.Value{
				DataType: v.DataType,
				Value:    model.GetCopyeValueFromProto(v),
			}
		}

		id := primitive.NewObjectID()

		now := time.Now()

		doc := model.Item{
			ID:          id,
			ItemID:      id.Hex(),
			AppID:       req.GetAppId(),
			DatastoreID: req.GetDatastoreId(),
			ItemMap:     itemMap,
			Owners:      req.GetOwners(),
			CheckStatus: "0",
			CreatedAt:   now,
			CreatedBy:   req.GetWriter(),
			UpdatedAt:   now,
			UpdatedBy:   req.GetWriter(),
			Status:      "1",
		}

		md := mongo.NewInsertOneModel()
		md.SetDocument(doc)

		items = append(items, md)
	}

	err := model.BulkAddItems(req.GetDatabase(), req.GetDatastoreId(), items)
	if err != nil {
		utils.ErrorLog(ActionCopyItems, err.Error())
		return err
	}

	utils.InfoLog(ActionCopyItems, utils.MsgProcessEnded)

	return nil
}
