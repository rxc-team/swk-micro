package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rxcsoft.cn/pit3/srv/report/proto/coldata"
	"rxcsoft.cn/pit3/srv/report/utils"
	database "rxcsoft.cn/utils/mongo"
)

// 总表数据结构
type ColData struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	Keiyakuno          string             `json:"keiyakuno" bson:"keiyakuno"`
	Keiyakuymd         time.Time          `json:"keiyakuymd" bson:"keiyakuymd"`
	Leasestymd         time.Time          `json:"leasestymd" bson:"leasestymd"`
	Leasekikan         int64              `json:"leasekikan" bson:"leasekikan"`
	ExtentionOption    int64              `json:"extentionOption" bson:"extentionOption"`
	Leaseexpireymd     time.Time          `json:"leaseexpireymd" bson:"leaseexpireymd"`
	Keiyakunm          string             `json:"keiyakunm" bson:"keiyakunm"`
	Biko1              string             `json:"biko1" bson:"biko1"`
	Paymentstymd       time.Time          `json:"paymentstymd" bson:"paymentstymd"`
	Paymentcycle       string             `json:"paymentcycle" bson:"paymentcycle"`
	Paymentday         string             `json:"paymentday" bson:"paymentday"`
	Paymentcounts      int64              `json:"paymentcounts" bson:"paymentcounts"`
	ResidualValue      int64              `json:"residualValue" bson:"residualValue"`
	Rishiritsu         float64            `json:"rishiritsu" bson:"rishiritsu"`
	InitialDirectCosts int64              `json:"initialDirectCosts" bson:"initialDirectCosts"`
	RestorationCosts   int64              `json:"restorationCosts" bson:"restorationCosts"`
	Sykshisankeisan    string             `json:"sykshisankeisan" bson:"sykshisankeisan"`
	Segmentcd          string             `json:"segmentcd" bson:"segmentcd"`
	Bunruicd           string             `json:"bunruicd" bson:"bunruicd"`
	Field_viw          string             `json:"field_viw" bson:"field_viw"`
	Field_22c          string             `json:"field_22c" bson:"field_22c"`
	Field_1av          string             `json:"field_1av" bson:"field_1av"`
	Field_206          string             `json:"field_206" bson:"field_206"`
	Field_14l          string             `json:"field_14l" bson:"field_14l"`
	Field_7p3          string             `json:"field_7p3" bson:"field_7p3"`
	Field_248          string             `json:"field_248" bson:"field_248"`
	Field_3k7          string             `json:"field_3k7" bson:"field_3k7"`
	Field_1vg          string             `json:"field_1vg" bson:"field_1vg"`
	Field_5fj          string             `json:"field_5fj" bson:"field_5fj"`
	Field_20h          string             `json:"field_20h" bson:"field_20h"`
	Field_2h1          string             `json:"field_2h1" bson:"field_2h1"`
	Field_qi4          string             `json:"field_qi4" bson:"field_qi4"`
	Field_1ck          string             `json:"field_1ck" bson:"field_1ck"`
	Field_u1q          string             `json:"field_u1q" bson:"field_u1q"`
	Hkkjitenzan        float64            `json:"hkkjitenzan" bson:"hkkjitenzan"`
	Sonnekigaku        float64            `json:"sonnekigaku" bson:"sonnekigaku"`
	Year               int64              `json:"year" bson:"year"`
	Month              int64              `json:"month" bson:"month"`
	Paymentleasefee    int64              `json:"paymentleasefee" bson:"paymentleasefee"`
	Interest           int64              `json:"interest" bson:"interest"`
	Repayment          int64              `json:"repayment" bson:"repayment"`
	Firstbalance       int64              `json:"firstbalance" bson:"firstbalance"`
	Balance            int64              `json:"balance" bson:"balance"`
	Endboka            int64              `json:"endboka" bson:"endboka"`
	Boka               int64              `json:"boka" bson:"boka"`
	Syokyaku           int64              `json:"syokyaku" bson:"syokyaku"`
	UpdateTime         time.Time          `json:"update_time" bson:"update_time"`
}

