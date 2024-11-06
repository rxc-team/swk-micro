package aclx

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/util"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/manage/proto/action"
	"rxcsoft.cn/pit3/srv/manage/proto/allow"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/level"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
)

// dashboardとか、3Tに書いていないパスが来たときにtrueを返す＝この時はkeymatch8は実行しない、
// ここに書いてあるとき＝falseなので、keymatch8で確認を行う
var actionMap = map[string]string{
	"/internal/api/v1/web/item/datastores/:d_id/items/:i_id/contract#PUT":       "contract_update",
	"/internal/api/v1/web/item/datastores/:d_id/items/:i_id/terminate#PUT":      "midway_cancel",
	"/internal/api/v1/web/item/datastores/:d_id/items/:i_id/debt#PUT":           "estimate_update",
	"/internal/api/v1/web/item/datastores/:d_id/items/:i_id/contractExpire#PUT": "contract_expire",
	"/internal/api/v1/web/item/datastores/:d_id/items/print#POST":               "pdf",
	"/internal/api/v1/web/item/clear/datastores/:d_id/items#DELETE":             "clear",
	"/internal/api/v1/web/item/datastores/:d_id/items#PATCH":                    "group",
	"/internal/api/v1/web/item/datastores/:d_id/items/owners#POST":              "group",
	"/internal/api/v1/web/history/datastores/:d_id/histories#GET":               "history",
	"/internal/api/v1/web/history/datastores/:d_id/download#GET":                "history",
	"/internal/api/v1/web/item/datastores/:d_id/items/search#POST":              "read",
	"/internal/api/v1/web/item/datastores/:d_id/items/:i_id#GET":                "read",
	"/internal/api/v1/web/item/datastores/:d_id/items#POST":                     "insert",
	"/internal/api/v1/web/item/datastores/:d_id/items/:i_id#PUT":                "update",
	"/internal/api/v1/web/item/datastores/:d_id/items/:i_id#DELETE":             "delete",
	"/internal/api/v1/web/mapping/datastores/:d_id/upload#POST":                 "mapping_upload",
	"/internal/api/v1/web/mapping/datastores/:d_id/download#POST":               "mapping_download",
	"/internal/api/v1/web/item/import/image/datastores/:d_id/items#POST":        "image",
	"/internal/api/v1/web/item/import/csv/datastores/:d_id/items#POST":          "csv",
	"/internal/api/v1/web/item/import/csv/datastores/:d_id/check/items#POST":    "inventory",
	"/internal/api/v1/web/item/datastores/:d_id/prs/download#POST":              "principal_repayment",
	"/internal/api/v1/web/item/datastores/:d_id/items/download#POST":            "data",
	"/internal/api/v1/web/report/reports/:rp_id#GET":                            "read",
	"/internal/api/v1/web/report/reports/:rp_id/data#POST":                      "read",
	"/internal/api/v1/web/report/gen/reports/:rp_id/data#POST":                  "read",
	"/internal/api/v1/web/report/reports/:rp_id/download#POST":                  "read",
	"/internal/api/v1/web/file/folders/:fo_id/files#GET":                        "read",
	"/internal/api/v1/web/file/download/folders/:fo_id/files/:file_id#GET":      "read",
	"/internal/api/v1/web/file/folders/:fo_id/upload#POST":                      "write",
	"/internal/api/v1/web/file/folders/:fo_id/files/:file_id#DELETE":            "delete",
	"/internal/api/v1/web/journal/journals#GET":                                 "read",
	"/internal/api/v1/web/journal/journals/:j_id#GET":                           "read",
	"/internal/api/v1/web/journal/journals#POST":                                "read",
	"/internal/api/v1/web/journal/compute/journals#GET":                         "read",
	"/internal/api/v1/web/journal/journals/:j_id#PUT":                           "read",
}

func PathExist(key1, method1, objectId string) bool {

	if strings.HasPrefix(key1, "/internal/api/v1/web/file/") {
		if objectId == "public" || objectId == "company" || objectId == "user" {
			return true
		}
	}

	hasExist := false

	for k := range actionMap {
		ks := strings.Split(k, "#")

		//リクエストのパスと直書きのパスを比較
		if util.KeyMatch2(key1, ks[0]) && method1 == ks[1] {
			hasExist = true
		}
	}

	//3Tにあるパスならfalseを返してkeymatch8で検証、書いてないならkeymatch8でfalseになるので、ここでtrueを返す
	return !hasExist
}

func KeyMatch8(key1, key2, method1, method2, db, objectId, appId, role string) bool {

	if !util.KeyMatch2(key1, key2) {
		return false
	}

	if method1 != method2 {
		return false
	}

	if strings.HasPrefix(key2, "/internal/api/v1/web/journal/") {
		return !checkAllow(db, "journal")
	}

	action, actionType := getActionInfo(key2, method2)
	if len(action) == 0 {
		return false
	}

	if !checkAction(db, action, objectId, actionType, appId, []string{role}) {
		return false
	}

	return true
}

