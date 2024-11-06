package leasex

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/configx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
	"rxcsoft.cn/pit3/srv/database/proto/template"
)

// InsertPay 生成支付数据
func InsertPay(db, appID, userID string, dsMap map[string]string, ps []typesx.Payment, insert bool) (result *typesx.InsertResult, err error) {
	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &typesx.InsertResult{}

	var tplItems typesx.TplData

	// 支付情报
	var payData []*template.ListItems
	for _, pay := range ps {
		payyear, _ := strconv.Atoi(pay.Paymentymd[0:4])
		paymonth, _ := strconv.Atoi(pay.Paymentymd[5:7])
		items := make(map[string]*template.Value)
		items["paymentcount"] = &template.Value{
			DataType: "number",
			Value:    strconv.Itoa(pay.Paymentcount),
		}
		items["paymentType"] = &template.Value{
			DataType: "text",
			Value:    pay.PaymentType,
		}
		items["paymentymd"] = &template.Value{
			DataType: "date",
			Value:    pay.Paymentymd,
		}
		items["paymentleasefee"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Paymentleasefee, 'f', -1, 64),
		}
		items["paymentleasefeehendo"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Paymentleasefeehendo, 'f', -1, 64),
		}
		items["incentives"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Incentives, 'f', -1, 64),
		}
		items["sonotafee"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Sonotafee, 'f', -1, 64),
		}
		items["kaiyakuson"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Kaiyakuson, 'f', -1, 64),
		}
		items["year"] = &template.Value{
			DataType: "number",
			Value:    strconv.Itoa(payyear),
		}
		items["month"] = &template.Value{
			DataType: "number",
			Value:    strconv.Itoa(paymonth),
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  dsMap["paymentStatus"],
			DatastoreKey: "paymentStatus",
			TemplateId:   templateID,
		}
		payData = append(payData, &data)
	}

	tplItems = append(tplItems, payData...)

	if insert {
		// 支付数据存入临时集合
		tplService := template.NewTemplateService("database", client.DefaultClient)

		var hsReq template.MutilAddRequest

		// 从body中获取参数
		hsReq.Data = tplItems
		// 从共通中获取参数
		hsReq.Writer = userID
		hsReq.Database = db
		hsReq.Collection = userID
		hsReq.AppId = appID

		_, err = tplService.MutilAddTemplateItem(context.TODO(), &hsReq)
		if err != nil {
			loggerx.ErrorLog("insertPay", err.Error())
			return nil, err
		}
	}

	result.TemplateID = templateID
	result.TplItems = tplItems

	return result, nil
}

