package ai

import (
	"strings"
	"testing"
)

func TestBuildPromptIncludesTitleAndDescription(t *testing.T) {
	got := buildPrompt(GenerateRequest{
		Title:       "Scaling an API in Go",
		Description: "Lessons on concurrency",
	})

	if !strings.Contains(got, "Scaling an API in Go") {
		t.Errorf("prompt does not contain the title: %q", got)
	}
	if !strings.Contains(got, "Lessons on concurrency") {
		t.Errorf("prompt does not contain the description: %q", got)
	}
}

func TestBuildPromptOmitsEmptyDescription(t *testing.T) {
	got := buildPrompt(GenerateRequest{Title: "Title only"})
	if strings.Contains(got, "Description/Context:") {
		t.Errorf("prompt should not include the description label when empty: %q", got)
	}
}
