package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/global/model"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Language 语言
type Language struct{}

// log出力使用
const (
	LanguageProcessName = "Language"

	ActionFindLanguages              = "FindLanguages"
	ActionFindLanguage               = "FindLanguage"
	ActionFindLanguageValue          = "FindLanguageValue"
	ActionAddLanguage                = "AddLanguage"
	ActionAddLanguageData            = "AddLanguageData"
	ActionAddGroupLanguageData       = "AddGroupLanguageData"
	ActionAddCommonData              = "AddCommonData"
	ActionAddAppLanguageData         = "AddAppLanguageData"
	ActionAddManyLanData             = "AddManyLanData"
	ActionDeleteAppLanguageData      = "DeleteAppLanguageData"
	ActionDeleteGroupLanguageData    = "DeleteGroupLanguageData"
	ActionDeleteLanguageData         = "DeleteLanguageData"
	ActionAddResourceLanguageData    = "AddResourceLanguageData"
	ActionDeleteResourceLanguageData = "DeleteResourceLanguageData"
)

// FindLanguages 获取所有语言
func (l *Language) FindLanguages(ctx context.Context, req *language.FindLanguagesRequest, rsp *language.FindLanguagesResponse) error {
	utils.InfoLog(ActionFindLanguages, utils.MsgProcessStarted)

	languages, err := model.FindLanguages(req.GetDatabase(), req.GetDomain())
	if err != nil {
		utils.ErrorLog(ActionFindLanguages, err.Error())
		return err
	}

	res := &language.FindLanguagesResponse{}
	for _, lang := range languages {
		res.LanguageList = append(res.LanguageList, lang.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindLanguages, utils.MsgProcessEnded)
	return nil
}

// FindLanguage 获取一条语言的数据
func (l *Language) FindLanguage(ctx context.Context, req *language.FindLanguageRequest, rsp *language.FindLanguageResponse) error {
	utils.InfoLog(ActionFindLanguage, utils.MsgProcessStarted)

	res, err := model.FindLanguage(req.GetDatabase(), req.GetDomain(), req.GetLangCd())
	if err != nil {
		utils.ErrorLog(ActionFindLanguage, err.Error())
		return err
	}

	apps := make(map[string]*language.App, len(res.Apps))

	for key, value := range res.Apps {
		apps[key] = value.ToProto()
	}
	common := res.Common.ToProto()

	rsp.Apps = apps
	rsp.Common = common

	utils.InfoLog(ActionFindLanguage, utils.MsgProcessEnded)
	return nil
}

// FindLanguageValue 通过当前domain、langcd和对应的key，获取下面的语言结果
func (l *Language) FindLanguageValue(ctx context.Context, req *language.FindLanguageValueRequest, rsp *language.FindLanguageValueResponse) error {
	utils.InfoLog(ActionFindLanguageValue, utils.MsgProcessStarted)

	res, err := model.FindLanguageValue(req.GetDatabase(), req.GetDomain(), req.GetLangCd(), req.GetKey())
	if err != nil {
		utils.ErrorLog(ActionFindLanguageValue, err.Error())
		return err
	}

	rsp.Name = res

	utils.InfoLog(ActionFindLanguageValue, utils.MsgProcessEnded)
	return nil
}

// AddLanguage 添加一种语言
func (l *Language) AddLanguage(ctx context.Context, req *language.AddLanguageRequest, rsp *language.AddLanguageResponse) error {
	utils.InfoLog(ActionAddLanguage, utils.MsgProcessStarted)

	params := model.Language{
		Domain:    req.GetDomain(),
		LangCD:    req.GetLangCd(),
		Text:      req.GetText(),
		Abbr:      req.GetAbbr(),
		CreatedAt: time.Now(),
		CreatedBy: req.GetWriter(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	res, err := model.AddLanguage(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddLanguage, err.Error())
		return err
	}

	rsp.LangCd = res

	utils.InfoLog(ActionAddLanguage, utils.MsgProcessEnded)
	return nil
}

// AddLanguageData 添加语言数据
func (l *Language) AddLanguageData(ctx context.Context, req *language.AddLanguageDataRequest, rsp *language.Response) error {
	utils.InfoLog(ActionAddLanguageData, utils.MsgProcessStarted)

	err := model.AddLanguageData(req.GetDatabase(), req.GetDomain(), req.GetLangCd(), req.GetAppId(), req.GetAppName(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionAddLanguageData, err.Error())
		return err
	}

	utils.InfoLog(ActionAddLanguageData, utils.MsgProcessEnded)
	return nil
}

// AddAppLanguageData 添加APP语言数据
func (l *Language) AddAppLanguageData(ctx context.Context, req *language.AddAppLanguageDataRequest, rsp *language.Response) error {
	utils.InfoLog(ActionAddAppLanguageData, utils.MsgProcessStarted)

	params := model.LanguageParam{
		Domain: req.GetDomain(),
		LangCd: req.GetLangCd(),
		AppID:  req.GetAppId(),
		Type:   req.GetType(),
		Key:    req.GetKey(),
		Value:  req.GetValue(),
		Writer: req.GetWriter(),
	}

	err := model.AddAppLanguageData(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionAddAppLanguageData, err.Error())
		return err
	}

	utils.InfoLog(ActionAddAppLanguageData, utils.MsgProcessEnded)
	return nil
}

// AddManyLanData 添加或更新多条多语言数据
func (l *Language) AddManyLanData(ctx context.Context, req *language.AddManyLanDataRequest, rsp *language.Response) error {
	utils.InfoLog(ActionAddManyLanData, utils.MsgProcessStarted)

	var lans []*model.LanItem
	for _, l := range req.GetLans() {
		lan := &model.LanItem{
			AppID: l.GetAppId(),
			Type:  l.GetType(),
			Key:   l.GetKey(),
			Value: l.GetValue(),
		}
		lans = append(lans, lan)
	}

	params := model.ManyLanParam{
		Domain: req.GetDomain(),
		LangCd: req.GetLangCd(),
		Lans:   lans,
		Writer: req.GetWriter(),
	}

	err := model.AddManyLanData(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionAddManyLanData, err.Error())
		return err
	}

	utils.InfoLog(ActionAddManyLanData, utils.MsgProcessEnded)
	return nil
}

