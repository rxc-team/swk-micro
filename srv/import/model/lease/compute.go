package lease

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	timeconv "github.com/Andrew-M-C/go.timeconv"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/import/common/filex"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
)

// InsertResult 少额预算返回
type InsertResult struct {
	TemplateID  string `json:"template_id" bson:"template_id"`
	attachItems attachData
}

// ComputeResult 新规契约预算返回
type ComputeResult struct {
	TemplateID  string  `json:"template_id" bson:"template_id"`
	KiSyuBoka   float64 `json:"kisyuboka" bson:"kisyuboka"` // 原始取得价值
	Hkkjitenzan float64 `json:"hkkjitenzan" bson:"hkkjitenzan"`
	Sonnekigaku float64 `json:"sonnekigaku" bson:"sonnekigaku"`
	attachItems attachData
}

// ChangeResult 契约情报变更返回
type ChangeResult struct {
	TemplateID  string `json:"template_id" bson:"template_id"`
	attachItems attachData
}

// CancelResult 中途解约预算返回
type CancelResult struct {
	TemplateID  string  `json:"template_id" bson:"template_id"` // 临时数据ID
	RemainDebt  float64 `json:"remaindebt" bson:"remaindebt"`   // 解約時元本残高
	Lossgaku    float64 `json:"lossgaku" bson:"lossgaku"`       // 中途解約による除却損金额
	attachItems attachData
}

// ExpireResult 满了预算返回
type ExpireResult struct {
	TemplateID  string  `json:"template_id" bson:"template_id"` // 临时数据ID
	Leftgaku    float64 `json:"leftgaku" bson:"leftgaku"`       // 满了時剩余价值
	attachItems attachData
}

// debtResult 利息偿还情报参数
type DebtResult struct {
	TemplateID         string  `json:"template_id" bson:"template_id"`                 // 临时数据ID
	KiSyuBoka          float64 `json:"kisyuboka" bson:"kisyuboka"`                     // 原始取得价值
	OShisannsougaku    float64 `json:"o_shisannsougaku" bson:"o_shisannsougaku"`       // 变更前使用権資産額
	Shisannsougaku     float64 `json:"shisannsougaku" bson:"shisannsougaku"`           // 变更后使用権資産額
	OLeasesaimusougaku float64 `json:"o_leasesaimusougaku" bson:"o_leasesaimusougaku"` // 变更前租赁负债额
	Leasesaimusougaku  float64 `json:"leasesaimusougaku" bson:"leasesaimusougaku"`     // 变更后租赁负债额
	Shisannsagaku      float64 `json:"shisannsagaku" bson:"shisannsagaku"`             // 使用权资产差额
	Leasesaimusagaku   float64 `json:"leasesaimusagaku" bson:"leasesaimusagaku"`       // 租赁负债差额
	Sonnekigaku        float64 `json:"sonnekigaku" bson:"sonnekigaku"`                 // 损益额
	attachItems        attachData
}

// PayParam 支付情报参数
type PayParam struct {
	Paymentstymd     time.Time `json:"paymentstymd" bson:"paymentstymd"`         // 支付开始日
	Paymentcycle     int       `json:"paymentcycle" bson:"paymentcycle"`         // 支付周期
	Paymentday       int       `json:"paymentday" bson:"paymentday"`             // 支付日
	Paymentcounts    int       `json:"paymentcounts" bson:"paymentcounts"`       // 支付回数
	ResidualValue    float64   `json:"residualValue" bson:"residualValue"`       // 残价保证额
	Paymentleasefee  float64   `json:"paymentleasefee" bson:"paymentleasefee"`   // 支付金额
	OptionToPurchase float64   `json:"optionToPurchase" bson:"optionToPurchase"` // 购入行使权金额
	Firstleasefee    float64   `json:"firstleasefee" bson:"firstleasefee"`       // 初回リース料
	Finalleasefee    float64   `json:"finalleasefee" bson:"firstleasefee"`       // 最终回リース料
	Keiyakuno        string    `json:"keiyakuno" bson:"keiyakuno"`               // 契约番号
	// Leasekaishacd    string    `json:"leasekaishacd" bson:"leasekaishacd"`       // 租赁会社
}

// LRParam 契约追加情报参数
type LRParam struct {
	ResidualValue           float64           `json:"residualValue" bson:"residualValue"`                     // 残价保证额
	Rishiritsu              float64           `json:"rishiritsu" bson:"rishiritsu"`                           // 割引率
	Leasestymd              time.Time         `json:"leasestymd" bson:"leasestymd"`                           // 租赁开始日
	CancellationRightOption bool              `json:"cancellationrightoption" bson:"cancellationrightoption"` // 解約行使権オプション
	Leasekikan              int               `json:"leasekikan" bson:"leasekikan"`                           // 租赁期间
	ExtentionOption         int               `json:"extentionOption" bson:"extentionOption"`                 // 延长租赁期间
	PaymentsAtOrPrior       float64           `json:"paymentsAtOrPrior" bson:"paymentsAtOrPrior"`             // 前払リース料
	IncentivesAtOrPrior     float64           `json:"incentivesAtOrPrior" bson:"incentivesAtOrPrior"`         // リース・インセンティブ（前払）
	InitialDirectCosts      float64           `json:"initialDirectCosts" bson:"initialDirectCosts"`           // 当初直接費用
	RestorationCosts        float64           `json:"restorationCosts" bson:"restorationCosts"`               // 原状回復コスト
	Assetlife               int               `json:"assetlife" bson:"assetlife"`                             // 耐用年限
	Torihikikbn             string            `json:"torihikikbn" bson:"torihikikbn"`                         // 取引判定区分
	HandleMonth             string            `json:"hadnle_month" bson:"hadnle_month"`                       // 处理月度
	BeginMonth              string            `json:"begin_month" bson:"begin_month"`                         // 期首月度
	Payments                []Payment         `json:"payments" bson:"payments"`                               // 支付情报
	DsMap                   map[string]string `json:"ds_map" bson:"ds_map"`                                   // 台账情报
	Sykshisankeisan         string            `json:"sykshisankeisan" bson:"sykshisankeisan"`                 // 使用権資産
	FirstMonth              string            `json:"firstMonth" bson:"firstMonth"`                           // 比較開始期首月
	// Leasekaishacd           string            `json:"leasekaishacd" bson:"leasekaishacd"`                     // 租赁会社
	Item *item.Item `json:"item_map" bson:"item_map"`
	seq  string
}

// ChangeParam 契约情报变更参数
type ChangeParam struct {
	Henkouymd   string                 `json:"henkouymd" bson:"henkouymd"`       // 变更年月
	HandleMonth string                 `json:"hadnle_month" bson:"hadnle_month"` // 处理月度
	DsMap       map[string]string      `json:"ds_map" bson:"ds_map"`             // 台账情报
	Item        *item.Item             `json:"item" bson:"item"`                 // 旧数据
	Change      map[string]*item.Value `json:"change" bson:"change"`             // 变更的数据
	seq         string
}

