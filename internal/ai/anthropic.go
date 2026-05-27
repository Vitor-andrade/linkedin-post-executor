package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// anthropicEndpoint is the Messages API URL (overridable for tests).
var anthropicEndpoint = "https://api.anthropic.com/v1/messages"

// Anthropic generates posts with Claude via the Messages API (bring your own
// key).
type Anthropic struct {
	apiKey string
	model  string
	client *http.Client
}

// NewAnthropic builds a Claude-backed provider.
func NewAnthropic(apiKey, model string) *Anthropic {
	return &Anthropic{apiKey: apiKey, model: model, client: &http.Client{Timeout: 120 * time.Second}}
}

// Name implements Provider.
func (a *Anthropic) Name() string { return "claude (" + a.model + ")" }

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Generate implements Provider, producing a full LinkedIn post.
func (a *Anthropic) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	return a.complete(ctx, systemPrompt, buildPrompt(req))
}

// Suggest implements Provider, producing post-topic ideas.
func (a *Anthropic) Suggest(ctx context.Context, category, query string) ([]Idea, error) {
	return suggestVia(ctx, a, category, query)
}

// complete performs a single system+user completion via the Messages API.
func (a *Anthropic) complete(ctx context.Context, system, user string) (string, error) {
	if a.apiKey == "" {
		return "", errors.New("Anthropic API key missing: set LPE_ANTHROPIC_API_KEY")
	}

	body, err := json.Marshal(anthropicRequest{
		Model:     a.model,
		MaxTokens: 2048,
		System:    system,
		Messages:  []anthropicMessage{{Role: "user", Content: user}},
	})
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("anthropic: %s", out.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic returned status %d", resp.StatusCode)
	}
	if len(out.Content) == 0 {
		return "", errors.New("anthropic: empty response")
	}
	return out.Content[0].Text, nil
}
