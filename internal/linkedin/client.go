package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client performs the stateless HTTP calls against LinkedIn. It carries no
// credentials of its own; tokens are passed in per call.
type Client struct {
	http *http.Client
}

// NewClient builds a Client with sensible timeouts.
func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 30 * time.Second}}
}

// Token is an OAuth token set with an absolute expiry.
type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// Expired reports whether the access token is at or past its expiry (with a
// one-minute safety margin). A zero expiry is treated as non-expiring.
func (t Token) Expired() bool {
	return !t.Expiry.IsZero() && time.Now().After(t.Expiry.Add(-time.Minute))
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// Exchange swaps an authorization code for a token set.
func (c *Client) Exchange(ctx context.Context, cfg Config, code string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", cfg.RedirectURL)
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	return c.postToken(ctx, form)
}

// Refresh obtains a new access token from a refresh token.
func (c *Client) Refresh(ctx context.Context, cfg Config, refreshToken string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	return c.postToken(ctx, form)
}

func (c *Client) postToken(ctx context.Context, form url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return Token{}, fmt.Errorf("linkedin token endpoint: decode: %w", err)
	}
	if resp.StatusCode != http.StatusOK || tr.Error != "" {
		return Token{}, fmt.Errorf("linkedin token endpoint (status %d): %s %s", resp.StatusCode, tr.Error, tr.ErrorDesc)
	}

	tok := Token{AccessToken: tr.AccessToken, RefreshToken: tr.RefreshToken}
	if tr.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	return tok, nil
}

// AuthorURN resolves the authenticated member's URN (urn:li:person:{sub}) via
// the OpenID Connect userinfo endpoint.
func (c *Client) AuthorURN(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userinfoEndpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("linkedin userinfo: status %d", resp.StatusCode)
	}

	var u struct {
		Sub string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}
	if u.Sub == "" {
		return "", errors.New("linkedin userinfo: empty subject")
	}
	return "urn:li:person:" + u.Sub, nil
}

// Publish creates a post on the member's profile and returns its URN. When
// assetURNs are provided (already-uploaded images), it becomes an image post.
func (c *Client) Publish(ctx context.Context, accessToken, authorURN, text string, assetURNs []string) (string, error) {
	share := map[string]any{
		"shareCommentary":    map[string]any{"text": text},
		"shareMediaCategory": "NONE",
	}
	if len(assetURNs) > 0 {
		media := make([]map[string]any, 0, len(assetURNs))
		for _, a := range assetURNs {
			media = append(media, map[string]any{
				"status": "READY",
				"media":  a,
			})
		}
		share["shareMediaCategory"] = "IMAGE"
		share["media"] = media
	}

	payload := map[string]any{
		"author":         authorURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]any{
			"com.linkedin.ugc.ShareContent": share,
		},
		"visibility": map[string]any{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ugcPostsEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("linkedin publish (status %d): %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	// LinkedIn returns the post URN in the X-RestLi-Id header; fall back to the
	// "id" field in the JSON body.
	if id := resp.Header.Get("X-RestLi-Id"); id != "" {
		return id, nil
	}
	var out struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out.ID, nil
}
