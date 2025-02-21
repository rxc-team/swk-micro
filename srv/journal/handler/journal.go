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

	ActionFindJournals            = "FindJournals"
	ActionFindJournal             = "FindJournal"
	ActionImportJournal           = "ImportJournal"
	ActionModifyJournal           = "ModifyJournal"
	ActionDeleteJournal           = "DeleteJournal"
	ActionAddDownloadSetting      = "AddDownloadSetting"
	ActionFindDownloadSetting     = "FindDownloadSetting"
	ActionAddSelectJournals       = "AddSelectJournals"
	ActionFindSelectJournals      = "FindSelectJournals"
	ActionFindConditionTemplates  = "FindConditionTemplates"
	ActionAddConditionTemplate    = "AddConditionTemplate"
	ActionDeleteConditionTemplate = "DeleteConditionTemplate"
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
		SubjectCd:       req.GetSubjectCd(),
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

// 添加分录下载设定
func (f *Journal) AddDownloadSetting(ctx context.Context, req *journal.AddDownloadSettingRequest, rsp *journal.AddDownloadSettingResponse) error {

	var fieldRules []*model.FieldRule

	for _, rule := range req.GetFieldRule() {
		var fieldConditions []*model.FieldCondition
		for _, condition := range rule.GetFieldConditions() {
			var fieldGroups []*model.FieldGroup
			for _, group := range condition.GetFieldGroups() {
				var fieldCons []*model.FieldCon
				for _, con := range group.GetFieldCons() {
					fieldCons = append(fieldCons, &model.FieldCon{
						ConID:       con.ConId,
						ConName:     con.ConName,
						ConField:    con.ConField,
						ConOperator: con.ConOperator,
						ConValue:    con.ConValue,
						ConDataType: con.ConDataType,
					})
				}

				fieldGroups = append(fieldGroups, &model.FieldGroup{
					GroupID:    group.GroupId,
					GroupName:  group.GroupName,
					Type:       group.Type,
					SwitchType: group.SwitchType,
					FieldCons:  fieldCons,
				})

			}

			var thenCustomFields []*model.CustomField
			for _, customfield := range condition.GetThenCustomFields() {
				thenCustomFields = append(thenCustomFields, &model.CustomField{
					CustomFieldType:     customfield.CustomFieldType,
					CustomFieldValue:    customfield.CustomFieldValue,
					CustomFieldDataType: customfield.CustomFieldDataType,
				})
			}

			var elseCustomFields []*model.CustomField
			for _, customfield := range condition.GetElseCustomFields() {
				elseCustomFields = append(elseCustomFields, &model.CustomField{
					CustomFieldType:     customfield.CustomFieldType,
					CustomFieldValue:    customfield.CustomFieldValue,
					CustomFieldDataType: customfield.CustomFieldDataType,
				})
			}

			fieldConditions = append(fieldConditions, &model.FieldCondition{
				ConditionID:       condition.ConditionId,
				ConditionName:     condition.ConditionName,
				ThenValue:         condition.ThenValue,
				ElseValue:         condition.ElseValue,
				ThenType:          condition.ThenType,
				ElseType:          condition.ElseType,
				ThenCustomType:    condition.ThenCustomType,
				ElseCustomType:    condition.ElseCustomType,
				FieldGroups:       fieldGroups,
				ThenCustomFields:  thenCustomFields,
				ElseCustomFields:  elseCustomFields,
				ThenValueDataType: condition.ThenValueDataType,
				ElseValueDataType: condition.ElseValueDataType,
			})
		}

		fieldRules = append(fieldRules, &model.FieldRule{
			DownloadName:      rule.DownloadName,
			FieldId:           rule.FieldId,
			FieldConditions:   fieldConditions,
			SettingMethod:     rule.SettingMethod,
			FieldType:         rule.FieldType,
			DatastoreId:       rule.DatastoreId,
			Format:            rule.Format,
			EditContent:       rule.EditContent,
			ElseValue:         rule.ElseValue,
			ElseType:          rule.ElseType,
			ElseValueDataType: rule.ElseValueDataType,
		})
	}

	params := model.FieldConf{
		AppId:         req.AppId,
		LayoutName:    req.LayoutName,
		CharEncoding:  req.CharEncoding,
		HeaderRow:     req.HeaderRow,
		SeparatorChar: req.SeparatorChar,
		LineBreaks:    req.LineBreaks,
		FixedLength:   req.FixedLength,
		NumberItems:   req.NumberItems,
		ValidFlag:     req.ValidFlag,
		FieldRule:     fieldRules,
	}

	err := model.AddDownloadSetting(req.GetDatabase(), req.GetAppId(), params)
	if err != nil {
		utils.ErrorLog(ActionAddDownloadSetting, err.Error())
		return err
	}

	utils.InfoLog(ActionAddDownloadSetting, utils.MsgProcessEnded)

	return nil
}

