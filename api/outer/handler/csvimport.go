package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/spf13/cast"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/srv/import/proto/upload"
)

func csvUpload(c *gin.Context, filePath, zipFilePath, payFilePath string) error {

	base := upload.Params{
		JobId:       c.PostForm("job_id"),
		Action:      c.PostForm("action"),
		Encoding:    c.PostForm("encoding"),
		ZipCharset:  c.PostForm("zip-charset"),
		EmptyChange: cast.ToBool(c.PostForm("empty_change")),
		UserId:      sessionx.GetAuthUserID(c),
		AppId:       sessionx.GetCurrentApp(c),
		Lang:        sessionx.GetCurrentLanguage(c),
		Domain:      sessionx.GetUserDomain(c),
		DatastoreId: c.Param("d_id"),
		GroupId:     sessionx.GetUserGroup(c),
		AccessKeys:  sessionx.GetUserAccessKeys(c, c.Param("d_id"), "W"),
		Owners:      sessionx.GetUserOwner(c),
		Roles:       sessionx.GetUserRoles(c),
		WfId:        c.Query("wf_id"),
		Database:    sessionx.GetUserCustomer(c),
	}

	file := upload.FileParams{
		FilePath:    filePath,
		ZipFilePath: zipFilePath,
		PayFilePath: payFilePath,
	}

	uploadService := upload.NewUploadService("import", client.DefaultClient)

	var req upload.CSVRequest
	// 从query获取
	req.BaseParams = &base
	req.FileParams = &file

	_, err := uploadService.CSVUpload(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("csvUpload", err.Error())
		return err
	}

	return nil
}
