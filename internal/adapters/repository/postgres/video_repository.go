package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	appvideo "github.com/SCAPUTO88/fit-brother/internal/application/video"
)

type VideoRepository struct {
	db *sqlx.DB
}

func NewVideoRepository(db *sqlx.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Save(ctx context.Context, v *appvideo.VideoRecord) error {
	query := `
		INSERT INTO videos (id, exercise_id, uploader_id, r2_key, filename, size_bytes, mime_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.ExerciseID, v.UploaderID, v.R2Key,
		v.Filename, v.SizeBytes, v.MimeType, v.CreatedAt,
	)
	return err
}

func (r *VideoRepository) GetByExerciseID(ctx context.Context, exerciseID uuid.UUID) (*appvideo.VideoRecord, error) {
	type row struct {
		ID         uuid.UUID `db:"id"`
		ExerciseID uuid.UUID `db:"exercise_id"`
		UploaderID uuid.UUID `db:"uploader_id"`
		R2Key      string    `db:"r2_key"`
		Filename   string    `db:"filename"`
		SizeBytes  int64     `db:"size_bytes"`
		MimeType   string    `db:"mime_type"`
	}
	query := `SELECT id, exercise_id, uploader_id, r2_key, filename, size_bytes, mime_type FROM videos WHERE exercise_id = $1 LIMIT 1`
	var r2 row
	if err := r.db.GetContext(ctx, &r2, query, exerciseID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &appvideo.VideoRecord{
		ID:         r2.ID,
		ExerciseID: r2.ExerciseID,
		UploaderID: r2.UploaderID,
		R2Key:      r2.R2Key,
		Filename:   r2.Filename,
		SizeBytes:  r2.SizeBytes,
		MimeType:   r2.MimeType,
	}, nil
}

func (r *VideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM videos WHERE id = $1`, id)
	return err
}
