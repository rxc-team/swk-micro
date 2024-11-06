package eventx

import (
	"github.com/micro/go-micro/v2/broker"
	"rxcsoft.cn/pit3/api/internal/system/aclx"
)

// LoadCasbinPolicy 更新用户和角色的关系
func LoadCasbinPolicy(event broker.Event) error {
	// 添加任务
	acl := aclx.GetCasbin()
	// 先删除该用户的所有角色关系
	acl.LoadPolicy()
	event.Ack()

	return nil
}
