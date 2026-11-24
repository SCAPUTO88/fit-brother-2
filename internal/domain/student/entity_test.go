package student_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/domain/student"
)

func TestStudentEntity(t *testing.T) {
	t.Run("Deve criar um Student válido a partir de CPF puro (sem máscara)", func(t *testing.T) {
		userID := uuid.New()
		trainerID := uuid.New()
		rawCPF := "12345678901"

		std, err := student.NewStudent(userID, trainerID, rawCPF)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, std.ID)
		assert.Equal(t, userID, std.UserID)
		require.NotNil(t, std.TrainerID)
		assert.Equal(t, trainerID, *std.TrainerID)
		assert.Equal(t, "12345678901", std.GetCleanCPF())
		assert.False(t, std.CreatedAt.IsZero())
	})

	t.Run("Deve limpar qualquer caractere não-numérico do CPF", func(t *testing.T) {
		userID := uuid.New()
		trainerID := uuid.New()
		dirtyCPF := " 123.456.789-01 !!"

		std, err := student.NewStudent(userID, trainerID, dirtyCPF)
		require.NoError(t, err)

		assert.Equal(t, "12345678901", std.GetCleanCPF())
	})

	t.Run("Deve falhar ao tentar criar aluno com CPF inválido/curto", func(t *testing.T) {
		userID := uuid.New()
		trainerID := uuid.New()

		_, err := student.NewStudent(userID, trainerID, "12345")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "o CPF deve conter exatamente 11 digitos")
	})

	t.Run("Deve falhar ao passar um UserID vazio", func(t *testing.T) {
		_, err := student.NewStudent(uuid.Nil, uuid.New(), "12345678901")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "o UserID é obrigatório")
	})

	t.Run("Deve falhar ao passar um TrainerID vazio", func(t *testing.T) {
		_, err := student.NewStudent(uuid.New(), uuid.Nil, "12345678901")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "o TrainerID é obrigatório")
	})
}

func TestBiometricEntity(t *testing.T) {
	t.Run("Deve calcular o IMC (BMI) perfeitamente dadas as medidas em kg e cm", func(t *testing.T) {
		studentID := uuid.New()
		trainerID := uuid.New()

		bio, err := student.NewBiometricRecord(studentID, trainerID, 80.0, 180.0, 15.5)

		require.NoError(t, err)
		assert.InDelta(t, 24.69, bio.BMI, 0.01)
	})

	t.Run("Deve estourar erro se os valores fisicos forem impossiveis", func(t *testing.T) {
		_, err := student.NewBiometricRecord(uuid.New(), uuid.New(), 0, 180.0, 10.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "peso e altura precisam ser maiores que zero")
	})
}
