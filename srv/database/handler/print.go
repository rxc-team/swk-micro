package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/database/model"
	"rxcsoft.cn/pit3/srv/database/proto/print"
	"rxcsoft.cn/pit3/srv/database/utils"
)

// Print 台账打印设置
type Print struct{}

// log出力使用
const (
	PrintProcessName       = "Print"
	ActionFindPrints       = "FindPrints"
	ActionFindPrint        = "FindPrint"
	ActionAddPrint         = "AddPrint"
	ActionModifyPrint      = "ModifyPrint"
	ActionHardDeletePrints = "HardDeletePrints"
)

// FindPrints 获取台账打印设置
func (f *Print) FindPrints(ctx context.Context, req *print.FindPrintsRequest, rsp *print.FindPrintsResponse) error {
	utils.InfoLog(ActionFindPrints, utils.MsgProcessStarted)

	prs, err := model.FindPrints(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId())
	if err != nil {
		utils.ErrorLog(ActionFindPrints, err.Error())
		return err
	}

	res := &print.FindPrintsResponse{}
	for _, ps := range prs {
		res.Prints = append(res.Prints, ps.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindPrints, utils.MsgProcessEnded)
	return nil
}

// FindPrint 通过AppID和台账ID获取台账打印设置
func (f *Print) FindPrint(ctx context.Context, req *print.FindPrintRequest, rsp *print.FindPrintResponse) error {
	utils.InfoLog(ActionFindPrint, utils.MsgProcessStarted)

	res, err := model.FindPrint(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId())
	if err != nil {
		utils.ErrorLog(ActionFindPrint, err.Error())
		return err
	}

	rsp.Print = res.ToProto()

	utils.InfoLog(ActionFindPrint, utils.MsgProcessEnded)
	return nil
}

// AddPrint 添加台账打印设置
func (f *Print) AddPrint(ctx context.Context, req *print.AddPrintRequest, rsp *print.AddPrintResponse) error {
	utils.InfoLog(ActionAddPrint, utils.MsgProcessStarted)
	var fs []*model.PrintField
	for _, f := range req.GetFields() {
		fs = append(fs, &model.PrintField{
			FieldID:   f.FieldId,
			FieldName: f.FieldName,
			FieldType: f.FieldType,
			IsImage:   f.IsImage,
			AsTitle:   f.AsTitle,
			Cols:      f.Cols,
			Rows:      f.Rows,
			X:         f.X,
			Y:         f.Y,
			Width:     f.Width,
			Precision: f.Precision,
		})
	}

	params := model.Print{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		Page:        req.GetPage(),
		Orientation: req.GetOrientation(),
		CheckField:  req.GetCheckField(),
		TitleWidth:  req.GetTitleWidth(),
		Fields:      fs,
		ShowSign:    req.ShowSign,
		SignName1:   req.SignName1,
		SignName2:   req.SignName2,
		ShowSystem:  req.ShowSystem,
		CreatedAt:   time.Now(),
		CreatedBy:   req.Writer,
	}

	err := model.AddPrint(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddPrint, err.Error())
		return err
	}

	utils.InfoLog(ActionAddPrint, utils.MsgProcessEnded)

	return nil
}

// ModifyPrint 修改台账打印设置
func (f *Print) ModifyPrint(ctx context.Context, req *print.ModifyPrintRequest, rsp *print.ModifyPrintResponse) error {
	utils.InfoLog(ActionModifyPrint, utils.MsgProcessStarted)

	var fs []*model.PrintField
	for _, f := range req.GetFields() {
		fs = append(fs, &model.PrintField{
			FieldID:   f.FieldId,
			FieldName: f.FieldName,
			FieldType: f.FieldType,
			IsImage:   f.IsImage,
			AsTitle:   f.AsTitle,
			Cols:      f.Cols,
			Rows:      f.Rows,
			X:         f.X,
			Y:         f.Y,
			Width:     f.Width,
			Precision: f.Precision,
		})
	}

	params := model.PrintModifyParam{
		AppID:       req.GetAppId(),
		DatastoreID: req.GetDatastoreId(),
		Page:        req.GetPage(),
		Orientation: req.GetOrientation(),
		CheckField:  req.GetCheckField(),
		TitleWidth:  req.GetTitleWidth(),
		Fields:      fs,
		ShowSign:    req.ShowSign,
		SignName1:   req.SignName1,
		SignName2:   req.SignName2,
		ShowSystem:  req.ShowSystem,
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.Writer,
	}

	err := model.ModifyPrint(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyPrint, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyPrint, utils.MsgProcessEnded)
	return nil
}

// HardDeletePrints 物理删除台账打印设置
func (f *Print) HardDeletePrints(ctx context.Context, req *print.HardDeletePrintsRequest, rsp *print.HardDeletePrintsResponse) error {
	utils.InfoLog(ActionHardDeletePrints, utils.MsgProcessStarted)

	err := model.HardDeletePrints(req.GetDatabase(), req.GetAppId(), req.GetDatastoreId())
	if err != nil {
		utils.ErrorLog(ActionHardDeletePrints, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeletePrints, utils.MsgProcessEnded)
	return nil
}
