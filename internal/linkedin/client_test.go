package linkedin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthCodeURL(t *testing.T) {
	cfg := Config{ClientID: "abc", RedirectURL: "http://localhost:8080/cb"}
	raw := cfg.AuthCodeURL("xyz")

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := u.Query()
	if q.Get("client_id") != "abc" || q.Get("state") != "xyz" {
		t.Errorf("missing params in %s", raw)
	}
	if q.Get("response_type") != "code" || q.Get("scope") != scopes {
		t.Errorf("unexpected response_type/scope in %s", raw)
	}
	if q.Get("redirect_uri") != "http://localhost:8080/cb" {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
}

func TestExchange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("grant_type") != "authorization_code" || r.FormValue("code") != "the-code" {
			t.Errorf("unexpected form: %v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at-123",
			"expires_in":    3600,
			"refresh_token": "rt-456",
		})
	}))
	defer srv.Close()
	restore := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = restore }()

	tok, err := NewClient().Exchange(context.Background(), Config{ClientID: "id", ClientSecret: "sec"}, "the-code")
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if tok.AccessToken != "at-123" || tok.RefreshToken != "rt-456" {
		t.Errorf("unexpected token: %+v", tok)
	}
	if tok.Expired() {
		t.Error("token should not be expired (~1h validity)")
	}
}

func TestExchangeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
	}))
	defer srv.Close()
	restore := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = restore }()

	if _, err := NewClient().Exchange(context.Background(), Config{}, "bad"); err == nil {
		t.Error("expected an error for a failed exchange")
	}
}

func TestAuthorURN(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer at-123" {
			t.Errorf("missing bearer token, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"sub": "MEMBER42"})
	}))
	defer srv.Close()
	restore := userinfoEndpoint
	userinfoEndpoint = srv.URL
	defer func() { userinfoEndpoint = restore }()

	urn, err := NewClient().AuthorURN(context.Background(), "at-123")
	if err != nil {
		t.Fatalf("author urn: %v", err)
	}
	if urn != "urn:li:person:MEMBER42" {
		t.Errorf("urn = %q", urn)
	}
}

func TestPublish(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "hello world") {
			t.Errorf("payload missing text: %s", body)
		}
		if !strings.Contains(string(body), "urn:li:person:ME") {
			t.Errorf("payload missing author: %s", body)
		}
		w.Header().Set("X-RestLi-Id", "urn:li:share:999")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	restore := ugcPostsEndpoint
	ugcPostsEndpoint = srv.URL
	defer func() { ugcPostsEndpoint = restore }()

	urn, err := NewClient().Publish(context.Background(), "at", "urn:li:person:ME", "hello world")
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if urn != "urn:li:share:999" {
		t.Errorf("urn = %q", urn)
	}
}
