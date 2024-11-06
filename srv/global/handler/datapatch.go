package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/datapatch"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// DataPatch
type DataPatch struct{}

// log出力使用
const (
	DataPatchProcessName = "Language"
	ActionDataPatch1216  = "DataPatch1216"
)

// DataPatch1216
func (d *DataPatch) DataPatch1216(ctx context.Context, req *datapatch.DataPatch1216Request, rsp *datapatch.DataPatch1216Response) error {
	utils.InfoLog(ActionDataPatch1216, utils.MsgProcessStarted)

	params := model.DataPatch1216Param{
		Domain: req.GetDomain(),
		LangCd: req.GetLangCd(),
		AppID:  req.GetAppId(),
		Kind:   req.GetKind(),
		Type:   req.GetType(),
		DelKbn: req.GetDelKbn(),
		Value:  req.GetValue(),
		Writer: req.GetWriter(),
	}

	err := model.DataPatch1216(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionDataPatch1216, err.Error())
		return err
	}

	utils.InfoLog(ActionDataPatch1216, utils.MsgProcessEnded)
	return nil
}
