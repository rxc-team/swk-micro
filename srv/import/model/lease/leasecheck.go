package lease

import (
	"errors"
	"reflect"
	"time"

	"github.com/spf13/cast"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
)

// leasestCheck 租赁开始日check
func leasestCheck(leasestymd, keiyakuymd string) bool {

	if len(leasestymd) != 0 && len(keiyakuymd) != 0 {
		if leasestymd < keiyakuymd {
			return true
		}
	}

	return false
}

// paystCheck 支付开始日check
func paystCheck(paymentymd, leasestymd string) bool {

	if len(paymentymd) != 0 && len(leasestymd) != 0 {
		if paymentymd < leasestymd {
			return true
		}
	}

	return false
}

// henkouCheck 变更年月日check
func henkouCheck(henkouymd, lastHenkouymd, leasestymd, handleMonth, beginMonth string) bool {

	henkouDate, err := time.Parse("2006-01-02", henkouymd)
	if err != nil {
		return true
	}

	henkouYm := henkouDate.Format("2006-01")

	lastHenkouDate, err := time.Parse("2006-01-02", lastHenkouymd)
	if err != nil {
		return true
	}

	lastHenkouYm := lastHenkouDate.Format("2006-01")

	// 如果变更年月日在租赁开始年月日之前，则无法通过
	if henkouymd < leasestymd {
		return true
	}

	// 当前期首的月份>处理年月的月份时，取上一年的期首日期
	kishuym := ""
	if len(handleMonth) > 0 {
		date, err := time.Parse("2006-01", handleMonth)
		if err != nil {
			return true
		}
		year := date.Year()
		month := date.Month()
		bMonth := cast.ToInt(beginMonth)
		local := time.Now().Location()

		if bMonth > int(month) {
			kishuym = time.Date(year-1, time.Month(bMonth), 1, 0, 0, 0, 0, local).Format("2006-01")
		} else {
			kishuym = time.Date(year, time.Month(bMonth), 1, 0, 0, 0, 0, local).Format("2006-01")
		}
	}
	// 从未变更的状态
	if len(lastHenkouymd) == 0 {
		// 1.当前期首之后~处理年月
		if henkouYm < kishuym {
			return true
		}
		if henkouYm > handleMonth {
			return true
		}
	}

	// 2.已经变更过了的，在变更过后的日期小于期首  当前期首之后~处理年月
	if lastHenkouYm < kishuym {
		if henkouYm < kishuym {
			return true
		}
		if henkouYm > handleMonth {
			return true
		}
		return false
	}
	// 3.在变更过后的日期大于期首 变更过后的日期 ~处理年月
	if lastHenkouYm > kishuym {
		if henkouymd < lastHenkouymd {
			return true
		}
		if henkouYm > handleMonth {
			return true
		}
		return false
	}

	return false
}

// expireDayCheck 满了年月日check
func expireDayCheck(henkouymd, handleMonth, leeseexpireymd string) bool {
	if len(henkouymd) < 10 {
		return true
	}
	ym := henkouymd[:7]
	ymd := henkouymd[:10]
	// 變更年月日(即滿了日) <= 処理月度
	hym := handleMonth[:7]
	if ym > hym {
		return true
	}
	// 變更年月日(即滿了日) >= 満了預定日付
	expireymd := leeseexpireymd[:10]
	return ymd < expireymd
}

