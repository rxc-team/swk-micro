package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/workflow/proto/node"
	"rxcsoft.cn/pit3/srv/workflow/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// NodeCollection schedule collection
	NodeCollection = "wf_node"
)

type (
	// Node 流程节点
	Node struct {
		ID          primitive.ObjectID `json:"id" bson:"_id"`
		NodeID      string             `json:"node_id" bson:"node_id"`
		NodeName    string             `json:"node_name" bson:"node_name"`
		WorkflowID  string             `json:"wf_id" bson:"wf_id"`
		NodeType    string             `json:"node_type" bson:"node_type"`
		PrevNode    string             `json:"prev_node" bson:"prev_node"`
		NextNode    string             `json:"next_node" bson:"next_node"`
		Assignees   []string           `json:"assignees" bson:"assignees"`
		ActType     string             `json:"act_type" bson:"act_type"`
		NodeGroupId string             `json:"node_group_id" bson:"node_group_id"`
		CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy   string             `json:"created_by" bson:"created_by"`
		UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	}
)

// ToProto 转换为proto数据
func (n *Node) ToProto() *node.Node {
	return &node.Node{
		NodeId:      n.NodeID,
		NodeName:    n.NodeName,
		WfId:        n.WorkflowID,
		NodeType:    n.NodeType,
		PrevNode:    n.PrevNode,
		NextNode:    n.NextNode,
		Assignees:   n.Assignees,
		ActType:     n.ActType,
		NodeGroupId: n.NodeGroupId,
		CreatedAt:   n.CreatedAt.String(),
		CreatedBy:   n.CreatedBy,
	}
}

// FindNodes 获取流程节点数据
func FindNodes(db, wfID string) (items []Node, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(NodeCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	if wfID != "" {
		query["wf_id"] = wfID
	}

	var result []Node
	opts := options.Find().SetSort(bson.D{{Key: "node_id", Value: 1}})
	nodes, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindNodes", err.Error())
		return nil, err
	}
	defer nodes.Close(ctx)
	for nodes.Next(ctx) {
		var nd Node
		err := nodes.Decode(&nd)
		if err != nil {
			utils.ErrorLog("error FindNodes", err.Error())
			return nil, err
		}
		result = append(result, nd)
	}

	return result, nil
}

// FindNode 获取流程节点数据
func FindNode(db, wfID, nodeID string) (n Node, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(NodeCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"wf_id":   wfID,
		"node_id": nodeID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindNode", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Node

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindNode", err.Error())
		return result, err
	}

	return result, nil
}

// AddNode 添加流程节点数据
func AddNode(db string, s *Node) (scheduleID string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(NodeCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.ID = primitive.NewObjectID()

	_, err = c.InsertOne(ctx, s)
	if err != nil {
		utils.ErrorLog("error AddNode", err.Error())
		return "", err
	}

	return s.NodeID, nil
}

// DeleteNode 删除流程节点数据
func DeleteNode(db string, wfID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(NodeCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"wf_id": wfID,
	}
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteNode", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteMany(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteNode", err.Error())
		return err
	}
	return nil
}
