package msg

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/global/proto/cache"
)

// MessageStore 使用redis作为消息服务器
type MessageStore struct {
	lang    string
	expired int64
}

// NewMessageStore 创建store
func NewMessageStore(lang string) *MessageStore {
	s := new(MessageStore)
	s.lang = lang
	s.expired = -1
	return s
}

// Set 插入数据
func (s *MessageStore) Set(value Msg) {
	cacheService := cache.NewCacheService("global", client.DefaultClient)

	for key, val := range value.Error {
		var req cache.SetRequest
		req.Key = []string{"message", s.lang, "error", key}
		req.Value = val
		req.Ttl = s.expired

		cacheService.SetCache(context.TODO(), &req)
	}
	for key, val := range value.Info {
		var req cache.SetRequest
		req.Key = []string{"message", s.lang, "info", key}
		req.Value = val
		req.Ttl = s.expired

		cacheService.SetCache(context.TODO(), &req)
	}
	for key, val := range value.Warn {
		var req cache.SetRequest
		req.Key = []string{"message", s.lang, "warn", key}
		req.Value = val
		req.Ttl = s.expired

		cacheService.SetCache(context.TODO(), &req)
	}
	for key, val := range value.Logger {
		var req cache.SetRequest
		req.Key = []string{"message", s.lang, "logger", key}
		req.Value = val
		req.Ttl = s.expired

		cacheService.SetCache(context.TODO(), &req)
	}
}

// Get 获取消息数据
func (s *MessageStore) Get(mtype, mkey string) string {
	cacheService := cache.NewCacheService("global", client.DefaultClient)

	var req cache.GetRequest
	req.Key = []string{"message", s.lang, mtype, mkey}
	req.Ttl = s.expired

	response, err := cacheService.GetCache(context.TODO(), &req)
	if err != nil {
		return ""
	}

	return response.GetValue()
}
