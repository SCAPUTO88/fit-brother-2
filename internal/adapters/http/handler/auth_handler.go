package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/application/auth"
)

type AuthHandler struct {
	loginUseCase   *auth.LoginUseCase
	refreshUseCase *auth.RefreshUseCase
	logoutUseCase  *auth.LogoutUseCase
	isProduction   bool
	refreshTTL     time.Duration
}

func NewAuthHandler(
	login *auth.LoginUseCase,
	refresh *auth.RefreshUseCase,
	logout *auth.LogoutUseCase,
	isProduction bool,
	refreshTTL time.Duration,
) *AuthHandler {
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &AuthHandler{
		loginUseCase:   login,
		refreshUseCase: refresh,
		logoutUseCase:  logout,
		isProduction:   isProduction,
		refreshTTL:     refreshTTL,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup, authLimiter gin.HandlerFunc) {
	group := router.Group("/auth")
	{
		if authLimiter != nil {
			group.POST("/login", authLimiter, h.Login)
			group.POST("/refresh", authLimiter, h.Refresh)
		} else {
			group.POST("/login", h.Login)
			group.POST("/refresh", h.Refresh)
		}
		group.POST("/logout", h.Logout)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input auth.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	input.ClientIP = getClientIP(c)

	output, err := h.loginUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if wantsCookieSession(c) {
		if err := h.setSessionCookies(c, output); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao emitir sessão segura"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"expires_in":   output.ExpiresIn,
			"role":         output.Role,
			"session_mode": "cookie",
		})
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input auth.RefreshInput
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
			return
		}
	}

	fromCookie := false
	if strings.TrimSpace(input.RefreshToken) == "" {
		if cookieRT, err := c.Cookie(middleware.RefreshCookieName); err == nil && strings.TrimSpace(cookieRT) != "" {
			input.RefreshToken = strings.TrimSpace(cookieRT)
			fromCookie = true
		}
	}

	if strings.TrimSpace(input.RefreshToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token obrigatório"})
		return
	}

	input.ClientIP = getClientIP(c)

	output, err := h.refreshUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		if fromCookie || wantsCookieSession(c) {
			h.clearSessionCookies(c)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if fromCookie || wantsCookieSession(c) {
		if err := h.setSessionCookies(c, output); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao renovar sessão segura"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"expires_in":   output.ExpiresIn,
			"role":         output.Role,
			"session_mode": "cookie",
		})
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input auth.LogoutInput
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
			return
		}
	}

	fromCookie := false
	if strings.TrimSpace(input.RefreshToken) == "" {
		if cookieRT, err := c.Cookie(middleware.RefreshCookieName); err == nil && strings.TrimSpace(cookieRT) != "" {
			input.RefreshToken = strings.TrimSpace(cookieRT)
			fromCookie = true
		}
	}

	if strings.TrimSpace(input.RefreshToken) != "" {
		if err := h.logoutUseCase.Execute(c.Request.Context(), input); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao encerrar sessão"})
			return
		}
	}

	if fromCookie || wantsCookieSession(c) || hasSessionCookies(c) {
		h.clearSessionCookies(c)
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) setSessionCookies(c *gin.Context, out *auth.TokenOutput) error {
	csrfToken, err := generateCSRFToken()
	if err != nil {
		return err
	}

	secure := h.shouldUseSecureCookies(c)
	refreshMaxAge := int(h.refreshTTL.Seconds())

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.AccessCookieName,
		Value:    out.AccessToken,
		Path:     "/v1",
		MaxAge:   out.ExpiresIn,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.RefreshCookieName,
		Value:    out.RefreshToken,
		Path:     "/v1/auth",
		MaxAge:   refreshMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		MaxAge:   refreshMaxAge,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (h *AuthHandler) clearSessionCookies(c *gin.Context) {
	secure := h.shouldUseSecureCookies(c)
	expired := time.Unix(0, 0)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.AccessCookieName,
		Value:    "",
		Path:     "/v1",
		MaxAge:   -1,
		Expires:  expired,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.RefreshCookieName,
		Value:    "",
		Path:     "/v1/auth",
		MaxAge:   -1,
		Expires:  expired,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  expired,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) shouldUseSecureCookies(c *gin.Context) bool {
	if h.isProduction || c.Request.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https")
}

func wantsCookieSession(c *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(c.GetHeader(middleware.SessionModeHeaderName)), "cookie")
}

func hasSessionCookies(c *gin.Context) bool {
	for _, name := range []string{middleware.AccessCookieName, middleware.RefreshCookieName, middleware.CSRFCookieName} {
		if val, err := c.Cookie(name); err == nil && val != "" {
			return true
		}
	}
	return false
}

func generateCSRFToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func getClientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		return "127.0.0.1"
	}
	if idx := strings.LastIndex(ip, ":"); idx != -1 && strings.Count(ip, ":") == 1 {
		ip = ip[:idx]
	}
	return ip
}
