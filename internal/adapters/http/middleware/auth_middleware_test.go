package middleware_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

type stubUserStatusChecker struct {
	active bool
	err    error
}

func (s *stubUserStatusChecker) IsUserActive(context.Context, uuid.UUID) (bool, error) {
	return s.active, s.err
}

func newTestJWTService(t *testing.T) *jwt.Service {
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

func TestRequireAuthAndSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := newTestJWTService(t)
	userID := uuid.New()
	token, err := jwtSvc.GenerateAccessToken(userID, "TRAINER", uuid.NewString(), 15*time.Minute)
	require.NoError(t, err)

	t.Run("Deve bloquear usuario inativo ou excluido mesmo com JWT valido", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.SecurityHeaders())
		r.GET("/protected", middleware.RequireAuth(jwtSvc, &stubUserStatusChecker{active: false}), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	})

	t.Run("Deve retornar 500 se consulta de status falhar", func(t *testing.T) {
		r := gin.New()
		r.GET("/protected", middleware.RequireAuth(jwtSvc, &stubUserStatusChecker{err: errors.New("db down")}), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db down")
	})

	t.Run("Deve liberar usuario ativo com claims no contexto", func(t *testing.T) {
		r := gin.New()
		r.GET("/protected", middleware.RequireAuth(jwtSvc, &stubUserStatusChecker{active: true}), func(c *gin.Context) {
			claims, ok := middleware.GetClaims(c)
			require.True(t, ok)
			c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID.String()})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), userID.String())
	})
}
