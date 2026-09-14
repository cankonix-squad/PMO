package spatial

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

// Sector (Sektor SDA — e.g. Irigasi, Bendungan, Sungai, Air Baku)
type Sector struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	Code           string         `gorm:"size:100;not null" json:"code"`
	Name           string         `gorm:"size:300;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	SortOrder      int            `gorm:"default:0" json:"sort_order"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      string         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      string         `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// Region (Wilayah administratif — e.g. Provinsi, Kabupaten)
type Region struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ParentID       *uuid.UUID     `gorm:"type:uuid" json:"parent_id,omitempty"`
	Code           string         `gorm:"size:100;not null" json:"code"`
	Name           string         `gorm:"size:300;not null" json:"name"`
	Level          int            `gorm:"default:1" json:"level"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	SortOrder      int            `gorm:"default:0" json:"sort_order"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      string         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      string         `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Children []Region `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// RiverBasin (DAS — Daerah Aliran Sungai)
type RiverBasin struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	RegionID       *uuid.UUID     `gorm:"type:uuid" json:"region_id,omitempty"`
	Code           string         `gorm:"size:100;not null" json:"code"`
	Name           string         `gorm:"size:300;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	AreaKm2        *float64       `gorm:"type:decimal(12,2)" json:"area_km2,omitempty"`
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

