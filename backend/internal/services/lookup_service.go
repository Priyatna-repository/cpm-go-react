package services

import (
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

type LookupService struct {
	db *gorm.DB
}

func NewLookupService(db *gorm.DB) *LookupService {
	return &LookupService{db: db}
}

func (s *LookupService) ListCountries() ([]models.Country, error) {
	var countries []models.Country
	err := s.db.Order("name").Find(&countries).Error
	return countries, err
}

func (s *LookupService) ListCurrencies() ([]models.Currency, error) {
	var currencies []models.Currency
	err := s.db.Order("name").Find(&currencies).Error
	return currencies, err
}
