package rbac

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrRoleNotFound is returned when a role lookup yields no result.
var ErrRoleNotFound = errors.New("role not found")

// Repository defines data access for RBAC entities.
type Repository interface {
	// Roles
	FindRoleByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindRoleByCode(ctx context.Context, orgID uuid.UUID, code string) (*Role, error)
	ListRoles(ctx context.Context, orgID uuid.UUID) ([]Role, error)
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// Permissions
	ListPermissions(ctx context.Context) ([]Permission, error)
	SyncRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	// User ↔ Role
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error)

	// Scopes
	SetUserScope(ctx context.Context, scope *UserScope) error
	GetUserScopes(ctx context.Context, userID uuid.UUID) ([]UserScope, error)
	DeleteUserScope(ctx context.Context, userID, scopeID uuid.UUID) error
}

// postgresRepository is the GORM implementation.
type postgresRepository struct {
	db *gorm.DB
}

// NewRepository creates a new RBAC Repository.
func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindRoleByID(ctx context.Context, id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Preload("Permissions").Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoleNotFound
	}
	return &role, err
}

func (r *postgresRepository) FindRoleByCode(ctx context.Context, orgID uuid.UUID, code string) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND code = ?", orgID, code).
		First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoleNotFound
	}
	return &role, err
}

func (r *postgresRepository) ListRoles(ctx context.Context, orgID uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("name").
		Find(&roles).Error
	return roles, err
}

func (r *postgresRepository) CreateRole(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *postgresRepository) UpdateRole(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *postgresRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Role{}, "id = ?", id).Error
}

func (r *postgresRepository) ListPermissions(ctx context.Context) ([]Permission, error) {
	var perms []Permission
	err := r.db.WithContext(ctx).Order("resource, action").Find(&perms).Error
	return perms, err
}

func (r *postgresRepository) SyncRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all existing
		if err := tx.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
			return err
		}
		// Insert new
		for _, pid := range permissionIDs {
			rp := RolePermission{RoleID: roleID, PermissionID: pid}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *postgresRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	ur := UserRole{UserID: userID, RoleID: roleID}
	return r.db.WithContext(ctx).
		Where(UserRole{UserID: userID, RoleID: roleID}).
		FirstOrCreate(&ur).Error
}

func (r *postgresRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&UserRole{}).Error
}

func (r *postgresRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *postgresRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error) {
	var perms []Permission
	err := r.db.WithContext(ctx).
		Distinct("permissions.*").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Find(&perms).Error
	return perms, err
}

func (r *postgresRepository) SetUserScope(ctx context.Context, scope *UserScope) error {
	return r.db.WithContext(ctx).Create(scope).Error
}

func (r *postgresRepository) GetUserScopes(ctx context.Context, userID uuid.UUID) ([]UserScope, error) {
	var scopes []UserScope
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&scopes).Error
	return scopes, err
}

func (r *postgresRepository) DeleteUserScope(ctx context.Context, userID, scopeID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", scopeID, userID).
		Delete(&UserScope{}).Error
}
