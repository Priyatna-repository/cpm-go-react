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

// IsAdmin reports whether roleName is the admin role — the one role that
// always bypasses permission checks. Use this everywhere instead of
// comparing against RoleAdmin directly, so there's one place to change if
// the semantics ever do.
func IsAdmin(roleName string) bool {
	return roleName == RoleAdmin
}

type Role struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"uniqueIndex;size: 50;not null"`
	Version     int          `gorm:"not null;default:1"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}