// kaiyakuCheck 解约年月日check
func kaiyakuCheck(kaiyakuymd, lastKaiyakuymd, leasestymd, handleMonth, beginMonth string) bool {
	// 解約年月日time
	kaiyakuDate, err := time.Parse("2006-01-02", kaiyakuymd)
	if err != nil {
		return true
	}
	// 解約年月str
	kaiyakuYm := kaiyakuDate.Format("2006-01")
	// 解約年月日str
	kykYmd := kaiyakuDate.Format("2006-01-02")

	// 上次變更年月日time
	lastKaiyakuDate, err := time.Parse("2006-01-02", lastKaiyakuymd)
	if err != nil {
		return true
	}
	// 上次變更年月str
	lastKaiyakuYm := lastKaiyakuDate.Format("2006-01")

	// 租賃開始年月日time
	leasestDate, err := time.Parse("2006-01-02", leasestymd)
	if err != nil {
		return true
	}
	// 租賃開始年月日str
	lstYmd := leasestDate.Format("2006-01-02")

	// 期首年月取得
	kishuym := ""
	if len(handleMonth) > 0 {
		date, err := time.Parse("2006-01", handleMonth)
		if err != nil {
			return true
		}
		year := date.Year()
		month := date.Month()
		bMonth := cast.ToInt(beginMonth)
		local := time.Now().Location()

		if bMonth > int(month) {
			kishuym = time.Date(year-1, time.Month(bMonth), 1, 0, 0, 0, 0, local).Format("2006-01")
		} else {
			kishuym = time.Date(year, time.Month(bMonth), 1, 0, 0, 0, 0, local).Format("2006-01")
		}
	}

	// 解约年月日不能在租赁开始日之前(包括租赁开始日)
	if kykYmd <= lstYmd {
		return true
	}

	// 从未变更的状态
	if len(lastKaiyakuymd) == 0 {
		// 期首之后~处理年月(包括边界值)
		if kaiyakuYm < kishuym {
			return true
		}
		if kaiyakuYm > handleMonth {
			return true
		}
	} else {
		// 已经变更过
		// 变更后日期小于期首----合法解约年月日 = 期首年月之后~处理年月(包括边界值)
		if lastKaiyakuYm < kishuym {
			if kaiyakuYm < kishuym {
				return true
			}
			if kaiyakuYm > handleMonth {
				return true
			}
			return false
		}
		// 变更后日期大于期首----合法解约年月日 = 变更后年月日之后~处理年月(包括边界值)
		if lastKaiyakuYm > kishuym {
			if kykYmd < lastKaiyakuymd {
				return true
			}
			if kaiyakuYm > handleMonth {
				return true
			}
			return false
		}
	}

	return false
}

// 処理月度の翌月からリース料変更可能 and 変更年月の翌月から変動リース料編集可能
func payChangeableCheck(hym, sym string, oldpays []Payment, newpays []Payment) (err error) {
	// 変更年月
	henkouym, err := time.Parse("2006-01", hym)
	if err != nil {
		loggerx.ErrorLog("payChangeableCheck", err.Error())
		return err
	}
	// 处理月度
	syoriym, err := time.Parse("2006-01", sym)
	if err != nil {
		loggerx.ErrorLog("payChangeableCheck", err.Error())
		return err
	}

	// 旧支払データから変更年月前（変更年月も含む）の支払データと処理月度前（処理月度も含む）の支払データの支払リース料を取得
	var oldpayfee []float64
	var oldPrePays []Payment
	for _, pay := range oldpays {
		// 支付年月
		paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("payChangeableCheck", err.Error())
			return err
		}
		// 支払データ
		if paymentym.Before(henkouym) || paymentym.Equal(henkouym) {
			pay.Fixed = true
			oldPrePays = append(oldPrePays, pay)
		}
		// 支払リース料
		if paymentym.Before(syoriym) || paymentym.Equal(syoriym) {
			oldpayfee = append(oldpayfee, pay.Paymentleasefee)
		}
	}

	// 新旧支払データから変更年月前（変更年月も含む）の支払データと処理月度前（処理月度も含む）の支払データの支払リース料を取得
	var newpayfee []float64
	var newPrePays []Payment
	for _, pay := range newpays {
		// 支付年月
		paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("payChangeableCheck", err.Error())
			return err
		}
		// 支払データ
		if paymentym.Before(henkouym) || paymentym.Equal(henkouym) {
			pay.Fixed = true
			newPrePays = append(newPrePays, pay)
		}
		// 支払リース料
		if paymentym.Before(syoriym) || paymentym.Equal(syoriym) {
			newpayfee = append(newpayfee, pay.Paymentleasefee)
		}
	}

	// チェック１：変更年月前（変更年月も含む）の支払データは一切変更出来ない
	if !reflect.DeepEqual(oldPrePays, newPrePays) {
		return errors.New("変更年月前（変更年月も含む）の支払データを編集することは禁じられております")
	}

	// チェック２：処理月度前（処理月度も含む）の支払データの支払リース料は一切変更出来ない
	if !reflect.DeepEqual(oldpayfee, newpayfee) {
		return errors.New("処理月度前（処理月度も含む）の支払データの支払リース料を編集することは禁じられております")
	}

	return nil
}

