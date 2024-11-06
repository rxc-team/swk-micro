package wsx

import (
	"time"

	"github.com/gorilla/websocket"
)

// ClientManager is a websocket manager
type ClientManager struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// Client is a websocket client
type Client struct {
	ID     string
	Domain string
	Socket *websocket.Conn
	Send   chan []byte
}

// Message is an object for websocket message which is mapped to json type
type Message struct {
	MessageID string    `json:"message_id,omitempty"`
	Sender    string    `json:"sender,omitempty"`
	Recipient string    `json:"recipient,omitempty"`
	Domain    string    `json:"domain,omitempty"`
	MsgType   string    `json:"msg_type,omitempty"`
	Code      string    `json:"code,omitempty"`
	Link      string    `json:"link,omitempty"`
	Content   string    `json:"content,omitempty"`
	Status    string    `json:"status,omitempty"`
	Object    string    `json:"object,omitempty"`
	SendTime  time.Time `json:"send_time,omitempty"`
}

// Manager define a ws server manager
var Manager = ClientManager{
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[*Client]bool),
}

// Start is to start a ws server
func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-manager.Register:
			manager.Clients[conn] = true
		case conn := <-manager.Unregister:
			if _, ok := manager.Clients[conn]; ok {
				close(conn.Send)
				delete(manager.Clients, conn)
			}
		case message := <-manager.Broadcast:
			for conn := range manager.Clients {
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
					delete(manager.Clients, conn)
				}
			}
		}
	}
}

// Send is to send ws message to ws client
func (manager *ClientManager) Send(message []byte, domain string) {
	for conn := range manager.Clients {
		if conn.Domain == domain {
			conn.Send <- message
		}
	}
}

// SendToUser is to send ws message to User
func (manager *ClientManager) SendToUser(message []byte, userID string) {
	for conn := range manager.Clients {
		if conn.ID == userID {
			conn.Send <- message
		}
	}
}

func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
		c.Socket.Close()
	}()

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			Manager.Unregister <- c
			c.Socket.Close()
			break
		}

		Manager.Broadcast <- message
	}
}

func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}
