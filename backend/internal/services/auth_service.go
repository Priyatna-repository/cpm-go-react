package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidGoogleToken  = errors.New("invalid google token")
	ErrGoogleUserNotFound  = errors.New("no user found with this google account")
	ErrGoogleNotConfigured = errors.New("google sign-in is not configured")
)

// dummyHash lets Login run a bcrypt compare even when the email doesn't
// exist, so a wrong-email response doesn't return measurably faster than a
// wrong-password one.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("cpm-timing-mitigation"), bcrypt.DefaultCost)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string
	AccessExpiresIn  int
	RefreshToken     string
	RefreshExpiresAt time.Time
}

func (s *AuthService) Login(email, password string) (*models.User, *TokenPair, error) {
	var user models.User
	if err := s.db.Preload("Role").Where("LOWER(email) = LOWER(?)", email).First(&user).Error; err != nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(&user)
	if err != nil {
		return nil, nil, err
	}
	return &user, pair, nil
}

func (s *AuthService) Refresh(rawToken string) (*models.User, *TokenPair, error) {
	hash := hashToken(rawToken)
	now := time.Now()

	result := s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, now).
		Update("revoked_at", now)
	if result.Error != nil {
		return nil, nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil, ErrInvalidRefreshToken
	}

	var stored models.RefreshToken
	if err := s.db.Where("token_hash = ?", hash).First(&stored).Error; err != nil {
		return nil, nil, ErrInvalidRefreshToken
	}

	var user models.User
	if err := s.db.Preload("Role").First(&user, stored.UserID).Error; err != nil {
		return nil, nil, ErrInvalidRefreshToken
	}

	pair, err := s.issueTokenPair(&user)
	if err != nil {
		return nil, nil, err
	}
	return &user, pair, nil
}

// GoogleLogin verifies a Google Identity Services ID token and logs in an
// EXISTING user matched by google_id (or, failing that, by a *verified*
// email — which backfills google_id). It never creates a new user, matching
// the reference app's behavior.
func (s *AuthService) GoogleLogin(ctx context.Context, idTokenStr string) (*models.User, *TokenPair, error) {
	if s.cfg.GoogleClientID == "" {
		return nil, nil, ErrGoogleNotConfigured
	}

	payload, err := idtoken.Validate(ctx, idTokenStr, s.cfg.GoogleClientID)
	if err != nil {
		return nil, nil, ErrInvalidGoogleToken
	}

	googleID, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	if googleID == "" || email == "" {
		return nil, nil, ErrInvalidGoogleToken
	}

	var user models.User
	err = s.db.Preload("Role").Where("google_id = ?", googleID).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if !emailVerified {
			return nil, nil, ErrGoogleUserNotFound
		}
		if err := s.db.Preload("Role").Where("LOWER(email) = LOWER(?)", email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, ErrGoogleUserNotFound
			}
			return nil, nil, err
		}
		if err := s.db.Model(&user).Update("google_id", googleID).Error; err != nil {
			return nil, nil, err
		}
	case err != nil:
		return nil, nil, err
	}

	pair, err := s.issueTokenPair(&user)
	if err != nil {
		return nil, nil, err
	}
	return &user, pair, nil
}

func (s *AuthService) Logout(rawToken string) error {
	hash := hashToken(rawToken)
	return s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", time.Now()).Error
}

func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) ParseAccessToken(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	return claims, nil
}

func (s *AuthService) issueTokenPair(user *models.User) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(s.cfg.AccessTokenTTL)

	claims := Claims{
		UserID: user.ID,
		Role:   user.Role.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
		},
	}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	rawRefresh, err := randomToken()
	if err != nil {
		return nil, err
	}
	refreshExp := now.Add(s.cfg.RefreshTokenTTL)

	if err := s.db.Create(&models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: refreshExp,
	}).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      access,
		AccessExpiresIn:  int(s.cfg.AccessTokenTTL.Seconds()),
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
