package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
)

// CreateMilestone creates a new milestone under a project.
func (s *Service) CreateMilestone(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateMilestoneRequest) (*Milestone, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	m := &Milestone{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Title:          req.Title,
		Description:    req.Description,
		DueDate:        req.DueDate,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateMilestone(ctx, m); err != nil {
		return nil, fmt.Errorf("milestone: create: %w", err)
	}
	return s.repo.FindMilestoneByID(ctx, projectID, m.ID, orgID)
}

// GetMilestone returns a milestone by ID, verifying org boundary.
func (s *Service) GetMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) (*Milestone, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindMilestoneByID(ctx, projectID, milestoneID, orgID)
}

// ListMilestones returns all milestones for a project.
func (s *Service) ListMilestones(ctx context.Context, projectID, orgID uuid.UUID) ([]Milestone, error) {
	// Verify the project belongs to this org first
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.ListMilestones(ctx, projectID, orgID)
}

// UpdateMilestone applies partial updates to a milestone.
func (s *Service) UpdateMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID, req *UpdateMilestoneRequest) (*Milestone, error) {
	m, err := s.GetMilestone(ctx, projectID, milestoneID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		m.Title = req.Title
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityMilestone, m.Status, req.Status); err != nil {
			return nil, fmt.Errorf("milestone: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityMilestone, m.Status),
			)
		}
		m.Status = req.Status
	}
	if req.ProgressPct != nil {
		m.ProgressPct = *req.ProgressPct
	}
	if req.DueDate != nil {
		m.DueDate = req.DueDate
	}
	if err := s.repo.UpdateMilestone(ctx, m); err != nil {
		return nil, fmt.Errorf("milestone: update: %w", err)
	}
	return m, nil
}

// DeleteMilestone soft-deletes a milestone, verifying org boundary.
func (s *Service) DeleteMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) error {
	m, err := s.GetMilestone(ctx, projectID, milestoneID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteMilestone(ctx, projectID, m.ID, orgID)
}
