package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for dev. Lock this down for production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(c *gin.Context) {
	hub := GetHub()

	// Expect JWT token as query param token=... OR passed via Authorization header
	// For demo, we'll accept user_id param to identify user (in production parse JWT)
	userIDInterface, _ := c.Get("claims")
	var userID uint
	if userIDInterface != nil {
		claims := userIDInterface.(map[string]interface{})
		if uidf, ok := claims["user_id"].(float64); ok {
			userID = uint(uidf)
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	client := &Client{
		Conn:   conn,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	hub.Register <- client

	go client.writer()
	client.reader(hub)
}

func (c *Client) reader(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		// broadcast to all (for demo). You should enrich messages with JSON that includes room id etc.
		hub.Broadcast <- msg
	}
}

func (c *Client) writer() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// hub closed channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			// send ping
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
