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

// ListClientUsers returns every user with the "client" role — used to
// populate the "assign client" picker on the ClientCompany form. This is
// not a Users management module (that's separate, unbuilt work); it's a
// narrow read-only lookup, same spirit as ListCountries/ListCurrencies.
func (s *UserLookupService) ListClientUsers() ([]models.User, error) {
	var users []models.User
	err := s.db.Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", models.RoleClient).
		Order("users.name").
		Find(&users).Error
	return users, err
}
