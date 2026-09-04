package models

type Permission struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Category    string `gorm:"size:100;not null" json:"category"`
	Description string `gorm:"size:255" json:"description,omitempty"`
}
