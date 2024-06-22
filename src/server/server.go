package server

import (
	"log"
	"net/http"
)

// Handler for the root URL
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		log.Fatal("Page is not found: ", r.Method)
	}
	http.ServeFile(w, r, "templates/index.html")
}

// Handler for the /templates path to serve static files
func TemplatesHandler(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("templates"))
	http.StripPrefix("/templates/", fs).ServeHTTP(w, r)
}


// func hello() {
// 	fmt.Println("hello world")
// }