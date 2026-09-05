package models

import "time"

// Known label types from the reference app. Only kontrak_label and
// pt_status are actually consumed anywhere yet (by Project); the rest
// exist so the type picker matches the reference catalog and are wired
// up when their owning module (Task, Inventory, WorkReport) ships.
const (
	LabelTypeKontrak                  = "kontrak_label"
	LabelTypeProjectTaskStatus        = "pt_status"
	LabelTypeProjectTaskBillingStatus = "ptb_status"
	LabelTypeTaskRelation             = "task_relation"
	LabelTypeInventoryStatus          = "inventory_status_label"
	LabelTypeInventoryType            = "inventory_type_label"
	LabelTypeTaskInventoryUnit        = "task_inventory_unit_label"
	LabelTypeTask                     = "task_type_label"
	LabelTypeWorkReportStatus         = "work_report_status"
	LabelTypePriority                 = "task_priority_label"
)

type Label struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"size:255;not null" json:"name"`
	Slug       string     `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Type       string     `gorm:"size:255;not null;index" json:"type"`
	Color      *string    `gorm:"size:20" json:"color,omitempty"`
	Icon       *string    `gorm:"size:255" json:"icon,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
