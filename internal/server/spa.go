package server

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// spaHandler serves static assets from the embedded UI filesystem and falls
// back to index.html for client-side routes (single-page application).
func spaHandler(ui fs.FS) http.Handler {
	fileServer := http.FileServerFS(ui)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" {
			clean = "index.html"
		}

		// If the requested file does not exist, serve the SPA entrypoint so
		// the client router can handle the route.
		if _, err := fs.Stat(ui, clean); err != nil {
			if os.IsNotExist(err) {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
