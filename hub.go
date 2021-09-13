package main

type Hub struct {
	// Registered client.
	clients map[*Client]bool

	// Broadcast message to all clients.
	broadcast chan []byte

	// Register requested client.
	register chan *Client

	// Unregister requested client.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 32*1024),
	}
}

// Run performs the action of each connected client in the backgrond (goroutine)
// It register, unregister, send (to a specific client), and broadcast(to a group of clients).
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			// Register socket client to a map.
			h.clients[client] = true
		case client := <-h.unregister:
			if _, exists := h.clients[client]; exists {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// Broadcast messages to all clients.
			for c := range h.clients {
				c.send <- message
			}
		}
	}
}
