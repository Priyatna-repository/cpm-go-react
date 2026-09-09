package services

import (
	"errors"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

type TaskSchedulingService struct {
	db *gorm.DB
}

func NewTaskSchedulingService(db *gorm.DB) *TaskSchedulingService {
	return &TaskSchedulingService{db: db}
}

// Reschedule sets taskID's dates, then cascades that change through every
// task that (transitively) depends on it, respecting each dependency's
// relation-type rule and lag. Returns every task whose dates actually
// changed (taskID included) so the caller can patch just those rows.
func (s *TaskSchedulingService) Reschedule(projectID, taskID uint, newStart, newEnd *time.Time) ([]models.Task, error) {
	var changed []models.Task

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var tasks []models.Task
		if err := tx.Where("project_id = ? AND archived_at IS NULL", projectID).Find(&tasks).Error; err != nil {
			return err
		}
		byID := make(map[uint]*models.Task, len(tasks))
		original := make(map[uint][2]*time.Time, len(tasks))
		for i := range tasks {
			byID[tasks[i].ID] = &tasks[i]
			original[tasks[i].ID] = [2]*time.Time{tasks[i].StartDate, tasks[i].EndDate}
		}
		if _, ok := byID[taskID]; !ok {
			return ErrTaskNotInProject
		}

		var deps []models.TaskDependency
		if err := tx.Where("task_id IN (SELECT id FROM tasks WHERE project_id = ? AND archived_at IS NULL)", projectID).
			Find(&deps).Error; err != nil {
			return err
		}
		predecessorsOf := make(map[uint][]models.TaskDependency)
		for _, d := range deps {
			predecessorsOf[d.TaskID] = append(predecessorsOf[d.TaskID], d)
		}

		order, err := topologicalOrder(tasks, deps)
		if err != nil {
			return err
		}

		byID[taskID].StartDate = newStart
		byID[taskID].EndDate = newEnd

		for _, id := range order {
			if id == taskID {
				continue // the dragged task's dates are the fixed input, not re-derived from its own predecessors
			}
			task := byID[id]
			for _, dep := range predecessorsOf[id] {
				pred := byID[dep.DependsOnTaskID]
				if pred == nil || pred.StartDate == nil || pred.EndDate == nil {
					continue
				}
				lag := time.Duration(dep.LagDays) * 24 * time.Hour
				switch dep.RelationType {
				case models.RelationFinishToStart:
					shiftIfBefore(task, pred.EndDate.Add(lag), true)
				case models.RelationStartToStart:
					shiftIfBefore(task, pred.StartDate.Add(lag), true)
				case models.RelationFinishToFinish:
					shiftIfBefore(task, pred.EndDate.Add(lag), false)
				case models.RelationStartToFinish:
					shiftIfBefore(task, pred.StartDate.Add(lag), false)
				}
			}
		}

		for _, id := range order {
			task := byID[id]
			orig := original[id]
			if datesEqual(orig[0], task.StartDate) && datesEqual(orig[1], task.EndDate) {
				continue
			}
			if err := tx.Model(&models.Task{}).Where("id = ?", id).
				Updates(map[string]interface{}{"start_date": task.StartDate, "end_date": task.EndDate}).Error; err != nil {
				return err
			}
			changed = append(changed, *task)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return changed, nil
}

// shiftIfBefore moves task forward (preserving its duration) if the
// relevant boundary is earlier than required. constrainsStart is true for
// rules that impose a minimum START (FS, SS) and false for rules that
// impose a minimum END (FF, SF).
func shiftIfBefore(task *models.Task, required time.Time, constrainsStart bool) {
	if task.StartDate == nil || task.EndDate == nil {
		return
	}
	if constrainsStart {
		if task.StartDate.Before(required) {
			shift := required.Sub(*task.StartDate)
			newStart, newEnd := required, task.EndDate.Add(shift)
			task.StartDate, task.EndDate = &newStart, &newEnd
		}
	} else {
		if task.EndDate.Before(required) {
			shift := required.Sub(*task.EndDate)
			newEnd, newStart := required, task.StartDate.Add(shift)
			task.StartDate, task.EndDate = &newStart, &newEnd
		}
	}
}

func datesEqual(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

// topologicalOrder returns task IDs ordered so every task appears after
// everything it depends on — Kahn's algorithm: repeatedly peel off nodes
// with no remaining unprocessed predecessors. Dependency creation already
// rejects cycles, so a valid order always exists for live data; the length
// check at the end is a defensive guard, not something that should ever
// actually trigger.
func topologicalOrder(tasks []models.Task, deps []models.TaskDependency) ([]uint, error) {
	inDegree := make(map[uint]int, len(tasks))
	successors := make(map[uint][]uint)
	for _, t := range tasks {
		inDegree[t.ID] = 0
	}
	for _, d := range deps {
		if _, ok := inDegree[d.TaskID]; !ok {
			continue
		}
		if _, ok := inDegree[d.DependsOnTaskID]; !ok {
			continue
		}
		inDegree[d.TaskID]++
		successors[d.DependsOnTaskID] = append(successors[d.DependsOnTaskID], d.TaskID)
	}

	var queue []uint
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	var order []uint
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)
		for _, next := range successors[current] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if len(order) != len(tasks) {
		return nil, errors.New("dependency graph contains a cycle")
	}
	return order, nil
}
