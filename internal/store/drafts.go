package store

import (
	"context"
	"time"
)

// Draft is a (possibly AI-generated) LinkedIn post the user is working on.
type Draft struct {
	ID                int64     `json:"id"`
	Title             string    `json:"title"`
	SourceDescription string    `json:"sourceDescription"`
	Content           string    `json:"content"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// CreateDraft inserts a new draft and returns it with its generated ID.
func (s *Store) CreateDraft(ctx context.Context, d Draft) (Draft, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO drafts (title, source_description, content, status)
		 VALUES (?, ?, ?, COALESCE(NULLIF(?, ''), 'draft'))`,
		d.Title, d.SourceDescription, d.Content, d.Status,
	)
	if err != nil {
		return Draft{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Draft{}, err
	}
	return s.GetDraft(ctx, id)
}

// GetDraft returns a single draft by ID.
func (s *Store) GetDraft(ctx context.Context, id int64) (Draft, error) {
	var d Draft
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, source_description, content, status, created_at, updated_at
		 FROM drafts WHERE id = ?`, id,
	).Scan(&d.ID, &d.Title, &d.SourceDescription, &d.Content, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

// ListDrafts returns all drafts, newest first.
func (s *Store) ListDrafts(ctx context.Context) ([]Draft, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, source_description, content, status, created_at, updated_at
		 FROM drafts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	drafts := []Draft{}
	for rows.Next() {
		var d Draft
		if err := rows.Scan(&d.ID, &d.Title, &d.SourceDescription, &d.Content, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		drafts = append(drafts, d)
	}
	return drafts, rows.Err()
}
