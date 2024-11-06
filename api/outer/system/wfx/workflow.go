package wfx

import (
	"context"
	"errors"
	"time"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/outer/common/containerx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/common/logic/mailx"
	"rxcsoft.cn/pit3/api/outer/system/wsx"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/workflow/proto/example"
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/proto/process"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
)

// Approve 审批
type Approve struct{}

// WfInfo 流程信息
type WfInfo struct {
	Workflow *workflow.Workflow
	Nodes    []*node.Node
}

const (
	defaultOriginURL = "http://localhost:4201"
	webuiUrlEnv      = "WEBUI_URL"
)

// AddExample 添加流程实例
func (a *Approve) AddExample(db, wfID, userID string) (string, error) {

	wf, err := findWfInfo(db, wfID)
	if err != nil {
		return "", err
	}

	// 开启一个流程实例
	exService := example.NewExampleService("workflow", client.DefaultClient)

	var req example.AddRequest
	req.WfId = wfID
	req.ExName = wf.Workflow.WfName + "_" + userID
	req.Status = 1
	req.Database = db
	req.UserId = userID
	req.Writer = userID

	response, err := exService.AddExample(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("StartExampleInstance", err.Error())
		return "", err
	}

	return response.GetExId(), nil
}

// StartExampleInstance 启动流程
func (a *Approve) StartExampleInstance(db, wfID, userID, exId, domain string) error {

	wf, err := findWfInfo(db, wfID)
	if err != nil {
		return err
	}

	// 根据node第一个节点，开启流程的进程
	rootNode := wf.Nodes[0]
	proceeService := process.NewProcessService("workflow", client.DefaultClient)

	// 更新group信息
	wkGroup := wf.Workflow.GroupId
	groupID := rootNode.NodeGroupId
	// 获取所有group数据
	groupService := group.NewGroupService("manage", client.DefaultClient)

	var gReq group.FindGroupsRequest
	// 当前用户的domain
	gReq.Domain = domain
	gReq.Database = db

	gResp, err := groupService.FindGroups(context.TODO(), &gReq)
	if err != nil {
		loggerx.ErrorLog("StartExampleInstance", err.Error())
		return err
	}
	// 获取用户信息
	userService := user.NewUserService("manage", client.DefaultClient)

	var ureq user.FindUserRequest
	ureq.UserId = userID
	ureq.Database = db

	uResp, err := userService.FindUser(context.TODO(), &ureq)
	if err != nil {
		loggerx.ErrorLog("StartExampleInstance", err.Error())
		return err
	}

	// 获取审批者
	approvers := findApprovers(db, domain, wkGroup, groupID, uResp.GetUser().GetGroup(), gResp.Groups, rootNode.Assignees)

	// 去除重复用户
	set := containerx.New()
	for _, u := range approvers {
		// 去除掉用户自己
		if u != userID {
			set.Add(u)
		}
	}

	approvers = set.ToList()

	// 当存在审批者的场合
	if len(approvers) > 0 {
		for _, uid := range approvers {
			var pReq process.AddRequest
			pReq.ExId = exId
			pReq.CurrentNode = rootNode.GetNodeId()
			pReq.UserId = uid
			// 设置默认过期时间为5天，如果5天未处理则自动回退
			pReq.ExpireDate = time.Now().Add(432000 * time.Second).Format("2006-01-02")
			pReq.Status = 0
			pReq.Database = db
			pReq.Writer = userID

			_, err = proceeService.AddProcess(context.TODO(), &pReq)
			if err != nil {
				loggerx.ErrorLog("StartExampleInstance", err.Error())
				return err
			}

			params := mailx.EmailParam{
				Database:       db,
				UserID:         uid,
				AppID:          wf.Workflow.GetAppId(),
				WorkflowID:     wfID,
				DatastoreID:    wf.Workflow.GetParams()["datastore"],
				Language:       "ja-JP",
				CreateUserName: uResp.GetUser().GetUserName(),
				Opreate:        wf.Workflow.GetParams()["action"],
			}

			err = mailx.SendEmailToApprover(params)
			if err != nil {
				loggerx.ErrorLog("StartExampleInstance", err.Error())
				return err
			}

			param := wsx.MessageParam{
				Sender:    "SYSTEM",
				Recipient: uid,
				MsgType:   "approve",
				Code:      "I_019",
				Content:   "新しい承認を処理する必要がありますので、確認してください。",
				Status:    "unread",
			}
			wsx.SendToUser(param)
		}
	} else {

		// 添加一条系统处理记录
		var pReq process.AddRequest
		pReq.ExId = exId
		pReq.CurrentNode = rootNode.GetNodeId()
		pReq.UserId = "system"
		// 设置默认过期时间为5天，如果5天未处理则自动回退
		pReq.ExpireDate = time.Now().Add(432000 * time.Second).Format("2006-01-02")
		pReq.Status = 0
		pReq.Database = db
		pReq.Writer = userID

		_, err = proceeService.AddProcess(context.TODO(), &pReq)
		if err != nil {
			loggerx.ErrorLog("StartExampleInstance", err.Error())
			return err
		}

		// 承认的场合
		if wf.Workflow.AcceptOrDismiss {
			approve := new(Approve)
			approve.Admit(db, exId, "system", domain, "この組織には承認者がいないため、システムは承認プロセスを実行しました。")
		} else {
			// 却下的场合
			approve := new(Approve)
			approve.Dismiss(db, exId, "system", "この組織には承認者がいないため、システムは却下プロセスを実行しました。")
		}
	}

	return nil
}

