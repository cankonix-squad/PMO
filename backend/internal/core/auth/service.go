package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/shared/utils"
	"go.uber.org/zap"
)

// Max consecutive failed logins before account lock.
const maxLoginAttempts = 5
const lockDuration = 15 * time.Minute

// ErrAccountLocked is returned when the user account is locked.
var ErrAccountLocked = errors.New("account is locked due to too many failed login attempts")

// ErrInvalidCredentials is returned when email/password do not match.
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrAccountInactive is returned when the user is disabled.
var ErrAccountInactive = errors.New("account is inactive")

// Service handles authentication business logic.
type Service struct {
	repo         Repository
	rbacRepo     rbac.Repository
	tokenService *TokenService
	cfg          *config.Config
	log          *zap.Logger
}

// NewService creates a new auth Service.
func NewService(repo Repository, rbacRepo rbac.Repository, tokenSvc *TokenService, cfg *config.Config, log *zap.Logger) *Service {
	return &Service{
		repo:         repo,
		rbacRepo:     rbacRepo,
		tokenService: tokenSvc,
		cfg:          cfg,
		log:          log,
	}
}

// Login authenticates a user by email + password and returns a token pair.
func (s *Service) Login(ctx context.Context, req *LoginRequest, userAgent, ip string) (*TokenPair, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth: login lookup: %w", err)
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	// Check account lock
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := utils.CheckPassword(req.Password, user.PasswordHash); err != nil {
		_ = s.repo.IncrementLoginFailed(ctx, user.ID)
		if user.LoginFailed+1 >= maxLoginAttempts {
			until := time.Now().Add(lockDuration)
			_ = s.repo.LockUser(ctx, user.ID, until)
			s.log.Warn("account locked due to failed login attempts",
				zap.String("user_id", user.ID.String()),
				zap.String("email", user.Email),
			)
		}
		return nil, ErrInvalidCredentials
	}

	// Fetch roles for the user
	roles, err := s.fetchUserRoles(ctx, user.ID)
	if err != nil {
		s.log.Error("auth: failed to fetch user roles", zap.Error(err))
		roles = []string{}
	}

	pair, err := s.tokenService.GenerateTokenPair(user, roles)
	if err != nil {
		return nil, fmt.Errorf("auth: generate tokens: %w", err)
	}

	// Persist session
	session := &UserSession{
		ID:         uuid.New(),
		UserID:     user.ID,
		JTI:        extractJTI(pair.AccessToken, s.tokenService),
		RefreshJTI: extractRefreshJTI(pair.RefreshToken, s.tokenService),
		UserAgent:  userAgent,
		IPAddress:  ip,
		ExpiresAt:  time.Now().Add(time.Duration(s.cfg.Auth.RefreshExpDays) * 24 * time.Hour),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		s.log.Error("auth: create session failed", zap.Error(err))
		// Non-fatal — token still valid
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	s.log.Info("user logged in",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
		zap.String("ip", ip),
	)

	return pair, nil
}

// Logout revokes the user's current session identified by the access token JTI.
func (s *Service) Logout(ctx context.Context, jti string) error {
	if err := s.repo.RevokeSessionByJTI(ctx, jti); err != nil {
		return fmt.Errorf("auth: logout revoke session: %w", err)
	}
	return nil
}

// LogoutAll revokes all active sessions for a user.
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllUserSessions(ctx, userID)
}

// Refresh validates a refresh token and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshToken, userAgent, ip string) (*TokenPair, error) {
	claims, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check refresh session is not revoked
	session, err := s.repo.FindSessionByRefreshJTI(ctx, claims.JTI)
	if err != nil || session.IsRevoked || session.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenInvalid
	}

	user, err := s.repo.FindUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	// Revoke old session
	_ = s.repo.RevokeSessionByJTI(ctx, session.JTI)

	roles, err := s.fetchUserRoles(ctx, user.ID)
	if err != nil {
		roles = []string{}
	}

	pair, err := s.tokenService.GenerateTokenPair(user, roles)
	if err != nil {
		return nil, fmt.Errorf("auth: refresh generate tokens: %w", err)
	}

	newSession := &UserSession{
		ID:         uuid.New(),
		UserID:     user.ID,
		JTI:        extractJTI(pair.AccessToken, s.tokenService),
		RefreshJTI: extractRefreshJTI(pair.RefreshToken, s.tokenService),
		UserAgent:  userAgent,
		IPAddress:  ip,
		ExpiresAt:  time.Now().Add(time.Duration(s.cfg.Auth.RefreshExpDays) * 24 * time.Hour),
	}
	if err := s.repo.CreateSession(ctx, newSession); err != nil {
		s.log.Error("auth: create refresh session failed", zap.Error(err))
	}

	return pair, nil
}

// ChangePassword changes a user's password after verifying the current one.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := utils.CheckPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("auth: hash new password: %w", err)
	}

	user.PasswordHash = hash
	user.MustChangePwd = false
	return s.repo.UpdateUser(ctx, user)
}

// ForgotPassword generates a reset token and returns it (caller is responsible for sending the email).
func (s *Service) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		// Return no error to avoid email enumeration
		return "", nil
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return "", fmt.Errorf("auth: generate reset token: %w", err)
	}
	tokenStr := hex.EncodeToString(rawToken)
	hash := sha256.Sum256(rawToken)
	tokenHash := hex.EncodeToString(hash[:])

	resetToken := &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := s.repo.CreateResetToken(ctx, resetToken); err != nil {
		return "", fmt.Errorf("auth: store reset token: %w", err)
	}

	return tokenStr, nil
}

// ResetPassword verifies a reset token and sets a new password.
func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	rawBytes, err := hex.DecodeString(req.Token)
	if err != nil {
		return errors.New("invalid token format")
	}
	hash := sha256.Sum256(rawBytes)
	tokenHash := hex.EncodeToString(hash[:])

	resetToken, err := s.repo.FindResetToken(ctx, tokenHash)
	if err != nil {
		return errors.New("token is invalid or has expired")
	}

	user, err := s.repo.FindUserByID(ctx, resetToken.UserID)
	if err != nil {
		return err
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}

	user.PasswordHash = newHash
	user.MustChangePwd = false
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	_ = s.repo.MarkResetTokenUsed(ctx, resetToken.ID)
	_ = s.repo.RevokeAllUserSessions(ctx, user.ID)

	return nil
}

// IsSessionRevoked checks if the JTI associated with a token has been revoked.
func (s *Service) IsSessionRevoked(ctx context.Context, jti string) bool {
	session, err := s.repo.FindSessionByJTI(ctx, jti)
	if err != nil {
		return true // treat as revoked if session not found
	}
	return session.IsRevoked
}

// --- helpers ---

func (s *Service) fetchUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	if s.rbacRepo == nil {
		return []string{}, nil
	}

	roles, err := s.rbacRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}
	return codes, nil
}

func extractJTI(tokenStr string, ts *TokenService) string {
	c, err := ts.ValidateAccessToken(tokenStr)
	if err != nil {
		return uuid.New().String()
	}
	return c.JTI
}

func extractRefreshJTI(tokenStr string, ts *TokenService) string {
	c, err := ts.ValidateRefreshToken(tokenStr)
	if err != nil {
		return uuid.New().String()
	}
	return c.JTI
}