// 查找分录下载设定
func (f *Journal) FindDownloadSetting(ctx context.Context, req *journal.FindDownloadSettingRequest, rsp *journal.FindDownloadSettingResponse) error {

	res, err := model.FindDownloadSetting(req.GetDatabase(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionFindDownloadSetting, err.Error())
		return err
	}

	*rsp = *res.ToProto()

	utils.InfoLog(ActionFindDownloadSetting, utils.MsgProcessEnded)

	return nil
}

// 查找分录下载设定
func (f *Journal) FindDownloadSettings(ctx context.Context, req *journal.FindDownloadSettingsRequest, rsp *journal.FindDownloadSettingsResponse) error {

	res, err := model.FindDownloadSettings(req.GetDatabase(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionFindDownloadSetting, err.Error())
		return err
	}

	// *rsp = *res.ToProto()
	*rsp = *model.ConvertToProto(res)

	utils.InfoLog(ActionFindDownloadSetting, utils.MsgProcessEnded)

	return nil
}

// 添加选择分录
func (f *Journal) AddSelectJournals(ctx context.Context, req *journal.AddSelectJournalsRequest, rsp *journal.AddSelectJournalsResponse) error {

	err := model.AddSelectJournals(req.GetSelectedJournal(), req.GetDatabase(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionFindDownloadSetting, err.Error())
		return err
	}

	utils.InfoLog(ActionFindDownloadSetting, utils.MsgProcessEnded)

	return nil
}

