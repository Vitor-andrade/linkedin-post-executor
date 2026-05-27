package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseSuggestions(t *testing.T) {
	raw := "1. First idea :: do this\n- Second idea :: and that\n\n  • Bare idea  \n"
	got := parseSuggestions(raw)
	if len(got) != 3 {
		t.Fatalf("got %d ideas %v, want 3", len(got), got)
	}
	if got[0].Title != "First idea" || got[0].Description != "do this" {
		t.Errorf("idea 0 = %+v", got[0])
	}
	if got[1].Title != "Second idea" || got[1].Description != "and that" {
		t.Errorf("idea 1 = %+v", got[1])
	}
	// A line without the separator keeps the whole text as the title.
	if got[2].Title != "Bare idea" || got[2].Description != "" {
		t.Errorf("idea 2 = %+v", got[2])
	}
}

func TestSuggestViaProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Idea A :: ctx A\nIdea B :: ctx B"}]}}]}`))
	}))
	defer srv.Close()
	restore := geminiBaseURL
	geminiBaseURL = srv.URL
	defer func() { geminiBaseURL = restore }()

	ideas, err := NewGemini("k", "gemini-2.0-flash").Suggest(context.Background(), "Backend", "Go")
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(ideas) != 2 || ideas[0].Title != "Idea A" || ideas[0].Description != "ctx A" {
		t.Errorf("got %+v", ideas)
	}
}

func TestCategoriesNonEmpty(t *testing.T) {
	cats := Categories()
	if len(cats) == 0 {
		t.Fatal("expected curated categories")
	}
	for _, c := range cats {
		if c.Name == "" || len(c.Ideas) == 0 {
			t.Errorf("category %q has no ideas", c.Name)
		}
		for _, idea := range c.Ideas {
			if idea.Title == "" || idea.Description == "" {
				t.Errorf("category %q has an incomplete idea: %+v", c.Name, idea)
			}
		}
	}
}
