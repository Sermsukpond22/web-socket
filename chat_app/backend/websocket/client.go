package websocket

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Client struct {
	UserID   uint
	Conn     *websocket.Conn
	Hub      *Hub
	writeMu  sync.Mutex
	isClosed bool
}

func NewClient(userID uint, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Hub:    hub,
	}
}

func (c *Client) WriteJSON(v interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.isClosed || c.Conn == nil {
		return nil
	}
	return c.Conn.WriteJSON(v)
}

func (c *Client) Close() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if !c.isClosed {
		c.isClosed = true
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	}
}
