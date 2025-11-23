package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte)}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}
