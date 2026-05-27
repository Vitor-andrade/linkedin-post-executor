package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/ai"
)

// handleSuggestions returns the curated, offline catalog of post-topic ideas.
func handleSuggestions(_ Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, ai.Categories())
	}
}

// handleSuggestionsGenerate asks the active AI provider for fresh ideas for a
// category and/or a specific technology query.
func handleSuggestionsGenerate(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Category string `json:"category"`
			Query    string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
		defer cancel()

		ideas, err := d.AI.Suggest(ctx, req.Category, req.Query)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"suggestions": ideas})
	}
}
