package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	appInv "github.com/SCAPUTO88/fit-brother/internal/application/invitation"
)

type InvitationHandler struct {
	uc           *appInv.UseCase
	isProduction bool
	refreshTTL   time.Duration
}

func NewInvitationHandler(uc *appInv.UseCase, isProduction bool, refreshTTL time.Duration) *InvitationHandler {
	return &InvitationHandler{
		uc:           uc,
		isProduction: isProduction,
		refreshTTL:   refreshTTL,
	}
}

func (h *InvitationHandler) RegisterRoutes(router *gin.RouterGroup, authLimiter, auther gin.HandlerFunc) {
	pub := router.Group("/auth/invitations")
	if authLimiter != nil {
		pub.POST("/accept", authLimiter, h.Accept)
	} else {
		pub.POST("/accept", h.Accept)
	}

	priv := router.Group("/invitations")
	if auther != nil {
		priv.Use(auther)
	}
	priv.POST("", h.Create)
	priv.GET("", h.List)
	priv.DELETE("/:id", h.Revoke)
}

func (h *InvitationHandler) Create(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	var input appInv.CreateInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	input.TrainerUserID = claims.UserID
	input.ClientIP = getClientIP(c)

	out, err := h.uc.Create(c.Request.Context(), claims.Role, input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar convite")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *InvitationHandler) List(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.uc.List(c.Request.Context(), claims.UserID, claims.Role)
	if err != nil {
		respondDomainError(c, err, "falha ao listar convites")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *InvitationHandler) Revoke(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	invID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do convite inválido"})
		return
	}
	if err := h.uc.Revoke(c.Request.Context(), invID, claims.UserID, claims.Role, getClientIP(c)); err != nil {
		respondDomainError(c, err, "falha ao revogar convite")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Convite revogado com sucesso."})
}

func (h *InvitationHandler) Accept(c *gin.Context) {
	var input appInv.AcceptInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos para ativação do convite"})
		return
	}
	input.ClientIP = getClientIP(c)
	input.UserAgent = c.GetHeader("User-Agent")

	out, err := h.uc.Accept(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "falha ao ativar convite")
		return
	}

	if wantsCookieSession(c) {
		sessionCopy := out.Session
		if err := setSessionCookiesWithConfig(c, &sessionCopy, h.isProduction, h.refreshTTL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "conta ativada, mas houve falha ao emitir sessão segura"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"student_id":      out.StudentID,
			"organization_id": out.OrganizationID,
			"session": gin.H{
				"expires_in":   out.Session.ExpiresIn,
				"role":         out.Session.Role,
				"session_mode": "cookie",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, out)
}
