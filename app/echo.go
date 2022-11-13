package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
		if r.Header.Get("origin") == "https://statbate.com" {
			return true
		}
		return false
	},
}

var (
	wsClients = make(map[*websocket.Conn]struct{})

	ws = struct {
		Count chan int
		Send  chan []byte
		Add   chan *websocket.Conn
	}{
		Count: make(chan int, 100),
		Send:  make(chan []byte, 100),
		Add:   make(chan *websocket.Conn, 100),
	}
)

func broadcast() {
	for {
		select {
		case conn := <-ws.Add:
			wsClients[conn] = struct{}{}

		case <-ws.Count:
			ws.Count <- len(wsClients)

		case message := <-ws.Send:
			for conn := range wsClients {
				if err := conn.WriteMessage(1, message); err != nil {
					conn.Close()
					delete(wsClients, conn)
				}
			}
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	ws.Add <- conn
}
