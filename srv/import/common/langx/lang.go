package langx

import (
	"context"
	"strings"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
)

const DefaultResult string = "(no translate)"

// GetLangData 通过key获取对应的语言数据
func GetLangData(db, domain, langCd, key string) string {
	languageService := language.NewLanguageService("global", client.DefaultClient)
	// 获取当前语言数据
	var req language.FindLanguageValueRequest
	req.LangCd = langCd
	req.Domain = domain
	req.Key = key
	req.Database = db

	response, err := languageService.FindLanguageValue(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("GetLangData", err.Error())
		return DefaultResult
	}
	// 返回当前APP的语言数据
	return response.GetName()
}

// 获取所有用户数据
func GetLanguageData(db, langCd, domain string) (a *language.Language) {
	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.LangCd = langCd
	req.Domain = domain
	req.Database = db

	response, err := languageService.FindLanguage(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("GetLanguageData", err.Error())
		return nil
	}

	return &language.Language{
		Apps:   response.GetApps(),
		Common: response.GetCommon(),
	}
}

func DeleteAppLanguageData(db, domain, appId, types, key string) error {
	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.DeleteAppLanguageDataRequest

	req.Domain = domain
	req.AppId = appId
	req.Type = types
	req.Key = key
	req.Database = db

	_, err := languageService.DeleteAppLanguageData(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("DeleteAppLanguageData", err.Error())
		return err
	}

	return nil
}

func GetLangValue(l *language.Language, key, defaultValue string) string {
	if len(key) == 0 || l == nil {
		return ""
	}

	// Apps
	if strings.HasPrefix(key, "apps") {
		keys := strings.Split(key, ".")
		apps := l.GetApps()
		if len(keys) > 0 && apps != nil {
			if appData, ok := apps[keys[1]]; ok && appData != nil {
				// 如果是app名字的场合
				if len(keys) == 3 && keys[2] == "app_name" {
					return appData.GetAppName()
				}
				// 其他场合
				ttype := keys[2]
				tkey := keys[3]
				switch ttype {
				case "datastores":
					if value, ok := appData.GetDatastores()[tkey]; ok {
						return value
					}
				case "fields":
					if value, ok := appData.GetFields()[tkey]; ok {
						return value
					}
				case "options":
					if value, ok := appData.GetOptions()[tkey]; ok {
						return value
					}
				case "reports":
					if value, ok := appData.GetReports()[tkey]; ok {
						return value
					}
				case "dashboards":
					if value, ok := appData.GetDashboards()[tkey]; ok {
						return value
					}
				case "mappings":
					if value, ok := appData.GetMappings()[tkey]; ok {
						return value
					}
				case "workflows":
					if value, ok := appData.GetWorkflows()[tkey]; ok {
						return value
					}
				}
			}
		}
	}

	// Common
	if strings.HasPrefix(key, "common") {
		keys := strings.Split(key, ".")
		common := l.GetCommon()
		if len(keys) > 0 && common != nil {
			ttype := keys[1]
			tkey := keys[2]
			switch ttype {
			case "groups":
				if value, ok := common.GetGroups()[tkey]; ok {
					return value
				}
			}
		}
	}

	return defaultValue
}

// GetAppKey 获取台账翻译key
func GetAppKey(appId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".app_name")
	return key.String()
}

// GetDatastoreKey 获取台账翻译key
func GetDatastoreKey(appId, datastoreId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".datastores.")
	key.WriteString(datastoreId)
	return key.String()
}

// GetReportKey 获取报表翻译key
func GetReportKey(appId, reportId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".reports.")
	key.WriteString(reportId)
	return key.String()
}

// GetOptionGroupKey 获取选项组翻译key
func GetOptionGroupKey(appId, groupId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".options.")
	key.WriteString(groupId)
	return key.String()
}

// GetOptionKey 获取选项翻译key
func GetOptionKey(appId, groupId, value string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".options.")
	key.WriteString(groupId)
	key.WriteString("_")
	key.WriteString(value)
	return key.String()
}

// GetFieldKey 获取字段翻译key
func GetFieldKey(appId, datastoreId, fieldId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".fields.")
	key.WriteString(datastoreId)
	key.WriteString("_")
	key.WriteString(fieldId)
	return key.String()
}

// GetWorkflowKey 获取流程翻译key
func GetWorkflowKey(appId, wfId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".workflows.")
	key.WriteString(wfId)
	return key.String()
}

// GetWorkflowMenuKey 获取流程菜单翻译key
func GetWorkflowMenuKey(appId, wfId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".workflows.")
	key.WriteString("menu_")
	key.WriteString(wfId)
	return key.String()
}

// GetMappingKey 获取映射翻译key
func GetMappingKey(appId, datastoreId, mappingId string) string {
	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(appId)
	key.WriteString(".mappings.")
	key.WriteString(datastoreId)
	key.WriteString("_")
	key.WriteString(mappingId)
	return key.String()
}

// GetGroupKey 获取组织翻译key
func GetGroupKey(groupId string) string {
	key := strings.Builder{}
	key.WriteString("common.groups.")
	key.WriteString(groupId)
	return key.String()
}

// 获取旧多语言集合map
func GetLansMap(db, domain, langCd string) (lms map[string]string) {
	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.Database = db
	req.Domain = domain
	req.LangCd = langCd

	response, err := languageService.FindLanguage(context.TODO(), &req)
	if err != nil {
		return nil
	}

	res := make(map[string]string)
	for appID, app := range response.GetApps() {
		key := "apps." + appID + ".app_name"
		res[key] = app.GetAppName()

		for k, v := range app.GetDatastores() {
			key := "apps." + appID + ".datastores." + k
			res[key] = v
		}
		for k, v := range app.GetFields() {
			key := "apps." + appID + ".fields." + k
			res[key] = v
		}
		for k, v := range app.GetQueries() {
			key := "apps." + appID + ".queries." + k
			res[key] = v
		}
		for k, v := range app.GetReports() {
			key := "apps." + appID + ".reports." + k
			res[key] = v
		}
		for k, v := range app.GetDashboards() {
			key := "apps." + appID + ".dashboards." + k
			res[key] = v
		}
		for k, v := range app.GetStatuses() {
			key := "apps." + appID + ".statuses." + k
			res[key] = v
		}
		for k, v := range app.GetOptions() {
			key := "apps." + appID + ".options." + k
			res[key] = v
		}
		for k, v := range app.GetMappings() {
			key := "apps." + appID + ".mappings." + k
			res[key] = v
		}
		for k, v := range app.GetWorkflows() {
			key := "apps." + appID + ".workflows." + k
			res[key] = v
		}
	}
	for k, v := range response.GetCommon().GetGroups() {
		key := "common" + ".groups." + k
		res[key] = v
	}

	return res
}

func GetOldName(key string, lms map[string]string) string {
	if _, ok := lms[key]; ok {
		return lms[key]
	}

	return ""
}

func GetLanKey(lan *language.LanData) string {
	rk := ""
	if lan.Type == "apps" {
		rk = "apps." + lan.AppId + ".app_name"
	} else {
		key := strings.Builder{}
		if lan.Type == "groups" {
			key.WriteString("common.")
			key.WriteString(lan.Type)
			key.WriteString(".")
			key.WriteString(lan.Key)
		} else {
			key.WriteString("apps.")
			key.WriteString(lan.AppId)
			key.WriteString(".")
			key.WriteString(lan.Type)
			key.WriteString(".")
			key.WriteString(lan.Key)
		}
		rk = key.String()
	}

	return rk
}
