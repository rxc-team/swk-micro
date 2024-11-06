package lease

// PayData 支付数据排序
type PayData []Payment

//排序规则：按displayOrder排序（由小到大）
func (list PayData) Len() int {
	return len(list)
}

func (list PayData) Less(i, j int) bool {
	return list[i].Paymentcount < list[j].Paymentcount
}

func (list PayData) Swap(i, j int) {
	var temp Payment = list[i]
	list[i] = list[j]
	list[j] = temp
}
