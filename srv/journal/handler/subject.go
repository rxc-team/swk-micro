package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/journal/model"
	"rxcsoft.cn/pit3/srv/journal/proto/subject"
	"rxcsoft.cn/pit3/srv/journal/utils"
)

// Subject 科目
type Subject struct{}

// log出力使用
const (
	SubjectProcessName = "Subject"

	ActionFindSubjects  = "FindSubjects"
	ActionFindSubject   = "FindSubject"
	ActionImportSubject = "ImportSubject"
	ActionModifySubject = "ModifySubject"
	ActionDeleteSubject = "DeleteSubject"
)

// FindSubjects 获取多个科目
func (f *Subject) FindSubjects(ctx context.Context, req *subject.SubjectsRequest, rsp *subject.SubjectsResponse) error {
	utils.InfoLog(ActionFindSubjects, utils.MsgProcessStarted)

	assetsType := req.GetAssetsType()
	// 类型为空的情况下，获取系统默认的科目
	if len(assetsType) == 0 {
		subjects, err := model.FindDefaultSubjects(req.GetDatabase(), req.GetAppId())
		if err != nil {
			utils.ErrorLog(ActionFindSubjects, err.Error())
			return err
		}

		res := &subject.SubjectsResponse{}
		for _, t := range subjects {
			res.Subjects = append(res.Subjects, t.ToProto())
		}

		*rsp = *res

		utils.InfoLog(ActionFindSubjects, utils.MsgProcessEnded)
		return nil
	}

	subjects, err := model.FindSubjects(req.GetDatabase(), req.GetAppId(), req.GetAssetsType())
	if err != nil {
		utils.ErrorLog(ActionFindSubjects, err.Error())
		return err
	}

	res := &subject.SubjectsResponse{}
	for _, t := range subjects {
		res.Subjects = append(res.Subjects, t.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindSubjects, utils.MsgProcessEnded)
	return nil
}

// FindSubject 通过JobID获取科目
func (f *Subject) FindSubject(ctx context.Context, req *subject.SubjectRequest, rsp *subject.SubjectResponse) error {
	utils.InfoLog(ActionFindSubject, utils.MsgProcessStarted)

	res, err := model.FindSubject(req.GetDatabase(), req.GetAppId(), req.GetAssetsType(), req.GetSubjectKey())
	if err != nil {
		utils.ErrorLog(ActionFindSubject, err.Error())
		return err
	}

	rsp.Subject = res.ToProto()

	utils.InfoLog(ActionFindSubject, utils.MsgProcessEnded)
	return nil
}

// ImportSubject 导入科目
func (f *Subject) ImportSubject(ctx context.Context, req *subject.ImportRequest, rsp *subject.ImportResponse) error {
	utils.InfoLog(ActionImportSubject, utils.MsgProcessStarted)

	var subjects []*model.Subject

	for _, j := range req.Subjects {
		subjects = append(subjects, &model.Subject{
			SubjectKey:  j.GetSubjectKey(),
			SubjectName: j.GetSubjectName(),
			AssetsType:  j.GetAssetsType(),
			AppID:       j.GetAppId(),
			CreatedAt:   time.Now(),
			CreatedBy:   req.GetWriter(),
			UpdatedAt:   time.Now(),
			UpdatedBy:   req.GetWriter(),
		})

	}

	err := model.ImportSubject(req.GetDatabase(), subjects)
	if err != nil {
		utils.ErrorLog(ActionImportSubject, err.Error())
		return err
	}

	utils.InfoLog(ActionImportSubject, utils.MsgProcessEnded)

	return nil
}

// ModifySubject 更新科目
func (f *Subject) ModifySubject(ctx context.Context, req *subject.ModifyRequest, rsp *subject.ModifyResponse) error {
	utils.InfoLog(ActionModifySubject, utils.MsgProcessStarted)

	err := model.ModifySubject(req.GetDatabase(), req.GetAppId(), req.GetAssetsType(), req.GetSubjectKey(), req.GetSubjectName(), req.GetDefaultName(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionModifySubject, err.Error())
		return err
	}

	utils.InfoLog(ActionModifySubject, utils.MsgProcessEnded)
	return nil
}

// DeleteSubject 删除科目
func (f *Subject) DeleteSubject(ctx context.Context, req *subject.DeleteRequest, rsp *subject.DeleteResponse) error {
	utils.InfoLog(ActionDeleteSubject, utils.MsgProcessStarted)

	err := model.DeleteSubject(req.GetDatabase(), req.GetAppId(), req.GetAssetsType(), req.GetSubjectKey())
	if err != nil {
		utils.ErrorLog(ActionDeleteSubject, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSubject, utils.MsgProcessEnded)
	return nil
}
