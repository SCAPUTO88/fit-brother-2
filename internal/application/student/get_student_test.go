package student_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

func TestGetStudentUseCase(t *testing.T) {
	ctx := context.TODO()

	// Simulador de chaves para que a decriptografia funcione no mock
	aesKey := "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=" // 32 bytes validos base64 ("12345678901234567890123456789012")
	encrptSvc, err := crypto.NewAESEncryptor(aesKey)
	require.NoError(t, err)

	t.Run("Deve retornar o Student descriptografando o CPF perfeitamente", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)

		sut := student.NewGetStudentUseCase(mockURepo, mockSRepo, encrptSvc)

		studID := uuid.New()
		userID := uuid.New()

		// Dado mockado no banco: CPF ENCRIPTADO
		encryptedVal, _ := encrptSvc.Encrypt("99988877766", []byte(studID.String()))

		mockSRepo.On("GetByID", ctx, studID).Return(&domainS.Student{
			ID:           studID,
			UserID:       userID,
			CPFEncrypted: encryptedVal,
			CreatedAt:    time.Now(),
		}, nil)

		mockURepo.On("GetByID", ctx, userID).Return(&domainU.User{
			ID:       userID,
			Email:    "busca@fit.com",
			FullName: "Buscador Silva",
		}, nil)

		// Executar UseCase
		out, err := sut.Execute(ctx, studID)

		require.NoError(t, err)
		assert.Equal(t, "busca@fit.com", out.Email)

		// MAGIA DA SEGURANÇA: O dado volta desencriptado, mas apenas nesse momento de Request
		// Não fica mantido em variavel global.
		assert.Equal(t, "99988877766", out.CPF)
	})
}
