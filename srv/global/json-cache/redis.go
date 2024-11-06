package jsoncache

import (
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"rxcsoft.cn/pit3/srv/global/utils"
	"rxcsoft.cn/utils/database"
)

var (

	// KeyExpiration 最大关键时间秒
	KeyExpiration = 10000000

	// DefaultKeySeperator 的默认redis键分隔符
	DefaultKeySeperator = ":"
)

// Set 将键和值字符串设置到redis
func Set(key, value string) error {
	s := time.Now()
	c := database.GetRedisCon()
	defer c.Close()

	if _, err := c.Do("SET", key, value); err != nil {
		utils.ErrorLog("erro Redis Set", err.Error())
		return err
	}
	log.Infof("Set took: %v", time.Since(s))
	return nil
}

// SetEx 设置kv，并设置过期时间
func SetEx(key, value string, ttl int64) error {
	s := time.Now()
	c := database.GetRedisCon()
	defer c.Close()

	if err := c.Send("MULTI"); err != nil {
		utils.ErrorLog("error Multi", err.Error())
		return err
	}

	if err := c.Send("SET", key, value); err != nil {
		utils.ErrorLog("error Setex", err.Error())
		return err
	}

	if ttl > 0 {
		if err := c.Send("EXPIRE", key, ttl); err != nil {
			utils.ErrorLog("error EXPIRE", err.Error())
			return err
		}
	}

	if _, err := c.Do("EXEC"); err != nil {
		utils.ErrorLog("error EXEC", err.Error())
		return err
	}
	log.Infof("SetEx took: %v", time.Since(s))
	return nil
}

// GetEx 通过key获取值，并更新过期时间
func GetEx(key string, ttl int64) ([]uint8, error) {
	s := time.Now()
	c := database.GetRedisCon()
	defer c.Close()

	if err := c.Send("MULTI"); err != nil {
		return []uint8{}, fmt.Errorf("error Multi %v", err)
	}

	if ttl > 0 {
		if err := c.Send("EXPIRE", key, ttl); err != nil {
			return []uint8{}, fmt.Errorf("error EXPIRE %v", err)
		}
		if err := c.Send("GET", key); err != nil {
			return []uint8{}, fmt.Errorf("error in GET getex %v", err)
		}

		str, err := redis.Values(c.Do("EXEC"))
		if err != nil || str[1] == nil {
			return []uint8{}, fmt.Errorf("error EXEC in GetEx %v key: %v", err, key)
		}
		log.Infof("GetEx took: %v", time.Since(s))
		return str[1].([]byte), nil
	}

	if err := c.Send("GET", key); err != nil {
		return []uint8{}, fmt.Errorf("error in GET getex %v", err)
	}

	str, err := redis.Values(c.Do("EXEC"))
	if err != nil || str[0] == nil {
		return []uint8{}, fmt.Errorf("error EXEC in GetEx %v key: %v", err, key)
	}
	for _, t := range str {
		log.Infof("GetEx value: %s", t)
	}

	log.Infof("GetEx took: %v", time.Since(s))
	return str[0].([]byte), nil

}

// Get 通过key获取值
func Get(key string) (string, error) {
	c := database.GetRedisCon()
	defer c.Close()

	str, err := redis.String(c.Do("GET", key))
	if err != nil {
		utils.ErrorLog("error Redis Get", err.Error())
		return "", err
	}

	return str, nil
}

// Delete 通过key删除数据
func Delete(key string) error {
	c := database.GetRedisCon()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		utils.ErrorLog("error Delete REdis key", err.Error())
		return err
	}

	return nil
}

// DelByPattern 删除多条数据
func DelByPattern(pattern ...string) error {
	s := time.Now()
	c := database.GetRedisCon()
	defer c.Close()

	keywords := JoinKeyWords(pattern...)
	keys, err := redis.Strings(c.Do("KEYS", keywords))
	if err != nil {
		return err
	}

	if err := c.Send("MULTI"); err != nil {
		utils.ErrorLog("error in DelByPattern MULTI", err.Error())
		return err
	}

	for _, key := range keys {
		if err := c.Send("DEL", key); err != nil {
			utils.ErrorLog("error EXEC", err.Error())
			continue
		}
	}

	if _, err := redis.Values(c.Do("EXEC")); err != nil {
		return fmt.Errorf("error EXEC %v", err)
	}
	log.Infof("DelByPattern took: %v keywords: %v keys: %v ", time.Since(s), keywords, keys)
	return nil
}

// JoinKeyWords 多关键字生成redis的键
func JoinKeyWords(keyword ...string) string {
	return strings.Join(keyword, DefaultKeySeperator)
}
