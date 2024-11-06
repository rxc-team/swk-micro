package wfx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
)

func GetUserWorkflow(db, groupId, appId, datastore, action string) []*workflow.Workflow {
	return findWorkflows(db, groupId, appId, datastore, action)
}

func findWorkflows(db, groupId, appId, datastore, action string) []*workflow.Workflow {

	groupInfo := getGroupInfo(db, groupId)

	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var req workflow.UserWorkflowsRequest
	req.AppId = appId
	req.ObjectId = datastore
	req.GroupId = groupId
	req.Action = action
	req.Database = db

	response, err := workflowService.FindUserWorkflows(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("GetUserWorkflow", err.Error())
		return nil
	}

	if len(response.GetWorkflows()) == 0 {
		if groupInfo.GetParentGroupId() == "root" {
			return nil
		}

		return findWorkflows(db, groupInfo.GetParentGroupId(), appId, datastore, action)
	}

	return response.GetWorkflows()
}

func getGroupInfo(db, groupId string) *group.Group {
	groupService := group.NewGroupService("manage", client.DefaultClient)

	var req group.FindGroupRequest
	req.GroupId = groupId
	req.Database = db
	response, err := groupService.FindGroup(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getGroupInfo", err.Error())
		return nil
	}

	return response.GetGroup()
}
