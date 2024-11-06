package wfx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
)

// CheckWfValid 判断流程是否有效
func CheckWfValid(db, wfID string) bool {
	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var req workflow.WorkflowRequest
	req.WfId = wfID
	req.Database = db

	response, err := workflowService.FindWorkflow(context.TODO(), &req)
	if err != nil {
		return false
	}

	if response.GetWorkflow().GetIsValid() {
		return true
	}
	return false
}
