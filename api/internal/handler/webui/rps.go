package webui

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"

	"rxcsoft.cn/pit3/api/internal/common/excelx"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/configx"
	"rxcsoft.cn/pit3/api/internal/common/logic/fieldx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
)

// Prs 报表
type Prs struct{}

// log出力
const (
	ActionDownloadPrs = "DownloadPrs"
)

// DownloadPrs 作成租赁物件本金返还预计表(明细表)数据,以csv文件的方式下载
// @Router /datastores/:d_id/prs/download [POST]
func (r *Prs) DownloadPrs(c *gin.Context) {
	type (
		// FieldItem 字段
		FieldItem struct {
			FieldID   string `json:"field_id" bson:"field_id"`
			FieldName string `json:"field_name" bson:"field_name"`
		}
		// TranslateItem 语言数据
		TranslateItem struct {
			Value string `json:"value" bson:"value"`
			Label string `json:"label" bson:"label"`
		}
		// DownloadRequest 下载
		DownloadRequest struct {
			FieldList     []*FieldItem                `json:"field_list" bson:"field_list"`
			ItemCondition item.ItemsRequest           `json:"item_condition" bson:"item_condition"`
			LookupMap     map[string]*[]TranslateItem `json:"lookup_map" bson:"lookup_map"`
			OptionMap     map[string]*[]TranslateItem `json:"option_map" bson:"option_map"`
		}
		// Keiyaku 契约
		Keiyaku struct {
			Leasekaisha string `json:"leasekaisha" bson:"leasekaisha"`
			Keiyakuno   string `json:"keiyakuno" bson:"keiyakuno"`
		}
		// Leas 利息数据
		Leas struct {
			Leasekaisha string  `json:"Leasekaisha" bson:"Leasekaisha"` // 租赁会社
			Keiyakuno   string  `json:"keiyakuno" bson:"keiyakuno"`     // 契约番号
			Repayment   float64 `json:"repayment" bson:"repayment"`     // 元本返済相当額
			Paymentymd  string  `json:"paymentymd" bson:"paymentymd"`   // 支付年月
		}
		// Payms 支付数据
		Payms struct {
			Leasekaisha     string  `json:"Leasekaisha" bson:"Leasekaisha"`         // 租赁会社
			Keiyakuno       string  `json:"keiyakuno" bson:"keiyakuno"`             // 契约番号
			Paymentleasefee float64 `json:"paymentleasefee" bson:"paymentleasefee"` // 支付金额
			Paymentymd      string  `json:"paymentymd" bson:"paymentymd"`           // 支付年月日
		}
		// 租赁物件本金返还预计表(明细表)
		PRSTable struct {
			Leasekaisha string  `json:"leasekaisha" bson:"leasekaisha"`
			Keiyakuno   string  `json:"keiyakuno" bson:"keiyakuno"`
			Type        string  `json:"type" bson:"type"`
			Repayment1  float64 `json:"repayment1" bson:"repayment1"`
			Repayment2  float64 `json:"repayment2" bson:"repayment2"`
			Repayment21 float64 `json:"repayment21" bson:"repayment21"`
			Repayment22 float64 `json:"repayment22" bson:"repayment22"`
			Repayment23 float64 `json:"repayment23" bson:"repayment23"`
			Repayment24 float64 `json:"repayment24" bson:"repayment24"`
			Repayment5  float64 `json:"repayment5" bson:"repayment5"`
			Repayment26 float64 `json:"repayment26" bson:"repayment26"`
			Repayment27 float64 `json:"repayment27" bson:"repayment27"`
			Repayment28 float64 `json:"repayment28" bson:"repayment28"`
			Repayment29 float64 `json:"repayment29" bson:"repayment29"`
			Repayment30 float64 `json:"repayment30" bson:"repayment30"`
			Repayment31 float64 `json:"repayment31" bson:"repayment31"`
			Repayment32 float64 `json:"repayment32" bson:"repayment32"`
		}
	)
	loggerx.InfoLog(c, ActionDownloadPrs, loggerx.MsgProcessStarted)

	// 参数取得
	jobID := c.Query("job_id")
	datastoreID := c.Param("d_id")
	appID := sessionx.GetCurrentApp(c)
	owners := sessionx.GetUserAccessKeys(c, datastoreID, "R")
	userID := sessionx.GetAuthUserID(c)
	roles := sessionx.GetUserRoles(c)
	domain := sessionx.GetUserDomain(c)
	db := sessionx.GetUserCustomer(c)
	lang := sessionx.GetCurrentLanguage(c)
	encoding := "utf-8"
	fileType := "csv"
	appRoot := "app_" + appID
	// 从body中获取参数
	var request DownloadRequest
	if err := c.BindJSON(&request); err != nil {
		httpx.GinHTTPError(c, ActionDownloadPrs, err)
		return
	}

	// 创建任务
	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "datastore file download(PRS)",
		Origin:       "apps." + appID + ".datastores." + datastoreID,
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "ds-csv-download",
		Steps:        []string{"start", "build-data", "write-to-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	// 契约台账验证
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	var dreq datastore.DatastoreRequest
	dreq.DatastoreId = datastoreID
	dreq.Database = sessionx.GetUserCustomer(c)
	dres, err := datastoreService.FindDatastore(context.TODO(), &dreq)
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadPrs, err)
		return
	}
	if dres.GetDatastore().ApiKey != "keiyakudaicho" {
		httpx.GinHTTPError(c, ActionDownloadPrs, errors.New("non-contractual ledger"))
		return
	}

	// 通过apikey获取支付台账情报
	psDs, err := getDatastoreInfo(db, appID, "paymentStatus")
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadPrs, err)
		return
	}

	// 通过apikey获取利息台账情报
	lsDs, err := getDatastoreInfo(db, appID, "paymentInterest")
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadPrs, err)
		return
	}

	// 处理月度取得
	cfg, err := configx.GetConfigVal(db, appID)
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadPrs, err)
		return
	}
	syoriYmStr := cfg.GetSyoriYm()
	if len(syoriYmStr) == 0 {
		httpx.GinHTTPError(c, ActionDownloadPrs, errors.New("syoriym is not set"))
		return
	}
	// 处理月度转换
	syoriym, err := time.Parse("2006-01", syoriYmStr)
	if err != nil {
		httpx.GinHTTPError(c, ActionDownloadPrs, err)
		return
	}

	// 获取出力台账的字段数据
	fields := fieldx.GetFields(db, lsDs.DatastoreId, appID, roles, false, false)
	// 排序
	sort.Sort(typesx.FieldList(fields))

	// 正式处理开始
	go func() {
		ct := grpc.NewClient(
			grpc.MaxSendMsgSize(100*1024*1024), grpc.MaxRecvMsgSize(100*1024*1024),
		)

		itemService := item.NewItemService("database", ct)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Minute * 10
			o.DialTimeout = time.Minute * 10
		}

		// 发送消息 开始编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_012"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		// 获取所有出力对象租赁契约情报
		kreq := request.ItemCondition
		kreq.DatastoreId = datastoreID
		kreq.AppId = appID
		kreq.Owners = owners
		kreq.Database = db
		keiyakures, err := itemService.FindItems(context.TODO(), &kreq, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}
		var kres []Keiyaku
		for _, tem := range keiyakures.GetItems() {
			var k Keiyaku
			k.Leasekaisha = tem.Items["leasekaishacd"].GetValue()
			k.Keiyakuno = tem.Items["keiyakuno"].GetValue()
			kres = append(kres, k)
		}

		// 获取所有利息情报
		var lreq item.ItemsRequest
		lreq.AppId = appID
		lreq.DatastoreId = lsDs.DatastoreId
		lreq.Owners = owners
		lreq.Database = db
		leaseres, err := itemService.FindItems(context.TODO(), &lreq, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		var lres []Leas
		for _, tem := range leaseres.GetItems() {
			var k Leas
			k.Leasekaisha = tem.Items["leasekaishacd"].GetValue()
			k.Keiyakuno = tem.Items["keiyakuno"].GetValue()
			repayment := tem.Items["repayment"].GetValue()
			repay, err := strconv.ParseFloat(repayment, 64)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "build-data",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			k.Repayment = repay
			k.Paymentymd = tem.Items["paymentymd"].GetValue()
			lres = append(lres, k)
		}

		// 获取所有支付情报
		var preq item.ItemsRequest
		preq.AppId = appID
		preq.DatastoreId = psDs.DatastoreId
		preq.Owners = owners
		preq.Database = db
		payres, err := itemService.FindItems(context.TODO(), &preq, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		var pres []Payms
		for _, tem := range payres.GetItems() {
			var k Payms
			k.Leasekaisha = tem.Items["leasekaishacd"].GetValue()
			k.Keiyakuno = tem.Items["keiyakuno"].GetValue()
			paymentleasefee := tem.Items["paymentleasefee"].GetValue()
			leasefee, err := strconv.ParseFloat(paymentleasefee, 64)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "build-data",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			k.Paymentleasefee = leasefee
			k.Paymentymd = tem.Items["paymentymd"].GetValue()
			pres = append(pres, k)
		}

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_033"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		timestamp := time.Now().Format("20060102150405")

		// 每次2000为一组数据
		total := float64(keiyakures.GetTotal())
		count := math.Ceil(total / 2000)

		// 设定csv头部
		var header []string
		var headers [][]string
		// 添加ID
		//header = append(header, "リース会社")
		header = append(header, "契約番号")
		header = append(header, "類別")
		header = append(header, "一年以内")
		header = append(header, "一年超")
		header = append(header, "一年超-二年以内")
		header = append(header, "二年超-三年以内")
		header = append(header, "三年超-四年以内")
		header = append(header, "四年超-五年以内")
		header = append(header, "五年超")
		headers = append(headers, header)

		excelFile := excelize.NewFile()
		// 创建一个工作表
		index := excelFile.NewSheet("Sheet1")
		// 设置工作簿的默认工作表
		excelFile.SetActiveSheet(index)

		// Excel文件下载
		if fileType == "xlsx" {
			for i, rows := range headers {
				for j, v := range rows {
					y := excelx.GetAxisY(j+1) + strconv.Itoa(i+1)
					excelFile.SetCellValue("Sheet1", y, v)
				}
			}

			for index := 0; index < int(count); index++ {
				startpos := index * 2000
				endpos := (index + 1) * 2000
				if endpos > len(kres) {
					endpos = len(kres)
				}
				var items [][]string
				for i := startpos; i < endpos; i++ {
					// 设置csv行
					var prs PRSTable
					prs.Leasekaisha = kres[i].Leasekaisha
					prs.Keiyakuno = kres[i].Keiyakuno
					// 租赁费用集计
					var itemData []string
					prs.Type = "支払リース料"
					for _, v := range pres {
						if v.Leasekaisha == kres[i].Leasekaisha && v.Keiyakuno == kres[i].Keiyakuno {
							// 支付年月
							paymentym, err := time.Parse("2006-01", v.Paymentymd[0:7])
							if err != nil {
								return
							}
							// 1年以内
							if paymentym.Before(syoriym.AddDate(0, 12, 0)) {
								prs.Repayment1 += v.Paymentleasefee
							}
							// 超过1年
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment2 += v.Paymentleasefee
							}
							// 超过1年-2年以内
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 24, 0))) {
								prs.Repayment21 += v.Paymentleasefee
							}
							// 超过2年-3年以内
							if paymentym.Equal(syoriym.AddDate(0, 24, 0)) || (paymentym.After(syoriym.AddDate(0, 24, 0)) && paymentym.Before(syoriym.AddDate(0, 36, 0))) {
								prs.Repayment22 += v.Paymentleasefee
							}
							// 超过3年-4年以内
							if paymentym.Equal(syoriym.AddDate(0, 36, 0)) || (paymentym.After(syoriym.AddDate(0, 36, 0)) && paymentym.Before(syoriym.AddDate(0, 48, 0))) {
								prs.Repayment23 += v.Paymentleasefee
							}
							// 超过4年-5年以内
							if paymentym.Equal(syoriym.AddDate(0, 48, 0)) || (paymentym.After(syoriym.AddDate(0, 48, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment24 += v.Paymentleasefee
							}
							// 超过5年
							if paymentym.Equal(syoriym.AddDate(0, 60, 0)) || paymentym.After(syoriym.AddDate(0, 60, 0)) {
								prs.Repayment5 += v.Paymentleasefee
							}
						}
					}
					// 编辑集计后租赁费用
					//itemData = append(itemData, prs.Leasekaisha)
					itemData = append(itemData, prs.Keiyakuno)
					itemData = append(itemData, prs.Type)
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment1, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment2, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment21, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment22, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment23, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment24, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment5, 'f', -1, 64))
					// 添加行
					items = append(items, itemData)

					// 元本集计
					var itemData2 []string
					prs.Type = "元本返済相当額"
					for _, v := range lres {
						if v.Leasekaisha == kres[i].Leasekaisha && v.Keiyakuno == kres[i].Keiyakuno {
							// 支付年月
							paymentym, err := time.Parse("2006-01", v.Paymentymd[0:7])
							if err != nil {
								return
							}
							// 1年以内
							if paymentym.Before(syoriym.AddDate(0, 12, 0)) {
								prs.Repayment26 += v.Repayment
							}
							// 超过1年
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment27 += v.Repayment
							}
							// 超过1年-2年以内
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 24, 0))) {
								prs.Repayment28 += v.Repayment
							}
							// 超过2年-3年以内
							if paymentym.Equal(syoriym.AddDate(0, 24, 0)) || (paymentym.After(syoriym.AddDate(0, 24, 0)) && paymentym.Before(syoriym.AddDate(0, 36, 0))) {
								prs.Repayment29 += v.Repayment
							}
							// 超过3年-4年以内
							if paymentym.Equal(syoriym.AddDate(0, 36, 0)) || (paymentym.After(syoriym.AddDate(0, 36, 0)) && paymentym.Before(syoriym.AddDate(0, 48, 0))) {
								prs.Repayment30 += v.Repayment
							}
							// 超过4年-5年以内
							if paymentym.Equal(syoriym.AddDate(0, 48, 0)) || (paymentym.After(syoriym.AddDate(0, 48, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment31 += v.Repayment
							}
							// 超过5年
							if paymentym.Equal(syoriym.AddDate(0, 60, 0)) || paymentym.After(syoriym.AddDate(0, 60, 0)) {
								prs.Repayment32 += v.Repayment
							}
						}
					}
					// 编辑集计后元本
					//itemData2 = append(itemData2, prs.Leasekaisha)
					itemData2 = append(itemData2, prs.Keiyakuno)
					itemData2 = append(itemData2, prs.Type)
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment26, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment27, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment28, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment29, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment30, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment31, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment32, 'f', -1, 64))
					// 添加行
					items = append(items, itemData2)
				}

				for k, rows := range items {
					for j, v := range rows {
						y := excelx.GetAxisY(j+1) + strconv.Itoa(int((index*2000 + k + 3)))
						excelFile.SetCellValue("Sheet1", y, v)
					}
				}
			}

			outFile := "text.xlsx"

			if err := excelFile.SaveAs(outFile); err != nil {
				fmt.Println(err)
			}

			fo, err := os.Open(outFile)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			defer fo.Close()

			// 写入文件到 minio
			minioClient, err := storagecli.NewClient(domain)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			filePath := path.Join("csv", "Prs_"+timestamp+".csv")
			path, err := minioClient.SavePublicObject(fo, filePath, "text/csv")
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			// 判断顾客上传文件是否在设置的最大存储空间以内
			canUpload := filex.CheckCanUpload(domain, float64(path.Size))
			if canUpload {
				// 如果没有超出最大值，就对顾客的已使用大小进行累加
				err = filex.ModifyUsedSize(domain, float64(path.Size))
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 保存文件失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "save-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
			} else {
				// 如果已达上限，则删除刚才上传的文件
				minioClient.DeleteObject(path.Name)
				path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     i18n.Tr(lang, "job.J_007"),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			os.Remove(outFile)

			// 发送消息 写入保存文件成功，返回下载路径，任务结束
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_028"),
				CurrentStep: "end",
				File: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

		} else {
			var writer *csv.Writer

			// 写入文件到本地
			filename := "temp/tmp" + "_" + timestamp + "_header" + ".csv"
			f, err := os.Create(filename)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "write-to-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			defer f.Close()

			if encoding == "sjis" {
				converter := transform.NewWriter(f, japanese.ShiftJIS.NewEncoder())
				writer = csv.NewWriter(converter)
			} else {
				writer = csv.NewWriter(f)
				// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
				headers[0][0] = "\xEF\xBB\xBF" + headers[0][0]
			}

			err = writer.WriteAll(headers)
			if err != nil {
				if err.Error() == "encoding: rune not supported by encoding." {
					path := filex.WriteAndSaveFile(domain, appID, []string{"現在のタイトルには、日本語の[shift-jis]エンコード以外の文字が含まれており、実行を続行できません。"})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "write-to-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}

				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 获取数据失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "write-to-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			writer.Flush() // 此时才会将缓冲区数据写入

			for index := int64(0); index < int64(count); index++ {
				startpos := index * 2000
				endpos := (index + 1) * 2000
				if endpos > int64(len(kres)) {
					endpos = int64(len(kres))
				}
				var items [][]string
				for i := startpos; i < endpos; i++ {
					// 设置csv行
					var prs PRSTable
					prs.Leasekaisha = kres[i].Leasekaisha
					prs.Keiyakuno = kres[i].Keiyakuno
					// 租赁费用集计
					var itemData []string
					prs.Type = "支払リース料"
					for _, v := range pres {
						if v.Leasekaisha == kres[i].Leasekaisha && v.Keiyakuno == kres[i].Keiyakuno {
							// 支付年月
							paymentym, err := time.Parse("2006-01", v.Paymentymd[0:7])
							if err != nil {
								return
							}
							// 1年以内
							if paymentym.Before(syoriym.AddDate(0, 12, 0)) {
								prs.Repayment1 += v.Paymentleasefee
							}
							// 超过1年
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment2 += v.Paymentleasefee
							}
							// 超过1年-2年以内
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 24, 0))) {
								prs.Repayment21 += v.Paymentleasefee
							}
							// 超过2年-3年以内
							if paymentym.Equal(syoriym.AddDate(0, 24, 0)) || (paymentym.After(syoriym.AddDate(0, 24, 0)) && paymentym.Before(syoriym.AddDate(0, 36, 0))) {
								prs.Repayment22 += v.Paymentleasefee
							}
							// 超过3年-4年以内
							if paymentym.Equal(syoriym.AddDate(0, 36, 0)) || (paymentym.After(syoriym.AddDate(0, 36, 0)) && paymentym.Before(syoriym.AddDate(0, 48, 0))) {
								prs.Repayment23 += v.Paymentleasefee
							}
							// 超过4年-5年以内
							if paymentym.Equal(syoriym.AddDate(0, 48, 0)) || (paymentym.After(syoriym.AddDate(0, 48, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment24 += v.Paymentleasefee
							}
							// 超过5年
							if paymentym.Equal(syoriym.AddDate(0, 60, 0)) || paymentym.After(syoriym.AddDate(0, 60, 0)) {
								prs.Repayment5 += v.Paymentleasefee
							}
						}
					}
					// 编辑集计后租赁费用
					//itemData = append(itemData, prs.Leasekaisha)
					itemData = append(itemData, prs.Keiyakuno)
					itemData = append(itemData, prs.Type)
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment1, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment2, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment21, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment22, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment23, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment24, 'f', -1, 64))
					itemData = append(itemData, strconv.FormatFloat(prs.Repayment5, 'f', -1, 64))
					// 添加行
					items = append(items, itemData)

					// 元本集计
					var itemData2 []string
					prs.Type = "元本返済相当額"
					for _, v := range lres {
						if v.Leasekaisha == kres[i].Leasekaisha && v.Keiyakuno == kres[i].Keiyakuno {
							// 支付年月
							paymentym, err := time.Parse("2006-01", v.Paymentymd[0:7])
							if err != nil {
								return
							}
							// 1年以内
							if paymentym.Before(syoriym.AddDate(0, 12, 0)) {
								prs.Repayment26 += v.Repayment
							}
							// 超过1年
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment27 += v.Repayment
							}
							// 超过1年-2年以内
							if paymentym.Equal(syoriym.AddDate(0, 12, 0)) || (paymentym.After(syoriym.AddDate(0, 12, 0)) && paymentym.Before(syoriym.AddDate(0, 24, 0))) {
								prs.Repayment28 += v.Repayment
							}
							// 超过2年-3年以内
							if paymentym.Equal(syoriym.AddDate(0, 24, 0)) || (paymentym.After(syoriym.AddDate(0, 24, 0)) && paymentym.Before(syoriym.AddDate(0, 36, 0))) {
								prs.Repayment29 += v.Repayment
							}
							// 超过3年-4年以内
							if paymentym.Equal(syoriym.AddDate(0, 36, 0)) || (paymentym.After(syoriym.AddDate(0, 36, 0)) && paymentym.Before(syoriym.AddDate(0, 48, 0))) {
								prs.Repayment30 += v.Repayment
							}
							// 超过4年-5年以内
							if paymentym.Equal(syoriym.AddDate(0, 48, 0)) || (paymentym.After(syoriym.AddDate(0, 48, 0)) && paymentym.Before(syoriym.AddDate(0, 60, 0))) {
								prs.Repayment31 += v.Repayment
							}
							// 超过5年
							if paymentym.Equal(syoriym.AddDate(0, 60, 0)) || paymentym.After(syoriym.AddDate(0, 60, 0)) {
								prs.Repayment32 += v.Repayment
							}
						}
					}
					// 编辑集计后元本
					//itemData2 = append(itemData2, prs.Leasekaisha)
					itemData2 = append(itemData2, prs.Keiyakuno)
					itemData2 = append(itemData2, prs.Type)
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment26, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment27, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment28, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment29, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment30, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment31, 'f', -1, 64))
					itemData2 = append(itemData2, strconv.FormatFloat(prs.Repayment32, 'f', -1, 64))
					// 添加行
					items = append(items, itemData2)
				}

				err = writer.WriteAll(items)
				if err != nil {
					if err.Error() == "encoding: rune not supported by encoding." {
						path := filex.WriteAndSaveFile(domain, appID, []string{"現在のデータには、日本語の[shift-jis]エンコーディング以外の文字があり、実行を続行できません。"})
						// 发送消息 获取数据失败，终止任务
						jobx.ModifyTask(task.ModifyRequest{
							JobId:       jobID,
							Message:     err.Error(),
							CurrentStep: "write-to-file",
							EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
							ErrorFile: &task.File{
								Url:  path.MediaLink,
								Name: path.Name,
							},
							Database: db,
						}, userID)

						return
					}

					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 获取数据失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "write-to-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
				writer.Flush() // 此时才会将缓冲区数据写入
			}

			f.Close()

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_029"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			// 发送消息 写入文件成功，开始保存文档到文件服务器
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_043"),
				CurrentStep: "save-file",
				Database:    db,
			}, userID)

			fo, err := os.Open(filename)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			defer func() {
				fo.Close()
				os.Remove(filename)
			}()

			// 写入文件到 minio
			minioClient, err := storagecli.NewClient(domain)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			filePath := path.Join(appRoot, "csv", "Prs_"+timestamp+".csv")
			path, err := minioClient.SavePublicObject(fo, filePath, "text/csv")
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}
			// 判断顾客上传文件是否在设置的最大存储空间以内
			canUpload := filex.CheckCanUpload(domain, float64(path.Size))
			if canUpload {
				// 如果没有超出最大值，就对顾客的已使用大小进行累加
				err = filex.ModifyUsedSize(domain, float64(path.Size))
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 保存文件失败，终止任务
					jobx.ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "save-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)

					return
				}
			} else {
				// 如果已达上限，则删除刚才上传的文件
				minioClient.DeleteObject(path.Name)
				path := filex.WriteAndSaveFile(domain, appID, []string{"最大ストレージ容量に達しました。ファイルのアップロードに失敗しました"})
				// 发送消息 保存文件失败，终止任务
				jobx.ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     i18n.Tr(lang, "job.J_007"),
					CurrentStep: "save-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)

				return
			}

			// 发送消息 写入保存文件成功，返回下载路径，任务结束
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     i18n.Tr(lang, "job.J_028"),
				CurrentStep: "end",
				File: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				EndTime:  time.Now().UTC().Format("2006-01-02 15:04:05"),
				Database: db,
			}, userID)

		}

	}()

	loggerx.InfoLog(c, ActionDownloadPrs, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, ActionDownloadPrs)),
		Data:    gin.H{},
	})
}

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
