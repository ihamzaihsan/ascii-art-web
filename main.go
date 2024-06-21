package main

import (
	"ascii/src/server"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Print("Server running on http://localhost:8080 \nTo stop the server press Ctrl+C")

	// Register the request handlers
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.HandleFunc("/", server.MainHandler)
	http.HandleFunc("/ascii-art", server.ResultHandler)

	//Start the http server
	log.Fatal(http.ListenAndServe(":8080", nil))

}
