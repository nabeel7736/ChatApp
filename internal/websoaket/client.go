package websocket

import (
	"chatapp/internal/controllers"
	"encoding/json"
)

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// Unmarshal incoming JSON message
		var msg WSMessage
		json.Unmarshal(data, &msg)

		// Save to database
		controllers.SaveMessage(msg.Sender, msg.Content)

		// Broadcast to all users
		c.hub.broadcast <- data
	}
}
