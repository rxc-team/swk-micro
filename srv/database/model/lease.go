package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/srv/database/utils"
	database "rxcsoft.cn/utils/mongo"
)

// ModifyContract 契约情报变更
func ModifyContract(db, collection string, p *ItemUpdateParam) (err error) {
	// 契约台账履历表取得
	dsrireki, err := FindDatastoreByKey(db, p.AppID, "rireki")
	if err != nil {
		utils.ErrorLog("ModifyContract", err.Error())
		return err
	}

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(p.DatastoreID))
	cr := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsrireki.DatastoreID))
	ct := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 事务处理开始
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("ModifyContract", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("ModifyContract", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// 查找自增字段
		param := &FindFieldsParam{
			AppID:       p.AppID,
			DatastoreID: p.DatastoreID,
		}
		allFields, err := FindFields(db, param)
		if err != nil {
			utils.ErrorLog("ModifyContract", err.Error())
			if err.Error() != mongo.ErrNoDocuments.Error() {
				return err
			}
		}

		// 变更前契约情报取得
		oldItem, e := getItem(db, p.ItemID, p.DatastoreID, p.Owners)
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				return errors.New("データが存在しないか、あなたまたはデータの申請者はデータを変更する権限がありません")
			}
			utils.ErrorLog("ModifyContract", e.Error())
			return e
		}

		var changeCount = 0

		// 自增字段不更新
		if len(allFields) > 0 {
			for _, f := range allFields {
				if f.FieldType == "autonum" {
					delete(oldItem.ItemMap, f.FieldID)
					delete(p.ItemMap, f.FieldID)
					continue
				}
				_, ok := p.ItemMap[f.FieldID]
				// 需要进行自算的情况
				if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
					if f.SelfCalculate == "add" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o + n
						continue
					}
					if f.SelfCalculate == "sub" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o - n
						continue
					}
				}
			}
		}

		hs := NewHistory(db, p.UpdatedBy, p.DatastoreID, p.Lang, p.Domain, sc, allFields)

		err = hs.Add("1", p.ItemID, oldItem.ItemMap)
		if err != nil {
			utils.ErrorLog("ModifyContract", err.Error())
			return err
		}

		// 变更后契约履历情报
		newItemMap := make(map[string]*Value, len(p.ItemMap)+12)
		// 变更前契约履历情报
		oldItemMap := make(map[string]*Value, len(p.ItemMap)+12)

		// 临时数据ID
		templateID := p.ItemMap["template_id"].Value.(string)
		// 删除临时数据ID
		delete(p.ItemMap, "template_id")

		// 循环取得变更前契约情报数据
		for key, value := range oldItem.ItemMap {
			oldItemMap[key] = value
		}
		// 契约变更情报编辑
		item := bson.M{
			"updated_at": p.UpdatedAt,
			"updated_by": p.UpdatedBy,
		}
		// 循环契约情报数据对比变更
		for key, value := range p.ItemMap {
			// 记录履历数据情报(包含未变更数据项)
			newItemMap[key] = value
			item["items."+key] = value
			// 对比前后数据值
			if _, ok := oldItem.ItemMap[key]; ok {
				// 该项数据历史存在,判断历史与当前是否变更
				if compare(value, oldItem.ItemMap[key]) {
					changeCount++
				}
			} else {
				// 该项数据历史不存在,判断当前是否为空
				if value.Value == "" || value.Value == "[]" {
					continue
				}

				changeCount++
			}
		}

		oldItemMap["henkouymd"] = newItemMap["henkouymd"]

		// 契约情报变更参数编辑
		update := bson.M{"$set": item}
		objectID, e := primitive.ObjectIDFromHex(p.ItemID)
		if e != nil {
			utils.ErrorLog("ModifyContract", e.Error())
			return e
		}
		query := bson.M{
			"_id": objectID,
		}
		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("ModifyContract", fmt.Sprintf("query: [ %s ]", queryJSON))
		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("ModifyContract", fmt.Sprintf("update: [ %s ]", updateJSON))

		// 第一步：更新契约台账情报
		if _, err := c.UpdateOne(sc, query, update); err != nil {
			utils.ErrorLog("ModifyContract", err.Error())
			return err
		}

		err = hs.Compare("1", p.ItemMap)
		if err != nil {
			utils.ErrorLog("ModifyContract", err.Error())
			return err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("ModifyContract", err.Error())
			return err
		}

		// 第二步：若字段有变更,添加契约历史情报和新旧履历情报
		if changeCount > 0 {
			// 契約履歴番号
			rirekiSeq, err := uuid.NewUUID()
			if err != nil {
				utils.ErrorLog("ModifyContract", err.Error())
				return err
			}
			newItemMap["no"] = &Value{
				DataType: "text",
				Value:    rirekiSeq.String(),
			}
			oldItemMap["no"] = &Value{
				DataType: "text",
				Value:    rirekiSeq.String(),
			}
			// 修正区分编辑
			newItemMap["zengokbn"] = &Value{
				DataType: "options",
				Value:    "after",
			}
			oldItemMap["zengokbn"] = &Value{
				DataType: "options",
				Value:    "before",
			}
			// 对接区分编辑
			newItemMap["dockkbn"] = &Value{
				DataType: "options",
				Value:    "undo",
			}
			oldItemMap["dockkbn"] = &Value{
				DataType: "options",
				Value:    "undo",
			}
			// 操作区分编辑
			newItemMap["actkbn"] = &Value{
				DataType: "options",
				Value:    "infoalter",
			}
			oldItemMap["actkbn"] = &Value{
				DataType: "options",
				Value:    "infoalter",
			}

			queryTmp := bson.M{
				"template_id":   templateID,
				"datastore_key": "rireki",
			}
			var rirekiTmp TemplateItem

			if err := ct.FindOne(ctx, queryTmp).Decode(&rirekiTmp); err != nil {
				utils.ErrorLog("ModifyContract", err.Error())
				return err
			}
			// リース開始日から変更年月日までの償却費の累計額
			newItemMap["oldDepreciationTotal"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["oldDepreciationTotal"].Value,
			}
			oldItemMap["oldDepreciationTotal"] = &Value{
				DataType: "number",
				Value:    0,
			}
			newItemMap["payTotalRemain"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["payTotalRemain"].Value,
			}
			oldItemMap["payTotalRemain"] = &Value{
				DataType: "number",
				Value:    0,
			}
			newItemMap["interestTotalRemain"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["interestTotalRemain"].Value,
			}
			oldItemMap["interestTotalRemain"] = &Value{
				DataType: "number",
				Value:    0,
			}

			// 将契约番号变成lookup类型
			keiyakuno := p.ItemMap["keiyakuno"]

			itMap := make(map[string]interface{})
			for key, val := range p.ItemMap {
				itMap[key] = val
			}

			keiyakuItem := keiyakuno.Value.(string)

			newItemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			oldItemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			// 添加新契约履历情报
			var newRirekiItem Item
			newRirekiItem.ID = primitive.NewObjectID()
			newRirekiItem.ItemID = newRirekiItem.ID.Hex()
			newRirekiItem.AppID = dsrireki.AppID
			newRirekiItem.DatastoreID = dsrireki.DatastoreID
			newRirekiItem.CreatedAt = p.UpdatedAt
			newRirekiItem.CreatedBy = p.UpdatedBy
			newRirekiItem.ItemMap = newItemMap
			newRirekiItem.Owners = oldItem.Owners

			queryJSON, _ = json.Marshal(newRirekiItem)
			utils.DebugLog("ModifyContract", fmt.Sprintf("newRirekiItem: [ %s ]", queryJSON))
			if _, err := cr.InsertOne(sc, newRirekiItem); err != nil {
				utils.ErrorLog("ModifyContract", err.Error())
				return err
			}
			// 添加旧契约履历情报
			var oldRirekiItem Item
			oldRirekiItem.ID = primitive.NewObjectID()
			oldRirekiItem.ItemID = oldRirekiItem.ID.Hex()
			oldRirekiItem.AppID = dsrireki.AppID
			oldRirekiItem.DatastoreID = dsrireki.DatastoreID
			oldRirekiItem.CreatedAt = p.UpdatedAt
			oldRirekiItem.CreatedBy = p.UpdatedBy
			oldRirekiItem.ItemMap = oldItemMap
			oldRirekiItem.Owners = oldItem.Owners

			queryJSON, _ = json.Marshal(oldRirekiItem)
			utils.DebugLog("ModifyContract", fmt.Sprintf("oldRirekiItem: [ %s ]", queryJSON))
			if _, err := cr.InsertOne(sc, oldRirekiItem); err != nil {
				utils.ErrorLog("ModifyContract", err.Error())
				return err
			}
		}

		// 删除临时数据
		tc := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
		query1 := bson.M{
			"template_id": templateID,
		}

		_, err = tc.DeleteMany(sc, query1)
		if err != nil {
			utils.ErrorLog("ModifyContract", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("ModifyContract", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("ModifyContract", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// ChangeDebt 债务变更
func ChangeDebt(db, collection string, p *ItemUpdateParam) (err error) {

	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(p.DatastoreID))
	ct := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 事务处理开始
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("ChangeDebt", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("ChangeDebt", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		// 获取所有台账
		dsList, e := FindDatastores(db, p.AppID, "", "", "")
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				dsList = []Datastore{}
			} else {
				utils.ErrorLog("ChangeDebt", e.Error())
				return err
			}
		}

		dsMap := make(map[string]string)
		for _, d := range dsList {
			dsMap[d.ApiKey] = d.DatastoreID
		}

		// 契约台账履历表取得
		dsrireki := dsMap["rireki"]
		dsPay := dsMap["paymentStatus"]
		dsInterest := dsMap["paymentInterest"]
		dsRepay := dsMap["repayment"]

		cr := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsrireki))
		cpay := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsPay))
		cinter := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsInterest))
		crepay := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsRepay))

		// 查找自增字段
		param := &FindFieldsParam{
			AppID:       p.AppID,
			DatastoreID: p.DatastoreID,
		}

		allFields, err := FindFields(db, param)
		if err != nil {
			utils.ErrorLog("ChangeDebt", err.Error())
			if err.Error() != mongo.ErrNoDocuments.Error() {
				return err
			}
		}

		// 变更前契约情报取得
		oldItem, e := getItem(db, p.ItemID, p.DatastoreID, p.Owners)
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				return errors.New("データが存在しないか、あなたまたはデータの申請者はデータを変更する権限がありません")
			}
			utils.ErrorLog("ChangeDebt", e.Error())
			return e
		}

		hs := NewHistory(db, p.UpdatedBy, p.DatastoreID, p.Lang, p.Domain, sc, allFields)

		err = hs.Add("1", p.ItemID, oldItem.ItemMap)
		if err != nil {
			utils.ErrorLog("ChangeDebt", err.Error())
			return err
		}

		// 变更箇所数记录用
		changeCount := 0

		// 自增字段不更新
		if len(allFields) > 0 {
			for _, f := range allFields {
				if f.FieldType == "autonum" {
					delete(oldItem.ItemMap, f.FieldID)
					delete(p.ItemMap, f.FieldID)
					continue
				}
				_, ok := p.ItemMap[f.FieldID]
				// 需要进行自算的情况
				if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
					if f.SelfCalculate == "add" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o + n
						continue
					}
					if f.SelfCalculate == "sub" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o - n
						continue
					}
				}
			}
		}

		// 变更后契约履历情报
		newItemMap := make(map[string]*Value, len(p.ItemMap)+12)
		// 变更前契约履历情报
		oldItemMap := make(map[string]*Value, len(p.ItemMap)+12)

		// 临时数据ID
		templateID := p.ItemMap["template_id"].Value.(string)
		// 删除临时数据ID
		delete(p.ItemMap, "template_id")

		// 循环取得变更前契约情报数据
		for key, value := range oldItem.ItemMap {
			oldItemMap[key] = value
		}

		// 租赁满了日算出
		leasestymd := GetValueFromModel(p.ItemMap["leasestymd"])[:10]
		leasekikan := GetValueFromModel(p.ItemMap["leasekikan"])
		extentionOption := GetValueFromModel(p.ItemMap["extentionOption"])
		expireymd, err := GetExpireymd(leasestymd, leasekikan, extentionOption)
		if err != nil {
			utils.ErrorLog("ChangeDebt", err.Error())
			return err
		}
		// 租赁满了日
		p.ItemMap["leaseexpireymd"] = &Value{
			DataType: "date",
			Value:    expireymd,
		}

		// 契约变更情报编辑
		item := bson.M{
			"updated_at": p.UpdatedAt,
			"updated_by": p.UpdatedBy,
		}

		// 循环契约情报数据对比变更
		for key, value := range p.ItemMap {
			// 记录履历数据情报(包含未变更数据项)
			newItemMap[key] = value

			// 项目[百分比]不做更新
			if key != "percentage" {
				item["items."+key] = value
			}

			// 对比前后数据值
			if _, ok := oldItem.ItemMap[key]; ok {
				// 该项数据历史存在,判断历史与当前是否变更
				if compare(value, oldItem.ItemMap[key]) {
					changeCount++
				}
			} else {
				// 该项数据历史不存在,判断当前是否为空
				if value.Value == "" || value.Value == "[]" {
					continue
				}
				changeCount++
			}
		}

		// 契约情报变更参数编辑
		update := bson.M{"$set": item}
		objectID, e := primitive.ObjectIDFromHex(p.ItemID)
		if e != nil {
			utils.ErrorLog("ChangeDebt", e.Error())
			return e
		}
		query := bson.M{
			"_id": objectID,
		}
		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("ModifyItem", fmt.Sprintf("query: [ %s ]", queryJSON))
		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("ModifyItem", fmt.Sprintf("update: [ %s ]", updateJSON))

		// 第一步：更新契约台账情报
		if _, err := c.UpdateOne(sc, query, update); err != nil {
			utils.ErrorLog("ChangeDebt", err.Error())
			return err
		}

		err = hs.Compare("1", p.ItemMap)
		if err != nil {
			utils.ErrorLog("ChangeDebt", err.Error())
			return err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("ChangeDebt", err.Error())
			return err
		}

		// 第二步：若字段有变更,添加契约历史情报和新旧履历情报
		if changeCount > 0 {
			// 契約履歴番号
			rirekiSeq, err := uuid.NewUUID()
			if err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}
			newItemMap["no"] = &Value{
				DataType: "text",
				Value:    rirekiSeq.String(),
			}
			oldItemMap["no"] = &Value{
				DataType: "text",
				Value:    rirekiSeq.String(),
			}
			// 修正区分编辑
			newItemMap["zengokbn"] = &Value{
				DataType: "options",
				Value:    "after",
			}
			oldItemMap["zengokbn"] = &Value{
				DataType: "options",
				Value:    "before",
			}
			// 对接区分编辑
			newItemMap["dockkbn"] = &Value{
				DataType: "options",
				Value:    "undo",
			}
			oldItemMap["dockkbn"] = &Value{
				DataType: "options",
				Value:    "undo",
			}
			// 操作区分编辑
			newItemMap["actkbn"] = &Value{
				DataType: "options",
				Value:    "debtchange",
			}
			oldItemMap["actkbn"] = &Value{
				DataType: "options",
				Value:    "debtchange",
			}

			// 将契约番号变成lookup类型
			keiyakuno := p.ItemMap["keiyakuno"]

			itMap := make(map[string]interface{})
			for key, val := range p.ItemMap {
				itMap[key] = val
			}

			keiyakuItem := keiyakuno.Value.(string)

			newItemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			oldItemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}

			queryTmp := bson.M{
				"template_id":   templateID,
				"datastore_key": "rireki",
			}
			var rirekiTmp TemplateItem

			if err := ct.FindOne(ctx, queryTmp).Decode(&rirekiTmp); err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}
			// 使用权资产总额
			newItemMap["shisannsougaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["shisannsougaku"].Value,
			}
			oldItemMap["shisannsougaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["o_shisannsougaku"].Value,
			}
			// 租赁债务总额
			newItemMap["leasesaimusougaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["leasesaimusougaku"].Value,
			}
			oldItemMap["leasesaimusougaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["o_leasesaimusougaku"].Value,
			}
			// 使用权资产差额
			newItemMap["shisannsagaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["shisannsagaku"].Value,
			}
			oldItemMap["shisannsagaku"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 租赁债务差额
			newItemMap["leasesaimusagaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["leasesaimusagaku"].Value,
			}
			oldItemMap["leasesaimusagaku"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 损益额
			newItemMap["sonnekigaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["sonnekigaku"].Value,
			}
			oldItemMap["sonnekigaku"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 変更時点の支払残額に対して、比例減少した金額
			newItemMap["gensyoPayTotal"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["gensyoPayTotal"].Value,
			}
			oldItemMap["gensyoPayTotal"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 変更時点の元本残高に対して、比例減少した金額
			newItemMap["gensyoBalance"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["gensyoBalance"].Value,
			}
			oldItemMap["gensyoBalance"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 変更時点の帳簿価額に対して、比例減少した金額
			newItemMap["gensyoBoka"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["gensyoBoka"].Value,
			}
			oldItemMap["gensyoBoka"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 再見積変更後現在価値
			newItemMap["leaseTotalAfter"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["leaseTotalAfter"].Value,
			}
			oldItemMap["leaseTotalAfter"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 変更時点の元本残高に対して、比例残の金額
			newItemMap["leaseTotalRemain"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["leaseTotalRemain"].Value,
			}
			oldItemMap["leaseTotalRemain"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 再見積変更後の支払総額
			newItemMap["payTotalAfter"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["payTotalAfter"].Value,
			}
			oldItemMap["payTotalAfter"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 変更時点の支払残額に対して、比例残の金額
			newItemMap["payTotalRemain"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["payTotalRemain"].Value,
			}
			oldItemMap["payTotalRemain"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 変更時点の支払残額に対して、比例残の金額
			newItemMap["payTotalChange"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["payTotalChange"].Value,
			}
			oldItemMap["payTotalChange"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 添加新契约履历情报
			var newRirekiItem Item
			newRirekiItem.ID = primitive.NewObjectID()
			newRirekiItem.ItemID = newRirekiItem.ID.Hex()
			newRirekiItem.AppID = p.AppID
			newRirekiItem.DatastoreID = dsrireki
			newRirekiItem.CreatedAt = p.UpdatedAt
			newRirekiItem.CreatedBy = p.UpdatedBy
			newRirekiItem.ItemMap = newItemMap
			newRirekiItem.Owners = oldItem.Owners

			queryJSON, _ = json.Marshal(newRirekiItem)
			utils.DebugLog("ModifyItem", fmt.Sprintf("newRirekiItem: [ %s ]", queryJSON))
			if _, err := cr.InsertOne(sc, newRirekiItem); err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}
			// 添加旧契约履历情报
			var oldRirekiItem Item
			oldRirekiItem.ID = primitive.NewObjectID()
			oldRirekiItem.ItemID = oldRirekiItem.ID.Hex()
			oldRirekiItem.AppID = p.AppID
			oldRirekiItem.DatastoreID = dsrireki
			oldRirekiItem.CreatedAt = p.UpdatedAt
			oldRirekiItem.CreatedBy = p.UpdatedBy
			oldRirekiItem.ItemMap = oldItemMap
			oldRirekiItem.Owners = oldItem.Owners

			queryJSON, _ = json.Marshal(oldRirekiItem)
			utils.DebugLog("ModifyItem", fmt.Sprintf("oldRirekiItem: [ %s ]", queryJSON))
			if _, err := cr.InsertOne(sc, oldRirekiItem); err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}

			// 获取契约情报
			keiyakunoValue := keiyakuno.Value.(string)

			/* ******************契约更新后根据契约番号删除以前的支付，利息，偿还的数据************* */
			querydel := bson.M{
				"items.keiyakuno.value": keiyakunoValue,
			}

			if _, err := cpay.DeleteMany(sc, querydel); err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}

			if _, err := cinter.DeleteMany(sc, querydel); err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}

			if _, err := crepay.DeleteMany(sc, querydel); err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}

			/* ******************************* */

			/* ******************登录契约更新后重新生成的支付，利息，偿还的数据************* */
			// 获取会社信息
			leasekaisha := p.ItemMap["leasekaishacd"]
			kaisyaItem := leasekaisha.Value.(string)

			// 获取分類コード信息
			bunruicd := p.ItemMap["bunruicd"]
			bunruicdItem := bunruicd.Value.(string)

			// 获取管理部門
			segmentcd := p.ItemMap["segmentcd"]
			segmentcdItem := segmentcd.Value.(string)

			// 取支付信息数据登录数据库 paymentStatus
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "paymentStatus",
				UserID:        p.UpdatedBy,
				Owners:        p.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
			})
			if err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}

			// 取利息数据登录数据库 paymentInterest
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "paymentInterest",
				UserID:        p.UpdatedBy,
				Owners:        p.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
			})
			if err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}

			// 取偿还数据登录数据库 repayment
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "repayment",
				UserID:        p.UpdatedBy,
				Owners:        p.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
			})
			if err != nil {
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}

			/* ******************************* */

		}

		// 删除临时数据
		tc := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
		query1 := bson.M{
			"template_id": templateID,
		}

		_, err = tc.DeleteMany(sc, query1)
		if err != nil {
			utils.ErrorLog("ChangeDebt", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("ChangeDebt", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("ChangeDebt", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// TerminateContract 中途解约
func TerminateContract(db, collection string, p *ItemUpdateParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(p.DatastoreID))
	ct := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 事务处理开始
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("TerminateContract", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("TerminateContract", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		// 获取所有台账
		dsList, e := FindDatastores(db, p.AppID, "", "", "")
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				dsList = []Datastore{}
			} else {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}
		}

		dsMap := make(map[string]string)
		for _, d := range dsList {
			dsMap[d.ApiKey] = d.DatastoreID
		}

		// 契约台账履历表取得
		dsrireki := dsMap["rireki"]
		dsPay := dsMap["paymentStatus"]
		dsInterest := dsMap["paymentInterest"]
		dsRepay := dsMap["repayment"]

		cr := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsrireki))
		cpay := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsPay))
		cinter := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsInterest))
		crepay := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsRepay))

		// 查找自增字段
		param := &FindFieldsParam{
			AppID:       p.AppID,
			DatastoreID: p.DatastoreID,
		}

		allFields, err := FindFields(db, param)
		if err != nil {
			utils.ErrorLog("TerminateContract", err.Error())
			if err.Error() != mongo.ErrNoDocuments.Error() {
				return err
			}
		}
		// 解约前契约情报取得
		oldItem, e := getItem(db, p.ItemID, p.DatastoreID, p.Owners)
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				return errors.New("データが存在しないか、あなたまたはデータの申請者はデータを変更する権限がありません")
			}
			utils.ErrorLog("TerminateContract", e.Error())
			return e
		}

		hs := NewHistory(db, p.UpdatedBy, p.DatastoreID, p.Lang, p.Domain, sc, allFields)

		err = hs.Add("1", p.ItemID, oldItem.ItemMap)
		if err != nil {
			utils.ErrorLog("TerminateContract", err.Error())
			return err
		}

		// 变更箇所数记录用
		changeCount := 0

		// 自增字段不更新
		if len(allFields) > 0 {
			for _, f := range allFields {
				if f.FieldType == "autonum" {
					delete(oldItem.ItemMap, f.FieldID)
					delete(p.ItemMap, f.FieldID)
					continue
				}
				_, ok := p.ItemMap[f.FieldID]
				// 需要进行自算的情况
				if f.FieldType == "number" && len(f.SelfCalculate) > 0 && ok {
					if f.SelfCalculate == "add" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o + n
						continue
					}
					if f.SelfCalculate == "sub" {
						o := GetNumberValue(oldItem.ItemMap[f.FieldID])
						n := GetNumberValue(p.ItemMap[f.FieldID])
						p.ItemMap[f.FieldID].Value = o - n
						continue
					}
				}
			}
		}

		p.ItemMap["status"] = &Value{
			DataType: "options",
			Value:    "cancel",
		}

		// 变更后契约履历情报
		newItemMap := make(map[string]*Value, len(p.ItemMap)+12)
		// 变更前契约履历情报
		oldItemMap := make(map[string]*Value, len(p.ItemMap)+12)

		//查询临时表取得契约履历的下列数据
		// 临时数据ID
		templateID := ""
		if val, exist := p.ItemMap["template_id"]; exist {
			templateID = val.Value.(string)
			// 删除临时数据ID
			delete(p.ItemMap, "template_id")
		}

		// 循环取得变更前契约情报数据
		for key, value := range oldItem.ItemMap {
			oldItemMap[key] = value
		}

		// 契约变更情报编辑
		item := bson.M{
			"updated_at": p.UpdatedAt,
			"updated_by": p.UpdatedBy,
		}

		// 循环契约情报数据对比变更
		for key, value := range p.ItemMap {
			// 记录履历数据情报(包含未变更数据项)
			newItemMap[key] = value
			item["items."+key] = value
			// 对比前后数据值
			if _, ok := oldItem.ItemMap[key]; ok {
				// 该项数据历史存在,判断历史与当前是否变更
				if compare(value, oldItem.ItemMap[key]) {

					changeCount++
				}
			} else {
				// 该项数据历史不存在,判断当前是否为空
				if value.Value == "" || value.Value == "[]" {
					continue
				}

				changeCount++
			}
		}

		// 契约情报变更参数编辑
		update := bson.M{"$set": item}
		objectID, e := primitive.ObjectIDFromHex(p.ItemID)
		if e != nil {
			utils.ErrorLog("TerminateContract", e.Error())
			return e
		}
		query := bson.M{
			"_id": objectID,
		}
		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("TerminateContract", fmt.Sprintf("query: [ %s ]", queryJSON))
		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("TerminateContract", fmt.Sprintf("update: [ %s ]", updateJSON))

		// 第一步：更新契约台账情报
		if _, err := c.UpdateOne(sc, query, update); err != nil {
			utils.ErrorLog("TerminateContract", err.Error())
			return err
		}

		err = hs.Compare("1", p.ItemMap)
		if err != nil {
			utils.ErrorLog("TerminateContract", err.Error())
			return err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("TerminateContract", err.Error())
			return err
		}

		// 第二步：若字段有变更,添加契约历史情报和新旧履历情报
		if changeCount > 0 {
			// 契約履歴番号
			rirekiSeq, err := uuid.NewUUID()
			if err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}
			newItemMap["no"] = &Value{
				DataType: "text",
				Value:    rirekiSeq.String(),
			}
			oldItemMap["no"] = &Value{
				DataType: "text",
				Value:    rirekiSeq.String(),
			}
			// 修正区分编辑
			newItemMap["zengokbn"] = &Value{
				DataType: "options",
				Value:    "after",
			}
			oldItemMap["zengokbn"] = &Value{
				DataType: "options",
				Value:    "before",
			}
			// 对接区分编辑
			newItemMap["dockkbn"] = &Value{
				DataType: "options",
				Value:    "undo",
			}
			oldItemMap["dockkbn"] = &Value{
				DataType: "options",
				Value:    "undo",
			}
			// 操作区分编辑
			newItemMap["actkbn"] = &Value{
				DataType: "options",
				Value:    "midcancel",
			}
			oldItemMap["actkbn"] = &Value{
				DataType: "options",
				Value:    "midcancel",
			}

			queryTmp := bson.M{
				"template_id":   templateID,
				"datastore_key": "rireki",
			}
			var rirekiTmp TemplateItem

			if err := ct.FindOne(ctx, queryTmp).Decode(&rirekiTmp); err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}
			// 剩余的债务
			newItemMap["remaindebt"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["remaindebt"].Value,
			}
			oldItemMap["remaindebt"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 除却損金额
			newItemMap["lossgaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["lossgaku"].Value,
			}
			oldItemMap["lossgaku"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 解约年月日
			newItemMap["kaiyakuymd"] = &Value{
				DataType: "text",
				Value:    rirekiTmp.ItemMap["kaiyakuymd"].Value,
			}
			// 中途解約時点の償却費の累計額
			newItemMap["syokyakuTotal"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["syokyakuTotal"].Value,
			}
			oldItemMap["syokyakuTotal"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 使用権資産の原始計上額
			oldItemMap["kisyuboka"] = &Value{
				DataType: "number",
				Value:    "0",
			}
			// 中途解約時点の支払リース料残額
			newItemMap["payTotalRemain"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["payTotalRemain"].Value,
			}
			oldItemMap["payTotalRemain"] = &Value{
				DataType: "number",
				Value:    0,
			}
			// 中途解約時点の利息残
			newItemMap["interestTotalRemain"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["interestTotalRemain"].Value,
			}
			oldItemMap["interestTotalRemain"] = &Value{
				DataType: "number",
				Value:    0,
			}

			// 将契约番号变成lookup类型
			keiyakuno := p.ItemMap["keiyakuno"]

			itMap := make(map[string]interface{})
			for key, val := range p.ItemMap {
				itMap[key] = val
			}

			keiyakuItem := keiyakuno.Value.(string)

			newItemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			oldItemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			// 添加新契约履历情报
			var newRirekiItem Item
			newRirekiItem.ID = primitive.NewObjectID()
			newRirekiItem.ItemID = newRirekiItem.ID.Hex()
			newRirekiItem.AppID = p.AppID
			newRirekiItem.DatastoreID = dsrireki
			newRirekiItem.CreatedAt = p.UpdatedAt
			newRirekiItem.CreatedBy = p.UpdatedBy
			newRirekiItem.ItemMap = newItemMap
			newRirekiItem.Owners = oldItem.Owners

			queryJSON, _ = json.Marshal(newRirekiItem)
			utils.DebugLog("TerminateContract", fmt.Sprintf("newRirekiItem: [ %s ]", queryJSON))
			if _, err := cr.InsertOne(sc, newRirekiItem); err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}
			// 添加旧契约履历情报
			var oldRirekiItem Item
			oldRirekiItem.ID = primitive.NewObjectID()
			oldRirekiItem.ItemID = oldRirekiItem.ID.Hex()
			oldRirekiItem.AppID = p.AppID
			oldRirekiItem.DatastoreID = dsrireki
			oldRirekiItem.CreatedAt = p.UpdatedAt
			oldRirekiItem.CreatedBy = p.UpdatedBy
			oldRirekiItem.ItemMap = oldItemMap
			oldRirekiItem.Owners = oldItem.Owners

			queryJSON, _ = json.Marshal(oldRirekiItem)
			utils.DebugLog("TerminateContract", fmt.Sprintf("oldRirekiItem: [ %s ]", queryJSON))
			if _, err := cr.InsertOne(sc, oldRirekiItem); err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}

			// 获取契约情报
			keiyakunoItem := keiyakuno.Value.(string)

			/* ******************契约更新后根据契约番号删除以前的支付，利息，偿还的数据************* */
			querydel := bson.M{
				"items.keiyakuno.value": keiyakunoItem,
			}

			if _, err := cpay.DeleteMany(sc, querydel); err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}

			if _, err := cinter.DeleteMany(sc, querydel); err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}

			if _, err := crepay.DeleteMany(sc, querydel); err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}

			/* ******************************* */

			/* ******************登录契约更新后重新生成的支付，利息，偿还的数据************* */
			// 获取会社信息
			leasekaisha := p.ItemMap["leasekaishacd"]
			kaisyaItem := leasekaisha.Value.(string)

			// 获取分類コード信息
			bunruicd := p.ItemMap["bunruicd"]
			bunruicdItem := bunruicd.Value.(string)

			// 获取管理部門
			segmentcd := p.ItemMap["segmentcd"]
			segmentcdItem := segmentcd.Value.(string)

			// 取支付信息数据登录数据库 paymentStatus
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "paymentStatus",
				UserID:        p.UpdatedBy,
				Owners:        p.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
			})
			if err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}

			// 取利息数据登录数据库 paymentInterest
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "paymentInterest",
				UserID:        p.UpdatedBy,
				Owners:        p.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
			})
			if err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}

			// 取偿还数据登录数据库 repayment
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "repayment",
				UserID:        p.UpdatedBy,
				Owners:        p.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
			})
			if err != nil {
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}

			/* ******************************* */

		}

		// 删除临时数据
		tc := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
		query1 := bson.M{
			"template_id": templateID,
		}

		_, err = tc.DeleteMany(sc, query1)
		if err != nil {
			utils.ErrorLog("TerminateContract", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("TerminateContract", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("TerminateContract", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// ContractExpire 契约满了
func ContractExpire(db, collection string, p *ItemUpdateParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(p.DatastoreID))
	ct := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 事务处理开始
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("ContractExpire", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("ContractExpire", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		// 获取所有台账
		dsList, e := FindDatastores(db, p.AppID, "", "", "")
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				dsList = []Datastore{}
			} else {
				utils.ErrorLog("ContractExpire", err.Error())
				return err
			}
		}
		dsMap := make(map[string]string)
		for _, d := range dsList {
			dsMap[d.ApiKey] = d.DatastoreID
		}

		dsrireki := dsMap["rireki"]
		dsRepay := dsMap["repayment"]

		cr := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsrireki))
		crepay := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(dsRepay))

		// 查找自增字段
		param := &FindFieldsParam{
			AppID:       p.AppID,
			DatastoreID: p.DatastoreID,
		}

		allFields, err := FindFields(db, param)
		if err != nil {
			utils.ErrorLog("TerminateContract", err.Error())
			if err.Error() != mongo.ErrNoDocuments.Error() {
				return err
			}
		}

		// 变更前契约情报取得
		oldItem, e := getItem(db, p.ItemID, p.DatastoreID, p.Owners)
		if e != nil {
			if e.Error() == mongo.ErrNoDocuments.Error() {
				return errors.New("データが存在しないか、あなたまたはデータの申請者はデータを変更する権限がありません")
			}
			utils.ErrorLog("ContractExpire", e.Error())
			return e
		}

		hs := NewHistory(db, p.UpdatedBy, p.DatastoreID, p.Lang, p.Domain, sc, allFields)

		err = hs.Add("1", p.ItemID, oldItem.ItemMap)
		if err != nil {
			utils.ErrorLog("ContractExpire", err.Error())
			return err
		}

		oldTtems := make(map[string]*Value, len(oldItem.ItemMap))
		for key, item := range oldItem.ItemMap {
			oldTtems[key] = item
		}

		// 临时数据ID
		templateID := ""
		if val, exist := p.ItemMap["template_id"]; exist {
			templateID = val.Value.(string)
			// 删除临时数据ID
			delete(p.ItemMap, "template_id")
		}

		// 变更后契约履历情报
		newItemMap := make(map[string]*Value)
		// 变更前契约履历情报
		oldItemMap := make(map[string]*Value)

		// 追加契约状态
		p.ItemMap["status"] = &Value{
			DataType: "options",
			Value:    "complete",
		}

		// 循环取得契约情报数据
		for key, value := range oldItem.ItemMap {
			oldItemMap[key] = value
			newItemMap[key] = value
		}

		// 契约变更情报编辑
		change := bson.M{
			"updated_at": p.UpdatedAt,
			"updated_by": p.UpdatedBy,
		}

		// 循环契约情报数据对比变更
		for key, value := range p.ItemMap {
			// 记录履历数据情报(包含未变更数据项)
			newItemMap[key] = value
			change["items."+key] = value
		}

		// 契约情报变更参数编辑
		update := bson.M{"$set": change}
		objectID, e := primitive.ObjectIDFromHex(p.ItemID)
		if e != nil {
			utils.ErrorLog("ContractExpire", e.Error())
			return e
		}
		query := bson.M{
			"_id": objectID,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("ContractExpire", fmt.Sprintf("query: [ %s ]", queryJSON))
		updateJSON, _ := json.Marshal(update)
		utils.DebugLog("ContractExpire", fmt.Sprintf("update: [ %s ]", updateJSON))

		// 更新契约台账情报
		if _, err := c.UpdateOne(sc, query, update); err != nil {
			utils.ErrorLog("ContractExpire", err.Error())
			return err
		}

		err = hs.Compare("1", p.ItemMap)
		if err != nil {
			utils.ErrorLog("ContractExpire", err.Error())
			return err
		}

		err = hs.Commit()
		if err != nil {
			utils.ErrorLog("ContractExpire", err.Error())
			return err
		}

		// 契约履历数据
		// 契約履歴番号
		rirekiSeq, err := uuid.NewUUID()
		if err != nil {
			utils.ErrorLog("ContractExpire", err.Error())
			return err
		}
		newItemMap["no"] = &Value{
			DataType: "text",
			Value:    rirekiSeq.String(),
		}
		oldItemMap["no"] = &Value{
			DataType: "text",
			Value:    rirekiSeq.String(),
		}
		// 循环取得契约情报数据
		for key, value := range oldItem.ItemMap {
			oldItemMap[key] = value
			newItemMap[key] = value
		}
		// 修正区分编辑
		newItemMap["zengokbn"] = &Value{
			DataType: "options",
			Value:    "after",
		}
		oldItemMap["zengokbn"] = &Value{
			DataType: "options",
			Value:    "before",
		}
		// 对接区分编辑
		newItemMap["dockkbn"] = &Value{
			DataType: "options",
			Value:    "undo",
		}
		oldItemMap["dockkbn"] = &Value{
			DataType: "options",
			Value:    "undo",
		}
		// 操作区分编辑
		newItemMap["actkbn"] = &Value{
			DataType: "options",
			Value:    "expire",
		}
		oldItemMap["actkbn"] = &Value{
			DataType: "options",
			Value:    "expire",
		}
		//查询临时表取得契约履历的下列数据

		if len(templateID) > 0 {
			queryTmp := bson.M{
				"template_id":   templateID,
				"datastore_key": "rireki",
			}
			var rirekiTmp TemplateItem

			if err := ct.FindOne(ctx, queryTmp).Decode(&rirekiTmp); err != nil {
				utils.ErrorLog("ContractExpire", err.Error())
				return err
			}
			// 除却損金额
			newItemMap["lossgaku"] = &Value{
				DataType: "number",
				Value:    rirekiTmp.ItemMap["lossgaku"].Value,
			}
		}

		// 将契约番号变成lookup类型
		keiyakuno := oldTtems["keiyakuno"]

		itMap := make(map[string]interface{})
		for key, val := range oldTtems {
			itMap[key] = val
		}

		keiyakuItem := keiyakuno.Value.(string)

		newItemMap["keiyakuno"] = &Value{
			DataType: "lookup",
			Value:    keiyakuItem,
		}
		oldItemMap["keiyakuno"] = &Value{
			DataType: "lookup",
			Value:    keiyakuItem,
		}
		// 添加新契约履历情报
		var newRirekiItem Item
		newRirekiItem.ID = primitive.NewObjectID()
		newRirekiItem.ItemID = newRirekiItem.ID.Hex()
		newRirekiItem.AppID = p.AppID
		newRirekiItem.DatastoreID = dsrireki
		newRirekiItem.CreatedAt = p.UpdatedAt
		newRirekiItem.CreatedBy = p.UpdatedBy
		newRirekiItem.ItemMap = newItemMap
		newRirekiItem.Owners = oldItem.Owners

		queryJSON, _ = json.Marshal(newRirekiItem)
		utils.DebugLog("ContractExpire", fmt.Sprintf("newRirekiItem: [ %s ]", queryJSON))
		if _, err := cr.InsertOne(sc, newRirekiItem); err != nil {
			utils.ErrorLog("ContractExpire", err.Error())
			return err
		}
		// 添加旧契约履历情报
		var oldRirekiItem Item
		oldRirekiItem.ID = primitive.NewObjectID()
		oldRirekiItem.ItemID = oldRirekiItem.ID.Hex()
		oldRirekiItem.AppID = p.AppID
		oldRirekiItem.DatastoreID = dsrireki
		oldRirekiItem.CreatedAt = p.UpdatedAt
		oldRirekiItem.CreatedBy = p.UpdatedBy
		oldRirekiItem.ItemMap = oldItemMap
		oldRirekiItem.Owners = oldItem.Owners

		queryJSON, _ = json.Marshal(oldRirekiItem)
		utils.DebugLog("ContractExpire", fmt.Sprintf("oldRirekiItem: [ %s ]", queryJSON))
		if _, err := cr.InsertOne(sc, oldRirekiItem); err != nil {
			utils.ErrorLog("ContractExpire", err.Error())
			return err
		}

		if len(templateID) > 0 {
			/* ******************契约更新后根据契约番号删除以前的偿还的数据************* */
			querydel := bson.M{
				"items.keiyakuno.value": keiyakuItem,
			}
			if _, err := crepay.DeleteMany(sc, querydel); err != nil {
				utils.ErrorLog("ContractExpire", err.Error())
				return err
			}
			/* ******************登录契约更新后重新生成的偿还的数据************* */
			// 获取会社信息
			leasekaisha := p.ItemMap["leasekaishacd"]
			kaisyaItem := leasekaisha.Value.(string)

			// 获取分類コード信息
			bunruicd := p.ItemMap["bunruicd"]
			bunruicdItem := bunruicd.Value.(string)

			// 获取管理部門
			segmentcd := p.ItemMap["segmentcd"]
			segmentcdItem := segmentcd.Value.(string)

			// 取偿还数据登录数据库 repayment
			err = insertTempData(client, sc, TmpParam{
				DB:            db,
				TemplateID:    templateID,
				APIKey:        "repayment",
				UserID:        p.UpdatedBy,
				Owners:        p.Owners,
				Datastores:    dsList,
				Keiyakuno:     keiyakuItem,
				Leasekaishacd: kaisyaItem,
				Bunruicd:      bunruicdItem,
				Segmentcd:     segmentcdItem,
			})
			if err != nil {
				utils.ErrorLog("ContractExpire", err.Error())
				return err
			}

			// 删除临时数据
			tc := client.Database(database.GetDBName(db)).Collection(genTplCollectionName(collection))
			query := bson.M{
				"template_id": templateID,
			}

			_, err = tc.DeleteMany(sc, query)
			if err != nil {
				utils.ErrorLog("ContractExpire", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("ContractExpire", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("ContractExpire", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

func insertTempData(client *mongo.Client, sc mongo.SessionContext, p TmpParam) error {

	c := client.Database(database.GetDBName(p.DB)).Collection(genTplCollectionName(p.UserID))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"template_id":   p.TemplateID,
		"datastore_key": p.APIKey,
	}

	var payData []TemplateItem

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("insertTempData", err.Error())
		return err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &payData)
	if err != nil {
		utils.ErrorLog("insertTempData", err.Error())
		return err
	}

	if len(payData) > 0 {
		var items []*Item
		datastoreID := ""
		appID := ""
		for index, it := range payData {
			if index == 0 {
				datastoreID = it.DatastoreID
				appID = it.AppID
			}

			itemMap := it.ItemMap

			itemMap["keiyakuno"] = &Value{
				DataType: "lookup",
				Value:    p.Keiyakuno,
			}

			itemMap["leasekaishacd"] = &Value{
				DataType: "lookup",
				Value:    p.Leasekaishacd,
			}

			itemMap["bunruicd"] = &Value{
				DataType: "lookup",
				Value:    p.Bunruicd,
			}

			itemMap["segmentcd"] = &Value{
				DataType: "lookup",
				Value:    p.Segmentcd,
			}

			id := primitive.NewObjectID()

			items = append(items, &Item{
				ID:          id,
				ItemID:      id.Hex(),
				AppID:       it.AppID,
				DatastoreID: it.DatastoreID,
				ItemMap:     itemMap,
				Owners:      p.Owners,
				CreatedAt:   time.Now(),
				CreatedBy:   p.UserID,
				UpdatedAt:   time.Now(),
				UpdatedBy:   p.UserID,
				Status:      "1",
			})
		}
		pc := client.Database(database.GetDBName(p.DB)).Collection(GetItemCollectionName(datastoreID))

		autoFields := []Field{}

		allFields := p.FileMap[datastoreID]
		if allFields == nil {
			// 获取所有字段
			params := FindFieldsParam{
				AppID:       appID,
				DatastoreID: datastoreID,
			}

			allFields, err = FindFields(p.DB, &params)
			if err != nil {
				utils.ErrorLog("insertTempData", err.Error())
				return err
			}
		}

		step := len(items)
		autoList := make(map[string][]string)

		for _, f := range allFields {
			if f.FieldType == "autonum" {
				list, err := autoNumListWithSession(sc, p.DB, &f, step)
				if err != nil {
					utils.ErrorLog("insertTempData", err.Error())
					return err
				}
				autoList[f.FieldID] = list
			}

		}

		var insertData []interface{}
		for index, it := range items {
			if len(autoFields) > 0 {
				for _, f := range autoFields {
					nums := autoList[f.FieldID]
					it.ItemMap[f.FieldID] = &Value{
						DataType: "autonum",
						Value:    nums[index],
					}
				}
			}

			insertData = append(insertData, it)
		}
		_, err = pc.InsertMany(sc, insertData)
		if err != nil {
			utils.ErrorLog("insertTempData", err.Error())
			return err
		}
	}

	return nil
}

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
