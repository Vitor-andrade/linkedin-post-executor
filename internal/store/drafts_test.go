package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

// newTestStore opens a fresh SQLite store backed by a temp file that the test
// framework removes automatically.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestCreateAndGetDraft(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	created, err := st.CreateDraft(ctx, Draft{
		Title:             "Scaling Go APIs",
		SourceDescription: "lessons on concurrency",
		Content:           "𝗛𝗼𝗼𝗸 line\n\nbody",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected a generated id")
	}
	if created.Status != "draft" {
		t.Errorf("status = %q, want default %q", created.Status, "draft")
	}
	if created.CreatedAt.IsZero() {
		t.Error("created_at was not set")
	}

	got, err := st.GetDraft(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Scaling Go APIs" || got.Content != created.Content {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestUpdateDraft(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	created, err := st.CreateDraft(ctx, Draft{Title: "Original", Content: "v1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := st.UpdateDraft(ctx, created.ID, Draft{
		Title:   "Edited",
		Content: "v2",
		Status:  "ready",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Title != "Edited" || updated.Content != "v2" || updated.Status != "ready" {
		t.Errorf("update did not persist: %+v", updated)
	}
	if updated.ID != created.ID {
		t.Errorf("update changed the id: %d != %d", updated.ID, created.ID)
	}
}

func TestUpdateDraftUnknownID(t *testing.T) {
	st := newTestStore(t)
	_, err := st.UpdateDraft(context.Background(), 9999, Draft{Title: "x"})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestListDraftsNewestFirst(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	for _, title := range []string{"first", "second", "third"} {
		if _, err := st.CreateDraft(ctx, Draft{Title: title}); err != nil {
			t.Fatalf("create %q: %v", title, err)
		}
	}

	drafts, err := st.ListDrafts(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(drafts) != 3 {
		t.Fatalf("got %d drafts, want 3", len(drafts))
	}
	// Newest (highest id) first; ties on timestamp fall back to id order.
	if drafts[0].ID < drafts[len(drafts)-1].ID {
		t.Errorf("drafts not ordered newest-first: %+v", drafts)
	}
}
