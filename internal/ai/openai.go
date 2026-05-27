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

// openaiEndpoint is the Chat Completions URL (overridable for tests).
var openaiEndpoint = "https://api.openai.com/v1/chat/completions"

// OpenAI generates posts via the Chat Completions API (bring your own key).
type OpenAI struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAI builds an OpenAI-backed provider.
func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{apiKey: apiKey, model: model, client: &http.Client{Timeout: 120 * time.Second}}
}

// Name implements Provider.
func (o *OpenAI) Name() string { return "openai (" + o.model + ")" }

type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Generate implements Provider by calling OpenAI's Chat Completions API.
func (o *OpenAI) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	if o.apiKey == "" {
		return "", errors.New("OpenAI API key missing: set LPE_OPENAI_API_KEY")
	}

	body, err := json.Marshal(openaiRequest{
		Model: o.model,
		Messages: []openaiMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildPrompt(req)},
		},
	})
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("openai: %s", out.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai returned status %d", resp.StatusCode)
	}
	if len(out.Choices) == 0 {
		return "", errors.New("openai: empty response")
	}
	return out.Choices[0].Message.Content, nil
}
