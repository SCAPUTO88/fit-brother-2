package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "x-auth-claims"
)

func RequireAuth(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "header de autorização faltando"})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato de header Authorization inválido"})
			return
		}

		if strings.ToLower(fields[0]) != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "tipo de token não suportado"})
			return
		}

		accessToken := fields[1]

		claims, err := jwtService.Verify(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token de acesso inválido ou expirado"})
			return
		}

		c.Set(authorizationPayloadKey, claims)
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
