package store

// schema is the initial database schema. It mirrors the ER diagram in
// docs/diagrams/er-diagram.mmd. For the foundation we apply it idempotently;
// a versioned migration system can replace this as the schema evolves.
const schema = `
CREATE TABLE IF NOT EXISTS drafts (
	id                 INTEGER PRIMARY KEY AUTOINCREMENT,
	title              TEXT NOT NULL,
	source_description TEXT NOT NULL DEFAULT '',
	content            TEXT NOT NULL DEFAULT '',
	status             TEXT NOT NULL DEFAULT 'draft',
	created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS draft_versions (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	draft_id     INTEGER NOT NULL REFERENCES drafts(id) ON DELETE CASCADE,
	content      TEXT NOT NULL,
	generated_by TEXT NOT NULL DEFAULT '',
	created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scheduled_posts (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	draft_id      INTEGER REFERENCES drafts(id) ON DELETE SET NULL,
	content       TEXT NOT NULL,
	scheduled_for DATETIME NOT NULL,
	status        TEXT NOT NULL DEFAULT 'pending',
	linkedin_urn  TEXT NOT NULL DEFAULT '',
	error         TEXT NOT NULL DEFAULT '',
	created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_scheduled_due
	ON scheduled_posts(status, scheduled_for);

CREATE TABLE IF NOT EXISTS oauth_tokens (
	id                INTEGER PRIMARY KEY AUTOINCREMENT,
	provider          TEXT NOT NULL UNIQUE,
	access_token_enc  BLOB,
	refresh_token_enc BLOB,
	expires_at        DATETIME,
	updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS settings (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT ''
);
`

func (s *Store) migrate() error {
	_, err := s.db.Exec(schema)
	return err
}
