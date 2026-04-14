package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/application/auth"
)

type AuthHandler struct {
	loginUseCase   *auth.LoginUseCase
	refreshUseCase *auth.RefreshUseCase
	logoutUseCase  *auth.LogoutUseCase
}

func NewAuthHandler(
	login *auth.LoginUseCase,
	refresh *auth.RefreshUseCase,
	logout *auth.LogoutUseCase,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:   login,
		refreshUseCase: refresh,
		logoutUseCase:  logout,
	}
}

// RegisterRoutes acopla os use-cases ao roteador Gin
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/auth")
	{
		group.POST("/login", h.Login)
		group.POST("/refresh", h.Refresh)
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

	c.JSON(http.StatusOK, output)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input auth.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	input.ClientIP = getClientIP(c)

	output, err := h.refreshUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input auth.LogoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	err := h.logoutUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Soft logout sempre retorna 200 OK ou 204 No Content se sucesso pra gente.
	c.Status(http.StatusNoContent)
}

// getClientIP obtém o IP limpo do request.
func getClientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		return "127.0.0.1"
	}
	// Limpa portas, caso venha do proxy "IP:Porta"
	if idx := strings.LastIndex(ip, ":"); idx != -1 && strings.Count(ip, ":") == 1 {
		ip = ip[:idx]
	}
	return ip
}
