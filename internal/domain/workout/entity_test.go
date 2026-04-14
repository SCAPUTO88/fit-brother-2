package workout_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

func TestNewWorkoutPlan(t *testing.T) {
	trainerID := uuid.New()
	studentID := uuid.New()

	t.Run("Deve criar plano de treino válido", func(t *testing.T) {
		plan, err := workout.NewWorkoutPlan(trainerID, studentID, "Hipertrofia 12 semanas", "Foco em volume")
		require.NoError(t, err)
		assert.Equal(t, trainerID, plan.TrainerID)
		assert.Equal(t, studentID, plan.StudentID)
		assert.Equal(t, "Hipertrofia 12 semanas", plan.Title)
		assert.True(t, plan.IsActive)
		assert.NotEqual(t, uuid.Nil, plan.ID)
	})

	t.Run("Deve falhar com título curto demais", func(t *testing.T) {
		_, err := workout.NewWorkoutPlan(trainerID, studentID, "AB", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "título")
	})

	t.Run("Deve falhar sem TrainerID", func(t *testing.T) {
		_, err := workout.NewWorkoutPlan(uuid.Nil, studentID, "Plano Válido", "")
		assert.Error(t, err)
	})

	t.Run("Deve falhar sem StudentID", func(t *testing.T) {
		_, err := workout.NewWorkoutPlan(trainerID, uuid.Nil, "Plano Válido", "")
		assert.Error(t, err)
	})
}

func TestNewWorkoutSession(t *testing.T) {
	planID := uuid.New()

	t.Run("Deve criar sessão válida", func(t *testing.T) {
		session, err := workout.NewWorkoutSession(planID, "Treino A - Peito e Tríceps", 1)
		require.NoError(t, err)
		assert.Equal(t, "Treino A - Peito e Tríceps", session.Title)
		assert.Equal(t, 1, session.DayOfWeek)
		assert.Equal(t, planID, session.WorkoutPlanID)
	})

	t.Run("Deve falhar com dia inválido", func(t *testing.T) {
		_, err := workout.NewWorkoutSession(planID, "Treino X", 7)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dia da semana")
	})
}

func TestNewExercise(t *testing.T) {
	sessionID := uuid.New()

	t.Run("Deve criar exercício válido", func(t *testing.T) {
		ex, err := workout.NewExercise(sessionID, "Supino Reto", 4, "8-12", 90, 0)
		require.NoError(t, err)
		assert.Equal(t, "Supino Reto", ex.Name)
		assert.Equal(t, 4, ex.Sets)
		assert.Equal(t, "8-12", ex.Reps)
	})

	t.Run("Deve falhar com nome curto", func(t *testing.T) {
		_, err := workout.NewExercise(sessionID, "X", 3, "10", 60, 0)
		assert.Error(t, err)
	})

	t.Run("Deve falhar com séries negativas", func(t *testing.T) {
		_, err := workout.NewExercise(sessionID, "Agachamento", -1, "10", 60, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "séries")
	})
}