// 支付数据合法性检查
func payDataValidCheck(cancellationRightOption bool, pays []Payment) (err error) {
	// 前回支付日保存用
	var prevPaymentymd time.Time
	// 前回支付回数保存用
	var prevPaymentCount int
	// 解约损失检查用
	var cancelLostCount float64
	for index, pay := range pays {
		// 支付年月日
		paymentymd, err := time.Parse("2006-01-02", pay.Paymentymd[0:10])
		if err != nil {
			loggerx.ErrorLog("payDataValidCheck", err.Error())
			return err
		}
		if index == 0 {
			// 初轮赋值
			prevPaymentymd = paymentymd
			prevPaymentCount = pay.Paymentcount
		} else {
			// 支付顺序检查
			if paymentymd.Before(prevPaymentymd) || pay.Paymentcount <= prevPaymentCount {
				return errors.New("支払順序チェックエラー")
			}
			// 支付日重复检查
			if paymentymd.Equal(prevPaymentymd) {
				return errors.New("支払日重複チェックエラー")
			}
			// 退避赋值
			prevPaymentymd = paymentymd
			prevPaymentCount = pay.Paymentcount
		}
		// 单条支付总额不能为负Check
		singlePayCount := pay.Paymentleasefee + pay.Paymentleasefeehendo + pay.Incentives + pay.Sonotafee + pay.Kaiyakuson
		if singlePayCount < 0 {
			return errors.New("一回の支払合計がマイナスであるチェックエラー")
		}
		// 单条支付前三个金额不能同时为零Check
		if pay.Paymentleasefee == 0 && pay.Paymentleasefeehendo == 0 && pay.Incentives == 0 {
			return errors.New("1回の支払の最初の3つの金額は同時にゼロであるチェックエラー")
		}

		// 解约损失检查
		if pay.Kaiyakuson < 0 {
			return errors.New("解約損がマイナスであるチェックエラー")
		}
		cancelLostCount += pay.Kaiyakuson
	}

	// 解约损失检查
	if !cancellationRightOption {
		if cancelLostCount != 0 {
			return errors.New("解約行使権オプションがチェックされていない場合、解約損の入力は禁じられています")
		}
	}

	return nil
}

// 残価保証額(IFRS)と購入オプション行使価額不能同时入力值。
func optionExclusiveCheck(residualValue, optionToPurchase string) bool {
	if residualValue == "0" || optionToPurchase == "0" {
		return false
	}
	if len(residualValue) > 0 && len(optionToPurchase) > 0 {
		return true
	}
	return false
}

// 契约状态check
// 契约状态是满了的场合，不能再满了，解约
// 契约状态是解约的场合，不能再解约，满了，情报变更，债务变更
func keiyakuStatusCheck(status string) bool {
	if status == "complete" || status == "cancel" {
		return true
	}
	return false
}

// 契约状态check2
// 契约状态是解约的场合，不能再解约，满了，情报变更，债务变更
func keiyakuStatusCheck2(status string) bool {
	return status == "cancel"
}

// 未满了的场合，不能满了
func expireCheck(leaseexpireymd, handleMondth string) bool {
	if len(leaseexpireymd) < 7 || len(handleMondth) < 7 {
		return true
	}
	if leaseexpireymd[:7] > handleMondth[:7] {
		return true
	}
	return false
}

// 审批的数据，不能满了，解约，情报变更，债务变更（status == 2）
func nonAdmitCheck(status string) bool {
	return status == "2"
}

// 前払リース料とリース・インセンティブ(前払) 不能为负数的check
func prepaidCheck(paymentsAtOrPrior, incentivesAtOrPrior string) bool {
	payments := cast.ToFloat64(paymentsAtOrPrior)
	incentives := cast.ToFloat64(incentivesAtOrPrior)

	if incentives+payments < 0 {
		return true
	}
	return false
}

// 短期リースまたは少額リース判定
func shortOrMinorJudge(db, currentAppID, minor, short string, leasekikan, extentionOption int, payments []Payment) (t string) {
	var leaseType string = "normal_lease"
	// 支付总额取得
	var payTotal float64 = 0
	for _, pay := range payments {
		payTotal += pay.Paymentleasefee
	}
	// 少額判断
	if payTotal <= cast.ToFloat64(minor) {
		leaseType = "minor_lease"
	}
	// 租赁总期间 = 租赁期间 + 延长租赁期间
	leasekikanTotal := leasekikan + extentionOption
	// 短期判断
	if leasekikanTotal <= cast.ToInt(short) {
		leaseType = "short_lease"
	}

	return leaseType
}

