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

// tickInterval controls how often the scheduler checks for due posts.
const tickInterval = time.Minute

// Publisher publishes post text immediately and returns the created post URN.
// It is satisfied by *linkedin.Service, keeping the scheduler decoupled from
// the LinkedIn package.
type Publisher interface {
	PublishNow(ctx context.Context, text string) (string, error)
}

// Scheduler periodically publishes posts whose scheduled time has passed.
type Scheduler struct {
	store *store.Store
	pub   Publisher
}

// New creates a Scheduler bound to the given store and publisher.
func New(st *store.Store, pub Publisher) *Scheduler {
	return &Scheduler{store: st, pub: pub}
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

// tick publishes any posts that are due, recording the outcome on each so a
// failure (e.g. LinkedIn not connected) is visible and never silently retried
// forever.
func (s *Scheduler) tick(ctx context.Context) {
	due, err := s.store.DueScheduledPosts(ctx, time.Now())
	if err != nil {
		log.Printf("scheduler: failed to fetch due posts: %v", err)
		return
	}
	for _, p := range due {
		urn, err := s.pub.PublishNow(ctx, p.Content)
		if err != nil {
			log.Printf("scheduler: post #%d failed: %v", p.ID, err)
			if err := s.store.MarkScheduledFailed(ctx, p.ID, err.Error()); err != nil {
				log.Printf("scheduler: could not mark post #%d failed: %v", p.ID, err)
			}
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
