package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/store"
)

func handleListScheduled(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := d.Store.ListScheduledPosts(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, posts)
	}
}

func handleCreateScheduled(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in store.ScheduledPost
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		content := strings.TrimSpace(in.Content)
		if content == "" && in.DraftID != nil {
			draft, err := d.Store.GetDraft(r.Context(), *in.DraftID)
			if err != nil {
				writeError(w, http.StatusBadRequest, "draft not found")
				return
			}
			content = strings.TrimSpace(draft.Content)
		}
		if content == "" {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}
		if in.ScheduledFor.IsZero() {
			writeError(w, http.StatusBadRequest, "scheduledFor is required")
			return
		}
		in.Content = content

		created, err := d.Store.CreateScheduledPost(r.Context(), in)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func handleCancelScheduled(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		err := d.Store.CancelScheduledPost(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "no pending scheduled post with that id")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
	}
}
