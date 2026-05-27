package schedule

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/store"
)

// fakePublisher records calls and returns a canned result.
type fakePublisher struct {
	urn   string
	err   error
	texts []string
}

func (f *fakePublisher) PublishNow(_ context.Context, text string) (string, error) {
	f.texts = append(f.texts, text)
	return f.urn, f.err
}

func newStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestTickPublishesDuePost(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()

	due, _ := st.CreateScheduledPost(ctx, store.ScheduledPost{
		Content:      "ready to ship",
		ScheduledFor: time.Now().Add(-time.Minute),
	})
	// A future post must not be touched.
	future, _ := st.CreateScheduledPost(ctx, store.ScheduledPost{
		Content:      "later",
		ScheduledFor: time.Now().Add(time.Hour),
	})

	pub := &fakePublisher{urn: "urn:li:share:42"}
	New(st, pub).tick(ctx)

	if len(pub.texts) != 1 || pub.texts[0] != "ready to ship" {
		t.Fatalf("publisher calls = %v", pub.texts)
	}
	if got, _ := st.GetScheduledPost(ctx, due.ID); got.Status != "published" || got.LinkedInURN != "urn:li:share:42" {
		t.Errorf("due post not marked published: %+v", got)
	}
	if got, _ := st.GetScheduledPost(ctx, future.ID); got.Status != "pending" {
		t.Errorf("future post should stay pending: %+v", got)
	}
}

func TestTickRetriesBeforeGivingUp(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	due, _ := st.CreateScheduledPost(ctx, store.ScheduledPost{
		Content:      "x",
		ScheduledFor: time.Now().Add(-time.Minute),
	})

	pub := &fakePublisher{err: errors.New("LinkedIn is not connected")}
	s := New(st, pub)

	// First failure: stays pending, scheduled for a future retry.
	s.tick(ctx)
	got, _ := st.GetScheduledPost(ctx, due.ID)
	if got.Status != "pending" || got.Attempts != 1 {
		t.Fatalf("after first failure: status=%q attempts=%d", got.Status, got.Attempts)
	}
	if got.NextAttemptAt == nil || !got.NextAttemptAt.After(time.Now()) {
		t.Errorf("expected a future next_attempt_at, got %v", got.NextAttemptAt)
	}
	if got.Error == "" {
		t.Error("expected the last error to be recorded")
	}

	// The post is no longer due (its retry is in the future), so a second tick
	// must not increment attempts.
	s.tick(ctx)
	if again, _ := st.GetScheduledPost(ctx, due.ID); again.Attempts != 1 {
		t.Errorf("attempts should not advance while not due, got %d", again.Attempts)
	}
}

func TestTickGivesUpAfterMaxAttempts(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	due, _ := st.CreateScheduledPost(ctx, store.ScheduledPost{
		Content:      "x",
		ScheduledFor: time.Now().Add(-time.Minute),
	})

	// A scheduler that gives up immediately and retries with no delay so the
	// post stays due across ticks.
	pub := &fakePublisher{err: errors.New("boom")}
	s := New(st, pub)
	s.maxAttempts = 2
	s.baseBackoff = 0

	s.tick(ctx) // attempt 1 → retry (next_attempt_at = now)
	s.tick(ctx) // attempt 2 → reaches max → failed

	got, _ := st.GetScheduledPost(ctx, due.ID)
	if got.Status != "failed" {
		t.Errorf("expected failed after max attempts, got %q (attempts=%d)", got.Status, got.Attempts)
	}
}
