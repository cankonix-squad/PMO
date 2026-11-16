package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
	"go.uber.org/zap"
)

// Service handles project business logic.
type Service struct {
	repo Repository
	fsm  *workflow.FSM
	log  *zap.Logger

	// storageRoot is the local dev storage root for document files.
	storageRoot string
	// maxFileSize caps uploaded document size in bytes.
	maxFileSize int64
}

// NewService creates a new project Service.
func NewService(repo Repository, fsm *workflow.FSM, log *zap.Logger) *Service {
	return &Service{
		repo:        repo,
		fsm:         fsm,
		log:         log,
		storageRoot: "storage/documents",
		maxFileSize: 20 * 1024 * 1024,
	}
}

// Create creates a new project.
func (s *Service) Create(ctx context.Context, req *CreateProjectRequest, orgID, createdBy uuid.UUID) (*Project, error) {
	if err := s.validateProjectMasterData(ctx, orgID, req.OrgUnitID, req.ProgramID, req.SectorID, req.RegionID, req.RiverBasinID); err != nil {
		return nil, err
	}

	p := &Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		OrgUnitID:      req.OrgUnitID,
		ProgramID:      req.ProgramID,
		SectorID:       req.SectorID,
		RegionID:       req.RegionID,
		RiverBasinID:   req.RiverBasinID,
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		Objectives:     req.Objectives,
		Status:         "DRAFT",
		Priority:       req.Priority,
		Category:       req.Category,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		BudgetTotal:    req.BudgetTotal,
		Currency:       req.Currency,
		ProgressPct:    *req.ProgressPct,
		ManagerID:      req.ManagerID,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Province:       req.Province,
		City:           req.City,
		LocationName:   req.LocationName,
		CreatedBy:      createdBy,
	}
	if p.Priority == "" {
		p.Priority = "MEDIUM"
	}
	if p.Currency == "" {
		p.Currency = "IDR"
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("project: create: %w", err)
	}

	// Auto-add creator as PROJECT_MANAGER
	_ = s.repo.AddTeamMember(ctx, &TeamMember{
		ID:        uuid.New(),
		ProjectID: p.ID,
		UserID:    createdBy,
		Role:      "PROJECT_MANAGER",
	})

	return s.repo.FindByID(ctx, p.ID, orgID)
}

// GetByID returns a project by ID, checking org boundary.
func (s *Service) GetByID(ctx context.Context, id, orgID uuid.UUID) (*Project, error) {
	return s.repo.FindByID(ctx, id, orgID)
}

// List returns paginated projects for an organization.
func (s *Service) List(ctx context.Context, filter ProjectListFilter) ([]Project, int64, error) {
	return s.repo.List(ctx, filter)
}

// Update modifies a project's details.
func (s *Service) Update(ctx context.Context, id, orgID uuid.UUID, req *UpdateProjectRequest) (*Project, error) {
	p, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if err := s.validateProjectMasterData(ctx, orgID, req.OrgUnitID, req.ProgramID, req.SectorID, req.RegionID, req.RiverBasinID); err != nil {
		return nil, err
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Objectives != "" {
		p.Objectives = req.Objectives
	}
	if req.Priority != "" {
		p.Priority = req.Priority
	}
	if req.Category != "" {
		p.Category = req.Category
	}
	if req.StartDate != nil {
		p.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		p.EndDate = req.EndDate
	}
	if req.BudgetTotal != nil {
		p.BudgetTotal = *req.BudgetTotal
	}
	if req.Currency != "" {
		p.Currency = req.Currency
	}
	if req.ManagerID != nil {
		p.ManagerID = req.ManagerID
	}
	if req.ProgressPct != nil {
		p.ProgressPct = *req.ProgressPct
	}
	if req.Latitude != nil {
		p.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		p.Longitude = req.Longitude
	}
	if req.Province != nil {
		p.Province = *req.Province
	}
	if req.City != nil {
		p.City = *req.City
	}
	if req.LocationName != nil {
		p.LocationName = *req.LocationName
	}
	if req.OrgUnitID != nil {
		p.OrgUnitID = req.OrgUnitID
	}
	if req.ProgramID != nil {
		p.ProgramID = req.ProgramID
	}
	if req.SectorID != nil {
		p.SectorID = req.SectorID
	}
	if req.RegionID != nil {
		p.RegionID = req.RegionID
	}
	if req.RiverBasinID != nil {
		p.RiverBasinID = req.RiverBasinID
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("project: update: %w", err)
	}

	return s.repo.FindByID(ctx, p.ID, orgID)
}

func (s *Service) validateProjectMasterData(ctx context.Context, orgID uuid.UUID, orgUnitID, programID, sectorID, regionID, riverBasinID *uuid.UUID) error {
	checks := []struct {
		table string
		id    *uuid.UUID
	}{
		{table: "org_units", id: orgUnitID},
		{table: "programs", id: programID},
		{table: "sectors", id: sectorID},
		{table: "regions", id: regionID},
		{table: "river_basins", id: riverBasinID},
	}

	for _, check := range checks {
		if check.id == nil {
			continue
		}
		exists, err := s.repo.MasterDataExists(ctx, check.table, *check.id, orgID)
		if err != nil {
			return fmt.Errorf("project: validate master data: %w", err)
		}
		if !exists {
			return ErrInvalidProjectMasterData
		}
	}
	return nil
}

// Transition moves a project to a new status via the FSM.
func (s *Service) Transition(ctx context.Context, id, orgID uuid.UUID, req *TransitionRequest) (*Project, error) {
	p, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if _, err := s.fsm.Transition(workflow.EntityProject, p.Status, req.ToStatus); err != nil {
		return nil, fmt.Errorf("project: %w — allowed: %v",
			workflow.ErrTransitionNotAllowed,
			s.fsm.AllowedTransitions(workflow.EntityProject, p.Status),
		)
	}

	p.Status = req.ToStatus
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("project: transition update: %w", err)
	}

	return p, nil
}

// Delete soft-deletes a project and cascades the soft delete to all business
// children (tasks, subtasks, milestones, issues, risks, budgets, documents,
// team members, task assignments) in a single transaction. It verifies the
// project belongs to the caller's organization first.
func (s *Service) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	p, err := s.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteCascade(ctx, p.ID, orgID)
}

// GetProgressHistory returns the progress history for a project, verifying org boundary.
func (s *Service) GetProgressHistory(ctx context.Context, projectID, orgID uuid.UUID) ([]ProgressHistory, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.GetProgressHistory(ctx, projectID)
}
