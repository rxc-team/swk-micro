package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"rxcsoft.cn/pit3/api/system/common/httpx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/utils/redisx"
)

type SystemInfo struct{}

const (
	statusProcessName      = "System"
	ActionGetReleaseStatus = "GetReleaseStatus"
	ActionSetStatusAndIP   = "SetStatusAndIP"
	ActionGetStatusAndIP   = "GetStatusAndIP"
)

type SystemConfigs struct {
	IPList       []string `json:"ip_list"`
	Status       string   `json:"status"`
	MaintSummary string   `json:"maint_summary"`
	MaintPeriod  string   `json:"maint_period"`
	MaintRemark  string   `json:"maint_remark"`
}

// 获取系统的更新状态
func (s *SystemInfo) GetReleaseStatus(c *gin.Context) {
	rdb := redisx.New()
	ctx := c.Request.Context()
	// 为true时，允许登录，为false时，不允许登录
	release := false
	count, err := rdb.Exists(ctx, "SystemStatus").Result()
	if err != nil {
		httpx.GinHTTPError(c, ActionGetReleaseStatus, err)
		return
	}
	ipExits, err := rdb.SIsMember(ctx, "SystemUserIP", c.ClientIP()).Result()
	if err != nil {
		httpx.GinHTTPError(c, ActionGetReleaseStatus, err)
		return
	}
	if count > 0 {
		if ipExits {
			release = false
		} else {
			release = true
		}
	}

	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, statusProcessName, ActionGetReleaseStatus)),
		Data:    release,
	})
}

// 设置全局的系统状态和更新时可操作系统的ip
func (s *SystemInfo) SetStatusAndIP(c *gin.Context) {
	var config SystemConfigs
	rdb := redisx.New()
	ctx := c.Request.Context()
	rdbPipe := rdb.Pipeline()

	err := c.BindJSON(&config)
	if err != nil {
		httpx.GinHTTPError(c, ActionSetStatusAndIP, err)
		return
	}

	// 重新设置IP白名单
	rdbPipe.Del(ctx, "SystemUserIP")
	if len(config.IPList) > 0 {
		rdbPipe.SAdd(ctx, "SystemUserIP", config.IPList)
	}
	// 重新设置系统状态更新提示情报
	rdbPipe.Del(ctx, "MaintSummary")
	if len(config.MaintSummary) > 0 {
		rdbPipe.MSet(ctx, "MaintSummary", config.MaintSummary)
	}
	rdbPipe.Del(ctx, "MaintPeriod")
	if len(config.MaintPeriod) > 0 {
		rdbPipe.MSet(ctx, "MaintPeriod", config.MaintPeriod)
	}
	rdbPipe.Del(ctx, "MaintRemark")
	if len(config.MaintRemark) > 0 {
		rdbPipe.MSet(ctx, "MaintRemark", config.MaintRemark)
	}

	if config.Status == "true" {
		// 设置或更新系统状态
		rdbPipe.MSet(ctx, "SystemStatus", "true")
		_, err = rdbPipe.Exec(ctx)
		if err != nil {
			httpx.GinHTTPError(c, ActionSetStatusAndIP, err)
			return
		}

		c.JSON(200, httpx.Response{
			Status:  0,
			Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, statusProcessName, ActionSetStatusAndIP)),
			Data:    gin.H{},
		})

		return
	}

	rdbPipe.Del(ctx, "SystemStatus")
	_, err = rdbPipe.Exec(ctx)
	if err != nil {
		httpx.GinHTTPError(c, ActionSetStatusAndIP, err)
		return
	}

	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, statusProcessName, ActionSetStatusAndIP)),
		Data:    gin.H{},
	})
}

// 获取全局的系统状态和更新时可操作系统的ip
func (s *SystemInfo) GetStatusAndIP(c *gin.Context) {
	rdb := redisx.New()
	ctx := c.Request.Context()

	status, err := rdb.Get(ctx, "SystemStatus").Result()
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			status = "false"
		} else {
			httpx.GinHTTPError(c, ActionGetStatusAndIP, err)
			return
		}
	}

	summary, err := rdb.Get(ctx, "MaintSummary").Result()
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			summary = ""
		} else {
			httpx.GinHTTPError(c, ActionGetStatusAndIP, err)
			return
		}
	}

	period, err := rdb.Get(ctx, "MaintPeriod").Result()
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			period = ""
		} else {
			httpx.GinHTTPError(c, ActionGetStatusAndIP, err)
			return
		}
	}

	remark, err := rdb.Get(ctx, "MaintRemark").Result()
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			remark = ""
		} else {
			httpx.GinHTTPError(c, ActionGetStatusAndIP, err)
			return
		}
	}

	ipList, err := rdb.SMembers(ctx, "SystemUserIP").Result()
	if err != nil {
		httpx.GinHTTPError(c, ActionGetStatusAndIP, err)
		return
	}

	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I005, fmt.Sprintf(httpx.Temp, statusProcessName, ActionGetStatusAndIP)),
		Data: gin.H{"status": status,
			"ip_list":       ipList,
			"maint_summary": summary,
			"maint_period":  period,
			"maint_remark":  remark},
	})
}