// Admit 承認
func (a *Approve) Admit(db, exID, userID, domain, comment string) error {
	// 获取流程实例
	exService := example.NewExampleService("workflow", client.DefaultClient)

	var req example.ExampleRequest
	req.ExId = exID
	req.Database = db

	exResp, err := exService.FindExample(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("admit", err.Error())
		return err
	}

	createUser := exResp.GetExample().GetUserId()

	// 查找该实例对应的进程
	proceeService := process.NewProcessService("workflow", client.DefaultClient)
	var pReq process.ProcessesRequest
	pReq.ExId = exID
	pReq.Database = db

	pResp, err := proceeService.FindProcesses(context.TODO(), &pReq)
	if err != nil {
		loggerx.ErrorLog("admit", err.Error())
		return err
	}
	var proc *process.Process
	// 获取当前用户的进程
	for i := 0; i < len(pResp.GetProcesses()); i++ {
		p := pResp.Processes[i]
		// 审批人员是当前用户，并且该审批是未审批状态
		if p.UserId == userID && p.Status == 0 {
			proc = p
			break
		}
	}

	if proc == nil {
		loggerx.ErrorLog("admit", "you're not the approver or this workflow has close")
		return errors.New("you're not the approver or this workflow has close")
	}

	// 更新当前进程的状态为【承认】
	var mReq process.ModifyRequest
	mReq.ProId = proc.ProId
	mReq.Status = "1"
	mReq.Comment = comment
	mReq.Database = db
	mReq.Writer = userID
	_, err = proceeService.ModifyProcess(context.TODO(), &mReq)
	if err != nil {
		loggerx.ErrorLog("admit", err.Error())
		return err
	}

	// 获取当前节点的信息
	nodeService := node.NewNodeService("workflow", client.DefaultClient)

	var nReq node.NodeRequest
	nReq.NodeId = proc.CurrentNode
	nReq.WfId = exResp.GetExample().GetWfId()
	nReq.Database = db

	nResp, err := nodeService.FindNode(context.TODO(), &nReq)
	if err != nil {
		loggerx.ErrorLog("admit", err.Error())
		return err
	}

	// 再次获取最新的进程信息
	npResp, err := proceeService.FindProcesses(context.TODO(), &pReq)
	if err != nil {
		loggerx.ErrorLog("admit", err.Error())
		return err
	}

	canNext := false
	// 根据节点的任务类型判断是否需要执行下一步
	if nResp.Node.ActType == "or" {
		// 循环当前进程，判断该节点的结果
		for i := 0; i < len(npResp.GetProcesses()); i++ {
			p := npResp.Processes[i]
			if p.CurrentNode == proc.CurrentNode && p.Status == 1 {
				canNext = true
				break
			}
		}
	} else {
		next := true
		// 循环当前进程，判断该节点的结果
		for i := 0; i < len(npResp.GetProcesses()); i++ {
			p := npResp.Processes[i]
			if p.CurrentNode == proc.CurrentNode && p.Status != 1 {
				next = false
				break
			}
		}

		canNext = next
	}

	if canNext {

		// 更新当前节点的剩余未审批的进程为完成状态
		for _, p := range npResp.GetProcesses() {
			if p.CurrentNode == proc.CurrentNode && p.Status == 0 && p.ProId != proc.ProId {
				var mReq process.ModifyRequest
				mReq.ProId = p.GetProId()
				mReq.Status = "1"
				mReq.Comment = "Approved by other approvers"
				mReq.Database = db
				mReq.Writer = "SYSTEM"
				_, err = proceeService.ModifyProcess(context.TODO(), &mReq)
				if err != nil {
					loggerx.ErrorLog("admit", err.Error())
					return err
				}
			}
		}

		// 判断当前节点是否为最后节点
		if nResp.Node.NextNode == "0" {
			// 更新实例
			var mexReq example.ModifyRequest
			mexReq.ExId = exID
			mexReq.Status = "2" //承认
			mexReq.Database = db
			// 如果承认者是系统的场合，使用创建者用户去更新该数据
			if userID == "system" {
				mexReq.Writer = createUser
			} else {
				mexReq.Writer = userID
			}

			_, err := exService.ModifyExample(context.TODO(), &mexReq)
			if err != nil {
				loggerx.ErrorLog("admit", err.Error())
				return err
			}

			// 结束流程，更新数据
			handler := createHandler("datastore")
			wk := &Work{
				WorkflowID: exResp.GetExample().GetWfId(),
				ExampleID:  exID,
				UserID:     createUser,
				Database:   db,
			}
			// 对数据进行承认处理
			_, er := handler.Admit(wk)
			if er != nil {
				loggerx.ErrorLog("admit", er.Error())
				// 回滚状态
				// 更新当前进程的状态为【却下】
				var mReq process.ModifyRequest
				mReq.ProId = proc.ProId
				mReq.Status = "0"
				mReq.Comment = ""
				mReq.Database = db
				mReq.Writer = "SYSTEM"
				_, e1 := proceeService.ModifyProcess(context.TODO(), &mReq)
				if e1 != nil {
					loggerx.ErrorLog("admit", e1.Error())
					return e1
				}
				// 回滚状态，更新实例
				var mexReq example.ModifyRequest
				mexReq.ExId = exID
				mexReq.Status = "1" //拒绝
				mexReq.Database = db
				// 如果承认者是系统的场合，使用创建者用户去更新该数据
				if userID == "system" {
					mexReq.Writer = createUser
				} else {
					mexReq.Writer = userID
				}

				_, e2 := exService.ModifyExample(context.TODO(), &mexReq)
				if e2 != nil {
					loggerx.ErrorLog("admit", e2.Error())
					return e2
				}

				return er
			}

			param := wsx.MessageParam{
				Sender:    "SYSTEM",
				Recipient: exResp.Example.GetUserId(),
				MsgType:   "approve",
				Code:      "I_020",
				Content:   "申請データが承認されましたので、ご確認してください。",
				Status:    "unread",
			}
			wsx.SendToUser(param)
		} else {
			// 开启流程的下一步的进程
			// 获取当前节点的信息
			nodeService := node.NewNodeService("workflow", client.DefaultClient)

			var nReq node.NodeRequest
			nReq.NodeId = nResp.Node.NextNode
			nReq.WfId = exResp.GetExample().GetWfId()
			nReq.Database = db

			nextResp, err := nodeService.FindNode(context.TODO(), &nReq)
			if err != nil {
				loggerx.ErrorLog("admit", err.Error())
				return err
			}
			proceeService := process.NewProcessService("workflow", client.DefaultClient)

			wf, err := findWfInfo(db, exResp.GetExample().GetWfId())
			if err != nil {
				loggerx.ErrorLog("admit", err.Error())
				return err
			}

			// 更新group信息
			wkGroup := wf.Workflow.GroupId
			groupID := nextResp.Node.NodeGroupId
			// 获取所有group数据
			groupService := group.NewGroupService("manage", client.DefaultClient)

			var gReq group.FindGroupsRequest
			// 当前用户的domain
			gReq.Domain = domain
			gReq.Database = db

			gResp, err := groupService.FindGroups(context.TODO(), &gReq)
			if err != nil {
				loggerx.ErrorLog("admit", err.Error())
				return err
			}

			// 获取用户信息
			userService := user.NewUserService("manage", client.DefaultClient)

			var ureq user.FindUserRequest
			ureq.UserId = createUser
			ureq.Database = db

			uResp, err := userService.FindUser(context.TODO(), &ureq)
			if err != nil {
				loggerx.ErrorLog("admit", err.Error())
				return err
			}

			// 获取审批者
			approvers := findApprovers(db, domain, wkGroup, groupID, uResp.GetUser().GetGroup(), gResp.Groups, nextResp.Node.Assignees)

			// 去除重复用户
			set := containerx.New()
			for _, u := range approvers {
				// 去除掉用户自己
				if u != createUser {
					set.Add(u)
				}
			}

			approvers = set.ToList()

			// 当存在审批者的场合
			if len(approvers) > 0 {
				for _, uid := range approvers {
					var pReq process.AddRequest
					pReq.ExId = exID
					pReq.CurrentNode = nextResp.Node.NodeId
					pReq.UserId = uid
					// 设置默认过期时间为5天，如果5天未处理则自动回退
					pReq.ExpireDate = time.Now().Add(432000 * time.Second).Format("2006-01-02")
					pReq.Status = 0
					pReq.Database = db
					// 如果承认者是系统的场合，使用创建者用户去更新该数据
					if userID == "system" {
						pReq.Writer = createUser
					} else {
						pReq.Writer = userID
					}

					_, err = proceeService.AddProcess(context.TODO(), &pReq)
					if err != nil {
						loggerx.ErrorLog("admit", err.Error())
						return err
					}

					params := mailx.EmailParam{
						Database:       db,
						UserID:         uid,
						AppID:          wf.Workflow.GetAppId(),
						WorkflowID:     exResp.GetExample().GetWfId(),
						DatastoreID:    wf.Workflow.GetParams()["datastore"],
						Language:       "ja-JP",
						CreateUserName: uResp.GetUser().GetUserName(),
						Opreate:        wf.Workflow.GetParams()["action"],
					}

					err = mailx.SendEmailToApprover(params)
					if err != nil {
						loggerx.ErrorLog("admit", err.Error())
						return err
					}

					param := wsx.MessageParam{
						Sender:    "SYSTEM",
						Recipient: uid,
						MsgType:   "approve",
						Code:      "I_019",
						Content:   "新しい承認を処理する必要がありますので、確認してください。",
						Status:    "unread",
					}
					wsx.SendToUser(param)
				}
			} else {

				var pReq process.AddRequest
				pReq.ExId = exID
				pReq.CurrentNode = nextResp.Node.NodeId
				pReq.UserId = "system"
				// 设置默认过期时间为5天，如果5天未处理则自动回退
				pReq.ExpireDate = time.Now().Add(432000 * time.Second).Format("2006-01-02")
				pReq.Status = 0
				pReq.Database = db
				pReq.Writer = userID

				_, err := proceeService.AddProcess(context.TODO(), &pReq)
				if err != nil {
					loggerx.ErrorLog("admit", err.Error())
					return err
				}

				// 承认的场合
				if wf.Workflow.AcceptOrDismiss {
					approve := new(Approve)
					approve.Admit(db, exID, "system", domain, "この組織には承認者がいないため、システムは承認プロセスを実行しました。")
				} else {
					// 却下的场合
					approve := new(Approve)
					approve.Dismiss(db, exID, "system", "この組織には承認者がいないため、システムは却下プロセスを実行しました。")
				}
			}
		}
	}

	return nil
}

