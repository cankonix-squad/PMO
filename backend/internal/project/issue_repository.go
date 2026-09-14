package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateIssue(ctx context.Context, i *Issue) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *postgresRepository) FindIssueByID(ctx context.Context, projectID, issueID, orgID uuid.UUID) (*Issue, error) {
	var i Issue
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", issueID, projectID, orgID).
		First(&i).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrIssueNotFound
	}
	return &i, err
}

func (r *postgresRepository) ListIssues(ctx context.Context, filter IssueListFilter) ([]Issue, int64, error) {
	q := r.db.WithContext(ctx).Model(&Issue{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if filter.AssignedTo != nil {
		q = q.Where("assigned_to = ?", filter.AssignedTo)
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

	var issues []Issue
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&issues).Error
	return issues, total, err
}

func (r *postgresRepository) UpdateIssue(ctx context.Context, i *Issue) error {
	return r.db.WithContext(ctx).Save(i).Error
}

func (r *postgresRepository) DeleteIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", issueID, projectID, orgID).
		Delete(&Issue{}).Error
}
