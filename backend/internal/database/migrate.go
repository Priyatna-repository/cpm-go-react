package database

import (
	"fmt"
	"log"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Migrate creates/updates tables and seeds the fixed role list, the
// permission catalog, and a bootstrap admin user (from ADMIN_* env vars) if
// one doesn't exist yet.
func Migrate(db *gorm.DB, cfg *config.Config) error {
	if err := db.AutoMigrate(&models.Role{}, &models.Permission{}, &models.User{}, &models.RefreshToken{}, &models.Country{}, &models.Currency{}, &models.OwnerCompany{}, &models.ClientCompany{}); err != nil {
		return err
	}

	if err := seedRoles(db); err != nil {
		return err
	}

	if err := seedPermissions(db); err != nil {
		return err
	}

	if err := seedLookups(db); err != nil {
		return err
	}

	if err := seedOwnerCompany(db); err != nil {
		return err
	}

	return seedAdmin(db, cfg)
}

func seedRoles(db *gorm.DB) error {
	for _, name := range models.AllRoles {
		if err := db.Where(models.Role{Name: name}).FirstOrCreate(&models.Role{Name: name}).Error; err != nil {
			return err
		}
	}
	return nil
}

// defaultPermission describes a permission to seed once, plus which roles
// get it by default. The role assignment only applies the FIRST time this
// permission is created — if it already exists, we leave its role
// assignments alone, since an admin may have already customized them via
// the management UI.
//
// Admin is deliberately never listed in `roles`: it always bypasses
// permission checks (models.IsAdmin), is excluded from the management UI,
// and can't be edited via UpdateRolePermissions — seeding it a row here
// would just be dead data nothing ever reads.
type defaultPermission struct {
	name        string
	category    string
	description string
	roles       []string
}

// permissionCatalog is the full set of permissions the app knows about.
// Each Phase 2+ module adds its own entries here as it ships (e.g.
// project.create) — these two are the first ones, gating the permission
// management feature itself.
var permissionCatalog = []defaultPermission{
	{
		name:        "roles.view",
		category:    "Roles & Permissions",
		description: "View roles and their assigned permissions",
		roles:       []string{models.RoleManager},
	},
	{
		name:        "roles.manage",
		category:    "Roles & Permissions",
		description: "Change which permissions are assigned to a role",
		roles:       []string{},
	},
	// Only admin gets owner_company.* by default (matches the reference
	// app — manager has no "Owner Company" permissions there either).
	// Deliberate, not an oversight; grant manager access via the Roles &
	// Permissions UI if that's ever actually needed.
	{
		name:        "owner_company.view",
		category:    "Owner Company",
		description: "View the owner company profile",
		roles:       []string{},
	},
	{
		name:        "owner_company.edit",
		category:    "Owner Company",
		description: "Edit the owner company profile",
		roles:       []string{},
	},
	{
		name:        "client_company.view",
		category:    "Client Companies",
		description: "View client companies",
		roles:       []string{models.RoleManager},
	},
	{
		name:        "client_company.create",
		category:    "Client Companies",
		description: "Create client companies",
		roles:       []string{models.RoleManager},
	},
	// Manager deliberately gets create/archive/restore but NOT edit here —
	// matches the reference app's actual default permission set exactly
	// (its "manager" role never had "edit client company" either). Not a
	// gap; grant it via the Roles & Permissions UI if needed.
	{
		name:        "client_company.edit",
		category:    "Client Companies",
		description: "Edit client companies",
		roles:       []string{},
	},
	{
		name:        "client_company.archive",
		category:    "Client Companies",
		description: "Archive client companies",
		roles:       []string{models.RoleManager},
	},
	{
		name:        "client_company.restore",
		category:    "Client Companies",
		description: "Restore archived client companies",
		roles:       []string{models.RoleManager},
	},
	{
		name:        "client_company.delete",
		category:    "Client Companies",
		description: "Permanently delete client companies",
		roles:       []string{},
	},
}

func seedPermissions(db *gorm.DB) error {
	var allRoles []models.Role
	if err := db.Find(&allRoles).Error; err != nil {
		return err
	}
	roleByName := make(map[string]models.Role, len(allRoles))
	for _, r := range allRoles {
		roleByName[r.Name] = r
	}

	for _, d := range permissionCatalog {
		var perm models.Permission
		result := db.Where(models.Permission{Name: d.name}).
			Attrs(models.Permission{Category: d.category, Description: d.description}).
			FirstOrCreate(&perm)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			continue // already existed — don't touch its role assignments
		}

		for _, roleName := range d.roles {
			role, ok := roleByName[roleName]
			if !ok {
				return fmt.Errorf("seedPermissions: role %q not found", roleName)
			}
			if err := db.Model(&role).Association("Permissions").Append(&perm); err != nil {
				return err
			}
		}
	}
	return nil
}

func seedAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		log.Println("ADMIN_EMAIL/ADMIN_PASSWORD not set, skipping admin bootstrap")
		return nil
	}

	var count int64
	if err := db.Model(&models.User{}).Where("email = ?", cfg.AdminEmail).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var adminRole models.Role
	if err := db.Where("name = ?", models.RoleAdmin).First(&adminRole).Error; err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Name:         cfg.AdminName,
		Email:        cfg.AdminEmail,
		PasswordHash: string(hash),
		RoleID:       adminRole.ID,
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	log.Printf("bootstrap admin user created: %s", cfg.AdminEmail)
	return nil
}

func seedOwnerCompany(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.OwnerCompany{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&models.OwnerCompany{Name: "My Company"}).Error
}
