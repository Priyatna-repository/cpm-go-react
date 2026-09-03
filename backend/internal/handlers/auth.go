package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/middleware"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	auth *services.AuthService
	cfg  *config.Config
}

func NewAuthHandler(auth *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required" example:"password"`
}

type UserResponse struct {
	ID    uint   `json:"id" example:"1"`
	Name  string `json:"name" example:"Admin User"`
	Email string `json:"email" example:"admin@example.com"`
	Role  string `json:"role" example:"admin"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type" example:"Bearer"`
	ExpiresIn   int          `json:"expires_in" example:"900"`
	User        UserResponse `json:"user"`
}

// Login godoc
// @Summary      Log in
// @Description  Validates credentials and returns an access token; sets an httpOnly refresh-token cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Credentials"
// @Success      200      {object}  LoginResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, pair, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	c.JSON(http.StatusOK, toLoginResponse(user, pair))
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Rotates the refresh-token cookie and returns a new access token.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  LoginResponse
// @Failure      401  {object}  map[string]string
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	raw, err := c.Cookie(refreshCookieName)
	if err != nil || raw == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	user, pair, err := h.auth.Refresh(raw)
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			h.clearRefreshCookie(c)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	c.JSON(http.StatusOK, toLoginResponse(user, pair))
}

// Logout godoc
// @Summary      Log out
// @Description  Revokes the current refresh token and clears its cookie.
// @Tags         auth
// @Success      204
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	if raw, err := c.Cookie(refreshCookieName); err == nil && raw != "" {
		_ = h.auth.Logout(raw)
	}
	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

// Me godoc
// @Summary      Current user
// @Description  Returns the authenticated user's profile.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  map[string]string
// @Router       /api/v1/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, _ := c.Get(middleware.ContextUserID)
	roleVal, _ := c.Get(middleware.ContextRole)

	id, _ := userIDVal.(uint)
	role, _ := roleVal.(string)

	user, err := h.auth.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, UserResponse{ID: user.ID, Name: user.Name, Email: user.Email, Role: role})
}

func toLoginResponse(user *models.User, pair *services.TokenPair) LoginResponse {
	return LoginResponse{
		AccessToken: pair.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   pair.AccessExpiresIn,
		User: UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role.Name,
		},
	}
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, token, maxAge, "/api/v1/auth", "", h.cfg.AppEnv == "production", true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, "/api/v1/auth", "", h.cfg.AppEnv == "production", true)
}
