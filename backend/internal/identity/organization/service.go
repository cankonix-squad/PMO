package organization

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Repository defines persistence operations for org units.
type Repository interface {
	ListOrgUnits(orgID uuid.UUID, includeInactive bool) ([]OrgUnit, error)
	GetOrgUnit(orgID, unitID uuid.UUID) (*OrgUnit, error)
	CreateOrgUnit(unit *OrgUnit) error
	UpdateOrgUnit(unit *OrgUnit) error
	DeleteOrgUnit(orgID, unitID uuid.UUID) error
}

type postgresRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListOrgUnits(orgID uuid.UUID, includeInactive bool) ([]OrgUnit, error) {
	var units []OrgUnit
	q := r.db.Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if !includeInactive {
		q = q.Where("is_active = true")
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&units).Error; err != nil {
		return nil, fmt.Errorf("orgunit: list: %w", err)
	}
	return units, nil
}

func (r *postgresRepository) GetOrgUnit(orgID, unitID uuid.UUID) (*OrgUnit, error) {
	var unit OrgUnit
	err := r.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", unitID, orgID).
		First(&unit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("orgunit: get: %w", err)
	}
	return &unit, nil
}

func (r *postgresRepository) CreateOrgUnit(unit *OrgUnit) error {
	if err := r.db.Create(unit).Error; err != nil {
		return fmt.Errorf("orgunit: create: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateOrgUnit(unit *OrgUnit) error {
	if err := r.db.Save(unit).Error; err != nil {
		return fmt.Errorf("orgunit: update: %w", err)
	}
	return nil
}

func (r *postgresRepository) DeleteOrgUnit(orgID, unitID uuid.UUID) error {
	res := r.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", unitID, orgID).
		Delete(&OrgUnit{})
	if res.Error != nil {
		return fmt.Errorf("orgunit: delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// Service provides business logic for org unit management.
type Service struct {
	repo Repository
	log  *zap.Logger
}

func NewService(db *gorm.DB, log *zap.Logger) *Service {
	return &Service{repo: NewRepository(db), log: log}
}

func (s *Service) ListOrgUnits(orgID uuid.UUID, includeInactive bool) ([]OrgUnit, error) {
	return s.repo.ListOrgUnits(orgID, includeInactive)
}

func (s *Service) GetOrgUnit(orgID, unitID uuid.UUID) (*OrgUnit, error) {
	return s.repo.GetOrgUnit(orgID, unitID)
}

func (s *Service) CreateOrgUnit(orgID uuid.UUID, req CreateOrgUnitRequest) (*OrgUnit, error) {
	unit := &OrgUnit{
		OrganizationID: orgID,
		ParentID:       req.ParentID,
		Code:           req.Code,
		Name:           req.Name,
		Level:          req.Level,
		SortOrder:      req.SortOrder,
		IsActive:       true,
	}
	if err := s.repo.CreateOrgUnit(unit); err != nil {
		return nil, err
	}
	return unit, nil
}

func (s *Service) UpdateOrgUnit(orgID, unitID uuid.UUID, req UpdateOrgUnitRequest) (*OrgUnit, error) {
	unit, err := s.repo.GetOrgUnit(orgID, unitID)
	if err != nil {
		return nil, err
	}

	if req.Code != "" {
		unit.Code = req.Code
	}
	if req.Name != "" {
		unit.Name = req.Name
	}
	if req.Level >= LevelKementerian && req.Level <= LevelUnit {
		unit.Level = req.Level
	}
	if req.ParentID != nil {
		unit.ParentID = req.ParentID
	}
	if req.HeadUserID != nil {
		unit.HeadUserID = req.HeadUserID
	}
	if req.IsActive != nil {
		unit.IsActive = *req.IsActive
	}
	if req.SortOrder != 0 {
		unit.SortOrder = req.SortOrder
	}

	if err := s.repo.UpdateOrgUnit(unit); err != nil {
		return nil, err
	}
	return unit, nil
}

func (s *Service) DeleteOrgUnit(orgID, unitID uuid.UUID) error {
	return s.repo.DeleteOrgUnit(orgID, unitID)
}
