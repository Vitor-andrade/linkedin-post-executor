package store

import "context"

// Image is an uploaded picture stored as a blob in SQLite, kept until it is
// attached to a published post.
type Image struct {
	ID        int64  `json:"id"`
	Mime      string `json:"mime"`
	Filename  string `json:"filename"`
	Data      []byte `json:"-"`
	CreatedAt string `json:"createdAt"`
}

// SaveImage stores the bytes of an uploaded image and returns its ID.
func (s *Store) SaveImage(ctx context.Context, mime, filename string, data []byte) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO images (mime, filename, data) VALUES (?, ?, ?)`,
		mime, filename, data,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetImage returns the full image (including bytes) by ID.
func (s *Store) GetImage(ctx context.Context, id int64) (Image, error) {
	var img Image
	err := s.db.QueryRowContext(ctx,
		`SELECT id, mime, filename, data, created_at FROM images WHERE id = ?`, id,
	).Scan(&img.ID, &img.Mime, &img.Filename, &img.Data, &img.CreatedAt)
	return img, err
}
