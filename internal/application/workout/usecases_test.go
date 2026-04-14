package workout_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appworkout "github.com/SCAPUTO88/fit-brother/internal/application/workout"
	domain "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type MockWorkoutRepo struct {
	mock.Mock
}

func (m *MockWorkoutRepo) CreatePlan(ctx context.Context, plan *domain.WorkoutPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}
func (m *MockWorkoutRepo) GetPlanByID(ctx context.Context, id uuid.UUID) (*domain.WorkoutPlan, error) {
	args := m.Called(ctx, id)
	if p := args.Get(0); p != nil {
		return p.(*domain.WorkoutPlan), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWorkoutRepo) ListPlansByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.WorkoutPlan, error) {
	args := m.Called(ctx, studentID)
	return args.Get(0).([]*domain.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutRepo) ListPlansByTrainer(ctx context.Context, trainerID uuid.UUID) ([]*domain.WorkoutPlan, error) {
	args := m.Called(ctx, trainerID)
	return args.Get(0).([]*domain.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutRepo) UpdatePlan(ctx context.Context, plan *domain.WorkoutPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}
func (m *MockWorkoutRepo) CreateSession(ctx context.Context, session *domain.WorkoutSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}
func (m *MockWorkoutRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.WorkoutSession, error) {
	args := m.Called(ctx, id)
	if s := args.Get(0); s != nil {
		return s.(*domain.WorkoutSession), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWorkoutRepo) ListSessionsByPlan(ctx context.Context, planID uuid.UUID) ([]*domain.WorkoutSession, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).([]*domain.WorkoutSession), args.Error(1)
}
func (m *MockWorkoutRepo) CreateExercise(ctx context.Context, exercise *domain.Exercise) error {
	args := m.Called(ctx, exercise)
	return args.Error(0)
}
func (m *MockWorkoutRepo) ListExercisesBySession(ctx context.Context, sessionID uuid.UUID) ([]*domain.Exercise, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*domain.Exercise), args.Error(1)
}
func (m *MockWorkoutRepo) DeleteExercise(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateWorkoutPlanUseCase(t *testing.T) {
	ctx := context.TODO()

	t.Run("Deve criar plano e persistir no repositório", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreatePlanUseCase(repo)

		trainerID := uuid.New()
		studentID := uuid.New()

		repo.On("CreatePlan", ctx, mock.MatchedBy(func(p *domain.WorkoutPlan) bool {
			return p.TrainerID == trainerID &&
				p.StudentID == studentID &&
				p.Title == "Hipertrofia 12 semanas" &&
				p.IsActive
		})).Return(nil)

		out, err := sut.Execute(ctx, appworkout.CreatePlanInput{
			TrainerID:   trainerID.String(),
			StudentID:   studentID.String(),
			Title:       "Hipertrofia 12 semanas",
			Description: "Foco em volume",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, out.ID)
		assert.Equal(t, "Hipertrofia 12 semanas", out.Title)
		assert.True(t, out.IsActive)
		repo.AssertExpectations(t)
	})

	t.Run("Deve falhar com input inválido (título curto)", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreatePlanUseCase(repo)

		_, err := sut.Execute(ctx, appworkout.CreatePlanInput{
			TrainerID: uuid.New().String(),
			StudentID: uuid.New().String(),
			Title:     "AB",
		})

		assert.Error(t, err)
	})
}

func TestCreateWorkoutSessionUseCase(t *testing.T) {
	ctx := context.TODO()

	t.Run("Deve criar sessão vinculada ao plano existente", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreateSessionUseCase(repo)

		planID := uuid.New()

		repo.On("GetPlanByID", ctx, planID).Return(&domain.WorkoutPlan{
			ID:       planID,
			IsActive: true,
		}, nil)

		repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.WorkoutSession) bool {
			return s.WorkoutPlanID == planID && s.Title == "Treino A - Peito"
		})).Return(nil)

		out, err := sut.Execute(ctx, appworkout.CreateSessionInput{
			WorkoutPlanID: planID.String(),
			Title:         "Treino A - Peito",
			DayOfWeek:     1,
		})

		require.NoError(t, err)
		assert.Equal(t, "Treino A - Peito", out.Title)
		repo.AssertExpectations(t)
	})

	t.Run("Deve falhar se plano não encontrado", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreateSessionUseCase(repo)

		planID := uuid.New()
		repo.On("GetPlanByID", ctx, planID).Return((*domain.WorkoutPlan)(nil), nil)

		_, err := sut.Execute(ctx, appworkout.CreateSessionInput{
			WorkoutPlanID: planID.String(),
			Title:         "Treino X",
			DayOfWeek:     3,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plano")
	})
}

func TestCreateExerciseUseCase(t *testing.T) {
	ctx := context.TODO()

	t.Run("Deve adicionar exercício a uma sessão existente", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreateExerciseUseCase(repo)

		sessionID := uuid.New()

		repo.On("GetSessionByID", ctx, sessionID).Return(&domain.WorkoutSession{
			ID: sessionID,
		}, nil)

		repo.On("CreateExercise", ctx, mock.MatchedBy(func(e *domain.Exercise) bool {
			return e.WorkoutSessionID == sessionID && e.Name == "Supino Reto"
		})).Return(nil)

		out, err := sut.Execute(ctx, appworkout.CreateExerciseInput{
			WorkoutSessionID: sessionID.String(),
			Name:             "Supino Reto",
			Sets:             4,
			Reps:             "8-12",
			RestSeconds:      90,
		})

		require.NoError(t, err)
		assert.Equal(t, "Supino Reto", out.Name)
		assert.Equal(t, 4, out.Sets)
		repo.AssertExpectations(t)
	})
}
