package services

import (
	"errors"
	"mime/multipart"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRoleScope string

const (
	UserScopeInternal UserRoleScope = "internal" // manager + team member
	UserScopeClient   UserRoleScope = "client"
)

var internalRoleNames = []string{models.RoleManager, models.RoleTeamMember}

var (
	ErrInvalidRole      = errors.New("invalid role for this user type")
	ErrDuplicateEmail   = errors.New("a user with this email already exists")
	ErrCannotSelfModify = errors.New("you cannot archive or delete your own account")
)

type UserService struct {
	db      *gorm.DB
	uploads *UploadService
}

func NewUserService(db *gorm.DB, uploads *UploadService) *UserService {
	return &UserService{db: db, uploads: uploads}
}

func (s *UserService) allowedRoleNames(scope UserRoleScope) []string {
	if scope == UserScopeClient {
		return []string{models.RoleClient}
	}
	return internalRoleNames
}

func (s *UserService) validRole(scope UserRoleScope, roleName string, role *models.Role) error {
	ok := false
	for _, n := range s.allowedRoleNames(scope) {
		if n == roleName {
			ok = true
			break
		}
	}
	if !ok {
		return ErrInvalidRole
	}
	return s.db.Where("name = ?", roleName).First(role).Error
}

type ListUsersParams struct {
	Page     int
	PerPage  int
	Search   string
	Archived bool
	Scope    UserRoleScope
}

type UserPage struct {
	Items       []models.User
	CurrentPage int
	LastPage    int
	Total       int64
}

func (s *UserService) List(p ListUsersParams) (*UserPage, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 12
	}

	query := s.db.Model(&models.User{}).Preload("Role").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ?", s.allowedRoleNames(p.Scope))

	if p.Archived {
		query = query.Where("users.archived_at IS NOT NULL")
	} else {
		query = query.Where("users.archived_at IS NULL")
	}
	if p.Search != "" {
		like := "%" + p.Search + "%"
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []models.User
	offset := (p.Page - 1) * p.PerPage
	if err := query.Order("users.name").Limit(p.PerPage).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	lastPage := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if lastPage < 1 {
		lastPage = 1
	}
	return &UserPage{Items: items, CurrentPage: p.Page, LastPage: lastPage, Total: total}, nil
}

// scopedFirst loads a user by id, but only if their role belongs to scope.
// Get/Update/Archive/Restore/ForceDelete all resolve their target through
// this instead of a bare id lookup, so a caller permissioned for one scope
// (e.g. client_user.*) can never reach a user from the other scope (e.g. an
// internal Manager/Admin account) — the same boundary List already enforces
// via its roles.name IN (...) join.
func (s *UserService) scopedFirst(scope UserRoleScope, id uint) (*models.User, error) {
	var user models.User
	err := s.db.Preload("Role").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ? AND users.id = ?", s.allowedRoleNames(scope), id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) Get(scope UserRoleScope, id uint) (*models.User, error) {
	return s.scopedFirst(scope, id)
}

type UserInput struct {
	Name     string
	Email    string
	Password string // empty on Update means "keep current password"
	RoleName string
	JobTitle *string
	Phone    *string
	Address  *string
}

func (s *UserService) Create(scope UserRoleScope, input UserInput, avatar *multipart.FileHeader) (*models.User, error) {
	var role models.Role
	if err := s.validRole(scope, input.RoleName, &role); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name: input.Name, Email: input.Email, PasswordHash: string(hash),
		RoleID: role.ID, JobTitle: input.JobTitle, Phone: input.Phone, Address: input.Address,
	}

	if avatar != nil {
		path, err := s.uploads.SaveImage(avatar, "avatars")
		if err != nil {
			return nil, err
		}
		user.Avatar = &path
	}

	if err := s.db.Create(&user).Error; err != nil {
		if user.Avatar != nil {
			_ = s.uploads.DeleteImage(*user.Avatar)
		}
		if isDuplicateKey(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	return s.Get(scope, user.ID)
}

func (s *UserService) Update(scope UserRoleScope, id uint, input UserInput, avatar *multipart.FileHeader) (*models.User, error) {
	user, err := s.scopedFirst(scope, id)
	if err != nil {
		return nil, err
	}

	var role models.Role
	if err := s.validRole(scope, input.RoleName, &role); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"name": input.Name, "email": input.Email, "role_id": role.ID,
		"job_title": input.JobTitle, "phone": input.Phone, "address": input.Address,
	}
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = string(hash)
	}

	var newAvatar, oldAvatar *string
	if avatar != nil {
		path, err := s.uploads.SaveImage(avatar, "avatars")
		if err != nil {
			return nil, err
		}
		newAvatar = &path
		oldAvatar = user.Avatar
		updates["avatar"] = path
	}

	if err := s.db.Model(user).Updates(updates).Error; err != nil {
		if newAvatar != nil {
			_ = s.uploads.DeleteImage(*newAvatar)
		}
		if isDuplicateKey(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	if oldAvatar != nil {
		_ = s.uploads.DeleteImage(*oldAvatar)
	}
	return s.Get(scope, id)
}

func (s *UserService) Archive(scope UserRoleScope, id, callerID uint) error {
	if id == callerID {
		return ErrCannotSelfModify
	}
	if _, err := s.scopedFirst(scope, id); err != nil {
		return err
	}
	result := s.db.Model(&models.User{}).Where("id = ?", id).Update("archived_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *UserService) Restore(scope UserRoleScope, id uint) error {
	if _, err := s.scopedFirst(scope, id); err != nil {
		return err
	}
	result := s.db.Model(&models.User{}).Where("id = ?", id).Update("archived_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ForceDelete permanently removes a user. Cleans up rows in other tables
// that reference this user directly (raw SQL rather than GORM
// associations, since User doesn't declare the inverse side of every
// many2many it participates in) rather than letting the DB reject the
// delete on a foreign-key violation.
func (s *UserService) ForceDelete(scope UserRoleScope, id, callerID uint) error {
	if id == callerID {
		return ErrCannotSelfModify
	}
	user, err := s.scopedFirst(scope, id)
	if err != nil {
		return err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&models.RefreshToken{}).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM client_company_user WHERE user_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM project_user_access WHERE user_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE projects SET client_user_id = NULL WHERE client_user_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(user).Error
	})
	if err != nil {
		return err
	}
	if user.Avatar != nil {
		_ = s.uploads.DeleteImage(*user.Avatar)
	}
	return nil
}