// Compute 计算利息和偿还数据(租赁系统用)
func Compute(db, appID, userID string, p typesx.LRParam, insert bool) (result *typesx.ComputeResult, err error) {

	// 使用権資産簿価
	if p.Sykshisankeisan == "1" {
		// 開始時点から計算

		// 生成临时数据ID
		uid := uuid.Must(uuid.NewRandom())
		templateID := uid.String()

		result = &typesx.ComputeResult{}

		var tplItems typesx.TplData

		// 支付数据合法性检查
		checkErr := payDataValidCheck(p.CancellationRightOption, p.Payments)
		if checkErr != nil {
			loggerx.ErrorLog("compute", checkErr.Error())
			return nil, checkErr
		}
		// 残价保证额
		residualValue := p.ResidualValue
		// 割引率
		rishiritsu := p.Rishiritsu
		// 租赁开始日
		leasestymd := p.Leasestymd
		// 比較開始期首月
		firstMonth := p.FirstMonth
		var firstMonthB, err = time.Parse("2006-1-02", firstMonth+"-01")
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		// 租赁期间
		leasekikan := p.Leasekikan
		// 延长租赁期间
		extentionOption := p.ExtentionOption
		// 租赁总期间算出(租赁期间 + 延长租赁期间)
		leasekikanTotal := leasekikan + extentionOption
		// 耐用年限(月单位)
		// yms := p.Assetlife * 12
		// 減価償却期間算出
		var genkakikan = leasekikanTotal
		/* genkakikan := yms
		if p.Torihikikbn != "1" {
			// 移転外
			if genkakikan > leasekikanTotal {
				// 取租赁期间与耐用期间中较短者
				genkakikan = leasekikanTotal
			}
		} */

		// 比較開始時点から計算
		hkkjitenzan, presentTotalRemain := getLeaseDebt(p.Payments, rishiritsu, firstMonth, residualValue)
		// 元本残高相当额(初回 = 租赁负债额) => 租赁负债额算出(现在价值累计)
		var principalAmount float64 = presentTotalRemain

		var leaseData []typesx.Lease
		var repayData []typesx.RePayment

		// 根据支付年月支付周期等情报对支付情报再整理
		payments, err := getArrangedPays(p.Payments)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}

		// **********利息情报算出**********
		leases, err := getLeaseDataStart(payments, firstMonthB, firstMonthB, principalAmount, rishiritsu)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		leaseData = append(leaseData, leases...)

		// **********偿还情报算出**********
		// 前払リース料
		//paymentsAtOrPrior := p.PaymentsAtOrPrior
		// リース・インセンティブ（前払）
		//incentivesAtOrPrior := p.IncentivesAtOrPrior
		// 当初直接費用
		//initialDirectCosts := p.InitialDirectCosts
		// 原状回復コスト
		//restorationCosts := p.RestorationCosts
		// 初期期首簿価 = 现在价值合计+ 当初直接費用 + 原状回復コスト
		//var boka float64 = principalAmount + initialDirectCosts + restorationCosts
		var boka = principalAmount

		// 原始取得价值 = 初期期首簿価
		result.KiSyuBoka = boka

		// 偿还情报算出处理
		repays, err := getRepayDataStart(firstMonthB, genkakikan, residualValue, boka, appID, db, leasestymd)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		repayData = append(repayData, repays...)

		// 处理月度取得
		cfg, err := configx.GetConfigVal(db, appID)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		syoriYm := cfg.GetSyoriYm()

		// 支付情报
		var payData []*template.ListItems
		for _, pay := range p.Payments {
			payyear, _ := strconv.Atoi(pay.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(pay.Paymentymd[5:7])
			items := make(map[string]*template.Value)
			items["paymentcount"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(pay.Paymentcount),
			}
			items["paymentType"] = &template.Value{
				DataType: "text",
				Value:    pay.PaymentType,
			}
			items["paymentymd"] = &template.Value{
				DataType: "date",
				Value:    pay.Paymentymd,
			}
			items["paymentleasefee"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Paymentleasefee, 'f', -1, 64),
			}
			items["paymentleasefeehendo"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Paymentleasefeehendo, 'f', -1, 64),
			}
			items["incentives"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Incentives, 'f', -1, 64),
			}
			items["sonotafee"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Sonotafee, 'f', -1, 64),
			}
			items["kaiyakuson"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Kaiyakuson, 'f', -1, 64),
			}
			items["year"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}

			data := template.ListItems{
				Items:        items,
				DatastoreId:  p.DsMap["paymentStatus"],
				DatastoreKey: "paymentStatus",
				TemplateId:   templateID,
			}

			payData = append(payData, &data)
		}

		tplItems = append(tplItems, payData...)

		// 利息情报
		var lsData []*template.ListItems
		for _, lease := range leaseData {
			payyear, _ := strconv.Atoi(lease.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(lease.Paymentymd[5:7])
			items := make(map[string]*template.Value)
			items["interest"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Interest, 'f', -1, 64),
			}
			items["repayment"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Repayment, 'f', -1, 64),
			}
			items["balance"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Balance, 'f', -1, 64),
			}
			items["firstbalance"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Firstbalance, 'f', -1, 64),
			}
			items["present"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Present, 'f', -1, 64),
			}
			items["paymentymd"] = &template.Value{
				DataType: "date",
				Value:    lease.Paymentymd,
			}
			items["year"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}

			data := template.ListItems{
				Items:        items,
				DatastoreId:  p.DsMap["paymentInterest"],
				DatastoreKey: "paymentInterest",
				TemplateId:   templateID,
			}

			lsData = append(lsData, &data)
		}

		tplItems = append(tplItems, lsData...)
		// 処理月度の先月までの償却費の累計額
		var preDepreciationTotal float64 = 0
		// 偿还情报
		var rpData []*template.ListItems
		for _, rp := range repayData {
			rpyear, _ := strconv.Atoi(rp.Syokyakuymd[0:4])
			rpmonth, _ := strconv.Atoi(rp.Syokyakuymd[5:7])
			items := make(map[string]*template.Value)
			items["endboka"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Endboka, 'f', -1, 64),
			}
			items["boka"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Boka, 'f', -1, 64),
			}
			items["syokyaku"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Syokyaku, 'f', -1, 64),
			}
			items["syokyakuymd"] = &template.Value{
				DataType: "date",
				Value:    rp.Syokyakuymd,
			}
			items["syokyakukbn"] = &template.Value{
				DataType: "text",
				Value:    rp.Syokyakukbn,
			}
			items["year"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpyear),
			}
			items["month"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpmonth),
			}

			data := template.ListItems{
				Items:        items,
				DatastoreId:  p.DsMap["repayment"],
				DatastoreKey: "repayment",
				TemplateId:   templateID,
			}

			if rp.Syokyakuymd[:7] < syoriYm {
				preDepreciationTotal += rp.Syokyaku
			}

			rpData = append(rpData, &data)
		}

		tplItems = append(tplItems, rpData...)

		// 履历情报
		var hsData []*template.ListItems

		items := make(map[string]*template.Value)
		items["leaseTotal"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(hkkjitenzan, 'f', -1, 64),
		}
		items["presentTotal"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(presentTotalRemain, 'f', -1, 64),
		}
		items["preDepreciationTotal"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(preDepreciationTotal, 'f', -1, 64),
		}
		items["hkkjitenzan"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(hkkjitenzan, 'f', -1, 64),
		}
		items["sonnekigaku"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(0, 'f', -1, 64),
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["rireki"],
			DatastoreKey: "rireki",
			TemplateId:   templateID,
		}

		hsData = append(hsData, &data)

		tplItems = append(tplItems, hsData...)

		if insert {
			// 支付数据存入临时集合
			tplService := template.NewTemplateService("database", client.DefaultClient)

			var hsReq template.MutilAddRequest

			// 从body中获取参数
			hsReq.Data = tplItems
			// 从共通中获取参数
			hsReq.Writer = userID
			hsReq.Database = db
			hsReq.Collection = userID
			hsReq.AppId = appID

			_, err = tplService.MutilAddTemplateItem(context.TODO(), &hsReq)
			if err != nil {
				loggerx.ErrorLog("compute", err.Error())
				return nil, err
			}
		}

		result.TemplateID = templateID
		result.TplItems = tplItems
		result.Hkkjitenzan = hkkjitenzan
		result.Sonnekigaku = 0

		return result, nil

	} else {
		// 取得時点に遡って計算

		// 生成临时数据ID
		uid := uuid.Must(uuid.NewRandom())
		templateID := uid.String()

		result = &typesx.ComputeResult{}

		var tplItems typesx.TplData

		// 支付数据合法性检查
		checkErr := payDataValidCheck(p.CancellationRightOption, p.Payments)
		if checkErr != nil {
			loggerx.ErrorLog("compute", checkErr.Error())
			return nil, checkErr
		}
		// 残价保证额
		residualValue := p.ResidualValue
		// 割引率
		rishiritsu := p.Rishiritsu
		// 比較開始期首月
		firstMonth := p.FirstMonth
		var firstMonthB, err = time.Parse("2006-1-02", firstMonth+"-01")
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		// 租赁开始日--利息数据编辑用
		leasestymd := p.Leasestymd
		leasestym := leasestymd.Format("200601")
		leasestymd, err = time.Parse("20060102", leasestym+"01")
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		// 租赁开始日--偿还数据编辑用
		leasestsyoymd := p.Leasestymd
		leasestsyoym := leasestsyoymd.Format("200601")
		leasestsyoymd, err = time.Parse("20060102", leasestsyoym+"01")
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		// 租赁期间
		leasekikan := p.Leasekikan
		// 延长租赁期间
		extentionOption := p.ExtentionOption
		// 租赁总期间算出(租赁期间 + 延长租赁期间)
		leasekikanTotal := leasekikan + extentionOption
		// 耐用年限(月单位)
		// yms := p.Assetlife * 12
		// 減価償却期間算出
		var genkakikan = leasekikanTotal
		/* genkakikan := yms
		if p.Torihikikbn != "1" {
			// 移転外
			if genkakikan > leasekikanTotal {
				// 取租赁期间与耐用期间中较短者
				genkakikan = leasekikanTotal
			}
		} */

		// 比較開始時点から計算
		hkkjitenzan, presentTotalRemain := getLeaseDebt(p.Payments, rishiritsu, firstMonth, residualValue)
		// 根据参数传入的支付情报算出现在价值合计
		presentTotal, leaseTotal := getLeaseTotal(p.Payments, leasestymd, rishiritsu)
		// 元本残高相当额(初回 = 租赁负债额) => 租赁负债额算出(现在价值累计)
		var principalAmount float64 = presentTotal

		var leaseData []typesx.Lease
		var repayData []typesx.RePayment

		// 根据支付年月支付周期等情报对支付情报再整理
		payments, err := getArrangedPays(p.Payments)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}

		// **********利息情报算出**********
		leases, err := getLeaseDataStart(payments, firstMonthB, firstMonthB, presentTotalRemain, rishiritsu)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		leaseData = append(leaseData, leases...)

		// **********偿还情报算出**********
		// 前払リース料
		//paymentsAtOrPrior := p.PaymentsAtOrPrior
		// リース・インセンティブ（前払）
		//incentivesAtOrPrior := p.IncentivesAtOrPrior
		// 当初直接費用
		initialDirectCosts := p.InitialDirectCosts
		// 原状回復コスト
		restorationCosts := p.RestorationCosts
		// 初期期首簿価 = 租赁负债额(现在价值合计 + 解约赔偿金) + 前払リース料 - リース・インセンティブ（前払）+ 当初直接費用 + 原状回復コスト
		// var boka float64 = principalAmount + paymentsAtOrPrior - incentivesAtOrPrior + initialDirectCosts + restorationCosts
		// 初期期首簿価 = 现在价值合计+ 当初直接費用 + 原状回復コスト
		var boka float64 = principalAmount + initialDirectCosts + restorationCosts
		// 使用権資産簿価 = 取得价值 - (取得价值-残价保证额)/租赁期间 * (租赁开始日到比較開始期首月的月数)
		var useBoka = boka - math.Floor(((boka-residualValue)/float64(leasekikan)))*float64(getGapMonths(leasestymd, firstMonthB))
		// 利益剰余金
		sonnekigaku := presentTotalRemain - useBoka

		// 原始取得价值 = 初期期首簿価
		result.KiSyuBoka = boka

		// 偿还情报算出处理
		repays, err := getRepayDataObtain(leasestymd, genkakikan, residualValue, boka, appID, db, firstMonthB)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		repayData = append(repayData, repays...)

		// 处理月度取得
		cfg, err := configx.GetConfigVal(db, appID)
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return nil, err
		}
		syoriYm := cfg.GetSyoriYm()

		// 支付情报
		var payData []*template.ListItems
		for _, pay := range p.Payments {
			payyear, _ := strconv.Atoi(pay.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(pay.Paymentymd[5:7])
			items := make(map[string]*template.Value)
			items["paymentcount"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(pay.Paymentcount),
			}
			items["paymentType"] = &template.Value{
				DataType: "text",
				Value:    pay.PaymentType,
			}
			items["paymentymd"] = &template.Value{
				DataType: "date",
				Value:    pay.Paymentymd,
			}
			items["paymentleasefee"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Paymentleasefee, 'f', -1, 64),
			}
			items["paymentleasefeehendo"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Paymentleasefeehendo, 'f', -1, 64),
			}
			items["incentives"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Incentives, 'f', -1, 64),
			}
			items["sonotafee"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Sonotafee, 'f', -1, 64),
			}
			items["kaiyakuson"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(pay.Kaiyakuson, 'f', -1, 64),
			}
			items["year"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}

			data := template.ListItems{
				Items:        items,
				DatastoreId:  p.DsMap["paymentStatus"],
				DatastoreKey: "paymentStatus",
				TemplateId:   templateID,
			}

			payData = append(payData, &data)
		}

		tplItems = append(tplItems, payData...)

		// 利息情报
		var lsData []*template.ListItems
		for _, lease := range leaseData {
			payyear, _ := strconv.Atoi(lease.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(lease.Paymentymd[5:7])
			items := make(map[string]*template.Value)
			items["interest"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Interest, 'f', -1, 64),
			}
			items["repayment"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Repayment, 'f', -1, 64),
			}
			items["balance"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Balance, 'f', -1, 64),
			}
			items["firstbalance"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Firstbalance, 'f', -1, 64),
			}
			items["present"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Present, 'f', -1, 64),
			}
			items["paymentymd"] = &template.Value{
				DataType: "date",
				Value:    lease.Paymentymd,
			}
			items["year"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}

			data := template.ListItems{
				Items:        items,
				DatastoreId:  p.DsMap["paymentInterest"],
				DatastoreKey: "paymentInterest",
				TemplateId:   templateID,
			}

			lsData = append(lsData, &data)
		}

		tplItems = append(tplItems, lsData...)
		// 処理月度の先月までの償却費の累計額
		var preDepreciationTotal float64 = 0
		// 偿还情报
		var rpData []*template.ListItems
		for _, rp := range repayData {
			rpyear, _ := strconv.Atoi(rp.Syokyakuymd[0:4])
			rpmonth, _ := strconv.Atoi(rp.Syokyakuymd[5:7])
			items := make(map[string]*template.Value)
			items["endboka"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Endboka, 'f', -1, 64),
			}
			items["boka"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Boka, 'f', -1, 64),
			}
			items["syokyaku"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Syokyaku, 'f', -1, 64),
			}
			items["syokyakuymd"] = &template.Value{
				DataType: "date",
				Value:    rp.Syokyakuymd,
			}
			items["syokyakukbn"] = &template.Value{
				DataType: "text",
				Value:    rp.Syokyakukbn,
			}
			items["year"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpyear),
			}
			items["month"] = &template.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpmonth),
			}

			data := template.ListItems{
				Items:        items,
				DatastoreId:  p.DsMap["repayment"],
				DatastoreKey: "repayment",
				TemplateId:   templateID,
			}

			if rp.Syokyakuymd[:7] < syoriYm {
				preDepreciationTotal += rp.Syokyaku
			}

			rpData = append(rpData, &data)
		}

		tplItems = append(tplItems, rpData...)

		// 履历情报
		var hsData []*template.ListItems

		items := make(map[string]*template.Value)
		items["leaseTotal"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(leaseTotal, 'f', -1, 64),
		}
		items["presentTotal"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(presentTotal, 'f', -1, 64),
		}
		items["preDepreciationTotal"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(preDepreciationTotal, 'f', -1, 64),
		}
		items["hkkjitenzan"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(hkkjitenzan, 'f', -1, 64),
		}
		items["sonnekigaku"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(sonnekigaku, 'f', -1, 64),
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["rireki"],
			DatastoreKey: "rireki",
			TemplateId:   templateID,
		}

		hsData = append(hsData, &data)

		tplItems = append(tplItems, hsData...)

		if insert {
			// 支付数据存入临时集合
			tplService := template.NewTemplateService("database", client.DefaultClient)

			var hsReq template.MutilAddRequest

			// 从body中获取参数
			hsReq.Data = tplItems
			// 从共通中获取参数
			hsReq.Writer = userID
			hsReq.Database = db
			hsReq.Collection = userID
			hsReq.AppId = appID

			_, err = tplService.MutilAddTemplateItem(context.TODO(), &hsReq)
			if err != nil {
				loggerx.ErrorLog("compute", err.Error())
				return nil, err
			}
		}

		result.TemplateID = templateID
		result.TplItems = tplItems
		result.Hkkjitenzan = hkkjitenzan
		result.Sonnekigaku = sonnekigaku

		return result, nil

	}

}

// ChangeCompute 计算情报变更数据(租赁系统用)
func ChangeCompute(db, appID, henkouymd, userID string, payData []typesx.Payment, leaseData []typesx.Lease, repayData []typesx.RePayment, dsMap map[string]string, insert bool) (result *typesx.ChangeResult, err error) {

	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &typesx.ChangeResult{}

	var tplItems typesx.TplData

	// 变更年月
	henkouym, err := time.Parse("2006-01", henkouymd[0:7])
	if err != nil {
		loggerx.ErrorLog("changeCompute", err.Error())
		return nil, err
	}

	// 翌月から最終回までの支払リース料合計額
	var payTotalRemain float64 = 0
	// 翌月から最終回までの利息合計額
	var interestTotalRemain float64 = 0
	for _, lease := range leaseData {
		// 当前利息支付年月
		paymentym, err := time.Parse("2006-01", lease.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("changeCompute", err.Error())
			return nil, err
		}
		// 翌月から最終回まで
		if paymentym.After(henkouym) {
			payTotalRemain += lease.Interest + lease.Repayment
			interestTotalRemain += lease.Interest
		}
	}

	// リース開始日から変更年月日までの償却費の累計額
	var oldDepreciationTotal float64 = 0

	for _, repay := range repayData {
		if repay.Syokyakuymd[:7] <= henkouymd[:7] {
			oldDepreciationTotal += repay.Syokyaku
		}
	}

	// 履历情报
	var hsData []*template.ListItems

	items := make(map[string]*template.Value)

	items["oldDepreciationTotal"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(oldDepreciationTotal, 'f', -1, 64),
	}
	items["payTotalRemain"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(payTotalRemain, 'f', -1, 64),
	}
	items["interestTotalRemain"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(interestTotalRemain, 'f', -1, 64),
	}

	data := template.ListItems{
		Items:        items,
		DatastoreId:  dsMap["rireki"],
		DatastoreKey: "rireki",
		TemplateId:   templateID,
	}

	hsData = append(hsData, &data)

	tplItems = append(tplItems, hsData...)

	if insert {
		// 支付数据存入临时集合
		tplService := template.NewTemplateService("database", client.DefaultClient)

		var hsReq template.MutilAddRequest

		// 从body中获取参数
		hsReq.Data = tplItems
		// 从共通中获取参数
		hsReq.Writer = userID
		hsReq.Database = db
		hsReq.Collection = userID
		hsReq.AppId = appID

		_, err = tplService.MutilAddTemplateItem(context.TODO(), &hsReq)
		if err != nil {
			loggerx.ErrorLog("changeCompute", err.Error())
			return nil, err
		}
	}

	result.TemplateID = templateID
	result.TplItems = tplItems

	return result, nil
}

