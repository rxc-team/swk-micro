package model

// 审批的数据，不能满了，解约，情报变更，债务变更（status == 2）
func NonAdmitCheck(status string) bool {
	return status == "2"
}
