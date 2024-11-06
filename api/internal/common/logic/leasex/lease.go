package leasex

import (
	"time"

	"rxcsoft.cn/pit3/api/internal/common/logic/configx"
	"rxcsoft.cn/pit3/api/internal/common/stringx"
	"rxcsoft.cn/pit3/api/internal/common/typesx"
)

// getGapMonths  获取二个日期间隔月数
func getGapMonths(startymd time.Time, endymd time.Time) (months int) {
	// 开始年
	startymdyear := startymd.Year()
	// 开始月
	startymdmonth := stringx.StringToInt(startymd.Format("01"))
	// 结束年
	endymdyear := endymd.Year()
	// 结束月
	endymdmonth := stringx.StringToInt(endymd.Format("01"))

	return endymdyear*12 + endymdmonth - startymdyear*12 - startymdmonth
}

// getPayDate  获取支付日
func getPayDate(date time.Time, payday int) (pdate time.Time) {
	// 年月日取得
	years := date.Year()
	month := stringx.StringToInt(date.Format("01"))
	nowday := date.Day()

	// 月末日取得
	lastday := 0
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			lastday = 30
		} else {
			lastday = 31
		}
	} else {
		if ((years%4) == 0 && (years%100) != 0) || (years%400) == 0 {
			lastday = 29
		} else {
			lastday = 28
		}
	}

	// 获取支付日
	if payday > lastday {
		pdate = date.AddDate(0, 0, lastday-nowday)
	} else {
		pdate = date.AddDate(0, 0, payday-nowday)
	}

	return pdate
}

// 短期リースまたは少額リース判定
func ShortOrMinorJudge(db, currentAppID string, leasekikan, extentionOption int, payments []typesx.Payment) (t string) {
	var leaseType string = "normal_lease"
	// 少額标准取得
	cfg, err := configx.GetConfigVal(db, currentAppID)
	if err != nil {
		return leaseType
	}
	minor := cfg.GetMinorBaseAmount()
	// 支付总额取得
	var payTotal float64 = 0
	for _, pay := range payments {
		payTotal += pay.Paymentleasefee
	}
	// 少額判断
	if payTotal <= float64(stringx.StringToInt(minor)) {
		leaseType = "minor_lease"
	}
	// 短期标准取得
	short := cfg.GetShortLeases()
	// 租赁总期间 = 租赁期间 + 延长租赁期间
	leasekikanTotal := leasekikan + extentionOption
	// 短期判断
	if leasekikanTotal <= stringx.StringToInt(short) {
		leaseType = "short_lease"
	}

	return leaseType
}
