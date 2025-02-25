package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

// deleteContractItem 删除租赁契约台账数据
func deleteContractItem(db, appId, datastoreID, itemID, userID, lang, domain string, owners []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(datastoreID))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 获取所有台账
	dsList, e := FindDatastores(db, appId, "", "", "")
	if e != nil {
		if e.Error() == mongo.ErrNoDocuments.Error() {
			dsList = []Datastore{}
		} else {
			utils.ErrorLog("deleteContractItem", e.Error())
			return e
		}
	}

	dsMap := make(map[string]string)
	for _, d := range dsList {
		dsMap[d.ApiKey] = d.DatastoreID
	}

	paymentStatus := dsMap["paymentStatus"]
	paymentInterest := dsMap["paymentInterest"]
	repayment := dsMap["repayment"]
	rireki := dsMap["rireki"]

	cp := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(paymentStatus))
	ci := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(paymentInterest))
	cr := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(repayment))
	ck := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(rireki))

	// 开启事务删除所有契约关联数据
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("deleteContractItem", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("deleteContractItem", err.Error())
		return err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		// 删除对象的契约数据
		oldItem, err := getItem(db, itemID, datastoreID, owners)
		if err != nil {
			if err.Error() == mongo.ErrNoDocuments.Error() {
				return errors.New("データが存在しないか、データを削除する権限がありません")
			}
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}
		// 删除对象的契约番号
		keiyakuno := GetValueFromModel(oldItem.ItemMap["keiyakuno"])
		query := bson.M{
			"items.keiyakuno.value": keiyakuno,
		}

		// 删除对象的支付数据
		if _, err := cp.DeleteMany(sc, query); err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		// 删除对象的试算数据
		if _, err := ci.DeleteMany(sc, query); err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		// 删除对象的偿还数据
		if _, err := cr.DeleteMany(sc, query); err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		// 删除对象的履历数据
		if _, err := ck.DeleteMany(sc, query); err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		hs := NewHistory(db, userID, datastoreID, lang, domain, sc, nil)

		err = hs.Add("1", itemID, oldItem.ItemMap)
		if err != nil {
			utils.ErrorLog("ModifyItem", err.Error())
			return err
		}

		objectID, err := primitive.ObjectIDFromHex(itemID)
		if err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		query1 := bson.M{
			"_id": objectID,
		}

		queryJSON, _ := json.Marshal(query1)
		utils.DebugLog("deleteContractItem", fmt.Sprintf("query: [ %s ]", queryJSON))

		if _, err := c.DeleteOne(sc, query1); err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		err = hs.Compare("1", nil)
		if err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("deleteContractItem", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				utils.ErrorLog("deleteContractItem", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("deleteContractItem", err.Error())
		return err
	}

	session.EndSession(ctx)
	return nil
}
