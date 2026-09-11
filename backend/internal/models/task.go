package models

import "time"

type TaskType string

const (
	TaskTypeTask      TaskType = "task"
	TaskTypeSummary   TaskType = "summary"
	TaskTypeMilestone TaskType = "milestone"
)

var AllTaskTypes = []TaskType{TaskTypeTask, TaskTypeSummary, TaskTypeMilestone}

type Task struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	ProjectID        uint       `gorm:"not null;index" json:"project_id"`
	Project          *Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	Type             TaskType   `gorm:"size:20;not null;default:'task'" json:"type"`
	ParentTaskID     *uint      `gorm:"index" json:"-"`
	Parent           *Task      `gorm:"foreignKey:ParentTaskID;constraint:OnDelete:CASCADE" json:"-"`
	Name             string     `gorm:"size:255;not null" json:"name"`
	Description      *string    `json:"description,omitempty"`
	Number           int        `gorm:"not null" json:"number"`
	OrderColumn      int        `gorm:"not null;default:0" json:"order_column"`
	AssignedToUserID *uint      `json:"-"`
	AssignedToUser   *User      `gorm:"foreignKey:AssignedToUserID" json:"assigned_to_user,omitempty"`
	StartDate        *time.Time `json:"start_date,omitempty"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	BudgetPlan       *float64   `gorm:"type:decimal(15,2)" json:"budget_plan,omitempty"`
	BudgetActual     *float64   `gorm:"type:decimal(15,2)" json:"budget_actual,omitempty"`
	Weight           float64    `gorm:"type:decimal(8,2);not null;default:0" json:"weight"`
	Progress         float64    `gorm:"type:decimal(5,2);not null;default:0" json:"progress"`
	StatusLabelID    *uint      `json:"-"`
	StatusLabel      *Label     `gorm:"foreignKey:StatusLabelID" json:"status_label,omitempty"`
	CreatedByUserID  *uint      `json:"-"`
	CreatedByUser    *User      `gorm:"foreignKey:CreatedByUserID" json:"-"`
	IsCompleted      bool       `gorm:"not null;default:false" json:"is_completed"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	ArchivedAt       *time.Time `json:"archived_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
