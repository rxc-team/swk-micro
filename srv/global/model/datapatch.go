package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"rxcsoft.cn/pit3/srv/global/utils"
	database "rxcsoft.cn/utils/mongo"
)

// DataPatch1216Param
type DataPatch1216Param struct {
	Domain string
	LangCd string
	AppID  string
	Kind   string
	Type   string
	DelKbn bool
	Value  map[string]string
	Writer string
}

func DataPatch1216(db string, p DataPatch1216Param) (err error) {
	// 连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain":  p.Domain,
		"lang_cd": p.LangCd,
	}

	key := ""
	if p.Kind == "apps" {
		key = "apps." + p.AppID + "." + p.Type
	} else {
		key = "common." + p.Type
	}

	val := p.Value
	if p.DelKbn {
		val = make(map[string]string)
	}

	// update := bson.M{
	// 	"$set": bson.M{
	// 		key:          val,
	// 		"updated_at": time.Now(),
	// 		"updated_by": p.Writer,
	// 	},
	// }

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": p.Writer,
	}
	if len(val) > 0 {
		for k, v := range val {
			change[key+"."+k] = v
		}
	} else {
		change[key] = val
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DataPatch1216", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DataPatch1216", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DataPatch1216", err.Error())
		return err
	}

	return nil
}
