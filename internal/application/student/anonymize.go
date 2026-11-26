package student

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type AnonymizeStudentUseCase struct {
	userRepo    domainU.Repository
	studRepo    domainS.Repository
	orgRepo     domainOrg.Repository
	privacyRepo domainPrivacy.Repository
	txManager   tx.Manager
}

func NewAnonymizeStudentUseCase(ur domainU.Repository, sr domainS.Repository, txManager tx.Manager) *AnonymizeStudentUseCase {
	return &AnonymizeStudentUseCase{
		userRepo:  ur,
		studRepo:  sr,
		txManager: txManager,
	}
}

func (c *AnonymizeStudentUseCase) WithMultiTenantAndPrivacy(
	orgRepo domainOrg.Repository,
	privacyRepo domainPrivacy.Repository,
) *AnonymizeStudentUseCase {
	c.orgRepo = orgRepo
	c.privacyRepo = privacyRepo
	return c
}

func (c *AnonymizeStudentUseCase) Execute(ctx context.Context, studentID, requesterID uuid.UUID, role string) error {
	run := func(execCtx context.Context) error {
		stud, err := c.studRepo.GetByID(execCtx, studentID)
		if err != nil {
			return errors.New("falha ao consultar aluno")
		}
		if stud == nil {
			return fmt.Errorf("%w: aluno não encontrado", errs.ErrNotFound)
		}

		if err := authorizeStudentAccessWithOrg(execCtx, c.orgRepo, stud, requesterID, role); err != nil {
			return err
		}

		usr, err := c.userRepo.GetByID(execCtx, stud.UserID)
		if err != nil {
			return errors.New("falha ao consultar usuário base")
		}
		if usr == nil {
			return fmt.Errorf("%w: perfil base corrompido ou inativo", errs.ErrNotFound)
		}

		if c.orgRepo != nil {
			if err := c.orgRepo.EndAllByStudent(execCtx, stud.ID); err != nil {
				return err
			}
		}

		now := time.Now().UTC()
		usr.Email = "deleted-" + usr.ID.String() + "@anon.local"
		usr.FullName = "Usuário Anonimizado"
		usr.PasswordHash = "ANONIMIZADO"
		usr.IsActive = false
		usr.DeletedAt = &now

		if err := c.userRepo.Update(execCtx, usr); err != nil {
			return err
		}

		if err := c.userRepo.RevokeAllByUserID(execCtx, usr.ID); err != nil {
			return err
		}

		if c.privacyRepo != nil {
			if err := c.privacyRepo.CompletePrivacyRequestsByUserAndType(execCtx, usr.ID, domainPrivacy.RequestAnonymize); err != nil {
				return err
			}
			actorID := requesterID
			targetID := usr.ID
			_ = c.privacyRepo.RecordAuditEvent(execCtx, &domainPrivacy.AuditEvent{
				ID:           uuid.New(),
				ActorUserID:  &actorID,
				Action:       "STUDENT_ANONYMIZED",
				TargetType:   "user",
				TargetID:     &targetID,
				MetadataJSON: "{}",
				CreatedAt:    now,
			})
		}

		return c.studRepo.Delete(execCtx, stud.ID)
	}

	if c.txManager != nil {
		return c.txManager.RunInTx(ctx, run)
	}
	return run(ctx)
}
