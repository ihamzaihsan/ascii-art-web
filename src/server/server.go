// Package server implements the HTML interface and HTTP input boundary.
package server

import (
	"ascii/src/asciiart"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"mime"
	"net/http"
)

const maxBodyBytes = 4096

type pageData struct {
	Input, Banner, Result string
	Code                  int
	ErrorMsg              string
}

type app struct {
	templates *template.Template
	generator *asciiart.Generator
}

// NewHandler validates bundled resources before accepting requests.
func NewHandler(files fs.FS) (http.Handler, error) {
	banners, err := fs.Sub(files, "banners")
	if err != nil {
		return nil, err
	}
	generator, err := asciiart.New(banners)
	if err != nil {
		return nil, err
	}
	templates, err := template.ParseFS(files, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	for _, name := range []string{"index.html", "ascii-art.html", "error.html"} {
		if templates.Lookup(name) == nil {
			return nil, fmt.Errorf("missing template %s", name)
		}
	}
	assets, err := fs.Sub(files, "assets")
	if err != nil {
		return nil, err
	}
	a := &app{templates: templates, generator: generator}
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("/", a.home)
	mux.HandleFunc("/ascii-art", a.result)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; img-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		mux.ServeHTTP(w, r)
	}), nil
}

// Buffer templates so rendering failures cannot send partial successful pages.
func (a *app) render(w http.ResponseWriter, status int, name string, data pageData) {
	var body bytes.Buffer
	if err := a.templates.ExecuteTemplate(&body, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := body.WriteTo(w); err != nil {
		log.Printf("write response: %v", err)
	}
}

func (a *app) fail(w http.ResponseWriter, code int, message string) {
	a.render(w, code, "error.html", pageData{Code: code, ErrorMsg: message})
}

func (a *app) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.fail(w, http.StatusNotFound, "This page does not exist.")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		a.fail(w, http.StatusMethodNotAllowed, "Use GET to open the generator.")
		return
	}
	a.render(w, http.StatusOK, "index.html", pageData{})
}

func (a *app) result(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		a.fail(w, http.StatusMethodNotAllowed, "Submit the form to generate ASCII art.")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/x-www-form-urlencoded" {
		a.fail(w, http.StatusUnsupportedMediaType, "Submit a URL-encoded form.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := r.ParseForm(); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			a.fail(w, http.StatusRequestEntityTooLarge, "The form exceeds 4 KiB.")
		} else {
			a.fail(w, http.StatusBadRequest, "The form could not be read.")
		}
		return
	}
	input, banner := r.PostForm.Get("input-text"), r.PostForm.Get("banner")
	result, err := a.generator.Render(input, banner)
	if err != nil {
		a.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	a.render(w, http.StatusOK, "ascii-art.html", pageData{Input: input, Banner: banner, Result: result})
}
