package services

import (
	"mime/multipart"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

type ClientCompanyService struct {
	db      *gorm.DB
	uploads *UploadService
}

func NewClientCompanyService(db *gorm.DB, uploads *UploadService) *ClientCompanyService {
	return &ClientCompanyService{db: db, uploads: uploads}
}

type ListClientCompaniesParams struct {
	Page     int
	PerPage  int
	Search   string
	Archived bool
}

type ClientCompanyPage struct {
	Items       []models.ClientCompany
	CurrentPage int
	LastPage    int
	Total       int64
}

func (s *ClientCompanyService) List(p ListClientCompaniesParams) (*ClientCompanyPage, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 12
	}

	query := s.db.Model(&models.ClientCompany{}).Preload("Country").Preload("Currency").Preload("Clients")
	if p.Archived {
		query = query.Where("archived_at IS NOT NULL")
	} else {
		query = query.Where("archived_at IS NULL")
	}
	if p.Search != "" {
		like := "%" + p.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.ClientCompany
	offset := (p.Page - 1) * p.PerPage
	if err := query.Order("name").Limit(p.PerPage).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	lastPage := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if lastPage < 1 {
		lastPage = 1
	}

	return &ClientCompanyPage{Items: items, CurrentPage: p.Page, LastPage: lastPage, Total: total}, nil
}

func (s *ClientCompanyService) Get(id uint) (*models.ClientCompany, error) {
	var company models.ClientCompany
	if err := s.db.Preload("Country").Preload("Currency").Preload("Clients").First(&company, id).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

type ClientCompanyInput struct {
	Name       string
	Address    *string
	PostalCode *string
	City       *string
	CountryID  *uint
	CurrencyID *uint
	Email      *string
	Phone      *string
	Web        *string
	ClientIDs  []uint
}

func (s *ClientCompanyService) Create(input ClientCompanyInput, logo *multipart.FileHeader) (*models.ClientCompany, error) {
	company := models.ClientCompany{
		Name: input.Name, Address: input.Address, PostalCode: input.PostalCode, City: input.City,
		CountryID: input.CountryID, CurrencyID: input.CurrencyID, Email: input.Email, Phone: input.Phone, Web: input.Web,
	}

	if logo != nil {
		path, err := s.uploads.SaveImage(logo, "client-companies")
		if err != nil {
			return nil, err
		}
		company.Logo = &path
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&company).Error; err != nil {
			return err
		}
		if len(input.ClientIDs) == 0 {
			return nil
		}
		var clients []models.User
		if err := tx.Where("id IN ?", input.ClientIDs).Find(&clients).Error; err != nil {
			return err
		}
		return tx.Model(&company).Association("Clients").Replace(clients)
	})
	if err != nil {
		if company.Logo != nil {
			_ = s.uploads.DeleteImage(*company.Logo)
		}
		return nil, err
	}

	return s.Get(company.ID)
}

func (s *ClientCompanyService) Update(id uint, input ClientCompanyInput, logo *multipart.FileHeader) (*models.ClientCompany, error) {
	var company models.ClientCompany
	if err := s.db.First(&company, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"name": input.Name, "address": input.Address, "postal_code": input.PostalCode, "city": input.City,
		"country_id": input.CountryID, "currency_id": input.CurrencyID, "email": input.Email, "phone": input.Phone, "web": input.Web,
	}

	var newLogo, oldLogo *string
	if logo != nil {
		path, err := s.uploads.SaveImage(logo, "client-companies")
		if err != nil {
			return nil, err
		}
		newLogo = &path
		oldLogo = company.Logo
		updates["logo"] = path
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&company).Updates(updates).Error; err != nil {
			return err
		}

		var clients []models.User
		if len(input.ClientIDs) > 0 {
			if err := tx.Where("id IN ?", input.ClientIDs).Find(&clients).Error; err != nil {
				return err
			}
		}
		// Always Replace, even with an empty slice (which clears every
		// assignment). The reference app's sync() only ran when the list
		// was non-empty, so an admin could never actually remove every
		// assigned client through the UI — we don't repeat that bug.
		return tx.Model(&company).Association("Clients").Replace(clients)
	})
	if err != nil {
		if newLogo != nil {
			_ = s.uploads.DeleteImage(*newLogo)
		}
		return nil, err
	}

	if oldLogo != nil {
		_ = s.uploads.DeleteImage(*oldLogo)
	}

	return s.Get(id)
}

func (s *ClientCompanyService) Archive(id uint) error {
	result := s.db.Model(&models.ClientCompany{}).Where("id = ?", id).Update("archived_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *ClientCompanyService) Restore(id uint) error {
	result := s.db.Model(&models.ClientCompany{}).Where("id = ?", id).Update("archived_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ForceDelete permanently removes a client company. Once the Project
// module exists (client_company_id FK on projects), this will need to
// translate a foreign-key-violation DB error into a friendly "has related
// records" response instead of a raw 500 — not needed yet, since nothing
// references client companies today.
func (s *ClientCompanyService) ForceDelete(id uint) error {
	var company models.ClientCompany
	if err := s.db.First(&company, id).Error; err != nil {
		return err
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&company).Association("Clients").Clear(); err != nil {
			return err
		}
		return tx.Delete(&company).Error
	})
	if err != nil {
		return err
	}

	if company.Logo != nil {
		_ = s.uploads.DeleteImage(*company.Logo)
	}
	return nil
}
