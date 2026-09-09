package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct {
	tasks      *services.TaskService
	scheduling *services.TaskSchedulingService
}

func NewTaskHandler(tasks *services.TaskService, scheduling *services.TaskSchedulingService) *TaskHandler {
	return &TaskHandler{tasks: tasks, scheduling: scheduling}
}

type TaskResponse struct {
	ID             uint             `json:"id"`
	ProjectID      uint             `json:"project_id"`
	Type           string           `json:"type"`
	ParentTaskID   *uint            `json:"parent_task_id,omitempty"`
	Name           string           `json:"name"`
	Description    *string          `json:"description,omitempty"`
	Number         int              `json:"number"`
	OrderColumn    int              `json:"order_column"`
	AssignedToUser *ProjectUserRef  `json:"assigned_to_user,omitempty"`
	StartDate      *string          `json:"start_date,omitempty"`
	EndDate        *string          `json:"end_date,omitempty"`
	BudgetPlan     *float64         `json:"budget_plan,omitempty"`
	BudgetActual   *float64         `json:"budget_actual,omitempty"`
	Weight         float64          `json:"weight"`
	Progress       float64          `json:"progress"`
	StatusLabel    *ProjectLabelRef `json:"status_label,omitempty"`
	CreatedBy      *ProjectUserRef  `json:"created_by,omitempty"`
	CreatedAt      string           `json:"created_at"`
	IsCompleted    bool             `json:"is_completed"`
	CompletedAt    *string          `json:"completed_at,omitempty"`
	ArchivedAt     *string          `json:"archived_at,omitempty"`
}

// toTaskResponse shapes a Task for JSON. Dates come back as plain
// YYYY-MM-DD strings (matching how the frontend already sends them for
// Project) rather than full RFC3339 timestamps — a Task's start/end date
// has no time-of-day component, so keeping the wire format date-only
// avoids a timezone-conversion footgun on the frontend.
func toTaskResponse(m *models.Task) TaskResponse {
	resp := TaskResponse{
		ID: m.ID, ProjectID: m.ProjectID, Type: string(m.Type), ParentTaskID: m.ParentTaskID,
		Name: m.Name, Description: m.Description,
		Number: m.Number, OrderColumn: m.OrderColumn, BudgetPlan: m.BudgetPlan, BudgetActual: m.BudgetActual,
		Weight: m.Weight, Progress: m.Progress, IsCompleted: m.IsCompleted,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if m.AssignedToUser != nil {
		resp.AssignedToUser = &ProjectUserRef{ID: m.AssignedToUser.ID, Name: m.AssignedToUser.Name}
	}
	if m.StatusLabel != nil {
		resp.StatusLabel = &ProjectLabelRef{ID: m.StatusLabel.ID, Name: m.StatusLabel.Name, Slug: m.StatusLabel.Slug, Color: m.StatusLabel.Color}
	}
	if m.CreatedByUser != nil {
		resp.CreatedBy = &ProjectUserRef{ID: m.CreatedByUser.ID, Name: m.CreatedByUser.Name}
	}
	if m.StartDate != nil {
		s := m.StartDate.Format("2006-01-02")
		resp.StartDate = &s
	}
	if m.EndDate != nil {
		s := m.EndDate.Format("2006-01-02")
		resp.EndDate = &s
	}
	if m.CompletedAt != nil {
		s := m.CompletedAt.Format("2006-01-02")
		resp.CompletedAt = &s
	}
	if m.ArchivedAt != nil {
		s := m.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ArchivedAt = &s
	}
	return resp
}

// ListTasks godoc
// @Summary      List a project's tasks
// @Description  Optionally filter by parent_task_id. Requires task.view + project access.
// @Tags         tasks
// @Security     BearerAuth
// @Produce      json
// @Param        id               path      int   true   "Project ID"
// @Param        parent_task_id   query     int   false  "Filter by parent task"
// @Param        archived         query     bool  false  "Show archived instead of active"
// @Success      200  {array}   TaskResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	archived := c.Query("archived") == "1" || c.Query("archived") == "true"

	params := services.ListTasksParams{ProjectID: projectID, Archived: archived}
	if raw := c.Query("parent_task_id"); raw != "" {
		parentID, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_task_id"})
			return
		}
		pid := uint(parentID)
		params.ParentTaskID = &pid
	}

	items, err := h.tasks.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tasks"})
		return
	}
	resp := make([]TaskResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toTaskResponse(&items[i]))
	}
	c.JSON(http.StatusOK, resp)
}

// GetTask godoc
// @Summary      Get a task
// @Tags         tasks
// @Security     BearerAuth
// @Produce      json
// @Param        id      path      int  true  "Project ID"
// @Param        taskId  path      int  true  "Task ID"
// @Success      200  {object}  TaskResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks/{taskId} [get]
func (h *TaskHandler) Get(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	task, err := h.tasks.Get(projectID, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(task))
}

type TaskRequest struct {
	Type             string   `json:"type" binding:"required"`
	ParentTaskID     *uint    `json:"parent_task_id"`
	Name             string   `json:"name" binding:"required"`
	Description      *string  `json:"description"`
	AssignedToUserID *uint    `json:"assigned_to_user_id"`
	StartDate        *string  `json:"start_date"`
	EndDate          *string  `json:"end_date"`
	BudgetPlan       *float64 `json:"budget_plan"`
	BudgetActual     *float64 `json:"budget_actual"`
	Weight           float64  `json:"weight"`
	StatusLabelID    *uint    `json:"status_label_id"`
}

