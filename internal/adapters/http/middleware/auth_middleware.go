package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "x-auth-claims"

	AccessCookieName      = "fitbrother_access_token"
	RefreshCookieName     = "fitbrother_refresh_token"
	CSRFCookieName        = "fitbrother_csrf"
	CSRFHeaderName        = "X-CSRF-Token"
	SessionModeHeaderName = "X-Session-Mode"
)

type UserStatusChecker interface {
	IsUserActive(ctx context.Context, userID uuid.UUID) (bool, error)
}

func RequireAuth(jwtService *jwt.Service, activeChecker UserStatusChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := extractAccessToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		claims, err := jwtService.Verify(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token de acesso inválido ou expirado"})
			return
		}

		if activeChecker != nil {
			active, err := activeChecker.IsUserActive(c.Request.Context(), claims.UserID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "falha ao validar status da sessão"})
				return
			}
			if !active {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "usuário inativo ou sessão revogada"})
				return
			}
		}

		c.Set(authorizationPayloadKey, claims)
		c.Next()
	}
}

func extractAccessToken(c *gin.Context) (string, error) {
	authorizationHeader := c.GetHeader(authorizationHeaderKey)
	if len(authorizationHeader) > 0 {
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			return "", errAuthInvalid("formato de header Authorization inválido")
		}
		if strings.ToLower(fields[0]) != authorizationTypeBearer {
			return "", errAuthInvalid("tipo de token não suportado")
		}
		return fields[1], nil
	}

	cookieToken, err := c.Cookie(AccessCookieName)
	if err == nil && strings.TrimSpace(cookieToken) != "" {
		return strings.TrimSpace(cookieToken), nil
	}

	return "", errAuthInvalid("credencial de autorização faltando")
}

type authExtractError string

func (e authExtractError) Error() string {
	return string(e)
}

func errAuthInvalid(msg string) error {
	return authExtractError(msg)
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, exists := c.Get(authorizationPayloadKey)
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*jwt.Claims)
	return jwtClaims, ok
}
