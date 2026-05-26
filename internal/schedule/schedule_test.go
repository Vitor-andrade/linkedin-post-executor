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

func TestTickMarksFailure(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	due, _ := st.CreateScheduledPost(ctx, store.ScheduledPost{
		Content:      "x",
		ScheduledFor: time.Now().Add(-time.Minute),
	})

	pub := &fakePublisher{err: errors.New("LinkedIn is not connected")}
	New(st, pub).tick(ctx)

	got, _ := st.GetScheduledPost(ctx, due.ID)
	if got.Status != "failed" || got.Error == "" {
		t.Errorf("expected failed status with error, got %+v", got)
	}
}