// 总表结构体（获取数据用）
type summaryData struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Keiyakuno       string             `json:"keiyakuno" bson:"keiyakuno"`
	Year            int64              `json:"year" bson:"year"`
	Month           int64              `json:"month" bson:"month"`
	Paymentstatus   Paymentstatussub   `json:"paymentstatus" bson:"paymentstatus"`
	Paymentinterest paymentinterestsub `json:"paymentinterest" bson:"paymentinterest"`
	Repayment       repaymentsub       `json:"repayment" bson:"repayment"`
	Keiyakudaicho   Keiyakudaichosub   `json:"keiyakudaicho" bson:"keiyakudaicho"`
}

type Paymentstatussub struct {
	Items ItemMap `json:"items" bson:"items"`
}

type paymentinterestsub struct {
	Items ItemMap `json:"items" bson:"items"`
}

type repaymentsub struct {
	Items ItemMap `json:"items" bson:"items"`
}

type Keiyakudaichosub struct {
	Items ItemMap `json:"items" bson:"items"`
}

// 支付表结构体（获取数据用）
type payData struct {
	Keiyakuno     string    `json:"Keiyakuno" bson:"_id"`
	Paymentymdmin time.Time `json:"paymentmdmin" bson:"min"`
	Paymentymdmax time.Time `json:"paymentmdmax" bson:"max"`
}

// 折旧表结构体（获取数据用）
type repayData struct {
	Keiyakuno      string    `json:"Keiyakuno" bson:"_id"`
	Syokyakuymdmin time.Time `json:"syokyakuymdmin" bson:"min"`
	Syokyakuymdmax time.Time `json:"syokyakuymdmax" bson:"max"`
}

// ColDataInfo 总表数据信息
type ColDataInfo struct {
	ColData []*ColData
	Total   int64
}

func (u *ColData) ToProto() *coldata.ColData {
	return &coldata.ColData{
		Keiyakuno:          u.Keiyakuno,
		Keiyakuymd:         u.Keiyakuymd.String(),
		Leasestymd:         u.Leasestymd.String(),
		Leasekikan:         u.Leasekikan,
		Extentionoption:    u.ExtentionOption,
		Leaseexpireymd:     u.Leaseexpireymd.String(),
		Keiyakunm:          u.Keiyakunm,
		Biko1:              u.Biko1,
		Paymentstymd:       u.Paymentstymd.String(),
		Paymentcycle:       u.Paymentcycle,
		Paymentday:         u.Paymentday,
		Paymentcounts:      u.Paymentcounts,
		Residualvalue:      u.ResidualValue,
		Rishiritsu:         strconv.FormatFloat(u.Rishiritsu, 'f', -1, 64),
		Initialdirectcosts: u.InitialDirectCosts,
		Restorationcosts:   u.RestorationCosts,
		Sykshisankeisan:    u.Sykshisankeisan,
		Segmentcd:          u.Segmentcd,
		Bunruicd:           u.Bunruicd,
		FieldViw:           u.Field_viw,
		Field_22C:          u.Field_22c,
		Field_1Av:          u.Field_1av,
		Field_206:          u.Field_206,
		Field_14L:          u.Field_14l,
		Field_7P3:          u.Field_7p3,
		Field_248:          u.Field_248,
		Field_3K7:          u.Field_3k7,
		Field_1Vg:          u.Field_1vg,
		Field_5Fj:          u.Field_5fj,
		Field_20H:          u.Field_20h,
		Field_2H1:          u.Field_2h1,
		FieldQi4:           u.Field_qi4,
		Field_1Ck:          u.Field_1ck,
		FieldU1Q:           u.Field_u1q,
		Hkkjitenzan:        strconv.FormatFloat(u.Hkkjitenzan, 'f', -1, 64),
		Sonnekigaku:        strconv.FormatFloat(u.Sonnekigaku, 'f', -1, 64),
		Year:               u.Year,
		Month:              u.Month,
		Paymentleasefee:    u.Paymentleasefee,
		Interest:           u.Interest,
		Repayment:          u.Repayment,
		Firstbalance:       u.Firstbalance,
		Balance:            u.Balance,
		Endboka:            u.Endboka,
		Boka:               u.Boka,
		Syokyaku:           u.Syokyaku,
		UpdateTime:         u.UpdateTime.String(),
	}
}

const (
	ColDataCollection = "summary_"
)

func ColName(id string) string {
	return ColDataCollection + id
}

