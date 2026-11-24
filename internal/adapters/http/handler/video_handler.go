package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	appvideo "github.com/SCAPUTO88/fit-brother/internal/application/video"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type VideoHandler struct {
	requestUploadURL *appvideo.RequestUploadURLUseCase
	confirmUpload    *appvideo.ConfirmUploadUseCase
	getSignedURL     *appvideo.GetSignedURLUseCase
}

func NewVideoHandler(
	ru *appvideo.RequestUploadURLUseCase,
	cu *appvideo.ConfirmUploadUseCase,
	gs *appvideo.GetSignedURLUseCase,
) *VideoHandler {
	return &VideoHandler{
		requestUploadURL: ru,
		confirmUpload:    cu,
		getSignedURL:     gs,
	}
}

func (h *VideoHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	g := router.Group("/videos")
	if auther != nil {
		g.Use(auther)
	}

	g.POST("/upload-url", h.RequestUploadURL)
	g.POST("", h.ConfirmUpload)
	g.GET("/exercises/:exercise_id/signed-url", h.GetSignedURL)
}

func (h *VideoHandler) RequestUploadURL(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem fazer upload de vídeos"})
		return
	}

	var input appvideo.RequestUploadURLInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	out, err := h.requestUploadURL.Execute(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao gerar URL de upload")
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *VideoHandler) ConfirmUpload(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem confirmar uploads"})
		return
	}

	var input appvideo.ConfirmUploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	out, err := h.confirmUpload.Execute(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao confirmar upload")
		return
	}

	c.JSON(http.StatusCreated, out)
}

func (h *VideoHandler) GetSignedURL(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("exercise_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exercise_id inválido"})
		return
	}

	out, err := h.getSignedURL.Execute(c.Request.Context(), exerciseID, claims.UserID, claims.Role)
	if err != nil {
		respondDomainError(c, err, "falha ao gerar URL do vídeo")
		return
	}

	c.JSON(http.StatusOK, out)
}
