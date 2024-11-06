package typesx

import (
	"time"

	"rxcsoft.cn/pit3/srv/database/proto/template"
)

type TplData []*template.ListItems

// InsertResult 少额预算返回
type InsertResult struct {
	TemplateID string  `json:"template_id" bson:"template_id"`
	TplItems   TplData `json:"-"`
}

// ComputeResult 新规契约预算返回
type ComputeResult struct {
	TemplateID  string  `json:"template_id" bson:"template_id"`
	KiSyuBoka   float64 `json:"kisyuboka" bson:"kisyuboka"` // 原始取得价值
	TplItems    TplData `json:"-"`
	Hkkjitenzan float64 `json:"hkkjitenzan" bson:"hkkjitenzan"` // 比較開始時点の残存リース料
	Sonnekigaku float64 `json:"sonnekigaku" bson:"sonnekigaku"` // 利益剰余金
}

// ChangeResult 契约情报变更返回
type ChangeResult struct {
	TemplateID string  `json:"template_id" bson:"template_id"`
	TplItems   TplData `json:"-"`
}

// CancelResult 中途解约预算返回
type CancelResult struct {
	TemplateID string  `json:"template_id" bson:"template_id"` // 临时数据ID
	RemainDebt float64 `json:"remaindebt" bson:"remaindebt"`   // 解約時元本残高
	Lossgaku   float64 `json:"lossgaku" bson:"lossgaku"`       // 中途解約による除却損金额
	TplItems   TplData `json:"-"`
}

// ExpireResult 满了预算返回
type ExpireResult struct {
	TemplateID string  `json:"template_id" bson:"template_id"` // 临时数据ID
	Leftgaku   float64 `json:"leftgaku" bson:"leftgaku"`       // 满了時剩余价值
	TplItems   TplData `json:"-"`
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
	TplItems           TplData `json:"-"`
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
	Payments                []Payment         `json:"payments" bson:"payments"`                               // 支付情报
	DsMap                   map[string]string `json:"ds_map" bson:"ds_map"`                                   // 台账情报
	Sykshisankeisan         string            `json:"sykshisankeisan" bson:"sykshisankeisan"`                 // 使用権資産
	FirstMonth              string            `json:"firstMonth" bson:"firstMonth"`                           // 比較開始期首月
	Hkkjitenzan             float64           `json:"hkkjitenzan" bson:"hkkjitenzan"`                         // 比較開始時点の残存リース料
	Sonnekigaku             float64           `json:"sonnekigaku" bson:"sonnekigaku"`                         // 利益剰余金
}

// ChangeParam 契约情报变更参数
type ChangeParam struct {
	Henkouymd string `json:"henkouymd" bson:"henkouymd"` // 变更年月
	ItemID    string `json:"item_id" bson:"item_id"`     // 契约ID
}

// DebtParam 债务变更情报参数
type DebtParam struct {
	Kaiyakuymd              string            `json:"kaiyakuymd" bson:"kaiyakuymd"`                           // 解约年月
	Henkouymd               string            `json:"henkouymd" bson:"henkouymd"`                             // 变更年月
	Leasestymd              string            `json:"leasestymd" bson:"leasestymd"`                           // 租赁开始日
	CancellationRightOption bool              `json:"cancellationrightoption" bson:"cancellationrightoption"` // 解約行使権オプション
	Leasekikan              int               `json:"leasekikan" bson:"leasekikan"`                           // 租赁期间
	ExtentionOption         int               `json:"extentionOption" bson:"extentionOption"`                 // 延长租赁期间
	Keiyakuno               string            `json:"keiyakuno" bson:"keiyakuno"`                             // 契约番号
	Rishiritsu              float64           `json:"rishiritsu" bson:"rishiritsu"`                           // 割引率
	ResidualValue           float64           `json:"residualValue" bson:"residualValue"`                     // 残价保证额
	Assetlife               int               `json:"assetlife" bson:"assetlife"`                             // 耐用年限
	Torihikikbn             string            `json:"torihikikbn" bson:"torihikikbn"`                         // 取引判定区分
	Percentage              float64           `json:"percentage" bson:"percentage"`                           // 剩余资产百分比
	Payments                []Payment         `json:"payments" bson:"payments"`                               // 支付情报
	DsMap                   map[string]string `json:"ds_map" bson:"ds_map"`                                   // 台账情报
}