type CreateSectorRequest struct {
	Code        string `json:"code" binding:"required,max=100"`
	Name        string `json:"name" binding:"required,max=300"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateSectorRequest struct {
	Code        string `json:"code" binding:"max=100"`
	Name        string `json:"name" binding:"max=300"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

type CreateRegionRequest struct {
	ParentID  *uuid.UUID `json:"parent_id"`
	Code      string     `json:"code" binding:"required,max=100"`
	Name      string     `json:"name" binding:"required,max=300"`
	Level     int        `json:"level" binding:"min=1,max=4"`
	SortOrder int        `json:"sort_order"`
}

type UpdateRegionRequest struct {
	ParentID  *uuid.UUID `json:"parent_id"`
	Code      string     `json:"code" binding:"max=100"`
	Name      string     `json:"name" binding:"max=300"`
	Level     int        `json:"level" binding:"min=0,max=4"`
	IsActive  *bool      `json:"is_active"`
	SortOrder int        `json:"sort_order"`
}

type CreateRiverBasinRequest struct {
	RegionID    *uuid.UUID `json:"region_id"`
	Code        string     `json:"code" binding:"required,max=100"`
	Name        string     `json:"name" binding:"required,max=300"`
	Description string     `json:"description"`
	AreaKm2     *float64   `json:"area_km2"`
	SortOrder   int        `json:"sort_order"`
}

type UpdateRiverBasinRequest struct {
	RegionID    *uuid.UUID `json:"region_id"`
	Code        string     `json:"code" binding:"max=100"`
	Name        string     `json:"name" binding:"max=300"`
	Description string     `json:"description"`
	AreaKm2     *float64   `json:"area_km2"`
	IsActive    *bool      `json:"is_active"`
	SortOrder   int        `json:"sort_order"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type Service struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewService(db *gorm.DB, log *zap.Logger) *Service {
	return &Service{db: db, log: log}
}

// ----- Sector -----

func (s *Service) ListSectors(orgID uuid.UUID, includeInactive bool) ([]Sector, error) {
	var list []Sector
	q := s.db.Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if !includeInactive {
		q = q.Where("is_active = true")
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("spatial: list sectors: %w", err)
	}
	return list, nil
}

func (s *Service) GetSector(orgID, id uuid.UUID) (*Sector, error) {
	var sec Sector
	err := s.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).First(&sec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("spatial: get sector: %w", err)
	}
	return &sec, nil
}

func (s *Service) CreateSector(orgID, createdBy uuid.UUID, req CreateSectorRequest) (*Sector, error) {
	sec := &Sector{
		OrganizationID: orgID,
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		SortOrder:      req.SortOrder,
		IsActive:       true,
		CreatedBy:      &createdBy,
	}
	if err := s.db.Create(sec).Error; err != nil {
		return nil, fmt.Errorf("spatial: create sector: %w", err)
	}
	return sec, nil
}

func (s *Service) UpdateSector(orgID, id uuid.UUID, req UpdateSectorRequest) (*Sector, error) {
	sec, err := s.GetSector(orgID, id)
	if err != nil {
		return nil, err
	}
	if req.Code != "" {
		sec.Code = req.Code
	}
	if req.Name != "" {
		sec.Name = req.Name
	}
	sec.Description = req.Description
	if req.IsActive != nil {
		sec.IsActive = *req.IsActive
	}
	if req.SortOrder != 0 {
		sec.SortOrder = req.SortOrder
	}
	if err := s.db.Save(sec).Error; err != nil {
		return nil, fmt.Errorf("spatial: update sector: %w", err)
	}
	return sec, nil
}

func (s *Service) DeleteSector(orgID, id uuid.UUID) error {
	res := s.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).Delete(&Sector{})
	if res.Error != nil {
		return fmt.Errorf("spatial: delete sector: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ----- Region -----

func (s *Service) ListRegions(orgID uuid.UUID, includeInactive bool) ([]Region, error) {
	var list []Region
	q := s.db.Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if !includeInactive {
		q = q.Where("is_active = true")
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("spatial: list regions: %w", err)
	}
	return list, nil
}

func (s *Service) GetRegion(orgID, id uuid.UUID) (*Region, error) {
	var reg Region
	err := s.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).First(&reg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("spatial: get region: %w", err)
	}
	return &reg, nil
}

func (s *Service) CreateRegion(orgID, createdBy uuid.UUID, req CreateRegionRequest) (*Region, error) {
	level := req.Level
	if level == 0 {
		level = 1
	}
	reg := &Region{
		OrganizationID: orgID,
		ParentID:       req.ParentID,
		Code:           req.Code,
		Name:           req.Name,
		Level:          level,
		SortOrder:      req.SortOrder,
		IsActive:       true,
		CreatedBy:      &createdBy,
	}
	if err := s.db.Create(reg).Error; err != nil {
		return nil, fmt.Errorf("spatial: create region: %w", err)
	}
	return reg, nil
}

func (s *Service) UpdateRegion(orgID, id uuid.UUID, req UpdateRegionRequest) (*Region, error) {
	reg, err := s.GetRegion(orgID, id)
	if err != nil {
		return nil, err
	}
	if req.Code != "" {
		reg.Code = req.Code
	}
	if req.Name != "" {
		reg.Name = req.Name
	}
	if req.ParentID != nil {
		reg.ParentID = req.ParentID
	}
	if req.Level > 0 {
		reg.Level = req.Level
	}
	if req.IsActive != nil {
		reg.IsActive = *req.IsActive
	}
	if req.SortOrder != 0 {
		reg.SortOrder = req.SortOrder
	}
	if err := s.db.Save(reg).Error; err != nil {
		return nil, fmt.Errorf("spatial: update region: %w", err)
	}
	return reg, nil
}

func (s *Service) DeleteRegion(orgID, id uuid.UUID) error {
	res := s.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).Delete(&Region{})
	if res.Error != nil {
		return fmt.Errorf("spatial: delete region: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ----- RiverBasin -----

func (s *Service) ListRiverBasins(orgID uuid.UUID, includeInactive bool) ([]RiverBasin, error) {
	var list []RiverBasin
	q := s.db.Where("organization_id = ? AND deleted_at IS NULL", orgID)
	if !includeInactive {
		q = q.Where("is_active = true")
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("spatial: list river basins: %w", err)
	}
	return list, nil
}

func (s *Service) GetRiverBasin(orgID, id uuid.UUID) (*RiverBasin, error) {
	var rb RiverBasin
	err := s.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).First(&rb).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("spatial: get river basin: %w", err)
	}
	return &rb, nil
}

func (s *Service) CreateRiverBasin(orgID, createdBy uuid.UUID, req CreateRiverBasinRequest) (*RiverBasin, error) {
	rb := &RiverBasin{
		OrganizationID: orgID,
		RegionID:       req.RegionID,
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		AreaKm2:        req.AreaKm2,
		SortOrder:      req.SortOrder,
		IsActive:       true,
		CreatedBy:      &createdBy,
	}
	if err := s.db.Create(rb).Error; err != nil {
		return nil, fmt.Errorf("spatial: create river basin: %w", err)
	}
	return rb, nil
}

func (s *Service) UpdateRiverBasin(orgID, id uuid.UUID, req UpdateRiverBasinRequest) (*RiverBasin, error) {
	rb, err := s.GetRiverBasin(orgID, id)
	if err != nil {
		return nil, err
	}
	if req.Code != "" {
		rb.Code = req.Code
	}
	if req.Name != "" {
		rb.Name = req.Name
	}
	rb.Description = req.Description
	if req.RegionID != nil {
		rb.RegionID = req.RegionID
	}
	if req.AreaKm2 != nil {
		rb.AreaKm2 = req.AreaKm2
	}
	if req.IsActive != nil {
		rb.IsActive = *req.IsActive
	}
	if req.SortOrder != 0 {
		rb.SortOrder = req.SortOrder
	}
	if err := s.db.Save(rb).Error; err != nil {
		return nil, fmt.Errorf("spatial: update river basin: %w", err)
	}
	return rb, nil
}

func (s *Service) DeleteRiverBasin(orgID, id uuid.UUID) error {
	res := s.db.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).Delete(&RiverBasin{})
	if res.Error != nil {
		return fmt.Errorf("spatial: delete river basin: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
