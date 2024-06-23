package main

import (
	"log"
	"net/http"
	// "fmt"
	"asciiart/src/server"
)

func main() {
	// Serve the root URL with the rootHandler
	http.HandleFunc("/", server.RootHandler)
	// Serve static files from the /templates directory
	http.HandleFunc("/templates/", server.TemplatesHandler)
	// Start the server on port 8080
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
