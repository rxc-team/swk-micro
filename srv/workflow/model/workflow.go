package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/workflow/proto/workflow"
	"rxcsoft.cn/pit3/srv/workflow/utils"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// WorkflowCollection schedule collection
	WorkflowCollection = "wf_workflows"
)

type (
	// Workflow 流程
	Workflow struct {
		ID              primitive.ObjectID `json:"id" bson:"_id"`
		WorkflowID      string             `json:"wf_id" bson:"wf_id"`
		WorkflowName    string             `json:"wf_name" bson:"wf_name"`
		MenuName        string             `json:"menu_name" bson:"menu_name"`
		IsValid         bool               `json:"is_valid" bson:"is_valid"`
		GroupID         string             `json:"group_id" bson:"group_id"`
		AppID           string             `json:"app_id" bson:"app_id"`
		WorkflowType    string             `json:"workflow_type" bson:"workflow_type"`
		AcceptOrDismiss bool               `json:"accept_or_dismiss" bson:"accept_or_dismiss"`
		Params          map[string]string  `json:"params" bson:"params"`
		CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
		CreatedBy       string             `json:"created_by" bson:"created_by"`
		UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
		UpdatedBy       string             `json:"updated_by" bson:"updated_by"`
	}
	// WorkflowUpdateParam 流程更新参数
	WorkflowUpdateParam struct {
		WfId            string
		IsValid         string
		AcceptOrDismiss string
		Params          map[string]string
		Writer          string
	}
)

// ToProto 转换为proto数据
func (w *Workflow) ToProto() *workflow.Workflow {
	return &workflow.Workflow{
		WfId:            w.WorkflowID,
		WfName:          w.WorkflowName,
		MenuName:        w.MenuName,
		IsValid:         w.IsValid,
		GroupId:         w.GroupID,
		AppId:           w.AppID,
		WorkflowType:    w.WorkflowType,
		AcceptOrDismiss: w.AcceptOrDismiss,
		Params:          w.Params,
		CreatedAt:       w.CreatedAt.String(),
		CreatedBy:       w.CreatedBy,
		UpdatedAt:       w.UpdatedAt.String(),
		UpdatedBy:       w.UpdatedBy,
	}
}

// FindWorkflows 获取流程数据
func FindWorkflows(db, appId, isValid, groupId, objectId, action string) (items []Workflow, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(WorkflowCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"app_id": appId,
	}

	if len(groupId) > 0 {
		query["group_id"] = groupId
	}
	if len(objectId) > 0 {
		query["params.datastore"] = objectId
	}
	if len(action) > 0 {
		query["params.action"] = action
	}

	if len(isValid) > 0 {
		result, err := strconv.ParseBool(isValid)
		if err != nil {
			result = false
		}
		query["is_valid"] = result
	}

	var result []Workflow

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	wf, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindWorkflows", err.Error())
		return nil, err
	}
	defer wf.Close(ctx)
	for wf.Next(ctx) {
		var wo Workflow
		err := wf.Decode(&wo)
		if err != nil {
			utils.ErrorLog("error FindWorkflows", err.Error())
			return nil, err
		}
		result = append(result, wo)
	}

	return result, nil
}

// FindUserWorkflows 获取当前用户需要走的流程
func FindUserWorkflows(db, appId, objectId, groupId, action string) (items []Workflow, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(WorkflowCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"is_valid":         true, // 默认必须是有效的流程
		"app_id":           appId,
		"group_id":         groupId,
		"params.datastore": objectId,
	}

	if len(action) > 0 {
		query["params.action"] = action
	}

	var result []Workflow

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	wf, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindWorkflows", err.Error())
		return nil, err
	}
	defer wf.Close(ctx)
	for wf.Next(ctx) {
		var wo Workflow
		err := wf.Decode(&wo)
		if err != nil {
			utils.ErrorLog("error FindWorkflows", err.Error())
			return nil, err
		}
		result = append(result, wo)
	}

	return result, nil
}

