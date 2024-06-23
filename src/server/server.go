package server

import (
	"log"
	"net/http"
	// "asciiart/src/asciiart"
	"html/template"
	// "strings"
)

// struct for error
type errorpage struct {
	httpcode string
	message string
}

// struct for result
type resultpage struct {
	input string
	asciiartbanner string
	result string
}

// Handler for the root URL
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		err := errorpage{httpcode: "404", message: "Page does not exist"}
		w.WriteHeader(http.StatusNotFound)
		errorhandler(w, r, &err)
		http.ServeFile(w, r, "templates/error.html")
		log.Print("error in request ", r.Method)
	}

	if r.Method != "GET" {
		err := errorpage{httpcode: "405", message: "Method is not allowed"}
		w.WriteHeader(http.StatusNotFound)
		errorhandler(w, r, &err)
		log.Print("error in request ", r.Method)
	}
	http.ServeFile(w, r, "templates/index.html")
	}

// Handler for the error
func errorhandler(w http.ResponseWriter, r *http.Request, err *errorpage) {
errorp := template.Must(template.ParseFiles("templates/error.html"))
errorp.Execute(w, err)
}

// Handler for the /templates path to serve static files
func TemplatesHandler(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("templates"))
	http.StripPrefix("/templates/", fs).ServeHTTP(w, r)
}
