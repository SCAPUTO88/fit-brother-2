package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type WorkoutRepository struct {
	db *sqlx.DB
}

func NewWorkoutRepository(db *sqlx.DB) *WorkoutRepository {
	return &WorkoutRepository{db: db}
}

func (r *WorkoutRepository) CreatePlan(ctx context.Context, plan *domain.WorkoutPlan) error {
	query := `
		INSERT INTO workout_plans (id, trainer_id, student_id, title, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		plan.ID, plan.TrainerID, plan.StudentID, plan.Title,
		plan.Description, plan.IsActive, plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (r *WorkoutRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*domain.WorkoutPlan, error) {
	type row struct {
		ID          uuid.UUID `db:"id"`
		TrainerID   uuid.UUID `db:"trainer_id"`
		StudentID   uuid.UUID `db:"student_id"`
		Title       string    `db:"title"`
		Description string    `db:"description"`
		IsActive    bool      `db:"is_active"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}
	query := `SELECT id, trainer_id, student_id, title, description, is_active, created_at, updated_at FROM workout_plans WHERE id = $1`
	var r2 row
	if err := r.db.GetContext(ctx, &r2, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.WorkoutPlan{
		ID:          r2.ID,
		TrainerID:   r2.TrainerID,
		StudentID:   r2.StudentID,
		Title:       r2.Title,
		Description: r2.Description,
		IsActive:    r2.IsActive,
		CreatedAt:   r2.CreatedAt,
		UpdatedAt:   r2.UpdatedAt,
	}, nil
}

func (r *WorkoutRepository) ListPlansByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.WorkoutPlan, error) {
	return r.listPlans(ctx, "student_id", studentID)
}

func (r *WorkoutRepository) ListPlansByTrainer(ctx context.Context, trainerID uuid.UUID) ([]*domain.WorkoutPlan, error) {
	return r.listPlans(ctx, "trainer_id", trainerID)
}

func (r *WorkoutRepository) listPlans(ctx context.Context, column string, id uuid.UUID) ([]*domain.WorkoutPlan, error) {
	type row struct {
		ID          uuid.UUID `db:"id"`
		TrainerID   uuid.UUID `db:"trainer_id"`
		StudentID   uuid.UUID `db:"student_id"`
		Title       string    `db:"title"`
		Description string    `db:"description"`
		IsActive    bool      `db:"is_active"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}
	query := `SELECT id, trainer_id, student_id, title, description, is_active, created_at, updated_at FROM workout_plans WHERE ` + column + ` = $1 ORDER BY created_at DESC`
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, id); err != nil {
		return nil, err
	}
	plans := make([]*domain.WorkoutPlan, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, &domain.WorkoutPlan{
			ID:          row.ID,
			TrainerID:   row.TrainerID,
			StudentID:   row.StudentID,
			Title:       row.Title,
			Description: row.Description,
			IsActive:    row.IsActive,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return plans, nil
}

func (r *WorkoutRepository) UpdatePlan(ctx context.Context, plan *domain.WorkoutPlan) error {
	query := `UPDATE workout_plans SET title=$1, description=$2, is_active=$3, updated_at=$4 WHERE id=$5`
	_, err := r.db.ExecContext(ctx, query, plan.Title, plan.Description, plan.IsActive, time.Now(), plan.ID)
	return err
}

func (r *WorkoutRepository) CreateSession(ctx context.Context, session *domain.WorkoutSession) error {
	query := `
		INSERT INTO workout_sessions (id, workout_plan_id, title, day_of_week, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.WorkoutPlanID, session.Title,
		session.DayOfWeek, session.Notes, session.CreatedAt,
	)
	return err
}

func (r *WorkoutRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.WorkoutSession, error) {
	type row struct {
		ID            uuid.UUID `db:"id"`
		WorkoutPlanID uuid.UUID `db:"workout_plan_id"`
		Title         string    `db:"title"`
		DayOfWeek     int       `db:"day_of_week"`
		Notes         string    `db:"notes"`
		CreatedAt     time.Time `db:"created_at"`
	}
	query := `SELECT id, workout_plan_id, title, day_of_week, notes, created_at FROM workout_sessions WHERE id = $1`
	var r2 row
	if err := r.db.GetContext(ctx, &r2, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.WorkoutSession{
		ID:            r2.ID,
		WorkoutPlanID: r2.WorkoutPlanID,
		Title:         r2.Title,
		DayOfWeek:     r2.DayOfWeek,
		Notes:         r2.Notes,
		CreatedAt:     r2.CreatedAt,
	}, nil
}

func (r *WorkoutRepository) ListSessionsByPlan(ctx context.Context, planID uuid.UUID) ([]*domain.WorkoutSession, error) {
	type row struct {
		ID            uuid.UUID `db:"id"`
		WorkoutPlanID uuid.UUID `db:"workout_plan_id"`
		Title         string    `db:"title"`
		DayOfWeek     int       `db:"day_of_week"`
		Notes         string    `db:"notes"`
		CreatedAt     time.Time `db:"created_at"`
	}
	query := `SELECT id, workout_plan_id, title, day_of_week, notes, created_at FROM workout_sessions WHERE workout_plan_id = $1 ORDER BY day_of_week, created_at`
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, planID); err != nil {
		return nil, err
	}
	sessions := make([]*domain.WorkoutSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, &domain.WorkoutSession{
			ID:            row.ID,
			WorkoutPlanID: row.WorkoutPlanID,
			Title:         row.Title,
			DayOfWeek:     row.DayOfWeek,
			Notes:         row.Notes,
			CreatedAt:     row.CreatedAt,
		})
	}
	return sessions, nil
}

func (r *WorkoutRepository) CreateExercise(ctx context.Context, exercise *domain.Exercise) error {
	query := `
		INSERT INTO exercises (id, workout_session_id, name, sets, reps, rest_seconds, notes, video_url, position, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		exercise.ID, exercise.WorkoutSessionID, exercise.Name,
		exercise.Sets, exercise.Reps, exercise.RestSeconds,
		exercise.Notes, exercise.VideoURL, exercise.Position, exercise.CreatedAt,
	)
	return err
}

func (r *WorkoutRepository) ListExercisesBySession(ctx context.Context, sessionID uuid.UUID) ([]*domain.Exercise, error) {
	type row struct {
		ID               uuid.UUID `db:"id"`
		WorkoutSessionID uuid.UUID `db:"workout_session_id"`
		Name             string    `db:"name"`
		Sets             int       `db:"sets"`
		Reps             string    `db:"reps"`
		RestSeconds      int       `db:"rest_seconds"`
		Notes            string    `db:"notes"`
		VideoURL         string    `db:"video_url"`
		Position         int       `db:"position"`
		CreatedAt        time.Time `db:"created_at"`
	}
	query := `SELECT id, workout_session_id, name, sets, reps, rest_seconds, notes, video_url, position, created_at FROM exercises WHERE workout_session_id = $1 ORDER BY position`
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, sessionID); err != nil {
		return nil, err
	}
	exercises := make([]*domain.Exercise, 0, len(rows))
	for _, row := range rows {
		exercises = append(exercises, &domain.Exercise{
			ID:               row.ID,
			WorkoutSessionID: row.WorkoutSessionID,
			Name:             row.Name,
			Sets:             row.Sets,
			Reps:             row.Reps,
			RestSeconds:      row.RestSeconds,
			Notes:            row.Notes,
			VideoURL:         row.VideoURL,
			Position:         row.Position,
			CreatedAt:        row.CreatedAt,
		})
	}
	return exercises, nil
}

func (r *WorkoutRepository) DeleteExercise(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM exercises WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
