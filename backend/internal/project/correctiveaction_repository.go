package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateCorrectiveAction(ctx context.Context, ca *CorrectiveAction) error {
	return r.db.WithContext(ctx).Create(ca).Error
}

func (r *postgresRepository) FindCorrectiveActionByID(ctx context.Context, projectID, caID, orgID uuid.UUID) (*CorrectiveAction, error) {
	var ca CorrectiveAction
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", caID, projectID, orgID).
		First(&ca).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCorrectiveActionNotFound
	}
	return &ca, err
}

func (r *postgresRepository) ListCorrectiveActions(ctx context.Context, filter CorrectiveActionListFilter) ([]CorrectiveAction, int64, error) {
	q := r.db.WithContext(ctx).Model(&CorrectiveAction{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.SourceType != "" {
		q = q.Where("source_type = ?", filter.SourceType)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ? OR deviation ILIKE ?", s, s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
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

	var cas []CorrectiveAction
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&cas).Error
	return cas, total, err
}

func (r *postgresRepository) UpdateCorrectiveAction(ctx context.Context, ca *CorrectiveAction) error {
	return r.db.WithContext(ctx).Save(ca).Error
}

func (r *postgresRepository) DeleteCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", caID, projectID, orgID).
		Delete(&CorrectiveAction{}).Error
}
