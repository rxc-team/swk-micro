package webui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/kataras/i18n"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/stringx"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/print"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// log出力
const (
	PrintProcessName = "Print"
	ActionPrintList  = "PrintList"
)

// PrintList 获取台账PDF打印设置,将台账数据按打印获取的设置打印生成PDF文件
// @Router /datastores/{d_id}/items/print [POST]
func (i *Item) PrintList(c *gin.Context) {

	loggerx.InfoLog(c, ActionPrintList, loggerx.MsgProcessStarted)

	datastoreID := c.Param("d_id")
	appID := sessionx.GetCurrentApp(c)
	userName := sessionx.GetUserName(c)
	owners := sessionx.GetUserAccessKeys(c, datastoreID, "R")
	lang := sessionx.GetCurrentLanguage(c)
	domain := sessionx.GetUserDomain(c)
	db := sessionx.GetUserCustomer(c)
	userID := sessionx.GetAuthUserID(c)
	appRoot := "app_" + appID
	jobID := "job_" + time.Now().Format("20060102150405")

	// 从body中获取参数
	var req item.ItemsRequest
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionPrintList, err)
		return
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		httpx.GinHTTPError(c, ActionPrintList, err)
		return
	}

	// 台账情报获取
	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)
	var dReq datastore.DatastoreRequest
	dReq.DatastoreId = datastoreID
	dReq.Database = db
	dResp, err := datastoreService.FindDatastore(context.TODO(), &dReq)
	if err != nil {
		httpx.GinHTTPError(c, ActionPrintList, err)
		return
	}

	jobx.CreateTask(task.AddRequest{
		JobId:        jobID,
		JobName:      "datastore pdf generation",
		Origin:       dResp.GetDatastore().GetDatastoreName(),
		UserId:       userID,
		ShowProgress: false,
		Message:      i18n.Tr(lang, "job.J_014"),
		TaskType:     "ds-pdf-generation",
		Steps:        []string{"start", "get-print-config", "build-data", "write-to-file", "save-file", "end"},
		CurrentStep:  "start",
		Database:     db,
		AppId:        appID,
	})

	go func() {
		printService := print.NewPrintService("database", client.DefaultClient)

		// 发送消息 开始获取PDF打印设置
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_026"),
			CurrentStep: "get-print-config",
			Database:    db,
		}, userID)

		var preq print.FindPrintRequest
		preq.AppId = appID
		preq.Database = db
		preq.DatastoreId = datastoreID

		printResp, err := printService.FindPrint(context.TODO(), &preq)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取PDF打印设置失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "get-print-config",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)

			return
		}

		// 发送消息 开始获取并编辑数据
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_025"),
			CurrentStep: "build-data",
			Database:    db,
		}, userID)

		printConfig := printResp.GetPrint()

		itemService := item.NewItemService("database", client.DefaultClient)

		var opss client.CallOption = func(o *client.CallOptions) {
			o.RequestTimeout = time.Hour * 1
			o.DialTimeout = time.Hour * 1
		}

		cReq := item.CountRequest{
			AppId:         appID,
			DatastoreId:   datastoreID,
			ConditionList: req.ConditionList,
			ConditionType: req.ConditionType,
			Owners:        owners,
			Database:      db,
		}

		cResp, err := itemService.FindCount(context.TODO(), &cReq, opss)
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

		// 获取当前app的语言数据
		langData := langx.GetLanguageData(db, lang, domain)
		dsName := langx.GetLangValue(langData, dResp.GetDatastore().GetDatastoreName(), langx.DefaultResult)

		// 打印数据上限为500件,超出500件则不打印直接返回错误消息
		if cResp.GetTotal() > 500 {
			path := filex.WriteAndSaveFile(domain, appID, []string{"The maximum print data volume supported by PDF is 500"})
			// 发送消息 获取数据失败，终止任务
			jobx.ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     "PDFでサポートされる最大印刷データ量は500です",
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

		// 发送消息 数据编辑完成，开始写入文件
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_034"),
			CurrentStep: "write-to-file",
			Database:    db,
		}, userID)

		page := printConfig.Page
		orientation := printConfig.Orientation

		// PDF 作成
		pdf := gofpdf.New(orientation, "mm", page, "")
		pdf.AddUTF8Font("HanaMinA", "B", "assets/font/HanaMinA.ttf")
		pdf.SetFont("HanaMinA", "B", 9)
		pdf.SetTopMargin(16)

		var maxCells int64 = 1

		for _, cf := range printConfig.Fields {
			if cf.GetX()+1 > maxCells {
				maxCells = cf.GetX() + 1
			}
		}

		// 从path中获取参数
		req.DatastoreId = datastoreID
		req.AppId = appID
		// 取消分组(2000/组)打印,设置上限500
		req.PageIndex = 1
		req.PageSize = 500
		// 从共通中获取参数
		req.AppId = appID
		req.Owners = owners
		req.ShowLookup = true
		req.Database = db

		response, err := itemService.FindItems(context.TODO(), &req, opss)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 数据写入文件失败，终止任务
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

		if len(response.GetItems()) == 0 {
			return
		}

		for _, dt := range response.GetItems() {
			printOnePage(maxCells, dsName, userName, pdf, printConfig, dt, langData, minioClient.GetEndpoint())
		}

		// 发送消息 写入文件成功，开始保存文档到文件服务器
		jobx.ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     i18n.Tr(lang, "job.J_043"),
			CurrentStep: "save-file",
			Database:    db,
		}, userID)

		timestamp := time.Now().Format("20060102150405")

		filex.Mkdir("temp/")

		pdfFileName := path.Join("temp", "datastore_"+timestamp+".pdf")

		err = pdf.OutputFileAndClose(pdfFileName)
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

		fo, err := os.Open(pdfFileName)
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

		filePath := path.Join(appRoot, "pdf", userID+"_datastore_"+timestamp+".pdf")
		path, err := minioClient.SavePublicObject(fo, filePath, "application/pdf")
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

		fo.Close()
		os.Remove(pdfFileName)

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

	}()

	// 设置下载的文件名
	loggerx.InfoLog(c, ActionPrintList, loggerx.MsgProcessEnded)
	// 设置文件类型以及输出数据
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I004, fmt.Sprintf(httpx.Temp, PrintProcessName, ActionPrintList)),
		Data:    gin.H{},
	})
}

