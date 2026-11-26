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
	resetUseCase   *auth.PasswordResetUseCase
	mfaUseCase     *auth.MFAUseCase
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

func (h *AuthHandler) WithPhase3(resetUC *auth.PasswordResetUseCase, mfaUC *auth.MFAUseCase) *AuthHandler {
	h.resetUseCase = resetUC
	h.mfaUseCase = mfaUC
	return h
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup, authLimiter gin.HandlerFunc) {
	group := router.Group("/auth")
	{
		if authLimiter != nil {
			group.POST("/login", authLimiter, h.Login)
			group.POST("/refresh", authLimiter, h.Refresh)
			if h.resetUseCase != nil {
				group.POST("/password-reset/request", authLimiter, h.RequestPasswordReset)
				group.POST("/password-reset/confirm", authLimiter, h.ConfirmPasswordReset)
			}
		} else {
			group.POST("/login", h.Login)
			group.POST("/refresh", h.Refresh)
			if h.resetUseCase != nil {
				group.POST("/password-reset/request", h.RequestPasswordReset)
				group.POST("/password-reset/confirm", h.ConfirmPasswordReset)
			}
		}
		group.POST("/logout", h.Logout)
	}
}

func (h *AuthHandler) RegisterProtectedRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	if h.mfaUseCase == nil {
		return
	}
	mfa := router.Group("/auth/mfa")
	if auther != nil {
		mfa.Use(auther)
	}
	mfa.GET("/status", h.MFAStatus)
	mfa.POST("/setup", h.MFASetup)
	mfa.POST("/enable", h.MFAEnable)
	mfa.POST("/disable", h.MFADisable)
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
		if errors.Is(err, auth.ErrMFARequired) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":        err.Error(),
				"mfa_required": true,
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if wantsCookieSession(c) {
		if err := setSessionCookiesWithConfig(c, output, h.isProduction, h.refreshTTL); err != nil {
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
			clearSessionCookiesWithConfig(c, h.isProduction)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if fromCookie || wantsCookieSession(c) {
		if err := setSessionCookiesWithConfig(c, output, h.isProduction, h.refreshTTL); err != nil {
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
		clearSessionCookiesWithConfig(c, h.isProduction)
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var input auth.PasswordResetRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "e-mail válido obrigatório"})
		return
	}
	input.ClientIP = getClientIP(c)
	if err := h.resetUseCase.Request(c.Request.Context(), input); err != nil {
		respondDomainError(c, err, "falha ao solicitar redefinição de senha")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Se o e-mail estiver cadastrado, enviaremos as instruções de redefinição.",
	})
}

func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var input auth.PasswordResetConfirmInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token e nova senha (mínimo 8 caracteres) obrigatórios"})
		return
	}
	input.ClientIP = getClientIP(c)
	if err := h.resetUseCase.Confirm(c.Request.Context(), input); err != nil {
		respondDomainError(c, err, "falha ao redefinir senha")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Senha redefinida com sucesso. Todas as sessões anteriores foram encerradas.",
	})
}

func (h *AuthHandler) MFAStatus(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.mfaUseCase.Status(c.Request.Context(), claims.UserID)
	if err != nil {
		respondDomainError(c, err, "falha ao consultar status MFA")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AuthHandler) MFASetup(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.mfaUseCase.Setup(c.Request.Context(), claims.UserID, claims.Role)
	if err != nil {
		respondDomainError(c, err, "falha ao iniciar configuração MFA")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AuthHandler) MFAEnable(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	var input auth.MFAVerifyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "código TOTP de 6 dígitos obrigatório"})
		return
	}
	if err := h.mfaUseCase.Enable(c.Request.Context(), claims.UserID, claims.Role, input.Code); err != nil {
		respondDomainError(c, err, "falha ao ativar MFA")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MFA TOTP ativado com sucesso."})
}

func (h *AuthHandler) MFADisable(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	var input auth.MFADisableInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "senha atual e código TOTP de 6 dígitos obrigatórios"})
		return
	}
	if err := h.mfaUseCase.Disable(c.Request.Context(), claims.UserID, claims.Role, input.Password, input.Code); err != nil {
		respondDomainError(c, err, "falha ao desativar MFA")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MFA TOTP desativado com sucesso."})
}

func setSessionCookiesWithConfig(c *gin.Context, out *auth.TokenOutput, isProduction bool, refreshTTL time.Duration) error {
	csrfToken, err := generateCSRFToken()
	if err != nil {
		return err
	}

	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	secure := shouldUseSecureCookies(c, isProduction)
	refreshMaxAge := int(refreshTTL.Seconds())

	// Secure é true em produção e HTTPS; dev local HTTP exige flexibilização via shouldUseSecureCookies.
	// nosemgrep: go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.AccessCookieName,
		Value:    out.AccessToken,
		Path:     "/v1",
		MaxAge:   out.ExpiresIn,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	// nosemgrep: go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     middleware.RefreshCookieName,
		Value:    out.RefreshToken,
		Path:     "/v1/auth",
		MaxAge:   refreshMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	// Cookie CSRF usa padrão Double-Submit Cookie (ADR-008): precisa ser legível via JS para X-CSRF-Token; não contém sessão/JWT.
	// nosemgrep: go.lang.security.audit.net.cookie-missing-httponly.cookie-missing-httponly,go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure
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

func clearSessionCookiesWithConfig(c *gin.Context, isProduction bool) {
	secure := shouldUseSecureCookies(c, isProduction)
	expired := time.Unix(0, 0)

	// nosemgrep: go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure
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
	// nosemgrep: go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure
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
	// nosemgrep: go.lang.security.audit.net.cookie-missing-httponly.cookie-missing-httponly,go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure
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

func shouldUseSecureCookies(c *gin.Context, isProduction bool) bool {
	if isProduction || c.Request.TLS != nil {
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
