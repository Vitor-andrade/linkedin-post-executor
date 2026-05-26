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

// Scheduler periodically publishes posts whose scheduled time has passed.
type Scheduler struct {
	store *store.Store
}

// New creates a Scheduler bound to the given store.
func New(st *store.Store) *Scheduler {
	return &Scheduler{store: st}
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

// tick processes any posts that are due. Publishing via the LinkedIn client
// will be wired in during the publish/schedule slices; for now it surfaces
// what would be published.
func (s *Scheduler) tick(ctx context.Context) {
	due, err := s.store.DueScheduledPosts(ctx, time.Now())
	if err != nil {
		log.Printf("scheduler: failed to fetch due posts: %v", err)
		return
	}
	for _, p := range due {
		log.Printf("scheduler: post #%d is due (publishing will be wired in the publish slice)", p.ID)
	}
}
