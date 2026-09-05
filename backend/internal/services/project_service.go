package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrProjectClientRequired = errors.New("choose either a client company or an individual client, not both or neither")
	ErrDuplicateProjectName  = errors.New("a project with this name already exists")
	ErrManageAccessRequired  = errors.New("changing the project's client requires the project.manage_access permission")
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// HasAccess reports whether a (non-admin) user can access a specific
// project: via client-company membership, being the project's direct
// individual client, or an explicit project_user_access grant. Callers
// should special-case admin themselves (this does no bypass logic).
func (s *ProjectService) HasAccess(userID, projectID uint) (bool, error) {
	var count int64
	err := s.db.Raw(`
		SELECT count(*) FROM projects p
		WHERE p.id = ?
		AND (
			EXISTS (SELECT 1 FROM client_company_user ccu WHERE ccu.client_company_id = p.client_company_id AND ccu.user_id = ?)
			OR p.client_user_id = ?
			OR EXISTS (SELECT 1 FROM project_user_access pua WHERE pua.project_id = p.id AND pua.user_id = ?)
		)`, projectID, userID, userID, userID).Scan(&count).Error
	return count > 0, err
}

type ListProjectsParams struct {
	Page     int
	PerPage  int
	Search   string
	Archived bool
	UserID   uint
	IsAdmin  bool
}

type ProjectPage struct {
	Items       []models.Project
	CurrentPage int
	LastPage    int
	Total       int64
}

func (s *ProjectService) List(p ListProjectsParams) (*ProjectPage, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 12
	}

	query := s.db.Model(&models.Project{}).
		Preload("ClientCompany").Preload("ClientUser").Preload("TypeLabel").Preload("StatusLabels")

	if p.Archived {
		query = query.Where("archived_at IS NOT NULL")
	} else {
		query = query.Where("archived_at IS NULL")
	}
	if p.Search != "" {
		like := "%" + p.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}
	if !p.IsAdmin {
		query = query.Where(
			"(EXISTS (SELECT 1 FROM client_company_user ccu WHERE ccu.client_company_id = projects.client_company_id AND ccu.user_id = ?) OR projects.client_user_id = ? OR EXISTS (SELECT 1 FROM project_user_access pua WHERE pua.project_id = projects.id AND pua.user_id = ?))",
			p.UserID, p.UserID, p.UserID,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.Project
	offset := (p.Page - 1) * p.PerPage
	if err := query.Order("name").Limit(p.PerPage).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	lastPage := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if lastPage < 1 {
		lastPage = 1
	}
	return &ProjectPage{Items: items, CurrentPage: p.Page, LastPage: lastPage, Total: total}, nil
}

func (s *ProjectService) Get(id uint) (*models.Project, error) {
	var project models.Project
	err := s.db.Preload("ClientCompany").Preload("ClientUser").Preload("TypeLabel").
		Preload("StatusLabels").Preload("Users").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

type ProjectInput struct {
	Name            string
	Description     *string
	ClientCompanyID *uint
	ClientUserID    *uint
	StartDate       *time.Time
	EndDate         *time.Time
	BudgetEstimate  *float64
	TypeLabelID     *uint
	StatusLabelIDs  []uint
	UserIDs         []uint
}

func (s *ProjectService) Create(input ProjectInput, creatorID uint) (*models.Project, error) {
	if err := validateProjectClient(input.ClientCompanyID, input.ClientUserID); err != nil {
		return nil, err
	}

	code, err := s.generateCode(input.ClientCompanyID)
	if err != nil {
		return nil, err
	}

	project := models.Project{
		Code: code, Name: input.Name, Description: input.Description,
		ClientCompanyID: input.ClientCompanyID, ClientUserID: input.ClientUserID,
		StartDate: input.StartDate, EndDate: input.EndDate,
		BudgetEstimate: input.BudgetEstimate, TypeLabelID: input.TypeLabelID,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&project).Error; err != nil {
			return err
		}
		if err := syncProjectStatusLabels(tx, &project, input.StatusLabelIDs); err != nil {
			return err
		}
		// The creator always gets access, independent of the UserIDs they
		// picked in the form — otherwise a manager creating a project for a
		// client they aren't personally a member of would immediately fail
		// RequireProjectAccess on their own next edit, with only an admin
		// able to grant it back via UpdateAccess.
		return syncProjectUsers(tx, &project, append(input.UserIDs, creatorID))
	})
	if err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateProjectName
		}
		return nil, err
	}
	return s.Get(project.ID)
}

