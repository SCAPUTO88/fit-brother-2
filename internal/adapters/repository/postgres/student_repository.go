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
	query := `INSERT INTO students (id, user_id, cpf_encrypted, cpf_hash, created_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.UserID, s.CPFEncrypted, s.CPFHash, s.CreatedAt,
	)
	return err
}

func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*student.Student, error) {
	query := `SELECT id, user_id, cpf_encrypted, cpf_hash, created_at FROM students WHERE id = $1`
	var std student.Student

	err := r.db.GetContext(ctx, &std, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &std, nil
}

func (r *StudentRepository) GetByCPFHash(ctx context.Context, cpfHash string) (*student.Student, error) {
	query := `SELECT id, user_id, cpf_encrypted, cpf_hash, created_at FROM students WHERE cpf_hash = $1`
	var std student.Student

	err := r.db.GetContext(ctx, &std, query, cpfHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &std, nil
}

// Delete faz a deleção. Pela LGPD nós também deveriamos zerar os CPFs aqui. O User recebe SoftDelete,
// a linha de Student pode ser apagada apagando também Biometrias ou zerando o Hash e EncryptedCpf pra texto generico.
func (r *StudentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *StudentRepository) SaveBiometric(ctx context.Context, b *student.BiometricData) error {
	query := `
		INSERT INTO biometric_data (id, student_id, weight_kg, height_cm, bmi, body_fat_percentage, recorded_by, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		b.ID, b.StudentID, b.WeightKg, b.HeightCm, b.BMI, b.BodyFatPercentage, b.RecordedBy, b.RecordedAt,
	)
	return err
}
