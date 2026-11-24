package student_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

func TestGetStudentUseCase(t *testing.T) {
	ctx := context.Background()

	aesKey := "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI="
	encrptSvc, err := crypto.NewAESEncryptor(aesKey)
	require.NoError(t, err)

	t.Run("Deve retornar o Student descriptografando o CPF para o treinador responsavel", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)

		sut := student.NewGetStudentUseCase(mockURepo, mockSRepo, encrptSvc)

		studID := uuid.New()
		userID := uuid.New()
		trainerID := uuid.New()

		encryptedVal, _ := encrptSvc.Encrypt("99988877766", []byte(studID.String()))

		mockSRepo.On("GetByID", ctx, studID).Return(&domainS.Student{
			ID:           studID,
			UserID:       userID,
			TrainerID:    &trainerID,
			CPFEncrypted: encryptedVal,
			CreatedAt:    time.Now(),
		}, nil)

		mockURepo.On("GetByID", ctx, userID).Return(&domainU.User{
			ID:       userID,
			Email:    "busca@fit.com",
			FullName: "Buscador Silva",
		}, nil)

		out, err := sut.Execute(ctx, studID, trainerID, domainU.RoleTrainer)

		require.NoError(t, err)
		assert.Equal(t, "busca@fit.com", out.Email)
		assert.Equal(t, "99988877766", out.CPF)
	})

	t.Run("Deve negar acesso quando outro treinador tentar consultar o aluno", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		sut := student.NewGetStudentUseCase(mockURepo, mockSRepo, encrptSvc)

		studID := uuid.New()
		trainerA := uuid.New()
		trainerB := uuid.New()

		mockSRepo.On("GetByID", ctx, studID).Return(&domainS.Student{
			ID:        studID,
			UserID:    uuid.New(),
			TrainerID: &trainerA,
		}, nil)

		out, err := sut.Execute(ctx, studID, trainerB, domainU.RoleTrainer)
		require.Nil(t, out)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})

	t.Run("Deve negar acesso quando outro aluno tentar consultar perfil de terceiro", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		sut := student.NewGetStudentUseCase(mockURepo, mockSRepo, encrptSvc)

		studID := uuid.New()
		trainerID := uuid.New()

		mockSRepo.On("GetByID", ctx, studID).Return(&domainS.Student{
			ID:        studID,
			UserID:    uuid.New(),
			TrainerID: &trainerID,
		}, nil)

		out, err := sut.Execute(ctx, studID, uuid.New(), domainU.RoleStudent)
		require.Nil(t, out)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
}