// DebtCompute 计算债务变更(租赁系统用)
func DebtCompute(db, appID, userID string, kisyuBoka float64, opayData []typesx.Payment, oleaseData []typesx.Lease, orepayData []typesx.RePayment, p typesx.DebtParam, insert bool) (result *typesx.DebtResult, err error) {

	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &typesx.DebtResult{}

	var tplItems typesx.TplData

	// 支付数据合法性检查
	checkErr := payDataValidCheck(p.CancellationRightOption, p.Payments)
	if checkErr != nil {
		loggerx.ErrorLog("debtCompute", checkErr.Error())
		return nil, checkErr
	}
	// 最终支付日取得(不含残价保证金和购入选项行使金)
	lastPayYmd := p.Payments[len(p.Payments)-1].Paymentymd
	if p.Payments[len(p.Payments)-1].PaymentType != "支払" {
		lastPayYmd = p.Payments[len(p.Payments)-2].Paymentymd
	}
	// 租赁满了年月日检查
	checkErr = expireymdCheck(p.Leasestymd[:10], lastPayYmd, p.Leasekikan, p.ExtentionOption)
	if checkErr != nil {
		loggerx.ErrorLog("debtCompute", checkErr.Error())
		return nil, checkErr
	}
	// 处理月度取得
	cfg, err := configx.GetConfigVal(db, appID)
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	syoriYmStr := cfg.GetSyoriYm()
	// 预定解约的场合,支付数据检查
	if p.Kaiyakuymd != "" {
		checkErrK := kaiyakuPayDataCheck(syoriYmStr, p.Kaiyakuymd, opayData, p.Payments)
		if checkErrK != nil {
			loggerx.ErrorLog("debtCompute", checkErrK.Error())
			return nil, checkErrK
		}
	}

	// 処理月度の翌月からリース料変更可能 and 変更年月の翌月から変動リース料編集可能
	isHasErr := payChangeableCheck(p.Henkouymd[0:7], syoriYmStr, opayData, p.Payments)
	if isHasErr != nil {
		loggerx.ErrorLog("debtCompute", isHasErr.Error())
		return nil, isHasErr
	}
	// 处理月度转换
	syoriym, err := time.Parse("2006-01", syoriYmStr)
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	// 割引率
	rishiritsu := p.Rishiritsu
	// 变更年月日
	henkouymd := p.Henkouymd
	// 新的基准日（租赁开始日）
	henkouym, err := time.Parse("2006-01", p.Henkouymd[0:7])
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	leasestymd := henkouym.AddDate(0, 1, 0)
	// 1--变更时点剩余元本残高
	var leftBalanceAfter float64 = 0
	// 2--变更时点剩余使用权资产
	var leftBokaAfter float64 = 0
	// 债务变更前数据处理
	var leaseData []typesx.Lease
	var repayData []typesx.RePayment
	// 债务变更调整区调整前数据
	var repayDataAdjBefore []typesx.RePayment
	// 循环旧利息数据集合,保存债务变更前利息数据,变更时点剩余元本残高取得
	for _, lease := range oleaseData {
		// 支付年月
		paymentym, err := time.Parse("2006-01", lease.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 变更年月
		henkouym, err := time.Parse("2006-01", henkouymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 变更时点剩余元本残高取得
		if paymentym.Before(henkouym) || paymentym.Equal(henkouym) {
			leftBalanceAfter = lease.Balance
			leaseData = append(leaseData, lease)
		}
	}

	// 循环旧偿还数据集合,保存债务变更前偿还数据,变更时点剩余使用权资产取得,债务变更调整区调整前偿还数据取得
	for _, repay := range orepayData {
		// 偿还年月
		syokyakuym, err := time.Parse("2006-01", repay.Syokyakuymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 变更年月
		henkouym, err := time.Parse("2006-01", henkouymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 变更时点剩余使用权资产取得
		if syokyakuym.Before(henkouym) || syokyakuym.Equal(henkouym) {
			leftBokaAfter = repay.Endboka
			repayData = append(repayData, repay)
		}
		// 债务变更调整区调整前数据取得
		if syokyakuym.After(henkouym) && (syokyakuym.Before(syoriym) || syokyakuym.Equal(syoriym)) {
			repayDataAdjBefore = append(repayDataAdjBefore, repay)
		}
	}

	// 变更后支付数据取得
	var henkouPayments []typesx.Payment
	// 现支付额合计(剩余支付期间现支払額的合计额)
	var payTotalAfter float64 = 0
	// 循环新支付情报,保存变更后支付数据
	for _, pay := range p.Payments {
		// 支付年月
		paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 变更年月
		henkouym, err := time.Parse("2006-01", henkouymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 债务变更年月后(不包含变更当月,当月按变更前处理,次月生效)
		if paymentym.After(henkouym) {
			henkouPayments = append(henkouPayments, pay)
			payTotalAfter += pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
		}
	}
	// 3--減少した支払総額(剩余支付期间元支払額×减少比例(1-剩余资产百分比)的合计额)
	var gensyoPayTotal float64 = 0
	// 原支付额比例减少后支付合计(剩余支付期间元支払額×剩余资产百分比的合计额)
	var payTotalRemain float64 = 0
	// 原支付额合计
	var payTotal float64 = 0

	// 循环旧支付情报,減少した支払総額取得
	for _, pay := range opayData {
		// 支付年月
		paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 变更年月
		henkouym, err := time.Parse("2006-01", henkouymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 债务变更年月后(不包含变更当月,当月按变更前处理,次月生效)
		if paymentym.After(henkouym) {
			gensyoPayTotal += math.Floor((pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo) * (1 - p.Percentage))
			payTotalRemain += math.Floor((pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo) * p.Percentage)
			payTotal += pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
		}
	}

	// 支付额变动额 = 再見積後リース料総額-再見積前リース料総額
	payTotalChange := payTotalAfter - payTotal

	// 根据变更后支付数据,算出现在价值合计
	presentTotal, _ := getLeaseTotal(henkouPayments, leasestymd, rishiritsu)
	// 9--变更后剩余リース負債 = (现在价值合计)
	var leaseTotal float64 = presentTotal

	// 分录使用
	// 変更時点の元本残高に対して、比例減少した金額
	var gensyoBalance float64 = 0
	// 変更時点の帳簿価額に対して、比例減少した金額
	var gensyoBoka float64 = 0
	// 再見積変更後現在価値
	var leaseTotalAfter float64 = 0
	// 変更時点の元本残高に対して、比例残の金額
	var leaseTotalRemain float64 = 0

	// ===========================================================================
	// =======================损益额等相关统计情报算出==============================
	// 比例减少
	if p.Percentage < 1 {
		// 4--リース範囲縮小の割合で減少分のリース債務 =变更前租赁负债额*減少比例(1-剩余资产百分比)
		gensyoBalance = math.Floor(leftBalanceAfter * (1 - p.Percentage))
		fmt.Printf("減少分のリース債務: %v \r\n", gensyoBalance)
		// 5--減少分に含まれる利息（未確認融資費用）= 減少した支払総額 - 減少分のリース債務
		gensyoLease := gensyoPayTotal - gensyoBalance
		fmt.Printf("減少分に含まれる利息（未確認融資費用）: %v \r\n", gensyoLease)
		// 6--リース範囲縮小の割合で減少分の使用権資産 = 变更时点剩余使用权资产*減少比例(1-剩余资产百分比)
		gensyoBoka = math.Floor(leftBokaAfter * (1 - p.Percentage))
		fmt.Printf("減少分の使用権資産: %v \r\n", gensyoBoka)
		// 7--比例減少によって発生する損益 = 減少分のリース債務 - 減少分の使用権資産
		result.Sonnekigaku = gensyoBalance - gensyoBoka
		// 8--リース範囲縮小の割合で減少し、残ったリース債務 = 变更前租赁负债额 = 变更时点剩余元本残高 * 剩余资产百分比
		result.OLeasesaimusougaku = math.Floor(leftBalanceAfter * p.Percentage)
		// 9--变更后租赁负债额 = 变更后剩余リース負債
		result.Leasesaimusougaku = leaseTotal
		// 10--リース債務の変動額 = 租赁负债差额 = 变更后租赁负债额 - 变更前租赁负债额
		result.Leasesaimusagaku = result.Leasesaimusougaku - result.OLeasesaimusougaku
		fmt.Printf("リース債務の変動額: %v \r\n", result.Leasesaimusagaku)
		// 11--リース範囲縮小の割合で減少し、残った使用権資産簿価 = 变更前使用権資産額 = 变更时点剩余使用权资产 * 剩余资产百分比
		result.OShisannsougaku = math.Floor(leftBokaAfter * p.Percentage)
		// 12--使用権資産簿価を調整 = 变更后使用権資産額 = 变更前使用権資産額 + 租赁负债差额（作为使用权調整額）
		result.Shisannsougaku = result.OShisannsougaku + result.Leasesaimusagaku
		// 13--増加した支払総額(剩余支付期间现支付额-元支払額×（1-减少比例）的合计额)
		payTotalAfteradd := payTotalAfter - payTotalRemain
		fmt.Printf("増加した支払総額: %v \r\n", payTotalAfteradd)
		// 増加した未確認融資費用額
		zoukaLease := payTotalAfteradd - result.Leasesaimusagaku
		fmt.Printf("増加した未確認融資費用額: %v \r\n", zoukaLease)
		// 使用权资产差额
		result.Shisannsagaku = result.Shisannsougaku - result.OShisannsougaku

		// 再見積変更後現在価値
		leaseTotalAfter = leaseTotal
		// 変更時点の元本残高に対して、比例残の金額
		leaseTotalRemain = math.Floor(leftBalanceAfter * p.Percentage)
	}

	// 非比例减少
	if p.Percentage == 1 {
		// 变更前租赁负债额 = 变更时点剩余元本残高
		result.OLeasesaimusougaku = leftBalanceAfter
		// 变更后租赁负债额 = 变更后剩余リース負債
		result.Leasesaimusougaku = leaseTotal
		// 租赁负债差额
		result.Leasesaimusagaku = result.Leasesaimusougaku - result.OLeasesaimusougaku

		// 变更前使用権資産額 = 变更时点剩余使用权资产
		result.OShisannsougaku = leftBokaAfter
		// 变更后使用権資産額 = 变更前使用権資産額 + 租赁负债差额（作为使用权調整額）
		result.Shisannsougaku = result.OShisannsougaku + result.Leasesaimusagaku
		// 使用权资产差额
		result.Shisannsagaku = result.Shisannsougaku - result.OShisannsougaku

		// 损益额 = 租赁负债差额 - 使用权资产差额
		result.Sonnekigaku = result.Leasesaimusagaku - result.Shisannsagaku
	}

	// 原始取得价值 = 原始值 + 使用权资产差额
	result.KiSyuBoka = kisyuBoka + result.Shisannsagaku

	// ==================================================================================
	// =======================变更年月后的利息和偿还情报再计算==============================
	// 变更后支付数据再整理
	payments, err := getArrangedPays(henkouPayments)
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	// 元本残高相当额
	var principalAmount float64 = result.Leasesaimusougaku
	// 拿到变更前的最后一条利息数据的支付年月
	prevPaymentymd := henkouym.AddDate(0, 1, 0)

	// **********利息情报算出**********
	leases, err := getLeaseData(payments, leasestymd, prevPaymentymd, principalAmount, rishiritsu)
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	leaseData = append(leaseData, leases...)

	// **********偿还情报算出**********
	// 租赁总期间
	leasekikanTotal := p.Leasekikan + p.ExtentionOption
	// 減価償却期間算出
	genkakikan := p.Assetlife * 12
	if p.Torihikikbn != "1" {
		// 移転外
		if genkakikan > leasekikanTotal {
			genkakikan = leasekikanTotal
		}
	}
	genkakikan = genkakikan - len(repayData)
	// 新減価开始日算出
	leasestsyoymd, err := time.Parse("2006-01-02", repayData[len(repayData)-1].Syokyakuymd)
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	leasestsyoymd = leasestsyoymd.AddDate(0, 1, 0)
	// 剩余期首簿価 = 变更后使用権資産額
	boka := result.Shisannsougaku

	// 偿还情报算出处理
	repays, err := getRepayData(leasestsyoymd, genkakikan, p.ResidualValue, boka, appID, db)
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	// 调整区调整前偿还总额
	var adjSyoTotalBefore float64 = 0
	// 添加调整区调整前数据,记录调整月和调整区调整前偿还总额
	for _, repay := range repayDataAdjBefore {
		adjSyoTotalBefore += repay.Syokyaku
		repayData = append(repayData, repay)
	}
	// 调整区调整后偿还总额
	var adjSyoTotalAfter float64 = 0
	// 债务变更调整区后数据
	var repayDataAdjAfter []typesx.RePayment
	// 债务变更调整区后数据取得和调整区调整后偿还总额取得
	for _, repay := range repays {
		// 偿还年月
		syokyakuym, err := time.Parse("2006-01", repay.Syokyakuymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		// 变更年月
		henkouym, err := time.Parse("2006-01", henkouymd[0:7])
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
		if syokyakuym.After(henkouym) && (syokyakuym.Before(syoriym) || syokyakuym.Equal(syoriym)) {
			// 调整区调整后偿还总额取得
			adjSyoTotalAfter += repay.Syokyaku
		} else {
			// 债务变更调整区后数据取得
			repayDataAdjAfter = append(repayDataAdjAfter, repay)
		}
	}
	syokyaku := adjSyoTotalAfter - adjSyoTotalBefore

	if syokyaku != 0 {
		// 插入調整額数据
		var repay typesx.RePayment
		// 调整年月
		repay.Syokyakuymd = syoriYmStr + "-01"
		// 調整額
		repay.Syokyaku = syokyaku
		// 调整区分
		repay.Syokyakukbn = "調整"
		// 添加偿却调整数据
		repayData = append(repayData, repay)
	}

	// 添加调整区后数据
	repayData = append(repayData, repayDataAdjAfter...)

	// 支付情报
	var payData []*template.ListItems
	for _, pay := range p.Payments {
		items := make(map[string]*template.Value)
		items["paymentcount"] = &template.Value{
			DataType: "number",
			Value:    strconv.Itoa(pay.Paymentcount),
		}
		items["paymentType"] = &template.Value{
			DataType: "text",
			Value:    pay.PaymentType,
		}
		items["paymentymd"] = &template.Value{
			DataType: "date",
			Value:    pay.Paymentymd,
		}
		items["paymentleasefee"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Paymentleasefee, 'f', -1, 64),
		}
		items["paymentleasefeehendo"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Paymentleasefeehendo, 'f', -1, 64),
		}
		items["incentives"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Incentives, 'f', -1, 64),
		}
		items["sonotafee"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Sonotafee, 'f', -1, 64),
		}
		items["kaiyakuson"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Kaiyakuson, 'f', -1, 64),
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["paymentStatus"],
			DatastoreKey: "paymentStatus",
			TemplateId:   templateID,
		}

		payData = append(payData, &data)
	}

	tplItems = append(tplItems, payData...)

	// 利息情报
	var lsData []*template.ListItems
	for _, lease := range leaseData {

		items := make(map[string]*template.Value)
		items["interest"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Interest, 'f', -1, 64),
		}
		items["repayment"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Repayment, 'f', -1, 64),
		}
		items["balance"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Balance, 'f', -1, 64),
		}
		items["present"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Present, 'f', -1, 64),
		}
		items["paymentymd"] = &template.Value{
			DataType: "date",
			Value:    lease.Paymentymd,
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["paymentInterest"],
			DatastoreKey: "paymentInterest",
			TemplateId:   templateID,
		}

		lsData = append(lsData, &data)
	}

	tplItems = append(tplItems, lsData...)

	// 偿还情报
	var rpData []*template.ListItems
	for _, rp := range repayData {

		items := make(map[string]*template.Value)
		items["endboka"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(rp.Endboka, 'f', -1, 64),
		}
		items["boka"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(rp.Boka, 'f', -1, 64),
		}
		items["syokyaku"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(rp.Syokyaku, 'f', -1, 64),
		}
		items["syokyakuymd"] = &template.Value{
			DataType: "date",
			Value:    rp.Syokyakuymd,
		}
		items["syokyakukbn"] = &template.Value{
			DataType: "text",
			Value:    rp.Syokyakukbn,
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["repayment"],
			DatastoreKey: "repayment",
			TemplateId:   templateID,
		}

		rpData = append(rpData, &data)
	}

	tplItems = append(tplItems, rpData...)

	// 履历情报
	var hsData []*template.ListItems

	items := make(map[string]*template.Value)
	items["shisannsougaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(result.Shisannsougaku, 'f', -1, 64),
	}
	items["o_shisannsougaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(result.OShisannsougaku, 'f', -1, 64),
	}
	items["leasesaimusougaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(result.Leasesaimusougaku, 'f', -1, 64),
	}
	items["o_leasesaimusougaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(result.OLeasesaimusougaku, 'f', -1, 64),
	}
	items["shisannsagaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(result.Shisannsagaku, 'f', -1, 64),
	}
	items["leasesaimusagaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(result.Leasesaimusagaku, 'f', -1, 64),
	}
	items["sonnekigaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(result.Sonnekigaku, 'f', -1, 64),
	}
	// 分录使用
	// 変更時点の支払残額に対して、比例減少した金額
	items["gensyoPayTotal"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(gensyoPayTotal, 'f', -1, 64),
	}
	// 変更時点の元本残高に対して、比例減少した金額
	items["gensyoBalance"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(gensyoBalance, 'f', -1, 64),
	}
	// 変更時点の帳簿価額に対して、比例減少した金額
	items["gensyoBoka"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(gensyoBoka, 'f', -1, 64),
	}
	// 再見積変更後現在価値
	items["leaseTotalAfter"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(leaseTotalAfter, 'f', -1, 64),
	}
	// 変更時点の元本残高に対して、比例残の金額
	items["leaseTotalRemain"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(leaseTotalRemain, 'f', -1, 64),
	}
	// 再見積変更後の支払総額
	items["payTotalAfter"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(payTotalAfter, 'f', -1, 64),
	}
	// 変更時点の支払残額に対して、比例残の金額
	items["payTotalRemain"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(payTotalRemain, 'f', -1, 64),
	}
	// 支付变动额
	items["payTotalChange"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(payTotalChange, 'f', -1, 64),
	}

	data := template.ListItems{
		Items:        items,
		DatastoreId:  p.DsMap["rireki"],
		DatastoreKey: "rireki",
		TemplateId:   templateID,
	}

	hsData = append(hsData, &data)

	tplItems = append(tplItems, hsData...)

	if insert {
		// 支付数据存入临时集合
		tplService := template.NewTemplateService("database", client.DefaultClient)

		var hsReq template.MutilAddRequest

		// 从body中获取参数
		hsReq.Data = tplItems
		// 从共通中获取参数
		hsReq.Writer = userID
		hsReq.Database = db
		hsReq.Collection = userID
		hsReq.AppId = appID

		_, err = tplService.MutilAddTemplateItem(context.TODO(), &hsReq)
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
	}

	result.TemplateID = templateID
	result.TplItems = tplItems

	return result, nil
}

