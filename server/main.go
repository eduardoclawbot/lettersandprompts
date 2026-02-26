package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/eduardoclawbot/lettersandprompts/chat"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from our domain
		origin := r.Header.Get("Origin")
		return origin == "" || 
			strings.HasPrefix(origin, "https://lettersandprompts.com") ||
			strings.HasPrefix(origin, "http://localhost")
	},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Serve static files from public/ (or ../public for local dev)
	publicDir := os.Getenv("PUBLIC_DIR")
	if publicDir == "" {
		// Try ../public for local dev, ./public for Docker
		if _, err := os.Stat("./public"); err == nil {
			publicDir = "./public"
		} else {
			publicDir = filepath.Join("..", "public")
		}
	}
	fs := http.FileServer(http.Dir(publicDir))
	
	// Initialize chat hub
	hub := chat.NewHub()
	go hub.Run()
	
	// Health check endpoint for Cloud Run
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// WebSocket endpoint for chat
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	// Serve all other requests as static files
	http.Handle("/", fs)

	log.Printf("Server starting on port %s, serving from %s\n", port, publicDir)
	log.Printf("WebSocket chat available at /ws")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleWebSocket(hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	// Get handle from query params
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		handle = "Guest"
	}

	// Sanitize handle
	handle = strings.TrimSpace(handle)
	if len(handle) > 20 {
		handle = handle[:20]
	}

	// Generate color for this handle
	color := chat.HandleColor(handle)

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client
	client := chat.NewClient(hub, conn, handle, color)
	hub.Register(client)

	// Start read and write pumps
	go client.WritePump()
	client.ReadPump() // Blocking call
}
