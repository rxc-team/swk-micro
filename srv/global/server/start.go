package server

import (
	"rxcsoft.cn/utils/config"
	"rxcsoft.cn/utils/database"
	database1 "rxcsoft.cn/utils/mongo"
)

const (
	// RedisKey redis key
	RedisKey = "redis"
	// MongoKey mongo key
	MongoKey = "mongo"
)

// LoadConfig 加载配置文件
func LoadConfig() {
	// 加载配置文件
	config.InitConfig()
}

// DBStart 启动DB服务
func DBStart() {
	// 获取DB配置文件
	redis := config.GetConf(RedisKey)
	// 启动redis
	database.StartRedis(redis)
	// 获取DB配置文件
	mongo := config.GetConf(MongoKey)
	// 启动mongo
	database1.StartMongodb(mongo)
}
