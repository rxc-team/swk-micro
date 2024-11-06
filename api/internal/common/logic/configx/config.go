package configx

import (
	"context"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
)

// GetConfigVal 获取用户配置情报
func GetConfigVal(db, appID string) (cfg *app.Configs, err error) {
	configService := app.NewAppService("manage", client.DefaultClient)

	var req app.FindAppRequest
	req.AppId = appID
	req.Database = db

	response, err := configService.FindApp(context.TODO(), &req)
	if err != nil {
		return nil, err
	}

	return response.GetApp().GetConfigs(), nil
}
