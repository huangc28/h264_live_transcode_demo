package main

import (
	"net/http"

	"github.com/golang/glog"
	gorillaws "github.com/gorilla/websocket"
)

var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func broadcastVideoStream(h *Hub, w http.ResponseWriter, r *http.Request) {
	// Upgrade from http request to websocket.
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		glog.Errorf("failed to upgrade to websocket %v", err)
		return
	}

	// Create a new socket client.
	c := NewClient(conn)
	h.register <- c

	// Spin up go  routines to handle read / write messages.
	go c.WritePump()
}
