// Package ai defines a pluggable content-generation layer. The default
// provider runs a local model via Ollama (zero cost); a "bring your own key"
// provider for Claude/OpenAI can be selected via configuration (see ADR-003).
package ai

import (
	"context"
	"os"
)

// GenerateRequest is the user-supplied seed for a LinkedIn post.
type GenerateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Provider generates a formatted LinkedIn post from a seed.
type Provider interface {
	// Name identifies the provider (e.g. "ollama").
	Name() string
	// Generate returns a copy-paste-ready LinkedIn post.
	Generate(ctx context.Context, req GenerateRequest) (string, error)
}

// NewFromEnv selects a provider based on environment configuration,
// defaulting to the local Ollama provider so the app works at zero cost.
func NewFromEnv() Provider {
	switch os.Getenv("LPE_AI_PROVIDER") {
	// Future: case "api": return newAPIProvider(...)
	default:
		return NewOllama(
			envOr("LPE_OLLAMA_URL", "http://localhost:11434"),
			envOr("LPE_OLLAMA_MODEL", "llama3.1"),
		)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
