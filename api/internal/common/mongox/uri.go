package mongox

import (
	"net/url"
	"strings"

	"rxcsoft.cn/utils/config"
)

// log出力使用
const (
	// MaxPoolSize 连接池大小
	MaxPoolSize uint64 = 1000
	// MongoKey mongo key
	MongoKey = "mongo"
)

// GetURI 获取mongodb的连接
func GetURI() string {
	// 获取DB配置文件
	env := config.GetConf(MongoKey)

	uri := strings.Builder{}
	uri.WriteString("mongodb://")
	if len(env.Username) > 0 && len(env.Password) > 0 {
		uri.WriteString(env.Username)
		uri.WriteString(":")
		uri.WriteString(url.QueryEscape(env.Password))
	}
	uri.WriteString("@")
	uri.WriteString(env.Host)
	uri.WriteString("/?")
	uri.WriteString("replicaSet=")
	uri.WriteString(env.ReplicaSetName)

	uri.WriteString("&authSource=")
	uri.WriteString(env.Source)

	return uri.String()
}
