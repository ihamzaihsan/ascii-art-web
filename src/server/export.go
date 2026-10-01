package server

import (
	"bytes"
	"errors"
	"image/png"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// Export accepts the displayed ASCII through an ordinary HTML POST form.
func (a *app) export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		a.fail(w, http.StatusMethodNotAllowed, "Use an artwork download button.")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/x-www-form-urlencoded" {
		a.fail(w, http.StatusUnsupportedMediaType, "Submit a URL-encoded export form.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 512<<10)
	if err := r.ParseForm(); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			a.fail(w, http.StatusRequestEntityTooLarge, "Export form exceeds 512 KiB.")
		} else {
			a.fail(w, http.StatusBadRequest, "Export form could not be read.")
		}
		return
	}
	// HTML form submission normalizes line breaks to CRLF. Keep exports aligned
	// with the LF-based renderer and reject lone carriage returns below.
	art := strings.ReplaceAll(r.PostForm.Get("art"), "\r\n", "\n")
	if len(art) == 0 || len(art) > 100000 {
		a.fail(w, http.StatusBadRequest, "Artwork must contain 1 to 100,000 ASCII characters.")
		return
	}
	for _, c := range art {
		if c != '\n' && (c < 32 || c > 126) {
			a.fail(w, http.StatusBadRequest, "Artwork must contain printable ASCII and newlines.")
			return
		}
	}
	var body []byte
	format := r.PostForm.Get("format")
	switch format {
	case "txt":
		body = []byte(art)
	case "png":
		frame := 0
		if text := r.PostForm.Get("frame"); text != "" {
			frame, err = strconv.Atoi(text)
			if err != nil {
				a.fail(w, http.StatusBadRequest, "PNG frame must be a whole number.")
				return
			}
		}
		img, err := a.generator.PNGImage(art, frame)
		if err != nil {
			a.fail(w, http.StatusBadRequest, err.Error())
			return
		}
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, img); err != nil {
			a.fail(w, http.StatusInternalServerError, "PNG could not be encoded.")
			return
		}
		body = encoded.Bytes()
	default:
		a.fail(w, http.StatusBadRequest, "Choose TXT or PNG.")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", `attachment; filename="ascii-art.`+format+`"`)
	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	if _, err := w.Write(body); err != nil {
		log.Printf("write export: %v", err)
	}
}
