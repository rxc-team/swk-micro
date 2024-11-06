package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/journal/model"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/pit3/srv/journal/utils"
)

// Journal 仕訳
type Journal struct{}

// log出力使用
const (
	JournalProcessName = "Journal"

	ActionFindJournals  = "FindJournals"
	ActionFindJournal   = "FindJournal"
	ActionImportJournal = "ImportJournal"
	ActionModifyJournal = "ModifyJournal"
	ActionDeleteJournal = "DeleteJournal"
)

// FindJournals 获取多个仕訳
func (f *Journal) FindJournals(ctx context.Context, req *journal.JournalsRequest, rsp *journal.JournalsResponse) error {
	utils.InfoLog(ActionFindJournals, utils.MsgProcessStarted)

	journals, err := model.FindJournals(req.GetDatabase(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionFindJournals, err.Error())
		return err
	}

	res := &journal.JournalsResponse{}
	for _, t := range journals {
		res.Journals = append(res.Journals, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindJournals, utils.MsgProcessEnded)
	return nil
}

// FindJournal 通过JobID获取仕訳
func (f *Journal) FindJournal(ctx context.Context, req *journal.JournalRequest, rsp *journal.JournalResponse) error {
	utils.InfoLog(ActionFindJournal, utils.MsgProcessStarted)

	res, err := model.FindJournal(req.GetDatabase(), req.GetAppId(), req.GetJournalId())
	if err != nil {
		utils.ErrorLog(ActionFindJournal, err.Error())
		return err
	}

	rsp.Journal = res.ToProto()

	utils.InfoLog(ActionFindJournal, utils.MsgProcessEnded)
	return nil
}

// ImportJournal 导入仕訳
func (f *Journal) ImportJournal(ctx context.Context, req *journal.ImportRequest, rsp *journal.ImportResponse) error {
	utils.InfoLog(ActionImportJournal, utils.MsgProcessStarted)

	var journals []*model.Journal

	for _, j := range req.Journals {
		var patterns []*model.Pattern
		for _, p := range j.GetPatterns() {
			var subs []*model.JSubject
			for _, s := range p.GetSubjects() {
				subs = append(subs, &model.JSubject{
					SubjectKey:      s.GetSubjectKey(),
					LendingDivision: s.GetLendingDivision(),
					ChangeFlag:      s.GetChangeFlag(),
					DefaultName:     s.GetDefaultName(),
					AmountName:      s.GetAmountName(),
					AmountField:     s.GetAmountField(),
					SubjectName:     s.GetSubjectName(),
				})
			}

			patterns = append(patterns, &model.Pattern{
				PatternID:   p.GetPatternId(),
				PatternName: p.GetPatternName(),
				Subjects:    subs,
			})
		}

		journals = append(journals, &model.Journal{
			JournalID:   j.GetJournalId(),
			JournalName: j.GetJournalName(),
			AppID:       j.GetAppId(),
			Patterns:    patterns,
			CreatedAt:   time.Now(),
			CreatedBy:   req.GetWriter(),
			UpdatedAt:   time.Now(),
			UpdatedBy:   req.GetWriter(),
		})

	}

	err := model.ImportJournal(req.GetDatabase(), journals)
	if err != nil {
		utils.ErrorLog(ActionImportJournal, err.Error())
		return err
	}

	utils.InfoLog(ActionImportJournal, utils.MsgProcessEnded)

	return nil
}

// ModifyJournal 更新仕訳
func (f *Journal) ModifyJournal(ctx context.Context, req *journal.ModifyRequest, rsp *journal.ModifyResponse) error {
	utils.InfoLog(ActionModifyJournal, utils.MsgProcessStarted)

	param := model.JournalParam{
		JournalID:       req.GetJournalId(),
		AppID:           req.GetAppId(),
		PatternID:       req.GetPatternId(),
		SubjectKey:      req.GetSubjectKey(),
		LendingDivision: req.GetLendingDivision(),
		ChangeFlag:      req.GetChangeFlag(),
		SubjectName:     req.GetSubjectName(),
		AmountName:      req.GetAmountName(),
		AmountField:     req.GetAmountField(),
	}

	err := model.ModifyJournal(req.GetDatabase(), req.GetWriter(), param)
	if err != nil {
		utils.ErrorLog(ActionModifyJournal, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyJournal, utils.MsgProcessEnded)
	return nil
}
