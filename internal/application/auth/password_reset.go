package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainMailer "github.com/SCAPUTO88/fit-brother/internal/domain/mailer"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

const passwordResetTTL = 30 * time.Minute

type PasswordResetUseCase struct {
	userRepo    domainU.Repository
	resetRepo   domainU.PasswordResetRepository
	privacyRepo domainPrivacy.Repository
	txManager   tx.Manager
	passSvc     *crypto.PasswordService
	mailer      domainMailer.Port
}

func NewPasswordResetUseCase(
	userRepo domainU.Repository,
	resetRepo domainU.PasswordResetRepository,
	privacyRepo domainPrivacy.Repository,
	txManager tx.Manager,
	passSvc *crypto.PasswordService,
	mailer domainMailer.Port,
) *PasswordResetUseCase {
	return &PasswordResetUseCase{
		userRepo:    userRepo,
		resetRepo:   resetRepo,
		privacyRepo: privacyRepo,
		txManager:   txManager,
		passSvc:     passSvc,
		mailer:      mailer,
	}
}

// Request inicia a redefinição sem vazar se o e-mail existe (anti-enumeração).
func (u *PasswordResetUseCase) Request(ctx context.Context, input PasswordResetRequestInput) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return fmt.Errorf("%w: e-mail obrigatório", errs.ErrInvalidInput)
	}

	usr, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("falha ao processar solicitação de redefinição")
	}
	if usr == nil || !usr.IsActive {
		return nil
	}

	rawToken, err := generateOpaqueToken()
	if err != nil {
		return errors.New("falha ao gerar token de redefinição")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(passwordResetTTL)
	record := &domainU.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    usr.ID,
		TokenHash: crypto.SHA256Raw(rawToken),
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	run := func(execCtx context.Context) error {
		if err := u.resetRepo.InvalidatePasswordResetTokensByUser(execCtx, usr.ID); err != nil {
			return err
		}
		if err := u.resetRepo.CreatePasswordResetToken(execCtx, record); err != nil {
			return err
		}
		if u.privacyRepo != nil {
			actor := usr.ID
			var ipPtr *string
			if strings.TrimSpace(input.ClientIP) != "" {
				ip := input.ClientIP
				ipPtr = &ip
			}
			_ = u.privacyRepo.RecordAuditEvent(execCtx, &domainPrivacy.AuditEvent{
				ID:           uuid.New(),
				ActorUserID:  &actor,
				Action:       "PASSWORD_RESET_REQUESTED",
				TargetType:   "user",
				TargetID:     &actor,
				MetadataJSON: "{}",
				IPAddress:    ipPtr,
				CreatedAt:    now,
			})
		}
		return nil
	}

	if u.txManager != nil {
		if err := u.txManager.RunInTx(ctx, run); err != nil {
			return errors.New("falha ao registrar solicitação de redefinição")
		}
	} else if err := run(ctx); err != nil {
		return errors.New("falha ao registrar solicitação de redefinição")
	}

	if u.mailer != nil {
		if err := u.mailer.SendPasswordReset(ctx, domainMailer.PasswordResetEmail{
			ToEmail:   usr.Email,
			FullName:  usr.FullName,
			RawToken:  rawToken,
			ExpiresAt: expiresAt,
		}); err != nil {
			return errors.New("falha ao enviar e-mail de redefinição")
		}
	}

	return nil
}

// Confirm valida token único/não expirado, atualiza a senha e revoga todas as sessões ativas.
func (u *PasswordResetUseCase) Confirm(ctx context.Context, input PasswordResetConfirmInput) error {
	rawToken := strings.TrimSpace(input.Token)
	if rawToken == "" || len(input.NewPassword) < 8 {
		return fmt.Errorf("%w: token válido e senha mínima de 8 caracteres são obrigatórios", errs.ErrInvalidInput)
	}

	tokenHash := crypto.SHA256Raw(rawToken)
	hashedPass, err := u.passSvc.Hash(input.NewPassword)
	if err != nil {
		return errors.New("falha ao processar nova senha")
	}

	now := time.Now().UTC()
	run := func(execCtx context.Context) error {
		record, err := u.resetRepo.GetPasswordResetTokenByHash(execCtx, tokenHash)
		if err != nil {
			return errors.New("falha ao consultar token")
		}
		if record == nil || !record.IsValid(now) {
			return fmt.Errorf("%w: token de redefinição inválido ou expirado", errs.ErrUnauthorized)
		}

		usr, err := u.userRepo.GetByID(execCtx, record.UserID)
		if err != nil {
			return errors.New("falha ao buscar usuário")
		}
		if usr == nil || !usr.IsActive {
			return fmt.Errorf("%w: usuário inativo ou inexistente", errs.ErrUnauthorized)
		}

		usr.PasswordHash = hashedPass
		usr.UpdatedAt = now
		if err := u.userRepo.Update(execCtx, usr); err != nil {
			return errors.New("falha ao atualizar senha")
		}
		if err := u.resetRepo.MarkPasswordResetTokenUsed(execCtx, record.ID); err != nil {
			return errors.New("falha ao invalidar token de redefinição")
		}
		if err := u.userRepo.RevokeAllByUserID(execCtx, usr.ID); err != nil {
			return errors.New("falha ao revogar sessões anteriores")
		}

		if u.privacyRepo != nil {
			actor := usr.ID
			var ipPtr *string
			if strings.TrimSpace(input.ClientIP) != "" {
				ip := input.ClientIP
				ipPtr = &ip
			}
			_ = u.privacyRepo.RecordAuditEvent(execCtx, &domainPrivacy.AuditEvent{
				ID:           uuid.New(),
				ActorUserID:  &actor,
				Action:       "PASSWORD_RESET_COMPLETED",
				TargetType:   "user",
				TargetID:     &actor,
				MetadataJSON: "{}",
				IPAddress:    ipPtr,
				CreatedAt:    now,
			})
		}

		return nil
	}

	if u.txManager != nil {
		return u.txManager.RunInTx(ctx, run)
	}
	return run(ctx)
}
