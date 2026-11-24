package workout_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appworkout "github.com/SCAPUTO88/fit-brother/internal/application/workout"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
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
	if p := args.Get(0); p != nil {
		return p.([]*domain.WorkoutPlan), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWorkoutRepo) ListPlansByTrainer(ctx context.Context, trainerID uuid.UUID) ([]*domain.WorkoutPlan, error) {
	args := m.Called(ctx, trainerID)
	if p := args.Get(0); p != nil {
		return p.([]*domain.WorkoutPlan), args.Error(1)
	}
	return nil, args.Error(1)
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
func (m *MockWorkoutRepo) GetExerciseByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
	args := m.Called(ctx, id)
	if e := args.Get(0); e != nil {
		return e.(*domain.Exercise), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWorkoutRepo) ListExercisesBySession(ctx context.Context, sessionID uuid.UUID) ([]*domain.Exercise, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*domain.Exercise), args.Error(1)
}
func (m *MockWorkoutRepo) DeleteExercise(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockWorkoutStudentRepo struct {
	mock.Mock
}

func (m *MockWorkoutStudentRepo) Create(context.Context, *domainS.Student) error { return nil }
func (m *MockWorkoutStudentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainS.Student, error) {
	args := m.Called(ctx, id)
	if s := args.Get(0); s != nil {
		return s.(*domainS.Student), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWorkoutStudentRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainS.Student, error) {
	args := m.Called(ctx, userID)
	if s := args.Get(0); s != nil {
		return s.(*domainS.Student), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWorkoutStudentRepo) GetByCPFHash(context.Context, string) (*domainS.Student, error) {
	return nil, nil
}
func (m *MockWorkoutStudentRepo) ListByTrainer(context.Context, uuid.UUID) ([]*domainS.StudentBaseInfo, error) {
	return nil, nil
}
func (m *MockWorkoutStudentRepo) Delete(context.Context, uuid.UUID) error                  { return nil }
func (m *MockWorkoutStudentRepo) SaveBiometric(context.Context, *domainS.BiometricData) error { return nil }

func TestCreateWorkoutPlanUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("Deve criar plano apenas para aluno vinculado ao treinador", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		studRepo := new(MockWorkoutStudentRepo)
		sut := appworkout.NewCreatePlanUseCase(repo, studRepo)

		trainerID := uuid.New()
		studentID := uuid.New()

		studRepo.On("GetByID", ctx, studentID).Return(&domainS.Student{
			ID:        studentID,
			TrainerID: &trainerID,
		}, nil)

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
		studRepo.AssertExpectations(t)
	})

	t.Run("Deve negar criacao de plano para aluno de outro treinador", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		studRepo := new(MockWorkoutStudentRepo)
		sut := appworkout.NewCreatePlanUseCase(repo, studRepo)

		trainerA := uuid.New()
		trainerB := uuid.New()
		studentID := uuid.New()

		studRepo.On("GetByID", ctx, studentID).Return(&domainS.Student{
			ID:        studentID,
			TrainerID: &trainerA,
		}, nil)

		_, err := sut.Execute(ctx, appworkout.CreatePlanInput{
			TrainerID: trainerB.String(),
			StudentID: studentID.String(),
			Title:     "Plano Indevido",
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
}

func TestCreateWorkoutSessionUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("Deve criar sessao vinculada ao plano do mesmo treinador", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreateSessionUseCase(repo)

		planID := uuid.New()
		trainerID := uuid.New()

		repo.On("GetPlanByID", ctx, planID).Return(&domain.WorkoutPlan{
			ID:        planID,
			TrainerID: trainerID,
			IsActive:  true,
		}, nil)

		repo.On("CreateSession", ctx, mock.MatchedBy(func(s *domain.WorkoutSession) bool {
			return s.WorkoutPlanID == planID && s.Title == "Treino A - Peito"
		})).Return(nil)

		out, err := sut.Execute(ctx, trainerID, appworkout.CreateSessionInput{
			WorkoutPlanID: planID.String(),
			Title:         "Treino A - Peito",
			DayOfWeek:     1,
		})

		require.NoError(t, err)
		assert.Equal(t, "Treino A - Peito", out.Title)
		repo.AssertExpectations(t)
	})

	t.Run("Deve bloquear adicao de sessao em plano de outro treinador", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreateSessionUseCase(repo)

		planID := uuid.New()
		repo.On("GetPlanByID", ctx, planID).Return(&domain.WorkoutPlan{
			ID:        planID,
			TrainerID: uuid.New(),
		}, nil)

		_, err := sut.Execute(ctx, uuid.New(), appworkout.CreateSessionInput{
			WorkoutPlanID: planID.String(),
			Title:         "Treino X",
			DayOfWeek:     3,
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
}

func TestCreateExerciseUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("Deve adicionar exercicio a uma sessao de plano do mesmo treinador", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreateExerciseUseCase(repo)

		sessionID := uuid.New()
		planID := uuid.New()
		trainerID := uuid.New()

		repo.On("GetSessionByID", ctx, sessionID).Return(&domain.WorkoutSession{
			ID:            sessionID,
			WorkoutPlanID: planID,
		}, nil)

		repo.On("GetPlanByID", ctx, planID).Return(&domain.WorkoutPlan{
			ID:        planID,
			TrainerID: trainerID,
		}, nil)

		repo.On("CreateExercise", ctx, mock.MatchedBy(func(e *domain.Exercise) bool {
			return e.WorkoutSessionID == sessionID && e.Name == "Supino Reto"
		})).Return(nil)

		out, err := sut.Execute(ctx, trainerID, appworkout.CreateExerciseInput{
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

	t.Run("Deve bloquear exercicio em sessao de plano de outro treinador", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		sut := appworkout.NewCreateExerciseUseCase(repo)

		sessionID := uuid.New()
		planID := uuid.New()
		repo.On("GetSessionByID", ctx, sessionID).Return(&domain.WorkoutSession{
			ID:            sessionID,
			WorkoutPlanID: planID,
		}, nil)
		repo.On("GetPlanByID", ctx, planID).Return(&domain.WorkoutPlan{
			ID:        planID,
			TrainerID: uuid.New(),
		}, nil)

		_, err := sut.Execute(ctx, uuid.New(), appworkout.CreateExerciseInput{
			WorkoutSessionID: sessionID.String(),
			Name:             "Agachamento",
			Sets:             4,
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
}

func TestStudentWorkoutIdentityResolution(t *testing.T) {
	ctx := context.Background()

	t.Run("ListPlans.ByStudent deve resolver students.id a partir do users.id do JWT", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		studRepo := new(MockWorkoutStudentRepo)
		sut := appworkout.NewListPlansUseCase(repo, studRepo)

		userID := uuid.New()
		studentID := uuid.New()
		planID := uuid.New()

		studRepo.On("GetByUserID", ctx, userID).Return(&domainS.Student{
			ID:     studentID,
			UserID: userID,
		}, nil)

		repo.On("ListPlansByStudent", ctx, studentID).Return([]*domain.WorkoutPlan{
			{ID: planID, StudentID: studentID, Title: "Treino do Aluno", IsActive: true},
		}, nil)

		plans, err := sut.ByStudent(ctx, userID)
		require.NoError(t, err)
		require.Len(t, plans, 1)
		assert.Equal(t, planID.String(), plans[0].ID)
	})

	t.Run("GetPlanDetails deve liberar acesso para aluno dono via students.id e negar para outro aluno", func(t *testing.T) {
		repo := new(MockWorkoutRepo)
		studRepo := new(MockWorkoutStudentRepo)
		sut := appworkout.NewGetPlanDetailsUseCase(repo, studRepo)

		userID := uuid.New()
		otherUserID := uuid.New()
		studentID := uuid.New()
		otherStudentID := uuid.New()
		planID := uuid.New()

		repo.On("GetPlanByID", ctx, planID).Return(&domain.WorkoutPlan{
			ID:        planID,
			TrainerID: uuid.New(),
			StudentID: studentID,
			Title:     "Plano A",
		}, nil)
		studRepo.On("GetByUserID", ctx, userID).Return(&domainS.Student{ID: studentID, UserID: userID}, nil)
		studRepo.On("GetByUserID", ctx, otherUserID).Return(&domainS.Student{ID: otherStudentID, UserID: otherUserID}, nil)
		repo.On("ListSessionsByPlan", ctx, planID).Return([]*domain.WorkoutSession{}, nil)

		details, err := sut.Execute(ctx, planID.String(), userID, domainU.RoleStudent)
		require.NoError(t, err)
		assert.Equal(t, "Plano A", details.Plan.Title)

		_, err = sut.Execute(ctx, planID.String(), otherUserID, domainU.RoleStudent)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
}
