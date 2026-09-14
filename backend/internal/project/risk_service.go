package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
)

// CreateRisk creates a new risk under a project, enforcing org boundary.
// The risk score and severity are always derived from probability × impact.
func (s *Service) CreateRisk(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateRiskRequest) (*Risk, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	probability := clampRiskLevel(req.Probability)
	if req.Probability == 0 {
		probability = 3
	}
	impact := clampRiskLevel(req.Impact)
	if req.Impact == 0 {
		impact = 3
	}

	r := &Risk{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         "IDENTIFIED",
		Probability:    probability,
		Impact:         impact,
		RiskScore:      RiskScore(probability, impact),
		Severity:       RiskSeverity(probability, impact),
		Mitigation:     req.Mitigation,
		OwnedBy:        req.OwnedBy,
		DueDate:        req.DueDate,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateRisk(ctx, r); err != nil {
		return nil, fmt.Errorf("risk: create: %w", err)
	}
	return s.repo.FindRiskByID(ctx, projectID, r.ID, orgID)
}

// GetRisk returns a risk by ID, verifying org and project boundaries.
func (s *Service) GetRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID) (*Risk, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindRiskByID(ctx, projectID, riskID, orgID)
}

// ListRisks returns paginated risks for a project.
func (s *Service) ListRisks(ctx context.Context, filter RiskListFilter) ([]Risk, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListRisks(ctx, filter)
}

// UpdateRisk applies partial updates to a risk, enforcing FSM for status and
// recomputing risk_score/severity whenever probability or impact changes.
func (s *Service) UpdateRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID, req *UpdateRiskRequest) (*Risk, error) {
	r, err := s.GetRisk(ctx, projectID, riskID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		r.Title = req.Title
	}
	if req.Description != "" {
		r.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityRisk, r.Status, req.Status); err != nil {
			return nil, fmt.Errorf("risk: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityRisk, r.Status),
			)
		}
		r.Status = req.Status
	}
	if req.Probability > 0 {
		r.Probability = clampRiskLevel(req.Probability)
	}
	if req.Impact > 0 {
		r.Impact = clampRiskLevel(req.Impact)
	}
	if req.Probability > 0 || req.Impact > 0 {
		r.RiskScore = RiskScore(r.Probability, r.Impact)
		r.Severity = RiskSeverity(r.Probability, r.Impact)
	}
	if req.Mitigation != "" {
		r.Mitigation = req.Mitigation
	}
	if req.OwnedBy != nil {
		r.OwnedBy = req.OwnedBy
	}
	if req.DueDate != nil {
		r.DueDate = req.DueDate
	}
	if err := s.repo.UpdateRisk(ctx, r); err != nil {
		return nil, fmt.Errorf("risk: update: %w", err)
	}
	return s.repo.FindRiskByID(ctx, projectID, r.ID, orgID)
}

// DeleteRisk soft-deletes a risk, verifying org boundary.
func (s *Service) DeleteRisk(ctx context.Context, projectID, riskID, orgID uuid.UUID) error {
	r, err := s.GetRisk(ctx, projectID, riskID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteRisk(ctx, projectID, r.ID, orgID)
}