// FindWorkflow 获取流程数据
func FindWorkflow(db, workflowID string) (items Workflow, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(WorkflowCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"wf_id": workflowID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindWorkflow", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Workflow

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindWorkflow", err.Error())
		return result, err
	}

	return result, nil
}

// AddWorkflow 添加流程数据
func AddWorkflow(db string, s *Workflow) (wfID string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(WorkflowCollection)
	// rc := client.Database(database.GetDBName(db)).Collection(RelationCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client.Database(database.GetDBName(db)).CreateCollection(ctx, WorkflowCollection)
	client.Database(database.GetDBName(db)).CreateCollection(ctx, RelationCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("AddWorkflow", err.Error())
		return "", err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("AddWorkflow", err.Error())
		return "", err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		s.ID = primitive.NewObjectID()
		s.WorkflowID = s.ID.Hex()

		s.WorkflowName = "apps." + s.AppID + ".workflows." + s.WorkflowID
		s.MenuName = "apps." + s.AppID + ".workflows.menu_" + s.WorkflowID

		_, err = c.InsertOne(sc, s)
		if err != nil {
			utils.ErrorLog("AddWorkflow", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("AddWorkflow", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("AddWorkflow", err.Error())
		return "", err
	}

	session.EndSession(ctx)

	return s.WorkflowID, nil
}

// ModifyWorkflow 更新流程数据
func ModifyWorkflow(db string, params *WorkflowUpdateParam) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(WorkflowCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wfID := params.WfId
	isValid := params.IsValid
	acceptOrDismiss := params.AcceptOrDismiss
	writer := params.Writer
	ps := params.Params

	objectID, err := primitive.ObjectIDFromHex(wfID)
	if err != nil {
		utils.ErrorLog("error ModifyWorkflow", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": writer,
	}
	if len(isValid) > 0 {
		ok, err := strconv.ParseBool(isValid)
		if err != nil {
			ok = false
		}
		change["is_valid"] = ok
	}
	if len(acceptOrDismiss) > 0 {
		ok, err := strconv.ParseBool(acceptOrDismiss)
		if err != nil {
			ok = false
		}
		change["accept_or_dismiss"] = ok
	}
	fs, exit := ps["fields"]
	if exit {
		change["params.fields"] = fs
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyWorkflow", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyWorkflow", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyWorkflow", err.Error())
		return err
	}

	return nil
}

// DeleteWorkflow 删除流程数据
func DeleteWorkflow(db string, workflows []string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(WorkflowCollection)
	rc := client.Database(database.GetDBName(db)).Collection(RelationCollection)
	ec := client.Database(database.GetDBName(db)).Collection(ExampleCollection)
	nc := client.Database(database.GetDBName(db)).Collection(NodeCollection)
	pc := client.Database(database.GetDBName(db)).Collection(ProcessCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("DeleteWorkflow", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("DeleteWorkflow", err.Error())
		return err
	}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, wfID := range workflows {
			objectID, err := primitive.ObjectIDFromHex(wfID)
			if err != nil {
				utils.ErrorLog("DeleteWorkflow", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteWorkflow", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("DeleteWorkflow", err.Error())
				return err
			}

			query1 := bson.M{
				"wf_id": wfID,
			}

			// 删除关系
			_, err = rc.DeleteMany(sc, query1)
			if err != nil {
				utils.ErrorLog("DeleteWorkflow", err.Error())
				return err
			}

			// 删除 example 的数据
			_, err = ec.DeleteMany(sc, query1)
			if err != nil {
				utils.ErrorLog("DeleteWorkflow", err.Error())
				return err
			}

			// 删除 node 的数据
			_, err = nc.DeleteMany(sc, query1)
			if err != nil {
				utils.ErrorLog("DeleteWorkflow", err.Error())
				return err
			}

			// 删除 node 的数据
			_, err = pc.DeleteMany(sc, query1)
			if err != nil {
				utils.ErrorLog("DeleteWorkflow", err.Error())
				return err
			}

		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("DeleteWorkflow", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("DeleteWorkflow", err.Error())
		return err
	}

	session.EndSession(ctx)
	return nil
}
