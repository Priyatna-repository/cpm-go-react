package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/middleware"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	projects *services.ProjectService
	perms    *services.PermissionService
}

func NewProjectHandler(projects *services.ProjectService, perms *services.PermissionService) *ProjectHandler {
	return &ProjectHandler{projects: projects, perms: perms}
}

func (h *ProjectHandler) callerCanManageAccess(role string) (bool, error) {
	if models.IsAdmin(role) {
		return true, nil
	}
	return h.perms.HasPermission(role, "project.manage_access")
}

type ProjectCompanyRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ProjectUserRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ProjectLabelRef struct {
	ID    uint    `json:"id"`
	Name  string  `json:"name"`
	Slug  string  `json:"slug"`
	Color *string `json:"color,omitempty"`
}

type ProjectResponse struct {
	ID             uint               `json:"id"`
	Code           string             `json:"code"`
	Name           string             `json:"name"`
	Description    *string            `json:"description,omitempty"`
	ClientCompany  *ProjectCompanyRef `json:"client_company,omitempty"`
	ClientUser     *ProjectUserRef    `json:"client_user,omitempty"`
	StartDate      *string            `json:"start_date,omitempty"`
	EndDate        *string            `json:"end_date,omitempty"`
	BudgetEstimate *float64           `json:"budget_estimate,omitempty"`
	TypeLabel      *ProjectLabelRef   `json:"type_label,omitempty"`
	StatusLabels   []ProjectLabelRef  `json:"status_labels"`
	IsCompleted    bool               `json:"is_completed"`
	CompletedAt    *time.Time         `json:"completed_at,omitempty"`
	ArchivedAt     *time.Time         `json:"archived_at,omitempty"`
	Users          []ProjectUserRef   `json:"users"`
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func toProjectResponse(m *models.Project) ProjectResponse {
	resp := ProjectResponse{
		ID: m.ID, Code: m.Code, Name: m.Name, Description: m.Description,
		StartDate: formatDate(m.StartDate), EndDate: formatDate(m.EndDate),
		BudgetEstimate: m.BudgetEstimate, IsCompleted: m.IsCompleted,
		CompletedAt: m.CompletedAt, ArchivedAt: m.ArchivedAt,
		StatusLabels: make([]ProjectLabelRef, 0, len(m.StatusLabels)),
		Users:        make([]ProjectUserRef, 0, len(m.Users)),
	}
	if m.ClientCompany != nil {
		resp.ClientCompany = &ProjectCompanyRef{ID: m.ClientCompany.ID, Name: m.ClientCompany.Name}
	}
	if m.ClientUser != nil {
		resp.ClientUser = &ProjectUserRef{ID: m.ClientUser.ID, Name: m.ClientUser.Name}
	}
	if m.TypeLabel != nil {
		resp.TypeLabel = &ProjectLabelRef{ID: m.TypeLabel.ID, Name: m.TypeLabel.Name, Slug: m.TypeLabel.Slug, Color: m.TypeLabel.Color}
	}
	for _, l := range m.StatusLabels {
		resp.StatusLabels = append(resp.StatusLabels, ProjectLabelRef{ID: l.ID, Name: l.Name, Slug: l.Slug, Color: l.Color})
	}
	for _, u := range m.Users {
		resp.Users = append(resp.Users, ProjectUserRef{ID: u.ID, Name: u.Name})
	}
	return resp
}

type ProjectListResponse struct {
	Data []ProjectResponse `json:"data"`
	Meta struct {
		CurrentPage int   `json:"current_page"`
		LastPage    int   `json:"last_page"`
		Total       int64 `json:"total"`
	} `json:"meta"`
}

func callerFromContext(c *gin.Context) (uint, string) {
	userIDVal, _ := c.Get(middleware.ContextUserID)
	roleVal, _ := c.Get(middleware.ContextRole)
	userID, _ := userIDVal.(uint)
	role, _ := roleVal.(string)
	return userID, role
}

// ListProjects godoc
// @Summary      List projects
// @Description  Paginated, searchable. Non-admins only see projects they have access to. Requires project.view.
// @Tags         projects
// @Security     BearerAuth
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        search    query     string  false  "Search by name or code"
// @Param        archived  query     bool    false  "Show archived instead of active"
// @Success      200  {object}  ProjectListResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	archived := c.Query("archived") == "1" || c.Query("archived") == "true"
	userID, role := callerFromContext(c)

	result, err := h.projects.List(services.ListProjectsParams{
		Page: page, PerPage: 12, Search: c.Query("search"), Archived: archived,
		UserID: userID, IsAdmin: models.IsAdmin(role),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load projects"})
		return
	}

	resp := ProjectListResponse{Data: make([]ProjectResponse, 0, len(result.Items))}
	for i := range result.Items {
		resp.Data = append(resp.Data, toProjectResponse(&result.Items[i]))
	}
	resp.Meta.CurrentPage = result.CurrentPage
	resp.Meta.LastPage = result.LastPage
	resp.Meta.Total = result.Total
	c.JSON(http.StatusOK, resp)
}

// GetProject godoc
// @Summary      Get a project
// @Tags         projects
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Project ID"
// @Success      200  {object}  ProjectResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id} [get]
func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	project, err := h.projects.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	c.JSON(http.StatusOK, toProjectResponse(project))
}

