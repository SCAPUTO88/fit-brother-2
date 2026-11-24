package student_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

func TestAnonymizeStudentUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("Deve limpar PII do User, revogar tokens e excluir ficha LGPD do aluno em transacao", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		txMgr := &MockTxManager{}

		sut := student.NewAnonymizeStudentUseCase(mockURepo, mockSRepo, txMgr)

		studID := uuid.New()
		userID := uuid.New()
		trainerID := uuid.New()

		mockSRepo.On("GetByID", ctx, studID).Return(&domainS.Student{
			ID:        studID,
			UserID:    userID,
			TrainerID: &trainerID,
		}, nil)

		mockURepo.On("GetByID", ctx, userID).Return(&domainU.User{
			ID:       userID,
			Email:    "aluno@fit.com",
			FullName: "Aluno Real",
			IsActive: true,
		}, nil)

		mockURepo.On("Update", ctx, mock.MatchedBy(func(u *domainU.User) bool {
			return u.ID == userID &&
				!u.IsActive &&
				u.DeletedAt != nil &&
				u.Email == "deleted-"+userID.String()+"@anon.local" &&
				u.FullName == "Usuário Anonimizado" &&
				u.PasswordHash == "ANONIMIZADO"
		})).Return(nil)

		mockURepo.On("RevokeAllByUserID", ctx, userID).Return(nil)
		mockSRepo.On("Delete", ctx, studID).Return(nil)

		err := sut.Execute(ctx, studID, trainerID, domainU.RoleTrainer)

		require.NoError(t, err)
		assert.Equal(t, 1, txMgr.calls)
		mockSRepo.AssertExpectations(t)
		mockURepo.AssertExpectations(t)
	})

	t.Run("Deve bloquear anonimizacao cruzada por outro treinador", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		txMgr := &MockTxManager{}

		studID := uuid.New()
		trainerA := uuid.New()
		trainerB := uuid.New()

		mockSRepo.On("GetByID", ctx, studID).Return(&domainS.Student{
			ID:        studID,
			UserID:    uuid.New(),
			TrainerID: &trainerA,
		}, nil)

		sut := student.NewAnonymizeStudentUseCase(mockURepo, mockSRepo, txMgr)

		err := sut.Execute(ctx, studID, trainerB, domainU.RoleTrainer)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
		mockURepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		mockSRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	})

	t.Run("Falha com ErrNotFound se nao encontrar aluno", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		txMgr := &MockTxManager{}
		mockSRepo.On("GetByID", ctx, mock.Anything).Return((*domainS.Student)(nil), nil)

		sut := student.NewAnonymizeStudentUseCase(mockURepo, mockSRepo, txMgr)

		err := sut.Execute(ctx, uuid.New(), uuid.New(), domainU.RoleTrainer)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})
}
