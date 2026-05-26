package store

import (
	"context"
	"time"
)

// ScheduledPost represents a post queued to be published at a future time.
type ScheduledPost struct {
	ID           int64     `json:"id"`
	DraftID      *int64    `json:"draftId"`
	Content      string    `json:"content"`
	ScheduledFor time.Time `json:"scheduledFor"`
	Status       string    `json:"status"`
	LinkedInURN  string    `json:"linkedinUrn"`
	Error        string    `json:"error"`
	CreatedAt    time.Time `json:"createdAt"`
}

// DueScheduledPosts returns pending posts whose scheduled time has passed.
func (s *Store) DueScheduledPosts(ctx context.Context, now time.Time) ([]ScheduledPost, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, draft_id, content, scheduled_for, status, linkedin_urn, error, created_at
		 FROM scheduled_posts
		 WHERE status = 'pending' AND scheduled_for <= ?
		 ORDER BY scheduled_for ASC`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []ScheduledPost{}
	for rows.Next() {
		var p ScheduledPost
		if err := rows.Scan(&p.ID, &p.DraftID, &p.Content, &p.ScheduledFor, &p.Status, &p.LinkedInURN, &p.Error, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
