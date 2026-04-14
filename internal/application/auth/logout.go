package auth

import (
	"context"
	"errors"

	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

type LogoutUseCase struct {
	userRepo user.Repository
}

func NewLogoutUseCase(repo user.Repository) *LogoutUseCase {
	return &LogoutUseCase{userRepo: repo}
}

// Execute faz o soft logout: pega o refresh token providenciado e a família dele e encerra
func (u *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	hashedRT := crypto.SHA256Raw(input.RefreshToken)

	rtRecord, err := u.userRepo.GetRefreshTokenByHash(ctx, hashedRT)
	if err != nil || rtRecord == nil {
		// Retornar suceso mesmo que não encontre pra previnir enumeracao.
		return nil
	}

	if rtRecord.IsRevoked {
		return nil
	}

	// Como é o logout, revogamos a família toda
	if err := u.userRepo.RevokeTokenFamily(ctx, rtRecord.FamilyID); err != nil {
		return errors.New("falha revogando a sessão")
	}

	return nil
}
