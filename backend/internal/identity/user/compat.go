// Package user is the production identity/user boundary.
// This compatibility facade lets new callers migrate independently from the
// legacy core package without changing API behavior.
package user

import coreuser "github.com/harmanto-49/cankora/internal/core/user"

type Service = coreuser.Service
type Repository = coreuser.Repository
type Handler = coreuser.Handler
type Profile = coreuser.Profile
type CreateUserRequest = coreuser.CreateUserRequest
type UpdateUserRequest = coreuser.UpdateUserRequest
type UpdateProfileRequest = coreuser.UpdateProfileRequest

var ErrEmailTaken = coreuser.ErrEmailTaken
