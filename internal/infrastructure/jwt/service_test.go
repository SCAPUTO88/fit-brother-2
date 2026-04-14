package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

// Helper para gerar um par de chaves RSA em arquivos temporários
func generateTempRSAKeys(t *testing.T) (privPath, pubPath string) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tempDir := t.TempDir()
	privPath = filepath.Join(tempDir, "private.pem")
	pubPath = filepath.Join(tempDir, "public.pem")

	// Grava Private
	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	require.NoError(t, os.WriteFile(privPath, privPEM, 0600))

	// Grava Public
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	require.NoError(t, os.WriteFile(pubPath, pubPEM, 0644))

	return privPath, pubPath
}

func TestJWTService(t *testing.T) {
	privPath, pubPath := generateTempRSAKeys(t)

	svc, err := jwt.NewService(privPath, pubPath, "test-issuer")
	require.NoError(t, err)

	t.Run("Gerar e validar Token", func(t *testing.T) {
		userID := uuid.New()
		role := "TRAINER"
		sessionID := uuid.New().String()

		tokenString, err := svc.GenerateAccessToken(userID, role, sessionID, 5*time.Minute)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		claims, err := svc.Verify(tokenString)
		require.NoError(t, err)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, sessionID, claims.SessionID)
		assert.Equal(t, "test-issuer", claims.Issuer)
	})

	t.Run("Falhar com token expirado", func(t *testing.T) {
		userID := uuid.New()
		
		// Criar com TTL negativo (expira no passado!)
		tokenString, err := svc.GenerateAccessToken(userID, "STUDENT", "sid", -1*time.Minute)
		require.NoError(t, err)

		claims, err := svc.Verify(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
		assert.Nil(t, claims)
	})

	t.Run("Falhar ao verificar token de outra fonte (assinatura invalida)", func(t *testing.T) {
		// Pega um svc mas agora eu gero UM NOVO par de chaves, pra simular um token "falso" ou 
		// de outro lugar.
		fakePriv, fakePub := generateTempRSAKeys(t)
		fakeSvc, _ := jwt.NewService(fakePriv, fakePub, "hacker-issuer")

		hackerToken, _ := fakeSvc.GenerateAccessToken(uuid.New(), "TRAINER", "123", 5*time.Minute)

		// O nosso Serviço Real tenta validar o hackertToken (espera falhar)
		claims, err := svc.Verify(hackerToken)
		assert.Error(t, err)
		// Detalhe: "signature is invalid" indica que nossa chave publica rejeitou a conta
		assert.Contains(t, err.Error(), "signature is invalid")
		assert.Nil(t, claims)
	})
}
