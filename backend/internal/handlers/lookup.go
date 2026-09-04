package handlers

import (
	"net/http"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type LookupHandler struct {
	lookups *services.LookupService
}

func NewLookupHandler(lookups *services.LookupService) *LookupHandler {
	return &LookupHandler{lookups: lookups}
}

type CountryResponse struct {
	ID   uint   `json:"id" example:"1"`
	Name string `json:"name" example:"Indonesia"`
}

// ListCountries godoc
// @Summary      List countries
// @Description  Static reference data, for populating dropdowns.
// @Tags         lookups
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   CountryResponse
// @Failure      401  {object}  map[string]string
// @Router       /api/v1/lookups/countries [get]
func (h *LookupHandler) ListCountries(c *gin.Context) {
	countries, err := h.lookups.ListCountries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load countries"})
		return
	}

	resp := make([]CountryResponse, 0, len(countries))
	for _, country := range countries {
		resp = append(resp, CountryResponse{ID: country.ID, Name: country.Name})
	}
	c.JSON(http.StatusOK, resp)
}

type CurrencyResponse struct {
	ID       uint   `json:"id" example:"1"`
	Name     string `json:"name" example:"Rupiah"`
	Code     string `json:"code" example:"IDR"`
	Symbol   string `json:"symbol" example:"Rp"`
	Decimals int16  `json:"decimals" example:"2"`
}

// ListCurrencies godoc
// @Summary      List currencies
// @Description  Static reference data, for populating dropdowns.
// @Tags         lookups
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   CurrencyResponse
// @Failure      401  {object}  map[string]string
// @Router       /api/v1/lookups/currencies [get]
func (h *LookupHandler) ListCurrencies(c *gin.Context) {
	currencies, err := h.lookups.ListCurrencies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load currencies"})
		return
	}

	resp := make([]CurrencyResponse, 0, len(currencies))
	for _, currency := range currencies {
		resp = append(resp, CurrencyResponse{
			ID: currency.ID, Name: currency.Name, Code: currency.Code,
			Symbol: currency.Symbol, Decimals: currency.Decimals,
		})
	}
	c.JSON(http.StatusOK, resp)
}
