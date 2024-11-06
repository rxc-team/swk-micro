package jsoncache

import (
	"encoding/json"
	"time"

	"rxcsoft.cn/pit3/srv/global/utils"
	"rxcsoft.cn/utils/logger"
)

var (
	separator = ":"
	log       = logger.New()
)

// JSONStringify 将对象转换为json字符串数据
func JSONStringify(data interface{}) (string, error) {
	str, err := json.Marshal(data)
	if err != nil {
		utils.ErrorLog("error JSONStringify", err.Error())
		return "", err
	}

	return string(str), nil
}

// CacheSetEx 简单缓存json数据到字符串值，ttl以秒为单位
func CacheSetEx(value string, keys []string, ttl int64) error {
	s := time.Now()

	if err := SetEx(joinKeysToSTR(keys...), value, ttl); err != nil {
		utils.ErrorLog("CacheSetEx", err.Error())
		return err
	}
	log.Infof("CacheSetEx took: %v key: %v ", time.Since(s), joinKeysToSTR(keys...))
	return nil
}

// CacheGetEx 获取字符串缓存并转换为json
func CacheGetEx(keys []string, ttl int64) ([]uint8, error) {
	s := time.Now()
	str, err := GetEx(joinKeysToSTR(keys...), ttl)
	if err != nil {
		utils.ErrorLog("CacheGetEx Hit miss!", err.Error())
		return str, err
	}

	log.Infof("CacheGetEx Hit! took: %v keys: %v", time.Since(s), joinKeysToSTR(keys...))
	return str, nil
}

// DeleteCache 删除缓存
func DeleteCache(keys []string) error {
	s := time.Now()

	if err := Delete(joinKeysToSTR(keys...)); err != nil {
		return err
	}

	log.Infof("CacheDelete took %v keys: %v", time.Since(s), joinKeysToSTR(keys...))
	return nil
}

// StringToStruct 将字符串转换为结构体
func StringToStruct(result string, target interface{}) error {
	if err := json.Unmarshal([]byte(result), &target); err != nil {
		utils.ErrorLog("error in StringToStruct", err.Error())
		return err
	}

	return nil
}

func joinKeysToSTR(keys ...string) string {
	return JoinKeyWords(keys...)
}
