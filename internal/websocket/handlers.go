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
	SenderPic string `json:"sender_pic"`
	Content   string `json:"content"`
	MediaType string `json:"media_type"`
	Timestamp string `json:"timestamp"`
}

type IncomingMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	MsgID     uint   `json:"msg_id,omitempty"`
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
	} else {
		tokenStr := c.Query("token")
		if tokenStr != "" {
			jwtSecret := []byte("secret123")

			tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err == nil && tok != nil {
				if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid {
					if uidf, ok := claims["user_id"].(float64); ok {
						userID = uint(uidf)
					}
					if uname, ok := claims["username"].(string); ok {
						username = uname
					}
				}
			} else {
				log.Println("invalid token in query:", err)
			}
		}
	}

	roomIDStr := c.Query("room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil || roomID == 0 {
		log.Println("Invalid room ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		log.Println("User not found for WS")
		return
	}

	client := &Client{
		Conn:       conn,
		UserID:     userID,
		Username:   username,
		ProfilePic: user.ProfilePic,
		RoomID:     uint(roomID),
		Send:       make(chan []byte, 256),
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

		switch inMsg.Type {
		case "delete":
			var msg models.Message
			if result := db.Limit(1).Find(&msg, inMsg.MsgID); result.RowsAffected > 0 {
				if msg.SenderID == c.UserID {
					db.Delete(&msg)
					outMsg := WSMessage{
						Type: "delete",
						ID:   msg.ID,
					}
					jsonBytes, _ := json.Marshal(outMsg)
					hub.Broadcast <- BroadcastMessage{RoomID: c.RoomID, Content: jsonBytes}
				}
			}

		case "edit":
			var msg models.Message
			if result := db.Limit(1).Find(&msg, inMsg.MsgID); result.RowsAffected > 0 {
				if msg.SenderID == c.UserID {
					// 2. Check 10-Minute Time Limit
					if time.Since(msg.CreatedAt) <= 10*time.Minute {
						msg.Content = inMsg.Content
						db.Save(&msg)

						outMsg := WSMessage{
							Type:      "edit",
							ID:        msg.ID,
							Content:   msg.Content,
							Sender:    c.Username,
							SenderPic: c.ProfilePic,
							MediaType: msg.MediaType,
							Timestamp: time.Now().Format("15:04"),
						}
						jsonBytes, _ := json.Marshal(outMsg)
						hub.Broadcast <- BroadcastMessage{RoomID: c.RoomID, Content: jsonBytes}
					} else {
						log.Println("Edit attempt blocked: time limit exceeded")
					}
				}
			}

		case "chat":
			mediaType := "text"
			if inMsg.MediaType != "" {
				mediaType = inMsg.MediaType
			}
			dbMessage := models.Message{
				RoomID:    c.RoomID,
				SenderID:  c.UserID,
				Content:   inMsg.Content,
				MediaType: mediaType,
			}
			db.Create(&dbMessage)

			outMsg := WSMessage{
				Type:      "chat",
				ID:        dbMessage.ID,
				Sender:    c.Username,
				SenderPic: c.ProfilePic,
				Content:   inMsg.Content,
				MediaType: mediaType,
				Timestamp: time.Now().Format("15:04"),
			}
			jsonBytes, _ := json.Marshal(outMsg)
			hub.Broadcast <- BroadcastMessage{RoomID: c.RoomID, Content: jsonBytes}
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
