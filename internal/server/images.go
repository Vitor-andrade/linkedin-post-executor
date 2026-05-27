package server

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strings"
)

const (
	maxImageBytes = 12 << 20 // 12 MB per image
	maxImages     = 9        // LinkedIn's per-post limit
)

// handleUploadImages stores one or more uploaded images (multipart field
// "files") as blobs and returns their metadata.
func handleUploadImages(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxImages * maxImageBytes); err != nil {
			writeError(w, http.StatusBadRequest, "invalid upload")
			return
		}
		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			writeError(w, http.StatusBadRequest, "no files provided")
			return
		}
		if len(files) > maxImages {
			writeError(w, http.StatusBadRequest, "too many images")
			return
		}

		type uploaded struct {
			ID       int64  `json:"id"`
			Filename string `json:"filename"`
			Mime     string `json:"mime"`
		}
		out := make([]uploaded, 0, len(files))

		for _, fh := range files {
			if fh.Size > maxImageBytes {
				writeError(w, http.StatusRequestEntityTooLarge, "image too large (max 12 MB)")
				return
			}
			mime := fh.Header.Get("Content-Type")
			if !strings.HasPrefix(mime, "image/") {
				writeError(w, http.StatusBadRequest, "only image files are allowed")
				return
			}
			f, err := fh.Open()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			data, err := io.ReadAll(io.LimitReader(f, maxImageBytes))
			f.Close()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			id, err := d.Store.SaveImage(r.Context(), mime, fh.Filename, data)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			out = append(out, uploaded{ID: id, Filename: fh.Filename, Mime: mime})
		}
		writeJSON(w, http.StatusCreated, out)
	}
}

// handleGetImage serves a stored image blob so the UI can preview it.
func handleGetImage(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		img, err := d.Store.GetImage(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "image not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", img.Mime)
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = w.Write(img.Data)
	}
}