// 租赁满了年月日检查
func expireymdCheck(leasestymd, lastpayymd string, leasekikan, extentionOption int) (err error) {
	// 租赁开始日转换
	stymd, err := time.Parse("2006-01-02", leasestymd)
	if err != nil {
		return errors.New("リース開始日付（" + leasestymd + "）が不正です")
	}
	// 最终支付日转换
	payymd, err := time.Parse("2006-01-02", lastpayymd)
	if err != nil {
		return errors.New("最終支払日付（" + lastpayymd + "）が不正です")
	}
	// 租赁满了日算出
	expireymd := stymd.AddDate(0, leasekikan+extentionOption, 0)

	// 判断
	if expireymd.Before(payymd) {
		return errors.New("リース満了年月日（" + expireymd.Format("2006/01/02") + "）は最終支払日（" + lastpayymd + "）以降（最終支払日含め）でなければなりません")
	}

	return nil
}

// MiraiKaiyakuCheck 未来解约年月日check
func miraiKaiyakuCheck(kaiyakuymd, leaseexpireymd, handleMonth string) bool {
	// 处理月度
	handleYM, err := time.Parse("2006-01", handleMonth)
	if err != nil {
		return true
	}
	// 解约年月
	kaiym, err := time.Parse("2006-01", kaiyakuymd[0:7])
	if err != nil {
		return true
	}
	// 解约年月日
	kaiymd, err := time.Parse("2006-01-02", kaiyakuymd)
	if err != nil {
		return true
	}
	// 满了年月日
	leaymd, err := time.Parse("2006-01-02", leaseexpireymd)
	if err != nil {
		return true
	}

	// 处理月度(不含) ~ 满了
	if kaiym.Equal(handleYM) || kaiym.Before(handleYM) {
		return true
	}
	if kaiymd.After(leaymd) {
		return true
	}

	return false
}

func kaiyakuPayDataCheck(syoriYmStr, kaiyakuymd string, oldpays, newpays []Payment) (err error) {
	// 处理月度
	syoriym, err := time.Parse("2006-01", syoriYmStr)
	if err != nil {
		loggerx.ErrorLog("kaiyakuPayDataCheck", err.Error())
		return err
	}
	// 解约年月
	kaiyakuym, err := time.Parse("2006-01", kaiyakuymd[0:7])
	if err != nil {
		loggerx.ErrorLog("kaiyakuPayDataCheck", err.Error())
		return err
	}

	var oldPastPays []Payment
	var oldLeftPays []Payment
	for _, pay := range oldpays {
		// 支付年月
		paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("kaiyakuPayDataCheck", err.Error())
			return err
		}
		// 已经完成支払データ
		if paymentym.Before(syoriym) {
			pay.Fixed = true
			oldPastPays = append(oldPastPays, pay)
		}
		// 待完成支払データ
		if (paymentym.Equal(syoriym) || paymentym.After(syoriym)) && (paymentym.Equal(kaiyakuym) || paymentym.Before(kaiyakuym)) {
			pay.Fixed = true
			pay.Kaiyakuson = 0
			oldLeftPays = append(oldLeftPays, pay)
		}
	}

	var newPastPays []Payment
	var newLeftPays []Payment
	for _, pay := range newpays {
		// 支付年月
		paymentym, err := time.Parse("2006-01", pay.Paymentymd[0:7])
		if err != nil {
			loggerx.ErrorLog("kaiyakuPayDataCheck", err.Error())
			return err
		}
		// 已经完成支払データ
		if paymentym.Before(syoriym) {
			pay.Fixed = true
			newPastPays = append(newPastPays, pay)
		}
		// 待完成支払データ
		if (paymentym.Equal(syoriym) || paymentym.After(syoriym)) && (paymentym.Equal(kaiyakuym) || paymentym.Before(kaiyakuym)) {
			pay.Fixed = true
			pay.Kaiyakuson = 0
			newLeftPays = append(newLeftPays, pay)
		}
	}

	// 処理月度から解約年月までの支払データの解約損項目しか編集できない
	if (!reflect.DeepEqual(oldPastPays, newPastPays)) || (!reflect.DeepEqual(oldLeftPays, newLeftPays)) {
		return errors.New("支払表データに対して、未来解約の場合、処理月度から解約年月までの支払データの解約損項目しか編集できない")
	}

	// 预定解约日后应该无支付数据
	if len(newpays) > len(newPastPays)+len(newLeftPays) {
		return errors.New("支払表データに対して、予定の解約年月以降の支払データが存在しないはずです")
	}

	return nil
}
