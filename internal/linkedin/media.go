package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// assetsEndpoint is the registerUpload API (overridable for tests).
var assetsEndpoint = "https://api.linkedin.com/v2/assets"

// UploadImage performs LinkedIn's two-step image upload: it registers an upload
// (reserving an asset URN) and then PUTs the bytes to the returned URL. The
// asset URN is what gets referenced when publishing.
func (c *Client) UploadImage(ctx context.Context, accessToken, ownerURN string, data []byte) (string, error) {
	uploadURL, asset, err := c.registerUpload(ctx, accessToken, ownerURN)
	if err != nil {
		return "", err
	}
	if err := c.putImageBytes(ctx, accessToken, uploadURL, data); err != nil {
		return "", err
	}
	return asset, nil
}

func (c *Client) registerUpload(ctx context.Context, accessToken, ownerURN string) (uploadURL, asset string, err error) {
	payload := map[string]any{
		"registerUploadRequest": map[string]any{
			"recipes": []string{"urn:li:digitalmediaRecipe:feedshare-image"},
			"owner":   ownerURN,
			"serviceRelationships": []map[string]any{
				{"relationshipType": "OWNER", "identifier": "urn:li:userGeneratedContent"},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, assetsEndpoint+"?action=registerUpload", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", "", fmt.Errorf("linkedin registerUpload (status %d): %s", resp.StatusCode, msg)
	}

	var out struct {
		Value struct {
			Asset           string `json:"asset"`
			UploadMechanism struct {
				Request struct {
					UploadURL string `json:"uploadUrl"`
				} `json:"com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest"`
			} `json:"uploadMechanism"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", err
	}
	uploadURL = out.Value.UploadMechanism.Request.UploadURL
	asset = out.Value.Asset
	if uploadURL == "" || asset == "" {
		return "", "", fmt.Errorf("linkedin registerUpload: missing upload URL or asset in response")
	}
	return uploadURL, asset, nil
}

func (c *Client) putImageBytes(ctx context.Context, accessToken, uploadURL string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("linkedin image upload (status %d): %s", resp.StatusCode, msg)
	}
	return nil
}
