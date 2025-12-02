package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type BroadcastMessage struct {
	RoomID  uint
	Content []byte // The JSON payload to send to clients
}

type Client struct {
	Conn       *websocket.Conn
	UserID     uint
	Username   string
	ProfilePic string
	Send       chan []byte
	RoomID     uint
}

type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan BroadcastMessage
	mu         sync.Mutex
}

var h *Hub
var once sync.Once

func GetHub() *Hub {
	once.Do(func() {
		h = &Hub{
			Clients:    make(map[*Client]bool),
			Register:   make(chan *Client),
			Unregister: make(chan *Client),
			Broadcast:  make(chan BroadcastMessage),
		}
		go h.run()
	})
	return h
}

func (hub *Hub) run() {
	for {
		select {
		case client := <-hub.Register:
			hub.mu.Lock()
			hub.Clients[client] = true
			hub.mu.Unlock()
			go hub.broadcastOnlineUsers()
		case client := <-hub.Unregister:
			hub.mu.Lock()
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.Send)
			}
			hub.mu.Unlock()
			go hub.broadcastOnlineUsers()
		case message := <-hub.Broadcast:
			hub.mu.Lock()
			for client := range hub.Clients {
				// Ideally add room filtering and message routing here
				if client.RoomID == message.RoomID {
					select {
					case client.Send <- message.Content:
					default:
						close(client.Send)
						delete(hub.Clients, client)
					}
				}
			}
			hub.mu.Unlock()
		}
	}
}

func (hub *Hub) broadcastOnlineUsers() {
	hub.mu.Lock()
	users := make([]string, 0)
	// Use a map to ensure unique usernames if a user has multiple tabs open
	seen := make(map[string]bool)
	for client := range hub.Clients {
		if !seen[client.Username] {
			users = append(users, client.Username)
			seen[client.Username] = true
		}
	}
	hub.mu.Unlock()

	msg := map[string]interface{}{
		"type":  "online_users",
		"users": users,
	}
	jsonBytes, _ := json.Marshal(msg)

	hub.mu.Lock()
	defer hub.mu.Unlock()
	for client := range hub.Clients {
		select {
		case client.Send <- jsonBytes:
		default:
			close(client.Send)
			delete(hub.Clients, client)
		}
	}
}