// CancelCompute 中途解约处理(租赁系统用)
func CancelCompute(db, appID, userID string, opayData []typesx.Payment, oleaseData []typesx.Lease, orepayData []typesx.RePayment, p typesx.CancelParam, insert bool) (result *typesx.CancelResult, err error) {

	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &typesx.CancelResult{}

	var tplItems typesx.TplData

	// 解约后数据
	var payData []typesx.Payment
	var leaseData []typesx.Lease
	var repayData []typesx.RePayment
	// 解约年月日
	kaiyakuymd := p.Kaiyakuymd
	// 解约年月转换
	kaiyakuym, err := time.Parse("2006-01", kaiyakuymd[0:7])
	if err != nil {
		loggerx.ErrorLog("cancelCompute", err.Error())
		return nil, err
	}
	// 处理月度取得
	cfg, err := configx.GetConfigVal(db, appID)
	if err != nil {
		loggerx.ErrorLog("cancelCompute", err.Error())
		return nil, err
	}
	syoriYmStr := cfg.GetSyoriYm()
	// 处理月度转换
	syoriym, err := time.Parse("2006-01", syoriYmStr)
	if err != nil {
		loggerx.ErrorLog("cancelCompute", err.Error())
		return nil, err
	}

	// 解约后支付数据整理
	for _, pay := range opayData {
		// 当前支付年月
		paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("cancelCompute", err.Error())
			return nil, err
		}
		// 解约年月和解约年月前的支付数据放入支付集合
		if paymentym.Before(kaiyakuym) || paymentym.Equal(kaiyakuym) {
			payData = append(payData, pay)
		} else {
			break
		}
	}
	// 中途解約時点の支払リース料残額
	var payTotalRemain float64 = 0
	// 中途解約時点の利息残
	var interestTotalRemain float64 = 0
	for _, lease := range oleaseData {
		// 当前利息支付年月
		paymentym, err := time.Parse("2006-01", lease.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("cancelCompute", err.Error())
			return nil, err
		}
		// 解约年月和解约年月前的利息数据放入利息集合
		if paymentym.After(kaiyakuym) {
			payTotalRemain += lease.Interest + lease.Repayment
			interestTotalRemain += lease.Interest
		}
	}

	// 解约后利息数据整理&解約時元本残高算出
	// 解約時元本残高
	var remainDebt float64 = 0
	for _, lease := range oleaseData {
		// 当前利息支付年月
		paymentym, err := time.Parse("2006-01", lease.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("cancelCompute", err.Error())
			return nil, err
		}
		// 解约年月和解约年月前的利息数据放入利息集合
		if paymentym.Before(kaiyakuym) || paymentym.Equal(kaiyakuym) {
			remainDebt = lease.Balance
			leaseData = append(leaseData, lease)
		} else {
			break
		}
	}

	// 解约后偿还数据整理&中途解約による除却損算出
	// 中途解約による除却損
	var lossgaku float64 = 0
	// 中途解約による調整額
	var tyoseigaku float64 = 0
	// 中途解約時点の償却費の累計額
	var syokyakuTotal float64 = 0
	for index, repay := range orepayData {
		// 当前偿还年月
		syokyakuym, err := time.Parse("2006-01", repay.Syokyakuymd[0:7])
		if err != nil {
			loggerx.ErrorLog("cancelCompute", err.Error())
			return nil, err
		}
		if index == 0 {
			// 除却損的计算-首次
			if syokyakuym.Equal(kaiyakuym) {
				// 解約年月日前月の月末簿価を除却損とする
				lossgaku = repay.Boka
			}
		}
		// 处理月度前的偿还数据放入偿还集合
		if syokyakuym.Before(kaiyakuym) || syokyakuym.Equal(kaiyakuym) {
			syokyakuTotal += repay.Syokyaku
		}

		// 偿却费的计算
		if syokyakuym.Before(syoriym) || syokyakuym.Equal(syoriym) {
			repayData = append(repayData, repay)
		}
		// 偿却费的计算
		if syokyakuym.After(kaiyakuym) && (syokyakuym.Before(syoriym) || syokyakuym.Equal(syoriym)) {
			// 解約年月日から処理月度の前月までの償却費は処理月度の調整償却費とする
			tyoseigaku += repay.Syokyaku
		}

		// 除却損的计算
		if syokyakuym.Before(kaiyakuym) {
			// 解約年月日前月の月末簿価を除却損とする
			lossgaku = repay.Endboka
		}
	}

	// 中途解約による調整額の添付
	if tyoseigaku != 0 {
		var re typesx.RePayment
		re.Syokyakuymd = syoriYmStr + "-01"
		re.Syokyaku = 0 - tyoseigaku
		re.Syokyakukbn = "調整"
		repayData = append(repayData, re)
	}

	// 返回结果编辑
	result.RemainDebt = remainDebt
	result.Lossgaku = lossgaku

	// 支付情报
	var payItems []*template.ListItems
	for _, pay := range payData {
		items := make(map[string]*template.Value)
		items["paymentcount"] = &template.Value{
			DataType: "number",
			Value:    strconv.Itoa(pay.Paymentcount),
		}
		items["paymentType"] = &template.Value{
			DataType: "text",
			Value:    pay.PaymentType,
		}
		items["paymentymd"] = &template.Value{
			DataType: "date",
			Value:    pay.Paymentymd,
		}
		items["paymentleasefee"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Paymentleasefee, 'f', -1, 64),
		}
		items["paymentleasefeehendo"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Paymentleasefeehendo, 'f', -1, 64),
		}
		items["incentives"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Incentives, 'f', -1, 64),
		}
		items["sonotafee"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Sonotafee, 'f', -1, 64),
		}
		items["kaiyakuson"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(pay.Kaiyakuson, 'f', -1, 64),
		}

		payItem := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["paymentStatus"],
			DatastoreKey: "paymentStatus",
			TemplateId:   templateID,
		}

		payItems = append(payItems, &payItem)
	}

	tplItems = append(tplItems, payItems...)

	// 利息情报
	var lsData []*template.ListItems
	for _, lease := range leaseData {
		items := make(map[string]*template.Value)
		items["interest"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Interest, 'f', -1, 64),
		}
		items["repayment"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Repayment, 'f', -1, 64),
		}
		items["balance"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Balance, 'f', -1, 64),
		}
		items["present"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(lease.Present, 'f', -1, 64),
		}
		items["paymentymd"] = &template.Value{
			DataType: "date",
			Value:    lease.Paymentymd,
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["paymentInterest"],
			DatastoreKey: "paymentInterest",
			TemplateId:   templateID,
		}

		lsData = append(lsData, &data)
	}

	tplItems = append(tplItems, lsData...)

	// 偿还情报
	var rpData []*template.ListItems
	for _, rp := range repayData {
		items := make(map[string]*template.Value)
		items["endboka"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(rp.Endboka, 'f', -1, 64),
		}
		items["boka"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(rp.Boka, 'f', -1, 64),
		}
		items["syokyaku"] = &template.Value{
			DataType: "number",
			Value:    strconv.FormatFloat(rp.Syokyaku, 'f', -1, 64),
		}
		items["syokyakuymd"] = &template.Value{
			DataType: "date",
			Value:    rp.Syokyakuymd,
		}
		items["syokyakukbn"] = &template.Value{
			DataType: "text",
			Value:    rp.Syokyakukbn,
		}

		data := template.ListItems{
			Items:        items,
			DatastoreId:  p.DsMap["repayment"],
			DatastoreKey: "repayment",
			TemplateId:   templateID,
		}

		rpData = append(rpData, &data)
	}

	tplItems = append(tplItems, rpData...)

	// 履历情报
	var hsData []*template.ListItems
	items := make(map[string]*template.Value)
	// 解約時元本残高
	items["remaindebt"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(remainDebt, 'f', -1, 64),
	}
	// 中途解約による除却損金额
	items["lossgaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(lossgaku, 'f', -1, 64),
	}
	// 解约年月日
	items["kaiyakuymd"] = &template.Value{
		DataType: "text",
		Value:    kaiyakuymd,
	}
	// 中途解約時点の償却費の累計額
	items["syokyakuTotal"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(syokyakuTotal, 'f', -1, 64),
	}
	// 中途解約時点の支払リース料残額
	items["payTotalRemain"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(payTotalRemain, 'f', -1, 64),
	}
	// 中途解約時点の利息残
	items["interestTotalRemain"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(interestTotalRemain, 'f', -1, 64),
	}

	data := template.ListItems{
		Items:        items,
		DatastoreId:  p.DsMap["rireki"],
		DatastoreKey: "rireki",
		TemplateId:   templateID,
	}

	hsData = append(hsData, &data)
	tplItems = append(tplItems, hsData...)

	if insert {
		// 支付数据存入临时集合
		tplService := template.NewTemplateService("database", client.DefaultClient)

		var hsReq template.MutilAddRequest

		// 从body中获取参数
		hsReq.Data = tplItems
		// 从共通中获取参数
		hsReq.Writer = userID
		hsReq.Database = db
		hsReq.Collection = userID
		hsReq.AppId = appID

		_, err = tplService.MutilAddTemplateItem(context.TODO(), &hsReq)
		if err != nil {
			loggerx.ErrorLog("debtCompute", err.Error())
			return nil, err
		}
	}

	result.TemplateID = templateID
	result.TplItems = tplItems

	return result, nil
}