// DebtParam 债务变更情报参数
type DebtParam struct {
	Kaiyakuymd              string                 `json:"kaiyakuymd" bson:"kaiyakuymd"`                           // 解约年月
	Henkouymd               string                 `json:"henkouymd" bson:"henkouymd"`                             // 变更年月
	Leasestymd              string                 `json:"leasestymd" bson:"leasestymd"`                           // 租赁开始日
	HandleMonth             string                 `json:"hadnle_month" bson:"hadnle_month"`                       // 处理月度
	BeginMonth              string                 `json:"begin_month" bson:"begin_month"`                         // 期首月度
	CancellationRightOption bool                   `json:"cancellationrightoption" bson:"cancellationrightoption"` // 解約行使権オプション
	Leasekikan              int                    `json:"leasekikan" bson:"leasekikan"`                           // 租赁期间
	ExtentionOption         int                    `json:"extentionOption" bson:"extentionOption"`                 // 延长租赁期间
	Keiyakuno               string                 `json:"keiyakuno" bson:"keiyakuno"`                             // 契约番号
	Rishiritsu              float64                `json:"rishiritsu" bson:"rishiritsu"`                           // 割引率
	ResidualValue           float64                `json:"residualValue" bson:"residualValue"`                     // 残价保证额
	Assetlife               int                    `json:"assetlife" bson:"assetlife"`                             // 耐用年限
	Torihikikbn             string                 `json:"torihikikbn" bson:"torihikikbn"`                         // 取引判定区分
	Percentage              float64                `json:"percentage" bson:"percentage"`                           // 剩余资产百分比
	Payments                []Payment              `json:"payments" bson:"payments"`                               // 支付情报
	DsMap                   map[string]string      `json:"ds_map" bson:"ds_map"`                                   // 台账情报
	Item                    *item.Item             `json:"item" bson:"item"`                                       // 旧数据
	Change                  map[string]*item.Value `json:"change" bson:"change"`                                   // 变更的数据
	seq                     string
}

// ExpireParam 契约满了情报参数
type ExpireParam struct {
	Henkouymd         string                 `json:"henkouymd" bson:"henkouymd"`                 // 变更年月
	Torihikikbn       string                 `json:"torihikikbn" bson:"torihikikbn"`             // 取引判定区分
	HandleMonth       string                 `json:"hadnle_month" bson:"hadnle_month"`           // 处理月度
	Expiresyokyakukbn string                 `json:"expiresyokyakukbn" bson:"expiresyokyakukbn"` // リース満了償却区分
	Keiyakuno         string                 `json:"keiyakuno" bson:"keiyakuno"`                 // 契约番号
	DsMap             map[string]string      `json:"ds_map" bson:"ds_map"`
	Item              *item.Item             `json:"item" bson:"item"`     // 旧数据
	Change            map[string]*item.Value `json:"change" bson:"change"` // 台账情报
	seq               string
}

// CancelParam 中途解约情报参数
type CancelParam struct {
	Kaiyakuymd  string                 `json:"kaiyakuymd" bson:"kaiyakuymd"`     // 解約年月日
	HandleMonth string                 `json:"hadnle_month" bson:"hadnle_month"` // 处理月度
	Keiyakuno   string                 `json:"keiyakuno" bson:"keiyakuno"`       // 契约番号
	DsMap       map[string]string      `json:"ds_map" bson:"ds_map"`             // 台账情报
	Item        *item.Item             `json:"item" bson:"item"`                 // 旧数据
	Change      map[string]*item.Value `json:"change" bson:"change"`
	seq         string
}

// Payment 支付数据
type Payment struct {
	// Leasekaishacd        string  `json:"leasekaishacd" bson:"leasekaishacd"`               // 租赁会社
	Paymentcount         int     `json:"paymentcount" bson:"paymentcount"`                 // 支付回数
	PaymentType          string  `json:"paymentType" bson:"paymentType"`                   // 支付类型
	Paymentymd           string  `json:"paymentymd" bson:"paymentymd"`                     // 支付年月日
	Paymentleasefee      float64 `json:"paymentleasefee" bson:"paymentleasefee"`           // 支付金额
	Paymentleasefeehendo float64 `json:"paymentleasefeehendo" bson:"paymentleasefeehendo"` // 变更支付金额
	Incentives           float64 `json:"incentives" bson:"incentives"`                     // 优惠金额
	Sonotafee            float64 `json:"sonotafee" bson:"sonotafee"`                       // 其他金额
	Kaiyakuson           float64 `json:"kaiyakuson" bson:"kaiyakuson"`                     // 解约损失
	Fixed                bool    `json:"fixed" bson:"fixed"`                               // 修正否
	Firstleasefee        float64 `json:"firstleasefee" bson:"firstleasefee"`               // 初回支付金额
	Finalleasefee        float64 `json:"finalleasefee" bson:"finalleasefee"`               // 最终回支付金额
	Keiyakuno            string  `json:"keiyakuno" bson:"keiyakuno"`                       // 契约番号
}

// Lease 利息数据
type Lease struct {
	// Leasekaishacd string  `json:"leasekaishacd" bson:"leasekaishacd"` // 租赁会社
	Keiyakuno    string  `json:"keiyakuno" bson:"keiyakuno"`       // 契约番号
	Interest     float64 `json:"interest" bson:"interest"`         // 支付利息相当额
	Repayment    float64 `json:"repayment" bson:"repayment"`       // 元本返済相当額
	Balance      float64 `json:"balance" bson:"balance"`           // 元本残高相当額
	Firstbalance float64 `json:"firstbalance" bson:"firstbalance"` // 期首元本残高
	Present      float64 `json:"present" bson:"present"`           // 現在価値
	Paymentymd   string  `json:"paymentymd" bson:"paymentymd"`     // 支付年月
}

// RePayment 偿还数据
type RePayment struct {
	// Leasekaishacd string  `json:"leasekaishacd" bson:"leasekaishacd"` // 租赁会社
	Keiyakuno   string  `json:"keiyakuno" bson:"keiyakuno"`     // 契约番号
	Syokyakukbn string  `json:"syokyakukbn" bson:"syokyakukbn"` // 償却区分
	Endboka     float64 `json:"endboka" bson:"endboka"`         // 月末薄价
	Boka        float64 `json:"boka" bson:"boka"`               // 期首薄价
	Syokyaku    float64 `json:"syokyaku" bson:"syokyaku"`       // 偿却额
	Syokyakuymd string  `json:"syokyakuymd" bson:"syokyakuymd"` // 偿却年月
}

// insertPay 生成支付数据
func insertPay(db, appID, userID string, dsMap map[string]string, ps []Payment) (result *InsertResult, err error) {
	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &InsertResult{}

	var attachItems attachData

	// 支付情报
	for _, pay := range ps {
		payyear, _ := strconv.Atoi(pay.Paymentymd[0:4])
		paymonth, _ := strconv.Atoi(pay.Paymentymd[5:7])
		items := make(map[string]*item.Value)
		items["paymentcount"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentcount),
		}
		items["paymentType"] = &item.Value{
			DataType: "text",
			Value:    pay.PaymentType,
		}
		items["paymentymd"] = &item.Value{
			DataType: "date",
			Value:    pay.Paymentymd,
		}
		items["paymentleasefee"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentleasefee),
		}
		items["paymentleasefeehendo"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentleasefeehendo),
		}
		items["incentives"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Incentives),
		}
		items["sonotafee"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Sonotafee),
		}
		items["kaiyakuson"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Kaiyakuson),
		}
		items["keiyakuno"] = &item.Value{
			DataType: "text",
			Value:    pay.Keiyakuno,
		}
		items["year"] = &item.Value{
			DataType: "number",
			Value:    strconv.Itoa(payyear),
		}
		items["month"] = &item.Value{
			DataType: "number",
			Value:    strconv.Itoa(paymonth),
		}
		data := item.AttachItems{
			Items:       items,
			DatastoreId: dsMap["paymentStatus"],
		}
		attachItems = append(attachItems, &data)
	}

	result.TemplateID = templateID
	result.attachItems = attachItems

	return result, nil
}

