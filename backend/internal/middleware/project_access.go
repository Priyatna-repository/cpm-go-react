package middleware

import (
	"net/http"
	"strconv"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// RequireProjectAccess must run after RequireAuth (and typically after
// RequirePermission — the exact composition CLAUDE.md/#3 called for:
// RequireAuth → RequirePermission("project.edit") → RequireProjectAccess).
// Admin always passes; everyone else needs company/direct-client ownership
// or an explicit project_user_access grant for the :id in the route.
func RequireProjectAccess(projects *services.ProjectService) gin.HandlerFunc {
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

		userIDVal, exists := c.Get(ContextUserID)
		userID, okType := userIDVal.(uint)
		if !exists || !okType {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication context"})
			return
		}

		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
			return
		}

		allowed, err := projects.HasAccess(userID, uint(projectID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "access check failed"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no access to this project"})
			return
		}
		c.Next()
	}
}