// FindSelectJournals 获取选择的仕訳
func (f *Journal) FindSelectJournals(ctx context.Context, req *journal.JournalsRequest, rsp *journal.JournalsResponse) error {
	utils.InfoLog(ActionFindSelectJournals, utils.MsgProcessStarted)

	journals, err := model.FindSelectJournals(req.GetDatabase(), req.GetAppId())
	if err != nil {
		utils.ErrorLog(ActionFindSelectJournals, err.Error())
		return err
	}

	res := &journal.JournalsResponse{}
	for _, t := range journals {
		res.Journals = append(res.Journals, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindSelectJournals, utils.MsgProcessEnded)
	return nil
}

// 查找所有自定义条件
func (f *Journal) FindConditionTemplates(ctx context.Context, req *journal.FindConditionTemplatesRequest, rsp *journal.FindConditionTemplatesResponse) error {
	res, err := model.FindConditionTemplates(req.GetDatabase(), req.GetAppId())

	if err != nil {
		utils.ErrorLog(ActionFindConditionTemplates, err.Error())
		return err
	}

	// *rsp = *res.ToProto(res)
	*rsp = *model.ToProto(res)

	utils.InfoLog(ActionFindConditionTemplates, utils.MsgProcessEnded)

	return nil
}

// 添加自定义条件模板
func (f *Journal) AddConditionTemplate(ctx context.Context, req *journal.AddConditionTemplateRequest, rsp *journal.AddConditionTemplateResponse) error {

	fieldCondition := req.GetConditionTemplate().GetFieldCondition()
	var groups []*model.FieldGroup
	for _, g := range fieldCondition.GetFieldGroups() {
		var cons []*model.FieldCon
		for _, con := range g.GetFieldCons() {
			cons = append(cons, &model.FieldCon{
				ConID:       con.ConId,
				ConName:     con.ConName,
				ConField:    con.ConField,
				ConOperator: con.ConOperator,
				ConValue:    con.ConValue,
				ConDataType: con.ConDataType,
			})
		}

		groups = append(groups, &model.FieldGroup{
			GroupID:    g.GroupId,
			GroupName:  g.GroupName,
			Type:       g.Type,
			SwitchType: g.SwitchType,
			FieldCons:  cons,
		})
	}

	var thenCustomFields []*model.CustomField
	for _, customfield := range req.GetConditionTemplate().GetFieldCondition().GetThenCustomFields() {
		thenCustomFields = append(thenCustomFields, &model.CustomField{
			CustomFieldType:     customfield.CustomFieldType,
			CustomFieldValue:    customfield.CustomFieldValue,
			CustomFieldDataType: customfield.CustomFieldDataType,
		})
	}

	var elseCustomFields []*model.CustomField
	for _, customfield := range req.GetConditionTemplate().GetFieldCondition().GetElseCustomFields() {
		elseCustomFields = append(elseCustomFields, &model.CustomField{
			CustomFieldType:     customfield.CustomFieldType,
			CustomFieldValue:    customfield.CustomFieldValue,
			CustomFieldDataType: customfield.CustomFieldDataType,
		})
	}

	fc := model.FieldCondition{
		ConditionID:       req.GetConditionTemplate().GetFieldCondition().GetConditionId(),
		ConditionName:     req.GetConditionTemplate().GetFieldCondition().GetConditionName(),
		ThenValue:         req.GetConditionTemplate().GetFieldCondition().GetThenValue(),
		ElseValue:         req.GetConditionTemplate().GetFieldCondition().GetElseValue(),
		ThenType:          req.GetConditionTemplate().GetFieldCondition().GetThenType(),
		ElseType:          req.GetConditionTemplate().GetFieldCondition().GetElseType(),
		ThenCustomType:    req.GetConditionTemplate().GetFieldCondition().GetThenCustomType(),
		ElseCustomType:    req.GetConditionTemplate().GetFieldCondition().GetElseCustomType(),
		FieldGroups:       groups,
		ThenCustomFields:  thenCustomFields,
		ElseCustomFields:  elseCustomFields,
		ThenValueDataType: req.GetConditionTemplate().GetFieldCondition().GetThenValueDataType(),
		ElseValueDataType: req.GetConditionTemplate().GetFieldCondition().GetElseValueDataType(),
	}

	params := model.ConditionTemplate{
		TemplateId:     req.GetConditionTemplate().GetTemplateId(),
		TemplateName:   req.GetConditionTemplate().GetTemplateName(),
		FieldCondition: fc,
		CreatedAt:      time.Now(),
		CreatedBy:      req.GetWriter(),
		AppID:          req.GetAppId(),
	}

	err := model.AddConditionTemplate(req.GetDatabase(), req.GetAppId(), params)
	if err != nil {
		utils.ErrorLog(ActionAddConditionTemplate, err.Error())
		return err
	}

	utils.InfoLog(ActionAddConditionTemplate, utils.MsgProcessEnded)

	return nil
}

// 删除自定义条件模板
func (f *Journal) DeleteConditionTemplate(ctx context.Context, req *journal.DeleteConditionTemplateRequest, rsp *journal.DeleteConditionTemplateResponse) error {
	err := model.DeleteConditionTemplate(req.GetDatabase(), req.GetAppId(), req.GetTemplateId())

	if err != nil {
		utils.ErrorLog(ActionDeleteConditionTemplate, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteConditionTemplate, utils.MsgProcessEnded)

	return nil
}
