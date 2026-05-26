// Package server wires the HTTP API and serves the embedded UI. It uses the
// standard library's net/http with the Go 1.22+ routing patterns, keeping
// dependencies minimal.
package server

import (
	"io/fs"
	"net/http"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/ai"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/store"
)

// Deps holds the collaborators the server needs.
type Deps struct {
	Store *store.Store
	AI    ai.Provider
	UI    fs.FS
}

// New builds the root http.Handler: JSON API under /api and the SPA UI for
// everything else.
func New(d Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth(d))
	mux.HandleFunc("POST /api/generate", handleGenerate(d))
	mux.HandleFunc("GET /api/drafts", handleListDrafts(d))
	mux.HandleFunc("POST /api/drafts", handleCreateDraft(d))
	mux.HandleFunc("GET /api/drafts/{id}", handleGetDraft(d))
	mux.HandleFunc("PUT /api/drafts/{id}", handleUpdateDraft(d))

	// Anything not under /api is served by the SPA.
	mux.Handle("/", spaHandler(d.UI))

	return mux
}
