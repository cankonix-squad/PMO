package user

import (
	"context"

	"github.com/harmanto-49/cankora/internal/identity/auth"
	"gorm.io/gorm"
)

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
