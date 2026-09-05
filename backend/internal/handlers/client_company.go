package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClientCompanyHandler struct {
	companies *services.ClientCompanyService
	users     *services.UserLookupService
}

func NewClientCompanyHandler(companies *services.ClientCompanyService, users *services.UserLookupService) *ClientCompanyHandler {
	return &ClientCompanyHandler{companies: companies, users: users}
}

type ClientUserResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ClientCompanyResponse struct {
	ID         uint                 `json:"id"`
	Name       string               `json:"name"`
	Logo       *string              `json:"logo,omitempty"`
	Address    *string              `json:"address,omitempty"`
	PostalCode *string              `json:"postal_code,omitempty"`
	City       *string              `json:"city,omitempty"`
	CountryID  *uint                `json:"country_id,omitempty"`
	CurrencyID *uint                `json:"currency_id,omitempty"`
	Email      *string              `json:"email,omitempty"`
	Phone      *string              `json:"phone,omitempty"`
	Web        *string              `json:"web,omitempty"`
	Clients    []ClientUserResponse `json:"clients"`
	ArchivedAt *time.Time           `json:"archived_at,omitempty"`
}

func toClientCompanyResponse(m *models.ClientCompany) ClientCompanyResponse {
	clients := make([]ClientUserResponse, 0, len(m.Clients))
	for _, u := range m.Clients {
		clients = append(clients, ClientUserResponse{ID: u.ID, Name: u.Name})
	}
	return ClientCompanyResponse{
		ID: m.ID, Name: m.Name, Logo: m.Logo, Address: m.Address, PostalCode: m.PostalCode,
		City: m.City, CountryID: m.CountryID, CurrencyID: m.CurrencyID, Email: m.Email,
		Phone: m.Phone, Web: m.Web, Clients: clients, ArchivedAt: m.ArchivedAt,
	}
}

type ClientCompanyListResponse struct {
	Data []ClientCompanyResponse `json:"data"`
	Meta struct {
		CurrentPage int   `json:"current_page"`
		LastPage    int   `json:"last_page"`
		Total       int64 `json:"total"`
	} `json:"meta"`
}

// ListClientCompanies godoc
// @Summary      List client companies
// @Description  Paginated, searchable. Requires client_company.view.
// @Tags         client-companies
// @Security     BearerAuth
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        search    query     string  false  "Search by name or email"
// @Param        archived  query     bool    false  "Show archived instead of active"
// @Success      200  {object}  ClientCompanyListResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/client-companies [get]
func (h *ClientCompanyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	archived := c.Query("archived") == "1" || c.Query("archived") == "true"

	result, err := h.companies.List(services.ListClientCompaniesParams{
		Page: page, PerPage: 12, Search: c.Query("search"), Archived: archived,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load client companies"})
		return
	}

	resp := ClientCompanyListResponse{Data: make([]ClientCompanyResponse, 0, len(result.Items))}
	for i := range result.Items {
		resp.Data = append(resp.Data, toClientCompanyResponse(&result.Items[i]))
	}
	resp.Meta.CurrentPage = result.CurrentPage
	resp.Meta.LastPage = result.LastPage
	resp.Meta.Total = result.Total
	c.JSON(http.StatusOK, resp)
}

// GetClientCompany godoc
// @Summary      Get a client company
// @Tags         client-companies
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Client company ID"
// @Success      200  {object}  ClientCompanyResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/client-companies/{id} [get]
func (h *ClientCompanyHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	company, err := h.companies.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client company not found"})
		return
	}
	c.JSON(http.StatusOK, toClientCompanyResponse(company))
}

type ClientCompanyRequest struct {
	Name       string  `form:"name" binding:"required"`
	Address    *string `form:"address"`
	PostalCode *string `form:"postal_code"`
	City       *string `form:"city"`
	CountryID  *uint   `form:"country_id"`
	CurrencyID *uint   `form:"currency_id"`
	Email      *string `form:"email"`
	Phone      *string `form:"phone"`
	Web        *string `form:"web"`
	ClientIDs  []uint  `form:"client_ids"`
}

