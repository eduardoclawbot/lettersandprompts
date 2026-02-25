package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

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
	
	// Health check endpoint for Cloud Run
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Serve all other requests as static files
	http.Handle("/", fs)

	log.Printf("Server starting on port %s, serving from %s\n", port, publicDir)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
