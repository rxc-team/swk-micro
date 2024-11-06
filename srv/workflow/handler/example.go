package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/workflow/model"
	"rxcsoft.cn/pit3/srv/workflow/proto/example"
	"rxcsoft.cn/pit3/srv/workflow/utils"
)

// Example 流程实例
type Example struct{}

// log出力使用
const (
	ExampleProcessName = "Example"

	ActionFindExamples  = "FindExamples"
	ActionFindExample   = "FindExample"
	ActionAddExample    = "AddExample"
	ActionModifyExample = "ModifyExample"
	ActionDeleteExample = "DeleteExample"
)

// FindExamples 获取多个流程实例
func (f *Example) FindExamples(ctx context.Context, req *example.ExamplesRequest, rsp *example.ExamplesResponse) error {
	utils.InfoLog(ActionFindExamples, utils.MsgProcessStarted)

	examples, err := model.FindExamples(req.GetDatabase(), req.GetWfId())
	if err != nil {
		utils.ErrorLog(ActionFindExamples, err.Error())
		return err
	}

	res := &example.ExamplesResponse{}
	for _, t := range examples {
		res.Examples = append(res.Examples, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindExamples, utils.MsgProcessEnded)
	return nil
}

// FindExample 通过JobID获取流程实例
func (f *Example) FindExample(ctx context.Context, req *example.ExampleRequest, rsp *example.ExampleResponse) error {
	utils.InfoLog(ActionFindExample, utils.MsgProcessStarted)

	res, err := model.FindExample(req.GetDatabase(), req.GetExId())
	if err != nil {
		utils.ErrorLog(ActionFindExample, err.Error())
		return err
	}

	rsp.Example = res.ToProto()

	utils.InfoLog(ActionFindExample, utils.MsgProcessEnded)
	return nil
}

// AddExample 添加流程实例
func (f *Example) AddExample(ctx context.Context, req *example.AddRequest, rsp *example.AddResponse) error {
	utils.InfoLog(ActionAddExample, utils.MsgProcessStarted)

	param := model.Example{
		WorkflowID:  req.GetWfId(),
		ExampleName: req.GetExName(),
		UserID:      req.GetUserId(),
		Status:      req.GetStatus(),
		CreatedAt:   time.Now(),
		CreatedBy:   req.GetWriter(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	id, err := model.AddExample(req.GetDatabase(), &param)
	if err != nil {
		utils.ErrorLog(ActionAddExample, err.Error())
		return err
	}

	rsp.ExId = id

	utils.InfoLog(ActionAddExample, utils.MsgProcessEnded)

	return nil
}

// ModifyExample 更新流程实例
func (f *Example) ModifyExample(ctx context.Context, req *example.ModifyRequest, rsp *example.ModifyResponse) error {
	utils.InfoLog(ActionModifyExample, utils.MsgProcessStarted)

	err := model.ModifyExample(req.GetDatabase(), req.GetExId(), req.GetStatus(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionModifyExample, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyExample, utils.MsgProcessEnded)
	return nil
}

// DeleteExample 删除流程实例
func (f *Example) DeleteExample(ctx context.Context, req *example.DeleteRequest, rsp *example.DeleteResponse) error {
	utils.InfoLog(ActionDeleteExample, utils.MsgProcessStarted)

	err := model.DeleteExample(req.GetDatabase(), req.GetExId())
	if err != nil {
		utils.ErrorLog(ActionDeleteExample, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteExample, utils.MsgProcessEnded)
	return nil
}