// ExpireCompute 满了处理(租赁系统用)
func ExpireCompute(db, appID, userID string, orepayData []typesx.RePayment, p typesx.ExpireParam, insert bool) (result *typesx.ExpireResult, err error) {
	// 返回数据
	result = &typesx.ExpireResult{}
	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()
	var tplItems typesx.TplData

	// 変更年月日(满了确定)
	henkouymd := p.Henkouymd
	// リース満了償却区分
	expiresyokyakukbn := p.Expiresyokyakukbn
	// 取引判定区分
	torihikikbn := p.Torihikikbn

	if torihikikbn != "1" {
		// 移転外の場合（期間短い方なので、償却完了しているはず、計算無し）
		return result, nil
	} else {
		// 移転の場合
		if expiresyokyakukbn == "keeppayoff" {
			// 償却持続なので、計算無し
			return result, nil
		}
	}

	// 以下は移転且つ償却停止で場合、残りの償却データを削除、残り価値を計算する
	var repayData []typesx.RePayment
	// 変更年月转换
	henkouym, err := time.Parse("2006-01", henkouymd[0:7])
	if err != nil {
		loggerx.ErrorLog("expireCompute", err.Error())
		return nil, err
	}

	// 满了后偿却数据整理和剩余价值算出
	var leftgaku float64 = 0
	for index, repay := range orepayData {
		// 当前偿还年月
		syokyakuym, err := time.Parse("2006-01", repay.Syokyakuymd[0:7])
		if err != nil {
			loggerx.ErrorLog("expireCompute", err.Error())
			return nil, err
		}

		if index == 0 && syokyakuym.Equal(henkouym) {
			// 解約年月日前月の月末簿価を除却損とする
			leftgaku = repay.Boka
		}

		// 変更年月(含)前的偿还数据放入偿还集合和剩余价值算出
		if syokyakuym.Before(henkouym) || syokyakuym.Equal(henkouym) {
			// 変更年月(含)前的偿还数据放入偿还集合
			repayData = append(repayData, repay)
			// 剩余价值记录
			leftgaku = repay.Endboka
		} else {
			break
		}
	}

	// 返回结果编辑
	result.Leftgaku = leftgaku

	// 偿还情报
	if leftgaku != 0 {
		var rpData []*template.ListItems
		for _, rp := range repayData {
			items := make(map[string]*template.Value)
			items["endboka"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Endboka, 'f', -1, 64),
			}
			items["boka"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Boka, 'f', -1, 64),
			}
			items["syokyaku"] = &template.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(rp.Syokyaku, 'f', -1, 64),
			}
			items["syokyakuymd"] = &template.Value{
				DataType: "date",
				Value:    rp.Syokyakuymd,
			}
			items["syokyakukbn"] = &template.Value{
				DataType: "text",
				Value:    rp.Syokyakukbn,
			}

			data := template.ListItems{
				Items:        items,
				DatastoreId:  p.DsMap["repayment"],
				DatastoreKey: "repayment",
				TemplateId:   templateID,
			}

			rpData = append(rpData, &data)
		}

		tplItems = append(tplItems, rpData...)
	}

	// 履历情报
	var hsData []*template.ListItems
	items := make(map[string]*template.Value)

	// 满了后剩余价值
	items["lossgaku"] = &template.Value{
		DataType: "number",
		Value:    strconv.FormatFloat(leftgaku, 'f', -1, 64),
	}

	data := template.ListItems{
		Items:        items,
		DatastoreId:  p.DsMap["rireki"],
		DatastoreKey: "rireki",
		TemplateId:   templateID,
	}

	hsData = append(hsData, &data)
	tplItems = append(tplItems, hsData...)

	if insert {
		// 支付数据存入临时集合
		tplService := template.NewTemplateService("database", client.DefaultClient)

		var hsReq template.MutilAddRequest

		// 从body中获取参数
		hsReq.Data = tplItems
		// 从共通中获取参数
		hsReq.Writer = userID
		hsReq.Database = db
		hsReq.Collection = userID
		hsReq.AppId = appID

		_, err = tplService.MutilAddTemplateItem(context.TODO(), &hsReq)
		if err != nil {
			loggerx.ErrorLog("expireCompute", err.Error())
			return nil, err
		}
	}

	result.TemplateID = templateID
	result.TplItems = tplItems

	return result, nil
}

