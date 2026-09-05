package models

import "time"

type Project struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Code            string         `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name            string         `gorm:"size:255;uniqueIndex;not null" json:"name"`
	Description     *string        `json:"description,omitempty"`
	ClientCompanyID *uint          `json:"-"`
	ClientCompany   *ClientCompany `gorm:"foreignKey:ClientCompanyID" json:"client_company,omitempty"`
	ClientUserID    *uint          `json:"-"`
	ClientUser      *User          `gorm:"foreignKey:ClientUserID" json:"client_user,omitempty"`
	StartDate       *time.Time     `json:"start_date,omitempty"`
	EndDate         *time.Time     `json:"end_date,omitempty"`
	BudgetEstimate  *float64       `gorm:"type:decimal(15,2)" json:"budget_estimate,omitempty"`
	TypeLabelID     *uint          `json:"-"`
	TypeLabel       *Label         `gorm:"foreignKey:TypeLabelID" json:"type_label,omitempty"`
	StatusLabels    []Label        `gorm:"many2many:project_status_labels;" json:"status_labels,omitempty"`
	IsCompleted     bool           `gorm:"not null;default:false" json:"is_completed"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`
	ArchivedAt      *time.Time     `json:"archived_at,omitempty"`
	Users           []User         `gorm:"many2many:project_user_access;" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
