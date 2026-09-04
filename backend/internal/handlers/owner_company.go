package handlers

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type OwnerCompanyHandler struct {
	company *services.OwnerCompanyService
}

func NewOwnerCompanyHandler(company *services.OwnerCompanyService) *OwnerCompanyHandler {
	return &OwnerCompanyHandler{company: company}
}

type OwnerCompanyResponse struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Logo       *string `json:"logo,omitempty"`
	Address    *string `json:"address,omitempty"`
	PostalCode *string `json:"postal_code,omitempty"`
	City       *string `json:"city,omitempty"`
	CountryID  *uint   `json:"country_id,omitempty"`
	CurrencyID *uint   `json:"currency_id,omitempty"`
	Email      *string `json:"email,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Web        *string `json:"web,omitempty"`
}

func toOwnerCompanyResponse(m *models.OwnerCompany) OwnerCompanyResponse {
	return OwnerCompanyResponse{
		ID: m.ID, Name: m.Name, Logo: m.Logo, Address: m.Address, PostalCode: m.PostalCode,
		City: m.City, CountryID: m.CountryID, CurrencyID: m.CurrencyID,
		Email: m.Email, Phone: m.Phone, Web: m.Web,
	}
}

// GetOwnerCompany godoc
// @Summary      Get owner company
// @Description  Returns the single owner company record. Requires owner_company.view.
// @Tags         owner-company
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  OwnerCompanyResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/owner-company [get]
func (h *OwnerCompanyHandler) Get(c *gin.Context) {
	company, err := h.company.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load owner company"})
		return
	}
	c.JSON(http.StatusOK, toOwnerCompanyResponse(company))
}

type UpdateOwnerCompanyRequest struct {
	Name       string  `form:"name" binding:"required"`
	Address    *string `form:"address"`
	PostalCode *string `form:"postal_code"`
	City       *string `form:"city"`
	CountryID  *uint   `form:"country_id"`
	CurrencyID *uint   `form:"currency_id"`
	Email      *string `form:"email"`
	Phone      *string `form:"phone"`
	Web        *string `form:"web"`
}

// UpdateOwnerCompany godoc
// @Summary      Update owner company
// @Description  multipart/form-data body (logo file is optional). Requires owner_company.edit.
// @Tags         owner-company
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Success      200  {object}  OwnerCompanyResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/owner-company [put]
func (h *OwnerCompanyHandler) Update(c *gin.Context) {
	var req UpdateOwnerCompanyRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	var logo *multipart.FileHeader
	if fh, err := c.FormFile("logo"); err == nil {
		logo = fh
	}

	company, err := h.company.Update(services.UpdateOwnerCompanyInput{
		Name: req.Name, Address: req.Address, PostalCode: req.PostalCode, City: req.City,
		CountryID: req.CountryID, CurrencyID: req.CurrencyID, Email: req.Email, Phone: req.Phone, Web: req.Web,
	}, logo)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrFileTooLarge):
			c.JSON(http.StatusBadRequest, gin.H{"error": "logo file is too large (max 2MB)"})
		case errors.Is(err, services.ErrUnsupportedImage):
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type (use jpg, png, or gif)"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update owner company"})
		}
		return
	}

	c.JSON(http.StatusOK, toOwnerCompanyResponse(company))
}
