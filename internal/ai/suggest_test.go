package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseSuggestions(t *testing.T) {
	raw := "1. First idea\n- Second idea\n\n  • Third idea  \n4) Fourth idea\n"
	got := parseSuggestions(raw)
	want := []string{"First idea", "Second idea", "Third idea", "Fourth idea"}
	if len(got) != len(want) {
		t.Fatalf("got %d ideas %v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("idea %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSuggestViaProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Idea A\nIdea B"}]}}]}`))
	}))
	defer srv.Close()
	restore := geminiBaseURL
	geminiBaseURL = srv.URL
	defer func() { geminiBaseURL = restore }()

	ideas, err := NewGemini("k", "gemini-2.0-flash").Suggest(context.Background(), "Backend", "Go")
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(ideas) != 2 || ideas[0] != "Idea A" {
		t.Errorf("got %v", ideas)
	}
}

func TestCategoriesNonEmpty(t *testing.T) {
	cats := Categories()
	if len(cats) == 0 {
		t.Fatal("expected curated categories")
	}
	for _, c := range cats {
		if c.Name == "" || len(c.Topics) == 0 {
			t.Errorf("category %q has no topics", c.Name)
		}
	}
}
