package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	perms *services.PermissionService
}

func NewPermissionHandler(perms *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{perms: perms}
}

type PermissionResponse struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"roles.manage"`
	Category    string `json:"category" example:"Roles & Permissions"`
	Description string `json:"description,omitempty" example:"Change which permissions are assigned to a role"`
}

// ListPermissions godoc
// @Summary      List all permissions
// @Description  Returns every permission in the catalog. Requires the roles.view permission.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   PermissionResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	perms, err := h.perms.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load permissions"})
		return
	}

	resp := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		resp = append(resp, PermissionResponse{ID: p.ID, Name: p.Name, Category: p.Category, Description: p.Description})
	}
	c.JSON(http.StatusOK, resp)
}

type RoleResponse struct {
	ID          uint     `json:"id" example:"2"`
	Name        string   `json:"name" example:"manager"`
	Version     int      `json:"version" example:"1"`
	Permissions []string `json:"permissions"`
}

// ListRoles godoc
// @Summary      List manageable roles with their permissions
// @Description  Returns every role except admin (whose access can't be restricted), each with its assigned permission names. Requires the roles.view permission.
// @Tags         permissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   RoleResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/roles [get]
func (h *PermissionHandler) ListRoles(c *gin.Context) {
	roles, err := h.perms.ListManageableRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load roles"})
		return
	}

	resp := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		resp = append(resp, RoleResponse{ID: r.ID, Name: r.Name, Version: r.Version, Permissions: r.Permissions})
	}
	c.JSON(http.StatusOK, resp)
}

type UpdateRolePermissionsRequest struct {
	Version     int      `json:"version" binding:"required"`
	Permissions []string `json:"permissions"`
}

// UpdateRolePermissions godoc
// @Summary      Update a role's assigned permissions
// @Description  Replaces the full set of permissions assigned to a role (admin's role can't be changed here). Requires the roles.manage permission.
// @Tags         permissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                            true  "Role ID"
// @Param        request  body      UpdateRolePermissionsRequest  true  "Permission names to assign"
// @Success      200      {object}  RoleResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Router       /api/v1/roles/{id}/permissions [put]
func (h *PermissionHandler) UpdateRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	var req UpdateRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version and permissions are required"})
		return
	}

	role, err := h.perms.UpdateRolePermissions(uint(roleID), req.Version, req.Permissions)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRoleNotManageable):
			c.JSON(http.StatusForbidden, gin.H{"error": "admin's permissions cannot be changed"})
		case errors.Is(err, services.ErrStaleRole):
			c.JSON(http.StatusConflict, gin.H{"error": "role was changed by someone else, reload and try again"})
		case errors.Is(err, services.ErrUnknownPermission):
			c.JSON(http.StatusBadRequest, gin.H{"error": "one or more permission names are not in the catalog"})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role permissions"})
		}
		return
	}
	c.JSON(http.StatusOK, RoleResponse{ID: role.ID, Name: role.Name, Version: role.Version, Permissions: role.Permissions})
}
