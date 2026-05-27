package ai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var seed = GenerateRequest{Title: "Go concurrency", Description: "channels"}

func TestGeminiGenerate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "k-123" {
			t.Errorf("missing api key in query: %s", r.URL.RawQuery)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "Go concurrency") {
			t.Errorf("prompt not forwarded: %s", body)
		}
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"a post"}]}}]}`))
	}))
	defer srv.Close()
	restore := geminiBaseURL
	geminiBaseURL = srv.URL
	defer func() { geminiBaseURL = restore }()

	got, err := NewGemini("k-123", "gemini-2.0-flash").Generate(context.Background(), seed)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if got != "a post" {
		t.Errorf("got %q", got)
	}
}

func TestAnthropicGenerate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "sk-ant" || r.Header.Get("anthropic-version") == "" {
			t.Errorf("missing auth headers: %v", r.Header)
		}
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"claude post"}]}`))
	}))
	defer srv.Close()
	restore := anthropicEndpoint
	anthropicEndpoint = srv.URL
	defer func() { anthropicEndpoint = restore }()

	got, err := NewAnthropic("sk-ant", "claude-haiku-4-5-20251001").Generate(context.Background(), seed)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if got != "claude post" {
		t.Errorf("got %q", got)
	}
}

func TestOpenAIGenerate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-oai" {
			t.Errorf("missing bearer: %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"oai post"}}]}`))
	}))
	defer srv.Close()
	restore := openaiEndpoint
	openaiEndpoint = srv.URL
	defer func() { openaiEndpoint = restore }()

	got, err := NewOpenAI("sk-oai", "gpt-4o-mini").Generate(context.Background(), seed)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if got != "oai post" {
		t.Errorf("got %q", got)
	}
}

func TestProvidersRequireAPIKey(t *testing.T) {
	cases := map[string]Provider{
		"gemini":    NewGemini("", "m"),
		"anthropic": NewAnthropic("", "m"),
		"openai":    NewOpenAI("", "m"),
	}
	for name, p := range cases {
		if _, err := p.Generate(context.Background(), seed); err == nil {
			t.Errorf("%s: expected an error when the API key is missing", name)
		}
	}
}

func TestNewFromEnvSelectsProvider(t *testing.T) {
	t.Setenv("LPE_AI_PROVIDER", "gemini")
	t.Setenv("LPE_GEMINI_API_KEY", "x")
	if name := NewFromEnv().Name(); !strings.HasPrefix(name, "gemini") {
		t.Errorf("got %q, want a gemini provider", name)
	}

	t.Setenv("LPE_AI_PROVIDER", "")
	if name := NewFromEnv().Name(); !strings.HasPrefix(name, "ollama") {
		t.Errorf("default got %q, want ollama", name)
	}
}
