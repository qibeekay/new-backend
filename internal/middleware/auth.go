package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"portfolio-api/internal/auth"
	"portfolio-api/internal/response"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok {
			response.Error(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}
		if err := auth.ValidateToken(secret, raw); err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}
		c.Next()
	}
}
