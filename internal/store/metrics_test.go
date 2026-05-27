package store

import (
	"context"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	// Two drafts (default status "draft") and one published draft.
	if _, err := st.CreateDraft(ctx, Draft{Title: "a"}); err != nil {
		t.Fatal(err)
	}
	pub, err := st.CreateDraft(ctx, Draft{Title: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetDraftStatus(ctx, pub.ID, "published"); err != nil {
		t.Fatal(err)
	}

	// Scheduling queue: one pending, one failed.
	if _, err := st.CreateScheduledPost(ctx, ScheduledPost{Content: "p", ScheduledFor: time.Now()}); err != nil {
		t.Fatal(err)
	}
	failed, _ := st.CreateScheduledPost(ctx, ScheduledPost{Content: "q", ScheduledFor: time.Now()})
	if err := st.MarkScheduledFailed(ctx, failed.ID, "boom"); err != nil {
		t.Fatal(err)
	}

	// Two successful publishes recorded.
	now := time.Now()
	for range 2 {
		if err := st.RecordPublished(ctx, now); err != nil {
			t.Fatal(err)
		}
	}

	m, err := st.Metrics(ctx)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}

	if m.DraftsTotal != 2 {
		t.Errorf("draftsTotal = %d, want 2", m.DraftsTotal)
	}
	if m.DraftsByStatus["draft"] != 1 || m.DraftsByStatus["published"] != 1 {
		t.Errorf("draftsByStatus = %v", m.DraftsByStatus)
	}
	if m.Scheduled.Pending != 1 || m.Scheduled.Failed != 1 || m.Scheduled.Published != 0 {
		t.Errorf("scheduled = %+v", m.Scheduled)
	}
	if m.PublishedTotal != 2 {
		t.Errorf("publishedTotal = %d, want 2", m.PublishedTotal)
	}
	if m.LastPublishedAt == nil {
		t.Error("expected lastPublishedAt to be set")
	}
}

func TestMetricsEmpty(t *testing.T) {
	m, err := newTestStore(t).Metrics(context.Background())
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if m.DraftsTotal != 0 || m.PublishedTotal != 0 || m.LastPublishedAt != nil {
		t.Errorf("expected zero-valued metrics, got %+v", m)
	}
}
