package ai

import "testing"

func TestBuildPromptIncludesTitleAndDescription(t *testing.T) {
	got := buildPrompt(GenerateRequest{
		Title:       "Escalando uma API em Go",
		Description: "Lições sobre concorrência",
	})

	if !contains(got, "Escalando uma API em Go") {
		t.Errorf("prompt não contém o título: %q", got)
	}
	if !contains(got, "Lições sobre concorrência") {
		t.Errorf("prompt não contém a descrição: %q", got)
	}
}

func TestBuildPromptOmitsEmptyDescription(t *testing.T) {
	got := buildPrompt(GenerateRequest{Title: "Apenas título"})
	if contains(got, "Descrição/Contexto:") {
		t.Errorf("prompt não deveria incluir rótulo de descrição quando vazia: %q", got)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
