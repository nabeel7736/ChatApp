package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	UserID uint
	Send   chan []byte
	RoomID uint
}

type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
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
			Broadcast:  make(chan []byte),
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
		case client := <-hub.Unregister:
			hub.mu.Lock()
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.Send)
			}
			hub.mu.Unlock()
		case message := <-hub.Broadcast:
			hub.mu.Lock()
			for client := range hub.Clients {
				// Ideally add room filtering and message routing here
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(hub.Clients, client)
				}
			}
			hub.mu.Unlock()
		}
	}
}
