package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler is instantiated twice in routes.go — once per UserRoleScope —
// so the same CRUD logic serves both the internal Users page and the
// Client Users page, matching the reference app's two-controller split
// without duplicating handler code.
type UserHandler struct {
	users *services.UserService
	scope services.UserRoleScope
}

func NewUserHandler(users *services.UserService, scope services.UserRoleScope) *UserHandler {
	return &UserHandler{users: users, scope: scope}
}

type UserResponses struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Avatar     *string    `json:"avatar,omitempty"`
	JobTitle   *string    `json:"job_title,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	Address    *string    `json:"address,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

func toUserResponse(m *models.User) UserResponses {
	return UserResponses{
		ID: m.ID, Name: m.Name, Email: m.Email, Role: m.Role.Name, Avatar: m.Avatar,
		JobTitle: m.JobTitle, Phone: m.Phone, Address: m.Address, ArchivedAt: m.ArchivedAt,
	}
}

type UserListResponses struct {
	Data []UserResponses `json:"data"`
	Meta struct {
		CurrentPage int   `json:"current_page"`
		LastPage    int   `json:"last_page"`
		Total       int64 `json:"total"`
	} `json:"meta"`
}

// ListUsers godoc
// @Summary      List users
// @Description  Paginated, searchable. Scope (internal staff vs client) is fixed per route. Requires the matching view permission.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        search    query     string  false  "Search by name or email"
// @Param        archived  query     bool    false  "Show archived instead of active"
// @Success      200  {object}  UserListResponses
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/users [get]
// @Router       /api/v1/client-user-accounts [get]
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	archived := c.Query("archived") == "1" || c.Query("archived") == "true"

	result, err := h.users.List(services.ListUsersParams{
		Page: page, PerPage: 12, Search: c.Query("search"), Archived: archived, Scope: h.scope,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}

	resp := UserListResponses{Data: make([]UserResponses, 0, len(result.Items))}
	for i := range result.Items {
		resp.Data = append(resp.Data, toUserResponse(&result.Items[i]))
	}
	resp.Meta.CurrentPage = result.CurrentPage
	resp.Meta.LastPage = result.LastPage
	resp.Meta.Total = result.Total
	c.JSON(http.StatusOK, resp)
}

// GetUser godoc
// @Summary      Get a user
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponses
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/users/{id} [get]
// @Router       /api/v1/client-user-accounts/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user, err := h.users.Get(h.scope, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

type UserRequest struct {
	Name     string  `form:"name" binding:"required"`
	Email    string  `form:"email" binding:"required,email"`
	Password string  `form:"password"`
	Role     string  `form:"role" binding:"required"`
	JobTitle *string `form:"job_title"`
	Phone    *string `form:"phone"`
	Address  *string `form:"address"`
}

// CreateUser godoc
// @Summary      Create a user
// @Description  multipart/form-data body (avatar file is optional). Password is required on create.
// @Tags         users
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Success      201  {object}  UserResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/users [post]
// @Router       /api/v1/client-user-accounts [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBind(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, role, and password are required"})
		return
	}
	avatar, _ := c.FormFile("avatar")

	user, err := h.users.Create(h.scope, services.UserInput{
		Name: req.Name, Email: req.Email, Password: req.Password, RoleName: req.Role,
		JobTitle: req.JobTitle, Phone: req.Phone, Address: req.Address,
	}, avatar)
	if err != nil {
		respondUserError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(user))
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  multipart/form-data body. Leave password blank to keep the current one.
// @Tags         users
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/users/{id} [put]
// @Router       /api/v1/client-user-accounts/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req UserRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, and role are required"})
		return
	}
	avatar, _ := c.FormFile("avatar")

	user, err := h.users.Update(h.scope, id, services.UserInput{
		Name: req.Name, Email: req.Email, Password: req.Password, RoleName: req.Role,
		JobTitle: req.JobTitle, Phone: req.Phone, Address: req.Address,
	}, avatar)
	if err != nil {
		respondUserError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

// ArchiveUser godoc
// @Summary      Archive a user
// @Description  Prevents the archived user from logging in. Cannot archive yourself.
// @Tags         users
// @Security     BearerAuth
// @Param        id   path  int  true  "User ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/users/{id} [delete]
// @Router       /api/v1/client-user-accounts/{id} [delete]
func (h *UserHandler) Archive(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	callerID, _ := callerFromContext(c)
	if err := h.users.Archive(h.scope, id, callerID); err != nil {
		respondUserError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RestoreUser godoc
// @Summary      Restore an archived user
// @Tags         users
// @Security     BearerAuth
// @Param        id   path  int  true  "User ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/users/{id}/restore [post]
// @Router       /api/v1/client-user-accounts/{id}/restore [post]
func (h *UserHandler) Restore(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.users.Restore(h.scope, id); err != nil {
		respondUserError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ForceDeleteUser godoc
// @Summary      Permanently delete a user
// @Tags         users
// @Security     BearerAuth
// @Param        id   path  int  true  "User ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/users/{id}/force [delete]
// @Router       /api/v1/client-user-accounts/{id}/force [delete]
func (h *UserHandler) ForceDelete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	callerID, _ := callerFromContext(c)
	if err := h.users.ForceDelete(h.scope, id, callerID); err != nil {
		respondUserError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidRole):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role for this user type"})
	case errors.Is(err, services.ErrDuplicateEmail):
		c.JSON(http.StatusConflict, gin.H{"error": "a user with this email already exists"})
	case errors.Is(err, services.ErrCannotSelfModify):
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot archive or delete your own account"})
	case errors.Is(err, services.ErrFileTooLarge):
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is too large (max 2MB)"})
	case errors.Is(err, services.ErrUnsupportedImage):
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type (use jpg, png, or gif)"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
	}
}
