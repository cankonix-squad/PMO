// Package auth is the production boundary for identity authentication.
//
// The aliases temporarily bridge the legacy core implementation while callers
// are migrated. New code should import this package; the bridge can be removed
// once the legacy package has no consumers.
package auth

import (
	coreauth "github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"go.uber.org/zap"
)

type User = coreauth.User
type Claims = coreauth.Claims
type TokenPair = coreauth.TokenPair
type LoginRequest = coreauth.LoginRequest
type RefreshRequest = coreauth.RefreshRequest
type Service = coreauth.Service
type TokenService = coreauth.TokenService
type Repository = coreauth.Repository

var ErrInvalidCredentials = coreauth.ErrInvalidCredentials
var ErrAccountLocked = coreauth.ErrAccountLocked
var ErrAccountInactive = coreauth.ErrAccountInactive
var ErrTokenInvalid = coreauth.ErrTokenInvalid
var ErrTokenExpired = coreauth.ErrTokenExpired

func NewTokenService(cfg *config.AuthConfig) *TokenService {
	return coreauth.NewTokenService(cfg)
}

func NewService(repo Repository, rbacRepo rbac.Repository, tokenSvc *TokenService, cfg *config.Config, log *zap.Logger) *Service {
	return coreauth.NewService(repo, rbacRepo, tokenSvc, cfg, log)
}
