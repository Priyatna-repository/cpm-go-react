package models

type Country struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:255;uniqueIndex;not null" json:"name"`
}
