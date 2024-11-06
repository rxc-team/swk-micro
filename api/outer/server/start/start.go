package start

import (
	"rxcsoft.cn/utils/config"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// MongoKey mongo key
	MongoKey = "mongo"
)

// DBStart 启动DB服务
func DBStart() {
	// 获取DB配置文件
	mongo := config.GetConf(MongoKey)
	// 启动mongo
	database.StartMongodb(mongo)
}
