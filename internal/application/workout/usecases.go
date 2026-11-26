package workout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	domain "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type CreatePlanUseCase struct {
	repo     domain.Repository
	studRepo domainS.Repository
	orgRepo  domainOrg.Repository
}

func NewCreatePlanUseCase(r domain.Repository, sr domainS.Repository) *CreatePlanUseCase {
	return &CreatePlanUseCase{repo: r, studRepo: sr}
}

func (uc *CreatePlanUseCase) WithOrganizationRepo(orgRepo domainOrg.Repository) *CreatePlanUseCase {
	uc.orgRepo = orgRepo
	return uc
}

func (uc *CreatePlanUseCase) Execute(ctx context.Context, input CreatePlanInput) (*PlanOutput, error) {
	trainerID, err := uuid.Parse(input.TrainerID)
	if err != nil {
		return nil, fmt.Errorf("%w: trainer_id inválido", errs.ErrInvalidInput)
	}
	studentID, err := uuid.Parse(input.StudentID)
	if err != nil {
		return nil, fmt.Errorf("%w: student_id inválido", errs.ErrInvalidInput)
	}

	stud, err := uc.studRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("falha ao verificar vínculo do aluno")
	}
	if stud == nil {
		return nil, fmt.Errorf("%w: aluno não encontrado", errs.ErrNotFound)
	}
	if uc.orgRepo != nil {
		active, err := uc.orgRepo.HasActiveRelationship(ctx, trainerID, studentID)
		if err != nil {
			return nil, errors.New("falha ao verificar vínculo ativo com o aluno")
		}
		if !active {
			return nil, fmt.Errorf("%w: este aluno não possui vínculo ativo com o treinador", errs.ErrForbidden)
		}
	} else if stud.TrainerID == nil || *stud.TrainerID != trainerID {
		return nil, fmt.Errorf("%w: este aluno não pertence ao treinador autenticado", errs.ErrForbidden)
	}

	plan, err := domain.NewWorkoutPlan(trainerID, studentID, input.Title, input.Description)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
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

func (uc *CreateSessionUseCase) Execute(ctx context.Context, trainerID uuid.UUID, input CreateSessionInput) (*SessionOutput, error) {
	planID, err := uuid.Parse(input.WorkoutPlanID)
	if err != nil {
		return nil, fmt.Errorf("%w: workout_plan_id inválido", errs.ErrInvalidInput)
	}

	plan, err := uc.repo.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, errors.New("falha ao buscar plano de treino")
	}
	if plan == nil {
		return nil, fmt.Errorf("%w: plano de treino não encontrado", errs.ErrNotFound)
	}
	if plan.TrainerID != trainerID {
		return nil, fmt.Errorf("%w: acesso negado ao plano de outro treinador", errs.ErrForbidden)
	}

	session, err := domain.NewWorkoutSession(planID, input.Title, input.DayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
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

func (uc *CreateExerciseUseCase) Execute(ctx context.Context, trainerID uuid.UUID, input CreateExerciseInput) (*ExerciseOutput, error) {
	sessionID, err := uuid.Parse(input.WorkoutSessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: workout_session_id inválido", errs.ErrInvalidInput)
	}

	session, err := uc.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("falha ao buscar sessão")
	}
	if session == nil {
		return nil, fmt.Errorf("%w: sessão não encontrada", errs.ErrNotFound)
	}

	plan, err := uc.repo.GetPlanByID(ctx, session.WorkoutPlanID)
	if err != nil {
		return nil, errors.New("falha ao buscar plano de treino")
	}
	if plan == nil {
		return nil, fmt.Errorf("%w: plano de treino não encontrado", errs.ErrNotFound)
	}
	if plan.TrainerID != trainerID {
		return nil, fmt.Errorf("%w: acesso negado ao plano de outro treinador", errs.ErrForbidden)
	}

	exercise, err := domain.NewExercise(sessionID, input.Name, input.Sets, input.Reps, input.RestSeconds, input.Position)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
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
	repo     domain.Repository
	studRepo domainS.Repository
}

