package middleware

import (
	"net/http"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// roleFromContext extracts the authenticated caller's role, set by
// RequireAuth. Shared by every permission/access middleware in this
// package (RequireProjectAccess in Phase 2 will use it too).
func roleFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextRole)
	role, ok := value.(string)
	return role, exists && ok
}

// RequirePermission must run after RequireAuth. Admin always passes,
// regardless of what's assigned in role_permissions — see the "Permission
// management" decision in CLAUDE.md: this is a deliberate bypass, fixing a
// real bug in the reference Laravel app where admin could get 403'd on a
// permission nobody remembered to seed.
func RequirePermission(perms *services.PermissionService, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := roleFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication context"})
			return
		}

		if models.IsAdmin(role) {
			c.Next()
			return
		}

		allowed, err := perms.HasPermission(role, permission)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing required permission"})
			return
		}
		c.Next()
	}
}
