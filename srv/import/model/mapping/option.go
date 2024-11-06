package mapping

import (
	"context"
	"strings"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/import/common/langx"
)

// 根据名称获取对应的ID  TODO 没有考虑重复选项名称的问题
func GetOptionValueByName(db, appID string, data *language.Language, name string) string {

	apps := data.GetApps()
	if apps != nil {
		if appData, ok := apps[appID]; ok && appData != nil {
			// 有效选项存在检查
			optionMap := appData.GetOptions()
			for key, option := range optionMap {
				if option == name {
					result := strings.Split(key, "_")
					if len(result) > 1 {
						return key[len(result[0])+1:]
					}
				}
			}
		}
	}

	return ""
}

// GetOptionMap 获取选项
func GetOptionMap(db, appId, optionId string, langData *language.Language) (mp OptionMap) {

	optionService := option.NewOptionService("database", client.DefaultClient)

	var req option.FindOptionRequest
	// 从path中获取参数
	req.OptionId = optionId
	req.Invalid = "false"
	// 从共通中获取参数
	req.AppId = appId
	req.Database = db

	response, err := optionService.FindOption(context.TODO(), &req)
	if err != nil {
		return
	}

	mp = make(OptionMap)

	for _, opt := range response.GetOptions() {
		label := langx.GetLangValue(langData, opt.OptionLabel, langx.DefaultResult)
		mp[label] = opt.OptionValue
	}

	return

}
