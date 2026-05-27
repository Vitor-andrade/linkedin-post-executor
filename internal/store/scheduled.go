package store

import (
	"context"
	"database/sql"
	"time"
)

// ScheduledPost represents a post queued to be published at a future time.
type ScheduledPost struct {
	ID            int64      `json:"id"`
	DraftID       *int64     `json:"draftId"`
	Content       string     `json:"content"`
	ScheduledFor  time.Time  `json:"scheduledFor"`
	Status        string     `json:"status"`
	LinkedInURN   string     `json:"linkedinUrn"`
	Error         string     `json:"error"`
	Attempts      int        `json:"attempts"`
	NextAttemptAt *time.Time `json:"nextAttemptAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// scheduledColumns is the canonical column list for the scans below.
const scheduledColumns = `id, draft_id, content, scheduled_for, status,
	linkedin_urn, error, attempts, next_attempt_at, created_at`

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanScheduled(sc scanner, p *ScheduledPost) error {
	var next sql.NullTime
	err := sc.Scan(&p.ID, &p.DraftID, &p.Content, &p.ScheduledFor, &p.Status,
		&p.LinkedInURN, &p.Error, &p.Attempts, &next, &p.CreatedAt)
	if next.Valid {
		t := next.Time
		p.NextAttemptAt = &t
	}
	return err
}

// CreateScheduledPost queues a post for future publishing (status=pending).
func (s *Store) CreateScheduledPost(ctx context.Context, p ScheduledPost) (ScheduledPost, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO scheduled_posts (draft_id, content, scheduled_for, status)
		 VALUES (?, ?, ?, 'pending')`,
		p.DraftID, p.Content, p.ScheduledFor,
	)
	if err != nil {
		return ScheduledPost{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return ScheduledPost{}, err
	}
	return s.GetScheduledPost(ctx, id)
}

// GetScheduledPost returns a single scheduled post by ID.
func (s *Store) GetScheduledPost(ctx context.Context, id int64) (ScheduledPost, error) {
	var p ScheduledPost
	err := scanScheduled(
		s.db.QueryRowContext(ctx, `SELECT `+scheduledColumns+` FROM scheduled_posts WHERE id = ?`, id),
		&p,
	)
	return p, err
}

// ListScheduledPosts returns all scheduled posts, soonest first.
func (s *Store) ListScheduledPosts(ctx context.Context) ([]ScheduledPost, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+scheduledColumns+` FROM scheduled_posts ORDER BY scheduled_for ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []ScheduledPost{}
	for rows.Next() {
		var p ScheduledPost
		if err := scanScheduled(rows, &p); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

// MarkScheduledPublished records a successful publish with its post URN.
func (s *Store) MarkScheduledPublished(ctx context.Context, id int64, urn string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE scheduled_posts SET status = 'published', linkedin_urn = ?, error = '' WHERE id = ?`,
		urn, id,
	)
	return err
}

// MarkScheduledFailed records a terminal failure (no further retries) with its
// error message.
func (s *Store) MarkScheduledFailed(ctx context.Context, id int64, msg string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE scheduled_posts SET status = 'failed', error = ? WHERE id = ?`,
		msg, id,
	)
	return err
}

// RecordFailedAttempt registers a retryable failure: it keeps the post pending,
// increments the attempt counter, stores the last error and schedules the next
// attempt for nextAttempt.
func (s *Store) RecordFailedAttempt(ctx context.Context, id int64, nextAttempt time.Time, msg string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE scheduled_posts
		 SET attempts = attempts + 1, next_attempt_at = ?, error = ?, status = 'pending'
		 WHERE id = ?`,
		nextAttempt, msg, id,
	)
	return err
}

// CancelScheduledPost deletes a still-pending scheduled post. It returns
// sql.ErrNoRows if the post does not exist or has already been processed.
func (s *Store) CancelScheduledPost(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM scheduled_posts WHERE id = ? AND status = 'pending'`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DueScheduledPosts returns pending posts whose effective time (the next
// retry time if set, otherwise the originally scheduled time) has passed.
func (s *Store) DueScheduledPosts(ctx context.Context, now time.Time) ([]ScheduledPost, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+scheduledColumns+`
		 FROM scheduled_posts
		 WHERE status = 'pending' AND COALESCE(next_attempt_at, scheduled_for) <= ?
		 ORDER BY COALESCE(next_attempt_at, scheduled_for) ASC`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []ScheduledPost{}
	for rows.Next() {
		var p ScheduledPost
		if err := scanScheduled(rows, &p); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
