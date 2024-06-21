package server

import (
	"html/template"
	"net/http"
	"ascii/src/asciiart"
	"strings"
)

// Struct to hold the error data
type ErrorPageData struct {
	Code     string
	ErrorMsg string
}

// Struct to hold the result data
type ResultPageData struct {
	Input  string
	Banner string
	Result string
}

// Function to render the error page
func errHandler(w http.ResponseWriter, r *http.Request, err *ErrorPageData) {
	errorTemp := template.Must(template.ParseFiles("templates/error.html"))
	errorTemp.Execute(w, err)

}

// Function to render the main page
func MainHandler(w http.ResponseWriter, r *http.Request) {
	//Validating the request path
	if r.URL.Path != "/" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		w.WriteHeader(http.StatusNotFound)
		errHandler(w, r, &err)
		return
	}
	// Validating the request method
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	// Validating the parsing of the main page
	main, err := template.ParseFiles("templates/index.html")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}

	mainTemp := template.Must(main, nil)
	mainTemp.Execute(w, nil)
}

// Function to render the result page
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	// Validating the paesing of the form
	if err := r.ParseForm(); err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}
	// Validation for the input
	input := r.PostFormValue("input-text")
	inputValidation := strings.ReplaceAll(input, "\r\n", "")

	for _, letter := range inputValidation {
		if letter < 32 || letter > 126 {
			err := ErrorPageData{Code: "400", ErrorMsg: "INVALID INPUT"}
			w.WriteHeader(http.StatusNotAcceptable)
			errHandler(w, r, &err)
			return
		}
	}
	// Validation for the banner
	banner := r.PostFormValue("banner")
	if banner != "standard" && banner != "shadow" && banner != "thinkertoy" {
		err := ErrorPageData{Code: "404", ErrorMsg: "BANNER NOT FOUND"}
		w.WriteHeader(http.StatusNotFound)
		errHandler(w, r, &err)
		return

	}
	//Validation for asciiart functions
	ascii, err := asciiart.AsciiArt(input, banner) 
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}

	resultTemp := template.Must(template.ParseFiles("templates/ascii-art.html"))

	output := ResultPageData{Input: input, Banner: banner, Result: ascii}

	resultTemp.Execute(w, output)

}