func (s *ProjectService) Update(id uint, input ProjectInput, canManageAccess bool) (*models.Project, error) {
	if err := validateProjectClient(input.ClientCompanyID, input.ClientUserID); err != nil {
		return nil, err
	}

	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return nil, err
	}

	if !canManageAccess && clientChanged(&project, input) {
		return nil, ErrManageAccessRequired
	}

	updates := map[string]interface{}{
		"name": input.Name, "description": input.Description,
		"client_company_id": input.ClientCompanyID, "client_user_id": input.ClientUserID,
		"start_date": input.StartDate, "end_date": input.EndDate,
		"budget_estimate": input.BudgetEstimate, "type_label_id": input.TypeLabelID,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&project).Updates(updates).Error; err != nil {
			return err
		}
		return syncProjectStatusLabels(tx, &project, input.StatusLabelIDs)
	})

	if err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateProjectName
		}
		return nil, err
	}
	return s.Get(id)
}

func (s *ProjectService) UpdateAccess(id uint, userIDs []uint) (*models.Project, error) {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	if err := syncProjectUsers(s.db, &project, userIDs); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *ProjectService) Archive(id uint) error {
	result := s.db.Model(&models.Project{}).Where("id = ?", id).Update("archived_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *ProjectService) Restore(id uint) error {
	result := s.db.Model(&models.Project{}).Where("id = ?", id).Update("archived_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *ProjectService) ForceDelete(id uint) error {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&project).Association("Users").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&project).Association("StatusLabels").Clear(); err != nil {
			return err
		}
		return tx.Delete(&project).Error
	})
}

func syncProjectStatusLabels(tx *gorm.DB, project *models.Project, labelIDs []uint) error {
	var labels []models.Label
	if len(labelIDs) > 0 {
		if err := tx.Where("id IN ?", labelIDs).Find(&labels).Error; err != nil {
			return err
		}
	}
	return tx.Model(project).Association("StatusLabels").Replace(labels)
}

func syncProjectUsers(tx *gorm.DB, project *models.Project, userIDs []uint) error {
	var users []models.User
	if len(userIDs) > 0 {
		if err := tx.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return err
		}
	}
	return tx.Model(project).Association("Users").Replace(users)
}

func validateProjectClient(companyID, userID *uint) error {
	if (companyID == nil) == (userID == nil) {
		return ErrProjectClientRequired
	}
	return nil
}

func clientChanged(project *models.Project, input ProjectInput) bool {
	return !uintPtrEqual(project.ClientCompanyID, input.ClientCompanyID) || !uintPtrEqual(project.ClientUserID, input.ClientUserID)
}

func uintPtrEqual(a, b *uint) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (s *ProjectService) generateCode(clientCompanyID *uint) (string, error) {
	prefix := "PRJ"
	if clientCompanyID != nil {
		var company models.ClientCompany
		if err := s.db.Select("name").First(&company, *clientCompanyID).Error; err == nil {
			prefix = projectCodePrefix(company.Name)
		}
	}

	var lastCode string
	err := s.db.Model(&models.Project{}).
		Where("code LIKE ?", prefix+"-%").
		Order("code DESC").
		Limit(1).
		Pluck("code", &lastCode).Error
	if err != nil {
		return "", err
	}

	next := 1
	if lastCode != "" {
		parts := strings.Split(lastCode, "-")
		if len(parts) == 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil {
				next = n + 1
			}
		}
	}
	return fmt.Sprintf("%s-%05d", prefix, next), nil
}

func projectCodePrefix(name string) string {
	words := strings.Fields(name)
	if len(words) > 3 {
		words = words[:3]
	}
	var b strings.Builder
	for _, w := range words {
		r := []rune(strings.ToUpper(w))
		if len(r) > 0 {
			b.WriteRune(r[0])
		}
	}
	for b.Len() < 3 {
		b.WriteByte('X')
	}
	return b.String()
}
