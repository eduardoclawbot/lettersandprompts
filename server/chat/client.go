package chat

import (
	"encoding/json"
	"html"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512

	// Rate limiting: max 1 message per second
	messageRateLimit = time.Second
)

// Client represents a WebSocket client
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan *Message
	handle string
	color  string

	// Rate limiting
	lastMessage time.Time
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn, handle, color string) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan *Message, 256),
		handle:      handle,
		color:       color,
		lastMessage: time.Now(),
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var inMsg struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}

		if err := json.Unmarshal(msgBytes, &inMsg); err != nil {
			log.Printf("JSON parse error: %v", err)
			continue
		}

		// Rate limiting
		if time.Since(c.lastMessage) < messageRateLimit {
			log.Printf("Rate limit exceeded for %s", c.handle)
			continue
		}
		c.lastMessage = time.Now()

		// Sanitize input
		text := html.EscapeString(strings.TrimSpace(inMsg.Text))
		if text == "" || len(text) > 500 {
			continue
		}

		// Create message
		msg := NewMessage(c.handle, text, c.color)

		// Broadcast to all clients
		c.hub.broadcast <- msg
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send JSON message
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
