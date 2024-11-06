package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/mongox"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/script"
)

// MongoScript mongodb脚本执行
type MongoScript struct{}

// log出力使用
const (
	MongoScriptProcessName = "MongoScript"
	ActionRun              = "Run"
	MONGO_SCRIPT_SERVER    = "MONGO_SCRIPT_SERVER"
)

// Run 执行appScript
// @Router /allows [get]
func (f *MongoScript) Run(c *gin.Context) {
	loggerx.InfoLog(c, ActionRun, loggerx.MsgProcessStarted)

	type Request struct {
		URI      string            `json:"uri"`
		ScriptId string            `json:"script_id"`
		Script   string            `json:"script"`
		Data     map[string]string `json:"data"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionRun, err)
		return
	}

	if len(req.ScriptId) > 0 {
		scriptService := script.NewScriptService("manage", client.DefaultClient)

		var freq script.FindScriptJobRequest
		freq.ScriptId = req.ScriptId
		freq.Database = sessionx.GetUserCustomer(c)

		response, err := scriptService.FindScriptJob(context.TODO(), &freq)
		if err != nil {
			httpx.GinHTTPError(c, ActionRun, err)
			return
		}

		scp := response.GetScriptJob()

		// 设置版本号
		v := os.Getenv("VERSION")
		if len(v) == 0 {
			v = "1.0.0"
		}

		if scp.GetScriptVersion() != v {
			httpx.GinHTTPError(c, ActionRun, fmt.Errorf("バージョンが一致しないため、スクリプトを実行できません。 サーバーバージョン%s、スクリプトバージョン%s", v, scp.GetScriptVersion()))
			return
		}

		var sreq script.StartRequest
		// 从path中获取参数
		sreq.ScriptId = req.ScriptId
		// 当前Script为更新者
		sreq.Writer = sessionx.GetAuthUserID(c)
		sreq.Database = sessionx.GetUserCustomer(c)

		_, err = scriptService.StartScriptJob(context.TODO(), &sreq)
		if err != nil {
			httpx.GinHTTPError(c, ActionRun, err)
			return
		}
	}

	req.URI = mongox.GetURI()

	client := resty.New()

	nReq := client.R()

	userID := sessionx.GetAuthUserID(c)

	nReq.SetBody(req)

	server := os.Getenv(MONGO_SCRIPT_SERVER)
	if len(server) == 0 {
		server = "http://localhost:8000"
	}

	result, err := nReq.Post(server + "/run/" + userID)
	if err != nil {
		httpx.GinHTTPError(c, ActionRun, err)
		return
	}

	print(string(result.Body()))

	loggerx.InfoLog(c, ActionRun, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I003, fmt.Sprintf(httpx.Temp, MongoScriptProcessName, ActionRun)),
		Data:    gin.H{},
	})
}
