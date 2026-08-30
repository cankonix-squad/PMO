package projectcategory

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service handles business logic for project categories.
type Service struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewService(db *gorm.DB, log *zap.Logger) *Service {
	return &Service{db: db, log: log}
}

// List returns all categories for an org. Pass includeInactive=true to include
// soft-deleted rows (deleted_at IS NOT NULL).
func (s *Service) List(orgID uuid.UUID, includeInactive bool) ([]ProjectCategory, error) {
	q := s.db.Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if !includeInactive {
		q = q.Where("is_active = true")
	}
	var list []ProjectCategory
	if err := q.Order("sort_order, name").Find(&list).Error; err != nil {
		s.log.Error("project_category: list", zap.Error(err))
		return nil, err
	}
	return list, nil
}

// GetByID fetches a single non-deleted category scoped to the org.
func (s *Service) GetByID(orgID, id uuid.UUID) (*ProjectCategory, error) {
	var cat ProjectCategory
	err := s.db.
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).
		First(&cat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &cat, err
}

// Create inserts a new category. Returns ErrCodeTaken on duplicate code.
func (s *Service) Create(orgID, createdBy uuid.UUID, req CreateRequest) (*ProjectCategory, error) {
	cat := ProjectCategory{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Code:           strings.TrimSpace(req.Code),
		Name:           strings.TrimSpace(req.Name),
		Description:    req.Description,
		IsActive:       true,
		SortOrder:      0,
		CreatedBy:      &createdBy,
	}
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}

	if err := s.db.Create(&cat).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeTaken
		}
		s.log.Error("project_category: create", zap.Error(err))
		return nil, err
	}
	return &cat, nil
}

// Update applies partial updates to an existing category.
func (s *Service) Update(orgID, id uuid.UUID, req UpdateRequest) (*ProjectCategory, error) {
	cat, err := s.GetByID(orgID, id)
	if err != nil {
		return nil, err
	}

	if req.Code != nil {
		cat.Code = strings.TrimSpace(*req.Code)
	}
	if req.Name != nil {
		cat.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		cat.Description = req.Description
	}
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}

	if err := s.db.Save(cat).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeTaken
		}
		s.log.Error("project_category: update", zap.Error(err))
		return nil, err
	}
	return cat, nil
}

// Delete soft-deletes a category (sets deleted_at).
func (s *Service) Delete(orgID, id uuid.UUID) error {
	res := s.db.
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).
		Delete(&ProjectCategory{})
	if res.Error != nil {
		s.log.Error("project_category: delete", zap.Error(res.Error))
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// isUniqueViolation detects PostgreSQL unique constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
