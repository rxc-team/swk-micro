package lease

import (
	"rxcsoft.cn/pit3/srv/database/proto/datastore"
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/database/proto/item"
	"rxcsoft.cn/pit3/srv/database/proto/option"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/permission"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

type data [][]string

// checkParam 检查参数
type checkParam struct {
	db          string
	lang        string // 当前语言
	jobID       string // 任务ID
	encoding    string // 编码格式
	action      string
	domain      string
	groupID     string
	datastoreID string
	appID       string
	userID      string
	firstMonth  string //比较开始期首月度
	handleMonth string // 处理年月度
	beginMonth  string // 期首月度
	smallAmount string // 少额租赁范围
	shortPeriod string // 短期租赁范围
	specialchar string
	fileData    []string
	headerData  []string
	roles       []string
	owners      []string
	allUsers    []*user.User
	allFields   []*field.Field
	options     []*option.Option
	actions     []*permission.Action
	relations   []*datastore.RelationItem
	langData    *language.Language // 语言文件
	payData     map[string]PayData
	dsMap       map[string]string
	gpMap       map[string]string
}

type attachData []*item.AttachItems

// 基础参数
type Params struct {
	JobId        string
	Action       string
	Encoding     string
	UserId       string
	AppId        string
	Lang         string
	Domain       string
	DatastoreId  string
	GroupId      string
	UpdateOwners []string
	Owners       []string
	Roles        []string
	Database     string
	FirstMonth   string
}

// 文件参数
type FileParams struct {
	FilePath    string
	PayFilePath string
}
