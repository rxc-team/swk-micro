package lease

import (
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
)

func getTrueData(value string) (data string) {

	if len(value) == 0 {
		return value
	}
	if len(value) == 8 {
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return value[0:4] + "-" + value[4:6] + "-" + value[6:8]
		}
	}
	if len(value) > 7 && len(value) < 11 {
		var ymdArr []string
		isRightDelimiter := true
		if strings.Contains(value, "-") {
			ymdArr = strings.Split(value, "-")
		} else if strings.Contains(value, "/") {
			ymdArr = strings.Split(value, "/")
		} else if strings.Contains(value, ".") {
			ymdArr = strings.Split(value, ".")
		} else {
			isRightDelimiter = false
		}

		// 合法年月日分隔符
		if isRightDelimiter {
			// 分割成年月日三份
			if len(ymdArr) == 3 {
				strY := ymdArr[0]
				strM := ymdArr[1]
				strD := ymdArr[2]
				// 年四位、月1~2位、日1~2位
				if len(strY) == 4 && len(strM) < 3 && len(strD) < 3 {
					// 月一位补位
					if len(strM) == 1 {
						strM = "0" + strM
					}
					// 日一位补位
					if len(strD) == 1 {
						strD = "0" + strD
					}
					return strY + "-" + strM + "-" + strD
				}
			}
		}
	}
	return " "
}

// getGapMonths  获取二个日期间隔月数
func getGapMonths(startymd time.Time, endymd time.Time) (months int) {
	// 开始年
	startymdyear := startymd.Year()
	// 开始月
	startymdmonth := cast.ToInt(startymd.Format("01"))
	// startymdmonth := stringx.StringToInt(startymd.Format("01"))
	// 结束年
	endymdyear := endymd.Year()
	// 结束月
	endymdmonth := cast.ToInt(endymd.Format("01"))
	// endymdmonth := stringx.StringToInt(endymd.Format("01"))

	if startymd.Format("01") == "08" {
		startymdmonth = 8
	}
	if startymd.Format("01") == "09" {
		startymdmonth = 9
	}

	if endymd.Format("01") == "08" {
		endymdmonth = 8
	}
	if endymd.Format("01") == "09" {
		endymdmonth = 9
	}
	return endymdyear*12 + endymdmonth - startymdyear*12 - startymdmonth
}

// getPayDate  获取支付日
func getPayDate(date time.Time, payday int) (pdate time.Time) {
	// 年月日取得
	years := date.Year()
	month := cast.ToInt(date.Format("01"))
	if date.Format("01") == "08" {
		month = 8
	}
	if date.Format("01") == "09" {
		month = 9
	}
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
