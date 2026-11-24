package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

type mockRefreshRepo struct {
	rtByHash         *user.RefreshToken
	usr              *user.User
	revokedTokenID   uuid.UUID
	revokedFamilyID  uuid.UUID
	createdNewTokens int
}

func (m *mockRefreshRepo) Create(context.Context, *user.User) error {
	return nil
}

func (m *mockRefreshRepo) GetByID(context.Context, uuid.UUID) (*user.User, error) {
	return m.usr, nil
}

func (m *mockRefreshRepo) GetByEmail(context.Context, string) (*user.User, error) {
	return nil, nil
}

func (m *mockRefreshRepo) Update(context.Context, *user.User) error {
	return nil
}

func (m *mockRefreshRepo) CreateRefreshToken(_ context.Context, _ *user.RefreshToken) error {
	m.createdNewTokens++
	return nil
}

func (m *mockRefreshRepo) GetRefreshTokenByHash(context.Context, string) (*user.RefreshToken, error) {
	return m.rtByHash, nil
}

func (m *mockRefreshRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID) error {
	m.revokedTokenID = id
	return nil
}

func (m *mockRefreshRepo) RevokeTokenFamily(_ context.Context, familyID uuid.UUID) error {
	m.revokedFamilyID = familyID
	return nil
}

func (m *mockRefreshRepo) RevokeAllByUserID(context.Context, uuid.UUID) error {
	return nil
}

type inlineTxManager struct {
	calls int
}

func (m *inlineTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	m.calls++
	return fn(ctx)
}

func testJWTService(t *testing.T) *jwt.Service {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	svc, err := jwt.NewServiceFromPEM(string(privPEM), string(pubPEM), "fit-brother-test")
	require.NoError(t, err)
	return svc
}

func TestRefreshUseCase_SecurityBaseline(t *testing.T) {
	jwtSvc := testJWTService(t)
	rawToken := "refresh-token-secreto"
	tokenHash := crypto.SHA256Raw(rawToken)
	userID := uuid.New()
	familyID := uuid.New()

	t.Run("Deve revogar toda a familia quando refresh token ja revogado for reutilizado", func(t *testing.T) {
		repo := &mockRefreshRepo{
			rtByHash: &user.RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: tokenHash,
				FamilyID:  familyID,
				IsRevoked: true,
				ExpiresAt: time.Now().Add(time.Hour),
			},
		}
		txMgr := &inlineTxManager{}
		uc := NewRefreshUseCase(repo, txMgr, jwtSvc, 15*time.Minute, 24*time.Hour)

		out, err := uc.Execute(context.Background(), RefreshInput{RefreshToken: rawToken})
		require.Nil(t, out)
		require.Error(t, err)
		assert.Equal(t, 1, txMgr.calls)
		assert.Equal(t, familyID, repo.revokedFamilyID)
		assert.Equal(t, 0, repo.createdNewTokens)
	})

	t.Run("Deve rejeitar refresh de usuario desativado antes de emitir novo token", func(t *testing.T) {
		repo := &mockRefreshRepo{
			rtByHash: &user.RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: tokenHash,
				FamilyID:  familyID,
				IsRevoked: false,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			usr: &user.User{
				ID:       userID,
				Role:     user.RoleStudent,
				IsActive: false,
			},
		}
		txMgr := &inlineTxManager{}
		uc := NewRefreshUseCase(repo, txMgr, jwtSvc, 15*time.Minute, 24*time.Hour)

		out, err := uc.Execute(context.Background(), RefreshInput{RefreshToken: rawToken})
		require.Nil(t, out)
		require.Error(t, err)
		assert.Equal(t, uuid.Nil, repo.revokedTokenID)
		assert.Equal(t, 0, repo.createdNewTokens)
	})

	t.Run("Deve rotacionar token validamente dentro de transacao para usuario ativo", func(t *testing.T) {
		rtID := uuid.New()
		repo := &mockRefreshRepo{
			rtByHash: &user.RefreshToken{
				ID:        rtID,
				UserID:    userID,
				TokenHash: tokenHash,
				FamilyID:  familyID,
				IsRevoked: false,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			usr: &user.User{
				ID:       userID,
				Role:     user.RoleTrainer,
				IsActive: true,
			},
		}
		txMgr := &inlineTxManager{}
		uc := NewRefreshUseCase(repo, txMgr, jwtSvc, 15*time.Minute, 24*time.Hour)

		out, err := uc.Execute(context.Background(), RefreshInput{RefreshToken: rawToken, ClientIP: "127.0.0.1"})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.NotEmpty(t, out.AccessToken)
		assert.NotEmpty(t, out.RefreshToken)
		assert.NotEqual(t, rawToken, out.RefreshToken)
		assert.Equal(t, rtID, repo.revokedTokenID)
		assert.Equal(t, 1, repo.createdNewTokens)
	})
}
