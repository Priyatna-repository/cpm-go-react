package models

import "time"

type TaskRelationType string

const (
	RelationStartToStart   TaskRelationType = "SS"
	RelationFinishToStart  TaskRelationType = "FS"
	RelationFinishToFinish TaskRelationType = "FF"
	RelationStartToFinish  TaskRelationType = "SF"
)

var AllTaskRelationTypes = []TaskRelationType{
	RelationStartToStart, RelationFinishToStart, RelationFinishToFinish, RelationStartToFinish,
}

type TaskDependency struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	TaskID          uint             `gorm:"not null;index;uniqueIndex:idx_task_dependency" json:"task_id"`
	Task            *Task            `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"-"`
	DependsOnTaskID uint             `gorm:"not null;index;uniqueIndex:idx_task_dependency" json:"depends_on_task_id"`
	DependsOnTask   *Task            `gorm:"foreignKey:DependsOnTaskID;constraint:OnDelete:CASCADE" json:"-"`
	RelationType    TaskRelationType `gorm:"size:2;not null" json:"relation_type"`
	LagDays         int              `gorm:"not null;default:0" json:"lag_days"`
	CreatedAt       time.Time        `json:"created_at"`
}
