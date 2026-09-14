package user

import (
	"context"

	"github.com/harmanto-49/cankora/internal/identity/auth"
)

// Repository is the user-specific data access interface that extends auth.Repository.
type Repository interface {
	auth.Repository
	ListUsers(ctx context.Context, filter UserListFilter) ([]auth.User, int64, error)
}
