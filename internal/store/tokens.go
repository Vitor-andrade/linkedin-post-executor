package store

import (
	"context"
	"time"
)

// OAuthToken holds the (already encrypted) OAuth credentials for a provider.
// The access/refresh tokens are stored as ciphertext; encryption is handled by
// the caller (see internal/secret), so the store never sees plaintext secrets.
type OAuthToken struct {
	Provider     string
	AccessToken  []byte // encrypted
	RefreshToken []byte // encrypted (may be empty)
	ExpiresAt    time.Time
	UpdatedAt    time.Time
}

// SaveOAuthToken upserts the token for its provider.
func (s *Store) SaveOAuthToken(ctx context.Context, t OAuthToken) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO oauth_tokens (provider, access_token_enc, refresh_token_enc, expires_at, updated_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(provider) DO UPDATE SET
		   access_token_enc  = excluded.access_token_enc,
		   refresh_token_enc = excluded.refresh_token_enc,
		   expires_at        = excluded.expires_at,
		   updated_at        = CURRENT_TIMESTAMP`,
		t.Provider, t.AccessToken, t.RefreshToken, t.ExpiresAt,
	)
	return err
}

// GetOAuthToken returns the stored token for provider, or sql.ErrNoRows if the
// provider has not been connected yet.
func (s *Store) GetOAuthToken(ctx context.Context, provider string) (OAuthToken, error) {
	var t OAuthToken
	err := s.db.QueryRowContext(ctx,
		`SELECT provider, access_token_enc, refresh_token_enc, expires_at, updated_at
		 FROM oauth_tokens WHERE provider = ?`, provider,
	).Scan(&t.Provider, &t.AccessToken, &t.RefreshToken, &t.ExpiresAt, &t.UpdatedAt)
	return t, err
}

// DeleteOAuthToken removes the stored token for provider (used on disconnect).
func (s *Store) DeleteOAuthToken(ctx context.Context, provider string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM oauth_tokens WHERE provider = ?`, provider)
	return err
}
