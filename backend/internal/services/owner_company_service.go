package services

import (
	"mime/multipart"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

type OwnerCompanyService struct {
	db      *gorm.DB
	uploads *UploadService
}

func NewOwnerCompanyService(db *gorm.DB, uploads *UploadService) *OwnerCompanyService {
	return &OwnerCompanyService{db: db, uploads: uploads}
}

func (s *OwnerCompanyService) Get() (*models.OwnerCompany, error) {
	var company models.OwnerCompany
	if err := s.db.Preload("Country").Preload("Currency").First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

type UpdateOwnerCompanyInput struct {
	Name       string
	Address    *string
	PostalCode *string
	City       *string
	CountryID  *uint
	CurrencyID *uint
	Email      *string
	Phone      *string
	Web        *string
}

// Update edits the single owner company row. If logo is non-nil, the new
// image is saved and the old one deleted — unlike the reference app, which
// leaks the old logo file on every replace.
func (s *OwnerCompanyService) Update(input UpdateOwnerCompanyInput, logo *multipart.FileHeader) (*models.OwnerCompany, error) {
	var company models.OwnerCompany
	if err := s.db.First(&company).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"name":        input.Name,
		"address":     input.Address,
		"postal_code": input.PostalCode,
		"city":        input.City,
		"country_id":  input.CountryID,
		"currency_id": input.CurrencyID,
		"email":       input.Email,
		"phone":       input.Phone,
		"web":         input.Web,
	}

	if logo == nil {
		if err := s.db.Model(&company).Updates(updates).Error; err != nil {
			return nil, err
		}
		return s.Get()
	}

	newPath, err := s.uploads.SaveImage(logo, "owner-company")
	if err != nil {
		return nil, err
	}
	updates["logo"] = newPath

	oldLogo := company.Logo
	if err := s.db.Model(&company).Updates(updates).Error; err != nil {
		_ = s.uploads.DeleteImage(newPath) // rollback the file we just saved
		return nil, err
	}
	if oldLogo != nil {
		_ = s.uploads.DeleteImage(*oldLogo)
	}

	return s.Get()
}