func NewListPlansUseCase(r domain.Repository, sr domainS.Repository) *ListPlansUseCase {
	return &ListPlansUseCase{repo: r, studRepo: sr}
}

func (uc *ListPlansUseCase) ByStudent(ctx context.Context, tokenUserID uuid.UUID) ([]*PlanOutput, error) {
	stud, err := uc.studRepo.GetByUserID(ctx, tokenUserID)
	if err != nil {
		return nil, errors.New("falha ao resolver perfil do aluno")
	}
	if stud == nil {
		return []*PlanOutput{}, nil
	}

	plans, err := uc.repo.ListPlansByStudent(ctx, stud.ID)
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

type GetPlanDetailsUseCase struct {
	repo     domain.Repository
	studRepo domainS.Repository
	orgRepo  domainOrg.Repository
}

func NewGetPlanDetailsUseCase(r domain.Repository, sr domainS.Repository) *GetPlanDetailsUseCase {
	return &GetPlanDetailsUseCase{repo: r, studRepo: sr}
}

func (uc *GetPlanDetailsUseCase) WithOrganizationRepo(orgRepo domainOrg.Repository) *GetPlanDetailsUseCase {
	uc.orgRepo = orgRepo
	return uc
}

func (uc *GetPlanDetailsUseCase) Execute(ctx context.Context, planIDStr string, tokenUserID uuid.UUID, role string) (*PlanDetailsOutput, error) {
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: id do plano inválido", errs.ErrInvalidInput)
	}

	plan, err := uc.repo.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("%w: plano não encontrado", errs.ErrNotFound)
	}

	switch role {
	case domainU.RoleStudent:
		stud, err := uc.studRepo.GetByUserID(ctx, tokenUserID)
		if err != nil {
			return nil, errors.New("falha ao verificar perfil do aluno")
		}
		if stud == nil || plan.StudentID != stud.ID {
			return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
		}
	case domainU.RoleTrainer:
		if uc.orgRepo != nil {
			active, err := uc.orgRepo.HasActiveRelationship(ctx, tokenUserID, plan.StudentID)
			if err != nil {
				return nil, errors.New("falha ao verificar vínculo ativo com o aluno")
			}
			if !active {
				return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
			}
		} else if plan.TrainerID != tokenUserID {
			return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
		}
	default:
		return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
	}

	sessions, err := uc.repo.ListSessionsByPlan(ctx, plan.ID)
	if err != nil {
		return nil, err
	}

	sessionDetails := make([]SessionDetailsOutput, 0, len(sessions))
	for _, s := range sessions {
		exercises, err := uc.repo.ListExercisesBySession(ctx, s.ID)
		if err != nil {
			return nil, err
		}

		exOut := make([]ExerciseOutput, 0, len(exercises))
		for _, e := range exercises {
			exOut = append(exOut, ExerciseOutput{
				ID:               e.ID.String(),
				WorkoutSessionID: e.WorkoutSessionID.String(),
				Name:             e.Name,
				Sets:             e.Sets,
				Reps:             e.Reps,
				RestSeconds:      e.RestSeconds,
				Notes:            e.Notes,
				VideoURL:         e.VideoURL,
				Position:         e.Position,
			})
		}

		sessionDetails = append(sessionDetails, SessionDetailsOutput{
			Session: SessionOutput{
				ID:            s.ID.String(),
				WorkoutPlanID: s.WorkoutPlanID.String(),
				Title:         s.Title,
				DayOfWeek:     s.DayOfWeek,
				Notes:         s.Notes,
				CreatedAt:     s.CreatedAt.Format(time.RFC3339),
			},
			Exercises: exOut,
		})
	}

	return &PlanDetailsOutput{
		Plan: PlanOutput{
			ID:          plan.ID.String(),
			TrainerID:   plan.TrainerID.String(),
			StudentID:   plan.StudentID.String(),
			Title:       plan.Title,
			Description: plan.Description,
			IsActive:    plan.IsActive,
			CreatedAt:   plan.CreatedAt.Format(time.RFC3339),
		},
		Sessions: sessionDetails,
	}, nil
}