// 现在价值累计算出
func getLeaseTotal(payments []typesx.Payment, leasestymd time.Time, rishiritsu float64) (presentTotal, leaseTotal float64) {
	var k float64 = 0
	for _, pay := range payments {
		// k幂数取得(租赁开始日到支付年月日的月数)
		payymd, _ := time.Parse("2006-01-02", pay.Paymentymd)
		k = float64(getGapMonths(leasestymd, payymd)) + 1
		// 当回实际支付金额取得(当回实际支付金额 = 当回支付金额 - 当回优惠 + 变动额)
		currentPay := pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
		// リース料累计
		leaseTotal += currentPay
		// 现在价值累计
		presentTotal += math.Floor(currentPay / math.Pow(1+(rishiritsu/12), k))
	}
	// 现在价值累计
	return
}

// 比較開始時点から計算
func getLeaseDebt(payments []typesx.Payment, rishiritsu float64, firstMonth string, residualValue float64) (leaseTotalPayment float64, presentTotalRemain float64) {
	var k float64 = 0
	// 比較開始期首月取得
	payFirstMonth, _ := time.Parse("2006-1", firstMonth)
	for _, pay := range payments {
		// 支付年月日取得
		payymd, _ := time.Parse("2006-01-02", pay.Paymentymd)
		// 满足支付年月在比較開始期首月之后的条件
		if !(payFirstMonth.After(payymd)) {
			// k幂数取得(比較開始期首月到支付年月日的月数)
			k = float64(getGapMonths(payFirstMonth, payymd)) + 1
			// 当回实际支付金额取得(当回实际支付金额 = 当回支付金额 - 当回优惠 + 变动额)
			currentPay := pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
			// 残存リース料累计
			leaseTotalPayment += currentPay
			// 残存价值累计
			presentTotalRemain += math.Floor(currentPay / math.Pow(1+(rishiritsu/12), k))
		}
	}
	return
}

// 支付情报整理
func getArrangedPays(oldPays []typesx.Payment) (pays []typesx.Payment, err error) {
	var payments []typesx.Payment
	// 前回支付日保存用
	var prevPaymentymd time.Time
	// 当回支付金额合计
	var paymentleasefee float64 = 0
	// 当回优惠合计
	var incentives float64 = 0
	// 当回变动额合计
	var paymentleasefeehendo float64 = 0
	// 循环支付情报再整理生成新支付情报
	for i, pay := range oldPays {
		// 支付年月日类型转换
		paymentymd, err := time.Parse("2006-01-02", pay.Paymentymd)
		if err != nil {
			loggerx.ErrorLog("getArrangedPays", err.Error())
			return payments, err
		}
		// 首条支付数据的场合,累计支付金额&优惠&变动额,支付年月日退避
		if i == 0 {
			// 仅有一条支付数据,首条即末条
			if len(oldPays) == 1 {
				// 数据出力
				payments = append(payments, pay)
			} else {
				// 数据累计退避
				paymentleasefee += pay.Paymentleasefee
				incentives += pay.Incentives
				paymentleasefeehendo += pay.Paymentleasefeehendo
				prevPaymentymd = paymentymd
			}
			continue
		}
		// 最后一条支付数据的场合
		if i == len(oldPays)-1 {
			// 前支付年月日与现支付年月日间隔月数取得
			cycle := getGapMonths(prevPaymentymd, paymentymd)
			// 间隔为0,即为同月,继续累计然后出力
			if cycle == 0 {
				// 累计
				paymentleasefee += pay.Paymentleasefee
				incentives += pay.Incentives
				paymentleasefeehendo += pay.Paymentleasefeehendo
				// 出力
				payInfo := typesx.Payment{
					Paymentymd:           pay.Paymentymd,
					Paymentleasefee:      paymentleasefee,
					Incentives:           incentives,
					Paymentleasefeehendo: paymentleasefeehendo,
				}
				payments = append(payments, payInfo)
				continue
			}
			// 间隔大于0,即为异月,先出力前已累计数据然后出力本末回数据
			if cycle > 0 {
				// 出力前已累计数据
				payInfo := typesx.Payment{
					Paymentymd:           prevPaymentymd.Format("2006-01-02"),
					Paymentleasefee:      paymentleasefee,
					Incentives:           incentives,
					Paymentleasefeehendo: paymentleasefeehendo,
				}
				payments = append(payments, payInfo)
				// 出力本末回数据
				payments = append(payments, pay)
				continue
			}
			// 其他情形(当前支付年月小于前回支付年月的场合)进入下面返回错误。
		}

		// 当前支付年月小于前回支付年月的场合,返回错误。
		if paymentymd.Before(prevPaymentymd) {
			return payments, fmt.Errorf("支払いデータ行%dで、現在の支払い年月が前の支払い年月よりも少ない。", i+1)
		}
		// 前支付年月日与现支付年月日间隔月数取得
		cycle := getGapMonths(prevPaymentymd, paymentymd)
		// 间隔为0,即为同月,继续累计
		if cycle == 0 {
			paymentleasefee += pay.Paymentleasefee
			incentives += pay.Incentives
			paymentleasefeehendo += pay.Paymentleasefeehendo
			prevPaymentymd = paymentymd
			continue
		}
		// 间隔大于0,即为异月,先出力前已累计数据然后累计本条数据
		if cycle > 0 {
			// 出力前已累计数据
			payInfo := typesx.Payment{
				Paymentymd:           prevPaymentymd.Format("2006-01-02"),
				Paymentleasefee:      paymentleasefee,
				Incentives:           incentives,
				Paymentleasefeehendo: paymentleasefeehendo,
			}
			payments = append(payments, payInfo)
			// 累计本条数据
			paymentleasefee = pay.Paymentleasefee
			incentives = pay.Incentives
			paymentleasefeehendo = pay.Paymentleasefeehendo
			prevPaymentymd = paymentymd
			continue
		}
	}
	return payments, nil
}

// 利息情报算出
func getLeaseData(payments []typesx.Payment, leasestymd time.Time, prevymd time.Time, principalAmount float64, rishiritsu float64) (ls []typesx.Lease, err error) {
	var k float64 = 0
	var leaseData []typesx.Lease
	// 前回支付日保存用
	var prevPaymentymd time.Time
	// 循环整理生成的新支付情报生成利息相关情报
	for i, pay := range payments {
		var lease typesx.Lease
		// 当前月份的支付年月日
		paymentymd, err := time.Parse("2006-01-02", pay.Paymentymd)
		if err != nil {
			loggerx.ErrorLog("getLeaseData", err.Error())
			return leaseData, err
		}
		// 利息情报-年月日
		lease.Paymentymd = paymentymd.Format("2006-01") + "-01"
		// 未支付单月份产生的利息保存用
		var interestsingle float64 = 0
		// 累加未支付月份产生的利息用保存用
		var interestcount float64 = 0
		// 临时计算原本保存用(计算利息用,不做下回支付退避)
		var principalAmountcal float64 = principalAmount
		// 首条数据的情形
		if i == 0 {
			// 获取首回支付与租赁开始日间隔月数
			firstGap := getGapMonths(prevymd, paymentymd)
			if leasestymd != prevymd {
				// 获取首回支付与前回支付年月间隔月数
				firstGap = firstGap - 1
			}
			// 累加从租赁开始日到首次支付日的利息
			if firstGap > 0 {
				// 累加计算未支付月利息
				for j := 0; j < firstGap; j++ {
					// 单月利息
					interestsingle = math.Floor(principalAmountcal * (rishiritsu / 12))
					// 临时计算原本 = 临时计算原本 + 未支付利息
					principalAmountcal = principalAmountcal + interestsingle
					// 未支付利息累计
					interestcount += interestsingle
				}
			}
			// 利息情报-支払利息相当額 = 当前支付月利息 + 累加的未支付月份产生的利息
			lease.Interest = math.Floor(principalAmountcal*(rishiritsu/12)) + interestcount
			// 当回实际支付额编辑 = 支付额 - 优惠 + 变动额
			currentPay := pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
			// 支付最终回的场合
			if i == len(payments)-1 {
				// 利息情报-支払利息相当額 = 当回实际支付额 - 元本残高相当額
				lease.Interest = math.Floor(currentPay - principalAmount)
			}
			// 利息情报-元本返済相当額
			lease.Repayment = math.Floor(currentPay - lease.Interest)
			// 利息情报-元本残高相当額
			lease.Balance = principalAmount - lease.Repayment
			// K幂数取得(k=租赁开始日到支付年月日的月数)
			payymd, _ := time.Parse("2006-01-02", pay.Paymentymd)
			k = float64(getGapMonths(leasestymd, payymd)) + 1
			// 利息情报-現在価値
			lease.Present = math.Floor(currentPay / math.Pow(1+(rishiritsu/12), k))
			// 前回元本残高相当額变化--下回计算用
			principalAmount = lease.Balance
			// 添加利息情报
			leaseData = append(leaseData, lease)
			// 支付年月日退避
			prevPaymentymd = paymentymd
			continue
		}
		// 获取前回支付与本回支付间隔月数
		cycleGap := getGapMonths(prevPaymentymd, paymentymd)
		// 累加计算间隔未支付月利息
		if cycleGap > 1 {
			// 累加计算未支付月利息
			for j := 0; j < (cycleGap - 1); j++ {
				// 单月利息
				interestsingle = math.Floor(principalAmountcal * (rishiritsu / 12))
				// 临时计算原本 = 临时计算原本 + 未支付利息
				principalAmountcal = principalAmountcal + interestsingle
				// 未支付利息累计
				interestcount += interestsingle
			}
		}
		// 利息情报-支払利息相当額 = 当前支付月利息 + 累加的未支付月份产生的利息
		lease.Interest = math.Floor(principalAmountcal*(rishiritsu/12)) + interestcount
		// 当回实际支付额编辑 = 支付额 - 优惠 + 变动额
		currentPay := pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
		// 支付最终回的场合
		if i == len(payments)-1 {
			// 利息情报-支払利息相当額 = 当回实际支付额 - 元本残高相当額
			lease.Interest = math.Floor(currentPay - principalAmount)
		}
		// 利息情报-元本返済相当額
		lease.Repayment = math.Floor(currentPay - lease.Interest)
		// 利息情报-元本残高相当額
		lease.Balance = principalAmount - lease.Repayment
		// k幂数取得(k=租赁开始日到支付年月日的月数)
		k = float64(getGapMonths(leasestymd, paymentymd)) + 1
		// 利息情报-現在価値
		lease.Present = math.Floor(currentPay / math.Pow(1+(rishiritsu/12), k))
		// 前回元本残高相当額变化--下回计算用
		principalAmount = lease.Balance
		// 添加利息情报
		leaseData = append(leaseData, lease)
		// 支付年月日退避
		prevPaymentymd = paymentymd
	}
	return leaseData, nil
}

