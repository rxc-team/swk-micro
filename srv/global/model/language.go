package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/global/utils"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	LanguagesCollection = "languages"
)

// Language 语言
type Language struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Domain    string             `json:"domain" bson:"domain"`
	LangCD    string             `json:"lang_cd" bson:"lang_cd"`
	Text      string             `json:"text" bson:"text"`
	Abbr      string             `json:"abbr" bson:"abbr"`
	Apps      map[string]*App    `json:"apps" bson:"apps"`
	Common    Common             `json:"common" bson:"common"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy string             `json:"created_by" bson:"created_by"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy string             `json:"updated_by" bson:"updated_by"`
	DeletedAt time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy string             `json:"deleted_by" bson:"deleted_by"`
}

// App 应用程序
type App struct {
	AppName    string            `json:"app_name" bson:"app_name"`
	Datastores map[string]string `json:"datastores" bson:"datastores"`
	Fields     map[string]string `json:"fields" bson:"fields"`
	Queries    map[string]string `json:"queries" bson:"queries"`
	Reports    map[string]string `json:"reports" bson:"reports"`
	Dashboards map[string]string `json:"dashboards" bson:"dashboards"`
	Statuses   map[string]string `json:"statuses" bson:"statuses"`
	Options    map[string]string `json:"options" bson:"options"`
	Mappings   map[string]string `json:"mappings" bson:"mappings"`
	Workflows  map[string]string `json:"workflows" bson:"workflows"`
}

// Common 共通数据
type Common struct {
	Groups    map[string]string `json:"groups" bson:"groups"`
	Workflows map[string]string `json:"workflows" bson:"workflows"`
}

// LanguageParam 添加App语言数据用参数
type LanguageParam struct {
	Domain string
	LangCd string
	AppID  string
	Type   string
	Key    string
	Value  string
	Writer string
}

// LanItem 语言数据子项
type LanItem struct {
	AppID string
	Type  string
	Key   string
	Value string
}

// ManyLanParam 添加或更新多条多语言数据用参数
type ManyLanParam struct {
	Domain string
	LangCd string
	Lans   []*LanItem
	Writer string
}

// ToProto 转换为proto数据
func (a *App) ToProto() *language.App {
	return &language.App{
		AppName:    a.AppName,
		Datastores: a.Datastores,
		Fields:     a.Fields,
		Queries:    a.Queries,
		Reports:    a.Reports,
		Dashboards: a.Dashboards,
		Statuses:   a.Statuses,
		Options:    a.Options,
		Mappings:   a.Mappings,
		Workflows:  a.Workflows,
	}
}

// ToProto 转换为proto数据
func (c *Common) ToProto() *language.Common {
	return &language.Common{
		Workflows: c.Workflows,
		Groups:    c.Groups,
	}
}

// ToProto 转换为proto数据
func (l *Language) ToProto() *language.Language {

	apps := map[string]*language.App{}
	for key, app := range l.Apps {
		apps[key] = app.ToProto()
	}

	return &language.Language{
		LangCd:    l.LangCD,
		Text:      l.Text,
		Abbr:      l.Abbr,
		Domain:    l.Domain,
		Apps:      apps,
		Common:    l.Common.ToProto(),
		CreatedAt: l.CreatedAt.String(),
		CreatedBy: l.CreatedBy,
		UpdatedAt: l.UpdatedAt.String(),
		UpdatedBy: l.UpdatedBy,
		DeletedAt: l.DeletedAt.String(),
		DeletedBy: l.DeletedBy,
	}
}

// FindLanguages 所有语言
func FindLanguages(db, domain string) (l []Language, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain": domain,
	}

	var result []Language
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindLanguages", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var lan Language
		err := cur.Decode(&lan)
		if err != nil {
			utils.ErrorLog("error FindLanguages", err.Error())
			return nil, err
		}
		result = append(result, lan)
	}
	return result, nil
}

// FindLanguage 语言的多语言数据
func FindLanguage(db, domain, langCD string) (l Language, err error) {
	//连接数据库
	client := database.New()

	opts := options.Collection()
	opts.SetReadPreference(readpref.Primary())

	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection, opts)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	//参数
	query := bson.M{
		"lang_cd": langCD,
		"domain":  domain,
	}
	//json格式转换
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindLanguage", fmt.Sprintf("query: [ %s ]", queryJSON))

	//查询单条数据
	var result Language
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindLanguage", err.Error())
		return result, err
	}
	//返回单条语言数据
	return result, nil
}

// FindLanguageValue 通过当前domain、langcd和对应的key，获取下面的语言结果
func FindLanguageValue(db, domain, langCD, key string) (n string, err error) {
	//连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	//参数
	match := bson.M{
		"$match": bson.M{
			"lang_cd": langCD,
			"domain":  domain,
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":  0,
			"name": "$" + key,
		},
	}

	query := []bson.M{
		match,
		project,
	}

	//json格式转换
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindLanguageValue", fmt.Sprintf("query: [ %s ]", queryJSON))

	//查询单条数据

	type Result struct {
		Name string `bson:"name"`
	}
	var result []Result

	cur, err := c.Aggregate(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindLanguageValue", err.Error())
		return "", err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		utils.ErrorLog("error FindLanguageValue", err.Error())
		return "", err
	}

	if len(result) >= 1 {
		return result[0].Name, nil
	}

	//返回单条语言数据
	return "", nil
}

