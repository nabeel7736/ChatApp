package websocket

import (
	"chatapp/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSMessage struct {
	Type      string `json:"type"`
	ID        uint   `json:"id,omitempty"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type IncomingMessage struct {
	Type    string `json:"type"` // "chat" or "delete"
	Content string `json:"content,omitempty"`
	MsgID   uint   `json:"msg_id,omitempty"` // ID of message to delete
}

func ServeWS(c *gin.Context, db *gorm.DB) {
	hub := GetHub()

	userIDInterface, _ := c.Get("claims")
	var userID uint
	var username string

	if userIDInterface != nil {
		if claims, ok := userIDInterface.(jwt.MapClaims); ok {
			if uidf, ok := claims["user_id"].(float64); ok {
				userID = uint(uidf)
			}
			if uname, ok := claims["username"].(string); ok {
				username = uname
			}
		}
	}

	roomIDStr := c.Query("room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil || roomID == 0 {
		log.Println("Invalid room ID")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID,
		Username: username,
		RoomID:   uint(roomID),
		Send:     make(chan []byte, 256),
	}

	hub.Register <- client

	go client.writer()
	client.reader(hub, db)
}

func (c *Client) reader(hub *Hub, db *gorm.DB) {
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
		_, msgBytes, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var inMsg IncomingMessage
		if err := json.Unmarshal(msgBytes, &inMsg); err != nil {
			log.Println("Invalid JSON:", err)
			continue
		}

		if inMsg.Type == "delete" {
			// Handle Deletion logic
			var msg models.Message
			if err := db.First(&msg, inMsg.MsgID).Error; err == nil {
				if msg.SenderID == c.UserID {
					db.Delete(&msg)
					outMsg := WSMessage{
						Type: "delete",
						ID:   msg.ID,
					}
					jsonBytes, _ := json.Marshal(outMsg)
					hub.Broadcast <- BroadcastMessage{
						RoomID:  c.RoomID,
						Content: jsonBytes,
					}
				}
			}
		} else {
			dbMessage := models.Message{
				RoomID:   c.RoomID,
				SenderID: c.UserID,
				Content:  inMsg.Content,
			}
			db.Create(&dbMessage)

			outMsg := WSMessage{
				Type:      "chat",
				ID:        dbMessage.ID,
				Sender:    c.Username,
				Content:   inMsg.Content,
				Timestamp: time.Now().Format("15:04"),
			}
			jsonBytes, _ := json.Marshal(outMsg)

			hub.Broadcast <- BroadcastMessage{
				RoomID:  c.RoomID,
				Content: jsonBytes,
			}
		}
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
