package models

// role users
const (
	RoleAdmin      = "admin"
	RoleManager    = "manager"
	RoleTeamMember = "team member"
	RoleClient     = "client"
)

var AllRoles = []string{
	RoleAdmin,
	RoleManager,
	RoleTeamMember,
	RoleClient,
}

type Role struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;size: 50;not null"`
}
