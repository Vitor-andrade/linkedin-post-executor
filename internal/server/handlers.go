package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/ai"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/store"
)

func handleHealth(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":     "ok",
			"aiProvider": d.AI.Name(),
		})
	}
}

func handleMetrics(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := d.Store.Metrics(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, m)
	}
}

func handleGenerate(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ai.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
		defer cancel()

		content, err := d.AI.Generate(ctx, req)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"content": content})
	}
}

func handleListDrafts(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := d.Store.ListDrafts(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, drafts)
	}
}

func handleCreateDraft(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in store.Draft
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(in.Title) == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		created, err := d.Store.CreateDraft(r.Context(), in)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func handleGetDraft(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		draft, err := d.Store.GetDraft(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "draft not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, draft)
	}
}

func handleUpdateDraft(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		var in store.Draft
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(in.Title) == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		updated, err := d.Store.UpdateDraft(r.Context(), id, in)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "draft not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func handleDeleteDraft(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		err := d.Store.DeleteDraft(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "draft not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// pathID parses the {id} path value, writing a 400 and returning false if it
// is missing or not a positive integer.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid draft id")
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