// CreateClientCompany godoc
// @Summary      Create a client company
// @Description  multipart/form-data body (logo file is optional). Requires client_company.create.
// @Tags         client-companies
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Success      201  {object}  ClientCompanyResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/client-companies [post]
func (h *ClientCompanyHandler) Create(c *gin.Context) {
	var req ClientCompanyRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	logo, _ := c.FormFile("logo")

	company, err := h.companies.Create(services.ClientCompanyInput{
		Name: req.Name, Address: req.Address, PostalCode: req.PostalCode, City: req.City,
		CountryID: req.CountryID, CurrencyID: req.CurrencyID, Email: req.Email, Phone: req.Phone,
		Web: req.Web, ClientIDs: req.ClientIDs,
	}, logo)
	if err != nil {
		respondClientCompanyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toClientCompanyResponse(company))
}

// UpdateClientCompany godoc
// @Summary      Update a client company
// @Description  multipart/form-data body (logo file is optional). Requires client_company.edit.
// @Tags         client-companies
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      int  true  "Client company ID"
// @Success      200  {object}  ClientCompanyResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/client-companies/{id} [put]
func (h *ClientCompanyHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ClientCompanyRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	logo, _ := c.FormFile("logo")

	company, err := h.companies.Update(id, services.ClientCompanyInput{
		Name: req.Name, Address: req.Address, PostalCode: req.PostalCode, City: req.City,
		CountryID: req.CountryID, CurrencyID: req.CurrencyID, Email: req.Email, Phone: req.Phone,
		Web: req.Web, ClientIDs: req.ClientIDs,
	}, logo)
	if err != nil {
		respondClientCompanyError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClientCompanyResponse(company))
}

// ArchiveClientCompany godoc
// @Summary      Archive a client company
// @Tags         client-companies
// @Security     BearerAuth
// @Param        id   path  int  true  "Client company ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/client-companies/{id} [delete]
func (h *ClientCompanyHandler) Archive(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.companies.Archive(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "client company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive client company"})
		return
	}
	c.Status(http.StatusNoContent)
}

// RestoreClientCompany godoc
// @Summary      Restore an archived client company
// @Tags         client-companies
// @Security     BearerAuth
// @Param        id   path  int  true  "Client company ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/client-companies/{id}/restore [post]
func (h *ClientCompanyHandler) Restore(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.companies.Restore(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "client company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore client company"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ForceDeleteClientCompany godoc
// @Summary      Permanently delete a client company
// @Tags         client-companies
// @Security     BearerAuth
// @Param        id   path  int  true  "Client company ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/client-companies/{id}/force [delete]
func (h *ClientCompanyHandler) ForceDelete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.companies.ForceDelete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "client company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete client company"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListClientUsers godoc
// @Summary      List users with the client role
// @Description  For the "assign clients" picker on the ClientCompany form.
// @Tags         client-companies
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   ClientUserResponse
// @Failure      401  {object}  map[string]string
// @Router       /api/v1/lookups/client-users [get]
// @Param        company_id  query  int  false  "Exclude: keep this company's own members visible too"
func (h *ClientCompanyHandler) ListClientUsers(c *gin.Context) {
	var excludeCompanyID *uint
	if raw := c.Query("company_id"); raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
			v := uint(id)
			excludeCompanyID = &v
		}
	}

	users, err := h.users.ListClientUsers(excludeCompanyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load client users"})
		return
	}

	resp := make([]ClientUserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, ClientUserResponse{ID: u.ID, Name: u.Name})
	}
	c.JSON(http.StatusOK, resp)
}

func respondClientCompanyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrFileTooLarge):
		c.JSON(http.StatusBadRequest, gin.H{"error": "logo file is too large (max 2MB)"})
	case errors.Is(err, services.ErrUnsupportedImage):
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type (use jpg, png, or gif)"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "client company not found"})
	case errors.Is(err, services.ErrClientAlreadyLinked):
		c.JSON(http.StatusConflict, gin.H{"error": "one or more selected clients already belong to a different client company"})
	case errors.Is(err, services.ErrClientCompanyHasProjects):
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete a client company that has projects"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save client company"})
	}
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// ListInternalUsers godoc
// @Summary      List non-client users
// @Description  For the "grant project access" picker. Requires project.view.
// @Tags         client-companies
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   ClientUserResponse
// @Failure      401  {object}  map[string]string
// @Router       /api/v1/lookups/internal-users [get]
func (h *ClientCompanyHandler) ListInternalUsers(c *gin.Context) {
	users, err := h.users.ListInternalUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}

	resp := make([]ClientUserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, ClientUserResponse{ID: u.ID, Name: u.Name})
	}
	c.JSON(http.StatusOK, resp)
}
