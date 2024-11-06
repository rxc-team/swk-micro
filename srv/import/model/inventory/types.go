package inventory

import (
	"rxcsoft.cn/pit3/srv/database/proto/field"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// checkReadParam 盘点文件读取参数
type checkReadParam struct {
	db          string // 数据库
	appID       string // appID
	domain      string // 域名
	datastoreID string // 台账ID
	jobID       string // 任务ID
	userID      string // 用户ID
	groupID     string // 用户组织ID
	encoding    string // 编码格式
	lang        string // 当前语言

	fileData   []string
	headerData []*field.Field
	allFields  []*field.Field
	langData   *language.Language // 语言文件
	allUsers   []*user.User
	mainKeys   map[string]struct{}

	checkType string
	checkedAt string
	checkedBy string

	index int
}

// 批量盘点基础参数
type Params struct {
	JobId        string
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
	MainKeys     []string
	CheckType    string
	CheckedAt    string
	CheckedBy    string
}
