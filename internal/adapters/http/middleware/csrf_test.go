package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
)

func TestCSRFProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(middleware.CSRFProtection([]string{"http://localhost:3000"}))
		r.POST("/v1/workouts", func(c *gin.Context) {
			c.Status(http.StatusCreated)
		})
		r.POST("/v1/auth/login", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		r.GET("/v1/workouts", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		return r
	}

	t.Run("Bloqueia POST via cookie sem X-CSRF-Token", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/workouts", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: "jwt-cookie"})
		req.AddCookie(&http.Cookie{Name: middleware.CSRFCookieName, Value: "csrf-secret-123"})
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Bloqueia POST via cookie com Origin nao permitido", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/workouts", nil)
		req.Header.Set("Origin", "https://evil.attacker.example")
		req.Header.Set(middleware.CSRFHeaderName, "csrf-secret-123")
		req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: "jwt-cookie"})
		req.AddCookie(&http.Cookie{Name: middleware.CSRFCookieName, Value: "csrf-secret-123"})
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Libera POST via cookie com Origin permitido e X-CSRF-Token correspondente", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/workouts", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set(middleware.CSRFHeaderName, "csrf-secret-123")
		req.AddCookie(&http.Cookie{Name: middleware.AccessCookieName, Value: "jwt-cookie"})
		req.AddCookie(&http.Cookie{Name: middleware.CSRFCookieName, Value: "csrf-secret-123"})
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Libera requisicao Bearer sem cookie CSRF", func(t *testing.T) {
		r := newRouter()
		req := httptest.NewRequest(http.MethodPost, "/v1/workouts", nil)
		req.Header.Set("Authorization", "Bearer mobile-jwt-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
