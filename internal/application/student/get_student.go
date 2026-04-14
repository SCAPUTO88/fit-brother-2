package student

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

type GetStudentUseCase struct {
	userRepo  domainU.Repository
	studRepo  domainS.Repository
	encryptor *crypto.AESEncryptor
}

func NewGetStudentUseCase(
	ur domainU.Repository,
	sr domainS.Repository,
	enc *crypto.AESEncryptor,
) *GetStudentUseCase {
	return &GetStudentUseCase{
		userRepo:  ur,
		studRepo:  sr,
		encryptor: enc,
	}
}

// Execute recupera um estudante mesclando dados da entidade de Autenticação e 
// descriptografando o CPF em Tempo Real GCM, consumindo o mínimo de memória temporária.
func (c *GetStudentUseCase) Execute(ctx context.Context, studentID uuid.UUID) (*StudentDetailOutput, error) {

	// 1. Puxamos nossos dados Biometricos / CPF Cifrado
	stud, err := c.studRepo.GetByID(ctx, studentID)
	if err != nil || stud == nil {
		return nil, errors.New("aluno não encontrado")
	}

	// 2. Puxamos nosso Perfil Base
	usr, err := c.userRepo.GetByID(ctx, stud.UserID)
	if err != nil || usr == nil {
		return nil, errors.New("perfil base corrompido ou inativo")
	}

	// 3. Descriptografia GCM 
	// Usamos o Validation-Binding AAD (O proprio ID é a prova de violação)!
	clearCPF, err := c.encryptor.Decrypt(stud.CPFEncrypted, []byte(stud.ID.String()))
	if err != nil {
		return nil, errors.New("falha grave de segurança: corrupção no CPF armazenado (AAD Invalido)")
	}

	return &StudentDetailOutput{
		ID:        stud.ID.String(),
		UserID:    stud.UserID.String(),
		Email:     usr.Email,
		FullName:  usr.FullName,
		CPF:       clearCPF,
		CreatedAt: stud.CreatedAt.Format(time.RFC3339),
	}, nil
}
