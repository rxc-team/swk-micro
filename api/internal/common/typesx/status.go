package typesx

// Status 审批状态
type Status struct {
	CanShow       bool   `json:"can_show"`       // 是否能够显示
	ApproveStatus int64  `json:"approve_status"` // 承认状态
	Approver      string `json:"approver"`       // 结束时的审批者
	Applicant     string `json:"applicant"`      // 申请者
	CurrentNode   string `json:"current_node"`   // 当前状态
}
