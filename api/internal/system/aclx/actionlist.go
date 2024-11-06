package aclx

import "net/http"

type ActionData struct {
	Path   string
	Method string
}

var datastoreActionMap = map[string][]*ActionData{
	"contract_update": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/:i_id/contract",
			Method: http.MethodPut,
		},
	},
	"midway_cancel": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/:i_id/terminate",
			Method: http.MethodPut,
		},
	},
	"estimate_update": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/:i_id/debt",
			Method: http.MethodPut,
		},
	},
	"contract_expire": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/:i_id/contractExpire",
			Method: http.MethodPut,
		},
	},
	"pdf": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/print",
			Method: http.MethodPost,
		},
	},
	"clear": {
		{
			Path:   "/internal/api/v1/web/item/clear/datastores/:d_id/items",
			Method: http.MethodDelete,
		},
	},
	"group": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items",
			Method: http.MethodPatch,
		},
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/owners",
			Method: http.MethodPost,
		},
	},
	"history": {
		{
			Path:   "/internal/api/v1/web/history/datastores/:d_id/histories",
			Method: http.MethodGet,
		},
		{
			Path:   "/internal/api/v1/web/history/datastores/:d_id/download",
			Method: http.MethodGet,
		},
	},
	"read": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/search",
			Method: http.MethodPost,
		},
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/:i_id",
			Method: http.MethodGet,
		},
	},
	"insert": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items",
			Method: http.MethodPost,
		},
	},
	"update": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/:i_id",
			Method: http.MethodPut,
		},
	},
	"delete": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/:i_id",
			Method: http.MethodDelete,
		},
	},
	"mapping_upload": {
		{
			Path:   "/internal/api/v1/web/mapping/datastores/:d_id/upload",
			Method: http.MethodPost,
		},
	},
	"mapping_download": {
		{
			Path:   "/internal/api/v1/web/mapping/datastores/:d_id/download",
			Method: http.MethodPost,
		},
	},
	"image": {
		{
			Path:   "/internal/api/v1/web/item/import/image/datastores/:d_id/items",
			Method: http.MethodPost,
		},
	},
	"csv": {
		{
			Path:   "/internal/api/v1/web/item/import/csv/datastores/:d_id/items",
			Method: http.MethodPost,
		},
	},
	"inventory": {
		{
			Path:   "/internal/api/v1/web/item/import/csv/datastores/:d_id/check/items",
			Method: http.MethodPost,
		},
	},
	"principal_repayment": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/prs/download",
			Method: http.MethodPost,
		},
	},
	"data": {
		{
			Path:   "/internal/api/v1/web/item/datastores/:d_id/items/download",
			Method: http.MethodPost,
		},
	},
}

var reportActionMap = map[string][]*ActionData{
	"read": {
		{
			Path:   "/internal/api/v1/web/report/reports/:rp_id",
			Method: http.MethodGet,
		},
		{
			Path:   "/internal/api/v1/web/report/reports/:rp_id/data",
			Method: http.MethodPost,
		},
		{
			Path:   "/internal/api/v1/web/report/gen/reports/:rp_id/data",
			Method: http.MethodPost,
		},
		{
			Path:   "/internal/api/v1/web/report/reports/:rp_id/download",
			Method: http.MethodPost,
		},
	},
}

var docActionMap = map[string][]*ActionData{
	"read": {
		{
			Path:   "/internal/api/v1/web/file/folders/:fo_id/files",
			Method: http.MethodGet,
		},
		{
			Path:   "/internal/api/v1/web/file/download/folders/:fo_id/files/:file_id",
			Method: http.MethodGet,
		},
	},
	"write": {
		{
			Path:   "/internal/api/v1/web/file/folders/:fo_id/upload",
			Method: http.MethodPost,
		},
	},
	"delete": {
		{
			Path:   "/internal/api/v1/web/file/folders/:fo_id/files/:file_id",
			Method: http.MethodDelete,
		},
	},
}
