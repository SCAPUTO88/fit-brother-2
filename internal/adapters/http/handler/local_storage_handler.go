package handler

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/storage"
	appvideo "github.com/SCAPUTO88/fit-brother/internal/application/video"
	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
)

type LocalStorageHandler struct {
	store *storage.LocalStorage
}

func NewLocalStorageHandler(store *storage.LocalStorage) *LocalStorageHandler {
	return &LocalStorageHandler{store: store}
}

func (h *LocalStorageHandler) RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/storage")
	{
		g.PUT("/local", h.UploadObject)
		g.GET("/local", h.DownloadObject)
	}
}

func (h *LocalStorageHandler) UploadObject(c *gin.Context) {
	key := c.Query("key")
	op := c.Query("op")
	signedContentType := c.Query("content_type")
	expires := c.Query("expires")
	sig := c.Query("sig")

	if op != "put" {
		c.JSON(http.StatusForbidden, gin.H{"error": "operação não autorizada"})
		return
	}

	if err := h.store.VerifySignedRequest("put", key, signedContentType, expires, sig); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "URL de upload inválida ou expirada"})
		return
	}

	headerMIME := normalizeContentType(c.GetHeader("Content-Type"))
	if signedContentType != "" && headerMIME != "" && headerMIME != normalizeContentType(signedContentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type não corresponde ao autorizado"})
		return
	}

	effectiveMIME := signedContentType
	if effectiveMIME == "" {
		effectiveMIME = headerMIME
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, appvideo.MaxVideoSizeBytes+1)
	if _, err := h.store.SaveObject(key, effectiveMIME, c.Request.Body, appvideo.MaxVideoSizeBytes); err != nil {
		if errors.Is(err, storage.ErrFileTooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "arquivo excede o limite de 50MB"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "falha ao armazenar arquivo"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *LocalStorageHandler) DownloadObject(c *gin.Context) {
	key := c.Query("key")
	op := c.Query("op")
	expires := c.Query("expires")
	sig := c.Query("sig")

	if op != "get" {
		c.JSON(http.StatusForbidden, gin.H{"error": "operação não autorizada"})
		return
	}

	if err := h.store.VerifySignedRequest("get", key, "", expires, sig); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "URL assinada inválida ou expirada"})
		return
	}

	file, meta, info, err := h.store.OpenObject(key)
	if err != nil {
		if errors.Is(err, domainStorage.ErrObjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "vídeo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao ler vídeo"})
		return
	}
	defer file.Close()

	if meta != nil && meta.MimeType != "" {
		c.Header("Content-Type", meta.MimeType)
	}
	c.Header("Cache-Control", "private, no-store")
	http.ServeContent(c.Writer, c.Request, filepath.Base(key), info.ModTime(), file)
}

func normalizeContentType(raw string) string {
	val := strings.ToLower(strings.TrimSpace(raw))
	if idx := strings.Index(val, ";"); idx != -1 {
		val = strings.TrimSpace(val[:idx])
	}
	return val
}
