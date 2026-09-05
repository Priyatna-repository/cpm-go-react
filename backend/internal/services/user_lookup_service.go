package services

import (
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

type UserLookupService struct {
	db *gorm.DB
}

func NewUserLookupService(db *gorm.DB) *UserLookupService {
	return &UserLookupService{db: db}
}

// ListClientUsers returns client-role users not currently linked to any
// OTHER client company (a client user belongs to at most one company).
// excludeCompanyID, when non-nil, lets that company's own members still
// show up (so editing an existing company doesn't hide its own clients).
func (s *UserLookupService) ListClientUsers(excludeCompanyID *uint) ([]models.User, error) {
	var users []models.User
	err := s.db.Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", models.RoleClient).
		Where("NOT EXISTS (SELECT 1 FROM client_company_user ccu WHERE ccu.user_id = users.id AND ccu.client_company_id != COALESCE(?, 0))", excludeCompanyID).
		Order("users.name").
		Find(&users).Error
	return users, err
}

// ListInternalUsers returns every user that is NOT a client — used to
// populate the "grant project access" picker (client access is automatic
// via company/individual ownership, so only internal staff need explicit
// grants).
func (s *UserLookupService) ListInternalUsers() ([]models.User, error) {
	var users []models.User
	err := s.db.Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name != ?", models.RoleClient).
		Order("users.name").
		Find(&users).Error
	return users, err
}
