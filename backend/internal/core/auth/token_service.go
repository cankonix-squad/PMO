package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/platform/config"
)

// ErrTokenExpired is returned when the JWT is expired.
var ErrTokenExpired = errors.New("token has expired")

// ErrTokenInvalid is returned when the JWT cannot be parsed or verified.
var ErrTokenInvalid = errors.New("token is invalid")

// jwtClaims is the internal JWT claims struct (unexported).
type jwtClaims struct {
	jwt.RegisteredClaims
	Email          string     `json:"email"`
	OrganizationID uuid.UUID  `json:"org_id"`
	OrgUnitID      *uuid.UUID `json:"org_unit_id,omitempty"`
	IsRefresh      bool       `json:"is_refresh,omitempty"`
}

// TokenService handles JWT generation and validation.
type TokenService struct {
	cfg *config.AuthConfig
}

// NewTokenService creates a new TokenService.
func NewTokenService(cfg *config.AuthConfig) *TokenService {
	return &TokenService{cfg: cfg}
}

// GenerateTokenPair creates a new access + refresh token pair for a user.
func (ts *TokenService) GenerateTokenPair(user *User, roles []string) (*TokenPair, error) {
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	accessExpiry := time.Now().Add(time.Duration(ts.cfg.AccessExpMinutes) * time.Minute)
	refreshExpiry := time.Now().Add(time.Duration(ts.cfg.RefreshExpDays) * 24 * time.Hour)

	accessToken, err := ts.generateToken(user, accessJTI, accessExpiry, false)
	if err != nil {
		return nil, fmt.Errorf("auth: generate access token: %w", err)
	}

	refreshToken, err := ts.generateToken(user, refreshJTI, refreshExpiry, true)
	if err != nil {
		return nil, fmt.Errorf("auth: generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		User: UserInfo{
			ID:             user.ID,
			OrganizationID: user.OrganizationID,
			OrgUnitID:      user.OrgUnitID,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Email:          user.Email,
			JobTitle:       user.JobTitle,
			AvatarURL:      user.AvatarURL,
			MustChangePwd:  user.MustChangePwd,
			Roles:          roles,
		},
	}, nil
}

// ValidateAccessToken parses and validates an access token.
// Returns Claims on success.
func (ts *TokenService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return ts.validateToken(tokenStr, false)
}

// ValidateRefreshToken parses and validates a refresh token.
func (ts *TokenService) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return ts.validateToken(tokenStr, true)
}

func (ts *TokenService) generateToken(user *User, jti string, expiry time.Time, isRefresh bool) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiry),
			Issuer:    "cankora",
			ID:        jti,
		},
		Email:          user.Email,
		OrganizationID: user.OrganizationID,
		OrgUnitID:      user.OrgUnitID,
		IsRefresh:      isRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(ts.cfg.JWTSecret))
}

func (ts *TokenService) validateToken(tokenStr string, expectRefresh bool) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(ts.cfg.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	if c.IsRefresh != expectRefresh {
		return nil, ErrTokenInvalid
	}

	userID, err := uuid.Parse(c.Subject)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	return &Claims{
		UserID:         userID,
		Email:          c.Email,
		OrganizationID: c.OrganizationID,
		OrgUnitID:      c.OrgUnitID,
		JTI:            c.ID,
		IsRefresh:      c.IsRefresh,
	}, nil
}
