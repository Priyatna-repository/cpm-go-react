package services

import (
	"errors"
	"strings"
	"unicode"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrInvalidColor  = errors.New("color must be a hex code like #3B82F6")
	ErrDuplicateSlug = errors.New("a label with this slug already exists")
	ErrLabelInUse    = errors.New("cannot delete a label that is assigned to a project")
)

type LabelService struct {
	db *gorm.DB
}

func NewLabelService(db *gorm.DB) *LabelService {
	return &LabelService{db: db}
}

type ListLabelsParams struct {
	Page     int
	PerPage  int
	Search   string
	Type     string
	Archived bool
}

type LabelPage struct {
	Items       []models.Label
	CurrentPage int
	LastPage    int
	Total       int64
}

func (s *LabelService) List(p ListLabelsParams) (*LabelPage, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 12
	}

	query := s.db.Model(&models.Label{})
	if p.Archived {
		query = query.Where("archived_at IS NOT NULL")
	} else {
		query = query.Where("archived_at IS NULL")
	}
	if p.Type != "" {
		query = query.Where("type = ?", p.Type)
	}
	if p.Search != "" {
		query = query.Where("name ILIKE ?", "%"+p.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.Label
	offset := (p.Page - 1) * p.PerPage
	if err := query.Order("type, name").Limit(p.PerPage).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	lastPage := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if lastPage < 1 {
		lastPage = 1
	}

	return &LabelPage{Items: items, CurrentPage: p.Page, LastPage: lastPage, Total: total}, nil
}

func (s *LabelService) Get(id uint) (*models.Label, error) {
	var label models.Label
	if err := s.db.First(&label, id).Error; err != nil {
		return nil, err
	}
	return &label, nil
}

type LabelInput struct {
	Name  string
	Slug  string
	Type  string
	Color string
	Icon  *string
}

func (s *LabelService) Create(input LabelInput) (*models.Label, error) {
	if !isHexColor(input.Color) {
		return nil, ErrInvalidColor
	}

	slug := input.Slug
	if slug == "" {
		slug = slugify(input.Name)
	}
	color := input.Color

	label := models.Label{Name: input.Name, Slug: slug, Type: input.Type, Color: &color, Icon: input.Icon}
	if err := s.db.Create(&label).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateSlug
		}
		return nil, err
	}
	return &label, nil
}

func (s *LabelService) Update(id uint, input LabelInput) (*models.Label, error) {
	if !isHexColor(input.Color) {
		return nil, ErrInvalidColor
	}

	var label models.Label
	if err := s.db.First(&label, id).Error; err != nil {
		return nil, err
	}

	slug := input.Slug
	if slug == "" {
		slug = slugify(input.Name)
	}

	updates := map[string]interface{}{
		"name": input.Name, "slug": slug, "type": input.Type, "color": input.Color, "icon": input.Icon,
	}
	if err := s.db.Model(&label).Updates(updates).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateSlug
		}
		return nil, err
	}
	return s.Get(id)
}

func (s *LabelService) Archive(id uint) error {
	result := s.db.Model(&models.Label{}).Where("id = ?", id).Update("archived_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *LabelService) Restore(id uint) error {
	result := s.db.Model(&models.Label{}).Where("id = ?", id).Update("archived_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *LabelService) ForceDelete(id uint) error {
	result := s.db.Delete(&models.Label{}, id)
	if result.Error != nil {
		if isForeignKeyViolation(result.Error) {
			return ErrLabelInUse
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func isHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, r := range s[1:] {
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
			return false
		}
	}
	return true
}

func slugify(name string) string {
	var b strings.Builder
	prevUnderscore := true
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevUnderscore = false
		} else if !prevUnderscore {
			b.WriteRune('_')
			prevUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}
