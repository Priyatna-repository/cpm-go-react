package middleware

import (
	"net/http"
	"strings"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "userID"
	ContextRole   = "role"
)

// RequireAuth validates the Authorization: Bearer <token> header and stores
// the caller's user ID and role in the Gin context for downstream handlers.
func RequireAuth(auth *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed Authorization header"})
			return
		}

		claims, err := auth.ParseAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// RequireRole must run after RequireAuth. It aborts with 403 unless the
// caller's role is one of the allowed roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		value, exists := c.Get(ContextRole)
		role, ok := value.(string)
		if !exists || !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication context"})
			return
		}
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}
		c.Next()
	}
}
