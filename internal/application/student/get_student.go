package student

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
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

func (c *GetStudentUseCase) Execute(ctx context.Context, studentID, requesterID uuid.UUID, role string) (*StudentDetailOutput, error) {
	stud, err := c.studRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("falha ao buscar dados do aluno")
	}
	if stud == nil {
		return nil, fmt.Errorf("%w: aluno não encontrado", errs.ErrNotFound)
	}

	if err := authorizeStudentAccess(stud, requesterID, role); err != nil {
		return nil, err
	}

	usr, err := c.userRepo.GetByID(ctx, stud.UserID)
	if err != nil {
		return nil, errors.New("falha ao buscar perfil do usuário")
	}
	if usr == nil {
		return nil, fmt.Errorf("%w: perfil base corrompido ou inativo", errs.ErrNotFound)
	}

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

func authorizeStudentAccess(stud *domainS.Student, requesterID uuid.UUID, role string) error {
	switch role {
	case domainU.RoleStudent:
		if stud.UserID != requesterID {
			return fmt.Errorf("%w: você não tem permissão para acessar outros perfis", errs.ErrForbidden)
		}
		return nil
	case domainU.RoleTrainer:
		if stud.TrainerID == nil || *stud.TrainerID != requesterID {
			return fmt.Errorf("%w: você não tem permissão para acessar este aluno", errs.ErrForbidden)
		}
		return nil
	default:
		return fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
	}
}