func printOnePage(maxCells int64, ds, userName string, pdf *gofpdf.Fpdf, config *print.Print, it *item.Item, langData *language.Language, endpoint string) {
	width, height := pdf.GetPageSize()

	title := ds
	printTime := "台帳出力時間: " + time.Now().Format("2006-01-02 15:04:05")
	link := "https://pitdemo.ml/datastores/" + it.GetDatastoreId() + "/items/" + it.GetItemId()

	// 添加一个新页面

	left, top, right, bottom := pdf.GetMargins()
	pdf.SetHeaderFuncMode(func() {

		pdf.SetXY(0, 0)
		pdf.SetFont("HanaMinA", "B", 12)
		pdf.CellFormat(width, 14, title, "0", 0, "C", false, pdf.AddLink(), link)
		if config.ShowSystem {
			// // 显示二维码
			// key2 := barcode.RegisterQR(pdf, link, qr.H, qr.Unicode)
			// var width1 float64 = 12
			// var height1 float64 = 12
			// barcode.BarcodeUnscalable(pdf, key2, width-width1-right, 2, &width1, &height1, false)

			// 设置帳票出力ID
			pdf.SetFont("HanaMinA", "B", 8)
			pdf.Text(width-100, 6, "帳票出力ID: "+it.GetItemId())

			// 设置最終更新日時
			pdf.SetFont("HanaMinA", "B", 8)
			pdf.Text(width-100, 12, "最終更新日時:"+it.UpdatedAt[:10])
		}
		// 显示打印时间
		pdf.SetXY(0, 8)
		pdf.SetFont("HanaMinA", "B", 8)
		pdf.CellFormat(width, 10, printTime, "0", 0, "C", false, 0, "")
		pdf.Ln(20)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.CellFormat(0, 10, fmt.Sprintf("-%d/{nb}-", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("")
	pdf.AddPage()

	_, lineHeight := pdf.GetFontSize()
	lineHeight += 3

	cellWidth := (width - left - right) / float64(maxCells)
	var titleWidth float64 = float64(config.TitleWidth)

	itemMap := it.Items

	pdf.SetFillColor(230, 247, 255)

	// File 字段的值
	type File struct {
		URL  string `json:"url" bson:"url"`
		Name string `json:"name" bson:"name"`
	}

	// 循环打印配置，设置内容
	for _, cf := range config.Fields {

		name := langx.GetLangValue(langData, cf.GetFieldName(), langx.DefaultResult)

		x := float64(cf.X)*cellWidth + left
		y := float64(cf.Y)*lineHeight + top

		tx := x + 2
		ty := y + 4
		vx := x + 2 + titleWidth
		vy := y + 4

		width := float64(cf.Cols) * cellWidth
		height := float64(cf.Rows) * lineHeight

		data, exist := itemMap[cf.FieldId]
		if !exist {
			if cf.AsTitle {
				pdf.SetTextColor(89, 159, 231)
				pdf.Rect(x, y, width, height, "FD")
				pdf.Text(tx, ty, "▼"+name)
				pdf.SetTextColor(0, 0, 0)
				continue
			}
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name)
			continue
		}

		dataType := data.DataType
		value := data.Value

		_, fontSize := pdf.GetFontSize()

		words := (width - titleWidth) / fontSize
		words = math.Floor(words)

		switch dataType {
		case "text":
			if cf.AsTitle {
				pdf.SetTextColor(89, 159, 231)
				pdf.Rect(x, y, width, height, "FD")
				pdf.Text(tx, ty, "▼"+name)
				pdf.SetTextColor(0, 0, 0)
			} else {
				pdf.Rect(x, y, width, height, "D")
				pdf.Rect(x, y, titleWidth, height, "D")
				pdf.Text(tx, ty, name+":")
				pdf.Text(vx, vy, stringx.AddEllipsis(value, words, int(cf.Rows)))
			}
		case "textarea":
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.SetXY(vx, y)
			words := (width - titleWidth - 2) / fontSize
			words = math.Floor(words)
			pdf.MultiCell(width-titleWidth-2, lineHeight, stringx.AddEllipsis(value, words-1, int(cf.Rows)), "", "L", false)
			// pdf.Text(vx, vy, stringx.AddEllipsis(value, words))
		case "number":
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.Text(vx, vy, stringx.AddEllipsis(value, words, int(cf.Rows)))
		case "date":
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.Text(vx, vy, stringx.AddEllipsis(value, words, int(cf.Rows)))
		case "time":
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.Text(vx, vy, stringx.AddEllipsis(value, words, int(cf.Rows)))
		case "switch":
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.Text(vx, vy, stringx.AddEllipsis(value, words, int(cf.Rows)))
		case "user":
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			var us []string
			err := json.Unmarshal([]byte(value), &us)
			if err == nil {
				pdf.Text(vx, vy, stringx.AddEllipsis(strings.Join(us, ","), words, int(cf.Rows)))
			}

		case "options":
			ov := langx.GetLangValue(langData, value, langx.DefaultResult)
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.Text(vx, vy, stringx.AddEllipsis(ov, words, int(cf.Rows)))
		case "lookup":
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.Text(vx, vy, stringx.AddEllipsis(value, words, int(cf.Rows)))
		case "file":
			pdf.Rect(x, y, width, height, "D")

			if cf.IsImage {
				var fs []File
				err := json.Unmarshal([]byte(value), &fs)
				if err == nil {
					mx := tx
					for _, fl := range fs {
						filex.Mkdir("temp/")
						localURL := path.Join("temp", fl.Name)
						scale, err := downloadFile(fl.URL, localURL, endpoint)
						if err == nil {
							mw := (height - 2) * scale
							pdf.Image(localURL, mx, y+1, mw, height-2, false, "", 0, "")
							mx += mw + 1
						}

						defer os.Remove(localURL)
					}
				}

			} else {
				var fs []File
				err := json.Unmarshal([]byte(value), &fs)
				if err == nil {
					pdf.Rect(x, y, titleWidth, height, "D")
					pdf.Text(tx, ty, name+":")
					fn := strings.Builder{}
					for _, fl := range fs {
						fn.WriteString(fl.Name)
					}
					pdf.Text(vx, vy, stringx.AddEllipsis(fn.String(), words, int(cf.Rows)))
				}
			}
		default:
			pdf.Rect(x, y, width, height, "D")
			pdf.Rect(x, y, titleWidth, height, "D")
			pdf.Text(tx, ty, name+":")
			pdf.Text(vx, vy, stringx.AddEllipsis(value, words, int(cf.Rows)))
		}
	}

	checkField := config.CheckField
	// 打印盘点图片
	if len(checkField) > 0 {
		value := itemMap[checkField].GetValue()

		var fs []File
		err := json.Unmarshal([]byte(value), &fs)
		if err == nil && len(fs) > 0 {
			checkFile := fs[len(fs)-1]
			filex.Mkdir("temp/")
			localURL := path.Join("temp", checkFile.Name)
			scale, err := downloadFile(checkFile.URL, localURL, endpoint)
			if err == nil {
				mx := left + 150 + 10
				my := height - bottom - 50
				pdf.Image(localURL, mx, my, 50*scale, 50, false, "", 0, "")
			}

			defer os.Remove(localURL)
		}
	}

	if config.ShowSign {

		var signWidth float64 = 30
		var signTitleHeight float64 = 8
		var signHeight float64 = 16

		signLeft := left
		signTop := height - signTitleHeight - signHeight - bottom - 26

		for i := float64(1); i < 6; i++ {
			// 打印标题
			x := ((i - 1) * signWidth) + signLeft
			y := signTop
			tx := x + 2
			ty := y + 6

			pdf.Rect(x, y, signWidth, signTitleHeight, "D")
			pdf.Text(tx, ty, config.SignName1+strconv.FormatFloat(i, 'f', -1, 64))
			// 打印空白格子
			bx := x
			by := y + signTitleHeight

			pdf.Rect(bx, by, signWidth, signHeight, "D")
		}

		signTop += 26
		for i := float64(1); i < 6; i++ {
			// 打印标题
			x := ((i - 1) * signWidth) + signLeft
			y := signTop
			tx := x + 2
			ty := y + 6

			pdf.Rect(x, y, signWidth, signTitleHeight, "D")
			pdf.Text(tx, ty, config.SignName2+strconv.FormatFloat(i, 'f', -1, 64))
			// 打印空白格子
			bx := x
			by := y + signTitleHeight

			pdf.Rect(bx, by, signWidth, signHeight, "D")
		}

	}

}

func downloadFile(URL, fileName, endpoint string) (float64, error) {

	//Get the response bytes from the url
	URL = strings.Replace(URL, "/storage/", "/", -1)

	// 获取DB配置文件
	URL = "http://" + endpoint + URL

	response, err := http.Get(URL)
	if err != nil {
		loggerx.ErrorLog("downloadFile", err.Error())
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return 0, errors.New("received non 200 response code")
	}

	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		loggerx.ErrorLog("downloadFile", err.Error())
		return 0, err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		loggerx.ErrorLog("downloadFile", err.Error())
		return 0, err
	}

	of, err := os.Open(fileName)
	if err != nil {
		loggerx.ErrorLog("downloadFile", err.Error())
		return 0, err
	}

	im, _, err := image.DecodeConfig(of)
	if err != nil {
		loggerx.ErrorLog("downloadFile", err.Error())
		return 1, err
	}

	scale := float64(im.Width) / float64(im.Height)

	return scale, nil
}
