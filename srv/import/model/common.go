package model

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/micro/go-micro/v2/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/import/common/containerx"
	"rxcsoft.cn/pit3/srv/import/common/filex"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/system/wsx"
	"rxcsoft.cn/pit3/srv/manage/proto/app"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/pit3/srv/task/utils"
	database "rxcsoft.cn/utils/mongo"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// 任务数据
type Task struct {
	JobId        string   `json:"job_id"`
	JobName      string   `json:"job_name,omitempty"`
	Origin       string   `json:"origin,omitempty"`
	UserId       string   `json:"user_id,omitempty"`
	ShowProgress bool     `json:"show_progress,omitempty"`
	Progress     int64    `json:"progress,omitempty"`
	StartTime    string   `json:"start_time,omitempty"`
	EndTime      string   `json:"end_time,omitempty"`
	Message      string   `json:"message,omitempty"`
	File         *File    `json:"file,omitempty"`
	ErrorFile    *File    `json:"error_file,omitempty"`
	TaskType     string   `json:"task_type,omitempty"`
	Steps        []string `json:"steps,omitempty"`
	CurrentStep  string   `json:"current_step,omitempty"`
	ScheduleId   string   `json:"schedule_id,omitempty"`
	AppId        string   `json:"app_id,omitempty"`
	Insert       int64    `json:"insert,omitempty"`
	Update       int64    `json:"update,omitempty"`
	Total        int64    `json:"total,omitempty"`
}

type File struct {
	Url  string `json:"url,omitempty"`
	Name string `json:"name,omitempty"`
}

const (
	// TimeFormat 日期格式化format
	TimeFormat = "2006-01-02 03:04:05"
	// DateFormat 日期格式化format
	DateFormat         = "2006-01-02"
	ActionValidSpecial = "ValidSpecial"
)

// Sequence 序列集合
type Sequence struct {
	ID            string `json:"id" bson:"_id"`
	SequenceValue int32  `json:"sequence_value" bson:"sequence_value"`
}

func ModifyTask(req task.ModifyRequest, userId string) error {
	taskService := task.NewTaskService("task", client.DefaultClient)
	_, err := taskService.ModifyTask(context.TODO(), &req)
	if err != nil {
		return err
	}

	var file File
	if req.File != nil {
		file.Url = req.File.Url
		file.Name = req.File.Name
	}

	var efile File
	if req.ErrorFile != nil {
		efile.Url = req.ErrorFile.Url
		efile.Name = req.ErrorFile.Name
	}

	tk := Task{
		JobId:       req.JobId,
		Progress:    req.Progress,
		EndTime:     req.EndTime,
		Message:     req.Message,
		File:        &file,
		ErrorFile:   &efile,
		CurrentStep: req.CurrentStep,
		Insert:      req.Insert,
		Update:      req.Update,
		Total:       req.Total,
	}

	content, err := json.Marshal(tk)
	if err != nil {
		loggerx.ErrorLog("ModifyTask", err.Error())
		return err
	}
	wsx.SendMsg(wsx.MessageParam{
		Sender:    "system",
		Recipient: userId,
		MsgType:   "job",
		Content:   string(content),
	})
	return nil
}

// // dateStandardFormat 日期字符串转指定格式日期字符串
// func DateStandardFormat(value string) (date string) {
// 	// 日期字符串为空
// 	if len(value) == 0 {
// 		return value
// 	}
// 	// yyyymmdd
// 	if len(value) == 8 {
// 		cast.ToTime()
// 		if _, err := strconv.ParseFloat(value, 64); err == nil {
// 			return value[0:4] + "-" + value[4:6] + "-" + value[6:8]
// 		}
// 	}
// 	// yyyy-mm-dd、yyyy-m-dd、yyyy-mm-d、yyyy-m-d
// 	// yyyy/mm/dd、yyyy/m/dd、yyyy/mm/d、yyyy/m/d
// 	// yyyy.mm.dd、yyyy.m.dd、yyyy.mm.d、yyyy.m.d
// 	if len(value) > 7 && len(value) < 11 {
// 		var ymdArr []string
// 		isRightDelimiter := true
// 		if strings.Contains(value, "-") {
// 			ymdArr = strings.Split(value, "-")
// 		} else if strings.Contains(value, "/") {
// 			ymdArr = strings.Split(value, "/")
// 		} else if strings.Contains(value, ".") {
// 			ymdArr = strings.Split(value, ".")
// 		} else {
// 			isRightDelimiter = false
// 		}