// compute 计算利息和偿还数据(租赁系统用)
func compute(db, appID, userID string, p LRParam) (result *ComputeResult, err error) {

	if p.Sykshisankeisan == "1" {
		// 開始時点から計算

		// 生成临时数据ID
		uid := uuid.Must(uuid.NewRandom())
		templateID := uid.String()

		result = &ComputeResult{}

		var attachItems attachData

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
		// if p.Torihikikbn != "1" {
		// 	// 移転外
		// 	if genkakikan > leasekikanTotal {
		// 		// 取租赁期间与耐用期间中较短者
		// 		genkakikan = leasekikanTotal
		// 	}
		// }

		// 比較開始時点から計算
		hkkjitenzan, presentTotalRemain := getLeaseDebt(p.Payments, rishiritsu, p.FirstMonth, residualValue)
		// 元本残高相当额(初回 = 租赁负债额) => 租赁负债额算出(现在价值累计)
		var principalAmount float64 = presentTotalRemain

		var leaseData []Lease
		var repayData []RePayment

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

		// 履历情报
		newItemMap := make(map[string]*item.Value)
		itMap := make(map[string]interface{})
		for key, value := range p.Item.GetItems() {
			newItemMap[key] = value
			itMap[key] = value
		}

		// 获取共通字段
		keiyakuItem := p.Item.Items["keiyakuno"].Value

		// 获取会社信息
		// kaisyaItem := p.Item.Items["leasekaishacd"].Value

		// 获取分類コード信息
		bunruicdItem := p.Item.Items["bunruicd"].Value

		// 获取管理部門
		segmentcdItem := p.Item.Items["segmentcd"].Value

		// 支付情报
		for _, pay := range p.Payments {
			payyear, _ := strconv.Atoi(pay.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(pay.Paymentymd[5:7])
			items := make(map[string]*item.Value)
			// 插入共通数据
			items["keiyakuno"] = &item.Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			// items["leasekaishacd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    kaisyaItem,
			// }
			items["bunruicd"] = &item.Value{
				DataType: "lookup",
				Value:    bunruicdItem,
			}
			items["segmentcd"] = &item.Value{
				DataType: "lookup",
				Value:    segmentcdItem,
			}
			items["paymentcount"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Paymentcount),
			}
			items["paymentType"] = &item.Value{
				DataType: "text",
				Value:    pay.PaymentType,
			}
			items["paymentymd"] = &item.Value{
				DataType: "date",
				Value:    pay.Paymentymd,
			}
			items["paymentleasefee"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Paymentleasefee),
			}
			items["paymentleasefeehendo"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Paymentleasefeehendo),
			}
			items["incentives"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Incentives),
			}
			items["sonotafee"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Sonotafee),
			}
			items["kaiyakuson"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Kaiyakuson),
			}
			items["year"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}
			data := item.AttachItems{
				Items:       items,
				DatastoreId: p.DsMap["paymentStatus"],
			}

			attachItems = append(attachItems, &data)
		}

		// 利息情报
		for _, lease := range leaseData {
			payyear, _ := strconv.Atoi(lease.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(lease.Paymentymd[5:7])
			items := make(map[string]*item.Value)
			// 插入共通数据
			items["keiyakuno"] = &item.Value{
				DataType: "lookup",
				Value:    string(keiyakuItem),
			}
			// items["leasekaishacd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    kaisyaItem,
			// }
			// items["bunruicd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    bunruicdItem,
			// }
			// items["segmentcd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    segmentcdItem,
			// }
			items["interest"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Interest),
			}
			items["repayment"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Repayment),
			}
			items["balance"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Balance),
			}
			items["firstbalance"] = &item.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Firstbalance, 'f', -1, 64),
			}
			items["present"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Present),
			}
			items["paymentymd"] = &item.Value{
				DataType: "date",
				Value:    lease.Paymentymd,
			}
			items["year"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}

			data := item.AttachItems{
				Items:       items,
				DatastoreId: p.DsMap["paymentInterest"],
			}

			attachItems = append(attachItems, &data)
		}

		// 処理月度の先月までの償却費の累計額
		var preDepreciationTotal float64 = 0
		// 偿还情报
		for _, rp := range repayData {
			rpyear, _ := strconv.Atoi(rp.Syokyakuymd[0:4])
			rpmonth, _ := strconv.Atoi(rp.Syokyakuymd[5:7])
			items := make(map[string]*item.Value)
			// 插入共通数据
			items["keiyakuno"] = &item.Value{
				DataType: "lookup",
				Value:    string(keiyakuItem),
			}
			// items["leasekaishacd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    kaisyaItem,
			// }
			items["bunruicd"] = &item.Value{
				DataType: "lookup",
				Value:    bunruicdItem,
			}
			items["segmentcd"] = &item.Value{
				DataType: "lookup",
				Value:    segmentcdItem,
			}

			items["endboka"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Endboka),
			}
			items["boka"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Boka),
			}
			items["syokyaku"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Syokyaku),
			}
			items["syokyakuymd"] = &item.Value{
				DataType: "date",
				Value:    rp.Syokyakuymd,
			}
			items["syokyakukbn"] = &item.Value{
				DataType: "text",
				Value:    rp.Syokyakukbn,
			}
			items["year"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpyear),
			}
			items["month"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpmonth),
			}

			data := item.AttachItems{
				Items:       items,
				DatastoreId: p.DsMap["repayment"],
			}

			if rp.Syokyakuymd[:7] < p.HandleMonth {
				preDepreciationTotal += rp.Syokyaku
			}

			attachItems = append(attachItems, &data)
		}

		rirekiNo := p.seq

		// 契約履歴番号
		newItemMap["no"] = &item.Value{
			DataType: "text",
			Value:    rirekiNo,
		}
		// 修正区分编辑
		newItemMap["zengokbn"] = &item.Value{
			DataType: "options",
			Value:    "after",
		}
		// 对接区分编辑
		newItemMap["dockkbn"] = &item.Value{
			DataType: "options",
			Value:    "undo",
		}

		newItemMap["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    string(keiyakuItem),
		}

		newItemMap["leaseTotal"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(hkkjitenzan),
		}
		newItemMap["presentTotal"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(presentTotalRemain),
		}
		newItemMap["preDepreciationTotal"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(preDepreciationTotal),
		}

		data := item.AttachItems{
			Items:       newItemMap,
			DatastoreId: p.DsMap["rireki"],
		}

		attachItems = append(attachItems, &data)

		result.TemplateID = templateID
		result.attachItems = attachItems
		result.Hkkjitenzan = hkkjitenzan
		result.Sonnekigaku = 0

		return result, nil

	} else {
		// 生成临时数据ID
		uid := uuid.Must(uuid.NewRandom())
		templateID := uid.String()

		result = &ComputeResult{}

		var attachItems attachData

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
		var firstMonthB, err1 = time.Parse("2006-1-02", firstMonth+"-01")
		//firstMonth, _ := time.Parse("2006-1", p.FirstMonth)
		if err1 != nil {
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
		//  耐用年限(月单位)
		// yms := p.Assetlife * 12
		// 減価償却期間算出
		genkakikan := leasekikanTotal
		// if p.Torihikikbn != "1" {
		// 	// 移転外
		// 	if genkakikan > leasekikanTotal {
		// 		// 取租赁期间与耐用期间中较短者
		// 		genkakikan = leasekikanTotal
		// 	}
		// }

		// 比較開始時点から計算
		hkkjitenzan, presentTotalRemain := getLeaseDebt(p.Payments, rishiritsu, p.FirstMonth, residualValue)
		// 根据参数传入的支付情报算出现在价值合计
		presentTotal, leaseTotal := getLeaseTotal(p.Payments, leasestymd, rishiritsu)
		// 元本残高相当额(初回 = 租赁负债额) => 租赁负债额算出(现在价值累计)
		var principalAmount float64 = presentTotal

		var leaseData []Lease
		var repayData []RePayment

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
		// 比較開始期首月取得
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

		// 履历情报
		newItemMap := make(map[string]*item.Value)
		itMap := make(map[string]interface{})
		for key, value := range p.Item.GetItems() {
			newItemMap[key] = value
			itMap[key] = value
		}

		// 获取共通字段
		keiyakuItem := p.Item.Items["keiyakuno"].Value

		// 获取会社信息
		// kaisyaItem := p.Item.Items["leasekaishacd"].Value

		// 获取分類コード信息
		bunruicdItem := p.Item.Items["bunruicd"].Value

		// 获取管理部門
		segmentcdItem := p.Item.Items["segmentcd"].Value

		// 支付情报
		for _, pay := range p.Payments {
			payyear, _ := strconv.Atoi(pay.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(pay.Paymentymd[5:7])
			items := make(map[string]*item.Value)
			// 插入共通数据
			items["keiyakuno"] = &item.Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			// items["leasekaishacd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    kaisyaItem,
			// }
			items["bunruicd"] = &item.Value{
				DataType: "lookup",
				Value:    bunruicdItem,
			}
			items["segmentcd"] = &item.Value{
				DataType: "lookup",
				Value:    segmentcdItem,
			}
			items["paymentcount"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Paymentcount),
			}
			items["paymentType"] = &item.Value{
				DataType: "text",
				Value:    pay.PaymentType,
			}
			items["paymentymd"] = &item.Value{
				DataType: "date",
				Value:    pay.Paymentymd,
			}
			items["paymentleasefee"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Paymentleasefee),
			}
			items["paymentleasefeehendo"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Paymentleasefeehendo),
			}
			items["incentives"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Incentives),
			}
			items["sonotafee"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Sonotafee),
			}
			items["kaiyakuson"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(pay.Kaiyakuson),
			}
			items["year"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}
			data := item.AttachItems{
				Items:       items,
				DatastoreId: p.DsMap["paymentStatus"],
			}

			attachItems = append(attachItems, &data)
		}

		// 利息情报
		for _, lease := range leaseData {
			payyear, _ := strconv.Atoi(lease.Paymentymd[0:4])
			paymonth, _ := strconv.Atoi(lease.Paymentymd[5:7])
			items := make(map[string]*item.Value)
			// 插入共通数据
			items["keiyakuno"] = &item.Value{
				DataType: "lookup",
				Value:    string(keiyakuItem),
			}
			// items["leasekaishacd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    kaisyaItem,
			// }
			// items["bunruicd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    bunruicdItem,
			// }
			// items["segmentcd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    segmentcdItem,
			// }
			items["interest"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Interest),
			}
			items["repayment"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Repayment),
			}
			items["balance"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Balance),
			}
			items["firstbalance"] = &item.Value{
				DataType: "number",
				Value:    strconv.FormatFloat(lease.Firstbalance, 'f', -1, 64),
			}
			items["present"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(lease.Present),
			}
			items["paymentymd"] = &item.Value{
				DataType: "date",
				Value:    lease.Paymentymd,
			}
			items["year"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(payyear),
			}
			items["month"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(paymonth),
			}

			data := item.AttachItems{
				Items:       items,
				DatastoreId: p.DsMap["paymentInterest"],
			}

			attachItems = append(attachItems, &data)
		}

		// 処理月度の先月までの償却費の累計額
		var preDepreciationTotal float64 = 0
		// 偿还情报
		for _, rp := range repayData {
			rpyear, _ := strconv.Atoi(rp.Syokyakuymd[0:4])
			rpmonth, _ := strconv.Atoi(rp.Syokyakuymd[5:7])
			items := make(map[string]*item.Value)
			// 插入共通数据
			items["keiyakuno"] = &item.Value{
				DataType: "lookup",
				Value:    string(keiyakuItem),
			}
			// items["leasekaishacd"] = &item.Value{
			// 	DataType: "lookup",
			// 	Value:    kaisyaItem,
			// }
			items["bunruicd"] = &item.Value{
				DataType: "lookup",
				Value:    bunruicdItem,
			}
			items["segmentcd"] = &item.Value{
				DataType: "lookup",
				Value:    segmentcdItem,
			}

			items["endboka"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Endboka),
			}
			items["boka"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Boka),
			}
			items["syokyaku"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Syokyaku),
			}
			items["syokyakuymd"] = &item.Value{
				DataType: "date",
				Value:    rp.Syokyakuymd,
			}
			items["syokyakukbn"] = &item.Value{
				DataType: "text",
				Value:    rp.Syokyakukbn,
			}
			items["year"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpyear),
			}
			items["month"] = &item.Value{
				DataType: "number",
				Value:    strconv.Itoa(rpmonth),
			}

			data := item.AttachItems{
				Items:       items,
				DatastoreId: p.DsMap["repayment"],
			}

			if rp.Syokyakuymd[:7] < p.HandleMonth {
				preDepreciationTotal += rp.Syokyaku
			}

			attachItems = append(attachItems, &data)
		}

		rirekiNo := p.seq

		// 契約履歴番号
		newItemMap["no"] = &item.Value{
			DataType: "text",
			Value:    rirekiNo,
		}
		// 修正区分编辑
		newItemMap["zengokbn"] = &item.Value{
			DataType: "options",
			Value:    "after",
		}
		// 对接区分编辑
		newItemMap["dockkbn"] = &item.Value{
			DataType: "options",
			Value:    "undo",
		}

		newItemMap["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    string(keiyakuItem),
		}

		newItemMap["leaseTotal"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(leaseTotal),
		}
		newItemMap["presentTotal"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(presentTotal),
		}
		newItemMap["preDepreciationTotal"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(preDepreciationTotal),
		}

		data := item.AttachItems{
			Items:       newItemMap,
			DatastoreId: p.DsMap["rireki"],
		}

		attachItems = append(attachItems, &data)

		result.TemplateID = templateID
		result.attachItems = attachItems
		result.Hkkjitenzan = hkkjitenzan
		result.Sonnekigaku = sonnekigaku

		return result, nil
	}
}

