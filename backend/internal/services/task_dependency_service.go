package services

import (
	"errors"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrInvalidRelationType = errors.New("relation_type must be one of SS, FS, FF, SF")
	ErrSelfDependency      = errors.New("a task cannot depend on itself")
	ErrDependencyCycle     = errors.New("this dependency would create a circular chain")
	ErrTaskNotInProject    = errors.New("task does not belong to this project")
	ErrDuplicateDependency = errors.New("this dependency already exists")
)

type TaskDependencyService struct {
	db *gorm.DB
}

func NewTaskDependencyService(db *gorm.DB) *TaskDependencyService {
	return &TaskDependencyService{db: db}
}

func validRelationType(rt models.TaskRelationType) bool {
	for _, v := range models.AllTaskRelationTypes {
		if v == rt {
			return true
		}
	}
	return false
}

func (s *TaskDependencyService) validateTaskInProject(tx *gorm.DB, projectID, taskID uint) error {
	var count int64
	if err := tx.Model(&models.Task{}).Where("id = ? AND project_id = ?", taskID, projectID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrTaskNotInProject
	}
	return nil
}

type CreateDependencyInput struct {
	TaskID          uint
	DependsOnTaskID uint
	RelationType    models.TaskRelationType
	LagDays         int
}

func (s *TaskDependencyService) Create(projectID uint, input CreateDependencyInput) (*models.TaskDependency, error) {
	if input.TaskID == input.DependsOnTaskID {
		return nil, ErrSelfDependency
	}
	if !validRelationType(input.RelationType) {
		return nil, ErrInvalidRelationType
	}

	var dep models.TaskDependency
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.validateTaskInProject(tx, projectID, input.TaskID); err != nil {
			return err
		}
		if err := s.validateTaskInProject(tx, projectID, input.DependsOnTaskID); err != nil {
			return err
		}

		cyclic, err := s.wouldCreateCycle(tx, input.TaskID, input.DependsOnTaskID)
		if err != nil {
			return err
		}
		if cyclic {
			return ErrDependencyCycle
		}

		dep = models.TaskDependency{
			TaskID: input.TaskID, DependsOnTaskID: input.DependsOnTaskID,
			RelationType: input.RelationType, LagDays: input.LagDays,
		}
		if err := tx.Create(&dep).Error; err != nil {
			if isDuplicateKey(err) {
				return ErrDuplicateDependency
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

// wouldCreateCycle checks whether "successorID depends on predecessorID"
// would close a loop. A TaskDependency row means predecessorID must
// happen before successorID — so walking FORWARD from successorID
// (following existing "X is a predecessor of Y" edges) tells us everything
// that already comes after successorID. If predecessorID shows up in that
// set, successorID already (transitively) precedes predecessorID — so
// adding "predecessorID before successorID" would close a loop.
func (s *TaskDependencyService) wouldCreateCycle(tx *gorm.DB, successorID, predecessorID uint) (bool, error) {
	visited := map[uint]bool{successorID: true}
	queue := []uint{successorID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == predecessorID {
			return true, nil
		}
		var successors []uint
		if err := tx.Model(&models.TaskDependency{}).
			Where("depends_on_task_id = ?", current).
			Pluck("task_id", &successors).Error; err != nil {
			return false, err
		}
		for _, next := range successors {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false, nil
}

func (s *TaskDependencyService) Delete(projectID, dependencyID uint) error {
	result := s.db.
		Where("id = ? AND task_id IN (SELECT id FROM tasks WHERE project_id = ?)", dependencyID, projectID).
		Delete(&models.TaskDependency{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListForProject returns every dependency edge among a project's tasks —
// the Gantt needs the full set at once to draw arrows, not paginated per task.
func (s *TaskDependencyService) ListForProject(projectID uint) ([]models.TaskDependency, error) {
	var deps []models.TaskDependency
	err := s.db.
		Where("task_id IN (SELECT id FROM tasks WHERE project_id = ?)", projectID).
		Find(&deps).Error
	return deps, err
}
