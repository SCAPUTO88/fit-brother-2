package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterBlocksExcessAndResetsAfterWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Unix(1700000000, 0)
	limiter := NewRateLimiter(3, time.Minute, time.Minute)
	limiter.nowFunc = func() time.Time { return now }

	r := gin.New()
	r.POST("/login", limiter.Middleware("auth_login"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "198.51.100.10:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 4a tentativa na mesma janela deve ser bloqueada com 429 e Retry-After
	reqBlocked := httptest.NewRequest(http.MethodPost, "/login", nil)
	reqBlocked.RemoteAddr = "198.51.100.10:12345"
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)
	assert.Equal(t, http.StatusTooManyRequests, wBlocked.Code)
	assert.NotEmpty(t, wBlocked.Header().Get("Retry-After"))

	// Outro IP continua independente
	reqOther := httptest.NewRequest(http.MethodPost, "/login", nil)
	reqOther.RemoteAddr = "198.51.100.20:54321"
	wOther := httptest.NewRecorder()
	r.ServeHTTP(wOther, reqOther)
	assert.Equal(t, http.StatusOK, wOther.Code)

	// Apos passar a janela, IP original volta a ser permitido e entradas expiradas sao limpas
	now = now.Add(2 * time.Minute)
	reqAfter := httptest.NewRequest(http.MethodPost, "/login", nil)
	reqAfter.RemoteAddr = "198.51.100.10:12345"
	wAfter := httptest.NewRecorder()
	r.ServeHTTP(wAfter, reqAfter)
	assert.Equal(t, http.StatusOK, wAfter.Code)
}