// changeCompute 计算情报变更数据(租赁系统用)
func changeCompute(db, appID, userID string, payData []Payment, leaseData []Lease, repayData []RePayment, p ChangeParam) (result *ChangeResult, err error) {

	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &ChangeResult{}

	var attachItems attachData

	// 变更年月
	henkouym, err := time.Parse("2006-01", p.Henkouymd[0:7])
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
		if repay.Syokyakuymd[:7] <= p.Henkouymd[:7] {
			oldDepreciationTotal += repay.Syokyaku
		}
	}

	// 履历情报
	oldItemMap := make(map[string]*item.Value)
	newItemMap := make(map[string]*item.Value)
	itMap := make(map[string]interface{})

	for key, value := range p.Item.GetItems() {
		oldItemMap[key] = value
		newItemMap[key] = value
		itMap[key] = value
	}

	for key, value := range p.Change {
		newItemMap[key] = value
		itMap[key] = value
	}

	// 契約履歴番号
	rirekiNo := p.seq
	// 履历番号
	newItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	oldItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	// 修正区分编辑
	newItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "after",
	}
	oldItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "before",
	}
	// 对接区分编辑
	newItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	oldItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	// 操作区分编辑
	newItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "infoalter",
	}
	oldItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "infoalter",
	}

	keiyakuno := p.Item.Items["keiyakuno"].GetValue()

	newItemMap["keiyakuno"] = &item.Value{
		DataType: "lookup",
		Value:    keiyakuno,
	}
	oldItemMap["keiyakuno"] = &item.Value{
		DataType: "lookup",
		Value:    keiyakuno,
	}

	newItemMap["oldDepreciationTotal"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(oldDepreciationTotal),
	}

	oldItemMap["oldDepreciationTotal"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	newItemMap["payTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalRemain),
	}

	oldItemMap["payTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	newItemMap["interestTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(interestTotalRemain),
	}

	oldItemMap["interestTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}

	attachItems = append(attachItems, &item.AttachItems{
		Items:       oldItemMap,
		DatastoreId: p.DsMap["rireki"],
	})

	attachItems = append(attachItems, &item.AttachItems{
		Items:       newItemMap,
		DatastoreId: p.DsMap["rireki"],
	})

	result.TemplateID = templateID
	result.attachItems = attachItems

	return result, nil
}

// debtCompute 计算债务变更(租赁系统用)
func debtCompute(db, appID, userID string, kiSyuBoka float64, opayData []Payment, oleaseData []Lease, orepayData []RePayment, p DebtParam) (result *DebtResult, err error) {

	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &DebtResult{}

	var attachItems attachData

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
	// 预定解约的场合,支付数据检查
	if p.Kaiyakuymd != "" {
		checkErrK := kaiyakuPayDataCheck(p.HandleMonth, p.Kaiyakuymd, opayData, p.Payments)
		if checkErrK != nil {
			loggerx.ErrorLog("debtCompute", checkErrK.Error())
			return nil, checkErrK
		}
	}

	// 処理月度の翌月からリース料変更可能 and 変更年月の翌月から変動リース料編集可能
	isHasErr := payChangeableCheck(p.Henkouymd[0:7], p.HandleMonth, opayData, p.Payments)
	if isHasErr != nil {
		loggerx.ErrorLog("debtCompute", isHasErr.Error())
		return nil, isHasErr
	}
	// 处理月度转换
	syoriym, err := time.Parse("2006-01", p.HandleMonth)
	if err != nil {
		loggerx.ErrorLog("debtCompute", err.Error())
		return nil, err
	}
	// 割引率
	rishiritsu := p.Rishiritsu
	// 解約行使権オプション
	// cancellationrightoption := p.CancellationRightOption
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
	var leaseData []Lease
	var repayData []RePayment
	// 债务变更调整区调整前数据
	var repayDataAdjBefore []RePayment
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
	var henkouPayments []Payment
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
		leaseTotalRemain = result.Leasesaimusagaku
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
	result.KiSyuBoka = kiSyuBoka + result.Shisannsagaku

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
	repays, err := getRepayData(leasestsyoymd, genkakikan, p.ResidualValue, boka, appID, p.BeginMonth, db)
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
	var repayDataAdjAfter []RePayment
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
		var repay RePayment
		// 调整年月
		repay.Syokyakuymd = p.HandleMonth + "-01"
		// 調整額
		repay.Syokyaku = syokyaku
		// 调整区分
		repay.Syokyakukbn = "調整"
		// 添加偿却调整数据
		repayData = append(repayData, repay)
	}

	// 添加调整区后数据
	repayData = append(repayData, repayDataAdjAfter...)

	// 履历情报
	oldItemMap := make(map[string]*item.Value)
	newItemMap := make(map[string]*item.Value)
	itMap := make(map[string]interface{})

	for key, value := range p.Item.GetItems() {
		oldItemMap[key] = value
		newItemMap[key] = value
		itMap[key] = value
	}

	for key, value := range p.Change {
		newItemMap[key] = value
		itMap[key] = value
	}

	// 获取共通字段
	keiyakuItem := p.Item.Items["keiyakuno"].Value

	// 获取会社信息
	// kaisyaItem := p.Item.Items["leasekaishacd"].Value

	// 获取分類コード信息
	bunruicdItem := p.Item.Items["bunruicd"].Value

	// 获取管理部門
	segmentcdItem := p.Item.Items["segmentcd"].Value

	// 支付情报
	for _, pay := range p.Payments {
		items := make(map[string]*item.Value)
		// 插入共通数据
		items["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    keiyakuItem,
		}
		// items["leasekaishacd"] = &item.Value{
		// 	DataType: "lookup",
		// 	Value:    kaisyaItem,
		// }
		items["bunruicd"] = &item.Value{
			DataType: "lookup",
			Value:    bunruicdItem,
		}
		items["segmentcd"] = &item.Value{
			DataType: "lookup",
			Value:    segmentcdItem,
		}

		items["paymentcount"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentcount),
		}
		items["paymentType"] = &item.Value{
			DataType: "text",
			Value:    pay.PaymentType,
		}
		items["paymentymd"] = &item.Value{
			DataType: "date",
			Value:    pay.Paymentymd,
		}
		items["paymentleasefee"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentleasefee),
		}
		items["paymentleasefeehendo"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentleasefeehendo),
		}
		items["incentives"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Incentives),
		}
		items["sonotafee"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Sonotafee),
		}
		items["kaiyakuson"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Kaiyakuson),
		}

		data := item.AttachItems{
			Items:       items,
			DatastoreId: p.DsMap["paymentStatus"],
		}

		attachItems = append(attachItems, &data)
	}

	// 利息情报
	for _, lease := range leaseData {

		items := make(map[string]*item.Value)
		// 插入共通数据
		items["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    string(keiyakuItem),
		}
		// items["leasekaishacd"] = &item.Value{
		// 	DataType: "lookup",
		// 	Value:    kaisyaItem,
		// }
		items["bunruicd"] = &item.Value{
			DataType: "lookup",
			Value:    bunruicdItem,
		}
		items["segmentcd"] = &item.Value{
			DataType: "lookup",
			Value:    segmentcdItem,
		}

		items["interest"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Interest),
		}
		items["repayment"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Repayment),
		}
		items["balance"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Balance),
		}
		items["present"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Present),
		}
		items["paymentymd"] = &item.Value{
			DataType: "date",
			Value:    lease.Paymentymd,
		}

		data := item.AttachItems{
			Items:       items,
			DatastoreId: p.DsMap["paymentInterest"],
		}

		attachItems = append(attachItems, &data)
	}

	// 偿还情报
	for _, rp := range repayData {

		items := make(map[string]*item.Value)
		// 插入共通数据
		items["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    string(keiyakuItem),
		}
		// items["leasekaishacd"] = &item.Value{
		// 	DataType: "lookup",
		// 	Value:    kaisyaItem,
		// }
		items["bunruicd"] = &item.Value{
			DataType: "lookup",
			Value:    bunruicdItem,
		}
		items["segmentcd"] = &item.Value{
			DataType: "lookup",
			Value:    segmentcdItem,
		}

		items["endboka"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(rp.Endboka),
		}
		items["boka"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(rp.Boka),
		}
		items["syokyaku"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(rp.Syokyaku),
		}
		items["syokyakuymd"] = &item.Value{
			DataType: "date",
			Value:    rp.Syokyakuymd,
		}
		items["syokyakukbn"] = &item.Value{
			DataType: "text",
			Value:    rp.Syokyakukbn,
		}

		data := item.AttachItems{
			Items:       items,
			DatastoreId: p.DsMap["repayment"],
		}

		attachItems = append(attachItems, &data)
	}

	// 契約履歴番号
	rirekiNo := p.seq
	newItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	oldItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	// 修正区分编辑
	newItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "after",
	}
	oldItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "before",
	}
	// 对接区分编辑
	newItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	oldItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	// 操作区分编辑
	newItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "debtchange",
	}
	oldItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "debtchange",
	}

	// 插入共通数据
	newItemMap["keiyakuno"] = &item.Value{
		DataType: "lookup",
		Value:    string(keiyakuItem),
	}
	// newItemMap["leasekaishacd"] = &item.Value{
	// 	DataType: "lookup",
	// 	Value:    kaisyaItem,
	// }
	newItemMap["bunruicd"] = &item.Value{
		DataType: "lookup",
		Value:    bunruicdItem,
	}
	newItemMap["segmentcd"] = &item.Value{
		DataType: "lookup",
		Value:    segmentcdItem,
	}
	oldItemMap["keiyakuno"] = &item.Value{
		DataType: "lookup",
		Value:    string(keiyakuItem),
	}
	// oldItemMap["leasekaishacd"] = &item.Value{
	// 	DataType: "lookup",
	// 	Value:    kaisyaItem,
	// }
	oldItemMap["bunruicd"] = &item.Value{
		DataType: "lookup",
		Value:    bunruicdItem,
	}
	oldItemMap["segmentcd"] = &item.Value{
		DataType: "lookup",
		Value:    segmentcdItem,
	}

	newItemMap["shisannsougaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(result.Shisannsougaku),
	}
	oldItemMap["o_shisannsougaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(result.OShisannsougaku),
	}
	newItemMap["leasesaimusougaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(result.Leasesaimusougaku),
	}
	oldItemMap["o_leasesaimusougaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(result.OLeasesaimusougaku),
	}
	newItemMap["shisannsagaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(result.Shisannsagaku),
	}
	oldItemMap["shisannsagaku"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	newItemMap["leasesaimusagaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(result.Leasesaimusagaku),
	}
	oldItemMap["leasesaimusagaku"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	newItemMap["sonnekigaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(result.Sonnekigaku),
	}
	oldItemMap["sonnekigaku"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	// 分录使用
	// 変更時点の支払残額に対して、比例減少した金額
	newItemMap["gensyoPayTotal"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(gensyoPayTotal),
	}
	oldItemMap["gensyoPayTotal"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(gensyoPayTotal),
	}
	// 変更時点の元本残高に対して、比例減少した金額
	newItemMap["gensyoBalance"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(gensyoBalance),
	}
	oldItemMap["gensyoBalance"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(gensyoBalance),
	}
	// 変更時点の帳簿価額に対して、比例減少した金額
	newItemMap["gensyoBoka"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(gensyoBoka),
	}
	oldItemMap["gensyoBoka"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(gensyoBoka),
	}
	// 再見積変更後現在価値
	newItemMap["leaseTotalAfter"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(leaseTotalAfter),
	}
	oldItemMap["leaseTotalAfter"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(leaseTotalAfter),
	}
	// 変更時点の元本残高に対して、比例残の金額
	newItemMap["leaseTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(leaseTotalRemain),
	}
	oldItemMap["leaseTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(leaseTotalRemain),
	}
	// 再見積変更後の支払総額
	newItemMap["payTotalAfter"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalAfter),
	}
	oldItemMap["payTotalAfter"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalAfter),
	}
	// 変更時点の支払残額に対して、比例残の金額
	newItemMap["payTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalRemain),
	}
	oldItemMap["payTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalRemain),
	}
	// 支付变动额
	newItemMap["payTotalChange"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalChange),
	}
	oldItemMap["payTotalChange"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalChange),
	}

	attachItems = append(attachItems, &item.AttachItems{
		Items:       newItemMap,
		DatastoreId: p.DsMap["rireki"],
	})
	attachItems = append(attachItems, &item.AttachItems{
		Items:       oldItemMap,
		DatastoreId: p.DsMap["rireki"],
	})

	result.TemplateID = templateID
	result.attachItems = attachItems

	return result, nil
}

