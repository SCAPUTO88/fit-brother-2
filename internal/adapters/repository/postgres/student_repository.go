package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/domain/student"
)

type StudentRepository struct {
	db *sqlx.DB
}

func NewStudentRepository(db *sqlx.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) Create(ctx context.Context, s *student.Student) error {
	query := `INSERT INTO students (id, user_id, trainer_id, cpf_encrypted, cpf_hash, created_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		s.ID, s.UserID, s.TrainerID, s.CPFEncrypted, s.CPFHash, s.CreatedAt,
	)
	return err
}

func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*student.Student, error) {
	query := `SELECT id, user_id, trainer_id, cpf_encrypted, cpf_hash, created_at FROM students WHERE id = $1`
	var std student.Student

	err := getExecutor(ctx, r.db).GetContext(ctx, &std, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &std, nil
}

func (r *StudentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*student.Student, error) {
	query := `SELECT id, user_id, trainer_id, cpf_encrypted, cpf_hash, created_at FROM students WHERE user_id = $1`
	var std student.Student

	err := getExecutor(ctx, r.db).GetContext(ctx, &std, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &std, nil
}

func (r *StudentRepository) GetByCPFHash(ctx context.Context, cpfHash string) (*student.Student, error) {
	query := `SELECT id, user_id, trainer_id, cpf_encrypted, cpf_hash, created_at FROM students WHERE cpf_hash = $1`
	var std student.Student

	err := getExecutor(ctx, r.db).GetContext(ctx, &std, query, cpfHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &std, nil
}

func (r *StudentRepository) ListByTrainer(ctx context.Context, trainerID uuid.UUID) ([]*student.StudentBaseInfo, error) {
	query := `
		SELECT s.id, u.full_name, u.email
		FROM students s
		INNER JOIN users u ON s.user_id = u.id
		WHERE s.trainer_id = $1 AND u.deleted_at IS NULL AND u.is_active = true
		ORDER BY u.full_name ASC
	`
	var items []*student.StudentBaseInfo
	err := getExecutor(ctx, r.db).SelectContext(ctx, &items, query, trainerID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*student.StudentBaseInfo, 0)
	}
	return items, nil
}

func (r *StudentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query, id)
	return err
}

func (r *StudentRepository) SaveBiometric(ctx context.Context, b *student.BiometricData) error {
	query := `
		INSERT INTO biometric_data (id, student_id, weight_kg, height_cm, bmi, body_fat_percentage, recorded_by, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		b.ID, b.StudentID, b.WeightKg, b.HeightCm, b.BMI, b.BodyFatPercentage, b.RecordedBy, b.RecordedAt,
	)
	return err
}
