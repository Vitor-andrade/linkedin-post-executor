package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

// stateCookie carries the anti-CSRF OAuth state between login and callback.
const stateCookie = "li_oauth_state"

func handleLinkedInStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := d.LinkedIn.Status(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, status)
	}
}

func handleLinkedInLogin(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !d.LinkedIn.Configured() {
			writeError(w, http.StatusPreconditionFailed,
				"LinkedIn credentials not configured (set LPE_LINKEDIN_CLIENT_ID and LPE_LINKEDIN_CLIENT_SECRET)")
			return
		}
		state := randomState()
		http.SetCookie(w, &http.Cookie{
			Name:     stateCookie,
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   600,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, d.LinkedIn.AuthCodeURL(state), http.StatusFound)
	}
}

func handleLinkedInCallback(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			http.Redirect(w, r, "/?linkedin=error", http.StatusFound)
			return
		}
		cookie, err := r.Cookie(stateCookie)
		if err != nil || cookie.Value == "" || cookie.Value != q.Get("state") {
			writeError(w, http.StatusBadRequest, "invalid OAuth state")
			return
		}
		// Clear the one-time state cookie.
		http.SetCookie(w, &http.Cookie{Name: stateCookie, Value: "", Path: "/", MaxAge: -1})

		if err := d.LinkedIn.Connect(r.Context(), q.Get("code")); err != nil {
			http.Redirect(w, r, "/?linkedin=error", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/?linkedin=connected", http.StatusFound)
	}
}

func handleLinkedInPublish(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Content  string  `json:"content"`
			DraftID  *int64  `json:"draftId"`
			ImageIDs []int64 `json:"imageIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		text := strings.TrimSpace(req.Content)
		if text == "" && req.DraftID != nil {
			draft, err := d.Store.GetDraft(r.Context(), *req.DraftID)
			if err != nil {
				writeError(w, http.StatusBadRequest, "draft not found")
				return
			}
			text = strings.TrimSpace(draft.Content)
		}
		if text == "" {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}

		var images [][]byte
		for _, id := range req.ImageIDs {
			img, err := d.Store.GetImage(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusBadRequest, "image not found")
				return
			}
			images = append(images, img.Data)
		}

		urn, err := d.LinkedIn.PublishNowWithImages(r.Context(), text, images)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if req.DraftID != nil {
			_ = d.Store.SetDraftStatus(r.Context(), *req.DraftID, "published")
		}
		writeJSON(w, http.StatusOK, map[string]string{"urn": urn})
	}
}

func handleLinkedInDisconnect(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := d.LinkedIn.Disconnect(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
	}
}

// randomState returns a cryptographically random hex string for OAuth CSRF
// protection.
func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
