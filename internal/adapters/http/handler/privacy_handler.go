package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	appPrivacy "github.com/SCAPUTO88/fit-brother/internal/application/privacy"
)

type PrivacyHandler struct {
	uc           *appPrivacy.UseCase
	isProduction bool
}

func NewPrivacyHandler(uc *appPrivacy.UseCase, isProduction bool) *PrivacyHandler {
	return &PrivacyHandler{
		uc:           uc,
		isProduction: isProduction,
	}
}

func (h *PrivacyHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	priv := router.Group("/privacy")
	if auther != nil {
		priv.Use(auther)
	}
	priv.POST("/consents", h.RecordConsent)
	priv.GET("/consents", h.ListConsents)
	priv.POST("/requests", h.CreateRequest)
	priv.GET("/requests", h.ListRequests)
	priv.GET("/export", h.ExportMyData)
	priv.POST("/anonymize", h.AnonymizeSelf)

	audit := router.Group("/audit-events")
	if auther != nil {
		audit.Use(auther)
	}
	audit.GET("", h.ListAuditEvents)
}

func (h *PrivacyHandler) RecordConsent(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	var input appPrivacy.RecordConsentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados de consentimento inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)
	input.UserAgent = c.GetHeader("User-Agent")

	out, err := h.uc.RecordConsent(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao registrar consentimento")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *PrivacyHandler) ListConsents(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.uc.ListConsents(c.Request.Context(), claims.UserID)
	if err != nil {
		respondDomainError(c, err, "falha ao listar consentimentos")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *PrivacyHandler) CreateRequest(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	var input appPrivacy.CreatePrivacyRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados da solicitação inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.CreateRequest(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar solicitação LGPD")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *PrivacyHandler) ListRequests(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.uc.ListRequests(c.Request.Context(), claims.UserID)
	if err != nil {
		respondDomainError(c, err, "falha ao listar solicitações LGPD")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *PrivacyHandler) ExportMyData(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.uc.ExportMyData(c.Request.Context(), claims.UserID, getClientIP(c))
	if err != nil {
		respondDomainError(c, err, "falha ao exportar dados pessoais")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *PrivacyHandler) AnonymizeSelf(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	if err := h.uc.AnonymizeSelf(c.Request.Context(), claims.UserID, getClientIP(c)); err != nil {
		respondDomainError(c, err, "falha ao processar anonimização")
		return
	}
	if wantsCookieSession(c) || hasSessionCookies(c) {
		clearSessionCookiesWithConfig(c, h.isProduction)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Conta anonimizada e sessões encerradas com sucesso."})
}

func (h *PrivacyHandler) ListAuditEvents(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.uc.ListAuditEvents(c.Request.Context(), claims.UserID, claims.Role)
	if err != nil {
		respondDomainError(c, err, "falha ao consultar auditoria")
		return
	}
	c.JSON(http.StatusOK, out)
}
