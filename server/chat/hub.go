package chat

import (
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	broadcast chan *Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- message
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			// Send join message to all clients
			joinMsg := &Message{
				Type:   "system",
				Text:   "*** " + client.handle + " has entered the room ***",
				Handle: "system",
			}
			h.BroadcastMessage(joinMsg)

			// Send user list update
			h.broadcastUserList()

			log.Printf("Client registered: %s (total: %d)", client.handle, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

			// Send leave message to remaining clients
			leaveMsg := &Message{
				Type:   "system",
				Text:   "*** " + client.handle + " has left the room ***",
				Handle: "system",
			}
			h.BroadcastMessage(leaveMsg)

			// Send updated user list
			h.broadcastUserList()

			log.Printf("Client unregistered: %s (total: %d)", client.handle, len(h.clients))

		case message := <-h.broadcast:
			h.BroadcastMessage(message)
		}
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *Hub) BroadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			// Client's send buffer is full, close it
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// broadcastUserList sends the current user list to all clients
func (h *Hub) broadcastUserList() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for client := range h.clients {
		users = append(users, client.handle)
	}

	userListMsg := &Message{
		Type:  "userlist",
		Users: users,
	}

	for client := range h.clients {
		select {
		case client.send <- userListMsg:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// GetUserCount returns the current number of connected users
func (h *Hub) GetUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
