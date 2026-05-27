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

// geminiBaseURL is the Generative Language API base; a package variable so
// tests can point it at a local server.
var geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Gemini generates posts with Google's Gemini models. It targets the free tier
// (e.g. gemini-2.0-flash), so users can bring a no-cost API key.
type Gemini struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGemini builds a Gemini-backed provider.
func NewGemini(apiKey, model string) *Gemini {
	return &Gemini{apiKey: apiKey, model: model, client: &http.Client{Timeout: 120 * time.Second}}
}

// Name implements Provider.
func (g *Gemini) Name() string { return "gemini (" + g.model + ")" }

type geminiRequest struct {
	SystemInstruction geminiContent   `json:"system_instruction"`
	Contents          []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Generate implements Provider by calling Gemini's generateContent endpoint.
func (g *Gemini) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	if g.apiKey == "" {
		return "", errors.New("Gemini API key missing: set LPE_GEMINI_API_KEY")
	}

	body, err := json.Marshal(geminiRequest{
		SystemInstruction: geminiContent{Parts: []geminiPart{{Text: systemPrompt}}},
		Contents:          []geminiContent{{Parts: []geminiPart{{Text: buildPrompt(req)}}}},
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiBaseURL, g.model, g.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("gemini: %s", out.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini returned status %d", resp.StatusCode)
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini: empty response")
	}
	return out.Candidates[0].Content.Parts[0].Text, nil
}
