/*
 * @Description:总表数据（handler）
 * @Author: PSD  王世佳
 * @Date: 2023-04-20 14:25:57
 * @LastEditors: PSD  王世佳、孫学霖、李景哲
 * @LastEditTime: 2023-04-20 14:25:57
 */
package handler

import (
	"context"

	"rxcsoft.cn/pit3/srv/report/model"
	"rxcsoft.cn/pit3/srv/report/proto/coldata"
	"rxcsoft.cn/pit3/srv/report/utils"
)

// ColData 用户信息
type ColData struct{}

// FindColDatas 获取总表数据
func (u *ColData) FindColDatas(ctx context.Context, req *coldata.FindColDatasRequest, rsp *coldata.FindColDatasResponse) error {
	utils.InfoLog("FindColDatas", utils.MsgProcessStarted)

	coldatas, err := model.FindColDatas(req.GetDatabase(), req.GetAppId(), req.GetPageIndex(), req.GetPageSize())
	if err != nil {
		utils.ErrorLog("FindColDatas", err.Error())
		return err
	}

	res := &coldata.FindColDatasResponse{}
	for _, u := range coldatas.ColData {
		res.ColDatas = append(res.ColDatas, u.ToProto())
	}
	rsp.Total = coldatas.Total
	rsp.ColDatas = res.ColDatas
	utils.InfoLog("FindColDatas", utils.MsgProcessEnded)
	return nil
}

// SelectColData  契约番号，年月获取总表数据
func (u *ColData) SelectColData(ctx context.Context, req *coldata.SelectColDataRequest, rsp *coldata.SelectColDataResponse) error {
	utils.InfoLog("FindColDatas", utils.MsgProcessStarted)

	coldatas, err := model.SelectColData(req.GetDatabase(), req.GetAppId(), req.GetKeiyakuno(), req.GetYear(), req.GetMonth(), req.GetPageIndex(), req.GetPageSize())
	if err != nil {
		utils.ErrorLog("FindColDatas", err.Error())
		return err
	}

	res := &coldata.SelectColDataResponse{}
	for _, u := range coldatas.ColData {
		res.ColDatas = append(res.ColDatas, u.ToProto())
	}
	rsp.Total = coldatas.Total
	rsp.ColDatas = res.ColDatas
	utils.InfoLog("FindColDatas", utils.MsgProcessEnded)
	return nil
}

// CreateColData 生成总表数据
func (u *ColData) CreateColData(ctx context.Context, req *coldata.CreateColDataRequest, rsp *coldata.CreateColDataResponse) error {
	utils.InfoLog(ActionGenerateReportData, utils.MsgProcessStarted)

	err := model.CreateColData(req.GetDatabase(), req.Items)
	if err != nil {
		utils.ErrorLog(ActionGenerateReportData, err.Error())
		return err
	}

	utils.InfoLog(ActionGenerateReportData, utils.MsgProcessEnded)
	return nil
}

// Download 总表CSV下载
func (u *ColData) Download(ctx context.Context, req *coldata.DownloadRequest, stream coldata.ColDataService_DownloadStream) error {
	utils.InfoLog("Download", utils.MsgProcessStarted)

	err := model.Download(req.GetDatabase(), req.GetAppId(), req.GetKeiyakuno(), req.GetYear(), req.GetMonth(), stream)
	if err != nil {
		utils.ErrorLog("FindColDatas", err.Error())
		return err
	}

	utils.InfoLog("Download", utils.MsgProcessEnded)
	return nil
}
