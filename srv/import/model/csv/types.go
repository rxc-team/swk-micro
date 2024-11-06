package csv

import (
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// checkParam 检查参数
type checkParam struct {
	db          string
	wfID        string // 流程ID
	lang        string // 当前语言
	jobID       string // 任务ID
	encoding    string // 编码格式
	action      string
	apiKey      string
	domain      string
	groupID     string
	datastoreID string
	appID       string
	userID      string
	specialchar string
	firstMonth  string //比较开始期首月
	emptyChange bool
	fileData    rowData
	headerData  []string
	roles       []string
	owners      []string
	options     []*option.Option
	allUsers    []*user.User
	allFields   []*field.Field
	relations   []*datastore.RelationItem
	actions     []*permission.Action
	langData    *language.Language // 语言文件
	gpMap       map[string]string
	fileMap     map[string]string
}

type rowData struct {
	index int
	item  *item.Item
	data  []string
}

// 基础参数
type Params struct {
	JobId        string
	Action       string
	Encoding     string
	ZipCharset   string
	UserId       string
	AppId        string
	Lang         string
	Domain       string
	DatastoreId  string
	GroupId      string
	UpdateOwners []string
	Owners       []string
	Roles        []string
	WfId         string
	Database     string
	EmptyChange  bool
}

// 文件参数
type FileParams struct {
	FilePath    string
	ZipFilePath string
}

// FileValue 文件类型
type FileValue struct {
	URL  string `json:"url" bson:"url"`
	Name string `json:"name" bson:"name"`
}
