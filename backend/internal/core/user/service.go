package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/shared/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrEmailTaken is returned when the requested email is already in use.
var ErrEmailTaken = errors.New("email already in use")

// Service handles user management business logic.
type Service struct {
	repo     Repository
	rbacRepo rbac.Repository
	log      *zap.Logger
}

// NewService creates a new user Service.
func NewService(repo Repository, rbacRepo rbac.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, rbacRepo: rbacRepo, log: log}
}

// List returns a paginated list of users for an organization.
func (s *Service) List(ctx context.Context, filter UserListFilter) ([]Profile, int64, error) {
	users, total, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	profiles := make([]Profile, 0, len(users))
	for _, u := range users {
		profiles = append(profiles, toProfile(u))
	}
	return profiles, total, nil
}

// GetByID returns a single user profile with roles.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p := toProfile(*user)

	roles, err := s.rbacRepo.GetUserRoles(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user: get roles: %w", err)
	}
	for _, r := range roles {
		p.Roles = append(p.Roles, RoleRef{ID: r.ID, Code: r.Code, Name: r.Name})
	}

	return &p, nil
}

// GetByIDForOrg returns a user profile only when it belongs to the organization.
func (s *Service) GetByIDForOrg(ctx context.Context, id, orgID uuid.UUID) (*Profile, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.OrganizationID != orgID {
		return nil, auth.ErrUserNotFound
	}
	return p, nil
}

// Create provisions a new user account.
func (s *Service) Create(ctx context.Context, req *CreateUserRequest, createdBy *uuid.UUID) (*Profile, error) {
	// Check email uniqueness
	_, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, auth.ErrUserNotFound) {
		return nil, fmt.Errorf("user: check email: %w", err)
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user: hash password: %w", err)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	u := &auth.User{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		OrgUnitID:      req.OrgUnitID,
		EmployeeID:     req.EmployeeID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		PasswordHash:   hash,
		Phone:          req.Phone,
		JobTitle:       req.JobTitle,
		IsActive:       isActive,
		MustChangePwd:  true,
		CreatedBy:      createdBy,
	}

	if err := s.repo.CreateUser(ctx, u); err != nil {
		return nil, fmt.Errorf("user: create: %w", err)
	}

	// Assign roles
	for _, roleID := range req.RoleIDs {
		if err := s.rbacRepo.AssignRoleToUser(ctx, u.ID, roleID); err != nil {
			s.log.Error("user: failed to assign role",
				zap.String("user_id", u.ID.String()),
				zap.String("role_id", roleID.String()),
				zap.Error(err),
			)
			return nil, fmt.Errorf("user: assign role: %w", err)
		}
	}

	return s.GetByID(ctx, u.ID)
}

// Update modifies an existing user's profile and role assignments.
func (s *Service) Update(ctx context.Context, id, orgID uuid.UUID, req *UpdateUserRequest) (*Profile, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.OrganizationID != orgID {
		return nil, auth.ErrUserNotFound
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.OrgUnitID != nil {
		user.OrgUnitID = req.OrgUnitID
	}
	if req.EmployeeID != nil {
		user.EmployeeID = req.EmployeeID
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.JobTitle != "" {
		user.JobTitle = req.JobTitle
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("user: update: %w", err)
	}

	// Sync roles if provided
	if req.RoleIDs != nil {
		existing, err := s.rbacRepo.GetUserRoles(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("user: get roles: %w", err)
		}
		existingMap := make(map[uuid.UUID]bool)
		for _, r := range existing {
			existingMap[r.ID] = true
		}
		requested := make(map[uuid.UUID]bool)
		for _, rid := range req.RoleIDs {
			requested[rid] = true
		}
		// Remove roles not in new set
		for _, r := range existing {
			if !requested[r.ID] {
				if err := s.rbacRepo.RemoveRoleFromUser(ctx, id, r.ID); err != nil {
					return nil, fmt.Errorf("user: remove role: %w", err)
				}
			}
		}
		// Add new roles
		for _, rid := range req.RoleIDs {
			if !existingMap[rid] {
				if err := s.rbacRepo.AssignRoleToUser(ctx, id, rid); err != nil {
					return nil, fmt.Errorf("user: assign role: %w", err)
				}
			}
		}
	}

	return s.GetByID(ctx, id)
}

// UpdateProfile allows a user to update their own non-sensitive profile fields.
func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, req *UpdateProfileRequest) (*Profile, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.JobTitle != "" {
		user.JobTitle = req.JobTitle
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("user: update profile: %w", err)
	}

	return s.GetByID(ctx, id)
}

// Deactivate soft-disables a user account.
func (s *Service) Deactivate(ctx context.Context, id, orgID uuid.UUID) error {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return err
	}
	if user.OrganizationID != orgID {
		return auth.ErrUserNotFound
	}
	user.IsActive = false
	return s.repo.UpdateUser(ctx, user)
}

// --- helpers ---

func toProfile(u auth.User) Profile {
	return Profile{
		ID:             u.ID,
		OrganizationID: u.OrganizationID,
		OrgUnitID:      u.OrgUnitID,
		EmployeeID:     u.EmployeeID,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		FullName:       u.FirstName + " " + u.LastName,
		Email:          u.Email,
		Phone:          u.Phone,
		JobTitle:       u.JobTitle,
		AvatarURL:      u.AvatarURL,
		IsActive:       u.IsActive,
		MustChangePwd:  u.MustChangePwd,
		LastLoginAt:    u.LastLoginAt,
		CreatedAt:      u.CreatedAt,
	}
}

// Repository is the user-specific data access interface that extends auth.Repository.
type Repository interface {
	auth.Repository
	ListUsers(ctx context.Context, filter UserListFilter) ([]auth.User, int64, error)
}

// postgresRepository wraps auth.Repository and adds user-listing capabilities.
type postgresRepository struct {
	auth.Repository
	db *gorm.DB
}

// NewRepository creates a user Repository backed by PostgreSQL.
func NewRepository(db *gorm.DB, authRepo auth.Repository) Repository {
	return &postgresRepository{Repository: authRepo, db: db}
}

func (r *postgresRepository) ListUsers(ctx context.Context, filter UserListFilter) ([]auth.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&auth.User{}).
		Where("organization_id = ?", filter.OrganizationID)

	if filter.OrgUnitID != nil {
		query = query.Where("org_unit_id = ?", filter.OrgUnitID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", s, s, s)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var users []auth.User
	err := query.Order("first_name, last_name").Limit(pageSize).Offset(offset).Find(&users).Error
	return users, total, err
}