// ExpireParam 契约满了情报参数
type ExpireParam struct {
	Henkouymd         string            `json:"henkouymd" bson:"henkouymd"`                 // 变更年月
	Torihikikbn       string            `json:"torihikikbn" bson:"torihikikbn"`             // 取引判定区分
	Expiresyokyakukbn string            `json:"expiresyokyakukbn" bson:"expiresyokyakukbn"` // リース満了償却区分
	Keiyakuno         string            `json:"keiyakuno" bson:"keiyakuno"`                 // 契约番号
	DsMap             map[string]string `json:"ds_map" bson:"ds_map"`                       // 台账情报
}

// CancelParam 中途解约情报参数
type CancelParam struct {
	Kaiyakuymd string            `json:"kaiyakuymd" bson:"kaiyakuymd"` // 解約年月日
	Keiyakuno  string            `json:"keiyakuno" bson:"keiyakuno"`   // 契约番号
	DsMap      map[string]string `json:"ds_map" bson:"ds_map"`         // 台账情报
}

// Payment 支付数据
type Payment struct {
	Leasekaishacd        string  `json:"leasekaishacd" bson:"leasekaishacd"`               // 租赁会社
	Keiyakuno            string  `json:"keiyakuno" bson:"keiyakuno"`                       // 契约番号
	Paymentcount         int     `json:"paymentcount" bson:"paymentcount"`                 // 支付回数
	PaymentType          string  `json:"paymentType" bson:"paymentType"`                   // 支付类型
	Paymentymd           string  `json:"paymentymd" bson:"paymentymd"`                     // 支付年月日
	Paymentleasefee      float64 `json:"paymentleasefee" bson:"paymentleasefee"`           // 支付金额
	Paymentleasefeehendo float64 `json:"paymentleasefeehendo" bson:"paymentleasefeehendo"` // 变更支付金额
	Incentives           float64 `json:"incentives" bson:"incentives"`                     // 优惠金额
	Sonotafee            float64 `json:"sonotafee" bson:"sonotafee"`                       // 其他金额
	Kaiyakuson           float64 `json:"kaiyakuson" bson:"kaiyakuson"`                     // 解约损失
	Fixed                bool    `json:"fixed" bson:"fixed"`                               // 修正否
}

// Lease 利息数据
type Lease struct {
	Leasekaishacd string  `json:"leasekaishacd" bson:"leasekaishacd"` // 租赁会社
	Keiyakuno     string  `json:"keiyakuno" bson:"keiyakuno"`         // 契约番号
	Interest      float64 `json:"interest" bson:"interest"`           // 支付利息相当额
	Repayment     float64 `json:"repayment" bson:"repayment"`         // 元本返済相当額
	Balance       float64 `json:"balance" bson:"balance"`             // 元本残高相当額
	Firstbalance  float64 `json:"firstbalance" bson:"firstbalance"`   // 期首元本残高
	Present       float64 `json:"present" bson:"present"`             // 現在価値
	Paymentymd    string  `json:"paymentymd" bson:"paymentymd"`       // 支付年月
}

// RePayment 偿还数据
type RePayment struct {
	Leasekaishacd string  `json:"leasekaishacd" bson:"leasekaishacd"` // 租赁会社
	Keiyakuno     string  `json:"keiyakuno" bson:"keiyakuno"`         // 契约番号
	Syokyakukbn   string  `json:"syokyakukbn" bson:"syokyakukbn"`     // 償却区分
	Endboka       float64 `json:"endboka" bson:"endboka"`             // 月末薄价
	Boka          float64 `json:"boka" bson:"boka"`                   // 期首薄价
	Syokyaku      float64 `json:"syokyaku" bson:"syokyaku"`           // 偿却额
	Syokyakuymd   string  `json:"syokyakuymd" bson:"syokyakuymd"`     // 偿却年月
}