// DeleteAppLanguageData 删除App中的语言数据
func (l *Language) DeleteAppLanguageData(ctx context.Context, req *language.DeleteAppLanguageDataRequest, rsp *language.Response) error {
	utils.InfoLog(ActionDeleteAppLanguageData, utils.MsgProcessStarted)

	err := model.DeleteAppLanguageData(req.GetDatabase(), req.GetDomain(), req.GetAppId(), req.GetType(), req.GetKey(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteAppLanguageData, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteAppLanguageData, utils.MsgProcessStarted)
	return nil
}

// AddCommonData 添加共通语言数据
func (l *Language) AddCommonData(ctx context.Context, req *language.AddCommonDataRequest, rsp *language.Response) error {
	utils.InfoLog(ActionAddCommonData, utils.MsgProcessStarted)

	params := model.LanguageParam{
		Domain: req.GetDomain(),
		LangCd: req.GetLangCd(),
		Type:   req.GetType(),
		Key:    req.GetKey(),
		Value:  req.GetValue(),
		Writer: req.GetWriter(),
	}

	err := model.AddCommonData(req.GetDatabase(), params)
	if err != nil {
		utils.ErrorLog(ActionAddCommonData, err.Error())
		return err
	}

	utils.InfoLog(ActionAddCommonData, utils.MsgProcessEnded)
	return nil
}

// DeleteCommonData 删除共通中的语言数据
func (l *Language) DeleteCommonData(ctx context.Context, req *language.DeleteCommonDataRequest, rsp *language.Response) error {
	utils.InfoLog(ActionDeleteAppLanguageData, utils.MsgProcessStarted)

	err := model.DeleteCommonData(req.GetDatabase(), req.GetDomain(), req.GetType(), req.GetKey(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteAppLanguageData, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteAppLanguageData, utils.MsgProcessStarted)
	return nil
}

// DeleteLanguageData 删除语言数据
func (l *Language) DeleteLanguageData(ctx context.Context, req *language.DeleteLanguageDataRequest, rsp *language.Response) error {
	utils.InfoLog(ActionDeleteLanguageData, utils.MsgProcessStarted)

	err := model.DeleteLanguageData(req.GetDatabase(), req.GetDomain(), req.GetAppId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteLanguageData, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteLanguageData, utils.MsgProcessEnded)
	return nil
}
