package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

type RefreshUseCase struct {
	userRepo   user.Repository
	txManager  tx.Manager
	jwtService *jwt.Service
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewRefreshUseCase(
	repo user.Repository,
	txManager tx.Manager,
	js *jwt.Service,
	accessTTL, refreshTTL time.Duration,
) *RefreshUseCase {
	return &RefreshUseCase{
		userRepo:   repo,
		txManager:  txManager,
		jwtService: js,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (u *RefreshUseCase) Execute(ctx context.Context, input RefreshInput) (*TokenOutput, error) {
	var out *TokenOutput
	var compromisedFamilyID uuid.UUID

	run := func(execCtx context.Context) error {
		hashedRT := crypto.SHA256Raw(input.RefreshToken)

		rtRecord, err := u.userRepo.GetRefreshTokenByHash(execCtx, hashedRT)
		if err != nil || rtRecord == nil {
			return errors.New("refresh token inválido")
		}

		if rtRecord.IsRevoked {
			compromisedFamilyID = rtRecord.FamilyID
			return errors.New("token comprometido, sessão encerrada por segurança")
		}

		if !rtRecord.IsValid() {
			return errors.New("refresh token expirado, faça login novamente")
		}

		usr, err := u.userRepo.GetByID(execCtx, rtRecord.UserID)
		if err != nil || usr == nil || !usr.IsActive || usr.DeletedAt != nil {
			return errors.New("usuário desativado ou não encontrado")
		}

		if err := u.userRepo.RevokeRefreshToken(execCtx, rtRecord.ID); err != nil {
			return errors.New("falha interna rotacionando chave")
		}

		newRawRefreshToken, err := generateOpaqueToken()
		if err != nil {
			return errors.New("erro interno")
		}

		newHashedRT := crypto.SHA256Raw(newRawRefreshToken)

		newRtRecord := &user.RefreshToken{
			UserID:    usr.ID,
			TokenHash: newHashedRT,
			FamilyID:  rtRecord.FamilyID,
			ExpiresAt: time.Now().Add(u.refreshTTL),
			CreatedIP: input.ClientIP,
		}

		if err := u.userRepo.CreateRefreshToken(execCtx, newRtRecord); err != nil {
			return errors.New("falha salvando nova sessão")
		}

		sessionID := uuid.New().String()
		accessToken, err := u.jwtService.GenerateAccessToken(usr.ID, usr.Role, sessionID, u.accessTTL)
		if err != nil {
			return err
		}

		out = &TokenOutput{
			AccessToken:  accessToken,
			RefreshToken: newRawRefreshToken,
			ExpiresIn:    int(u.accessTTL.Seconds()),
			Role:         usr.Role,
		}
		return nil
	}

	if u.txManager != nil {
		if err := u.txManager.RunInTx(ctx, run); err != nil {
			if compromisedFamilyID != uuid.Nil {
				_ = u.userRepo.RevokeTokenFamily(ctx, compromisedFamilyID)
			}
			return nil, err
		}
		return out, nil
	}

	if err := run(ctx); err != nil {
		if compromisedFamilyID != uuid.Nil {
			_ = u.userRepo.RevokeTokenFamily(ctx, compromisedFamilyID)
		}
		return nil, err
	}
	return out, nil
}
