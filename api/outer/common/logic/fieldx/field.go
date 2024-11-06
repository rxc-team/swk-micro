package fieldx

import (
	"context"
	"fmt"
	"sort"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/outer/common/containerx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/common/typesx"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
)

// 获取当前台账的字段
func GetFields(db, datastoreID, appID string, roles []string, showFile, showTitle bool) []*field.Field {
	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.FieldsRequest
	req.DatastoreId = datastoreID
	req.AppId = appID
	req.Database = db
	if !showTitle {
		req.AsTitle = "false"
	}

	response, err := fieldService.FindFields(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("getFields", err.Error())
		return nil
	}

	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	var preq permission.FindActionsRequest
	preq.RoleId = roles
	preq.PermissionType = "app"
	preq.AppId = appID
	preq.ActionType = "datastore"
	preq.ObjectId = req.DatastoreId
	preq.Database = db
	pResp, err := pmService.FindActions(context.TODO(), &preq)
	if err != nil {
		loggerx.ErrorLog("getFields", err.Error())
		return nil
	}

	set := containerx.New()
	for _, act := range pResp.GetActions() {
		if act.ObjectId == req.DatastoreId {
			set.AddAll(act.Fields...)
		}
	}

	fieldList := set.ToList()
	allFields := response.GetFields()
	var result []*field.Field
	for _, fieldID := range fieldList {
		f, err := FindField(fieldID, allFields)
		if err == nil {
			result = append(result, f)
		}
	}

	if showFile {
		// 排序
		sort.Sort(typesx.FieldList(result))
		return result
	}

	var fields []*field.Field
	// 去掉文件字段
	for _, f := range result {
		if f.GetFieldType() != "file" {
			fields = append(fields, f)
		}
	}

	// 排序
	sort.Sort(typesx.FieldList(fields))
	return fields
}

func FindField(fieldID string, fields []*field.Field) (r *field.Field, err error) {
	var reuslt *field.Field
	for _, f := range fields {
		if f.GetFieldId() == fieldID {
			reuslt = f
			break
		}
	}

	if reuslt == nil {
		return nil, fmt.Errorf("not found")
	}

	return reuslt, nil
}
