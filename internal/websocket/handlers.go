package websocket

import (
	"chatapp/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSMessage struct {
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func ServeWS(c *gin.Context, db *gorm.DB) {
	hub := GetHub()

	userIDInterface, _ := c.Get("claims")
	var userID uint
	var username string
	if userIDInterface != nil {
		claims := userIDInterface.(map[string]interface{})
		if uidf, ok := claims["user_id"].(float64); ok {
			userID = uint(uidf)
		}
		if uname, ok := claims["username"].(string); ok {
			username = uname
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

		// 1. Save to Database
		msgContent := string(msgBytes)
		dbMessage := models.Message{
			RoomID:   c.RoomID,
			SenderID: c.UserID,
			Content:  msgContent,
		}
		db.Create(&dbMessage)

		// 2. Prepare JSON for Broadcast (so frontend can display sender name)
		outMsg := WSMessage{
			Sender:    c.Username,
			Content:   msgContent,
			Timestamp: time.Now().Format("15:04"),
		}
		jsonBytes, _ := json.Marshal(outMsg)

		// 3. Send to Hub with RoomID
		hub.Broadcast <- BroadcastMessage{
			RoomID:  c.RoomID,
			Content: jsonBytes,
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