// FindColDatas 获取总表数据
func FindColDatas(db, appid string, pageIndex int64, pageSize int64) (colDataInfo *ColDataInfo, err error) {
	return SelectColData(db, appid, "", 0, 0, pageIndex, pageSize)
}

// SelectColData  契约番号，年月获取总表数据
func SelectColData(db, appid, keiyakuno string, year, month int64, pageIndex int64, pageSize int64) (colDataInfo *ColDataInfo, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ColName(appid))
	ctx, cancel := context.WithTimeout(context.Background(), 30000*time.Second)
	defer cancel()

	// 分页显示数据
	var result []*ColData
	limit := pageSize
	skip := (pageIndex - 1) * pageSize

	match := bson.M{}
	if keiyakuno != "" && year == 0 {
		match = bson.M{
			"keiyakuno": keiyakuno,
		}
	}
	if keiyakuno == "" && year != 0 {
		match = bson.M{
			"year":  year,
			"month": month,
		}
	}
	if keiyakuno != "" && year != 0 {
		match = bson.M{
			"keiyakuno": keiyakuno,
			"year":      year,
			"month":     month,
		}
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
	}

	// 排序
	pipe = append(pipe, bson.M{
		"$sort": bson.D{
			{Key: "keiyakuno", Value: 1},
			{Key: "year", Value: 1},
			{Key: "month", Value: 1},
		},
	})

	pipe = append(pipe, bson.M{
		"$skip": skip,
	})

	pipe = append(pipe, bson.M{
		"$limit": limit,
	})

	// 统计文档数量
	total, _ := c.CountDocuments(ctx, match)

	pipeJSON, _ := json.Marshal(pipe)
	utils.DebugLog("FindColDatas", fmt.Sprintf("pipe: [ %s ]", pipeJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("error FindColDatas", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var coldata ColData
		err := cur.Decode(&coldata)
		if err != nil {
			utils.ErrorLog("error FindColDatas", err.Error())
			return nil, err
		}
		result = append(result, &coldata)
	}

	collData := ColDataInfo{
		ColData: result,
		Total:   total,
	}
	return &collData, nil
}

// CreateColData  生成总表数据
func CreateColData(db string, i map[string]*coldata.Value) (err error) {

	var d_id []string
	// 支付信息
	var pay []payData
	// 折旧信息
	var repay []repayData
	// 最小年
	var summaryminyear int
	// 最小月
	var summaryminmonth int
	// 最大年
	var summarymaxyear int
	// 最大月
	var summarymaxmonth int
	// 索引
	var indexs bson.D
	var indext bson.D
	// 期首元本残高
	var firstbalance float64
	// 当月末元本残高
	var balance float64
	// 判断契约番号
	var tempkeiyakuno string
	// 期首月
	var firstmonth int64
	var boka interface{}
	var midvalue bool
	var tempfirstbalance float64

	//获取支付、试算、折旧的ID和appid
	paymentStatus_id := i["paymentStatus"].Value
	paymentInterest_id := i["paymentInterest"].Value
	repayment_id := i["repayment"].Value
	keiyakudaicho_id := i["keiyakudaicho"].Value
	appid := i["appID"].Value
	d_id = append(d_id, paymentStatus_id, paymentInterest_id, repayment_id)

	//创建总表
	client := database.New()
	client.Database(database.GetDBName(db)).CreateCollection(context.TODO(), "summary_"+appid)
	collection := client.Database(database.GetDBName(db)).Collection("summary_" + appid)
	collection.DeleteMany(context.TODO(), bson.D{{}})

	c_paymentStatus := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(paymentStatus_id))
	c_paymentInterest := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(paymentInterest_id))
	c_repayment := client.Database(database.GetDBName(db)).Collection(GetItemCollectionName(repayment_id))
	ctx, cancel := context.WithTimeout(context.Background(), 30000*time.Second)
	defer cancel()

	// 总表创建索引
	existMap := make(map[string]struct{})
	indexOpts := options.CreateIndexes().SetMaxTime(60 * time.Second)

	indexKeys := []string{"keiyakuno", "year", "month"}
	indexMap := bson.M{
		"keiyakuno": 1,
		"year":      1,
		"month":     1,
	}

	for _, key := range indexKeys {
		if _, exist := existMap[key]; !exist {
			existMap[key] = struct{}{}
			indexs = append(indexs, bson.E{
				Key: key, Value: indexMap[key],
			})
		}
	}

	index := mongo.IndexModel{
		Keys:    indexs,
		Options: options.Index().SetSparse(true).SetUnique(true),
	}

	if _, err := collection.Indexes().CreateOne(ctx, index, indexOpts); err != nil {
		utils.ErrorLog("CreateColData_index", err.Error())
	}

	// 支付、折旧、试算表中创建索引
	indexKeytotal := []string{"items.keiyakuno.value", "items.year.value", "items.month.value"}
	indexMaptotal := bson.M{
		"items.keiyakuno.value": 1,
		"items.year.value":      1,
		"items.month.value":     1,
	}

	for _, key := range indexKeytotal {
		if _, exist := existMap[key]; !exist {
			existMap[key] = struct{}{}
			indext = append(indext, bson.E{
				Key: key, Value: indexMaptotal[key],
			})
		}
	}

	indextotal := mongo.IndexModel{
		Keys:    indext,
		Options: options.Index().SetSparse(true).SetUnique(true),
	}

	if _, err := c_paymentInterest.Indexes().CreateOne(ctx, indextotal, indexOpts); err != nil {
		utils.ErrorLog("CreateColData_index", err.Error())
	}
	if _, err := c_paymentStatus.Indexes().CreateOne(ctx, indextotal, indexOpts); err != nil {
		utils.ErrorLog("CreateColData_index", err.Error())
	}
	if _, err := c_repayment.Indexes().CreateOne(ctx, indextotal, indexOpts); err != nil {
		utils.ErrorLog("CreateColData_index", err.Error())
	}

	//支付表查询契约番号、最小支付日期、最大支付日期
	pipe_pay := []bson.M{
		{
			"$group": bson.M{
				"_id": "$items.keiyakuno.value",
				"min": bson.M{
					"$min": "$items.paymentymd.value",
				},
				"max": bson.M{
					"$max": "$items.paymentymd.value",
				},
			},
		},
	}

	pipeJSON_pay, _ := json.Marshal(pipe_pay)
	utils.DebugLog("FindPaymentStatusData", fmt.Sprintf("pipe_pay: [ %s ]", pipeJSON_pay))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur_pay, _ := c_paymentStatus.Aggregate(ctx, pipe_pay, opt)
	defer cur_pay.Close(ctx)

	for cur_pay.Next(ctx) {
		var result payData
		cur_pay.Decode(&result)
		pay = append(pay, result)
	}

	//折旧表查询契约番号、最小折旧日期、最大折旧日期
	defer cancel()
	pipe_repay := []bson.M{
		{
			"$group": bson.M{
				"_id": "$items.keiyakuno.value",
				"min": bson.M{
					"$min": "$items.syokyakuymd.value",
				},
				"max": bson.M{
					"$max": "$items.syokyakuymd.value",
				},
			},
		},
	}
	pipeJSON_repay, _ := json.Marshal(pipe_repay)
	utils.DebugLog("FindReportData", fmt.Sprintf("pipe_repay: [ %s ]", pipeJSON_repay))
	cur_repay, _ := c_repayment.Aggregate(ctx, pipe_repay, opt)
	defer cur_repay.Close(ctx)

	for cur_repay.Next(ctx) {
		var result repayData
		cur_repay.Decode(&result)
		repay = append(repay, result)
	}

	// 支付表和折旧表日期进行比较，取最小年月和最大年月
	for _, pay := range pay {
		for _, repay := range repay {
			if pay.Keiyakuno == repay.Keiyakuno {
				if pay.Paymentymdmin.Before(repay.Syokyakuymdmin) {
					summaryminyear = pay.Paymentymdmin.Year()
					summaryminmonth = int(pay.Paymentymdmin.Month())
				} else {
					summaryminyear = repay.Syokyakuymdmin.Year()
					summaryminmonth = int(repay.Syokyakuymdmin.Month())
				}
				if pay.Paymentymdmax.Before(repay.Syokyakuymdmax) {
					summarymaxyear = repay.Syokyakuymdmax.Year()
					summarymaxmonth = int(repay.Syokyakuymdmax.Month())
				} else {
					summarymaxyear = pay.Paymentymdmax.Year()
					summarymaxmonth = int(pay.Paymentymdmax.Month())
				}

				// 相隔年数
				year_interval := summarymaxyear - summaryminyear
				// 根据最小年月和最大年月循环插入总表数据
				for year_count := 0; year_count <= year_interval; year_count++ {
					if summaryminyear < summarymaxyear {
						for count1 := 0; count1 < 12; count1++ {
							if summaryminmonth <= 12 {
								_, err := collection.InsertOne(context.TODO(), bson.M{
									"keiyakuno":          pay.Keiyakuno,
									"keiyakuymd":         "",
									"leasestymd":         "",
									"leasekikan":         0,
									"extentionOption":    0,
									"leaseexpireymd":     "",
									"keiyakunm":          "",
									"biko1":              "",
									"paymentstymd":       "",
									"paymentcycle":       0,
									"paymentday":         0,
									"paymentcounts":      0,
									"residualValue":      0,
									"rishiritsu":         0,
									"initialDirectCosts": 0,
									"restorationCosts":   0,
									"sykshisankeisan":    0,
									"segmentcd":          "",
									"bunruicd":           "",
									"field_viw":          "",
									"field_22c":          "",
									"field_1av":          "",
									"field_206":          "",
									"field_14l":          "",
									"field_7p3":          "",
									"field_248":          "",
									"field_3k7":          "",
									"field_1vg":          "",
									"field_5fj":          "",
									"field_20h":          "",
									"field_2h1":          "",
									"field_qi4":          "",
									"field_1ck":          "",
									"field_u1q":          "",
									"hkkjitenzan":        "",
									"sonnekigaku":        "",
									"year":               summaryminyear,
									"month":              summaryminmonth,
									"paymentleasefee":    0,
									"interest":           0,
									"repayment":          0,
									"firstbalance":       0,
									"balance":            0,
									"endboka":            0,
									"boka":               0,
									"syokyaku":           0,
									"update_time":        primitive.NewDateTimeFromTime(time.Now()),
								})
								if err != nil {
									utils.ErrorLog("insert ColData error", err.Error())
								}
								summaryminmonth = summaryminmonth + 1
							}
						}
						summaryminyear = summaryminyear + 1
						summaryminmonth = 1
					}
					//相同年情况循环插入总表数据
					if summaryminyear == summarymaxyear {
						for count2 := 0; count2 < 12; count2++ {
							if summaryminmonth <= summarymaxmonth {
								_, err := collection.InsertOne(context.TODO(), bson.M{
									"keiyakuno":          pay.Keiyakuno,
									"keiyakuymd":         "",
									"leasestymd":         "",
									"leasekikan":         0,
									"extentionOption":    0,
									"leaseexpireymd":     "",
									"keiyakunm":          "",
									"biko1":              "",
									"paymentstymd":       "",
									"paymentcycle":       0,
									"paymentday":         0,
									"paymentcounts":      0,
									"residualValue":      0,
									"rishiritsu":         0,
									"initialDirectCosts": 0,
									"restorationCosts":   0,
									"sykshisankeisan":    0,
									"segmentcd":          "",
									"bunruicd":           "",
									"field_viw":          "",
									"field_22c":          "",
									"field_1av":          "",
									"field_206":          "",
									"field_14l":          "",
									"field_7p3":          "",
									"field_248":          "",
									"field_3k7":          "",
									"field_1vg":          "",
									"field_5fj":          "",
									"field_20h":          "",
									"field_2h1":          "",
									"field_qi4":          "",
									"field_1ck":          "",
									"field_u1q":          "",
									"hkkjitenzan":        "",
									"sonnekigaku":        "",
									"year":               summaryminyear,
									"month":              summaryminmonth,
									"paymentleasefee":    0,
									"interest":           0,
									"repayment":          0,
									"firstbalance":       0,
									"balance":            0,
									"endboka":            0,
									"boka":               0,
									"syokyaku":           0,
									"update_time":        primitive.NewDateTimeFromTime(time.Now()),
								})
								if err != nil {
									utils.ErrorLog("insert ColData error", err.Error())
								}
								summaryminmonth = summaryminmonth + 1
							}
						}
					}
				}
			}
		}
	}

	// 睡眠5s钟
	time.Sleep(5 * time.Second)

	// 总表关联支付、试算、折旧获取所有数据
	var pipe []primitive.M
	mapName := map[string]string{
		paymentStatus_id:   "paymentstatus",
		paymentInterest_id: "paymentinterest",
		repayment_id:       "repayment",
		keiyakudaicho_id:   "keiyakudaicho",
	}
	for _, id := range d_id {
		let := bson.M{
			"keiyakuno": "$keiyakuno",
			"year":      "$year",
			"month":     "$month",
		}

		pp := []bson.M{
			{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{
								"$eq": []string{"$items.keiyakuno.value", "$$keiyakuno"},
							},
							{
								"$eq": []string{"$items.year.value", "$$year"},
							},
							{
								"$eq": []string{"$items.month.value", "$$month"},
							},
						},
					},
				},
			},
		}

		lookup := bson.M{
			"from":     "item_" + id,
			"let":      let,
			"pipeline": pp,
			"as":       mapName[id],
		}

		pipe = append(pipe, bson.M{
			"$lookup": lookup,
		})

		unwind := bson.M{
			"path":                       "$" + mapName[id],
			"preserveNullAndEmptyArrays": true,
		}

		pipe = append(pipe, bson.M{
			"$unwind": unwind,
		})
	}

	let := bson.M{
		"keiyakuno": "$keiyakuno",
	}

	pp := []bson.M{
		{
			"$match": bson.M{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$eq": []string{"$items.keiyakuno.value", "$$keiyakuno"},
						},
					},
				},
			},
		},
	}

	lookup := bson.M{
		"from":     "item_" + keiyakudaicho_id,
		"let":      let,
		"pipeline": pp,
		"as":       mapName[keiyakudaicho_id],
	}

	pipe = append(pipe, bson.M{
		"$lookup": lookup,
	})

	unwind := bson.M{
		"path":                       "$" + mapName[keiyakudaicho_id],
		"preserveNullAndEmptyArrays": true,
	}

	pipe = append(pipe, bson.M{
		"$unwind": unwind,
	})

	project := bson.M{
		"_id":                                    1,
		"keiyakuno":                              1,
		"year":                                   1,
		"month":                                  1,
		"paymentstatus.items.paymentleasefee":    1,
		"paymentinterest.items.interest":         1,
		"paymentinterest.items.repayment":        1,
		"paymentinterest.items.balance":          1,
		"repayment.items.endboka":                1,
		"repayment.items.boka":                   1,
		"repayment.items.syokyaku":               1,
		"keiyakudaicho.items.keiyakuymd":         1,
		"keiyakudaicho.items.leasestymd":         1,
		"keiyakudaicho.items.leasekikan":         1,
		"keiyakudaicho.items.leaseexpireymd":     1,
		"keiyakudaicho.items.extentionOption":    1,
		"keiyakudaicho.items.keiyakunm":          1,
		"keiyakudaicho.items.biko1":              1,
		"keiyakudaicho.items.paymentstymd":       1,
		"keiyakudaicho.items.paymentcycle":       1,
		"keiyakudaicho.items.paymentday":         1,
		"keiyakudaicho.items.paymentcounts":      1,
		"keiyakudaicho.items.residualValue":      1,
		"keiyakudaicho.items.rishiritsu":         1,
		"keiyakudaicho.items.initialDirectCosts": 1,
		"keiyakudaicho.items.restorationCosts":   1,
		"keiyakudaicho.items.sykshisankeisan":    1,
		"keiyakudaicho.items.segmentcd":          1,
		"keiyakudaicho.items.bunruicd":           1,
		"keiyakudaicho.items.field_viw":          1,
		"keiyakudaicho.items.field_22c":          1,
		"keiyakudaicho.items.field_1av":          1,
		"keiyakudaicho.items.field_206":          1,
		"keiyakudaicho.items.field_14l":          1,
		"keiyakudaicho.items.field_7p3":          1,
		"keiyakudaicho.items.field_248":          1,
		"keiyakudaicho.items.field_3k7":          1,
		"keiyakudaicho.items.field_1vg":          1,
		"keiyakudaicho.items.field_5fj":          1,
		"keiyakudaicho.items.field_20h":          1,
		"keiyakudaicho.items.field_2h1":          1,
		"keiyakudaicho.items.field_qi4":          1,
		"keiyakudaicho.items.field_1ck":          1,
		"keiyakudaicho.items.field_u1q":          1,
		"keiyakudaicho.items.hkkjitenzan":        1,
		"keiyakudaicho.items.sonnekigaku":        1,
		// "paymentinterest.items.firstbalance":     1,
	}

	pipe = append(pipe, bson.M{
		"$project": project,
	})

	// 排序
	pipe = append(pipe, bson.M{
		"$sort": bson.D{
			{Key: "keiyakuno", Value: 1},
			{Key: "year", Value: 1},
			{Key: "month", Value: 1},
		},
	})

	pipeJSON, _ := json.Marshal(pipe)
	utils.DebugLog("CreateColData", fmt.Sprintf("pipe: [ %s ]", pipeJSON))

	cur, _ := collection.Aggregate(ctx, pipe, opt)
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result summaryData
		cur.Decode(&result)

		query := bson.M{
			"keiyakuno": result.Keiyakuno,
			"year":      result.Year,
			"month":     result.Month,
		}

		add_colData := bson.M{}

		// 更新总表中契约信息
		if result.Keiyakudaicho.Items != nil {
			add_colData["keiyakuymd"] = result.Keiyakudaicho.Items["keiyakuymd"].Value
			add_colData["leasestymd"] = result.Keiyakudaicho.Items["leasestymd"].Value
			add_colData["extentionOption"] = result.Keiyakudaicho.Items["extentionOption"].Value
			add_colData["leasekikan"] = result.Keiyakudaicho.Items["leasekikan"].Value
			add_colData["leaseexpireymd"] = result.Keiyakudaicho.Items["leaseexpireymd"].Value
			add_colData["keiyakunm"] = result.Keiyakudaicho.Items["keiyakunm"].Value
			add_colData["biko1"] = result.Keiyakudaicho.Items["biko1"].Value
			add_colData["paymentstymd"] = result.Keiyakudaicho.Items["paymentstymd"].Value
			add_colData["paymentcycle"] = result.Keiyakudaicho.Items["paymentcycle"].Value
			add_colData["paymentday"] = result.Keiyakudaicho.Items["paymentday"].Value
			add_colData["paymentcounts"] = result.Keiyakudaicho.Items["paymentcounts"].Value
			add_colData["residualValue"] = result.Keiyakudaicho.Items["residualValue"].Value
			add_colData["rishiritsu"] = result.Keiyakudaicho.Items["rishiritsu"].Value
			add_colData["initialDirectCosts"] = result.Keiyakudaicho.Items["initialDirectCosts"].Value
			add_colData["restorationCosts"] = result.Keiyakudaicho.Items["restorationCosts"].Value
			add_colData["sykshisankeisan"] = result.Keiyakudaicho.Items["sykshisankeisan"].Value
			add_colData["segmentcd"] = result.Keiyakudaicho.Items["segmentcd"].Value
			add_colData["bunruicd"] = result.Keiyakudaicho.Items["bunruicd"].Value
			add_colData["field_viw"] = result.Keiyakudaicho.Items["field_viw"].Value
			add_colData["field_22c"] = result.Keiyakudaicho.Items["field_22c"].Value
			add_colData["field_1av"] = result.Keiyakudaicho.Items["field_1av"].Value
			add_colData["field_206"] = result.Keiyakudaicho.Items["field_206"].Value
			add_colData["field_14l"] = result.Keiyakudaicho.Items["field_14l"].Value
			add_colData["field_7p3"] = result.Keiyakudaicho.Items["field_7p3"].Value
			add_colData["field_248"] = result.Keiyakudaicho.Items["field_248"].Value
			add_colData["field_3k7"] = result.Keiyakudaicho.Items["field_3k7"].Value
			add_colData["field_1vg"] = result.Keiyakudaicho.Items["field_1vg"].Value
			add_colData["field_5fj"] = result.Keiyakudaicho.Items["field_5fj"].Value
			add_colData["field_20h"] = result.Keiyakudaicho.Items["field_20h"].Value
			add_colData["field_2h1"] = result.Keiyakudaicho.Items["field_2h1"].Value
			add_colData["field_qi4"] = result.Keiyakudaicho.Items["field_qi4"].Value
			add_colData["field_1ck"] = result.Keiyakudaicho.Items["field_1ck"].Value
			add_colData["field_u1q"] = result.Keiyakudaicho.Items["field_u1q"].Value
			add_colData["hkkjitenzan"] = result.Keiyakudaicho.Items["hkkjitenzan"].Value
			add_colData["sonnekigaku"] = result.Keiyakudaicho.Items["sonnekigaku"].Value
		}

		// 更新总表中支付信息
		if result.Paymentstatus.Items != nil {
			add_colData["paymentleasefee"] = result.Paymentstatus.Items["paymentleasefee"].Value
		}

		// 更新总表中折旧信息
		if result.Repayment.Items != nil {
			if result.Keiyakuno == tempkeiyakuno && boka != result.Repayment.Items["boka"].Value {
				firstmonth = result.Month
			}
			add_colData["endboka"] = result.Repayment.Items["endboka"].Value
			add_colData["boka"] = result.Repayment.Items["boka"].Value
			add_colData["syokyaku"] = result.Repayment.Items["syokyaku"].Value
			boka = result.Repayment.Items["boka"].Value
		}

		// 更新总表中试算信息
		if result.Paymentinterest.Items != nil {
			balance = result.Paymentinterest.Items["balance"].Value.(float64)
			// firstbalance = result.Paymentinterest.Items["firstbalance"].Value.(float64)
			if firstmonth == result.Month {
				firstbalance = result.Paymentinterest.Items["balance"].Value.(float64) + result.Paymentinterest.Items["repayment"].Value.(float64)
			}
			add_colData["interest"] = result.Paymentinterest.Items["interest"].Value
			add_colData["repayment"] = result.Paymentinterest.Items["repayment"].Value
			if firstbalance == 0 && !midvalue {
				tempfirstbalance = result.Paymentinterest.Items["repayment"].Value.(float64) + result.Paymentinterest.Items["balance"].Value.(float64)
				midvalue = true
			}
			if firstbalance == 0 {
				firstbalance = tempfirstbalance
			}
			add_colData["balance"] = balance
			add_colData["firstbalance"] = firstbalance
		} else {
			if result.Month == firstmonth {
				firstbalance = balance
				midvalue = false
			}
			if result.Keiyakuno != tempkeiyakuno {
				firstbalance = 0
				balance = 0
				midvalue = true
			}
			add_colData["balance"] = balance
			add_colData["firstbalance"] = firstbalance
		}

		update_colData := bson.M{"$set": add_colData}

		_, err = collection.UpdateOne(context.TODO(), query, update_colData)

		tempkeiyakuno = result.Keiyakuno

	}

	return nil
}

