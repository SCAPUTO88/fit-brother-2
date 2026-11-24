package auth

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

type loginUserRepository struct {
	user *user.User
}

func (r *loginUserRepository) Create(context.Context, *user.User) error { return nil }

func (r *loginUserRepository) GetByID(context.Context, uuid.UUID) (*user.User, error) {
	return nil, nil
}

func (r *loginUserRepository) GetByEmail(context.Context, string) (*user.User, error) {
	return r.user, nil
}

func (r *loginUserRepository) Update(context.Context, *user.User) error { return nil }

func (r *loginUserRepository) CreateRefreshToken(context.Context, *user.RefreshToken) error {
	return nil
}

func (r *loginUserRepository) GetRefreshTokenByHash(context.Context, string) (*user.RefreshToken, error) {
	return nil, nil
}

func (r *loginUserRepository) RevokeRefreshToken(context.Context, uuid.UUID) error { return nil }

func (r *loginUserRepository) RevokeTokenFamily(context.Context, uuid.UUID) error { return nil }

func (r *loginUserRepository) RevokeAllByUserID(context.Context, uuid.UUID) error { return nil }

func TestLoginUseCaseDoesNotLogPasswordHashOnInvalidPassword(t *testing.T) {
	passwordService, err := crypto.NewPasswordService("test-pepper")
	require.NoError(t, err)

	passwordHash, err := passwordService.Hash("correct-password")
	require.NoError(t, err)

	repository := &loginUserRepository{user: &user.User{
		ID:           uuid.New(),
		Email:        "trainer@fitbrother.dev",
		PasswordHash: passwordHash,
		Role:         user.RoleTrainer,
		IsActive:     true,
	}}
	useCase := NewLoginUseCase(repository, passwordService, nil, 15*time.Minute, 7*24*time.Hour)

	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previousOutput) })

	output, err := useCase.Execute(context.Background(), LoginInput{
		Email:    "trainer@fitbrother.dev",
		Password: "wrong-password",
	})

	require.Nil(t, output)
	require.EqualError(t, err, "credenciais inválidas")
	require.NotContains(t, logs.String(), passwordHash)
}
