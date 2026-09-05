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

type LabelHandler struct {
	labels *services.LabelService
}

func NewLabelHandler(labels *services.LabelService) *LabelHandler {
	return &LabelHandler{labels: labels}
}

type LabelResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Slug       string     `json:"slug"`
	Type       string     `json:"type"`
	Color      *string    `json:"color,omitempty"`
	Icon       *string    `json:"icon,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

func toLabelResponse(m *models.Label) LabelResponse {
	return LabelResponse{ID: m.ID, Name: m.Name, Slug: m.Slug, Type: m.Type, Color: m.Color, Icon: m.Icon, ArchivedAt: m.ArchivedAt}
}

type LabelListResponse struct {
	Data []LabelResponse `json:"data"`
	Meta struct {
		CurrentPage int   `json:"current_page"`
		LastPage    int   `json:"last_page"`
		Total       int64 `json:"total"`
	} `json:"meta"`
}

// ListLabels godoc
// @Summary      List labels
// @Description  Paginated, searchable, filterable by type. Requires labels.view.
// @Tags         labels
// @Security     BearerAuth
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        search    query     string  false  "Search by name"
// @Param        type      query     string  false  "Filter by label type"
// @Param        archived  query     bool    false  "Show archived instead of active"
// @Success      200  {object}  LabelListResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/labels [get]
func (h *LabelHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	archived := c.Query("archived") == "1" || c.Query("archived") == "true"

	result, err := h.labels.List(services.ListLabelsParams{
		Page: page, PerPage: 12, Search: c.Query("search"), Type: c.Query("type"), Archived: archived,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load labels"})
		return
	}

	resp := LabelListResponse{Data: make([]LabelResponse, 0, len(result.Items))}
	for i := range result.Items {
		resp.Data = append(resp.Data, toLabelResponse(&result.Items[i]))
	}
	resp.Meta.CurrentPage = result.CurrentPage
	resp.Meta.LastPage = result.LastPage
	resp.Meta.Total = result.Total
	c.JSON(http.StatusOK, resp)
}

// GetLabel godoc
// @Summary      Get a label
// @Tags         labels
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Label ID"
// @Success      200  {object}  LabelResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/labels/{id} [get]
func (h *LabelHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	label, err := h.labels.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
		return
	}
	c.JSON(http.StatusOK, toLabelResponse(label))
}

type LabelRequest struct {
	Name  string  `json:"name" binding:"required"`
	Slug  string  `json:"slug"`
	Type  string  `json:"type" binding:"required"`
	Color string  `json:"color" binding:"required"`
	Icon  *string `json:"icon"`
}

// CreateLabel godoc
// @Summary      Create a label
// @Description  Requires labels.create.
// @Tags         labels
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      LabelRequest  true  "Label"
// @Success      201  {object}  LabelResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/labels [post]
func (h *LabelHandler) Create(c *gin.Context) {
	var req LabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type, and color are required"})
		return
	}

	label, err := h.labels.Create(services.LabelInput{Name: req.Name, Slug: req.Slug, Type: req.Type, Color: req.Color, Icon: req.Icon})
	if err != nil {
		respondLabelError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toLabelResponse(label))
}

// UpdateLabel godoc
// @Summary      Update a label
// @Description  Requires labels.edit.
// @Tags         labels
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int           true  "Label ID"
// @Param        request  body      LabelRequest  true  "Label"
// @Success      200  {object}  LabelResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/labels/{id} [put]
func (h *LabelHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req LabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type, and color are required"})
		return
	}

	label, err := h.labels.Update(id, services.LabelInput{Name: req.Name, Slug: req.Slug, Type: req.Type, Color: req.Color, Icon: req.Icon})
	if err != nil {
		respondLabelError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabelResponse(label))
}

// ArchiveLabel godoc
// @Summary      Archive a label
// @Tags         labels
// @Security     BearerAuth
// @Param        id   path  int  true  "Label ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/labels/{id} [delete]
func (h *LabelHandler) Archive(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.labels.Archive(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive label"})
		return
	}
	c.Status(http.StatusNoContent)
}

// RestoreLabel godoc
// @Summary      Restore an archived label
// @Tags         labels
// @Security     BearerAuth
// @Param        id   path  int  true  "Label ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/labels/{id}/restore [post]
func (h *LabelHandler) Restore(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.labels.Restore(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restore label"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ForceDeleteLabel godoc
// @Summary      Permanently delete a label
// @Tags         labels
// @Security     BearerAuth
// @Param        id   path  int  true  "Label ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/labels/{id}/force [delete]
func (h *LabelHandler) ForceDelete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.labels.ForceDelete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete label"})
		return
	}
	c.Status(http.StatusNoContent)
}

func respondLabelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidColor):
		c.JSON(http.StatusBadRequest, gin.H{"error": "color must be a hex code like #3B82F6"})
	case errors.Is(err, services.ErrDuplicateSlug):
		c.JSON(http.StatusConflict, gin.H{"error": "a label with this slug already exists"})
	case errors.Is(err, services.ErrLabelInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete a label that is assigned to a project"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "label not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save label"})
	}
}