// 总表CSV下载
func Download(db, appid, keiyakuno string, year, month int64, stream coldata.ColDataService_DownloadStream) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(ColName(appid))
	ctx, cancel := context.WithTimeout(context.Background(), 30000*time.Second)
	defer cancel()

	// 匹配条件
	match := bson.M{}
	if keiyakuno != "" && year == 0 {
		match = bson.M{
			"keiyakuno": keiyakuno,
		}
	}
	if keiyakuno == "" && year != 0 {
		match = bson.M{
			"year":  year,
			"month": month,
		}
	}
	if keiyakuno != "" && year != 0 {
		match = bson.M{
			"keiyakuno": keiyakuno,
			"year":      year,
			"month":     month,
		}
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
	}

	// 排序
	pipe = append(pipe, bson.M{
		"$sort": bson.D{
			{Key: "keiyakuno", Value: 1},
			{Key: "year", Value: 1},
			{Key: "month", Value: 1},
		},
	})

	pipeJSON, _ := json.Marshal(pipe)
	utils.DebugLog("Download", fmt.Sprintf("pipe: [ %s ]", pipeJSON))

	opt := options.Aggregate()
	opt.SetAllowDiskUse(true)

	cur, err := c.Aggregate(ctx, pipe, opt)
	if err != nil {
		utils.ErrorLog("error Download", err.Error())
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var c ColData
		err := cur.Decode(&c)
		if err != nil {
			utils.ErrorLog("error Download", err.Error())
			return err
		}
		if err := stream.Send(&coldata.DownloadResponse{ColDatas: c.ToProto()}); err != nil {
			utils.ErrorLog("DownloadColDatas", err.Error())
			return err
		}
	}
	return nil
}
