package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/totp"
)

var (
	ErrMFARequired = errors.New("código MFA obrigatório")
	ErrInvalidMFA  = errors.New("código MFA inválido")
)

type LoginUseCase struct {
	userRepo        user.Repository
	mfaRepo         user.MFARepository
	privacyRepo     domainPrivacy.Repository
	passwordService *crypto.PasswordService
	encryptor       *crypto.AESEncryptor
	jwtService      *jwt.Service
	accessTTL       time.Duration
	refreshTTL      time.Duration
}

func NewLoginUseCase(
	repo user.Repository,
	ps *crypto.PasswordService,
	js *jwt.Service,
	accessTTL, refreshTTL time.Duration,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:        repo,
		passwordService: ps,
		jwtService:      js,
		accessTTL:       accessTTL,
		refreshTTL:      refreshTTL,
	}
}

func (u *LoginUseCase) WithMFA(
	mfaRepo user.MFARepository,
	encryptor *crypto.AESEncryptor,
	privacyRepo domainPrivacy.Repository,
) *LoginUseCase {
	u.mfaRepo = mfaRepo
	u.encryptor = encryptor
	u.privacyRepo = privacyRepo
	return u
}

func (u *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*TokenOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	usr, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("credenciais inválidas")
	}
	if usr == nil {
		return nil, errors.New("credenciais inválidas")
	}
	if !usr.IsActive {
		return nil, errors.New("credenciais inválidas")
	}

	if !u.passwordService.Check(input.Password, usr.PasswordHash) {
		return nil, errors.New("credenciais inválidas")
	}

	if u.mfaRepo != nil && u.encryptor != nil {
		mfa, err := u.mfaRepo.GetMFA(ctx, usr.ID)
		if err != nil {
			return nil, errors.New("erro interno ao validar MFA")
		}
		if mfa != nil && mfa.Enabled {
			if strings.TrimSpace(input.TOTPCode) == "" {
				return nil, ErrMFARequired
			}
			secret, err := u.encryptor.Decrypt(mfa.SecretEncrypted, []byte(usr.ID.String()))
			if err != nil {
				return nil, errors.New("erro interno ao validar MFA")
			}
			step, ok := totp.VerifyCode(secret, input.TOTPCode, time.Now().UTC())
			if !ok || step <= mfa.LastUsedStep {
				return nil, ErrInvalidMFA
			}
			if err := u.mfaRepo.UpdateMFALastUsedStep(ctx, usr.ID, step); err != nil {
				return nil, errors.New("erro ao registrar uso do MFA")
			}
		}
	}

	familyID := uuid.New()
	sessionID := uuid.New().String()

	rawRefreshToken, err := generateOpaqueToken()
	if err != nil {
		return nil, errors.New("erro interno ao gerar token")
	}

	hashedRT := crypto.SHA256Raw(rawRefreshToken)

	rtRecord := &user.RefreshToken{
		UserID:    usr.ID,
		TokenHash: hashedRT,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(u.refreshTTL),
		CreatedIP: input.ClientIP,
	}

	if err := u.userRepo.CreateRefreshToken(ctx, rtRecord); err != nil {
		return nil, errors.New("erro ao registrar sessão")
	}

	accessToken, err := u.jwtService.GenerateAccessToken(usr.ID, usr.Role, sessionID, u.accessTTL)
	if err != nil {
		return nil, errors.New("erro ao assinar token de acesso")
	}

	if u.privacyRepo != nil {
		actorID := usr.ID
		var ipPtr *string
		if strings.TrimSpace(input.ClientIP) != "" {
			ip := input.ClientIP
			ipPtr = &ip
		}
		_ = u.privacyRepo.RecordAuditEvent(ctx, &domainPrivacy.AuditEvent{
			ID:           uuid.New(),
			ActorUserID:  &actorID,
			Action:       "USER_LOGIN",
			TargetType:   "user",
			TargetID:     &actorID,
			MetadataJSON: "{}",
			IPAddress:    ipPtr,
			CreatedAt:    time.Now().UTC(),
		})
	}

	return &TokenOutput{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    int(u.accessTTL.Seconds()),
		Role:         usr.Role,
	}, nil
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
