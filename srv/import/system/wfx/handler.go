package wfx

// Work 任务
type Work struct {
	WorkflowID string // 流程ID
	ExampleID  string // 实例ID
	Database   string // 所属台账
	UserID     string // 操作者
}

// Handler 处理接口
type Handler interface {
	Admit(w *Work) (string, error)
	Dismiss(w *Work) (string, error)
}

func createHandler(wType string) Handler {
	var handler Handler = nil
	switch wType {
	case "datastore":
		handler = new(DsHandler)
	}

	return handler
}
