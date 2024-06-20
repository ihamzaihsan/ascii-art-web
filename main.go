package main

import (
	"fmt"
	"log"
	"net/http"
)

// Handler for the root URL
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "templates/index.html")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Perform the action you want here
		fmt.Println("Button clicked, running Go function!")
		// You can send a response back to the client if needed
		w.Write([]byte("Go function executed!"))
	}
}

// Handler for the /templates path to serve static files
func templatesHandler(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("templates"))
	http.StripPrefix("/templates/", fs).ServeHTTP(w, r)
}

func main() {
	// Serve the root URL with the rootHandler
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/run", handleRequest)
	// Serve static files from the /templates directory
	http.HandleFunc("/templates/", templatesHandler)

	// Start the server on port 8080
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
