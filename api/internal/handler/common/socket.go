package common

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/system/sessionx"
	"rxcsoft.cn/pit3/api/internal/system/wsx"
)

// WebSocket WebSocket
type WebSocket struct{}

// Ws is a websocket handler
func (w *WebSocket) Ws(c *gin.Context) {

	userID := c.Param("user_id")
	domain := sessionx.GetUserDomain(c)

	// change the reqest to websocket model
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if error != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	// websocket connect
	client := &wsx.Client{ID: userID, Domain: domain, Socket: conn, Send: make(chan []byte)}

	wsx.Manager.Register <- client

	go client.Read()
	go client.Write()
}

// Send is a websocket send message handler
func (w *WebSocket) Send(c *gin.Context) {
	var message wsx.Message
	// 从body中获取参数
	if err := c.BindJSON(&message); err != nil {
		httpx.GinHTTPError(c, "ws", err)
		return
	}

	if message.Recipient == "all user" {
		jsonMsg, _ := json.Marshal(message)
		domain := sessionx.GetUserDomain(c)
		wsx.Manager.Send(jsonMsg, domain)
	} else {
		jsonMsg, _ := json.Marshal(message)
		wsx.Manager.SendToUser(jsonMsg, message.Recipient)
	}

	c.JSON(200, httpx.Response{
		Status:  0,
		Message: "发送成功",
		Data:    nil,
	})
}
