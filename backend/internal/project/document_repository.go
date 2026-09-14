package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateDocument(ctx context.Context, d *ProjectDocument) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *postgresRepository) FindDocumentByID(ctx context.Context, projectID, documentID, orgID uuid.UUID) (*ProjectDocument, error) {
	// orgID scoping is enforced via parent project ownership check in the
	// service (project_documents has no organization_id column).
	var d ProjectDocument
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", documentID, projectID).
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDocumentNotFound
	}
	return &d, err
}

func (r *postgresRepository) ListDocuments(ctx context.Context, filter DocumentListFilter) ([]ProjectDocument, int64, error) {
	q := r.db.WithContext(ctx).Model(&ProjectDocument{}).
		Where("project_id = ?", filter.ProjectID)

	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("name ILIKE ? OR category ILIKE ?", s, s)
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

	var docs []ProjectDocument
	err := q.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&docs).Error
	return docs, total, err
}

func (r *postgresRepository) UpdateDocument(ctx context.Context, d *ProjectDocument) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *postgresRepository) DeleteDocument(ctx context.Context, projectID, documentID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", documentID, projectID).
		Delete(&ProjectDocument{}).Error
}