// cancelCompute 中途解约处理(租赁系统用)
func cancelCompute(db, appID, userID string, opayData []Payment, oleaseData []Lease, orepayData []RePayment, p CancelParam) (result *CancelResult, err error) {

	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()

	result = &CancelResult{}

	var attachItems attachData

	// 解约后数据
	var payData []Payment
	var leaseData []Lease
	var repayData []RePayment
	// 解约年月日
	kaiyakuymd := p.Kaiyakuymd
	// 解约年月转换
	kaiyakuym, err := time.Parse("2006-01", kaiyakuymd[0:7])
	if err != nil {
		loggerx.ErrorLog("cancelCompute", err.Error())
		return nil, err
	}

	// 处理月度转换
	syoriym, err := time.Parse("2006-01", p.HandleMonth)
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
		// 到处理月度为止的偿还费用合计
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
		var re RePayment
		re.Syokyakuymd = p.HandleMonth + "-01"
		re.Syokyaku = 0 - tyoseigaku
		re.Syokyakukbn = "調整"
		repayData = append(repayData, re)
	}

	// 返回结果编辑
	result.RemainDebt = remainDebt
	result.Lossgaku = lossgaku

	// 履历情报
	oldItemMap := make(map[string]*item.Value)
	newItemMap := make(map[string]*item.Value)
	itMap := make(map[string]interface{})

	for key, value := range p.Item.GetItems() {
		oldItemMap[key] = value
		newItemMap[key] = value
		itMap[key] = value
	}

	for key, value := range p.Change {
		newItemMap[key] = value
		itMap[key] = value
	}

	// 获取共通字段
	keiyakuItem := p.Item.Items["keiyakuno"].Value

	// 获取会社信息
	kaisyaItem := p.Item.Items["leasekaishacd"].Value

	// 获取分類コード信息
	bunruicdItem := p.Item.Items["bunruicd"].Value

	// 获取管理部門
	segmentcdItem := p.Item.Items["segmentcd"].Value

	// 支付情报
	for _, pay := range payData {
		items := make(map[string]*item.Value)
		// 插入共通数据
		items["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    keiyakuItem,
		}
		items["leasekaishacd"] = &item.Value{
			DataType: "lookup",
			Value:    kaisyaItem,
		}
		items["bunruicd"] = &item.Value{
			DataType: "lookup",
			Value:    bunruicdItem,
		}
		items["segmentcd"] = &item.Value{
			DataType: "lookup",
			Value:    segmentcdItem,
		}

		items["paymentcount"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentcount),
		}
		items["paymentType"] = &item.Value{
			DataType: "text",
			Value:    pay.PaymentType,
		}
		items["paymentymd"] = &item.Value{
			DataType: "date",
			Value:    pay.Paymentymd,
		}
		items["paymentleasefee"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentleasefee),
		}
		items["paymentleasefeehendo"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Paymentleasefeehendo),
		}
		items["incentives"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Incentives),
		}
		items["sonotafee"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Sonotafee),
		}
		items["kaiyakuson"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(pay.Kaiyakuson),
		}

		payItem := item.AttachItems{
			Items:       items,
			DatastoreId: p.DsMap["paymentStatus"],
		}

		attachItems = append(attachItems, &payItem)
	}

	// 利息情报
	for _, lease := range leaseData {
		items := make(map[string]*item.Value)
		// 插入共通数据
		items["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    string(keiyakuItem),
		}
		items["leasekaishacd"] = &item.Value{
			DataType: "lookup",
			Value:    kaisyaItem,
		}
		items["bunruicd"] = &item.Value{
			DataType: "lookup",
			Value:    bunruicdItem,
		}
		items["segmentcd"] = &item.Value{
			DataType: "lookup",
			Value:    segmentcdItem,
		}

		items["interest"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Interest),
		}
		items["repayment"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Repayment),
		}
		items["balance"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Balance),
		}
		items["present"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(lease.Present),
		}
		items["paymentymd"] = &item.Value{
			DataType: "date",
			Value:    lease.Paymentymd,
		}

		data := item.AttachItems{
			Items:       items,
			DatastoreId: p.DsMap["paymentInterest"],
		}

		attachItems = append(attachItems, &data)
	}

	// 偿还情报
	for _, rp := range repayData {
		items := make(map[string]*item.Value)
		// 插入共通数据
		items["keiyakuno"] = &item.Value{
			DataType: "lookup",
			Value:    string(keiyakuItem),
		}
		items["leasekaishacd"] = &item.Value{
			DataType: "lookup",
			Value:    kaisyaItem,
		}
		items["bunruicd"] = &item.Value{
			DataType: "lookup",
			Value:    bunruicdItem,
		}
		items["segmentcd"] = &item.Value{
			DataType: "lookup",
			Value:    segmentcdItem,
		}

		items["endboka"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(rp.Endboka),
		}
		items["boka"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(rp.Boka),
		}
		items["syokyaku"] = &item.Value{
			DataType: "number",
			Value:    cast.ToString(rp.Syokyaku),
		}
		items["syokyakuymd"] = &item.Value{
			DataType: "date",
			Value:    rp.Syokyakuymd,
		}
		items["syokyakukbn"] = &item.Value{
			DataType: "text",
			Value:    rp.Syokyakukbn,
		}

		data := item.AttachItems{
			Items:       items,
			DatastoreId: p.DsMap["repayment"],
		}

		attachItems = append(attachItems, &data)
	}

	// 履历情报
	// 契約履歴番号
	rirekiNo := p.seq
	newItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	oldItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	// 修正区分编辑
	newItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "after",
	}
	oldItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "before",
	}
	// 对接区分编辑
	newItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	oldItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	// 操作区分编辑
	newItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "midcancel",
	}
	oldItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "midcancel",
	}
	// 解約時元本残高
	newItemMap["remaindebt"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(remainDebt),
	}
	// 中途解約による除却損金额
	newItemMap["lossgaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(lossgaku),
	}
	// 解约年月日
	newItemMap["kaiyakuymd"] = &item.Value{
		DataType: "text",
		Value:    kaiyakuymd,
	}
	// 中途解約時点の償却費の累計額
	newItemMap["syokyakuTotal"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(syokyakuTotal),
	}
	// 使用権資産の原始計上額
	newItemMap["kisyuboka"] = &item.Value{
		DataType: "number",
		Value:    p.Item.Items["kisyuboka"].Value,
	}
	// 中途解約時点の支払リース料残額
	newItemMap["payTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(payTotalRemain),
	}
	// 中途解約時点の利息残
	newItemMap["interestTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(interestTotalRemain),
	}
	// 解約時元本残高
	oldItemMap["remaindebt"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	// 中途解約による除却損金额
	oldItemMap["lossgaku"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	// 解约年月日
	oldItemMap["kaiyakuymd"] = &item.Value{
		DataType: "text",
		Value:    kaiyakuymd,
	}
	// 中途解約時点の償却費の累計額
	oldItemMap["syokyakuTotal"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	// 使用権資産の原始計上額
	oldItemMap["kisyuboka"] = &item.Value{
		DataType: "number",
		Value:    p.Item.Items["kisyuboka"].Value,
	}
	// 中途解約時点の支払リース料残額
	oldItemMap["payTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}
	// 中途解約時点の利息残
	oldItemMap["interestTotalRemain"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}

	attachItems = append(attachItems, &item.AttachItems{
		Items:       newItemMap,
		DatastoreId: p.DsMap["rireki"],
	})
	attachItems = append(attachItems, &item.AttachItems{
		Items:       oldItemMap,
		DatastoreId: p.DsMap["rireki"],
	})

	result.TemplateID = templateID
	result.attachItems = attachItems

	return result, nil
}

// expireCompute 满了处理(租赁系统用)
func expireCompute(db, appID, userID string, orepayData []RePayment, p ExpireParam) (result *ExpireResult, err error) {
	// 返回数据
	result = &ExpireResult{}
	// 生成临时数据ID
	uid := uuid.Must(uuid.NewRandom())
	templateID := uid.String()
	var attachItems attachData

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
	var repayData []RePayment
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

	// 履历情报
	oldItemMap := make(map[string]*item.Value)
	newItemMap := make(map[string]*item.Value)
	itMap := make(map[string]interface{})

	for key, value := range p.Item.GetItems() {
		oldItemMap[key] = value
		newItemMap[key] = value
		itMap[key] = value
	}

	for key, value := range p.Change {
		newItemMap[key] = value
		itMap[key] = value
	}

	// 获取共通字段
	keiyakuItem := p.Item.Items["keiyakuno"].Value

	// 获取会社信息
	kaisyaItem := p.Item.Items["leasekaishacd"].Value

	// 获取分類コード信息
	bunruicdItem := p.Item.Items["bunruicd"].Value

	// 获取管理部門
	segmentcdItem := p.Item.Items["segmentcd"].Value

	// 偿还情报
	if leftgaku != 0 {
		for _, rp := range repayData {
			items := make(map[string]*item.Value)
			// 插入共通数据
			items["keiyakuno"] = &item.Value{
				DataType: "lookup",
				Value:    keiyakuItem,
			}
			items["leasekaishacd"] = &item.Value{
				DataType: "lookup",
				Value:    kaisyaItem,
			}
			items["bunruicd"] = &item.Value{
				DataType: "lookup",
				Value:    bunruicdItem,
			}
			items["segmentcd"] = &item.Value{
				DataType: "lookup",
				Value:    segmentcdItem,
			}

			items["endboka"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Endboka),
			}
			items["boka"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Boka),
			}
			items["syokyaku"] = &item.Value{
				DataType: "number",
				Value:    cast.ToString(rp.Syokyaku),
			}
			items["syokyakuymd"] = &item.Value{
				DataType: "date",
				Value:    rp.Syokyakuymd,
			}
			items["syokyakukbn"] = &item.Value{
				DataType: "text",
				Value:    rp.Syokyakukbn,
			}

			data := item.AttachItems{
				Items:       items,
				DatastoreId: p.DsMap["repayment"],
			}

			attachItems = append(attachItems, &data)
		}

	}

	// 履历情报
	// 契約履歴番号
	rirekiNo := p.seq
	newItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	oldItemMap["no"] = &item.Value{
		DataType: "text",
		Value:    rirekiNo,
	}
	// 修正区分编辑
	newItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "after",
	}
	oldItemMap["zengokbn"] = &item.Value{
		DataType: "options",
		Value:    "before",
	}
	// 对接区分编辑
	newItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	oldItemMap["dockkbn"] = &item.Value{
		DataType: "options",
		Value:    "undo",
	}
	// 操作区分编辑
	newItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "expire",
	}
	oldItemMap["actkbn"] = &item.Value{
		DataType: "options",
		Value:    "expire",
	}
	// 满了后剩余价值
	newItemMap["lossgaku"] = &item.Value{
		DataType: "number",
		Value:    cast.ToString(leftgaku),
	}
	oldItemMap["lossgaku"] = &item.Value{
		DataType: "number",
		Value:    "0",
	}

	attachItems = append(attachItems, &item.AttachItems{
		Items:       newItemMap,
		DatastoreId: p.DsMap["rireki"],
	})
	attachItems = append(attachItems, &item.AttachItems{
		Items:       oldItemMap,
		DatastoreId: p.DsMap["rireki"],
	})

	result.TemplateID = templateID
	result.attachItems = attachItems

	return result, nil
}

