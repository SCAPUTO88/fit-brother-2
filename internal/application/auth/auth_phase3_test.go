package auth_test

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/auth"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainMailer "github.com/SCAPUTO88/fit-brother/internal/domain/mailer"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/totp"
)

type memMFARepo struct {
	items map[uuid.UUID]*domainU.UserMFA
}

func newMemMFARepo() *memMFARepo {
	return &memMFARepo{items: make(map[uuid.UUID]*domainU.UserMFA)}
}

func (m *memMFARepo) GetMFA(_ context.Context, userID uuid.UUID) (*domainU.UserMFA, error) {
	item := m.items[userID]
	if item == nil {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}

func (m *memMFARepo) UpsertMFA(_ context.Context, mfa *domainU.UserMFA) error {
	cp := *mfa
	m.items[mfa.UserID] = &cp
	return nil
}

func (m *memMFARepo) UpdateMFALastUsedStep(_ context.Context, userID uuid.UUID, step int64) error {
	if m.items[userID] != nil {
		m.items[userID].LastUsedStep = step
	}
	return nil
}

func (m *memMFARepo) DisableMFA(_ context.Context, userID uuid.UUID) error {
	delete(m.items, userID)
	return nil
}

type memResetRepo struct {
	tokens map[string]*domainU.PasswordResetToken
}

func newMemResetRepo() *memResetRepo {
	return &memResetRepo{tokens: make(map[string]*domainU.PasswordResetToken)}
}

func (m *memResetRepo) CreatePasswordResetToken(_ context.Context, token *domainU.PasswordResetToken) error {
	cp := *token
	m.tokens[token.TokenHash] = &cp
	return nil
}

func (m *memResetRepo) GetPasswordResetTokenByHash(_ context.Context, tokenHash string) (*domainU.PasswordResetToken, error) {
	t := m.tokens[tokenHash]
	if t == nil {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (m *memResetRepo) MarkPasswordResetTokenUsed(_ context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	for _, t := range m.tokens {
		if t.ID == id {
			t.UsedAt = &now
		}
	}
	return nil
}

func (m *memResetRepo) InvalidatePasswordResetTokensByUser(_ context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	for _, t := range m.tokens {
		if t.UserID == userID && t.UsedAt == nil {
			t.UsedAt = &now
		}
	}
	return nil
}

type captureMailer struct {
	resetTokens  []string
	inviteTokens []string
}

func (m *captureMailer) SendInvitation(_ context.Context, msg domainMailer.InvitationEmail) error {
	m.inviteTokens = append(m.inviteTokens, msg.RawToken)
	return nil
}

func (m *captureMailer) SendPasswordReset(_ context.Context, msg domainMailer.PasswordResetEmail) error {
	m.resetTokens = append(m.resetTokens, msg.RawToken)
	return nil
}

func TestTrainerMFALoginAndReplayProtection(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryUserRepo()
	mfaRepo := newMemMFARepo()
	passSvc, _ := crypto.NewPasswordService("pepper")
	aesKey := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	encryptor, err := crypto.NewAESEncryptor(aesKey)
	require.NoError(t, err)
	jwtSvc := setupJWTService(t)

	passHash, _ := passSvc.Hash("TrainerPass!2026")
	trainer := &domainU.User{
		ID:           uuid.New(),
		Email:        "trainer.mfa@fit.test",
		FullName:     "Trainer MFA",
		PasswordHash: passHash,
		Role:         domainU.RoleTrainer,
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, trainer))

	mfaUC := auth.NewMFAUseCase(repo, mfaRepo, nil, passSvc, encryptor)
	setupOut, err := mfaUC.Setup(ctx, trainer.ID, domainU.RoleTrainer)
	require.NoError(t, err)

	// Gera código válido usando passo anterior para deixar passo atual livre ao login seguinte
	pastStep := time.Now().UTC().Unix()/totp.PeriodSeconds - 1
	// Decodifica secret via VerifyCode não é necessário: geramos URI/code via GenerateSecret
	// Vamos habilitar com passo -1 para que o passo atual ainda possa ser usado no login.
	codePrev := computeTOTPForTest(t, setupOut.Secret, pastStep)
	require.NoError(t, mfaUC.Enable(ctx, trainer.ID, domainU.RoleTrainer, codePrev))

	loginUC := auth.NewLoginUseCase(repo, passSvc, jwtSvc, 15*time.Minute, 24*time.Hour).
		WithMFA(mfaRepo, encryptor, nil)

	// Sem código TOTP deve exigir MFA
	_, err = loginUC.Execute(ctx, auth.LoginInput{
		Email:    trainer.Email,
		Password: "TrainerPass!2026",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrMFARequired))

	// Com código TOTP atual deve autenticar
	curStep := time.Now().UTC().Unix() / totp.PeriodSeconds
	codeCur := computeTOTPForTest(t, setupOut.Secret, curStep)
	out, err := loginUC.Execute(ctx, auth.LoginInput{
		Email:    trainer.Email,
		Password: "TrainerPass!2026",
		TOTPCode: codeCur,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)

	// Reuso do mesmo código na mesma janela deve ser bloqueado (anti-replay)
	_, err = loginUC.Execute(ctx, auth.LoginInput{
		Email:    trainer.Email,
		Password: "TrainerPass!2026",
		TOTPCode: codeCur,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrInvalidMFA))
}

func TestPasswordResetSingleUseAndRevokesSessions(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryUserRepo()
	resetRepo := newMemResetRepo()
	mailer := &captureMailer{}
	passSvc, _ := crypto.NewPasswordService("pepper")

	oldHash, _ := passSvc.Hash("OldPass!2026")
	usr := &domainU.User{
		ID:           uuid.New(),
		Email:        "user.reset@fit.test",
		FullName:     "Usuario Reset",
		PasswordHash: oldHash,
		Role:         domainU.RoleStudent,
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, usr))

	rt := &domainU.RefreshToken{
		ID:        uuid.New(),
		UserID:    usr.ID,
		TokenHash: "hash-active-rt",
		FamilyID:  uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	require.NoError(t, repo.CreateRefreshToken(ctx, rt))

	uc := auth.NewPasswordResetUseCase(repo, resetRepo, nil, &MockTxManager{}, passSvc, mailer)

	// Anti-enumeração: email inexistente retorna nil sem enviar email
	require.NoError(t, uc.Request(ctx, auth.PasswordResetRequestInput{Email: "naoexiste@fit.test"}))
	assert.Empty(t, mailer.resetTokens)

	// Email válido envia 1 token
	require.NoError(t, uc.Request(ctx, auth.PasswordResetRequestInput{Email: usr.Email}))
	require.Len(t, mailer.resetTokens, 1)
	token := mailer.resetTokens[0]

	require.NoError(t, uc.Confirm(ctx, auth.PasswordResetConfirmInput{
		Token:       token,
		NewPassword: "NewStrongPass!2026",
	}))

	updated, err := repo.GetByID(ctx, usr.ID)
	require.NoError(t, err)
	assert.True(t, passSvc.Check("NewStrongPass!2026", updated.PasswordHash))

	// Refresh tokens anteriores revogados
	revokedRT, _ := repo.GetRefreshTokenByHash(ctx, "hash-active-rt")
	assert.True(t, revokedRT.IsRevoked)

	// Reuso do mesmo token deve falhar
	err = uc.Confirm(ctx, auth.PasswordResetConfirmInput{
		Token:       token,
		NewPassword: "AnotherPass!2026",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, errs.ErrUnauthorized))
}

func computeTOTPForTest(t *testing.T, secret string, step int64) string {
	t.Helper()
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	require.NoError(t, err)
	return totp.GenerateCodeForStep(key, step)
}