// 利息情报算出(開始時点から計算)
func getLeaseDataStart(payments []typesx.Payment, leasestymd time.Time, prevymd time.Time, principalAmount float64, rishiritsu float64) (ls []typesx.Lease, err error) {
	var k float64 = 0
	var tf = true
	var firstbalance float64 = 0
	var leaseData []typesx.Lease
	// 前回支付日保存用
	var prevPaymentymd time.Time
	// 循环整理生成的新支付情报生成利息相关情报
	for i, pay := range payments {
		var lease typesx.Lease
		// 当前月份的支付年月日
		paymentymd, err := time.Parse("2006-01-02", pay.Paymentymd)
		if err != nil {
			loggerx.ErrorLog("getLeaseData", err.Error())
			return leaseData, err
		}
		if !(paymentymd.Before(leasestymd)) {
			// 利息情报-年月日
			lease.Paymentymd = paymentymd.Format("2006-01") + "-01"
			// 未支付单月份产生的利息保存用
			var interestsingle float64 = 0
			// 累加未支付月份产生的利息用保存用
			var interestcount float64 = 0
			// 临时计算原本保存用(计算利息用,不做下回支付退避)
			var principalAmountcal float64 = principalAmount
			// 首条数据的情形
			if tf {
				// 获取首回支付与比較開始期首月间隔月数
				firstGap := getGapMonths(prevymd, paymentymd)
				if leasestymd != prevymd {
					// 获取首回支付与前回支付年月间隔月数
					firstGap = firstGap - 1
				}
				// 累加从比較開始期首月到首次支付日的利息
				if firstGap > 0 {
					// 累加计算未支付月利息
					for j := 0; j < firstGap; j++ {
						// 单月利息
						interestsingle = math.Floor(principalAmountcal * (rishiritsu / 12))
						// 临时计算原本 = 临时计算原本 + 未支付利息
						principalAmountcal = principalAmountcal + interestsingle
						// 未支付利息累计
						interestcount += interestsingle
					}
				}
				// 利息情报-支払利息相当額 = 当前支付月利息 + 累加的未支付月份产生的利息
				lease.Interest = math.Floor(principalAmountcal*(rishiritsu/12)) + interestcount
				// 当回实际支付额编辑 = 支付额 - 优惠 + 变动额
				currentPay := pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
				// 支付最终回的场合
				if i == len(payments)-1 {
					// 利息情报-支払利息相当額 = 当回实际支付额 - 元本残高相当額
					lease.Interest = math.Floor(currentPay - principalAmount)
				}
				// 利息情报-元本返済相当額
				lease.Repayment = math.Floor(currentPay - lease.Interest)
				// 利息情报-元本残高相当額
				lease.Balance = principalAmount - lease.Repayment
				if paymentymd.Month() == prevymd.Month() {
					firstbalance = lease.Repayment + lease.Balance
				}
				lease.Firstbalance = firstbalance
				// K幂数取得(k=比較開始期首月到支付年月日的月数)
				payymd, _ := time.Parse("2006-01-02", pay.Paymentymd)
				k = float64(getGapMonths(leasestymd, payymd)) + 1
				// 利息情报-現在価値
				lease.Present = math.Floor(currentPay / math.Pow(1+(rishiritsu/12), k))
				// 前回元本残高相当額变化--下回计算用
				principalAmount = lease.Balance
				// 添加利息情报
				leaseData = append(leaseData, lease)
				// 支付年月日退避
				prevPaymentymd = paymentymd
				tf = false
				continue
			}
			// 获取前回支付与本回支付间隔月数
			cycleGap := getGapMonths(prevPaymentymd, paymentymd)
			// 累加计算间隔未支付月利息
			if cycleGap > 1 {
				// 累加计算未支付月利息
				for j := 0; j < (cycleGap - 1); j++ {
					// 单月利息
					interestsingle = math.Floor(principalAmountcal * (rishiritsu / 12))
					// 临时计算原本 = 临时计算原本 + 未支付利息
					principalAmountcal = principalAmountcal + interestsingle
					// 未支付利息累计
					interestcount += interestsingle
				}
			}
			// 利息情报-支払利息相当額 = 当前支付月利息 + 累加的未支付月份产生的利息
			lease.Interest = math.Floor(principalAmountcal*(rishiritsu/12)) + interestcount
			// 当回实际支付额编辑 = 支付额 - 优惠 + 变动额
			currentPay := pay.Paymentleasefee - pay.Incentives + pay.Paymentleasefeehendo
			// 支付最终回的场合
			if i == len(payments)-1 {
				// 利息情报-支払利息相当額 = 当回实际支付额 - 元本残高相当額
				lease.Interest = math.Floor(currentPay - principalAmount)
			}
			// 利息情报-元本返済相当額
			lease.Repayment = math.Floor(currentPay - lease.Interest)
			// 利息情报-元本残高相当額
			lease.Balance = principalAmount - lease.Repayment
			if paymentymd.Month() == prevymd.Month() {
				firstbalance = lease.Repayment + lease.Balance
			}
			lease.Firstbalance = firstbalance
			// k幂数取得(k=比較開始期首月到支付年月日的月数)
			k = float64(getGapMonths(leasestymd, paymentymd)) + 1
			// 利息情报-現在価値
			lease.Present = math.Floor(currentPay / math.Pow(1+(rishiritsu/12), k))
			// 前回元本残高相当額变化--下回计算用
			principalAmount = lease.Balance
			// 添加利息情报
			leaseData = append(leaseData, lease)
			// 支付年月日退避
			prevPaymentymd = paymentymd
		}
	}
	return leaseData, nil
}

// 偿还情报算出
func getRepayData(leasestsyoymd time.Time, genkakikan int, residualValue float64, boka float64, appID string, db string) (rps []typesx.RePayment, err error) {

	var repayData []typesx.RePayment
	// 默认期首月为1月
	var kishuMonth int64 = 1
	// 期首月取得
	cfg, err := configx.GetConfigVal(db, appID)
	if err != nil {
		loggerx.ErrorLog("getRepayData", err.Error())
		return nil, err
	}
	kishuYm := cfg.GetKishuYm()
	kishuMonth, _ = strconv.ParseInt(kishuYm, 10, 64)
	// 当期偿还月数算出
	var calMonths float64 = 12
	// 期首月份同租赁开始日月份相同时
	if int(kishuMonth) == int(leasestsyoymd.Month()) {
		if genkakikan < 12 {
			calMonths = float64(genkakikan)
		}
	}
	// 期首月份小于租赁开始日月份时
	if int(kishuMonth) < int(leasestsyoymd.Month()) {
		calMonths = float64(12 - int(leasestsyoymd.Month()) + int(kishuMonth))
		if calMonths > float64(genkakikan) {
			calMonths = float64(genkakikan)
		}
	}
	// 期首月份大于租赁开始日月份时
	if int(kishuMonth) > int(leasestsyoymd.Month()) {
		calMonths = float64(int(kishuMonth) - int(leasestsyoymd.Month()))
		if calMonths > float64(genkakikan) {
			calMonths = float64(genkakikan)
		}
	}
	// 剩余月数
	leftMonths := float64(genkakikan)

	// 计算
	for i := 0; i < genkakikan; i++ {
		// 当期偿还额合计
		var syokyakuCount float64 = 0
		// 当期偿还费算出
		var syoukyakucurrent float64 = 0
		// 使用権資産額期首簿価-残价保证额
		present := boka - residualValue
		// 使用権資産額期首簿価<>0&当期偿还月数<>0&残存月数<>0
		if boka != 0 && calMonths != 0 && leftMonths != 0 {
			// (使用権資産額期首簿価-残价保证额）/ 残存月数*当期偿还月数
			syoukyakucurrent = math.Floor(present / leftMonths * calMonths)
		}
		// 使用権資産額期首簿価-残价保证额 < 以上計算値の場合
		if math.Floor(present) < syoukyakucurrent {
			syoukyakucurrent = math.Floor(present)
		}
		// 当期偿还数据算出
		for j := 1; j <= int(calMonths); j++ {
			var repay typesx.RePayment
			// 期首薄价
			repay.Boka = boka
			// 偿却年月
			repay.Syokyakuymd = leasestsyoymd.Format("2006-01-02")
			// 偿却区分
			repay.Syokyakukbn = "通常"
			// 月别使用権償却額
			if j == 1 {
				// 首月使用権償却額 = 年额偿还费*对象月度/当期偿还月数
				repay.Syokyaku = math.Floor(syoukyakucurrent / calMonths)
			} else {
				// 年额偿还费 * 对象月度 / 当期偿还月数 - 年额偿却费 * (对象月度 - 1) / 当期偿还月数
				repay.Syokyaku = math.Floor(syoukyakucurrent*float64(j)/calMonths) - math.Floor(syoukyakucurrent*float64(j-1)/calMonths)
			}
			// 当期偿还额累计
			syokyakuCount = syokyakuCount + repay.Syokyaku
			// 期末薄价
			repay.Endboka = repay.Boka - syokyakuCount
			// 下月年月日算出
			leasestsyoym := leasestsyoymd.Format("200601")
			leasestsyoymd, _ = time.Parse("20060102", leasestsyoym+"01")
			leasestsyoymd = leasestsyoymd.AddDate(0, 1, 0)
			// 添加偿却数据
			repayData = append(repayData, repay)
		}
		// *************下期数据算出*************
		// 除去当期已计算月
		i = i + int(calMonths) - 1
		if i < genkakikan {
			// 使用権資産額期首簿価
			boka = boka - syokyakuCount
			// 偿还月数&剩余月数
			leftMonths = leftMonths - calMonths
			if leftMonths <= 12 {
				calMonths = leftMonths
			} else {
				calMonths = 12
			}
		}
	}
	return repayData, nil
}

