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

// RequireAuth cria um middleware que intercepta requisições HTTP,
// garante que um token Bearer foi enviado, valida a assinatura RS256 desse JWT
// e carrega as "Claims" no gin.Context.
func RequireAuth(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "header default faltando"})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato de header Authorization inválido"})
			return
		}

		if strings.ToLower(fields[0]) != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token não suportado ou faltando bearer"})
			return
		}

		accessToken := fields[1]

		// Valida via Public Key o token!
		claims, err := jwtService.Verify(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token de acesso inválido ou expirado"})
			return
		}

		// Injera na requisição local os dadoss
		c.Set(authorizationPayloadKey, claims)
		c.Next()
	}
}

// GetClaims extrai as Claims de forma segura depois de ter passado do RequireAuth.
func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, exists := c.Get(authorizationPayloadKey)
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*jwt.Claims)
	return jwtClaims, ok
}
