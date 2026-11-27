package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

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
	status := v.Status
	if status == "" {
		status = appvideo.StatusConfirmed
	}
	query := `
		INSERT INTO videos (id, exercise_id, uploader_id, r2_key, filename, size_bytes, mime_type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		v.ID, v.ExerciseID, v.UploaderID, v.R2Key,
		v.Filename, v.SizeBytes, v.MimeType, status, v.CreatedAt,
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
		Status     string    `db:"status"`
		CreatedAt  time.Time `db:"created_at"`
	}
	query := `
		SELECT id, exercise_id, uploader_id, r2_key, filename,
		       COALESCE(size_bytes, 0) AS size_bytes,
		       COALESCE(mime_type, '') AS mime_type,
		       status, created_at
		FROM videos
		WHERE exercise_id = $1 AND status = 'CONFIRMED'
		ORDER BY created_at DESC
		LIMIT 1
	`
	var r2 row
	if err := getExecutor(ctx, r.db).GetContext(ctx, &r2, query, exerciseID); err != nil {
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
		Status:     r2.Status,
		CreatedAt:  r2.CreatedAt,
	}, nil
}

func (r *VideoRepository) GetByKeyAndExercise(ctx context.Context, exerciseID uuid.UUID, r2Key string) (*appvideo.VideoRecord, error) {
	type row struct {
		ID         uuid.UUID `db:"id"`
		ExerciseID uuid.UUID `db:"exercise_id"`
		UploaderID uuid.UUID `db:"uploader_id"`
		R2Key      string    `db:"r2_key"`
		Filename   string    `db:"filename"`
		SizeBytes  int64     `db:"size_bytes"`
		MimeType   string    `db:"mime_type"`
		Status     string    `db:"status"`
		CreatedAt  time.Time `db:"created_at"`
	}
	query := `
		SELECT id, exercise_id, uploader_id, r2_key, filename,
		       COALESCE(size_bytes, 0) AS size_bytes,
		       COALESCE(mime_type, '') AS mime_type,
		       status, created_at
		FROM videos
		WHERE exercise_id = $1 AND r2_key = $2
		LIMIT 1
	`
	var r2 row
	if err := getExecutor(ctx, r.db).GetContext(ctx, &r2, query, exerciseID, r2Key); err != nil {
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
		Status:     r2.Status,
		CreatedAt:  r2.CreatedAt,
	}, nil
}

func (r *VideoRepository) ConfirmPending(ctx context.Context, id uuid.UUID, sizeBytes int64, mimeType string) error {
	query := `
		UPDATE videos
		SET status = 'CONFIRMED', size_bytes = $1, mime_type = $2
		WHERE id = $3 AND status = 'PENDING'
	`
	res, err := getExecutor(ctx, r.db).ExecContext(ctx, query, sizeBytes, mimeType, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *VideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, `DELETE FROM videos WHERE id = $1`, id)
	return err
}

func (r *VideoRepository) HasConfirmedVideosByExerciseIDs(ctx context.Context, exerciseIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool)
	if len(exerciseIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(exerciseIDs))
	args := make([]any, 0, len(exerciseIDs))
	for i, id := range exerciseIDs {
		placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
		args = append(args, id)
	}

	query := `
		SELECT DISTINCT exercise_id
		FROM videos
		WHERE status = 'CONFIRMED'
		  AND exercise_id IN (` + strings.Join(placeholders, ", ") + `)
	`
	var foundIDs []uuid.UUID
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &foundIDs, query, args...); err != nil {
		return nil, err
	}
	for _, id := range foundIDs {
		result[id] = true
	}
	return result, nil
}
