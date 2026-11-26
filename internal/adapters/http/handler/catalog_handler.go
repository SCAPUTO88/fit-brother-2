package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	appCatalog "github.com/SCAPUTO88/fit-brother/internal/application/catalog"
	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type CatalogHandler struct {
	uc *appCatalog.UseCase
}

func NewCatalogHandler(uc *appCatalog.UseCase) *CatalogHandler {
	return &CatalogHandler{uc: uc}
}

func (h *CatalogHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	g := router.Group("/catalog")
	if auther != nil {
		g.Use(auther)
	}

	g.GET("/filters", h.ListFilters)
	g.GET("/exercises", h.Search)
	g.GET("/exercises/:id", h.GetExercise)
	g.POST("/exercises", h.CreateCustom)
	g.PATCH("/exercises/:id", h.UpdateCustom)
	g.DELETE("/exercises/:id", h.DeleteCustom)
	g.POST("/exercises/:id/media/upload-url", h.RequestMediaUploadURL)
	g.POST("/exercises/:id/media/confirm", h.ConfirmMediaUpload)
	g.GET("/exercises/:id/media/:media_id/signed-url", h.GetMediaSignedURL)
}

func (h *CatalogHandler) ListFilters(c *gin.Context) {
	if _, ok := middleware.GetClaims(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	out, err := h.uc.ListFilters(c.Request.Context())
	if err != nil {
		respondDomainError(c, err, "falha ao listar filtros do catálogo")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CatalogHandler) Search(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	filter := domainCatalog.ExerciseFilter{
		Query:     strings.TrimSpace(c.Query("q")),
		Muscle:    strings.TrimSpace(c.Query("muscle")),
		Equipment: strings.TrimSpace(c.Query("equipment")),
		Category:  strings.TrimSpace(c.Query("category")),
		Scope:     strings.TrimSpace(c.Query("scope")),
		Limit:     parseQueryInt(c.Query("limit"), 30),
		Offset:    parseQueryInt(c.Query("offset"), 0),
	}

	out, err := h.uc.SearchExercises(c.Request.Context(), claims.UserID, claims.Role, filter)
	if err != nil {
		respondDomainError(c, err, "falha ao consultar biblioteca de exercícios")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CatalogHandler) GetExercise(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de exercício inválido"})
		return
	}

	out, err := h.uc.GetExercise(c.Request.Context(), claims.UserID, claims.Role, exerciseID)
	if err != nil {
		respondDomainError(c, err, "falha ao consultar exercício")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CatalogHandler) CreateCustom(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem criar exercícios customizados"})
		return
	}

	var input appCatalog.CreateExerciseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do exercício inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.CreateCustomExercise(c.Request.Context(), claims.UserID, claims.Role, input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar exercício customizado")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *CatalogHandler) UpdateCustom(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem editar exercícios customizados"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de exercício inválido"})
		return
	}

	var input appCatalog.UpdateExerciseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados do exercício inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.UpdateCustomExercise(c.Request.Context(), claims.UserID, claims.Role, exerciseID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao atualizar exercício customizado")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CatalogHandler) DeleteCustom(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem remover exercícios customizados"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de exercício inválido"})
		return
	}

	if err := h.uc.DeleteCustomExercise(c.Request.Context(), claims.UserID, claims.Role, exerciseID, getClientIP(c)); err != nil {
		respondDomainError(c, err, "falha ao remover exercício customizado")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Exercício customizado removido com sucesso."})
}

func (h *CatalogHandler) RequestMediaUploadURL(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem enviar mídias"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de exercício inválido"})
		return
	}

	var input appCatalog.RequestMediaUploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados de upload inválidos"})
		return
	}

	out, err := h.uc.RequestMediaUploadURL(c.Request.Context(), claims.UserID, claims.Role, exerciseID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao gerar URL de upload de mídia")
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CatalogHandler) ConfirmMediaUpload(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem confirmar mídias"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de exercício inválido"})
		return
	}

	var input appCatalog.ConfirmMediaUploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados de confirmação inválidos"})
		return
	}
	input.ClientIP = getClientIP(c)

	out, err := h.uc.ConfirmMediaUpload(c.Request.Context(), claims.UserID, claims.Role, exerciseID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao confirmar upload de mídia")
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *CatalogHandler) GetMediaSignedURL(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id de exercício inválido"})
		return
	}

	mediaID, err := uuid.Parse(c.Param("media_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "media_id inválido"})
		return
	}

	out, err := h.uc.GetMediaSignedURL(c.Request.Context(), claims.UserID, claims.Role, exerciseID, mediaID)
	if err != nil {
		respondDomainError(c, err, "falha ao gerar URL assinada da mídia")
		return
	}
	c.JSON(http.StatusOK, out)
}

func parseQueryInt(raw string, fallback int) int {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback
	}
	v, err := strconv.Atoi(trimmed)
	if err != nil {
		return fallback
	}
	return v
}
