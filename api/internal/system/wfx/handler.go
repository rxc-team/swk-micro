package wfx

// Work 任务
type Work struct {
	WorkflowID string // 流程ID
	ExampleID  string // 实例ID
	Database   string // 所属台账
	UserID     string // 操作者
}
