package services

import (
	"errors"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

var (
	// ErrRoleNotManageable is returned when trying to change the admin role's
	// permissions — admin always bypasses permission checks, so its assigned
	// permissions are meaningless and not editable here.
	ErrRoleNotManageable = errors.New("this role's permissions cannot be changed here")

	// ErrUnknownPermission is returned when UpdateRolePermissions is given a
	// permission name that isn't in the catalog — better to fail loudly than
	// silently drop it and let the caller believe it was applied.
	ErrUnknownPermission = errors.New("one or more permission names are not in the catalog")

	// ErrStaleRole is returned when the caller's expected version doesn't
	// match the role's current version — someone else changed it first.
	ErrStaleRole = errors.New("role was changed by someone else, reload and try again")
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// HasPermission checks whether a role (by name) has a permission (by name)
// assigned to it. Callers should special-case the admin role themselves —
// this method does no bypass logic, it's a literal DB lookup.
func (s *PermissionService) HasPermission(role, permission string) (bool, error) {
	var count int64
	err := s.db.Table("role_permissions").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("roles.name = ? AND permissions.name = ?", role, permission).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *PermissionService) ListPermissions() ([]models.Permission, error) {
	var perms []models.Permission
	err := s.db.Order("category, name").Find(&perms).Error
	return perms, err
}

type RoleWithPermissions struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Version     int      `json:"version"`
	Permissions []string `json:"permissions"`
}

// ListManageableRoles returns every role except admin — admin's access
// can't be restricted, so it's not shown in the management UI at all.
func (s *PermissionService) ListManageableRoles() ([]RoleWithPermissions, error) {
	var roles []models.Role
	if err := s.db.Preload("Permissions").Where("name != ?", models.RoleAdmin).Order("name").Find(&roles).Error; err != nil {
		return nil, err
	}

	result := make([]RoleWithPermissions, 0, len(roles))
	for _, r := range roles {
		result = append(result, RoleWithPermissions{
			ID: r.ID, Name: r.Name, Version: r.Version, Permissions: PermissionNames(r.Permissions),
		})
	}
	return result, nil
}

// UpdateRolePermissions replaces a role's full set of assigned permissions,
// but only if expectedVersion still matches the role's current version —
// otherwise ErrStaleRole, so a save from a stale page load can't silently
// clobber a concurrent change.
func (s *PermissionService) UpdateRolePermissions(roleID uint, expectedVersion int, permissionNames []string) (*RoleWithPermissions, error) {
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return nil, err
	}
	if models.IsAdmin(role.Name) {
		return nil, ErrRoleNotManageable
	}

	unique := uniquePermissionNames(permissionNames)

	var perms []models.Permission
	if len(unique) > 0 {
		if err := s.db.Where("name IN ?", unique).Find(&perms).Error; err != nil {
			return nil, err
		}
		if len(perms) != len(unique) {
			return nil, ErrUnknownPermission
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Role{}).
			Where("id = ? AND version = ?", role.ID, expectedVersion).
			Update("version", gorm.Expr("version + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrStaleRole
		}
		return tx.Model(&role).Association("Permissions").Replace(perms)
	})
	if err != nil {
		return nil, err
	}

	return &RoleWithPermissions{
		ID: role.ID, Name: role.Name, Version: expectedVersion + 1, Permissions: PermissionNames(perms),
	}, nil
}

func uniquePermissionNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			out = append(out, n)
		}
	}
	return out
}

func PermissionNames(perms []models.Permission) []string {
	names := make([]string, 0, len(perms))
	for _, p := range perms {
		names = append(names, p.Name)
	}
	return names
}
