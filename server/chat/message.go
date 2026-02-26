package chat

import "time"

// Message represents a chat message
type Message struct {
	Type   string   `json:"type"`             // "message", "system", "userlist", "join"
	Handle string   `json:"handle,omitempty"` // Username
	Text   string   `json:"text,omitempty"`   // Message text
	Color  string   `json:"color,omitempty"`  // Handle color
	Ts     int64    `json:"ts,omitempty"`     // Unix timestamp
	Users  []string `json:"users,omitempty"`  // For userlist messages
}

// NewMessage creates a new chat message
func NewMessage(handle, text, color string) *Message {
	return &Message{
		Type:   "message",
		Handle: handle,
		Text:   text,
		Color:  color,
		Ts:     time.Now().Unix(),
	}
}
