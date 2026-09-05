package models

import "time"

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"size:255;not null" json:"name"`
	Email        string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	GoogleID     *string    `gorm:"size:255;uniqueIndex" json:"-"`
	RoleID       uint       `gorm:"not null" json:"-"`
	Role         Role       `gorm:"foreignKey:RoleID" json:"role"`
	Avatar       *string    `gorm:"size:255" json:"avatar,omitempty"`
	JobTitle     *string    `gorm:"size:255" json:"job_title,omitempty"`
	Phone        *string    `gorm:"size:50" json:"phone,omitempty"`
	Address      *string    `json:"address,omitempty"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
