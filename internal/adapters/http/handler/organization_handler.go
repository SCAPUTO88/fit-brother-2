package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	appOrg "github.com/SCAPUTO88/fit-brother/internal/application/organization"
)

type OrganizationHandler struct {
	uc *appOrg.UseCase
}

func NewOrganizationHandler(uc *appOrg.UseCase) *OrganizationHandler {
	return &OrganizationHandler{uc: uc}
}

func (h *OrganizationHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	orgs := router.Group("/organizations")
	if auther != nil {
		orgs.Use(auther)
	}
	orgs.POST("", h.Create)
	orgs.GET("", h.ListMine)
	orgs.GET("/:id/members", h.ListMembers)
	orgs.POST("/:id/members", h.AddMember)
	orgs.PATCH("/:id/relationships/:student_id", h.UpdateRelationshipStatus)
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	var input appOrg.CreateOrganizationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.CreateOrganization(c.Request.Context(), claims.UserID, claims.Role, input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar organização")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *OrganizationHandler) ListMine(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	out, err := h.uc.ListMyOrganizations(c.Request.Context(), claims.UserID)
	if err != nil {
		respondDomainError(c, err, "falha ao listar organizações")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *OrganizationHandler) ListMembers(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization id inválido"})
		return
	}
	out, err := h.uc.ListMembers(c.Request.Context(), orgID, claims.UserID)
	if err != nil {
		respondDomainError(c, err, "falha ao listar membros")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *OrganizationHandler) AddMember(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization id inválido"})
		return
	}
	var input appOrg.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.AddMember(c.Request.Context(), orgID, claims.UserID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao adicionar membro")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *OrganizationHandler) UpdateRelationshipStatus(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization id inválido"})
		return
	}
	studentID, err := uuid.Parse(c.Param("student_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student_id inválido"})
		return
	}
	var input appOrg.UpdateRelationshipStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status inválido"})
		return
	}
	input.ClientIP = getClientIP(c)

	if err := h.uc.UpdateRelationshipStatus(c.Request.Context(), orgID, claims.UserID, claims.Role, studentID, input); err != nil {
		respondDomainError(c, err, "falha ao atualizar vínculo")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Vínculo atualizado com sucesso."})
}
