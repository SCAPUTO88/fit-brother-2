package student

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type AnonymizeStudentUseCase struct {
	userRepo domainU.Repository
	studRepo domainS.Repository
}

func NewAnonymizeStudentUseCase(ur domainU.Repository, sr domainS.Repository) *AnonymizeStudentUseCase {
	return &AnonymizeStudentUseCase{
		userRepo: ur,
		studRepo: sr,
	}
}

func (c *AnonymizeStudentUseCase) Execute(ctx context.Context, studentID uuid.UUID) error {
	stud, err := c.studRepo.GetByID(ctx, studentID)
	if err != nil || stud == nil {
		return errors.New("aluno não encontrado")
	}

	usr, err := c.userRepo.GetByID(ctx, stud.UserID)
	if err != nil || usr == nil {
		return errors.New("perfil base corrompido ou inativo")
	}

	usr.IsActive = false

	err = c.userRepo.Update(ctx, usr)
	if err != nil {
		return err
	}

	return c.studRepo.Delete(ctx, stud.ID)
}
