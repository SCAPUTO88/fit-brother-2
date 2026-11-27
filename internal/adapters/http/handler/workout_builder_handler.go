package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	appworkout "github.com/SCAPUTO88/fit-brother/internal/application/workout"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type WorkoutBuilderHandler struct {
	uc *appworkout.BuilderUseCase
}

func NewWorkoutBuilderHandler(uc *appworkout.BuilderUseCase) *WorkoutBuilderHandler {
	return &WorkoutBuilderHandler{uc: uc}
}

func (h *WorkoutBuilderHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	templates := router.Group("/workout-templates")
	if auther != nil {
		templates.Use(auther)
	}
	templates.GET("", h.ListTemplates)
	templates.POST("", h.CreateTemplate)
	templates.GET("/:id", h.GetTemplate)
	templates.PATCH("/:id", h.UpdateTemplate)
	templates.POST("/:id/clone", h.CloneTemplate)
	templates.POST("/:id/publish", h.PublishVersion)
	templates.GET("/:id/versions", h.ListVersions)
	templates.GET("/:id/versions/:version_id", h.GetVersion)
	templates.POST("/:id/assign", h.AssignTemplate)

	assignments := router.Group("/workout-assignments")
	if auther != nil {
		assignments.Use(auther)
	}
	assignments.GET("", h.ListAssignments)
	assignments.GET("/:id", h.GetAssignment)
	assignments.PATCH("/:id/status", h.UpdateAssignmentStatus)
}

func (h *WorkoutBuilderHandler) ListTemplates(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem listar templates de treino"})
		return
	}

	includeArchived := strings.EqualFold(c.Query("include_archived"), "true")
	out, err := h.uc.ListTemplates(c.Request.Context(), claims.UserID, claims.Role, includeArchived)
	if err != nil {
		respondDomainError(c, err, "falha ao listar templates de treino")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WorkoutBuilderHandler) CreateTemplate(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem criar templates de treino"})
		return
	}

	var input appworkout.CreateTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do template inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.CreateTemplate(c.Request.Context(), claims.UserID, claims.Role, input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar template de treino")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *WorkoutBuilderHandler) GetTemplate(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem visualizar templates de treino"})
		return
	}

	tplID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do template inválido"})
		return
	}

	out, err := h.uc.GetTemplate(c.Request.Context(), claims.UserID, claims.Role, tplID)
	if err != nil {
		respondDomainError(c, err, "falha ao carregar template de treino")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WorkoutBuilderHandler) UpdateTemplate(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem editar templates de treino"})
		return
	}

	tplID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do template inválido"})
		return
	}

	var input appworkout.UpdateTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do template inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.UpdateTemplate(c.Request.Context(), claims.UserID, claims.Role, tplID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao atualizar template de treino")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WorkoutBuilderHandler) CloneTemplate(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem clonar templates de treino"})
		return
	}

	tplID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do template inválido"})
		return
	}

	var input appworkout.CloneTemplateInput
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dados de clonagem inválidos"})
			return
		}
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.CloneTemplate(c.Request.Context(), claims.UserID, claims.Role, tplID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao clonar template de treino")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *WorkoutBuilderHandler) PublishVersion(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem publicar versões de treino"})
		return
	}

	tplID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do template inválido"})
		return
	}

	out, err := h.uc.PublishVersion(c.Request.Context(), claims.UserID, claims.Role, tplID, getClientIP(c))
	if err != nil {
		respondDomainError(c, err, "falha ao publicar versão do treino")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *WorkoutBuilderHandler) ListVersions(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem consultar histórico de versões"})
		return
	}

	tplID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do template inválido"})
		return
	}

	out, err := h.uc.ListVersions(c.Request.Context(), claims.UserID, claims.Role, tplID)
	if err != nil {
		respondDomainError(c, err, "falha ao listar versões do template")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WorkoutBuilderHandler) GetVersion(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem consultar versões"})
		return
	}

	tplID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do template inválido"})
		return
	}
	verID, err := uuid.Parse(c.Param("version_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da versão inválido"})
		return
	}

	out, err := h.uc.GetVersion(c.Request.Context(), claims.UserID, claims.Role, tplID, verID)
	if err != nil {
		respondDomainError(c, err, "falha ao consultar versão do template")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WorkoutBuilderHandler) AssignTemplate(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem atribuir treinos"})
		return
	}

	tplID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do template inválido"})
		return
	}

	var input appworkout.AssignTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados de atribuição inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.AssignTemplate(c.Request.Context(), claims.UserID, claims.Role, tplID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao atribuir treino")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *WorkoutBuilderHandler) ListAssignments(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	out, err := h.uc.ListAssignments(c.Request.Context(), claims.UserID, claims.Role, c.Query("student_id"))
	if err != nil {
		respondDomainError(c, err, "falha ao listar treinos atribuídos")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WorkoutBuilderHandler) GetAssignment(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da atribuição inválido"})
		return
	}

	out, err := h.uc.GetAssignment(c.Request.Context(), claims.UserID, claims.Role, assignmentID)
	if err != nil {
		respondDomainError(c, err, "falha ao carregar detalhes do treino")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *WorkoutBuilderHandler) UpdateAssignmentStatus(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem alterar status de atribuições"})
		return
	}

	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da atribuição inválido"})
		return
	}

	var input appworkout.UpdateAssignmentStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status inválido"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.UpdateAssignmentStatus(c.Request.Context(), claims.UserID, claims.Role, assignmentID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao atualizar status da atribuição")
		return
	}
	c.JSON(http.StatusOK, out)
}
