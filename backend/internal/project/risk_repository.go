package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateRisk(ctx context.Context, risk *Risk) error {
	return r.db.WithContext(ctx).Create(risk).Error
}

func (r *postgresRepository) FindRiskByID(ctx context.Context, projectID, riskID, orgID uuid.UUID) (*Risk, error) {
	var risk Risk
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", riskID, projectID, orgID).
		First(&risk).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRiskNotFound
	}
	return &risk, err
}

func (r *postgresRepository) ListRisks(ctx context.Context, filter RiskListFilter) ([]Risk, int64, error) {
	q := r.db.WithContext(ctx).Model(&Risk{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if filter.OwnedBy != nil {
		q = q.Where("owned_by = ?", filter.OwnedBy)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", s, s)
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

	var risks []Risk
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&risks).Error
	return risks, total, err
}

func (r *postgresRepository) UpdateRisk(ctx context.Context, risk *Risk) error {
	return r.db.WithContext(ctx).Save(risk).Error
}

func (r *postgresRepository) DeleteRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", riskID, projectID, orgID).
		Delete(&Risk{}).Error
}
