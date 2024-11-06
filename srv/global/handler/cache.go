package handler

import (
	"context"

	jsoncache "rxcsoft.cn/pit3/srv/global/json-cache"
	"rxcsoft.cn/pit3/srv/global/proto/cache"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Cache 缓存
type Cache struct{}

// log出力使用
const (
	CacheProcessName = "Cache"

	ActionSetCache    = "SetCache"
	ActionGetCache    = "GetCache"
	ActionDeleteCache = "DeleteCache"
)

// SetCache 设置缓存数据
func (ca *Cache) SetCache(ctx context.Context, req *cache.SetRequest, rsp *cache.Response) error {
	utils.InfoLog(ActionSetCache, utils.MsgProcessStarted)

	err := jsoncache.CacheSetEx(req.Value, req.Key, req.Ttl)
	if err != nil {
		utils.ErrorLog(ActionSetCache, err.Error())
		return err
	}

	utils.InfoLog(ActionSetCache, utils.MsgProcessEnded)
	return nil
}

// GetCache 获取缓存数据
func (ca *Cache) GetCache(ctx context.Context, req *cache.GetRequest, rsp *cache.GetResponse) error {
	utils.InfoLog(ActionGetCache, utils.MsgProcessStarted)

	str, err := jsoncache.CacheGetEx(req.Key, req.Ttl)
	if err != nil {
		utils.ErrorLog(ActionGetCache, err.Error())
		return err
	}

	rsp.Value = string(str)

	utils.InfoLog(ActionGetCache, utils.MsgProcessEnded)

	return nil
}

// DeleteCache 删除缓存数据
func (ca *Cache) DeleteCache(ctx context.Context, req *cache.DeleteRequest, rsp *cache.Response) error {
	utils.InfoLog(ActionDeleteCache, utils.MsgProcessStarted)

	err := jsoncache.DeleteCache(req.Key)
	if err != nil {
		utils.ErrorLog(ActionDeleteCache, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteCache, utils.MsgProcessEnded)

	return nil
}
