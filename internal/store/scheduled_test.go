package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestScheduledLifecycle(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	past := time.Now().Add(-time.Minute)
	future := time.Now().Add(time.Hour)

	duePost, err := st.CreateScheduledPost(ctx, ScheduledPost{Content: "due now", ScheduledFor: past})
	if err != nil {
		t.Fatalf("create due: %v", err)
	}
	if duePost.Status != "pending" {
		t.Errorf("status = %q, want pending", duePost.Status)
	}
	if _, err := st.CreateScheduledPost(ctx, ScheduledPost{Content: "later", ScheduledFor: future}); err != nil {
		t.Fatalf("create future: %v", err)
	}

	if all, err := st.ListScheduledPosts(ctx); err != nil || len(all) != 2 {
		t.Fatalf("list: got %d posts, err %v", len(all), err)
	}

	due, err := st.DueScheduledPosts(ctx, time.Now())
	if err != nil {
		t.Fatalf("due: %v", err)
	}
	if len(due) != 1 || due[0].ID != duePost.ID {
		t.Fatalf("expected only the past post to be due, got %+v", due)
	}

	if err := st.MarkScheduledPublished(ctx, duePost.ID, "urn:li:share:1"); err != nil {
		t.Fatalf("mark published: %v", err)
	}
	got, _ := st.GetScheduledPost(ctx, duePost.ID)
	if got.Status != "published" || got.LinkedInURN != "urn:li:share:1" {
		t.Errorf("after publish: %+v", got)
	}

	// A published post is no longer due.
	if due, _ := st.DueScheduledPosts(ctx, time.Now()); len(due) != 0 {
		t.Errorf("published post still due: %+v", due)
	}
}

func TestMarkScheduledFailed(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	p, _ := st.CreateScheduledPost(ctx, ScheduledPost{Content: "x", ScheduledFor: time.Now()})

	if err := st.MarkScheduledFailed(ctx, p.ID, "boom"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	got, _ := st.GetScheduledPost(ctx, p.ID)
	if got.Status != "failed" || got.Error != "boom" {
		t.Errorf("after fail: %+v", got)
	}
}

func TestCancelScheduledPost(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	p, _ := st.CreateScheduledPost(ctx, ScheduledPost{Content: "x", ScheduledFor: time.Now().Add(time.Hour)})

	if err := st.CancelScheduledPost(ctx, p.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := st.GetScheduledPost(ctx, p.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("post should be gone after cancel, err = %v", err)
	}
	// Cancelling again (or an unknown id) reports ErrNoRows.
	if err := st.CancelScheduledPost(ctx, p.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("re-cancel err = %v, want sql.ErrNoRows", err)
	}
}
