// Package linkedin implements the "bring your own credentials" integration
// with LinkedIn: the OAuth authorization-code flow and immediate publishing to
// the member's own profile via w_member_social (see ADR-004). Tokens are
// encrypted at rest; the member never publishes on behalf of anyone else.
package linkedin

import (
	"net/url"
	"os"
)

// provider is the key under which LinkedIn tokens are stored.
const provider = "linkedin"

// scopes requested during authorization: OpenID Connect to resolve the author
// URN, plus w_member_social to publish.
const scopes = "openid profile w_member_social"

// Endpoints are package variables so tests can point them at a local server.
var (
	authorizeEndpoint = "https://www.linkedin.com/oauth/v2/authorization"
	tokenEndpoint     = "https://www.linkedin.com/oauth/v2/accessToken"
	userinfoEndpoint  = "https://api.linkedin.com/v2/userinfo"
	ugcPostsEndpoint  = "https://api.linkedin.com/v2/ugcPosts"
)

// Config holds the user's LinkedIn app credentials (from the environment).
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// ConfigFromEnv reads credentials from LPE_LINKEDIN_* variables, defaulting the
// redirect URL to the local callback.
func ConfigFromEnv() Config {
	return Config{
		ClientID:     os.Getenv("LPE_LINKEDIN_CLIENT_ID"),
		ClientSecret: os.Getenv("LPE_LINKEDIN_CLIENT_SECRET"),
		RedirectURL:  envOr("LPE_LINKEDIN_REDIRECT_URL", "http://localhost:8080/api/linkedin/callback"),
	}
}

// Configured reports whether the mandatory credentials are present.
func (c Config) Configured() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// AuthCodeURL builds the authorization URL the user is redirected to, carrying
// an anti-CSRF state value.
func (c Config) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.RedirectURL)
	q.Set("scope", scopes)
	q.Set("state", state)
	return authorizeEndpoint + "?" + q.Encode()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
