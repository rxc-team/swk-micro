package mapping

// 基础参数
type Params struct {
	MappingID    string
	JobId        string
	UserId       string
	AppId        string
	Lang         string
	Domain       string
	DatastoreId  string
	UpdateOwners []string
	Owners       []string
	Roles        []string
	Database     string
	EmptyChange  bool
}

// 文件参数
type FileParams struct {
	FilePath    string
	PayFilePath string
}
