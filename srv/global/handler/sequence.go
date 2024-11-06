package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/sequence"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Sequence 採番
type Sequence struct{}

// log出力使用
const (
	ActionFindSequence = "FindSequence"
	ActionAddSequence  = "AddSequence"
)

// FindSequence 获取单个採番
func (q *Sequence) FindSequence(ctx context.Context, req *sequence.FindSequenceRequest, rsp *sequence.FindSequenceResponse) error {
	utils.InfoLog(ActionFindSequence, utils.MsgProcessStarted)

	seq, err := model.FindSequence(req.GetDatabase(), req.GetSequenceKey())
	if err != nil {
		utils.ErrorLog(ActionFindSequence, err.Error())
		return err
	}

	rsp.Sequence = seq

	utils.InfoLog(ActionFindSequence, utils.MsgProcessEnded)
	return nil
}

// AddSequence 添加採番
func (q *Sequence) AddSequence(ctx context.Context, req *sequence.AddRequest, rsp *sequence.AddResponse) error {
	utils.InfoLog(ActionAddSequence, utils.MsgProcessStarted)

	seq, err := model.AddSequence(req.GetDatabase(), req.GetSequenceKey(), req.GetStartValue())
	if err != nil {
		utils.ErrorLog(ActionAddSequence, err.Error())
		return err
	}

	rsp.Sequence = seq

	utils.InfoLog(ActionAddSequence, utils.MsgProcessEnded)
	return nil
}