type ProjectRequest struct {
	Name            string   `json:"name" binding:"required"`
	Description     *string  `json:"description"`
	ClientCompanyID *uint    `json:"client_company_id"`
	ClientUserID    *uint    `json:"client_user_id"`
	StartDate       *string  `json:"start_date"`
	EndDate         *string  `json:"end_date"`
	BudgetEstimate  *float64 `json:"budget_estimate"`
	TypeLabelID     *uint    `json:"type_label_id"`
	StatusLabelIDs  []uint   `json:"status_label_ids"`
	UserIDs         []uint   `json:"user_ids"`
}

func (req ProjectRequest) toInput() (services.ProjectInput, error) {
	start, err := parseOptionalDate(req.StartDate)
	if err != nil {
		return services.ProjectInput{}, err
	}
	end, err := parseOptionalDate(req.EndDate)
	if err != nil {
		return services.ProjectInput{}, err
	}
	return services.ProjectInput{
		Name: req.Name, Description: req.Description,
		ClientCompanyID: req.ClientCompanyID, ClientUserID: req.ClientUserID,
		StartDate: start, EndDate: end, BudgetEstimate: req.BudgetEstimate,
		TypeLabelID: req.TypeLabelID, StatusLabelIDs: req.StatusLabelIDs, UserIDs: req.UserIDs,
	}, nil
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateProject godoc
// @Summary      Create a project
// @Description  Requires project.create.
// @Tags         projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      ProjectRequest  true  "Project"
// @Success      201  {object}  ProjectResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/projects [post]
func (h *ProjectHandler) Create(c *gin.Context) {
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	input, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date/end_date must be YYYY-MM-DD"})
		return
	}

	callerID, _ := callerFromContext(c)
	project, err := h.projects.Create(input, callerID)
	if err != nil {
		respondProjectError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toProjectResponse(project))
}

// UpdateProject godoc
// @Summary      Update a project
// @Description  Requires project.edit + project access.
// @Tags         projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "Project ID"
// @Param        request  body      ProjectRequest  true  "Project"
// @Success      200  {object}  ProjectResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id} [put]
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	input, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date/end_date must be YYYY-MM-DD"})
		return
	}

	_, role := callerFromContext(c)
	canManageAccess, err := h.callerCanManageAccess(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
		return
	}

	project, err := h.projects.Update(id, input, canManageAccess)
	if err != nil {
		respondProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProjectResponse(project))
}

// ArchiveProject godoc
// @Summary      Archive a project
// @Tags         projects
// @Security     BearerAuth
// @Param        id   path  int  true  "Project ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id} [delete]
func (h *ProjectHandler) Archive(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.projects.Archive(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive project"})
		return
	}
	c.Status(http.StatusNoContent)
}

// RestoreProject godoc
// @Summary      Restore an archived project
// @Tags         projects
// @Security     BearerAuth
// @Param        id   path  int  true  "Project ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/restore [post]
func (h *ProjectHandler) Restore(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.projects.Restore(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore project"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ForceDeleteProject godoc
// @Summary      Permanently delete a project
// @Tags         projects
// @Security     BearerAuth
// @Param        id   path  int  true  "Project ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/force [delete]
func (h *ProjectHandler) ForceDelete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.projects.ForceDelete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}
	c.Status(http.StatusNoContent)
}

type ProjectAccessRequest struct {
	UserIDs []uint `json:"user_ids"`
}

// UpdateProjectAccess godoc
// @Summary      Manage project access
// @Description  Replaces the full list of users explicitly granted access to this project. Requires project.manage_access + project access.
// @Tags         projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                   true  "Project ID"
// @Param        request  body      ProjectAccessRequest  true  "User IDs"
// @Success      200  {object}  ProjectResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/access [put]
func (h *ProjectHandler) UpdateAccess(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req ProjectAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_ids must be an array"})
		return
	}
	project, err := h.projects.UpdateAccess(id, req.UserIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project access"})
		return
	}
	c.JSON(http.StatusOK, toProjectResponse(project))
}

func respondProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrProjectClientRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrDuplicateProjectName):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
	case errors.Is(err, services.ErrManageAccessRequired):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save project"})
	}
}
