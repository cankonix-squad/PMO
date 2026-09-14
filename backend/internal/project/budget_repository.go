package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateBudget(ctx context.Context, b *ProjectBudget) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *postgresRepository) FindBudgetByID(ctx context.Context, projectID, budgetID, orgID uuid.UUID) (*ProjectBudget, error) {
	// orgID is resolved through the parent project ownership check in the service.
	var b ProjectBudget
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", budgetID, projectID).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBudgetNotFound
	}
	return &b, err
}

func (r *postgresRepository) ListBudgets(ctx context.Context, filter BudgetListFilter) ([]ProjectBudget, int64, error) {
	q := r.db.WithContext(ctx).Model(&ProjectBudget{}).
		Where("project_id = ?", filter.ProjectID)

	if filter.Category != "" {
		q = q.Where("category ILIKE ?", "%"+filter.Category+"%")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("category ILIKE ? OR description ILIKE ?", s, s)
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

	var budgets []ProjectBudget
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&budgets).Error
	return budgets, total, err
}

func (r *postgresRepository) UpdateBudget(ctx context.Context, b *ProjectBudget) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *postgresRepository) DeleteBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", budgetID, projectID).
		Delete(&ProjectBudget{}).Error
}
