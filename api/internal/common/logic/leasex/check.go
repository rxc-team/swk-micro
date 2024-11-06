package leasex

import (
	"errors"
	"reflect"
	"time"

	"rxcsoft.cn/pit3/api/internal/common/typesx"
)

// 処理月度の翌月からリース料変更可能 and 変更年月の翌月から変動リース料編集可能
func payChangeableCheck(hym, sym string, oldpays []typesx.Payment, newpays []typesx.Payment) (err error) {
	// 変更年月
	henkouym, _ := time.Parse("2006-01", hym)
	// 处理月度
	syoriym, _ := time.Parse("2006-01", sym)

	// 旧支払データから変更年月前（変更年月も含む）の支払データと処理月度前（処理月度も含む）の支払データの支払リース料を取得
	var oldpayfee []float64
	var oldPrePays []typesx.Payment
	for _, pay := range oldpays {
		// 支付年月
		paymentym, _ := time.Parse("2006-01", pay.Paymentymd[0:7])
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
	var newPrePays []typesx.Payment
	for _, pay := range newpays {
		// 支付年月
		paymentym, _ := time.Parse("2006-01", pay.Paymentymd[0:7])
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

func kaiyakuPayDataCheck(syoriYmStr, kaiyakuymd string, oldpays, newpays []typesx.Payment) (err error) {
	// 处理月度
	syoriym, _ := time.Parse("2006-01", syoriYmStr)
	// 解约年月
	kaiyakuym, _ := time.Parse("2006-01", kaiyakuymd[0:7])

	var oldPastPays []typesx.Payment
	var oldLeftPays []typesx.Payment
	for _, pay := range oldpays {
		// 支付年月
		paymentym, _ := time.Parse("2006-01", pay.Paymentymd[0:7])
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

	var newPastPays []typesx.Payment
	var newLeftPays []typesx.Payment
	for _, pay := range newpays {
		// 支付年月
		paymentym, _ := time.Parse("2006-01", pay.Paymentymd[0:7])
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

// 支付数据合法性检查
func payDataValidCheck(cancellationRightOption bool, pays []typesx.Payment) (err error) {
	// 前回支付日保存用
	var prevPaymentymd time.Time
	// 前回支付回数保存用
	var prevPaymentCount int
	// 解约损失检查用
	var cancelLostCount float64
	for index, pay := range pays {
		// 支付年月日
		paymentymd, _ := time.Parse("2006-01-02", pay.Paymentymd[0:10])
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

// 租赁满了年月日检查
func expireymdCheck(leasestymd, lastpayymd string, leasekikan, extentionOption int) (err error) {
	// 租赁开始日转换
	stymd, err := time.Parse("2006-01-02", leasestymd)
	if err != nil {
		return errors.New("リース満了年月日は最終支払日以降（最終支払日含め）でなければなりません")
	}
	// 最终支付日转换
	payymd, err := time.Parse("2006-01-02", lastpayymd)
	if err != nil {
		return errors.New("リース満了年月日は最終支払日以降（最終支払日含め）でなければなりません")
	}
	// 租赁满了日算出
	expireymd := stymd.AddDate(0, leasekikan+extentionOption, 0)

	// 判断
	if expireymd.Before(payymd) {
		return errors.New("リース満了年月日は最終支払日以降（最終支払日含め）でなければなりません")
	}

	return nil
}
