package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/pem"
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
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/totp"
)

type memUserRepo struct {
	users  map[uuid.UUID]*domainU.User
	tokens map[string]*domainU.RefreshToken
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		users:  make(map[uuid.UUID]*domainU.User),
		tokens: make(map[string]*domainU.RefreshToken),
	}
}

func (m *memUserRepo) Create(_ context.Context, u *domainU.User) error {
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *memUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domainU.User, error) {
	u := m.users[id]
	if u == nil {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) GetByEmail(_ context.Context, email string) (*domainU.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memUserRepo) Update(_ context.Context, u *domainU.User) error {
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *memUserRepo) CreateRefreshToken(_ context.Context, rt *domainU.RefreshToken) error {
	cp := *rt
	m.tokens[rt.TokenHash] = &cp
	return nil
}

func (m *memUserRepo) GetRefreshTokenByHash(_ context.Context, tokenHash string) (*domainU.RefreshToken, error) {
	rt := m.tokens[tokenHash]
	if rt == nil {
		return nil, nil
	}
	cp := *rt
	return &cp, nil
}

func (m *memUserRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID) error {
	for _, rt := range m.tokens {
		if rt.ID == id {
			rt.IsRevoked = true
		}
	}
	return nil
}

func (m *memUserRepo) RevokeTokenFamily(_ context.Context, familyID uuid.UUID) error {
	for _, rt := range m.tokens {
		if rt.FamilyID == familyID {
			rt.IsRevoked = true
		}
	}
	return nil
}

func (m *memUserRepo) RevokeAllByUserID(_ context.Context, userID uuid.UUID) error {
	for _, rt := range m.tokens {
		if rt.UserID == userID {
			rt.IsRevoked = true
		}
	}
	return nil
}

type noopTxManager struct{}

func (m *noopTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func setupJWTService(t *testing.T) *jwt.Service {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	svc, err := jwt.NewServiceFromPEM(string(privPEM), string(pubPEM), "fit-auth-test")
	require.NoError(t, err)
	return svc
}

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
	repo := newMemUserRepo()
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

	pastStep := time.Now().UTC().Unix()/totp.PeriodSeconds - 1
	codePrev := computeTOTPForTest(t, setupOut.Secret, pastStep)
	require.NoError(t, mfaUC.Enable(ctx, trainer.ID, domainU.RoleTrainer, codePrev))

	loginUC := auth.NewLoginUseCase(repo, passSvc, jwtSvc, 15*time.Minute, 24*time.Hour).
		WithMFA(mfaRepo, encryptor, nil)

	_, err = loginUC.Execute(ctx, auth.LoginInput{
		Email:    trainer.Email,
		Password: "TrainerPass!2026",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrMFARequired))

	curStep := time.Now().UTC().Unix() / totp.PeriodSeconds
	codeCur := computeTOTPForTest(t, setupOut.Secret, curStep)
	out, err := loginUC.Execute(ctx, auth.LoginInput{
		Email:    trainer.Email,
		Password: "TrainerPass!2026",
		TOTPCode: codeCur,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)

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
	repo := newMemUserRepo()
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

	uc := auth.NewPasswordResetUseCase(repo, resetRepo, nil, &noopTxManager{}, passSvc, mailer)

	require.NoError(t, uc.Request(ctx, auth.PasswordResetRequestInput{Email: "naoexiste@fit.test"}))
	assert.Empty(t, mailer.resetTokens)

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

	revokedRT, _ := repo.GetRefreshTokenByHash(ctx, "hash-active-rt")
	assert.True(t, revokedRT.IsRevoked)

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
