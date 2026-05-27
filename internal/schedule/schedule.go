// Package schedule runs the background worker that publishes scheduled posts
// when their time arrives. It relies on goroutines, which is one of the main
// reasons Go was chosen for this project (see ADR-001).
package schedule

import (
	"context"
	"log"
	"time"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/store"
)

const (
	// tickInterval controls how often the scheduler checks for due posts.
	tickInterval = time.Minute
	// maxAttempts is the total number of publish tries before giving up.
	defaultMaxAttempts = 5
	// baseBackoff is the delay before the first retry; it doubles each attempt.
	defaultBaseBackoff = time.Minute
	// maxBackoff caps the exponential growth.
	maxBackoff = time.Hour
)

// Publisher publishes post text immediately and returns the created post URN.
// It is satisfied by *linkedin.Service, keeping the scheduler decoupled from
// the LinkedIn package.
type Publisher interface {
	PublishNow(ctx context.Context, text string) (string, error)
}

// Scheduler periodically publishes posts whose scheduled time has passed.
type Scheduler struct {
	store       *store.Store
	pub         Publisher
	maxAttempts int
	baseBackoff time.Duration
}

// New creates a Scheduler bound to the given store and publisher.
func New(st *store.Store, pub Publisher) *Scheduler {
	return &Scheduler{
		store:       st,
		pub:         pub,
		maxAttempts: defaultMaxAttempts,
		baseBackoff: defaultBaseBackoff,
	}
}

// backoff returns the delay before the given attempt number (1-based):
// baseBackoff * 2^(attempt-1), capped at maxBackoff. Doubling iteratively
// avoids shift overflow and keeps a zero base (used in tests) at zero.
func (s *Scheduler) backoff(attempt int) time.Duration {
	d := s.baseBackoff
	for i := 1; i < attempt; i++ {
		if d >= maxBackoff {
			return maxBackoff
		}
		d *= 2
	}
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}

// Start launches the scheduler loop in a background goroutine. It stops when
// ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

func (s *Scheduler) run(ctx context.Context) {
	log.Printf("scheduler started (interval %s)", tickInterval)
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// tick publishes any posts that are due. A failed attempt is retried with
// exponential backoff up to maxAttempts; after that the post is marked failed
// so it is never silently retried forever.
func (s *Scheduler) tick(ctx context.Context) {
	due, err := s.store.DueScheduledPosts(ctx, time.Now())
	if err != nil {
		log.Printf("scheduler: failed to fetch due posts: %v", err)
		return
	}
	for _, p := range due {
		urn, err := s.pub.PublishNow(ctx, p.Content)
		if err != nil {
			s.handleFailure(ctx, p, err)
			continue
		}
		if err := s.store.MarkScheduledPublished(ctx, p.ID, urn); err != nil {
			log.Printf("scheduler: could not mark post #%d published: %v", p.ID, err)
		}
		if p.DraftID != nil {
			_ = s.store.SetDraftStatus(ctx, *p.DraftID, "published")
		}
		log.Printf("scheduler: post #%d published as %s", p.ID, urn)
	}
}

// handleFailure decides whether a failed post is retried later or given up on.
func (s *Scheduler) handleFailure(ctx context.Context, p store.ScheduledPost, pubErr error) {
	attempt := p.Attempts + 1 // the attempt we just made
	if attempt >= s.maxAttempts {
		log.Printf("scheduler: post #%d failed after %d attempts, giving up: %v", p.ID, attempt, pubErr)
		if err := s.store.MarkScheduledFailed(ctx, p.ID, pubErr.Error()); err != nil {
			log.Printf("scheduler: could not mark post #%d failed: %v", p.ID, err)
		}
		return
	}
	next := time.Now().Add(s.backoff(attempt))
	log.Printf("scheduler: post #%d failed (attempt %d/%d), retrying around %s: %v",
		p.ID, attempt, s.maxAttempts, next.Format(time.RFC3339), pubErr)
	if err := s.store.RecordFailedAttempt(ctx, p.ID, next, pubErr.Error()); err != nil {
		log.Printf("scheduler: could not record retry for post #%d: %v", p.ID, err)
	}
}
