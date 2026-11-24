package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
)

func respondDomainError(c *gin.Context, err error, internalMsg string) {
	switch {
	case errors.Is(err, errs.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, errs.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, errs.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, errs.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, errs.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		if internalMsg == "" {
			internalMsg = "erro interno no servidor"
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": internalMsg})
	}
}
