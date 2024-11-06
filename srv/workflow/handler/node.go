package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/workflow/model"
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/utils"
)

// Node 流程节点
type Node struct{}

// log出力使用
const (
	NodeProcessName = "Node"

	ActionFindNodes  = "FindNodes"
	ActionFindNode   = "FindNode"
	ActionAddNode    = "AddNode"
	ActionModifyNode = "ModifyNode"
	ActionDeleteNode = "DeleteNode"
)

// FindNodes 获取多个流程节点
func (f *Node) FindNodes(ctx context.Context, req *node.NodesRequest, rsp *node.NodesResponse) error {
	utils.InfoLog(ActionFindNodes, utils.MsgProcessStarted)

	nodes, err := model.FindNodes(req.GetDatabase(), req.GetWfId())
	if err != nil {
		utils.ErrorLog(ActionFindNodes, err.Error())
		return err
	}

	res := &node.NodesResponse{}
	for _, t := range nodes {
		res.Nodes = append(res.Nodes, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindNodes, utils.MsgProcessEnded)
	return nil
}

// FindNode 通过JobID获取流程节点
func (f *Node) FindNode(ctx context.Context, req *node.NodeRequest, rsp *node.NodeResponse) error {
	utils.InfoLog(ActionFindNode, utils.MsgProcessStarted)

	res, err := model.FindNode(req.GetDatabase(), req.GetWfId(), req.GetNodeId())
	if err != nil {
		utils.ErrorLog(ActionFindNode, err.Error())
		return err
	}

	rsp.Node = res.ToProto()

	utils.InfoLog(ActionFindNode, utils.MsgProcessEnded)
	return nil
}

// AddNode 添加流程节点
func (f *Node) AddNode(ctx context.Context, req *node.AddRequest, rsp *node.AddResponse) error {
	utils.InfoLog(ActionAddNode, utils.MsgProcessStarted)

	param := model.Node{
		NodeID:      req.GetNodeId(),
		NodeName:    req.GetNodeName(),
		WorkflowID:  req.GetWfId(),
		NodeType:    req.GetNextNode(),
		PrevNode:    req.GetPrevNode(),
		NextNode:    req.GetNextNode(),
		Assignees:   req.GetAssignees(),
		ActType:     req.GetActType(),
		NodeGroupId: req.GetNodeGroupId(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	id, err := model.AddNode(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionAddNode, err.Error())
		return err
	}

	rsp.NodeId = id

	utils.InfoLog(ActionAddNode, utils.MsgProcessEnded)

	return nil
}

// DeleteNode 删除流程节点
func (f *Node) DeleteNode(ctx context.Context, req *node.DeleteRequest, rsp *node.DeleteResponse) error {
	utils.InfoLog(ActionDeleteNode, utils.MsgProcessStarted)

	err := model.DeleteNode(req.GetDatabase(), req.GetWfId())
	if err != nil {
		utils.ErrorLog(ActionDeleteNode, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteNode, utils.MsgProcessEnded)
	return nil
}
