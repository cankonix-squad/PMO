package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateContract(ctx context.Context, c *Contract) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrContractNumberTaken
		}
		return err
	}
	return nil
}

func (r *postgresRepository) FindContractByID(ctx context.Context, projectID, contractID, orgID uuid.UUID) (*Contract, error) {
	var c Contract
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", contractID, projectID, orgID).
		Preload("Vendor").
		Preload("Consultant").
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrContractNotFound
	}
	return &c, err
}

func (r *postgresRepository) ListContracts(ctx context.Context, filter ContractListFilter) ([]Contract, int64, error) {
	q := r.db.WithContext(ctx).Model(&Contract{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.VendorID != nil {
		q = q.Where("vendor_id = ?", filter.VendorID)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("contract_number ILIKE ? OR title ILIKE ?", s, s)
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

	var contracts []Contract
	err := q.Preload("Vendor").Preload("Consultant").
		Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&contracts).Error
	return contracts, total, err
}

func (r *postgresRepository) UpdateContract(ctx context.Context, c *Contract) error {
	if err := r.db.WithContext(ctx).Save(c).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrContractNumberTaken
		}
		return err
	}
	return nil
}

func (r *postgresRepository) DeleteContract(ctx context.Context, projectID, contractID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", contractID, projectID, orgID).
		Delete(&Contract{}).Error
}

// countContractsForVendor counts non-deleted contracts referencing a vendor as
// either the provider (vendor_id) or the supervisory consultant
// (consultant_id), within the given organization.
func (r *postgresRepository) countContractsForVendor(ctx context.Context, vendorID, orgID uuid.UUID, out *int64) error {
	return r.db.WithContext(ctx).Model(&Contract{}).
		Where("organization_id = ? AND (vendor_id = ? OR consultant_id = ?)", orgID, vendorID, vendorID).
		Count(out).Error
}