func (req TaskRequest) toInput() (services.TaskInput, error) {
	start, err := parseOptionalDate(req.StartDate)
	if err != nil {
		return services.TaskInput{}, err
	}
	end, err := parseOptionalDate(req.EndDate)
	if err != nil {
		return services.TaskInput{}, err
	}
	return services.TaskInput{
		Type: models.TaskType(req.Type), ParentTaskID: req.ParentTaskID, Name: req.Name, Description: req.Description,
		AssignedToUserID: req.AssignedToUserID, StartDate: start, EndDate: end,
		BudgetPlan: req.BudgetPlan, BudgetActual: req.BudgetActual, Weight: req.Weight,
		StatusLabelID: req.StatusLabelID,
	}, nil
}

// CreateTask godoc
// @Summary      Create a task
// @Description  Requires task.create + project access.
// @Tags         tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int          true  "Project ID"
// @Param        request  body      TaskRequest  true  "Task"
// @Success      201  {object}  TaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type and name are required"})
		return
	}
	input, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date/end_date must be YYYY-MM-DD"})
		return
	}
	callerID, _ := callerFromContext(c)
	task, err := h.tasks.Create(projectID, callerID, input)
	if err != nil {
		respondTaskError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTaskResponse(task))
}

// UpdateTask godoc
// @Summary      Update a task
// @Description  Requires task.edit + project access.
// @Tags         tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int          true  "Project ID"
// @Param        taskId   path      int          true  "Task ID"
// @Param        request  body      TaskRequest  true  "Task"
// @Success      200  {object}  TaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks/{taskId} [put]
func (h *TaskHandler) Update(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type and name are required"})
		return
	}
	input, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date/end_date must be YYYY-MM-DD"})
		return
	}
	task, err := h.tasks.Update(projectID, taskID, input)
	if err != nil {
		respondTaskError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(task))
}

// ArchiveTask godoc
// @Summary      Archive a task
// @Tags         tasks
// @Security     BearerAuth
// @Param        id      path  int  true  "Project ID"
// @Param        taskId  path  int  true  "Task ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks/{taskId} [delete]
func (h *TaskHandler) Archive(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	if err := h.tasks.Archive(projectID, taskID); err != nil {
		respondTaskError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RestoreTask godoc
// @Summary      Restore an archived task
// @Tags         tasks
// @Security     BearerAuth
// @Param        id      path  int  true  "Project ID"
// @Param        taskId  path  int  true  "Task ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks/{taskId}/restore [post]
func (h *TaskHandler) Restore(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	if err := h.tasks.Restore(projectID, taskID); err != nil {
		respondTaskError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ForceDeleteTask godoc
// @Summary      Permanently delete a task
// @Tags         tasks
// @Security     BearerAuth
// @Param        id      path  int  true  "Project ID"
// @Param        taskId  path  int  true  "Task ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks/{taskId}/force [delete]
func (h *TaskHandler) ForceDelete(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	if err := h.tasks.ForceDelete(projectID, taskID); err != nil {
		respondTaskError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type RescheduleTaskRequest struct {
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

// RescheduleTask godoc
// @Summary      Move a task's dates and cascade the change through its dependents
// @Description  Requires task.edit + project access. Returns every task whose dates changed (the moved task included).
// @Tags         tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Project ID"
// @Param        taskId   path      int                    true  "Task ID"
// @Param        request  body      RescheduleTaskRequest  true  "New dates"
// @Success      200  {array}   TaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks/{taskId}/reschedule [put]
func (h *TaskHandler) Reschedule(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	var req RescheduleTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	start, err := parseOptionalDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be YYYY-MM-DD"})
		return
	}
	end, err := parseOptionalDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be YYYY-MM-DD"})
		return
	}

	changed, err := h.scheduling.Reschedule(projectID, taskID, start, end)
	if err != nil {
		respondTaskError(c, err)
		return
	}
	resp := make([]TaskResponse, 0, len(changed))
	for i := range changed {
		resp = append(resp, toTaskResponse(&changed[i]))
	}
	c.JSON(http.StatusOK, resp)
}

type SetTaskCompletedRequest struct {
	Completed bool `json:"completed"`
}

// SetTaskCompleted godoc
// @Summary      Mark a task done or not done
// @Description  Requires task.edit + project access. Also stamps/clears completed_at.
// @Tags         tasks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "Project ID"
// @Param        taskId   path      int                      true  "Task ID"
// @Param        request  body      SetTaskCompletedRequest  true  "Completed flag"
// @Success      200  {object}  TaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/projects/{id}/tasks/{taskId}/complete [put]
func (h *TaskHandler) SetTaskCompleted(c *gin.Context) {
	projectID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	taskID, err := parseUintParam(c, "taskId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	var req SetTaskCompletedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	task, err := h.tasks.SetCompleted(projectID, taskID, req.Completed)
	if err != nil {
		respondTaskError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(task))
}

func respondTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrParentNotInProject):
		c.JSON(http.StatusBadRequest, gin.H{"error": "parent task does not belong to this project"})
	case errors.Is(err, services.ErrParentNotSummary):
		c.JSON(http.StatusBadRequest, gin.H{"error": "a task can only be nested under a summary task"})
	case errors.Is(err, services.ErrInvalidTaskType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be one of task, summary, milestone"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
	case errors.Is(err, services.ErrTaskNotInProject):
		c.JSON(http.StatusBadRequest, gin.H{"error": "task does not belong to this project"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save task"})
	}
}
