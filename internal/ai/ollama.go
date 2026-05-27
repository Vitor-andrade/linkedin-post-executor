package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Ollama generates posts using a model served locally by Ollama.
// This is the zero-cost default provider.
type Ollama struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllama builds an Ollama-backed provider pointing at baseURL (e.g.
// http://localhost:11434) using the given model.
func NewOllama(baseURL, model string) *Ollama {
	return &Ollama{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Name implements Provider.
func (o *Ollama) Name() string { return "ollama (" + o.model + ")" }

type ollamaRequest struct {
	Model  string `json:"model"`
	System string `json:"system"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// Generate implements Provider, producing a full LinkedIn post.
func (o *Ollama) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	return o.complete(ctx, systemPrompt, buildPrompt(req))
}

// Suggest implements Provider, producing post-topic ideas.
func (o *Ollama) Suggest(ctx context.Context, category, query string) ([]string, error) {
	return suggestVia(ctx, o, category, query)
}

// complete performs a single system+user completion against Ollama's
// /api/generate endpoint.
func (o *Ollama) complete(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(ollamaRequest{
		Model:  o.model,
		System: system,
		Prompt: user,
		Stream: false,
	})
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("could not reach Ollama at %s (is it running?): %w", o.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != "" {
		return "", fmt.Errorf("ollama: %s", out.Error)
	}
	return out.Response, nil
}
