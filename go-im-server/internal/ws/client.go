package ws

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan WSMessage
	Hub    *Hub
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg WSMessage
		if err := c.Conn.ReadJSON(&msg); err != nil {
			log.Printf("WS read error: %v", err)
			break
		}

		// Default msg.Type to "message" if not specified
		if msg.Type == "" {
			msg.Type = "message"
		}
		// Always set SenderID for security
		msg.SenderID = c.UserID
		c.Hub.Broadcast <- msg
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			log.Printf("WS write error: %v", err)
			break
		}
	}
}
