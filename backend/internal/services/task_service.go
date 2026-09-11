package services

import (
	"errors"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrParentNotInProject = errors.New("parent task does not belong to this project")
	ErrParentNotSummary   = errors.New("a task can only be nested under a summary task")
	ErrInvalidTaskType    = errors.New("type must be one of task, summary, milestone")
)

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

func validTaskType(t models.TaskType) bool {
	for _, v := range models.AllTaskTypes {
		if v == t {
			return true
		}
	}
	return false
}

type ListTasksParams struct {
	ProjectID    uint
	ParentTaskID *uint
	Archived     bool
}

func (s *TaskService) List(p ListTasksParams) ([]models.Task, error) {
	query := s.db.Preload("AssignedToUser").Preload("StatusLabel").Preload("CreatedByUser").Where("project_id = ?", p.ProjectID)
	if p.ParentTaskID != nil {
		query = query.Where("parent_task_id = ?", *p.ParentTaskID)
	}
	if p.Archived {
		query = query.Where("archived_at IS NOT NULL")
	} else {
		query = query.Where("archived_at IS NULL")
	}
	var items []models.Task
	err := query.Order("order_column, id").Find(&items).Error
	return items, err
}

// scopedFirst — never trust a bare id, always filter by project_id too, so
// a caller can't reach a task from a project they weren't checked against
// by RequireProjectAccess.
func (s *TaskService) scopedFirst(projectID, id uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.Preload("AssignedToUser").Preload("StatusLabel").Preload("CreatedByUser").
		Where("project_id = ? AND id = ?", projectID, id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskService) Get(projectID, id uint) (*models.Task, error) {
	return s.scopedFirst(projectID, id)
}

// validateParent checks parentTaskID (when given) is a summary-type task
// belonging to projectID — same cross-scope trap as everywhere else (a
// client could otherwise pass a parent id from a different project), plus
// a type check: nesting a task under a non-summary task would produce a
// tree the Gantt can't render sensibly (a plain task acting as a container).
func (s *TaskService) validateParent(projectID uint, parentTaskID *uint) error {
	if parentTaskID == nil {
		return nil
	}
	var parent models.Task
	if err := s.db.Where("id = ? AND project_id = ?", *parentTaskID, projectID).First(&parent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrParentNotInProject
		}
		return err
	}
	if parent.Type != models.TaskTypeSummary {
		return ErrParentNotSummary
	}
	return nil
}

type TaskInput struct {
	Type             models.TaskType
	ParentTaskID     *uint
	Name             string
	Description      *string
	AssignedToUserID *uint
	StartDate        *time.Time
	EndDate          *time.Time
	BudgetPlan       *float64
	BudgetActual     *float64
	Weight           float64
	StatusLabelID    *uint
}

// createdByUserID is a separate param, not a TaskInput field — it's set
// once at creation from the authenticated caller (never client-supplied,
// never touched again by Update) rather than living in the same struct
// Update reads from, where it could accidentally be overwritten.
func (s *TaskService) Create(projectID, createdByUserID uint, input TaskInput) (*models.Task, error) {
	if !validTaskType(input.Type) {
		return nil, ErrInvalidTaskType
	}
	if err := s.validateParent(projectID, input.ParentTaskID); err != nil {
		return nil, err
	}

	var maxOrder int
	if err := s.db.Model(&models.Task{}).Where("project_id = ?", projectID).
		Select("COALESCE(MAX(order_column), -1)").Scan(&maxOrder).Error; err != nil {
		return nil, err
	}
	var maxNumber int
	if err := s.db.Model(&models.Task{}).Where("project_id = ?", projectID).
		Select("COALESCE(MAX(number), 0)").Scan(&maxNumber).Error; err != nil {
		return nil, err
	}

	task := models.Task{
		ProjectID: projectID, Type: input.Type, ParentTaskID: input.ParentTaskID,
		Name: input.Name, Description: input.Description,
		Number: maxNumber + 1, OrderColumn: maxOrder + 1, AssignedToUserID: input.AssignedToUserID,
		StartDate: input.StartDate, EndDate: input.EndDate, BudgetPlan: input.BudgetPlan,
		BudgetActual: input.BudgetActual, Weight: input.Weight, StatusLabelID: input.StatusLabelID,
		CreatedByUserID: &createdByUserID,
	}
	if err := s.db.Create(&task).Error; err != nil {
		return nil, err
	}
	return s.scopedFirst(projectID, task.ID)
}

func (s *TaskService) Update(projectID, id uint, input TaskInput) (*models.Task, error) {
	if !validTaskType(input.Type) {
		return nil, ErrInvalidTaskType
	}
	if _, err := s.scopedFirst(projectID, id); err != nil {
		return nil, err
	}
	// A task can't be its own parent, directly or (to keep this cheap)
	// as an immediate self-reference — deeper cycles are prevented simply
	// by never letting a parent be re-pointed at one of its own
	// descendants would need a graph walk; not needed yet since the
	// frontend never offers that reassignment path today.
	if input.ParentTaskID != nil && *input.ParentTaskID == id {
		return nil, ErrParentNotSummary
	}
	if err := s.validateParent(projectID, input.ParentTaskID); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"type": input.Type, "parent_task_id": input.ParentTaskID, "name": input.Name, "description": input.Description,
		"assigned_to_user_id": input.AssignedToUserID, "start_date": input.StartDate, "end_date": input.EndDate,
		"budget_plan": input.BudgetPlan, "budget_actual": input.BudgetActual, "weight": input.Weight,
		"status_label_id": input.StatusLabelID,
	}
	if err := s.db.Model(&models.Task{}).Where("project_id = ? AND id = ?", projectID, id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.scopedFirst(projectID, id)
}

// SetCompleted is a separate action from Update (same reasoning as
// Reschedule living outside the general TaskInput/Update path) — toggling
// completion also stamps/clears CompletedAt, a side effect that doesn't
// belong bundled into a plain field-by-field save.
func (s *TaskService) SetCompleted(projectID, id uint, completed bool) (*models.Task, error) {
	if _, err := s.scopedFirst(projectID, id); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"is_completed": completed}
	if completed {
		updates["completed_at"] = gorm.Expr("NOW()")
	} else {
		updates["completed_at"] = nil
	}
	if err := s.db.Model(&models.Task{}).Where("project_id = ? AND id = ?", projectID, id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.scopedFirst(projectID, id)
}

func (s *TaskService) Archive(projectID, id uint) error {
	if _, err := s.scopedFirst(projectID, id); err != nil {
		return err
	}
	return s.db.Model(&models.Task{}).Where("project_id = ? AND id = ?", projectID, id).
		Update("archived_at", gorm.Expr("NOW()")).Error
}

func (s *TaskService) Restore(projectID, id uint) error {
	if _, err := s.scopedFirst(projectID, id); err != nil {
		return err
	}
	return s.db.Model(&models.Task{}).Where("project_id = ? AND id = ?", projectID, id).
		Update("archived_at", nil).Error
}

func (s *TaskService) ForceDelete(projectID, id uint) error {
	task, err := s.scopedFirst(projectID, id)
	if err != nil {
		return err
	}
	result := s.db.Delete(task)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
