package handlers

import (
	"errors"
	"net/http"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskDependencyHandler struct {
	deps *services.TaskDependencyService
}

func NewTaskDependencyHandler(deps *services.TaskDependencyService) *TaskDependencyHandler {
	return &TaskDependencyHandler{deps: deps}
}

type TaskDependencyResponse struct {
	ID              uint   `json:"id"`
	TaskID          uint   `json:"task_id"`
	DependsOnTaskID uint   `json:"depends_on_task_id"`
	RelationType    string `json:"relation_type"`
	LagDays         int    `json:"lag_days"`
}

func toTaskDependencyResponse(d *models.TaskDependency) TaskDependencyResponse {
	return TaskDependencyResponse{
		ID: d.ID, TaskID: d.TaskID, DependsOnTaskID: d.DependsOnTaskID,
		RelationType: string(d.RelationType), LagDays: d.LagDays,
	}
}

// ListTaskDependencies godoc
// @Summary      List every dependency edge among a project's tasks
// @Description  Not paginated — the Gantt needs the full set at once to draw arrows. Requires task.view + project access.
// @Tags         task-dependencies
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Project ID"
// @Success      200  {array}  TaskDependencyResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/projects/{id}/dependencies [get]
func (h *TaskDependencyHandler) List(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	items, err := h.deps.ListForProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dependencies"})
		return
	}
	resp := make([]TaskDependencyResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toTaskDependencyResponse(&items[i]))
	}
	c.JSON(http.StatusOK, resp)
}

type TaskDependencyRequest struct {
	TaskID          uint   `json:"task_id" binding:"required"`
	DependsOnTaskID uint   `json:"depends_on_task_id" binding:"required"`
	RelationType    string `json:"relation_type" binding:"required"`
	LagDays         int    `json:"lag_days"`
}

// CreateTaskDependency godoc
// @Summary      Create a dependency between two tasks
// @Description  Requires task.manage_dependencies + project access. Rejects self-dependencies and cycles.
// @Tags         task-dependencies
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Project ID"
// @Param        request  body      TaskDependencyRequest  true  "Dependency"
// @Success      201  {object}  TaskDependencyResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /api/v1/projects/{id}/dependencies [post]
func (h *TaskDependencyHandler) Create(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	var req TaskDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id, depends_on_task_id, and relation_type are required"})
		return
	}
	dep, err := h.deps.Create(projectID, services.CreateDependencyInput{
		TaskID: req.TaskID, DependsOnTaskID: req.DependsOnTaskID,
		RelationType: models.TaskRelationType(req.RelationType), LagDays: req.LagDays,
	})
	if err != nil {
		respondDependencyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTaskDependencyResponse(dep))
}

// DeleteTaskDependency godoc
// @Summary      Remove a dependency
// @Tags         task-dependencies
// @Security     BearerAuth
// @Param        id     path  int  true  "Project ID"
// @Param        depId  path  int  true  "Dependency ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/dependencies/{depId} [delete]
func (h *TaskDependencyHandler) Delete(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	depID, err := parseUintParam(c, "depId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dependency id"})
		return
	}
	if err := h.deps.Delete(projectID, depID); err != nil {
		respondDependencyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondDependencyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrSelfDependency):
		c.JSON(http.StatusBadRequest, gin.H{"error": "a task cannot depend on itself"})
	case errors.Is(err, services.ErrInvalidRelationType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "relation_type must be one of SS, FS, FF, SF"})
	case errors.Is(err, services.ErrTaskNotInProject):
		c.JSON(http.StatusBadRequest, gin.H{"error": "task does not belong to this project"})
	case errors.Is(err, services.ErrDependencyCycle):
		c.JSON(http.StatusConflict, gin.H{"error": "this dependency would create a circular chain"})
	case errors.Is(err, services.ErrDuplicateDependency):
		c.JSON(http.StatusConflict, gin.H{"error": "this dependency already exists"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "dependency not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save dependency"})
	}
}
