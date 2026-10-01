package server

import (
	"ascii/src/imageart"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"
	"strconv"
)

const maxImageBytes = 10 << 20
const maxImageFormBytes = 13 << 20 // upload + retained preview + form overhead

type imageControl struct {
	imageart.Control
	Value float64
}

func (a *app) renderImage(w http.ResponseWriter, data pageData) {
	for _, c := range imageart.Controls() {
		data.Controls = append(data.Controls, imageControl{c, data.Options.Values[c.Name]})
	}
	a.render(w, http.StatusOK, "image-ascii.html", data)
}

func imageOptions(r *http.Request) (imageart.Options, error) {
	o := imageart.Defaults()
	for _, c := range imageart.Controls() {
		if text := r.PostForm.Get(c.Name); text != "" {
			v, err := strconv.ParseFloat(text, 64)
			if err != nil {
				return o, fmt.Errorf("%s must be a number", c.Label)
			}
			o.Values[c.Name] = v
		}
	}
	if text := r.PostForm.Get("gradient"); text != "" {
		o.Gradient = text
	}
	if text := r.PostForm.Get("quality"); text != "" {
		o.Quality = text
	}
	o.Threshold = r.PostForm.Get("threshold-enabled") == "on"
	o.Sharpness = r.PostForm.Get("sharpness-enabled") == "on"
	o.Edges = r.PostForm.Get("edges-enabled") == "on"
	return o, o.Validate()
}

func decodeImage(data []byte) (image.Image, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("choose a valid PNG, JPEG, or GIF image")
	}
	// Check dimensions before allocating the decoded image, including overflow.
	if config.Width < 1 || config.Height < 1 || config.Width > 24000000/config.Height {
		return nil, fmt.Errorf("image must contain at most 24 million pixels")
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("image could not be decoded")
	}
	return src, nil
}

func (a *app) image(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodGet {
		a.renderImage(w, pageData{Options: imageart.Defaults()})
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		a.fail(w, http.StatusMethodNotAllowed, "Use GET or submit the image form.")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		a.fail(w, http.StatusUnsupportedMediaType, "Submit the multipart image form.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxImageFormBytes)
	err = r.ParseMultipartForm(maxImageFormBytes)
	// ParseMultipartForm can populate temporary files even when parsing fails.
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			a.fail(w, http.StatusRequestEntityTooLarge, "Image form exceeds 13 MiB.")
		} else {
			a.fail(w, http.StatusBadRequest, "Image form could not be read.")
		}
		return
	}
	o, err := imageOptions(r)
	if err != nil {
		a.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	var data []byte
	name := r.PostForm.Get("image-name")
	file, header, err := r.FormFile("image-file")
	if err == nil {
		defer file.Close()
		if header.Size > maxImageBytes {
			a.fail(w, http.StatusRequestEntityTooLarge, "Image exceeds 10 MiB.")
			return
		}
		data, err = io.ReadAll(io.LimitReader(file, maxImageBytes+1))
		name = header.Filename
		if err != nil {
			a.fail(w, http.StatusBadRequest, "Image could not be read.")
			return
		}
	} else if errors.Is(err, http.ErrMissingFile) {
		retained := r.PostForm.Get("image-data")
		if len(retained) > 2<<20 {
			a.fail(w, http.StatusRequestEntityTooLarge, "Retained image exceeds 2 MiB.")
			return
		}
		data, err = base64.StdEncoding.DecodeString(retained)
		if err != nil {
			a.fail(w, http.StatusBadRequest, "Retained image is invalid. Select the image again.")
			return
		}
	} else {
		a.fail(w, http.StatusBadRequest, "Image upload could not be read.")
		return
	}
	if len(data) == 0 {
		a.fail(w, http.StatusBadRequest, "Select an image to convert.")
		return
	}
	if len(data) > maxImageBytes {
		a.fail(w, http.StatusRequestEntityTooLarge, "Image exceeds 10 MiB.")
		return
	}
	src, err := decodeImage(data)
	if err != nil {
		a.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	// Retain a small PNG in the form so Apply can work without selecting the file
	// again. Every render uses this same bounded source; no images are stored.
	thumbnail := imageart.Thumbnail(src)
	art, rows, err := imageart.Render(thumbnail, o)
	if err != nil {
		a.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, thumbnail); err != nil {
		a.fail(w, http.StatusInternalServerError, "Image preview could not be encoded.")
		return
	}
	a.renderImage(w, pageData{Options: o, Image: base64.StdEncoding.EncodeToString(encoded.Bytes()), ImageName: name, Result: art, Rows: rows})
}
