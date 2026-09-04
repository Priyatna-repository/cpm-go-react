package database

import (
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
	if err := db.AutoMigrate(&models.Role{}, &models.Permission{}, &models.User{}, &models.RefreshToken{}); err != nil {
		return err
	}

	if err := seedRoles(db); err != nil {
		return err
	}

	if err := seedPermissions(db); err != nil {
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
		roles:       []string{models.RoleAdmin, models.RoleManager},
	},
	{
		name:        "roles.manage",
		category:    "Roles & Permissions",
		description: "Change which permissions are assigned to a role",
		roles:       []string{models.RoleAdmin},
	},
}

func seedPermissions(db *gorm.DB) error {
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
			var role models.Role
			if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
				return err
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
