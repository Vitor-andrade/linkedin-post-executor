package store

import (
	"context"
	"strconv"
	"time"
)

// Settings keys backing the publish counters. They are bumped on every
// successful publish (immediate or scheduled), giving an accurate total
// regardless of whether a draft row was involved.
const (
	keyPublishedTotal  = "metric_published_total"
	keyLastPublishedAt = "metric_last_published_at"
)

// ScheduledStats counts the scheduling queue by status.
type ScheduledStats struct {
	Pending   int `json:"pending"`
	Published int `json:"published"`
	Failed    int `json:"failed"`
}

// Metrics is a local, privacy-preserving snapshot derived entirely from the
// user's own data — no telemetry leaves the machine.
type Metrics struct {
	DraftsTotal     int            `json:"draftsTotal"`
	DraftsByStatus  map[string]int `json:"draftsByStatus"`
	Scheduled       ScheduledStats `json:"scheduled"`
	PublishedTotal  int            `json:"publishedTotal"`
	LastPublishedAt *time.Time     `json:"lastPublishedAt,omitempty"`
}

// RecordPublished increments the lifetime publish counter and stores the time
// of the latest publish.
func (s *Store) RecordPublished(ctx context.Context, at time.Time) error {
	cur, err := s.GetSetting(ctx, keyPublishedTotal)
	if err != nil {
		return err
	}
	n, _ := strconv.Atoi(cur)
	if err := s.SetSetting(ctx, keyPublishedTotal, strconv.Itoa(n+1)); err != nil {
		return err
	}
	return s.SetSetting(ctx, keyLastPublishedAt, at.UTC().Format(time.RFC3339))
}

// Metrics aggregates drafts, the scheduling queue and the publish counters.
func (s *Store) Metrics(ctx context.Context) (Metrics, error) {
	m := Metrics{DraftsByStatus: map[string]int{}}

	if err := s.eachStatusCount(ctx, "drafts", func(status string, n int) {
		m.DraftsByStatus[status] = n
		m.DraftsTotal += n
	}); err != nil {
		return Metrics{}, err
	}

	if err := s.eachStatusCount(ctx, "scheduled_posts", func(status string, n int) {
		switch status {
		case "pending":
			m.Scheduled.Pending = n
		case "published":
			m.Scheduled.Published = n
		case "failed":
			m.Scheduled.Failed = n
		}
	}); err != nil {
		return Metrics{}, err
	}

	total, err := s.GetSetting(ctx, keyPublishedTotal)
	if err != nil {
		return Metrics{}, err
	}
	m.PublishedTotal, _ = strconv.Atoi(total)

	if last, err := s.GetSetting(ctx, keyLastPublishedAt); err != nil {
		return Metrics{}, err
	} else if last != "" {
		if t, perr := time.Parse(time.RFC3339, last); perr == nil {
			m.LastPublishedAt = &t
		}
	}

	return m, nil
}

// eachStatusCount runs "SELECT status, COUNT(*) ... GROUP BY status" over the
// given table and calls fn for each row. The table name is a trusted constant.
func (s *Store) eachStatusCount(ctx context.Context, table string, fn func(status string, n int)) error {
	rows, err := s.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM `+table+` GROUP BY status`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return err
		}
		fn(status, n)
	}
	return rows.Err()
}
