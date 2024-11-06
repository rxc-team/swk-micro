package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/question"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Question 问题
type Question struct{}

// log出力使用
const (
	ActionFindQuestion    = "FindQuestion"
	ActionFindQuestions   = "FindQuestions"
	ActionAddQuestion     = "AddQuestion"
	ActionModifyQuestion  = "ModifyQuestion"
	ActionDeleteQuestion  = "DeleteQuestion"
	ActionDeleteQuestions = "DeleteQuestions"
)

// FindQuestion 获取单个问题
func (q *Question) FindQuestion(ctx context.Context, req *question.FindQuestionRequest, rsp *question.FindQuestionResponse) error {
	utils.InfoLog(ActionFindQuestion, utils.MsgProcessStarted)

	res, err := model.FindQuestion(req.GetQuestionId())
	if err != nil {
		utils.ErrorLog(ActionFindQuestion, err.Error())
		return err
	}

	rsp.Question = res.ToProto()

	utils.InfoLog(ActionFindQuestion, utils.MsgProcessEnded)
	return nil
}

// FindQuestions 获取多个问题
func (q *Question) FindQuestions(ctx context.Context, req *question.FindQuestionsRequest, rsp *question.FindQuestionsResponse) error {
	utils.InfoLog(ActionFindQuestions, utils.MsgProcessStarted)

	questionList, err := model.FindQuestions(req.GetTitle(), req.GetType(), req.GetFunction(), req.GetStatus(), req.GetDomain())
	if err != nil {
		utils.ErrorLog(ActionFindQuestions, err.Error())
		return err
	}

	res := &question.FindQuestionsResponse{}

	for _, question := range questionList {
		res.Questions = append(res.Questions, question.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindQuestions, utils.MsgProcessEnded)
	return nil
}

// AddQuestion 添加问题
func (q *Question) AddQuestion(ctx context.Context, req *question.AddQuestionRequest, rsp *question.AddQuestionResponse) error {
	utils.InfoLog(ActionAddQuestion, utils.MsgProcessStarted)

	params := model.Question{
		QuestionID:     req.GetQuestionId(),
		Title:          req.GetTitle(),
		Type:           req.GetType(),
		Function:       req.GetFunction(),
		Images:         req.GetImages(),
		Content:        req.GetContent(),
		Domain:         req.GetDomain(),
		QuestionerName: req.GetQuestionerName(),
		ResponderID:    req.GetResponderId(),
		ResponderName:  req.GetResponderName(),
		Locations:      req.GetLocations(),
		Status:         "open",
		CreatedAt:      time.Now(),
		CreatedBy:      req.GetWriter(),
		UpdatedAt:      time.Now(),
		UpdatedBy:      req.GetWriter(),
	}

	id, err := model.AddQuestion(&params)
	if err != nil {
		utils.ErrorLog(ActionAddQuestion, err.Error())
		return err
	}

	rsp.QuestionId = id

	utils.InfoLog(ActionAddQuestion, utils.MsgProcessEnded)
	return nil
}

// ModifyQuestion 更新问题
func (q *Question) ModifyQuestion(ctx context.Context, req *question.ModifyQuestionRequest, rsp *question.ModifyQuestionResponse) error {
	utils.InfoLog(ActionModifyQuestion, utils.MsgProcessStarted)

	var ps []model.Postscript
	if req.GetPostscript() != nil {
		content := req.GetPostscript().GetContent()
		link := req.GetPostscript().GetLink()
		images := req.GetPostscript().GetImages()
		if content != "" || link != "" || len(images) > 0 {
			ps = append(ps, model.Postscript{
				Postscripter:     req.GetPostscript().GetPostscripter(),
				PostscripterName: req.GetPostscript().GetPostscripterName(),
				Avatar:           req.GetPostscript().GetAvatar(),
				Content:          req.GetPostscript().GetContent(),
				Link:             req.GetPostscript().GetLink(),
				Images:           req.GetPostscript().GetImages(),
				PostscriptedAt:   time.Now(),
			})
		}
	} else {
		ps = nil
	}

	params := model.Question{
		QuestionID:  req.GetQuestionId(),
		Postscripts: ps,
		Status:      req.GetStatus(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.GetWriter(),
	}

	err := model.ModifyQuestion(&params)
	if err != nil {
		utils.ErrorLog(ActionModifyQuestion, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyQuestion, utils.MsgProcessEnded)
	return nil
}

// DeleteQuestion 硬删除问题
func (q *Question) DeleteQuestion(ctx context.Context, req *question.DeleteQuestionRequest, rsp *question.DeleteQuestionResponse) error {
	utils.InfoLog(ActionDeleteQuestion, utils.MsgProcessStarted)

	err := model.DeleteQuestion(req.GetQuestionId())
	if err != nil {
		utils.ErrorLog(ActionDeleteQuestion, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteQuestion, utils.MsgProcessEnded)
	return nil
}

// DeleteQuestions 硬删除多个问题
func (q *Question) DeleteQuestions(ctx context.Context, req *question.DeleteQuestionsRequest, rsp *question.DeleteQuestionsResponse) error {
	utils.InfoLog(ActionDeleteQuestions, utils.MsgProcessStarted)

	err := model.DeleteQuestions(req.GetQuestionIdList())
	if err != nil {
		utils.ErrorLog(ActionDeleteQuestions, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteQuestions, utils.MsgProcessEnded)
	return nil
}