// 比較開始時点から計算
func getLeaseDebt(payments []Payment, rishiritsu float64, firstMonth string, residualValue float64) (leaseTotalPayment float64, presentTotalRemain float64) {
	var k float64 = 0
	// 比較開始期首月取得
	payFirstMonth, _ := time.Parse("2006-1", firstMonth)
	for _, pay := range payments {
		// 支付年月日取得
		payymd, _ := time.Parse("2006-01-02", pay.Paymentymd)
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

// 读取支付数据
func buildPayData(path string) (map[string]PayData, error) {

	fileData := make(map[string]PayData)

	// 读取文件
	fs, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer fs.Close()

	r := csv.NewReader(fs)
	r.LazyQuotes = true

	//针对大文件，一行一行的读取文件
	index := 0
	indexMap := make(map[string]int)
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		// 验证行数据是否只包含逗号，只有逗号的行不合法
		isValid, errmsg := filex.CheckRowDataValid(row, index)
		if !isValid {
			return fileData, errors.New(errmsg)
		}

		if index == 1 {
			hasEmpty := false
			for i, h := range row {
				if h == "" || h == "　" {
					hasEmpty = true
				}
				indexMap[h] = i
			}
			if hasEmpty {
				loggerx.ErrorLog("buildPayData", "csvヘッダー行に空白の列名があります。修正してください。")
				return nil, errors.New("csvヘッダー行に空白の列名があります。修正してください。")
			}
		}

		if index > 1 {
			keiyakuno := row[indexMap["keiyakuno"]]
			paymentType := row[indexMap["paymentType"]]
			if paymentType == "1" {
				paymentType = "支払"
			} else if paymentType == "2" {
				paymentType = "残価保証額"
			} else {
				paymentType = " "
			}
			// else if paymentType == "3" {
			// 	paymentType = "購入オプション行使価額"
			// }
			paymentymd := getTrueData(row[indexMap["paymentymd"]])
			paymentcount, err := cast.ToIntE(row[indexMap["paymentcount"]])
			if err != nil {
				loggerx.ErrorLog("buildPayData", err.Error())
				return nil, err
			}
			paymentleasefee, err := cast.ToFloat64E(row[indexMap["paymentleasefee"]])
			if err != nil {
				loggerx.ErrorLog("buildPayData", err.Error())
				return nil, err
			}
			paymentleasefeehendo, err := cast.ToFloat64E(row[indexMap["paymentleasefeehendo"]])
			if err != nil {
				loggerx.ErrorLog("buildPayData", err.Error())
				return nil, err
			}
			incentives, err := cast.ToFloat64E(row[indexMap["incentives"]])
			if err != nil {
				loggerx.ErrorLog("buildPayData", err.Error())
				return nil, err
			}
			sonotafee, err := cast.ToFloat64E(row[indexMap["sonotafee"]])
			if err != nil {
				loggerx.ErrorLog("buildPayData", err.Error())
				return nil, err
			}
			// kaiyakuson, err := cast.ToFloat64E(row[indexMap["kaiyakuson"]])
			// if err != nil {
			// 	loggerx.ErrorLog("buildPayData", err.Error())
			// 	return nil, err
			// }
			pay := Payment{
				PaymentType:          paymentType,
				Paymentymd:           paymentymd,
				Paymentleasefee:      paymentleasefee,
				Paymentleasefeehendo: paymentleasefeehendo,
				Incentives:           incentives,
				Sonotafee:            sonotafee,
				Kaiyakuson:           0, //默认损失为0
				Paymentcount:         paymentcount,
				Keiyakuno:            keiyakuno,
			}
			fileData[keiyakuno] = append(fileData[keiyakuno], pay)
		}
		index++
	}

	for _, list := range fileData {
		sort.Sort(list)
	}

	return fileData, nil
}

// 现在价值累计算出
func getLeaseTotal(payments []Payment, leasestymd time.Time, rishiritsu float64) (presentTotal, leaseTotal float64) {
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

// 支付情报整理
func getArrangedPays(oldPays []Payment) (pays []Payment, err error) {
	var payments []Payment
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
				payInfo := Payment{
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
				payInfo := Payment{
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
			payInfo := Payment{
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
func getLeaseData(payments []Payment, leasestymd time.Time, prevymd time.Time, principalAmount float64, rishiritsu float64) (ls []Lease, err error) {
	var k float64 = 0
	var leaseData []Lease
	// 前回支付日保存用
	var prevPaymentymd time.Time
	// 循环整理生成的新支付情报生成利息相关情报
	for i, pay := range payments {
		var lease Lease
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
func getLeaseDataStart(payments []Payment, leasestymd time.Time, prevymd time.Time, principalAmount float64, rishiritsu float64) (ls []Lease, err error) {
	var k float64 = 0
	var tf = true
	var firstbalance float64 = 0
	var leaseData []Lease
	// 前回支付日保存用
	var prevPaymentymd time.Time
	// 循环整理生成的新支付情报生成利息相关情报
	for i, pay := range payments {
		var lease Lease
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
func getRepayData(leasestsyoymd time.Time, genkakikan int, residualValue float64, boka float64, appID, kishuYm, db string) (rps []RePayment, err error) {

	var repayData []RePayment
	// 默认期首月为1月
	var kishuMonth int64 = 1
	// 期首月取得
	kishuMonth = cast.ToInt64(kishuYm)
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
			var repay RePayment
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
func getRepayDataStart(firstMonth time.Time, genkakikan int, residualValue float64, boka float64, appID string, db string, leasestymd time.Time) (rps []RePayment, err error) {

	var repayData []RePayment
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
			var repay RePayment
			// 期首薄价
			repay.Boka = boka
			// 偿却年月
			repay.Syokyakuymd = firstMonth.Format("2006-01-02")
			// 偿却区分
			repay.Syokyakukbn = "通常"
			//syokyakuymd, _ := time.Parse("2006-01-02", repay.Syokyakuymd)
			//if !(syokyakuymd.Before(leasestsyoymd)) {
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
			//}
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
func getRepayDataObtain(leasestsyoymd time.Time, genkakikan int, residualValue float64, boka float64, appID string, db string, firstMonthB time.Time) (rps []RePayment, err error) {

	var repayData []RePayment
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
			var repay RePayment
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

// generatePay 生成支付数据(租赁系统用)
func generatePay(q PayParam) (payData []Payment, err error) {
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
	// 初回リース料
	firstleasefee := q.Firstleasefee
	// 最終回リース料
	finalleasefee := q.Finalleasefee
	//中转二回リース料
	var aaa = paymentleasefee
	//契约番号
	keiyakuno := q.Keiyakuno

	// 循环支付回数生成结果
	for i := 0; i < paymentcounts; i++ {
		var pay Payment
		if i == 0 {
			if firstleasefee != 0 {
				paymentleasefee = firstleasefee
			}
		}
		if i == 1 {
			paymentleasefee = aaa
		}
		if i == paymentcounts-1 {
			if finalleasefee != 0 {
				paymentleasefee = finalleasefee
			}
		}
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
		pay.Keiyakuno = keiyakuno
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
		var pay Payment
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
		var pay Payment
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

// getExpireymd 计算租赁满了日
func getExpireymd(leasestymd string, leasekikan, extentionOption int) (value string, err error) {
	var expireymd time.Time
	// 租赁开始日转换
	stymd, err := time.Parse("2006-01-02", leasestymd)
	if err != nil {
		return "", err
	}

	// 租赁满了日算出
	expireymd = timeconv.AddDate(stymd, 0, leasekikan+extentionOption, 0)

	return expireymd.Format("2006-01-02"), nil
}
