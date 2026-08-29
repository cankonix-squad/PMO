package portfolio

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

// Program represents a master program (umbrella) that groups projects.
type Program struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Code           string         `gorm:"size:100;not null;uniqueIndex:uniq_program_org_code" json:"code"`
	Name           string         `gorm:"size:300;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	FiscalYear     *int           `json:"fiscal_year,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	SortOrder      int            `gorm:"default:0" json:"sort_order"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      string         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      string         `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateProgramRequest struct {
	Code        string `json:"code" binding:"required,max=100"`
	Name        string `json:"name" binding:"required,max=300"`
	Description string `json:"description"`
	FiscalYear  *int   `json:"fiscal_year"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateProgramRequest struct {
	Code        string `json:"code" binding:"max=100"`
	Name        string `json:"name" binding:"max=300"`
	Description string `json:"description"`
	FiscalYear  *int   `json:"fiscal_year"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// ---------------------------------------------------------------------------
// Repository
// ---------------------------------------------------------------------------

type Repository interface {
	List(orgID uuid.UUID, includeInactive bool) ([]Program, error)
	Get(orgID, id uuid.UUID) (*Program, error)
	Create(p *Program) error
	Update(p *Program) error
	Delete(orgID, id uuid.UUID) error
}

type postgresRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) List(orgID uuid.UUID, includeInactive bool) ([]Program, error) {
	var list []Program
	q := r.db.Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if !includeInactive {
		q = q.Where("is_active = true")
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("portfolio: list: %w", err)
	}
	return list, nil
}

func (r *postgresRepo) Get(orgID, id uuid.UUID) (*Program, error) {
	var p Program
	err := r.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("portfolio: get: %w", err)
	}
	return &p, nil
}

func (r *postgresRepo) Create(p *Program) error {
	if err := r.db.Create(p).Error; err != nil {
		return fmt.Errorf("portfolio: create: %w", err)
	}
	return nil
}

func (r *postgresRepo) Update(p *Program) error {
	if err := r.db.Save(p).Error; err != nil {
		return fmt.Errorf("portfolio: update: %w", err)
	}
	return nil
}

func (r *postgresRepo) Delete(orgID, id uuid.UUID) error {
	res := r.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).Delete(&Program{})
	if res.Error != nil {
		return fmt.Errorf("portfolio: delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type Service struct {
	repo Repository
	log  *zap.Logger
}

func NewService(db *gorm.DB, log *zap.Logger) *Service {
	return &Service{repo: NewRepository(db), log: log}
}

func (s *Service) List(orgID uuid.UUID, includeInactive bool) ([]Program, error) {
	return s.repo.List(orgID, includeInactive)
}

func (s *Service) Get(orgID, id uuid.UUID) (*Program, error) {
	return s.repo.Get(orgID, id)
}

func (s *Service) Create(orgID, createdBy uuid.UUID, req CreateProgramRequest) (*Program, error) {
	p := &Program{
		OrganizationID: orgID,
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		FiscalYear:     req.FiscalYear,
		SortOrder:      req.SortOrder,
		IsActive:       true,
		CreatedBy:      &createdBy,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Update(orgID, id uuid.UUID, req UpdateProgramRequest) (*Program, error) {
	p, err := s.repo.Get(orgID, id)
	if err != nil {
		return nil, err
	}
	if req.Code != "" {
		p.Code = req.Code
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	p.Description = req.Description
	if req.FiscalYear != nil {
		p.FiscalYear = req.FiscalYear
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if req.SortOrder != 0 {
		p.SortOrder = req.SortOrder
	}
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(orgID, id uuid.UUID) error {
	return s.repo.Delete(orgID, id)
}
