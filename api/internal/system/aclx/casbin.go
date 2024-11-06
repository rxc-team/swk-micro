package aclx

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
	mongodbadapter "github.com/casbin/mongodb-adapter/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/utils/config"
)

const (
	// MaxPoolSize 连接池大小
	MaxPoolSize uint64 = 2000
	// MaxPoolSize uint64 = 1000
	// MongoKey mongo key
	MongoKey = "mongo"
)

var instance *casbin.SyncedEnforcer
var once sync.Once

// 権限更新、APP複製時などに使用⇒権限確認不要のためAddFunction除外
func GetCasbin() *casbin.SyncedEnforcer {
	once.Do(func() {
		// 获取DB配置文件
		mongo := config.GetConf(MongoKey)
		a, err := mongodbadapter.NewAdapterWithClientOption(getMongodbOption(mongo), mongo.Database)
		if err != nil {
			loggerx.FatalLog("GetCasbin", err.Error())
		}
		e, err := casbin.NewSyncedEnforcer("assets/conf/auth_model.conf", a)
		if err != nil {
			loggerx.FatalLog("GetCasbin", err.Error())
		}
		e.LoadPolicy()
		e.StartAutoLoadPolicy(5 * time.Second)
		// e.AddFunction("pathExist", PathExistFunc)
		// e.AddFunction("keyMatch8", KeyMatch8Func)
		instance = e
	})
	return instance
}

func GetCasbin_request(app string, roleid []string, userid string, path string, Method string, objectId string) (*casbin.SyncedEnforcer, bool) {

	// 获取DB配置文件
	mongo := config.GetConf(MongoKey)
	a, err := mongodbadapter.NewAdapterWithClientOption(getMongodbOption(mongo), mongo.Database)
	if err != nil {
		loggerx.FatalLog("GetCasbin_request", err.Error())
	}
	e, err := casbin.NewSyncedEnforcer("assets/conf/auth_model.conf", a)
	if err != nil {
		loggerx.FatalLog("GetCasbin_request", err.Error())
	}

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

	pathbool := true
	pathMatch := path
	if (strings.HasPrefix(path, "/internal/api/v1/web/file/")) && (objectId == "public" || objectId == "company" || objectId == "user") {
	} else {
		for k := range actionMap {
			ks := strings.Split(k, "#")

			if util.KeyMatch2(path, ks[0]) && Method == ks[1] { //リクエストのパスと直書きのパスを比較
				pathbool = false //権限が必要なパスは「false」で上書きして、後続処理で詳しく権限確認
				pathMatch = ks[0]
			}
		}
	}
	if pathbool { //権限が必要なければ、getcasbinを実行せずreturn
		return e, pathbool
	}

	//顧客権限のPとG、いずれも取得する
	//{[G] v0：userid かつ v2：app} または ｛[P] v0：roleid かつ　v1：Path かつ v2：メソッド｝
	filter := bson.M{"$or": []bson.M{{"$and": []bson.M{{"v0": userid}, {"v2": app}}}, {"$and": []bson.M{{"v0": roleid[0]}, {"v1": pathMatch}, {"v2": Method}}}}}
	e.LoadFilteredPolicy(filter)
	e.AddFunction("keyMatch8", KeyMatch8Func)

	//instance = e
	return e, pathbool

}

// getMongodbOption 获取mongodb的连接
func getMongodbOption(env config.DB) *options.ClientOptions {
	cfe, host := getMongodbHost(env)
	hosts := buildHost(host, cfe.Port)

	option := options.Client()
	option.SetHosts(hosts)                   // 设置连接host
	option.SetReplicaSet(env.ReplicaSetName) // 设置replica name
	option.SetMaxPoolSize(MaxPoolSize)       // 设置最大连接池的数量
	option.SetMinPoolSize(0)                 // 设定最小连接池大小
	option.SetRetryReads(true)               // 增加读的重试
	option.SetRetryWrites(true)              // 增加写的重试
	// 设置读的偏好
	r := readpref.Primary()
	//r := readpref.SecondaryPreferred()
	option.SetReadPreference(r)
	// 设置读隔离
	rn := readconcern.Local()
	option.SetReadConcern(rn)
	// 设置写隔离
	w := writeconcern.WriteConcern{}
	w.WithOptions(writeconcern.WMajority())
	option.SetWriteConcern(&w)
	if len(env.Username) > 0 && len(env.Password) > 0 {
		option.SetAuth(
			options.Credential{ // 设置认证信息
				AuthSource: env.Source,
				Username:   env.Username,
				Password:   env.Password,
			})
	}

	return option
}

// 构建db连接信息
func buildHost(hosts, port string) []string {
	mongodbHosts := splitMongodbInstances(hosts)
	// 编辑host主机信息
	for i, host := range mongodbHosts {
		if !strings.Contains(host, ":") {
			mongodbHosts[i] = fmt.Sprintf("%s:%s", host, port)
		}
	}

	return mongodbHosts
}

// splitMongodbInstances 将mongodb实例字符串拆分为切片
func splitMongodbInstances(instances string) []string {
	var hosts []string
	hosts = append(hosts, strings.Split(instances, ",")...)

	return hosts
}

func getMongodbHost(env config.DB) (config.DB, string) {
	return env, fmt.Sprintf("%v", env.Host)
}
