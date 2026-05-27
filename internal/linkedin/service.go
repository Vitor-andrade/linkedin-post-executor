package linkedin

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/secret"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/store"
)

// authorURNKey is the settings key under which the member URN is cached after
// the first connection, avoiding a userinfo round-trip on every publish.
const authorURNKey = "linkedin_author_urn"

// Service orchestrates the OAuth flow and publishing, persisting encrypted
// tokens through the store.
type Service struct {
	cfg    Config
	client *Client
	store  *store.Store
	cipher *secret.Cipher
}

// NewService wires the LinkedIn integration.
func NewService(cfg Config, st *store.Store, c *secret.Cipher) *Service {
	return &Service{cfg: cfg, client: NewClient(), store: st, cipher: c}
}

// Configured reports whether credentials are present.
func (s *Service) Configured() bool { return s.cfg.Configured() }

// AuthCodeURL returns the authorization URL for the given anti-CSRF state.
func (s *Service) AuthCodeURL(state string) string { return s.cfg.AuthCodeURL(state) }

// Status describes the current connection state for the UI.
type Status struct {
	Configured bool       `json:"configured"`
	Connected  bool       `json:"connected"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
}

// Status reports whether credentials are configured and a token is stored.
func (s *Service) Status(ctx context.Context) (Status, error) {
	st := Status{Configured: s.cfg.Configured()}
	rec, err := s.store.GetOAuthToken(ctx, provider)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	st.Connected = true
	if !rec.ExpiresAt.IsZero() {
		exp := rec.ExpiresAt
		st.ExpiresAt = &exp
	}
	return st, nil
}

// Connect exchanges the authorization code, resolves and caches the author
// URN, and stores the encrypted token.
func (s *Service) Connect(ctx context.Context, code string) error {
	tok, err := s.client.Exchange(ctx, s.cfg, code)
	if err != nil {
		return err
	}
	urn, err := s.client.AuthorURN(ctx, tok.AccessToken)
	if err != nil {
		return err
	}
	if err := s.saveToken(ctx, tok); err != nil {
		return err
	}
	return s.store.SetSetting(ctx, authorURNKey, urn)
}

// Disconnect removes the stored token.
func (s *Service) Disconnect(ctx context.Context) error {
	return s.store.DeleteOAuthToken(ctx, provider)
}

// PublishNow publishes text to the connected member's profile, refreshing the
// token first if needed, and returns the created post URN.
func (s *Service) PublishNow(ctx context.Context, text string) (string, error) {
	tok, err := s.validToken(ctx)
	if err != nil {
		return "", err
	}
	urn, err := s.store.GetSetting(ctx, authorURNKey)
	if err != nil {
		return "", err
	}
	if urn == "" {
		if urn, err = s.client.AuthorURN(ctx, tok.AccessToken); err != nil {
			return "", err
		}
		_ = s.store.SetSetting(ctx, authorURNKey, urn)
	}
	postURN, err := s.client.Publish(ctx, tok.AccessToken, urn, text)
	if err != nil {
		return "", err
	}
	// Best-effort local metric; never fail a publish over a counter.
	_ = s.store.RecordPublished(ctx, time.Now())
	return postURN, nil
}

// validToken loads the stored token and transparently refreshes it if expired.
func (s *Service) validToken(ctx context.Context) (Token, error) {
	tok, err := s.loadToken(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Token{}, errors.New("LinkedIn is not connected")
		}
		return Token{}, err
	}
	if !tok.Expired() {
		return tok, nil
	}
	if tok.RefreshToken == "" {
		return Token{}, errors.New("LinkedIn token expired and no refresh token is available; reconnect")
	}
	refreshed, err := s.client.Refresh(ctx, s.cfg, tok.RefreshToken)
	if err != nil {
		return Token{}, err
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = tok.RefreshToken // LinkedIn may omit it on refresh
	}
	if err := s.saveToken(ctx, refreshed); err != nil {
		return Token{}, err
	}
	return refreshed, nil
}

func (s *Service) saveToken(ctx context.Context, tok Token) error {
	access, err := s.cipher.Encrypt([]byte(tok.AccessToken))
	if err != nil {
		return err
	}
	var refresh []byte
	if tok.RefreshToken != "" {
		if refresh, err = s.cipher.Encrypt([]byte(tok.RefreshToken)); err != nil {
			return err
		}
	}
	return s.store.SaveOAuthToken(ctx, store.OAuthToken{
		Provider:     provider,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    tok.Expiry,
	})
}

func (s *Service) loadToken(ctx context.Context) (Token, error) {
	rec, err := s.store.GetOAuthToken(ctx, provider)
	if err != nil {
		return Token{}, err
	}
	access, err := s.cipher.Decrypt(rec.AccessToken)
	if err != nil {
		return Token{}, err
	}
	var refresh []byte
	if len(rec.RefreshToken) > 0 {
		if refresh, err = s.cipher.Decrypt(rec.RefreshToken); err != nil {
			return Token{}, err
		}
	}
	return Token{
		AccessToken:  string(access),
		RefreshToken: string(refresh),
		Expiry:       rec.ExpiresAt,
	}, nil
}
