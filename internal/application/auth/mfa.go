package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/totp"
)

type MFAUseCase struct {
	userRepo    domainU.Repository
	mfaRepo     domainU.MFARepository
	privacyRepo domainPrivacy.Repository
	passSvc     *crypto.PasswordService
	encryptor   *crypto.AESEncryptor
}

func NewMFAUseCase(
	userRepo domainU.Repository,
	mfaRepo domainU.MFARepository,
	privacyRepo domainPrivacy.Repository,
	passSvc *crypto.PasswordService,
	encryptor *crypto.AESEncryptor,
) *MFAUseCase {
	return &MFAUseCase{
		userRepo:    userRepo,
		mfaRepo:     mfaRepo,
		privacyRepo: privacyRepo,
		passSvc:     passSvc,
		encryptor:   encryptor,
	}
}

func (u *MFAUseCase) Status(ctx context.Context, userID uuid.UUID) (*MFAStatusOutput, error) {
	mfa, err := u.mfaRepo.GetMFA(ctx, userID)
	if err != nil {
		return nil, errors.New("falha ao consultar status MFA")
	}
	if mfa == nil || !mfa.Enabled {
		return &MFAStatusOutput{Enabled: false}, nil
	}
	var verifiedAt *string
	if mfa.VerifiedAt != nil {
		v := mfa.VerifiedAt.UTC().Format(time.RFC3339)
		verifiedAt = &v
	}
	return &MFAStatusOutput{
		Enabled:    true,
		VerifiedAt: verifiedAt,
	}, nil
}

func (u *MFAUseCase) Setup(ctx context.Context, userID uuid.UUID, role string) (*MFASetupOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: MFA TOTP disponível apenas para treinadores", errs.ErrForbidden)
	}
	usr, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || usr == nil || !usr.IsActive {
		return nil, fmt.Errorf("%w: usuário inválido", errs.ErrUnauthorized)
	}

	secret, err := totp.GenerateSecret()
	if err != nil {
		return nil, errors.New("falha ao gerar segredo MFA")
	}
	encSecret, err := u.encryptor.Encrypt(secret, []byte(userID.String()))
	if err != nil {
		return nil, errors.New("falha ao proteger segredo MFA")
	}

	now := time.Now().UTC()
	if err := u.mfaRepo.UpsertMFA(ctx, &domainU.UserMFA{
		UserID:          userID,
		SecretEncrypted: encSecret,
		Enabled:         false,
		LastUsedStep:    0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		return nil, errors.New("falha ao salvar configuração MFA")
	}

	return &MFASetupOutput{
		Secret:     secret,
		OTPAuthURI: totp.BuildOTPAuthURI("FitBrother", usr.Email, secret),
	}, nil
}

func (u *MFAUseCase) Enable(ctx context.Context, userID uuid.UUID, role, code string) error {
	if role != domainU.RoleTrainer {
		return fmt.Errorf("%w: MFA TOTP disponível apenas para treinadores", errs.ErrForbidden)
	}
	mfa, err := u.mfaRepo.GetMFA(ctx, userID)
	if err != nil {
		return errors.New("falha ao consultar configuração MFA")
	}
	if mfa == nil {
		return fmt.Errorf("%w: inicie a configuração MFA antes de ativar", errs.ErrInvalidInput)
	}
	secret, err := u.encryptor.Decrypt(mfa.SecretEncrypted, []byte(userID.String()))
	if err != nil {
		return errors.New("falha ao validar segredo MFA")
	}

	now := time.Now().UTC()
	step, ok := totp.VerifyCode(secret, code, now)
	if !ok {
		return fmt.Errorf("%w: código TOTP inválido", errs.ErrInvalidInput)
	}

	mfa.Enabled = true
	mfa.LastUsedStep = step
	mfa.VerifiedAt = &now
	mfa.UpdatedAt = now
	if err := u.mfaRepo.UpsertMFA(ctx, mfa); err != nil {
		return errors.New("falha ao ativar MFA")
	}

	if u.privacyRepo != nil {
		actor := userID
		_ = u.privacyRepo.RecordAuditEvent(ctx, &domainPrivacy.AuditEvent{
			ID:           uuid.New(),
			ActorUserID:  &actor,
			Action:       "MFA_ENABLED",
			TargetType:   "user",
			TargetID:     &actor,
			MetadataJSON: "{}",
			CreatedAt:    now,
		})
	}
	return nil
}

func (u *MFAUseCase) Disable(ctx context.Context, userID uuid.UUID, role, password, code string) error {
	if role != domainU.RoleTrainer {
		return fmt.Errorf("%w: MFA TOTP disponível apenas para treinadores", errs.ErrForbidden)
	}
	usr, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || usr == nil || !usr.IsActive {
		return fmt.Errorf("%w: usuário inválido", errs.ErrUnauthorized)
	}
	if !u.passSvc.Check(password, usr.PasswordHash) {
		return fmt.Errorf("%w: senha incorreta", errs.ErrUnauthorized)
	}

	mfa, err := u.mfaRepo.GetMFA(ctx, userID)
	if err != nil {
		return errors.New("falha ao consultar configuração MFA")
	}
	if mfa == nil || !mfa.Enabled {
		return fmt.Errorf("%w: MFA não está ativo", errs.ErrInvalidInput)
	}

	secret, err := u.encryptor.Decrypt(mfa.SecretEncrypted, []byte(userID.String()))
	if err != nil {
		return errors.New("falha ao validar segredo MFA")
	}
	step, ok := totp.VerifyCode(secret, code, time.Now().UTC())
	if !ok || step <= mfa.LastUsedStep {
		return fmt.Errorf("%w: código TOTP inválido", errs.ErrUnauthorized)
	}

	if err := u.mfaRepo.DisableMFA(ctx, userID); err != nil {
		return errors.New("falha ao desativar MFA")
	}

	if u.privacyRepo != nil {
		actor := userID
		_ = u.privacyRepo.RecordAuditEvent(ctx, &domainPrivacy.AuditEvent{
			ID:           uuid.New(),
			ActorUserID:  &actor,
			Action:       "MFA_DISABLED",
			TargetType:   "user",
			TargetID:     &actor,
			MetadataJSON: "{}",
			CreatedAt:    time.Now().UTC(),
		})
	}
	return nil
}
