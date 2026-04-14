package student_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

func TestAnonymizeStudentUseCase(t *testing.T) {
	ctx := context.TODO()

	t.Run("Deve acionar SoftDelete no User e HardDelete na ficha LGPD do Estudante", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)

		sut := student.NewAnonymizeStudentUseCase(mockURepo, mockSRepo)

		studID := uuid.New()
		userID := uuid.New()

		// Busca base de aluno para sabermos a qual User ele pertecia
		mockSRepo.On("GetByID", ctx, studID).Return(&domainS.Student{
			ID:     studID,
			UserID: userID,
		}, nil)

		// Soft delete do perfil core (Desativar)
		// Simula buscar o usuario root para o soft-delete
		mockURepo.On("GetByID", ctx, userID).Return(&domainU.User{
			ID:       userID,
			IsActive: true,
		}, nil)

		// Expectativa 1: O Usuario Root deve ser atualizado para Inativo e ter as confidenciais purgadas (IsActive false)
		mockURepo.On("Update", ctx, mock.MatchedBy(func(u *domainU.User) bool {
			return u.ID == userID && !u.IsActive
		})).Return(nil)

		// Expectativa 2: Exclusão Permanente (HardDelete) dos laudos medicos e CPF vinculados ao Student do BD
		mockSRepo.On("Delete", ctx, studID).Return(nil)

		// Executar o esqquecimento LGPD
		err := sut.Execute(ctx, studID)

		require.NoError(t, err)
		mockSRepo.AssertExpectations(t)
		mockURepo.AssertExpectations(t)
	})

	t.Run("Falha se não encontrar aluno", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		mockSRepo.On("GetByID", ctx, mock.Anything).Return((*domainS.Student)(nil), nil)

		sut := student.NewAnonymizeStudentUseCase(mockURepo, mockSRepo)

		err := sut.Execute(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "aluno não encontrado")
	})
}
