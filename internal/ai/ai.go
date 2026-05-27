// Package ai defines a pluggable content-generation layer. The default
// provider runs a local model via Ollama (zero cost); a "bring your own key"
// provider for Claude/OpenAI can be selected via configuration (see ADR-003).
package ai

import (
	"context"
	"os"
	"strings"
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

// NewFromEnv selects a provider based on LPE_AI_PROVIDER, defaulting to the
// local Ollama provider so the app works at zero cost. The "bring your own
// key" providers (Gemini, Claude, OpenAI) read their API key from the
// environment; a missing key surfaces as a clear error at generation time.
func NewFromEnv() Provider {
	switch strings.ToLower(os.Getenv("LPE_AI_PROVIDER")) {
	case "gemini":
		return NewGemini(
			os.Getenv("LPE_GEMINI_API_KEY"),
			envOr("LPE_GEMINI_MODEL", "gemini-2.0-flash"),
		)
	case "anthropic", "claude":
		return NewAnthropic(
			os.Getenv("LPE_ANTHROPIC_API_KEY"),
			envOr("LPE_ANTHROPIC_MODEL", "claude-haiku-4-5-20251001"),
		)
	case "openai":
		return NewOpenAI(
			os.Getenv("LPE_OPENAI_API_KEY"),
			envOr("LPE_OPENAI_MODEL", "gpt-4o-mini"),
		)
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
