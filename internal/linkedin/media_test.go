package linkedin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadImage(t *testing.T) {
	var uploaded []byte

	// The upload target the registerUpload step points to.
	uploadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("upload method = %s, want PUT", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing bearer on upload")
		}
		uploaded, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer uploadSrv.Close()

	// The assets registerUpload endpoint.
	assetsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "registerUpload" {
			t.Errorf("missing action=registerUpload: %s", r.URL.RawQuery)
		}
		resp := map[string]any{
			"value": map[string]any{
				"asset": "urn:li:digitalmediaAsset:XYZ",
				"uploadMechanism": map[string]any{
					"com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest": map[string]any{
						"uploadUrl": uploadSrv.URL,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer assetsSrv.Close()
	restore := assetsEndpoint
	assetsEndpoint = assetsSrv.URL
	defer func() { assetsEndpoint = restore }()

	asset, err := NewClient().UploadImage(context.Background(), "tok", "urn:li:person:ME", []byte("PNGDATA"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if asset != "urn:li:digitalmediaAsset:XYZ" {
		t.Errorf("asset = %q", asset)
	}
	if string(uploaded) != "PNGDATA" {
		t.Errorf("uploaded bytes = %q", uploaded)
	}
}