// 		// 合法年月日分隔符
// 		if isRightDelimiter {
// 			// 分割成年月日三份
// 			if len(ymdArr) == 3 {
// 				strY := ymdArr[0]
// 				strM := ymdArr[1]
// 				strD := ymdArr[2]
// 				// 年四位、月1~2位、日1~2位
// 				if len(strY) == 4 && len(strM) < 3 && len(strD) < 3 {
// 					// 月一位补位
// 					if len(strM) == 1 {
// 						strM = "0" + strM
// 					}
// 					// 日一位补位
// 					if len(strD) == 1 {
// 						strD = "0" + strD
// 					}
// 					return strY + "-" + strM + "-" + strD
// 				}
// 			}
// 		}
// 	}

// 	return ""
// }

// getConfig 获取用户配置情报
func GetConfig(db, appID string) (cfg *app.Configs, err error) {
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

// charCount 获取文本的字数
func CharCount(str string) int {
	r := []rune(str)
	return len(r)
}

// reTranUser 转换用户ID变更为用户名称
func ReTranUser(userName string, users []*user.User) string {
	for _, user := range users {
		if user.UserName == userName {
			return user.UserId
		}
	}

	return ""
}

// findField 根据ID查找字段
func FindField(fieldID string, fields []*field.Field) (r *field.Field, err error) {
	var reuslt *field.Field
	for _, f := range fields {
		if f.GetFieldId() == fieldID {
			reuslt = f
			break
		}
	}

	if reuslt == nil {
		return nil, fmt.Errorf("not found")
	}

	return reuslt, nil
}

// getUsers 获取所有用户数据
func GetUsers(db, app, domain string) (users []*user.User) {
	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUsersRequest
	// 从query中获取参数
	req.App = app
	// 从共通中获取参数
	req.Domain = domain
	req.Database = db

	response, err := userService.FindUsers(context.TODO(), &req)
	if err != nil {
		return users
	}

	return response.GetUsers()
}

// getLanguageData 获取所有用户数据
func GetLanguageData(db, langCd, domain string) (a *language.Language) {
	languageService := language.NewLanguageService("global", client.DefaultClient)

	var req language.FindLanguageRequest
	req.LangCd = langCd
	req.Domain = domain
	req.Database = db

	response, err := languageService.FindLanguage(context.TODO(), &req)
	if err != nil {
		return nil
	}

	return &language.Language{
		Apps:   response.GetApps(),
		Common: response.GetCommon(),
	}
}

// getFields 获取当前台账的字段
func GetFields(db, datastoreID, appID string, roles []string, showFile bool) []*field.Field {
	fieldService := field.NewFieldService("database", client.DefaultClient)

	var req field.FieldsRequest
	req.DatastoreId = datastoreID
	req.AppId = appID
	req.Database = db
	req.AsTitle = "false"

	response, err := fieldService.FindFields(context.TODO(), &req)
	if err != nil {
		return nil
	}
	pmService := permission.NewPermissionService("manage", client.DefaultClient)

	var preq permission.FindActionsRequest
	preq.RoleId = roles
	preq.PermissionType = "app"
	preq.AppId = appID
	preq.ActionType = "datastore"
	preq.ObjectId = datastoreID
	preq.Database = db
	pResp, err := pmService.FindActions(context.TODO(), &preq)
	if err != nil {
		return nil
	}

	set := containerx.New()
	for _, act := range pResp.GetActions() {
		if act.ObjectId == req.DatastoreId {
			set.AddAll(act.Fields...)
		}
	}
	fieldList := set.ToList()
	allFields := response.GetFields()
	var result []*field.Field
	for _, fieldID := range fieldList {
		f, err := FindField(fieldID, allFields)
		if err == nil {
			result = append(result, f)
		}
	}

	if showFile {
		// 排序
		sort.Sort(FieldList(result))
		return result
	}

	var fields []*field.Field
	// 去掉文件字段
	for _, f := range result {
		if f.GetFieldType() != "file" {
			fields = append(fields, f)
		}
	}

	// 排序
	sort.Sort(FieldList(fields))
	return fields
}

// getFile 从云端获取文件到本地，然后删除云端文件
func GetFile(domain, appID string, file string) error {

	if len(file) == 0 {
		return nil
	}

	minioClient, err := storagecli.NewClient(domain)
	if err != nil {
		return err
	}
	// 获取文件流
	object, err := minioClient.GetObject(file)
	if err != nil {
		return err
	}
	// 读取文件流
	var result []byte
	buffer := make([]byte, 1024)
	for {
		n, err := object.Read(buffer)
		result = append(result, buffer[:n]...)
		if err == io.EOF {
			break
		}
	}
	// 创建临时文件夹
	filex.Mkdir("public/app_" + appID + "/temp/")
	// 保存文件到本地的临时文件夹
	err = filex.SaveLocalFile(result, file)
	if err != nil {
		return err
	}
	// 删除云端文件
	err = minioClient.DeleteObject(file)
	if err != nil {
		return err
	}

	return nil
}

// excelToCsv 转换本地Excel到Csv
func ExcelToCsv(excelFilePath string) (csvFilePath string, err error) {
	csvTempPath := excelFilePath[:strings.LastIndex(excelFilePath, ".")+1] + "csv"

	// 读取对象EXCEL文件
	excelFile, err := excelize.OpenFile(excelFilePath)
	if err != nil {
		return "", err
	}

	// 设置读取对象sheet,默认Sheet1
	sheetDefault := "Sheet1"
	var sheetName, sheetFirst string
	for index, name := range excelFile.GetSheetMap() {
		if index == 1 {
			sheetFirst = name
		}
		if name == sheetDefault {
			sheetName = name
			break
		}
	}
	if sheetName == "" {
		sheetName = sheetFirst
	}

	// 读取行列内容
	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		return "", err
	}

	// 创建csv文件
	f, err := os.Create(csvTempPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 写入到csv文件
	writer := csv.NewWriter(f)
	rows[0][0] = "\xEF\xBB\xBF" + rows[0][0]
	writer.WriteAll(rows)
	writer.Flush()
	// 写完关闭csv临时文件
	f.Close()
	// 删除excel临时文件
	os.Remove(excelFilePath)

	return csvTempPath, nil
}

// getOptions 获取所有选项
func GetOptions(db, appID string) []*option.Option {
	optionService := option.NewOptionService("database", client.DefaultClient)

	var opReq option.FindOptionLabelsRequest
	// 从共通中获取参数
	opReq.Database = db
	opReq.AppId = appID

	opResponse, err := optionService.FindOptionLabels(context.TODO(), &opReq)
	if err != nil {
		return nil
	}

	return opResponse.GetOptions()
}

// 根据名称获取对应的ID
func GetOptionValue(group, name string, data *language.App) string {
	// 有效选项存在检查
	optionMap := data.GetOptions()
	for key, option := range optionMap {
		result := strings.Split(key, "_")
		groupID := result[0]
		if groupID == group && option == name {
			return strings.Join(result[1:], "")
		}
	}

	return ""
}

// 根据用户组名称获取对应的用户组ID
func GetGroupValue(name string, data map[string]string) string {
	for groupID, groupName := range data {
		if groupName == name {
			return groupID
		}
	}
	return ""
}

// 获取当前用户组及其下级用户组
func GetValidGroupsByID(pID string, groups []*group.Group) (child []*group.Group) {
	var validGroups []*group.Group

	var children []*group.Group
	for _, g := range groups {
		// 当前用户组情报
		if g.GroupId == pID {
			validGroups = append(validGroups, g)
		}
		// 当前用户组下级情报
		if g.ParentGroupId == pID {
			children = append(children, g)
		}
	}

	if len(children) == 0 {
		return validGroups
	}

	for _, cg := range children {
		validGroups = append(validGroups, GetValidGroupsByID(cg.GetGroupId(), groups)...)
	}

	return validGroups
}

// getSpecialChar 获取特殊字符
func GetSpecialChar(specialChars string) string {
	var specialchar string
	if len(specialChars) != 0 {
		// 编辑特殊字符
		for i := 0; i < len(specialChars); {
			specialchar += specialChars[i : i+1]
			i += 2
		}
	}
	return specialchar
}

// 验证特殊字符
func SpecialCheck(value string, special string) bool {

	if len(special) == 0 {
		return true
	}

	specialReg := regexp.QuoteMeta(special)
	// 判断特殊字符是否包含减号
	hasMinus := strings.Contains(specialReg, "-")
	if hasMinus {
		specialReg = strings.Replace(specialReg, "-", "\\-", 1)
	}
	re := regexp.MustCompile("[" + specialReg + "]")
	return !re.MatchString(value)
}

// GetNextSequenceValue 获取下个序列值
func GetNextSequenceValue(ctx context.Context, db, sequenceName string) (num int32, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("sequences")

	seq := createKeiyakunoSeq(db, sequenceName)
	if seq != nil {
		utils.ErrorLog("GetNextSequenceValue", seq.Error())
		return 0, seq
	}
	query := bson.M{
		"_id": sequenceName,
	}

	change := bson.M{
		"$inc": bson.M{
			"sequence_value": 1,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(1)
	var result Sequence
	if err := c.FindOneAndUpdate(ctx, query, change, opts).Decode(&result); err != nil {
		utils.ErrorLog("GetNextSequenceValue", err.Error())
		return result.SequenceValue, err
	}

	return result.SequenceValue, nil
}

func createKeiyakunoSeq(db, seqName string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection("sequences")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var s Sequence
	s.ID = seqName
	s.SequenceValue = 0

	var existing Sequence
	query := bson.M{"_id": seqName}
	seqErr := c.FindOne(ctx, query).Decode(&existing)
	if seqErr != nil {
		_, err = c.InsertOne(ctx, s)
		if err != nil {
			utils.ErrorLog("createKeiyakunoSeq", seqErr.Error())
			return err
		}
	}
	return nil
}
