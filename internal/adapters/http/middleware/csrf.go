package middleware

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func CSRFProtection(allowedOrigins []string) gin.HandlerFunc {
	normalizedAllowed := make([]string, 0, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if norm := normalizeOrigin(origin); norm != "" {
			normalizedAllowed = append(normalizedAllowed, norm)
		}
	}

	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		// Upload assinado via HMAC no local storage não utiliza autenticação por cookie.
		if strings.HasSuffix(c.Request.URL.Path, "/storage/local") {
			c.Next()
			return
		}

		// Requisições com Authorization: Bearer explícito não são vulneráveis a CSRF por cookie ambiente.
		authHeader := strings.TrimSpace(c.GetHeader(authorizationHeaderKey))
		if len(authHeader) >= len(authorizationTypeBearer)+1 &&
			strings.EqualFold(authHeader[:len(authorizationTypeBearer)], authorizationTypeBearer) {
			c.Next()
			return
		}

		hasSessionCookie := hasCookie(c, AccessCookieName) || hasCookie(c, RefreshCookieName)
		if !hasSessionCookie {
			c.Next()
			return
		}

		if origin := strings.TrimSpace(c.GetHeader("Origin")); origin != "" {
			if !isOriginAllowed(origin, c.Request.Host, normalizedAllowed) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "origem da requisição não autorizada"})
				return
			}
		} else if referer := strings.TrimSpace(c.GetHeader("Referer")); referer != "" {
			if !isOriginAllowed(referer, c.Request.Host, normalizedAllowed) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "referência da requisição não autorizada"})
				return
			}
		}

		// Login não usa cookie prévio para autenticação; permite reautenticação se restar cookie expirado.
		if strings.HasSuffix(c.Request.URL.Path, "/auth/login") {
			c.Next()
			return
		}

		csrfCookie, err := c.Cookie(CSRFCookieName)
		csrfHeader := strings.TrimSpace(c.GetHeader(CSRFHeaderName))
		if err != nil || strings.TrimSpace(csrfCookie) == "" || csrfHeader == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "token CSRF ausente"})
			return
		}

		if subtle.ConstantTimeCompare([]byte(csrfCookie), []byte(csrfHeader)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "token CSRF inválido"})
			return
		}

		c.Next()
	}
}

func hasCookie(c *gin.Context, name string) bool {
	val, err := c.Cookie(name)
	return err == nil && strings.TrimSpace(val) != ""
}

func normalizeOrigin(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
}

func isOriginAllowed(rawOrigin, requestHost string, allowed []string) bool {
	norm := normalizeOrigin(rawOrigin)
	if norm == "" {
		return false
	}

	if parsed, err := url.Parse(norm); err == nil && requestHost != "" {
		if strings.EqualFold(parsed.Host, requestHost) {
			return true
		}
	}

	for _, candidate := range allowed {
		if norm == candidate {
			return true
		}
	}
	return false
}
