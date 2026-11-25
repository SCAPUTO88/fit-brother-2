package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
)

func TestCORSConfigUsesConfiguredOrigins(t *testing.T) {
	corsConfig, err := newCORSConfig("http://localhost:3000, http://localhost:5173")
	require.NoError(t, err)

	require.Equal(t, []string{
		"http://localhost:3000",
		"http://localhost:5173",
	}, corsConfig.AllowOrigins)
}

func TestCORSConfigRejectsWildcard(t *testing.T) {
	_, err := newCORSConfig("http://localhost:3000, *")
	require.Error(t, err)
}

func TestNewJWTServiceUsesPEMConfig(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})

	service, err := newJWTService(config.JWTConfig{
		PrivateKeyPEM: string(privatePEM),
		PublicKeyPEM:  string(publicPEM),
	})

	require.NoError(t, err)
	require.NotNil(t, service)
}
