package workout_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

func TestNewWorkoutTemplateAndClone(t *testing.T) {
	orgID := uuid.New()
	trainerID := uuid.New()
	catalogID := uuid.New()
	mediaID := uuid.New()

	t.Run("Deve criar template válido e normalizar posições e IDs", func(t *testing.T) {
		tpl, err := workout.NewWorkoutTemplate(orgID, trainerID, "Hipertrofia AB", "Divisão superior/inferior", "Hipertrofia", workout.WorkoutStructure{
			Sessions: []workout.TemplateSession{
				{
					Title:     "Treino A - Superiores",
					DayOfWeek: 1,
					Exercises: []workout.TemplateExercise{
						{
							CatalogExerciseID: &catalogID,
							MediaID:           &mediaID,
							Name:              "Supino Reto com Barra",
							Sets:              4,
							Reps:              "8-10",
							RestSeconds:       90,
						},
					},
				},
			},
		})
		require.NoError(t, err)
		require.Len(t, tpl.DraftStructure.Sessions, 1)
		sess := tpl.DraftStructure.Sessions[0]
		assert.NotEqual(t, uuid.Nil, sess.ID)
		assert.Equal(t, 1, sess.Position)
		require.Len(t, sess.Exercises, 1)
		ex := sess.Exercises[0]
		assert.NotEqual(t, uuid.Nil, ex.ID)
		assert.Equal(t, 1, ex.Position)
		require.NotNil(t, ex.CatalogExerciseID)
		assert.Equal(t, catalogID, *ex.CatalogExerciseID)
		require.NotNil(t, ex.MediaID)
		assert.Equal(t, mediaID, *ex.MediaID)

		cloned := workout.CloneStructure(tpl.DraftStructure)
		require.Len(t, cloned.Sessions, 1)
		assert.NotEqual(t, sess.ID, cloned.Sessions[0].ID)
		assert.NotEqual(t, ex.ID, cloned.Sessions[0].Exercises[0].ID)
		assert.Equal(t, catalogID, *cloned.Sessions[0].Exercises[0].CatalogExerciseID)
		assert.Equal(t, mediaID, *cloned.Sessions[0].Exercises[0].MediaID)
		assert.Equal(t, "Supino Reto com Barra", cloned.Sessions[0].Exercises[0].Name)
	})

	t.Run("Deve rejeitar título curto ou exercício inválido", func(t *testing.T) {
		_, err := workout.NewWorkoutTemplate(orgID, trainerID, "AB", "", "", workout.WorkoutStructure{})
		assert.Error(t, err)

		_, err = workout.NewWorkoutTemplate(orgID, trainerID, "Template Válido", "", "", workout.WorkoutStructure{
			Sessions: []workout.TemplateSession{
				{Title: "A", DayOfWeek: 1},
			},
		})
		assert.Error(t, err)

		_, err = workout.NewWorkoutTemplate(orgID, trainerID, "Template Válido", "", "", workout.WorkoutStructure{
			Sessions: []workout.TemplateSession{
				{
					Title:     "Treino A",
					DayOfWeek: 1,
					Exercises: []workout.TemplateExercise{
						{Name: "Supino", Sets: -2, Reps: "10", RestSeconds: 60},
					},
				},
			},
		})
		assert.Error(t, err)
	})
}