// 偿还情报算出(開始時点から計算)
func getRepayDataStart(firstMonth time.Time, genkakikan int, residualValue float64, boka float64, appID string, db string, leasestymd time.Time) (rps []typesx.RePayment, err error) {

	var repayData []typesx.RePayment
	// 当期偿还月数算出
	var calMonths float64 = 12
	if firstMonth.Before(leasestymd) {
		firstMonth = leasestymd
	}
	// 剩余月数
	leftMonths := float64(genkakikan) - float64(getGapMonths(leasestymd, firstMonth))
	leftMonth := int(leftMonths)

	// 计算
	for i := 0; i < leftMonth; i++ {
		// 当期偿还额合计
		var syokyakuCount float64 = 0
		// 当期偿还费算出
		var syoukyakucurrent float64 = 0
		// 使用権資産額期首簿価-残价保证额
		present := boka - residualValue
		// 使用権資産額期首簿価<>0&当期偿还月数<>0&残存月数<>0
		if boka != 0 && calMonths != 0 && leftMonths != 0 {
			// (使用権資産額期首簿価-残价保证额）/ 残存月数*当期偿还月数
			syoukyakucurrent = math.Floor(present / leftMonths * calMonths)
		}
		// 使用権資産額期首簿価-残价保证额 < 以上計算値の場合
		if math.Floor(present) < syoukyakucurrent {
			syoukyakucurrent = math.Floor(present)
		}
		// 当期偿还数据算出
		for j := 1; j <= int(calMonths); j++ {
			var repay typesx.RePayment
			// 期首薄价
			repay.Boka = boka
			// 偿却年月
			repay.Syokyakuymd = firstMonth.Format("2006-01-02")
			// 偿却区分
			repay.Syokyakukbn = "通常"
			// 月别使用権償却額
			if j == 1 {
				// 首月使用権償却額 = 年额偿还费*对象月度/当期偿还月数
				repay.Syokyaku = math.Floor(syoukyakucurrent / calMonths)
			} else {
				// 年额偿还费 * 对象月度 / 当期偿还月数 - 年额偿却费 * (对象月度 - 1) / 当期偿还月数
				repay.Syokyaku = math.Floor(syoukyakucurrent*float64(j)/calMonths) - math.Floor(syoukyakucurrent*float64(j-1)/calMonths)
			}
			// 当期偿还额累计
			syokyakuCount = syokyakuCount + repay.Syokyaku
			// 期末薄价
			repay.Endboka = repay.Boka - syokyakuCount
			// 下月年月日算出
			firstym := firstMonth.Format("200601")
			firstMonth, _ = time.Parse("20060102", firstym+"01")
			firstMonth = firstMonth.AddDate(0, 1, 0)
			// 添加偿却数据
			repayData = append(repayData, repay)
		}
		// *************下期数据算出*************
		// 除去当期已计算月
		i = i + int(calMonths) - 1
		if i < genkakikan {
			// 使用権資産額期首簿価
			boka = boka - syokyakuCount
			// 偿还月数&剩余月数
			leftMonths = leftMonths - calMonths
			if leftMonths <= 12 {
				calMonths = leftMonths
			} else {
				calMonths = 12
			}
		}
	}
	return repayData, nil
}

// 偿还情报算出(取得時点に遡って計算)
func getRepayDataObtain(leasestsyoymd time.Time, genkakikan int, residualValue float64, boka float64, appID string, db string, firstMonthB time.Time) (rps []typesx.RePayment, err error) {

	var repayData []typesx.RePayment
	// 默认期首月为1月
	var kishuMonth int64 = 1
	// 期首月取得
	kishuMonth = int64(firstMonthB.Month())
	/* cfg, err := configx.GetConfigVal(db, appID)
	if err != nil {
		loggerx.ErrorLog("getRepayData", err.Error())
		return nil, err
	}
	kishuYm := cfg.GetKishuYm()
	kishuMonth, _ = strconv.ParseInt(kishuYm, 10, 64) */
	// 当期偿还月数算出
	var calMonths float64 = 12
	// 期首月份同租赁开始日月份相同时
	if int(kishuMonth) == int(leasestsyoymd.Month()) {
		if genkakikan < 12 {
			calMonths = float64(genkakikan)
		}
	}
	// 期首月份小于租赁开始日月份时
	if int(kishuMonth) < int(leasestsyoymd.Month()) {
		calMonths = float64(12 - int(leasestsyoymd.Month()) + int(kishuMonth))
		if calMonths > float64(genkakikan) {
			calMonths = float64(genkakikan)
		}
	}
	// 期首月份大于租赁开始日月份时
	if int(kishuMonth) > int(leasestsyoymd.Month()) {
		calMonths = float64(int(kishuMonth) - int(leasestsyoymd.Month()))
		if calMonths > float64(genkakikan) {
			calMonths = float64(genkakikan)
		}
	}
	// 剩余月数
	leftMonths := float64(genkakikan)

	// 计算
	for i := 0; i < genkakikan; i++ {
		// 当期偿还额合计
		var syokyakuCount float64 = 0
		// 当期偿还费算出
		var syoukyakucurrent float64 = 0
		// 使用権資産額期首簿価-残价保证额
		present := boka - residualValue
		// 使用権資産額期首簿価<>0&当期偿还月数<>0&残存月数<>0
		if boka != 0 && calMonths != 0 && leftMonths != 0 {
			// (使用権資産額期首簿価-残价保证额）/ 残存月数*当期偿还月数
			syoukyakucurrent = math.Floor(present / leftMonths * calMonths)
		}
		// 使用権資産額期首簿価-残价保证额 < 以上計算値の場合
		if math.Floor(present) < syoukyakucurrent {
			syoukyakucurrent = math.Floor(present)
		}
		// 当期偿还数据算出
		for j := 1; j <= int(calMonths); j++ {
			var repay typesx.RePayment
			// 期首薄价
			repay.Boka = boka
			// 偿却年月
			repay.Syokyakuymd = leasestsyoymd.Format("2006-01-02")
			// 偿却区分
			repay.Syokyakukbn = "通常"
			// 月别使用権償却額
			if j == 1 {
				// 首月使用権償却額 = 年额偿还费*对象月度/当期偿还月数
				repay.Syokyaku = math.Floor(syoukyakucurrent / calMonths)
			} else {
				// 年额偿还费 * 对象月度 / 当期偿还月数 - 年额偿却费 * (对象月度 - 1) / 当期偿还月数
				repay.Syokyaku = math.Floor(syoukyakucurrent*float64(j)/calMonths) - math.Floor(syoukyakucurrent*float64(j-1)/calMonths)
			}
			// 当期偿还额累计
			syokyakuCount = syokyakuCount + repay.Syokyaku
			// 期末薄价
			repay.Endboka = repay.Boka - syokyakuCount
			// 下月年月日算出
			leasestsyoym := leasestsyoymd.Format("200601")
			leasestsyoymd, _ = time.Parse("20060102", leasestsyoym+"01")
			leasestsyoymd = leasestsyoymd.AddDate(0, 1, 0)
			// 添加偿却数据
			repayData = append(repayData, repay)
		}
		// *************下期数据算出*************
		// 除去当期已计算月
		i = i + int(calMonths) - 1
		if i < genkakikan {
			// 使用権資産額期首簿価
			boka = boka - syokyakuCount
			// 偿还月数&剩余月数
			leftMonths = leftMonths - calMonths
			if leftMonths <= 12 {
				calMonths = leftMonths
			} else {
				calMonths = 12
			}
		}
	}
	return repayData, nil
}

// GeneratePay 生成支付数据(租赁系统用)
func GeneratePay(q typesx.PayParam) (payData []typesx.Payment, err error) {
	// 支付开始日
	paymentstymd := q.Paymentstymd
	// 支付周期
	paymentcycle := q.Paymentcycle
	// 支付日
	paymentday := q.Paymentday
	// 支付回数
	paymentcounts := q.Paymentcounts
	// 残价保证额
	residualValue := q.ResidualValue
	// 支付金额
	paymentleasefee := q.Paymentleasefee
	// 购入行使权金额
	optionToPurchase := q.OptionToPurchase

	// 循环支付回数生成结果
	for i := 0; i < paymentcounts; i++ {
		var pay typesx.Payment
		// 支付回数
		pay.Paymentcount = i + 1
		// 支付金额
		pay.Paymentleasefee = paymentleasefee
		// 支付年月日
		pay.Paymentymd = paymentstymd.Format("2006-01-02")
		// 其他设置
		pay.Paymentleasefeehendo = 0
		pay.Incentives = 0
		pay.Sonotafee = 0
		pay.Kaiyakuson = 0
		pay.PaymentType = "支払"
		pay.Fixed = false
		// -------获取下期支付年月-------
		// 支付月份加上支付周期
		if i == paymentcounts-1 {
			paymentstym := paymentstymd.Format("200601")
			paymentstymd, err = time.Parse("20060102", paymentstym+"01")
			paymentstymd = paymentstymd.AddDate(0, 1, 0)
			// 配上约定支付日
			paymentstymd = getPayDate(paymentstymd, paymentday)
		} else {
			paymentstym := paymentstymd.Format("200601")
			paymentstymd, err = time.Parse("20060102", paymentstym+"01")
			if err != nil {
				loggerx.ErrorLog("generatePay", err.Error())
				return
			}
			paymentstymd = paymentstymd.AddDate(0, paymentcycle, 0)
			// 配上约定支付日
			paymentstymd = getPayDate(paymentstymd, paymentday)
		}

		// 添加支付数据
		payData = append(payData, pay)
	}

	// 残价保证额&购入行使权金额==二选一
	if residualValue != 0 {
		var pay typesx.Payment
		// 支付回数
		pay.Paymentcount = len(payData) + 1
		// 支付金额
		pay.Paymentleasefee = residualValue
		// 支付年月
		pay.Paymentymd = paymentstymd.Format("2006-01-02")
		// 其他设置
		pay.Paymentleasefeehendo = 0
		pay.Incentives = 0
		pay.Sonotafee = 0
		pay.Kaiyakuson = 0
		pay.PaymentType = "残価保証額"
		pay.Fixed = true
		// -------获取下期支付年月-------
		// 支付月份加上支付周期
		paymentstym := paymentstymd.Format("200601")
		paymentstymd, err = time.Parse("20060102", paymentstym+"01")
		if err != nil {
			loggerx.ErrorLog("generatePay", err.Error())
			return
		}
		paymentstymd = paymentstymd.AddDate(0, 1, 0)
		// 配上约定支付日
		paymentstymd = getPayDate(paymentstymd, paymentday)

		// 添加支付数据
		payData = append(payData, pay)
	}
	if optionToPurchase != 0 {
		var pay typesx.Payment
		// 支付回数
		pay.Paymentcount = len(payData) + 1
		// 支付金额
		pay.Paymentleasefee = optionToPurchase
		// 支付年月
		pay.Paymentymd = paymentstymd.Format("2006-01-02")
		// 其他设置
		pay.Paymentleasefeehendo = 0
		pay.Incentives = 0
		pay.Sonotafee = 0
		pay.Kaiyakuson = 0
		pay.PaymentType = "購入オプション行使価額"
		pay.Fixed = true
		// -------获取下期支付年月-------
		// 支付月份加上支付周期
		paymentstym := paymentstymd.Format("200601")
		paymentstymd, err = time.Parse("20060102", paymentstym+"01")
		if err != nil {
			loggerx.ErrorLog("compute", err.Error())
			return
		}
		paymentstymd = paymentstymd.AddDate(0, 1, 0)
		// 配上约定支付日
		paymentstymd = getPayDate(paymentstymd, paymentday)

		// 添加支付数据
		payData = append(payData, pay)
	}

	return
}
