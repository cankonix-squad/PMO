package testkit

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/platform/http/router"
	"go.uber.org/zap"
)

// PermissionRepo deliberately implements only the calls used by HTTP guards.
// Any unexpected repository access panics instead of silently passing a test.
type PermissionRepo struct {
	rbac.Repository
	Permissions []rbac.Permission
	Err         error
}

func (r PermissionRepo) GetUserPermissions(context.Context, uuid.UUID) ([]rbac.Permission, error) {
	return r.Permissions, r.Err
}
func (r PermissionRepo) GetUserRoles(context.Context, uuid.UUID) ([]rbac.Role, error) {
	return nil, r.Err
}

type SessionRepo struct {
	auth.Repository
	Revoked bool
}

func (r SessionRepo) FindSessionByJTI(context.Context, string) (*auth.UserSession, error) {
	return &auth.UserSession{IsRevoked: r.Revoked}, nil
}

func AuthenticatedRouter(perms []rbac.Permission, revoked bool, configure ...func(*router.Dependencies)) (*gin.Engine, string, error) {
	var token string
	var err error
	engine := Router(func(d *router.Dependencies) {
		d.RBACRepo = PermissionRepo{Permissions: perms}
		d.AuthSvc = auth.NewService(SessionRepo{Revoked: revoked}, d.RBACRepo, d.TokenSvc, d.Config, zap.NewNop())
		pair, e := d.TokenSvc.GenerateTokenPair(&auth.User{ID: uuid.New(), OrganizationID: uuid.New()}, nil)
		err = e
		if e == nil {
			token = pair.AccessToken
		}
		for _, fn := range configure {
			fn(d)
		}
	})
	return engine, token, err
}
