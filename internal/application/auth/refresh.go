package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

type RefreshUseCase struct {
	userRepo   user.Repository
	jwtService *jwt.Service
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewRefreshUseCase(
	repo user.Repository,
	js *jwt.Service,
	accessTTL, refreshTTL time.Duration,
) *RefreshUseCase {
	return &RefreshUseCase{
		userRepo:   repo,
		jwtService: js,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (u *RefreshUseCase) Execute(ctx context.Context, input RefreshInput) (*TokenOutput, error) {
	hashedRT := crypto.SHA256Raw(input.RefreshToken)

	rtRecord, err := u.userRepo.GetRefreshTokenByHash(ctx, hashedRT)
	if err != nil || rtRecord == nil {
		return nil, errors.New("refresh token inválido")
	}

	if rtRecord.IsRevoked {
		_ = u.userRepo.RevokeTokenFamily(ctx, rtRecord.FamilyID)
		return nil, errors.New("token comprometido, sessão encerrada por segurança")
	}

	if !rtRecord.IsValid() {
		return nil, errors.New("refresh token expirado, faça login novamente")
	}

	if err := u.userRepo.RevokeRefreshToken(ctx, rtRecord.ID); err != nil {
		return nil, errors.New("falha interna rotacionando chave")
	}

	usr, err := u.userRepo.GetByID(ctx, rtRecord.UserID)
	if err != nil || usr == nil || !usr.IsActive {
		return nil, errors.New("usuário desativado ou não encontrado")
	}

	newRawRefreshToken, err := generateOpaqueToken()
	if err != nil {
		return nil, errors.New("erro interno")
	}

	newHashedRT := crypto.SHA256Raw(newRawRefreshToken)

	newRtRecord := &user.RefreshToken{
		UserID:    usr.ID,
		TokenHash: newHashedRT,
		FamilyID:  rtRecord.FamilyID,
		ExpiresAt: time.Now().Add(u.refreshTTL),
		CreatedIP: input.ClientIP,
	}

	if err := u.userRepo.CreateRefreshToken(ctx, newRtRecord); err != nil {
		return nil, errors.New("falha salvando nova sessão")
	}

	sessionID := uuid.New().String()
	accessToken, err := u.jwtService.GenerateAccessToken(usr.ID, usr.Role, sessionID, u.accessTTL)
	if err != nil {
		return nil, err
	}

	return &TokenOutput{
		AccessToken:  accessToken,
		RefreshToken: newRawRefreshToken,
		ExpiresIn:    int(u.accessTTL.Seconds()),
		Role:         usr.Role,
	}, nil
}
