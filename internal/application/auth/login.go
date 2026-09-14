package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

type LoginUseCase struct {
	userRepo        user.Repository
	passwordService *crypto.PasswordService
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

func (u *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*TokenOutput, error) {
	usr, err := u.userRepo.GetByEmail(ctx, input.Email)
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
