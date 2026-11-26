package student

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

type GetStudentUseCase struct {
	userRepo  domainU.Repository
	studRepo  domainS.Repository
	orgRepo   domainOrg.Repository
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

func (c *GetStudentUseCase) WithOrganizationRepo(orgRepo domainOrg.Repository) *GetStudentUseCase {
	c.orgRepo = orgRepo
	return c
}

func (c *GetStudentUseCase) Execute(ctx context.Context, studentID, requesterID uuid.UUID, role string) (*StudentDetailOutput, error) {
	stud, err := c.studRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("falha ao buscar dados do aluno")
	}
	if stud == nil {
		return nil, fmt.Errorf("%w: aluno não encontrado", errs.ErrNotFound)
	}

	if err := authorizeStudentAccessWithOrg(ctx, c.orgRepo, stud, requesterID, role); err != nil {
		return nil, err
	}

	usr, err := c.userRepo.GetByID(ctx, stud.UserID)
	if err != nil {
		return nil, errors.New("falha ao buscar perfil do usuário")
	}
	if usr == nil {
		return nil, fmt.Errorf("%w: perfil base corrompido ou inativo", errs.ErrNotFound)
	}

	var clearCPF string
	if stud.CPFEncrypted != nil && *stud.CPFEncrypted != "" {
		clearCPF, err = c.encryptor.Decrypt(*stud.CPFEncrypted, []byte(stud.ID.String()))
		if err != nil {
			return nil, errors.New("falha grave de segurança: corrupção no CPF armazenado (AAD Invalido)")
		}
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

func authorizeStudentAccessWithOrg(ctx context.Context, orgRepo domainOrg.Repository, stud *domainS.Student, requesterID uuid.UUID, role string) error {
	switch role {
	case domainU.RoleStudent:
		if stud.UserID != requesterID {
			return fmt.Errorf("%w: você não tem permissão para acessar outros perfis", errs.ErrForbidden)
		}
		return nil
	case domainU.RoleTrainer:
		if orgRepo != nil {
			hasRel, err := orgRepo.HasActiveRelationship(ctx, requesterID, stud.ID)
			if err != nil {
				return errors.New("falha ao verificar vínculo treinador-aluno")
			}
			if hasRel {
				return nil
			}
		}
		if stud.TrainerID != nil && *stud.TrainerID == requesterID && orgRepo == nil {
			return nil
		}
		return fmt.Errorf("%w: você não tem permissão para acessar este aluno", errs.ErrForbidden)
	default:
		return fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
	}
}
