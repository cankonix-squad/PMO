package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateVendor(ctx context.Context, v *Vendor) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *postgresRepository) FindVendorByID(ctx context.Context, id, orgID uuid.UUID) (*Vendor, error) {
	var v Vendor
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVendorNotFound
	}
	return &v, err
}

func (r *postgresRepository) ListVendors(ctx context.Context, filter VendorListFilter) ([]Vendor, int64, error) {
	q := r.db.WithContext(ctx).Model(&Vendor{}).
		Where("organization_id = ?", filter.OrganizationID)

	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.IsActive != nil {
		q = q.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("name ILIKE ? OR legal_name ILIKE ? OR tax_id ILIKE ?", s, s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var vendors []Vendor
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&vendors).Error
	return vendors, total, err
}

func (r *postgresRepository) UpdateVendor(ctx context.Context, v *Vendor) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *postgresRepository) DeleteVendor(ctx context.Context, id, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, orgID).
		Delete(&Vendor{}).Error
}
