package workout

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreatePlan(ctx context.Context, plan *WorkoutPlan) error
	GetPlanByID(ctx context.Context, id uuid.UUID) (*WorkoutPlan, error)
	ListPlansByStudent(ctx context.Context, studentID uuid.UUID) ([]*WorkoutPlan, error)
	ListPlansByTrainer(ctx context.Context, trainerID uuid.UUID) ([]*WorkoutPlan, error)
	UpdatePlan(ctx context.Context, plan *WorkoutPlan) error

	CreateSession(ctx context.Context, session *WorkoutSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*WorkoutSession, error)
	ListSessionsByPlan(ctx context.Context, planID uuid.UUID) ([]*WorkoutSession, error)

	CreateExercise(ctx context.Context, exercise *Exercise) error
	GetExerciseByID(ctx context.Context, id uuid.UUID) (*Exercise, error)
	ListExercisesBySession(ctx context.Context, sessionID uuid.UUID) ([]*Exercise, error)
	DeleteExercise(ctx context.Context, id uuid.UUID) error
}