// AddLanguage 添加多语言
func AddLanguage(db string, l *Language) (langCD string, err error) {
	//连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	l.ID = primitive.NewObjectID()
	if l.Apps == nil {
		l.Apps = make(map[string]*App)
	}
	if l.Common.Groups == nil {
		l.Common.Groups = make(map[string]string)
	}

	queryJSON, _ := json.Marshal(l)
	utils.DebugLog("FindLanguage", fmt.Sprintf("Language: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, l)
	if err != nil {
		utils.ErrorLog("error AddLanguage", err.Error())
		return "", err
	}

	return l.LangCD, nil
}

// AddLanguageData 添加应用名称多语言数据
func AddLanguageData(db, domain, langCD, appID, appName, userID string) (err error) {
	//连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain":  domain,
		"lang_cd": langCD,
	}

	key := "apps." + appID + ".app_name"
	update := bson.M{
		"$set": bson.M{
			key:          appName,
			"updated_at": time.Now(),
			"updated_by": userID,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("AddLanguageData", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("AddLanguageData", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error AddLanguageData", err.Error())
		return err
	}

	return nil
}

// AddAppLanguageData 添加某应用下属项目一条多语言数据
func AddAppLanguageData(db string, data LanguageParam) (err error) {
	//连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	query := bson.M{
		"domain":  data.Domain,
		"lang_cd": data.LangCd,
	}

	key := strings.Builder{}
	key.WriteString("apps.")
	key.WriteString(data.AppID)
	key.WriteString(".")
	key.WriteString(data.Type)
	key.WriteString(".")
	key.WriteString(data.Key)

	update := bson.M{
		"$set": bson.M{
			key.String(): data.Value,
			"updated_at": time.Now(),
			"updated_by": data.Writer,
		},
	}
	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("AddAppLanguageData", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("AddAppLanguageData", fmt.Sprintf("update: [ %s ]", updateSON))

	opt := options.Update()
	opt.SetUpsert(false)

	_, err = c.UpdateOne(ctx, query, update, opt)
	if err != nil {
		utils.ErrorLog("error AddAppLanguageData", err.Error())
		return err
	}

	return nil
}

// AddCommonData 添加一条共通多语言数据
func AddCommonData(db string, data LanguageParam) (err error) {
	// 连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain":  data.Domain,
		"lang_cd": data.LangCd,
	}

	key := strings.Builder{}
	key.WriteString("common.")
	key.WriteString(data.Type)
	key.WriteString(".")
	key.WriteString(data.Key)

	update := bson.M{
		"$set": bson.M{
			key.String(): data.Value,
			"updated_at": time.Now(),
			"updated_by": data.Writer,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("AddCommonData", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("AddCommonData", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error AddCommonData", err.Error())
		return err
	}

	return nil
}

// AddManyLanData 添加或更新多条多语言数据
func AddManyLanData(db string, data ManyLanParam) (err error) {
	// 连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain":  data.Domain,
		"lang_cd": data.LangCd,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": data.Writer,
	}

	for _, lan := range data.Lans {
		if lan.Type == "apps" {
			key := "apps." + lan.AppID + ".app_name"
			change[key] = lan.Value
		} else {
			key := strings.Builder{}
			if lan.Type == "groups" {
				key.WriteString("common.")
				key.WriteString(lan.Type)
				key.WriteString(".")
				key.WriteString(lan.Key)
			} else {
				key.WriteString("apps.")
				key.WriteString(lan.AppID)
				key.WriteString(".")
				key.WriteString(lan.Type)
				key.WriteString(".")
				key.WriteString(lan.Key)
			}
			change[key.String()] = lan.Value
		}
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("AddManyLanData", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("AddManyLanData", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error AddManyLanData", err.Error())
		return err
	}

	return nil
}

// DeleteAppLanguageData 删除语言数据
func DeleteAppLanguageData(db, domain, appID, dataType, key, userID string) (err error) {
	//连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain": domain,
	}

	setKey := strings.Builder{}
	setKey.WriteString("apps.")
	setKey.WriteString(appID)
	setKey.WriteString(".")
	setKey.WriteString(dataType)
	setKey.WriteString(".")
	setKey.WriteString(key)

	update := bson.M{
		"$unset": bson.M{
			setKey.String(): 1,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": userID,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteAppLanguageData", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteAppLanguageData", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteAppLanguageData", err.Error())
		return err
	}

	return nil
}

// DeleteCommonData 删除共通语言数据
func DeleteCommonData(db, domain, dataType, key, userID string) (err error) {
	//连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain": domain,
	}

	setKey := strings.Builder{}
	setKey.WriteString("common.")
	setKey.WriteString(dataType)
	setKey.WriteString(".")
	setKey.WriteString(key)

	update := bson.M{
		"$unset": bson.M{
			setKey.String(): 1,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": userID,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteCommonData", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteCommonData", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateMany(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteCommonData", err.Error())
		return err
	}

	return nil
}

// DeleteLanguageData 删除app的语言数据
func DeleteLanguageData(db, domain, appID, userID string) (err error) {
	//连接数据库
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(LanguagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{
		"domain": domain,
	}

	setKey := strings.Builder{}
	setKey.WriteString("apps.")
	setKey.WriteString(appID)

	update := bson.M{
		"$unset": bson.M{
			setKey.String(): 1,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": userID,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteLanguageData", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteLanguageData", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateMany(ctx, query, update)

	if err != nil {
		utils.ErrorLog("error DeleteLanguageData", err.Error())
		return err
	}

	return nil
}