// Dismiss 却下
func (a *Approve) Dismiss(db, exID, userID, comment string) error {
	// 获取流程实例
	exService := example.NewExampleService("workflow", client.DefaultClient)

	var req example.ExampleRequest
	req.ExId = exID
	req.Database = db

	exResp, err := exService.FindExample(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("Dismiss", err.Error())
		return err
	}

	createUser := exResp.GetExample().GetUserId()

	// 查找该实例对应的进程
	proceeService := process.NewProcessService("workflow", client.DefaultClient)
	var pReq process.ProcessesRequest
	pReq.ExId = exID
	pReq.Database = db

	pResp, err := proceeService.FindProcesses(context.TODO(), &pReq)
	if err != nil {
		loggerx.ErrorLog("Dismiss", err.Error())
		return err
	}
	var proc *process.Process
	// 获取当前用户的进程
	for i := 0; i < len(pResp.GetProcesses()); i++ {
		p := pResp.Processes[i]
		if p.UserId == userID && p.Status == 0 {
			proc = p
			break
		}
	}

	if proc == nil {
		loggerx.ErrorLog("Dismiss", "you're not the dismiss user or this workflow has close")
		return errors.New("you're not the dismiss user or this workflow has close")
	}

	// 更新当前进程的状态为【却下】
	var mReq process.ModifyRequest
	mReq.ProId = proc.ProId
	mReq.Status = "2"
	mReq.Comment = comment
	mReq.Database = db
	mReq.Writer = userID
	_, err = proceeService.ModifyProcess(context.TODO(), &mReq)
	if err != nil {
		loggerx.ErrorLog("Dismiss", err.Error())
		return err
	}

	// 更新其他未提交的审核进程为却下状态
	for _, p := range pResp.GetProcesses() {
		if p.CurrentNode == proc.CurrentNode && p.Status == 0 && p.ProId != proc.ProId {
			var mReq process.ModifyRequest
			mReq.ProId = p.GetProId()
			mReq.Status = "2"
			mReq.Comment = "Rejected by other approvers"
			mReq.Database = db
			mReq.Writer = "SYSTEM"
			_, err = proceeService.ModifyProcess(context.TODO(), &mReq)
			if err != nil {
				loggerx.ErrorLog("Dismiss", err.Error())
				return err
			}
		}
	}

	// 直接走却下处理
	// 更新实例
	var mexReq example.ModifyRequest
	mexReq.ExId = exID
	mexReq.Status = "3" //却下
	mexReq.Database = db
	// 如果承认者是系统的场合，使用创建者用户去更新该数据
	if userID == "system" {
		mexReq.Writer = createUser
	} else {
		mexReq.Writer = userID
	}

	_, er := exService.ModifyExample(context.TODO(), &mexReq)
	if er != nil {
		loggerx.ErrorLog("Dismiss", er.Error())
		return er
	}

	// 结束流程，更新数据
	handler := createHandler("datastore")
	wk := &Work{
		WorkflowID: exResp.GetExample().GetWfId(),
		ExampleID:  exID,
		UserID:     userID,
		Database:   db,
	}
	// 对数据进行承认处理
	handler.Dismiss(wk)
	param := wsx.MessageParam{
		Sender:    "SYSTEM",
		Recipient: exResp.Example.GetUserId(),
		MsgType:   "approve",
		Code:      "I_021",
		Content:   "申請データが拒否されましたので、ご確認ください",
		Status:    "unread",
	}
	wsx.SendToUser(param)

	return nil

}

func findWfInfo(db, wfID string) (*WfInfo, error) {
	workflowService := workflow.NewWfService("workflow", client.DefaultClient)

	var req workflow.WorkflowRequest
	req.WfId = wfID
	req.Database = db

	response, err := workflowService.FindWorkflow(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("findWfInfo", err.Error())
		return nil, err
	}

	nodeService := node.NewNodeService("workflow", client.DefaultClient)

	var nReq node.NodesRequest
	nReq.WfId = wfID
	nReq.Database = db

	nResp, err := nodeService.FindNodes(context.TODO(), &nReq)
	if err != nil {
		loggerx.ErrorLog("findWfInfo", err.Error())
		return nil, err
	}

	return &WfInfo{
		Workflow: response.Workflow,
		Nodes:    nResp.Nodes,
	}, nil

}
