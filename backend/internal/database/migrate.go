package database

import (
	"log"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Migrate creates/updates tables and seeds the fixed role list plus a
// bootstrap admin user (from ADMIN_* env vars) if one doesn't exist yet.
func Migrate(db *gorm.DB, cfg *config.Config) error {
	if err := db.AutoMigrate(&models.Role{}, &models.User{}, &models.RefreshToken{}); err != nil {
		return err
	}

	if err := seedRoles(db); err != nil {
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
