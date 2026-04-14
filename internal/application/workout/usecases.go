package workout

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domain "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type CreatePlanUseCase struct {
	repo domain.Repository
}

func NewCreatePlanUseCase(r domain.Repository) *CreatePlanUseCase {
	return &CreatePlanUseCase{repo: r}
}

func (uc *CreatePlanUseCase) Execute(ctx context.Context, input CreatePlanInput) (*PlanOutput, error) {
	trainerID, err := uuid.Parse(input.TrainerID)
	if err != nil {
		return nil, errors.New("trainer_id inválido")
	}
	studentID, err := uuid.Parse(input.StudentID)
	if err != nil {
		return nil, errors.New("student_id inválido")
	}

	plan, err := domain.NewWorkoutPlan(trainerID, studentID, input.Title, input.Description)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.CreatePlan(ctx, plan); err != nil {
		return nil, errors.New("falha ao salvar plano de treino")
	}

	return &PlanOutput{
		ID:          plan.ID.String(),
		TrainerID:   plan.TrainerID.String(),
		StudentID:   plan.StudentID.String(),
		Title:       plan.Title,
		Description: plan.Description,
		IsActive:    plan.IsActive,
		CreatedAt:   plan.CreatedAt.Format(time.RFC3339),
	}, nil
}

type CreateSessionUseCase struct {
	repo domain.Repository
}

func NewCreateSessionUseCase(r domain.Repository) *CreateSessionUseCase {
	return &CreateSessionUseCase{repo: r}
}

func (uc *CreateSessionUseCase) Execute(ctx context.Context, input CreateSessionInput) (*SessionOutput, error) {
	planID, err := uuid.Parse(input.WorkoutPlanID)
	if err != nil {
		return nil, errors.New("workout_plan_id inválido")
	}

	plan, err := uc.repo.GetPlanByID(ctx, planID)
	if err != nil || plan == nil {
		return nil, errors.New("plano de treino não encontrado")
	}

	session, err := domain.NewWorkoutSession(planID, input.Title, input.DayOfWeek)
	if err != nil {
		return nil, err
	}
	session.Notes = input.Notes

	if err := uc.repo.CreateSession(ctx, session); err != nil {
		return nil, errors.New("falha ao salvar sessão")
	}

	return &SessionOutput{
		ID:            session.ID.String(),
		WorkoutPlanID: session.WorkoutPlanID.String(),
		Title:         session.Title,
		DayOfWeek:     session.DayOfWeek,
		Notes:         session.Notes,
		CreatedAt:     session.CreatedAt.Format(time.RFC3339),
	}, nil
}

type CreateExerciseUseCase struct {
	repo domain.Repository
}

func NewCreateExerciseUseCase(r domain.Repository) *CreateExerciseUseCase {
	return &CreateExerciseUseCase{repo: r}
}

func (uc *CreateExerciseUseCase) Execute(ctx context.Context, input CreateExerciseInput) (*ExerciseOutput, error) {
	sessionID, err := uuid.Parse(input.WorkoutSessionID)
	if err != nil {
		return nil, errors.New("workout_session_id inválido")
	}

	session, err := uc.repo.GetSessionByID(ctx, sessionID)
	if err != nil || session == nil {
		return nil, errors.New("sessão não encontrada")
	}

	exercise, err := domain.NewExercise(sessionID, input.Name, input.Sets, input.Reps, input.RestSeconds, input.Position)
	if err != nil {
		return nil, err
	}
	exercise.Notes = input.Notes
	exercise.VideoURL = input.VideoURL

	if err := uc.repo.CreateExercise(ctx, exercise); err != nil {
		return nil, errors.New("falha ao salvar exercício")
	}

	return &ExerciseOutput{
		ID:               exercise.ID.String(),
		WorkoutSessionID: exercise.WorkoutSessionID.String(),
		Name:             exercise.Name,
		Sets:             exercise.Sets,
		Reps:             exercise.Reps,
		RestSeconds:      exercise.RestSeconds,
		Notes:            exercise.Notes,
		VideoURL:         exercise.VideoURL,
		Position:         exercise.Position,
	}, nil
}

type ListPlansUseCase struct {
	repo domain.Repository
}

func NewListPlansUseCase(r domain.Repository) *ListPlansUseCase {
	return &ListPlansUseCase{repo: r}
}

func (uc *ListPlansUseCase) ByStudent(ctx context.Context, studentID uuid.UUID) ([]*PlanOutput, error) {
	plans, err := uc.repo.ListPlansByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	return toPlansOutput(plans), nil
}

func (uc *ListPlansUseCase) ByTrainer(ctx context.Context, trainerID uuid.UUID) ([]*PlanOutput, error) {
	plans, err := uc.repo.ListPlansByTrainer(ctx, trainerID)
	if err != nil {
		return nil, err
	}
	return toPlansOutput(plans), nil
}

func toPlansOutput(plans []*domain.WorkoutPlan) []*PlanOutput {
	out := make([]*PlanOutput, 0, len(plans))
	for _, p := range plans {
		out = append(out, &PlanOutput{
			ID:          p.ID.String(),
			TrainerID:   p.TrainerID.String(),
			StudentID:   p.StudentID.String(),
			Title:       p.Title,
			Description: p.Description,
			IsActive:    p.IsActive,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}
