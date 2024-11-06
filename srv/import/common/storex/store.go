package storex

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/global/proto/cache"
)

// RedisStore 使用redis作为保存验证码的服务器
type RedisStore struct {
	// 过期时间（以秒为单位，超时自动删除数据）.
	expired int64
}

// NewRedisStore 创建store
func NewRedisStore(expired int64) *RedisStore {
	s := new(RedisStore)
	s.expired = expired
	return s
}

// Set 插入数据
func (s *RedisStore) Set(id string, value string) {
	cacheService := cache.NewCacheService("global", client.DefaultClient)

	valueMap := map[string]string{
		id: value,
	}

	str, err := json.Marshal(valueMap)
	if err != nil {
		fmt.Printf("value marshal has error: %v", err)
		return
	}

	var req cache.SetRequest
	req.Key = []string{"captcha", id}
	req.Value = string(str)
	req.Ttl = s.expired

	cacheService.SetCache(context.TODO(), &req)
}

// Get 获取数据
func (s *RedisStore) Get(id string, clear bool) string {
	cacheService := cache.NewCacheService("global", client.DefaultClient)

	var req cache.GetRequest
	req.Key = []string{"captcha", id}
	req.Ttl = s.expired

	response, err := cacheService.GetCache(context.TODO(), &req)
	if err != nil {
		return ""
	}

	var valueMap map[string]string

	e := json.Unmarshal([]byte(response.GetValue()), &valueMap)
	if e != nil {
		return ""
	}

	if clear {
		var del cache.DeleteRequest
		del.Key = []string{"captcha", id}
		_, err := cacheService.DeleteCache(context.TODO(), &del)
		if err != nil {
			fmt.Printf("delete key has error: %v", err)
		}
	}

	return valueMap[id]
}

// Verify 验证数据
func (s *RedisStore) Verify(id, answer string, clear bool) bool {
	result := s.Get(id, clear)
	if result == answer {
		return true
	}

	return false
}