func checkAllow(customerId, allowType string) bool {

	// 获取顾客信息
	customerService := customer.NewCustomerService("manage", client.DefaultClient)

	var req customer.FindCustomerRequest
	req.CustomerId = customerId
	response, err := customerService.FindCustomer(context.TODO(), &req)
	if err != nil {
		return false
	}
	// 通过顾客信息获取顾客的授权等级信息
	levelService := level.NewLevelService("manage", client.DefaultClient)

	var lreq level.FindLevelRequest
	lreq.LevelId = response.GetCustomer().GetLevel()
	levelResp, err := levelService.FindLevel(context.TODO(), &lreq)
	if err != nil {
		return false
	}

	if len(levelResp.GetLevel().GetAllows()) == 0 {
		return false
	}

	allowService := allow.NewAllowService("manage", client.DefaultClient)

	var alreq allow.FindLevelAllowsRequest
	// 从query获取
	alreq.AllowList = levelResp.GetLevel().GetAllows()

	allowResp, err := allowService.FindLevelAllows(context.TODO(), &alreq)
	if err != nil {
		return false
	}

	actionService := action.NewActionService("manage", client.DefaultClient)

	var areq action.FindActionsRequest
	aResp, err := actionService.FindActions(context.TODO(), &areq)
	if err != nil {
		return false
	}

	allowsData := allowResp.GetAllows()
	actions := aResp.GetActions()
	result := false

	for _, a := range allowsData {
		if a.AllowType == allowType {
			for _, x := range a.GetActions() {
				for _, y := range actions {
					if a.AllowType == y.ActionObject && x.ApiKey == y.ActionKey && x.ApiKey == "read" && x.GroupKey == y.ActionGroup {
						result = true
					}
				}
			}
		}

	}

	return result
}

func getActionInfo(key, method string) (string, string) {
	action := actionMap[key+"#"+method]

	actionType := "datastore"

	if strings.HasPrefix(key, "/internal/api/v1/web/item/") {
		return action, actionType
	}
	if strings.HasPrefix(key, "/internal/api/v1/web/mapping/") {
		return action, actionType
	}
	if strings.HasPrefix(key, "/internal/api/v1/web/history/") {
		return action, actionType
	}
	if strings.HasPrefix(key, "/internal/api/v1/web/report/") {
		actionType = "report"
		return action, actionType
	}
	if strings.HasPrefix(key, "/internal/api/v1/web/file/") {
		actionType = "folder"
		return action, actionType
	}

	return "", actionType

}

func checkAction(db, action, objectId, actionType, appId string, roles []string) bool {
	// 默认没有权限
	hasAccess := false

	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	var req permission.FindActionsRequest
	req.RoleId = roles
	req.PermissionType = "app"
	req.AppId = appId
	req.ActionType = actionType
	req.ObjectId = objectId
	req.Database = db

	if actionType == "folder" {
		req.AppId = ""
		req.PermissionType = "common"
	}

	resp, err := pmService.FindActions(context.TODO(), &req)
	if err != nil {
		return false
	}

LOOP:
	for _, act := range resp.GetActions() {
		if act.ObjectId == objectId {
			if val, exist := act.ActionMap[action]; exist {
				hasAccess = val
				break LOOP
			}
		}
	}

	return hasAccess

}

func KeyMatch8Func(args ...interface{}) (interface{}, error) {

	if err := validateVariadicArgs(8, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "keyMatch8", err)
	}
	// r.path, p.path, r.method, p.method, r.db, r.app, r.objectId, p.role

	//リクエストのパス
	key1 := args[0].(string)
	//3Tから取得したcasbinパス
	key2 := args[1].(string)
	method1 := args[2].(string)
	method2 := args[3].(string)
	db := args[4].(string)
	appId := args[5].(string)
	objectId := args[6].(string)
	role := args[7].(string)

	return (bool)(KeyMatch8(key1, key2, method1, method2, db, objectId, appId, role)), nil
}

func PathExistFunc(args ...interface{}) (interface{}, error) {

	if err := validateVariadicArgs(3, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "PathExist", err)
	}

	//リクエストのパス
	key1 := args[0].(string)
	method1 := args[1].(string)
	objectId := args[2].(string)

	return (bool)(PathExist(key1, method1, objectId)), nil
}

// validate the variadic parameter size and type as string
func validateVariadicArgs(expectedLen int, args ...interface{}) error {
	if len(args) != expectedLen {
		return fmt.Errorf("expected %d arguments, but got %d", expectedLen, len(args))
	}

	for _, p := range args {
		_, ok := p.(string)
		if !ok {
			return errors.New("argument must be a string")
		}
	}

	return nil
}
