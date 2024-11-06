package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/outer/common/containerx"
	"rxcsoft.cn/pit3/api/outer/common/httpx"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/api/outer/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
)

// Datastore datastore
type Datastore struct{}

// log出力
const (
	DatastoreProcessName = "Datastore"
	ActionFindDatastore  = "FindDatastore"
	ActionFindDatastores = "FindDatastores"
)

// FindDatastores 获取app下所有台账
// @Summary 获取app下所有台账
// @description 调用srv中的datastore服务，获取app下所有台账
// @Tags Datastore
// @Accept json
// @Security JWT
// @Produce  json
// @Param a_id path string true "AppID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /datastores [get]
func (d *Datastore) FindDatastores(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDatastores, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoresRequest
	// 从query获取
	req.DatastoreName = c.Query("datastore_name")
	req.CanCheck = c.Query("can_check")
	req.ShowInMenu = c.Query("show_in_menu")
	// 从共通获取
	req.AppId = sessionx.GetCurrentApp(c)
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.FindDatastores(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastores, err)
		return
	}

	needRole := c.Query("needRole")
	if needRole == "true" {
		roles := sessionx.GetUserRoles(c)
		set := containerx.New()

		pmService := permission.NewPermissionService("manage", client.DefaultClient)

		var preq permission.FindActionsRequest
		preq.RoleId = roles
		preq.PermissionType = "app"
		preq.AppId = sessionx.GetCurrentApp(c)
		preq.ActionType = "datastore"
		preq.Database = sessionx.GetUserCustomer(c)
		pResp, err := pmService.FindActions(context.TODO(), &preq)
		if err != nil {
			httpx.GinHTTPError(c, ActionFindDatastores, err)
			return
		}
		for _, act := range pResp.GetActions() {
			if act.ActionMap["read"] {
				set.Add(act.ObjectId)
			}
		}

		dsList := set.ToList()
		allDs := response.GetDatastores()
		var result []*datastore.Datastore
		for _, dsID := range dsList {
			f, err := findDatastore(dsID, allDs)
			if err == nil {
				result = append(result, f)
			}
		}

		loggerx.InfoLog(c, ActionFindDatastores, loggerx.MsgProcessEnded)
		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastores)),
			Data:    result,
		})

		return
	}

	loggerx.InfoLog(c, ActionFindDatastores, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastores)),
		Data:    response.GetDatastores(),
	})
}

func findDatastore(dsID string, dsList []*datastore.Datastore) (r *datastore.Datastore, err error) {
	var reuslt *datastore.Datastore
	for _, d := range dsList {
		if d.GetDatastoreId() == dsID {
			reuslt = d
			break
		}
	}

	if reuslt == nil {
		return nil, fmt.Errorf("not found")
	}

	return reuslt, nil
}

// FindDatastore 通过DatastoreID获取台账信息
// @Summary 通过DatastoreID获取台账信息
// @description 调用srv中的datastore服务，通过DatastoreID获取台账信息
// @Tags Datastore
// @Accept json
// @Security JWT
// @Produce  json
// @Param d_id path string true "DatastoreID"
// @Success 200 {object} handler.Response
// @Failure 401 {object} handler.ErrorResponse
// @Failure 403 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /datastores/{d_id} [get]
func (d *Datastore) FindDatastore(c *gin.Context) {
	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessStarted)

	datastoreService := datastore.NewDataStoreService("database", client.DefaultClient)

	var req datastore.DatastoreRequest
	// 从path获取
	req.DatastoreId = c.Param("d_id")
	req.Database = sessionx.GetUserCustomer(c)

	response, err := datastoreService.FindDatastore(context.TODO(), &req)
	if err != nil {
		httpx.GinHTTPError(c, ActionFindDatastore, err)
		return
	}

	loggerx.InfoLog(c, ActionFindDatastore, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, DatastoreProcessName, ActionFindDatastore)),
		Data:    response.GetDatastore(),
	})
}
